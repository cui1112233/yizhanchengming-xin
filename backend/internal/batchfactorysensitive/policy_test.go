package batchfactorysensitive

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestResolveModeDefaultsAndLegacyCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		settings batchfactoryv11.SettingsPatch
		want     Mode
	}{
		{name: "default", want: ModeSystem},
		{name: "legacy AI", settings: batchfactoryv11.SettingsPatch{"sensitiveAIEnabled": json.RawMessage(`true`)}, want: ModeAI},
		{name: "legacy rules", settings: batchfactoryv11.SettingsPatch{"sensitiveAIEnabled": json.RawMessage(`false`)}, want: ModeSystem},
		{name: "explicit off", settings: batchfactoryv11.SettingsPatch{"sensitiveMode": json.RawMessage(`"off"`), "sensitiveAIEnabled": json.RawMessage(`true`)}, want: ModeOff},
		{name: "explicit system", settings: batchfactoryv11.SettingsPatch{"sensitiveMode": json.RawMessage(`"system"`), "sensitiveAIEnabled": json.RawMessage(`true`)}, want: ModeSystem},
		{name: "explicit AI", settings: batchfactoryv11.SettingsPatch{"sensitiveMode": json.RawMessage(`"ai"`), "sensitiveAIEnabled": json.RawMessage(`false`)}, want: ModeAI},
		{name: "unknown mode falls back to legacy setting", settings: batchfactoryv11.SettingsPatch{"sensitiveMode": json.RawMessage(`"unknown"`), "sensitiveAIEnabled": json.RawMessage(`true`)}, want: ModeAI},
		{name: "invalid settings fall back to default", settings: batchfactoryv11.SettingsPatch{"sensitiveMode": json.RawMessage(`true`), "sensitiveAIEnabled": json.RawMessage(`"true"`)}, want: ModeSystem},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveMode(tt.settings); got != tt.want {
				t.Fatalf("ResolveMode(%v) = %q, want %q", tt.settings, got, tt.want)
			}
		})
	}
}

type ruleFunc func(context.Context, Input) (string, []Change, error)

func (f ruleFunc) ProcessRules(ctx context.Context, input Input) (string, []Change, error) {
	return f(ctx, input)
}

type aiFunc func(context.Context, Input) (string, []Change, error)

func (f aiFunc) ProcessAI(ctx context.Context, input Input) (string, []Change, error) {
	return f(ctx, input)
}

func TestProcessDispatchesOnlySelectedMode(t *testing.T) {
	tests := []struct {
		mode Mode
		want string
	}{
		{ModeSystem, "规则文案"},
		{ModeAI, "AI 文案"},
		{ModeOff, " 原文\r\n"},
	}
	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			input := Input{Mode: tt.mode, Owner: "alice", BookID: "book-1", Text: " 原文\r\n", TextModelID: "model-1"}
			changes := []Change{{Keyword: "原文", Before: input.Text, After: tt.want}}
			if tt.mode == ModeOff {
				changes = nil
			}
			p := Processor{
				Rules: ruleFunc(func(gotCtx context.Context, got Input) (string, []Change, error) {
					if tt.mode != ModeSystem || gotCtx != ctx || got != input {
						t.Fatalf("unexpected rules call with input %+v", got)
					}
					return "规则文案", changes, nil
				}),
				AI: aiFunc(func(gotCtx context.Context, got Input) (string, []Change, error) {
					if tt.mode != ModeAI || gotCtx != ctx || got != input {
						t.Fatalf("unexpected AI call with input %+v", got)
					}
					return "AI 文案", changes, nil
				}),
			}
			result, err := p.Process(ctx, input)
			if err != nil {
				t.Fatal(err)
			}
			if result.Mode != tt.mode || result.OriginalText != input.Text ||
				result.ProductionText != tt.want || !slices.Equal(result.Changes, changes) {
				t.Fatalf("unexpected result %+v", result)
			}
		})
	}
}

func TestAIFailureDoesNotFallBack(t *testing.T) {
	wantErr := errors.New("AI unavailable")
	p := Processor{
		Rules: ruleFunc(func(context.Context, Input) (string, []Change, error) {
			t.Fatal("rules called after AI error")
			return "", nil, nil
		}),
		AI: aiFunc(func(context.Context, Input) (string, []Change, error) {
			return "partial result", nil, wantErr
		}),
	}
	result, err := p.Process(context.Background(), Input{Mode: ModeAI, Text: "原文"})
	if !errors.Is(err, wantErr) || result.ProductionText != "" {
		t.Fatalf("Process() = %+v, %v, want AI error and no production text", result, err)
	}
}

func TestProcessRejectsUnsupportedMode(t *testing.T) {
	p := Processor{
		Rules: ruleFunc(func(context.Context, Input) (string, []Change, error) {
			t.Fatal("rules called with invalid mode")
			return "", nil, nil
		}),
		AI: aiFunc(func(context.Context, Input) (string, []Change, error) {
			t.Fatal("AI called with invalid mode")
			return "", nil, nil
		}),
	}
	for _, mode := range []Mode{"", "unknown"} {
		if _, err := p.Process(context.Background(), Input{Mode: mode}); err == nil {
			t.Fatalf("Process() accepted invalid mode %q", mode)
		}
	}
}

func TestProcessRequiresSelectedProcessor(t *testing.T) {
	for _, mode := range []Mode{ModeAI, ModeSystem} {
		if _, err := (Processor{}).Process(context.Background(), Input{Mode: mode}); err == nil {
			t.Fatalf("Process() accepted missing processor for mode %q", mode)
		}
	}
}
