package commands

import (
	"github.com/urfave/cli/v2"
)

func ListMeta() *cli.Command {
	return &cli.Command{
		Name:      "list",
		Usage:     "List available datasets",
		ArgsUsage: "region to search",
		Action: func(context *cli.Context) error {
			downloader := OsmDownloader{}
			err := downloader.Init()
			if err != nil {
				panic(err)
			}
			return downloader.List(context)
		},
	}
}
