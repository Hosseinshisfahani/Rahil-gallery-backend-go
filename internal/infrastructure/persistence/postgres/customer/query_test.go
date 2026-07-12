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

func TestBuildQuickSearchQuery_nameUsesLike(t *testing.T) {
	q := buildQuickSearchQuery("ali", nil)
	if q.where == "" {
		t.Fatal("expected where clause")
	}
	if !strings.Contains(q.where, nameSearchExpr) {
		t.Fatal("expected name expression")
	}
}

func TestBuildAdvancedQuery_crmFilters(t *testing.T) {
	q := buildAdvancedQuery(domain.ListFilter{
		CustomerAgeRange: "21-40",
		Gender:           "female",
		CustomerTypes:    []string{"vip", "public"},
	}, nil)
	for _, want := range []string{"customer_age_range = $", "gender = $", "customer_type = ANY($"} {
		if !strings.Contains(q.where, want) {
			t.Fatalf("expected %q in query: %s", want, q.where)
		}
	}
}

func TestBuildListQuery_advancedOverridesQuick(t *testing.T) {
	filter := domain.ListFilter{
		QuickSearch:      "should-ignore",
		CustomerAgeRange: "21-40",
	}
	q := buildListQuery(filter, nil)
	if strings.Contains(q.where, "should-ignore") {
		t.Fatal("quick search should be ignored when advanced filters set")
	}
}

func TestHasAdvancedFilters(t *testing.T) {
	if (domain.ListFilter{QuickSearch: "ali"}).HasAdvancedFilters() {
		t.Fatal("q alone is not advanced")
	}
	if !(domain.ListFilter{CustomerID: ptrID()}).HasAdvancedFilters() {
		t.Fatal("id is advanced")
	}
}

func ptrID() *shared.ID {
	id := shared.ID(uuid.New())
	return &id
}
