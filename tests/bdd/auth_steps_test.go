package bdd_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/gofiber/fiber/v2"
	"github.com/rahil-gallery/rahil-gallery-server/internal/test/fakeidentity"
	"github.com/rahil-gallery/rahil-gallery-server/internal/test/testauth"
)

type authScenario struct {
	app          *fiber.App
	store        *fakeidentity.Store
	lastResp     *http.Response
	lastBody     map[string]any
	accessToken  string
	refreshToken string
}

func TestAuthBDD(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			sc := &authScenario{}
			ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				sc.app, sc.store = testauth.NewApp()
				sc.lastResp = nil
				sc.lastBody = nil
				sc.accessToken = ""
				sc.refreshToken = ""
				return ctx, nil
			})

			ctx.Step(`^the auth API is running$`, sc.apiRunning)
			ctx.Step(`^no user exists with email "([^"]*)"$`, sc.noUserWithEmail)
			ctx.Step(`^a registered user:$`, sc.registeredUserTable)
			ctx.Step(`^I register with:$`, sc.registerWithTable)
			ctx.Step(`^I login with email "([^"]*)" and password "([^"]*)"$`, sc.loginWithEmailPassword)
			ctx.Step(`^I request my profile with the saved access token$`, sc.requestProfileWithToken)
			ctx.Step(`^I request my profile without a token$`, sc.requestProfileWithoutToken)
			ctx.Step(`^I refresh the session with the saved refresh token$`, sc.refreshWithSavedToken)
			ctx.Step(`^I logout with the saved refresh token$`, sc.logoutWithSavedToken)
			ctx.Step(`^the response status should be (\d+)$`, sc.responseStatusShouldBe)
			ctx.Step(`^the response should contain access and refresh tokens$`, sc.responseHasTokens)
			ctx.Step(`^the refresh token should be rotated$`, sc.refreshTokenRotated)
			ctx.Step(`^the profile email should be "([^"]*)"$`, sc.profileEmailShouldBe)
			ctx.Step(`^the error code should be "([^"]*)"$`, sc.errorCodeShouldBe)
		},
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../../features/auth"},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("bdd scenarios failed")
	}
}

func (s *authScenario) apiRunning() error { return nil }

func (s *authScenario) noUserWithEmail(email string) error {
	repo := fakeidentity.UserRepo{S: s.store}
	if _, err := repo.FindByEmail(context.Background(), email); err == nil {
		return fmt.Errorf("user %s already exists", email)
	}
	return nil
}

func (s *authScenario) registeredUserTable(table *godog.Table) error {
	email, err := tableValue(table, "email")
	if err != nil {
		return err
	}
	password, err := tableValue(table, "password")
	if err != nil {
		return err
	}
	return s.registerPayload(email, password, "BDD", "User")
}

func (s *authScenario) registerWithTable(table *godog.Table) error {
	email, err := tableValue(table, "email")
	if err != nil {
		return err
	}
	password, err := tableValue(table, "password")
	if err != nil {
		return err
	}
	first, err := tableValue(table, "first_name")
	if err != nil {
		return err
	}
	last, err := tableValue(table, "last_name")
	if err != nil {
		return err
	}
	return s.registerPayload(email, password, first, last)
}

// tableValue reads a two-column Gherkin table by row key (no header row in Godog).
func tableValue(table *godog.Table, key string) (string, error) {
	for _, row := range table.Rows {
		if len(row.Cells) < 2 {
			continue
		}
		if strings.TrimSpace(row.Cells[0].Value) == key {
			return strings.TrimSpace(row.Cells[1].Value), nil
		}
	}
	return "", fmt.Errorf("table missing key %q", key)
}

func (s *authScenario) registerPayload(email, password, first, last string) error {
	body := fmt.Sprintf(`{"email":%q,"password":%q,"first_name":%q,"last_name":%q}`,
		email, password, first, last)
	return s.doJSON(http.MethodPost, "/api/v1/auth/register", body)
}

func (s *authScenario) loginWithEmailPassword(email, password string) error {
	body := fmt.Sprintf(`{"email":%q,"password":%q}`, email, password)
	return s.doJSON(http.MethodPost, "/api/v1/auth/login", body)
}

func (s *authScenario) requestProfileWithToken() error {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+s.accessToken)
	return s.doRequest(req)
}

func (s *authScenario) requestProfileWithoutToken() error {
	return s.doJSON(http.MethodGet, "/api/v1/auth/me", "")
}

func (s *authScenario) refreshWithSavedToken() error {
	old := s.refreshToken
	body := fmt.Sprintf(`{"refresh_token":%q}`, s.refreshToken)
	if err := s.doJSON(http.MethodPost, "/api/v1/auth/refresh", body); err != nil {
		return err
	}
	// Rotation only applies to successful refresh (e.g. after logout we expect 401, same token)
	if s.lastResp != nil && s.lastResp.StatusCode == http.StatusOK && s.refreshToken == old {
		return fmt.Errorf("refresh token was not rotated")
	}
	return nil
}

func (s *authScenario) logoutWithSavedToken() error {
	body := fmt.Sprintf(`{"refresh_token":%q}`, s.refreshToken)
	return s.doJSON(http.MethodPost, "/api/v1/auth/logout", body)
}

func (s *authScenario) responseStatusShouldBe(code int) error {
	if s.lastResp == nil {
		return fmt.Errorf("no response")
	}
	if s.lastResp.StatusCode != code {
		return fmt.Errorf("expected status %d got %d body=%v", code, s.lastResp.StatusCode, s.lastBody)
	}
	return nil
}

func (s *authScenario) responseHasTokens() error {
	data, ok := s.lastBody["data"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing data: %v", s.lastBody)
	}
	access, _ := data["access_token"].(string)
	refresh, _ := data["refresh_token"].(string)
	if access == "" || refresh == "" {
		return fmt.Errorf("missing tokens in %v", data)
	}
	s.accessToken = access
	s.refreshToken = refresh
	return nil
}

func (s *authScenario) refreshTokenRotated() error {
	if s.refreshToken == "" {
		return fmt.Errorf("no refresh token saved")
	}
	return nil
}

func (s *authScenario) profileEmailShouldBe(email string) error {
	data, ok := s.lastBody["data"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing data")
	}
	if data["email"] != email {
		return fmt.Errorf("expected email %q got %v", email, data["email"])
	}
	return nil
}

func (s *authScenario) errorCodeShouldBe(code string) error {
	errObj, ok := s.lastBody["error"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing error: %v", s.lastBody)
	}
	if errObj["code"] != code {
		return fmt.Errorf("expected code %q got %v", code, errObj["code"])
	}
	return nil
}

func (s *authScenario) doJSON(method, path, body string) error {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	return s.doRequest(req)
}

func (s *authScenario) doRequest(req *http.Request) error {
	resp, err := s.app.Test(req, -1)
	if err != nil {
		return err
	}
	s.lastResp = resp
	s.lastBody = nil
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&s.lastBody); err != nil {
		// 204 or empty
		s.lastBody = map[string]any{}
	}
	if s.lastBody == nil {
		s.lastBody = map[string]any{}
	}

	// capture tokens on success responses
	if data, ok := s.lastBody["data"].(map[string]any); ok {
		if at, _ := data["access_token"].(string); at != "" {
			s.accessToken = at
		}
		if rt, _ := data["refresh_token"].(string); rt != "" {
			s.refreshToken = rt
		}
	}
	return nil
}
