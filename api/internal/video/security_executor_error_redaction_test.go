package video

import (
	"context"
	"strings"
	"testing"
)

func TestSecurityExecutorFailureDoesNotPersistSensitiveDetail(t *testing.T) {
	ctx := context.Background()
	store := newMemoryLocalExecutorStore()
	provider := NewDoubaoLocalExecutorProvider(store)
	submitted, err := provider.Submit(ctx, SubmitRequest{
		Model:        ModelDoubaoSeedance,
		Prompt:       "security error redaction",
		SourceTaskID: "source-sensitive-error",
	})
	if err != nil {
		t.Fatal(err)
	}
	service := NewLocalExecutorService(store, nil)
	registered, err := service.Register(ctx, LocalExecutorRegistrationInput{
		Name:        "executor-redaction",
		ProviderKey: ProviderDoubaoLocalExecutor,
		Model:       ModelDoubaoSeedance,
	})
	if err != nil {
		t.Fatal(err)
	}
	assignMemoryLocalExecutorTask(t, store, submitted.ProviderJobID, registered.Executor.ID)

	const raw = "Authorization: Bearer executor-super-secret; payload=/srv/private/provider-request.json"
	if err := service.FailTask(ctx, registered.Token, submitted.ProviderJobID, LocalExecutorFailInput{Code: "provider_failed", Message: raw}); err != nil {
		t.Fatal(err)
	}
	task, err := store.GetLocalExecutorTask(ctx, submitted.ProviderJobID)
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(task.ErrorMessage)
	for _, forbidden := range []string{"executor-super-secret", "authorization", "bearer", "/srv/private"} {
		if strings.Contains(lower, strings.ToLower(forbidden)) {
			t.Fatalf("sensitive executor failure leaked into durable/browser-visible status: %q", task.ErrorMessage)
		}
	}
}
