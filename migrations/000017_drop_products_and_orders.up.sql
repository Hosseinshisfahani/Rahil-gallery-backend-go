-- Drop catalog, cart, and order tables left by the initial schema.
-- Historical migrations that created them are left in place.

DROP TABLE IF EXISTS coupon_redemptions;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS order_status_history;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS cart_items;
DROP TABLE IF EXISTS carts;
DROP TABLE IF EXISTS coupons;
DROP TABLE IF EXISTS product_reviews;
DROP TABLE IF EXISTS wishlist_items;
DROP TABLE IF EXISTS inventory_items;
DROP TABLE IF EXISTS product_collections;
DROP TABLE IF EXISTS product_images;
DROP TABLE IF EXISTS product_variants;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS collections;

DROP TYPE IF EXISTS cart_status;
DROP TYPE IF EXISTS order_status;
DROP TYPE IF EXISTS payment_status;
DROP TYPE IF EXISTS payment_provider;
DROP TYPE IF EXISTS discount_type;
DROP TYPE IF EXISTS product_status;
DROP TYPE IF EXISTS jewelry_type;
DROP TYPE IF EXISTS metal_type;
DROP TYPE IF EXISTS gemstone_type;
