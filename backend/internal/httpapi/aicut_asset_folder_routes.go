package httpapi

import (
	"errors"
	"net/http"
	"qiantie/backend/internal/aicut"
	"strings"
	"time"
)

func aicutAssetFolders(store aicut.Store) aicut.AssetFolderStore {
	switch wrapped := store.(type) {
	case aicutRequestTimelineStore:
		return aicutAssetFolders(wrapped.Store)
	case aicutExportTimelineStore:
		return aicutAssetFolders(wrapped.Store)
	}
	folders, _ := store.(aicut.AssetFolderStore)
	return folders
}
func writeAICutAssetFolderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, aicut.ErrAssetFolderNotFound):
		writeJSON(w, 404, map[string]string{"error": "素材文件夹不存在"})
	case errors.Is(err, aicut.ErrInvalidAssetFolder):
		writeJSON(w, 400, map[string]string{"error": "请输入1到100字的文件夹名称"})
	default:
		writeAICutError(w, err)
	}
}
func RegisterAICutAssetFolderRoutes(mux *http.ServeMux, auth BridgeAuth, store aicut.Store) {
	folders := aicutAssetFolders(store)
	if mux == nil || folders == nil {
		return
	}
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := BridgeIdentityFromContext(r.Context())
			if !ok || strings.TrimSpace(identity.Username) == "" {
				writeJSON(w, 401, map[string]string{"error": "unauthorized"})
				return
			}
			fn(w, r, strings.TrimSpace(identity.Username))
		})))
	}
	base := "/api/ai-cut/projects/{id}/asset-folders"
	handle("GET "+base, func(w http.ResponseWriter, r *http.Request, o string) {
		items, err := folders.ListAssetFolders(r.Context(), o, r.PathValue("id"))
		if err != nil {
			writeAICutAssetFolderError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"folders": items})
	})
	handle("POST "+base, func(w http.ResponseWriter, r *http.Request, o string) {
		var in struct {
			Name string `json:"name"`
		}
		if !decodeAICutJSON(w, r, &in) {
			return
		}
		f := aicut.AssetFolder{ID: newAgentID("folder"), ProjectID: r.PathValue("id"), Name: strings.TrimSpace(in.Name), CreatedAt: time.Now().UTC()}
		if err := folders.CreateAssetFolder(r.Context(), o, f); err != nil {
			writeAICutAssetFolderError(w, err)
			return
		}
		writeJSON(w, 201, map[string]any{"folder": f})
	})
	handle("PATCH "+base+"/{folderID}", func(w http.ResponseWriter, r *http.Request, o string) {
		var in struct {
			Name string `json:"name"`
		}
		if !decodeAICutJSON(w, r, &in) {
			return
		}
		if err := folders.RenameAssetFolder(r.Context(), o, r.PathValue("id"), r.PathValue("folderID"), in.Name); err != nil {
			writeAICutAssetFolderError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true})
	})
	handle("DELETE "+base+"/{folderID}", func(w http.ResponseWriter, r *http.Request, o string) {
		if err := folders.DeleteAssetFolder(r.Context(), o, r.PathValue("id"), r.PathValue("folderID")); err != nil {
			writeAICutAssetFolderError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true})
	})
	handle("PATCH /api/ai-cut/projects/{id}/assets/{assetID}/folder", func(w http.ResponseWriter, r *http.Request, o string) {
		var in struct {
			FolderID *string `json:"folderId"`
		}
		if !decodeAICutJSON(w, r, &in) {
			return
		}
		if in.FolderID == nil {
			writeJSON(w, 400, map[string]string{"error": "请选择目标文件夹"})
			return
		}
		if err := folders.MoveAssetToFolder(r.Context(), o, r.PathValue("id"), r.PathValue("assetID"), *in.FolderID); err != nil {
			writeAICutAssetFolderError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"assetId": r.PathValue("assetID"), "folderId": *in.FolderID})
	})
}
