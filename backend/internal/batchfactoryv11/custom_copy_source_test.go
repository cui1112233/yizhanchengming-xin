package batchfactoryv11

import (
	"context"
	"strings"
	"testing"
)

func TestCustomCopyKeepsCompleteProductionText(t *testing.T) {
	text := "第一行\n第二行\n第三行\n第四行\n第五行\n第六行\n故事完整结尾。"
	const id = "2100242549395006289"
	parsed, err := ParseGroupedManualBookLists(ManualIntakeInput{ParseMode: "smart", Groups: []ManualIntakeGroup{{PlatformID: "15", PlatformName: "知乎付费", InputText: id, SourceTextByBookID: map[string]string{id: text}, SourceMetadataByBookID: map[string]map[string]any{id: {"bookName": "青霄赴", "sourceMode": "custom_copy", "customCopyRecordId": "280893"}}}}})
	if err != nil || len(parsed) != 1 {
		t.Fatalf("parse: %v", err)
	}
	if parsed[0].BookID != id || parsed[0].Title != "青霄赴" || parsed[0].SourceMetadata["customCopyRecordId"] != "280893" {
		t.Fatalf("lost selection: %+v", parsed[0])
	}
	if parsed[0].SourceMetadata["sourceContentVersion"] != "original" || parsed[0].SourceMetadata["sensitiveStatus"] != "skipped" {
		t.Fatal("custom copy must use original without cleanup")
	}
	book := Book{SourceText: parsed[0].SourceText, SourceMetadata: parsed[0].SourceMetadata}
	book.SourceMetadata["contentRangeLines"] = 5
	if got := productionTextForBook(book); got != text {
		t.Fatalf("custom copy truncated: %q", got)
	}
	service := &DirectorService{Store: NewMemoryStore()}
	produced, err := service.productionBook(context.Background(), "alice", "batch-test", book)
	if err != nil || produced.SourceText != text {
		t.Fatalf("production book lost copy: %q %v", produced.SourceText, err)
	}
	book.SourceText = strings.Repeat("完整句子。\n", 600)
	comic, err := service.comicProductionBook(context.Background(), "alice", "batch-test", book)
	if err != nil || comic.SourceText != strings.TrimSpace(book.SourceText) {
		t.Fatal("comic copy truncated", err)
	}
	book.SourceText = text
	book.SourceMetadata["sourceMode"] = "manual_original"
	if got := productionTextForBook(book); got == text {
		t.Fatal("normal source range unexpectedly ignored")
	}
}
