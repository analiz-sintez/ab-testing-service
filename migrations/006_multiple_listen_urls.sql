-- +goose Up
-- +goose StatementBegin
-- Create proxy_listen_urls table
CREATE TABLE proxy_listen_urls
(
    id         VARCHAR(255) PRIMARY KEY,
    proxy_id   VARCHAR(255) NOT NULL REFERENCES proxies (id) ON DELETE CASCADE,
    listen_url VARCHAR(255) NOT NULL,
    path_key   VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (listen_url),
    UNIQUE (path_key)
);

-- Create index for faster lookups
CREATE INDEX idx_proxy_listen_urls_proxy_id ON proxy_listen_urls (proxy_id);

-- Migrate existing listen_urls to the new table
INSERT INTO proxy_listen_urls (id, proxy_id, listen_url, path_key, created_at, updated_at)
SELECT gen_random_uuid(), id, listen_url, path_key, created_at, updated_at
FROM proxies
WHERE listen_url IS NOT NULL;

-- Remove old columns from proxies table
ALTER TABLE proxies
    DROP COLUMN listen_url,
    DROP COLUMN path_key;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Add back the columns to proxies table
ALTER TABLE proxies
    ADD COLUMN listen_url VARCHAR(255),
    ADD COLUMN path_key VARCHAR(255);

-- Migrate data back
UPDATE proxies p
SET listen_url = plu.listen_url,
    path_key   = plu.path_key
FROM proxy_listen_urls plu
WHERE p.id = plu.proxy_id;

-- Drop the new table
DROP TABLE proxy_listen_urls;
-- +goose StatementEnd 