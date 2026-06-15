-- Normalized phone digits for prefix search (handles +98… stored formats)

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS phone_digits text
    GENERATED ALWAYS AS (regexp_replace(COALESCE(phone, ''), '[^0-9]', '', 'g')) STORED;

DROP INDEX IF EXISTS idx_users_phone_normalized_prefix;

CREATE INDEX IF NOT EXISTS idx_users_phone_digits_prefix
    ON users (phone_digits text_pattern_ops)
    WHERE deleted_at IS NULL AND phone IS NOT NULL;
