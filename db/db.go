package db

import (
	"database/sql"
	_ "embed"
	"strings"
	"tevents"
)

//go:embed schema.sql
var DBSchema string

func SetupSchema(db *sql.DB) error {
	statements := strings.Split(DBSchema, ";")

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}

type EventService struct {
	db *sql.DB
}

var _ tevents.EventService = (*EventService)(nil)

func NewEventService(db *sql.DB) *EventService {
	return &EventService{db}
}

func (es *EventService) Insert(origin, event_type, body, owner string) error {
	_, err := es.db.Exec(`INSERT INTO events 
        (origin, event_type, body, owner)
    VALUES (?, ?, ?, ?)`, origin, event_type, body, owner)
	return err
}

func (es *EventService) Find() ([]*tevents.Event, error) {
	var events []*tevents.Event

	rows, err := es.db.Query(`SELECT origin, event_type, body, owner  FROM events ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var e tevents.Event
		err := rows.Scan(&e.Origin, &e.EventType, &e.Body, &e.Owner)
		if err != nil {
			return nil, err
		}
		events = append(events, &e)
	}
	return events, rows.Err()
}
