package lease

import (
	"testing"
	"time"
)

func TestAcquireLease(t *testing.T) {
	m := NewManager()

	// Initial lease acquisition
	success := m.AcquireLease("alice", "Drone-01", 50*time.Millisecond)
	if !success {
		t.Fatal("Expected alice to acquire control lease successfully")
	}

	// Exclusive block: Bob should be blocked from stealing alice's active lease
	success = m.AcquireLease("bob", "Drone-01", 50*time.Millisecond)
	if success {
		t.Error("Expected bob lease acquisition to be rejected during active lock")
	}

	// Renewal check: Alice can renew her own lock
	success = m.AcquireLease("alice", "Drone-01", 50*time.Millisecond)
	if !success {
		t.Error("Expected alice to successfully renew her own lease")
	}

	// Expiration verification
	time.Sleep(60 * time.Millisecond)

	// Bob should now succeed after Alice's lease expires
	success = m.AcquireLease("bob", "Drone-01", 50*time.Millisecond)
	if !success {
		t.Error("Expected bob to successfully lock the lease after expiration")
	}
}
