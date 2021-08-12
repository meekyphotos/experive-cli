package osm

import (
	"github.com/meekyphotos/experive-cli/core/commands/pbf"
	"google.golang.org/protobuf/proto"
)

type dataDecoder struct {
	q             []interface{}
	skipNodes     bool
	skipWays      bool
	skipRelations bool
	styler        *GazetteerStyler
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
	dec.parseNodes(pb, pg.GetNodes())
	dec.parseDenseNodes(pb, pg.GetDense())
	dec.parseWays(pb, pg.GetWays())
	dec.parseRelations(pb, pg.GetRelations())
}

func (dec *dataDecoder) parseNodes(pb *pbf.PrimitiveBlock, nodes []*pbf.Node) {
	if len(nodes) > 0 {
		for _, n := range nodes {
			println("Found node: ", n)
			for _, c := range dec.styler.Parse(pb, n) {
				dec.q = append(dec.q, &Node{Content: c})
			}
		}
	}

}

func (dec *dataDecoder) parseDenseNodes(pb *pbf.PrimitiveBlock, dn *pbf.DenseNodes) {
	if dn != nil {
		for _, c := range dec.styler.ParseDense(pb, dn) {
			dec.q = append(dec.q, &Node{Content: c})
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

		dec.q = append(dec.q, &Way{
			id, tags, nodeIDs,
		})
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

	for _, rel := range relations {
		id := rel.GetId()
		tags := extractTags(st, rel.GetKeys(), rel.GetVals())
		members := extractMembers(st, rel)

		dec.q = append(dec.q, &Relation{
			Content: map[string]interface{}{
				"osm_id":    id,
				"extratags": tags,
				"members":   members,
			},
		})
	}
}
