package external

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

type countingProvider struct{ calls int }

func (p *countingProvider) Submit(_ context.Context, _ CredentialInput, _ SubmissionIntent) (ProviderReference, error) {
	p.calls++
	return ProviderReference{Provider: Provider121, Reference: "ref-1", Status: "submitted"}, nil
}

func testService(provider Provider, enabled bool, adapter SubmissionProvider) (*Service, *MemoryStore) {
	store := NewMemoryStore()
	return &Service{Credentials: store, Intents: store, Audits: store, Providers: map[Provider]SubmissionProvider{provider: adapter}, Enabled: map[Provider]bool{provider: enabled}, Key: []byte(strings.Repeat("k", 32))}, store
}

func TestDisabledGateStopsBeforeCredentialAndTransport(t *testing.T) {
	provider := &countingProvider{}
	service, store := testService(ProviderYadi, false, provider)
	if _, err := service.SaveCredential(context.Background(), "alice", "yadi", CredentialInput{Name: "account", Secret: "secret"}); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected unavailable, got %v", err)
	}
	if _, err := service.Submit(context.Background(), "alice", "yadi", "intent"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected unavailable, got %v", err)
	}
	if provider.calls != 0 || len(store.credentials) != 0 {
		t.Fatalf("disabled gate touched state or transport")
	}
}

func TestSubmitRequiresConfirmationAndRedactsAudit(t *testing.T) {
	provider := &countingProvider{}
	service, store := testService(Provider121, true, provider)
	if _, err := service.SaveCredential(context.Background(), "alice", "121", CredentialInput{Name: "account", Secret: "secret-value"}); err != nil {
		t.Fatal(err)
	}
	intent, err := service.CreateIntent(context.Background(), "alice", "121", "batch-1", "", []byte(`{"title":"safe"}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Submit(context.Background(), "alice", "121", intent.ID); !errors.Is(err, ErrNotConfirmed) {
		t.Fatalf("expected confirmation error, got %v", err)
	}
	if _, err := service.ConfirmIntent(context.Background(), "alice", intent.ID); err != nil {
		t.Fatal(err)
	}
	ref, err := service.Submit(context.Background(), "alice", "121", intent.ID)
	if err != nil || ref.Reference != "ref-1" {
		t.Fatalf("submit failed: %#v %v", ref, err)
	}
	audits, _ := store.ListAudits(context.Background(), "alice", intent.ID)
	if len(audits) != 1 || strings.Contains(audits[0].Message, "secret-value") {
		t.Fatalf("audit leaked secret: %#v", audits)
	}
	if _, err := service.Submit(context.Background(), "alice", "121", intent.ID); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected single submit conflict, got %v", err)
	}
}

func TestCrossProviderIntentRejected(t *testing.T) {
	provider := &countingProvider{}
	service, _ := testService(Provider121, true, provider)
	store := service.Intents.(*MemoryStore)
	intent, _ := service.CreateIntent(context.Background(), "alice", "121", "batch-1", "", nil)
	if _, err := service.Submit(context.Background(), "alice", "yadi", intent.ID); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("disabled yadi must fail closed, got %v", err)
	}
	_ = store
}

type fakeBatchReader struct{ batch batchfactoryv11.Batch }

func (r fakeBatchReader) GetBatch(_ context.Context, owner, id string) (batchfactoryv11.Batch, error) {
	if owner != "alice" || id != r.batch.ID {
		return batchfactoryv11.Batch{}, batchfactoryv11.ErrNotFound
	}
	return r.batch, nil
}

func TestIntentChecksBatchOwnershipBeforeCreatingExternalAction(t *testing.T) {
	service, _ := testService(Provider121, true, &countingProvider{})
	service.BatchReader = fakeBatchReader{batch: batchfactoryv11.Batch{ID: "batch-1", Books: []batchfactoryv11.Book{{ID: "book-1"}}}}
	if _, err := service.CreateIntent(context.Background(), "mallory", "121", "batch-1", "book-1", nil); !errors.Is(err, batchfactoryv11.ErrNotFound) {
		t.Fatalf("expected owner-scoped batch failure, got %v", err)
	}
	if _, err := service.CreateIntent(context.Background(), "alice", "121", "batch-1", "book-missing", nil); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected book ownership failure, got %v", err)
	}
}

func TestPublishResultsProjectsReadOnlyStatusPerProvider(t *testing.T) {
	ctx := context.Background()
	service, store := testService(Provider121, true, &countingProvider{})
	service.Providers[ProviderYadi] = &countingProvider{}
	service.Enabled[ProviderYadi] = true

	// Create intents directly on the store: Service.CreateIntent derives IDs
	// from wall-clock nanoseconds, which can collide on coarse Windows timers.
	now := time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)
	confirmedAt := now
	submittedAt := now.Add(time.Minute)
	for _, intent := range []SubmissionIntent{
		{ID: "intent-121-book-1", Owner: "alice", Provider: Provider121, BatchID: "batch-1", BookID: "book-1", Payload: []byte(`{"title":"121"}`), ExpiresAt: now.Add(time.Hour), ConfirmedAt: &confirmedAt},
		{ID: "intent-yadi-book-1", Owner: "alice", Provider: ProviderYadi, BatchID: "batch-1", BookID: "book-1", Payload: []byte(`{"title":"yadi"}`), ExpiresAt: now.Add(time.Hour), SubmittedAt: &submittedAt},
		{ID: "intent-121-book-2", Owner: "alice", Provider: Provider121, BatchID: "batch-1", BookID: "book-2", Payload: []byte(`{}`), ExpiresAt: now.Add(time.Hour)},
	} {
		if _, err := store.CreateIntent(ctx, intent); err != nil {
			t.Fatal(err)
		}
	}

	results, err := service.PublishResults(ctx, "alice", "batch-1", "book-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("results=%+v, want one per provider", results)
	}
	statuses := map[Provider]string{}
	for _, result := range results {
		statuses[result.Provider] = result.Status
	}
	if statuses[Provider121] != "confirmed" || statuses[ProviderYadi] != "submitted" {
		t.Fatalf("statuses=%v", statuses)
	}

	list, err := store.ListIntentsByBatch(ctx, "alice", "batch-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Fatalf("intents=%d, want 3 (book filtering must not hide rows)", len(list))
	}
	for _, intent := range list {
		if len(intent.Payload) != 0 {
			t.Fatalf("payload leaked through projection query: %q", intent.Payload)
		}
	}
}
