package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
)

func TestComicSlotQueueWaitingAndAdmissionRefreshWatchdog(t *testing.T) {
	q := newAutomationComicSlotQueue(1)
	if err := q.acquire(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC)
	var clock, heartbeat atomic.Int64
	clock.Store(start.Add(automationStallTimeout + time.Minute).UnixNano())
	heartbeat.Store(start.UnixNano())
	ctx := withComicGenerationHeartbeat(context.Background(), func() { heartbeat.Store(clock.Load()) })
	waiting := make(chan bool, 1)
	admitted := make(chan error, 1)
	go func() {
		admitted <- q.acquire(ctx, func(int) {
			waiting <- !automationProgressStalled(time.Unix(0, heartbeat.Load()), time.Unix(0, clock.Load()))
		})
	}()
	select {
	case healthy := <-waiting:
		if !healthy {
			q.release()
			<-admitted
			t.Fatal("normal queue wait did not refresh stall heartbeat")
		}
	case <-time.After(time.Second):
		t.Fatal("queue waiter did not start")
	}
	clock.Add(int64(automationStallTimeout + time.Minute))
	q.release()
	if err := <-admitted; err != nil {
		t.Fatal(err)
	}
	defer q.release()
	last := time.Unix(0, heartbeat.Load())
	now := time.Unix(0, clock.Load())
	if automationProgressStalled(last, now) {
		t.Fatal("admission inherited time spent waiting in queue")
	}
	if !automationProgressStalled(last, now.Add(automationStallTimeout)) {
		t.Fatal("real stalled work must still time out")
	}
}

func TestComicImageOverloadRetriesAreBoundedAndCanRecover(t *testing.T) {
	t.Setenv("QIANTIE_IMAGE_PROVIDER_MAX_ATTEMPTS", "3")
	t.Setenv("QIANTIE_IMAGE_RETRY_BASE_DELAY_MS", "0")
	for _, tc := range []struct {
		name    string
		status  int
		recover bool
		want    int
	}{
		{"busy then succeeds", 503, true, 3},
		{"still busy stops", 503, false, 3},
		{"bad request stops", 400, false, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var attempts atomic.Int32
			provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n := attempts.Add(1)
				w.Header().Set("Content-Type", "application/json")
				if tc.recover && n == 3 {
					fmt.Fprintf(w, `{"data":[{"b64_json":"%s"}]}`, testTinyPNGBase64)
					return
				}
				w.WriteHeader(tc.status)
				fmt.Fprint(w, `{"error":{"message":"system cpu overloaded"}}`)
			}))
			defer provider.Close()
			files := localartifact.NewStore(t.TempDir(), 1<<20)
			target := shuihuoAssetImageTarget{Provider: "personal_api", ModelID: "comic-image", AspectRatio: "1:1", CatalogModel: map[string]any{"endpoint": provider.URL + "/v1/images/generations", "credential": "test-secret"}}
			_, ref, _, err := generateShuihuoAssetImageArtifactWithRetry(context.Background(), nil, files, nil, "alice", target, "测试画面", nil, nil)
			if got := int(attempts.Load()); got != tc.want {
				t.Fatalf("requests=%d, want %d", got, tc.want)
			}
			if tc.recover && (err != nil || ref == "") {
				t.Fatalf("did not recover image: %v", err)
			}
			if !tc.recover && err == nil {
				t.Fatal("failed response must remain failed")
			}
		})
	}
}

func TestImageOverloadIsTransientAndNotModelCapabilityFailure(t *testing.T) {
	for _, status := range []int{408, 429, 500, 502, 503, 504} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			message := shuihuoImageStatusRejectionMessage(status, "system cpu overloaded")
			if strings.Contains(message, "模型可能不支持") {
				t.Fatalf("misleading capability error: %s", message)
			}
			err := shuihuoImageUpstreamError{Message: message, Raw: "system cpu overloaded", StatusCode: status}
			if !retryableShuihuoAssetImageError(err) {
				t.Fatal("temporary provider rejection must allow bounded page retry")
			}
			if d := classifyAutomationFailure(err, "comic_images"); d.Class != automationFailureTransient {
				t.Fatalf("failure classification=%+v", d)
			}
		})
	}
	for _, status := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound} {
		err := shuihuoImageUpstreamError{Message: shuihuoImageStatusRejectionMessage(status, "invalid request"), StatusCode: status}
		if retryableShuihuoAssetImageError(err) {
			t.Fatalf("permanent HTTP %d must not be automatically retried", status)
		}
	}
}

func TestComicRetryAttemptDoesNotDisplayPreviousStallFailure(t *testing.T) {
	c := &BatchAutomationCoordinator{now: time.Now}
	entry := map[string]any{"error": errAutomationBookStalled.Error(), "logs": []any{"previous failure"}}
	_, err := c.generateComicWithRetry(context.Background(), entry, func() {}, func(context.Context) (batchfactoryv11.ComicCollection, error) {
		if automationStringValue(entry["error"], "") != "" {
			t.Error("active generation still displays previous failure")
		}
		if entry["logs"] == nil {
			t.Error("failure history was erased")
		}
		return batchfactoryv11.ComicCollection{}, nil
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPausedBooksResumeWithoutLosingStageOrTouchingFailures(t *testing.T) {
	paused := map[string]any{"status": "paused", "stage": "comic_images", "error": "old error", "retryAt": "old time", "nodes": map[string]any{"successfulPage": "kept"}}
	failed := map[string]any{"status": "failed", "error": "needs review"}
	done := map[string]any{"status": "completed"}
	job := map[string]any{"books": map[string]any{"paused": paused, "failed": failed, "done": done}}
	preparePausedAutomationBooksForResume(job)
	if paused["status"] != "running" || paused["stage"] != "comic_images" {
		t.Fatalf("resume lost live state/stage: %#v", paused)
	}
	if paused["error"] != nil || paused["retryAt"] != nil || paused["nodes"] == nil {
		t.Fatalf("resume must clear stale notice but keep ledger: %#v", paused)
	}
	if failed["status"] != "failed" || done["status"] != "completed" {
		t.Fatal("resume changed unrelated terminal books")
	}
}
