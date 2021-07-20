package pipeline

import (
	"github.com/meekyphotos/experive-cli/core/commands/osm"
	"github.com/meekyphotos/experive-cli/core/dataproviders"
	"io"
	"os"
	"runtime"
)

func ReadFromPbf(path string, heartbeat Heartbeat) (chan *dataproviders.INode, chan *osm.Way, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	outNodes := make(chan *dataproviders.INode, 100000)
	outWays := make(chan *osm.Way, 100000)
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

				outNodes <- &dataproviders.INode{
					Id:      node.Id,
					Content: node.Content,
				}
			case *osm.Way:
				node := v.(*osm.Way)
				outWays <- node
			case *osm.Relation:
			default:
			}
		}
		close(outNodes)
		close(outWays)
	}()
	return outNodes, outWays, nil
}
