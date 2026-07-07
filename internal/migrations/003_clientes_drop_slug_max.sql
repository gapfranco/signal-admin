-- +goose Up
ALTER TABLE clientes DROP COLUMN slug_turso;
ALTER TABLE clientes DROP COLUMN max_instalacoes;

-- +goose Down
ALTER TABLE clientes ADD COLUMN slug_turso TEXT NOT NULL DEFAULT '';
ALTER TABLE clientes ADD COLUMN max_instalacoes INTEGER NOT NULL DEFAULT 1;
