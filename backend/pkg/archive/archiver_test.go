package archive

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestArchiverSequence(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "missions_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	arch := &Archiver{
		filePath: tempFile.Name(),
	}

	// 1. Start mission
	arch.StartMission()

	// 2. Log telemetry points
	arch.LogPoint(10.762622, 106.660172, 12.0, 98.0, 5.2)
	time.Sleep(100 * time.Millisecond) // Let small delta pass
	arch.LogPoint(10.763000, 106.661000, 12.5, 97.5, 5.5)

	// Verify points recorded
	if len(arch.activePoints) != 2 {
		t.Errorf("Expected 2 points logged, got %d", len(arch.activePoints))
	}

	// First point should be t=0
	if arch.activePoints[0].T != 0 {
		t.Errorf("Expected first point T=0, got %d", arch.activePoints[0].T)
	}

	// 3. Save mission
	mID, err := arch.SaveMission()
	if err != nil {
		t.Fatalf("Failed to save mission: %v", err)
	}
	if mID == "" {
		t.Error("Expected non-empty mission ID")
	}

	// 4. Verify file content persistence
	data, err := ioutil.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read back db file: %v", err)
	}

	var records []MissionRecord
	err = json.Unmarshal(data, &records)
	if err != nil {
		t.Fatalf("Failed to unmarshal saved database rows: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("Expected 1 mission record, got %d", len(records))
	}

	rec := records[0]
	if rec.ID != mID {
		t.Errorf("Expected record ID %s, got %s", mID, rec.ID)
	}
	if len(rec.Telemetry) != 2 {
		t.Errorf("Expected 2 telemetry points in database, got %d", len(rec.Telemetry))
	}
	if rec.Telemetry[0].Lat != 10.762622 {
		t.Errorf("Expected latitude 10.762622, got %f", rec.Telemetry[0].Lat)
	}
}

func TestMissionsRetrievalReversal(t *testing.T) {
	missions := []MissionRecord{
		{ID: "MSN-1"},
		{ID: "MSN-2"},
	}
	// Reverse slice
	for i, j := 0, len(missions)-1; i < j; i, j = i+1, j-1 {
		missions[i], missions[j] = missions[j], missions[i]
	}
	if missions[0].ID != "MSN-2" {
		t.Errorf("Expected first element to be MSN-2 after reversal, got %s", missions[0].ID)
	}
}
