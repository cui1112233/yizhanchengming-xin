package agentstudio

import (
	"context"
	"errors"
	"testing"
)

func TestContinuePersistsUnavailableExecutionWithoutFabricatingReply(t *testing.T) {
	store := newMemoryStore()
	service := NewService(store, UnavailableExecutor{})
	project, err := service.CreateProject(context.Background(), Actor{UserID: 7}, CreateProjectInput{Title: "新项目"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	result, err := service.Continue(context.Background(), Actor{UserID: 7}, project.ID, ContinueInput{Content: "写一个开头"})
	if !errors.Is(err, ErrExecutorUnavailable) {
		t.Fatalf("error = %v, want ErrExecutorUnavailable", err)
	}
	if result.Execution.Status != ExecutionUnavailable || result.Execution.ErrorCode != "executor_unavailable" {
		t.Fatalf("execution = %#v", result.Execution)
	}
	if len(store.messages[project.ID]) != 1 || store.messages[project.ID][0].Role != RoleUser {
		t.Fatalf("messages = %#v", store.messages[project.ID])
	}
	if len(store.messages[project.ID]) > 1 {
		t.Fatal("unavailable executor must not fabricate an assistant message")
	}
}
