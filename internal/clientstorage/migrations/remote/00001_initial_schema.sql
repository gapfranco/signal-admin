-- +goose Up
CREATE TABLE IF NOT EXISTS event (
	id TEXT PRIMARY KEY,
	local TEXT NOT NULL,
	central INTEGER NOT NULL,
	link INTEGER NOT NULL,
	device_id INTEGER NOT NULL,
	device TEXT NOT NULL,
	device_type TEXT,
	event_type TEXT NOT NULL,
	ts_unix_ms INTEGER NOT NULL,
	inst_id TEXT,
	type_id TEXT
);

CREATE TABLE IF NOT EXISTS cluster_lock (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	instance_name TEXT NOT NULL,
	last_heartbeat INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS remote_metadata (
	key TEXT PRIMARY KEY,
	value TEXT
);

-- +goose Down
DROP TABLE IF EXISTS remote_metadata;
DROP TABLE IF EXISTS cluster_lock;
DROP TABLE IF EXISTS event;
