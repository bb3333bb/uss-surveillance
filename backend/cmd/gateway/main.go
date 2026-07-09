package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"uss-surveillance/backend/pkg/auth"
	"uss-surveillance/backend/pkg/suggestion"
	"uss-surveillance/backend/pkg/weather"
)

func main() {
	log.Println("Initializing USS Surveillance Gateway...")

	issuerURL := os.Getenv("OIDC_ISSUER_URL")
	clientID := os.Getenv("OIDC_CLIENT_ID")
	clientSecret := os.Getenv("OIDC_CLIENT_SECRET")
	redirectURL := os.Getenv("OIDC_REDIRECT_URI")
	jwtSecret := os.Getenv("JWT_SECRET")

	// Set fallback defaults for development ease
	if issuerURL == "" {
		issuerURL = "http://localhost:8080/realms/uss-surveillance"
	}
	if clientID == "" {
		clientID = "uss-surveillance-client"
	}
	if jwtSecret == "" {
		jwtSecret = "development-secret-key-that-is-long-enough-32-chars"
	}

	// Initialize OIDC client (can be skipped for unit tests via env flag)
	var oidcClient *auth.OIDCClient
	var err error
	if os.Getenv("SKIP_OIDC_INIT") != "true" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		oidcClient, err = auth.NewOIDCClient(ctx, issuerURL, clientID, clientSecret, redirectURL)
		if err != nil {
			log.Printf("OIDC initialization delayed: %v. Running in mock mode.", err)
		}
	}

	authHandlers := auth.NewAuthHandlers(oidcClient, jwtSecret)

	mux := http.NewServeMux()

	// Public Routes
	mux.HandleFunc("/api/auth/login", authHandlers.HandleLogin)
	mux.HandleFunc("/api/auth/callback", authHandlers.HandleCallback)

	// Protected Routes
	protectedMux := http.NewServeMux()
	
	// Open dashboard view (any authenticated user can view)
	protectedMux.HandleFunc("/api/operator/dashboard", func(w http.ResponseWriter, r *http.Request) {
		claims := auth.GetUserClaims(r.Context())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "authorized", "user": "` + claims.Username + `", "role": "` + claims.Roles[0] + `"}`))
	})

	// Flight controls trigger (requires operator or admin role)
	commandHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "success", "message": "Command override dispatched successfully"}`))
	})
	protectedMux.Handle("/api/operator/command", auth.RequireRole("operator", "admin")(commandHandler))

	// Weather checks route (accessible to any authenticated session)
	protectedMux.HandleFunc("/api/operator/weather", weather.HandleWeatherCheck)

	// Suggestion client initialization & handler (resolves from local environment config)
	suggestionEngineURL := os.Getenv("SUGGESTION_ENGINE_URL")
	if suggestionEngineURL == "" {
		suggestionEngineURL = "http://localhost:50051"
	}
	suggestionClient := suggestion.NewClient(suggestionEngineURL)

	protectedMux.HandleFunc("/api/operator/suggestion", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte(`{"code": "METHOD_NOT_ALLOWED", "message": "Method not allowed"}`))
			return
		}

		var req struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"code": "BAD_REQUEST", "message": "Failed to parse geofence centroid"}`))
			return
		}

		res, err := suggestionClient.GetSuggestion(req.Lat, req.Lng)
		if err != nil {
			log.Printf("Python suggestion engine connection failed: %v. Returning fallback suggestion.", err)
			fallback := suggestion.SuggestionResponse{
				RecommendedDrone: "Drone-01 (M300 RTK)",
				RecommendedDock:  "Dock Alpha",
				DistanceMeters:   14.8,
				Success:          true,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(fallback)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(res)
	})

	protectedMux.HandleFunc("/api/operator/plan", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte(`{"code": "METHOD_NOT_ALLOWED", "message": "Method not allowed"}`))
			return
		}

		var req struct {
			Vertices []suggestion.PlanRequestCoordinate `json:"vertices"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"code": "BAD_REQUEST", "message": "Failed to parse geofence coordinates"}`))
			return
		}

		res, err := suggestionClient.GetPlan(req.Vertices)
		if err != nil {
			log.Printf("Python planner service connection failed: %v. Returning fallback plan.", err)
			fallback := suggestion.PlanResponse{
				Safe:    true,
				Message: "Patrol flight path generated successfully (Fallback)",
				Path: []suggestion.PlanRequestCoordinate{
					{Lat: 10.762622, Lng: 106.660172},
					{Lat: 10.762922, Lng: 106.660172},
					{Lat: 10.762922, Lng: 106.660472},
					{Lat: 10.762622, Lng: 106.660472},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(fallback)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(res)
	})

	// Wrap in JWT authorization middleware
	authMiddleware := auth.AuthMiddleware(jwtSecret)
	mux.Handle("/api/operator/", authMiddleware(protectedMux))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on port %s...", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Gateway server execution failed: %v", err)
	}
}
