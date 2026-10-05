package generation

import (
	"context"
	"strings"
	"testing"
)

func TestHookDisabledDoesNotRequireHookPrompt(t *testing.T) {
	service, store, provider := serviceFixture()
	delete(store.prompts, PromptHook)
	provider.responses = []string{"SCRIPT", `{"cards":[]}`}

	result, err := service.RunBook(context.Background(), RunBookRequest{
		BatchProjectID: 3,
		BookID: 11,
		HookEnabled: false,
		DirectorMode: DirectorNormal,
		RequestID: "hook-off-no-prompt",
	})
	if err != nil {
		t.Fatalf("Hook disabled must skip without resolving Hook prompt: %v", err)
	}
	hook, err := store.LatestStageRun(context.Background(), result.Run.ID, StageHook)
	if err != nil {
		t.Fatal(err)
	}
	if hook.Status != StatusSkipped {
		t.Fatalf("hook status = %s", hook.Status)
	}
}

func TestH3DirectorRejectsUnstructuredOutputAndPersistsFailure(t *testing.T) {
	service, store, provider := serviceFixture()
	provider.responses = []string{"SCRIPT", "HOOK", "not-json"}

	result, err := service.RunBook(context.Background(), RunBookRequest{
		BatchProjectID: 3,
		BookID: 11,
		HookEnabled: true,
		DirectorMode: DirectorH3,
		RequestID: "h3-invalid",
	})
	if err == nil {
		t.Fatal("expected invalid H3 output to fail")
	}
	director, readErr := store.LatestStageRun(context.Background(), result.Run.ID, StageDirector)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if director.Status != StatusFailed {
		t.Fatalf("director status = %s", director.Status)
	}
	if strings.TrimSpace(director.ErrorMessage) == "" {
		t.Fatal("H3 format failure must be persisted")
	}
}

func TestFinalPromptSnapshotPreservesCompileContextForRetry(t *testing.T) {
	service, store, _ := serviceFixture()
	result, err := service.RunBook(context.Background(), RunBookRequest{
		BatchProjectID: 3,
		BookID: 11,
		HookEnabled: true,
		DirectorMode: DirectorNormal,
		RequestID: "final-context",
		ProcessingRules: "RULE-X",
		KnowledgeBase: "KB-X",
		ProjectConfig: "PROJECT-X",
		UserConfig: "USER-X",
		ModelConfig: "MODEL-X",
	})
	if err != nil {
		t.Fatal(err)
	}
	final, err := store.LatestStageRun(context.Background(), result.Run.ID, StageFinalPrompt)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"RULE-X", "KB-X", "PROJECT-X", "USER-X", "MODEL-X"} {
		if !strings.Contains(final.InputSnapshot, value) {
			t.Fatalf("final input snapshot lost %s: %s", value, final.InputSnapshot)
		}
	}

	failed := final
	failed.ID = 0
	failed.Attempt = final.Attempt + 1
	failed.Status = StatusFailed
	failed.OutputText = ""
	failed.ErrorMessage = "compiler interrupted"
	if _, err := store.CreateStageRun(context.Background(), failed); err != nil {
		t.Fatal(err)
	}
	run := result.Run
	run.Status = StatusFailed
	run.ErrorMessage = "compiler interrupted"
	if _, err := store.UpdateBookRun(context.Background(), run); err != nil {
		t.Fatal(err)
	}

	retried, err := service.RetryStage(context.Background(), RetryStageRequest{
		BatchProjectID: 3,
		BookID: 11,
		Stage: StageFinalPrompt,
		RequestID: "retry-final-context",
	})
	if err != nil {
		t.Fatal(err)
	}
	latest := retried.Latest[StageFinalPrompt]
	for _, value := range []string{"RULE-X", "KB-X", "PROJECT-X", "USER-X", "MODEL-X"} {
		if !strings.Contains(latest.OutputText, value) {
			t.Fatalf("retried final prompt lost %s: %s", value, latest.OutputText)
		}
	}
}
