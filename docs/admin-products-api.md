# Admin products API

Base path: `/api/v1/admin` — requires Bearer token with `admin` or `staff` role.

## Products

| Method | Path | Description |
|--------|------|-------------|
| GET | `/products` | Paginated list (`page`, `perPage`, `q`, `status`, `categoryId`, `jewelryType`, `metal`, `gemstone`, `featured`, `priceMin`, `priceMax`) |
| POST | `/products` | Create product |
| GET | `/products/:id` | Product detail |
| PATCH | `/products/:id` | Update product |
| DELETE | `/products/:id` | Archive (soft delete) |

### Create product

```json
{
  "categoryId": "33333333-3333-4333-8333-333333333301",
  "sku": "RG-NEW-001",
  "name": "New Ring",
  "slug": "new-ring",
  "jewelryType": "ring",
  "status": "draft",
  "basePrice": 50000000,
  "gemstoneType": "diamond",
  "metalType": "gold_white",
  "karat": 18
}
```

## Variants

| Method | Path | Description |
|--------|------|-------------|
| POST | `/products/:id/variants` | Add variant + initial inventory |
| PATCH | `/products/:id/variants/:variantId` | Update variant |
| DELETE | `/products/:id/variants/:variantId` | Deactivate variant |

### Create variant

```json
{
  "sku": "RG-NEW-001-14",
  "name": "Size 14",
  "sizeLabel": "14",
  "priceAdjustment": 0,
  "initialQuantity": 10,
  "lowStockThreshold": 3,
  "isDefault": true
}
```

## Images

| Method | Path | Description |
|--------|------|-------------|
| POST | `/products/:id/images` | Add image |
| DELETE | `/products/:id/images/:imageId` | Remove image |

## Inventory

| Method | Path | Description |
|--------|------|-------------|
| PATCH | `/inventory/:variantId` | Set stock quantity |

```json
{ "quantity": 42, "lowStockThreshold": 5 }
```

## Examples

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@rehil.gallery","password":"Admin1234"}' \
  | jq -r '.data.access_token')

curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/admin/products
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/admin/products/44444444-4444-4444-8444-444444444401
```
