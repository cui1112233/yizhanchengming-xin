package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/screenplay"
	"strings"
	"testing"
	"time"
)

type subjectPreflightService struct {
	fakeScreenplayVideoService
	checked int
}

func (s *subjectPreflightService) ValidateInput(_ context.Context, _ string, in scriptVideoLocalInput) error {
	s.checked++
	if in.Prompt == "bad" {
		return fmt.Errorf("bad provider input")
	}
	return nil
}
func TestScreenplaySubjectPreflightIsWholeEpisode(t *testing.T) {
	c := NewUserConfigStore(t.TempDir())
	_ = c.write("alice", map[string]any{"scriptDefaults": map[string]any{"videoModelKey": "own"}})
	s := &subjectPreflightService{}
	repo := screenplay.NewMemoryProductionRepository()
	r := &screenplayVideoRunner{service: s, configs: c, repo: repo}
	_, err := r.ProduceEpisode(context.Background(), screenplay.Project{ID: "p", OwnerUsername: "alice"}, 1, screenplay.Storyboard{ID: "b", Shots: []screenplay.StoryboardShot{{Seq: 1, Duration: 5, VideoDesc: "good"}, {Seq: 2, Duration: 5, VideoDesc: "bad"}}})
	tasks, _ := repo.ListShotTasks(context.Background(), "alice", "p")
	if err == nil || len(s.inputs) != 0 || len(tasks) != 0 || s.checked != 2 {
		t.Fatalf("err=%v submits=%d tasks=%d checked=%d", err, len(s.inputs), len(tasks), s.checked)
	}
}
func TestScreenplaySubjectResolverStrategiesAndBoundaries(t *testing.T) {
	ctx := context.Background()
	store := screenplay.NewMemoryStore()
	p, _ := store.CreateProject(ctx, screenplay.Project{ID: "p", OwnerUsername: "alice"})
	for _, a := range []screenplay.Asset{{ID: "c", Name: "Alice", Type: screenplay.AssetCharacter, Extra: map[string]string{"sora2_info": "asset-alice"}}, {ID: "s", Name: "Park", Type: screenplay.AssetScene, Extra: map[string]string{"sora2_info": "asset-park"}}, {ID: "x", Name: "Book", Type: screenplay.AssetProp, Extra: map[string]string{"sora2_info": "asset-book"}}} {
		a.OwnerUsername = "alice"
		a.ProjectID = "p"
		if _, err := store.CreateAsset(ctx, a); err != nil {
			t.Fatal(err)
		}
	}
	media := &ScreenplayMediaWorkbench{Projects: store}
	shot := screenplay.StoryboardShot{Characters: "Alice", Scene: "Park", Props: "Book", VideoDesc: "@Alice , @Park @Book @Unknown @Alice。"}
	p.CreativeConfig = &screenplay.CreativeConfig{AutoReplaceSora2Info: true, VideoGenerationMode: "text_to_video"}
	got, err := media.VideoSubjects(ctx, p, shot)
	if err != nil || got.Prompt != "@asset-alice , @asset-park @Book @Unknown @Alice。" {
		t.Fatalf("Sora: %+v %v", got, err)
	}
	p.CreativeConfig = &screenplay.CreativeConfig{UseSubjectMode: true}
	got, err = media.VideoSubjects(ctx, p, shot)
	if err != nil || got.Prompt != shot.VideoDesc || len(got.URIs) != 3 || got.URIs[0] != "asset://asset-alice" {
		t.Fatalf("Seedance: %+v %v", got, err)
	}
	shot.Characters = "Missing"
	if _, err = media.VideoSubjects(ctx, p, shot); err == nil {
		t.Fatal("missing subject accepted")
	}
	shot.Characters = "Alice"
	shot.ProtocolFields = []byte(`{"character_ids":["unknown"]}`)
	if _, err = media.VideoSubjects(ctx, p, shot); err == nil {
		t.Fatal("unknown explicit id fell back to name")
	}
	for _, raw := range []string{`{"asset_id":"x"}`, `{"volcengineAssetId":"x"}`, "asset://", "bad id", "asset-abc.def", "abc", "asset-own:part", `{bad asset-own`, `["asset-own"]`} {
		if screenplaySubjectURI(raw) != "" {
			t.Fatalf("opaque archive became URI: %s", raw)
		}
	}
	if screenplaySubjectArchive(screenplay.Asset{Extra: map[string]string{"sora2_info": "", "subject_upload_result": `{"assetId":"old"}`}}) != "" {
		t.Fatal("clear revived snapshot")
	}
}
func TestScreenplaySubjectGenerationModeAndHash(t *testing.T) {
	ctx := context.Background()
	p := screenplay.Project{ID: "p", OwnerUsername: "alice", CreativeConfig: &screenplay.CreativeConfig{VideoGenerationMode: "text_to_video"}}
	r := &screenplayVideoRunner{References: func(context.Context, screenplay.Project, screenplay.StoryboardShot) ([]string, error) {
		t.Fatal("text mode resolves images")
		return nil, nil
	}}
	shot := screenplay.StoryboardShot{Seq: 1, Duration: 5, VideoDesc: "@Alice"}
	b := screenplay.Storyboard{ID: "b"}
	in, _, err := r.prepareInput(ctx, p, b, shot, "own", "720p")
	if err != nil || len(in.ImageURLs) > 0 {
		t.Fatal(in, err)
	}
	p.CreativeConfig.UseSubjectMode = true
	signature := "a"
	stable := "media:one"
	r.Subjects = func(context.Context, screenplay.Project, screenplay.StoryboardShot) (screenplayVideoSubjects, error) {
		return screenplayVideoSubjects{Prompt: shot.VideoDesc, URIs: []string{"https://media.example/a?sig=" + signature}, IDs: []string{stable}}, nil
	}
	_, h1, err := r.prepareInput(ctx, p, b, shot, "own", "720p")
	if err != nil {
		t.Fatal(err)
	}
	signature = "b"
	_, h2, _ := r.prepareInput(ctx, p, b, shot, "own", "720p")
	if h1 != h2 {
		t.Fatal("signature refresh changed hash")
	}
	stable = "media:two"
	_, h3, _ := r.prepareInput(ctx, p, b, shot, "own", "720p")
	if h3 == h2 {
		t.Fatal("changed subject kept hash")
	}
}
func TestScreenplaySubjectConfigHTTP(t *testing.T) {
	now := time.Now()
	s := screenplay.NewMemoryStore()
	_, _ = s.CreateProject(context.Background(), screenplay.Project{ID: "p", OwnerUsername: "alice"})
	mux := http.NewServeMux()
	RegisterScreenplayCreativeRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, s, nil)
	w := signedJSONRequest(t, mux, now, "alice", "PUT", "/api/screenplays/p/creative-config", map[string]any{"expected_revision": 0, "use_subject_mode": true, "video_generation_mode": "text_to_video"})
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	w = signedJSONRequest(t, mux, now, "alice", "GET", "/api/screenplays/p/creative-config", nil)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"use_subject_mode":true`) {
		t.Fatal(w.Code, w.Body.String())
	}
	w = signedJSONRequest(t, mux, now, "alice", "PUT", "/api/screenplays/p/creative-config", map[string]any{"expected_revision": 1, "auto_replace_sora2_info": true})
	if w.Code != 400 {
		t.Fatal("dual mode accepted", w.Code, w.Body.String())
	}
	w = signedJSONRequest(t, mux, now, "alice", "PUT", "/api/screenplays/p/creative-config", map[string]any{"expected_revision": 1, "use_subject_mode": false, "auto_replace_sora2_info": true})
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	w = signedJSONRequest(t, mux, now, "alice", "PUT", "/api/screenplays/p/creative-config", map[string]any{"expected_revision": 2, "use_subject_mode": false, "auto_replace_sora2_info": false, "video_generation_mode": ""})
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	p, _ := s.GetProject(context.Background(), "alice", "p")
	if p.CreativeConfig.AutoReplaceSora2Info || p.CreativeConfig.UseSubjectMode || p.CreativeConfig.VideoGenerationMode != "" {
		t.Fatal("explicit disable not saved")
	}
}
func TestScreenplaySubjectExactNamesAndDuplicateExplicitID(t *testing.T) {
	assets := []screenplay.Asset{{ID: "short", Type: screenplay.AssetCharacter, Name: "林晚"}, {ID: "long", Type: screenplay.AssetCharacter, Name: "林晚晴"}, {ID: "dup", Type: screenplay.AssetCharacter, Name: "林晚"}}
	if _, err := screenplayBoundSubjects(screenplay.StoryboardShot{Characters: "林晚"}, assets, true); err == nil {
		t.Fatal("duplicate guessed")
	}
	got, err := screenplayBoundSubjects(screenplay.StoryboardShot{Characters: "林晚晴"}, assets, true)
	if err != nil || len(got) != 1 || got[0].ID != "long" {
		t.Fatal(got, err)
	}
	got, err = screenplayBoundSubjects(screenplay.StoryboardShot{Characters: "林晚", ProtocolFields: []byte(`{"character_ids":["short","short"]}`)}, assets, true)
	if err != nil || len(got) != 1 || got[0].ID != "short" {
		t.Fatal(got, err)
	}
}
func TestScreenplaySubjectFallbackUsesOnlyEntitySelection(t *testing.T) {
	library, first, _, _, _ := assetLibraryServiceFixture(t)
	m := library.Media
	ctx := context.Background()
	p, err := m.Projects.GetProject(ctx, "alice", first.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	p.CreativeConfig = &screenplay.CreativeConfig{UseSubjectMode: true}
	m.ReferencesProvider = agentImageReferenceProvider(func(_ context.Context, a agentworkspace.Asset, _ time.Duration) (string, error) {
		return "https://fixture.invalid/" + a.ID, nil
	})
	shot := screenplay.StoryboardShot{ID: "shot", Characters: first.Name, VideoDesc: "@" + first.Name}
	got, err := m.VideoSubjects(ctx, p, shot)
	if err != nil || len(got.URIs) != 1 || !strings.HasPrefix(got.URIs[0], "https://fixture.invalid/") || !strings.HasPrefix(got.IDs[0], "media:") {
		t.Fatal(got, err)
	}
	_, err = m.Media.MutateMedia(ctx, "alice", p.ID, func(st *screenplay.MediaState) error {
		st.Selected[shot.ID] = st.Selected[first.ID]
		delete(st.Selected, first.ID)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.VideoSubjects(ctx, p, shot); err == nil {
		t.Fatal("used other shot image for missing entity")
	}
}
func TestScreenplaySoraTokenMatchesJavaScriptWhitespace(t *testing.T) {
	for _, space := range []string{"\u00a0", "\u1680", "\u2003", "\u2028", "\u2029", "\u202f", "\u205f", "\u3000", "\ufeff", "\v"} {
		source := "@Alice" + space + "walks"
		got := screenplaySoraToken.FindString(source)
		if got != "@Alice" {
			t.Fatalf("whitespace %q got %q", space, got)
		}
	}
	if got := screenplaySoraToken.FindString("@Alice。 next"); got != "@Alice。" {
		t.Fatal(got)
	}
}

func TestScreenplaySubjectJSONArchiveBoundary(t *testing.T) {
	for _, raw := range []string{`{"asset_id":"asset-own"}`, `{"assetId":"asset-own"}`, `{"assetID":"asset-own"}`, `{"sora2_info":"asset-own"}`, `{"volcengineAssetId":"asset-own"}`, `{"unrelated":"asset-own"}`, `{"asset_id":"wrong","assetId":"asset-own"}`} {
		if uri := screenplaySubjectURI(raw); uri != "asset://asset-own" {
			t.Fatal(raw, uri)
		}
	}
}

type subjectChangedConfigService struct{ fakeScreenplayVideoService }

func (s *subjectChangedConfigService) Submit(context.Context, string, scriptVideoLocalInput) (scriptVideoSubmission, error) {
	return scriptVideoSubmission{}, &scriptVideoPreflightError{Err: fmt.Errorf("model disabled before send")}
}
func TestScreenplaySubjectSubmitRevalidationIsNotUnknown(t *testing.T) {
	c := NewUserConfigStore(t.TempDir())
	_ = c.write("alice", map[string]any{"scriptDefaults": map[string]any{"videoModelKey": "own"}})
	repo := screenplay.NewMemoryProductionRepository()
	r := &screenplayVideoRunner{service: &subjectChangedConfigService{}, configs: c, repo: repo}
	_, err := r.ProduceEpisode(context.Background(), screenplay.Project{ID: "p", OwnerUsername: "alice"}, 1, screenplay.Storyboard{ID: "b", Shots: []screenplay.StoryboardShot{{Seq: 1, Duration: 5, VideoDesc: "shot"}}})
	task, e := repo.GetShotTask(context.Background(), "alice", "p", "b", 1)
	if err == nil || e != nil || task.Status != "not_submitted" {
		t.Fatal(err, e, task)
	}
}
func TestScreenplaySubjectDedupAndClearedArchive(t *testing.T) {
	ctx := context.Background()
	store := screenplay.NewMemoryStore()
	p, _ := store.CreateProject(ctx, screenplay.Project{ID: "p", OwnerUsername: "alice"})
	p.CreativeConfig = &screenplay.CreativeConfig{UseSubjectMode: true}
	m := &ScreenplayMediaWorkbench{Projects: store}
	var character screenplay.Asset
	for _, kind := range []screenplay.AssetType{screenplay.AssetCharacter, screenplay.AssetScene} {
		a, e := store.CreateAsset(ctx, screenplay.Asset{ProjectID: "p", OwnerUsername: "alice", Type: kind, Name: "Same", Extra: map[string]string{"sora2_info": "asset-one"}})
		if e != nil {
			t.Fatal(e)
		}
		if kind == screenplay.AssetCharacter {
			character = a
		}
	}
	shot := screenplay.StoryboardShot{Characters: "Same", Scene: "Same", VideoDesc: "@Same"}
	got, e := m.VideoSubjects(ctx, p, shot)
	if e != nil || len(got.URIs) != 1 || got.URIs[0] != "asset://asset-one" {
		t.Fatal(got, e)
	}
	character.Extra = map[string]string{"sora2_info": "", "subject_upload_result": `{"assetId":"asset-stale"}`}
	if _, e = store.UpdateAsset(ctx, "alice", "p", character.ID, character, character.Revision); e != nil {
		t.Fatal(e)
	}
	if _, e = m.VideoSubjects(ctx, p, shot); e == nil {
		t.Fatal("empty archive revived upload snapshot")
	}
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	if _, e = m.VideoSubjects(ctx, p, shot); e == nil {
		t.Fatal("cancellation ignored")
	}
}
