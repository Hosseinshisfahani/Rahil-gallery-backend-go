package customer

import "testing"

func TestSavedFilterPayload_ToListFilter(t *testing.T) {
	from := "2024-01-01"
	p := SavedFilterPayload{
		Segment:        "vip",
		LTVMin:         ptrFloat(1_000_000),
		RegisteredFrom: &from,
	}
	f, err := p.ToListFilter()
	if err != nil {
		t.Fatal(err)
	}
	if f.Segment != "vip" || f.LTVMin == nil || *f.LTVMin != 1_000_000 {
		t.Fatalf("unexpected filter %+v", f)
	}
	if f.RegisteredFrom == nil {
		t.Fatal("expected registeredFrom")
	}
}

func TestSavedFilterPayload_invalidDate(t *testing.T) {
	bad := "not-a-date"
	_, err := (SavedFilterPayload{RegisteredFrom: &bad}).ToListFilter()
	if err == nil {
		t.Fatal("expected date parse error")
	}
}

func ptrFloat(v float64) *float64 { return &v }
