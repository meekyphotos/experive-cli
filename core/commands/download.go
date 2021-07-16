package commands

import (
	"github.com/urfave/cli/v2"
)

func DownloadMeta() *cli.Command {
	return &cli.Command{
		Name:      "download",
		Usage:     "Download an extract from geofabrik",
		ArgsUsage: "region to download",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "f", Aliases: []string{"format"}, Value: "pbf", Usage: "Specify format"},
		},
		Action: func(context *cli.Context) error {
			downloader := OsmDownloader{}
			err := downloader.Init()
			if err != nil {
				panic(err)
			}
			return downloader.OsmDownload(context)
		},
	}
}
