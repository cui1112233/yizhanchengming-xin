package httpapi

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"qiantie/backend/internal/agentworkspace"
)

const agentAssetUploadMaxBytes = 128 << 20

func RegisterAgentAssetRoutes(root *http.ServeMux, middleware BridgeAuth, projects agentworkspace.Store, assets agentworkspace.AssetStore, objects agentworkspace.AssetObjectStore, clock func() time.Time) {
	if root == nil || projects == nil || assets == nil {
		return
	}
	if clock == nil {
		clock = time.Now
	}
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		root.Handle(pattern, middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity, ok := BridgeIdentityFromContext(req.Context())
			if !ok || strings.TrimSpace(identity.Username) == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			fn(w, req, strings.TrimSpace(identity.Username))
		})))
	}
	projectFor := func(w http.ResponseWriter, req *http.Request, owner string) bool {
		if _, err := projects.Project(req.Context(), owner, req.PathValue("id")); err != nil {
			writeAgentStoreError(w, err)
			return false
		}
		return true
	}
	handle("GET /api/agent/tasks/{id}/assets", func(w http.ResponseWriter, req *http.Request, owner string) {
		if !projectFor(w, req, owner) {
			return
		}
		kind := strings.TrimSpace(req.URL.Query().Get("kind"))
		if kind != "" && !agentworkspace.ValidAssetKind(kind) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "素材类型仅支持图片或视频"})
			return
		}
		items, err := assets.ListAssets(req.Context(), owner, req.PathValue("id"), kind)
		if err != nil {
			writeAgentAssetError(w, err)
			return
		}
		for index := range items {
			items[index].URL = agentAssetContentURL(req.PathValue("id"), items[index].ID)
		}
		writeJSON(w, http.StatusOK, map[string]any{"assets": items})
	})
	handle("POST /api/agent/tasks/{id}/assets", func(w http.ResponseWriter, req *http.Request, owner string) {
		if !projectFor(w, req, owner) {
			return
		}
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "素材存储尚未配置"})
			return
		}
		req.Body = http.MaxBytesReader(w, req.Body, agentAssetUploadMaxBytes+(1<<20))
		if err := req.ParseMultipartForm(agentAssetUploadMaxBytes); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "素材文件无效或超过大小限制"})
			return
		}
		file, header, err := req.FormFile("file")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择素材文件"})
			return
		}
		defer file.Close()
		contents, err := io.ReadAll(io.LimitReader(file, agentAssetUploadMaxBytes+1))
		if err != nil || len(contents) > agentAssetUploadMaxBytes {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "素材文件读取失败或超过大小限制"})
			return
		}
		contentType, kind := detectedAgentAssetType(header.Header.Get("Content-Type"), header.Filename, contents)
		if kind == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "仅支持图片或视频素材"})
			return
		}
		assetID := newAgentID("asset")
		key, err := agentworkspace.AgentAssetObjectKey(owner, req.PathValue("id"), assetID, path.Ext(header.Filename))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "素材文件名无效"})
			return
		}
		asset := agentworkspace.Asset{ID: assetID, ProjectID: req.PathValue("id"), Kind: kind, Name: path.Base(header.Filename), ContentType: contentType, ObjectKey: key, ByteSize: int64(len(contents)), CreatedAt: clock().UTC()}
		if err := objects.Put(req.Context(), asset, bytes.NewReader(contents)); err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "素材上传失败"})
			return
		}
		if err := assets.CreateAsset(req.Context(), owner, asset); err != nil {
			_ = objects.Delete(req.Context(), asset)
			writeAgentAssetError(w, err)
			return
		}
		asset.URL = agentAssetContentURL(asset.ProjectID, asset.ID)
		writeJSON(w, http.StatusCreated, map[string]any{"asset": asset})
	})
	handle("GET /api/agent/tasks/{id}/assets/{assetID}/content", func(w http.ResponseWriter, req *http.Request, owner string) {
		if !projectFor(w, req, owner) {
			return
		}
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "素材存储尚未配置"})
			return
		}
		asset, err := assets.Asset(req.Context(), owner, req.PathValue("id"), req.PathValue("assetID"))
		if err != nil {
			writeAgentAssetError(w, err)
			return
		}
		body, err := objects.Open(req.Context(), asset)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "素材读取失败"})
			return
		}
		defer body.Close()
		w.Header().Set("Content-Type", asset.ContentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", asset.ByteSize))
		w.Header().Set("Cache-Control", "private, max-age=300")
		_, _ = io.Copy(w, body)
	})
}

func detectedAgentAssetType(requestType, filename string, contents []byte) (string, string) {
	declared := strings.ToLower(strings.TrimSpace(strings.Split(requestType, ";")[0]))
	detected := http.DetectContentType(contents)
	for _, contentType := range []string{detected, declared, mime.TypeByExtension(strings.ToLower(path.Ext(filename)))} {
		contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
		if strings.HasPrefix(contentType, "image/") {
			return contentType, "image"
		}
		if strings.HasPrefix(contentType, "video/") {
			return contentType, "video"
		}
	}
	return "", ""
}

func agentAssetContentURL(projectID, assetID string) string {
	return "/api/agent/tasks/" + projectID + "/assets/" + assetID + "/content"
}

func writeAgentAssetError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, agentworkspace.ErrAssetNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "素材不存在"})
	case errors.Is(err, agentworkspace.ErrAssetExists):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "素材已存在"})
	case errors.Is(err, agentworkspace.ErrInvalidAsset):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "素材信息无效"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "素材保存失败"})
	}
}
