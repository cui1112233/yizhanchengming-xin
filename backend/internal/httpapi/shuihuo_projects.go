package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
)

var shuihuoProductionConfigs sync.Map

// registerShuihuoProjectCompatRoutes keeps the legacy shared-work library
// contract available while the browser still reads its project cards through
// /api/platform-projects. The source of truth is the Go batch store.
func registerShuihuoProjectCompatRoutes(root *http.ServeMux, auth BridgeAuth, store batchfactoryv11.Store, files *localartifact.Store) {
	if store == nil {
		return
	}
	root.Handle("/api/shuihuo-production/health", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := bridgeOwner(r); !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ok": true, "database": map[string]any{"ready": true},
			"redis":   map[string]any{"ready": false, "reason": "本地 Go 迁移阶段暂未启用 Redis"},
			"storage": map[string]any{"ready": true}, "enabledModelKinds": []string{},
		})
	})))
	root.Handle("/api/shuihuo-production/models", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := bridgeOwner(r); !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"models": []any{}})
	})))
	root.Handle("/api/shuihuo-production/preset-slots", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := bridgeOwner(r); !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"presets": []any{}})
	})))
	root.Handle("/api/shuihuo-production/config", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if r.Method == http.MethodGet {
			if value, exists := shuihuoProductionConfigs.Load(owner); exists {
				writeJSON(w, http.StatusOK, value)
				return
			}
			writeJSON(w, http.StatusOK, shuihuoDefaultConfig())
			return
		}
		if r.Method == http.MethodPut {
			var input map[string]any
			if !decodeJSON(w, r, &input) {
				return
			}
			config := shuihuoDefaultConfig()
			for key, value := range input {
				config[key] = value
			}
			shuihuoProductionConfigs.Store(owner, config)
			writeJSON(w, http.StatusOK, config)
			return
		}
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	})))
	root.Handle("/api/shuihuo-production/projects", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		switch r.Method {
		case http.MethodGet:
			if id := strings.TrimSpace(r.URL.Query().Get("id")); id != "" {
				writeShuihuoProject(w, r, store, owner, id)
				return
			}
			batches, err := store.ListBatches(r.Context(), owner)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			projects := make([]map[string]any, 0, len(batches))
			for _, batch := range batches {
				projects = append(projects, shuihuoProject(batch))
			}
			writeJSON(w, http.StatusOK, map[string]any{"projects": projects})
		case http.MethodPost:
			var input struct {
				Name           string `json:"name"`
				ProductionMode string `json:"productionMode,omitempty"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			name := strings.TrimSpace(input.Name)
			if name == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
				return
			}
			batch, err := store.CreateBatch(r.Context(), owner, batchfactoryv11.CreateBatchInput{
				Title: name,
				InitialBatchSettings: batchfactoryv11.SettingsPatch{
					"shuihuoProject": json.RawMessage("true"),
				},
				// Keep one editable source book so the legacy source PUT can update
				// a newly-created project before any segmentation work starts.
				Books: []batchfactoryv11.CreateBookInput{{Title: name}},
			})
			if err != nil {
				writeStoreError(w, err)
				return
			}
			project := shuihuoProject(batch)
			if input.ProductionMode != "" {
				project["productionMode"] = input.ProductionMode
			}
			writeJSON(w, http.StatusCreated, project)
		case http.MethodDelete:
			id := strings.TrimSpace(r.URL.Query().Get("id"))
			if id == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id is required"})
				return
			}
			if err := store.DeleteBatch(r.Context(), owner, id); err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	})))

	root.Handle("/api/shuihuo-production/projects/", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/api/shuihuo-production/projects/")
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "project id is required"})
			return
		}
		batch, err := store.GetBatch(r.Context(), owner, parts[0])
		if err != nil {
			writeStoreError(w, err)
			return
		}
		if len(parts) == 1 && r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, shuihuoProjectDetail(batch))
			return
		}
		if len(parts) == 1 && r.Method == http.MethodDelete {
			if err := store.DeleteBatch(r.Context(), owner, batch.ID); err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return
		}
		if len(parts) == 2 && parts[1] == "segments" && r.Method == http.MethodPost {
			var input map[string]any
			if !decodeJSON(w, r, &input) {
				return
			}
			if len(batch.Books) == 0 {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "project has no editable source book"})
				return
			}
			text, _ := input["sourceText"].(string)
			if strings.TrimSpace(text) == "" {
				text, _ = input["text"].(string)
			}
			if strings.TrimSpace(text) == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "sourceText is required"})
				return
			}
			segments := storedShuihuoSegments(batch.Books[0])
			segment := normalizeShuihuoSegments([]map[string]any{input}, batch.Books[0].ID)
			segment[0]["id"] = fmt.Sprintf("segment_%s_%d", batch.Books[0].ID, len(segments)+1)
			segments = append(segments, segment[0])
			if err := saveShuihuoSegments(r, store, owner, batch, segments); err != nil {
				writeStoreError(w, err)
				return
			}
			updated, err := store.GetBatch(r.Context(), owner, batch.ID)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, shuihuoProjectDetail(updated))
			return
		}
		if len(parts) == 3 && parts[1] == "segments" && parts[2] == "order" && r.Method == http.MethodPut {
			var input struct {
				SegmentIDs []string `json:"segmentIds"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			segments := storedShuihuoSegments(batch.Books[0])
			ordered := reorderShuihuoSegments(segments, input.SegmentIDs)
			if err := saveShuihuoSegments(r, store, owner, batch, ordered); err != nil {
				writeStoreError(w, err)
				return
			}
			updated, err := store.GetBatch(r.Context(), owner, batch.ID)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, shuihuoProjectDetail(updated))
			return
		}
		if len(parts) == 2 && parts[1] == "voice-settings" && (r.Method == http.MethodGet || r.Method == http.MethodPut) {
			writeJSON(w, http.StatusOK, map[string]any{"settings": map[string]any{}})
			return
		}
		if len(parts) == 2 && parts[1] == "asset-generation-config" && (r.Method == http.MethodGet || r.Method == http.MethodPut) {
			writeJSON(w, http.StatusOK, map[string]any{})
			return
		}
		if len(parts) == 2 && parts[1] == "assets" && r.Method == http.MethodPost {
			var input struct {
				Category string `json:"category"`
				Kind     string `json:"kind"`
				Name     string `json:"name"`
				Prompt   string `json:"prompt"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			if len(batch.Books) == 0 || strings.TrimSpace(input.Name) == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "asset name is required"})
				return
			}
			kind := strings.TrimSpace(input.Kind)
			if kind == "" {
				kind = strings.TrimSpace(input.Category)
			}
			if kind == "" {
				kind = "character"
			}
			if _, err := store.CreateBookAsset(r.Context(), owner, batch.ID, batch.Books[0].ID, batchfactoryv11.CreateBookAssetInput{Kind: kind, Name: strings.TrimSpace(input.Name), Prompt: strings.TrimSpace(input.Prompt)}); err != nil {
				writeStoreError(w, err)
				return
			}
			updated, err := store.GetBatch(r.Context(), owner, batch.ID)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, shuihuoProjectDetail(updated))
			return
		}
		if len(parts) == 3 && parts[1] == "assets" && parts[2] == "image" && r.Method == http.MethodPost {
			if files == nil || len(batch.Books) == 0 {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "asset image storage is unavailable"})
				return
			}
			var input struct {
				Category string `json:"category"`
				Kind     string `json:"kind"`
				Name     string `json:"name"`
				Prompt   string `json:"prompt"`
				DataURL  string `json:"dataUrl"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			mediaType, payload, err := decodeImageDataURL(input.DataURL)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid image upload"})
				return
			}
			kind := strings.TrimSpace(input.Kind)
			if kind == "" {
				kind = strings.TrimSpace(input.Category)
			}
			if kind == "" {
				kind = "character"
			}
			asset, err := store.CreateBookAsset(r.Context(), owner, batch.ID, batch.Books[0].ID, batchfactoryv11.CreateBookAssetInput{Kind: kind, Name: strings.TrimSpace(input.Name), Prompt: strings.TrimSpace(input.Prompt)})
			if err != nil {
				writeStoreError(w, err)
				return
			}
			artifactID, err := localexecutor.NewArtifactID()
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "image id generation failed"})
				return
			}
			saved, err := files.SaveImage(artifactID, mediaType, bytes.NewReader(payload))
			if err != nil {
				writeBookAssetImageStorageError(w, err)
				return
			}
			if _, err := store.CreateBookAssetImage(r.Context(), owner, batch.ID, batch.Books[0].ID, asset.ID, batchfactoryv11.CreateBookAssetImageInput{StorageRef: saved.StorageRef, MediaType: saved.MediaType, Source: "upload"}); err != nil {
				_ = files.Remove(saved.StorageRef)
				writeStoreError(w, err)
				return
			}
			updated, err := store.GetBatch(r.Context(), owner, batch.ID)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, shuihuoProjectDetail(updated))
			return
		}
		if len(parts) == 2 && parts[1] == "segmentation" && r.Method == http.MethodPost {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "segmentation action is required"})
			return
		}
		if len(parts) == 2 && parts[1] == "segmentation" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "segmentation action is required"})
			return
		}
		if len(parts) == 3 && parts[1] == "segmentation" && (parts[2] == "paragraphs" || parts[2] == "fixed") && r.Method == http.MethodPost {
			candidates := paragraphCandidates(batch)
			writeJSON(w, http.StatusOK, map[string]any{"candidates": candidates})
			return
		}
		if len(parts) == 3 && parts[1] == "segmentation" && parts[2] == "confirm" && r.Method == http.MethodPost {
			var input struct {
				Candidates []map[string]any `json:"candidates"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			if len(batch.Books) == 0 || len(input.Candidates) == 0 {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "candidates are required"})
				return
			}
			segments := normalizeShuihuoSegments(input.Candidates, batch.Books[0].ID)
			metadata := cloneMetadata(batch.Books[0].SourceMetadata)
			metadata["shuihuoSegments"] = segments
			metadata["shuihuoSegmentationStatus"] = "confirmed"
			if _, err := store.UpdateBookMetadata(r.Context(), owner, batch.ID, batch.Books[0].ID, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata, ExpectedRevision: batch.Books[0].Revision}); err != nil {
				writeStoreError(w, err)
				return
			}
			updated, err := store.GetBatch(r.Context(), owner, batch.ID)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, shuihuoProjectDetail(updated))
			return
		}
		if len(parts) == 2 && parts[1] == "source" && (r.Method == http.MethodPut || r.Method == http.MethodPost) {
			var input struct {
				SourceText       string `json:"sourceText"`
				SegmentationMode string `json:"segmentationMode,omitempty"`
				SourceTitle      string `json:"sourceTitle,omitempty"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			if strings.TrimSpace(input.SourceText) == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "sourceText is required"})
				return
			}
			if len(batch.Books) == 0 {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "project has no editable source book"})
				return
			}
			book := batch.Books[0]
			updated, err := store.CaptureBookSource(r.Context(), owner, batch.ID, book.ID, batchfactoryv11.CaptureBookSourceInput{
				SourceText: strings.TrimSpace(input.SourceText), ReplaceSource: true, SourceTitle: input.SourceTitle, ExpectedRevision: book.Revision,
			})
			if err != nil {
				writeStoreError(w, err)
				return
			}
			batch.Books[0] = updated
			project := shuihuoProject(batch)
			project["segmentationMode"] = input.SegmentationMode
			writeJSON(w, http.StatusOK, shuihuoProjectDetailFromProject(project))
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "project route not found"})
	})))

	root.Handle("/api/shuihuo-production/segments/", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		segmentID := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/shuihuo-production/segments/"), "/")
		if segmentID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "segment id is required"})
			return
		}
		batches, err := store.ListBatches(r.Context(), owner)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		for _, batch := range batches {
			// ListBatches may return compact book records without source metadata.
			// Always hydrate the batch because editable segments live in that metadata.
			full, getErr := store.GetBatch(r.Context(), owner, batch.ID)
			if getErr != nil {
				continue
			}
			batch = full
			if len(batch.Books) == 0 {
				continue
			}
			segments := storedShuihuoSegments(batch.Books[0])
			index := -1
			for i, segment := range segments {
				if id, _ := segment["id"].(string); id == segmentID {
					index = i
					break
				}
			}
			if index < 0 {
				continue
			}
			switch r.Method {
			case http.MethodPut:
				var input map[string]any
				if !decodeJSON(w, r, &input) {
					return
				}
				for _, key := range []string{"sourceText", "subtitleText", "speaker", "imagePrompt", "videoPrompt"} {
					if value, exists := input[key]; exists {
						segments[index][key] = value
					}
				}
				if err := saveShuihuoSegments(r, store, owner, batch, segments); err != nil {
					writeStoreError(w, err)
					return
				}
			case http.MethodDelete:
				segments = append(segments[:index], segments[index+1:]...)
				if err := saveShuihuoSegments(r, store, owner, batch, segments); err != nil {
					writeStoreError(w, err)
					return
				}
			default:
				writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
				return
			}
			updated, err := store.GetBatch(r.Context(), owner, batch.ID)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, shuihuoProjectDetail(updated))
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "segment not found"})
	})))

	root.Handle("/api/shuihuo-production/assets/", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/shuihuo-production/assets/"), "/")
		parts := strings.Split(path, "/")
		if len(parts) == 0 || parts[0] == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "asset id is required"})
			return
		}
		assetID := parts[0]
		batches, err := store.ListBatches(r.Context(), owner)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		for _, compact := range batches {
			batch, getErr := store.GetBatch(r.Context(), owner, compact.ID)
			if getErr != nil {
				continue
			}
			if len(batch.Books) == 0 {
				continue
			}
			var asset *batchfactoryv11.BookAsset
			for index := range batch.Books[0].AssetRecords {
				if batch.Books[0].AssetRecords[index].ID == assetID {
					asset = &batch.Books[0].AssetRecords[index]
					break
				}
			}
			if asset == nil {
				continue
			}
			if len(parts) == 2 && parts[1] == "images" && r.Method == http.MethodGet {
				images, listErr := store.ListBookAssetImages(r.Context(), owner, batch.ID, batch.Books[0].ID, assetID)
				if listErr != nil {
					writeStoreError(w, listErr)
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"images": bookAssetImageResponses(images, batch.ID, batch.Books[0].ID, assetID)})
				return
			}
			if len(parts) == 1 && r.Method == http.MethodPut {
				var input struct {
					Name             string `json:"name"`
					Prompt           string `json:"prompt"`
					ExpectedRevision int64  `json:"expectedRevision"`
				}
				if !decodeJSON(w, r, &input) {
					return
				}
				if input.ExpectedRevision == 0 {
					input.ExpectedRevision = asset.Revision
				}
				updated, updateErr := store.UpdateBookAsset(r.Context(), owner, batch.ID, batch.Books[0].ID, assetID, batchfactoryv11.UpdateBookAssetInput{Name: input.Name, Prompt: input.Prompt, ExpectedRevision: input.ExpectedRevision})
				if updateErr != nil {
					writeStoreError(w, updateErr)
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"asset": updated})
				return
			}
			if len(parts) == 1 && r.Method == http.MethodDelete {
				if deleter, ok := store.(interface {
					DeleteBookAsset(context.Context, string, string, string, string) error
				}); ok {
					if deleteErr := deleter.DeleteBookAsset(r.Context(), owner, batch.ID, batch.Books[0].ID, assetID); deleteErr != nil {
						writeStoreError(w, deleteErr)
						return
					}
					writeJSON(w, http.StatusOK, map[string]any{"ok": true})
					return
				}
				writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "asset deletion is not enabled in this Go store"})
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "asset not found"})
	})))

	root.Handle("/api/shuihuo-production/asset-images/", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		imageID := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/shuihuo-production/asset-images/"), "/")
		if !strings.HasSuffix(imageID, "/primary") || r.Method != http.MethodPut {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "asset image route not found"})
			return
		}
		imageID = strings.TrimSuffix(imageID, "/primary")
		batches, err := store.ListBatches(r.Context(), owner)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		for _, compact := range batches {
			batch, getErr := store.GetBatch(r.Context(), owner, compact.ID)
			if getErr != nil || len(batch.Books) == 0 {
				continue
			}
			for _, asset := range batch.Books[0].AssetRecords {
				images, listErr := store.ListBookAssetImages(r.Context(), owner, batch.ID, batch.Books[0].ID, asset.ID)
				if listErr != nil {
					continue
				}
				for _, image := range images {
					if image.ID != imageID {
						continue
					}
					primary, primaryErr := store.SetPrimaryBookAssetImage(r.Context(), owner, batch.ID, batch.Books[0].ID, asset.ID, imageID)
					if primaryErr != nil {
						writeStoreError(w, primaryErr)
						return
					}
					writeJSON(w, http.StatusOK, map[string]any{"image": bookAssetImageResponse(primary, batch.ID, batch.Books[0].ID, asset.ID)})
					return
				}
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "asset image not found"})
	})))
}

func isShuihuoProjectBatch(batch batchfactoryv11.Batch) bool {
	value, ok := batch.SettingsState.Patch["shuihuoProject"]
	return ok && strings.TrimSpace(string(value)) == "true"
}

func shuihuoDefaultConfig() map[string]any {
	return map[string]any{
		"characterPrefix": "", "imagePrefix": "", "imageSuffix": "", "videoPrefix": "", "videoSuffix": "",
		"textModelId": nil, "imageModelId": nil, "videoModelId": nil, "audioModelId": nil, "jianyingDraftDirectory": "",
	}
}

func shuihuoProject(batch batchfactoryv11.Batch) map[string]any {
	createdAt := batch.CreatedAt.UTC().Format(time.RFC3339)
	updatedAt := batch.UpdatedAt.UTC().Format(time.RFC3339)
	status := "pending"
	var sourceText string
	segments := []map[string]any{}
	if len(batch.Books) > 0 {
		sourceText = batch.Books[0].SourceText
		segments = storedShuihuoSegments(batch.Books[0])
		if len(segments) > 0 {
			status = "confirmed"
		}
	}
	return map[string]any{
		"id": batch.ID, "name": batch.Title, "createdAt": createdAt, "updatedAt": updatedAt,
		"sourceText": sourceText, "segmentationStatus": status, "segments": segments, "assets": shuihuoAssets(batch),
		"books": batch.Books,
	}
}

func shuihuoAssets(batch batchfactoryv11.Batch) []map[string]any {
	if len(batch.Books) == 0 {
		return []map[string]any{}
	}
	assets := make([]map[string]any, 0, len(batch.Books[0].AssetRecords))
	for _, asset := range batch.Books[0].AssetRecords {
		assets = append(assets, map[string]any{
			"id": asset.ID, "category": asset.Kind, "kind": asset.Kind, "name": asset.Name, "prompt": asset.Prompt,
			"source": asset.Source, "revision": asset.Revision, "createdAt": asset.CreatedAt, "updatedAt": asset.UpdatedAt,
		})
	}
	return assets
}

// The studio page consumes a single read model rather than the compact card
// returned by the project library. Keep the envelope stable even while the
// Go migration is filling in the remaining segmentation/media stores.
func shuihuoProjectDetail(batch batchfactoryv11.Batch) map[string]any {
	return shuihuoProjectDetailFromProject(shuihuoProject(batch))
}

func shuihuoProjectDetailFromProject(project map[string]any) map[string]any {
	segments, _ := project["segments"].([]map[string]any)
	assets, _ := project["assets"].([]map[string]any)
	return map[string]any{
		"project":         project,
		"segments":        segments,
		"assets":          assets,
		"media":           []any{},
		"segmentAssetIDs": map[string][]string{},
		"settings":        map[string]any{},
	}
}

func cloneMetadata(input map[string]any) map[string]any {
	output := make(map[string]any, len(input)+2)
	for key, value := range input {
		output[key] = value
	}
	return output
}

func saveShuihuoSegments(r *http.Request, store batchfactoryv11.Store, owner string, batch batchfactoryv11.Batch, segments []map[string]any) error {
	if len(batch.Books) == 0 {
		return fmt.Errorf("project has no editable source book")
	}
	for index, segment := range segments {
		segment["orderIndex"] = index + 1
	}
	metadata := cloneMetadata(batch.Books[0].SourceMetadata)
	metadata["shuihuoSegments"] = segments
	metadata["shuihuoSegmentationStatus"] = "confirmed"
	_, err := store.UpdateBookMetadata(r.Context(), owner, batch.ID, batch.Books[0].ID, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata, ExpectedRevision: batch.Books[0].Revision})
	return err
}

func reorderShuihuoSegments(segments []map[string]any, ids []string) []map[string]any {
	byID := make(map[string]map[string]any, len(segments))
	for _, segment := range segments {
		if id, _ := segment["id"].(string); id != "" {
			byID[id] = segment
		}
	}
	ordered := make([]map[string]any, 0, len(segments))
	for _, id := range ids {
		if segment, ok := byID[id]; ok {
			ordered = append(ordered, segment)
			delete(byID, id)
		}
	}
	for _, segment := range segments {
		if id, _ := segment["id"].(string); id != "" {
			if _, ok := byID[id]; ok {
				ordered = append(ordered, segment)
				delete(byID, id)
			}
		}
	}
	return ordered
}

func storedShuihuoSegments(book batchfactoryv11.Book) []map[string]any {
	if value, ok := book.SourceMetadata["shuihuoSegments"].([]any); ok {
		segments := make([]map[string]any, 0, len(value))
		for _, item := range value {
			if segment, ok := item.(map[string]any); ok {
				segments = append(segments, segment)
			}
		}
		return segments
	}
	if value, ok := book.SourceMetadata["shuihuoSegments"].([]map[string]any); ok {
		return value
	}
	return []map[string]any{}
}

func paragraphCandidates(batch batchfactoryv11.Batch) []map[string]any {
	if len(batch.Books) == 0 {
		return []map[string]any{}
	}
	text := strings.TrimSpace(batch.Books[0].SourceText)
	if text == "" {
		return []map[string]any{}
	}
	lines := strings.FieldsFunc(text, func(r rune) bool { return r == '\n' || r == '\r' })
	if len(lines) == 1 {
		lines = splitShuihuoSentences(lines[0])
	}
	candidates := make([]map[string]any, 0, len(lines))
	for index, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		candidates = append(candidates, map[string]any{"id": fmt.Sprintf("candidate_%s_%d", batch.Books[0].ID, index+1), "text": line, "sourceText": line, "subtitleText": line, "speaker": "旁白"})
	}
	return candidates
}

func splitShuihuoSentences(text string) []string {
	var out []string
	start := 0
	for index, r := range text {
		if r == '。' || r == '！' || r == '？' || r == '!' || r == '?' || r == '；' || r == ';' {
			if line := strings.TrimSpace(text[start : index+len(string(r))]); line != "" {
				out = append(out, line)
			}
			start = index + len(string(r))
		}
	}
	if line := strings.TrimSpace(text[start:]); line != "" {
		out = append(out, line)
	}
	return out
}

func normalizeShuihuoSegments(candidates []map[string]any, bookID string) []map[string]any {
	segments := make([]map[string]any, 0, len(candidates))
	for index, candidate := range candidates {
		text, _ := candidate["text"].(string)
		if strings.TrimSpace(text) == "" {
			text, _ = candidate["sourceText"].(string)
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		id, _ := candidate["id"].(string)
		if strings.TrimSpace(id) == "" {
			id = fmt.Sprintf("segment_%s_%d", bookID, index+1)
		}
		segment := map[string]any{"id": id, "orderIndex": index + 1, "sourceText": text, "subtitleText": text, "speaker": "旁白", "confirmed": true, "imagePrompt": "", "videoPrompt": ""}
		for _, key := range []string{"subtitleText", "speaker", "imagePrompt", "videoPrompt"} {
			if value, ok := candidate[key].(string); ok && strings.TrimSpace(value) != "" {
				segment[key] = strings.TrimSpace(value)
			}
		}
		segments = append(segments, segment)
	}
	return segments
}

func writeShuihuoProject(w http.ResponseWriter, r *http.Request, store batchfactoryv11.Store, owner, id string) {
	batch, err := store.GetBatch(r.Context(), owner, id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, shuihuoProject(batch))
}
