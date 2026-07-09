package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthMiddleware(t *testing.T) {
	secret := "test-secret-key-that-is-long-enough-32-chars"
	claims := UserClaims{
		Username: "operator_sarah",
		Roles:    []string{"operator"},
	}

	tokenString, err := SignJWT(claims, secret, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxClaims := GetUserClaims(r.Context())
		if ctxClaims == nil {
			t.Fatal("Expected user claims to be injected into context")
		}
		if ctxClaims.Username != "operator_sarah" {
			t.Errorf("Expected username operator_sarah, got %s", ctxClaims.Username)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(secret)(nextHandler)

	// Case 1: Valid Bearer token
	req := httptest.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	// Case 2: Query param token (WebSocket fallback)
	req = httptest.NewRequest("GET", "/api/ws?token="+tokenString, nil)
	rr = httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	// Case 3: Missing token
	req = httptest.NewRequest("GET", "/api/protected", nil)
	rr = httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	// Case 4: Invalid signature
	req = httptest.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString+"bad")
	rr = httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestRequireRole(t *testing.T) {
	secret := "test-secret-key-that-is-long-enough-32-chars"
	
	operatorToken, _ := SignJWT(UserClaims{Username: "sarah", Roles: []string{"operator"}}, secret, 1*time.Hour)
	viewerToken, _ := SignJWT(UserClaims{Username: "visitor", Roles: []string{"viewer"}}, secret, 1*time.Hour)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	routeFlow := AuthMiddleware(secret)(RequireRole("operator")(nextHandler))

	// Case 1: Operator accesses (Allowed)
	req := httptest.NewRequest("POST", "/api/drone/takeoff", nil)
	req.Header.Set("Authorization", "Bearer "+operatorToken)
	rr := httptest.NewRecorder()
	routeFlow.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d for operator, got %d", http.StatusOK, rr.Code)
	}

	// Case 2: Viewer accesses (Forbidden)
	req = httptest.NewRequest("POST", "/api/drone/takeoff", nil)
	req.Header.Set("Authorization", "Bearer "+viewerToken)
	rr = httptest.NewRecorder()
	routeFlow.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status code %d for viewer, got %d", http.StatusForbidden, rr.Code)
	}
}
