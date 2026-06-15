-- Denormalized segment on customer_profiles for indexed admin filters

ALTER TABLE customer_profiles
    ADD COLUMN IF NOT EXISTS segment TEXT;

CREATE OR REPLACE FUNCTION compute_customer_segment(
    p_is_vip BOOLEAN,
    p_total_orders INT,
    p_last_activity_at TIMESTAMPTZ,
    p_user_created_at TIMESTAMPTZ
)
RETURNS TEXT
LANGUAGE sql
STABLE
AS $$
SELECT CASE
    WHEN COALESCE(p_is_vip, FALSE) THEN 'vip'
    WHEN COALESCE(p_total_orders, 0) = 0 THEN 'new'
    WHEN COALESCE(p_total_orders, 0) >= 2 THEN 'returning'
    WHEN COALESCE(p_last_activity_at, p_user_created_at) < NOW() - INTERVAL '90 days' THEN 'inactive'
    ELSE 'active'
END;
$$;

UPDATE customer_profiles cp
SET segment = compute_customer_segment(cp.is_vip, cp.total_orders, cp.last_activity_at, u.created_at)
FROM users u
WHERE u.id = cp.user_id;

ALTER TABLE customer_profiles
    ALTER COLUMN segment SET NOT NULL,
    ALTER COLUMN segment SET DEFAULT 'new';

CREATE INDEX IF NOT EXISTS idx_customer_profiles_segment
    ON customer_profiles (segment);

CREATE OR REPLACE FUNCTION trg_customer_profiles_sync_segment()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    user_created_at TIMESTAMPTZ;
BEGIN
    SELECT u.created_at INTO user_created_at FROM users u WHERE u.id = NEW.user_id;
    NEW.segment := compute_customer_segment(
        NEW.is_vip,
        NEW.total_orders,
        NEW.last_activity_at,
        user_created_at
    );
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_customer_profiles_sync_segment ON customer_profiles;

CREATE TRIGGER trg_customer_profiles_sync_segment
    BEFORE INSERT OR UPDATE OF is_vip, total_orders, last_activity_at ON customer_profiles
    FOR EACH ROW
    EXECUTE FUNCTION trg_customer_profiles_sync_segment();

-- Keep segment in sync when commerce stats refresh runs
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
        segment = compute_customer_segment(
            cp.is_vip,
            COALESCE(stats.order_count, 0),
            cp.last_activity_at,
            u.created_at
        ),
        updated_at = NOW()
    FROM (
        SELECT
            COUNT(*)::int AS order_count,
            COALESCE(SUM(total_amount), 0) AS ltv,
            MAX(placed_at) AS last_at,
            MIN(placed_at) AS first_at
        FROM orders
        WHERE user_id = p_user_id
    ) stats,
    users u
    WHERE cp.user_id = p_user_id
      AND u.id = cp.user_id;
END;
$$;

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
        segment = compute_customer_segment(
            cp.is_vip,
            COALESCE(stats.order_count, 0),
            cp.last_activity_at,
            u.created_at
        ),
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
    ) stats,
    users u
    WHERE cp.user_id = stats.user_id
      AND u.id = cp.user_id;

    UPDATE customer_profiles cp
    SET
        total_orders = 0,
        total_ltv = 0,
        last_purchase_at = NULL,
        first_purchase_at = NULL,
        segment = compute_customer_segment(cp.is_vip, 0, cp.last_activity_at, u.created_at),
        updated_at = NOW()
    FROM users u
    WHERE cp.user_id = u.id
      AND NOT EXISTS (SELECT 1 FROM orders o WHERE o.user_id = cp.user_id);
END;
$$;

CREATE OR REPLACE FUNCTION refresh_all_customer_segments()
RETURNS VOID
LANGUAGE plpgsql
AS $$
BEGIN
    UPDATE customer_profiles cp
    SET segment = compute_customer_segment(cp.is_vip, cp.total_orders, cp.last_activity_at, u.created_at)
    FROM users u
    WHERE u.id = cp.user_id;
END;
$$;
