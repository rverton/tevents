package tevents

import "time"

type Event struct {
	Id        int
	Origin    string
	EventType string
	Body      string
	Owner     string
	CreatedAt time.Time
}

type EventService interface {
	Insert(origin, event_type, body, owner string) error
	Find(event_type string) ([]*Event, error)
}
