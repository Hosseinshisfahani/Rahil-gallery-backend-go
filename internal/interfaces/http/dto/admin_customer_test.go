package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/customer"
)

func TestToCustomerSummary_includesCRMFields(t *testing.T) {
	customerType := "vip"
	ageRange := "21-40"
	gender := "female"
	row := domain.ListRow{
		ID:                  uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		FullName:            "Sara Mohammadi",
		Phone:               "+989121234567",
		RegisteredAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Segment:             domain.SegmentVIP,
		Status:              "active",
		CustomerType:        &customerType,
		PurchasedCategories: []string{"gold_and_gemstones", "silver_and_stones"},
		CustomerAgeRange:    &ageRange,
		Gender:              &gender,
	}

	resp := ToCustomerSummary(row)
	if resp.CustomerType != "vip" {
		t.Fatalf("customerType = %q, want vip", resp.CustomerType)
	}
	if len(resp.PurchasedCategories) != 2 {
		t.Fatalf("purchasedCategories = %#v", resp.PurchasedCategories)
	}
	if resp.CustomerAgeRange == nil || *resp.CustomerAgeRange != "21-40" {
		t.Fatalf("customerAgeRange = %#v", resp.CustomerAgeRange)
	}
	if resp.Gender == nil || *resp.Gender != "female" {
		t.Fatalf("gender = %#v", resp.Gender)
	}

	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	if raw["customerType"] != "vip" {
		t.Fatalf("json customerType = %#v", raw["customerType"])
	}
	cats, ok := raw["purchasedCategories"].([]any)
	if !ok || len(cats) != 2 {
		t.Fatalf("json purchasedCategories = %#v", raw["purchasedCategories"])
	}
}

func TestEnrichListRowCRMFromImportProfile(t *testing.T) {
	row := domain.ListRow{}
	profile := json.RawMessage(`{
		"customerType": "public",
		"customerAgeRange": "14-21",
		"gender": "male",
		"purchasedCategories": ["gold_and_stones"]
	}`)

	enrichListRowCRMFromImportProfile(&row, profile)

	if row.CustomerType == nil || *row.CustomerType != "public" {
		t.Fatalf("customerType = %#v", row.CustomerType)
	}
	if row.CustomerAgeRange == nil || *row.CustomerAgeRange != "14-21" {
		t.Fatalf("customerAgeRange = %#v", row.CustomerAgeRange)
	}
	if row.Gender == nil || *row.Gender != "male" {
		t.Fatalf("gender = %#v", row.Gender)
	}
	if len(row.PurchasedCategories) != 1 || row.PurchasedCategories[0] != "gold_and_stones" {
		t.Fatalf("purchasedCategories = %#v", row.PurchasedCategories)
	}
}
