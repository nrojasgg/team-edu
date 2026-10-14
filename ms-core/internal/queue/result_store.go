package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisResultStore implements ResultStore using a go-redis client.
// Job results are stored as JSON strings with a configurable TTL so that
// the polling handler can retrieve them without keeping any in-process state.
type RedisResultStore struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisResultStore creates a RedisResultStore backed by the provided client.
// If ttl is zero, ResultTTL (30 min) is used as the default.
func NewRedisResultStore(client *redis.Client, ttl time.Duration) *RedisResultStore {
	if ttl <= 0 {
		ttl = ResultTTL
	}
	return &RedisResultStore{
		client: client,
		ttl:    ttl,
	}
}

// SaveResult persists result as a JSON string in Redis under ResultKey(result.JobID).
// Any previous value for the same key is overwritten, which allows the worker
// to transition a job from "processing" → "completed" / "failed".
func (s *RedisResultStore) SaveResult(ctx context.Context, result JobResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("result_store: marshal job result: %w", err)
	}

	key := ResultKey(result.JobID)
	if err := s.client.Set(ctx, key, data, s.ttl).Err(); err != nil {
		return fmt.Errorf("result_store: save result for job %s: %w", result.JobID, err)
	}
	return nil
}

// GetResult retrieves the JobResult for jobID from Redis.
// Returns ErrNotFound when the key does not exist (job is still pending or
// the result has already expired).
func (s *RedisResultStore) GetResult(ctx context.Context, jobID string) (JobResult, error) {
	key := ResultKey(jobID)

	data, err := s.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return JobResult{}, ErrNotFound
		}
		return JobResult{}, fmt.Errorf("result_store: get result for job %s: %w", jobID, err)
	}

	var result JobResult
	if err := json.Unmarshal(data, &result); err != nil {
		return JobResult{}, fmt.Errorf("result_store: unmarshal result for job %s: %w", jobID, err)
	}
	return result, nil
}
