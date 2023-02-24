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

func (es *EventService) Find(event_type string) ([]*tevents.Event, error) {
	var events []*tevents.Event

	sql := `SELECT origin, event_type, body, owner, created_at FROM events WHERE event_type = ? ORDER BY created_at`

	rows, err := es.db.Query(sql, event_type)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var e tevents.Event
		err := rows.Scan(&e.Origin, &e.EventType, &e.Body, &e.Owner, &e.CreatedAt)
		if err != nil {
			return nil, err
		}
		events = append(events, &e)
	}
	return events, rows.Err()
}

func (es *EventService) ClearAll() error {
	_, err := es.db.Exec(`DELETE FROM events`)
	return err
}
