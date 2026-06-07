package friction

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/MisterVVP/logarift/backend/internal/domain"
	"github.com/MisterVVP/logarift/backend/internal/llmadapter"
	"github.com/MisterVVP/logarift/backend/internal/serviceerror"
	"github.com/MisterVVP/logarift/backend/internal/store"
	"github.com/MisterVVP/logarift/backend/internal/store/commands"
	"github.com/MisterVVP/logarift/backend/internal/store/queries"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s *Service) processLLMEnrichmentJob(ctx context.Context, jobID bson.ObjectID) {
	if s == nil || s.llmAdapter == nil || s.dispatcher == nil || jobID.IsZero() {
		return
	}
	job, err := s.getLLMJob(ctx, jobID)
	if err != nil {
		slog.Error("llm enrichment job lookup failed", "job_id", jobID.Hex(), "error", err)
		return
	}
	event, err := s.getEventByObjectID(ctx, job.EventID)
	if err != nil {
		s.completeFailedJob(ctx, job, nil, domain.LLMStatusFailed, "event_lookup_failed", err)
		return
	}

	now := s.clock.Now().UTC()
	job.Status = domain.LLMStatusRunning
	job.Attempt++
	job.ClaimedAt = &now
	job.UpdatedAt = now
	if _, err := s.dispatcher.SendCommand(commands.UpdateLLMEnrichmentJob{Context: ctx, Job: job}); err != nil {
		slog.Error("llm enrichment job claim failed", "job_id", job.ID.Hex(), "event_id", job.EventID.Hex(), "trace_id", job.TraceID, "error", err)
		return
	}
	setEventEnrichment(event, job, domain.LLMStatusRunning, "Local LLM enrichment is running.", nil)
	_, _ = s.dispatcher.SendCommand(commands.UpdateFrictionEvent{Context: ctx, Event: event})

	request := llmadapter.RequestFromEvent(job.RequestID, *event, s.llmIncludeMarkdown)
	adapterCtx := llmadapter.ContextWithMetadata(ctx, llmadapter.Metadata{TraceID: job.TraceID, RequestID: job.RequestID, EventID: job.EventID.Hex(), JobID: job.ID.Hex()})
	started := time.Now()
	resp, err := s.llmAdapter.Enrich(adapterCtx, request)
	if err != nil {
		status := domain.LLMStatusFailed
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(adapterCtx.Err(), context.DeadlineExceeded) {
			status = domain.LLMStatusTimedOut
		}
		recordAdapterFailure(event, job.RequestID, err)
		s.completeFailedJob(ctx, job, event, status, "adapter_unavailable", err)
		return
	}
	merge := mergeAdapterResponseWithJob(event, resp, s.llmMinConfidence, job.ID.Hex())
	merge.EventID = event.ID.Hex()
	merge.JobID = job.ID.Hex()
	setEventEnrichment(event, job, merge.LLMStatus, "Local LLM enrichment completed.", &merge)
	if _, err := s.dispatcher.SendCommand(commands.UpdateFrictionEvent{Context: ctx, Event: event}); err != nil {
		s.completeFailedJob(ctx, job, nil, domain.LLMStatusFailed, "event_update_failed", err)
		return
	}
	completed := s.clock.Now().UTC()
	job.Status = merge.LLMStatus
	job.CompletedAt = &completed
	job.WarningCount = merge.AdapterResult.WarningCount
	job.ModelName = merge.AdapterResult.ModelName
	job.PromptVersion = merge.AdapterResult.PromptVersion
	job.MergeSummary = &merge
	job.UpdatedAt = completed
	_, err = s.dispatcher.SendCommand(commands.UpdateLLMEnrichmentJob{Context: ctx, Job: job})
	if err != nil {
		slog.Error("llm enrichment job completion update failed", "job_id", job.ID.Hex(), "event_id", job.EventID.Hex(), "trace_id", job.TraceID, "error", err)
		return
	}
	slog.Info("llm enrichment merge completed", "trace_id", job.TraceID, "event_id", job.EventID.Hex(), "job_id", job.ID.Hex(), "status", merge.LLMStatus, "duration_ms", time.Since(started).Milliseconds(), "accepted_field_count", merge.AcceptedFieldCount, "rejected_field_count", merge.RejectedFieldCount, "fallback_field_count", merge.FallbackFieldCount, "field_decisions", merge.FieldDecisions)
}

func (s *Service) completeFailedJob(ctx context.Context, job *domain.LLMEnrichmentJob, event *domain.FrictionEvent, status, code string, err error) {
	now := s.clock.Now().UTC()
	job.Status = status
	job.ErrorCode = code
	if err != nil {
		job.LastError = err.Error()
	}
	job.CompletedAt = &now
	job.UpdatedAt = now
	if event != nil {
		setEventEnrichment(event, job, status, "Local LLM enrichment failed; deterministic enrichment was kept.", nil)
		_, _ = s.dispatcher.SendCommand(commands.UpdateFrictionEvent{Context: ctx, Event: event})
	}
	_, _ = s.dispatcher.SendCommand(commands.UpdateLLMEnrichmentJob{Context: ctx, Job: job})
	slog.Warn("llm enrichment job failed", "trace_id", job.TraceID, "event_id", job.EventID.Hex(), "job_id", job.ID.Hex(), "status", status, "error_code", code, "error", job.LastError)
}

func (s *Service) getLLMJob(ctx context.Context, id bson.ObjectID) (*domain.LLMEnrichmentJob, error) {
	out, err := s.dispatcher.SendQuery(queries.GetLLMEnrichmentJobByID{Context: ctx, ID: id})
	if err != nil {
		return nil, mapErr(err)
	}
	job, ok := out.(*domain.LLMEnrichmentJob)
	if !ok || job == nil {
		return nil, store.ErrNotFound
	}
	return job, nil
}

func (s *Service) getEventByObjectID(ctx context.Context, id bson.ObjectID) (*domain.FrictionEvent, error) {
	out, err := s.dispatcher.SendQuery(queries.GetFrictionEventByID{Context: ctx, ID: id})
	if err != nil {
		return nil, mapErr(err)
	}
	event, ok := out.(*domain.FrictionEvent)
	if !ok || event == nil {
		return nil, store.ErrNotFound
	}
	return event, nil
}

func setEventEnrichment(event *domain.FrictionEvent, job *domain.LLMEnrichmentJob, status, message string, merge *domain.LLMEnrichmentMergeResult) {
	if event.Enrichment == nil {
		event.Enrichment = &domain.FrictionEnrichment{DeterministicStatus: "applied"}
	}
	event.Enrichment.LLMStatus = status
	event.Enrichment.JobID = job.ID.Hex()
	event.Enrichment.TraceID = job.TraceID
	event.Enrichment.UserMessage = message
	event.Enrichment.UpdatedAt = time.Now().UTC()
	event.Enrichment.MergeSummary = merge
}

func (s *Service) GetLLMEnrichmentJob(ctx context.Context, id string) (domain.LLMEnrichmentJob, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return domain.LLMEnrichmentJob{}, serviceerror.ValidationError{Fields: []serviceerror.FieldError{fe("id", "must be a valid ObjectID")}}
	}
	job, err := s.getLLMJob(ctx, objectID)
	if err != nil {
		return domain.LLMEnrichmentJob{}, err
	}
	return *job, nil
}
