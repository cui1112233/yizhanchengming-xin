package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"qiantie/backend/internal/screenplay"
	"strings"
	"testing"
	"time"
)

type assetCheckRouteGenerator struct{ calls *int }

func (g assetCheckRouteGenerator) Generate(context.Context, screenplay.TextRequest) (screenplay.TextResponse, error) {
	*g.calls++
	return screenplay.TextResponse{}, fmt.Errorf("HTTP must not call generator")
}
func assetCheckRouteFixture(t *testing.T) (*screenplay.MemoryStore, *screenplay.AuthoringService, *http.ServeMux, time.Time, *int) {
	t.Helper()
	ctx := context.Background()
	store := screenplay.NewMemoryStore()
	_, err := store.CreateProject(ctx, screenplay.Project{ID: "asset-check-p", OwnerUsername: "alice", ModelID: "own-text", SourceText: "真实原文", PermissionMode: screenplay.PermissionAuto})
	if err != nil {
		t.Fatal(err)
	}
	authors := &screenplay.AuthoringService{Store: store, Repository: screenplay.NewMemoryAuthoringRepository()}
	state, _ := authors.Load(ctx, "alice", "asset-check-p")
	state.Members = []screenplay.MemberGrant{{Username: "reader", Read: true}, {Username: "editor", Read: true, Edit: true}, {Username: "generator", Read: true, Generate: true}}
	if _, err = authors.Save(ctx, "alice", "asset-check-p", state, state.Revision); err != nil {
		t.Fatal(err)
	}
	calls := new(int)
	runtime := &screenplay.AssetCheckRuntime{Store: store, Repository: store, Authoring: authors, RedisDo: func(context.Context, ...string) (any, error) { return "OK", nil }, Generator: func(string) screenplay.TextGenerator { return assetCheckRouteGenerator{calls} }, BuildSteps: func(_ screenplay.Project, eps []screenplay.AssetCheckEpisode, _ []screenplay.Asset, model string) ([]screenplay.AssetCheckStep, error) {
		return []screenplay.AssetCheckStep{{ID: "step", ChapterIndex: eps[0].ChapterIndex, EpisodeNo: eps[0].EpisodeNo, Status: "pending", Request: screenplay.TextRequest{Model: model, SystemPrompt: "secret-system", UserPrompt: "private-source"}}}, nil
	}}
	now := time.Now()
	mux := http.NewServeMux()
	RegisterScreenplayAssetCheckRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, runtime, authors)
	return store, authors, mux, now, calls
}

const assetCheckRouteBase = "/api/comic-screenplays/projects/asset-check-p/asset-checks"

func createAssetCheckRouteJob(t *testing.T, s *screenplay.MemoryStore) screenplay.AssetCheckJob {
	t.Helper()
	j, err := s.CreateAssetCheckJob(context.Background(), screenplay.AssetCheckJob{ID: "private-job", Owner: "alice", ProjectID: "asset-check-p", Actor: "alice", RequestKey: "private-request", InputDigest: "private-digest", Model: "own-text", Status: "queued", LeaseToken: "secret-lease", Episodes: []screenplay.AssetCheckEpisode{{EpisodeNo: 1, Title: "第一集", Content: "private-source"}}, Steps: []screenplay.AssetCheckStep{{ID: "step", Status: "received", RawResponse: "private-raw", Request: screenplay.TextRequest{SystemPrompt: "secret-system"}}}})
	if err != nil {
		t.Fatal(err)
	}
	return j
}
func TestScreenplayAssetCheckRoutesStartIsDurableAndSummaryIsPrivate(t *testing.T) {
	s, _, mux, now, calls := assetCheckRouteFixture(t)
	input := map[string]any{"requestKey": "start-key", "sourceMode": "original", "episodeNumbers": []int{1}, "model": "own-text", "confirmed": true}
	response := signedJSONRequest(t, mux, now, "alice", "POST", assetCheckRouteBase, input)
	if response.Code != 202 || *calls != 0 {
		t.Fatalf("start %d calls=%d %s", response.Code, *calls, response.Body.String())
	}
	var data struct {
		Job screenplay.AssetCheckSummary `json:"job"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
	if data.Job.Status != "queued" || data.Job.Model != "own-text" || data.Job.RequestKey != "start-key" {
		t.Fatal(data)
	}
	repeat := signedJSONRequest(t, mux, now, "alice", "POST", assetCheckRouteBase, input)
	if repeat.Code != 202 {
		t.Fatal(repeat.Code, repeat.Body.String())
	}
	var again struct {
		Job screenplay.AssetCheckSummary `json:"job"`
	}
	json.Unmarshal(repeat.Body.Bytes(), &again)
	if again.Job.ID != data.Job.ID {
		t.Fatal("request replay duplicated job")
	}
	for _, path := range []string{assetCheckRouteBase, assetCheckRouteBase + "/" + data.Job.ID} {
		r := signedJSONRequest(t, mux, now, "reader", "GET", path, nil)
		if r.Code != 200 {
			t.Fatal(r.Code, r.Body.String())
		}
		for _, secret := range []string{"secret-system", "private-source", "input_digest", "lease_token", "raw_response", "system_prompt"} {
			if strings.Contains(r.Body.String(), secret) {
				t.Fatalf("private field exposed: %s", r.Body.String())
			}
		}
	}
	jobs, _ := s.ListAssetCheckJobs(context.Background(), "alice", "asset-check-p", 20)
	if len(jobs) != 1 || len(jobs[0].Steps) != 1 || jobs[0].Steps[0].Request.SystemPrompt != "secret-system" {
		t.Fatal(jobs)
	}
	denied := httptest.NewRecorder()
	mux.ServeHTTP(denied, httptest.NewRequest("GET", assetCheckRouteBase, nil))
	if denied.Code != 401 {
		t.Fatal(denied.Code)
	}
}
func TestScreenplayAssetCheckRoutesRoleSeparationAndOwnerIsolation(t *testing.T) {
	s, _, mux, now, calls := assetCheckRouteFixture(t)
	j := createAssetCheckRouteJob(t, s)
	for _, test := range []struct {
		actor, method, suffix string
		body                  any
	}{{"reader", "POST", "/" + j.ID + "/cancel", map[string]any{}}, {"editor", "POST", "/" + j.ID + "/cancel", map[string]any{}}, {"generator", "POST", "/" + j.ID + "/apply", map[string]any{"expectedRevision": j.Revision}}, {"reader", "POST", "", map[string]any{"requestKey": "x", "confirmed": true}}} {
		r := signedJSONRequest(t, mux, now, test.actor, test.method, assetCheckRouteBase+test.suffix, test.body)
		if r.Code != 403 {
			t.Fatalf("role %+v: %d %s", test, r.Code, r.Body.String())
		}
	}
	cancelled := signedJSONRequest(t, mux, now, "generator", "POST", assetCheckRouteBase+"/"+j.ID+"/cancel", map[string]any{})
	if cancelled.Code != 200 || *calls != 0 {
		t.Fatal(cancelled.Code, *calls)
	}
	s.CreateProject(context.Background(), screenplay.Project{ID: "asset-check-p", OwnerUsername: "bob"})
	other := signedJSONRequest(t, mux, now, "bob", "GET", assetCheckRouteBase+"/"+j.ID, nil)
	if other.Code != 404 {
		t.Fatal("foreign job leaked", other.Code, other.Body.String())
	}
}
func TestScreenplayAssetCheckRoutesExplicitApplyCASAndDiscard(t *testing.T) {
	s, _, mux, now, _ := assetCheckRouteFixture(t)
	j := createAssetCheckRouteJob(t, s)
	j, err := s.MutateAssetCheckJob(context.Background(), "alice", "asset-check-p", j.ID, func(j *screenplay.AssetCheckJob) error {
		j.Status = "completed"
		j.Candidates = []screenplay.AssetCheckCandidate{{ID: "candidate", Asset: screenplay.Asset{Type: screenplay.AssetCharacter, Name: "甲", Prompt: "人物词"}, Original: json.RawMessage(`{"name":"甲","full":"原始完整字段"}`)}}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	path := assetCheckRouteBase + "/" + j.ID
	for _, input := range []map[string]any{{}, {"expectedRevision": -1}, {"expectedRevision": j.Revision, "extra": "reject"}} {
		r := signedJSONRequest(t, mux, now, "editor", "POST", path+"/apply", input)
		if r.Code != 400 {
			t.Fatal(r.Code, r.Body.String())
		}
	}
	stale := signedJSONRequest(t, mux, now, "editor", "POST", path+"/apply", map[string]any{"expectedRevision": j.Revision - 1})
	if stale.Code != 409 {
		t.Fatal(stale.Code)
	}
	before, _ := s.ListAssets(context.Background(), "alice", "asset-check-p", "")
	if len(before) != 0 {
		t.Fatal("discovery wrote assets")
	}
	applied := signedJSONRequest(t, mux, now, "editor", "POST", path+"/apply", map[string]any{"expectedRevision": j.Revision})
	if applied.Code != 200 {
		t.Fatal(applied.Code, applied.Body.String())
	}
	repeated := signedJSONRequest(t, mux, now, "editor", "POST", path+"/apply", map[string]any{"expectedRevision": j.Revision})
	if repeated.Code != 200 {
		t.Fatal(repeated.Code)
	}
	assets, _ := s.ListAssets(context.Background(), "alice", "asset-check-p", "")
	if len(assets) != 1 || assets[0].Name != "甲" {
		t.Fatal(assets)
	}
	fresh := screenplay.AssetCheckJob{ID: "discard-job", Owner: "alice", ProjectID: "asset-check-p", RequestKey: "discard-request", InputDigest: "discard-digest", Status: "completed", Candidates: j.Candidates}
	fresh, err = s.CreateAssetCheckJob(context.Background(), fresh)
	if err != nil {
		t.Fatal(err)
	}
	discarded := signedJSONRequest(t, mux, now, "editor", "POST", assetCheckRouteBase+"/"+fresh.ID+"/discard", map[string]any{"expectedRevision": fresh.Revision})
	if discarded.Code != 200 {
		t.Fatal(discarded.Code, discarded.Body.String())
	}
	invalid := signedJSONRequest(t, mux, now, "editor", "POST", assetCheckRouteBase+"/"+fresh.ID+"/apply", map[string]any{"expectedRevision": fresh.Revision + 1})
	if invalid.Code != 409 {
		t.Fatal(invalid.Code)
	}
}
func TestScreenplayAssetCheckRoutesRejectUnknownRequestData(t *testing.T) {
	_, _, mux, now, calls := assetCheckRouteFixture(t)
	r := signedJSONRequest(t, mux, now, "alice", "POST", assetCheckRouteBase, map[string]any{"requestKey": "safe", "sourceMode": "original", "episodeNumbers": []int{1}, "model": "own-text", "confirmed": true, "content": "client must not replace saved source"})
	if r.Code != 400 || *calls != 0 {
		t.Fatal(r.Code, *calls)
	}
}

type assetCheckGrantRaceStore struct {
	*screenplay.MemoryStore
	beforeLock func()
}

func (s *assetCheckGrantRaceStore) AcquireLock(ctx context.Context, name, token string, ttl time.Duration) error {
	if s.beforeLock != nil {
		fn := s.beforeLock
		s.beforeLock = nil
		fn()
	}
	return s.MemoryStore.AcquireLock(ctx, name, token, ttl)
}
func TestScreenplayAssetCheckRoutesRecheckExactOwnerAfterGrantRevoked(t *testing.T) {
	memory, authors, mux, now, _ := assetCheckRouteFixture(t)
	j := createAssetCheckRouteJob(t, memory)
	j, _ = memory.MutateAssetCheckJob(context.Background(), "alice", "asset-check-p", j.ID, func(j *screenplay.AssetCheckJob) error {
		j.Status = "completed"
		j.Candidates = []screenplay.AssetCheckCandidate{{Asset: screenplay.Asset{Type: screenplay.AssetCharacter, Name: "不可写入甲"}, Original: json.RawMessage(`{"name":"不可写入甲"}`)}}
		return nil
	})
	race := &assetCheckGrantRaceStore{MemoryStore: memory}
	authors.Store = race
	race.beforeLock = func() {
		state, err := authors.Load(context.Background(), "alice", "asset-check-p")
		if err != nil {
			t.Fatal(err)
		}
		state.Members = nil
		if _, err = authors.Repository.Save(context.Background(), state, state.Revision); err != nil {
			t.Fatal(err)
		}
		memory.CreateProject(context.Background(), screenplay.Project{ID: "asset-check-p", OwnerUsername: "editor"})
	}
	result := signedJSONRequest(t, mux, now, "editor", "POST", assetCheckRouteBase+"/"+j.ID+"/apply", map[string]any{"expectedRevision": j.Revision})
	if result.Code != 403 {
		t.Fatal(result.Code, result.Body.String())
	}
	assets, _ := memory.ListAssets(context.Background(), "alice", "asset-check-p", "")
	if len(assets) != 0 {
		t.Fatal("revoked writer applied assets", assets)
	}
}
func TestScreenplayAssetCheckRoutesStrictJSONAndLimit(t *testing.T) {
	for _, body := range []string{`{} {}`, `{"expectedRevision":1} null`, `{"unexpected":true}`} {
		rec := httptest.NewRecorder()
		request := httptest.NewRequest("POST", "/", strings.NewReader(body))
		var input struct {
			Expected *int64 `json:"expectedRevision"`
		}
		if decodeAssetCheckRequest(rec, request, &input, false) || rec.Code != 400 {
			t.Fatal(body, rec.Code)
		}
	}
	_, _, mux, now, _ := assetCheckRouteFixture(t)
	for _, limit := range []string{"-1", "300", "not-a-number"} {
		result := signedJSONRequest(t, mux, now, "alice", "GET", assetCheckRouteBase+"?limit="+limit, nil)
		if result.Code != 400 {
			t.Fatal(limit, result.Code)
		}
	}
}

func TestScreenplayAssetCheckWorkflowRegistrationIsReachableWithoutStartingWorker(t *testing.T) {
	now := time.Now()
	store := screenplay.NewMemoryStore()
	store.CreateProject(context.Background(), screenplay.Project{ID: "asset-check-p", OwnerUsername: "alice"})
	mux := http.NewServeMux()
	registerScreenplayWorkflow(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, RouterOptions{Screenplay: store}, nil, nil)
	result := signedJSONRequest(t, mux, now, "alice", "GET", assetCheckRouteBase, nil)
	if result.Code != 200 || !strings.Contains(result.Body.String(), `"jobs":[]`) {
		t.Fatal(result.Code, result.Body.String())
	}
}
