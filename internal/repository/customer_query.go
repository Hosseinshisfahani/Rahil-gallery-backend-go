package repository

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/rahil-gallery/rahil-gallery-server/internal/model"
)

const listBaseFrom = `
FROM customers c
WHERE c.deleted_at IS NULL`

const (
	nameSearchExpr = `lower(trim(c.first_name || ' ' || c.last_name))`
	phoneNormExpr  = `c.phone_digits`
)

type quickSearchMode int

const (
	quickSearchNone quickSearchMode = iota
	quickSearchName
	quickSearchPhone
	quickSearchUnion
)

type listQuery struct {
	where string
	args  []any
}

func buildListQuery(filter model.ListFilter, args []any) listQuery {
	if filter.HasAdvancedFilters() {
		return buildAdvancedQuery(filter, args)
	}

	if q := strings.TrimSpace(filter.QuickSearch); q != "" {
		return buildQuickSearchQuery(q, args)
	}

	return listQuery{args: args}
}

func buildQuickSearchQuery(q string, args []any) listQuery {
	mode, namePattern, phonePattern := analyzeQuickSearch(q)
	next := len(args) + 1

	switch mode {
	case quickSearchName:
		clause := fmt.Sprintf("%s LIKE $%d", nameSearchExpr, next)
		args = append(args, namePattern)
		return listQuery{where: "AND " + clause, args: args}

	case quickSearchPhone:
		clause := fmt.Sprintf("%s LIKE $%d", phoneNormExpr, next)
		args = append(args, phonePattern)
		return listQuery{where: "AND " + clause, args: args}

	case quickSearchUnion:
		clause := fmt.Sprintf("(%s LIKE $%d OR %s LIKE $%d)", nameSearchExpr, next, phoneNormExpr, next+1)
		args = append(args, namePattern, phonePattern)
		return listQuery{where: "AND " + clause, args: args}
	default:
		return listQuery{args: args}
	}
}

func analyzeQuickSearch(q string) (quickSearchMode, string, string) {
	digits := extractDigits(q)
	hasText := hasNonDigitText(q)

	switch {
	case len(digits) >= 4 && hasText:
		return quickSearchUnion, "%" + strings.ToLower(q) + "%", digits + "%"
	case len(digits) >= 4:
		return quickSearchPhone, "", digits + "%"
	default:
		return quickSearchName, "%" + strings.ToLower(q) + "%", ""
	}
}

func hasNonDigitText(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) && !unicode.IsSpace(r) {
			return true
		}
	}
	return false
}

func extractDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func buildAdvancedQuery(filter model.ListFilter, args []any) listQuery {
	var conds []string

	if filter.CustomerID != nil {
		next := len(args) + 1
		args = append(args, *filter.CustomerID)
		conds = append(conds, fmt.Sprintf("c.id = $%d", next))
	}

	if email := strings.TrimSpace(filter.Email); email != "" {
		next := len(args) + 1
		args = append(args, "%"+strings.ToLower(email)+"%")
		conds = append(conds, fmt.Sprintf("lower(c.email) LIKE $%d", next))
	}

	if age := strings.TrimSpace(filter.CustomerAgeRange); age != "" {
		next := len(args) + 1
		args = append(args, age)
		conds = append(conds, fmt.Sprintf("c.customer_age_range = $%d", next))
	}

	if gender := strings.TrimSpace(filter.Gender); gender != "" && gender != "all" {
		next := len(args) + 1
		args = append(args, gender)
		conds = append(conds, fmt.Sprintf("c.gender = $%d", next))
	}

	if len(filter.CustomerTypes) > 0 {
		next := len(args) + 1
		args = append(args, filter.CustomerTypes)
		conds = append(conds, fmt.Sprintf("c.customer_type = ANY($%d)", next))
	}

	if len(filter.PurchaseTypes) > 0 {
		next := len(args) + 1
		args = append(args, filter.PurchaseTypes)
		conds = append(conds, fmt.Sprintf("c.purchased_categories && $%d::text[]", next))
	}

	appendDateRange(&conds, &args, "c.first_visit_date", filter.FirstVisitFrom, filter.FirstVisitTo)
	appendDateRange(&conds, &args, "c.birthday", filter.BirthdayFrom, filter.BirthdayTo)
	appendDateRange(&conds, &args, "c.marriage_date", filter.MarriageFrom, filter.MarriageTo)

	where := ""
	if len(conds) > 0 {
		where = "AND " + strings.Join(conds, " AND ")
	}
	return listQuery{where: where, args: args}
}

func appendDateRange(conds *[]string, args *[]any, column string, from, to *time.Time) {
	if from != nil {
		next := len(*args) + 1
		*args = append(*args, *from)
		*conds = append(*conds, fmt.Sprintf("%s >= $%d", column, next))
	}
	if to != nil {
		next := len(*args) + 1
		*args = append(*args, *to)
		*conds = append(*conds, fmt.Sprintf("%s <= $%d", column, next))
	}
}

const listSelectColumns = `
SELECT
  c.id,
  TRIM(c.first_name || ' ' || c.last_name) AS full_name,
  c.phone,
  c.email,
  c.customer_type,
  c.customer_age_range,
  c.gender,
  c.purchased_categories,
  c.created_at`
