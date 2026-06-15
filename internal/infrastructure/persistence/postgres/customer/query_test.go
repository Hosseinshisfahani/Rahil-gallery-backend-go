package customer

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/customer"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

func TestAnalyzeQuickSearch_nameOnly(t *testing.T) {
	mode, name, phone := analyzeQuickSearch("Sara")
	if mode != quickSearchName {
		t.Fatalf("mode = %v", mode)
	}
	if name != "%sara%" || phone != "" {
		t.Fatalf("name=%q phone=%q", name, phone)
	}
}

func TestAnalyzeQuickSearch_phoneOnly(t *testing.T) {
	mode, name, phone := analyzeQuickSearch("98900123456")
	if mode != quickSearchPhone {
		t.Fatalf("mode = %v", mode)
	}
	if name != "" || phone != "98900123456%" {
		t.Fatalf("name=%q phone=%q", name, phone)
	}
}

func TestAnalyzeQuickSearch_union(t *testing.T) {
	mode, name, phone := analyzeQuickSearch("ali98900")
	if mode != quickSearchUnion {
		t.Fatalf("mode = %v", mode)
	}
	if name != "%ali98900%" || phone != "98900%" {
		t.Fatalf("name=%q phone=%q", name, phone)
	}
}

func TestBuildQuickSearchQuery_nameUsesTrigramPath(t *testing.T) {
	q := buildQuickSearchQuery("ali", []any{"customer"})
	if q.where == "" {
		t.Fatal("expected where clause")
	}
	if strings.Contains(q.where, " OR ") {
		t.Fatal("name-only search must not use OR")
	}
	if strings.Contains(q.where, phoneNormExpr) {
		t.Fatal("name-only search must not query phone")
	}
	if len(q.args) != 2 {
		t.Fatalf("args = %v", q.args)
	}
}

func TestBuildQuickSearchQuery_phonePrefixOnly(t *testing.T) {
	q := buildQuickSearchQuery("98900123456", []any{"customer"})
	if strings.Contains(q.where, " OR ") {
		t.Fatal("phone-only search must not use OR")
	}
	if !strings.Contains(q.where, phoneNormExpr) {
		t.Fatal("expected phone expression")
	}
	if q.args[1] != "98900123456%" {
		t.Fatalf("phone pattern = %v", q.args[1])
	}
}

func TestBuildQuickSearchQuery_unionJoin(t *testing.T) {
	q := buildQuickSearchQuery("ali98900", []any{"customer"})
	if !strings.Contains(q.from, "UNION") {
		t.Fatal("mixed input should use UNION join")
	}
	if q.where != "" {
		t.Fatal("union path should not add extra where")
	}
	if len(q.args) != 3 {
		t.Fatalf("args = %v", q.args)
	}
}

func TestBuildAdvancedQuery_usesProfileCTE(t *testing.T) {
	vip := true
	q := buildAdvancedQuery(domain.ListFilter{VIP: &vip}, []any{"customer"})
	if !strings.Contains(q.cte, "profile_filter") {
		t.Fatal("VIP filter should use profile CTE")
	}
	if !strings.Contains(q.cte, "is_vip") {
		t.Fatal("expected is_vip in CTE")
	}
}

func TestBuildAdvancedQuery_segmentUsesStoredColumn(t *testing.T) {
	q := buildAdvancedQuery(domain.ListFilter{Segment: "returning"}, []any{"customer"})
	if !strings.Contains(q.cte, "profile_filter") {
		t.Fatal("segment filter should use profile CTE")
	}
	if !strings.Contains(q.cte, "segment = $") {
		t.Fatal("segment should filter on stored segment column")
	}
	if strings.Contains(q.cte, "CASE") {
		t.Fatal("segment should not use CASE expression")
	}
}

func TestBuildAdvancedQuery_userOnlyEmail(t *testing.T) {
	q := buildAdvancedQuery(domain.ListFilter{Email: "sara@example.com"}, []any{"customer"})
	if q.cte != "" {
		t.Fatal("email-only should not use profile CTE")
	}
	if !strings.Contains(q.where, "lower(u.email)") {
		t.Fatal("expected email condition")
	}
}

func TestBuildAdvancedQuery_crmFiltersUseIndexedProfileColumns(t *testing.T) {
	q := buildAdvancedQuery(domain.ListFilter{
		CustomerAgeRange: "21-40",
		Gender:           "female",
		CustomerTypes:    []string{"vip", "public"},
	}, []any{"customer"})
	if !strings.Contains(q.cte, "profile_filter") {
		t.Fatal("CRM filters should use profile CTE")
	}
	for _, want := range []string{"crm_age_range = $", "crm_gender = $", "crm_customer_type = ANY($"} {
		if !strings.Contains(q.cte, want) {
			t.Fatalf("expected %q in query: %s", want, q.cte)
		}
	}
}

func TestBuildAdvancedQuery_purchaseTypesUseGINJsonOperator(t *testing.T) {
	q := buildAdvancedQuery(domain.ListFilter{
		PurchaseTypes: []string{"gold_and_stones", "silver_and_stones"},
	}, []any{"customer"})
	if !strings.Contains(q.cte, "import_profile->'purchasedCategories') ?| $") {
		t.Fatalf("expected indexed JSONB category overlap condition: %s", q.cte)
	}
}

func TestBuildListQuery_advancedOverridesQuick(t *testing.T) {
	filter := domain.ListFilter{
		QuickSearch: "should-ignore",
		Segment:     "vip",
	}
	q := buildListQuery(filter, []any{"customer"})
	if q.cte == "" && q.where == "" {
		t.Fatal("expected advanced filter clause")
	}
	if strings.Contains(q.cte, "should-ignore") || strings.Contains(q.where, "should-ignore") {
		t.Fatal("quick search should be ignored when advanced filters set")
	}
}

func TestHasAdvancedFilters(t *testing.T) {
	if (domain.ListFilter{QuickSearch: "ali"}).HasAdvancedFilters() {
		t.Fatal("q alone is not advanced")
	}
	if !(domain.ListFilter{Segment: "vip"}).HasAdvancedFilters() {
		t.Fatal("segment is advanced")
	}
	if !(domain.ListFilter{CustomerID: ptrID()}).HasAdvancedFilters() {
		t.Fatal("id is advanced")
	}
}

func ptrID() *shared.ID {
	id := shared.ID(uuid.New())
	return &id
}
