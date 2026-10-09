package generation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/task9runtime"
)

const outcomeFailureMessage = "生成阶段执行失败，请稍后重试"

var outcomeCanaries = []string{
	"provider-canary-short", "password=pass-canary", "Cookie: session=cookie-canary",
	"Authorization: Bearer bearer-canary", "provider_api_key=key-canary",
	"user:dsn-canary@tcp(localhost:3306)/db",
}

func TestRuntimeOutcomeErrorMarksMissingAudioAsNonRetryable(t *testing.T) {
	err := NewRuntimeOutcomeError(ErrAudioMeasurementRequired)
	var safe task9runtime.SafeExecutionError
	if !errors.As(err, &safe) || safe.Retryable() {
		t.Fatalf("safe=%T %#v", safe, safe)
	}
}

func TestGenerationSafeOutcomeCauseCatalogue(t *testing.T) {
	for _, tc := range []struct {
		err     error
		message string
	}{
		{nil, ""}, {ErrInvalid, "生成参数无效，请检查后重试"},
		{ErrNotFound, "生成记录不存在，请刷新后重试"},
		{ErrConflict, "当前生成状态不允许此操作，请刷新后重试"},
		{ErrAudioMeasurementRequired, "请先生成或检测音频"},
		{ErrAudioProbeUnavailable, "音频检测服务暂不可用，请稍后重试"},
		{ErrTimelineValidation, "导演分镜时长校验失败，请重试导演阶段"},
		{ErrUnavailable, "生成服务暂不可用，请稍后重试"},
	} {
		t.Run(tc.message, func(t *testing.T) {
			cause := tc.err
			if cause != nil {
				cause = fmt.Errorf("%w: provider-canary-short", cause)
			}
			if got := safeError(cause); got != tc.message {
				t.Fatalf("public message = %q, want %q", got, tc.message)
			}
		})
	}
}

func TestGenerationSafeOutcomeProducers(t *testing.T) {
	for _, diagnostic := range append(append([]string{}, outcomeCanaries...), strings.Join(outcomeCanaries, "\n")) {
		t.Run(diagnostic, func(t *testing.T) {
			ctx := context.Background()
			s, store, provider := serviceFixture()
			cause := errors.New(diagnostic)
			provider.errors[2] = cause
			failed, err := s.RunBook(ctx, RunBookRequest{BatchProjectID: 3, BookID: 11, HookEnabled: true, RequestID: "execution-first"})
			if err != cause {
				t.Fatalf("internal cause lost: %v", err)
			}
			if failed.Error != outcomeFailureMessage || failed.Run.ErrorMessage != outcomeFailureMessage || failed.Latest[StageDirector].ErrorMessage != outcomeFailureMessage {
				t.Errorf("unsafe future outcome: %#v", failed)
			}
			scriptBefore, hookBefore := failed.Latest[StageScript], failed.Latest[StageHook]
			provider.errors[3] = cause
			retried, err := s.RetryStage(ctx, RetryStageRequest{BatchProjectID: 3, BookID: 11, Stage: StageDirector, RequestID: "execution-retry"})
			if err != cause || retried.Latest[StageDirector].Attempt != 2 || retried.Run.ID != failed.Run.ID {
				t.Fatalf("retry semantics changed: %#v %v", retried, err)
			}
			if retried.Error != outcomeFailureMessage || retried.Latest[StageDirector].ErrorMessage != outcomeFailureMessage {
				t.Errorf("unsafe retry: %#v", retried)
			}
			delete(provider.errors, 4)
			success, err := s.RetryStage(ctx, RetryStageRequest{BatchProjectID: 3, BookID: 11, Stage: StageDirector, RequestID: "execution-success"})
			if err != nil || success.Run.ErrorMessage != "" || success.Error != "" || success.Latest[StageDirector].ErrorMessage != "" || success.Latest[StageDirector].Attempt != 3 {
				t.Fatalf("retry did not clear error: %#v %v", success, err)
			}
			if !reflect.DeepEqual(success.Latest[StageScript], scriptBefore) || !reflect.DeepEqual(success.Latest[StageHook], hookBefore) {
				t.Fatal("retry changed prerequisite outputs")
			}
			for _, run := range store.bookRuns {
				if run.ErrorMessage != "" && run.ErrorMessage != outcomeFailureMessage {
					t.Errorf("persisted run error = %q", run.ErrorMessage)
				}
			}
			for _, stage := range store.stageRuns {
				if stage.ErrorMessage != "" && stage.ErrorMessage != outcomeFailureMessage {
					t.Errorf("persisted stage error = %q", stage.ErrorMessage)
				}
			}
		})
	}
}

func TestGenerationSafeOutcomeBatchOriginalCause(t *testing.T) {
	s, _, provider := serviceFixture()
	diagnostic := strings.Join(outcomeCanaries, "\n")
	provider.errors[0] = fmt.Errorf("%w: %s", ErrInvalid, diagnostic)
	result, err := s.RunBatch(context.Background(), RunBatchRequest{BatchProjectID: 3, RequestID: "batch-execution"})
	if !errors.Is(err, ErrUnavailable) || errors.Is(err, ErrInvalid) {
		t.Fatalf("aggregate classification changed: %v", err)
	}
	if !strings.Contains(err.Error(), diagnostic) || !strings.Contains(err.Error(), "book 11") {
		t.Errorf("original batch cause lost: %v", err)
	}
	if result.Failed != 1 || result.Completed != 2 || len(result.Books) != 3 || result.Books[0].Error != "生成参数无效，请检查后重试" {
		t.Errorf("batch outcomes = %#v", result)
	}
}

func TestGenerationSafeOutcomeAudioValidation(t *testing.T) {
	for _, providerFailure := range []bool{true, false} {
		t.Run(fmt.Sprint(providerFailure), func(t *testing.T) {
			s, store, provider := serviceFixture()
			s.audioProber = &fakeAudioProber{durationMS: 28000}
			if _, err := s.MeasureAudio(context.Background(), AudioMeasurementRequest{BatchProjectID: 3, BookID: 11, AudioAsset: "/audio/11.mp3"}); err != nil {
				t.Fatal(err)
			}
			message := "导演分镜时长校验失败，请重试导演阶段"
			if providerFailure {
				provider.errors[1] = errors.New(strings.Join(outcomeCanaries, "\n"))
				message = outcomeFailureMessage
			}
			provider.responses = []string{"SCRIPT", "invalid timeline"}
			result, err := s.RunBook(context.Background(), RunBookRequest{BatchProjectID: 3, BookID: 11, MatchAudio: true, RequestID: "audio-execution"})
			if err == nil {
				t.Fatal("expected validation failure")
			}
			if !providerFailure && !errors.Is(err, ErrTimelineValidation) {
				t.Fatalf("timeline cause = %v", err)
			}
			stage, _ := store.LatestStageRun(context.Background(), result.Run.ID, StageDirector)
			var validation map[string]any
			if json.Unmarshal([]byte(stage.ValidationResult), &validation) != nil {
				t.Fatal("missing validation facts")
			}
			if validation["error"] != message || validation["valid"] != false || validation["matchAudio"] != true || stage.ErrorMessage != message {
				t.Fatalf("validation = %#v, stage = %#v", validation, stage)
			}
			if !strings.Contains(stage.InputSnapshot, `"audioDurationSec":28`) || !strings.Contains(stage.InputSnapshot, `"script":"SCRIPT"`) {
				t.Fatal("retry inputs lost")
			}
		})
	}
}
