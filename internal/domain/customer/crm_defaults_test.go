package customer

import "testing"

func TestNormalizeListRowCRM_appliesDefaults(t *testing.T) {
	row := ListRow{}
	NormalizeListRowCRM(&row)

	if row.CustomerType == nil || *row.CustomerType != DefaultCustomerType {
		t.Fatalf("customerType = %#v", row.CustomerType)
	}
	if len(row.PurchasedCategories) != len(DefaultPurchasedCategories) {
		t.Fatalf("purchasedCategories = %#v", row.PurchasedCategories)
	}
}

func TestNormalizeListRowCRM_preservesExisting(t *testing.T) {
	customerType := "vip"
	row := ListRow{
		CustomerType:        &customerType,
		PurchasedCategories: []string{"gold_and_gemstones"},
	}
	NormalizeListRowCRM(&row)

	if *row.CustomerType != "vip" {
		t.Fatalf("customerType = %q", *row.CustomerType)
	}
	if len(row.PurchasedCategories) != 1 || row.PurchasedCategories[0] != "gold_and_gemstones" {
		t.Fatalf("purchasedCategories = %#v", row.PurchasedCategories)
	}
}
