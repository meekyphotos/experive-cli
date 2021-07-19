package osm

import (
	"bytes"
	"regexp"
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
var addressBytes = []byte(`addr:`)
var carriageReturn = regexp.MustCompile(`[\n\r\t"\\]`)
var escapeQuote = regexp.MustCompile(`"`)

type json struct {
	started bool
	buffer  *strings.Builder
}

func newJson() json {
	return json{
		buffer: &strings.Builder{},
	}
}
func (js *json) close() {
	if js.started {
		js.buffer.Write(endPar)
	}
}

func (js *json) toString() string {
	return js.buffer.String()
}
func (js *json) add(key []byte, val []byte) {
	if js.started {
		js.buffer.Write(openWithComma)
	} else {
		js.started = true
		js.buffer.Write(openPar)
		js.buffer.Write(quotes)
	}
	js.buffer.Write(key)
	js.buffer.Write(keyVal)
	cleaned := carriageReturn.ReplaceAll(val, []byte{})
	js.buffer.Write(cleaned)
	js.buffer.Write(quotes)
}

// Make tags map from stringtable and array of IDs (used in DenseNodes encoding).
func (tu *tagUnpacker) next() (string, string, string, string, string) {
	var class, osmType string
	tagsJson := newJson()
	nameJson := newJson()
	addressJson := newJson()
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
				break // add key anyway
			}
		}

		if bytes.Contains(keyBytes, nameBytes) {
			nameJson.add(keyBytes, valBytes)
		} else if bytes.HasPrefix(keyBytes, addressBytes) {
			addressJson.add(keyBytes, valBytes)
		} else {
			tagsJson.add(keyBytes, valBytes)
		}
	}
	tagsJson.close()
	nameJson.close()
	addressJson.close()
	return tagsJson.toString(), nameJson.toString(), addressJson.toString(), class, osmType
}
