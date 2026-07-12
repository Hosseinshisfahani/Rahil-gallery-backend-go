package customer

import (
	"strings"
	"time"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

const DefaultCustomerType = "public"

var DefaultPurchasedCategories = []string{"gold_and_stones"}

type Customer struct {
	ID                  shared.ID
	FirstName           string
	LastName            string
	Job                 *string
	Phone               string
	Email               *string
	Address             *string
	Birthday            *time.Time
	MarriageDate        *time.Time
	ImportantDate       *time.Time
	FirstVisitDate      *time.Time
	Gender              *string
	CustomerType        string
	CustomerAgeRange    *string
	PurchasedCategories []string
	Description         *string
	SignatureURL        *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (c Customer) FullName() string {
	return strings.TrimSpace(c.FirstName + " " + c.LastName)
}

type ListRow struct {
	ID                  shared.ID
	FullName            string
	Phone               string
	Email               *string
	CustomerType        string
	CustomerAgeRange    *string
	Gender              *string
	PurchasedCategories []string
	CreatedAt           time.Time
}

type ListFilter struct {
	QuickSearch string

	CustomerID *shared.ID
	Email      string

	CustomerAgeRange string
	Gender           string
	CustomerTypes    []string
	PurchaseTypes    []string
	FirstVisitFrom   *time.Time
	FirstVisitTo     *time.Time
	BirthdayFrom     *time.Time
	BirthdayTo       *time.Time
	MarriageFrom     *time.Time
	MarriageTo       *time.Time

	SkipCount bool
}

func (f ListFilter) HasAdvancedFilters() bool {
	return f.CustomerID != nil ||
		strings.TrimSpace(f.Email) != "" ||
		strings.TrimSpace(f.CustomerAgeRange) != "" ||
		strings.TrimSpace(f.Gender) != "" ||
		len(f.CustomerTypes) > 0 ||
		len(f.PurchaseTypes) > 0 ||
		f.FirstVisitFrom != nil ||
		f.FirstVisitTo != nil ||
		f.BirthdayFrom != nil ||
		f.BirthdayTo != nil ||
		f.MarriageFrom != nil ||
		f.MarriageTo != nil
}

type ListResult struct {
	Items      []ListRow
	Total      int
	Page       int
	PerPage    int
	TotalPages int
	HasMore    bool
	TotalExact bool
}

type Input struct {
	FirstName           string
	LastName            string
	Job                 *string
	Phone               string
	Email               *string
	Address             *string
	Birthday            *time.Time
	MarriageDate        *time.Time
	ImportantDate       *time.Time
	FirstVisitDate      *time.Time
	Gender              *string
	CustomerType        string
	CustomerAgeRange    *string
	PurchasedCategories []string
	Description         *string
	SignatureURL        *string
}

func NormalizeListRow(row *ListRow) {
	if row == nil {
		return
	}
	if row.CustomerType == "" {
		row.CustomerType = DefaultCustomerType
	}
	if len(row.PurchasedCategories) == 0 {
		row.PurchasedCategories = append([]string(nil), DefaultPurchasedCategories...)
	}
}

func NormalizeInput(in *Input) {
	if in == nil {
		return
	}
	if strings.TrimSpace(in.CustomerType) == "" {
		in.CustomerType = DefaultCustomerType
	}
	if len(in.PurchasedCategories) == 0 {
		in.PurchasedCategories = append([]string(nil), DefaultPurchasedCategories...)
	}
}

func ParseDate(s *string) (*time.Time, error) {
	if s == nil {
		return nil, nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func FormatDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}
