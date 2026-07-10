package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"math"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"uss-surveillance/backend/pkg/archive"
	"uss-surveillance/backend/pkg/auth"
	"uss-surveillance/backend/pkg/lease"
	"uss-surveillance/backend/pkg/mqtt"
	"uss-surveillance/backend/pkg/suggestion"
	"uss-surveillance/backend/pkg/weather"
)

// DroneTelemetryState tracks the live telemetry variables of the active drone.
type DroneTelemetryState struct {
	mu           sync.Mutex
	IsFlying     bool
	Path         []suggestion.PlanRequestCoordinate
	CurrentIndex int
	Battery      float64
	Altitude     float64
	Speed        float64
}

var globalDroneState = &DroneTelemetryState{
	Battery:  92.0,
	Altitude: 0.0,
	Speed:    0.0,
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var globalHubDoorsState string = "closed"
var globalHubDoorsMutex sync.Mutex

func setHubDoorsState(state string) {
	globalHubDoorsMutex.Lock()
	globalHubDoorsState = state
	globalHubDoorsMutex.Unlock()
}

func getHubDoorsState() string {
	globalHubDoorsMutex.Lock()
	defer globalHubDoorsMutex.Unlock()
	return globalHubDoorsState
}

func main() {
	log.Println("Initializing USS Surveillance Gateway...")

	// Initialize mechanical Drone Hub simulator
	mqtt.StartDroneHubSimulator(mqtt.DefaultClient)

	// Background ticker advancing drone coordinates step-by-step along active plans (1 Hz)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			globalDroneState.mu.Lock()
			if globalDroneState.IsFlying {
				if len(globalDroneState.Path) > 0 {
					if globalDroneState.CurrentIndex < len(globalDroneState.Path)-1 {
						globalDroneState.CurrentIndex++
						globalDroneState.Battery -= 0.5
					} else {
						// Final coordinate reached - check precision match (AD-6) and trigger interlocks
						lastCoord := globalDroneState.Path[globalDroneState.CurrentIndex]
						diffLat := math.Abs(lastCoord.Lat - 10.762622)
						diffLng := math.Abs(lastCoord.Lng - 106.660172)
						if diffLat < 1e-7 && diffLng < 1e-7 {
							globalDroneState.IsFlying = false
							_, errArch := archive.DefaultArchiver.SaveMission()
							if errArch != nil {
								log.Printf("Archiver error on dock: %v", errArch)
							}
							
							// Trigger mechanical closures
							go func() {
								setHubDoorsState("opening")
								time.Sleep(1 * time.Second)
								setHubDoorsState("open")
								time.Sleep(1 * time.Second)
								setHubDoorsState("closing")
								time.Sleep(1 * time.Second)
								setHubDoorsState("recharging")
							}()
						} else {
							globalDroneState.IsFlying = false
							_, errArch := archive.DefaultArchiver.SaveMission()
							if errArch != nil {
								log.Printf("Archiver error on landing: %v", errArch)
							}
						}
					}
				}
				globalDroneState.Altitude = 12.0
				globalDroneState.Speed = 5.2

				// Log point
				var curLat, curLng float64
				if len(globalDroneState.Path) > 0 {
					pt := globalDroneState.Path[globalDroneState.CurrentIndex]
					curLat, curLng = pt.Lat, pt.Lng
				} else {
					curLat, curLng = 10.762622, 106.660172
				}
				archive.DefaultArchiver.LogPoint(curLat, curLng, globalDroneState.Altitude, globalDroneState.Battery, globalDroneState.Speed)
			} else {
				// Docked battery replenishment ticks
				if getHubDoorsState() == "recharging" {
					if globalDroneState.Battery < 100.0 {
						globalDroneState.Battery += 1.5
						if globalDroneState.Battery > 100.0 {
							globalDroneState.Battery = 100.0
						}
					}
				}
				globalDroneState.Altitude = 0.0
				globalDroneState.Speed = 0.0
			}
			globalDroneState.mu.Unlock()
		}
	}()

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

	// Command validation handler (requires operator or admin role)
	commandHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte(`{"code": "METHOD_NOT_ALLOWED", "message": "Method not allowed"}`))
			return
		}

		var req struct {
			Command string `json:"command"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"code": "BAD_REQUEST", "message": "Failed to parse command payload"}`))
			return
		}

		claims := auth.GetUserClaims(r.Context())
		operatorName := claims.Username
		if operatorName == "" {
			operatorName = "anonymous"
		}

		// Verify operator control lease
		leaseholder, hasLease := lease.DefaultManager.GetLeaseHolder("Drone-01")
		if hasLease && leaseholder != operatorName {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"code": "FORBIDDEN", "message": "Exclusive command lock held by operator: ` + leaseholder + `"}`))
			return
		}

		// Auto acquire lease if unheld
		if !hasLease {
			lease.DefaultManager.AcquireLease(operatorName, "Drone-01", 10*time.Second)
		}

		// Publish command to MQTT
		var mqttCmd string
		if req.Command == "pause" {
			mqttCmd = "hover"
		} else if req.Command == "rth" {
			mqttCmd = "rth"
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"code": "BAD_REQUEST", "message": "Unsupported manual override command"}`))
			return
		}

		errPub := mqtt.DefaultClient.Publish("drone/hub/command", mqttCmd)
		if errPub != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"code": "INTERNAL_SERVER_ERROR", "message": "Failed to publish override command"}`))
			return
		}

		// Update active simulation states
		globalDroneState.mu.Lock()
		if req.Command == "rth" {
			globalDroneState.IsFlying = false
			_, errArch := archive.DefaultArchiver.SaveMission()
			if errArch != nil {
				log.Printf("Archiver error on RTH: %v", errArch)
			}
			go func() {
				setHubDoorsState("opening")
				time.Sleep(1 * time.Second)
				setHubDoorsState("open")
				time.Sleep(1 * time.Second)
				setHubDoorsState("closing")
				time.Sleep(1 * time.Second)
				setHubDoorsState("recharging")
			}()
		}
		globalDroneState.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "success", "message": "Override command '" + req.Command + "' dispatched successfully"}`))
	})
	protectedMux.Handle("/api/operator/command", auth.RequireRole("operator", "admin")(commandHandler))

	// Weather checks route (accessible to any authenticated session)
	protectedMux.HandleFunc("/api/operator/weather", weather.HandleWeatherCheck)

	// Historical missions log retrieval endpoint (requires operator or admin role)
	missionsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte(`{"code": "METHOD_NOT_ALLOWED", "message": "Method not allowed"}`))
			return
		}

		data, err := os.ReadFile("data/missions.json")
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[]`))
			return
		}

		var missions []archive.MissionRecord
		if err := json.Unmarshal(data, &missions); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"code": "INTERNAL_SERVER_ERROR", "message": "Failed to parse records"}`))
			return
		}

		for i, j := 0, len(missions)-1; i < j; i, j = i+1, j-1 {
			missions[i], missions[j] = missions[j], missions[i]
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(missions)
	})
	protectedMux.Handle("/api/operator/missions", auth.RequireRole("operator", "admin")(missionsHandler))

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
			globalDroneState.mu.Lock()
			globalDroneState.Path = fallback.Path
			globalDroneState.mu.Unlock()

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(fallback)
			return
		}

		globalDroneState.mu.Lock()
		globalDroneState.Path = res.Path
		globalDroneState.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(res)
	})

	// Mechanical Takeoff Interlock Launch Orchestrator (requires operator or admin role)
	launchHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte(`{"code": "METHOD_NOT_ALLOWED", "message": "Method not allowed"}`))
			return
		}

		success, msg := mqtt.RunLaunchSequence(mqtt.DefaultClient)
		if !success {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"status": "error", "message": "` + msg + `"}`))
			return
		}

		globalDroneState.mu.Lock()
		globalDroneState.IsFlying = true
		globalDroneState.CurrentIndex = 0
		globalDroneState.Battery = 98.0
		archive.DefaultArchiver.StartMission()
		globalDroneState.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "launched", "message": "` + msg + `"}`))
	})
	protectedMux.Handle("/api/operator/launch", auth.RequireRole("operator", "admin")(launchHandler))



	// WebSocket 1 Hz Telemetry & Mutex Control Lease Handler
	protectedMux.HandleFunc("/api/operator/telemetry", func(w http.ResponseWriter, r *http.Request) {
		claims := auth.GetUserClaims(r.Context())
		operatorName := claims.Username
		if operatorName == "" {
			operatorName = "anonymous"
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		// Secure exclusive control lease for 10 seconds initially
		acquired := lease.DefaultManager.AcquireLease(operatorName, "Drone-01", 10*time.Second)
		if !acquired {
			log.Printf("[Lease block] Operator %s denied exclusive control lease.", operatorName)
		}

		// Broadcaster loop (1 Hz)
		stopChan := make(chan bool)
		defer close(stopChan)

		go func() {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					globalDroneState.mu.Lock()
					var lat, lng float64
					if len(globalDroneState.Path) > 0 && globalDroneState.IsFlying {
						pt := globalDroneState.Path[globalDroneState.CurrentIndex]
						lat, lng = pt.Lat, pt.Lng
					} else {
						lat, lng = 10.762622, 106.660172
					}

					leaseholder, _ := lease.DefaultManager.GetLeaseHolder("Drone-01")

					// Build real-time safety alert arrays (AC-2)
					var alerts []string
					if leaseholder != "" && leaseholder != operatorName {
						alerts = append(alerts, "LEASE_CONFLICT")
					}
					if globalDroneState.IsFlying {
						if lat > 10.77 {
							alerts = append(alerts, "WEATHER_BREACH_WIND", "WEATHER_BREACH_RAIN")
						}
						// Restricted airspace (TSN airport NFZ geofence) checking distance
						dx := lat - 10.7725
						dy := lng - 106.69
						dist := math.Sqrt(dx*dx + dy*dy)
						if dist < 0.0072 { // ~800m threshold radius
							alerts = append(alerts, "RESTRICTED_AIRSPACE")
						}
					}

					payload := struct {
						Lat         float64  `json:"lat"`
						Lng         float64  `json:"lng"`
						Altitude    float64  `json:"altitude"`
						Battery     float64  `json:"battery"`
						Speed       float64  `json:"speed"`
						IsFlying    bool     `json:"is_flying"`
						Leaseholder string   `json:"leaseholder"`
						HubDoors    string   `json:"hub_doors"`
						Alerts      []string `json:"alerts"`
					}{
						Lat:         lat,
						Lng:         lng,
						Altitude:    globalDroneState.Altitude,
						Battery:     globalDroneState.Battery,
						Speed:       globalDroneState.Speed,
						IsFlying:    globalDroneState.IsFlying,
						Leaseholder: leaseholder,
						HubDoors:    getHubDoorsState(),
						Alerts:      alerts,
					}
					globalDroneState.mu.Unlock()

					if err := conn.WriteJSON(payload); err != nil {
						return
					}
				case <-stopChan:
					return
				}
			}
		}()

		// Heartbeat watchdog (AD-5) channel
		heartbeatChan := make(chan bool, 10)

		go func() {
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					break
				}
				var msg struct {
					Type string `json:"type"`
				}
				if err := json.Unmarshal(message, &msg); err == nil && msg.Type == "ping" {
					// Renew lease
					lease.DefaultManager.AcquireLease(operatorName, "Drone-01", 10*time.Second)
					select {
					case heartbeatChan <- true:
					default:
					}
				}
			}
			close(heartbeatChan)
		}()

		for {
			select {
			case active, ok := <-heartbeatChan:
				if !ok {
					log.Printf("[WebSocket] Connection closed for operator %s. Releasing control lease.", operatorName)
					leaseholder, _ := lease.DefaultManager.GetLeaseHolder("Drone-01")
					if leaseholder == operatorName {
						lease.DefaultManager.ReleaseLease("Drone-01")
					}
					return
				}
				_ = active // Heartbeat received successfully
			case <-time.After(10 * time.Second):
				// AD-5 Watchdog timeout!
				leaseholder, _ := lease.DefaultManager.GetLeaseHolder("Drone-01")
				if leaseholder == operatorName {
					log.Printf("[Safety Watchdog] Heartbeat timeout for operator %s. Issuing drone HOVER and releasing lease.", operatorName)
					lease.DefaultManager.ReleaseLease("Drone-01")
					_ = mqtt.DefaultClient.Publish("drone/hub/command", "hover")

					globalDroneState.mu.Lock()
					globalDroneState.IsFlying = false
					globalDroneState.mu.Unlock()
				}
				return
			}
		}
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
