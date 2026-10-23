package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/batchfactoryv11"
)

type batchFactoryDeleteGuard struct {
	*batchfactoryv11.MemoryStore
	deleteBookAssetCalls int
}

type duplicateBindingRaceStore struct {
	*agentworkspace.MemoryStore
}

func (s *duplicateBindingRaceStore) BindBatchFactoryProject(ctx context.Context, binding agentworkspace.BatchFactoryBinding) (agentworkspace.BatchFactoryBinding, error) {
	if _, err := s.MemoryStore.BindBatchFactoryProject(ctx, binding); err != nil {
		return agentworkspace.BatchFactoryBinding{}, err
	}
	// Match the MySQL unique-key race: another request has already persisted
	// the deterministic binding, while this insert reports a duplicate.
	return agentworkspace.BatchFactoryBinding{}, agentworkspace.ErrBatchFactoryBindingConflict
}

func (s *batchFactoryDeleteGuard) DeleteBookAsset(ctx context.Context, owner, batchID, bookID, assetID string) error {
	s.deleteBookAssetCalls++
	return s.MemoryStore.DeleteBookAsset(ctx, owner, batchID, bookID, assetID)
}

func TestBindBatchFactoryReturnsSameProjectForSameOwner(t *testing.T) {
	now := time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC)
	batches := batchfactoryv11.NewMemoryStore()
	batch, err := batches.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "private batch title", Books: []batchfactoryv11.CreateBookInput{{Title: "private book title", SourceText: "private source text"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	projects := agentworkspace.NewMemoryStore()
	mux := http.NewServeMux()
	RegisterAgentBatchFactoryBridge(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, projects, batches, nil, func() time.Time { return now })

	path := "/api/batch-factory/v11/batches/" + batch.ID + "/agent-project"
	first := agentRequest(t, mux, now, http.MethodPost, path, "")
	second := agentRequest(t, mux, now, http.MethodPost, path, "")
	if first.Code != http.StatusOK || second.Code != http.StatusOK {
		t.Fatalf("first=%d %s second=%d %s", first.Code, first.Body.String(), second.Code, second.Body.String())
	}
	var one, two agentBatchFactoryBindingResponse
	if err := json.Unmarshal(first.Body.Bytes(), &one); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(second.Body.Bytes(), &two); err != nil {
		t.Fatal(err)
	}
	if one.ProjectID == "" || one.ProjectID != two.ProjectID || one.State != "bound" || one.Source != "batch_factory" {
		t.Fatalf("first=%+v second=%+v", one, two)
	}
	if _, err := projects.Project(context.Background(), "alice", one.ProjectID); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(first.Body.String(), "private batch title") || strings.Contains(first.Body.String(), "private book title") || strings.Contains(first.Body.String(), "private source text") {
		t.Fatalf("binding leaked private Batch Factory fields: %s", first.Body.String())
	}
}

func TestBindBatchFactoryRecoversConcurrentDuplicateBinding(t *testing.T) {
	now := time.Date(2026, 10, 23, 12, 1, 0, 0, time.UTC)
	batches := batchfactoryv11.NewMemoryStore()
	batch, err := batches.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch"})
	if err != nil {
		t.Fatal(err)
	}
	projects := &duplicateBindingRaceStore{MemoryStore: agentworkspace.NewMemoryStore()}
	mux := http.NewServeMux()
	RegisterAgentBatchFactoryBridge(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, projects, batches, nil, func() time.Time { return now })

	path := "/api/batch-factory/v11/batches/" + batch.ID + "/agent-project"
	start := make(chan struct{})
	responses := make(chan *httptest.ResponseRecorder, 2)
	var group sync.WaitGroup
	for range 2 {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			req := httptest.NewRequest(http.MethodPost, path, nil)
			SignBridgeRequest(req, "alice", true, now, "secret")
			recorder := httptest.NewRecorder()
			mux.ServeHTTP(recorder, req)
			responses <- recorder
		}()
	}
	close(start)
	group.Wait()
	close(responses)
	for response := range responses {
		if response.Code != http.StatusOK {
			t.Fatalf("concurrent bind status=%d body=%s", response.Code, response.Body.String())
		}
	}
	projectsList, err := projects.ListProjects(context.Background(), "alice")
	if err != nil || len(projectsList) != 1 {
		t.Fatalf("projects=%+v err=%v", projectsList, err)
	}
}

func TestBatchFactoryAgentBridgeRegistersThroughRouter(t *testing.T) {
	now := time.Date(2026, 10, 23, 12, 2, 0, 0, time.UTC)
	batches := batchfactoryv11.NewMemoryStore()
	batch, err := batches.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch"})
	if err != nil {
		t.Fatal(err)
	}
	api := NewRouter(RouterOptions{
		BridgeSecret: "secret", Now: func() time.Time { return now }, Store: batches, AgentWorkspace: agentworkspace.NewMemoryStore(),
		UserConfigStore: NewUserConfigStore(t.TempDir()),
	})
	response := agentRequest(t, api, now, http.MethodPost, "/api/batch-factory/v11/batches/"+batch.ID+"/agent-project", "")
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestSyncBookRejectsAnotherOwnersBatch(t *testing.T) {
	now := time.Date(2026, 10, 23, 12, 5, 0, 0, time.UTC)
	batches := batchfactoryv11.NewMemoryStore()
	batch, err := batches.CreateBatch(context.Background(), "bob", batchfactoryv11.CreateBatchInput{
		Title: "bob private batch", Books: []batchfactoryv11.CreateBookInput{{Title: "bob private book", SourceText: "bob private source"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	projects := agentworkspace.NewMemoryStore()
	project := agentworkspace.Project{ID: "agent-alice", Title: "Agent", CreatedAt: now, UpdatedAt: now, Canvas: agentworkspace.EmptyCanvas()}
	if err := projects.CreateProject(context.Background(), "alice", project); err != nil {
		t.Fatal(err)
	}
	if _, err := projects.BindBatchFactoryProject(context.Background(), agentworkspace.BatchFactoryBinding{Owner: "alice", BatchID: batch.ID, ProjectID: project.ID, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterAgentBatchFactoryBridge(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, projects, batches, nil, func() time.Time { return now })

	response := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/"+project.ID+"/batch-factory/books/"+batch.Books[0].ID+"/sync", "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "bob private batch") || strings.Contains(response.Body.String(), "bob private book") || strings.Contains(response.Body.String(), "bob private source") {
		t.Fatalf("cross-owner response leaked Batch Factory fields: %s", response.Body.String())
	}
}

func TestSyncBookDoesNotPersistSignedMediaURL(t *testing.T) {
	now := time.Date(2026, 10, 23, 12, 7, 0, 0, time.UTC)
	batches := batchfactoryv11.NewMemoryStore()
	batch, err := batches.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{Title: "book"}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	asset, err := batches.CreateBookAsset(context.Background(), "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "character", Name: "hero", Prompt: "portrait"})
	if err != nil {
		t.Fatal(err)
	}
	const signedURL = "https://media.example/image.png?X-Amz-Signature=private-signature&token=private-token"
	if _, err := batches.CreateBookAssetImage(context.Background(), "alice", batch.ID, book.ID, asset.ID, batchfactoryv11.CreateBookAssetImageInput{URL: signedURL, MediaType: "image/png"}); err != nil {
		t.Fatal(err)
	}
	projects := agentworkspace.NewMemoryStore()
	mux := http.NewServeMux()
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	RegisterAgentWorkspaceCompat(mux, auth, projects, "", nil, func() time.Time { return now })
	RegisterAgentBatchFactoryBridge(mux, auth, projects, batches, nil, func() time.Time { return now })

	bound := agentRequest(t, mux, now, http.MethodPost, "/api/batch-factory/v11/batches/"+batch.ID+"/agent-project", "")
	if bound.Code != http.StatusOK {
		t.Fatalf("bind status=%d body=%s", bound.Code, bound.Body.String())
	}
	var binding agentBatchFactoryBindingResponse
	if err := json.Unmarshal(bound.Body.Bytes(), &binding); err != nil {
		t.Fatal(err)
	}
	synced := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/"+binding.ProjectID+"/batch-factory/books/"+book.ID+"/sync", "")
	if synced.Code != http.StatusOK {
		t.Fatalf("sync status=%d body=%s", synced.Code, synced.Body.String())
	}
	stored, err := projects.Project(context.Background(), "alice", binding.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	storedJSON, _ := json.Marshal(stored)
	if strings.Contains(string(storedJSON), "private-signature") || strings.Contains(string(storedJSON), "private-token") {
		t.Fatalf("signed media URL persisted in canvas: %s", storedJSON)
	}
	read := agentRequest(t, mux, now, http.MethodGet, "/api/agent/tasks/"+binding.ProjectID, "")
	if read.Code != http.StatusOK || strings.Contains(read.Body.String(), "private-signature") || strings.Contains(read.Body.String(), "private-token") {
		t.Fatalf("task response leaked signed media URL: status=%d body=%s", read.Code, read.Body.String())
	}
}

func TestBatchFactoryProjectionMediaReadUsesCanonicalBookID(t *testing.T) {
	now := time.Date(2026, 10, 23, 19, 50, 0, 0, time.UTC)
	batches := batchfactoryv11.NewMemoryStore()
	batch, err := batches.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{
		BookID: "2091926811672429633", Title: "legacy source book",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	if book.ID == book.BookID || book.BookID != "2091926811672429633" {
		t.Fatalf("fixture must keep distinct canonical and legacy book IDs: %+v", book)
	}
	asset, err := batches.CreateBookAsset(context.Background(), "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "character", Name: "hero", Prompt: "portrait"})
	if err != nil {
		t.Fatal(err)
	}
	const mediaURL = "https://media.example/hero.png"
	image, err := batches.CreateBookAssetImage(context.Background(), "alice", batch.ID, book.ID, asset.ID, batchfactoryv11.CreateBookAssetImageInput{URL: mediaURL, MediaType: "image/png", Source: "provider"})
	if err != nil {
		t.Fatal(err)
	}
	projects := agentworkspace.NewMemoryStore()
	mux := http.NewServeMux()
	auth := BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}
	RegisterAgentWorkspaceCompat(mux, auth, projects, "", nil, func() time.Time { return now })
	RegisterAgentBatchFactoryBridge(mux, auth, projects, batches, nil, func() time.Time { return now })

	bound := agentRequest(t, mux, now, http.MethodPost, "/api/batch-factory/v11/batches/"+batch.ID+"/agent-project", "")
	if bound.Code != http.StatusOK {
		t.Fatalf("bind status=%d body=%s", bound.Code, bound.Body.String())
	}
	var binding agentBatchFactoryBindingResponse
	if err := json.Unmarshal(bound.Body.Bytes(), &binding); err != nil {
		t.Fatal(err)
	}
	synced := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/"+binding.ProjectID+"/batch-factory/books/"+book.ID+"/sync", "")
	if synced.Code != http.StatusOK {
		t.Fatalf("sync status=%d body=%s", synced.Code, synced.Body.String())
	}
	canvas, err := projects.Canvas(context.Background(), "alice", binding.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	var nodeID string
	for _, node := range canvas.Nodes {
		if node.Config["resourceType"] == "book_asset" && node.Config["resourceId"] == asset.ID && node.AssetRef["id"] == image.ID {
			nodeID = node.ID
			break
		}
	}
	if nodeID == "" {
		t.Fatalf("missing projected image node: %+v", canvas.Nodes)
	}
	canvasResponse := agentRequest(t, mux, now, http.MethodGet, "/api/agent/tasks/"+binding.ProjectID+"/canvas", "")
	wantMediaPath := "/api/agent/tasks/" + binding.ProjectID + "/batch-factory/nodes/" + nodeID + "/media"
	if canvasResponse.Code != http.StatusOK || !strings.Contains(canvasResponse.Body.String(), wantMediaPath) || strings.Contains(canvasResponse.Body.String(), "/assets/"+image.ID+"/content") {
		t.Fatalf("canvas response must expose the mapped media reader only: status=%d body=%s", canvasResponse.Code, canvasResponse.Body.String())
	}

	media := agentRequest(t, mux, now, http.MethodGet, "/api/agent/tasks/"+binding.ProjectID+"/batch-factory/nodes/"+nodeID+"/media", "")
	if media.Code != http.StatusFound || media.Header().Get("Location") != mediaURL {
		t.Fatalf("media status=%d location=%q body=%s", media.Code, media.Header().Get("Location"), media.Body.String())
	}
}

func TestBatchFactorySyncUsesMappedStateWithoutAgentGenerationTask(t *testing.T) {
	now := time.Date(2026, 10, 23, 12, 8, 0, 0, time.UTC)
	batches := batchfactoryv11.NewMemoryStore()
	batch, err := batches.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{Title: "batch", Books: []batchfactoryv11.CreateBookInput{{Title: "book", Videos: []batchfactoryv11.CreateVideoInput{{Label: "VIDEO 01", VideoPrompt: "prompt"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	if _, err := batches.CreateProductionJob(context.Background(), batchfactoryv11.ProductionJob{
		Owner: "alice", BatchID: batch.ID, BookID: book.ID, RequestID: "failed-production", Status: batchfactoryv11.ProductionFailed,
		Tasks: []batchfactoryv11.ProductionTask{{VideoID: book.Videos[0].ID, Status: batchfactoryv11.ProductionFailed, ErrorMessage: "temporary provider failure"}}, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	projects := agentworkspace.NewMemoryStore()
	mux := http.NewServeMux()
	RegisterAgentBatchFactoryBridge(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, projects, batches, nil, func() time.Time { return now })

	bound := agentRequest(t, mux, now, http.MethodPost, "/api/batch-factory/v11/batches/"+batch.ID+"/agent-project", "")
	if bound.Code != http.StatusOK {
		t.Fatalf("bind status=%d body=%s", bound.Code, bound.Body.String())
	}
	var binding agentBatchFactoryBindingResponse
	if err := json.Unmarshal(bound.Body.Bytes(), &binding); err != nil {
		t.Fatal(err)
	}
	synced := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/"+binding.ProjectID+"/batch-factory/books/"+book.ID+"/sync", "")
	if synced.Code != http.StatusOK {
		t.Fatalf("sync status=%d body=%s", synced.Code, synced.Body.String())
	}
	canvas, err := projects.Canvas(context.Background(), "alice", binding.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	var video agentworkspace.Node
	for _, node := range canvas.Nodes {
		if node.Type == "video" && node.Config["resourceType"] == "video" {
			video = node
			break
		}
	}
	if video.ID == "" || video.Status != "failed" || video.Config["canRetry"] != true || video.Config["errorMessage"] != "temporary provider failure" {
		t.Fatalf("video=%+v", video)
	}
	tasks, err := projects.ListGenerationTasks(context.Background(), "alice", binding.ProjectID)
	if err != nil || len(tasks) != 0 {
		t.Fatalf("agent tasks=%+v err=%v", tasks, err)
	}
}

func TestDeleteBoundNodeOnlyRemovesCanvasReference(t *testing.T) {
	now := time.Date(2026, 10, 23, 12, 10, 0, 0, time.UTC)
	base := batchfactoryv11.NewMemoryStore()
	batch, err := base.CreateBatch(context.Background(), "alice", batchfactoryv11.CreateBatchInput{
		Title: "batch", Books: []batchfactoryv11.CreateBookInput{{Title: "book", SourceText: "source"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	asset, err := base.CreateBookAsset(context.Background(), "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "character", Name: "hero", Prompt: "portrait"})
	if err != nil {
		t.Fatal(err)
	}
	batches := &batchFactoryDeleteGuard{MemoryStore: base}
	projects := agentworkspace.NewMemoryStore()
	mux := http.NewServeMux()
	RegisterAgentBatchFactoryBridge(mux, BridgeAuth{Secret: "secret", Now: func() time.Time { return now }}, projects, batches, nil, func() time.Time { return now })

	bound := agentRequest(t, mux, now, http.MethodPost, "/api/batch-factory/v11/batches/"+batch.ID+"/agent-project", "")
	if bound.Code != http.StatusOK {
		t.Fatalf("bind status=%d body=%s", bound.Code, bound.Body.String())
	}
	var binding agentBatchFactoryBindingResponse
	if err := json.Unmarshal(bound.Body.Bytes(), &binding); err != nil {
		t.Fatal(err)
	}
	synced := agentRequest(t, mux, now, http.MethodPost, "/api/agent/tasks/"+binding.ProjectID+"/batch-factory/books/"+book.ID+"/sync", "")
	if synced.Code != http.StatusOK {
		t.Fatalf("sync status=%d body=%s", synced.Code, synced.Body.String())
	}
	origins, err := projects.BatchFactoryNodeOrigins(context.Background(), "alice", binding.ProjectID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	var nodeID string
	for _, origin := range origins {
		if origin.ResourceID == asset.ID {
			nodeID = origin.NodeID
			break
		}
	}
	if nodeID == "" {
		t.Fatalf("asset origin missing: %+v", origins)
	}
	canvasBeforeDelete, err := projects.Canvas(context.Background(), "alice", binding.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	var scriptNodeID string
	for _, origin := range origins {
		if origin.ResourceType == "book_script" {
			scriptNodeID = origin.NodeID
			break
		}
	}
	if scriptNodeID == "" {
		t.Fatalf("script origin missing: %+v", origins)
	}
	canvasBeforeDelete.Edges = append(canvasBeforeDelete.Edges, agentworkspace.Edge{ID: "manual-reference", Source: scriptNodeID, Target: nodeID, Relation: "manual"})
	if _, err := projects.SaveCanvas(context.Background(), "alice", binding.ProjectID, canvasBeforeDelete, canvasBeforeDelete.Revision, now); err != nil {
		t.Fatal(err)
	}

	deleted := agentRequest(t, mux, now, http.MethodDelete, "/api/agent/tasks/"+binding.ProjectID+"/batch-factory/nodes/"+nodeID+"%20", "")
	if deleted.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", deleted.Code, deleted.Body.String())
	}
	if batches.deleteBookAssetCalls != 0 {
		t.Fatalf("DeleteBookAsset calls=%d, want 0", batches.deleteBookAssetCalls)
	}
	assets, err := base.ListBookAssets(context.Background(), "alice", batch.ID, book.ID)
	if err != nil || len(assets) != 1 || assets[0].ID != asset.ID {
		t.Fatalf("assets=%+v err=%v", assets, err)
	}
	canvas, err := projects.Canvas(context.Background(), "alice", binding.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	for _, node := range canvas.Nodes {
		if node.ID == nodeID {
			t.Fatalf("deleted node remains: %+v", node)
		}
	}
	for _, edge := range canvas.Edges {
		if edge.Source == nodeID || edge.Target == nodeID {
			t.Fatalf("deleted node edge remains: %+v", edge)
		}
	}
	origins, err = projects.BatchFactoryNodeOrigins(context.Background(), "alice", binding.ProjectID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, origin := range origins {
		if origin.NodeID == nodeID {
			t.Fatalf("deleted origin remains: %+v", origin)
		}
	}
}
