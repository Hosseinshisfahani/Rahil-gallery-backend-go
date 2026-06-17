package seed

import (
	"encoding/json"
	"fmt"
	"time"
)

// CRM enum values — keep in sync with client mock-customers / admin i18n keys.
var (
	CRMCustomerTypes = []string{
		"foreign_and_tour_guidance",
		"vip",
		"public",
		"colleagues",
		"family_and_friends",
	}

	CRMPurchasedCategories = []string{
		"gold_and_gemstones",
		"gold_and_stones",
		"silver_and_gemstones",
		"silver_and_stones",
		"gemstones_and_special_roughs",
		"stones_and_roughs",
	}
)

type seedImportProfile struct {
	FirstName           string   `json:"firstName"`
	LastName            string   `json:"lastName"`
	Phone               string   `json:"phone"`
	Email               string   `json:"email,omitempty"`
	Job                 string   `json:"job,omitempty"`
	Address             string   `json:"address,omitempty"`
	CustomerType        string   `json:"customerType"`
	CustomerAgeRange    string   `json:"customerAgeRange,omitempty"`
	Gender              string   `json:"gender,omitempty"`
	PurchasedCategories []string `json:"purchasedCategories"`
	FirstVisitDate      string   `json:"firstVisitDate,omitempty"`
	Birthday            string   `json:"birthday,omitempty"`
	MarriageDate        string   `json:"marriageDate,omitempty"`
	Description         string   `json:"description,omitempty"`
	Signature           string   `json:"signature,omitempty"`
}

func marshalImportProfile(p seedImportProfile) (string, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func fixtureImportProfile(
	first, last, phone string,
	email *string,
	customerType string,
	categories []string,
	now time.Time,
	registeredDaysAgo int,
) (string, error) {
	p := seedImportProfile{
		FirstName:           first,
		LastName:            last,
		Phone:               phone,
		CustomerType:        customerType,
		CustomerAgeRange:    "21-40",
		Gender:              "female",
		PurchasedCategories: categories,
		FirstVisitDate:      now.AddDate(0, 0, -registeredDaysAgo+14).Format("2006-01-02"),
		Birthday:            "1990-05-12",
		Description:         fmt.Sprintf("%s %s — fixture CRM profile", first, last),
	}
	if email != nil {
		p.Email = *email
	}
	return marshalImportProfile(p)
}
