package feature_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rahil-gallery/rahil-gallery-server/internal/test/testauth"
)

func TestAuthFeature_RegisterLoginMeRefreshLogout(t *testing.T) {
	app, _ := testauth.NewApp()

	// Register
	regBody := `{"email":"feature@example.com","password":"Secret12","first_name":"Feature","last_name":"User"}`
	reg := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(regBody))
	reg.Header.Set("Content-Type", "application/json")
	regResp, err := app.Test(reg, -1)
	if err != nil {
		t.Fatalf("register request: %v", err)
	}
	if regResp.StatusCode != http.StatusCreated {
		t.Fatalf("register status = %d body = %s", regResp.StatusCode, readBody(regResp.Body))
	}
	regJSON := parseEnvelope(t, regResp.Body)
	access1 := tokenFromData(t, regJSON["data"])
	refresh1 := refreshFromData(t, regJSON["data"])

	// Me with access token
	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+access1)
	meResp, _ := app.Test(meReq, -1)
	if meResp.StatusCode != http.StatusOK {
		t.Fatalf("me status = %d", meResp.StatusCode)
	}
	meJSON := parseEnvelope(t, meResp.Body)
	profile := meJSON["data"].(map[string]any)
	if profile["email"] != "feature@example.com" {
		t.Fatalf("profile email = %v", profile["email"])
	}

	// Login
	loginBody := `{"email":"feature@example.com","password":"Secret12"}`
	login := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(loginBody))
	login.Header.Set("Content-Type", "application/json")
	loginResp, _ := app.Test(login, -1)
	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("login status = %d", loginResp.StatusCode)
	}

	// Refresh
	refBody, _ := json.Marshal(map[string]string{"refresh_token": refresh1})
	ref := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(refBody))
	ref.Header.Set("Content-Type", "application/json")
	refResp, _ := app.Test(ref, -1)
	if refResp.StatusCode != http.StatusOK {
		t.Fatalf("refresh status = %d body=%s", refResp.StatusCode, readBody(refResp.Body))
	}
	refJSON := parseEnvelope(t, refResp.Body)
	refresh2 := refreshFromData(t, refJSON["data"])

	// Logout
	logoutBody, _ := json.Marshal(map[string]string{"refresh_token": refresh2})
	logout := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewReader(logoutBody))
	logout.Header.Set("Content-Type", "application/json")
	logoutResp, _ := app.Test(logout, -1)
	if logoutResp.StatusCode != http.StatusOK {
		t.Fatalf("logout status = %d", logoutResp.StatusCode)
	}

	// Refresh after logout should fail
	ref2 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(logoutBody))
	ref2.Header.Set("Content-Type", "application/json")
	ref2Resp, _ := app.Test(ref2, -1)
	if ref2Resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("refresh after logout status = %d", ref2Resp.StatusCode)
	}

}

func TestAuthFeature_RegisterValidation(t *testing.T) {
	app, _ := testauth.NewApp()
	body := `{"email":"bad","password":"weak","first_name":"","last_name":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestAuthFeature_MeWithoutToken(t *testing.T) {
	app, _ := testauth.NewApp()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func parseEnvelope(t *testing.T, body io.Reader) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.NewDecoder(body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if success, _ := out["success"].(bool); !success {
		t.Fatalf("success=false envelope: %+v", out)
	}
	return out
}

func tokenFromData(t *testing.T, data any) string {
	t.Helper()
	m := data.(map[string]any)
	tok, _ := m["access_token"].(string)
	if tok == "" {
		t.Fatal("missing access_token")
	}
	return tok
}

func refreshFromData(t *testing.T, data any) string {
	t.Helper()
	m := data.(map[string]any)
	tok, _ := m["refresh_token"].(string)
	if tok == "" {
		t.Fatal("missing refresh_token")
	}
	return tok
}

func readBody(body io.ReadCloser) string {
	b, _ := io.ReadAll(body)
	return string(b)
}
