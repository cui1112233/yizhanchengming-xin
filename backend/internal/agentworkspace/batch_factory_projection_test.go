package agentworkspace

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

func TestSyncBatchFactoryBookDoesNotExposeMergeFailureDetails(t *testing.T) {
	ctx := context.Background()
	service, binding, book, _, _, production, merge := newProjectionFixture(t, ctx)
	source := service.BatchFactory.(*batchfactoryv11.MemoryStore)
	secretDetail := "provider failure: Authorization Bearer secret-token"
	failureTime := merge.UpdatedAt.Add(time.Minute)
	failed, err := source.CreateMergeJob(ctx, batchfactoryv11.MergeJob{
		Owner: "alice", BatchID: binding.BatchID, BookID: book.ID, RequestID: "merge-failed",
		Status: batchfactoryv11.MergeFailed, ErrorMessage: secretDetail,
		Sources:   []batchfactoryv11.MergeMedia{{ProductionJobID: production.ID, VideoID: book.Videos[0].ID, MediaURL: production.Tasks[0].MediaURL}},
		CreatedAt: failureTime, UpdatedAt: failureTime,
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.SyncBatchFactoryBook(ctx, "alice", binding, book.ID); err != nil {
		t.Fatal(err)
	}
	mergeNode := projectedNode(t, service.CanvasStore, ctx, binding, book.ID, "merge", failed.ID)
	visible, err := json.Marshal(struct {
		Summary  string         `json:"summary"`
		Config   map[string]any `json:"config"`
		AssetRef map[string]any `json:"assetRef"`
	}{Summary: mergeNode.Summary, Config: mergeNode.Config, AssetRef: mergeNode.AssetRef})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(visible), secretDetail) {
		t.Fatalf("merge failure detail leaked into canvas node: %s", visible)
	}
}

func TestProjectBookFingerprintsIgnoreUpstreamErrorDetails(t *testing.T) {
	ctx := context.Background()
	service, binding, book, _, _, _, _ := newProjectionFixture(t, ctx)
	source := service.BatchFactory.(*batchfactoryv11.MemoryStore)
	batch, err := source.GetBatch(ctx, "alice", binding.BatchID)
	if err != nil {
		t.Fatal(err)
	}
	book, found := projectionBook(batch, book.ID)
	if !found {
		t.Fatal("fixture book missing")
	}
	assets, err := source.ListBookAssets(ctx, "alice", binding.BatchID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	productionJobs, err := source.ListProductionJobs(ctx, "alice", binding.BatchID)
	if err != nil {
		t.Fatal(err)
	}
	mergeJobs, err := source.ListMergeJobs(ctx, "alice", binding.BatchID)
	if err != nil {
		t.Fatal(err)
	}
	base, _, err := service.projectBook(ctx, "alice", binding, book, assets, productionJobs, mergeJobs, EmptyCanvas(), nil)
	if err != nil {
		t.Fatal(err)
	}

	changedProduction := append([]batchfactoryv11.ProductionJob(nil), productionJobs...)
	changedProduction[0].Tasks = append([]batchfactoryv11.ProductionTask(nil), productionJobs[0].Tasks...)
	changedProduction[0].Tasks[0].ErrorMessage = "executor secret detail"
	changedMerges := append([]batchfactoryv11.MergeJob(nil), mergeJobs...)
	changedMerges[0].ErrorMessage = "provider secret detail"
	changed, _, err := service.projectBook(ctx, "alice", binding, book, assets, changedProduction, changedMerges, EmptyCanvas(), nil)
	if err != nil {
		t.Fatal(err)
	}

	baseVersions := make(map[string]string, len(base))
	for _, item := range base {
		baseVersions[projectionResourceKey(item.origin.ResourceType, item.origin.ResourceID)] = item.origin.VersionID
	}
	for _, item := range changed {
		key := projectionResourceKey(item.origin.ResourceType, item.origin.ResourceID)
		if item.origin.VersionID != baseVersions[key] {
			t.Fatalf("error detail changed %s fingerprint: base=%s changed=%s", key, baseVersions[key], item.origin.VersionID)
		}
	}
}

func TestSyncBatchFactoryBookLeavesCanvasAndOriginsUnchangedWhenProjectionSaveFails(t *testing.T) {
	ctx := context.Background()
	service, binding, book, _, _, _, _ := newProjectionFixture(t, ctx)
	store := service.CanvasStore.(*MemoryStore)
	beforeCanvas, err := store.Canvas(ctx, "alice", binding.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	beforeOrigins, err := store.BatchFactoryNodeOrigins(ctx, "alice", binding.ProjectID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	wantErr := errors.New("origin reconciliation failed")
	service.CanvasStore = &failingProjectionStore{MemoryStore: store, err: wantErr}

	if _, err := service.SyncBatchFactoryBook(ctx, "alice", binding, book.ID); !errors.Is(err, wantErr) {
		t.Fatalf("err=%v, want %v", err, wantErr)
	}
	afterCanvas, err := store.Canvas(ctx, "alice", binding.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	afterOrigins, err := store.BatchFactoryNodeOrigins(ctx, "alice", binding.ProjectID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(afterCanvas, beforeCanvas) || !reflect.DeepEqual(afterOrigins, beforeOrigins) {
		t.Fatalf("projection failure changed stored state: canvas before=%+v after=%+v origins before=%+v after=%+v", beforeCanvas, afterCanvas, beforeOrigins, afterOrigins)
	}
}

func TestMemoryStoreSaveBatchFactoryProjectionRejectsInvalidOriginWithoutMutation(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	now := time.Date(2026, 10, 23, 11, 0, 0, 0, time.UTC)
	project := Project{ID: "project-1", Title: "作品", Canvas: EmptyCanvas(), CreatedAt: now, UpdatedAt: now}
	if err := store.CreateProject(ctx, "alice", project); err != nil {
		t.Fatal(err)
	}
	binding, err := store.BindBatchFactoryProject(ctx, BatchFactoryBinding{Owner: "alice", BatchID: "batch-1", ProjectID: project.ID, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	existing := BatchFactoryNodeOrigin{ProjectID: project.ID, NodeID: "node-old", BatchID: binding.BatchID, BookID: "book-1", ResourceType: "book_script", ResourceID: "book-1", VersionID: "v1"}
	if err := store.BindBatchFactoryNode(ctx, "alice", existing); err != nil {
		t.Fatal(err)
	}
	beforeCanvas, err := store.Canvas(ctx, "alice", project.ID)
	if err != nil {
		t.Fatal(err)
	}
	beforeOrigins, err := store.BatchFactoryNodeOrigins(ctx, "alice", project.ID, "book-1")
	if err != nil {
		t.Fatal(err)
	}
	next := beforeCanvas
	next.Nodes = append(next.Nodes, Node{ID: "node-new", Type: "script", Title: "新剧本", Version: 1})
	desired := []BatchFactoryNodeOrigin{
		{ProjectID: project.ID, NodeID: "node-new", BatchID: binding.BatchID, BookID: "book-1", ResourceType: "book_script", ResourceID: "book-1", VersionID: "v2"},
		{ProjectID: project.ID, NodeID: "node-invalid", BatchID: binding.BatchID, BookID: "book-1", ResourceType: "book_asset"},
	}

	_, err = store.SaveBatchFactoryProjection(ctx, "alice", binding, "book-1", next, beforeCanvas.Revision, desired, now.Add(time.Minute))
	if !errors.Is(err, ErrBatchFactoryNodeOriginInvalid) {
		t.Fatalf("err=%v, want %v", err, ErrBatchFactoryNodeOriginInvalid)
	}
	afterCanvas, err := store.Canvas(ctx, "alice", project.ID)
	if err != nil {
		t.Fatal(err)
	}
	afterOrigins, err := store.BatchFactoryNodeOrigins(ctx, "alice", project.ID, "book-1")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(afterCanvas, beforeCanvas) || !reflect.DeepEqual(afterOrigins, beforeOrigins) {
		t.Fatalf("invalid origin changed stored state: canvas before=%+v after=%+v origins before=%+v after=%+v", beforeCanvas, afterCanvas, beforeOrigins, afterOrigins)
	}
}

func TestSyncBatchFactoryBookProjectsOnlyExistingNodeTypes(t *testing.T) {
	ctx := context.Background()
	service, binding, book, asset, image, production, merge := newProjectionFixture(t, ctx)

	canvas, err := service.SyncBatchFactoryBook(ctx, "alice", binding, book.ID)
	if err != nil {
		t.Fatal(err)
	}

	wantTypes := map[string]int{"script": 1, "image": 1, "audio": 1, "video": 2}
	gotTypes := map[string]int{}
	for _, node := range canvas.Nodes {
		gotTypes[node.Type]++
		if node.GroupID != "book:"+book.ID {
			t.Fatalf("node %s group=%q", node.ID, node.GroupID)
		}
		if node.Type == "asset" {
			t.Fatalf("invented asset node: %+v", node)
		}
	}
	if len(gotTypes) != len(wantTypes) {
		t.Fatalf("node types=%v, want %v", gotTypes, wantTypes)
	}
	for nodeType, count := range wantTypes {
		if gotTypes[nodeType] != count {
			t.Fatalf("node type %q count=%d, want %d", nodeType, gotTypes[nodeType], count)
		}
	}

	imageNode := projectedNode(t, service.CanvasStore, ctx, binding, book.ID, "book_asset", asset.ID)
	if imageNode.Config["prompt"] != asset.Prompt || imageNode.AssetRef["id"] != image.ID || imageNode.AssetRef["url"] != nil {
		t.Fatalf("image node=%+v", imageNode)
	}
	videoNode := projectedNode(t, service.CanvasStore, ctx, binding, book.ID, "video", book.Videos[0].ID)
	if videoNode.AssetRef["id"] != production.Tasks[0].ID || videoNode.AssetRef["url"] != nil {
		t.Fatalf("video node=%+v", videoNode)
	}
	if videoNode.Config["videoProvider"] != batchfactoryv11.VideoProviderDoubaoLocal || videoNode.Config["videoModelId"] != "local-doubao-executor-video" {
		t.Fatalf("video execution settings=%+v", videoNode.Config)
	}
	audioSeconds, audioOK := projectionConfigPositiveNumber(videoNode.Config["connectedAudioDurationSeconds"])
	resolvedSeconds, resolvedOK := projectionConfigPositiveNumber(videoNode.Config["resolvedDurationSeconds"])
	resourceRevision, revisionOK := projectionConfigPositiveNumber(videoNode.Config["resourceRevision"])
	if !audioOK || absDurationDelta(audioSeconds, 7.4) > 0.001 || !resolvedOK || absDurationDelta(resolvedSeconds, 5) > 0.001 || !revisionOK || resourceRevision < 1 {
		t.Fatalf("video timing config=%+v", videoNode.Config)
	}
	audioNode := projectedNode(t, service.CanvasStore, ctx, binding, book.ID, "audio", merge.AudioArtifactID)
	mergeNode := projectedNode(t, service.CanvasStore, ctx, binding, book.ID, "merge", merge.ID)
	scriptNode := projectedNode(t, service.CanvasStore, ctx, binding, book.ID, "book_script", book.ID)
	if mergeNode.Title != "最终合成视频" || mergeNode.AssetRef["url"] != nil {
		t.Fatalf("merge node=%+v", mergeNode)
	}
	assertProjectionEdge(t, canvas.Edges, scriptNode.ID, imageNode.ID)
	assertProjectionEdge(t, canvas.Edges, imageNode.ID, videoNode.ID)
	assertProjectionEdge(t, canvas.Edges, videoNode.ID, mergeNode.ID)
	assertProjectionEdge(t, canvas.Edges, audioNode.ID, mergeNode.ID)
}

func TestProjectionVideoAssetIDsUsesSavedSelectionOrBatchFactoryDefault(t *testing.T) {
	assets := []batchfactoryv11.BookAsset{{ID: "asset-1"}, {ID: "asset-2"}}
	video := batchfactoryv11.Video{}
	if got := projectionVideoAssetIDs(video, assets); !reflect.DeepEqual(got, []string{"asset-1", "asset-2"}) {
		t.Fatalf("default asset ids=%v", got)
	}
	video.SettingsState.Patch = batchfactoryv11.SettingsPatch{
		"assetSelection": json.RawMessage(`{"autoAssetIds":["asset-2"],"addedAssetIds":["asset-missing"],"excludedAssetIds":["asset-1"]}`),
	}
	if got := projectionVideoAssetIDs(video, assets); !reflect.DeepEqual(got, []string{"asset-2"}) {
		t.Fatalf("selected asset ids=%v", got)
	}
}

func projectionConfigPositiveNumber(value any) (float64, bool) {
	switch number := value.(type) {
	case float64:
		return number, number > 0
	case int64:
		return float64(number), number > 0
	case int:
		return float64(number), number > 0
	default:
		return 0, false
	}
}

func TestSyncBatchFactoryBookKeepsStableNodesPositionsAndManualContent(t *testing.T) {
	ctx := context.Background()
	service, binding, book, asset, _, _, _ := newProjectionFixture(t, ctx)
	first, err := service.SyncBatchFactoryBook(ctx, "alice", binding, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	imageNode := projectedNode(t, service.CanvasStore, ctx, binding, book.ID, "book_asset", asset.ID)
	for index := range first.Nodes {
		if first.Nodes[index].ID == imageNode.ID {
			first.Nodes[index].Position = Position{X: 777, Y: 333}
		}
	}
	manual := Node{ID: "manual-note", Type: "text", Title: "我的备注", Summary: "不要覆盖", Version: 4, Status: "draft", Position: Position{X: 9, Y: 11}}
	first.Nodes = append(first.Nodes, manual)
	manualEdge := Edge{ID: "manual-edge", Source: manual.ID, Target: imageNode.ID, Relation: "manual_reference"}
	first.Edges = append(first.Edges, manualEdge)
	if _, err := service.CanvasStore.SaveCanvas(ctx, "alice", binding.ProjectID, first, first.Revision, time.Date(2026, 10, 23, 10, 1, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}

	second, err := service.SyncBatchFactoryBook(ctx, "alice", binding, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Nodes) != len(first.Nodes) {
		t.Fatalf("node count after repeat=%d, want %d", len(second.Nodes), len(first.Nodes))
	}
	seen := map[string]int{}
	for _, node := range second.Nodes {
		seen[node.ID]++
		if node.ID == imageNode.ID && node.Position != (Position{X: 777, Y: 333}) {
			t.Fatalf("image position=%+v", node.Position)
		}
		if node.ID == manual.ID && (node.Summary != manual.Summary || node.Version != manual.Version || node.GroupID != "") {
			t.Fatalf("manual node changed: %+v", node)
		}
	}
	for id, count := range seen {
		if count != 1 {
			t.Fatalf("node %s appears %d times", id, count)
		}
	}
	if !hasEdge(second.Edges, manualEdge.Source, manualEdge.Target, manualEdge.Relation) {
		t.Fatalf("manual edge was removed: %+v", second.Edges)
	}
	projectionEdges := map[string]int{}
	for _, edge := range second.Edges {
		if edge.Relation == batchFactoryProjectionRelation {
			projectionEdges[edge.ID]++
		}
	}
	for id, count := range projectionEdges {
		if count != 1 {
			t.Fatalf("projection edge %s appears %d times", id, count)
		}
	}
}

func TestSyncBatchFactoryBookRecoversMissingOriginWithoutDuplicateNode(t *testing.T) {
	ctx := context.Background()
	service, binding, book, asset, _, _, _ := newProjectionFixture(t, ctx)
	if _, err := service.SyncBatchFactoryBook(ctx, "alice", binding, book.ID); err != nil {
		t.Fatal(err)
	}
	imageNode := projectedNode(t, service.CanvasStore, ctx, binding, book.ID, "book_asset", asset.ID)
	if err := service.CanvasStore.RemoveBatchFactoryNodeBinding(ctx, "alice", binding.ProjectID, imageNode.ID); err != nil {
		t.Fatal(err)
	}

	canvas, err := service.SyncBatchFactoryBook(ctx, "alice", binding, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, node := range canvas.Nodes {
		if node.ID == imageNode.ID {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("stable image node %s appears %d times", imageNode.ID, count)
	}
}

func TestSyncBatchFactoryBookDoesNotProjectMissingAudioRecord(t *testing.T) {
	ctx := context.Background()
	service, binding, book, _, _, _, merge := newProjectionFixture(t, ctx)
	service.AudioRecords = batchFactoryAudioRecords{}

	if _, err := service.SyncBatchFactoryBook(ctx, "alice", binding, book.ID); err != nil {
		t.Fatal(err)
	}
	origins, err := service.CanvasStore.BatchFactoryNodeOrigins(ctx, "alice", binding.ProjectID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, origin := range origins {
		if origin.ResourceType == "audio" || origin.ResourceID == merge.AudioArtifactID && origin.ResourceType != "merge" {
			t.Fatalf("missing audio record was projected: %+v", origin)
		}
	}
}

func TestSyncBatchFactoryBookBumpsVersionWhenPrimaryImageChanges(t *testing.T) {
	ctx := context.Background()
	service, binding, book, asset, _, _, _ := newProjectionFixture(t, ctx)
	if _, err := service.SyncBatchFactoryBook(ctx, "alice", binding, book.ID); err != nil {
		t.Fatal(err)
	}
	first := projectedNode(t, service.CanvasStore, ctx, binding, book.ID, "book_asset", asset.ID)
	source := service.BatchFactory.(*batchfactoryv11.MemoryStore)
	secondImage, err := source.CreateBookAssetImage(ctx, "alice", binding.BatchID, book.ID, asset.ID, batchfactoryv11.CreateBookAssetImageInput{URL: "https://cdn.example.com/lin-wan-2.png", MediaType: "image/png", Source: "provider"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := source.SetPrimaryBookAssetImage(ctx, "alice", binding.BatchID, book.ID, asset.ID, secondImage.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SyncBatchFactoryBook(ctx, "alice", binding, book.ID); err != nil {
		t.Fatal(err)
	}
	second := projectedNode(t, service.CanvasStore, ctx, binding, book.ID, "book_asset", asset.ID)
	if second.AssetRef["id"] != secondImage.ID || second.Version != first.Version+1 {
		t.Fatalf("first=%+v second=%+v", first, second)
	}
}

func TestSyncBatchFactoryBookRecoversAllMissingOriginsWithoutDuplicateEdges(t *testing.T) {
	ctx := context.Background()
	service, binding, book, _, _, _, _ := newProjectionFixture(t, ctx)
	if _, err := service.SyncBatchFactoryBook(ctx, "alice", binding, book.ID); err != nil {
		t.Fatal(err)
	}
	origins, err := service.CanvasStore.BatchFactoryNodeOrigins(ctx, "alice", binding.ProjectID, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, origin := range origins {
		if err := service.CanvasStore.RemoveBatchFactoryNodeBinding(ctx, "alice", binding.ProjectID, origin.NodeID); err != nil {
			t.Fatal(err)
		}
	}

	canvas, err := service.SyncBatchFactoryBook(ctx, "alice", binding, book.ID)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]int{}
	for _, edge := range canvas.Edges {
		if edge.Relation == batchFactoryProjectionRelation {
			seen[edge.ID]++
		}
	}
	for id, count := range seen {
		if count != 1 {
			t.Fatalf("projection edge %s appears %d times", id, count)
		}
	}
}

func newProjectionFixture(t *testing.T, ctx context.Context) (*BatchFactoryProjectionService, BatchFactoryBinding, batchfactoryv11.Book, batchfactoryv11.BookAsset, batchfactoryv11.BookAssetImage, batchfactoryv11.ProductionJob, batchfactoryv11.MergeJob) {
	t.Helper()
	now := time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)
	canvasStore := NewMemoryStore()
	if err := canvasStore.CreateProject(ctx, "alice", Project{ID: "project-1", Title: "作品", CreatedAt: now, UpdatedAt: now, Canvas: EmptyCanvas()}); err != nil {
		t.Fatal(err)
	}
	source := batchfactoryv11.NewMemoryStore()
	batch, err := source.CreateBatch(ctx, "alice", batchfactoryv11.CreateBatchInput{
		Title: "作品",
		InitialBatchSettings: batchfactoryv11.SettingsPatch{
			"videoProvider": json.RawMessage(`"doubao_local_executor"`),
			"videoModelId":  json.RawMessage(`"local-doubao-executor-video"`),
		},
		Books: []batchfactoryv11.CreateBookInput{{Title: "第一本书", SourceText: "第一章，雨夜。", Videos: []batchfactoryv11.CreateVideoInput{{Label: "VIDEO 01", VideoPrompt: "雨夜追逐", DurationSeconds: 5}}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	book := batch.Books[0]
	binding, err := canvasStore.BindBatchFactoryProject(ctx, BatchFactoryBinding{Owner: "alice", BatchID: batch.ID, ProjectID: "project-1", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	asset, err := source.CreateBookAsset(ctx, "alice", batch.ID, book.ID, batchfactoryv11.CreateBookAssetInput{Kind: "character", Name: "林晚", Prompt: "黑发女性，红色风衣", ImageAspectRatio: "9:16"})
	if err != nil {
		t.Fatal(err)
	}
	image, err := source.CreateBookAssetImage(ctx, "alice", batch.ID, book.ID, asset.ID, batchfactoryv11.CreateBookAssetImageInput{URL: "https://cdn.example.com/lin-wan.png", MediaType: "image/png", Source: "provider"})
	if err != nil {
		t.Fatal(err)
	}
	production, err := source.CreateProductionJob(ctx, batchfactoryv11.ProductionJob{Owner: "alice", BatchID: batch.ID, BookID: book.ID, RequestID: "production-1", Status: batchfactoryv11.ProductionSucceeded, Tasks: []batchfactoryv11.ProductionTask{{VideoID: book.Videos[0].ID, Status: batchfactoryv11.ProductionSucceeded, Attempt: 1, ReferenceImageURLs: []string{image.URL}, MediaURL: "https://cdn.example.com/video-1.mp4", CreatedAt: now, UpdatedAt: now}}, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	merge, err := source.CreateMergeJob(ctx, batchfactoryv11.MergeJob{Owner: "alice", BatchID: batch.ID, BookID: book.ID, RequestID: "merge-1", AudioSource: "tts", AudioArtifactID: "audio-1", Status: batchfactoryv11.MergeSucceeded, Sources: []batchfactoryv11.MergeMedia{{ProductionJobID: production.ID, VideoID: book.Videos[0].ID, MediaURL: production.Tasks[0].MediaURL, Order: 0}}, OutputURL: "https://cdn.example.com/final.mp4", CreatedAt: now, UpdatedAt: now})
	if err != nil {
		t.Fatal(err)
	}
	service := &BatchFactoryProjectionService{
		CanvasStore: canvasStore, BatchFactory: source, Now: func() time.Time { return now },
		AudioRecords: batchFactoryAudioRecords{merge.AudioArtifactID: {ID: merge.AudioArtifactID, VersionID: "audio-version-1", Status: "succeeded", DurationSeconds: 7.4, MediaURL: "/api/audio/audio-1"}},
	}
	return service, binding, book, asset, image, production, merge
}

type batchFactoryAudioRecords map[string]BatchFactoryAudioRecord

type failingProjectionStore struct {
	*MemoryStore
	err error
}

func (s *failingProjectionStore) SaveBatchFactoryProjection(context.Context, string, BatchFactoryBinding, string, Canvas, int64, []BatchFactoryNodeOrigin, time.Time) (Canvas, error) {
	return Canvas{}, s.err
}

func (s *failingProjectionStore) BindBatchFactoryNode(context.Context, string, BatchFactoryNodeOrigin) error {
	return s.err
}

func (records batchFactoryAudioRecords) BatchFactoryAudioRecord(_ context.Context, _, _, _, audioID string) (BatchFactoryAudioRecord, bool, error) {
	record, found := records[audioID]
	return record, found, nil
}

func projectedNode(t *testing.T, store Store, ctx context.Context, binding BatchFactoryBinding, bookID, resourceType, resourceID string) Node {
	t.Helper()
	origins, err := store.BatchFactoryNodeOrigins(ctx, binding.Owner, binding.ProjectID, bookID)
	if err != nil {
		t.Fatal(err)
	}
	canvas, err := store.Canvas(ctx, binding.Owner, binding.ProjectID)
	if err != nil {
		t.Fatal(err)
	}
	for _, origin := range origins {
		if origin.ResourceType != resourceType || origin.ResourceID != resourceID {
			continue
		}
		for _, node := range canvas.Nodes {
			if node.ID == origin.NodeID {
				return node
			}
		}
	}
	t.Fatalf("missing projected node %s/%s", resourceType, resourceID)
	return Node{}
}

func assertProjectionEdge(t *testing.T, edges []Edge, source, target string) {
	t.Helper()
	if !hasEdge(edges, source, target, batchFactoryProjectionRelation) {
		t.Fatalf("missing projection edge %s -> %s in %+v", source, target, edges)
	}
}

func hasEdge(edges []Edge, source, target, relation string) bool {
	for _, edge := range edges {
		if edge.Source == source && edge.Target == target && edge.Relation == relation {
			return true
		}
	}
	return false
}

func TestResolveVideoDurationUsesNearestAllowedAudioDuration(t *testing.T) {
	got := ResolveVideoDuration(7.4, []float64{5, 10, 15}, 0)
	if got.RequestedSeconds != 5 {
		t.Fatalf("requested=%v, want 5", got.RequestedSeconds)
	}
	if !got.Adjusted {
		t.Fatalf("resolution=%+v, want adjusted", got)
	}
}

func TestResolveVideoDurationUsesAbsoluteNearestAllowedValue(t *testing.T) {
	got := ResolveVideoDuration(11, []float64{5, 10, 15}, 0)
	if got.RequestedSeconds != 10 {
		t.Fatalf("requested=%v, want 10", got.RequestedSeconds)
	}
}

func TestResolveVideoDurationPreservesManualOverride(t *testing.T) {
	got := ResolveVideoDuration(7.4, []float64{5, 10, 15}, 15)
	if got.RequestedSeconds != 15 {
		t.Fatalf("requested=%v, want 15", got.RequestedSeconds)
	}
	if !got.HasMismatchWarning {
		t.Fatalf("resolution=%+v, want mismatch warning", got)
	}
}

func TestNormalizeMappedNodeStateUsesOnlyCanvasStatesAndSafeErrors(t *testing.T) {
	failed := NormalizeMappedNodeState("error", "Authorization Bearer secret-token")
	if failed.Status != "failed" || !failed.CanRetry || failed.ErrorMessage != "批量工厂任务失败，请重试" {
		t.Fatalf("failed=%+v", failed)
	}
	running := NormalizeMappedNodeState("awaiting_confirmation", "")
	if running.Status != "running" || running.CanRetry || running.ErrorMessage != "" {
		t.Fatalf("running=%+v", running)
	}
	if cancelled := NormalizeMappedNodeState("canceled", ""); cancelled.Status != "cancelled" || !cancelled.CanRetry {
		t.Fatalf("cancelled=%+v", cancelled)
	}
}
