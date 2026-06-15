package seed

import (
	"strings"
	"testing"
)

func TestProductImageURLUsesLocalStaticPath(t *testing.T) {
	url := productImageURL("ring", 0)
	if !strings.HasPrefix(url, defaultAssetBaseURL+catalogStaticPrefix+"/") {
		t.Fatalf("expected local static url, got %s", url)
	}
	if !strings.HasSuffix(url, ".jpg") {
		t.Fatalf("expected jpg asset, got %s", url)
	}
}

func TestBulkImageURLMatchesJewelryType(t *testing.T) {
	ring := bulkImageURL(0)
	necklace := bulkImageURL(1)
	if ring == necklace {
		t.Fatalf("expected different urls per jewelry type: %s", ring)
	}
}
