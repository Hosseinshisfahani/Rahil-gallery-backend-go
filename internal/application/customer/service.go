package customer

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	domain "github.com/rahil-gallery/rahil-gallery-server/internal/domain/customer"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/identity"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
)

var (
	ErrPhoneAlreadyExists    = errors.New("phone number already registered")
	ErrInvalidBlockReason    = errors.New("invalid block reason")
	ErrInvalidTag            = errors.New("invalid customer tag")
	ErrInvalidSavedView      = errors.New("invalid saved view")
	ErrSavedViewNameRequired = errors.New("saved view name is required")
)

var validBlockReasons = map[domain.BlockReason]struct{}{
	domain.BlockReasonFraudSuspicion: {},
	domain.BlockReasonPaymentIssues:  {},
	domain.BlockReasonReturnAbuse:    {},
	domain.BlockReasonSystemMisuse:   {},
}

var validTags = map[string]struct{}{
	"VIP":              {},
	"Bridal customer":  {},
	"High spender":     {},
	"At-risk":          {},
	"Influencer lead":  {},
}

type CreateInput struct {
	ImportMode      string
	FullName        string
	Phone           string
	Email           string
	Locale          string
	DefaultRingSize string
	IsVIP           bool
	Tags            []string
	ImportProfile   *domain.ImportProfileInput
}

type UpdateInput struct {
	FullName        *string
	Phone           *string
	Email           *string
	Locale          *string
	DefaultRingSize *string
	IsVIP           *bool
	Tags            *[]string
	ImportProfile   *domain.ImportProfileInput
}

type Service struct {
	users    identity.UserRepository
	roles    identity.RoleRepository
	customer domain.Repository
}

func NewService(
	users identity.UserRepository,
	roles identity.RoleRepository,
	customer domain.Repository,
) *Service {
	return &Service{users: users, roles: roles, customer: customer}
}

func (s *Service) List(ctx context.Context, filter domain.ListFilter, page, perPage int) (domain.ListResult, error) {
	return s.customer.List(ctx, filter, page, perPage)
}

type SavedViewInput struct {
	Name     string
	ViewType domain.ListViewType
	Filters  domain.SavedFilterPayload
	IsShared bool
	Position int
}

func (s *Service) ListSegments(ctx context.Context, staffID shared.ID) (domain.SegmentsResult, error) {
	builtin, err := s.customer.CountCustomersBySegment(ctx)
	if err != nil {
		return domain.SegmentsResult{}, err
	}
	segType := domain.ListViewTypeSegment
	saved, err := s.customer.ListSavedViews(ctx, staffID, &segType)
	if err != nil {
		return domain.SegmentsResult{}, err
	}
	return domain.SegmentsResult{Builtin: builtin, Saved: saved}, nil
}

func (s *Service) ListSavedViews(ctx context.Context, staffID shared.ID, viewType *domain.ListViewType) ([]domain.SavedListView, error) {
	return s.customer.ListSavedViews(ctx, staffID, viewType)
}

func (s *Service) GetSavedView(ctx context.Context, staffID, viewID shared.ID) (*domain.SavedListView, error) {
	return s.customer.GetSavedView(ctx, viewID, staffID)
}

func (s *Service) CreateSavedView(ctx context.Context, staffID shared.ID, in SavedViewInput) (*domain.SavedListView, error) {
	if err := validateSavedViewInput(in); err != nil {
		return nil, err
	}
	now := time.Now()
	view := &domain.SavedListView{
		ID:        shared.ID(uuid.New()),
		OwnerID:   staffID,
		Name:      strings.TrimSpace(in.Name),
		ViewType:  in.ViewType,
		Filters:   in.Filters,
		IsShared:  in.IsShared,
		Position:  in.Position,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.customer.CreateSavedView(ctx, view); err != nil {
		return nil, err
	}
	return view, nil
}

func (s *Service) UpdateSavedView(ctx context.Context, staffID, viewID shared.ID, in SavedViewInput) (*domain.SavedListView, error) {
	if err := validateSavedViewInput(in); err != nil {
		return nil, err
	}
	existing, err := s.customer.GetSavedView(ctx, viewID, staffID)
	if err != nil {
		return nil, err
	}
	if existing.OwnerID != staffID {
		return nil, shared.ErrForbidden
	}
	existing.Name = strings.TrimSpace(in.Name)
	existing.ViewType = in.ViewType
	existing.Filters = in.Filters
	existing.IsShared = in.IsShared
	existing.Position = in.Position
	existing.UpdatedAt = time.Now()
	if err := s.customer.UpdateSavedView(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *Service) DeleteSavedView(ctx context.Context, staffID, viewID shared.ID) error {
	return s.customer.DeleteSavedView(ctx, viewID, staffID)
}

func validateSavedViewInput(in SavedViewInput) error {
	if strings.TrimSpace(in.Name) == "" {
		return ErrSavedViewNameRequired
	}
	if in.ViewType != domain.ListViewTypeFilter && in.ViewType != domain.ListViewTypeSegment {
		return ErrInvalidSavedView
	}
	filter, err := in.Filters.ToListFilter()
	if err != nil {
		return shared.ErrInvalidInput
	}
	switch in.ViewType {
	case domain.ListViewTypeSegment:
		if filter.Segment == "" {
			return ErrInvalidSavedView
		}
	case domain.ListViewTypeFilter:
		if !filter.HasAdvancedFilters() && strings.TrimSpace(filter.QuickSearch) == "" {
			return ErrInvalidSavedView
		}
	}
	if filter.Segment != "" && !isValidSegment(filter.Segment) {
		return ErrInvalidSavedView
	}
	return nil
}

func isValidSegment(s string) bool {
	switch domain.Segment(s) {
	case domain.SegmentNew, domain.SegmentActive, domain.SegmentReturning, domain.SegmentVIP, domain.SegmentInactive:
		return true
	default:
		return false
	}
}

func (s *Service) Get(ctx context.Context, id shared.ID) (*domain.Detail, error) {
	return s.customer.GetDetail(ctx, id)
}

func (s *Service) Create(ctx context.Context, adminID shared.ID, in CreateInput) (*domain.Detail, error) {
	phone, first, last, err := s.resolveIdentity(in)
	if err != nil {
		return nil, err
	}

	exists, err := s.customer.PhoneExists(ctx, phone, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrPhoneAlreadyExists
	}

	if in.Email != "" {
		if _, err := s.users.FindByEmail(ctx, normalizeEmail(in.Email)); err == nil {
			return nil, shared.ErrConflict
		} else if !errors.Is(err, shared.ErrNotFound) {
			return nil, err
		}
	}

	role, err := s.roles.FindByName(ctx, identity.RoleCustomer)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &identity.User{
		ID:        uuid.New(),
		RoleID:    role.ID,
		FirstName: first,
		LastName:  last,
		Phone:     strPtr(phone),
		Status:    identity.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if in.Email != "" {
		user.Email = strPtr(normalizeEmail(in.Email))
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	locale := in.Locale
	if locale == "" {
		locale = "fa"
	}

	isVIP := in.IsVIP
	var importMode *domain.ImportMode
	var importProfile json.RawMessage

	if in.ImportMode == string(domain.ImportModeHistoryIncluded) && in.ImportProfile != nil {
		mode := domain.ImportModeHistoryIncluded
		importMode = &mode
		if in.ImportProfile.CustomerType == "vip" {
			isVIP = true
		}
		importProfile, err = domain.MarshalImportProfile(*in.ImportProfile)
		if err != nil {
			return nil, err
		}
	} else if in.ImportMode == string(domain.ImportModeQuick) || in.ImportMode == "" {
		mode := domain.ImportModeQuick
		importMode = &mode
	}

	var vipSource *domain.VIPSource
	if isVIP {
		manual := domain.VIPSourceManual
		vipSource = &manual
	}

	tags := in.Tags
	if tags == nil {
		tags = []string{}
	}

	profile := &domain.Profile{
		UserID:         user.ID,
		Locale:         locale,
		DefaultRingSize: optionalString(in.DefaultRingSize),
		IsVIP:          isVIP,
		VIPSource:      vipSource,
		ImportMode:     importMode,
		ImportProfile:  importProfile,
		Tags:           tags,
		LastActivityAt: &now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if in.DefaultRingSize != "" {
		profile.DefaultRingSize = &in.DefaultRingSize
	}

	if err := s.customer.UpsertProfile(ctx, profile); err != nil {
		return nil, err
	}

	details := "Customer account created"
	if importMode != nil && *importMode == domain.ImportModeHistoryIncluded {
		details = "Customer imported with history"
	}
	_ = s.audit(ctx, adminID, user.ID, domain.AuditActionCreated, nil, &details)

	return s.customer.GetDetail(ctx, user.ID)
}

func (s *Service) Update(ctx context.Context, adminID, userID shared.ID, in UpdateInput) (*domain.Detail, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if in.Phone != nil {
		phone := strings.TrimSpace(*in.Phone)
		if phone == "" {
			return nil, shared.ErrInvalidInput
		}
		exists, err := s.customer.PhoneExists(ctx, phone, &userID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrPhoneAlreadyExists
		}
		user.Phone = &phone
	}

	if in.FullName != nil {
		first, last := splitName(*in.FullName)
		if first == "" {
			return nil, shared.ErrInvalidInput
		}
		user.FirstName = first
		user.LastName = last
	}

	if in.Email != nil {
		if *in.Email == "" {
			user.Email = nil
		} else {
			email := normalizeEmail(*in.Email)
			user.Email = &email
		}
	}

	user.UpdatedAt = time.Now()
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}

	profile, err := s.customer.GetProfile(ctx, userID)
	if err != nil {
		if !errors.Is(err, shared.ErrNotFound) {
			return nil, err
		}
		profile = &domain.Profile{
			UserID:    userID,
			Locale:    "fa",
			Tags:      []string{},
			CreatedAt: time.Now(),
		}
	}

	if in.Locale != nil {
		profile.Locale = *in.Locale
	}
	if in.DefaultRingSize != nil {
		profile.DefaultRingSize = optionalString(*in.DefaultRingSize)
	}
	if in.IsVIP != nil {
		profile.IsVIP = *in.IsVIP
		if *in.IsVIP {
			manual := domain.VIPSourceManual
			profile.VIPSource = &manual
		} else {
			profile.VIPSource = nil
		}
	}
	if in.Tags != nil {
		profile.Tags = *in.Tags
	}
	if in.ImportProfile != nil {
		profile.ImportProfile, err = domain.MarshalImportProfile(*in.ImportProfile)
		if err != nil {
			return nil, err
		}
		mode := domain.ImportModeHistoryIncluded
		profile.ImportMode = &mode
	}

	profile.UpdatedAt = time.Now()
	if err := s.customer.UpsertProfile(ctx, profile); err != nil {
		return nil, err
	}

	details := "Profile updated"
	_ = s.audit(ctx, adminID, userID, domain.AuditActionProfileEdit, nil, &details)

	return s.customer.GetDetail(ctx, userID)
}

func (s *Service) Delete(ctx context.Context, userID shared.ID) error {
	return s.users.SoftDelete(ctx, userID)
}

func (s *Service) Block(ctx context.Context, adminID, userID shared.ID, reason domain.BlockReason, note string) (*domain.Detail, error) {
	if _, ok := validBlockReasons[reason]; !ok {
		return nil, ErrInvalidBlockReason
	}

	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.Status = identity.UserStatusBanned
	user.UpdatedAt = time.Now()
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}

	profile, err := s.loadOrInitProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	profile.BlockReason = &reason
	if note != "" {
		profile.BlockNote = &note
	}
	profile.UpdatedAt = time.Now()
	if err := s.customer.UpsertProfile(ctx, profile); err != nil {
		return nil, err
	}

	reasonStr := string(reason)
	_ = s.audit(ctx, adminID, userID, domain.AuditActionBlock, &reasonStr, optionalStringPtr(note))

	return s.customer.GetDetail(ctx, userID)
}

func (s *Service) Unblock(ctx context.Context, adminID, userID shared.ID, justification string) (*domain.Detail, error) {
	if strings.TrimSpace(justification) == "" {
		return nil, shared.ErrInvalidInput
	}

	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.Status = identity.UserStatusActive
	user.UpdatedAt = time.Now()
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}

	profile, err := s.loadOrInitProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	profile.BlockReason = nil
	profile.BlockNote = nil
	profile.UpdatedAt = time.Now()
	if err := s.customer.UpsertProfile(ctx, profile); err != nil {
		return nil, err
	}

	_ = s.audit(ctx, adminID, userID, domain.AuditActionUnblock, nil, &justification)

	return s.customer.GetDetail(ctx, userID)
}

func (s *Service) ToggleVIP(ctx context.Context, adminID, userID shared.ID) (*domain.Detail, error) {
	profile, err := s.loadOrInitProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	action := domain.AuditActionVIPRemove
	if !profile.IsVIP {
		action = domain.AuditActionVIPAssign
		manual := domain.VIPSourceManual
		profile.VIPSource = &manual
	} else {
		profile.VIPSource = nil
	}
	profile.IsVIP = !profile.IsVIP
	profile.UpdatedAt = time.Now()

	if err := s.customer.UpsertProfile(ctx, profile); err != nil {
		return nil, err
	}

	details := "VIP status toggled"
	_ = s.audit(ctx, adminID, userID, action, nil, &details)

	return s.customer.GetDetail(ctx, userID)
}

func (s *Service) ToggleTag(ctx context.Context, adminID, userID shared.ID, tag string) (*domain.Detail, error) {
	if _, ok := validTags[tag]; !ok {
		return nil, ErrInvalidTag
	}

	profile, err := s.loadOrInitProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	action := domain.AuditActionTagAdd
	found := false
	for i, t := range profile.Tags {
		if t == tag {
			found = true
			profile.Tags = append(profile.Tags[:i], profile.Tags[i+1:]...)
			action = domain.AuditActionTagRemove
			break
		}
	}
	if !found {
		profile.Tags = append(profile.Tags, tag)
	}
	profile.UpdatedAt = time.Now()

	if err := s.customer.UpsertProfile(ctx, profile); err != nil {
		return nil, err
	}

	details := "Tag: " + tag
	_ = s.audit(ctx, adminID, userID, action, nil, &details)

	return s.customer.GetDetail(ctx, userID)
}

func (s *Service) AddNote(ctx context.Context, adminID, userID shared.ID, body string) (*domain.Detail, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, shared.ErrInvalidInput
	}

	now := time.Now()
	note := &domain.Note{
		ID:        uuid.New(),
		UserID:    userID,
		AuthorID:  adminID,
		Body:      body,
		CreatedAt: now,
	}
	if err := s.customer.AddNote(ctx, note); err != nil {
		return nil, err
	}

	details := "Internal note added"
	_ = s.audit(ctx, adminID, userID, domain.AuditActionNoteAdd, nil, &details)

	return s.customer.GetDetail(ctx, userID)
}

func (s *Service) resolveIdentity(in CreateInput) (phone, first, last string, err error) {
	if in.ImportMode == string(domain.ImportModeHistoryIncluded) && in.ImportProfile != nil {
		p := in.ImportProfile
		phone = strings.TrimSpace(p.Phone)
		first = strings.TrimSpace(p.FirstName)
		last = strings.TrimSpace(p.LastName)
		if phone == "" || first == "" || last == "" {
			return "", "", "", shared.ErrInvalidInput
		}
		return phone, first, last, nil
	}

	phone = strings.TrimSpace(in.Phone)
	first, last = splitName(in.FullName)
	if phone == "" || first == "" {
		return "", "", "", shared.ErrInvalidInput
	}
	return phone, first, last, nil
}

func (s *Service) loadOrInitProfile(ctx context.Context, userID shared.ID) (*domain.Profile, error) {
	profile, err := s.customer.GetProfile(ctx, userID)
	if err == nil {
		return profile, nil
	}
	if !errors.Is(err, shared.ErrNotFound) {
		return nil, err
	}
	now := time.Now()
	return &domain.Profile{
		UserID:    userID,
		Locale:    "fa",
		Tags:      []string{},
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *Service) audit(ctx context.Context, adminID, targetID shared.ID, action domain.AuditAction, reason, details *string) error {
	return s.customer.AppendAudit(ctx, &domain.AuditEntry{
		ID:           uuid.New(),
		AdminID:      adminID,
		TargetUserID: targetID,
		Action:       action,
		Reason:       reason,
		Details:      details,
		CreatedAt:    time.Now(),
	})
}

func splitName(full string) (string, string) {
	full = strings.TrimSpace(full)
	if full == "" {
		return "", ""
	}
	parts := strings.Fields(full)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func strPtr(s string) *string {
	return &s
}

func optionalString(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func optionalStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
