package commands

import (
	"github.com/meekyphotos/experive-cli/core/commands/connectors"
	"github.com/meekyphotos/experive-cli/core/commands/pipeline"
	"github.com/meekyphotos/experive-cli/core/utils"
	"github.com/urfave/cli/v2"
	"sync"
	"time"
)

type OsmRunner struct {
	NodeConnector      connectors.Connector
	WaysConnector      connectors.Connector
	RelationsConnector connectors.Connector
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

var wayFields = []connectors.Column{
	{Name: "osm_id", Type: connectors.Bigint, Indexed: true},
	{Name: "extratags", Type: connectors.Jsonb},
	{Name: "node_ids", Type: connectors.Jsonb},
}

var relationFields = []connectors.Column{
	{Name: "osm_id", Type: connectors.Bigint, Indexed: true},
	{Name: "extratags", Type: connectors.Jsonb},
	{Name: "members", Type: connectors.Jsonb},
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

	r.WaysConnector = &connectors.PgConnector{
		Config: c, TableName: c.TableName + "_ways", Db: pg.Db,
	}
	r.RelationsConnector = &connectors.PgConnector{
		Config: c, TableName: c.TableName + "_relations", Db: pg.Db,
	}
	// no need to close ways & relations
	defer r.NodeConnector.Close()
	dbErr = r.NodeConnector.Init(determineNodesCols(c))
	if dbErr != nil {
		return dbErr
	}
	dbErr = r.WaysConnector.Init(wayFields)
	if dbErr != nil {
		return dbErr
	}
	dbErr = r.RelationsConnector.Init(relationFields)
	if dbErr != nil {
		return dbErr
	}

	nodeChannel, wayChannel, relationChannel, err := pipeline.ReadFromPbf(c.File, &pipeline.NoopBeat{})
	if err != nil {
		return err
	}
	nodeRequests := pipeline.BatchRequest(nodeChannel, 10000, time.Second)
	wayRequests := pipeline.BatchRequest(wayChannel, 10000, time.Second)
	relationRequests := pipeline.BatchRequest(relationChannel, 10000, time.Second)
	var pgWorkers sync.WaitGroup

	nodeBeat := &pipeline.ProgressBarBeat{OperationName: "Nodes"}
	relationBeat := &pipeline.ProgressBarBeat{OperationName: "Relations"}
	waysBeat := &pipeline.ProgressBarBeat{OperationName: "Ways"}

	pgWorkers.Add(1)
	go func() {
		err := pipeline.ProcessChannel(nodeRequests, r.NodeConnector, nodeBeat)
		if err != nil {
			panic(err)
		}
		pgWorkers.Done()
	}()

	pgWorkers.Add(1)
	go func() {
		err := pipeline.ProcessChannel(wayRequests, r.WaysConnector, waysBeat)
		if err != nil {
			panic(err)
		}
		pgWorkers.Done()
	}()

	pgWorkers.Add(1)
	go func() {
		err := pipeline.ProcessChannel(relationRequests, r.RelationsConnector, relationBeat)
		if err != nil {
			panic(err)
		}
		pgWorkers.Done()
	}()

	pgWorkers.Wait()
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
