package httpapi

import (
	"reflect"
	"testing"
)

func TestLegacyOperationResponseKeepsTopLevelPollingFields(t *testing.T) {
	item := map[string]any{"id": "op-1", "status": "done", "result": map[string]any{"ok": true}}
	got := legacyOperationResponse(item)
	if got["id"] != "op-1" || got["status"] != "done" {
		t.Fatalf("polling fields were not flattened: %#v", got)
	}
	if !reflect.DeepEqual(got["operation"], item) {
		t.Fatalf("nested operation was not preserved: %#v", got)
	}
}

func TestWorkshopPreviewKeepsSourceLineForSelectedBatch(t *testing.T) {
	docs, _, _, _ := parseWorkshopDocuments("2081147306087401012\t低保户妹妹非要看演唱会", 4000)
	if len(docs) != 1 || docs[0].Meta["sourceLine"] != "2081147306087401012\t低保户妹妹非要看演唱会" {
		t.Fatalf("source line was not preserved: %#v", docs)
	}
}
