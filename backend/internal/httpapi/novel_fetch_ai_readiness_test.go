package httpapi

import (
	"qiantie/backend/internal/novelfetchworkshop"
	"testing"
)

func TestNovelFetchOriginalRawNeverCountsAsGeneratedAI(t *testing.T) {
	doc := novelfetchworkshop.Document{BookID: "test", Meta: map[string]any{"aiStatus": "failed"}, Versions: map[string]any{"original_raw": map[string]any{"state": "ready"}}, BodyRefs: map[string]novelfetchworkshop.BodyRef{"original_raw": {State: "ready"}, "original": {State: "ready"}}}
	view := legacyNovelFetchTask(doc)
	if files, ok := view["ai_files"].([]string); ok && len(files) > 0 {
		t.Fatalf("original counted as AI: %v", files)
	}
	if view["ai_status"] == "ai_done" {
		t.Fatal("failed task incorrectly marked AI done")
	}
	doc.BodyRefs["ai1"] = novelfetchworkshop.BodyRef{State: "ready"}
	view = legacyNovelFetchTask(doc)
	files, _ := view["ai_files"].([]string)
	if len(files) != 1 || files[0] != "ai1" {
		t.Fatalf("real AI1 result disappeared: %v", files)
	}
}
