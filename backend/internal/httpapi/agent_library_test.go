package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
)

func TestAgentMaterialLibrarySavesOwnerAssetAndInsertsItWithProvenance(t *testing.T) {
	now := time.Date(2026, 10, 24, 13, 0, 0, 0, time.UTC)
	store := agentworkspace.NewMemoryStore()
	source := generationHistoryProject("project-source", "image-1", "image", now)
	target := generationHistoryProject("project-target", "text-1", "text", now)
	other := generationHistoryProject("project-other", "image-2", "image", now)
	for _, project := range []agentworkspace.Project{source, target} {
		if err := store.CreateProject(context.Background(), "alice", project); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.CreateProject(context.Background(), "bob", other); err != nil {
		t.Fatal(err)
	}
	assets := agentworkspace.NewMemoryAssetStore()
	if err := assets.CreateAsset(context.Background(), "alice", agentworkspace.Asset{ID: "asset-source", ProjectID: source.ID, Kind: "image", Name: "女主参考.png", ContentType: "image/png", ObjectKey: "agent-assets/alice/project-source/asset-source.png", ByteSize: 4, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := assets.CreateAsset(context.Background(), "bob", agentworkspace.Asset{ID: "asset-other", ProjectID: other.ID, Kind: "image", Name: "他人素材.png", ContentType: "image/png", ObjectKey: "agent-assets/bob/project-other/asset-other.png", ByteSize: 4, CreatedAt: now}); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	RegisterAgentMaterialLibraryRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, assets, func() time.Time { return now })
	folder := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/material-library/folders", map[string]any{"name": "角色参考"})
	if folder.Code != http.StatusCreated {
		t.Fatalf("folder=%d %s", folder.Code, folder.Body.String())
	}
	var folderBody struct {
		Folder agentworkspace.LibraryFolder `json:"folder"`
	}
	if err := json.Unmarshal(folder.Body.Bytes(), &folderBody); err != nil || folderBody.Folder.ID == "" {
		t.Fatalf("folder body=%s err=%v", folder.Body.String(), err)
	}
	saved := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/material-library/folders/"+folderBody.Folder.ID+"/items", map[string]any{"projectId": source.ID, "assetId": "asset-source", "tags": []string{"女主", "古装"}})
	if saved.Code != http.StatusCreated {
		t.Fatalf("save=%d %s", saved.Code, saved.Body.String())
	}
	var savedBody struct {
		Item agentworkspace.LibraryItem `json:"item"`
	}
	if err := json.Unmarshal(saved.Body.Bytes(), &savedBody); err != nil || savedBody.Item.ID == "" {
		t.Fatalf("saved body=%s err=%v", saved.Body.String(), err)
	}
	foreign := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/material-library/folders/"+folderBody.Folder.ID+"/items", map[string]any{"projectId": other.ID, "assetId": "asset-other"})
	if foreign.Code != http.StatusNotFound {
		t.Fatalf("foreign=%d %s", foreign.Code, foreign.Body.String())
	}
	inserted := signedJSONRequest(t, mux, now, "alice", http.MethodPost, "/api/agent/tasks/"+target.ID+"/material-library/items/"+savedBody.Item.ID+"/insert", nil)
	if inserted.Code != http.StatusCreated {
		t.Fatalf("insert=%d %s", inserted.Code, inserted.Body.String())
	}
	canvas, err := store.Canvas(context.Background(), "alice", target.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(canvas.Nodes) != 2 {
		t.Fatalf("nodes=%+v", canvas.Nodes)
	}
	var imported agentworkspace.Node
	for _, node := range canvas.Nodes {
		if node.AssetRef != nil {
			imported = node
		}
	}
	if imported.AssetRef["originProjectId"] != source.ID || imported.AssetRef["originAssetId"] != "asset-source" || imported.AssetRef["libraryItemId"] != savedBody.Item.ID {
		t.Fatalf("imported=%+v", imported)
	}
	importedAssetID, _ := imported.AssetRef["assetId"].(string)
	importedAsset, err := assets.Asset(context.Background(), "alice", target.ID, importedAssetID)
	if err != nil || importedAsset.ObjectKey != "agent-assets/alice/project-source/asset-source.png" {
		t.Fatalf("asset=%+v err=%v", importedAsset, err)
	}
}

func TestAgentMaterialLibraryListsOnlyCurrentOwnerFolders(t *testing.T) {
	now := time.Date(2026, 10, 24, 13, 30, 0, 0, time.UTC)
	store := agentworkspace.NewMemoryStore()
	if _, err := store.CreateLibraryFolder(context.Background(), "alice", agentworkspace.CreateLibraryFolderInput{ID: "folder-alice", Name: "Alice", CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateLibraryFolder(context.Background(), "bob", agentworkspace.CreateLibraryFolderInput{ID: "folder-bob", Name: "Bob", CreatedAt: now}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAgentMaterialLibraryRoutes(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, store, nil, func() time.Time { return now })
	response := signedJSONRequest(t, mux, now, "alice", http.MethodGet, "/api/agent/material-library/folders", nil)
	if response.Code != http.StatusOK || !containsText(response.Body.String(), "folder-alice") || containsText(response.Body.String(), "folder-bob") {
		t.Fatalf("folders=%d %s", response.Code, response.Body.String())
	}
}
