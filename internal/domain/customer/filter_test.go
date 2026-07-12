package customer

import "testing"

func TestListFilter_HasAdvancedFilters(t *testing.T) {
	if (ListFilter{QuickSearch: "ali"}).HasAdvancedFilters() {
		t.Fatal("quick search alone is not advanced")
	}
	if !(ListFilter{Email: "a@b.com"}).HasAdvancedFilters() {
		t.Fatal("email is advanced")
	}
	if !(ListFilter{CustomerAgeRange: "21-40"}).HasAdvancedFilters() {
		t.Fatal("age range is advanced")
	}
	if !(ListFilter{PurchaseTypes: []string{"gold_and_stones"}}).HasAdvancedFilters() {
		t.Fatal("purchase type is advanced")
	}
}
