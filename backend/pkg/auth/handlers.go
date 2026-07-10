package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"
)

// AuthHandlers aggregates OIDC clients and JWT configurations.
type AuthHandlers struct {
	oidcClient *OIDCClient
	jwtSecret  string
}

// NewAuthHandlers instantiates OIDC and JWT HTTP handler functions.
func NewAuthHandlers(oidcClient *OIDCClient, jwtSecret string) *AuthHandlers {
	return &AuthHandlers{
		oidcClient: oidcClient,
		jwtSecret:  jwtSecret,
	}
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
		
		// Redirect back to frontend dev server with query parameters
		redirectURL := "http://localhost:5173/?token=" + tokenString + "&user=" + userClaims.Username + "&role=" + userClaims.Roles[0]
		http.Redirect(w, r, redirectURL, http.StatusFound)
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

	// Return token response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
		"role":  roles[0],
		"user":  username,
	})
}
