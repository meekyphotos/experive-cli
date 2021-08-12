package pipeline

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/meekyphotos/experive-cli/core/commands/formats"
	"strings"
)

var cols = []string{
	"osm_id",
	"osm_type",
	"geometry",
	"class",
	"type",
	"admin_level",
	"name",
	"address",
	"extratags",
}

func prepareCopyIn(tx *sql.Tx, table string, cols []string) (*sql.Stmt, error) {
	return tx.Prepare(fmt.Sprintf("COPY %s (%s) FROM STDIN WITH (FORMAT CSV, DELIMITER '|', QUOTE '\"')", table, strings.Join(cols, ",")))
}

func toLine(row []interface{}, separator string) string {
	buffer := &bytes.Buffer{}
	for i, r := range row {
		if i > 0 {
			buffer.WriteString(separator)
		}
		formats.AppendWithQuote(buffer, r)
	}
	return buffer.String()
}

func CopyIn(channel chan [][]interface{}, db *sql.DB) chan bool {
	done := make(chan bool)
	go func() {
		for {
			content, more := <-channel
			if more {
				if len(content) == 0 {
					continue
				}
				if tx, err := db.BeginTx(context.Background(), nil); err == nil {
					stmt, err := prepareCopyIn(tx, "place", cols)

					if err != nil {
						panic(err)
					}

					for _, row := range content {
						line := toLine(row, "|")
						if _, err = stmt.Exec(line); err != nil {
							panic(err)
						}
					}

					if _, err = stmt.Exec(); err != nil {
						panic(err)
					}
					if err = stmt.Close(); err != nil {
						panic(err)
					}

					if err = tx.Commit(); err != nil {
						panic(err)
					}
				} else {
					panic(err)
				}
			} else {
				done <- true
				close(done)
				break
			}
		}
	}()
	return done
}
