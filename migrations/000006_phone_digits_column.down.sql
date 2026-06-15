DROP INDEX IF EXISTS idx_users_phone_digits_prefix;

ALTER TABLE users DROP COLUMN IF EXISTS phone_digits;

CREATE INDEX IF NOT EXISTS idx_users_phone_normalized_prefix
    ON users (replace(COALESCE(phone, ''), ' ', '') text_pattern_ops)
    WHERE deleted_at IS NULL AND phone IS NOT NULL;
