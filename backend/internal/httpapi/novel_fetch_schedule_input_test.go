package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

func TestNovelFetchScheduleRejectsInvalidInputWithoutSaving(t *testing.T) {
	for _, tc := range []struct {
		name, field, value, message string
	}{
		{"empty_books", "input_text", "  ", "书籍清单"},
		{"empty_time", "run_at", "", "时间"},
		{"invalid_time", "run_at", "tomorrow", "时间"},
		{"no_book_ids", "input_text", "只有书名没有编号", "有效书籍"},
		{"unknown_platform", "platform_id", "not-a-platform", "平台"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := novelfetchworkshop.NewMemoryStore()
			now := time.Unix(1_800_000_000, 0)
			handler := NewRouter(RouterOptions{BridgeSecret: "schedule-input", Now: func() time.Time { return now }, NovelFetchStore: store})
			input := map[string]any{"input_text": "12345678901 测试书籍", "run_at": "2027-01-01T12:00", "platform_id": "2"}
			input[tc.field] = tc.value
			r := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/schedules", input, now, "schedule-input")
			if r.Code != http.StatusBadRequest || !strings.Contains(r.Body.String(), tc.message) {
				t.Fatalf("invalid schedule accepted or unclear error: status=%d body=%s", r.Code, r.Body.String())
			}
			cfg, _ := store.GetConfig(context.Background(), "tester")
			if entries, _ := cfg["novel_fetch_schedules"].([]any); len(entries) != 0 {
				t.Fatalf("invalid schedule persisted: %#v", entries)
			}
		})
	}
}

func TestNovelFetchScheduleNormalizesTimeAndPreservesExecutionSettings(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	now := time.Unix(1_800_000_000, 0)
	handler := NewRouter(RouterOptions{BridgeSecret: "schedule-input", Now: func() time.Time { return now }, NovelFetchStore: store})
	r := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/schedules", map[string]any{
		"inputText": "12345678901 测试书籍", "runAt": "2027-01-01T12:00", "platformId": "2",
		"selected_versions": []string{"ai1", "ai2"}, "text_model_id": "chosen-model", "rewrite_methods": map[string]any{"ai2": "opening"},
	}, now, "schedule-input")
	if r.Code != http.StatusOK {
		t.Fatalf("valid schedule rejected: %d %s", r.Code, r.Body.String())
	}
	var response struct {
		Schedule map[string]any `json:"schedule"`
	}
	if err := json.Unmarshal(r.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	item := response.Schedule
	if item["run_at"] != "2027-01-01T04:00:00Z" || item["platform_id"] != "2" || item["input_text"] != "12345678901 测试书籍" {
		t.Fatalf("canonical input/time not saved: %#v", item)
	}
	if item["text_model_id"] != "chosen-model" || item["auto_submit"] != true || len(scheduleStringSlice(item["selected_versions"])) != 2 || scheduleStringMap(item["rewrite_methods"])["ai2"] != "opening" {
		t.Fatalf("execution choices lost: %#v", item)
	}
}

func TestNovelFetchInvalidHistoricalScheduleFailsClearlyWithoutQueueing(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	if err := store.PutConfig(ctx, "tester", map[string]any{"novel_fetch_schedules": []any{map[string]any{
		"id": "broken-time", "status": "scheduled", "run_at": "broken", "input_text": "12345678901 测试书籍",
	}}}); err != nil {
		t.Fatal(err)
	}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", time.Now()); err != nil {
		t.Fatal(err)
	}
	cfg, _ := store.GetConfig(ctx, "tester")
	item := cfg["novel_fetch_schedules"].([]any)[0].(map[string]any)
	if item["status"] != "failed" || !strings.Contains(firstText(item["error"]), "时间") {
		t.Fatalf("historical schedule silently left waiting: %#v", item)
	}
	claimCtx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	if job, err := queue.Claim(claimCtx, 10*time.Millisecond); err == nil {
		t.Fatalf("invalid schedule queued: %#v", job)
	}
}

func TestNovelFetchScheduleTimeEditValidatesBeforeChangingSavedTask(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	now := time.Unix(1_800_000_000, 0)
	const originalTime = "2027-01-01T04:00:00Z"
	if err := store.PutConfig(ctx, "tester", map[string]any{"novel_fetch_schedules": []any{map[string]any{
		"id": "edit-time", "status": "scheduled", "run_at": originalTime, "input_text": "12345678901 测试书籍",
	}}}); err != nil {
		t.Fatal(err)
	}
	handler := NewRouter(RouterOptions{BridgeSecret: "schedule-input", Now: func() time.Time { return now }, NovelFetchStore: store})
	r := doNovelFetchRequest(t, handler, http.MethodPatch, "/api/novel-fetch-workshop/schedules/edit-time", map[string]any{"run_at": "broken"}, now, "schedule-input")
	if r.Code != http.StatusBadRequest || !strings.Contains(r.Body.String(), "时间") {
		t.Fatalf("invalid edit accepted: %d %s", r.Code, r.Body.String())
	}
	cfg, _ := store.GetConfig(ctx, "tester")
	item := cfg["novel_fetch_schedules"].([]any)[0].(map[string]any)
	if item["run_at"] != originalTime || item["status"] != "scheduled" {
		t.Fatalf("rejected edit changed task: %#v", item)
	}
	r = doNovelFetchRequest(t, handler, http.MethodPatch, "/api/novel-fetch-workshop/schedules/edit-time", map[string]any{"runAt": "2027-01-02T12:00"}, now, "schedule-input")
	if r.Code != http.StatusOK {
		t.Fatalf("valid edit rejected: %d %s", r.Code, r.Body.String())
	}
	cfg, _ = store.GetConfig(ctx, "tester")
	item = cfg["novel_fetch_schedules"].([]any)[0].(map[string]any)
	if item["run_at"] != "2027-01-02T04:00:00Z" {
		t.Fatalf("edited time not used: %#v", item)
	}
}
