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

func TestGenerateAIRewritesUsesCompactInstructionBlock(t *testing.T) {
	store := NewMemoryStore()
	original := strings.Join([]string{
		"第一行原文。", "第二行原文。", "第三行原文。", "第四行原文。", "第五行原文。",
		"第六行原文。", "第七行原文。", "第八行原文。", "第九行原文。", "第十行原文。",
		"第十一行原文。", "第十二行原文。", "第十三行原文。", "第十四行原文。", "第十五行原文。",
		"第十六行原文。", "第十七行原文。", "第十八行原文。", "第十九行原文。", "第二十行原文。",
		"第二十一行衔接。", "第二十二行衔接。", "第二十三行衔接。", "第二十四行衔接。", "第二十五行衔接。", "第二十六行正文。",
	}, "\n")
	if err := store.PutDocument(context.Background(), "alice", Document{
		BookID:   "book-compact",
		Gender:   "女频",
		BookInfo: BookInfo{Category: "女频", Genre: "现代言情"},
		Meta:     map[string]any{"style": "现代通用", "bookName": "测试小说"},
		Original: original,
	}); err != nil {
		t.Fatal(err)
	}
	var seen RewriteRequest
	document, err := GenerateAIRewrites(context.Background(), store, "alice", "book-compact", 1, rewriteProviderFunc(func(_ context.Context, request RewriteRequest) (string, error) {
		seen = request
		return "改写第一行。\n改写第二行。", nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if seen.Temperature != 0.45 || seen.MaxTokens != 1200 {
		t.Fatalf("expected compact model settings, got temperature=%v maxTokens=%d", seen.Temperature, seen.MaxTokens)
	}
	for _, want := range []string{"待处理片段", "女频", "现代通用", "测试小说"} {
		if !strings.Contains(seen.UserPrompt, want) {
			t.Fatalf("prompt missing %q: %s", want, seen.UserPrompt)
		}
	}
	if strings.Contains(seen.UserPrompt, "第二十六行正文") {
		t.Fatalf("prompt sent text outside the target block: %s", seen.UserPrompt)
	}
	if got := document.Versions["ai1"]; got != "改写第一行。\n改写第二行。\n第二十一行衔接。\n第二十二行衔接。\n第二十三行衔接。\n第二十四行衔接。\n第二十五行衔接。\n第二十六行正文。" {
		t.Fatalf("unexpected merged rewrite: %q", got)
	}
}

func TestGenerateAIRewritesForVersionsOnlySavesExplicitSelections(t *testing.T) {
	store := NewMemoryStore()
	if err := store.PutDocument(context.Background(), "alice", Document{BookID: "book-selected", Original: "原文"}); err != nil {
		t.Fatal(err)
	}
	calls := 0
	document, err := GenerateAIRewritesForVersions(context.Background(), store, "alice", "book-selected", []string{"ai3", "ai3", "original"}, rewriteProviderFunc(func(_ context.Context, _ RewriteRequest) (string, error) {
		calls++
		return "ai3改写", nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 || document.Versions["ai3"] != "ai3改写" {
		t.Fatalf("only selected version should be generated: calls=%d versions=%#v", calls, document.Versions)
	}
	if _, exists := document.Versions["ai1"]; exists {
		t.Fatalf("unselected version was generated: %#v", document.Versions)
	}
}
