package pipeline

//func getConnection() *sql.DB {
//	connStr := fmt.Sprintf("user=%s dbname=%s password=%s host=%s sslmode=disable",
//		"meeky", "test", "meeky", "127.0.0.1")
//	conn, err := sql.Open("postgres", connStr)
//	if err != nil {
//		panic(err)
//	}
//	return conn
//}

//func TestCopyActuallyWorks(t *testing.T) {
//	conn := getConnection()
//	//var chn = make(chan [][]interface{})
//	//row := []interface{}{
//	//	696261627,                             // "osm_id",
//	//	"N",                                   // "osm_type",
//	//	"SRID=4326;POINT(43.731031 7.422659)", // "geometry",
//	//	"shop",                                // "class",
//	//	"gift",                                // "type",
//	//	0,                                     // "admin_level",
//	//	"{}",                                  // "name",
//	//	"{\"street\":\"Rue Basse\",\"country\":\"MC\",\"postcode\":\"98000\",\"housenumber\":\"14\"}", // "address",
//	//	"{}", // "extratags",
//	//}
//	tx, _ := conn.BeginTx(context.Background(), nil)
//	stmt, err := tx.Prepare(`COPY place ("osm_id", "osm_type", "geometry", "class", "type", "admin_level", "name", "extratags") FROM STDIN WITH (FORMAT CSV, DELIMITER '|', QUOTE '"');`)
//	if err != nil {
//		panic(err)
//	}
//	var buf = &bytes.Buffer{}
//	formats.AppendWithQuote(buf, 696261627)
//	buf.WriteString("|")
//	formats.AppendWithQuote(buf, "N")
//	buf.WriteString("|")
//	formats.AppendWithQuote(buf, "SRID=4326;POINT(43.731031 7.422659)")
//	buf.WriteString("|")
//	formats.AppendWithQuote(buf, "shop")
//	buf.WriteString("|")
//	formats.AppendWithQuote(buf, "gift")
//	buf.WriteString("|")
//	formats.AppendWithQuote(buf, 0)
//	buf.WriteString("|")
//	formats.AppendWithQuote(buf, "{\"street\":\"Rue Basse\",\"country\":\"MC\"}")
//	buf.WriteString("|")
//	formats.AppendWithQuote(buf, "{}")
//	fmt.Println(buf.String())
//	if _, err = stmt.Exec(buf.String()); err != nil {
//		panic(err)
//	}
//	if _, err = stmt.Exec(); err != nil {
//		panic(err)
//	}
//	stmt.Close()
//	tx.Commit()
//
//}
