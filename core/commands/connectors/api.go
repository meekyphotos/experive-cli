package connectors

type DbType int

const (
	Snowflake       DbType = iota
	Bigint          DbType = iota
	DoublePrecision DbType = iota
	Text            DbType = iota
	Jsonb           DbType = iota
	Point           DbType = iota
)

type Column struct {
	Name    string
	Type    DbType
	Indexed bool
}

type DataConverter interface {
	Convert(data interface{}) (string, error)
}

type Connector interface {
	Connect() error
	Close()
	Init(columns []Column) error
	Write(data []map[string]interface{}) error
	CreateIndexes() error
}
