package screenplay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type assetCheckGeneratorFunc func(context.Context, TextRequest) (TextResponse, error)

func (f assetCheckGeneratorFunc) Generate(c context.Context, r TextRequest) (TextResponse, error) {
	return f(c, r)
}

func assetCheckFakeRedis() func(context.Context, ...string) (any, error) {
	var mu sync.Mutex
	leases := map[string]string{}
	return func(ctx context.Context, args ...string) (any, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		mu.Lock()
		defer mu.Unlock()
		switch args[0] {
		case "PING":
			return "PONG", nil
		case "SET":
			if leases[args[1]] != "" {
				return nil, nil
			}
			leases[args[1]] = args[2]
			return "OK", nil
		case "EVAL":
			if leases[args[3]] != args[4] {
				return int64(0), nil
			}
			if strings.Contains(args[1], "'DEL'") {
				delete(leases, args[3])
			}
			return int64(1), nil
		case "ZADD", "ZREM":
			return int64(1), nil
		case "ZRANGEBYSCORE":
			return []any{}, nil
		}
		return nil, fmt.Errorf("unexpected redis: %v", args)
	}
}

func assetCheckRuntimeFixture(t *testing.T, count int, generator assetCheckGeneratorFunc) (*AssetCheckRuntime, *MemoryStore, AssetCheckJob) {
	t.Helper()
	ctx := context.Background()
	store := NewMemoryStore()
	_, err := store.CreateProject(ctx, Project{ID: "p1", OwnerUsername: "alice", SourceText: "测试原文", ModelID: "our-model"})
	if err != nil {
		t.Fatal(err)
	}
	runtime := &AssetCheckRuntime{Store: store, Repository: store, RedisDo: assetCheckFakeRedis(), Generator: func(string) TextGenerator { return generator }, BuildSteps: BuildAssetCheckSteps}
	job := newCheckFixture()
	job.Existing = []Asset{}
	job.Candidates = []AssetCheckCandidate{}
	for i := 0; i < count; i++ {
		job.Episodes = append(job.Episodes, AssetCheckEpisode{ChapterIndex: i, EpisodeNo: i + 1, Title: fmt.Sprintf("第%d集", i+1), Content: fmt.Sprintf("正文%d", i+1)})
	}
	project, _ := store.GetProject(ctx, "alice", "p1")
	job.Steps, err = BuildAssetCheckSteps(project, job.Episodes, nil, "our-model")
	if err != nil {
		t.Fatal(err)
	}
	job, err = store.CreateAssetCheckJob(ctx, job)
	if err != nil {
		t.Fatal(err)
	}
	return runtime, store, job
}

func assetCheckEmptyResponse(TextRequest) TextResponse {
	return TextResponse{Content: `{"characters":[],"scenes":[],"props":[]}`}
}

func TestAssetCheckRuntimeConcurrentEpisodesAndCooperativeCancel(t *testing.T) {
	var active, peak, calls atomic.Int32
	ready := make(chan struct{}, 9)
	release := make(chan struct{})
	r, store, job := assetCheckRuntimeFixture(t, 6, func(ctx context.Context, req TextRequest) (TextResponse, error) {
		calls.Add(1)
		n := active.Add(1)
		defer active.Add(-1)
		for old := peak.Load(); n > old; old = peak.Load() {
			if peak.CompareAndSwap(old, n) {
				break
			}
		}
		ready <- struct{}{}
		select {
		case <-release:
		case <-ctx.Done():
			return TextResponse{}, ctx.Err()
		}
		return TextResponse{Content: `{"characters":[{"name":"新人物","prompt":"候选"}],"scenes":[],"props":[]}`}, nil
	})
	done := make(chan error, 1)
	go func() { done <- r.RunOne(context.Background(), job) }()
	for i := 0; i < 9; i++ {
		select {
		case <-ready:
		case <-time.After(5 * time.Second):
			close(release)
			t.Fatal("three episodes × three asset kinds not concurrent", calls.Load())
		}
	}
	if _, err := r.Cancel(context.Background(), "alice", "p1", job.ID); err != nil {
		t.Fatal(err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	current, _ := store.GetAssetCheckJob(context.Background(), "alice", "p1", job.ID)
	if calls.Load() != 9 || peak.Load() != 9 || current.Status != "cancelled" || len(current.Candidates) != 1 || current.Summary().CheckedCount != 3 {
		t.Fatalf("calls=%d peak=%d job=%+v", calls.Load(), peak.Load(), current.Summary())
	}
	assets, _ := store.ListAssets(context.Background(), "alice", "p1", "")
	if len(assets) != 0 {
		t.Fatal("discovery wrote assets", assets)
	}
	applied, err := store.ApplyAssetCheckJob(context.Background(), "alice", "p1", job.ID, current.Revision)
	if err != nil || applied.AddedCount != 1 {
		t.Fatal(applied, err)
	}
}

func TestAssetCheckRuntimeUncertainRequestNeverAutomaticallyResent(t *testing.T) {
	var calls atomic.Int32
	r, store, job := assetCheckRuntimeFixture(t, 1, func(context.Context, TextRequest) (TextResponse, error) { calls.Add(1); return TextResponse{}, nil })
	job, _ = store.MutateAssetCheckJob(context.Background(), "alice", "p1", job.ID, func(j *AssetCheckJob) error {
		j.Status = "running"
		j.Steps[0].Status = "calling"
		j.LeaseToken = "old"
		return nil
	})
	if err := r.RunOne(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	current, _ := store.GetAssetCheckJob(context.Background(), "alice", "p1", job.ID)
	if calls.Load() != 0 || current.Status != "unknown" || current.Steps[0].Status != "unknown" {
		t.Fatal(current.Status, calls.Load())
	}
}

func TestAssetCheckRuntimeReceivedResponseParsesWithoutModelCall(t *testing.T) {
	var calls atomic.Int32
	r, store, job := assetCheckRuntimeFixture(t, 1, func(context.Context, TextRequest) (TextResponse, error) { calls.Add(1); return TextResponse{}, nil })
	job, _ = store.MutateAssetCheckJob(context.Background(), "alice", "p1", job.ID, func(j *AssetCheckJob) error {
		for i := range j.Steps {
			j.Steps[i].Status = "received"
			j.Steps[i].RawResponse = `{"characters":[{"name":"人物","custom":{"a":1}}],"scenes":[],"props":[]}`
		}
		return nil
	})
	if err := r.RunOne(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	current, _ := store.GetAssetCheckJob(context.Background(), "alice", "p1", job.ID)
	if calls.Load() != 0 || current.Status != "completed" || len(current.Candidates) != 1 || !strings.Contains(string(current.Candidates[0].Original), `"custom"`) {
		t.Fatal(current.Summary(), calls.Load())
	}
}

func TestAssetCheckRuntimeLateResultCannotOverwriteNewLease(t *testing.T) {
	ready := make(chan struct{}, 3)
	release := make(chan struct{})
	r, store, job := assetCheckRuntimeFixture(t, 1, func(context.Context, TextRequest) (TextResponse, error) {
		ready <- struct{}{}
		<-release
		return TextResponse{Content: `{"characters":[{"name":"late"}],"scenes":[],"props":[]}`}, nil
	})
	done := make(chan error, 1)
	go func() { done <- r.RunOne(context.Background(), job) }()
	for i := 0; i < 3; i++ {
		select {
		case <-ready:
		case <-time.After(5 * time.Second):
			close(release)
			t.Fatal("not started")
		}
	}
	_, err := store.MutateAssetCheckJob(context.Background(), "alice", "p1", job.ID, func(j *AssetCheckJob) error { j.LeaseToken = "replacement"; j.Status = "unknown"; return nil })
	if err != nil {
		t.Fatal(err)
	}
	close(release)
	if err = <-done; !errors.Is(err, ErrLockBusy) {
		t.Fatal(err)
	}
	current, _ := store.GetAssetCheckJob(context.Background(), "alice", "p1", job.ID)
	if current.LeaseToken != "replacement" || current.Status != "unknown" || len(current.Candidates) != 0 {
		t.Fatal(current.Summary())
	}
}

func TestAssetCheckStartIdempotentSnapshotsAndValidation(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	_, _ = store.CreateProject(ctx, Project{ID: "p", OwnerUsername: "alice", SourceText: "正文", ModelID: "our-model"})
	r := &AssetCheckRuntime{Store: store, Repository: store, RedisDo: assetCheckFakeRedis(), Generator: func(string) TextGenerator {
		return assetCheckGeneratorFunc(func(context.Context, TextRequest) (TextResponse, error) {
			t.Fatal("Start must not call provider")
			return TextResponse{}, nil
		})
	}, BuildSteps: BuildAssetCheckSteps}
	in := AssetCheckStartRequest{RequestKey: "request-safe-001", SourceMode: "original", EpisodeNumbers: []int{1}, Confirmed: true}
	j, err := r.Start(ctx, "alice", "p", "alice", in)
	if err != nil {
		t.Fatal(err)
	}
	again, err := r.Start(ctx, "alice", "p", "alice", in)
	if err != nil || j.ID != again.ID {
		t.Fatal(again, err)
	}
	if len(j.Steps) != 3 || j.Steps[0].Request.Model != "our-model" {
		t.Fatal(j.Steps)
	}
	_, _ = store.CreateAsset(ctx, Asset{OwnerUsername: "alice", ProjectID: "p", Type: AssetCharacter, Name: "检查启动后新建的人物"})
	project, _ := store.GetProject(ctx, "alice", "p")
	project.SourceText = "启动后修改的正文"
	if _, err = store.UpdateProject(ctx, "alice", "p", project, project.Revision); err != nil {
		t.Fatal(err)
	}
	frozen, err := r.Start(ctx, "alice", "p", "alice", in)
	if err != nil || frozen.ID != j.ID || frozen.Episodes[0].Content != "正文" || frozen.Steps[0].RequestDigest != j.Steps[0].RequestDigest {
		t.Fatal("lost start response did not recover frozen request", frozen.Summary(), err)
	}
	in.EpisodeNumbers = []int{1, 1}
	if _, err = r.Start(ctx, "alice", "p", "alice", in); !errors.Is(err, ErrInvalidArgument) {
		t.Fatal(err)
	}
	in.EpisodeNumbers = []int{2}
	if _, err = r.Start(ctx, "alice", "p", "alice", in); !errors.Is(err, ErrRevisionConflict) {
		t.Fatal(err)
	}
	if _, err = r.Start(ctx, "mallory", "p", "mallory", in); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
}

func TestAssetCheckEndUnknownRecoversReceivedWithoutCallingAndUnlocks(t *testing.T) {
	var calls atomic.Int32
	r, store, job := assetCheckRuntimeFixture(t, 1, func(context.Context, TextRequest) (TextResponse, error) { calls.Add(1); return TextResponse{}, nil })
	job, err := store.MutateAssetCheckJob(context.Background(), "alice", "p1", job.ID, func(j *AssetCheckJob) error {
		j.Status = "unknown"
		j.Steps[0].Status = "received"
		j.Steps[0].RawResponse = `{"characters":[{"name":"已收到人物"}]}`
		j.Steps[1].Status = "unknown"
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	stopped, err := r.Cancel(context.Background(), "alice", "p1", job.ID)
	if err != nil || calls.Load() != 0 || stopped.Status != "cancelled" || len(stopped.Candidates) != 1 {
		t.Fatal(stopped.Summary(), err, calls.Load())
	}
	next := newCheckFixture()
	next.ID = "new-job"
	next.RequestKey = "new-request"
	if _, err = store.CreateAssetCheckJob(context.Background(), next); err != nil {
		t.Fatal("ended unknown still blocks new jobs", err)
	}
	if _, err = store.ApplyAssetCheckJob(context.Background(), "alice", "p1", job.ID, stopped.Revision); err != nil {
		t.Fatal(err)
	}
}

func TestAssetCheckLeaseLossAfterJournalDoesNotSubmitProvider(t *testing.T) {
	var calls atomic.Int32
	r, store, job := assetCheckRuntimeFixture(t, 1, func(context.Context, TextRequest) (TextResponse, error) { calls.Add(1); return TextResponse{}, nil })
	inner := r.RedisDo
	r.RedisDo = func(ctx context.Context, args ...string) (any, error) {
		if args[0] == "EVAL" && strings.Contains(args[1], "PEXPIRE") {
			current, _ := store.GetAssetCheckJob(ctx, "alice", "p1", job.ID)
			for _, s := range current.Steps {
				if s.Status == "calling" {
					return int64(0), nil
				}
			}
		}
		return inner(ctx, args...)
	}
	if err := r.RunOne(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	current, _ := store.GetAssetCheckJob(context.Background(), "alice", "p1", job.ID)
	if calls.Load() != 0 || current.Status != "queued" {
		t.Fatal("provider called after lease lost", calls.Load(), current.Summary())
	}
}

func TestAssetCheckInvalidJSONIsVisibleFailureAndSummaryPrivate(t *testing.T) {
	r, store, job := assetCheckRuntimeFixture(t, 1, func(context.Context, TextRequest) (TextResponse, error) {
		return TextResponse{Content: "not JSON"}, nil
	})
	if err := r.RunOne(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	j, _ := store.GetAssetCheckJob(context.Background(), "alice", "p1", job.ID)
	if j.Status != "failed" || j.Summary().Episodes[0].Status != "failed" {
		t.Fatal(j.Summary())
	}
	raw, err := json.Marshal(j.Summary())
	if err != nil {
		t.Fatal(err)
	}
	for _, private := range []string{"system_prompt", "SystemPrompt", "lease_token", "raw_response", "测试原文", "正文1"} {
		if strings.Contains(string(raw), private) {
			t.Fatal("private field leak", private)
		}
	}
}

func TestAssetCheckCandidateCompletionOrderAndExactNames(t *testing.T) {
	job := AssetCheckJob{Existing: []Asset{{Type: AssetCharacter, Name: "A"}}, EpisodeCompletionOrder: []int{1, 0}, Steps: []AssetCheckStep{
		{ChapterIndex: 0, Candidates: []AssetCheckCandidate{candidate("B", "later"), candidate("a", "lower")}},
		{ChapterIndex: 1, Candidates: []AssetCheckCandidate{candidate("B", "first"), candidate("A", "do not replace")}},
	}}
	rows := collectAssetCheckCandidates(job)
	if len(rows) != 2 || rows[0].Asset.Prompt != "first" || rows[1].Asset.Name != "a" {
		t.Fatal(rows)
	}
}
