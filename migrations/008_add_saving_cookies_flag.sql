-- +goose Up
-- +goose StatementBegin
ALTER TABLE proxies
    ADD COLUMN saving_cookies_flg BOOLEAN NOT NULL DEFAULT false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE proxies
    DROP COLUMN saving_cookies_flg;
-- +goose StatementEnd 