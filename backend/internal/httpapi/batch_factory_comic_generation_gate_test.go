package httpapi

import (
	"context"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestComicPageGenerationContextSurvivesRequestCancellation(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	ctx, stop := comicPageGenerationContext(parent)
	defer stop()
	cancel()
	select {
	case <-ctx.Done():
		t.Fatal("generation context was cancelled with the browser request")
	default:
	}
	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) <= 0 {
		t.Fatal("generation context should have a bounded deadline")
	}
}

func TestComicPageGenerationContextDoesNotInheritShortParentDeadline(t *testing.T) {
	parent, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	ctx, stop := comicPageGenerationContext(parent)
	defer stop()

	time.Sleep(30 * time.Millisecond)
	if err := ctx.Err(); err != nil {
		t.Fatalf("page generation inherited the short parent deadline: %v", err)
	}
	deadline, ok := ctx.Deadline()
	if !ok || time.Until(deadline) < 5*time.Hour {
		t.Fatalf("page generation deadline=%v ok=%t, want its own long bounded lifetime", deadline, ok)
	}
}

func TestComicPageGenerationGateRejectsDuplicatePage(t *testing.T) {
	gate := newComicPageGenerationGate()
	release, acquired := gate.tryAcquire("alice/batch/book/video-1")
	if !acquired || release == nil {
		t.Fatal("first request should acquire the page generation gate")
	}
	defer release()

	duplicateRelease, duplicateAcquired := gate.tryAcquire("alice/batch/book/video-1")
	if duplicateAcquired || duplicateRelease != nil {
		t.Fatal("duplicate request should be rejected while the page is generating")
	}

	if _, acquired := gate.tryAcquire("alice/batch/book/video-2"); !acquired {
		t.Fatal("a different page should still be able to acquire the gate")
	}
}

func TestComicPageGenerationGateReleasesAfterRequestFinishes(t *testing.T) {
	gate := newComicPageGenerationGate()
	release, acquired := gate.tryAcquire("alice/batch/book/video-1")
	if !acquired {
		t.Fatal("first request should acquire the page generation gate")
	}
	release()

	secondRelease, acquired := gate.tryAcquire("alice/batch/book/video-1")
	if !acquired || secondRelease == nil {
		t.Fatal("page should be available after the first request releases it")
	}
	secondRelease()
}

func TestComicPageGenerationUsesRelayWaitBudget(t *testing.T) {
	if shuihuoRelayImageRequestTimeout < 10*time.Minute {
		t.Fatalf("relay request timeout=%s, want at least 10 minutes", shuihuoRelayImageRequestTimeout)
	}
}

func TestComicPageGenerationTaskStateBlocksOnlyUnfinishedTask(t *testing.T) {
	if !comicPageGenerationTaskExists(batchfactoryv11.ComicPage{Status: "generating"}) {
		t.Fatal("a page with a generating status must be treated as an existing task")
	}
	if comicPageGenerationTaskExists(batchfactoryv11.ComicPage{
		Status: "image_ready",
		Images: []batchfactoryv11.ComicPageImage{{IsPrimary: true}},
	}) {
		t.Fatal("a completed page must allow an explicit regeneration")
	}
}

func TestComicPageGenerationTaskOnlyBecomesStaleAfterSafetyWindow(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	page := batchfactoryv11.ComicPage{Status: "generating", UpdatedAt: now.Add(-comicPageGenerationStaleAfter + time.Second)}
	if comicPageGenerationTaskIsStale(page, now) {
		t.Fatal("a recently updated generation task must remain protected")
	}
	page.UpdatedAt = now.Add(-comicPageGenerationStaleAfter - time.Second)
	if !comicPageGenerationTaskIsStale(page, now) {
		t.Fatal("an old generation lease must be recoverable")
	}
}

func TestRecentDurableComicGenerationLeaseSurvivesProcessRestart(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	page := batchfactoryv11.ComicPage{Status: "generating", UpdatedAt: now.Add(-time.Minute)}
	if comicPageGenerationLeaseRecoverable(page, true, now) {
		t.Fatal("a new process-local gate must not reclaim a recent remote generation")
	}
	page.UpdatedAt = now.Add(-comicPageGenerationStaleAfter - time.Second)
	if !comicPageGenerationLeaseRecoverable(page, true, now) {
		t.Fatal("a durable generation lease may be recovered after the safety window")
	}
}
