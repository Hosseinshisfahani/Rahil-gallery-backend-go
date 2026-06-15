-- Rahil Gallery — Jewelry E-commerce
-- Bounded contexts: identity, catalog, inventory, cart, order, payment, promotion, review, wishlist

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ---------------------------------------------------------------------------
-- Identity
-- ---------------------------------------------------------------------------

CREATE TYPE user_status AS ENUM ('active', 'inactive', 'banned');
CREATE TYPE address_type AS ENUM ('shipping', 'billing', 'both');

CREATE TABLE roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id           UUID NOT NULL REFERENCES roles (id),
    email             VARCHAR(255) NOT NULL,
    phone             VARCHAR(20),
    password_hash     VARCHAR(255) NOT NULL,
    first_name        VARCHAR(100) NOT NULL,
    last_name         VARCHAR(100) NOT NULL,
    status            user_status NOT NULL DEFAULT 'active',
    email_verified_at TIMESTAMPTZ,
    last_login_at     TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,
    CONSTRAINT users_email_unique UNIQUE (email)
);

CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_addresses (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    address_type   address_type NOT NULL DEFAULT 'both',
    label          VARCHAR(50),
    recipient_name VARCHAR(200) NOT NULL,
    phone          VARCHAR(20) NOT NULL,
    province       VARCHAR(100) NOT NULL,
    city           VARCHAR(100) NOT NULL,
    postal_code    VARCHAR(20) NOT NULL,
    address_line   TEXT NOT NULL,
    is_default     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ---------------------------------------------------------------------------
-- Catalog (jewelry-specific attributes)
-- ---------------------------------------------------------------------------

CREATE TYPE product_status AS ENUM ('draft', 'published', 'archived');
CREATE TYPE jewelry_type AS ENUM (
    'ring', 'necklace', 'bracelet', 'earring',
    'pendant', 'anklet', 'brooch', 'set', 'other'
);
CREATE TYPE metal_type AS ENUM (
    'gold_yellow', 'gold_white', 'gold_rose',
    'silver', 'platinum', 'titanium', 'mixed', 'other'
);
CREATE TYPE gemstone_type AS ENUM (
    'none', 'diamond', 'ruby', 'sapphire', 'emerald',
    'pearl', 'turquoise', 'amethyst', 'mixed', 'other'
);

CREATE TABLE categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id   UUID REFERENCES categories (id) ON DELETE SET NULL,
    name        VARCHAR(150) NOT NULL,
    slug        VARCHAR(180) NOT NULL UNIQUE,
    description TEXT,
    sort_order  INT NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE collections (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(150) NOT NULL,
    slug        VARCHAR(180) NOT NULL UNIQUE,
    description TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE products (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id        UUID NOT NULL REFERENCES categories (id),
    sku                VARCHAR(64) NOT NULL UNIQUE,
    name               VARCHAR(255) NOT NULL,
    slug               VARCHAR(280) NOT NULL UNIQUE,
    description        TEXT,
    short_description  VARCHAR(500),
    jewelry_type       jewelry_type NOT NULL,
    status             product_status NOT NULL DEFAULT 'draft',
    base_price         NUMERIC(14, 2) NOT NULL CHECK (base_price >= 0),
    compare_at_price   NUMERIC(14, 2) CHECK (compare_at_price IS NULL OR compare_at_price >= 0),
    currency           CHAR(3) NOT NULL DEFAULT 'IRR',
    metal_type         metal_type,
    karat              SMALLINT CHECK (karat IS NULL OR karat IN (14, 18, 21, 22, 24)),
    gemstone_type      gemstone_type NOT NULL DEFAULT 'none',
    weight_grams       NUMERIC(10, 3) CHECK (weight_grams IS NULL OR weight_grams > 0),
    purity_percent     NUMERIC(5, 2) CHECK (purity_percent IS NULL OR (purity_percent > 0 AND purity_percent <= 100)),
    certificate_number VARCHAR(100),
    is_handmade        BOOLEAN NOT NULL DEFAULT FALSE,
    is_featured        BOOLEAN NOT NULL DEFAULT FALSE,
    meta_title         VARCHAR(255),
    meta_description   VARCHAR(500),
    published_at       TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMPTZ
);

CREATE TABLE product_variants (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id       UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    sku              VARCHAR(64) NOT NULL UNIQUE,
    name             VARCHAR(150) NOT NULL,
    size_label       VARCHAR(50),
    color_label      VARCHAR(50),
    price_adjustment NUMERIC(14, 2) NOT NULL DEFAULT 0,
    weight_grams     NUMERIC(10, 3) CHECK (weight_grams IS NULL OR weight_grams > 0),
    barcode          VARCHAR(64),
    is_default       BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order       INT NOT NULL DEFAULT 0,
    is_active        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT product_variants_unique_per_product UNIQUE (product_id, name)
);

CREATE TABLE product_images (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    variant_id UUID REFERENCES product_variants (id) ON DELETE CASCADE,
    url        TEXT NOT NULL,
    alt_text   VARCHAR(255),
    sort_order INT NOT NULL DEFAULT 0,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE product_collections (
    product_id    UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    collection_id UUID NOT NULL REFERENCES collections (id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, collection_id)
);

-- ---------------------------------------------------------------------------
-- Inventory
-- ---------------------------------------------------------------------------

CREATE TABLE inventory_items (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    variant_id          UUID NOT NULL UNIQUE REFERENCES product_variants (id) ON DELETE CASCADE,
    quantity            INT NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    reserved_quantity   INT NOT NULL DEFAULT 0 CHECK (reserved_quantity >= 0),
    low_stock_threshold INT NOT NULL DEFAULT 5 CHECK (low_stock_threshold >= 0),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT inventory_reserved_lte_quantity CHECK (reserved_quantity <= quantity)
);

-- ---------------------------------------------------------------------------
-- Cart
-- ---------------------------------------------------------------------------

CREATE TYPE cart_status AS ENUM ('active', 'merged', 'abandoned', 'converted');

CREATE TABLE carts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES users (id) ON DELETE SET NULL,
    guest_token VARCHAR(64) UNIQUE,
    status      cart_status NOT NULL DEFAULT 'active',
    expires_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT carts_owner_check CHECK (user_id IS NOT NULL OR guest_token IS NOT NULL)
);

CREATE TABLE cart_items (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id             UUID NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    variant_id          UUID NOT NULL REFERENCES product_variants (id),
    quantity            INT NOT NULL CHECK (quantity > 0),
    unit_price_snapshot NUMERIC(14, 2) NOT NULL CHECK (unit_price_snapshot >= 0),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT cart_items_unique_variant UNIQUE (cart_id, variant_id)
);

-- ---------------------------------------------------------------------------
-- Order
-- ---------------------------------------------------------------------------

CREATE TYPE order_status AS ENUM (
    'pending', 'confirmed', 'processing',
    'shipped', 'delivered', 'cancelled', 'refunded'
);

CREATE TABLE orders (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number          VARCHAR(32) NOT NULL UNIQUE,
    user_id               UUID NOT NULL REFERENCES users (id),
    status                order_status NOT NULL DEFAULT 'pending',
    subtotal              NUMERIC(14, 2) NOT NULL CHECK (subtotal >= 0),
    discount_amount       NUMERIC(14, 2) NOT NULL DEFAULT 0 CHECK (discount_amount >= 0),
    shipping_amount       NUMERIC(14, 2) NOT NULL DEFAULT 0 CHECK (shipping_amount >= 0),
    tax_amount            NUMERIC(14, 2) NOT NULL DEFAULT 0 CHECK (tax_amount >= 0),
    total_amount          NUMERIC(14, 2) NOT NULL CHECK (total_amount >= 0),
    currency              CHAR(3) NOT NULL DEFAULT 'IRR',
    customer_note         TEXT,
    shipping_address_id   UUID REFERENCES user_addresses (id),
    billing_address_id    UUID REFERENCES user_addresses (id),
    placed_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE order_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    variant_id      UUID NOT NULL REFERENCES product_variants (id),
    product_name    VARCHAR(255) NOT NULL,
    variant_name    VARCHAR(150) NOT NULL,
    sku             VARCHAR(64) NOT NULL,
    jewelry_type    jewelry_type NOT NULL,
    metal_type      metal_type,
    karat           SMALLINT,
    quantity        INT NOT NULL CHECK (quantity > 0),
    unit_price      NUMERIC(14, 2) NOT NULL CHECK (unit_price >= 0),
    line_total      NUMERIC(14, 2) NOT NULL CHECK (line_total >= 0),
    weight_grams    NUMERIC(10, 3)
);

CREATE TABLE order_status_history (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    from_status order_status,
    to_status   order_status NOT NULL,
    note        TEXT,
    changed_by  UUID REFERENCES users (id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ---------------------------------------------------------------------------
-- Payment
-- ---------------------------------------------------------------------------

CREATE TYPE payment_status AS ENUM ('pending', 'authorized', 'paid', 'failed', 'refunded');
CREATE TYPE payment_provider AS ENUM ('zarinpal', 'idpay', 'manual', 'other');

CREATE TABLE payments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    provider    payment_provider NOT NULL,
    external_id VARCHAR(255),
    amount      NUMERIC(14, 2) NOT NULL CHECK (amount > 0),
    currency    CHAR(3) NOT NULL DEFAULT 'IRR',
    status      payment_status NOT NULL DEFAULT 'pending',
    paid_at     TIMESTAMPTZ,
    metadata    JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ---------------------------------------------------------------------------
-- Promotion
-- ---------------------------------------------------------------------------

CREATE TYPE discount_type AS ENUM ('percentage', 'fixed_amount');

CREATE TABLE coupons (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code             VARCHAR(50) NOT NULL UNIQUE,
    discount_type    discount_type NOT NULL,
    discount_value   NUMERIC(14, 2) NOT NULL CHECK (discount_value > 0),
    min_order_amount NUMERIC(14, 2) NOT NULL DEFAULT 0,
    max_uses         INT,
    used_count       INT NOT NULL DEFAULT 0 CHECK (used_count >= 0),
    valid_from       TIMESTAMPTZ NOT NULL,
    valid_until      TIMESTAMPTZ NOT NULL,
    is_active        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT coupons_valid_range CHECK (valid_until > valid_from)
);

CREATE TABLE coupon_redemptions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coupon_id       UUID NOT NULL REFERENCES coupons (id),
    order_id        UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users (id),
    discount_amount NUMERIC(14, 2) NOT NULL CHECK (discount_amount >= 0),
    redeemed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT coupon_redemptions_order_unique UNIQUE (order_id)
);

-- ---------------------------------------------------------------------------
-- Review & Wishlist
-- ---------------------------------------------------------------------------

CREATE TABLE product_reviews (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id  UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    user_id     UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    rating      SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title       VARCHAR(200),
    body        TEXT,
    is_approved BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT product_reviews_user_product_unique UNIQUE (product_id, user_id)
);

CREATE TABLE wishlist_items (
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    variant_id UUID NOT NULL REFERENCES product_variants (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, variant_id)
);

-- ---------------------------------------------------------------------------
-- Indexes
-- ---------------------------------------------------------------------------

CREATE INDEX idx_users_role_id ON users (role_id);
CREATE INDEX idx_users_status ON users (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);
CREATE INDEX idx_user_addresses_user_id ON user_addresses (user_id);

CREATE INDEX idx_categories_parent_id ON categories (parent_id);
CREATE INDEX idx_products_category_id ON products (category_id);
CREATE INDEX idx_products_status ON products (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_jewelry_type ON products (jewelry_type);
CREATE INDEX idx_products_featured ON products (is_featured) WHERE status = 'published';
CREATE INDEX idx_product_variants_product_id ON product_variants (product_id);
CREATE INDEX idx_product_images_product_id ON product_images (product_id);

CREATE INDEX idx_carts_user_id ON carts (user_id);
CREATE INDEX idx_cart_items_cart_id ON cart_items (cart_id);

CREATE INDEX idx_orders_user_id ON orders (user_id);
CREATE INDEX idx_orders_status ON orders (status);
CREATE INDEX idx_order_items_order_id ON order_items (order_id);
CREATE INDEX idx_order_status_history_order_id ON order_status_history (order_id);

CREATE INDEX idx_payments_order_id ON payments (order_id);
CREATE INDEX idx_payments_status ON payments (status);

CREATE INDEX idx_product_reviews_product_id ON product_reviews (product_id);
CREATE INDEX idx_wishlist_items_user_id ON wishlist_items (user_id);

-- ---------------------------------------------------------------------------
-- Seed: default roles
-- ---------------------------------------------------------------------------

INSERT INTO roles (name, description) VALUES
    ('customer', 'Store customer'),
    ('admin', 'Full platform administrator'),
    ('staff', 'Catalog and order management');
