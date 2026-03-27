DROP INDEX IF EXISTS idx_shorten_urls_user_deleted;
DROP INDEX IF EXISTS idx_shorten_urls_is_deleted;
DROP INDEX IF EXISTS idx_shorten_urls_user_id;

ALTER TABLE shorten_urls 
    DROP COLUMN IF EXISTS user_id,
    DROP COLUMN IF EXISTS is_deleted,
    DROP COLUMN IF EXISTS created_at;