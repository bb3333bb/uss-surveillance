package lease

import (
	"sync"
	"time"
)

// Manager coordinates exclusive control leases over drones. Implementations
// must guarantee that AcquireLease is safe under concurrent callers - it
// backs the mutex gating manual override commands (FR-11) and the
// telemetry WebSocket's heartbeat watchdog.
type Manager interface {
	AcquireLease(operator string, droneID string, duration time.Duration) bool
	GetLeaseHolder(droneID string) (string, bool)
	ReleaseLease(droneID string)
}

type leaseEntry struct {
	Operator  string
	ExpiresAt time.Time
}

// InMemoryManager is a single-process lease broker. It's used for local
// dev/CI and as the automatic fallback when REDIS_URL is unset - safe only
// for a single gateway instance. See RedisManager for multi-instance
// deployments.
type InMemoryManager struct {
	mu     sync.Mutex
	leases map[string]leaseEntry
}

// NewManager instantiates an in-memory lease broker.
func NewManager() *InMemoryManager {
	return &InMemoryManager{
		leases: make(map[string]leaseEntry),
	}
}

// AcquireLease attempts to lock or renew command access to a drone.
func (m *InMemoryManager) AcquireLease(operator string, droneID string, duration time.Duration) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	current, exists := m.leases[droneID]

	// If an unexpired lease belongs to another operator, deny allocation
	if exists && current.ExpiresAt.After(now) && current.Operator != operator {
		return false
	}

	m.leases[droneID] = leaseEntry{
		Operator:  operator,
		ExpiresAt: now.Add(duration),
	}
	return true
}

// GetLeaseHolder returns the current operator owning the drone control lock.
func (m *InMemoryManager) GetLeaseHolder(droneID string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	current, exists := m.leases[droneID]
	if exists && current.ExpiresAt.After(time.Now()) {
		return current.Operator, true
	}
	return "", false
}

// ReleaseLease forces eviction of the current control lock lease.
func (m *InMemoryManager) ReleaseLease(droneID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.leases, droneID)
}

// DefaultManager is the package-level shared lease provider, selected via
// NewFromEnv: Redis-backed when REDIS_URL is set (required for any
// multi-instance gateway deployment), in-memory otherwise.
var DefaultManager Manager = NewFromEnv()
