package httpapi

import (
	"context"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryexecution"
)

type routeExecutionRepository struct {
	execution batchfactoryexecution.Execution
	created   bool
}

func (r *routeExecutionRepository) Ensure(_ context.Context, record batchfactoryexecution.EnsureRecord) (batchfactoryexecution.EnsureResult, error) {
	if r.execution.ID == "" {
		r.execution = batchfactoryexecution.Execution{ID: "execution-1", OwnerUsername: record.Owner, BatchID: record.BatchID, Mode: record.Mode, Target: record.Target, ClientRequestID: record.ClientRequestID, Status: batchfactoryexecution.ExecutionPlanned}
		r.created = true
		return batchfactoryexecution.EnsureResult{Execution: r.execution, Created: true, QueuedBooks: []string{"book-1"}}, nil
	}
	return batchfactoryexecution.EnsureResult{Execution: r.execution, Created: false, QueuedBooks: []string{"book-1"}}, nil
}
func (r *routeExecutionRepository) Get(_ context.Context, owner, id string) (batchfactoryexecution.Execution, error) {
	if r.execution.ID == id && r.execution.OwnerUsername == owner {
		return r.execution, nil
	}
	return batchfactoryexecution.Execution{}, batchfactoryexecution.ErrNotFound
}
func (r *routeExecutionRepository) GetByClientRequest(context.Context, string, string, string) (batchfactoryexecution.Execution, bool, error) {
	return batchfactoryexecution.Execution{}, false, nil
}
func (r *routeExecutionRepository) ListBookTimeline(context.Context, string, string, string) ([]batchfactoryexecution.TimelineEvent, error) {
	return []batchfactoryexecution.TimelineEvent{}, nil
}
func (r *routeExecutionRepository) AttachLegacyJob(context.Context, string, string, string) error {
	return nil
}
func (r *routeExecutionRepository) UpdateBookStatus(context.Context, string, string, batchfactoryexecution.BookStatus, batchfactoryexecution.Failure) error {
	return nil
}
func (r *routeExecutionRepository) UpdateExecutionStatus(_ context.Context, owner, id string, status batchfactoryexecution.ExecutionStatus) error {
	if r.execution.ID != id || r.execution.OwnerUsername != owner {
		return batchfactoryexecution.ErrNotFound
	}
	r.execution.Status = status
	return nil
}
func (r *routeExecutionRepository) AppendOutbox(context.Context, batchfactoryexecution.OutboxEvent) error {
	return nil
}

func TestBatchExecutionRoutesRequireAuthentication(t *testing.T) {
	mux := http.NewServeMux()
	service := batchfactoryexecution.NewService(&routeExecutionRepository{}, nil)
	registerBatchExecutionRoutes(mux, BridgeAuth{Secret: "secret", Now: time.Now}, service)
	rec := signedJSONRequest(t, mux, time.Now(), "", http.MethodPost, "/api/batch-factory/v12/batches/batch-1/executions/ensure", map[string]any{"bookIds": []string{"book-1"}, "mode": "video", "target": "published_confirmed", "clientRequestId": "request-1", "config": map[string]any{}})
	// Empty bridge usernames are rejected by the middleware.
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestBatchExecutionRoutesCreateReuseAndValidate(t *testing.T) {
	now := time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC)
	repo := &routeExecutionRepository{}
	service := batchfactoryexecution.NewService(repo, nil)
	mux := http.NewServeMux()
	registerBatchExecutionRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, service)
	body := map[string]any{"bookIds": []string{"book-1"}, "mode": "video", "target": "published_confirmed", "clientRequestId": "request-1", "config": map[string]any{}}
	first := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/batch-factory/v12/batches/batch-1/executions/ensure", body)
	second := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/batch-factory/v12/batches/batch-1/executions/ensure", body)
	if first.Code != http.StatusCreated || second.Code != http.StatusOK {
		t.Fatalf("first=%d %s second=%d %s", first.Code, first.Body.String(), second.Code, second.Body.String())
	}
	one := decodeBody[batchfactoryexecution.EnsureResponse](t, first)
	two := decodeBody[batchfactoryexecution.EnsureResponse](t, second)
	if one.ExecutionID == "" || one.ExecutionID != two.ExecutionID {
		t.Fatalf("first=%#v second=%#v", one, two)
	}
	bad := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/batch-factory/v12/batches/batch-1/executions/ensure", map[string]any{"bookIds": []string{"book-1"}, "mode": "invalid", "target": "published_confirmed", "clientRequestId": "request-2"})
	if bad.Code != http.StatusBadRequest {
		t.Fatalf("bad status=%d body=%s", bad.Code, bad.Body.String())
	}
}

func TestBatchExecutionRoutesIsolateOwners(t *testing.T) {
	now := time.Date(2026, 11, 22, 10, 0, 0, 0, time.UTC)
	repo := &routeExecutionRepository{execution: batchfactoryexecution.Execution{ID: "execution-1", OwnerUsername: "alice", BatchID: "batch-1", Status: batchfactoryexecution.ExecutionRunning}}
	service := batchfactoryexecution.NewService(repo, nil)
	mux := http.NewServeMux()
	registerBatchExecutionRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, service)
	allowed := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/batch-factory/v12/executions/execution-1", nil)
	denied := signedJSONRequest(t, mux, now, "bob", http.MethodGet, "/api/batch-factory/v12/executions/execution-1", nil)
	if allowed.Code != http.StatusOK || denied.Code != http.StatusNotFound {
		t.Fatalf("allowed=%d denied=%d", allowed.Code, denied.Code)
	}
}
