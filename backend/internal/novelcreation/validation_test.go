package novelcreation

import (
	"errors"
	"testing"
)

func TestNormalizeSendMessageInputRejectsBlank(t *testing.T) {
	_, err := NormalizeSendMessageInput(SendMessageInput{SessionID: "session-1", Content: " \n "})
	var domainErr *DomainError
	if !errors.As(err, &domainErr) || domainErr.Code != ErrInputRequired {
		t.Fatalf("err=%v, want %s", err, ErrInputRequired)
	}
}

func TestNormalizeSendMessageInputRejectsEachRequiredField(t *testing.T) {
	valid := SendMessageInput{
		SessionID:      "session-1",
		ProjectID:      "project-1",
		Content:        "继续写下一章",
		ModelID:        "model-1",
		ThinkingLevel:  "balanced",
		IdempotencyKey: "send-1",
	}

	tests := []struct {
		name  string
		input SendMessageInput
	}{
		{name: "session", input: func() SendMessageInput { in := valid; in.SessionID = " \t"; return in }()},
		{name: "content", input: func() SendMessageInput { in := valid; in.Content = "\n"; return in }()},
		{name: "model", input: func() SendMessageInput { in := valid; in.ModelID = "  "; return in }()},
		{name: "idempotency", input: func() SendMessageInput { in := valid; in.IdempotencyKey = " \n"; return in }()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeSendMessageInput(tt.input)
			var domainErr *DomainError
			if !errors.As(err, &domainErr) || domainErr.Code != ErrInputRequired {
				t.Fatalf("err=%v, want %s", err, ErrInputRequired)
			}
		})
	}
}

func TestNormalizeSendMessageInputTrimsAllStrings(t *testing.T) {
	got, err := NormalizeSendMessageInput(SendMessageInput{
		SessionID:      " session-1 ",
		ProjectID:      " project-1 ",
		Content:        " 继续写下一章 ",
		ModelID:        " model-1 ",
		ThinkingLevel:  " balanced ",
		IdempotencyKey: " send-1 ",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := SendMessageInput{
		SessionID:      "session-1",
		ProjectID:      "project-1",
		Content:        "继续写下一章",
		ModelID:        "model-1",
		ThinkingLevel:  "balanced",
		IdempotencyKey: "send-1",
	}
	if got.SessionID != want.SessionID || got.ProjectID != want.ProjectID || got.Content != want.Content ||
		got.ModelID != want.ModelID || got.ThinkingLevel != want.ThinkingLevel || got.IdempotencyKey != want.IdempotencyKey ||
		got.ModeInputs != nil {
		t.Fatalf("got=%+v, want=%+v", got, want)
	}
}

func TestNormalizeCreateProjectInputTrimsStringsAndDefaultsTargetWords(t *testing.T) {
	got, err := NormalizeCreateProjectInput(CreateProjectInput{
		Title:              " 迷雾记者 ",
		Genre:              " 都市悬疑 ",
		World:              " 近未来城市 ",
		Protagonist:        " 林默 ",
		CoreConflict:       " 真相与代价 ",
		ModelID:            " model-1 ",
		TargetChapterWords: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := CreateProjectInput{
		Title:              "迷雾记者",
		Genre:              "都市悬疑",
		World:              "近未来城市",
		Protagonist:        "林默",
		CoreConflict:       "真相与代价",
		ModelID:            "model-1",
		TargetChapterWords: 2000,
	}
	if got != want {
		t.Fatalf("got=%+v, want=%+v", got, want)
	}
}

func TestNormalizeCreateProjectInputAcceptsTargetWordBoundaries(t *testing.T) {
	for _, words := range []int{800, 8000} {
		got, err := NormalizeCreateProjectInput(CreateProjectInput{TargetChapterWords: words})
		if err != nil || got.TargetChapterWords != words {
			t.Fatalf("words=%d got=%+v err=%v", words, got, err)
		}
	}
}

func TestNormalizeCreateProjectInputRejectsTargetWordsOutsideRange(t *testing.T) {
	for _, words := range []int{799, 8001} {
		_, err := NormalizeCreateProjectInput(CreateProjectInput{TargetChapterWords: words})
		var domainErr *DomainError
		if !errors.As(err, &domainErr) || domainErr.Code != ErrInputRequired {
			t.Fatalf("words=%d err=%v, want %s", words, err, ErrInputRequired)
		}
	}
}

func TestCanTransitionRejectsSucceededToWriting(t *testing.T) {
	if CanTransition(RunSucceeded, RunWriting) {
		t.Fatal("terminal run must not return to writing")
	}
}

func TestNormalizeModeInputsRequiresModeSpecificFields(t *testing.T) {
	tests := []struct{ mode, key string }{
		{mode: "long_imitation", key: "referenceText"},
		{mode: "short_imitation", key: "referenceText"},
		{mode: "fanfiction", key: "existingSetting"},
		{mode: "continuation", key: "previousText"},
	}
	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			if _, err := NormalizeModeInputs(tt.mode, nil); !IsCode(err, ErrInputRequired) {
				t.Fatalf("missing input err=%v, want %s", err, ErrInputRequired)
			}
			got, err := NormalizeModeInputs(tt.mode, map[string]string{tt.key: "  内容  ", "unknown": "x"})
			if err != nil {
				t.Fatal(err)
			}
			if got[tt.key] != "内容" || len(got) != 1 {
				t.Fatalf("got=%v, want only %s", got, tt.key)
			}
		})
	}
}

func TestNormalizeModeInputsDropsExtraInputsForPlainModes(t *testing.T) {
	got, err := NormalizeModeInputs("long_original", map[string]string{"referenceText": "x"})
	if err != nil || got != nil {
		t.Fatalf("got=%v err=%v, plain modes ignore extra inputs", got, err)
	}
}

func TestCanTransitionAllowsConfirmedCheckpointToSucceed(t *testing.T) {
	if !CanTransition(RunAwaitingConfirmation, RunSucceeded) {
		t.Fatal("confirmed checkpoint must be able to reach succeeded")
	}
}

func TestCanTransitionAllowsNovelCreationPipeline(t *testing.T) {
	transitions := [][2]RunStatus{
		{RunDraft, RunClarifying},
		{RunClarifying, RunAwaitingConfirmation},
		{RunAwaitingConfirmation, RunQueued},
		{RunQueued, RunPlanning},
		{RunPlanning, RunWriting},
		{RunWriting, RunNormalizing},
		{RunNormalizing, RunAuditing},
		{RunAuditing, RunRevising},
		{RunAuditing, RunSettling},
		{RunRevising, RunAuditing},
		{RunSettling, RunIndexing},
		{RunIndexing, RunSucceeded},
		{RunIndexing, RunSucceededWithWarnings},
	}
	for _, transition := range transitions {
		if !CanTransition(transition[0], transition[1]) {
			t.Errorf("CanTransition(%q, %q) = false, want true", transition[0], transition[1])
		}
	}
}

func TestCanTransitionAllowsFailureOrCancellationFromActiveRuns(t *testing.T) {
	for _, from := range []RunStatus{
		RunDraft, RunClarifying, RunAwaitingConfirmation, RunQueued, RunPlanning,
		RunWriting, RunNormalizing, RunAuditing, RunRevising, RunSettling, RunIndexing,
	} {
		if !CanTransition(from, RunFailed) {
			t.Errorf("CanTransition(%q, %q) = false, want true", from, RunFailed)
		}
		if !CanTransition(from, RunCancelled) {
			t.Errorf("CanTransition(%q, %q) = false, want true", from, RunCancelled)
		}
	}
}

func TestCanTransitionRejectsLeavingTerminalRuns(t *testing.T) {
	for _, from := range []RunStatus{RunSucceeded, RunSucceededWithWarnings, RunFailed, RunCancelled} {
		if CanTransition(from, RunWriting) {
			t.Errorf("CanTransition(%q, %q) = true, want false", from, RunWriting)
		}
	}
}

func TestRunStatusValuesKeepStableLiteralContract(t *testing.T) {
	tests := []struct {
		name string
		got  RunStatus
		want string
	}{
		{name: "draft", got: RunDraft, want: "draft"},
		{name: "clarifying", got: RunClarifying, want: "clarifying"},
		{name: "awaiting confirmation", got: RunAwaitingConfirmation, want: "awaiting_confirmation"},
		{name: "queued", got: RunQueued, want: "queued"},
		{name: "planning", got: RunPlanning, want: "planning"},
		{name: "writing", got: RunWriting, want: "writing"},
		{name: "normalizing", got: RunNormalizing, want: "normalizing"},
		{name: "auditing", got: RunAuditing, want: "auditing"},
		{name: "revising", got: RunRevising, want: "revising"},
		{name: "settling", got: RunSettling, want: "settling"},
		{name: "indexing", got: RunIndexing, want: "indexing"},
		{name: "succeeded", got: RunSucceeded, want: "succeeded"},
		{name: "succeeded with warnings", got: RunSucceededWithWarnings, want: "succeeded_with_warnings"},
		{name: "failed", got: RunFailed, want: "failed"},
		{name: "cancelled", got: RunCancelled, want: "cancelled"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Fatalf("got %q, want stable literal %q", tt.got, tt.want)
			}
		})
	}
}

func TestNovelErrorCodesKeepStableLiteralContract(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "input required", got: ErrInputRequired, want: "novel_input_required"},
		{name: "project not found", got: ErrProjectNotFound, want: "novel_project_not_found"},
		{name: "session not found", got: ErrSessionNotFound, want: "novel_session_not_found"},
		{name: "run in progress", got: ErrRunInProgress, want: "novel_run_in_progress"},
		{name: "project locked", got: ErrProjectLocked, want: "novel_project_locked"},
		{name: "invalid transition", got: ErrInvalidTransition, want: "novel_invalid_transition"},
		{name: "model unavailable", got: ErrModelUnavailable, want: "novel_model_unavailable"},
		{name: "generation timeout", got: ErrGenerationTimeout, want: "novel_generation_timeout"},
		{name: "persistence failed", got: ErrPersistenceFailed, want: "novel_persistence_failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("got %q, want stable literal %q", tt.got, tt.want)
			}
		})
	}
}

func TestCanTransitionRejectsInvalidJumps(t *testing.T) {
	tests := []struct {
		name string
		from RunStatus
		to   RunStatus
	}{
		{name: "draft skips to writing", from: RunDraft, to: RunWriting},
		{name: "draft skips to succeeded", from: RunDraft, to: RunSucceeded},
		{name: "clarifying skips confirmation", from: RunClarifying, to: RunWriting},
		{name: "queued skips planning", from: RunQueued, to: RunWriting},
		{name: "planning skips writing", from: RunPlanning, to: RunNormalizing},
		{name: "writing skips normalizing", from: RunWriting, to: RunAuditing},
		{name: "normalizing skips auditing", from: RunNormalizing, to: RunSettling},
		{name: "auditing skips revision or settlement", from: RunAuditing, to: RunSucceeded},
		{name: "revising returns to writing", from: RunRevising, to: RunWriting},
		{name: "settling skips indexing", from: RunSettling, to: RunSucceeded},
		{name: "indexing returns to writing", from: RunIndexing, to: RunWriting},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if CanTransition(tt.from, tt.to) {
				t.Fatalf("CanTransition(%q, %q) = true, want false", tt.from, tt.to)
			}
		})
	}
}

func TestCanTransitionRejectsUnknownStatuses(t *testing.T) {
	unknown := RunStatus("unknown")
	tests := [][2]RunStatus{
		{unknown, RunDraft},
		{RunDraft, unknown},
		{unknown, unknown},
	}
	for _, transition := range tests {
		if CanTransition(transition[0], transition[1]) {
			t.Errorf("CanTransition(%q, %q) = true, want false", transition[0], transition[1])
		}
	}
}

func TestCanTransitionRejectsTransitionsBetweenDistinctTerminalRuns(t *testing.T) {
	terminals := []RunStatus{RunSucceeded, RunSucceededWithWarnings, RunFailed, RunCancelled}
	for _, from := range terminals {
		for _, to := range terminals {
			if from == to {
				continue
			}
			if CanTransition(from, to) {
				t.Errorf("CanTransition(%q, %q) = true, want false", from, to)
			}
		}
	}
}

func TestCanTransitionAllowsTerminalSelfLoopsAsIdempotent(t *testing.T) {
	for _, terminal := range []RunStatus{RunSucceeded, RunSucceededWithWarnings, RunFailed, RunCancelled} {
		if !CanTransition(terminal, terminal) {
			t.Errorf("CanTransition(%q, %q) = false, want true for idempotent self-loop", terminal, terminal)
		}
	}
}
