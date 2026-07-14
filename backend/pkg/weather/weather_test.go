package weather

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleWeatherCheckStubFallback(t *testing.T) {
	handler := NewHandler("") // no API key -> deterministic stub

	// Case 1: Centroid inside safe weather threshold (lat <= 10.77)
	reqBody1, _ := json.Marshal(WeatherRequest{Lat: 10.762622, Lng: 106.660172})
	req1 := httptest.NewRequest("POST", "/api/operator/weather", bytes.NewBuffer(reqBody1))
	rr1 := httptest.NewRecorder()

	handler(rr1, req1)
	if rr1.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr1.Code)
	}

	var res1 WeatherResponse
	_ = json.Unmarshal(rr1.Body.Bytes(), &res1)
	if !res1.Safe {
		t.Error("Expected weather to be marked safe")
	}
	if res1.WindSpeed != 4.2 {
		t.Errorf("Expected wind speed 4.2, got %f", res1.WindSpeed)
	}

	// Case 2: Centroid inside dangerous weather threshold (lat > 10.77)
	reqBody2, _ := json.Marshal(WeatherRequest{Lat: 10.782622, Lng: 106.660172})
	req2 := httptest.NewRequest("POST", "/api/operator/weather", bytes.NewBuffer(reqBody2))
	rr2 := httptest.NewRecorder()

	handler(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr2.Code)
	}

	var res2 WeatherResponse
	_ = json.Unmarshal(rr2.Body.Bytes(), &res2)
	if res2.Safe {
		t.Error("Expected weather to be marked unsafe")
	}
	if res2.WindSpeed != 18.5 {
		t.Errorf("Expected wind speed 18.5, got %f", res2.WindSpeed)
	}
}

func TestHandleWeatherCheckLiveAPI(t *testing.T) {
	owm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"weather": [{"main": "Rain"}],
			"main": {"temp": 22.5},
			"wind": {"speed": 19.3}
		}`))
	}))
	defer owm.Close()

	client := NewClient("test-key")
	client.baseURL = owm.URL

	live, err := client.FetchCurrent(10.76, 106.66)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if live.WindSpeed != 19.3 {
		t.Errorf("expected wind speed 19.3, got %f", live.WindSpeed)
	}
	if live.Precipitation != "rain" {
		t.Errorf("expected precipitation 'rain', got %q", live.Precipitation)
	}
	if live.Safe {
		t.Error("expected unsafe: wind speed exceeds MaxSafeWindSpeedMS")
	}
}

func TestHandleWeatherCheckFallsBackToStubOnFetchError(t *testing.T) {
	owm := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer owm.Close()

	client := NewClient("test-key")
	client.baseURL = owm.URL
	handler := newHandler(client)

	reqBody, _ := json.Marshal(WeatherRequest{Lat: 10.762622, Lng: 106.660172})
	req := httptest.NewRequest("POST", "/api/operator/weather", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 with stub fallback even when the live fetch fails, got %d", rr.Code)
	}

	var res WeatherResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &res)
	if res.WindSpeed != 4.2 {
		t.Errorf("expected stub fallback wind speed 4.2, got %f", res.WindSpeed)
	}
}
