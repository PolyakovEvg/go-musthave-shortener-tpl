ALTER TABLE shorten_urls
    ADD COLUMN IF NOT EXISTS user_id TEXT;
    ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN NOT NULL DEFAULT FALSE;
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_shorten_urls_user_id ON shorten_urls(user_id);
CREATE INDEX IF NOT EXISTS idx_shorten_urls_is_deleted ON shorten_urls(is_deleted);
CREATE INDEX IF NOT EXISTS idx_shorten_urls_user_deleted ON shorten_urls(user_id, is_deleted);