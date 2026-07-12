-- +goose Up
CREATE TABLE IF NOT EXISTS license (
    nome TEXT,
    cnpj TEXT PRIMARY KEY NOT NULL,
    validade TEXT
);

-- +goose Down
DROP TABLE IF EXISTS license;
