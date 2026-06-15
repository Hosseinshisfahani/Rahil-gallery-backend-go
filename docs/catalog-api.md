# Public catalog API

Base path: `/api/v1` — no authentication required.

## Categories

| Method | Path | Description |
|--------|------|-------------|
| GET | `/categories` | Active category tree |
| GET | `/categories/:slug` | Category by slug |

## Collections

| Method | Path | Description |
|--------|------|-------------|
| GET | `/collections` | Active collections |
| GET | `/collections/:slug` | Collection detail |
| GET | `/collections/:slug/products` | Published products in collection (`page`, `perPage`) |

## Products

| Method | Path | Query params |
|--------|------|--------------|
| GET | `/products` | `page`, `perPage`, `category`, `collection`, `jewelryType`, `metal`, `gemstone`, `priceMin`, `priceMax`, `featured`, `q`, `sort` |
| GET | `/products/:slug` | Product detail with variants, images, availability |

### Sort values

- `newest` (default)
- `price_asc`
- `price_desc`
- `featured`

### List response

```json
{
  "data": [
    {
      "id": "uuid",
      "slug": "solitaire-diamond-ring",
      "name": "Solitaire Diamond Ring",
      "title": { "en": "Solitaire Diamond Ring", "fa": "Solitaire Diamond Ring" },
      "category": "Rings",
      "categorySlug": "rings",
      "jewelryType": "ring",
      "basePrice": 185000000,
      "priceFrom": 185000000,
      "priceTo": 190000000,
      "currency": "IRR",
      "gemstoneType": "diamond",
      "availability": "in_stock",
      "primaryImageUrl": "https://..."
    }
  ],
  "meta": { "page": 1, "perPage": 12, "total": 3, "totalPages": 1 }
}
```

`title.fa` mirrors `name` until bilingual columns are added.

## Examples

```bash
curl -s http://localhost:8080/api/v1/products
curl -s http://localhost:8080/api/v1/products/solitaire-diamond-ring
curl -s "http://localhost:8080/api/v1/products?category=rings&sort=price_desc"
```
