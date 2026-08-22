# Strangler Fig — schema ownership freeze

Commerce HTTP APIs move to **Rahil-gallery-backend-django**. Go keeps Auth, CRM customers, and catalog.

## Frozen for Go migrations (Django-owned)

Do **not** create new golang-migrate files that `CREATE`, `ALTER`, or `DROP`:

- `carts`, `cart_items`
- `orders`, `order_items`, `order_status_history`
- `payments`
- `coupons`, `coupon_redemptions`
- Related enums: `cart_status`, `order_status`, `payment_status`, `payment_provider`, `discount_type`

## Still owned by Go

- Identity: `users`, `roles`, `refresh_tokens`, `user_addresses`
- CRM: `customers` (+ admin saved views if present)
- Catalog / inventory / observability

## Auth contract (Django verifies)

- Secret: `JWT_ACCESS_SECRET` (HS256)
- Claims: `sub` (user UUID), `email`, `role`, `exp`, `iat`, `jti`
- Refresh: opaque tokens via Go `/api/v1/auth/refresh` only

See also: `Rahil-gallery-backend-django/docs/adr-table-ownership.md`.

PR checklist: if SQL touches a frozen commerce table → reject; route the change through Django migrations instead.
