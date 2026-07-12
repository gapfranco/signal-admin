-- +goose Up
ALTER TABLE clientes ADD COLUMN onfly INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE clientes DROP COLUMN onfly;
