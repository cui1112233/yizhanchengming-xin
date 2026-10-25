package httpapi

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
	"qiantie/backend/internal/prompt"
)

func RegisterBatchComicReferenceRoutes(root *http.ServeMux, auth BridgeAuth, store batchfactoryv11.Store, files *localartifact.Store, executors *localexecutor.Service) {
	if root == nil || store == nil || files == nil {
		return
	}
	uploadPattern := "POST /api/batch-factory/v12/batches/{batchId}/comic-reference/upload"
	previewPattern := "GET /api/batch-factory/v12/batches/{batchId}/comic-references/{assetId}/content"
	root.Handle(uploadPattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if _, err := store.GetBatch(r.Context(), owner, r.PathValue("batchId")); err != nil {
			writeStoreError(w, err)
			return
		}
		var input struct {
			DataURL string `json:"dataUrl"`
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		mediaType, payload, err := decodeImageDataURL(input.DataURL)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "参考图格式无效"})
			return
		}
		assetID, err := localexecutor.NewArtifactID()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "参考图编号生成失败"})
			return
		}
		saved, err := files.SaveImage(assetID, mediaType, bytes.NewReader(payload))
		if err != nil {
			writeBookAssetImageStorageError(w, err)
			return
		}
		batchID := r.PathValue("batchId")
		reference := batchfactoryv11.ComicGlobalReferenceImage{
			AssetID:    assetID,
			URL:        "/api/local-executor/v1/batch-comic-references/" + batchID + "/" + assetID,
			PreviewURL: "/api/batch-factory/v12/batches/" + batchID + "/comic-references/" + assetID + "/content",
			Enabled:    true, ApplyToAssets: true, ApplyToPages: true,
		}
		_ = saved
		writeJSON(w, http.StatusCreated, map[string]any{"reference": reference})
	})))
	root.Handle(previewPattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		serveBatchComicReference(w, r, store, files, owner, r.PathValue("batchId"), r.PathValue("assetId"))
	})))
	if executors == nil {
		return
	}
	root.HandleFunc("GET /api/local-executor/v1/batch-comic-references/{batchId}/{assetId}", func(w http.ResponseWriter, r *http.Request) {
		token, ok := executorBearerToken(r)
		if !ok {
			writeExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		owner, err := executors.OwnerForToken(r.Context(), token)
		if err != nil {
			writeLocalJobError(w, err)
			return
		}
		serveBatchComicReference(w, r, store, files, owner, r.PathValue("batchId"), r.PathValue("assetId"))
	})
	root.HandleFunc("GET /api/local-executor/v1/batch-comic-style-previews/{presetId}", func(w http.ResponseWriter, r *http.Request) {
		token, ok := executorBearerToken(r)
		if !ok {
			writeExecutorError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if _, err := executors.OwnerForToken(r.Context(), token); err != nil {
			writeLocalJobError(w, err)
			return
		}
		payload, mediaType, ok := prompt.OpenComicStylePreview(r.PathValue("presetId"))
		if !ok {
			writeExecutorError(w, http.StatusNotFound, "comic style preview not found")
			return
		}
		w.Header().Set("Content-Type", mediaType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
		w.Header().Set("Cache-Control", "private, max-age=86400")
		_, _ = w.Write(payload)
	})
}

func serveBatchComicReference(w http.ResponseWriter, r *http.Request, store batchfactoryv11.Store, files *localartifact.Store, owner, batchID, assetID string) {
	batch, err := store.GetBatch(r.Context(), owner, batchID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	allowed := batchfactoryv11.ComicGlobalReference(batch.SettingsState.Patch).AssetID == assetID
	for _, book := range batch.Books {
		if batchfactoryv11.ComicGlobalReference(batchfactoryv11.ResolveSettings(batch.SettingsState.Patch, book.SettingsState.Patch)).AssetID == assetID {
			allowed = true
			break
		}
	}
	if !allowed || strings.TrimSpace(assetID) == "" || strings.HasPrefix(assetID, "preset:") {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "漫画参考图不存在"})
		return
	}
	file, err := files.Open(assetID + comicReferenceExtension(files, assetID))
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "漫画参考图文件不存在"})
		return
	}
	defer file.Close()
	first := make([]byte, 512)
	n, _ := file.Read(first)
	mediaType := http.DetectContentType(first[:n])
	_, _ = file.Seek(0, io.SeekStart)
	w.Header().Set("Content-Type", mediaType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, assetID, batch.UpdatedAt, file)
}

func comicReferenceExtension(files *localartifact.Store, assetID string) string {
	for _, extension := range []string{".png", ".jpg", ".webp"} {
		file, err := files.Open(assetID + extension)
		if err == nil {
			_ = file.Close()
			return extension
		}
	}
	return ""
}
