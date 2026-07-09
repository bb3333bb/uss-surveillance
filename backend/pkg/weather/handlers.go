package weather

import (
	"encoding/json"
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

// HandleWeatherCheck computes deterministic forecast warnings based on geofence coordinates.
func HandleWeatherCheck(w http.ResponseWriter, r *http.Request) {
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

	// Deterministic weather thresholds for developer simulation (AD-6 coordinate rules apply)
	windSpeed := 4.2
	precipitation := "none"
	temp := 31.2
	safe := true

	// Simulate high wind storm for geofence centroids north of 10.77
	if req.Lat > 10.77 {
		windSpeed = 18.5
		precipitation = "heavy rain"
		temp = 25.4
		safe = false
	}

	res := WeatherResponse{
		WindSpeed:     windSpeed,
		Precipitation: precipitation,
		Temp:          temp,
		Safe:          safe,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(res)
}
