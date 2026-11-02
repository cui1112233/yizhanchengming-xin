package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"qiantie/backend/internal/agentworkspace"
)

func RegisterAgentMaterialLibraryRoutes(mux *http.ServeMux, auth BridgeAuth, store agentworkspace.Store, assets agentworkspace.AssetStore, now func() time.Time) {
	if mux == nil || store == nil {
		return
	}
	mux.Handle("GET /api/agent/material-library/folders", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		owner, ok := agentLibraryOwner(w, req)
		if !ok {
			return
		}
		folders, err := store.ListLibraryFolders(req.Context(), owner)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "素材库文件夹读取失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"folders": folders})
	})))
	mux.Handle("POST /api/agent/material-library/folders", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		owner, ok := agentLibraryOwner(w, req)
		if !ok {
			return
		}
		var input struct {
			Name string `json:"name"`
		}
		if !decodeJSON(w, req, &input) {
			return
		}
		folder, err := store.CreateLibraryFolder(req.Context(), owner, agentworkspace.CreateLibraryFolderInput{ID: newAgentID("library-folder"), Name: input.Name, CreatedAt: agentLibraryNow(now)})
		if err != nil {
			writeAgentLibraryError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"folder": folder})
	})))
	mux.Handle("GET /api/agent/material-library/folders/{folderID}/items", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		owner, ok := agentLibraryOwner(w, req)
		if !ok {
			return
		}
		items, err := store.ListLibraryItems(req.Context(), owner, req.PathValue("folderID"), agentworkspace.LibraryItemFilter{Kind: strings.TrimSpace(req.URL.Query().Get("kind")), Query: strings.TrimSpace(req.URL.Query().Get("q"))})
		if err != nil {
			writeAgentLibraryError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	})))
	mux.Handle("POST /api/agent/material-library/folders/{folderID}/items", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		owner, ok := agentLibraryOwner(w, req)
		if !ok {
			return
		}
		if assets == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "素材存储尚未配置"})
			return
		}
		var input struct {
			ProjectID string   `json:"projectId"`
			AssetID   string   `json:"assetId"`
			Tags      []string `json:"tags"`
		}
		if !decodeJSON(w, req, &input) {
			return
		}
		source, err := assets.Asset(req.Context(), owner, strings.TrimSpace(input.ProjectID), strings.TrimSpace(input.AssetID))
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "要保存的项目素材不存在"})
			return
		}
		item, err := store.SaveLibraryItem(req.Context(), owner, agentworkspace.SaveLibraryItemInput{ID: newAgentID("library-item"), FolderID: req.PathValue("folderID"), SourceProjectID: source.ProjectID, SourceAssetID: source.ID, Kind: source.Kind, Name: source.Name, Tags: input.Tags, CreatedAt: agentLibraryNow(now)})
		if err != nil {
			writeAgentLibraryError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"item": item})
	})))
	mux.Handle("PATCH /api/agent/material-library/items/{itemID}", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		owner, ok := agentLibraryOwner(w, req)
		if !ok {
			return
		}
		var input struct {
			FolderID string `json:"folderId"`
		}
		if !decodeJSON(w, req, &input) {
			return
		}
		item, err := store.MoveLibraryItem(req.Context(), owner, req.PathValue("itemID"), input.FolderID, agentLibraryNow(now))
		if err != nil {
			writeAgentLibraryError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	})))
	mux.Handle("POST /api/agent/tasks/{id}/material-library/items/{itemID}/insert", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		owner, ok := agentLibraryOwner(w, req)
		if !ok {
			return
		}
		if assets == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "素材存储尚未配置"})
			return
		}
		item, err := store.LibraryItem(req.Context(), owner, req.PathValue("itemID"))
		if err != nil {
			writeAgentLibraryError(w, err)
			return
		}
		source, err := assets.Asset(req.Context(), owner, item.SourceProjectID, item.SourceAssetID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "素材库来源素材不存在"})
			return
		}
		targetProjectID := strings.TrimSpace(req.PathValue("id"))
		canvas, err := store.Canvas(req.Context(), owner, targetProjectID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "当前项目不存在"})
			return
		}
		createdAt := agentLibraryNow(now)
		assetID := newAgentID("library-asset")
		importedAsset := agentworkspace.Asset{ID: assetID, ProjectID: targetProjectID, Kind: source.Kind, Name: source.Name, ContentType: source.ContentType, ObjectKey: source.ObjectKey, ByteSize: source.ByteSize, CreatedAt: createdAt}
		if err := assets.CreateAsset(req.Context(), owner, importedAsset); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "素材库素材写入失败"})
			return
		}
		node := agentworkspace.Node{ID: newAgentID("library-node"), Type: source.Kind, Title: "素材库：" + source.Name, Version: 1, Status: "confirmed", Position: agentGenerationHistoryNodePosition(canvas), AssetRef: map[string]any{
			"id": assetID, "assetId": assetID, "kind": source.Kind, "name": source.Name, "url": agentAssetContentURL(targetProjectID, assetID),
			"originProjectId": item.SourceProjectID, "originAssetId": item.SourceAssetID, "libraryItemId": item.ID,
		}}
		canvas.Nodes = append(canvas.Nodes, node)
		updated, err := store.SaveCanvas(req.Context(), owner, targetProjectID, canvas, canvas.Revision, createdAt)
		if err != nil {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "当前项目已更新，请刷新后重试"})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"canvas": updated, "node": node, "asset": importedAsset, "libraryItem": item})
	})))
}

func agentLibraryOwner(w http.ResponseWriter, req *http.Request) (string, bool) {
	identity, ok := BridgeIdentityFromContext(req.Context())
	if !ok || strings.TrimSpace(identity.Username) == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return "", false
	}
	return identity.Username, true
}

func agentLibraryNow(now func() time.Time) time.Time {
	if now != nil {
		return now().UTC()
	}
	return time.Now().UTC()
}

func writeAgentLibraryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, agentworkspace.ErrLibraryInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "素材库参数不正确"})
	case errors.Is(err, agentworkspace.ErrLibraryFolderNotFound), errors.Is(err, agentworkspace.ErrLibraryItemNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "素材库项目不存在"})
	case errors.Is(err, agentworkspace.ErrLibraryFolderExists), errors.Is(err, agentworkspace.ErrLibraryItemExists):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "素材库中已有相同素材"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "素材库操作失败"})
	}
}
