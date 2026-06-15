package customer

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type Segment string

const (
	SegmentNew       Segment = "new"
	SegmentActive    Segment = "active"
	SegmentReturning Segment = "returning"
	SegmentVIP       Segment = "vip"
	SegmentInactive  Segment = "inactive"
)

type ImportMode string

const (
	ImportModeQuick           ImportMode = "quick"
	ImportModeHistoryIncluded ImportMode = "history_included"
)

type VIPSource string

const (
	VIPSourceManual    VIPSource = "manual"
	VIPSourceAutomatic VIPSource = "automatic"
)

type BlockReason string

const (
	BlockReasonFraudSuspicion BlockReason = "fraud_suspicion"
	BlockReasonPaymentIssues  BlockReason = "payment_issues"
	BlockReasonReturnAbuse    BlockReason = "return_abuse"
	BlockReasonSystemMisuse   BlockReason = "system_misuse"
)

type AuditAction string

const (
	AuditActionBlock       AuditAction = "block"
	AuditActionUnblock     AuditAction = "unblock"
	AuditActionVIPAssign   AuditAction = "vip_assign"
	AuditActionVIPRemove   AuditAction = "vip_remove"
	AuditActionTagAdd      AuditAction = "tag_add"
	AuditActionTagRemove   AuditAction = "tag_remove"
	AuditActionProfileEdit AuditAction = "profile_edit"
	AuditActionNoteAdd     AuditAction = "note_add"
	AuditActionExport      AuditAction = "export"
	AuditActionCreated     AuditAction = "account_created"
)

type Profile struct {
	UserID          shared.ID
	Locale          string
	DefaultRingSize *string
	IsVIP           bool
	VIPSource       *VIPSource
	ImportMode      *ImportMode
	ImportProfile   json.RawMessage
	Tags            []string
	BlockReason     *BlockReason
	BlockNote       *string
	LastActivityAt  *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Note struct {
	ID        shared.ID
	UserID    shared.ID
	AuthorID  shared.ID
	Author    string
	Body      string
	CreatedAt time.Time
}

type AuditEntry struct {
	ID           shared.ID
	AdminID      shared.ID
	AdminName    string
	TargetUserID shared.ID
	Action       AuditAction
	Reason       *string
	Details      *string
	CreatedAt    time.Time
}

type OrderSummary struct {
	ID        shared.ID
	Date      time.Time
	Total     float64
	Status    string
	ItemCount int
	HasReturn bool
}

type WishlistSummary struct {
	ID                   shared.ID
	ProductName          string
	Category             string
	Price                float64
	SavedAt              time.Time
	IsConfiguration      bool
	ConfigurationSummary *string
}

type CategoryInsight struct {
	Category   string
	Count      int
	Percentage float64
}

type ListRow struct {
	ID               shared.ID
	FullName         string
	Phone            string
	Email            *string
	RegisteredAt     time.Time
	LastActivityAt   *time.Time
	LastPurchaseDate *time.Time
	TotalOrders      int
	TotalLTV         float64
	Segment          Segment
	Status           string
	IsVIP            bool
	Tags             []string
}

type Detail struct {
	ListRow
	Locale                 string
	DefaultRingSize        *string
	AverageOrderValue      float64
	FirstPurchaseDate      *time.Time
	VIPSource              *VIPSource
	TopCategories          []CategoryInsight
	WishlistCount          int
	WishlistAdditions      int
	WishlistRemovals       int
	WishlistConversionRate float64
	CartAbandonmentCount   int
	ConfiguratorUsageCount int
	EngagementScore        int
	RepeatPurchaseRate     float64
	PurchaseFrequency      float64
	FunnelPosition         string
	BlockReason            *BlockReason
	BlockNote              *string
	ImportMode             *ImportMode
	ImportProfile          json.RawMessage
	Orders                 []OrderSummary
	Wishlist               []WishlistSummary
	Notes                  []Note
	AuditLog               []AuditEntry
}

// ListFilter supports two mutually exclusive modes (enforced in HTTP layer):
//   - Quick search: QuickSearch (q) — text searches name; 4+ digits search phone prefix.
//   - Advanced: structured filters below (segment, LTV, dates, email, id, …).
type ListFilter struct {
	QuickSearch string

	CustomerID       *shared.ID
	Email            string
	Segment          string
	Status           string
	VIP              *bool
	LTVMin           *float64
	LTVMax           *float64
	OrdersMin        *int
	OrdersMax        *int
	RegisteredFrom   *time.Time
	RegisteredTo     *time.Time
	LastPurchaseFrom *time.Time
	LastPurchaseTo   *time.Time
	LastActivityFrom *time.Time
	LastActivityTo   *time.Time
	Tags             []string
	HasPurchased     string
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

	// SkipCount skips COUNT(*) and uses LIMIT+1 for hasMore (HTTP: includeTotal=false).
	SkipCount bool
}

func (f ListFilter) HasAdvancedFilters() bool {
	return f.CustomerID != nil ||
		strings.TrimSpace(f.Email) != "" ||
		f.Segment != "" ||
		f.Status != "" ||
		f.VIP != nil ||
		f.LTVMin != nil ||
		f.LTVMax != nil ||
		f.OrdersMin != nil ||
		f.OrdersMax != nil ||
		f.RegisteredFrom != nil ||
		f.RegisteredTo != nil ||
		f.LastPurchaseFrom != nil ||
		f.LastPurchaseTo != nil ||
		f.LastActivityFrom != nil ||
		f.LastActivityTo != nil ||
		len(f.Tags) > 0 ||
		f.HasPurchased != "" ||
		strings.TrimSpace(f.CustomerAgeRange) != "" ||
		strings.TrimSpace(f.Gender) != "" ||
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
	TotalExact bool // false when COUNT was skipped (includeTotal=false)
}
