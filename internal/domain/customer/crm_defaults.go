package customer

import "encoding/json"

const DefaultCustomerType = "public"

var DefaultPurchasedCategories = []string{"gold_and_stones"}

func NormalizeListRowCRM(row *ListRow) {
	if row == nil {
		return
	}
	if row.CustomerType == nil || *row.CustomerType == "" {
		t := DefaultCustomerType
		row.CustomerType = &t
	}
	if len(row.PurchasedCategories) == 0 {
		row.PurchasedCategories = append([]string(nil), DefaultPurchasedCategories...)
	}
}

func MarshalMinimalImportProfile(first, last, phone string) (json.RawMessage, error) {
	return MarshalImportProfile(ImportProfileInput{
		FirstName:           first,
		LastName:            last,
		Phone:               phone,
		CustomerType:        DefaultCustomerType,
		PurchasedCategories: append([]string(nil), DefaultPurchasedCategories...),
	})
}
