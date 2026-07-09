package auth

import (
	"testing"
	"time"
)

func TestSignAndVerifyJWT(t *testing.T) {
	secret := "test-secret-key-that-is-long-enough-32-chars"
	claims := UserClaims{
		Username: "operator_sarah",
		Roles:    []string{"operator"},
	}

	// 1. Sign token (This should fail to compile initially until UserClaims and SignJWT are defined)
	tokenString, err := SignJWT(claims, secret, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to sign JWT: %v", err)
	}

	if tokenString == "" {
		t.Fatal("Expected token string to be non-empty")
	}

	// 2. Verify token
	parsedClaims, err := VerifyJWT(tokenString, secret)
	if err != nil {
		t.Fatalf("Failed to verify JWT: %v", err)
	}

	if parsedClaims.Username != claims.Username {
		t.Errorf("Expected username %s, got %s", claims.Username, parsedClaims.Username)
	}

	if len(parsedClaims.Roles) != 1 || parsedClaims.Roles[0] != "operator" {
		t.Errorf("Expected roles ['operator'], got %v", parsedClaims.Roles)
	}
}

func TestVerifyExpiredJWT(t *testing.T) {
	secret := "test-secret-key-that-is-long-enough-32-chars"
	claims := UserClaims{
		Username: "operator_sarah",
		Roles:    []string{"operator"},
	}

	// Sign with negative duration to simulate immediate expiry
	tokenString, err := SignJWT(claims, secret, -1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to sign JWT: %v", err)
	}

	_, err = VerifyJWT(tokenString, secret)
	if err == nil {
		t.Error("Expected error when verifying expired JWT, got nil")
	}
}
