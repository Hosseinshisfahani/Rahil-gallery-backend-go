DROP TABLE IF EXISTS customer_audit_log;
DROP TABLE IF EXISTS customer_notes;
DROP TABLE IF EXISTS customer_profiles;

DROP TYPE IF EXISTS vip_source;
DROP TYPE IF EXISTS customer_import_mode;

DROP INDEX IF EXISTS idx_users_phone_search;
DROP INDEX IF EXISTS users_phone_unique;
DROP INDEX IF EXISTS users_email_unique;

ALTER TABLE users
    ALTER COLUMN email SET NOT NULL,
    ALTER COLUMN password_hash SET NOT NULL;

ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email);
