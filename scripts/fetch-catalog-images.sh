#!/usr/bin/env bash
# Downloads jewelry photos into data/catalog-images/ (served at /static/catalog).
# Primary: loremflickr.com (Creative Commons Flickr pool). Fallback: placehold.co labels.
# Re-run safely; existing non-empty files are skipped.

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEST="$ROOT/data/catalog-images"
mkdir -p "$DEST"/{categories,collections,products/{rings,necklaces,bracelets,earrings,pendants}}

download() {
  local rel="$1"
  local tags="$2"
  local lock="$3"
  local label="$4"
  local out="$DEST/$rel"

  if [[ -f "$out" && -s "$out" ]]; then
    echo "skip  $rel"
    return 0
  fi

  mkdir -p "$(dirname "$out")"
  local url="https://loremflickr.com/900/900/${tags}?lock=${lock}"
  echo "fetch $rel ($tags)"
  if curl -fsSL --max-time 45 "$url" -o "$out"; then
    return 0
  fi

  echo "fallback $rel (placehold)"
  curl -fsSL --max-time 20 \
    "https://placehold.co/900x900/jpeg?text=${label// /+}" \
    -o "$out"
}

# Categories
download categories/rings.jpg      "jewelry,ring"      101 "Rings"
download categories/necklaces.jpg  "jewelry,necklace" 102 "Necklaces"
download categories/bracelets.jpg  "jewelry,bracelet" 103 "Bracelets"
download categories/earrings.jpg   "jewelry,earring"  104 "Earrings"
download categories/pendants.jpg     "jewelry,pendant"  105 "Pendants"

# Collections
download collections/bridal.jpg    "wedding,jewelry" 201 "Bridal"
download collections/everyday.jpg  "jewelry,fashion" 202 "Everyday"
download collections/signature.jpg "gold,jewelry"    203 "Signature"

# Rings
download products/rings/01.jpg "jewelry,ring"      301 "Ring 1"
download products/rings/02.jpg "diamond,ring"      302 "Ring 2"
download products/rings/03.jpg "gold,ring"        303 "Ring 3"
download products/rings/04.jpg "engagement,ring"  304 "Ring 4"
download products/rings/05.jpg "jewelry,ring"      305 "Ring 5"
download products/rings/06.jpg "silver,ring"      306 "Ring 6"

# Necklaces
download products/necklaces/01.jpg "jewelry,necklace" 401 "Necklace 1"
download products/necklaces/02.jpg "gold,necklace"     402 "Necklace 2"
download products/necklaces/03.jpg "pearl,necklace"    403 "Necklace 3"
download products/necklaces/04.jpg "diamond,necklace"  404 "Necklace 4"
download products/necklaces/05.jpg "jewelry,necklace"  405 "Necklace 5"

# Bracelets
download products/bracelets/01.jpg "jewelry,bracelet" 501 "Bracelet 1"
download products/bracelets/02.jpg "gold,bracelet"     502 "Bracelet 2"
download products/bracelets/03.jpg "diamond,bracelet"  503 "Bracelet 3"
download products/bracelets/04.jpg "jewelry,bracelet"  504 "Bracelet 4"

# Earrings
download products/earrings/01.jpg "jewelry,earring" 601 "Earring 1"
download products/earrings/02.jpg "gold,earring"     602 "Earring 2"
download products/earrings/03.jpg "pearl,earring"    603 "Earring 3"
download products/earrings/04.jpg "diamond,earring"  604 "Earring 4"

# Pendants
download products/pendants/01.jpg "jewelry,pendant" 701 "Pendant 1"
download products/pendants/02.jpg "gold,pendant"     702 "Pendant 2"
download products/pendants/03.jpg "silver,pendant"    703 "Pendant 3"

count="$(find "$DEST" -type f -name '*.jpg' | wc -l)"
echo "done: ${count} images in $DEST"
