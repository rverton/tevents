package main

import (
	"database/sql"
	"flag"
	"log"
	"tevents"

	tdb "tevents/db"

	_ "modernc.org/sqlite"
	"tailscale.com/tsnet"
)

var (
	dburl    = flag.String("dburl", "db.sqlite", "database url")
	hostname = flag.String("hostname", "tevents", "hostname for the tailnet")
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

	ts := &tsnet.Server{
		Hostname: *hostname,
	}

	defer ts.Close()

	ln, err := ts.Listen("tcp", ":80")
	if err != nil {
		log.Fatal(err)
	}

	defer ln.Close()

	lc, err := ts.LocalClient()
	if err != nil {
		log.Fatal(err)
	}

	s := tevents.NewServer(":8080", db, ln, lc)
	s.EventService = tdb.NewEventService(db)

	if err := s.Start(); err != nil {
		log.Fatal("http server failed:", err)
	}

}
