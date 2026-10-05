package handler

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rahil-gallery/rahil-gallery-server/internal/model"
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
	MelliCode           *string  `json:"melliCode,omitempty"`
	PostalCode          *string  `json:"postalCode,omitempty"`
	Birthday            *string  `json:"birthday,omitempty"`
	MarriageDate        *string  `json:"marriageDate,omitempty"`
	ImportantDate       *string  `json:"importantDate,omitempty"`
	FirstVisitDate      *string  `json:"firstVisitDate,omitempty"`
	Gender              *string  `json:"gender,omitempty"`
	CustomerType        string   `json:"customerType"`
	CustomerAgeRange    *string  `json:"customerAgeRange,omitempty"`
	PurchasedCategories []string `json:"purchasedCategories"`
	Description         *string  `json:"description,omitempty"`
	MarketerNote        *string  `json:"marketerNote,omitempty"`
	Signature           *string  `json:"signature,omitempty"`
}

type CustomerDetailResponse struct {
	CustomerSummaryResponse
	Email          *string               `json:"email,omitempty"`
	Job            *string               `json:"job,omitempty"`
	Address        *string               `json:"address,omitempty"`
	MelliCode      *string               `json:"melliCode,omitempty"`
	PostalCode     *string               `json:"postalCode,omitempty"`
	Birthday       *string               `json:"birthday,omitempty"`
	MarriageDate   *string               `json:"marriageDate,omitempty"`
	ImportantDate  *string               `json:"importantDate,omitempty"`
	FirstVisitDate *string               `json:"firstVisitDate,omitempty"`
	Description    *string               `json:"description,omitempty"`
	MarketerNote   *string               `json:"marketerNote,omitempty"`
	SignatureUrl   *string               `json:"signatureUrl,omitempty"`
	ImportProfile  ImportProfileResponse `json:"importProfile"`
}

type CreateCustomerRequest struct {
	ImportMode    string                `json:"importMode"`
	ImportProfile *ImportProfileRequest `json:"importProfile"`
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
	MelliCode           *string  `json:"melliCode,omitempty"`
	PostalCode          *string  `json:"postalCode,omitempty"`
	Birthday            *string  `json:"birthday,omitempty"`
	MarriageDate        *string  `json:"marriageDate,omitempty"`
	ImportantDate       *string  `json:"importantDate,omitempty"`
	FirstVisitDate      *string  `json:"firstVisitDate,omitempty"`
	Gender              *string  `json:"gender,omitempty"`
	CustomerType        string   `json:"customerType"`
	CustomerAgeRange    *string  `json:"customerAgeRange,omitempty"`
	PurchasedCategories []string `json:"purchasedCategories"`
	Description         *string  `json:"description,omitempty"`
	MarketerNote        *string  `json:"marketerNote,omitempty"`
	Signature           *string  `json:"signature,omitempty"`
}

func ToCustomerSummary(row model.ListRow) CustomerSummaryResponse {
	model.NormalizeListRow(&row)
	id := row.ID.String()
	return CustomerSummaryResponse{
		ID:                  id,
		FullName:            row.FullName,
		Phone:               row.Phone,
		CustomerType:        row.CustomerType,
		PurchasedCategories: model.StringSliceOrEmpty(row.PurchasedCategories),
		CustomerAgeRange:    row.CustomerAgeRange,
		Gender:              row.Gender,
		CreatedAt:           formatDate(row.CreatedAt),
		Href:                "/admin/customers/" + id,
	}
}

func ToCustomerDetail(c *model.Customer) CustomerDetailResponse {
	summary := ToCustomerSummary(model.ListRow{
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
		MelliCode:               c.MelliCode,
		PostalCode:              c.PostalCode,
		Birthday:                model.FormatDate(c.Birthday),
		MarriageDate:            model.FormatDate(c.MarriageDate),
		ImportantDate:           model.FormatDate(c.ImportantDate),
		FirstVisitDate:          model.FormatDate(c.FirstVisitDate),
		Description:             c.Description,
		MarketerNote:            c.MarketerNote,
		SignatureUrl:            c.SignatureURL,
		ImportProfile:           profile,
	}
}

func toImportProfileResponse(c *model.Customer) ImportProfileResponse {
	sig := c.SignatureURL
	return ImportProfileResponse{
		FirstName:           c.FirstName,
		LastName:            c.LastName,
		Job:                 c.Job,
		Phone:               c.Phone,
		Email:               c.Email,
		Address:             c.Address,
		MelliCode:           c.MelliCode,
		PostalCode:          c.PostalCode,
		Birthday:            model.FormatDate(c.Birthday),
		MarriageDate:        model.FormatDate(c.MarriageDate),
		ImportantDate:       model.FormatDate(c.ImportantDate),
		FirstVisitDate:      model.FormatDate(c.FirstVisitDate),
		Gender:              c.Gender,
		CustomerType:        c.CustomerType,
		CustomerAgeRange:    c.CustomerAgeRange,
		PurchasedCategories: model.StringSliceOrEmpty(c.PurchasedCategories),
		Description:         c.Description,
		MarketerNote:        c.MarketerNote,
		Signature:           sig,
	}
}

func ToPaginatedCustomers(result model.ListResult) PaginatedCustomersResponse {
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

func (r CreateCustomerRequest) ToCustomer() (*model.Customer, error) {
	if r.ImportProfile == nil {
		return nil, model.ErrInvalidInput
	}
	return importProfileToCustomer(*r.ImportProfile)
}

func (r UpdateCustomerRequest) ToCustomer() (*model.Customer, error) {
	if r.ImportProfile == nil {
		return nil, model.ErrInvalidInput
	}
	return importProfileToCustomer(*r.ImportProfile)
}

func importProfileToCustomer(p ImportProfileRequest) (*model.Customer, error) {
	birthday, err := model.ParseDate(p.Birthday)
	if err != nil {
		return nil, model.ErrInvalidInput
	}
	marriage, err := model.ParseDate(p.MarriageDate)
	if err != nil {
		return nil, model.ErrInvalidInput
	}
	important, err := model.ParseDate(p.ImportantDate)
	if err != nil {
		return nil, model.ErrInvalidInput
	}
	firstVisit, err := model.ParseDate(p.FirstVisitDate)
	if err != nil {
		return nil, model.ErrInvalidInput
	}

	return &model.Customer{
		FirstName:           strings.TrimSpace(p.FirstName),
		LastName:            strings.TrimSpace(p.LastName),
		Job:                 p.Job,
		Phone:               strings.TrimSpace(p.Phone),
		Email:               p.Email,
		Address:             p.Address,
		MelliCode:           p.MelliCode,
		PostalCode:          p.PostalCode,
		Birthday:            birthday,
		MarriageDate:        marriage,
		ImportantDate:       important,
		FirstVisitDate:      firstVisit,
		Gender:              p.Gender,
		CustomerType:        p.CustomerType,
		CustomerAgeRange:    p.CustomerAgeRange,
		PurchasedCategories: model.StringSliceOrEmpty(p.PurchasedCategories),
		Description:         p.Description,
		MarketerNote:        p.MarketerNote,
		SignatureURL:        p.Signature,
	}, nil
}

func formatDate(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

func ParseCustomerID(id string) (model.ID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return model.ID{}, model.ErrInvalidInput
	}
	return model.ID(parsed), nil
}
