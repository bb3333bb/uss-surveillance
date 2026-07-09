package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

// UserClaimsKey is the context key for injected user claims.
const UserClaimsKey contextKey = "user_claims"

// AuthMiddleware intercepts HTTP requests to validate OIDC-signed JWTs.
func AuthMiddleware(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Upgrade WebSocket connections may contain token in query params
			tokenStr := ""
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					tokenStr = parts[1]
				}
			}

			// Fallback: URL query parameter (for WebSocket handshakes)
			if tokenStr == "" {
				tokenStr = r.URL.Query().Get("token")
			}

			if tokenStr == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"code": "UNAUTHORIZED", "message": "Missing authentication credentials"}`))
				return
			}

			claims, err := VerifyJWT(tokenStr, secretKey)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"code": "UNAUTHORIZED", "message": "Invalid or expired session token"}`))
				return
			}

			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserClaims helper retrieves injected claims from context.
func GetUserClaims(ctx context.Context) *UserClaims {
	claims, ok := ctx.Value(UserClaimsKey).(*UserClaims)
	if !ok {
		return nil
	}
	return claims
}

// RequireRole verifies that the user holds at least one of the permitted role scopes.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserClaims(r.Context())
			if claims == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"code": "UNAUTHORIZED", "message": "Authentication required"}`))
				return
			}

			authorized := false
			for _, allowed := range allowedRoles {
				for _, userRole := range claims.Roles {
					if strings.ToLower(userRole) == strings.ToLower(allowed) {
						authorized = true
						break
					}
				}
				if authorized {
					break
				}
			}

			if !authorized {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"code": "FORBIDDEN", "message": "Access denied for this role"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
