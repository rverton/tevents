package tevents

import (
	"testing"
	"time"
)

func TestMonitorMap(t *testing.T) {
	now := time.Date(2000, 1, 1, 1, 30, 0, 0, time.UTC)
	events := []*Event{
		{Origin: "event2", CreatedAt: now.Add(-1 * time.Hour)},
		{Origin: "event3", CreatedAt: now.Add(-2 * time.Hour)},
	}

	// get results for the last 6 hours
	results := MonitorMap(now, events, 6)
	l := len(results)

	if results[l-1] != false || results[l-2] != true || results[l-3] != true {
		t.Error("monitor map is wrong", results)
	}
}

func TestMonitorMapBounds(t *testing.T) {
	now := time.Date(2000, 1, 1, 1, 0, 0, 0, time.UTC)
	events := []*Event{
		{Origin: "event1", CreatedAt: now},
		{Origin: "event2", CreatedAt: now.Add(-7 * time.Hour)},
	}

	// get results for the last 6 hours
	results := MonitorMap(now, events, 6)

	if results[len(results)-1] != true {
		t.Error("monitor map bound is wrong")
	}

	if results[0] != false {
		t.Error("monitor map bound is wrong")
	}
}
