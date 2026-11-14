package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/videomanagementconnector"
)

// Exercise the production dispatcher, worker, multipart connector and readback
// together. All upstreams and storage are isolated; no real book is submitted.
func TestScheduledBookFullChainAndReadbackRecovery(t *testing.T) {
	for _, failReadback := range []bool{false, true} {
		t.Run(fmt.Sprintf("readback_failure_%v", failReadback), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			const owner = "schedule-acceptance"
			const bookID = "12345678901"
			var fetches, rewrites, uploads, readbacks atomic.Int32
			original := strings.Repeat("她推开书店的门，发现多年未见的朋友正在整理书架。两人坐在窗边，终于说清了当年的误会。", 80)
			rewritten := strings.Repeat("她走进熟悉的书店，阳光洒在书架之间。多年未见的朋友抬起头，两人终于平静地说出了心里话。", 70)
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("bookid") != bookID || r.URL.Query().Get("platform") != "2" {
					t.Errorf("source request lost book/platform: %s", r.URL.RawQuery)
				}
				fetches.Add(1)
				_ = json.NewEncoder(w).Encode(map[string]any{"code": 200, "data": original, "bookinfo": map[string]any{"category": "女频", "genre": "现代甜文", "book_name": "定时链路验证"}})
			}))
			defer upstream.Close()
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if !strings.Contains(r.Header.Get("Cookie"), "sid=acceptance") {
					t.Error("upload session lost")
				}
				switch r.URL.Path {
				case "/tttadmin/api/zbooklist_upload.php":
					uploads.Add(1)
					if err := r.ParseMultipartForm(1 << 20); err != nil {
						t.Error(err)
						http.Error(w, "bad form", 400)
						return
					}
					if r.FormValue("bookid") != bookID {
						t.Errorf("upload book=%s", r.FormValue("bookid"))
					}
					file, _, err := r.FormFile("files[]")
					if err != nil {
						t.Error(err)
						return
					}
					defer file.Close()
					body, _ := io.ReadAll(file)
					if !strings.Contains(string(body), "书店") {
						t.Error("generated prose absent from upload")
					}
					_, _ = w.Write([]byte(`{"success":true,"id":"fixture-record"}`))
				case "/tttadmin/api/zbooklist_get.php":
					attempt := readbacks.Add(1)
					if failReadback && attempt == 1 {
						http.Error(w, "controlled readback failure", 503)
						return
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "data": []any{map[string]any{"bookid": bookID, "id": "fixture-record", "addtime": time.Now().In(time.FixedZone("CST", 8*3600)).Format(time.DateTime)}}})
				default:
					http.NotFound(w, r)
				}
			}))
			defer target.Close()
			store := novelfetchworkshop.NewMemoryStore()
			queue := novelfetchworkshop.NewMemoryQueue()
			due := time.Now().UTC().Add(time.Second)
			err := store.PutConfig(ctx, owner, map[string]any{
				"novel_fetch_automation": map[string]any{"enabled": true},
				"appConfig":              map[string]any{"web_submit": map[string]any{"enabled": true, "submit_versions": []any{"ai1"}}},
			})
			if err != nil {
				t.Fatal(err)
			}
			// Go through the real save endpoint before dispatching, not a fixture
			// schedule inserted directly into the database.
			mux := http.NewServeMux()
			registerNovelFetchLegacyWorkflowRoutes(mux, store, queue, "", nil)
			payload, _ := json.Marshal(map[string]any{"run_at": due.Format(time.RFC3339Nano), "input_text": bookID + "\t定时链路验证", "platform_id": "2", "selected_versions": []string{"ai1"}, "sensitive_ai_enabled": false, "max_txt": 4000})
			request := httptest.NewRequest(http.MethodPost, "/api/novel-fetch-workshop/schedules", strings.NewReader(string(payload)))
			request = request.WithContext(context.WithValue(ctx, bridgeIdentityKey{}, BridgeIdentity{Username: owner}))
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("save schedule: %d %s", response.Code, response.Body.String())
			}
			saved, err := store.GetDocument(ctx, owner, bookID)
			if err != nil || saved.Meta["scheduledHold"] != true {
				t.Fatalf("no held book after save: %#v %v", saved, err)
			}
			usersDir := t.TempDir()
			if err := os.MkdirAll(filepath.Join(usersDir, owner), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(usersDir, owner, "upload-target.json"), []byte(`{"cookie":"sid=acceptance"}`), 0600); err != nil {
				t.Fatal(err)
			}
			connector := &videomanagementconnector.TargetHTTPService{BaseURL: target.URL, Client: target.Client()}
			worker := &workshopProcessWorker{store: store, queue: queue, registry: newWorkshopProcessRegistry(), fetcher: &novelfetchworkshop.UpstreamFetcher{Client: upstream.Client(), Endpoint: upstream.URL},
				resolveRewriteProvider: func(context.Context, string) (novelfetchworkshop.RewriteProvider, error) {
					return rewriteProviderFunc(func(_ context.Context, request novelfetchworkshop.RewriteRequest) (string, error) {
						if strings.Contains(request.SystemPrompt, "小说分类器") {
							return `{"style":"现代甜文"}`, nil
						}
						rewrites.Add(1)
						payload, err := json.Marshal(map[string]any{"id": bookID, "changed_lines": []any{map[string]any{"line_no": 1, "text": rewritten}}})
						return string(payload), err
					}), nil
				},
				autoSubmit: func(ctx context.Context, job novelfetchworkshop.ProcessJob) error {
					result, err := executeWebSubmit(ctx, owner, usersDir, store, webSubmitRequest{Queue: queue, TaskIDs: job.BookIDs, Versions: job.SelectedVersions}, nil, connector)
					if err != nil {
						return err
					}
					if result.Failed > 0 {
						return fmt.Errorf("readback pending: %d", result.Failed)
					}
					return nil
				},
			}
			// Wall-clock execution: no job exists before due time.
			if err := dispatchDueNovelFetchSchedules(ctx, store, queue, owner, due.Add(-time.Millisecond)); err != nil {
				t.Fatal(err)
			}
			if _, err := queue.Claim(ctx, time.Millisecond); err == nil {
				t.Fatal("schedule ran before due time")
			}
			ticker := time.NewTicker(20 * time.Millisecond)
			defer ticker.Stop()
			for time.Now().Before(due) {
				select {
				case <-ticker.C:
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				}
			}
			if err := dispatchDueNovelFetchSchedules(ctx, store, queue, owner, time.Now().UTC()); err != nil {
				t.Fatal(err)
			}
			job, err := queue.Claim(ctx, time.Second)
			if err != nil {
				t.Fatal(err)
			}
			if !job.AutoSubmit {
				t.Fatal("schedule lost automatic upload default")
			}
			worker.process(ctx, job)
			doc, err := store.GetDocument(ctx, owner, bookID)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("stage=%v error=%v aiError=%v fetch=%d rewrite=%d upload=%d", doc.Meta["status"], doc.Meta["error"], doc.Meta["aiError"], fetches.Load(), rewrites.Load(), uploads.Load())
			if failReadback {
				if doc.Meta["siteSubmitStatus"] != "pending_confirmation" {
					t.Fatalf("unconfirmed upload marked successful: %v", doc.Meta["siteSubmitStatus"])
				}
				// Simulate the existing retry runner invoking the same saved job.
				rewriteCount := rewrites.Load()
				job.Attempts = 1
				worker.process(ctx, job)
				if rewrites.Load() != rewriteCount {
					t.Fatal("upload retry regenerated prose")
				}
				doc, err = store.GetDocument(ctx, owner, bookID)
				if err != nil {
					t.Fatal(err)
				}
			}
			if doc.OriginalRaw != original || strings.TrimSpace(doc.ProcessedText) == "" {
				t.Fatal("raw or processed prose not persisted")
			}
			if doc.Meta["siteSubmitStatus"] != "submitted" {
				t.Fatalf("upload not confirmed: %v", doc.Meta["siteSubmitStatus"])
			}
			if fetches.Load() != 1 || uploads.Load() != 1 || rewrites.Load() == 0 {
				t.Fatalf("unexpected calls fetch=%d rewrite=%d upload=%d", fetches.Load(), rewrites.Load(), uploads.Load())
			}
			cfg, _ := store.GetConfig(ctx, owner)
			if cfg["novel_fetch_schedules"].([]any)[0].(map[string]any)["status"] != "completed" {
				t.Fatal("schedule did not finish")
			}
			if err := dispatchDueNovelFetchSchedules(ctx, store, queue, owner, time.Now().UTC()); err != nil {
				t.Fatal(err)
			}
			// Check completed versions independently from queue retry bookkeeping.
			result, err := executeWebSubmit(ctx, owner, usersDir, store, webSubmitRequest{Queue: queue, TaskIDs: []string{bookID}, Versions: []string{"ai1"}}, nil, connector)
			if err != nil || result.Failed != 0 || uploads.Load() != 1 {
				t.Fatalf("completed version uploaded twice: %v %#v", err, result)
			}
			t.Logf("wall-clock schedule completed: fetch=%d rewrite=%d upload=%d readback=%d; no duplicate upload", fetches.Load(), rewrites.Load(), uploads.Load(), readbacks.Load())
		})
	}
}
