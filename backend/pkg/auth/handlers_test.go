package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleLoginMockModeRedirectsWithToken(t *testing.T) {
	h := NewAuthHandlers(nil, "test-secret-at-least-32-characters-long", "http://localhost:5173")

	req := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	rr := httptest.NewRecorder()
	h.HandleLogin(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect, got %d", rr.Code)
	}
	loc := rr.Header().Get("Location")
	if !strings.HasPrefix(loc, "http://localhost:5173/?") {
		t.Errorf("expected redirect to the configured frontend URL, got %q", loc)
	}
	if !strings.Contains(loc, "user=operator-dev") || !strings.Contains(loc, "role=operator") {
		t.Errorf("expected mock user/role in redirect URL, got %q", loc)
	}
	if !strings.Contains(loc, "token=") {
		t.Errorf("expected a signed token in the redirect URL, got %q", loc)
	}
}

func TestHandleLoginRealOIDCRedirectsToProviderNotFrontend(t *testing.T) {
	// A non-nil OIDCClient (even a bare one) should take the "real SSO"
	// branch: set a CSRF state cookie and redirect to the IdP, never
	// straight to the frontend the way mock mode does.
	h := NewAuthHandlers(&OIDCClient{}, "test-secret-at-least-32-characters-long", "http://localhost:5173")

	req := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	rr := httptest.NewRecorder()
	h.HandleLogin(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect, got %d", rr.Code)
	}
	if loc := rr.Header().Get("Location"); strings.HasPrefix(loc, "http://localhost:5173") {
		t.Errorf("expected a redirect to the IdP, not straight to the frontend, got %q", loc)
	}
	found := false
	for _, c := range rr.Result().Cookies() {
		if c.Name == "oauthstate" && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Error("expected an oauthstate CSRF cookie to be set")
	}
}

func TestHandleCallbackMockModeReturns400(t *testing.T) {
	h := NewAuthHandlers(nil, "test-secret-at-least-32-characters-long", "http://localhost:5173")

	req := httptest.NewRequest(http.MethodGet, "/api/auth/callback", nil)
	rr := httptest.NewRecorder()
	h.HandleCallback(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 when OIDC isn't configured, got %d", rr.Code)
	}
}

func TestHandleCallbackMissingStateCookieReturns401(t *testing.T) {
	h := NewAuthHandlers(&OIDCClient{}, "test-secret-at-least-32-characters-long", "http://localhost:5173")

	req := httptest.NewRequest(http.MethodGet, "/api/auth/callback?state=abc&code=xyz", nil)
	rr := httptest.NewRecorder()
	h.HandleCallback(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with no state cookie, got %d", rr.Code)
	}
}

func TestHandleCallbackStateMismatchReturns401(t *testing.T) {
	h := NewAuthHandlers(&OIDCClient{}, "test-secret-at-least-32-characters-long", "http://localhost:5173")

	req := httptest.NewRequest(http.MethodGet, "/api/auth/callback?state=wrong-value&code=xyz", nil)
	req.AddCookie(&http.Cookie{Name: "oauthstate", Value: "expected-value"})
	rr := httptest.NewRecorder()
	h.HandleCallback(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 on state mismatch, got %d", rr.Code)
	}
}

func TestHandleCallbackMissingCodeReturns400(t *testing.T) {
	h := NewAuthHandlers(&OIDCClient{}, "test-secret-at-least-32-characters-long", "http://localhost:5173")

	req := httptest.NewRequest(http.MethodGet, "/api/auth/callback?state=matching-value", nil)
	req.AddCookie(&http.Cookie{Name: "oauthstate", Value: "matching-value"})
	rr := httptest.NewRecorder()
	h.HandleCallback(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 with no authorization code, got %d", rr.Code)
	}
}

// Note: the success path (HandleCallback exchanging a real code and
// redirecting with a token) isn't covered here - that requires a working
// fake OIDC provider server (coreos/go-oidc fetches provider metadata over
// HTTP), which is a larger separate effort. The guard clauses above and
// the mock-mode path are what changed in this fix; the redirect-vs-JSON
// bug this closes lived in those paths.
