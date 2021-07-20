package commands

import (
	"github.com/meekyphotos/experive-cli/core/commands/connectors"
	"github.com/meekyphotos/experive-cli/core/commands/pipeline"
	"github.com/meekyphotos/experive-cli/core/dataproviders"
	"github.com/meekyphotos/experive-cli/core/utils"
	"github.com/urfave/cli/v2"
	"github.com/valyala/fastjson"
	"os"
	"sync"
	"time"
)

type OsmRunner struct {
	store         dataproviders.Store
	NodeConnector connectors.Connector
}

// 	ID      int64
//	Tags    map[string]string
//	NodeIDs []int64

var osmFields = []connectors.Column{
	{Name: "osm_id", Type: connectors.Bigint, Indexed: true},
	{Name: "osm_type", Type: connectors.Text},
	{Name: "class", Type: connectors.Text},
	{Name: "type", Type: connectors.Text},
	{Name: "name", Type: connectors.Jsonb},
	{Name: "address", Type: connectors.Jsonb},
}

func determineNodesCols(c *utils.Config) []connectors.Column {
	cols := make([]connectors.Column, 0)
	cols = append(cols, osmFields...)
	if c.InclKeyValues {
		cols = append(cols, connectors.Column{Name: "extratags", Type: connectors.Jsonb})
	}
	if c.UseGeom {
		cols = append(cols, geomFields...)
	} else {
		cols = append(cols, latLngFields...)
	}
	return cols
}

func (r OsmRunner) Run(c *utils.Config) error {
	pg := &connectors.PgConnector{
		Config: c, TableName: c.TableName + "_node",
	}
	r.NodeConnector = pg
	dbErr := r.NodeConnector.Connect()
	if dbErr != nil {
		return dbErr
	}
	r.store = dataproviders.Store{}
	os.RemoveAll("./db.tmp") // try to delete all to cleanup previous run
	r.store.Open("./db.tmp")
	defer func() {
		os.RemoveAll("./db.tmp")
	}()
	defer r.store.Close()
	// no need to close ways & relations
	defer r.NodeConnector.Close()
	dbErr = r.NodeConnector.Init(determineNodesCols(c))
	if dbErr != nil {
		return dbErr
	}

	nodeChannel, wayChannel, err := pipeline.ReadFromPbf(c.File, &pipeline.NoopBeat{})
	if err != nil {
		return err
	}
	nodeRequests := pipeline.BatchINodes(nodeChannel, 10000, time.Second)
	var postProcessingWorkers sync.WaitGroup

	nodeBeat := &pipeline.ProgressBarBeat{OperationName: "Nodes"}
	waysBeat := &pipeline.ProgressBarBeat{OperationName: "Ways"}
	actualBeat := &pipeline.ProgressBarBeat{OperationName: "Node written"}

	postProcessingWorkers.Add(1)
	go func() {
		err := pipeline.ProcessINodes(nodeRequests, r.store, nodeBeat)
		if err != nil {
			panic(err)
		}
		postProcessingWorkers.Done()
	}()

	postProcessingWorkers.Add(1)
	go func() {
		pipeline.ProcessNodeEnrichment(wayChannel, r.store, waysBeat)
		postProcessingWorkers.Done()
	}()

	postProcessingWorkers.Wait()

	// I'm completely done with post processing.. now I should start writing stuff
	storeChannel := r.store.Stream(func(value *fastjson.Value) map[string]interface{} {
		baseObject := map[string]interface{}{
			"osm_id":    value.GetInt64("osm_id"),
			"osm_type":  string(value.GetStringBytes("osm_type")),
			"class":     string(value.GetStringBytes("class")),
			"type":      string(value.GetStringBytes("type")),
			"latitude":  value.GetFloat64("latitude"),
			"longitude": value.GetFloat64("longitude"),
			"name":      "",
			"address":   "",
			"extratags": "",
		}
		name := value.GetObject("name")
		if name != nil {
			baseObject["name"] = string(name.MarshalTo([]byte{}))
		}
		address := value.GetObject("address")
		if address != nil {
			baseObject["address"] = string(address.MarshalTo([]byte{}))
		}

		extratags := value.GetObject("extratags")
		if extratags != nil {
			baseObject["extratags"] = string(extratags.MarshalTo([]byte{}))
		}
		return baseObject
	})
	rowsChannel := pipeline.BatchRequest(storeChannel, 10000, time.Second)

	var pgWorker sync.WaitGroup
	pgWorker.Add(1)
	go func() {
		err := pipeline.ProcessChannel(rowsChannel, r.NodeConnector, actualBeat)
		if err != nil {
			panic(err)
		}
		pgWorker.Done()
	}()
	pgWorker.Wait()
	return r.NodeConnector.CreateIndexes()
}

func LoadOsmMeta() *cli.Command {
	stdAction := utils.DatabaseLoader{
		Runner:           OsmRunner{},
		PasswordProvider: utils.TerminalPasswordReader{},
	}

	return &cli.Command{
		Name:        "osm",
		Usage:       "Load a osm dataset into target postgres",
		Description: "Load a osm dataset into target postgres",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "a", Aliases: []string{"append"}, Value: false, Usage: "Run in append mode. Adds the OSM change file into the database without removing existing data."},
			&cli.BoolFlag{Name: "c", Aliases: []string{"create"}, Value: true, Usage: "Run in create mode. This is the default if -a, --append is not specified. Removes existing data from the database tables!"},
			// DATABASE OPTIONS
			&cli.StringFlag{Name: "d", Aliases: []string{"database"}, Value: "", Required: true, Usage: "The name of the PostgreSQL database to connect to. If this parameter contains an = sign or starts with a valid URI prefix (postgresql:// or postgres://), it is treated as a conninfo string. See the PostgreSQL manual for details."},
			&cli.StringFlag{Name: "U", Aliases: []string{"username"}, Value: "postgres", Usage: "Postgresql user name."},
			&cli.BoolFlag{Name: "W", Aliases: []string{"password"}, Value: false, Usage: "Force password prompt."},
			&cli.StringFlag{Name: "H", Aliases: []string{"host"}, Value: "localhost", Usage: "Database server hostname or unix domain socket location."},
			&cli.IntFlag{Name: "P", Aliases: []string{"port"}, Value: 5432, Usage: "Database server port."},
			&cli.IntFlag{Name: "workers", Value: 4, Usage: "Number of workers"},

			// OUTPUT FORMAT
			&cli.BoolFlag{Name: "latlong", Value: false, Usage: "Store coordinates in degrees of latitude & longitude."},
			&cli.StringFlag{Name: "t", Aliases: []string{"table"}, Value: "planet_osm", Usage: "Prefix of table"},

			&cli.BoolFlag{Name: "j", Aliases: []string{"json"}, Value: true, Usage: "Add tags without column to an additional json (key/value) column in the database tables."},

			&cli.StringFlag{Name: "schema", Value: "public", Usage: "Use PostgreSQL schema SCHEMA for all tables, indexes, and functions in the pgsql output (default is no schema, i.e. the public schema is used)."},
		},
		UseShortOptionHandling: true,
		Action:                 stdAction.DoLoad,
	}
}
