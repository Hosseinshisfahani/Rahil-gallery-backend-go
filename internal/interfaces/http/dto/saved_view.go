package dto

import (
	"time"

	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/customer"
)

type SegmentsResponse struct {
	Builtin []SegmentSummaryResponse `json:"builtin"`
	Saved   []SavedListViewResponse  `json:"saved"`
}

type SegmentSummaryResponse struct {
	Segment string `json:"segment"`
	Label   string `json:"label"`
	Count   int    `json:"count"`
}

type SavedListViewResponse struct {
	ID        string                    `json:"id"`
	Name      string                    `json:"name"`
	ViewType  string                    `json:"viewType"`
	Filters   domain.SavedFilterPayload `json:"filters"`
	IsShared  bool                      `json:"isShared"`
	Position  int                       `json:"position"`
	OwnerID   string                    `json:"ownerId"`
	CreatedAt string                    `json:"createdAt"`
	UpdatedAt string                    `json:"updatedAt"`
}

type CreateSavedViewRequest struct {
	Name     string                    `json:"name"`
	ViewType string                    `json:"viewType"`
	Filters  domain.SavedFilterPayload `json:"filters"`
	IsShared bool                      `json:"isShared"`
	Position int                       `json:"position"`
}

type UpdateSavedViewRequest struct {
	Name     string                    `json:"name"`
	ViewType string                    `json:"viewType"`
	Filters  domain.SavedFilterPayload `json:"filters"`
	IsShared bool                      `json:"isShared"`
	Position int                       `json:"position"`
}

func ToSegmentsResponse(r domain.SegmentsResult) SegmentsResponse {
	builtin := make([]SegmentSummaryResponse, 0, len(r.Builtin))
	for _, s := range r.Builtin {
		builtin = append(builtin, SegmentSummaryResponse{
			Segment: string(s.Segment),
			Label:   s.Label,
			Count:   s.Count,
		})
	}
	saved := make([]SavedListViewResponse, 0, len(r.Saved))
	for _, v := range r.Saved {
		saved = append(saved, ToSavedListView(v))
	}
	return SegmentsResponse{Builtin: builtin, Saved: saved}
}

func ToSavedListView(v domain.SavedListView) SavedListViewResponse {
	return SavedListViewResponse{
		ID:        v.ID.String(),
		Name:      v.Name,
		ViewType:  string(v.ViewType),
		Filters:   v.Filters,
		IsShared:  v.IsShared,
		Position:  v.Position,
		OwnerID:   v.OwnerID.String(),
		CreatedAt: v.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: v.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func (r CreateSavedViewRequest) ToInput() (domain.ListViewType, domain.SavedFilterPayload, error) {
	return domain.ListViewType(r.ViewType), r.Filters, nil
}
