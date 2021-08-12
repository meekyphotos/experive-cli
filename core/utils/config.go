package utils

import (
	"github.com/urfave/cli/v2"
	"log"
)

type Config struct {
	File          string
	Create        bool
	DbName        string
	UserName      string
	Password      string
	Host          string
	Port          int
	UseLatLng     bool
	UseGeom       bool
	TableName     string
	InclKeyValues bool
	Schema        string
	WorkerCount   int
}

func ParseDbConfig(pwdProvider PasswordProvider, context *cli.Context) Config {
	cfg := Config{}
	if context.NArg() >= 1 {
		cfg.File = context.Args().Get(0)
	}
	cfg.Create = context.Bool("c")
	if context.Bool("a") {
		cfg.Create = false
	}
	cfg.DbName = context.String("d")
	cfg.UserName = context.String("U")
	if context.Bool("W") {
		// should prompt password and set it
		log.Println("Now, please type in the password (mandatory): ")
		pwd, _ := pwdProvider.ReadPassword()
		cfg.Password = pwd
	} else {
		cfg.Password = cfg.UserName
	}
	cfg.Host = context.String("H")
	cfg.Port = context.Int("P")
	cfg.UseGeom = true
	if context.Bool("latlong") {
		cfg.UseLatLng = true
		cfg.UseGeom = false
	}
	cfg.TableName = context.String("p")
	cfg.InclKeyValues = context.Bool("json")

	cfg.Schema = context.String("schema")
	cfg.TableName = context.String("t")
	cfg.WorkerCount = context.Int("workers")

	return cfg
}
