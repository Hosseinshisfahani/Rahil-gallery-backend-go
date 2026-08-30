package handler

import "testing"

func TestFilterFromBulkBody_genderAllKeepsQuickSearch(t *testing.T) {
	f := filterFromBulkBody(bulkSMSFilterBody{
		Query:  "حسین",
		Gender: "all",
	})
	if f.QuickSearch != "حسین" {
		t.Fatalf("QuickSearch = %q, want حسین", f.QuickSearch)
	}
	if f.Gender != "" {
		t.Fatalf("Gender = %q, want empty after normalizing all", f.Gender)
	}
	if f.HasAdvancedFilters() {
		t.Fatal("expected no advanced filters")
	}
}

func TestFilterFromBulkBody_advancedClearsQuickSearch(t *testing.T) {
	f := filterFromBulkBody(bulkSMSFilterBody{
		Query: "حسین",
		Email: "a@b.com",
	})
	if f.QuickSearch != "" {
		t.Fatalf("QuickSearch = %q, want empty when advanced filters present", f.QuickSearch)
	}
	if f.Email != "a@b.com" {
		t.Fatalf("Email = %q, want a@b.com", f.Email)
	}
}
