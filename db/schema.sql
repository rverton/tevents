CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    origin TEXT,
    event_type TEXT,
    body TEXT,
    owner TEXT,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS events_origin ON events (origin);
CREATE INDEX IF NOT EXISTS events_type ON events (event_type);
CREATE INDEX IF NOT EXISTS events_owner ON events (owner);
