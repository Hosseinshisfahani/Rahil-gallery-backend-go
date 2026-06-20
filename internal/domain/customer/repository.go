package customer

import (
	"context"
	"encoding/json"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type Repository interface {
	List(ctx context.Context, filter ListFilter, page, perPage int) (ListResult, error)
	GetDetail(ctx context.Context, userID shared.ID) (*Detail, error)
	PhoneExists(ctx context.Context, phone string, excludeUserID *shared.ID) (bool, error)

	UpsertProfile(ctx context.Context, profile *Profile) error
	GetProfile(ctx context.Context, userID shared.ID) (*Profile, error)

	AddNote(ctx context.Context, note *Note) error
	ListNotes(ctx context.Context, userID shared.ID) ([]Note, error)

	AppendAudit(ctx context.Context, entry *AuditEntry) error
	ListAudit(ctx context.Context, userID shared.ID) ([]AuditEntry, error)

	ListOrders(ctx context.Context, userID shared.ID) ([]OrderSummary, error)
	ListWishlist(ctx context.Context, userID shared.ID) ([]WishlistSummary, error)

	ListSavedViews(ctx context.Context, ownerID shared.ID, viewType *ListViewType) ([]SavedListView, error)
	GetSavedView(ctx context.Context, id, requesterID shared.ID) (*SavedListView, error)
	CreateSavedView(ctx context.Context, view *SavedListView) error
	UpdateSavedView(ctx context.Context, view *SavedListView) error
	DeleteSavedView(ctx context.Context, id, ownerID shared.ID) error
	CountCustomersBySegment(ctx context.Context) ([]SegmentSummary, error)
}

type ImportProfileInput struct {
	FirstName           string   `json:"firstName"`
	LastName            string   `json:"lastName"`
	Job                 *string  `json:"job,omitempty"`
	Phone               string   `json:"phone"`
	Email               *string  `json:"email,omitempty"`
	Address             *string  `json:"address,omitempty"`
	Birthday            *string  `json:"birthday,omitempty"`
	MarriageDate        *string  `json:"marriageDate,omitempty"`
	ImportantDate       *string  `json:"importantDate,omitempty"`
	FirstVisitDate      *string  `json:"firstVisitDate,omitempty"`
	CustomerType        string   `json:"customerType"`
	CustomerAgeRange    *string  `json:"customerAgeRange,omitempty"`
	Gender              *string  `json:"gender,omitempty"`
	PurchasedCategories []string `json:"purchasedCategories"`
	Description         *string  `json:"description,omitempty"`
	Signature           *string  `json:"signature,omitempty"`
}

func MarshalImportProfile(v ImportProfileInput) (json.RawMessage, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(b), nil
}
