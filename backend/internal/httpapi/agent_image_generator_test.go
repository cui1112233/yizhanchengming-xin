package httpapi

import (
	"context"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
)

type agentImageReferenceProvider func(context.Context, agentworkspace.Asset, time.Duration) (string, error)

func (fn agentImageReferenceProvider) ReferenceURL(ctx context.Context, asset agentworkspace.Asset, ttl time.Duration) (string, error) {
	return fn(ctx, asset, ttl)
}

func TestAgentLocalDoubaoImageUsesSelectedModel(t *testing.T) {
	input := AgentImageGenerationInput{Provider: "doubao_local_executor", ModelID: "Seedream 5.0 Pro"}
	model, ok := agentLocalDoubaoImageModel(input)
	if !ok || model != "Seedream 5.0 Pro" {
		t.Fatalf("model=%q local=%v", model, ok)
	}
}

func TestAgentLocalDoubaoImageMapsCanvasQualityToSupportedModel(t *testing.T) {
	for quality, want := range map[string]string{
		"low":      "Seedream 5.0 Flash",
		"standard": "Seedream 4.5",
		"high":     "Seedream 5.0 Pro",
	} {
		model, ok := agentLocalDoubaoImageModel(AgentImageGenerationInput{
			Provider: "doubao_local_executor", ModelID: agentLocalDoubaoImageModelID, Quality: quality,
		})
		if !ok || model != want {
			t.Fatalf("quality=%q model=%q local=%v, want %q", quality, model, ok, want)
		}
	}
}

func TestAgentConfiguredImageDoesNotEnterLocalExecutor(t *testing.T) {
	input := AgentImageGenerationInput{Provider: "personal_api", ModelID: "configured-image"}
	if model, ok := agentLocalDoubaoImageModel(input); ok || model != "" {
		t.Fatalf("model=%q local=%v", model, ok)
	}
}

func TestAgentLocalDoubaoImagePromptRequestsAllSelectedImagesInOneSubmission(t *testing.T) {
	prompt := agentLocalDoubaoImagePrompt("雨夜街道", 4)
	if !strings.Contains(prompt, "雨夜街道") || !strings.Contains(prompt, "同一次生成中输出 4 张") {
		t.Fatalf("prompt=%q", prompt)
	}
	if got := agentLocalDoubaoImagePrompt("雨夜街道", 1); got != "雨夜街道" {
		t.Fatalf("single prompt=%q", got)
	}
}

func TestAgentLocalDoubaoImageResolvesProjectReferenceURLs(t *testing.T) {
	assets := agentworkspace.NewMemoryAssetStore()
	image := agentworkspace.Asset{ID: "reference-1", ProjectID: "project-1", Kind: "image", Name: "reference.png", ContentType: "image/png", ObjectKey: "agent-assets/alice/project-1/reference-1.png", ByteSize: 12, CreatedAt: time.Now()}
	if err := assets.CreateAsset(context.Background(), "alice", image); err != nil {
		t.Fatal(err)
	}
	urls, err := agentImageReferenceURLs(context.Background(), assets, agentImageReferenceProvider(func(_ context.Context, got agentworkspace.Asset, ttl time.Duration) (string, error) {
		if got.ID != image.ID || ttl != 15*time.Minute {
			t.Fatalf("asset=%+v ttl=%s", got, ttl)
		}
		return "https://assets.example/reference.png?signature=test", nil
	}), "alice", "project-1", []string{image.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 1 || urls[0] != "https://assets.example/reference.png?signature=test" {
		t.Fatalf("urls=%v", urls)
	}
}
