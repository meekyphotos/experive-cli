package commands

import (
	"fmt"
	"github.com/urfave/cli/v2"
)

func LoadOsmMeta() *cli.Command {
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
			&cli.StringFlag{Name: "t", Aliases: []string{"table"}, Value: "planet_data", Usage: "Output table name"},

			&cli.BoolFlag{Name: "j", Aliases: []string{"json"}, Value: false, Usage: "Add tags without column to an additional json (key/value) column in the database tables."},

			&cli.StringFlag{Name: "schema", Value: "public", Usage: "Use PostgreSQL schema SCHEMA for all tables, indexes, and functions in the pgsql output (default is no schema, i.e. the public schema is used)."},
		},
		UseShortOptionHandling: true,
		Action: func(context *cli.Context) error {
			fmt.Println("work in progress")
			return nil
		},
	}
}
