package commands

import (
	"github.com/meekyphotos/experive-cli/core/commands/connectors"
	"github.com/meekyphotos/experive-cli/core/commands/pipeline"
	"github.com/meekyphotos/experive-cli/core/utils"
	"github.com/urfave/cli/v2"
	"github.com/valyala/fastjson"
	"sync"
	"time"
)

type WofRunner struct {
	Connector connectors.Connector
}

var fields = []connectors.Column{
	{Name: "id", Type: connectors.Snowflake},
	{Name: "wof_id", Type: connectors.Bigint, Indexed: true},
	{Name: "continent_id", Type: connectors.Bigint},
	{Name: "country_id", Type: connectors.Bigint},
	{Name: "country_code", Type: connectors.Text},
	{Name: "county_id", Type: connectors.Bigint},
	{Name: "locality_id", Type: connectors.Bigint},
	{Name: "region_id", Type: connectors.Bigint},
	{Name: "preferred_names", Type: connectors.Jsonb},
	{Name: "variant_names", Type: connectors.Jsonb},
}

var latLngFields = []connectors.Column{
	{Name: "latitude", Type: connectors.DoublePrecision},
	{Name: "longitude", Type: connectors.DoublePrecision},
}

var geomFields = []connectors.Column{
	{Name: "geometry", Type: connectors.Point, Indexed: true},
}

func determineCols(c *utils.Config) []connectors.Column {
	cols := make([]connectors.Column, 0)
	cols = append(cols, fields...)
	if c.InclKeyValues {
		cols = append(cols, connectors.Column{Name: "metadata", Type: connectors.Jsonb})
	}
	if c.UseGeom {
		cols = append(cols, geomFields...)
	} else {
		cols = append(cols, latLngFields...)
	}
	return cols
}

func (r WofRunner) Run(c *utils.Config) error {
	r.Connector = &connectors.PgConnector{
		Config: c, TableName: c.TableName,
	}
	dbErr := r.Connector.Connect()
	if dbErr != nil {
		return dbErr
	}
	defer r.Connector.Close()
	dbErr = r.Connector.Init(determineCols(c))
	if dbErr != nil {
		return dbErr
	}
	channel, err := pipeline.ReadFromTar(c.File, &pipeline.NoopBeat{})
	if err != nil {
		return err
	}
	pool := fastjson.ParserPool{}
	channelOut, jsonWorkers := pipeline.GeojsonParser(channel, c, &pool)
	requests := pipeline.BatchRequest(channelOut, 10000, time.Second)
	var pgWorkers sync.WaitGroup
	pgWorkers.Add(1)
	beat := &pipeline.ProgressBarBeat{OperationName: "Writing"}
	go func() {
		err := pipeline.ProcessChannel(requests, r.Connector, beat)
		if err != nil {
			panic(err)
		}
		pgWorkers.Done()
	}()
	jsonWorkers.Wait()
	close(requests)
	pgWorkers.Wait()
	return r.Connector.CreateIndexes()
}

func LoadWofMeta() *cli.Command {
	stdAction := utils.DatabaseLoader{
		Runner:           WofRunner{},
		PasswordProvider: utils.TerminalPasswordReader{},
	}

	return &cli.Command{
		Name:        "wof",
		Usage:       "Load a whosonfirst dataset into target postgres",
		Description: "Load a whosonfirst dataset into target postgres",
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
			&cli.StringFlag{Name: "t", Aliases: []string{"table"}, Value: "planet_data", Usage: "Output table name"},

			&cli.BoolFlag{Name: "j", Aliases: []string{"json"}, Value: false, Usage: "Add tags without column to an additional json (key/value) column in the database tables."},

			&cli.StringFlag{Name: "schema", Value: "public", Usage: "Use PostgreSQL schema SCHEMA for all tables, indexes, and functions in the pgsql output (default is no schema, i.e. the public schema is used)."},
		},
		UseShortOptionHandling: true,
		Action:                 stdAction.DoLoad,
	}
}
