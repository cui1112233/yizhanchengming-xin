package video

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSecurityLocalExecutorCannotCompleteUnassignedTask(t *testing.T) {
	ctx := context.Background()
	store := newMemoryLocalExecutorStore()
	provider := NewDoubaoLocalExecutorProvider(store)
	submitted, err := provider.Submit(ctx, SubmitRequest{
		Model:        ModelDoubaoSeedance,
		Prompt:       "security assignment test",
		SourceTaskID: "source-task-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	service := NewLocalExecutorService(store, nil)
	registered, err := service.Register(ctx, LocalExecutorRegistrationInput{
		Name:        "executor-a",
		ProviderKey: ProviderDoubaoLocalExecutor,
		Model:       ModelDoubaoSeedance,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = service.CompleteTask(ctx, registered.Token, submitted.ProviderJobID, LocalExecutorCompleteInput{
		ArtifactURL: "https://cdn.example/result.mp4",
	})
	if !errors.Is(err, ErrLocalExecutorUnauthorized) {
		t.Fatalf("CompleteTask error = %v, want ErrLocalExecutorUnauthorized for unassigned task", err)
	}
}

func TestSecurityLocalExecutorCannotFailUnassignedTask(t *testing.T) {
	ctx := context.Background()
	store := newMemoryLocalExecutorStore()
	provider := NewDoubaoLocalExecutorProvider(store)
	submitted, err := provider.Submit(ctx, SubmitRequest{
		Model:        ModelDoubaoSeedance,
		Prompt:       "security assignment test",
		SourceTaskID: "source-task-2",
	})
	if err != nil {
		t.Fatal(err)
	}

	service := NewLocalExecutorService(store, nil)
	registered, err := service.Register(ctx, LocalExecutorRegistrationInput{
		Name:        "executor-a",
		ProviderKey: ProviderDoubaoLocalExecutor,
		Model:       ModelDoubaoSeedance,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = service.FailTask(ctx, registered.Token, submitted.ProviderJobID, LocalExecutorFailInput{
		Code:    "provider_failed",
		Message: "failed",
	})
	if !errors.Is(err, ErrLocalExecutorUnauthorized) {
		t.Fatalf("FailTask error = %v, want ErrLocalExecutorUnauthorized for unassigned task", err)
	}
}

func TestSecurityExecutorACannotCompleteOrFailExecutorBAssignment(t *testing.T) {
	ctx := context.Background()
	store := newMemoryLocalExecutorStore()
	provider := NewDoubaoLocalExecutorProvider(store)
	service := NewLocalExecutorService(store, nil)

	a, err := service.Register(ctx, LocalExecutorRegistrationInput{Name: "executor-a", ProviderKey: ProviderDoubaoLocalExecutor, Model: ModelDoubaoSeedance})
	if err != nil { t.Fatal(err) }
	b, err := service.Register(ctx, LocalExecutorRegistrationInput{Name: "executor-b", ProviderKey: ProviderDoubaoLocalExecutor, Model: ModelDoubaoSeedance})
	if err != nil { t.Fatal(err) }
	submitted, err := provider.Submit(ctx, SubmitRequest{Model: ModelDoubaoSeedance, Prompt: "private task", SourceTaskID: "source-task-b"})
	if err != nil { t.Fatal(err) }
	assignMemoryLocalExecutorTask(t, store, submitted.ProviderJobID, b.Executor.ID)

	if err := service.CompleteTask(ctx, a.Token, submitted.ProviderJobID, LocalExecutorCompleteInput{ArtifactURL: "https://cdn.example/a.mp4"}); !errors.Is(err, ErrLocalExecutorUnauthorized) {
		t.Fatalf("executor A CompleteTask = %v, want unauthorized for executor B assignment", err)
	}
	if err := service.FailTask(ctx, a.Token, submitted.ProviderJobID, LocalExecutorFailInput{Code: "failed", Message: "guess"}); !errors.Is(err, ErrLocalExecutorUnauthorized) {
		t.Fatalf("executor A FailTask = %v, want unauthorized for executor B assignment", err)
	}
}

func TestSecurityUnknownExecutorCredentialCannotMutateGuessedTaskID(t *testing.T) {
	ctx := context.Background()
	store := newMemoryLocalExecutorStore()
	service := NewLocalExecutorService(store, nil)
	if err := service.CompleteTask(ctx, "unknown-executor-token", "guessed-task-id", LocalExecutorCompleteInput{ArtifactURL: "https://cdn.example/guess.mp4"}); !errors.Is(err, ErrLocalExecutorUnauthorized) {
		t.Fatalf("unknown token CompleteTask = %v, want unauthorized", err)
	}
	if err := service.FailTask(ctx, "unknown-executor-token", "guessed-task-id", LocalExecutorFailInput{Code: "failed", Message: "guess"}); !errors.Is(err, ErrLocalExecutorUnauthorized) {
		t.Fatalf("unknown token FailTask = %v, want unauthorized", err)
	}
}

func TestSecurityStaleExecutorCannotCompleteOrFailAssignedTask(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 10, 5, 14, 0, 0, 0, time.UTC)
	store := newMemoryLocalExecutorStore()
	provider := NewDoubaoLocalExecutorProvider(store)
	service := NewLocalExecutorService(store, func() time.Time { return now })
	registered, err := service.Register(ctx, LocalExecutorRegistrationInput{Name: "executor-stale", ProviderKey: ProviderDoubaoLocalExecutor, Model: ModelDoubaoSeedance})
	if err != nil { t.Fatal(err) }

	completeTask, err := provider.Submit(ctx, SubmitRequest{Model: ModelDoubaoSeedance, Prompt: "complete stale", SourceTaskID: "source-stale-complete"})
	if err != nil { t.Fatal(err) }
	failTask, err := provider.Submit(ctx, SubmitRequest{Model: ModelDoubaoSeedance, Prompt: "fail stale", SourceTaskID: "source-stale-fail"})
	if err != nil { t.Fatal(err) }
	assignMemoryLocalExecutorTask(t, store, completeTask.ProviderJobID, registered.Executor.ID)
	assignMemoryLocalExecutorTask(t, store, failTask.ProviderJobID, registered.Executor.ID)

	now = now.Add(LocalExecutorOnlineThreshold + time.Second)
	identity, err := service.Identity(ctx, registered.Token)
	if err != nil { t.Fatal(err) }
	if identity.Online {
		t.Fatal("fixture must be stale/offline before mutation attempt")
	}

	if err := service.CompleteTask(ctx, registered.Token, completeTask.ProviderJobID, LocalExecutorCompleteInput{ArtifactURL: "https://cdn.example/stale.mp4"}); !errors.Is(err, ErrLocalExecutorUnauthorized) {
		t.Fatalf("stale CompleteTask = %v, want unauthorized", err)
	}
	if err := service.FailTask(ctx, registered.Token, failTask.ProviderJobID, LocalExecutorFailInput{Code: "failed", Message: "stale"}); !errors.Is(err, ErrLocalExecutorUnauthorized) {
		t.Fatalf("stale FailTask = %v, want unauthorized", err)
	}
}
