package seed

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

var bulkNamespace = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")

var (
	firstNames = []string{
		"Sara", "Ali", "Neda", "Reza", "Maryam", "Hossein", "Leila", "Omid",
		"Fatima", "Amir", "Zahra", "Mohammad", "Parisa", "Kian", "Shirin", "Babak",
		"Nazanin", "Arash", "Mahsa", "Saeed", "Yasmin", "Hamid", "Roya", "Peyman",
	}
	lastNames = []string{
		"Mohammadi", "Rezaei", "Karimi", "Ahmadi", "Hosseini", "Moradi", "Shirazi", "Farhadi",
		"Jafari", "Rahimi", "Salehi", "Nouri", "Ghasemi", "Akbari", "Ebrahimi", "Mousavi",
	}
)

func bulkCustomerID(index int) uuid.UUID {
	return uuid.NewSHA1(bulkNamespace, []byte(fmt.Sprintf("customer:%d", index)))
}

func bulkPhone(index int) string {
	return fmt.Sprintf("%s%08d", bulkPhonePrefix, index)
}

func bulkEmail(index int) string {
	return fmt.Sprintf("seed-%08d@rehil.dev", index)
}

func pickName(index int) (first, last string) {
	return firstNames[index%len(firstNames)], lastNames[(index/len(firstNames))%len(lastNames)]
}

func bulkUserStatus(index int) string {
	if index%50 == 0 {
		return "banned"
	}
	return "active"
}

func bulkLocale(index int) string {
	if index%10 == 0 {
		return "en"
	}
	return "fa"
}

func bulkIsVIP(index int) bool {
	return index%20 == 0
}

func bulkTags(index int) []string {
	switch index % 17 {
	case 0:
		return []string{"VIP", "High spender"}
	case 1:
		return []string{"Bridal customer"}
	case 2:
		return []string{"At-risk"}
	case 3:
		return []string{"Influencer lead"}
	default:
		return []string{}
	}
}

func bulkRegisteredAt(now time.Time, index int) time.Time {
	days := index % 1095 // ~3 years
	return now.AddDate(0, 0, -days)
}

// ── CRM profile generators ────────────────────────────────────────────────────

var crmGenders = []string{"male", "female", "other"}

var crmAgeRanges = []string{"1-7", "7-14", "14-21", "21-40", "40+"}

var crmCustomerTypes = CRMCustomerTypes

var crmCategories = CRMPurchasedCategories

// bulkHasCRMProfile returns true for all bulk customers.
func bulkHasCRMProfile(index int) bool {
	return true
}

// bulkCRMGender returns a gender string; ~14 % of CRM customers have it unset.
// Uses integer division (index/3) to avoid the mod-3 alignment of bulkHasCRMProfile.
func bulkCRMGender(index int) string {
	if index%7 == 0 {
		return ""
	}
	return crmGenders[(index/3)%len(crmGenders)]
}

func bulkCRMAgeRange(index int) string {
	return crmAgeRanges[(index/3)%len(crmAgeRanges)]
}

func bulkCRMCustomerType(index int) string {
	return crmCustomerTypes[(index/3)%len(crmCustomerTypes)]
}

// bulkCRMCategories returns 1 or 2 purchased categories.
// Uses integer division to avoid the mod-3 alignment of bulkHasCRMProfile.
func bulkCRMCategories(index int) []string {
	a := crmCategories[(index/3)%len(crmCategories)]
	b := crmCategories[(index/3+2)%len(crmCategories)]
	if index%5 != 0 && b != a {
		return []string{a, b}
	}
	return []string{a}
}

// bulkCRMFirstVisit returns an ISO date string 30–1830 days in the past.
func bulkCRMFirstVisit(now time.Time, index int) string {
	days := 30 + (index % 1800)
	return now.AddDate(0, 0, -days).Format("2006-01-02")
}

// bulkCRMBirthday returns an ISO birthday string (birth years 1960–2009).
func bulkCRMBirthday(index int) string {
	year := 1960 + (index % 50)
	month := 1 + (index % 12)
	day := 1 + (index % 28)
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

// bulkCRMMarriageDate returns an ISO marriage date or "" for ~33 % of customers.
func bulkCRMMarriageDate(index int) string {
	if index%3 == 0 {
		return ""
	}
	year := 2000 + (index % 24)
	month := 1 + (index % 12)
	day := 1 + (index % 28)
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}
