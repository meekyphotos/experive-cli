package osm

import (
	_ "embed"
	"fmt"
)

type LatLng struct {
	Lat float64
	Lon float64
}

func (l *LatLng) ToString() string {
	return fmt.Sprintf("SRID=4326;POINT(%f %f)", l.Lat, l.Lon)
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
