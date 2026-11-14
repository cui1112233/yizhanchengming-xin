package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

func createPreparedSchedule(t *testing.T, store novelfetchworkshop.LifecycleStore, input map[string]any) map[string]any {
	t.Helper()
	now := time.Now().UTC()
	if input["run_at"] == nil {
		input["run_at"] = now.Add(time.Hour).Format(time.RFC3339Nano)
	}
	if input["platform_id"] == nil {
		input["platform_id"] = "2"
	}
	handler := NewRouter(RouterOptions{BridgeSecret: "schedule-prepare", Now: func() time.Time { return now }, NovelFetchStore: store})
	r := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/schedules", input, now, "schedule-prepare")
	if r.Code != http.StatusOK {
		t.Fatalf("create: %d %s", r.Code, r.Body.String())
	}
	var response struct {
		Schedule map[string]any `json:"schedule"`
	}
	if err := json.Unmarshal(r.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	return response.Schedule
}

func TestScheduleSaveCreatesHeldBooksAndDueDispatchDoesNotSkipThem(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	due := time.Now().UTC().Add(time.Hour)
	item := createPreparedSchedule(t, store, map[string]any{"run_at": due.Format(time.RFC3339Nano), "input_text": "12345678901 预先入库", "selected_versions": []string{"ai1"}})
	doc, err := store.GetDocument(ctx, "tester", "12345678901")
	if err != nil {
		t.Fatalf("saved card has no book record: %v", err)
	}
	if item["tasks_prepared"] != true || len(scheduleStringSlice(item["book_ids"])) != 1 || doc.Meta["scheduledHold"] != true || doc.Meta["status"] != "scheduled_waiting" {
		t.Fatalf("schedule not safely prepared: item=%#v meta=%#v", item, doc.Meta)
	}
	if novelFetchAutomationEligible(doc, due.Add(time.Hour)) {
		t.Fatal("ordinary automation stole a held scheduled book")
	}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", due.Add(-time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", due.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	claimCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	job, err := queue.Claim(claimCtx, time.Second)
	if err != nil || len(job.BookIDs) != 1 || job.BookIDs[0] != doc.BookID {
		t.Fatalf("prepared book skipped at due time: %#v %v", job, err)
	}
	doc, _ = store.GetDocument(ctx, "tester", doc.BookID)
	if doc.Meta["scheduledHold"] == true {
		t.Fatal("due queue did not release schedule hold")
	}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", due.Add(2*time.Second)); err != nil {
		t.Fatal(err)
	}
	noJobCtx, stop := context.WithTimeout(ctx, 20*time.Millisecond)
	defer stop()
	if duplicate, err := queue.Claim(noJobCtx, time.Millisecond); err == nil {
		t.Fatalf("duplicate scheduled job: %#v", duplicate)
	}
}

func TestScheduleResumePreservesBodiesAndOnlySkipsConfirmedRequestedVersions(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	for _, id := range []string{"12345678901", "12345678902"} {
		doc := novelfetchworkshop.Document{BookID: id, OriginalRaw: "完整原文", Original: "处理原文", ProcessedText: "处理原文", Versions: map[string]any{"ai1": "第一版文案"}, Meta: map[string]any{"status": "completed", "originalStatus": "done", "siteSubmitStatus": "submitted", "siteSubmitDoneVersions": []any{"ai1"}}}
		if id == "12345678902" {
			doc.Meta["siteSubmitDoneVersions"] = []any{"ai1", "ai2"}
			doc.Versions["ai2"] = "第二版文案"
		}
		if err := store.PutDocument(ctx, "tester", doc); err != nil {
			t.Fatal(err)
		}
	}
	item := createPreparedSchedule(t, store, map[string]any{"input_text": "12345678901 部分完成\n12345678902 全部完成", "selected_versions": []string{"ai1", "ai2"}})
	ids := scheduleStringSlice(item["book_ids"])
	if len(ids) != 1 || ids[0] != "12345678901" {
		t.Fatalf("existing record confused with complete work: %#v", item)
	}
	doc, _ := store.GetDocument(ctx, "tester", ids[0])
	if doc.OriginalRaw != "完整原文" || doc.Versions["ai1"] != "第一版文案" || len(scheduleStringSlice(doc.Meta["siteSubmitDoneVersions"])) != 1 {
		t.Fatalf("saved results overwritten: %#v", doc)
	}
	complete, _ := store.GetDocument(ctx, "tester", "12345678902")
	if complete.Meta["scheduledHold"] == true || complete.Meta["status"] != "completed" {
		t.Fatal("complete book was modified")
	}
}

func TestScheduleDoesNotTakeActiveOrManuallyStoppedBooks(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	for i, status := range []string{"running", "stopped"} {
		id := []string{"12345678901", "12345678902"}[i]
		if err := store.PutDocument(ctx, "tester", novelfetchworkshop.Document{BookID: id, Meta: map[string]any{"status": status, "currentProcessJobId": "existing-job"}}); err != nil {
			t.Fatal(err)
		}
	}
	item := createPreparedSchedule(t, store, map[string]any{"input_text": "12345678901 活动任务\n12345678902 手动停止"})
	if len(scheduleStringSlice(item["book_ids"])) != 0 || item["status"] == "completed" || len(arrayOrEmpty(item["skipped_books"])) != 2 {
		t.Fatalf("busy/manual books treated as successful new work: %#v", item)
	}
	for _, id := range []string{"12345678901", "12345678902"} {
		doc, _ := store.GetDocument(ctx, "tester", id)
		if doc.Meta["scheduledHold"] == true {
			t.Fatalf("book stolen: %s", id)
		}
	}
}

func TestScheduleExplicitColumnsMatchSavedBookMetadata(t *testing.T) {
	store := novelfetchworkshop.NewMemoryStore()
	createPreparedSchedule(t, store, map[string]any{"input_text": "女频\t推荐理由\t实际书名\t12345678901", "input_format": "custom", "column_order": []string{"gender", "reason", "bookName", "bookId"}})
	doc, err := store.GetDocument(context.Background(), "tester", "12345678901")
	if err != nil || firstText(doc.Meta["bookName"]) != "实际书名" || doc.Gender != "女频" || firstText(doc.Meta["reason"]) != "推荐理由" {
		t.Fatalf("column settings ignored: %#v %v", doc, err)
	}
}

func TestScheduleParserPreservesEmptyCellsAndLegacySmartTitles(t *testing.T) {
	for _, tc := range []struct {
		name          string
		item          map[string]any
		title, gender string
	}{
		{"header", map[string]any{"input_text": "男女频\t书名\t书籍ID\n女频\t表头书名\t12345678901"}, "表头书名", "女频"},
		{"empty_gender", map[string]any{"input_text": "12345678901\t真实标题\t\t适合男频读者", "input_normalized": true}, "真实标题", ""},
		{"legacy_smart", map[string]any{"input_text": "12345678901 旧版智能标题", "input_format": "smart", "column_order": []string{"bookId", "bookName", "gender", "reason"}}, "旧版智能标题", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			docs, _, _, _ := parseScheduleDocuments(tc.item)
			if len(docs) != 1 || firstText(docs[0].Meta["bookName"]) != tc.title || docs[0].Gender != tc.gender {
				t.Fatalf("parser changed row meaning: %#v", docs)
			}
		})
	}
}

type scheduleFailOnceStore struct {
	novelfetchworkshop.LifecycleStore
	writes int
}

func (s *scheduleFailOnceStore) PutDocument(ctx context.Context, owner string, doc novelfetchworkshop.Document) error {
	s.writes++
	if s.writes == 2 {
		return errors.New("controlled preparation interruption")
	}
	return s.LifecycleStore.PutDocument(ctx, owner, doc)
}

func TestScheduleInterruptedPreparationRecoversWithoutStartingEarly(t *testing.T) {
	ctx := context.Background()
	store := &scheduleFailOnceStore{LifecycleStore: novelfetchworkshop.NewMemoryStore()}
	now := time.Now().UTC()
	handler := NewRouter(RouterOptions{BridgeSecret: "prepare-failure", Now: func() time.Time { return now }, NovelFetchStore: store})
	r := doNovelFetchRequest(t, handler, http.MethodPost, "/api/novel-fetch-workshop/schedules", map[string]any{"run_at": now.Add(time.Hour).Format(time.RFC3339Nano), "input_text": "12345678901 第一条\n12345678902 第二条"}, now, "prepare-failure")
	if r.Code != http.StatusInternalServerError {
		t.Fatalf("interruption not surfaced: %d %s", r.Code, r.Body.String())
	}
	cfg, _ := store.GetConfig(ctx, "tester")
	item := cfg["novel_fetch_schedules"].([]any)[0].(map[string]any)
	if item["status"] != "preparing" {
		t.Fatalf("preparation not recoverable: %#v", item)
	}
	queue := novelfetchworkshop.NewMemoryQueue()
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", now); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"12345678901", "12345678902"} {
		doc, err := store.GetDocument(ctx, "tester", id)
		if err != nil || doc.Meta["scheduledHold"] != true {
			t.Fatalf("book not recovered: %s %#v %v", id, doc, err)
		}
	}
	noJobCtx, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()
	if job, err := queue.Claim(noJobCtx, time.Millisecond); err == nil {
		t.Fatalf("preparation generated paid work early: %#v", job)
	}
}

func TestScheduleDueDoesNotRestartBookManuallyStoppedAfterSaving(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	due := time.Now().Add(time.Hour)
	item := createPreparedSchedule(t, store, map[string]any{"run_at": due.Format(time.RFC3339Nano), "input_text": "12345678901 稍后停止"})
	doc, _ := store.GetDocument(ctx, "tester", "12345678901")
	doc.Meta["status"] = "stopped"
	doc.Meta["queueState"] = "stopped"
	if err := store.PutDocument(ctx, "tester", doc); err != nil {
		t.Fatal(err)
	}
	if err := dispatchDueNovelFetchSchedules(ctx, store, queue, "tester", due.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	claimCtx, cancel := context.WithTimeout(ctx, 20*time.Millisecond)
	defer cancel()
	if job, err := queue.Claim(claimCtx, time.Millisecond); err == nil {
		t.Fatalf("manual stop overridden by schedule %v: %#v", item["id"], job)
	}
	doc, _ = store.GetDocument(ctx, "tester", doc.BookID)
	if doc.Meta["status"] != "stopped" {
		t.Fatalf("manual stop overwritten: %#v", doc.Meta)
	}
}

func TestScheduleCancelPreservesProseGeneratedBeforeDue(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	queue := novelfetchworkshop.NewMemoryQueue()
	item := createPreparedSchedule(t, store, map[string]any{"input_text": "12345678901 取消保留"})
	doc, _ := store.GetDocument(ctx, "tester", "12345678901")
	doc.OriginalRaw = "保留已生成原文"
	if err := store.PutDocument(ctx, "tester", doc); err != nil {
		t.Fatal(err)
	}
	r := scheduleControlRequest(store, queue, "tester", http.MethodPost, "/api/novel-fetch-workshop/schedules/"+firstText(item["id"])+"/cancel")
	if r.Code != http.StatusOK {
		t.Fatalf("cancel: %d %s", r.Code, r.Body.String())
	}
	doc, err := store.GetDocument(ctx, "tester", "12345678901")
	if err != nil || doc.OriginalRaw != "保留已生成原文" {
		t.Fatalf("cancel removed completed content: %#v %v", doc, err)
	}
}
