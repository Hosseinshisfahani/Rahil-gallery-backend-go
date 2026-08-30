package customer

import (
	"strings"
	"time"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

const DefaultCustomerType = "public"

type Customer struct {
	ID                  shared.ID
	FirstName           string
	LastName            string
	Job                 *string
	Phone               string
	Email               *string
	Address             *string
	MelliCode           *string
	PostalCode          *string
	Birthday            *time.Time
	MarriageDate        *time.Time
	ImportantDate       *time.Time
	FirstVisitDate      *time.Time
	Gender              *string
	CustomerType        string
	CustomerAgeRange    *string
	PurchasedCategories []string
	Description         *string
	MarketerNote        *string
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
	gender := strings.TrimSpace(f.Gender)
	return f.CustomerID != nil ||
		strings.TrimSpace(f.Email) != "" ||
		strings.TrimSpace(f.CustomerAgeRange) != "" ||
		(gender != "" && !strings.EqualFold(gender, "all")) ||
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
	MelliCode           *string
	PostalCode          *string
	Birthday            *time.Time
	MarriageDate        *time.Time
	ImportantDate       *time.Time
	FirstVisitDate      *time.Time
	Gender              *string
	CustomerType        string
	CustomerAgeRange    *string
	PurchasedCategories []string
	Description         *string
	MarketerNote        *string
	SignatureURL        *string
}

func NormalizeListRow(row *ListRow) {
	if row == nil {
		return
	}
	if row.CustomerType == "" {
		row.CustomerType = DefaultCustomerType
	}
}

func NormalizeInput(in *Input) {
	if in == nil {
		return
	}
	if strings.TrimSpace(in.CustomerType) == "" {
		in.CustomerType = DefaultCustomerType
	}
	in.Phone = strings.TrimSpace(NormalizeDigits(in.Phone))
	if in.MelliCode != nil {
		v := strings.TrimSpace(NormalizeDigits(*in.MelliCode))
		if v == "" {
			in.MelliCode = nil
		} else {
			in.MelliCode = &v
		}
	}
	if in.PostalCode != nil {
		v := strings.TrimSpace(NormalizeDigits(*in.PostalCode))
		if v == "" {
			in.PostalCode = nil
		} else {
			in.PostalCode = &v
		}
	}
	in.PurchasedCategories = StringSliceOrEmpty(in.PurchasedCategories)
}

// NormalizeDigits converts Persian (۰-۹) and Arabic-Indic (٠-٩) digits to their
// ASCII equivalents, leaving all other characters untouched. This keeps values
// such as phone numbers stored consistently in English digits.
func NormalizeDigits(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= '\u06F0' && r <= '\u06F9': // Persian digits ۰-۹
			b.WriteRune('0' + (r - '\u06F0'))
		case r >= '\u0660' && r <= '\u0669': // Arabic-Indic digits ٠-٩
			b.WriteRune('0' + (r - '\u0660'))
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// StringSliceOrEmpty returns a non-nil slice so Postgres TEXT[] columns receive
// '{}' instead of NULL when no categories are selected.
func StringSliceOrEmpty(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	return append([]string(nil), values...)
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
