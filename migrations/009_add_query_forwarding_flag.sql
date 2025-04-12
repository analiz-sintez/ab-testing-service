-- +goose Up
-- +goose StatementBegin
ALTER TABLE proxies
    ADD COLUMN query_forwarding_flg BOOLEAN NOT NULL DEFAULT true;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE proxies
    DROP COLUMN query_forwarding_flg;
-- +goose StatementEnd 