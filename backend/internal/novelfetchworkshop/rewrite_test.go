package novelfetchworkshop

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type rewriteProviderFunc func(context.Context, RewriteRequest) (string, error)

func (f rewriteProviderFunc) Complete(ctx context.Context, request RewriteRequest) (string, error) {
	return f(ctx, request)
}

func TestGenerateAIRewritesSavesVersionsAndBackendPrompt(t *testing.T) {
	store := NewMemoryStore()
	if err := store.PutDocument(context.Background(), "alice", Document{BookID: "book-1", Original: "原文第一段"}); err != nil {
		t.Fatal(err)
	}
	var seen RewriteRequest
	document, err := GenerateAIRewrites(context.Background(), store, "alice", "book-1", 2, rewriteProviderFunc(func(_ context.Context, request RewriteRequest) (string, error) {
		seen = request
		return "改写后的正文", nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if document.Versions["ai1"] != "改写后的正文" || document.Versions["ai2"] != "改写后的正文" {
		t.Fatalf("versions not saved: %#v", document.Versions)
	}
	if document.Meta["aiStatus"] != "done" || document.Meta["aiGeneratedCount"] != 2 {
		t.Fatalf("unexpected status: %#v", document.Meta)
	}
	if !strings.Contains(seen.SystemPrompt, "保持人物身份") || !strings.Contains(seen.UserPrompt, "原文第一段") {
		t.Fatalf("backend prompt did not include rewrite contract: %#v", seen)
	}
}

func TestGenerateAIRewritesDoesNotSaveEmptyOrFailedVersion(t *testing.T) {
	store := NewMemoryStore()
	if err := store.PutDocument(context.Background(), "alice", Document{BookID: "book-2", Original: "原文"}); err != nil {
		t.Fatal(err)
	}
	_, err := GenerateAIRewrites(context.Background(), store, "alice", "book-2", 1, rewriteProviderFunc(func(context.Context, RewriteRequest) (string, error) {
		return "", errors.New("模型不可用")
	}))
	if err == nil || err.Error() != "模型不可用" {
		t.Fatalf("expected provider error, got %v", err)
	}
	document, getErr := store.GetDocument(context.Background(), "alice", "book-2")
	if getErr != nil {
		t.Fatal(getErr)
	}
	if _, exists := document.Versions["ai1"]; exists {
		t.Fatalf("failed rewrite must not save version: %#v", document.Versions)
	}
	if document.Meta["aiStatus"] != "failed" {
		t.Fatalf("failed status missing: %#v", document.Meta)
	}
}
