package main

import (
	"testing"
	"time"

	"uss-surveillance/backend/pkg/lease"
	"uss-surveillance/backend/pkg/suggestion"
)

func TestPaginationBoundsNoParamsReturnsFullRange(t *testing.T) {
	start, end := paginationBounds(5, "", "")
	if start != 0 || end != 5 {
		t.Errorf("expected [0,5), got [%d,%d)", start, end)
	}
}

func TestPaginationBoundsOffsetAndLimit(t *testing.T) {
	start, end := paginationBounds(10, "3", "2")
	if start != 3 || end != 5 {
		t.Errorf("expected [3,5), got [%d,%d)", start, end)
	}
}

func TestPaginationBoundsLimitClampedToTotal(t *testing.T) {
	start, end := paginationBounds(4, "2", "100")
	if start != 2 || end != 4 {
		t.Errorf("expected [2,4), got [%d,%d)", start, end)
	}
}

func TestPaginationBoundsOffsetBeyondTotal(t *testing.T) {
	start, end := paginationBounds(4, "50", "10")
	if start != 4 || end != 4 {
		t.Errorf("expected empty range [4,4), got [%d,%d)", start, end)
	}
}

func TestPaginationBoundsInvalidParamsIgnored(t *testing.T) {
	start, end := paginationBounds(5, "not-a-number", "also-not-a-number")
	if start != 0 || end != 5 {
		t.Errorf("expected fallback to full range [0,5), got [%d,%d)", start, end)
	}
}

func TestPaginationBoundsNegativeParamsIgnored(t *testing.T) {
	start, end := paginationBounds(5, "-1", "-1")
	if start != 0 || end != 5 {
		t.Errorf("expected fallback to full range [0,5) for negative params, got [%d,%d)", start, end)
	}
}

// resetGlobalDroneStateForTest restores globalDroneState to a clean,
// docked-and-ready baseline so applyFleetReadiness tests don't leak state
// into each other.
func resetGlobalDroneStateForTest() {
	globalDroneState.mu.Lock()
	globalDroneState.IsFlying = false
	globalDroneState.IsPaused = false
	globalDroneState.Battery = 92.0
	globalDroneState.mu.Unlock()
	lease.DefaultManager.ReleaseLease("Drone-01")
}

func TestApplyFleetReadinessAvailableWhenDockedAndCharged(t *testing.T) {
	resetGlobalDroneStateForTest()
	defer resetGlobalDroneStateForTest()

	res := &suggestion.SuggestionResponse{Success: true}
	applyFleetReadiness(res, "operator-dev")

	if !res.Success {
		t.Errorf("expected available drone to report success, got message %q", res.Message)
	}
}

func TestApplyFleetReadinessUnavailableWhileFlying(t *testing.T) {
	resetGlobalDroneStateForTest()
	defer resetGlobalDroneStateForTest()

	globalDroneState.mu.Lock()
	globalDroneState.IsFlying = true
	globalDroneState.mu.Unlock()

	res := &suggestion.SuggestionResponse{Success: true}
	applyFleetReadiness(res, "operator-dev")

	if res.Success {
		t.Error("expected an in-flight drone to be reported unavailable")
	}
	if res.Message == "" {
		t.Error("expected a message explaining why the drone is unavailable")
	}
}

func TestApplyFleetReadinessUnavailableWhenLockedByAnotherOperator(t *testing.T) {
	resetGlobalDroneStateForTest()
	defer resetGlobalDroneStateForTest()

	lease.DefaultManager.AcquireLease("someone-else", "Drone-01", 10*time.Second)

	res := &suggestion.SuggestionResponse{Success: true}
	applyFleetReadiness(res, "operator-dev")

	if res.Success {
		t.Error("expected a drone locked by another operator to be reported unavailable")
	}
}

func TestApplyFleetReadinessAvailableWhenLockedByRequestingOperator(t *testing.T) {
	resetGlobalDroneStateForTest()
	defer resetGlobalDroneStateForTest()

	// The requesting operator's own dashboard WebSocket connection
	// acquires this same lease automatically - it must not make their own
	// suggestion request look unavailable.
	lease.DefaultManager.AcquireLease("operator-dev", "Drone-01", 10*time.Second)

	res := &suggestion.SuggestionResponse{Success: true}
	applyFleetReadiness(res, "operator-dev")

	if !res.Success {
		t.Errorf("expected the requesting operator's own lease not to block their suggestion, got message %q", res.Message)
	}
}

func TestApplyFleetReadinessUnavailableWhenBatteryLow(t *testing.T) {
	resetGlobalDroneStateForTest()
	defer resetGlobalDroneStateForTest()

	globalDroneState.mu.Lock()
	globalDroneState.Battery = 15.0
	globalDroneState.mu.Unlock()

	res := &suggestion.SuggestionResponse{Success: true}
	applyFleetReadiness(res, "operator-dev")

	if res.Success {
		t.Error("expected a low-battery drone to be reported unavailable")
	}
}
