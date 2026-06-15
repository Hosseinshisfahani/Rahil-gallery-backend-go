DROP TRIGGER IF EXISTS trg_orders_sync_commerce_stats ON orders;
DROP FUNCTION IF EXISTS trg_orders_sync_commerce_stats();
DROP FUNCTION IF EXISTS refresh_all_customer_commerce_stats();
DROP FUNCTION IF EXISTS refresh_customer_commerce_stats(UUID);

DROP INDEX IF EXISTS idx_users_phone_trgm;
DROP INDEX IF EXISTS idx_users_full_name_trgm;
DROP INDEX IF EXISTS idx_users_customer_created;
DROP INDEX IF EXISTS idx_customer_profiles_last_purchase;
DROP INDEX IF EXISTS idx_customer_profiles_total_ltv;
DROP INDEX IF EXISTS idx_customer_profiles_total_orders;

ALTER TABLE customer_profiles
    DROP COLUMN IF EXISTS first_purchase_at,
    DROP COLUMN IF EXISTS last_purchase_at,
    DROP COLUMN IF EXISTS total_ltv,
    DROP COLUMN IF EXISTS total_orders;
