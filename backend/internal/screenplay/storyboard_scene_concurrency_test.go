package screenplay

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestStoryboardSceneConcurrencyStartsSecondBeforeFirstReturnsAndKeepsOrder(t *testing.T) {
	t.Setenv("QIANTIE_SCREENPLAY_SCENE_CONCURRENCY", "2")
	_, store, _ := newStoryboardService(t, PermissionBypass)
	second := make(chan struct{})
	firstStarted := make(chan struct{})
	var once sync.Once
	var active, peak atomic.Int32
	g := storyboardGenerationTestFunc(func(ctx context.Context, in TextRequest) (TextResponse, error) {
		n := active.Add(1)
		defer active.Add(-1)
		for old := peak.Load(); n > old; old = peak.Load() {
			if peak.CompareAndSwap(old, n) {
				break
			}
		}
		_, scene := readZIPSceneMessage(t, in.UserPrompt)
		text, _ := scene["content"].(string)
		marker := "第一场"
		if strings.Contains(text, "乙走出") {
			marker = "第二场"
			select {
			case <-firstStarted:
			case <-ctx.Done():
				return TextResponse{}, ctx.Err()
			case <-time.After(3 * time.Second):
				return TextResponse{}, fmt.Errorf("first scene not started")
			}
			once.Do(func() { close(second) })
		} else {
			close(firstStarted)
			select {
			case <-second:
			case <-ctx.Done():
				return TextResponse{}, ctx.Err()
			case <-time.After(3 * time.Second):
				return TextResponse{}, fmt.Errorf("第二场未在第一场返回前启动")
			}
			// Force the second response to be checkpointed before the first returns.
			timer := time.NewTimer(3 * time.Second)
			defer timer.Stop()
			ticker := time.NewTicker(time.Millisecond)
			defer ticker.Stop()
			for {
				state, e := store.LoadStoryboardGenerations(ctx, "alice", "p1")
				if e == nil && state.Attempts["concurrent-order"].Episodes[1].Calls["storyboard_scene_2"].Status == "received" {
					break
				}
				select {
				case <-ctx.Done():
					return TextResponse{}, ctx.Err()
				case <-timer.C:
					return TextResponse{}, fmt.Errorf("second scene not checkpointed")
				case <-ticker.C:
				}
			}
		}
		return TextResponse{Content: strings.Replace(zip164JSON, "图1至图4依次推进。", marker, 1)}, nil
	})
	s := &StoryboardService{Store: store, Generator: g}
	_, err := s.Generate(context.Background(), StoryboardRequest{IdempotencyKey: "concurrent-order", Owner: "alice", ProjectID: "p1", PromptID: "164", Content: "# 第1集\n## 1-1 屋内\n甲在屋内开窗。\n## 1-2 门外\n乙走出门外。"})
	if err != nil {
		t.Fatal(err)
	}
	if peak.Load() != 2 {
		t.Fatal("requests did not overlap", peak.Load())
	}
	b, err := store.GetStoryboard(context.Background(), "alice", "p1", 1, "prompt_164")
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Shots) != 2 || !strings.HasPrefix(b.Shots[0].VideoDesc, "第一场") || !strings.HasPrefix(b.Shots[1].VideoDesc, "第二场") || b.Shots[0].SourceSceneIndex != 0 || b.Shots[1].SourceSceneIndex != 1 {
		t.Fatal("completion order changed source order")
	}
}

func concurrentSceneInput(key string) StoryboardRequest {
	return StoryboardRequest{IdempotencyKey: key, Owner: "alice", ProjectID: "p1", PromptID: "164", Content: "# 第1集\n## 1-1 房间\n甲在屋内开窗。\n## 1-2 门外\n乙走出门外。\n## 1-3 车站\n丙上车。\n## 1-4 客厅\n丁坐下。"}
}

func TestStoryboardSceneConcurrencySharesProcessLimitAcrossServices(t *testing.T) {
	t.Setenv("QIANTIE_SCREENPLAY_SCENE_CONCURRENCY", "2")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	limit := cap(storyboardSceneSlots)
	services := limit/2 + 1
	started := make(chan struct{}, services*4)
	gate := make(chan struct{})
	var active, peak atomic.Int32
	var wg sync.WaitGroup
	errs := make(chan error, services)
	for i := 0; i < services; i++ {
		_, store, _ := newStoryboardService(t, PermissionBypass)
		g := storyboardGenerationTestFunc(func(ctx context.Context, _ TextRequest) (TextResponse, error) {
			n := active.Add(1)
			defer active.Add(-1)
			for old := peak.Load(); n > old; old = peak.Load() {
				if peak.CompareAndSwap(old, n) {
					break
				}
			}
			started <- struct{}{}
			select {
			case <-gate:
				return TextResponse{Content: zip164JSON}, nil
			case <-ctx.Done():
				return TextResponse{}, ctx.Err()
			}
		})
		s := &StoryboardService{Store: store, Generator: g}
		wg.Add(1)
		go func() { defer wg.Done(); _, e := s.Generate(ctx, concurrentSceneInput("shared-limit")); errs <- e }()
	}
	for i := 0; i < limit; i++ {
		select {
		case <-started:
		case <-ctx.Done():
			close(gate)
			wg.Wait()
			t.Fatal("shared limit did not fill", i)
		}
	}
	close(gate)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if peak.Load() != int32(limit) {
		t.Fatal("process-wide limit ignored", peak.Load(), limit)
	}
	if len(storyboardSceneSlots) != 0 {
		t.Fatal("scene capacity leaked")
	}
}

func TestStoryboardSceneConcurrencyRetryKeepsOtherSceneCheckpoints(t *testing.T) {
	t.Setenv("QIANTIE_SCREENPLAY_SCENE_CONCURRENCY", "2")
	_, store, _ := newStoryboardService(t, PermissionBypass)
	var failed atomic.Bool
	var firstCalls, secondCalls atomic.Int32
	g := storyboardGenerationTestFunc(func(_ context.Context, in TextRequest) (TextResponse, error) {
		_, scene := readZIPSceneMessage(t, in.UserPrompt)
		text, _ := scene["content"].(string)
		if strings.Contains(text, "乙走出") {
			secondCalls.Add(1)
			if !failed.Swap(true) {
				return TextResponse{Content: "not-json"}, nil
			}
		} else if strings.Contains(text, "甲在屋内") {
			firstCalls.Add(1)
		}
		return TextResponse{Content: zip164JSON}, nil
	})
	s := &StoryboardService{Store: store, Generator: g}
	req := concurrentSceneInput("parallel-failed-parent")
	if _, err := s.Generate(context.Background(), req); err == nil {
		t.Fatal("invalid scene accepted")
	}
	state, _ := store.LoadStoryboardGenerations(context.Background(), "alice", "p1")
	for _, phase := range []string{"storyboard_scene_1", "storyboard_scene_3", "storyboard_scene_4"} {
		if state.Attempts[req.IdempotencyKey].Episodes[1].Calls[phase].Status != "received" {
			t.Fatal("successful checkpoint lost", phase)
		}
	}
	req.RetryOf = req.IdempotencyKey
	req.IdempotencyKey = "parallel-failed-retry"
	req.ConfirmedRetry = true
	if _, err := s.Generate(context.Background(), req); err != nil {
		t.Fatal(err)
	}
	if firstCalls.Load() != 1 || secondCalls.Load() != 2 {
		t.Fatal("success rebilled or failed scene skipped", firstCalls.Load(), secondCalls.Load())
	}
	b, err := store.GetStoryboard(context.Background(), "alice", "p1", 1, "prompt_164")
	if err != nil || len(b.Shots) != 4 {
		t.Fatal(err)
	}
	for i, shot := range b.Shots {
		if shot.SourceSceneIndex != i || shot.Seq != i+1 {
			t.Fatal("source order changed")
		}
	}
}

func TestStoryboardSceneConcurrencyCancellationStopsQueuedRequests(t *testing.T) {
	t.Setenv("QIANTIE_SCREENPLAY_SCENE_CONCURRENCY", "2")
	_, store, _ := newStoryboardService(t, PermissionBypass)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var calls atomic.Int32
	g := storyboardGenerationTestFunc(func(ctx context.Context, _ TextRequest) (TextResponse, error) {
		if calls.Add(1) == 2 {
			cancel()
		}
		<-ctx.Done()
		return TextResponse{Content: zip164JSON}, nil
	})
	s := &StoryboardService{Store: store, Generator: g}
	_, err := s.Generate(ctx, concurrentSceneInput("parallel-cancel"))
	if !errors.Is(err, context.Canceled) || calls.Load() != 2 {
		t.Fatal("queued work started after cancel", err, calls.Load())
	}
	b, _ := store.ListStoryboards(context.Background(), "alice", "p1", "prompt_164")
	if len(b) != 0 {
		t.Fatal("late response saved as success")
	}
	if len(storyboardSceneSlots) != 0 {
		t.Fatal("cancel leaked capacity")
	}
}

func TestStoryboardSceneConcurrencyUnknownResponseStopsAndCannotRebill(t *testing.T) {
	t.Setenv("QIANTIE_SCREENPLAY_SCENE_CONCURRENCY", "2")
	_, store, _ := newStoryboardService(t, PermissionBypass)
	started := make(chan struct{})
	var once sync.Once
	var calls atomic.Int32
	g := storyboardGenerationTestFunc(func(ctx context.Context, in TextRequest) (TextResponse, error) {
		calls.Add(1)
		_, scene := readZIPSceneMessage(t, in.UserPrompt)
		text, _ := scene["content"].(string)
		if strings.Contains(text, "甲在屋内") {
			select {
			case <-started:
			case <-time.After(3 * time.Second):
				return TextResponse{}, fmt.Errorf("parallel peer missing")
			}
			return TextResponse{}, fmt.Errorf("upstream outcome unknown")
		}
		once.Do(func() { close(started) })
		<-ctx.Done()
		return TextResponse{Content: zip164JSON}, nil
	})
	s := &StoryboardService{Store: store, Generator: g}
	req := concurrentSceneInput("parallel-unknown")
	if _, err := s.Generate(context.Background(), req); err == nil {
		t.Fatal("unknown outcome succeeded")
	}
	state, _ := store.LoadStoryboardGenerations(context.Background(), "alice", "p1")
	if state.Attempts[req.IdempotencyKey].Status != "unknown" || calls.Load() != 2 {
		t.Fatal("unknown lost or queued work called", state.Attempts[req.IdempotencyKey].Status, calls.Load())
	}
	if _, err := s.Generate(context.Background(), req); err == nil || calls.Load() != 2 {
		t.Fatal("unknown response was rebilled", err, calls.Load())
	}
}

func TestStoryboardSceneConcurrencyPreservesHistoryDependency(t *testing.T) {
	t.Setenv("QIANTIE_SCREENPLAY_SCENE_CONCURRENCY", "2")
	_, store, _ := newStoryboardService(t, PermissionBypass)
	ctx := context.Background()
	p, _ := store.GetProject(ctx, "alice", "p1")
	cfg := EffectiveCreativeConfig(p)
	cfg.StoryboardUseContextHistory = true
	if _, err := store.SaveCreativeConfig(ctx, "alice", "p1", cfg, p.Revision); err != nil {
		t.Fatal(err)
	}
	calls := 0
	g := storyboardGenerationTestFunc(func(_ context.Context, in TextRequest) (TextResponse, error) {
		calls++
		if calls > 1 && !strings.Contains(in.UserPrompt, "前场成功标记") {
			return TextResponse{}, fmt.Errorf("previous generated scene missing")
		}
		return TextResponse{Content: strings.Replace(zip164JSON, "图1至图4依次推进。", "前场成功标记", 1)}, nil
	})
	s := &StoryboardService{Store: store, Generator: g}
	if _, err := s.Generate(ctx, concurrentSceneInput("parallel-history")); err != nil || calls != 4 {
		t.Fatal(err, calls)
	}
}

func TestStoryboardSceneConcurrencyNeverStartsCallAfterPersistedUnknown(t *testing.T) {
	_, store, g := newStoryboardService(t, PermissionBypass)
	ctx := context.Background()
	state, err := store.MutateStoryboardGenerations(ctx, "alice", "p1", func(v *StoryboardGenerationState) error {
		v.Attempts["unknown-boundary"] = &StoryboardGenerationAttempt{RequestKey: "unknown-boundary", Status: "running", Episodes: map[int]*StoryboardGenerationEpisode{1: {EpisodeNo: 1, Calls: map[string]StoryboardGenerationCall{"storyboard_scene_1": {Status: "unknown", Error: "outcome not confirmed"}}}}}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	session := &storyboardGenerationSession{store: store, owner: "alice", project: "p1", key: "unknown-boundary", attempt: *state.Attempts["unknown-boundary"]}
	_, err = session.call(ctx, 1, "storyboard_scene_3", TextRequest{UserPrompt: "new work must wait"}, g)
	if !errors.Is(err, ErrInvalidState) || g.callCount() != 0 {
		t.Fatal("new call submitted after unknown checkpoint", err, g.callCount())
	}
}

func TestStoryboardSceneConcurrencySharedLimitAlsoCoversSerialHistoryPath(t *testing.T) {
	for i := 0; i < cap(storyboardSceneSlots); i++ {
		storyboardSceneSlots <- struct{}{}
	}
	defer func() {
		for i := 0; i < cap(storyboardSceneSlots); i++ {
			<-storyboardSceneSlots
		}
	}()
	s, _, g := newStoryboardService(t, PermissionBypass)
	g.responses = []string{zip164JSON}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := s.Generate(ctx, concurrentSceneInput("serial-capacity-full"))
	if err == nil || g.callCount() != 0 {
		t.Fatal("serial path bypassed shared scene capacity", err, g.callCount())
	}
}
