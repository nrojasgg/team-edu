package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ms-core/internal/queue"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ── Async HTTP endpoints ───────────────────────────────────────────────────────

// FlexibleFormTeamsAsync enqueues a team formation job and returns immediately
// with a job_id that the client can poll via GET /jobs/:id.
//
// POST /api/v1/team-formation/form-flexible/async
func (h *TeamFormationHandler) FlexibleFormTeamsAsync(c *fiber.Ctx) error {
	if h.broker == nil {
		// Graceful degradation: if no broker is configured fall back to sync.
		return h.FlexibleFormTeams(c)
	}

	// Validate the body can be parsed before enqueuing so the client gets an
	// immediate 400 instead of a silent failure deep in the worker.
	var req FlexibleFormTeamsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   errInvalidRequest,
			"details": err.Error(),
		})
	}
	if req.SectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "section_id is required"})
	}
	if len(req.Candidates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "candidates is required"})
	}
	if req.TeamSize < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"success": false, "error": "team_size must be at least 2"})
	}

	// Re-encode the validated body so the worker receives the canonical form.
	rawBody, err := json.Marshal(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "failed to encode request",
		})
	}

	jobID := uuid.New().String()
	payload := queue.FlexibleFormTeamsPayload{
		JobID:   jobID,
		Request: rawBody,
	}

	// Write a "pending" sentinel BEFORE enqueuing so GetJobResult can
	// distinguish a brand-new job from an unknown job ID while the task
	// is in transit to asynq.
	if h.resultStore != nil {
		if sentinelErr := h.resultStore.SaveResult(c.Context(), queue.JobResult{
			JobID:      jobID,
			Status:     queue.JobStatusPending,
			EnqueuedAt: time.Now(),
		}); sentinelErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "failed to initialise job state",
				"details": sentinelErr.Error(),
			})
		}
	}

	if _, err := h.broker.EnqueueFlexibleFormTeams(c.Context(), payload); err != nil {
		// Roll back the sentinel so the job ID does not linger as "pending"
		// forever without a worker ever picking it up.
		if h.resultStore != nil {
			_ = h.resultStore.SaveResult(c.Context(), queue.JobResult{
				JobID:  jobID,
				Status: queue.JobStatusFailed,
				Error:  fmt.Sprintf("enqueue failed: %s", err.Error()),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "failed to enqueue job",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"success":  true,
		"job_id":   jobID,
		"status":   queue.JobStatusPending,
		"poll_url": fmt.Sprintf("/api/v1/jobs/%s", jobID),
	})
}

// EvaluateGroupsAsync enqueues a group evaluation job and returns immediately
// with a job_id.
//
// POST /api/v1/team-formation/evaluate-groups/async
func (h *TeamFormationHandler) EvaluateGroupsAsync(c *fiber.Ctx) error {
	if h.broker == nil {
		return h.EvaluateGroups(c)
	}

	var req EvaluateGroupsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": fmt.Errorf("invalid request body: %w", err).Error(),
		})
	}
	if req.SectionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": true, "message": "section_id is required"})
	}
	if len(req.Groups) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": true, "message": "at least one group is required"})
	}
	if len(req.CriteriaParameters) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": true, "message": "criteria_parameters is required"})
	}

	rawBody, err := json.Marshal(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "failed to encode request",
		})
	}

	jobID := uuid.New().String()
	payload := queue.EvaluateGroupsPayload{
		JobID:   jobID,
		Request: rawBody,
	}

	if h.resultStore != nil {
		if sentinelErr := h.resultStore.SaveResult(c.Context(), queue.JobResult{
			JobID:      jobID,
			Status:     queue.JobStatusPending,
			EnqueuedAt: time.Now(),
		}); sentinelErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "failed to initialise job state",
				"details": sentinelErr.Error(),
			})
		}
	}

	if _, err := h.broker.EnqueueEvaluateGroups(c.Context(), payload); err != nil {
		if h.resultStore != nil {
			_ = h.resultStore.SaveResult(c.Context(), queue.JobResult{
				JobID:  jobID,
				Status: queue.JobStatusFailed,
				Error:  fmt.Sprintf("enqueue failed: %s", err.Error()),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "failed to enqueue job",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"success":  true,
		"job_id":   jobID,
		"status":   queue.JobStatusPending,
		"poll_url": fmt.Sprintf("/api/v1/jobs/%s", jobID),
	})
}

// GetJobResult polls Redis for the result of a previously enqueued job.
//
// GET /api/v1/jobs/:id
func (h *TeamFormationHandler) GetJobResult(c *fiber.Ctx) error {
	if h.resultStore == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"error":   "async job store is not available",
		})
	}

	jobID := c.Params("id")
	if jobID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "job id is required",
		})
	}

	result, err := h.resultStore.GetResult(c.Context(), jobID)
	if err != nil {
		if errors.Is(err, queue.ErrNotFound) {
			// The job was never written to Redis (broker down, expired TTL, or
			// unknown ID).  Return 404 with enough context to diagnose.
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"error":   "job not found: the job may have expired, never been enqueued successfully, or the provided id is incorrect",
				"job_id":  jobID,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "failed to retrieve job result",
			"details": err.Error(),
		})
	}

	// While the job is still pending/processing return 202 so the client
	// knows to keep polling without misinterpreting it as success.
	httpStatus := fiber.StatusOK
	if result.Status == queue.JobStatusPending || result.Status == queue.JobStatusProcessing {
		httpStatus = fiber.StatusAccepted
	}

	resp := fiber.Map{
		"success":     result.Status == queue.JobStatusCompleted,
		"job_id":      result.JobID,
		"status":      result.Status,
		"enqueued_at": result.EnqueuedAt,
	}
	if result.CompletedAt != nil {
		resp["completed_at"] = result.CompletedAt
	}
	if result.Status == queue.JobStatusCompleted && result.Data != nil {
		// Embed the raw JSON data directly so the client sees the same shape
		// as the synchronous endpoint.
		resp["data"] = json.RawMessage(result.Data)
	}
	if result.Status == queue.JobStatusFailed {
		resp["error"] = result.Error
	}

	return c.Status(httpStatus).JSON(resp)
}

// ── Payload-execution methods (called by worker) ──────────────────────────────

// ExecuteFlexibleFormTeamsFromPayload deserialises the JSON request stored in
// the queue payload and runs the full synchronous FlexibleFormTeams logic,
// returning the result without touching Fiber's context.
//
// This method is the single source of truth for the business logic so both the
// sync HTTP handler and the async worker delegate to the same code path.
func (h *TeamFormationHandler) ExecuteFlexibleFormTeamsFromPayload(ctx context.Context, raw json.RawMessage) (*TeamFormationAsyncResult, error) {
	var req FlexibleFormTeamsRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, fmt.Errorf("execute flexible form teams: unmarshal request: %w", err)
	}

	if req.MinTeamSize == 0 {
		req.MinTeamSize = req.TeamSize - 1
	}
	if req.MaxTeamSize == 0 {
		req.MaxTeamSize = req.TeamSize + 1
	}

	algorithm := req.Algorithm
	if algorithm == "" {
		algorithm = "genetic"
	}

	useLegacyMode := req.ProviderConfig != nil && len(req.CriteriaParameters) == 0
	useNewMode := len(req.CriteriaParameters) > 0

	if !useLegacyMode && !useNewMode {
		return nil, fmt.Errorf("execute flexible form teams: either provider_config or criteria_parameters must be provided")
	}

	candidates := make([]TeamMember, len(req.Candidates))
	for i, id := range req.Candidates {
		candidates[i] = TeamMember{ID: id}
	}

	// runLegacyFormation and runCriteriaFormation both accept a *fiber.Ctx for
	// the plugin RPC call.  We build a minimal shim that only provides a
	// context.Context so the core logic is reachable without an HTTP round-trip.
	var result interface{}
	var err error

	if useLegacyMode {
		result, err = h.runLegacyFormationCtx(ctx, req, candidates, algorithm)
	} else {
		result, err = h.runCriteriaFormationCtx(ctx, req, candidates, algorithm)
	}

	if err != nil {
		return nil, fmt.Errorf("execute flexible form teams: %w", err)
	}

	go h.saveGroupHistory(req.SectionID, extractTeamFormationResult(result))

	return &TeamFormationAsyncResult{Payload: result}, nil
}

// ExecuteEvaluateGroupsFromPayload deserialises the JSON request stored in
// the queue payload and runs the full EvaluateGroups logic without Fiber.
func (h *TeamFormationHandler) ExecuteEvaluateGroupsFromPayload(ctx context.Context, raw json.RawMessage) (*EvaluateGroupsAsyncResult, error) {
	var req EvaluateGroupsRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, fmt.Errorf("execute evaluate groups: unmarshal request: %w", err)
	}

	resp, err := h.evaluateGroupsCore(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("execute evaluate groups: %w", err)
	}
	return &EvaluateGroupsAsyncResult{Response: resp}, nil
}

// ── Result wrappers ────────────────────────────────────────────────────────────

// TeamFormationAsyncResult wraps the polymorphic result produced by either the
// legacy or criteria formation path so the worker can serialise it uniformly.
type TeamFormationAsyncResult struct {
	Payload interface{}
}

// EvaluateGroupsAsyncResult wraps the EvaluateGroupsResponse so the worker
// does not need to import the concrete type.
type EvaluateGroupsAsyncResult struct {
	Response *EvaluateGroupsResponse
}

// ── Dependency setters ────────────────────────────────────────────────────────

// WithBroker injects the queue Broker used by the async endpoints.
// Calling this method is optional; if not called the async endpoints fall back
// to synchronous execution.
func (h *TeamFormationHandler) WithBroker(b queue.Broker) *TeamFormationHandler {
	h.broker = b
	return h
}

// WithResultStore injects the ResultStore used by the async endpoints and the
// job-polling handler.
func (h *TeamFormationHandler) WithResultStore(rs queue.ResultStore) *TeamFormationHandler {
	h.resultStore = rs
	return h
}
