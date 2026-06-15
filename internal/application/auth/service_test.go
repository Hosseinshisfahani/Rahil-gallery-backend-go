package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rahil-gallery/rahil-gallery-server/internal/application/auth"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/identity"
	"github.com/rahil-gallery/rahil-gallery-server/internal/domain/shared"
	"github.com/rahil-gallery/rahil-gallery-server/internal/infrastructure/security"
	"github.com/rahil-gallery/rahil-gallery-server/internal/test/fakeidentity"
	"github.com/rahil-gallery/rahil-gallery-server/internal/test/testauth"
)

func TestService_Register_Login_Refresh_Logout_Me(t *testing.T) {
	ctx := context.Background()
	store := fakeidentity.NewStore()
	cfg := testauth.TestConfig()
	wire := testauth.WireFromStore(store)
	svc := auth.NewService(cfg, wire.Users, wire.Roles, wire.RefreshTokens,
		security.NewJWTProvider(cfg), security.NewBcryptHasher())

	t.Run("register returns tokens", func(t *testing.T) {
		pair, err := svc.Register(ctx, auth.RegisterInput{
			Email: "unit@example.com", Password: "Secret12",
			FirstName: "Unit", LastName: "Test",
		})
		if err != nil {
			t.Fatalf("register: %v", err)
		}
		if pair.AccessToken == "" || pair.RefreshToken == "" {
			t.Fatal("expected tokens")
		}
		if pair.TokenType != "Bearer" {
			t.Fatalf("token type = %q", pair.TokenType)
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		_, err := svc.Register(ctx, auth.RegisterInput{
			Email: "unit@example.com", Password: "Secret12",
			FirstName: "A", LastName: "B",
		})
		if !errors.Is(err, auth.ErrEmailAlreadyExists) {
			t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
		}
	})

	t.Run("login success", func(t *testing.T) {
		pair, err := svc.Login(ctx, auth.LoginInput{
			Email: "unit@example.com", Password: "Secret12",
		})
		if err != nil {
			t.Fatalf("login: %v", err)
		}
		if pair.AccessToken == "" {
			t.Fatal("missing access token")
		}
	})

	t.Run("login wrong password", func(t *testing.T) {
		_, err := svc.Login(ctx, auth.LoginInput{
			Email: "unit@example.com", Password: "WrongPass1",
		})
		if !errors.Is(err, auth.ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	var refreshToken string
	t.Run("refresh rotates token", func(t *testing.T) {
		login, _ := svc.Login(ctx, auth.LoginInput{
			Email: "unit@example.com", Password: "Secret12",
		})
		refreshToken = login.RefreshToken

		pair, err := svc.Refresh(ctx, refreshToken)
		if err != nil {
			t.Fatalf("refresh: %v", err)
		}
		if pair.RefreshToken == refreshToken {
			t.Fatal("refresh token should be rotated")
		}
		refreshToken = pair.RefreshToken
	})

	t.Run("old refresh invalid after rotation", func(t *testing.T) {
		old := loginRefresh(t, svc, ctx)
		_, _ = svc.Refresh(ctx, old)
		_, err := svc.Refresh(ctx, old)
		if !errors.Is(err, auth.ErrInvalidRefreshToken) {
			t.Fatalf("expected invalid refresh, got %v", err)
		}
	})

	t.Run("me returns profile", func(t *testing.T) {
		user, _ := wire.Users.FindByEmail(ctx, "unit@example.com")
		profile, err := svc.Me(ctx, user.ID)
		if err != nil {
			t.Fatalf("me: %v", err)
		}
		if profile.Email != "unit@example.com" || profile.Role != identity.RoleCustomer {
			t.Fatalf("unexpected profile: %+v", profile)
		}
	})

	t.Run("logout revokes refresh", func(t *testing.T) {
		if err := svc.Logout(ctx, refreshToken); err != nil {
			t.Fatalf("logout: %v", err)
		}
		_, err := svc.Refresh(ctx, refreshToken)
		if !errors.Is(err, auth.ErrInvalidRefreshToken) {
			t.Fatalf("expected invalid refresh after logout, got %v", err)
		}
	})
}

func TestService_Register_validation(t *testing.T) {
	ctx := context.Background()
	store := fakeidentity.NewStore()
	cfg := testauth.TestConfig()
	wire := testauth.WireFromStore(store)
	svc := auth.NewService(cfg, wire.Users, wire.Roles, wire.RefreshTokens,
		security.NewJWTProvider(cfg), security.NewBcryptHasher())

	cases := []struct {
		name  string
		input auth.RegisterInput
	}{
		{"empty email", auth.RegisterInput{Email: "", Password: "Secret12", FirstName: "A", LastName: "B"}},
		{"short password", auth.RegisterInput{Email: "a@b.com", Password: "short", FirstName: "A", LastName: "B"}},
		{"no digit", auth.RegisterInput{Email: "a@b.com", Password: "NoDigitsHere", FirstName: "A", LastName: "B"}},
		{"no letter", auth.RegisterInput{Email: "a@b.com", Password: "12345678", FirstName: "A", LastName: "B"}},
		{"empty name", auth.RegisterInput{Email: "a@b.com", Password: "Secret12", FirstName: "", LastName: "B"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Register(ctx, tc.input)
			if !errors.Is(err, shared.ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestService_Login_inactiveAccount(t *testing.T) {
	ctx := context.Background()
	store := fakeidentity.NewStore()
	cfg := testauth.TestConfig()
	wire := testauth.WireFromStore(store)
	svc := auth.NewService(cfg, wire.Users, wire.Roles, wire.RefreshTokens,
		security.NewJWTProvider(cfg), security.NewBcryptHasher())

	role, _ := wire.Roles.FindByName(ctx, identity.RoleCustomer)
	hash, _ := security.NewBcryptHasher().Hash("Secret12")
	email := "inactive@example.com"
	user := &identity.User{
		ID: uuid.New(), RoleID: role.ID, Email: &email,
		PasswordHash: &hash, FirstName: "In", LastName: "Active",
		Status: identity.UserStatusBanned, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	store.SeedUser(user)

	_, err := svc.Login(ctx, auth.LoginInput{Email: "inactive@example.com", Password: "Secret12"})
	if !errors.Is(err, auth.ErrInactiveAccount) {
		t.Fatalf("expected ErrInactiveAccount, got %v", err)
	}
}

func loginRefresh(t *testing.T, svc *auth.Service, ctx context.Context) string {
	t.Helper()
	pair, err := svc.Login(ctx, auth.LoginInput{Email: "unit@example.com", Password: "Secret12"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	return pair.RefreshToken
}
