package httpapi

import (
	"context"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/screenplay"
)

func TestScreenplayTextGeneratorOutputBudget(t *testing.T) {
	for _, tc := range []struct {
		name        string
		input, want int
	}{
		{"omitted", 0, 16000},
		{"negative", -1, 16000},
		{"explicit", 6000, 6000},
		{"explicit-small", 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			provider := screenplayProviderFunc(func(_ context.Context, req batchfactoryv11.TextCompletionRequest) (string, error) {
				calls++
				if req.MaxTokens != tc.want {
					t.Fatalf("provider output budget = %d, want %d", req.MaxTokens, tc.want)
				}
				if req.SystemPrompt != "system" || req.UserPrompt != "user" || req.Temperature != 0.4 || req.DisableJSONResponse {
					t.Fatalf("other completion fields changed: %#v", req)
				}
				return "generated", nil
			})
			generator := screenplayTextGenerator{owner: "alice", resolver: func(_ context.Context, owner, model string) (batchfactoryv11.DirectorProvider, error) {
				if owner != "alice" || model != "selected-text" {
					t.Fatalf("resolver selection = %s/%s", owner, model)
				}
				return provider, nil
			}}
			out, err := generator.Generate(context.Background(), screenplay.TextRequest{Model: "selected-text", SystemPrompt: "system", UserPrompt: "user", JSONResponse: true, Temperature: 0.4, MaxOutputTokens: tc.input})
			if err != nil || out.Content != "generated" || calls != 1 {
				t.Fatalf("completion result = %#v, calls = %d, error = %v", out, calls, err)
			}
		})
	}
}
