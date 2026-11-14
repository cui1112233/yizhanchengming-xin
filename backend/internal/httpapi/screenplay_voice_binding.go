package httpapi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/screenplay"
	"strconv"
	"strings"
	"time"
)

func (s *ScreenplayMediaWorkbench) voiceBindingTarget(ctx context.Context, p screenplay.Project, target string, st screenplay.MediaState) error {
	if target == "" {
		return screenplay.ErrInvalidArgument
	}
	if st.VoiceDesign != nil {
		for _, draft := range st.VoiceDesign.Items {
			if target == "voice-draft:"+strconv.Itoa(draft.ID) {
				return nil
			}
		}
	}
	assets, e := s.Projects.ListAssets(ctx, p.OwnerUsername, p.ID, "")
	if e != nil {
		return e
	}
	for _, a := range assets {
		if a.ID == target && a.Type == screenplay.AssetCharacter {
			return nil
		}
	}
	return fmt.Errorf("%w: 请选择当前作品角色或已保存音色条目", screenplay.ErrInvalidArgument)
}
func voiceAudioBinding(item screenplay.MediaItem, source string) screenplay.VoiceBinding {
	binding := screenplay.VoiceBinding{MediaItemID: item.ID, Engine: "custom", ReferenceOnly: true, Source: source}
	if item.VoiceID != "" && (item.VoiceEngine == "minimax" || item.VoiceEngine == "qwen-tts") {
		binding.VoiceID = item.VoiceID
		binding.Engine = item.VoiceEngine
		binding.ReferenceOnly = false
		binding.LibraryVoiceName = item.Name
	}
	return binding
}
func (s *ScreenplayMediaWorkbench) bindVoiceAudio(ctx context.Context, p screenplay.Project, in screenplayMediaAction) (any, error) {
	if s.Cut == nil || s.Objects == nil {
		return nil, fmt.Errorf("音频素材存储未配置")
	}
	st, e := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if e != nil {
		return nil, e
	}
	if st.Revision != in.ExpectedRevision {
		return nil, screenplay.ErrRevisionConflict
	}
	if e = s.voiceBindingTarget(ctx, p, in.TargetID, st); e != nil {
		return nil, e
	}
	sourceID := firstText(in.SourceProjectID, p.ID)
	if sourceID != p.ID {
		actor, ok := BridgeIdentityFromContext(ctx)
		if !ok || actor.Username != p.OwnerUsername {
			return nil, screenplay.ErrAccessDenied
		}
		if _, e = s.Projects.GetProject(ctx, p.OwnerUsername, sourceID); e != nil {
			return nil, e
		}
		source, e := s.Media.LoadMedia(ctx, p.OwnerUsername, sourceID)
		if e != nil {
			return nil, e
		}
		item, e := mediaItem(source, in.ItemID)
		if e != nil || item.Kind != "audio" || voiceAudioBinding(item, "library").ReferenceOnly {
			return nil, fmt.Errorf("%w: 音色库只能选择当前账号真实已设计音色", screenplay.ErrInvalidArgument)
		}
		if s.Cut == nil || s.Objects == nil {
			return nil, fmt.Errorf("音频存储服务未配置")
		}
		asset, e := s.Cut.Asset(ctx, p.OwnerUsername, sourceID, item.AssetID)
		if e != nil {
			return nil, e
		}
		reader, e := s.Objects.Open(ctx, aicutObjectAsset(asset))
		if e != nil {
			return nil, e
		}
		defer reader.Close()
		saved, binding, e := s.storeVoiceBindingAudio(ctx, p, in.TargetID, in.ExpectedRevision, item.Name, asset.ContentType, reader, item, "library")
		return map[string]any{"item": saved, "binding": binding}, e
	}
	item, e := mediaItem(st, in.ItemID)
	if e != nil || item.Kind != "audio" {
		return nil, fmt.Errorf("%w: 请选择本作品已保存的音频", screenplay.ErrInvalidArgument)
	}
	asset, e := s.Cut.Asset(ctx, p.OwnerUsername, p.ID, item.AssetID)
	if e != nil {
		return nil, e
	}
	if asset.Kind != "audio" || !strings.HasPrefix(asset.ContentType, "audio/") {
		return nil, fmt.Errorf("%w: 目标音频记录无效", screenplay.ErrInvalidArgument)
	}
	binding := voiceAudioBinding(item, "library")
	if binding.ReferenceOnly {
		binding.Source = "upload"
	}
	_, e = s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(st *screenplay.MediaState) error {
		if st.Revision != in.ExpectedRevision {
			return screenplay.ErrRevisionConflict
		}
		if st.VoiceBindings == nil {
			st.VoiceBindings = map[string]screenplay.VoiceBinding{}
		}
		st.VoiceBindings[in.TargetID] = binding
		return nil
	})
	return map[string]any{"item": item, "binding": binding}, e
}

// Object storage and asset records are prepared first, then media and binding commit under one revision.
func (s *ScreenplayMediaWorkbench) storeVoiceBindingAudio(ctx context.Context, p screenplay.Project, target string, expected int64, name, contentType string, body io.Reader, metadata screenplay.MediaItem, source string) (screenplay.MediaItem, screenplay.VoiceBinding, error) {
	var zero screenplay.MediaItem
	var empty screenplay.VoiceBinding
	if s.Objects == nil || s.Cut == nil {
		return zero, empty, fmt.Errorf("音频存储服务未配置")
	}
	if !strings.HasPrefix(contentType, "audio/") {
		return zero, empty, screenplay.ErrInvalidArgument
	}
	raw, e := io.ReadAll(io.LimitReader(body, (32<<20)+1))
	if e != nil || len(raw) == 0 || len(raw) > 32<<20 {
		return zero, empty, fmt.Errorf("音频为空、无法读取或超过32MB")
	}
	if e = s.ensureCut(ctx, p); e != nil {
		return zero, empty, e
	}
	id := newAgentID("sp-voice-audio")
	ext := strings.TrimPrefix(aicutTTSExtension(contentType), ".")
	key, e := agentworkspace.AgentAssetObjectKey(p.OwnerUsername, p.ID, id, ext)
	if e != nil {
		return zero, empty, e
	}
	key = strings.Replace(key, "agent-assets/", "aicut-assets/", 1)
	asset := aicut.Asset{ID: id, ProjectID: p.ID, Kind: "audio", Name: path.Base(name), ContentType: contentType, ObjectKey: key, ByteSize: int64(len(raw)), CreatedAt: time.Now().UTC()}
	if e = s.Objects.Put(ctx, aicutObjectAsset(asset), bytes.NewReader(raw)); e != nil {
		return zero, empty, e
	}
	if e = s.Cut.CreateAsset(ctx, p.OwnerUsername, asset); e != nil {
		return zero, empty, e
	}
	if s.Assets != nil {
		if e = s.Assets.CreateAsset(ctx, p.OwnerUsername, aicutObjectAsset(asset)); e != nil {
			return zero, empty, e
		}
	}
	item := screenplay.MediaItem{ID: id, AssetID: id, Name: asset.Name, Kind: "audio", Status: "succeeded", TargetID: target, CreatedAt: asset.CreatedAt, VoiceID: metadata.VoiceID, VoiceEngine: metadata.VoiceEngine, VoiceTargetModel: metadata.VoiceTargetModel, VoiceDescription: metadata.VoiceDescription, VoicePreviewText: metadata.VoicePreviewText}
	binding := voiceAudioBinding(item, source)
	if binding.ReferenceOnly {
		item.VoiceID = ""
		item.VoiceEngine = "custom"
	}
	_, e = s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(st *screenplay.MediaState) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if st.Revision != expected {
			return screenplay.ErrRevisionConflict
		}
		st.Items = append(st.Items, item)
		if st.VoiceBindings == nil {
			st.VoiceBindings = map[string]screenplay.VoiceBinding{}
		}
		st.VoiceBindings[target] = binding
		return nil
	})
	return item, binding, e
}

func RegisterScreenplayVoiceBindingRoutes(mux *http.ServeMux, auth BridgeAuth, s *ScreenplayMediaWorkbench) {
	if mux == nil || s == nil || s.Projects == nil || s.Media == nil {
		return
	}
	mux.Handle("GET /api/screenplays/voice-library", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor, ok := BridgeIdentityFromContext(r.Context())
		if !ok {
			writeJSON(w, 401, map[string]string{"error": "unauthorized"})
			return
		}
		if s.Cut == nil || s.Objects == nil {
			writeJSON(w, 200, map[string]any{"items": []any{}, "capabilities": map[string]any{"available": false, "reason": "音频素材存储未配置"}})
			return
		}
		projects, e := s.Projects.ListProjects(r.Context(), actor.Username)
		if e != nil {
			writeScreenplayError(w, e)
			return
		}
		items := []map[string]any{}
		for _, p := range projects {
			st, e := s.Media.LoadMedia(r.Context(), actor.Username, p.ID)
			if e != nil {
				writeScreenplayMediaError(w, e)
				return
			}
			for _, item := range st.Items {
				if item.Kind != "audio" || item.Status != "succeeded" || item.AssetID == "" || voiceAudioBinding(item, "library").ReferenceOnly {
					continue
				}
				asset, e := s.Cut.Asset(r.Context(), actor.Username, p.ID, item.AssetID)
				if e != nil || asset.Kind != "audio" {
					continue
				}
				items = append(items, map[string]any{"project_id": p.ID, "project_title": p.Title, "media_item_id": item.ID, "name": item.Name, "voice_id": item.VoiceID, "engine": item.VoiceEngine, "target_model": item.VoiceTargetModel, "url": aicutSignedAssetContentURL(auth, actor.Username, p.ID, item.AssetID, "content", time.Now())})
			}
		}
		writeJSON(w, 200, map[string]any{"items": items})
	})))
	mux.Handle("POST /api/comic-screenplays/projects/{id}/media/voice-upload", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor, ok := BridgeIdentityFromContext(r.Context())
		if !ok {
			writeJSON(w, 401, map[string]string{"error": "unauthorized"})
			return
		}
		p, e := s.resolve(r.Context(), actor.Username, r.PathValue("id"), "edit")
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 34<<20)
		if e = r.ParseMultipartForm(4 << 20); e != nil {
			writeJSON(w, 400, map[string]string{"error": "音频上传请求无效或过大"})
			return
		}
		if r.MultipartForm != nil {
			defer r.MultipartForm.RemoveAll()
		}
		expected, e := strconv.ParseInt(r.FormValue("expectedRevision"), 10, 64)
		if e != nil {
			writeJSON(w, 400, map[string]string{"error": "请传入当前媒体版本"})
			return
		}
		st, e := s.Media.LoadMedia(r.Context(), p.OwnerUsername, p.ID)
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		target := r.FormValue("targetId")
		if st.Revision != expected {
			writeScreenplayMediaError(w, screenplay.ErrRevisionConflict)
			return
		}
		if e = s.voiceBindingTarget(r.Context(), p, target, st); e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		file, header, e := r.FormFile("file")
		if e != nil {
			writeJSON(w, 400, map[string]string{"error": "请选择WAV或MP3音频"})
			return
		}
		defer file.Close()
		ext := strings.ToLower(path.Ext(header.Filename))
		if ext != ".mp3" && ext != ".wav" {
			writeJSON(w, 400, map[string]string{"error": "只支持WAV或MP3音频"})
			return
		}
		head := make([]byte, 512)
		n, _ := io.ReadFull(file, head)
		ct := http.DetectContentType(head[:n])
		if ct != "audio/mpeg" && ct != "audio/wave" && ct != "audio/wav" && ct != "audio/x-wav" {
			writeJSON(w, 400, map[string]string{"error": "文件内容不是有效WAV或MP3音频"})
			return
		}
		item, binding, e := s.storeVoiceBindingAudio(r.Context(), p, target, expected, header.Filename, ct, io.MultiReader(bytes.NewReader(head[:n]), file), screenplay.MediaItem{}, "upload")
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		item.URL = aicutSignedAssetContentURL(auth, p.OwnerUsername, p.ID, item.AssetID, "content", time.Now())
		writeJSON(w, 201, map[string]any{"item": item, "binding": binding})
	})))
}
