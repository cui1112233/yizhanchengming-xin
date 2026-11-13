package httpapi

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestQueueStatusDoesNotRepeatCompletedBookBodies(t *testing.T) {
	job := map[string]any{"id": "job", "status": "completed", "updated_at": "now", "result": map[string]any{"tasks": []any{map[string]any{"originalRaw": strings.Repeat("正文", 100000)}}}, "logs": []any{strings.Repeat("日志", 100000)}}
	before, _ := json.Marshal(job)
	after, _ := json.Marshal(novelFetchJobStatusSummary(job))
	if len(after) > 1024 || len(after)*100 > len(before) {
		t.Fatalf("queue response remains large: before=%d after=%d", len(before), len(after))
	}
	if novelFetchJobStatusSummary(job)["status"] != "completed" || job["result"] == nil {
		t.Fatal("status or original detail lost")
	}
}
