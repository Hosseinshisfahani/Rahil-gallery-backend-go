package customer

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/customer"
)

// Denormalized segment on customer_profiles (see migration 000008).
const segmentExpr = `COALESCE(cp.segment, 'new')`

const statusExpr = `CASE WHEN u.status = 'banned' THEN 'blocked' ELSE 'active' END`

const crmCustomerTypeExpr = `COALESCE(NULLIF(cp.crm_customer_type, ''), 'public')`

const crmPurchasedCategoriesExpr = `COALESCE(
  (SELECT array_agg(elem ORDER BY elem)
   FROM jsonb_array_elements_text(COALESCE(cp.import_profile->'purchasedCategories', '[]'::jsonb)) AS elem),
  ARRAY['gold_and_stones']::text[]
)`

const crmAgeRangeExpr = `NULLIF(COALESCE(cp.crm_age_range, cp.import_profile->>'customerAgeRange'), '')`

const crmGenderExpr = `NULLIF(COALESCE(cp.crm_gender, cp.import_profile->>'gender'), '')`

const listBaseFrom = `
FROM users u
INNER JOIN roles r ON r.id = u.role_id AND r.name = $1
LEFT JOIN customer_profiles cp ON cp.user_id = u.id
WHERE u.deleted_at IS NULL`

const (
	nameSearchExpr  = `lower(trim(u.first_name || ' ' || u.last_name))`
	phoneNormExpr   = `u.phone_digits`
	nameSearchExpr2 = `lower(trim(u2.first_name || ' ' || u2.last_name))`
	phoneNormExpr2  = `u2.phone_digits`
)

var crmDateColumns = map[string]string{
	"firstVisit": "crm_first_visit_date",
	"birthday":   "crm_birthday",
	"marriage":   "crm_marriage_date",
}

type quickSearchMode int

const (
	quickSearchNone quickSearchMode = iota
	quickSearchName
	quickSearchPhone
	quickSearchUnion
)

type listQuery struct {
	cte   string // optional leading WITH clause
	from  string
	where string
	args  []any
}

const advancedProfileCTETpl = `WITH profile_filter AS MATERIALIZED (
  SELECT user_id FROM customer_profiles
  WHERE %s
)`

// Role check in JOIN avoids hashing all customers when resolving profile_filter rows.
const advancedProfileFrom = `
FROM profile_filter pf
INNER JOIN users u ON u.id = pf.user_id
  AND u.deleted_at IS NULL
  AND u.role_id = (SELECT id FROM roles WHERE name = $1)
LEFT JOIN customer_profiles cp ON cp.user_id = u.id`

// buildListQuery assembles FROM/WHERE for list and count queries.
func buildListQuery(filter domain.ListFilter, args []any) listQuery {
	if filter.HasAdvancedFilters() {
		return buildAdvancedQuery(filter, args)
	}

	if q := strings.TrimSpace(filter.QuickSearch); q != "" {
		return buildQuickSearchQuery(q, args)
	}

	return listQuery{from: listBaseFrom, args: args}
}

func buildQuickSearchQuery(q string, args []any) listQuery {
	mode, namePattern, phonePattern := analyzeQuickSearch(q)
	next := 2

	switch mode {
	case quickSearchName:
		clause := fmt.Sprintf("%s LIKE $%d", nameSearchExpr, next)
		args = append(args, namePattern)
		return listQuery{from: listBaseFrom, where: "AND " + clause, args: args}

	case quickSearchPhone:
		clause := fmt.Sprintf("%s LIKE $%d", phoneNormExpr, next)
		args = append(args, phonePattern)
		return listQuery{from: listBaseFrom, where: "AND " + clause, args: args}

	case quickSearchUnion:
		unionJoin := fmt.Sprintf(`
INNER JOIN (
  SELECT u2.id FROM users u2
  INNER JOIN roles r2 ON r2.id = u2.role_id AND r2.name = $1
  WHERE u2.deleted_at IS NULL AND %s LIKE $%d
  UNION
  SELECT u2.id FROM users u2
  INNER JOIN roles r2 ON r2.id = u2.role_id AND r2.name = $1
  WHERE u2.deleted_at IS NULL AND %s LIKE $%d
) qs ON qs.id = u.id`, nameSearchExpr2, next, phoneNormExpr2, next+1)
		args = append(args, namePattern, phonePattern)
		return listQuery{from: listBaseFrom + unionJoin, args: args}
	default:
		return listQuery{from: listBaseFrom, args: args}
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

type advancedFilterPlan struct {
	profileConds []string
	userConds    []string
}

func buildAdvancedQuery(filter domain.ListFilter, args []any) listQuery {
	plan, args, _ := buildAdvancedFilterPlan(filter, args, 2)

	if len(plan.profileConds) > 0 {
		return listQuery{
			cte:   fmt.Sprintf(advancedProfileCTETpl, strings.Join(plan.profileConds, " AND ")),
			from:  advancedProfileFrom,
			where: formatWhere(plan.userConds),
			args:  args,
		}
	}

	return listQuery{from: listBaseFrom, where: appendWhere(plan.userConds), args: args}
}

func appendWhere(conds []string) string {
	if len(conds) == 0 {
		return ""
	}
	return "AND " + strings.Join(conds, " AND ")
}

func formatWhere(conds []string) string {
	if len(conds) == 0 {
		return ""
	}
	return "WHERE " + strings.Join(conds, " AND ")
}

func buildAdvancedFilterPlan(filter domain.ListFilter, args []any, next int) (advancedFilterPlan, []any, int) {
	var plan advancedFilterPlan

	if filter.CustomerID != nil {
		plan.userConds = append(plan.userConds, fmt.Sprintf("u.id = $%d", next))
		args = append(args, *filter.CustomerID)
		next++
	}

	if email := strings.TrimSpace(strings.ToLower(filter.Email)); email != "" {
		plan.userConds = append(plan.userConds, fmt.Sprintf("lower(u.email) LIKE $%d", next))
		args = append(args, "%"+email+"%")
		next++
	}

	if filter.Segment != "" {
		plan.profileConds = append(plan.profileConds, fmt.Sprintf("segment = $%d", next))
		args = append(args, filter.Segment)
		next++
	}

	if filter.Status == "active" {
		plan.userConds = append(plan.userConds, "u.status = 'active'")
	} else if filter.Status == "blocked" {
		plan.userConds = append(plan.userConds, "u.status = 'banned'")
	}

	if filter.VIP != nil {
		plan.profileConds = append(plan.profileConds, fmt.Sprintf("is_vip = $%d", next))
		args = append(args, *filter.VIP)
		next++
	}

	if filter.LTVMin != nil {
		plan.profileConds = append(plan.profileConds, fmt.Sprintf("COALESCE(total_ltv, 0) >= $%d", next))
		args = append(args, *filter.LTVMin)
		next++
	}
	if filter.LTVMax != nil {
		plan.profileConds = append(plan.profileConds, fmt.Sprintf("COALESCE(total_ltv, 0) <= $%d", next))
		args = append(args, *filter.LTVMax)
		next++
	}

	if filter.OrdersMin != nil {
		plan.profileConds = append(plan.profileConds, fmt.Sprintf("COALESCE(total_orders, 0) >= $%d", next))
		args = append(args, *filter.OrdersMin)
		next++
	}
	if filter.OrdersMax != nil {
		plan.profileConds = append(plan.profileConds, fmt.Sprintf("COALESCE(total_orders, 0) <= $%d", next))
		args = append(args, *filter.OrdersMax)
		next++
	}

	if filter.RegisteredFrom != nil {
		plan.userConds = append(plan.userConds, fmt.Sprintf("u.created_at >= $%d", next))
		args = append(args, *filter.RegisteredFrom)
		next++
	}
	if filter.RegisteredTo != nil {
		plan.userConds = append(plan.userConds, fmt.Sprintf("u.created_at <= $%d", next))
		args = append(args, *filter.RegisteredTo)
		next++
	}

	if filter.LastPurchaseFrom != nil {
		plan.profileConds = append(plan.profileConds, fmt.Sprintf("last_purchase_at >= $%d", next))
		args = append(args, *filter.LastPurchaseFrom)
		next++
	}
	if filter.LastPurchaseTo != nil {
		plan.profileConds = append(plan.profileConds, fmt.Sprintf("last_purchase_at <= $%d", next))
		args = append(args, *filter.LastPurchaseTo)
		next++
	}

	if filter.LastActivityFrom != nil {
		plan.profileConds = append(plan.profileConds, fmt.Sprintf("last_activity_at >= $%d", next))
		args = append(args, *filter.LastActivityFrom)
		next++
	}
	if filter.LastActivityTo != nil {
		plan.profileConds = append(plan.profileConds, fmt.Sprintf("last_activity_at <= $%d", next))
		args = append(args, *filter.LastActivityTo)
		next++
	}

	if len(filter.Tags) > 0 {
		plan.profileConds = append(plan.profileConds, fmt.Sprintf("tags && $%d", next))
		args = append(args, filter.Tags)
		next++
	}

	switch filter.HasPurchased {
	case "yes":
		plan.profileConds = append(plan.profileConds, "COALESCE(total_orders, 0) > 0")
	case "no":
		plan.profileConds = append(plan.profileConds, "COALESCE(total_orders, 0) = 0")
	}

	if ageRange := strings.TrimSpace(filter.CustomerAgeRange); ageRange != "" {
		plan.profileConds = append(plan.profileConds, fmt.Sprintf("crm_age_range = $%d", next))
		args = append(args, ageRange)
		next++
	}

	if gender := strings.TrimSpace(filter.Gender); gender != "" {
		plan.profileConds = append(plan.profileConds, fmt.Sprintf("crm_gender = $%d", next))
		args = append(args, gender)
		next++
	}

	if len(filter.CustomerTypes) > 0 {
		plan.profileConds = append(plan.profileConds, fmt.Sprintf("crm_customer_type = ANY($%d)", next))
		args = append(args, filter.CustomerTypes)
		next++
	}

	if len(filter.PurchaseTypes) > 0 {
		plan.profileConds = append(plan.profileConds, fmt.Sprintf("(import_profile->'purchasedCategories') ?| $%d", next))
		args = append(args, filter.PurchaseTypes)
		next++
	}

	plan, args, next = addCRMDateRange(plan, args, next, "firstVisit", filter.FirstVisitFrom, filter.FirstVisitTo)
	plan, args, next = addCRMDateRange(plan, args, next, "birthday", filter.BirthdayFrom, filter.BirthdayTo)
	plan, args, next = addCRMDateRange(plan, args, next, "marriage", filter.MarriageFrom, filter.MarriageTo)

	return plan, args, next
}

func addCRMDateRange(plan advancedFilterPlan, args []any, next int, key string, from, to *time.Time) (advancedFilterPlan, []any, int) {
	column := crmDateColumns[key]
	if from != nil {
		plan.profileConds = append(plan.profileConds, fmt.Sprintf("%s >= $%d", column, next))
		args = append(args, from.Format("2006-01-02"))
		next++
	}
	if to != nil {
		plan.profileConds = append(plan.profileConds, fmt.Sprintf("%s <= $%d", column, next))
		args = append(args, to.Format("2006-01-02"))
		next++
	}
	return plan, args, next
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
