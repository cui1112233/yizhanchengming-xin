package httpapi

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestComicInFlightPagesStayOutOfFailureLedger(t *testing.T) {
	work := []comicPageWorkItem{{page: batchfactoryv11.Video{ID: "live"}}, {page: batchfactoryv11.Video{ID: "bad"}}}
	err := runComicPageWork(context.Background(), work, 2, func(_ context.Context, item comicPageWorkItem) error {
		if item.page.ID == "live" {
			return fmt.Errorf("%w: 当前漫画页正在生成，请等待本次结果完成后再重试", batchfactoryv11.ErrConflict)
		}
		return errors.New("upstream HTTP 503")
	})
	var failures *comicPageWorkError
	if !errors.As(err, &failures) || len(failures.Failures) != 1 || failures.Failures[0].VideoID != "bad" {
		t.Fatalf("only real failures belong in ledger: %#v, err=%v", failures, err)
	}
	if err := runComicPageWork(context.Background(), work[:1], 1, func(context.Context, comicPageWorkItem) error {
		return fmt.Errorf("%w: 当前漫画页已有生成任务", batchfactoryv11.ErrConflict)
	}); err != nil {
		t.Fatalf("live task must proceed to settle/refill, got %v", err)
	}
}
