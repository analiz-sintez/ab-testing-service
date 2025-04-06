-- +goose Up
-- +goose StatementBegin
ALTER TABLE proxies
    ADD COLUMN name VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE proxies
    DROP COLUMN name;
-- +goose StatementEnd 