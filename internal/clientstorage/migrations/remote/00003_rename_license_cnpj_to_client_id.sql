-- +goose Up
ALTER TABLE license RENAME COLUMN cnpj TO client_id;

-- +goose Down
ALTER TABLE license RENAME COLUMN client_id TO cnpj;
