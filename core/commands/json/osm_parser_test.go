package json

import (
	"fmt"
	"github.com/meekyphotos/experive-cli/core/commands/osm"
	"io"
	"os"
	"runtime"
	"testing"
)

func (o *OsmParser) CountNodes() {
	open, err := os.Open("netherlands.osm.pbf")
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
			o.node++
			//fmt.Println(v)
		case *osm.Way:
			o.ways++
		case *osm.Relation:
		default:
		}
	}
}

// 9.991.103 in 14.69s only counting and skipping ways and relations
func TestParsing(t *testing.T) {
	parser := OsmParser{}
	parser.CountNodes()
	fmt.Println(parser.node, parser.ways)
}
