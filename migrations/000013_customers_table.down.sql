-- Best-effort rollback: recreate legacy CRM tables (data not restored).

-- CREATE TYPE customer_list_view_type AS ENUM ('filter', 'segment');

-- CREATE TYPE customer_import_mode AS ENUM ('quick', 'history_included');
-- CREATE TYPE vip_source AS ENUM ('manual', 'automatic');

-- CREATE TABLE customer_profiles (
--     user_id            UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
--     locale             VARCHAR(5) NOT NULL DEFAULT 'fa',
--     default_ring_size  VARCHAR(10),
--     is_vip             BOOLEAN NOT NULL DEFAULT FALSE,
--     vip_source         vip_source,
--     import_mode        customer_import_mode,
--     import_profile     JSONB,
--     tags               TEXT[] NOT NULL DEFAULT '{}',
--     block_reason       VARCHAR(50),
--     block_note         TEXT,
--     last_activity_at   TIMESTAMPTZ,
--     total_orders       INT NOT NULL DEFAULT 0,
--     total_ltv          NUMERIC(14, 2) NOT NULL DEFAULT 0,
--     last_purchase_at   TIMESTAMPTZ,
--     first_purchase_at  TIMESTAMPTZ,
--     segment            TEXT NOT NULL DEFAULT 'new',
--     crm_age_range      TEXT,
--     crm_gender         TEXT,
--     crm_customer_type  TEXT,
--     crm_first_visit_date TEXT,
--     crm_birthday       TEXT,
--     crm_marriage_date  TEXT,
--     created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--     updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
-- );

-- CREATE TABLE customer_notes (
--     id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
--     author_id  UUID NOT NULL REFERENCES users (id),
--     body       TEXT NOT NULL,
--     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
-- );

-- CREATE TABLE customer_audit_log (
--     id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     admin_id       UUID NOT NULL REFERENCES users (id),
--     target_user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
--     action         VARCHAR(50) NOT NULL,
--     reason         TEXT,
--     details        TEXT,
--     created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
-- );

-- CREATE TABLE admin_customer_list_views (
--     id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     owner_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
--     name        VARCHAR(120) NOT NULL,
--     view_type   customer_list_view_type NOT NULL DEFAULT 'filter',
--     filters     JSONB NOT NULL DEFAULT '{}',
--     is_shared   BOOLEAN NOT NULL DEFAULT FALSE,
--     position    INT NOT NULL DEFAULT 0,
--     created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--     updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
--     CONSTRAINT admin_customer_list_views_owner_name_key UNIQUE (owner_id, name)
-- );

-- DROP TABLE IF EXISTS customers;
