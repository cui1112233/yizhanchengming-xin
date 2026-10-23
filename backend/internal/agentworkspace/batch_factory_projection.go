package agentworkspace

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

const batchFactoryProjectionRelation = "batch_factory_reference"

// BatchFactoryProjectionReader is the read-only part of Batch Factory used by
// the canvas projection. Both Batch Factory stores already satisfy this port.
type BatchFactoryProjectionReader interface {
	GetBatch(context.Context, string, string) (batchfactoryv11.Batch, error)
	ListBookAssets(context.Context, string, string, string) ([]batchfactoryv11.BookAsset, error)
	ListBookAssetImages(context.Context, string, string, string, string) ([]batchfactoryv11.BookAssetImage, error)
	ListProductionJobs(context.Context, string, string) ([]batchfactoryv11.ProductionJob, error)
	ListMergeJobs(context.Context, string, string) ([]batchfactoryv11.MergeJob, error)
}

type BatchFactoryAudioRecord struct {
	ID              string
	VersionID       string
	Status          string
	MediaURL        string
	DurationSeconds float64
}

type BatchFactoryAudioReader interface {
	BatchFactoryAudioRecord(context.Context, string, string, string, string) (BatchFactoryAudioRecord, bool, error)
}

type BatchFactoryProjectionStore interface {
	Store
	SaveBatchFactoryProjection(context.Context, string, BatchFactoryBinding, string, Canvas, int64, []BatchFactoryNodeOrigin, time.Time) (Canvas, error)
}

type BatchFactoryProjectionService struct {
	CanvasStore  BatchFactoryProjectionStore
	BatchFactory BatchFactoryProjectionReader
	AudioRecords BatchFactoryAudioReader
	Now          func() time.Time
}

type batchFactoryProjectionItem struct {
	node   Node
	origin BatchFactoryNodeOrigin
}

func (s *BatchFactoryProjectionService) SyncBatchFactoryBook(ctx context.Context, owner string, binding BatchFactoryBinding, bookID string) (Canvas, error) {
	owner, bookID = strings.TrimSpace(owner), strings.TrimSpace(bookID)
	if s == nil || s.CanvasStore == nil || s.BatchFactory == nil || owner == "" || bookID == "" {
		return Canvas{}, ErrBatchFactoryBindingInvalid
	}
	if err := ValidateBatchFactoryBinding(binding); err != nil || binding.Owner != owner {
		return Canvas{}, ErrBatchFactoryBindingInvalid
	}
	persisted, err := s.CanvasStore.BatchFactoryBinding(ctx, owner, binding.BatchID)
	if err != nil {
		return Canvas{}, err
	}
	if persisted.ProjectID != binding.ProjectID {
		return Canvas{}, ErrBatchFactoryBindingConflict
	}

	batch, err := s.BatchFactory.GetBatch(ctx, owner, binding.BatchID)
	if err != nil {
		return Canvas{}, err
	}
	book, found := projectionBook(batch, bookID)
	if !found {
		return Canvas{}, batchfactoryv11.ErrNotFound
	}
	assets, err := s.BatchFactory.ListBookAssets(ctx, owner, binding.BatchID, book.ID)
	if err != nil {
		return Canvas{}, err
	}
	productionJobs, err := s.BatchFactory.ListProductionJobs(ctx, owner, binding.BatchID)
	if err != nil {
		return Canvas{}, err
	}
	mergeJobs, err := s.BatchFactory.ListMergeJobs(ctx, owner, binding.BatchID)
	if err != nil {
		return Canvas{}, err
	}
	canvas, err := s.CanvasStore.Canvas(ctx, owner, binding.ProjectID)
	if err != nil {
		return Canvas{}, err
	}
	origins, err := s.CanvasStore.BatchFactoryNodeOrigins(ctx, owner, binding.ProjectID, book.ID)
	if err != nil {
		return Canvas{}, err
	}

	items, edges, err := s.projectBook(ctx, owner, binding, book, assets, productionJobs, mergeJobs, canvas, origins)
	if err != nil {
		return Canvas{}, err
	}
	next := reconcileBatchFactoryCanvas(canvas, origins, items, edges)
	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	desiredOrigins := make([]BatchFactoryNodeOrigin, 0, len(items))
	for _, item := range items {
		desiredOrigins = append(desiredOrigins, item.origin)
	}
	return s.CanvasStore.SaveBatchFactoryProjection(ctx, owner, binding, book.ID, next, canvas.Revision, desiredOrigins, now)
}

func projectionBook(batch batchfactoryv11.Batch, bookID string) (batchfactoryv11.Book, bool) {
	for _, book := range batch.Books {
		if book.ID == bookID {
			return book, true
		}
	}
	return batchfactoryv11.Book{}, false
}

func (s *BatchFactoryProjectionService) projectBook(ctx context.Context, owner string, binding BatchFactoryBinding, book batchfactoryv11.Book, assets []batchfactoryv11.BookAsset, productionJobs []batchfactoryv11.ProductionJob, mergeJobs []batchfactoryv11.MergeJob, canvas Canvas, origins []BatchFactoryNodeOrigin) ([]batchFactoryProjectionItem, []Edge, error) {
	groupID := "book:" + book.ID
	existingNodes := make(map[string]Node, len(canvas.Nodes))
	for _, node := range canvas.Nodes {
		existingNodes[node.ID] = node
	}
	originByResource := projectionOriginsByResource(origins, existingNodes)
	items := make([]batchFactoryProjectionItem, 0, 1+len(assets)+len(book.Videos)+2)
	nodeIDByResource := map[string]string{}
	imageNodeByURL := map[string]string{}

	add := func(resourceType, resourceID, versionID string, node Node) string {
		key := projectionResourceKey(resourceType, resourceID)
		origin, existed := originByResource[key]
		if !existed {
			origin = BatchFactoryNodeOrigin{ProjectID: binding.ProjectID, NodeID: projectionStableID(binding.BatchID, book.ID, resourceType, resourceID), BatchID: binding.BatchID, BookID: book.ID, ResourceType: resourceType, ResourceID: resourceID}
		}
		node.ID = origin.NodeID
		node.GroupID = groupID
		if previous, ok := existingNodes[node.ID]; ok {
			node.Position = previous.Position
			node.Version = previous.Version
			if !existed {
				node.Version++
			}
		}
		if node.Version < 1 {
			node.Version = 1
		} else if existed && origin.VersionID != "" && origin.VersionID != versionID {
			node.Version++
		}
		origin.VersionID = versionID
		items = append(items, batchFactoryProjectionItem{node: node, origin: origin})
		nodeIDByResource[key] = node.ID
		return node.ID
	}

	add("book_script", book.ID, projectionFingerprint(book.Revision, book.Title, book.SourceText), Node{
		Type: "script", Title: projectionTitle(book.Title, "剧本"), Summary: book.SourceText, Status: "confirmed",
		Position: Position{X: 40, Y: 80}, Config: projectionConfig(binding.BatchID, book.ID, "book_script", book.ID),
	})
	sort.SliceStable(assets, func(i, j int) bool { return assets[i].ID < assets[j].ID })
	for index, asset := range assets {
		images, err := s.BatchFactory.ListBookAssetImages(ctx, owner, binding.BatchID, book.ID, asset.ID)
		if err != nil {
			return nil, nil, err
		}
		primary := projectionPrimaryImage(images)
		assetRef := map[string]any{"assetId": asset.ID, "kind": "image"}
		if primary.ID != "" {
			assetRef["id"] = primary.ID
			assetRef["mediaType"], assetRef["source"] = primary.MediaType, primary.Source
		}
		config := projectionConfig(binding.BatchID, book.ID, "book_asset", asset.ID)
		config["prompt"], config["assetKind"] = asset.Prompt, asset.Kind
		if asset.ImageModelID != "" {
			config["modelId"] = asset.ImageModelID
		}
		if asset.ImageAspectRatio != "" {
			config["ratio"] = asset.ImageAspectRatio
		}
		nodeID := add("book_asset", asset.ID, projectionFingerprint(asset.Revision, primary.ID, primary.Revision, primary.URL), Node{
			Type: "image", Title: projectionTitle(asset.Name, "图片素材"), Summary: asset.Prompt, Status: projectionMediaStatus(primary.ID != ""),
			Position: Position{X: 360, Y: 80 + float64(index*180)}, AssetRef: assetRef, Config: config,
		})
		for _, image := range images {
			if url := strings.TrimSpace(image.URL); url != "" {
				imageNodeByURL[url] = nodeID
			}
		}
	}

	tasksByVideo := projectionTasksByVideo(book, productionJobs)
	for index, video := range book.Videos {
		task, hasTask := projectionCurrentTask(video, tasksByVideo[video.ID])
		assetRef := map[string]any{"videoId": video.ID, "kind": "video"}
		versionID := projectionFingerprint(video.Revision, video.Label, video.VideoPrompt, video.VisualPrompt, video.DurationSeconds)
		status := "draft"
		if hasTask {
			assetRef["id"] = task.ID
			versionID = projectionFingerprint(video.Revision, task.ID, task.Status, task.MediaURL, task.UpdatedAt, task.ReferenceImageURLs)
			status = string(task.Status)
		}
		config := projectionConfig(binding.BatchID, book.ID, "video", video.ID)
		config["prompt"], config["visualPrompt"] = video.VideoPrompt, video.VisualPrompt
		if video.DurationSeconds > 0 {
			config["durationSeconds"] = video.DurationSeconds
		}
		nodeID := add("video", video.ID, versionID, Node{
			Type: "video", Title: projectionTitle(video.Label, fmt.Sprintf("VIDEO %02d", index+1)), Summary: video.VideoPrompt, Status: status,
			Position: Position{X: 700, Y: 80 + float64(index*200)}, AssetRef: assetRef, Config: config,
		})
		if hasTask {
			for _, imageURL := range task.ReferenceImageURLs {
				if sourceID := imageNodeByURL[strings.TrimSpace(imageURL)]; sourceID != "" {
					items[len(items)-1].node.SourceNodeIDs = appendUniqueString(items[len(items)-1].node.SourceNodeIDs, sourceID)
				}
			}
		}
		nodeIDByResource[projectionResourceKey("video", video.ID)] = nodeID
	}

	if merge, ok := projectionLatestMerge(book.ID, mergeJobs); ok {
		var audioNodeID string
		if audioID := strings.TrimSpace(merge.AudioArtifactID); audioID != "" && s.AudioRecords != nil {
			record, found, err := s.AudioRecords.BatchFactoryAudioRecord(ctx, owner, binding.BatchID, book.ID, audioID)
			if err != nil {
				return nil, nil, err
			}
			if found && strings.TrimSpace(record.ID) == audioID {
				assetRef := map[string]any{"id": audioID, "kind": "audio"}
				config := projectionConfig(binding.BatchID, book.ID, "audio", audioID)
				if record.DurationSeconds > 0 {
					config["durationSeconds"] = record.DurationSeconds
				}
				versionID := strings.TrimSpace(record.VersionID)
				if versionID == "" {
					versionID = projectionFingerprint(record.ID, record.Status, record.MediaURL, record.DurationSeconds)
				}
				audioNodeID = add("audio", audioID, versionID, Node{
					Type: "audio", Title: "配音", Summary: merge.AudioSource, Status: projectionTitle(record.Status, "draft"), Position: Position{X: 360, Y: 120 + float64(len(assets)*180)},
					AssetRef: assetRef, Config: config,
				})
			}
		}
		assetRef := map[string]any{"id": merge.ID, "kind": "video"}
		mergeNodeID := add("merge", merge.ID, projectionFingerprint(merge.ID, merge.Status, merge.OutputURL, merge.UpdatedAt, merge.Sources), Node{
			Type: "video", Title: "最终合成视频", Status: string(merge.Status), Position: Position{X: 1040, Y: 80},
			AssetRef: assetRef, Config: projectionConfig(binding.BatchID, book.ID, "merge", merge.ID),
		})
		for _, source := range merge.Sources {
			if sourceID := nodeIDByResource[projectionResourceKey("video", source.VideoID)]; sourceID != "" {
				items[len(items)-1].node.SourceNodeIDs = appendUniqueString(items[len(items)-1].node.SourceNodeIDs, sourceID)
			}
		}
		if audioNodeID != "" {
			items[len(items)-1].node.SourceNodeIDs = appendUniqueString(items[len(items)-1].node.SourceNodeIDs, audioNodeID)
		}
		nodeIDByResource[projectionResourceKey("merge", merge.ID)] = mergeNodeID
	}

	edges := projectionEdges(binding.ProjectID, book.ID, items)
	return items, edges, nil
}

func projectionOriginsByResource(origins []BatchFactoryNodeOrigin, nodes map[string]Node) map[string]BatchFactoryNodeOrigin {
	sort.SliceStable(origins, func(i, j int) bool {
		_, leftExists := nodes[origins[i].NodeID]
		_, rightExists := nodes[origins[j].NodeID]
		if leftExists != rightExists {
			return leftExists
		}
		return origins[i].NodeID < origins[j].NodeID
	})
	out := make(map[string]BatchFactoryNodeOrigin, len(origins))
	for _, origin := range origins {
		key := projectionResourceKey(origin.ResourceType, origin.ResourceID)
		if _, exists := out[key]; !exists {
			out[key] = origin
		}
	}
	return out
}

func reconcileBatchFactoryCanvas(canvas Canvas, origins []BatchFactoryNodeOrigin, items []batchFactoryProjectionItem, projectedEdges []Edge) Canvas {
	oldProjected := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		oldProjected[origin.NodeID] = struct{}{}
	}
	desired := make(map[string]struct{}, len(items))
	for _, item := range items {
		desired[item.node.ID] = struct{}{}
	}
	nodes := make([]Node, 0, len(canvas.Nodes)+len(items))
	for _, node := range canvas.Nodes {
		if _, projected := oldProjected[node.ID]; projected {
			continue
		}
		if _, projected := desired[node.ID]; projected {
			continue
		}
		nodes = append(nodes, node)
	}
	for _, item := range items {
		node := item.node
		sort.Strings(node.SourceNodeIDs)
		nodes = append(nodes, node)
	}
	removed := make(map[string]struct{})
	for nodeID := range oldProjected {
		if _, keep := desired[nodeID]; !keep {
			removed[nodeID] = struct{}{}
		}
	}
	edges := make([]Edge, 0, len(canvas.Edges)+len(projectedEdges))
	for _, edge := range canvas.Edges {
		if _, sourceRemoved := removed[edge.Source]; sourceRemoved {
			continue
		}
		if _, targetRemoved := removed[edge.Target]; targetRemoved {
			continue
		}
		_, sourceProjected := oldProjected[edge.Source]
		_, targetProjected := oldProjected[edge.Target]
		_, sourceDesired := desired[edge.Source]
		_, targetDesired := desired[edge.Target]
		if edge.Relation == batchFactoryProjectionRelation && (sourceProjected || targetProjected || sourceDesired || targetDesired) {
			continue
		}
		edges = append(edges, edge)
	}
	edges = append(edges, projectedEdges...)
	canvas.Nodes, canvas.Edges = nodes, edges
	return canvas
}

func projectionEdges(projectID, bookID string, items []batchFactoryProjectionItem) []Edge {
	edges := make([]Edge, 0)
	seen := map[string]struct{}{}
	for _, item := range items {
		for _, sourceID := range item.node.SourceNodeIDs {
			if sourceID == "" || sourceID == item.node.ID {
				continue
			}
			key := sourceID + "\x00" + item.node.ID
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			edges = append(edges, Edge{ID: projectionStableID(projectID, bookID, sourceID, item.node.ID), Source: sourceID, Target: item.node.ID, Relation: batchFactoryProjectionRelation})
		}
	}
	sort.Slice(edges, func(i, j int) bool { return edges[i].ID < edges[j].ID })
	return edges
}

func projectionTasksByVideo(book batchfactoryv11.Book, jobs []batchfactoryv11.ProductionJob) map[string][]batchfactoryv11.ProductionTask {
	out := map[string][]batchfactoryv11.ProductionTask{}
	for _, job := range jobs {
		if job.BookID != book.ID {
			continue
		}
		if book.DirectorRevision != nil && job.DirectorRevisionID != book.DirectorRevision.ID {
			continue
		}
		for _, task := range job.Tasks {
			if task.OpeningVariantIndex == 0 {
				out[task.VideoID] = append(out[task.VideoID], task)
			}
		}
	}
	return out
}

func projectionCurrentTask(video batchfactoryv11.Video, tasks []batchfactoryv11.ProductionTask) (batchfactoryv11.ProductionTask, bool) {
	primaryID := projectionJSONText(video.SettingsState.Patch["primaryMediaTaskId"])
	if primaryID != "" {
		for _, task := range tasks {
			if task.ID == primaryID && task.Status == batchfactoryv11.ProductionSucceeded && strings.TrimSpace(task.MediaURL) != "" {
				return task, true
			}
		}
	}
	for index := len(tasks) - 1; index >= 0; index-- {
		if tasks[index].Status == batchfactoryv11.ProductionSucceeded && strings.TrimSpace(tasks[index].MediaURL) != "" {
			return tasks[index], true
		}
	}
	if len(tasks) > 0 {
		return tasks[len(tasks)-1], true
	}
	return batchfactoryv11.ProductionTask{}, false
}

func projectionLatestMerge(bookID string, jobs []batchfactoryv11.MergeJob) (batchfactoryv11.MergeJob, bool) {
	var latest batchfactoryv11.MergeJob
	found := false
	for _, job := range jobs {
		if job.BookID != bookID {
			continue
		}
		if !found || latest.UpdatedAt.Before(job.UpdatedAt) || (latest.UpdatedAt.Equal(job.UpdatedAt) && (latest.CreatedAt.Before(job.CreatedAt) || (latest.CreatedAt.Equal(job.CreatedAt) && latest.ID < job.ID))) {
			latest, found = job, true
		}
	}
	return latest, found
}

func projectionPrimaryImage(images []batchfactoryv11.BookAssetImage) batchfactoryv11.BookAssetImage {
	for _, image := range images {
		if image.IsPrimary {
			return image
		}
	}
	return batchfactoryv11.BookAssetImage{}
}

func projectionConfig(batchID, bookID, resourceType, resourceID string) map[string]any {
	return map[string]any{"source": "batch_factory", "batchId": batchID, "bookId": bookID, "resourceType": resourceType, "resourceId": resourceID}
}

func projectionTitle(value, fallback string) string {
	if value = strings.TrimSpace(value); value != "" {
		return value
	}
	return fallback
}

func projectionMediaStatus(hasMedia bool) string {
	if hasMedia {
		return "confirmed"
	}
	return "draft"
}

func projectionJSONText(raw json.RawMessage) string {
	var value string
	if len(raw) == 0 || json.Unmarshal(raw, &value) != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

func projectionResourceKey(resourceType, resourceID string) string {
	return strings.TrimSpace(resourceType) + "\x00" + strings.TrimSpace(resourceID)
}

func projectionStableID(parts ...string) string {
	hash := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return "bf-" + hex.EncodeToString(hash[:12])
}

func projectionFingerprint(values ...any) string {
	raw, _ := json.Marshal(values)
	hash := sha256.Sum256(raw)
	return hex.EncodeToString(hash[:16])
}

func appendUniqueString(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

var (
	_ BatchFactoryProjectionReader = (*batchfactoryv11.MemoryStore)(nil)
	_ BatchFactoryProjectionReader = (*batchfactoryv11.MySQLStore)(nil)
	_ BatchFactoryProjectionStore  = (*MemoryStore)(nil)
	_ BatchFactoryProjectionStore  = (*MySQLStore)(nil)
)
