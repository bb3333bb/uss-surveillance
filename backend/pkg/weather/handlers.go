package weather

import (
	"encoding/json"
	"log"
	"net/http"
)

// WeatherRequest holds the coordinates payload for weather forecast analysis.
type WeatherRequest struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// WeatherResponse models the assessed safety state.
type WeatherResponse struct {
	WindSpeed     float64 `json:"wind_speed"`
	Precipitation string  `json:"precipitation"`
	Temp          float64 `json:"temp"`
	Safe          bool    `json:"safe"`
}

// NewHandler returns the FR-4 weather-check HTTP handler. When apiKey is
// non-empty it queries OpenWeatherMap for live data; otherwise (and on any
// live-fetch error) it falls back to a deterministic stub - the same
// behavior this replaces, kept as the dev/CI default so a missing key
// degrades gracefully instead of breaking mission planning.
func NewHandler(apiKey string) http.HandlerFunc {
	var client *Client
	if apiKey != "" {
		client = NewClient(apiKey)
	}
	return NewHandlerWithClient(client)
}

// NewHandlerWithClient builds the handler around an already-constructed
// client (or nil for stub-only mode), so callers that need the same client
// elsewhere (e.g. the gateway's in-flight weather polling) or tests
// pointing at a mock server don't have to construct a second one.
func NewHandlerWithClient(client *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte(`{"code": "METHOD_NOT_ALLOWED", "message": "Method not allowed"}`))
			return
		}

		var req WeatherRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"code": "BAD_REQUEST", "message": "Failed to parse coordinate points"}`))
			return
		}

		res := CheckSafety(client, req.Lat, req.Lng)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(res)
	}
}

// CheckSafety returns live weather data via client when non-nil, falling
// back to the deterministic stub on a nil client or a live-fetch error.
// Exported so callers outside an HTTP handler (e.g. the gateway's in-flight
// telemetry loop, which needs the same real-vs-stub behavior to raise
// WEATHER_BREACH alerts) can reuse the exact same logic.
func CheckSafety(client *Client, lat, lng float64) WeatherResponse {
	if client != nil {
		if live, err := client.FetchCurrent(lat, lng); err != nil {
			log.Printf("OpenWeatherMap fetch failed, falling back to stub: %v", err)
		} else {
			return *live
		}
	}
	return stubWeather(lat)
}

// stubWeather is the deterministic dev/CI fallback (AD-6 coordinate rules
// apply): simulates a high-wind storm north of lat 10.77.
func stubWeather(lat float64) WeatherResponse {
	if lat > 10.77 {
		return WeatherResponse{WindSpeed: 18.5, Precipitation: "heavy rain", Temp: 25.4, Safe: false}
	}
	return WeatherResponse{WindSpeed: 4.2, Precipitation: "none", Temp: 31.2, Safe: true}
}
