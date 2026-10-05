package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/rahil-gallery/rahil-gallery-server/internal/config"
	"github.com/rahil-gallery/rahil-gallery-server/internal/model"
	"github.com/rahil-gallery/rahil-gallery-server/internal/repository"
)

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrEmailAlreadyExists  = errors.New("email already registered")
	ErrInactiveAccount     = errors.New("account is not active")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrPasswordNotSet      = errors.New("password not set; complete account setup")
)

type RegisterInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Phone     *string
}

type LoginInput struct {
	Email    string
	Password string
}

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	TokenType        string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

type UserProfile struct {
	ID        model.ID `json:"id"`
	Email     string   `json:"email"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Phone     *string  `json:"phone,omitempty"`
	Role      string   `json:"role"`
	Status    string   `json:"status"`
}

type AuthService struct {
	users         *repository.UserRepository
	roles         *repository.RoleRepository
	refreshTokens *repository.RefreshTokenRepository
	tokens        JWTProvider
	passwords     BcryptHasher
	refreshTTL    time.Duration
}

func NewAuthService(
	cfg config.Config,
	users *repository.UserRepository,
	roles *repository.RoleRepository,
	refreshTokens *repository.RefreshTokenRepository,
	tokens JWTProvider,
) *AuthService {
	return &AuthService{
		users:         users,
		roles:         roles,
		refreshTokens: refreshTokens,
		tokens:        tokens,
		passwords:     NewBcryptHasher(),
		refreshTTL:    cfg.JWTRefreshTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (TokenPair, error) {
	if err := validateRegister(in); err != nil {
		return TokenPair{}, err
	}

	email := normalizeEmail(in.Email)
	if _, err := s.users.FindByEmail(ctx, email); err == nil {
		return TokenPair{}, ErrEmailAlreadyExists
	} else if !errors.Is(err, model.ErrNotFound) {
		return TokenPair{}, err
	}

	role, err := s.roles.FindByName(ctx, model.RoleCustomer)
	if err != nil {
		return TokenPair{}, err
	}

	hash, err := s.passwords.Hash(in.Password)
	if err != nil {
		return TokenPair{}, err
	}

	now := time.Now()
	user := &model.User{
		ID:           uuid.New(),
		RoleID:       role.ID,
		Email:        strPtr(email),
		Phone:        in.Phone,
		PasswordHash: &hash,
		FirstName:    strings.TrimSpace(in.FirstName),
		LastName:     strings.TrimSpace(in.LastName),
		Status:       model.UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return TokenPair{}, err
	}

	return s.issueTokenPair(ctx, user, role.Name)
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (TokenPair, error) {
	email := normalizeEmail(in.Email)
	if email == "" || in.Password == "" {
		return TokenPair{}, model.ErrInvalidInput
	}

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return TokenPair{}, ErrInvalidCredentials
		}
		return TokenPair{}, err
	}

	if user.Status != model.UserStatusActive {
		return TokenPair{}, ErrInactiveAccount
	}

	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return TokenPair{}, ErrPasswordNotSet
	}

	if err := s.passwords.Compare(*user.PasswordHash, in.Password); err != nil {
		return TokenPair{}, ErrInvalidCredentials
	}

	role, err := s.roles.FindByID(ctx, user.RoleID)
	if err != nil {
		return TokenPair{}, err
	}

	now := time.Now()
	user.LastLoginAt = &now
	user.UpdatedAt = now
	_ = s.users.Update(ctx, user)

	return s.issueTokenPair(ctx, user, role.Name)
}

func (s *AuthService) Refresh(ctx context.Context, refreshTokenPlain string) (TokenPair, error) {
	if refreshTokenPlain == "" {
		return TokenPair{}, ErrInvalidRefreshToken
	}

	hash := HashRefreshToken(refreshTokenPlain)
	stored, err := s.refreshTokens.FindByHash(ctx, hash)
	if err != nil {
		return TokenPair{}, ErrInvalidRefreshToken
	}

	if stored.RevokedAt != nil || time.Now().After(stored.ExpiresAt) {
		return TokenPair{}, ErrInvalidRefreshToken
	}

	user, err := s.users.FindByID(ctx, stored.UserID)
	if err != nil {
		return TokenPair{}, ErrInvalidRefreshToken
	}

	if user.Status != model.UserStatusActive {
		return TokenPair{}, ErrInactiveAccount
	}

	role, err := s.roles.FindByID(ctx, user.RoleID)
	if err != nil {
		return TokenPair{}, err
	}

	if err := s.refreshTokens.Revoke(ctx, stored.ID); err != nil && !errors.Is(err, model.ErrNotFound) {
		return TokenPair{}, err
	}

	return s.issueTokenPair(ctx, user, role.Name)
}

func (s *AuthService) Logout(ctx context.Context, refreshTokenPlain string) error {
	if refreshTokenPlain == "" {
		return model.ErrInvalidInput
	}

	hash := HashRefreshToken(refreshTokenPlain)
	stored, err := s.refreshTokens.FindByHash(ctx, hash)
	if err != nil {
		return nil
	}

	return s.refreshTokens.Revoke(ctx, stored.ID)
}

func (s *AuthService) Me(ctx context.Context, userID model.ID) (UserProfile, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return UserProfile{}, err
	}

	role, err := s.roles.FindByID(ctx, user.RoleID)
	if err != nil {
		return UserProfile{}, err
	}

	return toProfile(user, role.Name), nil
}

func (s *AuthService) issueTokenPair(ctx context.Context, user *model.User, roleName string) (TokenPair, error) {
	access, accessExp, err := s.tokens.IssueAccess(AccessClaims{
		UserID: user.ID,
		Email:  emailString(user.Email),
		Role:   roleName,
	})
	if err != nil {
		return TokenPair{}, err
	}

	plain, hash, err := GenerateRefreshToken()
	if err != nil {
		return TokenPair{}, err
	}

	now := time.Now()
	refreshExp := now.Add(s.refreshTTL)
	rt := &model.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: refreshExp,
		CreatedAt: now,
	}

	if err := s.refreshTokens.Create(ctx, rt); err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:      access,
		RefreshToken:     plain,
		TokenType:        "Bearer",
		AccessExpiresAt:  accessExp,
		RefreshExpiresAt: refreshExp,
	}, nil
}

func toProfile(user *model.User, roleName string) UserProfile {
	return UserProfile{
		ID:        user.ID,
		Email:     emailString(user.Email),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		Role:      roleName,
		Status:    string(user.Status),
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func emailString(email *string) string {
	if email == nil {
		return ""
	}
	return *email
}

func validateRegister(in RegisterInput) error {
	if normalizeEmail(in.Email) == "" {
		return model.ErrInvalidInput
	}
	if len(in.Password) < 8 {
		return model.ErrInvalidInput
	}
	if strings.TrimSpace(in.FirstName) == "" || strings.TrimSpace(in.LastName) == "" {
		return model.ErrInvalidInput
	}
	if !isPasswordStrongEnough(in.Password) {
		return model.ErrInvalidInput
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isPasswordStrongEnough(password string) bool {
	var hasLetter, hasDigit bool
	for _, r := range password {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}
