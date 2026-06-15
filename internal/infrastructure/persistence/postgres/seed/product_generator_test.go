package seed

import (
	"strings"
	"testing"
)

func TestBulkProductSKU(t *testing.T) {
	sku := bulkProductSKU(42)
	if !strings.HasPrefix(sku, bulkProductSKUPrefix) {
		t.Fatalf("unexpected sku prefix: %s", sku)
	}
	if sku != "RG-SEED-00000042" {
		t.Fatalf("unexpected sku: %s", sku)
	}
}

func TestBulkProductIDDeterministic(t *testing.T) {
	a := bulkProductID(10)
	b := bulkProductID(10)
	if a != b {
		t.Fatal("expected deterministic product id")
	}
	if bulkProductID(10) == bulkProductID(11) {
		t.Fatal("expected different ids for different indexes")
	}
}

func TestBulkProductNameIncludesJewelryType(t *testing.T) {
	name, slug := bulkProductName(0)
	if name == "" || slug == "" {
		t.Fatal("expected non-empty name and slug")
	}
	if !strings.Contains(slug, "-") {
		t.Fatalf("expected slug with index suffix: %s", slug)
	}
}

func TestSlugify(t *testing.T) {
	if slugify("Hello World!") != "hello-world" {
		t.Fatalf("unexpected slug: %s", slugify("Hello World!"))
	}
}

func TestIsBulkProductSKU(t *testing.T) {
	if !isBulkProductSKU("RG-SEED-00000001") {
		t.Fatal("expected bulk sku match")
	}
	if isBulkProductSKU("RG-SOL-001") {
		t.Fatal("fixture sku should not match bulk prefix")
	}
}
