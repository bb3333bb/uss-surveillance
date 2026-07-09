package weather

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleWeatherCheck(t *testing.T) {
	// Case 1: Centroid inside safe weather threshold (lat <= 10.77)
	reqBody1, _ := json.Marshal(WeatherRequest{Lat: 10.762622, Lng: 106.660172})
	req1 := httptest.NewRequest("POST", "/api/operator/weather", bytes.NewBuffer(reqBody1))
	rr1 := httptest.NewRecorder()

	HandleWeatherCheck(rr1, req1)
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

	HandleWeatherCheck(rr2, req2)
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
