package osm

import (
	"bytes"
	"strings"
)

var classDefiningAttributes = [][]byte{
	[]byte("tourism"),
	[]byte("leisure"),
	[]byte("shop"),
	[]byte("amenity"),
	[]byte("place"),
	[]byte("highway"),
	[]byte("man_made"),
	[]byte("railway"),
	[]byte("office"),
	[]byte("historic"),
	[]byte("aeroway"),
	[]byte("boundary"),
	[]byte("emergency"),
	[]byte("landuse"),
	[]byte("natural"),
	[]byte("club"),
	[]byte("craft"),
	[]byte("tunnel"),
	[]byte("public_transport"),
}
var ignoredTags = [][]byte{
	[]byte("created_by"),
	[]byte("source"),
	[]byte("source:website"),
	[]byte("ref"),
	[]byte("crossing"),
	[]byte("crossing_ref"),
	[]byte("traffic_signals"),
}

func extractTags(stringTable [][]byte, keyIDs, valueIDs []uint32) map[string]string {
	tags := make(map[string]string, len(keyIDs))
	for index, keyID := range keyIDs {
		key := string(stringTable[keyID])

		val := string(stringTable[valueIDs[index]])
		tags[key] = val
	}
	return tags
}

func ExtractInfo(stringTable [][]byte, keyIDs, valueIDs []uint32) (string, string, string, string) {
	tags := make(map[string]string)
	names := make(map[string]string)
	var class, osmType string

keyLoop:
	for index, keyID := range keyIDs {
		keyBytes := stringTable[keyID]
		for _, b := range ignoredTags {
			if bytes.Equal(b, keyBytes) {
				continue keyLoop
			}
		}
		for _, b := range classDefiningAttributes {
			if bytes.Equal(b, keyBytes) {
				class = string(b)
				osmType = string(stringTable[valueIDs[index]])
				continue keyLoop
			}
		}
		key := string(keyBytes)
		val := string(stringTable[valueIDs[index]])
		if strings.Contains(key, "name") {
			names[key] = val
		} else {
			tags[key] = val
		}
	}
	return "{}", "{}", class, osmType
}

type tagUnpacker struct {
	stringTable [][]byte
	keysVals    []int32
	index       int
}

var openPar = []byte(`{`)
var openWithComma = []byte(`,"`)
var keyVal = []byte(`":"`)
var quotes = []byte(`"`)
var endPar = []byte(`}`)
var nameBytes = []byte(`name`)

// Make tags map from stringtable and array of IDs (used in DenseNodes encoding).
func (tu *tagUnpacker) next() (string, string, string, string) {
	var class, osmType string
	tagsJson := strings.Builder{}
	nameJson := strings.Builder{}
	tagsJson.Write(openPar)
	nameJson.Write(openPar)
	firstName := true
	firstTag := true
keyLoop:
	for tu.index < len(tu.keysVals) {
		keyID := tu.keysVals[tu.index]
		tu.index++
		if keyID == 0 {
			break
		}
		valID := tu.keysVals[tu.index]
		tu.index++

		keyBytes := tu.stringTable[keyID]
		for _, b := range ignoredTags {
			if bytes.Equal(b, keyBytes) {
				continue keyLoop
			}
		}

		valBytes := tu.stringTable[valID]
		for _, b := range classDefiningAttributes {
			if bytes.Equal(b, keyBytes) {
				class = string(b)
				osmType = string(valBytes)
				continue keyLoop
			}
		}

		if bytes.Contains(keyBytes, nameBytes) {
			if !firstName {
				nameJson.Write(openWithComma)
			} else {
				firstName = false
				nameJson.Write(quotes)
			}
			nameJson.Write(keyBytes)
			nameJson.Write(keyVal)
			nameJson.Write(valBytes)
			nameJson.Write(quotes)
		} else {
			if !firstTag {
				tagsJson.Write(openWithComma)
			} else {
				firstTag = false
				tagsJson.Write(quotes)
			}
			tagsJson.Write(keyBytes)
			tagsJson.Write(keyVal)
			tagsJson.Write(valBytes)
			tagsJson.Write(quotes)

		}
	}
	tagsJson.Write(endPar)
	nameJson.Write(endPar)
	return tagsJson.String(), nameJson.String(), class, osmType
}
