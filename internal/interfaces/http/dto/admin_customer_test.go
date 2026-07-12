package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/customer"
)

func TestToCustomerSummary_includesCRMFields(t *testing.T) {
	ageRange := "21-40"
	gender := "female"
	row := domain.ListRow{
		ID:                  uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		FullName:            "Sara Mohammadi",
		Phone:               "+989121234567",
		CustomerType:        "vip",
		PurchasedCategories: []string{"gold_and_gemstones", "silver_and_stones"},
		CustomerAgeRange:    &ageRange,
		Gender:              &gender,
		CreatedAt:           time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
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
}

func TestImportProfileToInput(t *testing.T) {
	job := "Designer"
	email := "sara@example.com"
	input, err := importProfileToInput(ImportProfileRequest{
		FirstName:           "Sara",
		LastName:            "Mohammadi",
		Job:                 &job,
		Phone:               "+989121234567",
		Email:               &email,
		CustomerType:        "public",
		PurchasedCategories: []string{"gold_and_stones"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if input.FirstName != "Sara" || input.Phone != "+989121234567" {
		t.Fatalf("input = %#v", input)
	}
}
