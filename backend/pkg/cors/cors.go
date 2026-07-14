// Package cors provides a small CORS middleware for the gateway.
//
// The frontend (Vite dev server, or any deployed origin different from the
// gateway's own) sends a preflight OPTIONS request before any cross-origin
// GET/POST call that carries an Authorization header or a JSON body -
// including every /api/operator/* call the operator dashboard makes.
// Browsers never attach custom headers (like Authorization) to preflight
// requests, so if the preflight itself has to pass the JWT auth middleware,
// it always fails with 401 and the browser aborts the real request before
// it's even sent. This middleware must run outside/before AuthMiddleware
// so OPTIONS requests are answered without needing a token.
package cors

import "net/http"

// Middleware answers CORS preflight requests directly and adds the
// necessary headers to every response so allowedOrigin can call the API.
func Middleware(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
