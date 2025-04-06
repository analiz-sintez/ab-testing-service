-- +goose Up
-- +goose StatementBegin
-- Add name column to targets table
ALTER TABLE targets ADD COLUMN name VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove name column from targets table
ALTER TABLE targets DROP COLUMN name;
-- +goose StatementEnd
