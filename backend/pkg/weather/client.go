package weather

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// MaxSafeWindSpeedMS is the FR-4 safety threshold: mission confirmation is
// blocked when wind speed exceeds this value.
const MaxSafeWindSpeedMS = 15.0

const owmBaseURL = "https://api.openweathermap.org/data/2.5/weather"

// Client wraps the OpenWeatherMap Current Weather Data API.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient constructs a client for the given OpenWeatherMap API key.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		baseURL:    owmBaseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

type owmResponse struct {
	Weather []struct {
		Main string `json:"main"`
	} `json:"weather"`
	Main struct {
		Temp float64 `json:"temp"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
}

// FetchCurrent retrieves live wind/temperature/precipitation data for the
// given WGS84 coordinates.
func (c *Client) FetchCurrent(lat, lng float64) (*WeatherResponse, error) {
	query := url.Values{
		"lat":   {strconv.FormatFloat(lat, 'f', 7, 64)},
		"lon":   {strconv.FormatFloat(lng, 'f', 7, 64)},
		"appid": {c.apiKey},
		"units": {"metric"},
	}

	resp, err := c.httpClient.Get(c.baseURL + "?" + query.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to contact OpenWeatherMap: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenWeatherMap returned status %d", resp.StatusCode)
	}

	var owm owmResponse
	if err := json.NewDecoder(resp.Body).Decode(&owm); err != nil {
		return nil, fmt.Errorf("failed to decode OpenWeatherMap response: %w", err)
	}

	precipitation := "none"
	if len(owm.Weather) > 0 {
		switch owm.Weather[0].Main {
		case "Rain", "Drizzle", "Thunderstorm":
			precipitation = "rain"
		case "Snow":
			precipitation = "snow"
		}
	}

	return &WeatherResponse{
		WindSpeed:     owm.Wind.Speed,
		Precipitation: precipitation,
		Temp:          owm.Main.Temp,
		Safe:          owm.Wind.Speed <= MaxSafeWindSpeedMS,
	}, nil
}
