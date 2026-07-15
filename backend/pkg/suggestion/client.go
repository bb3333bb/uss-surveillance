package suggestion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SuggestionRequest models the centroid input payload.
type SuggestionRequest struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// SuggestionResponse models the recommendation payload.
type SuggestionResponse struct {
	RecommendedDrone string  `json:"recommended_drone"`
	RecommendedDock  string  `json:"recommended_dock"`
	DistanceMeters   float64 `json:"distance_meters"`
	Success          bool    `json:"success"`
	Message          string  `json:"message,omitempty"`
}

// Client wraps connections to the Python Suggestion Engine service.
type Client struct {
	endpoint   string
	httpClient *http.Client
}

// NewClient instantiates the Suggestion service client.
func NewClient(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetSuggestion retrieves recommendation mappings from the Python Suggestion Engine.
func (c *Client) GetSuggestion(lat, lng float64) (*SuggestionResponse, error) {
	reqBody, err := json.Marshal(SuggestionRequest{Lat: lat, Lng: lng})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal suggestion payload: %w", err)
	}

	resp, err := c.httpClient.Post(c.endpoint+"/api/suggest", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to contact recommendation service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("recommendation service returned code: %d", resp.StatusCode)
	}

	var res SuggestionResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode suggestion response: %w", err)
	}

	return &res, nil
}

// PlanRequestCoordinate maps lat/lng pairs inside REST queries.
type PlanRequestCoordinate struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// PlanRequest models the planning vertices query payload.
type PlanRequest struct {
	Vertices []PlanRequestCoordinate `json:"vertices"`
}

// PlanResponse models the path grid results.
type PlanResponse struct {
	Safe    bool                    `json:"safe"`
	Message string                  `json:"message"`
	Path    []PlanRequestCoordinate `json:"path"`
}

// GetPlan sends drawn vertices to the Python service to compute sweep plans.
func (c *Client) GetPlan(vertices []PlanRequestCoordinate) (*PlanResponse, error) {
	reqBody, err := json.Marshal(PlanRequest{Vertices: vertices})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal planner payload: %w", err)
	}

	resp, err := c.httpClient.Post(c.endpoint+"/api/plan", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to contact planner service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("planner service returned code: %d", resp.StatusCode)
	}

	var res PlanResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode planner response: %w", err)
	}

	return &res, nil
}
