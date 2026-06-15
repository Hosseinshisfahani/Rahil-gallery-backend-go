-- Admin CRM: saved customer list filters and segment shortcuts

CREATE TYPE customer_list_view_type AS ENUM ('filter', 'segment');

CREATE TABLE admin_customer_list_views (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name        VARCHAR(120) NOT NULL,
    view_type   customer_list_view_type NOT NULL DEFAULT 'filter',
    filters     JSONB NOT NULL DEFAULT '{}',
    is_shared   BOOLEAN NOT NULL DEFAULT FALSE,
    position    INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT admin_customer_list_views_owner_name_key UNIQUE (owner_id, name)
);

CREATE INDEX idx_admin_customer_list_views_owner
    ON admin_customer_list_views (owner_id, position);

CREATE INDEX idx_admin_customer_list_views_shared
    ON admin_customer_list_views (is_shared, view_type)
    WHERE is_shared = TRUE;
