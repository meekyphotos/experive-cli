package osm

import (
	"github.com/meekyphotos/experive-cli/core/commands/pbf"
	"google.golang.org/protobuf/proto"
	"time"
)

type dataDecoder struct {
	q             []interface{}
	skipNodes     bool
	skipWays      bool
	skipRelations bool
}

func (dec *dataDecoder) Decode(blob *pbf.Blob) ([]interface{}, error) {
	dec.q = make([]interface{}, 0, 8000) // typical PrimitiveBlock contains 8k OSM entities

	data, err := getData(blob)
	if err != nil {
		return nil, err
	}

	primitiveBlock := &pbf.PrimitiveBlock{}
	if err := proto.Unmarshal(data, primitiveBlock); err != nil {
		return nil, err
	}

	dec.parsePrimitiveBlock(primitiveBlock)
	return dec.q, nil
}

func (dec *dataDecoder) parsePrimitiveBlock(pb *pbf.PrimitiveBlock) {
	for _, pg := range pb.GetPrimitivegroup() {
		dec.parsePrimitiveGroup(pb, pg)
	}
}

func (dec *dataDecoder) parsePrimitiveGroup(pb *pbf.PrimitiveBlock, pg *pbf.PrimitiveGroup) {
	if !dec.skipNodes {
		dec.parseNodes(pb, pg.GetNodes())
		dec.parseDenseNodes(pb, pg.GetDense())
	}
	if !dec.skipWays {
		dec.parseWays(pb, pg.GetWays())
	}
	if !dec.skipRelations {
		dec.parseRelations(pb, pg.GetRelations())
	}
}

func (dec *dataDecoder) parseNodes(pb *pbf.PrimitiveBlock, nodes []*pbf.Node) {
	st := pb.GetStringtable().GetS()
	granularity := int64(pb.GetGranularity())

	latOffset := pb.GetLatOffset()
	lonOffset := pb.GetLonOffset()
	// identify unwanted keys
	for _, node := range nodes {
		id := node.GetId()
		lat := node.GetLat()
		lon := node.GetLon()

		latitude := 1e-9 * float64(latOffset+(granularity*lat))
		longitude := 1e-9 * float64(lonOffset+(granularity*lon))

		tags, names, class, osmType := ExtractInfo(st, node.GetKeys(), node.GetVals())
		if len(tags) != 2 || len(names) != 2 {
			dec.q = append(dec.q, &Node{
				OsmId: id,
				Lat:   latitude,
				Lon:   longitude,
				Tags:  tags,
				Names: names,
				Class: class,
				Type:  osmType,
			})
		}
	}

}

func (dec *dataDecoder) parseDenseNodes(pb *pbf.PrimitiveBlock, dn *pbf.DenseNodes) {
	st := pb.GetStringtable().GetS()
	granularity := int64(pb.GetGranularity())
	latOffset := pb.GetLatOffset()
	lonOffset := pb.GetLonOffset()
	ids := dn.GetId()
	lats := dn.GetLat()
	lons := dn.GetLon()

	tu := tagUnpacker{st, dn.GetKeysVals(), 0}
	var id, lat, lon int64
	for index := range ids {
		id = ids[index] + id
		lat = lats[index] + lat
		lon = lons[index] + lon
		latitude := 1e-9 * float64(latOffset+(granularity*lat))
		longitude := 1e-9 * float64(lonOffset+(granularity*lon))
		tags, names, class, osmType := tu.next()
		if len(tags) != 2 || len(names) != 2 {
			dec.q = append(dec.q, &Node{id, class, osmType, latitude, longitude, tags, names})
		}
	}
}

func (dec *dataDecoder) parseWays(pb *pbf.PrimitiveBlock, ways []*pbf.Way) {
	st := pb.GetStringtable().GetS()

	for _, way := range ways {
		id := way.GetId()

		tags := extractTags(st, way.GetKeys(), way.GetVals())

		refs := way.GetRefs()
		var nodeID int64
		nodeIDs := make([]int64, len(refs))
		for index := range refs {
			nodeID = refs[index] + nodeID // delta encoding
			nodeIDs[index] = nodeID
		}
		dec.q = append(dec.q, &Way{ID: id, Tags: tags, NodeIDs: nodeIDs})
	}
}

// Make relation members from stringtable and three parallel arrays of IDs.
func extractMembers(stringTable [][]byte, rel *pbf.Relation) []Member {
	memIDs := rel.GetMemids()
	types := rel.GetTypes()
	roleIDs := rel.GetRolesSid()

	var memID int64
	members := make([]Member, len(memIDs))
	for index := range memIDs {
		memID = memIDs[index] + memID // delta encoding

		var memType MemberType
		switch types[index] {
		case pbf.Relation_NODE:
			memType = NodeType
		case pbf.Relation_WAY:
			memType = WayType
		case pbf.Relation_RELATION:
			memType = RelationType
		}

		role := stringTable[roleIDs[index]]

		members[index] = Member{memID, memType, string(role)}
	}

	return members
}

func (dec *dataDecoder) parseRelations(pb *pbf.PrimitiveBlock, relations []*pbf.Relation) {
	st := pb.GetStringtable().GetS()
	dateGranularity := int64(pb.GetDateGranularity())

	for _, rel := range relations {
		id := rel.GetId()
		tags := extractTags(st, rel.GetKeys(), rel.GetVals())
		members := extractMembers(st, rel)
		info := extractInfo(st, rel.GetInfo(), dateGranularity)

		dec.q = append(dec.q, &Relation{id, tags, members, info})
	}
}

func extractInfo(stringTable [][]byte, i *pbf.Info, dateGranularity int64) Info {
	info := Info{Visible: true}

	if i != nil {
		info.Version = i.GetVersion()

		millisec := time.Duration(i.GetTimestamp()*dateGranularity) * time.Millisecond
		info.Timestamp = time.Unix(0, millisec.Nanoseconds()).UTC()

		info.Changeset = i.GetChangeset()

		info.Uid = i.GetUid()

		info.User = string(stringTable[i.GetUserSid()])

		if i.Visible != nil {
			info.Visible = i.GetVisible()
		}
	}

	return info
}