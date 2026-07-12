-- Replace users + customer_profiles CRM with a single customers table.

CREATE TABLE customers (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name           VARCHAR(100) NOT NULL,
    last_name            VARCHAR(100) NOT NULL,
    job                  VARCHAR(200),
    phone                VARCHAR(20) NOT NULL,
    phone_digits         TEXT GENERATED ALWAYS AS (regexp_replace(COALESCE(phone, ''), '[^0-9]', '', 'g')) STORED,
    email                VARCHAR(255),
    address              TEXT,
    birthday             DATE,
    marriage_date        DATE,
    important_date       DATE,
    first_visit_date     DATE,
    gender               VARCHAR(20),
    customer_type        VARCHAR(50) NOT NULL DEFAULT 'public',
    customer_age_range   VARCHAR(10),
    purchased_categories TEXT[] NOT NULL DEFAULT '{}',
    description          TEXT,
    signature_url        TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at           TIMESTAMPTZ
);

CREATE UNIQUE INDEX customers_phone_unique
    ON customers (phone)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX customers_email_unique
    ON customers (email)
    WHERE deleted_at IS NULL AND email IS NOT NULL;

CREATE INDEX idx_customers_created_at
    ON customers (created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_customers_phone_digits_prefix
    ON customers (phone_digits text_pattern_ops)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_customers_name_trgm
    ON customers USING gin (lower(trim(first_name || ' ' || last_name)) gin_trgm_ops)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_customers_email_trgm
    ON customers USING gin (lower(email) gin_trgm_ops)
    WHERE deleted_at IS NULL AND email IS NOT NULL;

CREATE INDEX idx_customers_customer_type
    ON customers (customer_type)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_customers_age_range
    ON customers (customer_age_range)
    WHERE deleted_at IS NULL AND customer_age_range IS NOT NULL;

CREATE INDEX idx_customers_gender
    ON customers (gender)
    WHERE deleted_at IS NULL AND gender IS NOT NULL;

CREATE INDEX idx_customers_purchased_categories
    ON customers USING gin (purchased_categories);

CREATE INDEX idx_customers_first_visit_date
    ON customers (first_visit_date)
    WHERE deleted_at IS NULL AND first_visit_date IS NOT NULL;

CREATE INDEX idx_customers_birthday
    ON customers (birthday)
    WHERE deleted_at IS NULL AND birthday IS NOT NULL;

CREATE INDEX idx_customers_marriage_date
    ON customers (marriage_date)
    WHERE deleted_at IS NULL AND marriage_date IS NOT NULL;


-- Drop commerce/segment triggers that targeted customer_profiles.
DROP TRIGGER IF EXISTS trg_orders_sync_commerce_stats ON orders;
DROP TRIGGER IF EXISTS trg_customer_profiles_sync_segment ON customer_profiles;

DROP FUNCTION IF EXISTS trg_orders_sync_commerce_stats();
DROP FUNCTION IF EXISTS trg_customer_profiles_sync_segment();
DROP FUNCTION IF EXISTS refresh_customer_commerce_stats(UUID);
DROP FUNCTION IF EXISTS refresh_all_customer_commerce_stats();
DROP FUNCTION IF EXISTS refresh_all_customer_segments();
DROP FUNCTION IF EXISTS compute_customer_segment(BOOLEAN, INT, TIMESTAMPTZ, TIMESTAMPTZ);

DROP TABLE IF EXISTS customer_notes;
DROP TABLE IF EXISTS customer_audit_log;
DROP TABLE IF EXISTS admin_customer_list_views;
DROP TABLE IF EXISTS customer_profiles;

DROP TYPE IF EXISTS customer_list_view_type;

-- Soft-delete migrated customer users (admin/staff users remain).
UPDATE users u
SET deleted_at = NOW(), updated_at = NOW()
FROM roles r
WHERE u.role_id = r.id
  AND r.name = 'customer'
  AND u.deleted_at IS NULL;
