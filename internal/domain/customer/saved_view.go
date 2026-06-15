package customer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type ListViewType string

const (
	ListViewTypeFilter  ListViewType = "filter"
	ListViewTypeSegment ListViewType = "segment"
)

// SavedFilterPayload mirrors list query params for round-trip with the admin UI.
type SavedFilterPayload struct {
	Q                string   `json:"q,omitempty"`
	Segment          string   `json:"segment,omitempty"`
	Email            string   `json:"email,omitempty"`
	Status           string   `json:"status,omitempty"`
	VIP              *bool    `json:"vip,omitempty"`
	LTVMin           *float64 `json:"ltvMin,omitempty"`
	LTVMax           *float64 `json:"ltvMax,omitempty"`
	OrdersMin        *int     `json:"ordersMin,omitempty"`
	OrdersMax        *int     `json:"ordersMax,omitempty"`
	RegisteredFrom   *string  `json:"registeredFrom,omitempty"`
	RegisteredTo     *string  `json:"registeredTo,omitempty"`
	LastPurchaseFrom *string  `json:"lastPurchaseFrom,omitempty"`
	LastPurchaseTo   *string  `json:"lastPurchaseTo,omitempty"`
	LastActivityFrom *string  `json:"lastActivityFrom,omitempty"`
	LastActivityTo   *string  `json:"lastActivityTo,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	HasPurchased     string   `json:"hasPurchased,omitempty"`
	IncludeTotal     *bool    `json:"includeTotal,omitempty"`
}

func (p SavedFilterPayload) ToListFilter() (ListFilter, error) {
	f := ListFilter{
		QuickSearch:  p.Q,
		Email:        p.Email,
		Segment:      p.Segment,
		Status:       p.Status,
		VIP:          p.VIP,
		LTVMin:       p.LTVMin,
		LTVMax:       p.LTVMax,
		OrdersMin:    p.OrdersMin,
		OrdersMax:    p.OrdersMax,
		Tags:         p.Tags,
		HasPurchased: p.HasPurchased,
	}
	if p.IncludeTotal != nil && !*p.IncludeTotal {
		f.SkipCount = true
	}

	var err error
	if f.RegisteredFrom, err = parseOptionalDate(p.RegisteredFrom); err != nil {
		return ListFilter{}, fmt.Errorf("registeredFrom: %w", err)
	}
	if f.RegisteredTo, err = parseOptionalDate(p.RegisteredTo); err != nil {
		return ListFilter{}, fmt.Errorf("registeredTo: %w", err)
	}
	if f.LastPurchaseFrom, err = parseOptionalDate(p.LastPurchaseFrom); err != nil {
		return ListFilter{}, fmt.Errorf("lastPurchaseFrom: %w", err)
	}
	if f.LastPurchaseTo, err = parseOptionalDate(p.LastPurchaseTo); err != nil {
		return ListFilter{}, fmt.Errorf("lastPurchaseTo: %w", err)
	}
	if f.LastActivityFrom, err = parseOptionalDate(p.LastActivityFrom); err != nil {
		return ListFilter{}, fmt.Errorf("lastActivityFrom: %w", err)
	}
	if f.LastActivityTo, err = parseOptionalDate(p.LastActivityTo); err != nil {
		return ListFilter{}, fmt.Errorf("lastActivityTo: %w", err)
	}
	return f, nil
}

func parseOptionalDate(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func ListFilterToPayload(f ListFilter) SavedFilterPayload {
	p := SavedFilterPayload{
		Q:            f.QuickSearch,
		Email:        f.Email,
		Segment:      f.Segment,
		Status:       f.Status,
		VIP:          f.VIP,
		LTVMin:       f.LTVMin,
		LTVMax:       f.LTVMax,
		OrdersMin:    f.OrdersMin,
		OrdersMax:    f.OrdersMax,
		Tags:         f.Tags,
		HasPurchased: f.HasPurchased,
	}
	p.RegisteredFrom = formatOptionalDate(f.RegisteredFrom)
	p.RegisteredTo = formatOptionalDate(f.RegisteredTo)
	p.LastPurchaseFrom = formatOptionalDate(f.LastPurchaseFrom)
	p.LastPurchaseTo = formatOptionalDate(f.LastPurchaseTo)
	p.LastActivityFrom = formatOptionalDate(f.LastActivityFrom)
	p.LastActivityTo = formatOptionalDate(f.LastActivityTo)
	if f.SkipCount {
		inc := false
		p.IncludeTotal = &inc
	}
	return p
}

func formatOptionalDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

var SegmentLabels = map[Segment]string{
	SegmentVIP:       "VIP",
	SegmentNew:       "New",
	SegmentActive:    "Active",
	SegmentReturning: "Returning",
	SegmentInactive:  "Inactive",
}

type SavedListView struct {
	ID        shared.ID
	OwnerID   shared.ID
	Name      string
	ViewType  ListViewType
	Filters   SavedFilterPayload
	IsShared  bool
	Position  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (v SavedListView) FiltersJSON() (json.RawMessage, error) {
	return json.Marshal(v.Filters)
}

type SegmentSummary struct {
	Segment Segment
	Label   string
	Count   int
}

type SegmentsResult struct {
	Builtin []SegmentSummary
	Saved   []SavedListView
}
