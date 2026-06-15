# Catalog seed images

Jewelry photos for dev seeds, served by the API at `/static/catalog/...`.

## Fetch

```bash
make fetch-catalog-images
```

Downloads ~27 JPEGs into this folder (Flickr CC pool via loremflickr.com, with placehold.co fallback).

## Seed URLs

Seeds store absolute URLs like `http://localhost:8080/static/catalog/products/rings/01.jpg`.

Override for Docker or staging:

```bash
SEED_ASSET_BASE_URL=http://localhost:8080 make seed-reset
```

## Layout

See `manifest.json` for category, collection, and product image mappings.
