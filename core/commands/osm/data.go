package osm

import (
	"time"
)

type BoundingBox struct {
	Left   float64
	Right  float64
	Top    float64
	Bottom float64
}

type Header struct {
	BoundingBox                      *BoundingBox
	RequiredFeatures                 []string
	OptionalFeatures                 []string
	WritingProgram                   string
	Source                           string
	OsmosisReplicationTimestamp      time.Time
	OsmosisReplicationSequenceNumber int64
	OsmosisReplicationBaseUrl        string
}

type Info struct {
	Version   int32
	Uid       int32
	Timestamp time.Time
	Changeset int64
	User      string
	Visible   bool
}

type Node struct {
	Content []interface{}
}

type Way struct {
	Id      int64
	Tags    map[string]string
	NodeIds []int64
}

type MemberType int

const (
	NodeType MemberType = iota
	WayType
	RelationType
)

type Member struct {
	ID   int64
	Type MemberType
	Role string
}

type Relation struct {
	Content map[string]interface{}
}
