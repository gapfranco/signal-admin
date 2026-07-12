-- +goose Up
ALTER TABLE license ADD COLUMN status TEXT DEFAULT 'active';

-- +goose Down
ALTER TABLE license DROP COLUMN status;
