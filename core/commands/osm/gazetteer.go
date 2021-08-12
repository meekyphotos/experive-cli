package osm

import (
	"bytes"
	_ "embed"
	"github.com/meekyphotos/experive-cli/core/commands/formats"
	"github.com/meekyphotos/experive-cli/core/commands/pbf"
	"github.com/meekyphotos/experive-cli/core/utils/patterns"
	"strconv"
)

type GazetteerStyler struct {
	rules []patterns.Rule
}

//go:embed import-full.style
var Style []byte

func NewFullStyler() *GazetteerStyler {
	styler := &GazetteerStyler{}
	styler.LoadBytes(Style)
	return styler
}

func (g *GazetteerStyler) Load(path string) {
	styleJson, err := patterns.ParseRules(path)
	if err != nil {
		panic(err)
	}
	rules := make([]patterns.Rule, 0, len(styleJson))
	for _, s := range styleJson {
		rules = append(rules, *patterns.NewRule(s))
	}
	g.rules = rules
}

func (g *GazetteerStyler) LoadBytes(path []byte) {
	styleJson, err := patterns.ParseStringRules(path)
	if err != nil {
		panic(err)
	}
	rules := make([]patterns.Rule, 0, len(styleJson))
	for _, s := range styleJson {
		rules = append(rules, *patterns.NewRule(s))
	}
	g.rules = rules
}

type clause struct {
	name  []byte
	value []byte
}

const (
	addrLen = 5
	isInLen = 6
)

type template struct {
	Name       formats.Json
	Refs       formats.Json
	AdminLevel int
	Address    formats.Json
	ExtraTags  formats.Json
}

func (g *GazetteerStyler) ParseDense(pb *pbf.PrimitiveBlock, dn *pbf.DenseNodes) [][]interface{} {
	lats := dn.GetLat()
	lons := dn.GetLon()
	lenKeys := len(dn.KeysVals)
	index := 0
	out := make([][]interface{}, 0, 10)
	var id, lat, lon int64
	for i, currId := range dn.Id {
		id = currId + id
		keys := make([]uint32, 0, 10)
		vals := make([]uint32, 0, 10)
		if dn.KeysVals[index] == 0 {
			index++
		} else {
			for index < lenKeys && dn.KeysVals[index] != 0 {
				keys = append(keys, uint32(dn.KeysVals[index]))
				vals = append(vals, uint32(dn.KeysVals[index+1]))
				index += 2
			}
			index++
		}
		lat = lats[i] + lat
		lon = lons[i] + lon
		out = append(out, g.Parse(pb, &pbf.Node{
			Id:   &id,
			Keys: keys,
			Vals: vals,
			Info: nil,
			Lat:  &lat,
			Lon:  &lon,
		})...)
	}
	return out
}

func (g *GazetteerStyler) findFlag(k []byte, v string) int {
	for _, r := range g.rules {
		if r.Matches(k) {
			return r.GetFlag(v)
		}
	}
	return 0
}

func (g *GazetteerStyler) Parse(pb *pbf.PrimitiveBlock, node *pbf.Node) [][]interface{} {
	granularity := int64(pb.GetGranularity())
	latlng := LatLng{
		Lat: 1e-9 * float64(pb.GetLatOffset()+(granularity*node.GetLat())),
		Lon: 1e-9 * float64(pb.GetLonOffset()+(granularity*node.GetLon())),
	}
	initial := []interface{}{*node.Id, "N", latlng.ToString()}

	place := template{
		Name:      formats.NewJson(),
		Address:   formats.NewJson(),
		ExtraTags: formats.NewJson(),
		Refs:      formats.NewJson(),
	}
	stringTable := pb.Stringtable.S
	classesWithNameClauses := make([]clause, 0, 3)
	classesWithNamedKeyClauses := make([]clause, 0, 3)
	classesClauses := make([]clause, 0, 3)
	var fallback *clause
	for index, k := range node.Keys {
		keyBytes := stringTable[k]
		value := string(stringTable[node.Vals[index]])
		valueBytes := stringTable[node.Vals[index]]
		if bytes.Equal(keyBytes, adminLevelKey) {
			i, parse := strconv.Atoi(value)
			if parse == nil {
				place.AdminLevel = i
			}
		}

		flag := g.findFlag(keyBytes, value)
		if flag == 0 {
			continue
		}

		if flag&patterns.SfMainNamed == patterns.SfMainNamed {
			classesWithNameClauses = append(classesWithNameClauses, clause{name: keyBytes, value: valueBytes})
		} else if flag&patterns.SfMainNamedKey == patterns.SfMainNamedKey {
			classesWithNamedKeyClauses = append(classesWithNamedKeyClauses, clause{name: keyBytes, value: valueBytes})
		} else if flag&patterns.SfMain == patterns.SfMain {
			classesClauses = append(classesClauses, clause{name: keyBytes, value: valueBytes})
		} else if flag&patterns.SfMainFallback == patterns.SfMainFallback && fallback == nil {
			fallback = &clause{
				name:  keyBytes,
				value: valueBytes,
			}
		}
		if flag&patterns.SfName == patterns.SfName {
			place.Name.Add(keyBytes, valueBytes)
		}
		if flag&patterns.SfRef == patterns.SfRef {
			place.Refs.Add(keyBytes, valueBytes)
		}
		if flag&patterns.SfAddress == patterns.SfAddress {
			place.Address.Add(cleanAddressPrefix(keyBytes), value)
		}
		if flag&patterns.SfHouse == patterns.SfHouse {
			fallback = &clause{name: placeKey, value: houseKey}
		}
		if flag&patterns.SfPostcode == patterns.SfPostcode {
			place.Address.Add(postcodeKey, valueBytes)
		}
		if flag&patterns.SfCountry == patterns.SfCountry && len(valueBytes) == 2 {
			place.Address.Add(countryKey, valueBytes)
		}
		if flag&patterns.SfExtra == patterns.SfExtra {
			place.ExtraTags.Add(keyBytes, valueBytes)
		}

	}

	hasClauses := len(classesClauses)+len(classesWithNameClauses)+len(classesWithNamedKeyClauses) > 0

	out := make([][]interface{}, 0, 10)
	for _, c := range classesClauses {
		row := make([]interface{}, 0, 9)
		row = append(row, initial...)
		row = append(row,
			string(c.name),
			string(c.value),
			place.AdminLevel,
			place.Name.ToString(),
			place.Address.ToString(),
			place.ExtraTags.ToString(),
		)
		out = append(out, row)
	}
	if place.Name.HasContent() {
		for _, c := range classesWithNameClauses {
			row := make([]interface{}, 0, 9)
			row = append(row, initial...)
			row = append(row,
				string(c.name),
				string(c.value),
				place.AdminLevel,
				place.Name.ToString(),
				place.Address.ToString(),
				place.ExtraTags.ToString(),
			)
			out = append(out, row)

		}
	}
	if fallback != nil && !hasClauses {
		row := make([]interface{}, 0, 9)
		row = append(row, initial...)
		row = append(row,
			string(fallback.name),
			string(fallback.value),
			place.AdminLevel,
			place.Name.ToString(),
			place.Address.ToString(),
			place.ExtraTags.ToString(),
		)
		out = append(out, row)

	}
	return out
}

func cleanAddressPrefix(key []byte) []byte {
	if bytes.HasPrefix(key, addrPrefix) {
		return key[addrLen:]
	}
	if bytes.HasPrefix(key, isInPrefix) {
		return key[isInLen:]
	}
	return key
}

var addrPrefix = []byte(`addr:`)
var isInPrefix = []byte(`is_in:`)
var placeKey = []byte(`place`)
var postcodeKey = []byte(`postcode`)
var houseKey = []byte(`house`)
var countryKey = []byte(`country`)
var adminLevelKey = []byte(`admin_level`)
