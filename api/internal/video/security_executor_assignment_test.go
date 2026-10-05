package video

import (
	"context"
	"errors"
	"testing"
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
