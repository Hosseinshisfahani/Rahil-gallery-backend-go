package feature_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rahil-gallery/rahil-gallery-server/internal/test/testauth"
	httpx "github.com/rahil-gallery/rahil-gallery-server/internal/interfaces/http"
)

func TestRoutesFeature_HealthAndAPIRootRegistered(t *testing.T) {
	app := httpx.NewApp(httpx.RouterDeps{Config: testauth.TestConfig()})

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/health"},
		{http.MethodGet, "/health/ready"},
		{http.MethodGet, "/api/v1/"},
	}

	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			resp, err := app.Test(req, -1)
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			if resp.StatusCode == http.StatusNotFound {
				t.Fatalf("%s %s must be registered (got 404)", tc.method, tc.path)
			}
		})
	}
}

func TestRoutesFeature_AuthLoginRegistered(t *testing.T) {
	app, _ := testauth.NewApp()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"x@example.com","password":"Secret12"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		t.Fatal("POST /api/v1/auth/login must be registered (got 404)")
	}
}
