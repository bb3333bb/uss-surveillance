package lease

import (
	"sync"
	"time"
)

// Lease represents a temporal ownership lease over a drone.
type Lease struct {
	Operator  string
	DroneID   string
	ExpiresAt time.Time
}

// Manager coordinates concurrent lease acquisitions thread-safely.
type Manager struct {
	mu     sync.Mutex
	leases map[string]Lease
}

// NewManager instantiates a lease broker.
func NewManager() *Manager {
	return &Manager{
		leases: make(map[string]Lease),
	}
}

// DefaultManager is the package-level shared lease provider.
var DefaultManager = NewManager()

// AcquireLease attempts to lock or renew command access to a drone.
func (m *Manager) AcquireLease(operator string, droneID string, duration time.Duration) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	current, exists := m.leases[droneID]

	// If an unexpired lease belongs to another operator, deny allocation
	if exists && current.ExpiresAt.After(now) && current.Operator != operator {
		return false
	}

	m.leases[droneID] = Lease{
		Operator:  operator,
		DroneID:   droneID,
		ExpiresAt: now.Add(duration),
	}
	return true
}

// GetLeaseHolder returns the current operator owning the drone control lock.
func (m *Manager) GetLeaseHolder(droneID string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	current, exists := m.leases[droneID]
	if exists && current.ExpiresAt.After(time.Now()) {
		return current.Operator, true
	}
	return "", false
}

// ReleaseLease forces eviction of the current control lock lease.
func (m *Manager) ReleaseLease(droneID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.leases, droneID)
}
