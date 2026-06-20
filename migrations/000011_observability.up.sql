-- Application errors and client reports (retained ~7 days via app cleanup job).
CREATE TABLE IF NOT EXISTS observability_events (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source       TEXT NOT NULL CHECK (source IN ('api', 'client')),
    level        TEXT NOT NULL CHECK (level IN ('error', 'warn', 'info')),
    message      TEXT NOT NULL,
    stack_trace  TEXT,
    route        TEXT,
    method       TEXT,
    status_code  INT,
    request_id   TEXT,
    user_agent   TEXT,
    metadata     JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_observability_events_created_at
    ON observability_events (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_observability_events_source_level
    ON observability_events (source, level, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_observability_events_route
    ON observability_events (route)
    WHERE route IS NOT NULL;
