package suggestion

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSuggestion(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/suggest" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		res := SuggestionResponse{
			RecommendedDrone: "Drone-01 (M300 RTK)",
			RecommendedDock:  "Dock Alpha",
			DistanceMeters:   14.8,
			Success:          true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(res)
	}))
	defer mockServer.Close()

	client := NewClient(mockServer.URL)
	res, err := client.GetSuggestion(10.762622, 106.660172)
	if err != nil {
		t.Fatalf("Failed to fetch recommendation suggestion: %v", err)
	}

	if !res.Success {
		t.Error("Expected successful suggestion allocation status")
	}

	if res.RecommendedDrone != "Drone-01 (M300 RTK)" {
		t.Errorf("Expected Drone-01 (M300 RTK), got %s", res.RecommendedDrone)
	}

	if res.RecommendedDock != "Dock Alpha" {
		t.Errorf("Expected Dock Alpha, got %s", res.RecommendedDock)
	}
}

func TestGetPlan(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/plan" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		res := PlanResponse{
			Safe:    true,
			Message: "Patrol flight path generated successfully",
			Path: []PlanRequestCoordinate{
				{Lat: 10.762622, Lng: 106.660172},
				{Lat: 10.762922, Lng: 106.660172},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(res)
	}))
	defer mockServer.Close()

	client := NewClient(mockServer.URL)
	vertices := []PlanRequestCoordinate{
		{Lat: 10.762, Lng: 106.660},
		{Lat: 10.763, Lng: 106.660},
		{Lat: 10.763, Lng: 106.661},
	}
	res, err := client.GetPlan(vertices)
	if err != nil {
		t.Fatalf("Failed to fetch plan suggestion: %v", err)
	}

	if !res.Safe {
		t.Error("Expected plan to be safe")
	}

	if len(res.Path) != 2 {
		t.Errorf("Expected path length 2, got %d", len(res.Path))
	}
}
