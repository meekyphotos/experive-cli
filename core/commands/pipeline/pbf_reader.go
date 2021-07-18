package pipeline

import (
	"github.com/meekyphotos/experive-cli/core/commands/osm"
	"io"
	"os"
	"runtime"
)

func ReadFromPbf(path string, heartbeat Heartbeat) (chan map[string]interface{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	out := make(chan map[string]interface{}, 100000)
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
		d.Skip(false, true, true)
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
				m := make(map[string]interface{}, 7)
				node := v.(osm.Node)
				m["osm_id"] = node.OsmId
				m["class"] = node.Class
				m["type"] = node.Type
				m["latitude"] = node.Lat
				m["longitude"] = node.Lon
				m["metadata"] = node.Tags
				m["names"] = node.Names
				out <- m
			case *osm.Way:
			case *osm.Relation:
			default:
			}
		}
		close(out)
	}()
	return out, nil
}
