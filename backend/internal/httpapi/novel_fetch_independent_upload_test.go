package httpapi

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
)

func readyIndependentUploadDocument(id string) novelfetchworkshop.Document {
	return novelfetchworkshop.Document{BookID: id, Gender: "女频", Original: "她推开门，决定把所有真相说出来。", ProcessedText: "她推开门，决定把所有真相说出来。", Meta: map[string]any{"originalStatus": "done", "style": "现代悬疑"}, Versions: map[string]any{}}
}

func TestWorkshopUploadsSuccessfulBookEvenWhenSiblingRewriteFails(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	good := readyIndependentUploadDocument("good")
	good.Versions["ai1"] = "她站在门边，终于鼓起勇气说出真相。"
	_ = store.PutDocument(ctx, "tester", good)
	_ = store.PutDocument(ctx, "tester", readyIndependentUploadDocument("bad"))
	uploaded := []string{}
	worker := &workshopProcessWorker{store: store, queue: novelfetchworkshop.NewMemoryQueue(), registry: newWorkshopProcessRegistry(),
		resolveRewriteProvider: func(context.Context, string) (novelfetchworkshop.RewriteProvider, error) {
			return rewriteProviderFunc(func(context.Context, novelfetchworkshop.RewriteRequest) (string, error) {
				return "", errors.New("upstream offline")
			}), nil
		},
		autoSubmit: func(_ context.Context, job novelfetchworkshop.ProcessJob) error {
			uploaded = append(uploaded, job.BookIDs...)
			return nil
		},
	}
	worker.process(ctx, novelfetchworkshop.ProcessJob{ID: "independent", Owner: "tester", BatchID: "batch", BookIDs: []string{"bad", "good"}, AutoSubmit: true, SelectedVersions: []string{"ai1"}, Attempts: 2})
	if len(uploaded) != 1 || uploaded[0] != "good" {
		t.Fatalf("successful book was blocked or failed book uploaded: %v", uploaded)
	}
}

func TestWorkshopSlowUploadDoesNotBlockSiblingProduction(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store := novelfetchworkshop.NewMemoryStore()
	good := readyIndependentUploadDocument("upload-first")
	good.Versions["ai1"] = "她站在门边，终于鼓起勇气说出真相。"
	_ = store.PutDocument(ctx, "tester", good)
	_ = store.PutDocument(ctx, "tester", readyIndependentUploadDocument("still-producing"))
	uploadStarted, produced, releaseUpload, finished := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
	defer close(releaseUpload)
	worker := &workshopProcessWorker{store: store, queue: novelfetchworkshop.NewMemoryQueue(), registry: newWorkshopProcessRegistry(),
		resolveRewriteProvider: func(context.Context, string) (novelfetchworkshop.RewriteProvider, error) {
			return rewriteProviderFunc(func(ctx context.Context, _ novelfetchworkshop.RewriteRequest) (string, error) {
				select {
				case <-uploadStarted:
				case <-ctx.Done():
					return "", ctx.Err()
				}
				close(produced)
				return "", errors.New("controlled production failure")
			}), nil
		},
		autoSubmit: func(ctx context.Context, _ novelfetchworkshop.ProcessJob) error {
			close(uploadStarted)
			select {
			case <-releaseUpload:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}
	go func() {
		defer close(finished)
		worker.process(ctx, novelfetchworkshop.ProcessJob{ID: "slow-upload", Owner: "tester", BatchID: "batch", BookIDs: []string{good.BookID, "still-producing"}, AutoSubmit: true, SelectedVersions: []string{"ai1"}, Attempts: 2})
	}()
	select {
	case <-uploadStarted:
	case <-time.After(3 * time.Second):
		t.Fatal("ready book waited for unfinished sibling before upload")
	}
	select {
	case <-produced:
	case <-time.After(3 * time.Second):
		t.Fatal("slow upload blocked sibling production")
	}
	cancel()
	select {
	case <-finished:
	case <-time.After(3 * time.Second):
		t.Fatal("worker did not finish after cancellation")
	}
}

func TestWorkshopUploadRetrySkipsModelResolution(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	document := readyIndependentUploadDocument("retry-upload")
	document.Versions["ai1"] = "她站在门边，终于鼓起勇气说出真相。"
	_ = store.PutDocument(ctx, "tester", document)
	uploads, models := 0, 0
	worker := &workshopProcessWorker{store: store, queue: novelfetchworkshop.NewMemoryQueue(), registry: newWorkshopProcessRegistry(),
		resolveRewriteProvider: func(context.Context, string) (novelfetchworkshop.RewriteProvider, error) {
			models++
			if uploads > 0 {
				return nil, errors.New("model disabled")
			}
			return nil, nil
		},
		autoSubmit: func(context.Context, novelfetchworkshop.ProcessJob) error {
			uploads++
			if uploads == 1 {
				return errors.New("upload temporarily offline")
			}
			return nil
		},
	}
	job := novelfetchworkshop.ProcessJob{ID: "upload-retry", Owner: "tester", BatchID: "batch", BookIDs: []string{document.BookID}, AutoSubmit: true, SelectedVersions: []string{"ai1"}}
	worker.process(ctx, job)
	resolved := models
	job.Attempts = 1
	job.SensitiveAIEnabled = true
	worker.process(ctx, job)
	if uploads != 2 || models != resolved {
		t.Fatalf("upload retry called model resolver: uploads=%d models=%d first=%d", uploads, models, resolved)
	}
	stored, _ := store.GetDocument(ctx, "tester", document.BookID)
	if stored.Versions["ai1"] != document.Versions["ai1"] {
		t.Fatal("retry changed completed AI body")
	}
	stored.Original += "新的原文"
	_ = store.PutDocument(ctx, "tester", stored)
	if err := worker.execute(ctx, job); err == nil || !strings.Contains(err.Error(), "model disabled") {
		t.Fatalf("changed original incorrectly reused upload checkpoint: %v", err)
	}
}

func TestCancelledNovelFetchScheduleCannotBeReopenedByWorker(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	_ = store.PutConfig(ctx, "tester", map[string]any{"novel_fetch_schedules": []any{map[string]any{"id": "cancelled", "status": "cancelled"}}})
	if err := updateNovelFetchScheduleRuntimeStatus(ctx, store, "tester", "cancelled", "completed", ""); err != nil {
		t.Fatal(err)
	}
	settings, _ := store.GetConfig(ctx, "tester")
	if settings["novel_fetch_schedules"].([]any)[0].(map[string]any)["status"] != "cancelled" {
		t.Fatal("worker reopened cancelled schedule")
	}
}

func TestWebSubmitRechecksAIBodyAndBlocksRefusalBeforeUpload(t *testing.T) {
	ctx := context.Background()
	store := novelfetchworkshop.NewMemoryStore()
	_ = store.PutConfig(ctx, "tester", map[string]any{"appConfig": map[string]any{"web_submit": map[string]any{"enabled": true, "min_file_size": 0, "organization": "org-1"}}})
	document := readyIndependentUploadDocument("12345678")
	document.Meta["platformId"] = "1"
	_ = store.PutDocument(ctx, "tester", document)
	_, _ = store.PutBody(ctx, "tester", novelfetchworkshop.BodyRecord{BookID: document.BookID, BodyRef: novelfetchworkshop.BodyRef{VersionID: "ai1"}, Content: "抱歉，我无法协助生成这段正文。"})
	usersDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(usersDir, "tester"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(usersDir, "tester", "upload-target.json"), []byte(`{"cookie":"session=ok"}`), 0600); err != nil {
		t.Fatal(err)
	}
	connector := &configuredSubmitConnector{}
	result, err := executeWebSubmit(ctx, "tester", usersDir, store, webSubmitRequest{TaskIDs: []string{document.BookID}, Versions: []string{"ai1"}}, nil, connector)
	if err != nil || result.Failed != 0 || connector.submitCalls != 0 {
		t.Fatalf("bad AI body reached remote: %#v err=%v submits=%d", result, err, connector.submitCalls)
	}
	if len(result.Results) != 1 || result.Results[0].(map[string]any)["status"] != "waiting_generate" {
		t.Fatalf("invalid AI body must wait for regeneration: %#v", result.Results)
	}
	stored, _ := store.GetDocument(ctx, "tester", document.BookID)
	if stored.Meta["siteSubmitStatus"] != "waiting_generate" {
		t.Fatalf("missing waiting_generate status: %#v", stored.Meta)
	}
	logged := false
	for _, raw := range stored.Logs {
		if entry, ok := raw.(map[string]any); ok && strings.Contains(firstText(entry["error"]), "正文复检未通过") {
			logged = true
		}
	}
	if !logged {
		t.Fatalf("missing actionable failure log: %#v", stored.Logs)
	}
}
