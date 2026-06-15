package catalog

import (
	"strings"
	"testing"

	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/catalog"
)

func TestBuildPublishedListQueryIncludesPublishedFilter(t *testing.T) {
	q := buildPublishedListQuery(domain.PublicListFilter{})
	if !strings.Contains(q.where, "status = 'published'") {
		t.Fatalf("expected published filter, got %q", q.where)
	}
}

func TestBuildAdminListQuerySearch(t *testing.T) {
	q := buildAdminListQuery(domain.AdminListFilter{Query: "ring"})
	if len(q.args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(q.args))
	}
	if !strings.Contains(q.where, "ILIKE") {
		t.Fatalf("expected ILIKE clause, got %q", q.where)
	}
}

func TestPublishedSort(t *testing.T) {
	if publishedSort("price_asc") != "COALESCE(vc.price_from, p.base_price) ASC" {
		t.Fatal("unexpected price_asc sort")
	}
}
