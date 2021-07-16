package utils

import "github.com/urfave/cli/v2"

type Runner interface {
	Run(c *Config) error
}

type DatabaseLoader struct {
	PasswordProvider PasswordProvider
	Config           Config
	Runner           Runner
}

func (db *DatabaseLoader) DoLoad(context *cli.Context) error {
	db.Config = ParseDbConfig(db.PasswordProvider, context)
	return db.Runner.Run(&db.Config)
}
