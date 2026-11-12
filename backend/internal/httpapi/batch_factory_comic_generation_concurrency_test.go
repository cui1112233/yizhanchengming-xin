package httpapi

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

// 资产缺图降级：只有漫画角色/场景（关键资产）全部缺失才算阻断条件。
func TestComicCriticalAssetCoverage(t *testing.T) {
	assets := []batchfactoryv11.BookAsset{
		{ID: "c1", Kind: "character", Source: "comic"},
		{ID: "s1", Kind: "scene", Source: "comic"},
		{ID: "p1", Kind: "prop", Source: "comic"},
		{ID: "c2", Kind: "character", Source: "novel"},
	}
	cases := []struct {
		name        string
		missing     []string
		wantTotal   int
		wantMissing int
	}{
		{name: "只缺道具不阻断", missing: []string{"p1"}, wantTotal: 2, wantMissing: 0},
		{name: "关键资产部分缺失可继续", missing: []string{"c1"}, wantTotal: 2, wantMissing: 1},
		{name: "关键资产全缺才阻断", missing: []string{"c1", "s1", "p1"}, wantTotal: 2, wantMissing: 2},
	}
	for _, tc := range cases {
		total, missing := comicCriticalAssetCoverage(assets, tc.missing)
		if total != tc.wantTotal || missing != tc.wantMissing {
			t.Fatalf("%s: total=%d missing=%d, want %d/%d", tc.name, total, missing, tc.wantTotal, tc.wantMissing)
		}
	}
}

// 资产缺图记账：缺失项可读、已计费未取回可计数，且不阻断整本。
func TestRecordComicAssetShortfallWritesLedger(t *testing.T) {
	now := time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC)
	entry := map[string]any{}
	recordComicAssetShortfall(entry, comicAssetShortfall{
		MissingAssetIDs: []string{"a2", "a3"},
		CriticalTotal:   2,
		CriticalMissing: 1,
		Billed:          true,
		Err:             &shuihuoImageBilledReadFailure{message: "200 后读取响应体失败"},
	}, now)

	ledger := automationFailureLedger(entry, "comic_asset_images")
	if len(ledger) != 2 {
		t.Fatalf("asset ledger=%+v, want 2 entries", ledger)
	}
	for _, assetID := range []string{"a2", "a3"} {
		record, _ := ledger[assetID].(map[string]any)
		if record == nil {
			t.Fatalf("asset ledger missing %s: %+v", assetID, ledger)
		}
		if record["lastErrorCode"] != automationCodeBilledNotRetrieved {
			t.Fatalf("%s code=%v, want %s", assetID, record["lastErrorCode"], automationCodeBilledNotRetrieved)
		}
		if !automationBoolValue(record["billedNotRetrieved"]) {
			t.Fatalf("%s 未标记已计费未取回: %+v", assetID, record)
		}
	}
	if got := automationNumberValue(entry["billedNotRetrievedCount"], -1); got != 2 {
		t.Fatalf("billedNotRetrievedCount=%d, want 2", got)
	}
	if got := automationNumberValue(entry["assetShortfallCount"], -1); got != 2 {
		t.Fatalf("assetShortfallCount=%d, want 2", got)
	}
}

// 默认并发从 2 提到 4：按本串行出图后一本书独占通道，4 路能把单本耗时从
// 约 28 分钟压到约 8 分钟；上限仍是 4，避免 ggwk 长响应被并发拖到断流。
func TestShuihuoImageGenerationConcurrencyDefaultsToFourLanes(t *testing.T) {
	t.Setenv(shuihuoImageConcurrencyEnv, "")
	if got := shuihuoImageGenerationConcurrency(); got != 4 {
		t.Fatalf("default image concurrency=%d, want 4", got)
	}
	if got := shuihuoClampImageConcurrency(99); got != 4 {
		t.Fatalf("clamped image concurrency=%d, want 4", got)
	}
}

// 并发提升后页级隔离仍成立：8 路并发下 1 页失败，其余页必须全部完成。
func TestRunComicPageWorkIsolatesFailureUnderHighConcurrency(t *testing.T) {
	const total = 12
	work := make([]comicPageWorkItem, 0, total)
	for index := 0; index < total; index++ {
		work = append(work, comicPageWorkItem{index: index})
	}
	var completed atomic.Int32
	err := runComicPageWork(context.Background(), work, 8, func(_ context.Context, item comicPageWorkItem) error {
		if item.index == 5 {
			return errors.New("provider failed")
		}
		completed.Add(1)
		return nil
	})
	var pageErr *comicPageWorkError
	if !errors.As(err, &pageErr) || len(pageErr.Failures) != 1 {
		t.Fatalf("result error=%v, want exactly one page failure", err)
	}
	if got := completed.Load(); got != total-1 {
		t.Fatalf("completed pages=%d, want %d", got, total-1)
	}
}

// 页级隔离：一页失败后，排队页必须继续执行，不能被 stopQueue 掐断。
func TestRunComicPageWorkContinuesQueuedPagesAfterFailure(t *testing.T) {
	work := []comicPageWorkItem{{index: 0}, {index: 1}, {index: 2}}
	startedTwo := make(chan struct{})
	releaseFirst := make(chan struct{})
	var calls atomic.Int32
	var completed atomic.Int32
	result := make(chan error, 1)
	go func() {
		result <- runComicPageWork(context.Background(), work, 2, func(ctx context.Context, _ comicPageWorkItem) error {
			call := calls.Add(1)
			if call == 1 {
				<-startedTwo
				<-releaseFirst
				if ctx.Err() != nil {
					return ctx.Err()
				}
				completed.Add(1)
				return nil
			}
			if call == 2 {
				close(startedTwo)
				return errors.New("provider failed")
			}
			return nil
		})
	}()

	select {
	case <-startedTwo:
	case <-time.After(time.Second):
		t.Fatal("two pages did not start")
	}
	deadline := time.Now().Add(time.Second)
	for calls.Load() != 3 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if got := calls.Load(); got != 3 {
		t.Fatalf("started pages=%d, want queued page to keep running after failure", got)
	}
	close(releaseFirst)
	select {
	case err := <-result:
		var pageErr *comicPageWorkError
		if !errors.As(err, &pageErr) || len(pageErr.Failures) != 1 || pageErr.Failures[0].Err.Error() != "provider failed" {
			t.Fatalf("result error=%v, want aggregated provider failure", err)
		}
	case <-time.After(time.Second):
		t.Fatal("in-flight page did not finish")
	}
	if got := completed.Load(); got != 1 {
		t.Fatalf("completed in-flight pages=%d, want 1", got)
	}
}

func TestRunComicPageWorkContinuesAfterContentPolicyRejection(t *testing.T) {
	work := []comicPageWorkItem{{index: 0}, {index: 1}}
	var calls atomic.Int32
	err := runComicPageWork(context.Background(), work, 1, func(_ context.Context, item comicPageWorkItem) error {
		calls.Add(1)
		if item.index == 0 {
			return errShuihuoImageContentPolicy
		}
		return nil
	})
	var pageErr *comicPageWorkError
	if !errors.As(err, &pageErr) || len(pageErr.Failures) != 1 {
		t.Fatalf("content policy should be recorded as one page failure, got %v", err)
	}
	if pageErr.Failures[0].Code != automationCodeContentPolicy {
		t.Fatalf("failure code=%s, want %s", pageErr.Failures[0].Code, automationCodeContentPolicy)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("generated pages=%d, want both pages attempted", got)
	}
}

// 页级聚合不连坐：已有成功页时只降级为可续跑的部分失败；一页都没成功
// 才保持整本失败语义。
func TestComicPageGenerationErrorDegradesToPartialWhenPagesSucceeded(t *testing.T) {
	failure := comicPageFailure{VideoID: "v2", Label: "漫画页 02", Code: automationCodeTransient, Err: errors.New("provider failed")}

	wholeBook := comicPageGenerationError([]comicPageFailure{failure}, 0, 2)
	var partial *comicPagePartialError
	if errors.As(wholeBook, &partial) {
		t.Fatalf("没有成功页时必须整本失败，实际得到 %v", wholeBook)
	}
	var pageErr *comicPageWorkError
	if !errors.As(wholeBook, &pageErr) || len(pageErr.Failures) != 1 {
		t.Fatalf("整本失败错误=%v, want 聚合页级失败", wholeBook)
	}

	degraded := comicPageGenerationError([]comicPageFailure{failure}, 3, 4)
	if !errors.As(degraded, &partial) {
		t.Fatalf("已有成功页时必须降级为部分失败，实际得到 %v", degraded)
	}
	if partial.Completed != 3 || partial.Total != 4 || len(partial.Failures) != 1 {
		t.Fatalf("partial=%+v, want completed=3 total=4 failures=1", partial)
	}
	if retryableComicAutomationError(degraded) {
		t.Fatal("部分失败不得在原地重复调上游，必须交给带退避的续跑")
	}
}

// 部分失败进页级账本：失败页可读、可续跑，且不升级成整本人工处理。
func TestComicPagePartialErrorRecordsPageLedger(t *testing.T) {
	now := time.Date(2026, 11, 6, 10, 0, 0, 0, time.UTC)
	err := &comicPagePartialError{
		comicPageWorkError: &comicPageWorkError{Failures: []comicPageFailure{
			{VideoID: "v2", Label: "漫画页 02", Code: automationCodeTransient, Err: errors.New("provider failed")},
		}},
		Completed: 3,
		Total:     4,
	}
	entry := map[string]any{}
	outcome := recordAutomationPageLedger(entry, err, "comic_images", now)
	if !outcome.HasFailures {
		t.Fatal("部分失败必须写入页级账本")
	}
	if outcome.NeedsManual {
		t.Fatal("首次页级失败不得直接转人工")
	}
	ledger := automationFailureLedger(entry, "comic_images")
	record, _ := ledger["v2"].(map[string]any)
	if record == nil {
		t.Fatalf("ledger missing page v2: %+v", ledger)
	}
	if attempts, _ := record["attempts"].(int); attempts != 1 {
		t.Fatalf("attempts=%v, want 1", record["attempts"])
	}
}
