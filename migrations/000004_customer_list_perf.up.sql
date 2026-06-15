-- Customer list/search performance: denormalized commerce stats + trigram search indexes

ALTER TABLE customer_profiles
    ADD COLUMN IF NOT EXISTS total_orders       INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_ltv          NUMERIC(14, 2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_purchase_at   TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS first_purchase_at  TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_customer_profiles_total_orders
    ON customer_profiles (total_orders);

CREATE INDEX IF NOT EXISTS idx_customer_profiles_total_ltv
    ON customer_profiles (total_ltv);

CREATE INDEX IF NOT EXISTS idx_customer_profiles_last_purchase
    ON customer_profiles (last_purchase_at)
    WHERE last_purchase_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_users_customer_created
    ON users (created_at DESC)
    WHERE deleted_at IS NULL;

-- Backfill commerce stats from orders
UPDATE customer_profiles cp
SET
    total_orders = stats.order_count,
    total_ltv = stats.ltv,
    last_purchase_at = stats.last_at,
    first_purchase_at = stats.first_at,
    updated_at = NOW()
FROM (
    SELECT
        user_id,
        COUNT(*)::int AS order_count,
        COALESCE(SUM(total_amount), 0) AS ltv,
        MAX(placed_at) AS last_at,
        MIN(placed_at) AS first_at
    FROM orders
    GROUP BY user_id
) stats
WHERE cp.user_id = stats.user_id;

-- Refresh one customer (used by trigger)
CREATE OR REPLACE FUNCTION refresh_customer_commerce_stats(p_user_id UUID)
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE customer_profiles cp
    SET
        total_orders = COALESCE(stats.order_count, 0),
        total_ltv = COALESCE(stats.ltv, 0),
        last_purchase_at = stats.last_at,
        first_purchase_at = stats.first_at,
        updated_at = NOW()
    FROM (
        SELECT
            COUNT(*)::int AS order_count,
            COALESCE(SUM(total_amount), 0) AS ltv,
            MAX(placed_at) AS last_at,
            MIN(placed_at) AS first_at
        FROM orders
        WHERE user_id = p_user_id
    ) stats
    WHERE cp.user_id = p_user_id;
END;
$$;

-- Full backfill (used after bulk seed)
CREATE OR REPLACE FUNCTION refresh_all_customer_commerce_stats()
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE customer_profiles cp
    SET
        total_orders = COALESCE(stats.order_count, 0),
        total_ltv = COALESCE(stats.ltv, 0),
        last_purchase_at = stats.last_at,
        first_purchase_at = stats.first_at,
        updated_at = NOW()
    FROM (
        SELECT
            user_id,
            COUNT(*)::int AS order_count,
            COALESCE(SUM(total_amount), 0) AS ltv,
            MAX(placed_at) AS last_at,
            MIN(placed_at) AS first_at
        FROM orders
        GROUP BY user_id
    ) stats
    WHERE cp.user_id = stats.user_id;

    UPDATE customer_profiles cp
    SET
        total_orders = 0,
        total_ltv = 0,
        last_purchase_at = NULL,
        first_purchase_at = NULL,
        updated_at = NOW()
    WHERE NOT EXISTS (SELECT 1 FROM orders o WHERE o.user_id = cp.user_id);
END;
$$;

CREATE OR REPLACE FUNCTION trg_orders_sync_commerce_stats()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    uid UUID;
BEGIN
    uid := COALESCE(NEW.user_id, OLD.user_id);
    PERFORM refresh_customer_commerce_stats(uid);
    RETURN COALESCE(NEW, OLD);
END;
$$;

DROP TRIGGER IF EXISTS trg_orders_sync_commerce_stats ON orders;

CREATE TRIGGER trg_orders_sync_commerce_stats
    AFTER INSERT OR UPDATE OF user_id, total_amount, placed_at OR DELETE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION trg_orders_sync_commerce_stats();

-- Trigram indexes for admin customer search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_users_full_name_trgm
    ON users USING gin (lower(trim(first_name || ' ' || last_name)) gin_trgm_ops)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_users_phone_trgm
    ON users USING gin (replace(COALESCE(phone, ''), ' ', '') gin_trgm_ops)
    WHERE deleted_at IS NULL AND phone IS NOT NULL;
