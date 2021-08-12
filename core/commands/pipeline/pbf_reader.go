package pipeline

import (
	"github.com/meekyphotos/experive-cli/core/commands/osm"
	"io"
	"log"
	"os"
	"runtime"
)

func ReadFromPbf(path string, heartbeat Heartbeat) (chan []interface{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	outNodes := make(chan []interface{}, 100000)
	heartbeat.Start()
	go func() {
		defer f.Close()
		defer heartbeat.Done()
		open, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		d := osm.NewDecoder(open)
		d.SetBufferSize(osm.MaxBlobSize)
		d.Skip(false, false, true)
		if err := d.Start(runtime.GOMAXPROCS(-1), osm.NewFullStyler()); err != nil {
			panic(err)
		}
		for {
			v, err := d.Decode()
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}

			switch t := v.(type) {
			case *osm.Node:
				log.Println(t.Content...)
				outNodes <- t.Content
			case *osm.Way:
			case *osm.Relation:
			default:
			}
		}
		close(outNodes)
	}()
	return outNodes, nil
}
