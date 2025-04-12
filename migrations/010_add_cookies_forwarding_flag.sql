-- +goose Up
-- +goose StatementBegin
ALTER TABLE proxies
    ADD COLUMN cookies_forwarding_flg BOOLEAN NOT NULL DEFAULT false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE proxies
    DROP COLUMN cookies_forwarding_flg;
-- +goose StatementEnd 