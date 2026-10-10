package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
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
func registerShuihuoProjectCompatRoutes(root *http.ServeMux, auth BridgeAuth, store batchfactoryv11.Store, files *localartifact.Store, director *batchfactoryv11.DirectorService, production *batchfactoryv11.ProductionService) {
	if store == nil {
		return
	}
	root.Handle("/api/shuihuo-production/health", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := bridgeOwner(r); !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		enabledKinds := []string{}
		if director != nil && director.Provider != nil {
			enabledKinds = append(enabledKinds, "text")
		}
		if production != nil && production.Enabled {
			enabledKinds = append(enabledKinds, "video")
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ok": true, "database": map[string]any{"ready": true},
			"redis":   map[string]any{"ready": false, "reason": "本地 Go 迁移阶段暂未启用 Redis"},
			"storage": map[string]any{"ready": true}, "enabledModelKinds": enabledKinds,
		})
	})))
	root.Handle("/api/shuihuo-production/models", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		models := []any{}
		if durable, durableOK := store.(batchfactoryv11.ShuihuoConfigStore); durableOK {
			if config, err := durable.GetShuihuoConfig(r.Context(), owner); err == nil {
				if catalog, catalogOK := config["modelCatalog"].([]any); catalogOK {
					for _, raw := range catalog {
						if model, modelOK := raw.(map[string]any); modelOK {
							// Never return credentials from the account-level catalog.
							copy := map[string]any{}
							for _, key := range []string{"id", "name", "displayName", "kind", "provider", "adapterKind", "modelId", "enabled", "configured"} {
								if value, exists := model[key]; exists {
									copy[key] = value
								}
							}
							models = append(models, copy)
						}
					}
				}
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"models": models})
	})))
	root.Handle("/api/shuihuo-production/preset-slots", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := bridgeOwner(r); !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"presets": []any{}})
	})))
	root.Handle("/api/shuihuo-production/admin/models", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := BridgeIdentityFromContext(r.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if !identity.IsOwner {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "无权管理模型"})
			return
		}
		config, err := loadShuihuoOwnerConfig(r.Context(), store, identity.Username)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		models := shuihuoCollection(config, "modelCatalog", nil)
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, map[string]any{"models": safeShuihuoModels(models, false)})
			return
		}
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var input map[string]any
		if !decodeJSON(w, r, &input) {
			return
		}
		modelID, _ := input["modelId"].(string)
		name, _ := input["name"].(string)
		kind, _ := input["kind"].(string)
		if strings.TrimSpace(modelID) == "" || strings.TrimSpace(name) == "" || strings.TrimSpace(kind) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "modelId, name and kind are required"})
			return
		}
		for _, model := range models {
			if model["id"] == modelID || model["modelId"] == modelID {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "模型 ID 已存在"})
				return
			}
		}
		model := map[string]any{}
		for _, key := range []string{"modelId", "name", "kind", "adapterKind", "enabled", "parameterSchema", "credentialRef", "endpoint", "tasksEndpoint", "resultEndpoint", "workflowId", "requestTemplate", "responseMapping", "credential", "apiKey"} {
			if value, exists := input[key]; exists {
				model[key] = value
			}
		}
		model["id"] = modelID
		models = append(models, model)
		config["modelCatalog"] = models
		if err := saveShuihuoOwnerConfig(r.Context(), store, identity.Username, config); err != nil {
			writeStoreError(w, err)
			return
		}
		if production != nil && production.ProviderRegistry != nil && strings.EqualFold(strings.TrimSpace(kind), "video") {
			if syncErr := syncShuihuoVideoModel(r.Context(), production.ProviderRegistry, identity.Username, model); syncErr != nil {
				writeStoreError(w, syncErr)
				return
			}
		}
		writeJSON(w, http.StatusCreated, map[string]any{"model": safeShuihuoModels([]map[string]any{model}, false)[0]})
	})))
	root.Handle("/api/shuihuo-production/asset-types", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		config, err := loadShuihuoOwnerConfig(r.Context(), store, owner)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		types := shuihuoCollection(config, "assetTypes", []map[string]any{
			{"id": "system-character", "name": "角色", "category": "character", "system": true},
			{"id": "system-scene", "name": "场景", "category": "scene", "system": true},
			{"id": "system-prop", "name": "道具", "category": "prop", "system": true},
		})
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, map[string]any{"assetTypes": types})
			return
		}
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var input map[string]any
		if !decodeJSON(w, r, &input) {
			return
		}
		name, _ := input["name"].(string)
		category, _ := input["category"].(string)
		if strings.TrimSpace(name) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "type name is required"})
			return
		}
		id, idErr := localexecutor.NewArtifactID()
		if idErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "type id generation failed"})
			return
		}
		types = append(types, map[string]any{"id": id, "name": strings.TrimSpace(name), "category": strings.TrimSpace(category), "system": false})
		config["assetTypes"] = types
		if err := saveShuihuoOwnerConfig(r.Context(), store, owner, config); err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"assetType": types[len(types)-1]})
	})))
	root.Handle("/api/shuihuo-production/asset-types/", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/shuihuo-production/asset-types/"), "/")
		config, err := loadShuihuoOwnerConfig(r.Context(), store, owner)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		types := shuihuoCollection(config, "assetTypes", nil)
		for index, item := range types {
			if item["id"] != id {
				continue
			}
			if item["system"] == true {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "system asset type cannot be deleted"})
				return
			}
			if r.Method == http.MethodDelete {
				types = append(types[:index], types[index+1:]...)
				config["assetTypes"] = types
				if err := saveShuihuoOwnerConfig(r.Context(), store, owner, config); err != nil {
					writeStoreError(w, err)
					return
				}
				writeJSON(w, http.StatusOK, map[string]any{"ok": true})
				return
			}
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "asset type not found"})
	})))
	root.Handle("/api/shuihuo-production/asset-templates", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		config, err := loadShuihuoOwnerConfig(r.Context(), store, owner)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		templates := shuihuoCollection(config, "assetTemplates", nil)
		if r.Method == http.MethodGet {
			if filter := strings.TrimSpace(r.URL.Query().Get("assetTypeId")); filter != "" {
				filtered := []map[string]any{}
				for _, item := range templates {
					if item["assetTypeId"] == filter {
						filtered = append(filtered, item)
					}
				}
				templates = filtered
			}
			writeJSON(w, http.StatusOK, map[string]any{"assetTemplates": templates})
			return
		}
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var input map[string]any
		if !decodeJSON(w, r, &input) {
			return
		}
		name, _ := input["name"].(string)
		typeID, _ := input["assetTypeId"].(string)
		prompt, _ := input["prompt"].(string)
		if strings.TrimSpace(name) == "" || strings.TrimSpace(typeID) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "template name and assetTypeId are required"})
			return
		}
		id, idErr := localexecutor.NewArtifactID()
		if idErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "template id generation failed"})
			return
		}
		item := map[string]any{"id": id, "assetTypeId": strings.TrimSpace(typeID), "name": strings.TrimSpace(name), "prompt": strings.TrimSpace(prompt), "source": "manual"}
		templates = append(templates, item)
		config["assetTemplates"] = templates
		if err := saveShuihuoOwnerConfig(r.Context(), store, owner, config); err != nil {
			writeStoreError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"assetTemplate": item})
	})))
	root.Handle("/api/shuihuo-production/asset-templates/", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/shuihuo-production/asset-templates/"), "/")
		config, err := loadShuihuoOwnerConfig(r.Context(), store, owner)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		templates := shuihuoCollection(config, "assetTemplates", nil)
		for index, item := range templates {
			if item["id"] != id {
				continue
			}
			if r.Method != http.MethodDelete {
				writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
				return
			}
			templates = append(templates[:index], templates[index+1:]...)
			config["assetTemplates"] = templates
			if err := saveShuihuoOwnerConfig(r.Context(), store, owner, config); err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "asset template not found"})
	})))
	root.Handle("/api/shuihuo-production/config", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if r.Method == http.MethodGet {
			if durable, ok := store.(batchfactoryv11.ShuihuoConfigStore); ok {
				config, err := durable.GetShuihuoConfig(r.Context(), owner)
				if err != nil {
					writeStoreError(w, err)
					return
				}
				merged := shuihuoDefaultConfig()
				for key, value := range config {
					merged[key] = value
				}
				writeJSON(w, http.StatusOK, merged)
				return
			}
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
			if durable, ok := store.(batchfactoryv11.ShuihuoConfigStore); ok {
				if err := durable.SaveShuihuoConfig(r.Context(), owner, config); err != nil {
					writeStoreError(w, err)
					return
				}
				writeJSON(w, http.StatusOK, config)
				return
			}
			shuihuoProductionConfigs.Store(owner, config)
			writeJSON(w, http.StatusOK, config)
			return
		}
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	})))
	root.Handle("/api/shuihuo-production/projects/import", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		var input struct {
			Name     string `json:"name"`
			Filename string `json:"filename"`
			DataURL  string `json:"dataUrl"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		name := strings.TrimSpace(input.Name)
		if name == "" || strings.TrimSpace(input.DataURL) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and dataUrl are required"})
			return
		}
		parts := strings.SplitN(strings.TrimSpace(input.DataURL), ",", 2)
		if len(parts) != 2 || !strings.HasSuffix(parts[0], ";base64") {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid source file"})
			return
		}
		raw, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil || len(raw) == 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid source file"})
			return
		}
		ext := strings.ToLower(filepath.Ext(strings.TrimSpace(input.Filename)))
		if ext == ".docx" {
			writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "Go 本地迁移阶段暂未接入 DOCX 正文解析，请先保存为 TXT 或 SRT"})
			return
		}
		if ext != "" && ext != ".txt" && ext != ".srt" && ext != ".vtt" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "仅支持 TXT、SRT 或 VTT 文件"})
			return
		}
		batch, err := store.CreateBatch(r.Context(), owner, batchfactoryv11.CreateBatchInput{Title: name, InitialBatchSettings: batchfactoryv11.SettingsPatch{"shuihuoProject": json.RawMessage("true")}, Books: []batchfactoryv11.CreateBookInput{{Title: name}}})
		if err != nil {
			writeStoreError(w, err)
			return
		}
		if len(batch.Books) == 0 {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "project has no editable source book"})
			return
		}
		book, err := store.CaptureBookSource(r.Context(), owner, batch.ID, batch.Books[0].ID, batchfactoryv11.CaptureBookSourceInput{SourceText: string(raw), ReplaceSource: true, SourceTitle: input.Filename, ExpectedRevision: batch.Books[0].Revision})
		if err != nil {
			_ = store.DeleteBatch(r.Context(), owner, batch.ID)
			writeStoreError(w, err)
			return
		}
		batch.Books[0] = book
		writeJSON(w, http.StatusCreated, shuihuoProjectDetail(batch))
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
			if len(batch.Books) == 0 {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "project has no editable source book"})
				return
			}
			config := storedShuihuoAssetGenerationConfig(batch)
			if r.Method == http.MethodPut {
				var input map[string]any
				if !decodeJSON(w, r, &input) {
					return
				}
				config = input
				metadata := cloneMetadata(batch.Books[0].SourceMetadata)
				encoded, _ := json.Marshal(config)
				var normalized map[string]any
				if json.Unmarshal(encoded, &normalized) == nil {
					metadata["shuihuoAssetGenerationConfig"] = normalized
				}
				if _, saveErr := store.UpdateBookMetadata(r.Context(), owner, batch.ID, batch.Books[0].ID, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata, ExpectedRevision: batch.Books[0].Revision}); saveErr != nil {
					writeStoreError(w, saveErr)
					return
				}
			}
			writeJSON(w, http.StatusOK, config)
			return
		}
		if len(parts) == 2 && parts[1] == "files" && r.Method == http.MethodGet {
			if len(batch.Books) == 0 {
				writeJSON(w, http.StatusOK, map[string]any{"hasSourceFile": false, "media": []any{}})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"hasSourceFile": strings.TrimSpace(batch.Books[0].TxtFileName) != "" || strings.TrimSpace(batch.Books[0].SourceText) != "", "media": storedShuihuoMedia(batch.Books[0])})
			return
		}
		if len(parts) == 2 && parts[1] == "media" && r.Method == http.MethodPost {
			if files == nil || len(batch.Books) == 0 {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "media storage is unavailable"})
				return
			}
			var input struct {
				Kind      string `json:"kind"`
				Filename  string `json:"filename"`
				DataURL   string `json:"dataUrl"`
				SegmentID string `json:"segmentId,omitempty"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			mediaType, payload, err := decodeShuihuoMediaDataURL(input.DataURL)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid media upload"})
				return
			}
			kind := strings.ToLower(strings.TrimSpace(input.Kind))
			if kind != "image" && kind != "video" && kind != "audio" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported media kind"})
				return
			}
			artifactID, err := localexecutor.NewArtifactID()
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "media id generation failed"})
				return
			}
			extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(strings.TrimSpace(input.Filename))), ".")
			if extension == "" {
				extension = map[string]string{"image": "png", "video": "mp4", "audio": "mp3"}[kind]
			}
			var saved localartifact.Artifact
			if kind == "image" {
				saved, err = files.SaveImage(artifactID, mediaType, bytes.NewReader(payload))
			} else {
				saved, err = files.SaveMedia(artifactID, mediaType, extension, bytes.NewReader(payload))
			}
			if err != nil {
				writeBookAssetImageStorageError(w, err)
				return
			}
			media := storedShuihuoMedia(batch.Books[0])
			media = append(media, map[string]any{"id": artifactID, "kind": kind, "filename": input.Filename, "mediaType": mediaType, "storageRef": saved.StorageRef, "segmentId": strings.TrimSpace(input.SegmentID), "isPrimary": false, "createdAt": time.Now().UTC().Format(time.RFC3339)})
			metadata := cloneMetadata(batch.Books[0].SourceMetadata)
			metadata["shuihuoMedia"] = media
			if _, err := store.UpdateBookMetadata(r.Context(), owner, batch.ID, batch.Books[0].ID, batchfactoryv11.UpdateBookMetadataInput{Metadata: metadata, ExpectedRevision: batch.Books[0].Revision}); err != nil {
				_ = files.Remove(saved.StorageRef)
				writeStoreError(w, err)
				return
			}
			updated, err := store.GetBatch(r.Context(), owner, batch.ID)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]any{"media": map[string]any{"id": artifactID, "kind": kind, "filename": input.Filename, "mediaType": mediaType, "segmentId": strings.TrimSpace(input.SegmentID), "isPrimary": false}, "project": shuihuoProjectDetail(updated)})
			return
		}
		if len(parts) == 2 && parts[1] == "tasks" && r.Method == http.MethodGet {
			if production == nil {
				writeJSON(w, http.StatusOK, map[string]any{"tasks": []any{}})
				return
			}
			status, statusErr := production.GetBatchStatus(r.Context(), owner, batch.ID)
			if statusErr != nil {
				writeStoreError(w, statusErr)
				return
			}
			tasks := make([]map[string]any, 0)
			for _, job := range status.Jobs {
				for _, task := range job.Tasks {
					tasks = append(tasks, map[string]any{"id": task.ID, "status": task.Status, "kind": "video", "mediaUrl": task.MediaURL, "errorMessage": task.ErrorMessage, "requestId": job.RequestID, "createdAt": task.CreatedAt, "updatedAt": task.UpdatedAt})
				}
			}
			writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
			return
		}
		if len(parts) == 2 && parts[1] == "tasks" && r.Method == http.MethodPost {
			if production == nil || len(batch.Books) == 0 {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "当前 Go 本地环境尚未配置可用的生成模型"})
				return
			}
			var input struct {
				RequestID string `json:"requestId"`
				Provider  string `json:"provider,omitempty"`
				ModelID   string `json:"modelId,omitempty"`
				Kind      string `json:"kind,omitempty"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			if kind := strings.ToLower(strings.TrimSpace(input.Kind)); kind != "" && kind != "video" {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "Go 本地迁移阶段暂未接入图片或配音任务，请先使用视频任务"})
				return
			}
			if production.ProviderRegistry != nil {
				if syncErr := syncShuihuoConfiguredVideoModels(r.Context(), store, production.ProviderRegistry, owner); syncErr != nil {
					writeStoreError(w, syncErr)
					return
				}
			}
			if strings.TrimSpace(input.RequestID) == "" {
				input.RequestID = "shuihuo-" + time.Now().UTC().Format("20060102T150405.000000000")
			}
			provider := input.Provider
			if strings.TrimSpace(input.ModelID) != "" {
				provider = batchfactoryv11.VideoProviderForModel(input.ModelID, provider)
			}
			job, submitErr := production.SubmitBookProductionWithProvider(r.Context(), owner, batch.ID, batch.Books[0].ID, input.RequestID, provider)
			if submitErr != nil {
				writeStoreError(w, submitErr)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]any{"tasks": job.Tasks, "job": job})
			return
		}
		if len(parts) == 3 && parts[1] == "tasks" && parts[2] == "batch" && r.Method == http.MethodPost {
			if production == nil {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "当前 Go 本地环境尚未配置可用的生成模型"})
				return
			}
			var input struct {
				RequestID string `json:"requestId"`
				Provider  string `json:"provider,omitempty"`
				ModelID   string `json:"modelId,omitempty"`
			}
			if !decodeJSON(w, r, &input) {
				return
			}
			if production.ProviderRegistry != nil {
				if syncErr := syncShuihuoConfiguredVideoModels(r.Context(), store, production.ProviderRegistry, owner); syncErr != nil {
					writeStoreError(w, syncErr)
					return
				}
			}
			if strings.TrimSpace(input.RequestID) == "" {
				input.RequestID = "shuihuo-batch-" + time.Now().UTC().Format("20060102T150405.000000000")
			}
			provider := input.Provider
			if strings.TrimSpace(input.ModelID) != "" {
				provider = batchfactoryv11.VideoProviderForModel(input.ModelID, provider)
			}
			status, submitErr := production.SubmitBatchProductionWithProvider(r.Context(), owner, batch.ID, input.RequestID, provider)
			if submitErr != nil {
				writeStoreError(w, submitErr)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]any{"tasks": status.Jobs, "status": status})
			return
		}
		if len(parts) == 3 && parts[1] == "analysis" && r.Method == http.MethodPost {
			if director == nil || len(batch.Books) == 0 {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "当前 Go 本地环境尚未配置文本模型，暂不能进行智能分析"})
				return
			}
			assets, analysisErr := director.RunAssetExtraction(r.Context(), owner, batch.ID, batch.Books[0].ID)
			if analysisErr != nil {
				writeStoreError(w, analysisErr)
				return
			}
			updated, getErr := store.GetBatch(r.Context(), owner, batch.ID)
			if getErr != nil {
				writeStoreError(w, getErr)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"assets": assets, "project": shuihuoProjectDetail(updated)})
			return
		}
		if len(parts) == 3 && parts[1] == "prompt-candidates" && r.Method == http.MethodPost {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "当前 Go 本地环境尚未配置文本模型，暂不能生成提示词候选"})
			return
		}
		if len(parts) == 4 && parts[1] == "prompt-candidates" && r.Method == http.MethodPut {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "当前 Go 本地环境尚未配置文本模型，暂不能应用提示词候选"})
			return
		}
		if len(parts) == 2 && parts[1] == "export" && r.Method == http.MethodGet {
			if files == nil || len(batch.Books) == 0 {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "导出素材存储不可用"})
				return
			}
			if err := writeShuihuoProjectExport(w, r, files, batch); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "项目导出失败"})
			}
			return
		}
		if len(parts) == 2 && parts[1] == "media" && r.Method == http.MethodPost {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "Go 本地迁移阶段暂未接入媒体素材存储"})
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

	root.Handle("/api/shuihuo-production/media/", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/shuihuo-production/media/"), "/")
		parts := strings.Split(path, "/")
		if len(parts) == 0 || parts[0] == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "media id is required"})
			return
		}
		mediaID := parts[0]
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
			media := storedShuihuoMedia(batch.Books[0])
			index := -1
			for i, item := range media {
				if id, _ := item["id"].(string); id == mediaID {
					index = i
					break
				}
			}
			if index < 0 {
				continue
			}
			if len(parts) == 2 && parts[1] == "download" && r.Method == http.MethodGet {
				if files == nil {
					writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "media storage is unavailable"})
					return
				}
				storageRef, _ := media[index]["storageRef"].(string)
				file, openErr := files.Open(storageRef)
				if openErr != nil {
					writeJSON(w, http.StatusNotFound, map[string]string{"error": "media file not found"})
					return
				}
				defer file.Close()
				mediaType, _ := media[index]["mediaType"].(string)
				if mediaType == "" {
					mediaType = "application/octet-stream"
				}
				w.Header().Set("Content-Type", mediaType)
				http.ServeContent(w, r, mediaID, time.Now(), file)
				return
			}
			if len(parts) == 2 && parts[1] == "segment" && r.Method == http.MethodPut {
				var input struct {
					SegmentID string `json:"segmentId"`
				}
				if !decodeJSON(w, r, &input) {
					return
				}
				media[index]["segmentId"] = strings.TrimSpace(input.SegmentID)
			} else if len(parts) == 2 && parts[1] == "primary" && r.Method == http.MethodPut {
				for i := range media {
					media[i]["isPrimary"] = i == index
				}
			} else if len(parts) == 1 && r.Method == http.MethodDelete {
				if files != nil {
					if storageRef, _ := media[index]["storageRef"].(string); storageRef != "" {
						_ = files.Remove(storageRef)
					}
				}
				media = append(media[:index], media[index+1:]...)
			} else if len(parts) == 1 {
				writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
				return
			} else {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "media route not found"})
				return
			}
			metadata := cloneMetadata(batch.Books[0].SourceMetadata)
			metadata["shuihuoMedia"] = media
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
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "media not found"})
	})))

	// The legacy task drawer addresses tasks by task id, while the Go
	// production service cancels by the owning batch and request id. Resolve
	// that mapping here so the drawer no longer falls into a 404 hole.
	root.Handle("/api/shuihuo-production/tasks/", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if production == nil {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "当前 Go 本地环境尚未配置可用的生成模型"})
			return
		}
		path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/shuihuo-production/tasks/"), "/")
		parts := strings.Split(path, "/")
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "task route not found"})
			return
		}
		taskID, action := parts[0], parts[1]
		if action == "retry" && r.Method == http.MethodPost {
			if director == nil {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "当前 Go 本地环境尚未配置导演服务，暂不能重试任务"})
				return
			}
			batches, err := store.ListBatches(r.Context(), owner)
			if err != nil {
				writeStoreError(w, err)
				return
			}
			for _, compact := range batches {
				status, statusErr := production.GetBatchStatus(r.Context(), owner, compact.ID)
				if statusErr != nil {
					continue
				}
				for _, job := range status.Jobs {
					for _, task := range job.Tasks {
						if task.ID != taskID {
							continue
						}
						stages := &batchfactoryv11.BookStageService{Store: store, Director: director, Production: production}
						summary, retryErr := stages.RetryLastFailed(r.Context(), owner, compact.ID, job.BookID, "shuihuo-retry-"+time.Now().UTC().Format("20060102T150405.000000000"), task.VideoID)
						if retryErr != nil {
							writeStoreError(w, retryErr)
							return
						}
						writeJSON(w, http.StatusOK, summary)
						return
					}
				}
			}
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
			return
		}
		if action != "cancel" || r.Method != http.MethodPut {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		batches, err := store.ListBatches(r.Context(), owner)
		if err != nil {
			writeStoreError(w, err)
			return
		}
		for _, compact := range batches {
			status, statusErr := production.GetBatchStatus(r.Context(), owner, compact.ID)
			if statusErr != nil {
				continue
			}
			for _, job := range status.Jobs {
				for _, task := range job.Tasks {
					if task.ID != taskID {
						continue
					}
					result, cancelErr := production.CancelBatch(r.Context(), owner, compact.ID, job.RequestID)
					if cancelErr != nil {
						writeStoreError(w, cancelErr)
						return
					}
					writeJSON(w, http.StatusOK, result)
					return
				}
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
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
		"media": shuihuoMedia(batch),
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
	media, _ := project["media"].([]map[string]any)
	return map[string]any{
		"project":         project,
		"segments":        segments,
		"assets":          assets,
		"media":           media,
		"segmentAssetIDs": map[string][]string{},
		"settings":        map[string]any{},
	}
}

func shuihuoMedia(batch batchfactoryv11.Batch) []map[string]any {
	if len(batch.Books) == 0 {
		return []map[string]any{}
	}
	return storedShuihuoMedia(batch.Books[0])
}

func loadShuihuoOwnerConfig(ctx context.Context, store batchfactoryv11.Store, owner string) (map[string]any, error) {
	if durable, ok := store.(batchfactoryv11.ShuihuoConfigStore); ok {
		return durable.GetShuihuoConfig(ctx, owner)
	}
	if value, ok := shuihuoProductionConfigs.Load(owner); ok {
		if config, ok := value.(map[string]any); ok {
			return config, nil
		}
	}
	config := shuihuoDefaultConfig()
	shuihuoProductionConfigs.Store(owner, config)
	return config, nil
}

func saveShuihuoOwnerConfig(ctx context.Context, store batchfactoryv11.Store, owner string, config map[string]any) error {
	if durable, ok := store.(batchfactoryv11.ShuihuoConfigStore); ok {
		return durable.SaveShuihuoConfig(ctx, owner, config)
	}
	shuihuoProductionConfigs.Store(owner, config)
	return nil
}

func shuihuoCollection(config map[string]any, key string, fallback []map[string]any) []map[string]any {
	if raw, ok := config[key].([]any); ok {
		items := make([]map[string]any, 0, len(raw))
		for _, value := range raw {
			if item, ok := value.(map[string]any); ok {
				items = append(items, item)
			}
		}
		return items
	}
	if raw, ok := config[key].([]map[string]any); ok {
		return raw
	}
	return append([]map[string]any(nil), fallback...)
}

func safeShuihuoModels(models []map[string]any, includeSecrets bool) []map[string]any {
	result := make([]map[string]any, 0, len(models))
	for _, model := range models {
		copy := map[string]any{}
		for _, key := range []string{"id", "modelId", "name", "displayName", "kind", "provider", "adapterKind", "enabled", "configured", "parameterSchema", "credentialRef", "endpoint", "requestTemplate", "responseMapping"} {
			if value, exists := model[key]; exists {
				copy[key] = value
			}
		}
		if includeSecrets {
			if value, exists := model["credential"]; exists {
				copy["credential"] = value
			}
			if value, exists := model["apiKey"]; exists {
				copy["apiKey"] = value
			}
		}
		result = append(result, copy)
	}
	return result
}

func syncShuihuoVideoModel(ctx context.Context, registry batchfactoryv11.VideoProviderRegistry, owner string, model map[string]any) error {
	modelID, _ := model["modelId"].(string)
	provider, _ := model["provider"].(string)
	provider = batchfactoryv11.VideoProviderForModel(modelID, provider)
	apiKey, _ := model["apiKey"].(string)
	if strings.TrimSpace(apiKey) == "" {
		apiKey, _ = model["credential"].(string)
	}
	endpoint, _ := model["endpoint"].(string)
	tasksEndpoint, _ := model["tasksEndpoint"].(string)
	resultEndpoint, _ := model["resultEndpoint"].(string)
	workflowID, _ := model["workflowId"].(string)
	return registry.Put(ctx, owner, batchfactoryv11.VideoProviderConfig{Provider: provider, APIKey: apiKey, Model: modelID, CreateURL: endpoint, TasksURL: tasksEndpoint, ResultURL: resultEndpoint, WorkflowID: workflowID})
}

func syncShuihuoConfiguredVideoModels(ctx context.Context, store batchfactoryv11.Store, registry batchfactoryv11.VideoProviderRegistry, owner string) error {
	config, err := loadShuihuoOwnerConfig(ctx, store, owner)
	if err != nil {
		return err
	}
	for _, model := range shuihuoCollection(config, "modelCatalog", nil) {
		kind, _ := model["kind"].(string)
		enabled, enabledSet := model["enabled"].(bool)
		if !strings.EqualFold(strings.TrimSpace(kind), "video") || (enabledSet && !enabled) {
			continue
		}
		apiKey, _ := model["apiKey"].(string)
		if strings.TrimSpace(apiKey) == "" {
			apiKey, _ = model["credential"].(string)
		}
		provider, _ := model["provider"].(string)
		if strings.TrimSpace(apiKey) == "" && batchfactoryv11.VideoProviderForModel(modelString(model, "modelId"), provider) != batchfactoryv11.VideoProviderDoubaoLocal {
			continue
		}
		if err := syncShuihuoVideoModel(ctx, registry, owner, model); err != nil {
			return err
		}
	}
	return nil
}

func modelString(model map[string]any, key string) string {
	value, _ := model[key].(string)
	return strings.TrimSpace(value)
}

func writeShuihuoProjectExport(w http.ResponseWriter, r *http.Request, files *localartifact.Store, batch batchfactoryv11.Batch) error {
	var archive bytes.Buffer
	zw := zip.NewWriter(&archive)
	book := batch.Books[0]
	writeEntry := func(name string, data []byte) error {
		entry, err := zw.Create(name)
		if err != nil {
			return err
		}
		_, err = entry.Write(data)
		return err
	}
	if err := writeEntry("source.txt", []byte(book.SourceText)); err != nil {
		return err
	}
	segments, err := json.MarshalIndent(storedShuihuoSegments(book), "", "  ")
	if err != nil {
		return err
	}
	if err := writeEntry("segments.json", segments); err != nil {
		return err
	}
	project, err := json.MarshalIndent(shuihuoProjectDetail(batch), "", "  ")
	if err != nil {
		return err
	}
	if err := writeEntry("project.json", project); err != nil {
		return err
	}
	for _, media := range storedShuihuoMedia(book) {
		storageRef, _ := media["storageRef"].(string)
		mediaID, _ := media["id"].(string)
		if strings.TrimSpace(storageRef) == "" || strings.TrimSpace(mediaID) == "" {
			continue
		}
		file, openErr := files.Open(storageRef)
		if openErr != nil {
			continue
		}
		name := "media/" + mediaID
		if filename, ok := media["filename"].(string); ok && strings.TrimSpace(filename) != "" {
			name = "media/" + filepath.Base(filename)
		}
		entry, createErr := zw.Create(name)
		if createErr == nil {
			_, createErr = io.Copy(entry, file)
		}
		_ = file.Close()
		if createErr != nil {
			return createErr
		}
	}
	if err := zw.Close(); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=shuihuo-project-%s.zip", batch.ID))
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(archive.Bytes())
	return err
}

func storedShuihuoMedia(book batchfactoryv11.Book) []map[string]any {
	value, ok := book.SourceMetadata["shuihuoMedia"]
	if !ok {
		return []map[string]any{}
	}
	if items, ok := value.([]map[string]any); ok {
		return items
	}
	if items, ok := value.([]any); ok {
		output := make([]map[string]any, 0, len(items))
		for _, item := range items {
			if media, ok := item.(map[string]any); ok {
				output = append(output, media)
			}
		}
		return output
	}
	return []map[string]any{}
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

func storedShuihuoAssetGenerationConfig(batch batchfactoryv11.Batch) map[string]any {
	defaults := map[string]any{"textModelId": nil, "imageModelId": nil, "audioModelId": nil, "assetPresetId": "shuihuo-extract-assets", "characterSheetPresetId": nil, "aspectRatio": "16:9"}
	if len(batch.Books) == 0 || batch.Books[0].SourceMetadata == nil {
		return defaults
	}
	raw, ok := batch.Books[0].SourceMetadata["shuihuoAssetGenerationConfig"]
	if !ok {
		return defaults
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return defaults
	}
	var stored map[string]any
	if json.Unmarshal(encoded, &stored) != nil {
		return defaults
	}
	for key, value := range stored {
		defaults[key] = value
	}
	return defaults
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

func decodeShuihuoMediaDataURL(value string) (string, []byte, error) {
	parts := strings.SplitN(strings.TrimSpace(value), ",", 2)
	if len(parts) != 2 || !strings.HasPrefix(parts[0], "data:") || !strings.HasSuffix(parts[0], ";base64") {
		return "", nil, fmt.Errorf("invalid data url")
	}
	mediaType, _, err := mime.ParseMediaType(strings.TrimSuffix(strings.TrimPrefix(parts[0], "data:"), ";base64"))
	if err != nil {
		return "", nil, err
	}
	payload, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil || len(payload) == 0 {
		return "", nil, fmt.Errorf("invalid base64")
	}
	return strings.ToLower(mediaType), payload, nil
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
