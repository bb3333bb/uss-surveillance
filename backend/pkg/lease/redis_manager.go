package lease

import (
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisKeyPrefix = "uss:lease:"

// RedisManager backs control leases with Redis, safe for multi-instance
// gateway deployments (retro action item, epics 2/3). Acquisition/renewal
// runs inside a WATCH/MULTI transaction so a concurrent acquire from
// another operator can never race a renewal from the current holder.
// Expiry is delegated to Redis' own key TTL rather than a stored
// timestamp: a missing key is indistinguishable from - and treated the
// same as - an expired one.
type RedisManager struct {
	client *redis.Client
}

// NewRedisManager wraps an existing Redis client.
func NewRedisManager(client *redis.Client) *RedisManager {
	return &RedisManager{client: client}
}

func (m *RedisManager) key(droneID string) string {
	return redisKeyPrefix + droneID
}

// AcquireLease attempts to lock or renew command access to a drone.
func (m *RedisManager) AcquireLease(operator string, droneID string, duration time.Duration) bool {
	ctx := context.Background()
	key := m.key(droneID)
	acquired := false

	err := m.client.Watch(ctx, func(tx *redis.Tx) error {
		current, err := tx.Get(ctx, key).Result()
		if err != nil && err != redis.Nil {
			return err
		}
		if err == nil && current != operator {
			// Held by a different operator and not yet expired.
			return nil
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, key, operator, duration)
			return nil
		})
		if err != nil {
			return err
		}
		acquired = true
		return nil
	}, key)

	if err != nil {
		return false
	}
	return acquired
}

// GetLeaseHolder returns the current operator owning the drone control lock.
func (m *RedisManager) GetLeaseHolder(droneID string) (string, bool) {
	ctx := context.Background()
	val, err := m.client.Get(ctx, m.key(droneID)).Result()
	if err != nil {
		return "", false
	}
	return val, true
}

// ReleaseLease forces eviction of the current control lock lease.
func (m *RedisManager) ReleaseLease(droneID string) {
	ctx := context.Background()
	_ = m.client.Del(ctx, m.key(droneID)).Err()
}

// NewFromEnv selects a Redis-backed manager when REDIS_URL is set, and
// falls back to an in-memory manager (see docs/CONFIGURATION.md) otherwise.
func NewFromEnv() Manager {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return NewManager()
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return NewManager()
	}

	return NewRedisManager(redis.NewClient(opt))
}
