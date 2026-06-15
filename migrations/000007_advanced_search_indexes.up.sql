-- Advanced search: trigram index for partial email match

CREATE INDEX IF NOT EXISTS idx_users_email_trgm
    ON users USING gin (lower(email) gin_trgm_ops)
    WHERE deleted_at IS NULL AND email IS NOT NULL;
