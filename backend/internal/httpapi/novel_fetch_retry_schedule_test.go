package httpapi

import (
	"errors"
	"testing"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

func TestNovelFetchSafeBookIDPrefersExplicitParameters(t *testing.T) {
	cases := map[string]string{
		"https://example.com/read/999999?bookid=123456":            "123456",
		"https://example.com/read?chapterId=888888&book_id=778899": "778899",
		"bookId = 2100242549395006289":                             "2100242549395006289",
		"u_author_123456":                                          "u_author_123456",
		"123456 书名":                                                "123456",
		"没有书号":                                                     "",
	}
	for input, want := range cases {
		if got := novelFetchSafeBookID(input); got != want {
			t.Fatalf("safeBookID(%q)=%q, want %q", input, got, want)
		}
	}
}

func TestNovelFetchRetryScheduleBacksOffAndHoldsManualStages(t *testing.T) {
	now := time.Date(2026, 11, 14, 3, 0, 0, 0, time.UTC)
	meta := map[string]any{}

	first := scheduleNovelFetchTaskRetry(meta, novelFetchRetryStageRewrite, errors.New("模型超时"), 0, now)
	if attempt := intValue(first["attempt"]); attempt != 1 {
		t.Fatalf("attempt=%v, want 1", first["attempt"])
	}
	next, ok := novelFetchParseRetryTime(firstText(first["next_retry_at"]))
	if !ok || !next.Equal(now.Add(30*time.Second)) {
		t.Fatalf("next retry=%v, want %v", first["next_retry_at"], now.Add(30*time.Second))
	}
	if novelFetchTaskRetryDue(meta, novelFetchRetryStageRewrite, now.Add(29*time.Second)) {
		t.Fatal("29 秒时不应到期")
	}
	if !novelFetchTaskRetryDue(meta, novelFetchRetryStageRewrite, now.Add(31*time.Second)) {
		t.Fatal("31 秒后应到期")
	}

	second := scheduleNovelFetchTaskRetry(meta, novelFetchRetryStageRewrite, errors.New("模型超时"), 0, now)
	if attempt := intValue(second["attempt"]); attempt != 2 {
		t.Fatalf("attempt=%v, want 2", second["attempt"])
	}
	next, _ = novelFetchParseRetryTime(firstText(second["next_retry_at"]))
	if !next.Equal(now.Add(2 * time.Minute)) {
		t.Fatalf("第二次退避=%v, want %v", next, now.Add(2*time.Minute))
	}

	manual := scheduleNovelFetchTaskRetry(meta, novelFetchRetryStageRewrite, errors.New("文本模型输出被长度截断"), 0, now)
	if manual["manual_only"] != true {
		t.Fatal("模型确定性失败必须写 manual_only")
	}
	if firstText(manual["next_retry_at"]) != "" {
		t.Fatal("manual_only 不应再安排自动重试时间")
	}
	if meta["manual_ai_retry_required"] != true || firstText(meta["manual_ai_retry_stage"]) != novelFetchRetryStageRewrite {
		t.Fatalf("人工保持字段缺失：%v", meta)
	}
	if !novelFetchRetryScheduleHasManualHold(meta) {
		t.Fatal("重试表应能识别人工保持")
	}
	if novelFetchTaskRetryDue(meta, "", now.Add(24*time.Hour)) {
		t.Fatal("人工保持阶段永远不自动到期")
	}
}

func TestNovelFetchRetryDelayHintSetsMinimumDelay(t *testing.T) {
	now := time.Date(2026, 11, 14, 3, 0, 0, 0, time.UTC)
	meta := map[string]any{}
	failure := errors.New("请求过于频繁，请 45 秒后重试")
	if got := novelFetchAIRetryDelayHintSeconds(failure); got != 45*time.Second {
		t.Fatalf("hint=%v, want 45s", got)
	}
	state := scheduleNovelFetchTaskRetry(meta, novelFetchRetryStageSensitive, failure, novelFetchAIRetryDelayHintSeconds(failure), now)
	next, ok := novelFetchParseRetryTime(firstText(state["next_retry_at"]))
	if !ok || !next.Equal(now.Add(45*time.Second)) {
		t.Fatalf("next retry=%v, want %v", state["next_retry_at"], now.Add(45*time.Second))
	}
}

func TestNovelFetchAutomationEligibleRespectsStageRetryDeadline(t *testing.T) {
	now := time.Date(2026, 11, 14, 3, 0, 0, 0, time.UTC)
	document := novelfetchworkshop.Document{BookID: "book-1", Meta: map[string]any{"status": "failed"}}
	scheduleNovelFetchTaskRetry(document.Meta, novelFetchRetryStageRewrite, errors.New("模型超时"), 0, now)

	if novelFetchAutomationEligible(document, now.Add(time.Second)) {
		t.Fatal("退避未到期前不应重新入队")
	}
	if !novelFetchAutomationEligible(document, now.Add(31*time.Second)) {
		t.Fatal("退避到期后应重新入队")
	}

	scheduleNovelFetchTaskRetry(document.Meta, novelFetchRetryStageRewrite, errors.New("文本模型输出被长度截断"), 0, now)
	if novelFetchAutomationEligible(document, now.Add(24*time.Hour)) {
		t.Fatal("人工保持阶段不应被自动扫描接管")
	}
}

func TestResetNovelFetchAutomationRetryStateClearsStageSchedule(t *testing.T) {
	now := time.Date(2026, 11, 14, 3, 0, 0, 0, time.UTC)
	document := novelfetchworkshop.Document{BookID: "book-1", Meta: map[string]any{"status": "failed", "manualOnly": true}}
	scheduleNovelFetchTaskRetry(document.Meta, novelFetchRetryStageRewrite, errors.New("文本模型输出被长度截断"), 0, now)

	if !resetNovelFetchAutomationRetryState(&document) {
		t.Fatal("有重试状态时 reset 必须报告已变更")
	}
	if len(novelFetchRetrySchedule(document.Meta)) != 0 {
		t.Fatalf("重试表未清空：%v", document.Meta["retry_schedule"])
	}
	for _, key := range []string{"retry_stage", "retry_count", "next_retry_at", "last_retry_error", "manual_ai_retry_required", "manual_ai_retry_stage", "manual_ai_retry_error"} {
		if _, ok := document.Meta[key]; ok {
			t.Fatalf("%s 未清理", key)
		}
	}
	if resetNovelFetchAutomationRetryState(&document) {
		t.Fatal("没有可清理状态时 reset 不应报告变更")
	}
}

func TestNovelFetchResolvedPlatformIDUsesCatalogName(t *testing.T) {
	if got := novelFetchResolvedPlatformID(nil, map[string]any{"platformName": "知乎付费"}); got != "15" {
		t.Fatalf("platform by name=%q, want 15", got)
	}
	if got := novelFetchResolvedPlatformID(nil, map[string]any{"platformId": "3", "platformName": "知乎付费"}); got != "3" {
		t.Fatalf("explicit platform id must win, got %q", got)
	}
	if got := novelFetchResolvedPlatformID(nil, map[string]any{"platformName": "不存在的平台"}); got != "2" {
		t.Fatalf("unknown platform=%q, want default 2", got)
	}
}

func TestParseWorkshopDocumentsKeepsURLBookID(t *testing.T) {
	input := "书籍ID\t书名\nhttps://example.com/read?chapterId=888888&bookid=123456\n123457\t书名\t男频"
	documents, parsed, emptyIDs, duplicates := parseWorkshopDocuments(input, 4000)
	if parsed != 2 || emptyIDs != 0 || duplicates != 0 {
		t.Fatalf("parsed=%d empty=%d duplicates=%d", parsed, emptyIDs, duplicates)
	}
	if len(documents) != 2 {
		t.Fatalf("documents=%d, want 2", len(documents))
	}
	if documents[0].BookID != "123456" {
		t.Fatalf("URL 书号=%q, want 123456", documents[0].BookID)
	}
	if name := firstText(documents[0].Meta["bookName"]); name != "" {
		t.Fatalf("链接查询参数不应被当成书名：%q", name)
	}
	if documents[1].BookID != "123457" || firstText(documents[1].Meta["bookName"]) != "书名" || documents[1].Gender != "男频" {
		t.Fatalf("普通行解析错误：%+v", documents[1])
	}
}
