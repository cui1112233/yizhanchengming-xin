package agentworkspace

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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
	// Comic pages and collections are projected as read-only canvas state so
	// the Agent canvas can show the same comic deliverable the workbench shows.
	ListComicPages(context.Context, string, string, string) ([]batchfactoryv11.ComicPage, error)
	ListComicCollections(context.Context, string, string, string) ([]batchfactoryv11.ComicCollection, error)
	// GetDraft reads the persisted N8 comic publish record. The projection
	// never writes drafts back.
	GetDraft(context.Context, string, string, string, string) (batchfactoryv11.Draft, error)
}

// n8ComicPublishDraftKind is the durable draft kind written by Batch Factory
// after an N8 comic collection publish. The projection reads it as status only.
const n8ComicPublishDraftKind = "n8-comic-publish"

// BatchFactoryPublishResult is the read-only outcome of an external publish
// intent (providers 121 and yadi). It carries no payload or credential data.
type BatchFactoryPublishResult struct {
	Provider    string
	Status      string
	SubmittedAt *time.Time
}

// BatchFactoryPublishResultReader lets the canvas see 121/YD publish outcomes
// without giving the canvas any write path into the publish system.
type BatchFactoryPublishResultReader interface {
	BatchFactoryPublishResults(context.Context, string, string, string) ([]BatchFactoryPublishResult, error)
}

// BatchFactoryMergeStatusReader refreshes in-flight merge jobs before a
// projection is rendered. Batch Factory owns this polling behavior so the
// canvas and workbench always show the same job state.
type BatchFactoryMergeStatusReader interface {
	GetBatchStatus(context.Context, string, string) ([]batchfactoryv11.MergeJob, error)
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

// DurationResolution records how a Batch Factory video duration was chosen.
// A connected audio track is a default, never a silent overwrite of a manual
// choice made by the user.
type DurationResolution struct {
	RequestedSeconds   float64
	Adjusted           bool
	HasMismatchWarning bool
}

// ResolveVideoDuration chooses the nearest duration accepted by a model when
// audio supplies the default target. A positive override remains authoritative
// and is surfaced as a warning when it no longer matches the audio.
func ResolveVideoDuration(audioSeconds float64, allowedSeconds []float64, overrideSeconds float64) DurationResolution {
	allowed := make([]float64, 0, len(allowedSeconds))
	for _, value := range allowedSeconds {
		if value > 0 {
			allowed = append(allowed, value)
		}
	}
	sort.Float64s(allowed)
	if overrideSeconds > 0 {
		return DurationResolution{RequestedSeconds: overrideSeconds, HasMismatchWarning: audioSeconds > 0 && absDurationDelta(overrideSeconds, audioSeconds) > 0.01}
	}
	if audioSeconds <= 0 || len(allowed) == 0 {
		return DurationResolution{RequestedSeconds: audioSeconds}
	}
	requested := allowed[0]
	for _, candidate := range allowed[1:] {
		if absDurationDelta(candidate, audioSeconds) < absDurationDelta(requested, audioSeconds) {
			requested = candidate
		}
	}
	return DurationResolution{RequestedSeconds: requested, Adjusted: absDurationDelta(requested, audioSeconds) > 0.01}
}

func absDurationDelta(left, right float64) float64 {
	if left < right {
		return right - left
	}
	return left - right
}

// NormalizedMappedNodeState keeps the Agent canvas truthful without exposing
// provider-specific status values or unsafe diagnostic payloads.
type NormalizedMappedNodeState struct {
	Status       string
	ErrorMessage string
	CanRetry     bool
}

func NormalizeMappedNodeState(status, errorMessage string) NormalizedMappedNodeState {
	normalized := strings.ToLower(strings.TrimSpace(status))
	switch normalized {
	case "succeeded", "success", "completed", "confirmed", "applied":
		normalized = "succeeded"
	case "failed", "error", "abandoned":
		normalized = "failed"
	case "cancelled", "canceled":
		normalized = "cancelled"
	case "running", "processing", "submitting", "preparing", "ingesting", "result_ready", "awaiting_confirmation":
		normalized = "running"
	default:
		normalized = "queued"
	}
	state := NormalizedMappedNodeState{Status: normalized, CanRetry: normalized == "failed" || normalized == "cancelled"}
	if normalized == "failed" {
		state.ErrorMessage = safeMappedNodeError(errorMessage)
	}
	return state
}

func safeMappedNodeError(value string) string {
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	if value == "" {
		return "批量工厂任务失败，请重试"
	}
	lower := strings.ToLower(value)
	for _, sensitive := range []string{"authorization", "bearer ", "token", "secret", "password", "http://", "https://"} {
		if strings.Contains(lower, sensitive) {
			return "批量工厂任务失败，请重试"
		}
	}
	if len(value) > 200 {
		return value[:200] + "..."
	}
	return value
}

type BatchFactoryProjectionStore interface {
	Store
	SaveBatchFactoryProjection(context.Context, string, BatchFactoryBinding, string, Canvas, int64, []BatchFactoryNodeOrigin, time.Time) (Canvas, error)
}

type BatchFactoryProjectionService struct {
	CanvasStore    BatchFactoryProjectionStore
	BatchFactory   BatchFactoryProjectionReader
	MergeStatus    BatchFactoryMergeStatusReader
	AudioRecords   BatchFactoryAudioReader
	PublishResults BatchFactoryPublishResultReader
	Now            func() time.Time
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
	var mergeJobs []batchfactoryv11.MergeJob
	if s.MergeStatus != nil {
		mergeJobs, err = s.MergeStatus.GetBatchStatus(ctx, owner, binding.BatchID)
	} else {
		mergeJobs, err = s.BatchFactory.ListMergeJobs(ctx, owner, binding.BatchID)
	}
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

	items, edges, err := s.projectBook(ctx, owner, binding, book, assets, productionJobs, mergeJobs, canvas, origins, batch.SettingsState.Patch)
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
	// Status polling asks Batch Factory for the latest task state. If that
	// read did not change this book's projection, keep the existing canvas
	// revision so an active canvas is not continuously saved or conflicted.
	if sameBatchFactoryProjectionCanvas(canvas, next) && sameBatchFactoryProjectionOrigins(origins, desiredOrigins) {
		return canvas, nil
	}
	return s.CanvasStore.SaveBatchFactoryProjection(ctx, owner, binding, book.ID, next, canvas.Revision, desiredOrigins, now)
}

func sameBatchFactoryProjectionCanvas(left, right Canvas) bool {
	// Projection config values can carry equivalent numeric values with
	// different Go representations after a store round trip. JSON is the
	// persisted canvas format and gives us the durable comparison here.
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && string(leftJSON) == string(rightJSON)
}

func sameBatchFactoryProjectionOrigins(left, right []BatchFactoryNodeOrigin) bool {
	if len(left) != len(right) {
		return false
	}
	byNodeID := make(map[string]BatchFactoryNodeOrigin, len(left))
	for _, origin := range left {
		byNodeID[origin.NodeID] = origin
	}
	for _, origin := range right {
		if current, found := byNodeID[origin.NodeID]; !found || current != origin {
			return false
		}
	}
	return true
}

func projectionBook(batch batchfactoryv11.Batch, bookID string) (batchfactoryv11.Book, bool) {
	for _, book := range batch.Books {
		if book.ID == bookID {
			return book, true
		}
	}
	return batchfactoryv11.Book{}, false
}

func (s *BatchFactoryProjectionService) projectBook(ctx context.Context, owner string, binding BatchFactoryBinding, book batchfactoryv11.Book, assets []batchfactoryv11.BookAsset, productionJobs []batchfactoryv11.ProductionJob, mergeJobs []batchfactoryv11.MergeJob, canvas Canvas, origins []BatchFactoryNodeOrigin, batchSettings ...batchfactoryv11.SettingsPatch) ([]batchFactoryProjectionItem, []Edge, error) {
	groupID := "book:" + book.ID
	existingNodes := make(map[string]Node, len(canvas.Nodes))
	for _, node := range canvas.Nodes {
		existingNodes[node.ID] = node
	}
	originByResource := projectionOriginsByResource(origins, existingNodes)
	items := make([]batchFactoryProjectionItem, 0, 1+len(assets)+len(book.Videos)+2)
	nodeIDByResource := map[string]string{}
	imageNodeByURL := map[string]string{}
	imageNodeByAssetID := map[string]string{}
	merge, hasMerge := projectionLatestMerge(book.ID, mergeJobs)
	audioSeconds := projectionJSONNumber(book.SettingsState.Patch["audioDurationSeconds"])
	audioID := projectionJSONText(book.SettingsState.Patch["ttsAudioArtifactId"])
	if hasMerge && strings.TrimSpace(merge.AudioArtifactID) != "" {
		// A completed merge is the authoritative record of the track it used.
		audioID = strings.TrimSpace(merge.AudioArtifactID)
	}
	var audioRecord BatchFactoryAudioRecord
	hasAudioRecord := false
	if audioID != "" && s.AudioRecords != nil {
		record, found, err := s.AudioRecords.BatchFactoryAudioRecord(ctx, owner, binding.BatchID, book.ID, audioID)
		if err != nil {
			return nil, nil, err
		}
		if found && strings.TrimSpace(record.ID) == audioID {
			audioRecord, hasAudioRecord = record, true
			if record.DurationSeconds > 0 {
				audioSeconds = record.DurationSeconds
			}
		}
	}

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

	scriptConfig := projectionConfig(binding.BatchID, book.ID, "book_script", book.ID)
	scriptConfig["sourceText"] = book.SourceText
	scriptConfig["contentRangeLines"] = projectionContentRangeLines(book.SourceMetadata)
	scriptConfig["resourceRevision"] = book.Revision
	effectiveScriptSettings := append([]batchfactoryv11.SettingsPatch(nil), batchSettings...)
	effectiveScriptSettings = append(effectiveScriptSettings, book.SettingsState.Patch)
	if textModelID := projectionJSONText(batchfactoryv11.ResolveSettings(effectiveScriptSettings...)["textModelId"]); textModelID != "" {
		scriptConfig["modelId"], scriptConfig["textModelId"] = textModelID, textModelID
	}
	applyMappedNodeState(scriptConfig, "confirmed", "")
	scriptNodeID := add("book_script", book.ID, projectionFingerprint(book.Revision, book.Title, book.SourceText), Node{
		Type: "script", Title: projectionTitle(book.Title, "剧本"), Summary: book.SourceText, Status: "succeeded",
		Position: Position{X: 40, Y: 80}, Config: scriptConfig,
	})
	sort.SliceStable(assets, func(i, j int) bool { return assets[i].ID < assets[j].ID })
	for index, asset := range assets {
		images, err := s.BatchFactory.ListBookAssetImages(ctx, owner, binding.BatchID, book.ID, asset.ID)
		if err != nil {
			return nil, nil, err
		}
		primary := projectionPrimaryImage(images)
		assetRef := map[string]any{"assetId": asset.ID, "kind": "image", "source": "batch_factory"}
		if primary.ID != "" {
			assetRef["id"] = primary.ID
			assetRef["mediaType"], assetRef["providerSource"] = primary.MediaType, primary.Source
		}
		config := projectionConfig(binding.BatchID, book.ID, "book_asset", asset.ID)
		config["prompt"], config["assetKind"] = asset.Prompt, asset.Kind
		config["assetName"], config["formalName"] = asset.Name, asset.FormalName
		config["facts"], config["appearance"], config["relationships"] = asset.Facts, asset.Appearance, asset.Relationships
		config["resourceRevision"] = asset.Revision
		if asset.ImageModelID != "" {
			config["modelId"], config["imageModelId"] = asset.ImageModelID, asset.ImageModelID
		}
		if asset.ImageAspectRatio != "" {
			config["ratio"], config["imageAspectRatio"] = asset.ImageAspectRatio, asset.ImageAspectRatio
		}
		imageStatus := "queued"
		if primary.ID != "" {
			imageStatus = "succeeded"
		}
		applyMappedNodeState(config, imageStatus, "")
		nodeID := add("book_asset", asset.ID, projectionFingerprint(asset.Revision, primary.ID, primary.Revision, primary.URL), Node{
			Type: "image", Title: projectionTitle(asset.Name, "图片素材"), Summary: asset.Prompt, Status: imageStatus,
			Position: Position{X: 360, Y: 80 + float64(index*180)}, AssetRef: assetRef, Config: config, SourceNodeIDs: []string{scriptNodeID},
		})
		imageNodeByAssetID[asset.ID] = nodeID
		for _, image := range images {
			if url := strings.TrimSpace(image.URL); url != "" {
				imageNodeByURL[url] = nodeID
			}
		}
	}

	tasksByVideo := projectionTasksByVideo(book, productionJobs)
	for index, video := range book.Videos {
		task, hasTask := projectionCurrentTask(video, tasksByVideo[video.ID])
		assetRef := map[string]any{"videoId": video.ID, "kind": "video", "source": "batch_factory"}
		versionID := projectionFingerprint(video.Revision, video.Label, video.VideoPrompt, video.VisualPrompt, video.DurationSeconds)
		status := "draft"
		if hasTask {
			assetRef["id"] = task.ID
			versionID = projectionFingerprint(video.Revision, task.ID, task.Status, task.MediaURL, task.UpdatedAt, task.ReferenceImageURLs)
			status = string(task.Status)
		}
		config := projectionConfig(binding.BatchID, book.ID, "video", video.ID)
		config["prompt"], config["visualPrompt"] = video.VideoPrompt, video.VisualPrompt
		layers := append([]batchfactoryv11.SettingsPatch(nil), batchSettings...)
		layers = append(layers, book.SettingsState.Patch, video.SettingsState.Patch)
		effective := batchfactoryv11.ResolveSettings(layers...)
		modelID := projectionJSONText(effective["videoModelId"])
		provider := batchfactoryv11.VideoProviderForModel(modelID, projectionJSONText(effective["videoProvider"]))
		if modelID != "" {
			config["modelId"], config["videoModelId"] = modelID, modelID
		}
		if provider != "" {
			config["provider"], config["videoProvider"] = provider, provider
		}
		config["resourceRevision"] = video.Revision
		allowedDurations := []float64{5, 10, 15}
		config["allowedDurationSeconds"] = allowedDurations
		if audioSeconds > 0 {
			config["connectedAudioDurationSeconds"] = audioSeconds
		}
		manualDuration := projectionJSONNumber(video.SettingsState.Patch["duration"])
		if manualDuration > 0 {
			config["durationOverrideSeconds"] = manualDuration
		}
		resolution := ResolveVideoDuration(audioSeconds, allowedDurations, manualDuration)
		if resolution.RequestedSeconds > 0 {
			config["resolvedDurationSeconds"] = resolution.RequestedSeconds
		}
		if video.DurationSeconds > 0 {
			config["durationSeconds"] = video.DurationSeconds
		}
		applyMappedNodeState(config, status, task.ErrorMessage)
		status = NormalizeMappedNodeState(status, task.ErrorMessage).Status
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
		for _, assetID := range projectionVideoAssetIDs(video, assets) {
			if sourceID := imageNodeByAssetID[assetID]; sourceID != "" {
				items[len(items)-1].node.SourceNodeIDs = appendUniqueString(items[len(items)-1].node.SourceNodeIDs, sourceID)
			}
		}
		nodeIDByResource[projectionResourceKey("video", video.ID)] = nodeID
	}

	// Comic pages and collections are read-only deliverables. They hang off the
	// video that produced them and feed the collection used by N8 publishing.
	comicPageNodeByID := map[string]string{}
	{
		pages, err := s.BatchFactory.ListComicPages(ctx, owner, binding.BatchID, book.ID)
		if err != nil {
			return nil, nil, err
		}
		videoOrder := make(map[string]int, len(book.Videos))
		for index, video := range book.Videos {
			videoOrder[video.ID] = index
		}
		sort.SliceStable(pages, func(i, j int) bool {
			left, leftOK := videoOrder[pages[i].VideoID]
			right, rightOK := videoOrder[pages[j].VideoID]
			if leftOK != rightOK {
				return leftOK
			}
			if leftOK && left != right {
				return left < right
			}
			if !pages[i].CreatedAt.Equal(pages[j].CreatedAt) {
				return pages[i].CreatedAt.Before(pages[j].CreatedAt)
			}
			return pages[i].ID < pages[j].ID
		})
		comicBaseY := 340 + float64(len(assets)*180)
		for index, page := range pages {
			primary := projectionPrimaryComicImage(page.Images)
			pageStatus := page.Status
			if pageStatus == "" {
				pageStatus = "draft"
			}
			if page.ConfigStale {
				pageStatus = "stale"
			}
			pageConfig := projectionConfig(binding.BatchID, book.ID, "comic_page", page.ID)
			pageConfig["videoId"] = page.VideoID
			pageConfig["comicPrompt"], pageConfig["imagePrompt"] = page.ComicPrompt, page.ImagePrompt
			pageConfig["resourceRevision"] = page.Revision
			pageConfig["configStale"] = page.ConfigStale
			if page.ModelID != "" {
				pageConfig["modelId"] = page.ModelID
			}
			if page.AspectRatio != "" {
				pageConfig["ratio"] = page.AspectRatio
			}
			if page.SnapshotHash != "" {
				pageConfig["snapshotHash"] = page.SnapshotHash
			}
			pageAssetRef := map[string]any{"kind": "comic_page", "source": "batch_factory", "comicPageId": page.ID}
			if primary.ID != "" {
				pageAssetRef["id"] = primary.ID
				pageAssetRef["mediaType"] = primary.MediaType
				if primary.URL != "" {
					pageAssetRef["url"] = primary.URL
				}
			}
			applyMappedNodeState(pageConfig, pageStatus, "")
			pageNodeID := add("comic_page", page.ID, projectionFingerprint(page.Revision, page.Status, page.ConfigStale, primary.ID, primary.Revision, primary.URL), Node{
				Type: "image", Title: fmt.Sprintf("漫画页 %02d", index+1), Summary: page.ComicPrompt,
				Status: NormalizeMappedNodeState(pageStatus, "").Status, Position: Position{X: 360, Y: comicBaseY + float64(index*180)},
				AssetRef: pageAssetRef, Config: pageConfig,
			})
			comicPageNodeByID[page.ID] = pageNodeID
			if sourceID := nodeIDByResource[projectionResourceKey("video", page.VideoID)]; sourceID != "" {
				items[len(items)-1].node.SourceNodeIDs = appendUniqueString(items[len(items)-1].node.SourceNodeIDs, sourceID)
			}
			for _, reference := range page.ReferenceImageURLs {
				if sourceID := imageNodeByURL[strings.TrimSpace(reference)]; sourceID != "" {
					items[len(items)-1].node.SourceNodeIDs = appendUniqueString(items[len(items)-1].node.SourceNodeIDs, sourceID)
				}
			}
		}

		collections, err := s.BatchFactory.ListComicCollections(ctx, owner, binding.BatchID, book.ID)
		if err != nil {
			return nil, nil, err
		}
		sort.SliceStable(collections, func(i, j int) bool {
			if !collections[i].CreatedAt.Equal(collections[j].CreatedAt) {
				return collections[i].CreatedAt.Before(collections[j].CreatedAt)
			}
			return collections[i].ID < collections[j].ID
		})
		for index, collection := range collections {
			draft, err := s.BatchFactory.GetDraft(ctx, owner, collection.ID, n8ComicPublishDraftKind, binding.BatchID)
			if err != nil && !errors.Is(err, batchfactoryv11.ErrNotFound) {
				return nil, nil, err
			}
			var record struct {
				Status      string    `json:"status"`
				PublishedAt time.Time `json:"publishedAt"`
				CarouselID  int64     `json:"carouselId"`
				WorkIDs     []int64   `json:"workIds"`
			}
			if err == nil && strings.TrimSpace(draft.Content) != "" {
				_ = json.Unmarshal([]byte(draft.Content), &record)
			}
			n8Status := strings.TrimSpace(record.Status)
			collectionConfig := projectionConfig(binding.BatchID, book.ID, "comic_collection", collection.ID)
			collectionConfig["comicStatus"] = collection.Status
			collectionConfig["pageCount"] = len(collection.Items)
			collectionConfig["n8PublishStatus"] = n8Status
			if record.CarouselID > 0 {
				collectionConfig["n8PublishCarouselId"] = record.CarouselID
			}
			if len(record.WorkIDs) > 0 {
				collectionConfig["n8PublishWorkIds"] = append([]int64(nil), record.WorkIDs...)
			}
			if !record.PublishedAt.IsZero() {
				collectionConfig["n8PublishedAt"] = record.PublishedAt.UTC().Format(time.RFC3339)
			}
			collectionStatus := "queued"
			switch n8Status {
			case "published":
				collectionStatus = "succeeded"
			case "uploaded":
				collectionStatus = "running"
			}
			applyMappedNodeState(collectionConfig, collectionStatus, "")
			collectionAssetRef := map[string]any{"kind": "comic_collection", "source": "batch_factory", "collectionId": collection.ID}
			if len(collection.Items) > 0 && collection.Items[0].URL != "" {
				collectionAssetRef["url"] = collection.Items[0].URL
			}
			collectionNodeID := add("comic_collection", collection.ID, projectionFingerprint(collection.Status, len(collection.Items), n8Status, record.CarouselID, record.PublishedAt), Node{
				Type: "image", Title: "漫画合集", Summary: fmt.Sprintf("%d 页", len(collection.Items)),
				Status: collectionStatus, Position: Position{X: 1040, Y: 420 + float64(index*200)},
				AssetRef: collectionAssetRef, Config: collectionConfig,
			})
			for _, item := range collection.Items {
				if sourceID := comicPageNodeByID[item.ComicPageID]; sourceID != "" {
					items[len(items)-1].node.SourceNodeIDs = appendUniqueString(items[len(items)-1].node.SourceNodeIDs, sourceID)
				}
			}
			nodeIDByResource[projectionResourceKey("comic_collection", collection.ID)] = collectionNodeID
		}
	}

	// The final video is a stable book-level node. It must be available before
	// the first merge so the canvas can finish the same chain as Batch Factory.
	mergeReady, mergeReason := projectionMergeReadiness(book.Videos, tasksByVideo)
	// Audio is a first-class production step, not an optional result of a
	// completed merge. Project a stable pending card before the first TTS run
	// so a newly bound book can complete the full Batch Factory chain on canvas.
	audioResourceID := audioID
	if audioResourceID == "" {
		audioResourceID = "audio-plan:" + book.ID
	}
	audioAssetRef := map[string]any{"kind": "audio", "source": "batch_factory"}
	config := projectionConfig(binding.BatchID, book.ID, "audio", audioResourceID)
	config["audioPlanPending"] = !hasAudioRecord
	versionID := projectionFingerprint(book.Revision, audioResourceID, "pending-audio-plan")
	status, summary := "queued", "生成配音并读取时长"
	if hasAudioRecord {
		record := audioRecord
		audioAssetRef["id"] = audioID
		if record.DurationSeconds > 0 {
			config["durationSeconds"] = record.DurationSeconds
		}
		versionID = strings.TrimSpace(record.VersionID)
		if versionID == "" {
			versionID = projectionFingerprint(record.ID, record.Status, record.MediaURL, record.DurationSeconds)
		}
		status = NormalizeMappedNodeState(record.Status, "").Status
		summary = "系统配音"
		if hasMerge && strings.TrimSpace(merge.AudioSource) != "" {
			summary = merge.AudioSource
		}
	}
	applyMappedNodeState(config, status, "")
	audioNodeID := add("audio", audioResourceID, versionID, Node{
		Type: "audio", Title: "配音", Summary: summary, Status: status, Position: Position{X: 360, Y: 120 + float64(len(assets)*180)},
		AssetRef: audioAssetRef, Config: config,
	})
	mergeConfig := projectionConfig(binding.BatchID, book.ID, "merge", book.ID)
	mergeConfig["bookRevision"] = book.Revision
	publishSettings := projectionJSONMap(batchfactoryv11.ResolveSettings(append(append([]batchfactoryv11.SettingsPatch{}, batchSettings...), book.SettingsState.Patch)...)["publishSettings"])
	mergeConfig["canMerge"] = mergeReady
	mergeConfig["publishStatus"] = projectionMetadataText(book.SourceMetadata, "websiteSubmitStatus", "publishStatus")
	if progress, ok := book.SourceMetadata["websiteSubmitProgress"].(map[string]any); ok {
		mergeConfig["publishProgress"] = progress
	}
	if organization := projectionMetadataText(publishSettings, "organization"); organization != "" {
		mergeConfig["publishOrganization"] = organization
	}
	// 121/YD publish outcomes come from the read-only external publish records.
	// They are never written back; the canvas only surfaces what already
	// happened.
	publish121Status, publishYadiStatus := "", ""
	if s.PublishResults != nil {
		results, err := s.PublishResults.BatchFactoryPublishResults(ctx, owner, binding.BatchID, book.ID)
		if err != nil {
			return nil, nil, err
		}
		for _, result := range results {
			switch strings.ToLower(strings.TrimSpace(result.Provider)) {
			case "121":
				publish121Status = result.Status
				mergeConfig["publish121Status"] = result.Status
				if result.SubmittedAt != nil {
					mergeConfig["publish121SubmittedAt"] = result.SubmittedAt.UTC().Format(time.RFC3339)
				}
			case "yadi":
				publishYadiStatus = result.Status
				mergeConfig["publishYadiStatus"] = result.Status
				if result.SubmittedAt != nil {
					mergeConfig["publishYadiSubmittedAt"] = result.SubmittedAt.UTC().Format(time.RFC3339)
				}
			}
		}
	}
	assetRef := map[string]any{"kind": "video", "source": "batch_factory"}
	mergeStatus, mergeVersion, mergeSummary := "queued", projectionFingerprint(book.Revision, "merge-pending", mergeReady, mergeReason, publish121Status, publishYadiStatus), mergeReason
	if hasMerge {
		assetRef["id"] = merge.ID
		mergeConfig["mergeJobId"] = merge.ID
		mergeConfig["canPublish"] = merge.Status == batchfactoryv11.MergeSucceeded && strings.TrimSpace(merge.OutputURL) != ""
		applyMappedNodeState(mergeConfig, string(merge.Status), merge.ErrorMessage)
		mergeStatus = NormalizeMappedNodeState(string(merge.Status), merge.ErrorMessage).Status
		mergeVersion = projectionFingerprint(merge.ID, merge.Status, merge.OutputURL, merge.UpdatedAt, merge.Sources, publish121Status, publishYadiStatus)
		mergeSummary = ""
	} else {
		mergeConfig["canPublish"] = false
		mergeConfig["mergeState"] = map[bool]string{true: "ready", false: "blocked"}[mergeReady]
		if mergeReason != "" {
			mergeConfig["mergeReason"] = mergeReason
		}
		applyMappedNodeState(mergeConfig, "queued", "")
	}
	mergeNodeID := add("merge", book.ID, mergeVersion, Node{
		Type: "video", Title: "最终合成视频", Summary: mergeSummary, Status: mergeStatus, Position: Position{X: 1040, Y: 80},
		AssetRef: assetRef, Config: mergeConfig,
	})
	if hasMerge {
		for _, source := range merge.Sources {
			if sourceID := nodeIDByResource[projectionResourceKey("video", source.VideoID)]; sourceID != "" {
				items[len(items)-1].node.SourceNodeIDs = appendUniqueString(items[len(items)-1].node.SourceNodeIDs, sourceID)
			}
		}
	} else {
		for _, video := range book.Videos {
			if sourceID := nodeIDByResource[projectionResourceKey("video", video.ID)]; sourceID != "" {
				items[len(items)-1].node.SourceNodeIDs = appendUniqueString(items[len(items)-1].node.SourceNodeIDs, sourceID)
			}
		}
	}
	if audioNodeID != "" {
		items[len(items)-1].node.SourceNodeIDs = appendUniqueString(items[len(items)-1].node.SourceNodeIDs, audioNodeID)
	}
	nodeIDByResource[projectionResourceKey("merge", book.ID)] = mergeNodeID

	edges := projectionEdges(binding.ProjectID, book.ID, items)
	return items, edges, nil
}

// projectionVideoAssetIDs mirrors Batch Factory's selection rule for the
// canvas only. A saved selection is authoritative; without one, Batch
// Factory's production compiler falls back to every available book asset.
// Keeping that fallback here makes the visible lines truthful before a video
// is first submitted, instead of showing unrelated isolated nodes.
func projectionVideoAssetIDs(video batchfactoryv11.Video, assets []batchfactoryv11.BookAsset) []string {
	type selection struct {
		AutoAssetIDs     []string `json:"autoAssetIds"`
		AddedAssetIDs    []string `json:"addedAssetIds"`
		ExcludedAssetIDs []string `json:"excludedAssetIds"`
	}

	raw, hasSelection := video.SettingsState.Patch["assetSelection"]
	if hasSelection && len(raw) > 0 {
		var saved selection
		if json.Unmarshal(raw, &saved) == nil {
			selected := map[string]bool{}
			for _, id := range append(saved.AutoAssetIDs, saved.AddedAssetIDs...) {
				if id = strings.TrimSpace(id); id != "" {
					selected[id] = true
				}
			}
			for _, id := range saved.ExcludedAssetIDs {
				delete(selected, strings.TrimSpace(id))
			}
			ids := make([]string, 0, len(selected))
			for _, asset := range assets {
				if selected[asset.ID] {
					ids = append(ids, asset.ID)
				}
			}
			return ids
		}
	}

	ids := make([]string, 0, len(assets))
	for _, asset := range assets {
		if id := strings.TrimSpace(asset.ID); id != "" {
			ids = append(ids, id)
		}
	}
	return ids
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
	projectionScopes := make(map[string]struct{}, len(items))
	for _, item := range items {
		desired[item.node.ID] = struct{}{}
		if scope := batchFactoryProjectionNodeScope(item.node); scope != "" {
			projectionScopes[scope] = struct{}{}
		}
	}
	staleMapped := make(map[string]struct{})
	for _, node := range canvas.Nodes {
		if _, keep := desired[node.ID]; keep {
			continue
		}
		if _, projected := projectionScopes[batchFactoryProjectionNodeScope(node)]; projected {
			staleMapped[node.ID] = struct{}{}
		}
	}
	authoritativeEdgePairs := make(map[string]struct{}, len(projectedEdges))
	for _, edge := range projectedEdges {
		authoritativeEdgePairs[edge.Source+"\x00"+edge.Target] = struct{}{}
	}
	nodes := make([]Node, 0, len(canvas.Nodes)+len(items))
	for _, node := range canvas.Nodes {
		if _, projected := oldProjected[node.ID]; projected {
			continue
		}
		if _, stale := staleMapped[node.ID]; stale {
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
	for nodeID := range staleMapped {
		removed[nodeID] = struct{}{}
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
		if _, authoritative := authoritativeEdgePairs[edge.Source+"\x00"+edge.Target]; authoritative &&
			(sourceProjected || targetProjected || sourceDesired || targetDesired) {
			// Batch Factory owns this exact source/target relationship. Keeping an
			// older manual relation here renders the same visual line twice.
			continue
		}
		edges = append(edges, edge)
	}
	edges = append(edges, projectedEdges...)
	canvas.Nodes, canvas.Edges = nodes, edges
	return canvas
}

// batchFactoryProjectionNodeScope identifies only nodes written by the
// projection. Keeping the batch and book together prevents a book refresh
// from touching manual nodes or another book's group in the same project.
func batchFactoryProjectionNodeScope(node Node) string {
	if strings.TrimSpace(fmt.Sprint(node.Config["source"])) != "batch_factory" {
		return ""
	}
	batchID := strings.TrimSpace(fmt.Sprint(node.Config["batchId"]))
	bookID := strings.TrimSpace(fmt.Sprint(node.Config["bookId"]))
	if batchID == "" || bookID == "" || batchID == "<nil>" || bookID == "<nil>" {
		return ""
	}
	return batchID + "\x00" + bookID
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

func projectionMergeReadiness(videos []batchfactoryv11.Video, tasksByVideo map[string][]batchfactoryv11.ProductionTask) (bool, string) {
	if len(videos) == 0 {
		return false, "请先生成至少一个视频分镜"
	}
	for _, video := range videos {
		task, found := projectionCurrentTask(video, tasksByVideo[video.ID])
		if !found || task.Status != batchfactoryv11.ProductionSucceeded || strings.TrimSpace(task.MediaURL) == "" {
			if task.Status == batchfactoryv11.ProductionFailed {
				return false, "有分镜视频生成失败，请重试后再合成"
			}
			return false, "等待所有分镜视频生成完成"
		}
	}
	return true, ""
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

func projectionPrimaryComicImage(images []batchfactoryv11.ComicPageImage) batchfactoryv11.ComicPageImage {
	for _, image := range images {
		if image.IsPrimary {
			return image
		}
	}
	if len(images) > 0 {
		return images[0]
	}
	return batchfactoryv11.ComicPageImage{}
}

func projectionConfig(batchID, bookID, resourceType, resourceID string) map[string]any {
	return map[string]any{"source": "batch_factory", "batchId": batchID, "bookId": bookID, "resourceType": resourceType, "resourceId": resourceID}
}

func projectionContentRangeLines(metadata map[string]any) int {
	value := 5
	switch raw := metadata["contentRangeLines"].(type) {
	case int:
		value = raw
	case int64:
		value = int(raw)
	case float64:
		value = int(raw)
	case json.Number:
		if parsed, err := raw.Int64(); err == nil {
			value = int(parsed)
		}
	}
	if value < 1 {
		return 1
	}
	if value > 500 {
		return 500
	}
	return value
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

func applyMappedNodeState(config map[string]any, status, errorMessage string) {
	state := NormalizeMappedNodeState(status, errorMessage)
	config["mappedState"] = state.Status
	config["canRetry"] = state.CanRetry
	if state.ErrorMessage != "" {
		config["errorMessage"] = state.ErrorMessage
	} else {
		delete(config, "errorMessage")
	}
}

func projectionJSONText(raw json.RawMessage) string {
	var value string
	if len(raw) == 0 || json.Unmarshal(raw, &value) != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

func projectionJSONMap(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var value map[string]any
	if json.Unmarshal(raw, &value) != nil || value == nil {
		return map[string]any{}
	}
	return value
}

func projectionMetadataText(metadata map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(fmt.Sprint(metadata[key])); value != "" && value != "<nil>" {
			return value
		}
	}
	return ""
}

func projectionJSONNumber(raw json.RawMessage) float64 {
	var value float64
	if len(raw) == 0 || json.Unmarshal(raw, &value) != nil || value <= 0 {
		return 0
	}
	return value
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
