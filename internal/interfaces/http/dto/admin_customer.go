package dto

import (
	"strings"
	"time"

	"github.com/google/uuid"
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
	ID                  string   `json:"id"`
	FullName            string   `json:"fullName"`
	Phone               string   `json:"phone"`
	CustomerType        string   `json:"customerType"`
	PurchasedCategories []string `json:"purchasedCategories"`
	CustomerAgeRange    *string  `json:"customerAgeRange,omitempty"`
	Gender              *string  `json:"gender,omitempty"`
	CreatedAt           string   `json:"createdAt"`
	Href                string   `json:"href"`
}

type ImportProfileResponse struct {
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
	Gender              *string  `json:"gender,omitempty"`
	CustomerType        string   `json:"customerType"`
	CustomerAgeRange    *string  `json:"customerAgeRange,omitempty"`
	PurchasedCategories []string `json:"purchasedCategories"`
	Description         *string  `json:"description,omitempty"`
	Signature           *string  `json:"signature,omitempty"`
}

type CustomerDetailResponse struct {
	CustomerSummaryResponse
	Email         *string                 `json:"email,omitempty"`
	Job           *string                 `json:"job,omitempty"`
	Address       *string                 `json:"address,omitempty"`
	Birthday      *string                 `json:"birthday,omitempty"`
	MarriageDate  *string                 `json:"marriageDate,omitempty"`
	ImportantDate *string                 `json:"importantDate,omitempty"`
	FirstVisitDate *string                `json:"firstVisitDate,omitempty"`
	Description   *string                 `json:"description,omitempty"`
	SignatureUrl  *string                 `json:"signatureUrl,omitempty"`
	ImportProfile ImportProfileResponse     `json:"importProfile"`
}

type CreateCustomerRequest struct {
	ImportMode    string                   `json:"importMode"`
	ImportProfile *ImportProfileRequest    `json:"importProfile"`
}

type UpdateCustomerRequest struct {
	ImportProfile *ImportProfileRequest `json:"importProfile"`
}

type ImportProfileRequest struct {
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
	Gender              *string  `json:"gender,omitempty"`
	CustomerType        string   `json:"customerType"`
	CustomerAgeRange    *string  `json:"customerAgeRange,omitempty"`
	PurchasedCategories []string `json:"purchasedCategories"`
	Description         *string  `json:"description,omitempty"`
	Signature           *string  `json:"signature,omitempty"`
}

func ToCustomerSummary(row domain.ListRow) CustomerSummaryResponse {
	domain.NormalizeListRow(&row)
	id := row.ID.String()
	return CustomerSummaryResponse{
		ID:                  id,
		FullName:            row.FullName,
		Phone:               row.Phone,
		CustomerType:        row.CustomerType,
		PurchasedCategories: domain.StringSliceOrEmpty(row.PurchasedCategories),
		CustomerAgeRange:    row.CustomerAgeRange,
		Gender:              row.Gender,
		CreatedAt:           formatDate(row.CreatedAt),
		Href:                "/admin/customers/" + id,
	}
}

func ToCustomerDetail(c *domain.Customer) CustomerDetailResponse {
	summary := ToCustomerSummary(domain.ListRow{
		ID:                  c.ID,
		FullName:            c.FullName(),
		Phone:               c.Phone,
		CustomerType:        c.CustomerType,
		PurchasedCategories: c.PurchasedCategories,
		CustomerAgeRange:    c.CustomerAgeRange,
		Gender:              c.Gender,
		CreatedAt:           c.CreatedAt,
	})

	profile := toImportProfileResponse(c)

	return CustomerDetailResponse{
		CustomerSummaryResponse: summary,
		Email:                   c.Email,
		Job:                     c.Job,
		Address:                 c.Address,
		Birthday:                domain.FormatDate(c.Birthday),
		MarriageDate:            domain.FormatDate(c.MarriageDate),
		ImportantDate:           domain.FormatDate(c.ImportantDate),
		FirstVisitDate:          domain.FormatDate(c.FirstVisitDate),
		Description:             c.Description,
		SignatureUrl:            c.SignatureURL,
		ImportProfile:           profile,
	}
}

func toImportProfileResponse(c *domain.Customer) ImportProfileResponse {
	sig := c.SignatureURL
	return ImportProfileResponse{
		FirstName:           c.FirstName,
		LastName:            c.LastName,
		Job:                 c.Job,
		Phone:               c.Phone,
		Email:               c.Email,
		Address:             c.Address,
		Birthday:            domain.FormatDate(c.Birthday),
		MarriageDate:        domain.FormatDate(c.MarriageDate),
		ImportantDate:       domain.FormatDate(c.ImportantDate),
		FirstVisitDate:      domain.FormatDate(c.FirstVisitDate),
		Gender:              c.Gender,
		CustomerType:        c.CustomerType,
		CustomerAgeRange:    c.CustomerAgeRange,
		PurchasedCategories: domain.StringSliceOrEmpty(c.PurchasedCategories),
		Description:         c.Description,
		Signature:           sig,
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

func (r CreateCustomerRequest) ToInput() (domain.Input, error) {
	if r.ImportProfile == nil {
		return domain.Input{}, shared.ErrInvalidInput
	}
	return importProfileToInput(*r.ImportProfile)
}

func (r UpdateCustomerRequest) ToInput() (domain.Input, error) {
	if r.ImportProfile == nil {
		return domain.Input{}, shared.ErrInvalidInput
	}
	return importProfileToInput(*r.ImportProfile)
}

func importProfileToInput(p ImportProfileRequest) (domain.Input, error) {
	birthday, err := domain.ParseDate(p.Birthday)
	if err != nil {
		return domain.Input{}, shared.ErrInvalidInput
	}
	marriage, err := domain.ParseDate(p.MarriageDate)
	if err != nil {
		return domain.Input{}, shared.ErrInvalidInput
	}
	important, err := domain.ParseDate(p.ImportantDate)
	if err != nil {
		return domain.Input{}, shared.ErrInvalidInput
	}
	firstVisit, err := domain.ParseDate(p.FirstVisitDate)
	if err != nil {
		return domain.Input{}, shared.ErrInvalidInput
	}

	return domain.Input{
		FirstName:           strings.TrimSpace(p.FirstName),
		LastName:            strings.TrimSpace(p.LastName),
		Job:                 p.Job,
		Phone:               strings.TrimSpace(p.Phone),
		Email:               p.Email,
		Address:             p.Address,
		Birthday:            birthday,
		MarriageDate:        marriage,
		ImportantDate:       important,
		FirstVisitDate:      firstVisit,
		Gender:              p.Gender,
		CustomerType:        p.CustomerType,
		CustomerAgeRange:    p.CustomerAgeRange,
		PurchasedCategories: domain.StringSliceOrEmpty(p.PurchasedCategories),
		Description:         p.Description,
		SignatureURL:        p.Signature,
	}, nil
}

func formatDate(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

func ParseCustomerID(id string) (shared.ID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return shared.ID{}, shared.ErrInvalidInput
	}
	return shared.ID(parsed), nil
}
