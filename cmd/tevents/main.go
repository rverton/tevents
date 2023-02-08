package main

import (
	"database/sql"
	"flag"
	"log"
	"tevents"

	tdb "tevents/db"

	_ "modernc.org/sqlite"
)

var (
	dburl = flag.String("dburl", "db.sqlite", "database url")
)

func main() {
	flag.Parse()

	db, err := sql.Open("sqlite", *dburl)
	if err != nil {
		log.Fatal("cant connect to db", err)
	}

	if err := tdb.SetupSchema(db); err != nil {
		log.Fatal("cant setup schema", err)
	}

	s := tevents.NewServer(":8080", db)
	s.EventService = tdb.NewEventService(db)

	if err := s.Start(); err != nil {
		log.Fatal("http server failed:", err)
	}

}
