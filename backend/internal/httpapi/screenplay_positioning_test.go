package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/screenplay"
	"strings"
	"testing"
	"time"
)

func TestPositioningRecoveredGaaPromptEvidence(t *testing.T) {
	// Original: F:/软件/洛水2026/洛水2026/dist/assets/index-DszfAkKD.js,
	// Gaa JS string in UTF-16 [11442466,11443327), before qaa declaration.
	// Summary: lock reference identities/clothes and scene, modify positions only,
	// use left-hand-drive seating, never clone characters; retain reference style.
	// The analysis rule remains independently authored, not copied from a hidden backend.
	if got := fmt.Sprintf("%x", sha256.Sum256([]byte(prompt.ScreenplayPositioningPrefix))); got != "b4199db7643dc6bfcff1a4366a55625b4c9e53646b9f54b8be6fb25f2b196068" {
		t.Fatal("recovered Gaa text changed:", got)
	}
}

func positioningFixture(t *testing.T) (*ScreenplayMediaWorkbench, screenplay.Project) {
	t.Helper()
	ctx := context.Background()
	projects := screenplay.NewMemoryStore()
	p, e := projects.CreateProject(ctx, screenplay.Project{ID: "position-p", OwnerUsername: "alice", Title: "站位", AspectRatio: "9:16", SourceText: "# 第1集\n完整剧本内容", CreativeConfig: &screenplay.CreativeConfig{ImageModelID: "gen", ImageEditModelID: "edit"}})
	if e != nil {
		t.Fatal(e)
	}
	s := &ScreenplayMediaWorkbench{Projects: projects, Media: screenplay.NewMediaMemoryStore(), Cut: aicut.NewMemoryStore(), Objects: &memoryImageToolObjects{files: map[string][]byte{}}, Assets: agentworkspace.NewMemoryAssetStore()}
	_, e = s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(st *screenplay.MediaState) error {
		st.Positioning = &screenplay.PositioningState{Items: []screenplay.PositioningItem{{ID: "one", Episode: "1", SceneName: "房间", Characters: []string{"乙", "甲"}, PositionDetail: "甲在左乙在右", Remark: "不换座"}}, ShowCharacterNames: true}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	return s, p
}
func rev(v int64) *int64 { return &v }

func TestPositioningCRUDRevisionOwnershipAndProtectedURL(t *testing.T) {
	s, p := positioningFixture(t)
	mux := http.NewServeMux()
	now := time.Now()
	RegisterScreenplayPositioningRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, s)
	url := "/api/comic-screenplays/projects/" + p.ID + "/positioning"
	if resp := signedJSONRequest(t, mux, now, "bob", "GET", url, nil); resp.Code == 200 {
		t.Fatal("foreign project read")
	}
	body := map[string]any{"expected_revision": 0, "items": []map[string]any{{"id": "two", "episode": 2, "scene_name": "房间", "position_detail": "站位2"}, {"id": "one", "position_detail": "站位1", "image_path": "https://evil/foreign", "image_media_item_id": "foreign"}}, "show_character_names": true}
	resp := signedJSONRequest(t, mux, now, "alice", "PUT", url, body)
	if resp.Code != 200 {
		t.Fatal(resp.Code, resp.Body.String())
	}
	var out struct {
		State screenplay.PositioningState `json:"state"`
	}
	if e := json.Unmarshal(resp.Body.Bytes(), &out); e != nil {
		t.Fatal(e)
	}
	if out.State.Revision != 1 || out.State.Items[0].Index != 1 || out.State.Items[0].Episode != "2" || out.State.Items[1].ImagePath != "" || out.State.Items[1].ImageMediaItemID != "" || !out.State.ShowCharacterNames {
		t.Fatalf("state=%+v", out.State)
	}
	if resp = signedJSONRequest(t, mux, now, "alice", "PUT", url, body); resp.Code != 409 {
		t.Fatal(resp.Code, resp.Body.String())
	}
	body["expected_revision"] = 1
	body["items"] = []map[string]any{{"id": "same"}, {"id": "same"}}
	if resp = signedJSONRequest(t, mux, now, "alice", "PUT", url, body); resp.Code != 400 {
		t.Fatal(resp.Code, resp.Body.String())
	}
	body["items"] = []any{}
	if resp = signedJSONRequest(t, mux, now, "alice", "PUT", url, body); resp.Code != 200 {
		t.Fatal(resp.Code, resp.Body.String())
	}
}

func TestPositioningRealImageRequestReferencesAndIdempotency(t *testing.T) {
	s, p := positioningFixture(t)
	ctx := context.Background()
	referenceIDs := map[string]string{}
	for _, name := range []string{"甲", "乙", "房间"} {
		kind := screenplay.AssetCharacter
		if name == "房间" {
			kind = screenplay.AssetScene
		}
		asset, e := s.Projects.CreateAsset(ctx, screenplay.Asset{ID: "asset-" + fmt.Sprint(len(referenceIDs)), OwnerUsername: p.OwnerUsername, ProjectID: p.ID, Name: name, Type: kind})
		if e != nil {
			t.Fatal(e)
		}
		m, e := s.saveBytes(ctx, p, "image", name+".png", "image/png", bytes.NewReader(testImageToolPNG(t)), asset.ID)
		if e != nil {
			t.Fatal(e)
		}
		referenceIDs[name] = m.AssetID
	}
	marked, e := s.saveBytes(ctx, p, "image", "marked.png", "image/png", bytes.NewReader(testImageToolPNG(t)), "")
	if e != nil {
		t.Fatal(e)
	}
	s.ReferencesProvider = agentImageReferenceProvider(func(_ context.Context, a agentworkspace.Asset, _ time.Duration) (string, error) {
		return "https://assets.example/" + a.ID, nil
	})
	calls := 0
	s.Images = func(_ context.Context, owner string, in AgentImageGenerationInput) ([]string, int, error) {
		calls++
		want := []string{marked.AssetID, referenceIDs["乙"], referenceIDs["甲"], referenceIDs["房间"]}
		if owner != "alice" || in.ModelID != "edit" || in.AspectRatio != "16:9" || in.Size != "2048x1152" || in.Resolution != "2K" || in.Count != 1 || strings.Join(in.ReferenceAssetIDs, ",") != strings.Join(want, ",") || len(in.ReferenceImages) != 4 || !strings.Contains(in.Prompt, "完整剧本内容") || !strings.Contains(in.Prompt, "不换座") || !strings.Contains(in.Prompt, "禁止复制人物") {
			t.Fatalf("owner=%s request=%+v", owner, in)
		}
		return []string{"data:image/png;base64," + base64.StdEncoding.EncodeToString(testImageToolPNG(t))}, 200, nil
	}
	in := positioningRequest{ExpectedRevision: rev(0), IdempotencyKey: "once", MarkedImageItemID: marked.ID}
	item, missing, e := s.generatePositioning(ctx, p, "one", in)
	if e != nil || len(missing) != 0 || item.Status != "succeeded" || item.ImageMediaItemID == "" {
		t.Fatal(item, missing, e)
	}
	again, _, e := s.generatePositioning(ctx, p, "one", in)
	if e != nil || again.ImageMediaItemID != item.ImageMediaItemID || calls != 1 {
		t.Fatal(again, e, calls)
	}
	mux := http.NewServeMux()
	now := time.Now()
	RegisterScreenplayPositioningRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, s)
	response := signedJSONRequest(t, mux, now, "alice", "GET", "/api/comic-screenplays/projects/"+p.ID+"/positioning", nil)
	if response.Code != 200 || !strings.Contains(response.Body.String(), "access=") {
		t.Fatal(response.Code, response.Body.String())
	}
}

func TestPositioningSelectedReferenceOverridesLatest(t *testing.T) {
	s, p := positioningFixture(t)
	ctx := context.Background()
	a, e := s.Projects.CreateAsset(ctx, screenplay.Asset{ID: "chosen-role", OwnerUsername: p.OwnerUsername, ProjectID: p.ID, Name: "甲", Type: screenplay.AssetCharacter})
	if e != nil {
		t.Fatal(e)
	}
	old, e := s.saveBytes(ctx, p, "image", "old.png", "image/png", bytes.NewReader(testImageToolPNG(t)), a.ID)
	if e != nil {
		t.Fatal(e)
	}
	newest, e := s.saveBytes(ctx, p, "image", "new.png", "image/png", bytes.NewReader(testImageToolPNG(t)), a.ID)
	if e != nil {
		t.Fatal(e)
	}
	_, e = s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(st *screenplay.MediaState) error { st.Selected[a.ID] = old.ID; return nil })
	if e != nil {
		t.Fatal(e)
	}
	st, _ := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	refs, _, e := s.positioningReferences(ctx, p, st, screenplay.PositioningItem{Characters: []string{"甲"}}, "")
	if e != nil || len(refs) != 1 || refs[0] != old.AssetID || refs[0] == newest.AssetID {
		t.Fatal(refs, e)
	}
	st.Selected[a.ID] = "foreign-owner-media"
	if _, _, e = s.positioningReferences(ctx, p, st, screenplay.PositioningItem{Characters: []string{"甲"}}, ""); e == nil {
		t.Fatal("foreign selected media accepted")
	}
}

func TestPositioningMissingReferencesCancellationAndFailure(t *testing.T) {
	s, p := positioningFixture(t)
	calls := 0
	s.Images = func(context.Context, string, AgentImageGenerationInput) ([]string, int, error) {
		calls++
		return nil, 500, fmt.Errorf("offline")
	}
	in := positioningRequest{ExpectedRevision: rev(0), IdempotencyKey: "missing"}
	_, missing, e := s.generatePositioning(context.Background(), p, "one", in)
	if e == nil || len(missing) != 3 || calls != 0 {
		t.Fatal(missing, e, calls)
	}
	in.AllowMissingRefs = true
	_, _, e = s.generatePositioning(context.Background(), p, "one", in)
	if e == nil || calls != 1 {
		t.Fatal(e, calls)
	}
	st, _ := s.Media.LoadMedia(context.Background(), p.OwnerUsername, p.ID)
	if st.Positioning.Items[0].Status != "failed" {
		t.Fatal(st.Positioning)
	}
	in.ExpectedRevision = rev(st.Positioning.Revision)
	in.IdempotencyKey = "cancel"
	ctx, cancel := context.WithCancel(context.Background())
	s.Images = func(context.Context, string, AgentImageGenerationInput) ([]string, int, error) {
		cancel()
		return []string{"data:image/png;base64," + base64.StdEncoding.EncodeToString(testImageToolPNG(t))}, 200, nil
	}
	_, _, e = s.generatePositioning(ctx, p, "one", in)
	if !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	st, _ = s.Media.LoadMedia(context.Background(), p.OwnerUsername, p.ID)
	if st.Positioning.Items[0].Status != "cancelled" || st.Positioning.Items[0].ImageMediaItemID != "" {
		t.Fatal(st.Positioning)
	}
}

func TestPositioningAnalysisModesContextAndAtomicFailure(t *testing.T) {
	s, p := positioningFixture(t)
	ctx := context.Background()
	_, e := s.Projects.SaveScreenplay(ctx, screenplay.Document{ID: "doc", OwnerUsername: p.OwnerUsername, ProjectID: p.ID, Status: "active", Content: "# 第1集\n真正激活剧本"}, nil, nil, nil)
	if e != nil {
		t.Fatal(e)
	}
	calls := 0
	s.PositioningText = func(_ context.Context, owner string, in screenplay.TextRequest) (screenplay.TextResponse, error) {
		calls++
		if owner != "alice" || in.Model != "text-model" || !in.JSONResponse || !strings.Contains(in.UserPrompt, "真正激活剧本") || strings.Contains(in.UserPrompt, "完整剧本内容") || !strings.Contains(in.SystemPrompt, "复杂") {
			t.Fatalf("owner=%s request=%+v", owner, in)
		}
		return screenplay.TextResponse{Content: `{"items":[]}`}, nil
	}
	in := positioningRequest{ExpectedRevision: rev(0), Model: "text-model", AnalyzeMode: "complex"}
	if e = s.analyzePositioning(ctx, p, in); e != nil {
		t.Fatal(e)
	}
	st, _ := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if len(st.Positioning.Items) != 0 || st.Positioning.Revision != 1 {
		t.Fatal(st.Positioning)
	}
	in.ExpectedRevision = rev(1)
	in.AnalyzeMode = "global"
	s.PositioningText = func(_ context.Context, _ string, req screenplay.TextRequest) (screenplay.TextResponse, error) {
		if !strings.Contains(req.SystemPrompt, "全部") {
			t.Fatal(req.SystemPrompt)
		}
		return screenplay.TextResponse{Content: `{"bad":[]}`}, nil
	}
	if e = s.analyzePositioning(ctx, p, in); !errors.Is(e, screenplay.ErrAIInvalidJSON) {
		t.Fatal(e)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	s.PositioningText = func(context.Context, string, screenplay.TextRequest) (screenplay.TextResponse, error) {
		cancel()
		return screenplay.TextResponse{Content: `{"items":[{"id":"late"}]}`}, nil
	}
	if e = s.analyzePositioning(cancelCtx, p, in); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	st, _ = s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if st.Positioning.Revision != 1 || len(st.Positioning.Items) != 0 {
		t.Fatal(st.Positioning)
	}
	if calls != 1 {
		t.Fatal(calls)
	}
}

func TestPositioningConcurrentEditBlocksLateImageAndAnalysis(t *testing.T) {
	s, p := positioningFixture(t)
	ctx := context.Background()
	s.Images = func(context.Context, string, AgentImageGenerationInput) ([]string, int, error) {
		_, e := s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(st *screenplay.MediaState) error {
			st.Positioning.Revision++
			st.Positioning.Items[0].PositionDetail = "用户刚改的位置"
			st.Positioning.Items[0].RequestKey = ""
			st.Positioning.Items[0].Status = ""
			return nil
		})
		if e != nil {
			t.Fatal(e)
		}
		return []string{"data:image/png;base64," + base64.StdEncoding.EncodeToString(testImageToolPNG(t))}, 200, nil
	}
	_, _, e := s.generatePositioning(ctx, p, "one", positioningRequest{ExpectedRevision: rev(0), IdempotencyKey: "race", AllowMissingRefs: true})
	if !errors.Is(e, screenplay.ErrRevisionConflict) {
		t.Fatal(e)
	}
	st, _ := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if st.Positioning.Items[0].PositionDetail != "用户刚改的位置" || st.Positioning.Items[0].ImageMediaItemID != "" || len(st.Items) != 0 {
		t.Fatal(st)
	}
	version := st.Positioning.Revision
	s.PositioningText = func(context.Context, string, screenplay.TextRequest) (screenplay.TextResponse, error) {
		_, e := s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(st *screenplay.MediaState) error { st.Positioning.Revision++; return nil })
		return screenplay.TextResponse{Content: `{"items":[]}`}, e
	}
	if e = s.analyzePositioning(ctx, p, positioningRequest{ExpectedRevision: rev(version), Model: "model", AnalyzeMode: "global"}); !errors.Is(e, screenplay.ErrRevisionConflict) {
		t.Fatal(e)
	}
	st, _ = s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if len(st.Positioning.Items) != 1 || st.Positioning.Items[0].PositionDetail != "用户刚改的位置" {
		t.Fatal(st)
	}
}

func TestPositioningUploadedImageBindingAndExplicitEdit(t *testing.T) {
	s, p := positioningFixture(t)
	ctx := context.Background()
	uploaded, e := s.saveBytes(ctx, p, "image", "站位.png", "image/png", bytes.NewReader(testImageToolPNG(t)), "")
	if e != nil {
		t.Fatal(e)
	}
	if e = s.bindPositioningImage(ctx, p, "one", positioningRequest{ExpectedRevision: rev(0), MediaItemID: "foreign-id"}); !errors.Is(e, screenplay.ErrInvalidArgument) {
		t.Fatal(e)
	}
	if e = s.bindPositioningImage(ctx, p, "one", positioningRequest{ExpectedRevision: rev(0), MediaItemID: uploaded.ID}); e != nil {
		t.Fatal(e)
	}
	st, _ := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if st.Positioning.Revision != 1 || st.Positioning.Items[0].ImageMediaItemID != uploaded.ID {
		t.Fatal(st)
	}
	s.ReferencesProvider = agentImageReferenceProvider(func(_ context.Context, a agentworkspace.Asset, _ time.Duration) (string, error) {
		return "https://assets.example/" + a.ID, nil
	})
	calls := 0
	s.Images = func(_ context.Context, owner string, in AgentImageGenerationInput) ([]string, int, error) {
		calls++
		if owner != "alice" || in.ModelID != "edit" || in.AspectRatio != "16:9" || len(in.ReferenceAssetIDs) != 1 || in.ReferenceAssetIDs[0] != uploaded.AssetID || !strings.Contains(in.Prompt, "将甲稍向左移动") || !strings.Contains(in.Prompt, "只按以下修改要求编辑") {
			t.Fatalf("owner=%s request=%+v", owner, in)
		}
		return []string{"data:image/png;base64," + base64.StdEncoding.EncodeToString(testImageToolPNG(t))}, 200, nil
	}
	in := positioningRequest{ExpectedRevision: rev(1), IdempotencyKey: "edit-once", EditPrompt: "将甲稍向左移动", AllowMissingRefs: true}
	item, _, e := s.generatePositioning(ctx, p, "one", in)
	if e != nil || item.ImageMediaItemID == uploaded.ID || calls != 1 {
		t.Fatal(item, e, calls)
	}
	mux := http.NewServeMux()
	now := time.Now()
	RegisterScreenplayPositioningRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, s)
	response := signedJSONRequest(t, mux, now, "alice", "POST", "/api/comic-screenplays/projects/"+p.ID+"/positioning/one/image/edit", map[string]any{"expected_revision": item.Index, "idempotencyKey": "empty"})
	if response.Code != 400 || calls != 1 {
		t.Fatal(response.Code, response.Body.String(), calls)
	}
}
