-- SMS campaign jobs (bulk) + birthday send log (Go-owned CRM)

CREATE TABLE sms_jobs (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_by              UUID REFERENCES users (id) ON DELETE SET NULL,
    message                 TEXT NOT NULL,
    filter_snapshot         JSONB NOT NULL DEFAULT '{}',
    status                  VARCHAR(40) NOT NULL DEFAULT 'pending',
    matched_count           INT NOT NULL DEFAULT 0,
    skipped_invalid_phone   INT NOT NULL DEFAULT 0,
    sent_count              INT NOT NULL DEFAULT 0,
    failed_count            INT NOT NULL DEFAULT 0,
    batch_count             INT NOT NULL DEFAULT 0,
    seller_note             TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at            TIMESTAMPTZ
);

CREATE INDEX idx_sms_jobs_created_at ON sms_jobs (created_at DESC);

CREATE TABLE sms_birthday_log (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id   UUID NOT NULL REFERENCES customers (id) ON DELETE CASCADE,
    sent_on       DATE NOT NULL,
    status        VARCHAR(40) NOT NULL DEFAULT 'sent',
    error_message TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT sms_birthday_log_customer_day_unique UNIQUE (customer_id, sent_on)
);

CREATE INDEX idx_sms_birthday_log_sent_on ON sms_birthday_log (sent_on DESC);
