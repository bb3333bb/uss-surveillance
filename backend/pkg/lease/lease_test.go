package lease

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// testManagerContract is run against every Manager implementation so the
// Redis-backed and in-memory brokers are held to the exact same safety
// contract (FR-11's exclusive control mutex). advance moves time forward
// past a lease's expiry - a real sleep for the in-memory manager, and
// miniredis's FastForward for the Redis-backed one, since miniredis
// doesn't expire keys against the real wall clock.
func testManagerContract(t *testing.T, m Manager, advance func(time.Duration)) {
	t.Helper()

	// Initial lease acquisition
	success := m.AcquireLease("alice", "Drone-01", 150*time.Millisecond)
	if !success {
		t.Fatal("Expected alice to acquire control lease successfully")
	}

	// Exclusive block: Bob should be blocked from stealing alice's active lease
	success = m.AcquireLease("bob", "Drone-01", 150*time.Millisecond)
	if success {
		t.Error("Expected bob lease acquisition to be rejected during active lock")
	}

	// Renewal check: Alice can renew her own lock
	success = m.AcquireLease("alice", "Drone-01", 150*time.Millisecond)
	if !success {
		t.Error("Expected alice to successfully renew her own lease")
	}

	holder, held := m.GetLeaseHolder("Drone-01")
	if !held || holder != "alice" {
		t.Errorf("Expected alice to be the reported lease holder, got %q (held=%v)", holder, held)
	}

	// Expiration verification
	advance(200 * time.Millisecond)

	// Bob should now succeed after Alice's lease expires
	success = m.AcquireLease("bob", "Drone-01", 150*time.Millisecond)
	if !success {
		t.Error("Expected bob to successfully lock the lease after expiration")
	}

	m.ReleaseLease("Drone-01")
	if _, held := m.GetLeaseHolder("Drone-01"); held {
		t.Error("Expected no lease holder after ReleaseLease")
	}
}

func TestInMemoryManagerContract(t *testing.T) {
	testManagerContract(t, NewManager(), time.Sleep)
}

func TestRedisManagerContract(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	testManagerContract(t, NewRedisManager(client), mr.FastForward)
}

func TestNewFromEnvFallsBackToInMemoryWhenUnset(t *testing.T) {
	t.Setenv("REDIS_URL", "")
	m := NewFromEnv()
	if _, ok := m.(*InMemoryManager); !ok {
		t.Errorf("expected in-memory manager when REDIS_URL unset, got %T", m)
	}
}

func TestNewFromEnvUsesRedisWhenSet(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	t.Setenv("REDIS_URL", "redis://"+mr.Addr())
	m := NewFromEnv()
	if _, ok := m.(*RedisManager); !ok {
		t.Errorf("expected Redis-backed manager when REDIS_URL set, got %T", m)
	}
}
