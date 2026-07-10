package mqtt

import (
	"testing"
)

func TestRunLaunchSequence(t *testing.T) {
	client := NewMockClient()
	StartDroneHubSimulator(client)

	success, msg := RunLaunchSequence(client)
	if !success {
		t.Fatalf("Expected launch sequence to succeed, failed: %s", msg)
	}

	if msg != "Drone takeoff sequence completed successfully" {
		t.Errorf("Expected success confirmation message, got: %s", msg)
	}
}
