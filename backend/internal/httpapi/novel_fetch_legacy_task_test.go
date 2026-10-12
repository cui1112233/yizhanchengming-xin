package httpapi

import (
	"testing"

	"qiantie/backend/internal/novelfetchworkshop"
)

func TestLegacyNovelFetchTaskPreservesSavedMetadataAndStatuses(t *testing.T) {
	task := novelfetchworkshop.Document{
		BookID:           "2081147306087401012",
		Gender:           "女频",
		OriginalRawChars: 4711,
		ProcessedChars:   3982,
		MaxTxt:           4000,
		Meta: map[string]any{
			"bookName":     "低保户妹妹非要看演唱会",
			"platformName": "知乎付费",
			"aiStatus":     "done",
		},
		Versions: map[string]any{
			"original_raw": map[string]any{"state": "ready"},
			"ai1":          map[string]any{"state": "ready"},
		},
	}
	legacy := legacyNovelFetchTask(task)
	for key, want := range map[string]string{
		"id": "2081147306087401012", "book_id": "2081147306087401012",
		"book_name": "低保户妹妹非要看演唱会", "platform_name": "知乎付费",
		"gender": "女频", "classify_status": "classified", "original_status": "done",
		"ai_status": "ai_done",
	} {
		if got := legacy[key]; got != want {
			t.Fatalf("%s = %#v, want %q", key, got, want)
		}
	}
	if got := legacy["original_raw_chars"]; got != 4711 {
		t.Fatalf("original_raw_chars = %#v, want 4711", got)
	}
}
