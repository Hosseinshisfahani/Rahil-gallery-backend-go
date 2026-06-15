package seed

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

const (
	catalogImagesDir     = "data/catalog-images"
	catalogStaticPrefix  = "/static/catalog"
	defaultAssetBaseURL  = "http://localhost:8080"
)

type catalogManifest struct {
	Categories  map[string]string   `json:"categories"`
	Collections map[string]string   `json:"collections"`
	Products    map[string][]string `json:"products"`
}

var (
	manifestOnce sync.Once
	manifestData catalogManifest
	manifestErr  error
	assetBaseURL string
)

func init() {
	assetBaseURL = os.Getenv("SEED_ASSET_BASE_URL")
	if assetBaseURL == "" {
		assetBaseURL = defaultAssetBaseURL
	}
}

func loadManifest() {
	manifestOnce.Do(func() {
		path := filepath.Join(catalogImagesDir, "manifest.json")
		raw, err := os.ReadFile(path)
		if err != nil {
			manifestErr = err
			return
		}
		manifestErr = json.Unmarshal(raw, &manifestData)
	})
}

func catalogAssetURL(relPath string) string {
	return assetBaseURL + catalogStaticPrefix + "/" + relPath
}

func categoryImageURL(slug string) string {
	loadManifest()
	if manifestErr != nil {
		return fallbackProductImageURL("ring", 0)
	}
	if rel, ok := manifestData.Categories[slug]; ok {
		return catalogAssetURL(rel)
	}
	return catalogAssetURL("categories/rings.jpg")
}

func collectionImageURL(slug string) string {
	loadManifest()
	if manifestErr != nil {
		return catalogAssetURL("collections/bridal.jpg")
	}
	if rel, ok := manifestData.Collections[slug]; ok {
		return catalogAssetURL(rel)
	}
	return catalogAssetURL("collections/everyday.jpg")
}

func productImageURL(jewelryType string, index int) string {
	loadManifest()
	if manifestErr != nil {
		return fallbackProductImageURL(jewelryType, index)
	}
	paths, ok := manifestData.Products[jewelryType]
	if !ok || len(paths) == 0 {
		return fallbackProductImageURL(jewelryType, index)
	}
	return catalogAssetURL(paths[index%len(paths)])
}

func fallbackProductImageURL(jewelryType string, index int) string {
	switch jewelryType {
	case "necklace":
		return catalogAssetURL("products/necklaces/01.jpg")
	case "bracelet":
		return catalogAssetURL("products/bracelets/01.jpg")
	case "earring":
		return catalogAssetURL("products/earrings/01.jpg")
	case "pendant":
		return catalogAssetURL("products/pendants/01.jpg")
	default:
		fallbacks := []string{
			"products/rings/01.jpg",
			"products/rings/02.jpg",
			"products/rings/03.jpg",
		}
		return catalogAssetURL(fallbacks[index%len(fallbacks)])
	}
}

func bulkImageURL(index int) string {
	jt := bulkJewelryType(index)
	return productImageURL(jt, index)
}

func fixtureProductImageURL(jewelryType string, slot int) string {
	return productImageURL(jewelryType, slot)
}
