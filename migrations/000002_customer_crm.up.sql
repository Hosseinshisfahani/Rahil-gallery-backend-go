-- Customer CRM: nullable credentials for OTP-only accounts, profiles, notes, audit

ALTER TABLE users
    ALTER COLUMN email DROP NOT NULL,
    ALTER COLUMN password_hash DROP NOT NULL;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_unique;

CREATE UNIQUE INDEX users_email_unique ON users (email)
    WHERE email IS NOT NULL AND deleted_at IS NULL;

CREATE UNIQUE INDEX users_phone_unique ON users (phone)
    WHERE phone IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX idx_users_phone_search ON users (phone)
    WHERE deleted_at IS NULL;

CREATE TYPE customer_import_mode AS ENUM ('quick', 'history_included');
CREATE TYPE vip_source AS ENUM ('manual', 'automatic');

CREATE TABLE customer_profiles (
    user_id            UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    locale             VARCHAR(5) NOT NULL DEFAULT 'fa',
    default_ring_size  VARCHAR(10),
    is_vip             BOOLEAN NOT NULL DEFAULT FALSE,
    vip_source         vip_source,
    import_mode        customer_import_mode,
    import_profile     JSONB,
    tags               TEXT[] NOT NULL DEFAULT '{}',
    block_reason       VARCHAR(50),
    block_note         TEXT,
    last_activity_at   TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE customer_notes (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    author_id  UUID NOT NULL REFERENCES users (id),
    body       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE customer_audit_log (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id       UUID NOT NULL REFERENCES users (id),
    target_user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    action         VARCHAR(50) NOT NULL,
    reason         TEXT,
    details        TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_customer_profiles_is_vip ON customer_profiles (is_vip) WHERE is_vip = TRUE;
CREATE INDEX idx_customer_profiles_tags ON customer_profiles USING GIN (tags);
CREATE INDEX idx_customer_profiles_last_activity ON customer_profiles (last_activity_at);

CREATE INDEX idx_customer_notes_user_id ON customer_notes (user_id);
CREATE INDEX idx_customer_audit_log_target ON customer_audit_log (target_user_id, created_at DESC);
