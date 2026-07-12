-- +goose Up
CREATE TABLE IF NOT EXISTS antena_users (
	id INTEGER PRIMARY KEY,
	usuario TEXT NOT NULL UNIQUE,
	nome TEXT NOT NULL,
	senha TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS antena_users;
