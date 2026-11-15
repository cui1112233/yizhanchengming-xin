package novelfetchworkshop

import (
	"context"
	"strings"
	"testing"
)

func TestParseRewriteOutputRepairsTransportLineBreaksInsideStrings(t *testing.T) {
	// A real OpenAI-compatible gateway wrapped this object every ~30 characters
	// and inserted the breaks inside string values, including the book ID.
	raw := "{\n  \"id\": \"20945115735660195\n32\",\n  \"changed_lines\": [\n    {\n      \"line_no\": 1,\n      \n\"text\": \"我家浴室水管突然爆了，男友却说：你能干。\"\n    }\n  ]\n}"
	result, err := parseRewriteOutput(raw, "changed_lines_json", "2094511573566019532", 1)
	if err != nil {
		t.Fatalf("transport line breaks rejected: %v", err)
	}
	if result != "我家浴室水管突然爆了，男友却说：你能干。" {
		t.Fatalf("unexpected repair result: %q", result)
	}
}

func TestParseRewriteOutputAcceptsFencedAndPrefixedJSON(t *testing.T) {
	body := `{"id":"123","changed_lines":[{"line_no":1,"text":"改后首行"}]}`
	cases := []string{
		"```json\r\n" + body + "\r\n```",
		"以下是改写结果：\n" + body + "\n以上。",
	}
	for _, raw := range cases {
		result, err := parseRewriteOutput(raw, "changed_lines_json", "123", 1)
		if err != nil || result != "改后首行" {
			t.Fatalf("raw=%q result=%q err=%v", raw, result, err)
		}
	}
}

func TestRewriteSuccessClearsStaleFailureState(t *testing.T) {
	ctx := context.Background()
	store := plainRewriteTestStore("owner")
	_ = store.PutDocument(ctx, "owner", Document{
		BookID:   "123",
		Original: "原文",
		Meta: map[string]any{
			"error":                      "旧的 503 错误",
			"aiError":                    "旧的 503 错误",
			"autoSubmitReady":            true,
			"autoSubmitReadyVersions":    []any{"ai1"},
			"autoSubmitReadyFingerprint": "stale",
		},
	})
	if _, err := GenerateAIRewrites(ctx, store, "owner", "123", 2, rewriteProviderFunc(func(context.Context, RewriteRequest) (string, error) {
		return "改写正文", nil
	})); err != nil {
		t.Fatal(err)
	}
	doc, _ := store.GetDocument(ctx, "owner", "123")
	if _, exists := doc.Meta["error"]; exists {
		t.Fatalf("stale task error kept: %#v", doc.Meta)
	}
	if doc.Meta["aiError"] != "" {
		t.Fatalf("stale ai error kept: %#v", doc.Meta["aiError"])
	}
	versions := metaStringList(doc.Meta["autoSubmitReadyVersions"])
	if strings.Join(versions, ",") != "ai1,ai2" {
		t.Fatalf("ready versions not refreshed: %#v", doc.Meta["autoSubmitReadyVersions"])
	}
	if _, exists := doc.Meta["autoSubmitReadyFingerprint"]; exists {
		t.Fatalf("stale ready fingerprint kept: %#v", doc.Meta)
	}
}
