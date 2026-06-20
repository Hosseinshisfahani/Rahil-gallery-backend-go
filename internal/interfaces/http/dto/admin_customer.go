package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	appcustomer "github.com/rahil-gallery/rahil-gallery-server/internal/application/customer"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/customer"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

type PaginatedCustomersResponse struct {
	Data []CustomerSummaryResponse `json:"data"`
	Meta PaginationMeta            `json:"meta"`
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"perPage"`
	Total      *int  `json:"total,omitempty"`
	TotalPages *int  `json:"totalPages,omitempty"`
	HasMore    *bool `json:"hasMore,omitempty"`
}

type CustomerSummaryResponse struct {
	ID               string   `json:"id"`
	FullName         string   `json:"fullName"`
	Phone            string   `json:"phone"`
	RegisteredAt     string   `json:"registeredAt"`
	LastActivityAt   string   `json:"lastActivityAt"`
	LastPurchaseDate *string  `json:"lastPurchaseDate,omitempty"`
	TotalOrders      int      `json:"totalOrders"`
	TotalLtv         float64  `json:"totalLtv"`
	Segment          string   `json:"segment"`
	Status           string   `json:"status"`
	IsVip               bool     `json:"isVip"`
	Tags                []string `json:"tags"`
	CustomerType        string   `json:"customerType"`
	PurchasedCategories []string `json:"purchasedCategories"`
	CustomerAgeRange    *string  `json:"customerAgeRange,omitempty"`
	Gender              *string  `json:"gender,omitempty"`
	Country             string   `json:"country"`
	Href             string   `json:"href"`
}

type CustomerDetailResponse struct {
	CustomerSummaryResponse
	Email                    *string                  `json:"email,omitempty"`
	Locale                   string                   `json:"locale"`
	DefaultRingSize          *string                  `json:"defaultRingSize,omitempty"`
	AverageOrderValue        float64                  `json:"averageOrderValue"`
	FirstPurchaseDate        *string                  `json:"firstPurchaseDate,omitempty"`
	VipSource                *string                  `json:"vipSource"`
	TopCategories            []CategoryInsightResponse `json:"topCategories"`
	WishlistCount            int                      `json:"wishlistCount"`
	WishlistAdditions        int                      `json:"wishlistAdditions"`
	WishlistRemovals         int                      `json:"wishlistRemovals"`
	WishlistConversionRate   float64                  `json:"wishlistConversionRate"`
	CartAbandonmentCount     int                      `json:"cartAbandonmentCount"`
	ConfiguratorUsageCount   int                      `json:"configuratorUsageCount"`
	EngagementScore          int                      `json:"engagementScore"`
	RepeatPurchaseRate       float64                  `json:"repeatPurchaseRate"`
	PurchaseFrequency        float64                  `json:"purchaseFrequency"`
	FunnelPosition           string                   `json:"funnelPosition"`
	BlockReason              *string                  `json:"blockReason,omitempty"`
	BlockNote                *string                  `json:"blockNote,omitempty"`
	Orders                   []CustomerOrderResponse  `json:"orders"`
	Wishlist                 []WishlistItemResponse   `json:"wishlist"`
	Notes                    []CustomerNoteResponse   `json:"notes"`
	AuditLog                 []AuditLogEntryResponse  `json:"auditLog"`
	ImportMode               *string                  `json:"importMode,omitempty"`
	ImportProfile            json.RawMessage          `json:"importProfile,omitempty"`
}

type CategoryInsightResponse struct {
	Category   string  `json:"category"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type CustomerOrderResponse struct {
	ID        string  `json:"id"`
	Date      string  `json:"date"`
	Total     float64 `json:"total"`
	Status    string  `json:"status"`
	ItemCount int     `json:"itemCount"`
	HasReturn bool    `json:"hasReturn"`
	Href      string  `json:"href"`
}

type WishlistItemResponse struct {
	ID                   string  `json:"id"`
	ProductName          string  `json:"productName"`
	Category             string  `json:"category"`
	Price                float64 `json:"price"`
	SavedAt              string  `json:"savedAt"`
	IsConfiguration      bool    `json:"isConfiguration"`
	ConfigurationSummary *string `json:"configurationSummary,omitempty"`
}

type CustomerNoteResponse struct {
	ID        string `json:"id"`
	Author    string `json:"author"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
}

type AuditLogEntryResponse struct {
	ID           string  `json:"id"`
	AdminID      string  `json:"adminId"`
	AdminName    string  `json:"adminName"`
	Action       string  `json:"action"`
	TargetUserID string  `json:"targetUserId"`
	Timestamp    string  `json:"timestamp"`
	Reason       *string `json:"reason,omitempty"`
	Details      *string `json:"details,omitempty"`
}

type CreateCustomerRequest struct {
	ImportMode      string                         `json:"importMode"`
	FullName        string                         `json:"fullName"`
	Phone           string                         `json:"phone"`
	Email           string                         `json:"email"`
	Locale          string                         `json:"locale"`
	DefaultRingSize string                         `json:"defaultRingSize"`
	IsVip           bool                           `json:"isVip"`
	Tags            []string                       `json:"tags"`
	ImportProfile   *domain.ImportProfileInput     `json:"importProfile"`
}

type UpdateCustomerRequest struct {
	FullName        *string                    `json:"fullName"`
	Phone           *string                    `json:"phone"`
	Email           *string                    `json:"email"`
	Locale          *string                    `json:"locale"`
	DefaultRingSize *string                    `json:"defaultRingSize"`
	IsVip           *bool                      `json:"isVip"`
	Tags            *[]string                  `json:"tags"`
	ImportProfile   *domain.ImportProfileInput `json:"importProfile"`
}

type BlockCustomerRequest struct {
	Reason string `json:"reason"`
	Note   string `json:"note"`
}

type UnblockCustomerRequest struct {
	Justification string `json:"justification"`
}

type ToggleTagRequest struct {
	Tag string `json:"tag"`
}

type AddNoteRequest struct {
	Body string `json:"body"`
}

func ToCustomerSummary(row domain.ListRow) CustomerSummaryResponse {
	domain.NormalizeListRowCRM(&row)
	id := row.ID.String()
	resp := CustomerSummaryResponse{
		ID:                  id,
		FullName:            row.FullName,
		Phone:               row.Phone,
		RegisteredAt:        formatDate(row.RegisteredAt),
		LastActivityAt:      formatDatePtr(row.LastActivityAt, row.RegisteredAt),
		TotalOrders:         row.TotalOrders,
		TotalLtv:            row.TotalLTV,
		Segment:             string(row.Segment),
		Status:              row.Status,
		IsVip:               row.IsVIP,
		Tags:                row.Tags,
		CustomerType:        *row.CustomerType,
		PurchasedCategories: row.PurchasedCategories,
		CustomerAgeRange:    row.CustomerAgeRange,
		Gender:              row.Gender,
		Country:             "IR",
		Href:                "/admin/customers/" + id,
	}
	if row.LastPurchaseDate != nil {
		d := formatDate(*row.LastPurchaseDate)
		resp.LastPurchaseDate = &d
	}
	if resp.Tags == nil {
		resp.Tags = []string{}
	}
	if resp.PurchasedCategories == nil {
		resp.PurchasedCategories = append([]string(nil), domain.DefaultPurchasedCategories...)
	}
	return resp
}

func ToCustomerDetail(d *domain.Detail) CustomerDetailResponse {
	enrichListRowCRMFromImportProfile(&d.ListRow, d.ImportProfile)
	summary := ToCustomerSummary(d.ListRow)
	detail := CustomerDetailResponse{
		CustomerSummaryResponse: summary,
		Email:                   d.Email,
		Locale:                  d.Locale,
		DefaultRingSize:         d.DefaultRingSize,
		AverageOrderValue:       d.AverageOrderValue,
		VipSource:               vipSourceString(d.VIPSource),
		TopCategories:           toCategoryInsights(d.TopCategories),
		WishlistCount:           d.WishlistCount,
		WishlistAdditions:       d.WishlistAdditions,
		WishlistRemovals:        d.WishlistRemovals,
		WishlistConversionRate:  d.WishlistConversionRate,
		CartAbandonmentCount:    d.CartAbandonmentCount,
		ConfiguratorUsageCount:  d.ConfiguratorUsageCount,
		EngagementScore:         d.EngagementScore,
		RepeatPurchaseRate:      d.RepeatPurchaseRate,
		PurchaseFrequency:       d.PurchaseFrequency,
		FunnelPosition:          d.FunnelPosition,
		Orders:                  toOrders(d.Orders),
		Wishlist:                toWishlist(d.Wishlist),
		Notes:                   toNotes(d.Notes),
		AuditLog:                toAuditLog(d.AuditLog),
	}
	if d.FirstPurchaseDate != nil {
		fd := formatDate(*d.FirstPurchaseDate)
		detail.FirstPurchaseDate = &fd
	}
	if d.BlockReason != nil {
		br := string(*d.BlockReason)
		detail.BlockReason = &br
	}
	detail.BlockNote = d.BlockNote
	if d.ImportMode != nil {
		im := string(*d.ImportMode)
		detail.ImportMode = &im
	}
	if len(d.ImportProfile) > 0 {
		detail.ImportProfile = d.ImportProfile
	}
	return detail
}

func enrichListRowCRMFromImportProfile(row *domain.ListRow, importProfile json.RawMessage) {
	if row == nil || len(importProfile) == 0 {
		return
	}
	var profile struct {
		CustomerType        string   `json:"customerType"`
		CustomerAgeRange    string   `json:"customerAgeRange"`
		Gender              string   `json:"gender"`
		PurchasedCategories []string `json:"purchasedCategories"`
	}
	if err := json.Unmarshal(importProfile, &profile); err != nil {
		return
	}
	if (row.CustomerType == nil || *row.CustomerType == "") && profile.CustomerType != "" {
		ct := profile.CustomerType
		row.CustomerType = &ct
	}
	if row.CustomerAgeRange == nil && profile.CustomerAgeRange != "" {
		ar := profile.CustomerAgeRange
		row.CustomerAgeRange = &ar
	}
	if row.Gender == nil && profile.Gender != "" {
		g := profile.Gender
		row.Gender = &g
	}
	if len(row.PurchasedCategories) == 0 && len(profile.PurchasedCategories) > 0 {
		row.PurchasedCategories = profile.PurchasedCategories
	}
}

func ToPaginatedCustomers(result domain.ListResult) PaginatedCustomersResponse {
	items := make([]CustomerSummaryResponse, 0, len(result.Items))
	for _, row := range result.Items {
		items = append(items, ToCustomerSummary(row))
	}
	meta := PaginationMeta{
		Page:    result.Page,
		PerPage: result.PerPage,
	}
	if result.TotalExact {
		meta.Total = &result.Total
		meta.TotalPages = &result.TotalPages
	} else {
		hasMore := result.HasMore
		meta.HasMore = &hasMore
	}
	return PaginatedCustomersResponse{
		Data: items,
		Meta: meta,
	}
}

func (r CreateCustomerRequest) ToInput() appcustomer.CreateInput {
	return appcustomer.CreateInput{
		ImportMode:      r.ImportMode,
		FullName:        r.FullName,
		Phone:           r.Phone,
		Email:           r.Email,
		Locale:          r.Locale,
		DefaultRingSize: r.DefaultRingSize,
		IsVIP:           r.IsVip,
		Tags:            r.Tags,
		ImportProfile:   r.ImportProfile,
	}
}

func (r UpdateCustomerRequest) ToInput() appcustomer.UpdateInput {
	return appcustomer.UpdateInput{
		FullName:        r.FullName,
		Phone:           r.Phone,
		Email:           r.Email,
		Locale:          r.Locale,
		DefaultRingSize: r.DefaultRingSize,
		IsVIP:           r.IsVip,
		Tags:            r.Tags,
		ImportProfile:   r.ImportProfile,
	}
}

func vipSourceString(v *domain.VIPSource) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
}

func toCategoryInsights(items []domain.CategoryInsight) []CategoryInsightResponse {
	out := make([]CategoryInsightResponse, 0, len(items))
	for _, item := range items {
		out = append(out, CategoryInsightResponse(item))
	}
	return out
}

func toOrders(orders []domain.OrderSummary) []CustomerOrderResponse {
	out := make([]CustomerOrderResponse, 0, len(orders))
	for _, o := range orders {
		id := o.ID.String()
		out = append(out, CustomerOrderResponse{
			ID:        id,
			Date:      formatDate(o.Date),
			Total:     o.Total,
			Status:    o.Status,
			ItemCount: o.ItemCount,
			HasReturn: o.HasReturn,
			Href:      "/admin/orders/" + id,
		})
	}
	return out
}

func toWishlist(items []domain.WishlistSummary) []WishlistItemResponse {
	out := make([]WishlistItemResponse, 0, len(items))
	for _, w := range items {
		out = append(out, WishlistItemResponse{
			ID:                   w.ID.String(),
			ProductName:          w.ProductName,
			Category:             w.Category,
			Price:                w.Price,
			SavedAt:              formatDate(w.SavedAt),
			IsConfiguration:      w.IsConfiguration,
			ConfigurationSummary: w.ConfigurationSummary,
		})
	}
	return out
}

func toNotes(notes []domain.Note) []CustomerNoteResponse {
	out := make([]CustomerNoteResponse, 0, len(notes))
	for _, n := range notes {
		out = append(out, CustomerNoteResponse{
			ID:        n.ID.String(),
			Author:    n.Author,
			Body:      n.Body,
			CreatedAt: n.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return out
}

func toAuditLog(entries []domain.AuditEntry) []AuditLogEntryResponse {
	out := make([]AuditLogEntryResponse, 0, len(entries))
	for _, e := range entries {
		out = append(out, AuditLogEntryResponse{
			ID:           e.ID.String(),
			AdminID:      e.AdminID.String(),
			AdminName:    e.AdminName,
			Action:       string(e.Action),
			TargetUserID: e.TargetUserID.String(),
			Timestamp:    e.CreatedAt.UTC().Format(time.RFC3339),
			Reason:       e.Reason,
			Details:      e.Details,
		})
	}
	return out
}

func formatDate(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

func formatDatePtr(t *time.Time, fallback time.Time) string {
	if t != nil {
		return formatDate(*t)
	}
	return formatDate(fallback)
}

func ParseCustomerID(id string) (shared.ID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return shared.ID{}, shared.ErrInvalidInput
	}
	return shared.ID(parsed), nil
}
