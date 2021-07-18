package connectors

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/godruoyi/go-snowflake"
	"github.com/lib/pq"
	"github.com/meekyphotos/experive-cli/core/utils"
	"regexp"
	"strings"
)

type PgConnector struct {
	Config    *utils.Config
	TableName string
	db        *sql.DB
	columns   []Column
}

func (p *PgConnector) Connect() error {
	config := p.Config
	connStr := fmt.Sprintf("user=%s dbname=%s password=%s host=%s sslmode=disable",
		config.UserName, config.DbName, config.Password, config.Host)
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	p.db = conn
	return nil
}

func (p *PgConnector) Close() {
	err := p.db.Close()
	if err != nil {
		panic(err)
	}
}

var carriageReturn = regexp.MustCompile("[\n\r]")

func (p *PgConnector) Write(data []map[string]interface{}) error {
	if p.columns == nil {
		return errors.New("no columns found, call init before starting")
	}
	txn, err := p.db.Begin()
	if err != nil {
		return err
	}
	columnNames := make([]string, len(p.columns))
	for i, c := range p.columns {
		columnNames[i] = c.Name
	}
	stmt, err := txn.Prepare(pq.CopyIn(p.TableName,
		columnNames...,
	))
	if err != nil {
		return err
	}
	for _, row := range data {
		vals := make([]interface{}, len(p.columns))
		for i, c := range p.columns {
			switch c.Type {
			case Snowflake:
				vals[i] = snowflake.ID()
			case Bigint:
				vals[i] = row[c.Name]
			case DoublePrecision:
				vals[i] = row[c.Name]
			case Text:
				vals[i] = row[c.Name]
			case Jsonb:
				value := row[c.Name]
				switch value.(type) {
				case string: // already marshalled
					repairedString := string(carriageReturn.ReplaceAll([]byte(value.(string)), []byte{}))
					vals[i] = repairedString
				default:
					marshal, err := json.Marshal(value)
					if err != nil {
						return err
					}
					vals[i] = string(marshal)
				}
			case Point:
				if lat, ok := row["latitude"]; ok {
					if lng, ok := row["longitude"]; ok {
						vals[i] = fmt.Sprintf("SRID=4326;POINT(%f %f)", lat, lng)
					}
				}
			}
		}
		_, err := stmt.Exec(vals...)
		if err != nil {
			return err
		}
	}
	_, err = stmt.Exec()
	if err != nil {
		return err
	}
	return txn.Commit()
}

func (p *PgConnector) Init(columns []Column) error {
	p.columns = columns
	txn, err := p.db.Begin()
	if err != nil {
		return err
	}

	if p.Config.Create {
		_, err := txn.Exec("DROP TABLE IF EXISTS " + p.Config.Schema + "." + p.Config.TableName)
		if err != nil {
			return err
		}
	}
	stmt := strings.Builder{}
	stmt.WriteString("CREATE TABLE IF NOT EXISTS ")
	stmt.WriteString(p.Config.Schema)
	stmt.WriteString(".")
	stmt.WriteString(p.Config.TableName)
	stmt.WriteString("\n(")
	for i, f := range columns {
		stmt.WriteString(f.Name)
		stmt.WriteString(" ")
		switch f.Type {
		case Snowflake:
			stmt.WriteString("bigint")
		case Bigint:
			stmt.WriteString("bigint")
		case DoublePrecision:
			stmt.WriteString("double precision")
		case Text:
			stmt.WriteString("text")
		case Jsonb:
			stmt.WriteString("jsonb")
		case Point:
			stmt.WriteString("geography(POINT)")
		}

		if f.Name == "id" {
			stmt.WriteString(" primary key")
		}
		if i < len(columns)-1 {
			stmt.WriteString(", \n")
		}
	}
	stmt.WriteString("\n)")
	sqlStatement := stmt.String()
	_, err = txn.Exec(sqlStatement)
	if err != nil {
		return txn.Rollback()
	}
	return txn.Commit()
}

func (p *PgConnector) CreateIndexes() error {
	if p.columns == nil {
		return errors.New("no columns found, call init before starting")
	}
	txn, err := p.db.Begin()
	if err != nil {
		return err
	}
	for _, c := range p.columns {
		if c.Indexed {
			if c.Type == Point {
				_, err := txn.Exec("CREATE INDEX ON " + p.Config.Schema + "." + p.Config.TableName + " using BRIN (" + c.Name + ")")
				if err != nil {
					_ = txn.Rollback()
					return err
				}
			} else {
				_, err := txn.Exec("CREATE INDEX ON " + p.Config.Schema + "." + p.Config.TableName + " (" + c.Name + ")")
				if err != nil {
					_ = txn.Rollback()
					return err
				}
			}

		}
	}
	return txn.Commit()
}
