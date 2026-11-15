package httpapi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/screenplay"
)

const screenplayLibraryImageLimit = 40 << 20
const screenplayLibraryObjectProject = "screenplay-library"

type ScreenplayAssetLibrary struct {
	Repository screenplay.AssetLibraryRepository
	Media      *ScreenplayMediaWorkbench
}

type screenplayLibrarySaveInput struct {
	RequestKey             string `json:"requestKey"`
	AssetID                string `json:"assetId"`
	ExpectedAssetRevision  *int64 `json:"expectedAssetRevision"`
	Category               string `json:"category"`
	ExpectedSelectedItemID string `json:"expectedSelectedItemId"`
}

// SaveSource snapshots the chosen real image into independent library storage.
// It never invokes a model, and never mutates the source project or its assets.
func (s *ScreenplayAssetLibrary) SaveSource(ctx context.Context, actor, projectID string, in screenplayLibrarySaveInput) (screenplay.AssetLibraryEntry, error) {
	if s == nil || s.Repository == nil || s.Media == nil || s.Media.Projects == nil || s.Media.Media == nil || s.Media.Cut == nil || s.Media.Objects == nil {
		return screenplay.AssetLibraryEntry{}, fmt.Errorf("%w: 普通资产库存储未配置", screenplay.ErrInvalidState)
	}
	if !validAssetLibraryRequestKey(in.RequestKey) || strings.TrimSpace(in.AssetID) == "" || strings.TrimSpace(in.ExpectedSelectedItemID) == "" || in.ExpectedAssetRevision == nil || *in.ExpectedAssetRevision < 1 {
		return screenplay.AssetLibraryEntry{}, screenplay.ErrInvalidArgument
	}
	project, err := s.Media.resolve(ctx, actor, projectID, "read")
	if err != nil {
		return screenplay.AssetLibraryEntry{}, err
	}
	// The ordinary library is personal. Shared-project access does not grant the
	// right to copy another owner's resources into a different account's library.
	if project.OwnerUsername != actor {
		return screenplay.AssetLibraryEntry{}, screenplay.ErrAccessDenied
	}
	var result screenplay.AssetLibraryEntry
	save := func() error {
		current, e := s.Media.resolve(ctx, actor, projectID, "read")
		if e != nil {
			return e
		}
		if current.OwnerUsername != actor {
			return screenplay.ErrAccessDenied
		}
		result, e = s.saveSourceLocked(ctx, actor, current, in)
		return e
	}
	if s.Media.Authoring != nil {
		err = s.Media.Authoring.WithLock(ctx, actor, projectID, save)
	} else {
		err = save()
	}
	return result, err
}

func validAssetLibraryRequestKey(key string) bool {
	if len(key) < 8 || len(key) > 191 {
		return false
	}
	for _, c := range key {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

func (s *ScreenplayAssetLibrary) saveSourceLocked(ctx context.Context, actor string, project screenplay.Project, in screenplayLibrarySaveInput) (screenplay.AssetLibraryEntry, error) {
	sum := sha256.Sum256([]byte(actor + "\x00" + in.RequestKey))
	id := fmt.Sprintf("sp-lib-%x", sum[:16])
	previous, err := s.Repository.GetAssetLibraryEntry(ctx, actor, id)
	if err == nil {
		if previous.SourceProjectID != project.ID || previous.SourceAssetID != in.AssetID || previous.SourceMediaID != in.ExpectedSelectedItemID || previous.Snapshot.Revision != *in.ExpectedAssetRevision || previous.Category != strings.TrimSpace(in.Category) {
			return screenplay.AssetLibraryEntry{}, screenplay.ErrRevisionConflict
		}
		return previous, nil
	}
	if !errors.Is(err, screenplay.ErrNotFound) {
		return screenplay.AssetLibraryEntry{}, err
	}
	assets, err := s.Media.Projects.ListAssets(ctx, actor, project.ID, "")
	if err != nil {
		return screenplay.AssetLibraryEntry{}, err
	}
	var source *screenplay.Asset
	for i := range assets {
		if assets[i].ID == in.AssetID {
			source = &assets[i]
			break
		}
	}
	if source == nil {
		return screenplay.AssetLibraryEntry{}, screenplay.ErrNotFound
	}
	if source.Revision != *in.ExpectedAssetRevision {
		return screenplay.AssetLibraryEntry{}, screenplay.ErrRevisionConflict
	}
	state, err := s.Media.Media.LoadMedia(ctx, actor, project.ID)
	if err != nil {
		return screenplay.AssetLibraryEntry{}, err
	}
	if state.Selected[source.ID] != in.ExpectedSelectedItemID {
		return screenplay.AssetLibraryEntry{}, fmt.Errorf("%w: 当前选中图片已变化，请重新读取后保存", screenplay.ErrRevisionConflict)
	}
	selected, err := mediaItem(state, state.Selected[source.ID])
	if err != nil || selected.Kind != "image" {
		return screenplay.AssetLibraryEntry{}, fmt.Errorf("%w: 请先选择该资产已保存的图片，再保存到库", screenplay.ErrInvalidArgument)
	}
	stored, err := s.Media.Cut.Asset(ctx, actor, project.ID, selected.AssetID)
	if err != nil {
		return screenplay.AssetLibraryEntry{}, err
	}
	extensions := map[string]string{"image/png": "png", "image/jpeg": "jpg", "image/webp": "webp", "image/gif": "gif", "image/avif": "avif"}
	extension := extensions[stored.ContentType]
	if stored.Kind != "image" || extension == "" || stored.ByteSize > screenplayLibraryImageLimit {
		return screenplay.AssetLibraryEntry{}, fmt.Errorf("%w: 库图片需为不超过40MB的已保存图片", screenplay.ErrInvalidArgument)
	}
	reader, err := s.Media.Objects.Open(ctx, aicutObjectAsset(stored))
	if err != nil {
		return screenplay.AssetLibraryEntry{}, fmt.Errorf("读取来源图片失败：%w", err)
	}
	raw, readErr := io.ReadAll(io.LimitReader(reader, screenplayLibraryImageLimit+1))
	closeErr := reader.Close()
	if readErr != nil {
		return screenplay.AssetLibraryEntry{}, readErr
	}
	if closeErr != nil {
		return screenplay.AssetLibraryEntry{}, closeErr
	}
	if len(raw) == 0 || len(raw) > screenplayLibraryImageLimit {
		return screenplay.AssetLibraryEntry{}, screenplay.ErrInvalidArgument
	}
	imageID := newAgentID("sp-library-image")
	key, err := agentworkspace.AgentAssetObjectKey(actor, screenplayLibraryObjectProject, imageID, extension)
	if err != nil {
		return screenplay.AssetLibraryEntry{}, err
	}
	image := agentworkspace.Asset{ID: imageID, ProjectID: screenplayLibraryObjectProject, Kind: "image", Name: source.Name + "." + extension, ContentType: stored.ContentType, ObjectKey: key, ByteSize: int64(len(raw)), CreatedAt: time.Now().UTC()}
	if err = s.Media.Objects.Put(ctx, image, bytes.NewReader(raw)); err != nil {
		cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		_ = s.Media.Objects.Delete(cleanup, image)
		return screenplay.AssetLibraryEntry{}, err
	}
	entry := screenplay.AssetLibraryEntry{ID: id, Owner: actor, Type: source.Type, Name: source.Name, Description: firstText(source.Description, source.Identity), Category: strings.TrimSpace(in.Category), Snapshot: *source, SourceProjectID: project.ID, SourceAssetID: source.ID, SourceMediaID: selected.ID, Image: screenplay.MediaItem{ID: imageID, AssetID: imageID, Name: image.Name, Kind: "image", Status: "succeeded", CreatedAt: image.CreatedAt}, ImageObject: screenplay.AssetLibraryImageObject{ObjectKey: key, ContentType: image.ContentType, ByteSize: image.ByteSize}}
	saved, err := s.Repository.SaveAssetLibraryEntry(ctx, entry, 0)
	if err != nil {
		// A failed commit acknowledgement does not prove rollback. Reconcile
		// first; never delete an object that a durable row may already reference.
		reconcile, reconcileCancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer reconcileCancel()
		winner, e := s.Repository.GetAssetLibraryEntry(reconcile, actor, id)
		if errors.Is(e, screenplay.ErrNotFound) || e == nil && winner.ImageObject.ObjectKey != image.ObjectKey {
			_ = s.Media.Objects.Delete(reconcile, image)
		}
		if e == nil && winner.SourceProjectID == project.ID && winner.SourceAssetID == source.ID && winner.Snapshot.Revision == source.Revision && winner.SourceMediaID == selected.ID && winner.Category == entry.Category {
			return winner, nil
		}
		if e != nil && !errors.Is(e, screenplay.ErrNotFound) {
			return screenplay.AssetLibraryEntry{}, fmt.Errorf("保存结果尚未核对，请使用同一请求标识重新读取：%w", err)
		}
		return screenplay.AssetLibraryEntry{}, err
	}
	return saved, nil
}

// Explicit DTO keeps private object keys and source snapshots out of responses.
func screenplayLibraryPublic(entry screenplay.AssetLibraryEntry) map[string]any {
	image := entry.Image
	image.URL = "/api/comic-screenplays/asset-library/" + url.PathEscape(entry.ID) + "/image"
	return map[string]any{"id": entry.ID, "type": entry.Type, "name": entry.Name, "description": entry.Description, "category": entry.Category, "revision": entry.Revision, "image": image, "created_at": entry.CreatedAt, "updated_at": entry.UpdatedAt, "prompt": entry.Snapshot.Extra[screenplayLibraryPromptKey], "voice": entry.Snapshot.Extra[screenplayLibraryVoiceKey]}
}

func RegisterScreenplayAssetLibraryRoutes(mux *http.ServeMux, auth BridgeAuth, service *ScreenplayAssetLibrary) {
	if mux == nil || service == nil || service.Repository == nil {
		return
	}
	register := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor, ok := BridgeIdentityFromContext(r.Context())
			if !ok || strings.TrimSpace(actor.Username) == "" {
				writeJSON(w, 401, map[string]string{"error": "unauthorized"})
				return
			}
			w.Header().Set("Cache-Control", "no-store")
			fn(w, r, actor.Username)
		})))
	}
	registerScreenplayLibraryManager(register, service)
	register("GET /api/comic-screenplays/asset-library", func(w http.ResponseWriter, r *http.Request, actor string) {
		rows, err := service.Repository.ListAssetLibrary(r.Context(), actor, screenplay.AssetType(r.URL.Query().Get("type")), r.URL.Query().Get("q"), r.URL.Query().Get("category"))
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		entries := []map[string]any{}
		for _, entry := range rows {
			entries = append(entries, screenplayLibraryPublic(entry))
		}
		writeJSON(w, 200, map[string]any{"entries": entries})
	})
	register("POST /api/comic-screenplays/projects/{id}/asset-library", func(w http.ResponseWriter, r *http.Request, actor string) {
		var in screenplayLibrarySaveInput
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
		decoder.DisallowUnknownFields()
		var extra any
		if decoder.Decode(&in) != nil || decoder.Decode(&extra) != io.EOF {
			writeScreenplayError(w, fmt.Errorf("%w: 资产库请求格式不正确", screenplay.ErrInvalidArgument))
			return
		}
		entry, err := service.SaveSource(r.Context(), actor, r.PathValue("id"), in)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		writeJSON(w, 201, map[string]any{"entry": screenplayLibraryPublic(entry)})
	})
	register("POST /api/comic-screenplays/projects/{id}/asset-library/{entryID}/stage", func(w http.ResponseWriter, r *http.Request, actor string) {
		var in screenplayLibraryStageInput
		if !decodeAssetCheckRequest(w, r, &in, false) {
			return
		}
		item, err := service.StageImage(r.Context(), actor, r.PathValue("id"), r.PathValue("entryID"), in)
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		item.URL = aicutSignedAssetContentURL(auth, actor, r.PathValue("id"), item.AssetID, "content", time.Now())
		writeJSON(w, http.StatusOK, map[string]any{"item": item})
	})
	register("GET /api/comic-screenplays/asset-library/{entryID}/image", func(w http.ResponseWriter, r *http.Request, actor string) {
		entry, err := service.Repository.GetAssetLibraryEntry(r.Context(), actor, r.PathValue("entryID"))
		if err != nil {
			writeScreenplayError(w, err)
			return
		}
		if service.Media == nil || service.Media.Objects == nil || entry.ImageObject.ObjectKey == "" {
			writeJSON(w, 422, map[string]string{"error": "库图片尚未持久保存，请重新保存来源图片"})
			return
		}
		image := agentworkspace.Asset{ID: entry.Image.AssetID, ProjectID: screenplayLibraryObjectProject, Kind: "image", Name: entry.Image.Name, ObjectKey: entry.ImageObject.ObjectKey, ContentType: entry.ImageObject.ContentType, ByteSize: entry.ImageObject.ByteSize}
		reader, err := service.Media.Objects.Open(r.Context(), image)
		if err != nil {
			writeJSON(w, 502, map[string]string{"error": "库图片读取失败，请重试"})
			return
		}
		defer reader.Close()
		w.Header().Set("Content-Type", image.ContentType)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
		w.Header().Set("Cache-Control", "private, max-age=60")
		_, _ = io.Copy(w, reader)
	})
}
