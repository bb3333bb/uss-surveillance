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
	return newHandler(client)
}

// newHandler builds the handler around an already-constructed client (or
// nil for stub-only mode), so tests can point it at a mock server.
func newHandler(client *Client) http.HandlerFunc {
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

		res := stubWeather(req.Lat)
		if client != nil {
			live, err := client.FetchCurrent(req.Lat, req.Lng)
			if err != nil {
				log.Printf("OpenWeatherMap fetch failed, falling back to stub: %v", err)
			} else {
				res = *live
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(res)
	}
}

// stubWeather is the deterministic dev/CI fallback (AD-6 coordinate rules
// apply): simulates a high-wind storm north of lat 10.77.
func stubWeather(lat float64) WeatherResponse {
	if lat > 10.77 {
		return WeatherResponse{WindSpeed: 18.5, Precipitation: "heavy rain", Temp: 25.4, Safe: false}
	}
	return WeatherResponse{WindSpeed: 4.2, Precipitation: "none", Temp: 31.2, Safe: true}
}
