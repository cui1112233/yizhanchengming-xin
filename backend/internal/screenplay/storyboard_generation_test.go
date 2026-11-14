package screenplay

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type storyboardGenerationTestFunc func(context.Context, TextRequest) (TextResponse, error)

func (f storyboardGenerationTestFunc) Generate(ctx context.Context, in TextRequest) (TextResponse, error) {
	return f(ctx, in)
}
func generationTestRequest(key string) StoryboardRequest {
	return StoryboardRequest{IdempotencyKey: key, Owner: "alice", ProjectID: "p1", Engine: EngineXiaoluoV7, Content: "# 第1集\n完整原文"}
}

func TestStoryboardGenerationRequiresStableKeyAndReplaysPersistentSuccess(t *testing.T) {
	service, store, generator := newStoryboardService(t, PermissionBypass)
	ctx := context.Background()
	missing := generationTestRequest("")
	if _, e := service.Generate(ctx, missing); !errors.Is(e, ErrInvalidArgument) || generator.callCount() != 0 {
		t.Fatal(e, generator.callCount())
	}
	in := generationTestRequest("stable-request")
	first, e := service.Generate(ctx, in)
	if e != nil {
		t.Fatal(e)
	}
	saved, _ := store.GetStoryboard(ctx, "alice", "p1", 1, EngineXiaoluoV7)
	second, e := (&StoryboardService{Store: store, Generator: generator}).Generate(ctx, in)
	if e != nil || generator.callCount() != 1 || second.GenerationAttempt.Status != "succeeded" || second.Storyboards[0].Revision != saved.Revision || first.Storyboards[0].ID != second.Storyboards[0].ID {
		t.Fatal(second, e, generator.callCount())
	}
	state, _ := store.LoadStoryboardGenerations(ctx, "alice", "p1")
	if state.Attempts[in.IdempotencyKey].Episodes[1].Calls["storyboard"].RawResponse != storyboardJSON {
		t.Fatal("raw response not checkpointed")
	}
	other, _ := store.LoadStoryboardGenerations(ctx, "bob", "p1")
	if len(other.Attempts) != 0 {
		t.Fatal("owner leak")
	}
	in.Content = "# 第1集\n修改后的原文"
	if _, e = service.Generate(ctx, in); !errors.Is(e, ErrRevisionConflict) || generator.callCount() != 1 {
		t.Fatal(e, generator.callCount())
	}
}

func TestStoryboardGenerationPartialRetryOnlyCallsFailedEpisodeAndDoesNotOverwriteSuccess(t *testing.T) {
	service, store, generator := newStoryboardService(t, PermissionBypass)
	generator.responses = []string{storyboardJSON, "not-json", storyboardJSON}
	ctx := context.Background()
	p, _ := store.GetProject(ctx, "alice", "p1")
	cfg := EffectiveCreativeConfig(p)
	cfg.StoryboardUseContextHistory = true
	_, _ = store.SaveCreativeConfig(ctx, "alice", "p1", cfg, p.Revision)
	in := generationTestRequest("partial-parent")
	in.Content = ""
	in.Episodes = map[int]string{1: "第一集完整输入", 2: "第二集完整输入"}
	first, e := service.Generate(ctx, in)
	if !errors.Is(e, ErrPartialFailure) || first.GenerationAttempt.Status != "partial" || generator.callCount() != 2 {
		t.Fatal(first, e, generator.callCount())
	}
	one, _ := store.GetStoryboard(ctx, "alice", "p1", 1, EngineXiaoluoV7)
	if _, e = service.Generate(ctx, in); !errors.Is(e, ErrPartialFailure) || generator.callCount() != 2 {
		t.Fatal(e, generator.callCount())
	}
	unlinked := in
	unlinked.IdempotencyKey = "unlinked-new-key"
	if _, e = service.Generate(ctx, unlinked); !errors.Is(e, ErrInvalidState) || generator.callCount() != 2 {
		t.Fatal(e, generator.callCount())
	}
	retry := in
	retry.IdempotencyKey = "retry-child"
	retry.RetryOf = in.IdempotencyKey
	retry.ConfirmedRetry = true
	result, e := service.Generate(ctx, retry)
	if e != nil || len(result.Storyboards) != 2 || generator.callCount() != 3 || result.GenerationAttempt.Status != "succeeded" {
		t.Fatal(result, e, generator.callCount())
	}
	after, _ := store.GetStoryboard(ctx, "alice", "p1", 1, EngineXiaoluoV7)
	if after.Revision != one.Revision || after.ID != one.ID {
		t.Fatal("inherited successful episode overwritten", one, after)
	}
	if !strings.Contains(generator.requests[2].UserPrompt, "第二集完整输入") {
		t.Fatal(generator.requests[2])
	}
}

type generationAssetWriteFailStore struct {
	*MemoryStore
	writes int
	fail   bool
	cancel context.CancelFunc
}

func (s *generationAssetWriteFailStore) CreateAsset(ctx context.Context, asset Asset) (Asset, error) {
	s.writes++
	if s.fail && s.writes == 2 {
		if s.cancel != nil {
			s.cancel()
			return Asset{}, ctx.Err()
		}
		return Asset{}, fmt.Errorf("local asset database unavailable")
	}
	return s.MemoryStore.CreateAsset(ctx, asset)
}

func TestStoryboardGenerationReceivedAssetSyncIsNotBilledAgainAfterLocalFailure(t *testing.T) {
	for _, mode := range []string{"failure", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			service, base, generator := newStoryboardService(t, PermissionBypass)
			ctx := context.Background()
			p, _ := base.GetProject(ctx, "alice", "p1")
			cfg := EffectiveCreativeConfig(p)
			cfg.StoryboardAutoSyncAssets = true
			_, _ = base.SaveCreativeConfig(ctx, "alice", "p1", cfg, p.Revision)
			store := &generationAssetWriteFailStore{MemoryStore: base, fail: true}
			service.Store = store
			runCtx := ctx
			if mode == "cancel" {
				var cancel context.CancelFunc
				runCtx, cancel = context.WithCancel(ctx)
				store.cancel = cancel
				defer cancel()
			}
			generator.responses = []string{storyboardJSON, `{"characters":[{"name":"顾苒苒","description":"角色甲","prompt":"角色甲形象"},{"name":"吴姐","description":"角色乙","prompt":"角色乙形象"}],"scenes":[{"name":"蓝雨传媒28楼·经纪人办公室","description":"办公室","prompt":"办公室图"},{"name":"顾苒苒公寓","description":"公寓","prompt":"公寓图"}],"props":[{"name":"解约协议","description":"协议纸张","prompt":"白纸合同"}]}`}
			in := generationTestRequest("asset-sync-parent")
			result, e := service.Generate(runCtx, in)
			if e == nil || generator.callCount() != 2 || !result.GenerationAttempt.CanRetry {
				t.Fatal(result, e, generator.callCount())
			}
			state, _ := base.LoadStoryboardGenerations(ctx, "alice", "p1")
			if len(state.Attempts[in.IdempotencyKey].Episodes[1].PendingAssets) != 5 {
				t.Fatal("validated asset plan not durable before writes")
			}
			store.fail = false
			in.RetryOf = in.IdempotencyKey
			in.IdempotencyKey = "asset-sync-resume"
			in.ConfirmedRetry = true
			result, e = service.Generate(ctx, in)
			if e != nil || generator.callCount() != 2 || result.GenerationAttempt.Status != "succeeded" {
				t.Fatal("received asset sync was billed twice", result, e, generator.callCount())
			}
			assets, _ := base.ListAssets(ctx, "alice", "p1", "")
			if len(assets) != 5 {
				t.Fatal(len(assets))
			}
		})
	}
}

func TestStoryboardGenerationSavedConfigPrefixAndPlanningChangesRejectOldKey(t *testing.T) {
	for _, kind := range []string{"config", "prefix", "planning", "original"} {
		t.Run(kind, func(t *testing.T) {
			service, store, generator := newStoryboardService(t, PermissionBypass)
			ctx := context.Background()
			p, _ := store.GetProject(ctx, "alice", "p1")
			cfg := EffectiveCreativeConfig(p)
			cfg.StoryboardPromptID = "storyboard_xiaoluo_v7_15s"
			p, _ = store.SaveCreativeConfig(ctx, "alice", "p1", cfg, p.Revision)
			selected, configDigest, _, _ := storyboardPlanningContext(p)
			authors := &AuthoringService{Store: store, Repository: NewMemoryAuthoringRepository()}
			_, _ = authors.Save(ctx, "alice", "p1", AuthoringState{Chapters: []Chapter{{ID: "c", Title: "章节", SourceText: "完整原文", ScriptText: "完整剧本", PlanningText: "已有规划", PlanningPromptID: selected.ID, PlanningPromptDigest: contentDigest(selected.Body), PlanningConfigDigest: configDigest}}}, 0)
			service.Authoring = authors
			_, _ = store.MutateStoryboardPreferences(ctx, "alice", func(v *StoryboardPreferences) error {
				v.PrefixPresets = []StoryboardPrefixPreset{{ID: "p", Name: "前缀", Content: "原前缀"}}
				v.SelectedPrefixID = "p"
				return nil
			})
			yes := true
			in := generationTestRequest("frozen-input")
			in.ChapterID = "c"
			in.UsePlanning = &yes
			in.UseOriginalContentDirectly = true
			if _, e := service.Generate(ctx, in); e != nil {
				t.Fatal(e)
			}
			switch kind {
			case "config":
				p, _ := store.GetProject(ctx, "alice", "p1")
				cfg := EffectiveCreativeConfig(p)
				cfg.StoryboardUseContextHistory = true
				_, _ = store.SaveCreativeConfig(ctx, "alice", "p1", cfg, p.Revision)
			case "prefix":
				_, _ = store.MutateStoryboardPreferences(ctx, "alice", func(v *StoryboardPreferences) error { v.PrefixPresets[0].Content = "新前缀"; return nil })
			case "planning", "original":
				state, _ := authors.Load(ctx, "alice", "p1")
				if kind == "planning" {
					state.Chapters[0].PlanningText = "新规划"
				} else {
					state.Chapters[0].SourceText = "新原文"
				}
				_, _ = authors.Save(ctx, "alice", "p1", state, state.Revision)
			}
			if _, e := service.Generate(ctx, in); !errors.Is(e, ErrRevisionConflict) || generator.callCount() != 1 {
				t.Fatal(kind, e, generator.callCount())
			}
		})
	}
}

func TestStoryboardGenerationRunningNewKeyAndUnknownTransportCannotRebill(t *testing.T) {
	service, _, _ := newStoryboardService(t, PermissionBypass)
	ctx := context.Background()
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	var calls atomic.Int32
	service.Generator = storyboardGenerationTestFunc(func(context.Context, TextRequest) (TextResponse, error) {
		if calls.Add(1) == 1 {
			started <- struct{}{}
			<-release
		}
		return TextResponse{Content: storyboardJSON}, nil
	})
	done := make(chan error, 1)
	in := generationTestRequest("running-first")
	go func() { _, e := service.Generate(ctx, in); done <- e }()
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		close(release)
		t.Fatal("did not start")
	}
	other := in
	other.IdempotencyKey = "different-key"
	result, e := service.Generate(ctx, other)
	count := calls.Load()
	close(release)
	firstErr := <-done
	if !errors.Is(e, ErrInvalidState) || count != 1 || result.GenerationAttempt.RequestKey != in.IdempotencyKey || firstErr != nil {
		t.Fatal(result, e, count, firstErr)
	}
	uncertain, _, generator := newStoryboardService(t, PermissionBypass)
	generator.err = fmt.Errorf("socket disappeared after submitting")
	unknown := generationTestRequest("unknown-request")
	out, e := uncertain.Generate(ctx, unknown)
	if e == nil || out.GenerationAttempt.Status != "unknown" || out.GenerationAttempt.CanRetry {
		t.Fatal(out, e)
	}
	unknown.IdempotencyKey = "retry-unknown"
	unknown.RetryOf = "unknown-request"
	unknown.ConfirmedRetry = true
	if _, e = uncertain.Generate(ctx, unknown); !errors.Is(e, ErrInvalidState) || generator.callCount() != 1 {
		t.Fatal(e, generator.callCount())
	}
}

func TestStoryboardGenerationCancelledReceivedOutputCanResumeWithoutSecondModelCall(t *testing.T) {
	service, store, _ := newStoryboardService(t, PermissionBypass)
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	service.Generator = storyboardGenerationTestFunc(func(context.Context, TextRequest) (TextResponse, error) {
		calls++
		cancel()
		return TextResponse{Content: storyboardJSON}, nil
	})
	in := generationTestRequest("cancel-received")
	out, e := service.Generate(ctx, in)
	if e == nil || out.GenerationAttempt.Status != "cancelled" || !out.GenerationAttempt.CanRetry {
		t.Fatal(out, e)
	}
	if _, e = store.GetStoryboard(context.Background(), "alice", "p1", 1, EngineXiaoluoV7); !errors.Is(e, ErrNotFound) {
		t.Fatal("cancelled response saved as successful board", e)
	}
	service.Generator = storyboardGenerationTestFunc(func(context.Context, TextRequest) (TextResponse, error) {
		t.Fatal("received valid output billed again")
		return TextResponse{}, nil
	})
	retry := in
	retry.IdempotencyKey = "cancel-resume"
	retry.RetryOf = in.IdempotencyKey
	retry.ConfirmedRetry = true
	out, e = service.Generate(context.Background(), retry)
	if e != nil || out.GenerationAttempt.Status != "succeeded" || calls != 1 {
		t.Fatal(out, e, calls)
	}
}

type generationCommitFailStore struct {
	*MemoryStore
	fail bool
}

func (s *generationCommitFailStore) CommitStoryboardGeneration(ctx context.Context, owner, project, key string, no int, board Storyboard, warnings []string) (StoryboardGenerationAttempt, error) {
	if s.fail {
		return StoryboardGenerationAttempt{}, fmt.Errorf("simulated SQL commit failed")
	}
	return s.MemoryStore.CommitStoryboardGeneration(ctx, owner, project, key, no, board, warnings)
}
func TestStoryboardGenerationSaveFailureRetainsReceiptAndRetryDoesNotCallModel(t *testing.T) {
	service, base, generator := newStoryboardService(t, PermissionBypass)
	store := &generationCommitFailStore{MemoryStore: base, fail: true}
	service.Store = store
	in := generationTestRequest("save-failure")
	out, e := service.Generate(context.Background(), in)
	if e == nil || out.GenerationAttempt.Status != "failed" || generator.callCount() != 1 {
		t.Fatal(out, e, generator.callCount())
	}
	store.fail = false
	in.RetryOf = in.IdempotencyKey
	in.IdempotencyKey = "save-retry"
	in.ConfirmedRetry = true
	out, e = service.Generate(context.Background(), in)
	if e != nil || out.GenerationAttempt.Status != "succeeded" || generator.callCount() != 1 {
		t.Fatal(out, e, generator.callCount())
	}
}
