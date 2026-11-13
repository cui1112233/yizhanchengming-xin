package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"qiantie/backend/internal/batchfactoryv11"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestBatchPriorityHTTPRouteOwnershipAndActualResponse(t *testing.T) {
	c, b, _ := priorityFixture(t)
	mux := http.NewServeMux()
	registerBatchAutomationRoutes(mux, BridgeAuth{Secret: "test-priority", Now: c.now}, c, "v12")
	for _, tc := range []struct {
		owner string
		code  int
	}{{"", 401}, {"bob", 404}, {"alice", 200}} {
		request := httptest.NewRequest("POST", "/api/batch-factory/v12/batches/"+b.ID+"/automation/priority", strings.NewReader(`{"bookId":"`+b.Books[2].ID+`"}`))
		if tc.owner != "" {
			SignBridgeRequest(request, tc.owner, false, c.now(), "test-priority")
		}
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, request)
		if response.Code != tc.code {
			t.Fatalf("%s: code=%d body=%s", tc.owner, response.Code, response.Body.String())
		}
		if tc.code == 200 && (!strings.Contains(response.Body.String(), `"priorityRequested":true`) || !strings.Contains(response.Body.String(), `"priorityBookIds"`)) {
			t.Fatal(response.Body.String())
		}
	}
}

type failingPriorityStore struct {
	*memoryAutomationStateStore
	fail bool
}

type failingPriorityClaimStore struct {
	*memoryAutomationStateStore
	bookID   string
	failures int
}

func (s *failingPriorityClaimStore) Save(ctx context.Context, state automationStateFile) error {
	job := state.Jobs["job-active"]
	if job != nil && job["state"] == "running" && s.failures > 0 {
		entries, _ := job["books"].(map[string]any)
		entry, _ := entries[s.bookID].(map[string]any)
		if entry != nil && entry["status"] == "running" {
			s.failures--
			return errors.New("temporary next-book save failure")
		}
	}
	return s.memoryAutomationStateStore.Save(ctx, state)
}

func (s *failingPriorityStore) Save(ctx context.Context, state automationStateFile) error {
	if s.fail {
		return errors.New("priority storage unavailable")
	}
	return s.memoryAutomationStateStore.Save(ctx, state)
}

func TestBatchPriorityPersistenceFailureDoesNotChangeCachedOrder(t *testing.T) {
	c, b, s := priorityFixture(t)
	storage := &failingPriorityStore{memoryAutomationStateStore: s, fail: true}
	c.SetAutomationStateStore(storage)
	if _, err := c.PrioritizeBook(context.Background(), "alice", b.ID, b.Books[2].ID); err == nil {
		t.Fatal("save failure reported as success")
	}
	c.mu.Lock()
	state, _ := c.loadLocked()
	ids := automationPriorityIDs(state.Jobs["job-active"])
	c.mu.Unlock()
	if len(ids) > 0 {
		t.Fatal("failed save left a phantom cached priority")
	}
	storage.fail = false
	if _, err := c.PrioritizeBook(context.Background(), "alice", b.ID, b.Books[2].ID); err != nil {
		t.Fatal(err)
	}
	storage.fail = true
	if _, _, err := c.claimAutomationPriorityBook(context.Background(), "alice", b.ID, "job-active", 0, b.Books[1]); err == nil {
		t.Fatal("claim ignored storage failure")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	state, _ = c.loadLocked()
	if len(automationPriorityIDs(state.Jobs["job-active"])) != 1 {
		t.Fatal("failed claim consumed durable priority")
	}
}

func TestBatchPriorityActualRunnerDoesNotInterruptAndStartsChosenNext(t *testing.T) {
	for _, scenario := range []struct {
		name                   string
		target                 int
		completed, onlyCurrent bool
		failSaves              int
	}{
		{name: "queued", target: 2}, {name: "new", target: 3}, {name: "last-book-new", target: 3, onlyCurrent: true}, {name: "resume-mixed-completed", target: 2, completed: true}, {name: "transient-claim-save", target: 2, failSaves: 1}, {name: "claim-exhaustion-and-resume", target: 2, failSaves: 3},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			target := scenario.target
			c, b, s := priorityFixture(t)
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			config := json.RawMessage(`{"comic":{"enabled":true,"body":"完整中文漫画页","assets":{"id":"batch-comic-assets","body":"提取漫画人物、场景和道具"},"imageDerivation":{"body":"整理成可直接出图的提示词"}}}`)
			_, err := c.store.SaveSettings(ctx, "alice", batchfactoryv11.ScopeRef{Kind: batchfactoryv11.ScopeBatch, BatchID: b.ID}, batchfactoryv11.SettingsUpdate{ExpectedRevision: b.Revision, Patch: batchfactoryv11.SettingsPatch{"productionMode": json.RawMessage(`"comic"`), "aiPromptConfig": config, "textModelId": json.RawMessage(`"text-1"`), "imageModelId": json.RawMessage(`"image-1"`), "assetImagesEnabled": json.RawMessage(`false`)}})
			if err != nil {
				t.Fatal(err)
			}
			b, err = c.store.GetBatch(ctx, "alice", b.ID)
			if err != nil {
				t.Fatal(err)
			}
			c.stages.Director = &batchfactoryv11.DirectorService{Store: c.store, Provider: &concurrentComicTextProvider{}}
			entered := make(chan string, 8)
			gate := make(chan struct{})
			var release sync.Once
			defer release.Do(func() { close(gate) })
			c.SetComicBookGenerator(func(ctx context.Context, owner, batchID, bookID string) (batchfactoryv11.ComicCollection, error) {
				entered <- bookID
				if bookID == b.Books[0].ID {
					select {
					case <-gate:
					case <-ctx.Done():
						return batchfactoryv11.ComicCollection{}, ctx.Err()
					}
				}
				return batchfactoryv11.ComicCollection{ID: "collection-" + bookID}, nil
			})
			job := s.state.Jobs["job-active"]
			if scenario.failSaves > 0 {
				c.SetAutomationStateStore(&failingPriorityClaimStore{memoryAutomationStateStore: s, bookID: b.Books[target].ID, failures: scenario.failSaves})
			}
			job["runMode"], job["concurrency"] = "video_no_submit", 1
			job["books"].(map[string]any)[b.Books[0].ID].(map[string]any)["status"] = "queued"
			runBatch := b
			runBatch.Books = append([]batchfactoryv11.Book(nil), b.Books[:3]...)
			if scenario.onlyCurrent {
				runBatch.Books = runBatch.Books[:1]
				delete(job["books"].(map[string]any), b.Books[1].ID)
				delete(job["books"].(map[string]any), b.Books[2].ID)
			}
			if scenario.completed {
				job["books"].(map[string]any)[b.Books[0].ID].(map[string]any)["status"] = "completed"
				if _, err := c.PrioritizeBook(ctx, "alice", b.ID, b.Books[target].ID); err != nil {
					t.Fatal(err)
				}
				c.mu.Lock()
				state, _ := c.loadLocked()
				job = cloneAutomationMap(state.Jobs["job-active"])
				c.mu.Unlock()
			}
			done := make(chan struct{})
			go func() { c.run(ctx, 0, "alice", b.ID, "job-active", cloneAutomationMap(job), runBatch); close(done) }()
			next := func() string {
				select {
				case id := <-entered:
					return id
				case <-ctx.Done():
					t.Fatal("runner did not reach expected generator stage")
					return ""
				}
			}
			if scenario.completed {
				if id := next(); id != b.Books[target].ID {
					t.Fatalf("resumed first=%s", id)
				}
				if id := next(); id != b.Books[1].ID {
					t.Fatalf("completed book blocked B: %s", id)
				}
			} else {
				if id := next(); id != b.Books[0].ID {
					t.Fatalf("first=%s", id)
				}
				if _, err := c.PrioritizeBook(ctx, "alice", b.ID, b.Books[target].ID); err != nil {
					t.Fatal(err)
				}
				select {
				case id := <-entered:
					t.Fatalf("started %s before current completed", id)
				default:
				}
				release.Do(func() { close(gate) })
				if scenario.failSaves == 3 {
					select {
					case <-done:
					case <-ctx.Done():
						t.Fatal("exhausted claim left a hanging dispatcher")
					}
					c.mu.Lock()
					attention, _ := c.loadLocked()
					attentionState := automationStringValue(attention.Jobs["job-active"]["state"], "")
					attentionError := automationStringValue(attention.Jobs["job-active"]["lastError"], "")
					c.mu.Unlock()
					if attentionState != "needs_attention" || !strings.Contains(attentionError, "领取下一本失败") {
						t.Fatalf("hidden queue failure: %s %s", attentionState, attentionError)
					}
					if _, err := c.Resume(ctx, "alice", b.ID); err != nil {
						t.Fatal(err)
					}
				}
				if id := next(); id != b.Books[target].ID {
					t.Fatalf("next=%s", id)
				}
				if !scenario.onlyCurrent {
					if id := next(); id != b.Books[1].ID {
						t.Fatalf("ordinary B lost: %s", id)
					}
					if target == 3 {
						if id := next(); id != b.Books[2].ID {
							t.Fatalf("ordinary C lost: %s", id)
						}
					}
				}
			}
			select {
			case <-done:
			case <-ctx.Done():
				t.Fatal("runner did not finish")
			}
			for {
				state, _, err := c.currentJobControl("alice", b.ID, "job-active")
				if err != nil {
					t.Fatal(err)
				}
				if state == "completed" {
					break
				}
				select {
				case <-ctx.Done():
					t.Fatal("follow-up did not finish")
				case <-time.After(10 * time.Millisecond):
				}
			}
			c.mu.Lock()
			final, _ := c.loadLocked()
			remaining := automationPriorityIDs(final.Jobs["job-active"])
			c.mu.Unlock()
			if len(remaining) > 0 {
				t.Fatalf("consumed priority restored: %v", remaining)
			}
			select {
			case id := <-entered:
				t.Fatalf("duplicate generation %s", id)
			default:
			}
		})
	}
}

func priorityFixture(t *testing.T) (*BatchAutomationCoordinator, batchfactoryv11.Batch, *memoryAutomationStateStore) {
	t.Helper()
	store := batchfactoryv11.NewMemoryStore()
	batch, err := store.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "插队验收", Books: []batchfactoryv11.CreateBookInput{{Title: "A", SourceText: "A正文"}, {Title: "B", SourceText: "B正文"}, {Title: "C", SourceText: "C正文"}, {Title: "D", SourceText: "D正文"}}})
	if err != nil {
		t.Fatal(err)
	}
	entries := map[string]any{}
	for i, b := range batch.Books[:3] {
		status := "queued"
		if i == 0 {
			status = "running"
		}
		entries[b.ID] = map[string]any{"bookId": b.ID, "title": b.Title, "status": status, "stage": "queued"}
	}
	state := &memoryAutomationStateStore{found: true, state: automationStateFile{Version: 1, Jobs: map[string]map[string]any{"job-active": {"id": "job-active", "owner": "alice", "batchId": batch.ID, "state": "running", "books": entries}}}}
	c := NewBatchAutomationCoordinator(store, nil, &batchfactoryv11.BookStageService{Store: store}, nil, nil, nil, nil, nil, func() time.Time { return time.Date(2026, 11, 13, 8, 0, 0, 0, time.UTC) })
	c.SetAutomationStateStore(state)
	return c, batch, state
}

func TestBatchPriorityPausedClaimPreservesRequest(t *testing.T) {
	c, b, _ := priorityFixture(t)
	if _, err := c.PrioritizeBook(context.Background(), "alice", b.ID, b.Books[2].ID); err != nil {
		t.Fatal(err)
	}
	c.mu.Lock()
	state, _ := c.loadLocked()
	state.Jobs["job-active"]["state"] = "paused"
	c.mu.Unlock()
	if _, _, err := c.claimAutomationPriorityBook(context.Background(), "alice", b.ID, "job-active", 0, b.Books[1]); err == nil {
		t.Fatal("paused worker consumed priority")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	state, _ = c.loadLocked()
	if len(automationPriorityIDs(state.Jobs["job-active"])) != 1 {
		t.Fatal("paused priority lost")
	}
}

func TestBatchPriorityConcurrentPauseCannotOverwriteAcceptedPriority(t *testing.T) {
	c, b, _ := priorityFixture(t)
	entered := make(chan struct{})
	release := make(chan struct{})
	controlled := make(chan error, 1)
	prioritized := make(chan error, 1)
	go func() {
		_, err := c.control("alice", b.ID, "paused", func(map[string]any) error { close(entered); <-release; return nil })
		controlled <- err
	}()
	<-entered
	go func() {
		_, err := c.PrioritizeBook(context.Background(), "alice", b.ID, b.Books[2].ID)
		prioritized <- err
	}()
	alreadyReturned := false
	select {
	case err := <-prioritized:
		if err != nil {
			t.Fatal(err)
		}
		alreadyReturned = true
	case <-time.After(100 * time.Millisecond):
	}
	close(release)
	if err := <-controlled; err != nil {
		t.Fatal(err)
	}
	if !alreadyReturned {
		if err := <-prioritized; err != nil {
			t.Fatal(err)
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	state, _ := c.loadLocked()
	if len(automationPriorityIDs(state.Jobs["job-active"])) != 1 {
		t.Fatal("pause saved its stale snapshot over an accepted priority")
	}
}

func TestBatchPriorityTerminalFallbackDoesNotStopRemainingBooks(t *testing.T) {
	c, b, _ := priorityFixture(t)
	c.mu.Lock()
	state, _ := c.loadLocked()
	state.Jobs["job-active"]["books"].(map[string]any)[b.Books[0].ID].(map[string]any)["status"] = "completed"
	c.mu.Unlock()
	_, entry, err := c.claimAutomationPriorityBook(context.Background(), "alice", b.ID, "job-active", 0, b.Books[0])
	if err != nil || entry["status"] != "completed" {
		t.Fatalf("terminal fallback must be skipped by worker: %v %v", entry, err)
	}
	chosen, _, err := c.claimAutomationPriorityBook(context.Background(), "alice", b.ID, "job-active", 0, b.Books[1])
	if err != nil || chosen.ID != b.Books[1].ID {
		t.Fatalf("remaining B lost: %v", err)
	}
}

func TestBatchPriorityNextSlotKeepsActiveAndDeduplicates(t *testing.T) {
	c, b, s := priorityFixture(t)
	ctx := context.Background()
	for range 2 {
		if _, err := c.PrioritizeBook(ctx, "alice", b.ID, b.Books[2].ID); err != nil {
			t.Fatal(err)
		}
	}
	job := s.state.Jobs["job-active"]
	if len(automationPriorityIDs(job)) != 1 {
		t.Fatal(job)
	}
	entries := job["books"].(map[string]any)
	if entries[b.Books[0].ID].(map[string]any)["status"] != "running" {
		t.Fatal("active book interrupted")
	}
	chosen, entry, err := c.claimAutomationPriorityBook(ctx, "alice", b.ID, "job-active", 0, b.Books[1])
	if err != nil || chosen.ID != b.Books[2].ID || entry["status"] != "running" {
		t.Fatalf("chosen=%s entry=%v err=%v", chosen.Title, entry, err)
	}
	if len(automationPriorityIDs(s.state.Jobs["job-active"])) != 0 {
		t.Fatal("priority not consumed")
	}
	pending := replaceAutomationQueuedBook(b.Books[1:3], 1, b.Books[1], chosen)
	if len(pending) != 2 || pending[1].ID != b.Books[1].ID {
		t.Fatalf("displaced book lost: %v", pending)
	}
}

func TestBatchPriorityRejectsActiveCompletedAndForeign(t *testing.T) {
	c, b, s := priorityFixture(t)
	ctx := context.Background()
	for _, owner := range []string{"alice", "bob"} {
		if _, err := c.PrioritizeBook(ctx, owner, b.ID, b.Books[0].ID); err == nil {
			t.Fatalf("accepted active/foreign %s", owner)
		}
	}
	s.state.Jobs["job-active"]["books"].(map[string]any)[b.Books[1].ID].(map[string]any)["status"] = "completed"
	if _, err := c.PrioritizeBook(ctx, "alice", b.ID, b.Books[1].ID); !errors.Is(err, batchfactoryv11.ErrConflict) {
		t.Fatal(err)
	}
}

func TestBatchPriorityCanInsertNewBookAndSurvivesCoordinatorReload(t *testing.T) {
	c, b, s := priorityFixture(t)
	ctx := context.Background()
	if _, err := c.PrioritizeBook(ctx, "alice", b.ID, b.Books[3].ID); err != nil {
		t.Fatal(err)
	}
	c2 := NewBatchAutomationCoordinator(c.store, nil, nil, nil, nil, nil, nil, nil, c.now)
	c2.SetAutomationStateStore(s)
	chosen, _, err := c2.claimAutomationPriorityBook(ctx, "alice", b.ID, "job-active", 0, b.Books[1])
	if err != nil || chosen.ID != b.Books[3].ID {
		t.Fatalf("lost persisted priority: %v %v", chosen, err)
	}
	if len(s.state.Jobs) != 1 {
		t.Fatal("created second run")
	}
}
