// Package queue provides a Redis-backed task queue for offloading CPU-intensive
// operations (team formation, group evaluation) from the HTTP request path.
//
// Architecture:
//
//	HTTP handler  →  Enqueue(task)  →  Redis (asynq)  →  Worker.Process(task)
//	             ←  job_id                             →  StoreResult(job_id)
//	             ←  GET /jobs/:id  ←  Redis result key
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// ── Task type constants ────────────────────────────────────────────────────────

const (
	// TypeFlexibleFormTeams is the task type for async flexible team formation.
	TypeFlexibleFormTeams = "team_formation:flexible"

	// TypeEvaluateGroups is the task type for async group evaluation.
	TypeEvaluateGroups = "team_formation:evaluate"

	// DefaultQueue is the default asynq queue name.
	DefaultQueue = "default"

	// CriticalQueue is used for user-facing jobs that must be processed first.
	CriticalQueue = "critical"

	// resultKeyPrefix is the Redis key prefix for storing job results.
	resultKeyPrefix = "job:result:"

	// ResultTTL is how long a completed job result is kept in Redis.
	ResultTTL = 30 * time.Minute
)

// ── Status ─────────────────────────────────────────────────────────────────────

// JobStatus represents the lifecycle state of an enqueued job.
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// ── Payloads ───────────────────────────────────────────────────────────────────

// FlexibleFormTeamsPayload carries the full request data needed by the worker
// to execute a flexible team formation job.
type FlexibleFormTeamsPayload struct {
	JobID   string          `json:"job_id"`
	Request json.RawMessage `json:"request"`
}

// EvaluateGroupsPayload carries the full request data needed by the worker
// to execute a group evaluation job.
type EvaluateGroupsPayload struct {
	JobID   string          `json:"job_id"`
	Request json.RawMessage `json:"request"`
}

// ── Result envelope ────────────────────────────────────────────────────────────

// JobResult is the envelope stored in Redis once a job completes or fails.
type JobResult struct {
	JobID       string          `json:"job_id"`
	Status      JobStatus       `json:"status"`
	Data        json.RawMessage `json:"data,omitempty"`
	Error       string          `json:"error,omitempty"`
	EnqueuedAt  time.Time       `json:"enqueued_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
}

// ── Broker interface ───────────────────────────────────────────────────────────

// Broker abstracts the enqueue side of the task queue so handlers remain
// decoupled from the asynq implementation.
type Broker interface {
	// EnqueueFlexibleFormTeams enqueues a flexible team formation job and
	// returns the assigned job ID.
	EnqueueFlexibleFormTeams(ctx context.Context, payload FlexibleFormTeamsPayload) (string, error)

	// EnqueueEvaluateGroups enqueues a group evaluation job and returns the
	// assigned job ID.
	EnqueueEvaluateGroups(ctx context.Context, payload EvaluateGroupsPayload) (string, error)

	// Close releases underlying resources.
	Close() error
}

// ResultStore abstracts reading and writing job results so the worker and the
// polling handler do not depend on a concrete Redis client.
type ResultStore interface {
	// SaveResult persists a JobResult under its job_id with ResultTTL.
	SaveResult(ctx context.Context, result JobResult) error

	// GetResult retrieves a previously stored JobResult by job_id.
	// Returns ErrNotFound if the key does not exist yet.
	GetResult(ctx context.Context, jobID string) (JobResult, error)
}

// ErrNotFound is returned by ResultStore.GetResult when no result exists yet.
var ErrNotFound = fmt.Errorf("job result not found")

// ── asynq Broker implementation ────────────────────────────────────────────────

// AsynqBroker implements Broker using hibiken/asynq backed by Redis.
type AsynqBroker struct {
	client *asynq.Client
}

// NewAsynqBroker creates a new AsynqBroker connected to the given Redis address.
//
//	broker, err := queue.NewAsynqBroker("localhost:6379")
func NewAsynqBroker(redisAddr string) (*AsynqBroker, error) {
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &AsynqBroker{client: client}, nil
}

// EnqueueFlexibleFormTeams implements Broker.
func (b *AsynqBroker) EnqueueFlexibleFormTeams(ctx context.Context, payload FlexibleFormTeamsPayload) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("queue: marshal flexible form teams payload: %w", err)
	}

	task := asynq.NewTask(TypeFlexibleFormTeams, data, asynq.TaskID(payload.JobID))
	info, err := b.client.EnqueueContext(ctx, task,
		asynq.Queue(CriticalQueue),
		asynq.MaxRetry(1),
		asynq.Timeout(5*time.Minute),
		asynq.Unique(ResultTTL),
	)
	if err != nil {
		// asynq returns ErrTaskIDConflict when a task with the same ID is
		// already queued/active. Since the job_id is a fresh UUID generated
		// by the handler, a conflict here means the task was already accepted
		// (idempotent success) — return the existing ID instead of an error.
		if err == asynq.ErrTaskIDConflict {
			return payload.JobID, nil
		}
		return "", fmt.Errorf("queue: enqueue flexible form teams: %w", err)
	}
	return info.ID, nil
}

// EnqueueEvaluateGroups implements Broker.
func (b *AsynqBroker) EnqueueEvaluateGroups(ctx context.Context, payload EvaluateGroupsPayload) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("queue: marshal evaluate groups payload: %w", err)
	}

	task := asynq.NewTask(TypeEvaluateGroups, data, asynq.TaskID(payload.JobID))
	info, err := b.client.EnqueueContext(ctx, task,
		asynq.Queue(CriticalQueue),
		asynq.MaxRetry(1),
		asynq.Timeout(5*time.Minute),
		asynq.Unique(ResultTTL),
	)
	if err != nil {
		if err == asynq.ErrTaskIDConflict {
			return payload.JobID, nil
		}
		return "", fmt.Errorf("queue: enqueue evaluate groups: %w", err)
	}
	return info.ID, nil
}

// Close implements Broker.
func (b *AsynqBroker) Close() error {
	return b.client.Close()
}

// ── asynq Server constructor ───────────────────────────────────────────────────

// NewAsynqServer creates a pre-configured asynq.Server ready to register
// handlers.  The caller must call srv.Start(mux) after registering task
// handlers on the returned mux.
//
//	srv, mux := queue.NewAsynqServer("localhost:6379", 4)
//	mux.HandleFunc(queue.TypeFlexibleFormTeams, myHandler)
//	srv.Start(mux)
func NewAsynqServer(redisAddr string, concurrency int) (*asynq.Server, *asynq.ServeMux) {
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	if concurrency <= 0 {
		// Default to 8 workers. Tune this value based on the number of CPU
		// cores available and the expected job throughput. For stress tests
		// with burst enqueue patterns, values between 8 and 32 are recommended.
		concurrency = 8
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: concurrency,
			Queues: map[string]int{
				CriticalQueue: 6,
				DefaultQueue:  2,
			},
			// Retry delay with exponential back-off capped at 30 s.
			RetryDelayFunc: func(n int, _ error, _ *asynq.Task) time.Duration {
				d := time.Duration(n*n) * time.Second
				if d > 30*time.Second {
					d = 30 * time.Second
				}
				return d
			},
		},
	)

	mux := asynq.NewServeMux()
	return srv, mux
}

// ResultKey returns the Redis key used to store a job result.
func ResultKey(jobID string) string {
	return resultKeyPrefix + jobID
}
