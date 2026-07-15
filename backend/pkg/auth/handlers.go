package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"
)

// AuthHandlers aggregates OIDC clients and JWT configurations.
type AuthHandlers struct {
	oidcClient  *OIDCClient
	jwtSecret   string
	frontendURL string
}

// NewAuthHandlers instantiates OIDC and JWT HTTP handler functions.
// frontendURL is where the browser is redirected back to after both the
// mock-mode and real-OIDC login flows, carrying the issued token/user/role
// as query params (see AuthContext.jsx's callback parsing).
func NewAuthHandlers(oidcClient *OIDCClient, jwtSecret string, frontendURL string) *AuthHandlers {
	return &AuthHandlers{
		oidcClient:  oidcClient,
		jwtSecret:   jwtSecret,
		frontendURL: frontendURL,
	}
}

// redirectToFrontend sends the browser back to the SPA with the issued
// session encoded as query params, matching AuthContext.jsx's expected
// shape. Used by both the mock-mode login and the real OIDC callback so
// there's exactly one place this format is defined.
func (h *AuthHandlers) redirectToFrontend(w http.ResponseWriter, r *http.Request, token, username, role string) {
	redirectURL := h.frontendURL + "/?token=" + token + "&user=" + username + "&role=" + role
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// HandleLogin generates OAuth state cookies and redirects users to OIDC provider login.
func (h *AuthHandlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if h.oidcClient == nil {
		// Mock Mode: Generate default developer operator claims
		userClaims := UserClaims{
			Username: "operator-dev",
			Roles:    []string{"operator"},
		}
		tokenString, err := SignJWT(userClaims, h.jwtSecret, 8*time.Hour)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"code": "INTERNAL_ERROR", "message": "Failed to sign mock token"}`))
			return
		}

		h.redirectToFrontend(w, r, tokenString, userClaims.Username, userClaims.Roles[0])
		return
	}

	b := make([]byte, 16)
	_, _ = rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	// Set state cookie to prevent CSRF attacks
	http.SetCookie(w, &http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
		Path:     "/",
	})

	authURL := h.oidcClient.AuthCodeURL(state)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// HandleCallback processes OIDC redirects, exchanges credentials, and issues custom JWTs.
func (h *AuthHandlers) HandleCallback(w http.ResponseWriter, r *http.Request) {
	if h.oidcClient == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code": "BAD_REQUEST", "message": "OIDC client is not initialized"}`))
		return
	}

	stateCookie, err := r.Cookie("oauthstate")
	if err != nil || r.URL.Query().Get("state") != stateCookie.Value {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code": "UNAUTHORIZED", "message": "Invalid OAuth state parameter"}`))
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"code": "BAD_REQUEST", "message": "Authorization code is missing"}`))
		return
	}

	// Validate against OIDC Provider
	idToken, err := h.oidcClient.ExchangeAndVerify(context.Background(), code)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"code": "INTERNAL_ERROR", "message": "Failed to exchange authorization credentials"}`))
		return
	}

	var claims struct {
		Subject  string   `json:"sub"`
		Username string   `json:"preferred_username"`
		Groups   []string `json:"groups"`
	}
	if err := idToken.Claims(&claims); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"code": "INTERNAL_ERROR", "message": "Failed to parse security token details"}`))
		return
	}

	// Map OIDC groups to dashboard RBAC roles
	roles := []string{"viewer"}
	for _, g := range claims.Groups {
		if g == "uss-operator" || g == "operator" {
			roles = []string{"operator"}
			break
		} else if g == "uss-admin" || g == "admin" {
			roles = []string{"admin"}
			break
		}
	}

	username := claims.Username
	if username == "" {
		username = claims.Subject
	}

	userClaims := UserClaims{
		Username: username,
		Roles:    roles,
	}

	tokenString, err := SignJWT(userClaims, h.jwtSecret, 8*time.Hour)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"code": "INTERNAL_ERROR", "message": "Failed to sign authentication token"}`))
		return
	}

	// The browser reached this endpoint via a top-level navigation from the
	// IdP, not an XHR/fetch call - it must be redirected back to the SPA
	// with the token, not sent a JSON body nobody on the frontend reads.
	h.redirectToFrontend(w, r, tokenString, username, roles[0])
}
