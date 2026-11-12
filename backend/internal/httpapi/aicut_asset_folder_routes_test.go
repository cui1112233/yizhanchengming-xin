package httpapi

import (
	"net/http"
	"net/http/httptest"
	"qiantie/backend/internal/aicut"
	"testing"
	"time"
)

func TestAICutAssetFolderRoutesPersistAndReturnDeletedFolderAssetsToRoot(t *testing.T) {
	now := time.Now().UTC()
	s := aicut.NewMemoryStore()
	ctx := t.Context()
	if err := s.CreateProject(ctx, "alice", aicut.Project{ID: "p", Title: "工程", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateAsset(ctx, "alice", aicut.Asset{ID: "a", ProjectID: "p", Kind: "image", Name: "a.png", ContentType: "image/png", ObjectKey: "a.png", ByteSize: 1, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAICutAssetRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, s, nil, func() time.Time { return now })
	request := func(owner, method, path string, body any) *httptest.ResponseRecorder {
		return signedJSONRequest(t, mux, now, owner, method, path, body)
	}
	created := request("alice", "POST", "/api/ai-cut/projects/p/asset-folders", map[string]any{"name": "片头"})
	if created.Code != 201 {
		t.Fatalf("create=%d %s", created.Code, created.Body.String())
	}
	var payload struct {
		Folder aicut.AssetFolder `json:"folder"`
	}
	decodeTestJSON(t, created.Body.Bytes(), &payload)
	id := payload.Folder.ID
	for _, test := range []struct {
		owner, method, path string
		body                any
		status              int
	}{
		{"bob", "GET", "/api/ai-cut/projects/p/asset-folders", nil, 404},
		{"bob", "PATCH", "/api/ai-cut/projects/p/assets/a/folder", map[string]any{"folderId": id}, 404},
		{"alice", "PATCH", "/api/ai-cut/projects/p/assets/a/folder", map[string]any{}, 400},
		{"alice", "PATCH", "/api/ai-cut/projects/p/assets/a/folder", map[string]any{"folderId": id}, 200},
		{"alice", "PATCH", "/api/ai-cut/projects/p/asset-folders/" + id, map[string]any{"name": "新片头"}, 200},
	} {
		r := request(test.owner, test.method, test.path, test.body)
		if r.Code != test.status {
			t.Fatalf("%s %s as %s=%d %s", test.method, test.path, test.owner, r.Code, r.Body.String())
		}
	}
	listed := request("alice", "GET", "/api/ai-cut/projects/p/assets", nil)
	var assets struct {
		Assets []aicut.Asset `json:"assets"`
	}
	decodeTestJSON(t, listed.Body.Bytes(), &assets)
	if len(assets.Assets) != 1 || assets.Assets[0].FolderID != id {
		t.Fatalf("assets=%s", listed.Body.String())
	}
	deleted := request("alice", "DELETE", "/api/ai-cut/projects/p/asset-folders/"+id, nil)
	if deleted.Code != 200 {
		t.Fatalf("delete=%s", deleted.Body.String())
	}
	listed = request("alice", "GET", "/api/ai-cut/projects/p/assets", nil)
	assets.Assets = nil
	decodeTestJSON(t, listed.Body.Bytes(), &assets)
	if len(assets.Assets) != 1 || assets.Assets[0].FolderID != "" {
		t.Fatalf("delete lost assets or kept membership: %s", listed.Body.String())
	}
}
func TestAICutAssetFoldersRemainProjectScopedOnNamedTimeline(t *testing.T) {
	now := time.Now().UTC()
	store := aicut.NewMemoryStore()
	if err := store.CreateProject(t.Context(), "alice", aicut.Project{ID: "p", Title: "工程", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	mux := http.NewServeMux()
	wrapped := aicutRequestTimelineStore{aicutExportTimelineStore{Store: store}}
	RegisterAICutAssetRoutes(mux, auth, wrapped, nil, nil)
	result := signedJSONRequestWithTimeline(t, withAICutTimelineScope(mux, auth), now, "POST", "/api/ai-cut/projects/p/asset-folders", "timeline-two", map[string]any{"name": "共享素材"})
	if result.Code != 201 {
		t.Fatalf("named timeline folder=%d %s", result.Code, result.Body.String())
	}
}
