package commands

import (
	"database/sql"
	"fmt"
	"github.com/meekyphotos/experive-cli/core/commands/pipeline"
	"github.com/meekyphotos/experive-cli/core/utils"
	"github.com/urfave/cli/v2"
	"time"
)

type OsmRunner struct {
	config *utils.Config
	db     *sql.DB
}

func (r *OsmRunner) prepare() error {
	fmt.Println("Connecting to database")
	connStr := fmt.Sprintf("user=%s dbname=%s password=%s host=%s sslmode=disable",
		r.config.UserName, r.config.DbName, r.config.Password, r.config.Host)
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	r.db = conn
	return nil
}

func (r *OsmRunner) createTables() error {
	fmt.Println("Creating table & index")
	_, err := r.db.Exec("DROP TABLE IF EXISTS place")
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`
			CREATE TABLE place (
				osm_id bigint not null,
				osm_type text not null,
				class text not null,
				type text not null,
				name jsonb, 
				admin_level smallint, 
				address jsonb,
				extratags jsonb, 
				geometry geometry(point),
				primary key (osm_id, osm_type, class)
		   )`)
	if err != nil {
		return err
	}

	_, err = r.db.Exec("CREATE INDEX place_id_idx ON place USING BTREE(osm_type, osm_id)")

	return err
}

func (r *OsmRunner) runLoading() error {
	fmt.Println("Importing data")
	pbf, err := pipeline.ReadFromPbf(r.config.File, &pipeline.NoopBeat{})
	if err != nil {
		return err
	}
	buffered := pipeline.BatchRequests(pbf, 100_000, 30*time.Second)
	channel := pipeline.CopyIn(buffered, r.db)
	<-channel
	return err
}

func (r *OsmRunner) cleanup() error {
	fmt.Println("Cleaning up")
	return r.db.Close()
}

func (r *OsmRunner) Run(c *utils.Config) error {
	r.config = c
	pipe := pipeline.Pipeline{}
	pipe.Add(
		r.prepare,
		r.createTables,
		r.runLoading,
		r.cleanup)
	return pipe.RunPipe()
}

func LoadOsmMeta() *cli.Command {
	stdAction := utils.DatabaseLoader{
		Runner:           &OsmRunner{},
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
