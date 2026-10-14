// Package worker contains asynq task handlers that execute CPU-intensive
// team formation and group evaluation jobs outside the HTTP request path.
//
// Each handler:
//  1. Deserialises the payload from the asynq task.
//  2. Marks the job as "processing" in the ResultStore.
//  3. Runs the heavy computation (genetic algorithm, criteria evaluation, …).
//  4. Stores the result (or error) back in the ResultStore so the polling
//     endpoint can return it to the client.
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"ms-core/internal/handlers"
	"ms-core/internal/queue"

	"github.com/hibiken/asynq"
)

// TeamFormationWorker holds the dependencies shared across all task handlers.
// It is intentionally thin: actual business logic lives in TeamFormationHandler.
type TeamFormationWorker struct {
	handler     *handlers.TeamFormationHandler
	resultStore queue.ResultStore
}

// NewTeamFormationWorker creates a TeamFormationWorker and registers its task
// handlers on mux so the caller only needs to start the asynq.Server.
//
//	w := worker.NewTeamFormationWorker(tfHandler, resultStore, mux)
func NewTeamFormationWorker(
	handler *handlers.TeamFormationHandler,
	resultStore queue.ResultStore,
	mux *asynq.ServeMux,
) *TeamFormationWorker {
	w := &TeamFormationWorker{
		handler:     handler,
		resultStore: resultStore,
	}
	mux.HandleFunc(queue.TypeFlexibleFormTeams, w.HandleFlexibleFormTeams)
	mux.HandleFunc(queue.TypeEvaluateGroups, w.HandleEvaluateGroups)
	return w
}

// ── FlexibleFormTeams ──────────────────────────────────────────────────────────

// HandleFlexibleFormTeams processes a TypeFlexibleFormTeams asynq task.
func (w *TeamFormationWorker) HandleFlexibleFormTeams(ctx context.Context, t *asynq.Task) error {
	var payload queue.FlexibleFormTeamsPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("worker: unmarshal flexible form teams payload: %w", err)
	}

	log.Printf("[worker] FlexibleFormTeams start job_id=%s", payload.JobID)

	// Mark as processing so the client gets a useful status during computation.
	w.markProcessing(ctx, payload.JobID)

	result, err := w.handler.ExecuteFlexibleFormTeamsFromPayload(ctx, payload.Request)

	now := time.Now()
	jobResult := queue.JobResult{
		JobID:       payload.JobID,
		CompletedAt: &now,
	}

	if err != nil {
		log.Printf("[worker] FlexibleFormTeams failed job_id=%s err=%v", payload.JobID, err)
		jobResult.Status = queue.JobStatusFailed
		jobResult.Error = err.Error()
	} else {
		data, merr := json.Marshal(map[string]interface{}{"success": true, "data": result})
		if merr != nil {
			jobResult.Status = queue.JobStatusFailed
			jobResult.Error = fmt.Sprintf("marshal result: %v", merr)
		} else {
			jobResult.Status = queue.JobStatusCompleted
			jobResult.Data = data
		}
		log.Printf("[worker] FlexibleFormTeams completed job_id=%s", payload.JobID)
	}

	if serr := w.resultStore.SaveResult(ctx, jobResult); serr != nil {
		log.Printf("[worker] failed to save result job_id=%s: %v", payload.JobID, serr)
	}

	// Return nil so asynq does not retry on business-logic failures; retries
	// are meaningful only for infrastructure errors (Redis down, etc.).
	return nil
}

// ── EvaluateGroups ─────────────────────────────────────────────────────────────

// HandleEvaluateGroups processes a TypeEvaluateGroups asynq task.
func (w *TeamFormationWorker) HandleEvaluateGroups(ctx context.Context, t *asynq.Task) error {
	var payload queue.EvaluateGroupsPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("worker: unmarshal evaluate groups payload: %w", err)
	}

	log.Printf("[worker] EvaluateGroups start job_id=%s", payload.JobID)

	w.markProcessing(ctx, payload.JobID)

	result, err := w.handler.ExecuteEvaluateGroupsFromPayload(ctx, payload.Request)

	now := time.Now()
	jobResult := queue.JobResult{
		JobID:       payload.JobID,
		CompletedAt: &now,
	}

	if err != nil {
		log.Printf("[worker] EvaluateGroups failed job_id=%s err=%v", payload.JobID, err)
		jobResult.Status = queue.JobStatusFailed
		jobResult.Error = err.Error()
	} else {
		data, merr := json.Marshal(map[string]interface{}{"success": true, "data": result})
		if merr != nil {
			jobResult.Status = queue.JobStatusFailed
			jobResult.Error = fmt.Sprintf("marshal result: %v", merr)
		} else {
			jobResult.Status = queue.JobStatusCompleted
			jobResult.Data = data
		}
		log.Printf("[worker] EvaluateGroups completed job_id=%s", payload.JobID)
	}

	if serr := w.resultStore.SaveResult(ctx, jobResult); serr != nil {
		log.Printf("[worker] failed to save result job_id=%s: %v", payload.JobID, serr)
	}

	return nil
}

// ── helpers ────────────────────────────────────────────────────────────────────

// markProcessing writes a "processing" sentinel to the ResultStore so that the
// polling handler can distinguish "not yet started" from "in progress".
func (w *TeamFormationWorker) markProcessing(ctx context.Context, jobID string) {
	sentinel := queue.JobResult{
		JobID:      jobID,
		Status:     queue.JobStatusProcessing,
		EnqueuedAt: time.Now(),
	}
	if err := w.resultStore.SaveResult(ctx, sentinel); err != nil {
		log.Printf("[worker] markProcessing failed job_id=%s: %v", jobID, err)
	}
}
