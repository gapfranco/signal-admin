-- +goose Up
CREATE TABLE IF NOT EXISTS clientes (
    cliente_id       TEXT PRIMARY KEY,
    nome             TEXT NOT NULL,
    cnpj             TEXT NOT NULL DEFAULT '',
    email            TEXT NOT NULL DEFAULT '',
    telefone         TEXT NOT NULL DEFAULT '',
    slug_turso       TEXT NOT NULL DEFAULT '',
    valid_until      TEXT,
    max_instalacoes  INTEGER NOT NULL DEFAULT 1,
    status           TEXT NOT NULL DEFAULT 'active',
    observacao       TEXT NOT NULL DEFAULT '',
    created_at       TEXT NOT NULL,
    updated_at       TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS clientes;
