DROP TRIGGER IF EXISTS trg_customer_profiles_sync_segment ON customer_profiles;
DROP FUNCTION IF EXISTS trg_customer_profiles_sync_segment();
DROP FUNCTION IF EXISTS refresh_all_customer_segments();
DROP INDEX IF EXISTS idx_customer_profiles_segment;

ALTER TABLE customer_profiles DROP COLUMN IF EXISTS segment;

DROP FUNCTION IF EXISTS compute_customer_segment();

-- Restore refresh functions without segment (from 000004)
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
