package archive

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TelemetryPoint represents a single logged telemetry frame.
type TelemetryPoint struct {
	T        int     `json:"t"` // t=0 relative index
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Altitude float64 `json:"altitude"`
	Battery  float64 `json:"battery"`
	Speed    float64 `json:"speed"`
}

// MissionRecord represents the persisted database row structure.
type MissionRecord struct {
	ID        string           `json:"id"`
	StartTime time.Time        `json:"start_time"`
	EndTime   time.Time        `json:"end_time"`
	Duration  float64          `json:"duration_sec"`
	VideoPath string           `json:"video_path"`
	Telemetry []TelemetryPoint `json:"telemetry"`
}

type Archiver struct {
	mu           sync.Mutex
	activePoints []TelemetryPoint
	takeoffTime  time.Time
	filePath     string
}

var DefaultArchiver = &Archiver{
	filePath: "data/missions.json",
}

// SetFilePath updates target database location.
func (a *Archiver) SetFilePath(path string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.filePath = path
}

// StartMission resets buffers and records start time.
func (a *Archiver) StartMission() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.activePoints = []TelemetryPoint{}
	a.takeoffTime = time.Now()
}

// LogPoint appends a new 1 Hz telemetry frame.
func (a *Archiver) LogPoint(lat, lng, alt, battery, speed float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.takeoffTime.IsZero() {
		a.takeoffTime = time.Now()
	}
	tSec := int(time.Since(a.takeoffTime).Seconds())
	a.activePoints = append(a.activePoints, TelemetryPoint{
		T:        tSec,
		Lat:      lat,
		Lng:      lng,
		Altitude: alt,
		Battery:  battery,
		Speed:    speed,
	})
}

// SaveMission persists the buffered history to JSON flat-file db.
func (a *Archiver) SaveMission() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.activePoints) == 0 {
		return "", fmt.Errorf("no active flight points to archive")
	}

	endTime := time.Now()
	durationSec := endTime.Sub(a.takeoffTime).Seconds()
	missionID := fmt.Sprintf("MSN-%d", a.takeoffTime.Unix())

	record := MissionRecord{
		ID:        missionID,
		StartTime: a.takeoffTime,
		EndTime:   endTime,
		Duration:  durationSec,
		VideoPath: fmt.Sprintf("/videos/%s.mp4", missionID),
		Telemetry: a.activePoints,
	}

	// Read existing missions database
	var missions []MissionRecord
	data, err := ioutil.ReadFile(a.filePath)
	if err == nil {
		_ = json.Unmarshal(data, &missions)
	}

	missions = append(missions, record)

	// Ensure output directory exists
	dir := filepath.Dir(a.filePath)
	_ = os.MkdirAll(dir, 0755)

	out, err := json.MarshalIndent(missions, "", "  ")
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(a.filePath, out, 0644)
	if err != nil {
		return "", err
	}

	// Reset buffers
	a.activePoints = []TelemetryPoint{}
	a.takeoffTime = time.Time{}

	return missionID, nil
}
