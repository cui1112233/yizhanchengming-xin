package batchfactoryexecution

import (
	"strings"
	"testing"
	"time"
)

func TestObservabilityEventContainsCorrelationFieldsAndRedactsOwner(t *testing.T) {
	event := NewObservabilityEvent("execution_reconcile", "alice@example.com")
	event.ExecutionID = "execution-1"
	event.BatchID = "batch-1"
	event.BookID = "book-1"
	event.StageKey = "publish"
	event.AttemptNo = 2
	event.InputFingerprint = "sha256:input"
	event.LegacyJobID = "legacy-1"
	event.Status = "reconciling"
	event.DurationMS = 31

	encoded := event.JSON()
	for _, required := range []string{
		`"execution_id":"execution-1"`, `"owner_hash":"`, `"batch_id":"batch-1"`,
		`"book_id":"book-1"`, `"stage_key":"publish"`, `"attempt_no":2`,
		`"input_fingerprint":"sha256:input"`, `"legacy_job_id":"legacy-1"`,
		`"status":"reconciling"`, `"duration_ms":31`,
	} {
		if !strings.Contains(encoded, required) {
			t.Fatalf("event %s does not contain %s", encoded, required)
		}
	}
	if strings.Contains(encoded, "alice@example.com") {
		t.Fatalf("event leaked raw owner: %s", encoded)
	}
}

func TestObservabilityEventHasNoPayloadSurfaceForSecrets(t *testing.T) {
	event := NewObservabilityEvent("transition_rejected", "owner")
	event.Status = "rejected"
	event.DurationMS = time.Millisecond.Milliseconds()
	encoded := strings.ToLower(event.JSON())

	for _, forbidden := range []string{
		"api_key", "apikey", "cookie", "password", "source_text",
		"prompt_body", "provider_payload", "authorization", "bearer ",
	} {
		if strings.Contains(encoded, forbidden) {
			t.Fatalf("event exposed forbidden field %q: %s", forbidden, encoded)
		}
	}
}

func TestCompletionInvariantRequiresConfirmedPublicationReceipt(t *testing.T) {
	confirmedAt := time.Now().UTC()
	confirmed := PublicationReceipt{RemoteWorkID: "work-1", RemoteStatus: "published", ConfirmedAt: &confirmedAt}

	if CanCompleteBook("published_confirmed", nil) {
		t.Fatal("published target completed without a receipt")
	}
	if CanCompleteBook("published_confirmed", &PublicationReceipt{RemoteWorkID: "work-1", RemoteStatus: "published"}) {
		t.Fatal("published target completed without confirmed readback")
	}
	if !CanCompleteBook("published_confirmed", &confirmed) {
		t.Fatal("confirmed publication receipt should satisfy completion invariant")
	}
	if !CanCompleteBook("video_ready", nil) {
		t.Fatal("non-publication target should not require a publication receipt")
	}
}
