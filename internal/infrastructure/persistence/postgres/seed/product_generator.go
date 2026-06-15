package seed

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var productBulkNamespace = uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")

const bulkProductSKUPrefix = "RG-SEED-"

var fixtureCategories = []struct {
	id          uuid.UUID
	name, slug  string
	description string
	sort        int
}{
	{IDCategoryRings, "Rings", "rings", "Engagement and fashion rings", 1},
	{IDCategoryNecklaces, "Necklaces", "necklaces", "Chains, pendants, and statement necklaces", 2},
	{IDCategoryBracelets, "Bracelets", "bracelets", "Bangles, cuffs, and tennis bracelets", 3},
	{IDCategoryEarrings, "Earrings", "earrings", "Studs, hoops, and drop earrings", 4},
	{IDCategoryPendants, "Pendants", "pendants", "Charms and standalone pendants", 5},
}

var fixtureCollections = []struct {
	id          uuid.UUID
	name, slug  string
	description string
}{
	{IDCollectionBridal, "Bridal", "bridal", "Wedding and engagement favorites"},
	{IDCollectionEveryday, "Everyday", "everyday", "Wearable pieces for daily elegance"},
	{IDCollectionSignature, "Signature", "signature", "House bestsellers and hero pieces"},
}

var jewelryTypes = []string{"ring", "necklace", "bracelet", "earring", "pendant"}
var metalTypes = []string{"gold_yellow", "gold_white", "gold_rose", "silver", "platinum"}
var gemstoneTypes = []string{"diamond", "ruby", "sapphire", "emerald", "pearl", "turquoise", "none"}

var productNameStems = []string{
	"Aurora", "Celeste", "Noor", "Roya", "Shirin", "Laleh", "Mahi", "Sepideh",
	"Yasmin", "Parisa", "Darya", "Setareh", "Golshan", "Narges", "Bahareh",
}

var productNameKinds = map[string][]string{
	"ring":      {"Solitaire Ring", "Eternity Band", "Cocktail Ring", "Signet Ring", "Halo Ring"},
	"necklace":  {"Pendant Necklace", "Tennis Necklace", "Chain Necklace", "Choker", "Lariat"},
	"bracelet":  {"Tennis Bracelet", "Bangle", "Cuff Bracelet", "Chain Bracelet", "Charm Bracelet"},
	"earring":   {"Stud Earrings", "Hoop Earrings", "Drop Earrings", "Huggie Earrings", "Chandelier Earrings"},
	"pendant":   {"Heart Pendant", "Teardrop Pendant", "Medallion Pendant", "Cross Pendant", "Coin Pendant"},
}

func bulkProductID(index int) uuid.UUID {
	return uuid.NewSHA1(productBulkNamespace, []byte(fmt.Sprintf("product:%d", index)))
}

func bulkVariantID(productIndex, variantIndex int) uuid.UUID {
	return uuid.NewSHA1(productBulkNamespace, []byte(fmt.Sprintf("variant:%d:%d", productIndex, variantIndex)))
}

func bulkProductImageID(productIndex int) uuid.UUID {
	return uuid.NewSHA1(productBulkNamespace, []byte(fmt.Sprintf("image:%d", productIndex)))
}

func bulkProductSKU(index int) string {
	return fmt.Sprintf("%s%08d", bulkProductSKUPrefix, index)
}

func bulkVariantSKU(productIndex, variantIndex int) string {
	return fmt.Sprintf("%s%08d-%02d", bulkProductSKUPrefix, productIndex, variantIndex+1)
}

func bulkProductName(index int) (name, slug string) {
	jt := jewelryTypes[index%len(jewelryTypes)]
	stem := productNameStems[index%len(productNameStems)]
	kinds := productNameKinds[jt]
	kind := kinds[index%len(kinds)]
	name = fmt.Sprintf("%s %s", stem, kind)
	slug = slugify(name) + fmt.Sprintf("-%d", index)
	return name, slug
}

func bulkCategoryID(index int) uuid.UUID {
	jt := jewelryTypes[index%len(jewelryTypes)]
	switch jt {
	case "ring":
		return IDCategoryRings
	case "necklace":
		return IDCategoryNecklaces
	case "bracelet":
		return IDCategoryBracelets
	case "earring":
		return IDCategoryEarrings
	default:
		return IDCategoryPendants
	}
}

func bulkJewelryType(index int) string {
	return jewelryTypes[index%len(jewelryTypes)]
}

func bulkMetalType(index int) string {
	return metalTypes[index%len(metalTypes)]
}

func bulkGemstoneType(index int) string {
	return gemstoneTypes[index%len(gemstoneTypes)]
}

func bulkBasePrice(index int) float64 {
	tiers := []float64{28_000_000, 48_000_000, 72_000_000, 95_000_000, 125_000_000, 185_000_000, 240_000_000}
	return tiers[index%len(tiers)]
}

func bulkCompareAtPrice(index int, base float64) *float64 {
	if index%4 != 0 {
		return nil
	}
	v := base * 1.12
	return &v
}

func bulkKarat(index int) int16 {
	karats := []int16{14, 18, 18, 21, 22}
	return karats[index%len(karats)]
}

func bulkWeightGrams(index int) float64 {
	return 2.5 + float64(index%20)*0.35
}

func bulkIsFeatured(index int) bool {
	return index%11 == 0
}

func bulkIsHandmade(index int) bool {
	return index%7 == 0
}

func bulkProductStatus(index int) string {
	if index%17 == 0 {
		return "draft"
	}
	if index%53 == 0 {
		return "archived"
	}
	return "published"
}

func bulkVariantCount(index int) int {
	return (index % 3) + 1
}

func bulkVariantLabel(jewelryType string, variantIndex int) (name string, sizeLabel *string) {
	switch jewelryType {
	case "ring":
		size := fmt.Sprintf("%d", 12+variantIndex*2)
		return "Size " + size, &size
	case "bracelet":
		size := fmt.Sprintf("%d cm", 16+variantIndex)
		return "Wrist " + size, &size
	case "necklace":
		lengths := []string{"40 cm", "45 cm", "50 cm"}
		l := lengths[variantIndex%len(lengths)]
		return l, &l
	default:
		names := []string{"Standard", "Petite", "Grande"}
		n := names[variantIndex%len(names)]
		return n, nil
	}
}

func bulkVariantPriceAdjustment(variantIndex int) float64 {
	if variantIndex == 0 {
		return 0
	}
	return float64(variantIndex) * 4_500_000
}

func bulkInventoryQuantity(index, variantIndex int) int {
	return 5 + ((index + variantIndex*3) % 40)
}

func bulkLowStockThreshold(index int) int {
	if index%9 == 0 {
		return 8
	}
	return 3
}

func bulkPublishedAt(now time.Time, index int) *time.Time {
	if bulkProductStatus(index) != "published" {
		return nil
	}
	days := index % 800
	t := now.AddDate(0, 0, -days)
	return &t
}

func bulkCollectionForProduct(index int) *uuid.UUID {
	switch index % 5 {
	case 0:
		return &IDCollectionBridal
	case 1:
		return &IDCollectionEveryday
	case 2:
		return &IDCollectionSignature
	default:
		return nil
	}
}

func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash {
			b.WriteByte('-')
			prevDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func isBulkProductSKU(sku string) bool {
	return strings.HasPrefix(sku, bulkProductSKUPrefix)
}
