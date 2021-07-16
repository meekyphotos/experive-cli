package initializers

import (
	"github.com/meekyphotos/experive-cli/core/commands"
	"github.com/urfave/cli/v2"
)

func InitApp() *cli.App {
	return &cli.App{
		Name:        "experive",
		Description: "Data management tool",
		Commands: []*cli.Command{
			commands.DownloadMeta(),
			commands.ListMeta(),
			{
				Name: "load",
				Subcommands: []*cli.Command{
					commands.LoadWofMeta(),
					commands.LoadOsmMeta(),
				},
			},
		},
		UseShortOptionHandling: true,
	}
}
