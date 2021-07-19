package pipeline

import (
	"github.com/meekyphotos/experive-cli/core/commands/osm"
	"io"
	"os"
	"runtime"
)

func ReadFromPbf(path string, heartbeat Heartbeat) (chan map[string]interface{}, chan map[string]interface{}, chan map[string]interface{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, nil, err
	}
	outNodes := make(chan map[string]interface{}, 100000)
	outWays := make(chan map[string]interface{}, 100000)
	outRelations := make(chan map[string]interface{}, 100000)
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
		if err := d.Start(runtime.GOMAXPROCS(-1)); err != nil {
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

			switch v.(type) {
			case *osm.Node:
				node := v.(*osm.Node)
				outNodes <- node.Content
			case *osm.Way:
				node := v.(*osm.Way)
				outWays <- node.Content
			case *osm.Relation:
				node := v.(*osm.Relation)
				outRelations <- node.Content

			default:
			}
		}
		close(outNodes)
		close(outWays)
		close(outRelations)
	}()
	return outNodes, outWays, outRelations, nil
}
