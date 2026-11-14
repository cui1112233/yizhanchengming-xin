package httpapi

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/screenplay"
	"sort"
	"strings"
	"time"
)

type ScreenplayMediaWorkbench struct {
	VoiceDesigner      screenplayVoiceDesigner
	PositioningText    func(context.Context, string, screenplay.TextRequest) (screenplay.TextResponse, error)
	Projects           screenplay.Store
	Media              screenplay.MediaStore
	Cut                aicut.Store
	Objects            agentworkspace.AssetObjectStore
	Assets             agentworkspace.AssetStore
	ReferencesProvider agentworkspace.ReferenceURLProvider
	Images             AgentImageGenerator
	Catalog            novelfetchworkshop.Store
	ASR                *aicut.ASRService
	Exports            *AICutExportService
	Authoring          *screenplay.AuthoringService
	Production         screenplay.ProductionRepository
	Executors          *localexecutor.Service
	Files              *localartifact.Store
}
type screenplayMediaAction struct {
	VoiceDesignDrafts []screenplay.VoiceDesignDraft `json:"items"`
	Rows              int                           `json:"rows"`
	Columns           int                           `json:"columns"`
	TargetIDs         []string                      `json:"targetIds"`
	IdempotencyKey    string                        `json:"idempotencyKey"`
	Action            string                        `json:"action"`
	ItemID            string                        `json:"itemId"`
	TargetID          string                        `json:"targetId"`
	VoiceID           string                        `json:"voiceId"`
	Text              string                        `json:"text"`
	Speed             float64                       `json:"speed"`
	ModelID           string                        `json:"modelId"`
	Prompt            string                        `json:"prompt"`
	Format            string                        `json:"format"`
	ReferenceItemIDs  []string                      `json:"referenceItemIds"`
	ExpectedRevision  int64                         `json:"expectedRevision"`
	Clips             []screenplayMediaClip         `json:"clips"`
	Subtitles         []screenplayMediaSubtitle     `json:"subtitles"`
}
type screenplayMediaClip struct {
	ItemID        string  `json:"itemId"`
	StartMS       int64   `json:"startMs"`
	DurationMS    int64   `json:"durationMs"`
	SourceStartMS int64   `json:"sourceStartMs"`
	Scale         float64 `json:"scale"`
	Order         int     `json:"order"`
}
type screenplayMediaSubtitle struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	StartMS    int64  `json:"startMs"`
	DurationMS int64  `json:"durationMs"`
}

func (s *ScreenplayMediaWorkbench) resolve(ctx context.Context, actor, id, permission string) (screenplay.Project, error) {
	if s.Authoring != nil {
		owner, err := s.Authoring.Resolve(ctx, actor, id, permission)
		if err != nil {
			return screenplay.Project{}, err
		}
		return s.Projects.GetProject(ctx, owner, id)
	}
	return s.Projects.GetProject(ctx, actor, id)
}
func (s *ScreenplayMediaWorkbench) ensureCut(ctx context.Context, p screenplay.Project) error {
	if s.Cut == nil {
		return fmt.Errorf("剪辑服务未配置")
	}
	_, err := s.Cut.Project(ctx, p.OwnerUsername, p.ID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, aicut.ErrProjectNotFound) {
		return err
	}
	err = s.Cut.CreateProject(ctx, p.OwnerUsername, aicut.Project{ID: p.ID, Title: p.Title, AspectRatio: p.AspectRatio, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()})
	if errors.Is(err, aicut.ErrProjectExists) {
		return nil
	}
	return err
}
func RegisterScreenplayMediaWorkbenchRoutes(mux *http.ServeMux, auth BridgeAuth, s *ScreenplayMediaWorkbench) {
	if mux == nil || s == nil || s.Projects == nil || s.Media == nil {
		return
	}
	handle := func(pattern, permission string, fn func(http.ResponseWriter, *http.Request, screenplay.Project)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, ok := BridgeIdentityFromContext(r.Context())
			if !ok {
				writeJSON(w, 401, map[string]string{"error": "unauthorized"})
				return
			}
			p, e := s.resolve(r.Context(), id.Username, r.PathValue("id"), permission)
			if e != nil {
				writeScreenplayMediaError(w, e)
				return
			}
			fn(w, r, p)
		})))
	}
	handle("GET /api/comic-screenplays/projects/{id}/media", "read", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		// Import already-produced results only; this never submits model work.
		s.syncProduced(r.Context(), p)
		state, e := s.Media.LoadMedia(r.Context(), p.OwnerUsername, p.ID)
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		for i := range state.Items {
			if state.Items[i].AssetID != "" {
				state.Items[i].URL = aicutSignedAssetContentURL(auth, p.OwnerUsername, p.ID, state.Items[i].AssetID, "content", time.Now())
				if s.Cut != nil {
					if asset, e := s.Cut.Asset(r.Context(), p.OwnerUsername, p.ID, state.Items[i].AssetID); e == nil {
						state.Items[i].Name = asset.Name
						state.Items[i].DurationMS = asset.DurationMS
					}
				}
			}
		}
		capability := func(ok bool, why string) map[string]any { return map[string]any{"available": ok, "reason": why} }
		caps := map[string]any{"split_grid": capability(s.Cut != nil && s.Objects != nil, "按明确行列拆分已保存图片"), "import": capability(s.Cut != nil && s.Objects != nil, "导入本系统媒体资料包"), "upload": capability(s.Objects != nil && s.Cut != nil, "需要素材存储"), "image": capability(s.Images != nil && s.Assets != nil && s.Objects != nil, "需要图片模型及素材存储"), "tts": capability(s.Objects != nil && s.Cut != nil, "使用当前账号配音设置"), "asr": capability(s.ASR != nil, "需要字幕识别服务"), "timeline": capability(s.Cut != nil, "需要剪辑服务"), "render": capability(s.Exports != nil, "需要剪辑导出队列"), "package": capability(s.Cut != nil && s.Objects != nil, "媒体与时间线资料包"), "retry_import": capability(s.Production != nil && s.Cut != nil && s.Objects != nil, "重新下载已生成视频，不提交生成任务"), "jianying": capability(s.Cut != nil && s.Objects != nil, "基础轨道兼容草稿，尚未验证新版剪映导入")}
		out := map[string]any{"items": state.Items, "settings": state, "capabilities": caps, "aicutProjectId": p.ID}
		caps["asset_process"] = capability(s.Images != nil && s.Assets != nil && s.Objects != nil && s.Cut != nil, "仅用于当前项目角色参考图的资产处理")
		_, voiceErr := resolveScreenplayVoiceDesignTarget(r.Context(), s.Catalog, p.OwnerUsername, screenplay.EffectiveCreativeConfig(p).VoiceDesignModel)
		voiceReason := "独立音色设计并保存试听音频，不自动绑定普通配音"
		if voiceErr != nil {
			voiceReason = voiceErr.Error()
		} else if s.VoiceDesigner == nil {
			voiceReason = "音色设计执行器未配置"
		} else if s.Objects == nil || s.Cut == nil {
			voiceReason = "试听音频素材存储未配置"
		}
		caps["voice_design"] = capability(voiceErr == nil && s.VoiceDesigner != nil && s.Objects != nil && s.Cut != nil, voiceReason)
		caps["voice_analyze"] = capability(s.PositioningText != nil, "使用原版127提示词生成并保存音色描述，不自动生成音频")
		if s.Cut != nil {
			if _, e := s.Cut.Project(r.Context(), p.OwnerUsername, p.ID); e == nil {
				tl, _ := s.Cut.Timeline(r.Context(), p.OwnerUsername, p.ID)
				out["timeline"] = tl
				clips := []screenplayMediaClip{}
				subtitles := []screenplayMediaSubtitle{}
				for _, track := range tl.Tracks {
					for _, c := range track.Clips {
						if c.Kind == aicut.ClipSubtitle {
							subtitles = append(subtitles, screenplayMediaSubtitle{ID: c.ID, Text: c.Text, StartMS: c.StartMS, DurationMS: c.DurationMS})
							continue
						}
						itemID := ""
						for _, item := range state.Items {
							if item.AssetID == c.AssetID {
								itemID = item.ID
								break
							}
						}
						if itemID == "" {
							continue
						}
						scale := 1.0
						if c.CanvasTransform != nil {
							scale = c.CanvasTransform.ScaleX
						}
						clips = append(clips, screenplayMediaClip{ItemID: itemID, StartMS: c.StartMS, DurationMS: c.DurationMS, SourceStartMS: c.SourceStartMS, Scale: scale, Order: len(clips)})
					}
				}
				out["clips"] = clips
				out["subtitles"] = subtitles
				if s.Exports != nil {
					var invalidatedAt time.Time
					if batch, e := s.Projects.GetBatch(r.Context(), p.OwnerUsername, p.ID); e == nil {
						if step := batch.Step("8_finalize"); step != nil && step.Status == screenplay.StepStale {
							invalidatedAt = batch.UpdatedAt
						}
					}
					jobs, e := s.Exports.List(r.Context(), p.OwnerUsername, p.ID)
					if e == nil {
						out["exportJobs"] = jobs
						visible := state.Items[:0]
						for _, existing := range state.Items {
							if existing.Kind != "export" {
								visible = append(visible, existing)
							}
						}
						state.Items = visible
						for _, job := range jobs {
							if job.Format != "mp4" {
								continue
							}
							item := screenplay.MediaItem{ID: job.ID, Kind: "export", Name: "MP4成片", Status: string(job.State), Error: job.Error, CreatedAt: job.CreatedAt}
							if job.State == aicut.ExportSucceeded {
								saved, saveErr := s.persistExport(r.Context(), p, job, false)
								if saveErr != nil {
									item.Status = "failed"
									item.ProviderID = job.ID
									item.Error = saveErr.Error()
								} else {
									item = saved
								}
							}
							stale := job.ExpectedRevision != tl.Revision || invalidatedAt.After(job.CreatedAt)
							if stale && item.Status != "failed" {
								item.Status = "needs_review"
								item.Error = "作品媒体或参考图已修改，请核对时间线后重新合成"
							}
							if item.Status == "succeeded" && !stale {
								item.URL = "/api/comic-screenplays/projects/" + p.ID + "/media/exports/" + job.ID
							}
							state.Items = append(state.Items, item)
						}
						out["items"] = state.Items
					}
				}
				if s.ASR != nil {
					out["asrJobs"], _ = s.ASR.ListForTimeline(r.Context(), p.OwnerUsername, p.ID, "")
				}
			}
		}
		writeJSON(w, 200, out)
	})
	handle("POST /api/comic-screenplays/projects/{id}/media/upload", "edit", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		r.Body = http.MaxBytesReader(w, r.Body, 128<<20)
		if e := r.ParseMultipartForm(8 << 20); e != nil {
			writeJSON(w, 400, map[string]string{"error": "上传文件无效或超过128MB"})
			return
		}
		defer r.MultipartForm.RemoveAll()
		f, h, e := r.FormFile("file")
		if e != nil {
			writeJSON(w, 400, map[string]string{"error": "请选择文件"})
			return
		}
		defer f.Close()
		kind := r.FormValue("kind")
		if !agentworkspace.ValidAssetKind(kind) {
			writeJSON(w, 400, map[string]string{"error": "素材类型无效"})
			return
		}
		head := make([]byte, 512)
		n, _ := io.ReadFull(f, head)
		ct := http.DetectContentType(head[:n])
		if !strings.HasPrefix(ct, kind+"/") {
			writeJSON(w, 400, map[string]string{"error": "文件内容与素材类型不符"})
			return
		}
		target := firstText(r.FormValue("targetId"), r.FormValue("shotId"), r.FormValue("assetId"))
		if e = s.validTarget(r.Context(), p, target); e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		item, e := s.saveBytes(r.Context(), p, kind, path.Base(h.Filename), ct, io.MultiReader(bytes.NewReader(head[:n]), f), target)
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		writeJSON(w, 201, map[string]any{"item": item})
	})
	handle("POST /api/comic-screenplays/projects/{id}/media/actions", "read", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		var in screenplayMediaAction
		if e := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20)).Decode(&in); e != nil {
			writeJSON(w, 400, map[string]string{"error": "请求无效"})
			return
		}
		{
			permission := "edit"
			if in.Action == "image" || in.Action == "asset_process" || in.Action == "voice_design" || in.Action == "voice_analyze" || in.Action == "tts" || in.Action == "asr" || in.Action == "render" {
				permission = "generate"
			}
			id, _ := BridgeIdentityFromContext(r.Context())
			var e error
			p, e = s.resolve(r.Context(), id.Username, p.ID, permission)
			if e != nil {
				writeScreenplayMediaError(w, e)
				return
			}
		}
		result, e := s.action(r.Context(), p, in)
		if e == nil {
			kind := ""
			switch in.Action {
			case "select", "split_grid":
				kind = "references"
			case "voice":
				kind = "voice"
			case "timeline", "subtitles":
				kind = "edit"
			}
			if kind != "" {
				e = screenplay.InvalidateMediaContent(r.Context(), s.Projects, p.OwnerUsername, p.ID, kind)
			}
		}
		if e != nil {
			writeJSON(w, 422, map[string]string{"error": e.Error()})
			return
		}
		writeJSON(w, 200, result)
	})
	handle("GET /api/comic-screenplays/projects/{id}/media/package", "read", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		if e := s.packageMedia(w, r, p); e != nil {
			writeJSON(w, 422, map[string]string{"error": e.Error()})
		}
	})
	handle("POST /api/comic-screenplays/projects/{id}/media/import", "edit", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) { s.importMedia(w, r, p) })
	handle("GET /api/comic-screenplays/projects/{id}/media/voices", "read", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		writeJSON(w, 200, s.voiceDirectory(r.Context(), p.OwnerUsername))
	})
	handle("GET /api/comic-screenplays/projects/{id}/media/exports/{jobId}", "read", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		if s.Exports == nil {
			writeJSON(w, 503, map[string]string{"error": "导出服务未配置"})
			return
		}
		job, e := s.Exports.Job(r.Context(), p.OwnerUsername, p.ID, r.PathValue("jobId"))
		if e != nil {
			writeAICutError(w, e)
			return
		}
		if job.State != aicut.ExportSucceeded || job.OutputPath == "" {
			writeJSON(w, 422, map[string]string{"error": "成片尚未就绪"})
			return
		}
		item, e := s.persistExport(r.Context(), p, job, false)
		if e != nil {
			writeJSON(w, 422, map[string]string{"error": e.Error()})
			return
		}
		asset, e := s.Cut.Asset(r.Context(), p.OwnerUsername, p.ID, item.AssetID)
		if e != nil {
			writeAICutError(w, e)
			return
		}
		body, e := s.Objects.Open(r.Context(), aicutObjectAsset(asset))
		if e != nil {
			writeJSON(w, 502, map[string]string{"error": "成片对象读取失败，请重试"})
			return
		}
		defer body.Close()
		w.Header().Set("Content-Disposition", `attachment; filename="screenplay.mp4"`)
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Length", fmt.Sprint(asset.ByteSize))
		_, _ = io.Copy(w, body)
	})
}
func (s *ScreenplayMediaWorkbench) validTarget(ctx context.Context, p screenplay.Project, target string) error {
	if target == "" || target == "narrator" {
		return nil
	}
	assets, e := s.Projects.ListAssets(ctx, p.OwnerUsername, p.ID, "")
	if e != nil {
		return e
	}
	for _, a := range assets {
		if a.ID == target {
			return nil
		}
	}
	boards, e := s.Projects.ListStoryboards(ctx, p.OwnerUsername, p.ID, "")
	if e != nil {
		return e
	}
	for _, b := range boards {
		for _, shot := range b.Shots {
			if shot.ID == target {
				return nil
			}
		}
	}
	return fmt.Errorf("目标资产或镜头不存在")
}
func (s *ScreenplayMediaWorkbench) append(ctx context.Context, p screenplay.Project, item screenplay.MediaItem) error {
	_, e := s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(st *screenplay.MediaState) error { st.Items = append(st.Items, item); return nil })
	return e
}
func (s *ScreenplayMediaWorkbench) saveBytes(ctx context.Context, p screenplay.Project, kind, name, ct string, body io.Reader, target string) (screenplay.MediaItem, error) {
	if s.Objects == nil {
		return screenplay.MediaItem{}, fmt.Errorf("素材存储未配置")
	}
	if e := s.ensureCut(ctx, p); e != nil {
		return screenplay.MediaItem{}, e
	}
	id := newAgentID("sp-media")
	ext := strings.TrimPrefix(path.Ext(name), ".")
	if ext == "" {
		ext = "bin"
	}
	key, e := agentworkspace.AgentAssetObjectKey(p.OwnerUsername, p.ID, id, ext)
	if e != nil {
		return screenplay.MediaItem{}, e
	}
	key = strings.Replace(key, "agent-assets/", "aicut-assets/", 1)
	raw, e := io.ReadAll(io.LimitReader(body, (128<<20)+1))
	if e != nil || len(raw) > 128<<20 {
		return screenplay.MediaItem{}, fmt.Errorf("素材读取失败或超过128MB")
	}
	a := aicut.Asset{ID: id, ProjectID: p.ID, Kind: kind, Name: name, ContentType: ct, ObjectKey: key, ByteSize: int64(len(raw)), CreatedAt: time.Now().UTC()}
	if e = s.Objects.Put(ctx, aicutObjectAsset(a), bytes.NewReader(raw)); e != nil {
		return screenplay.MediaItem{}, e
	}
	if kind != "image" {
		if d, e := probeAICutAssetDuration(ctx, a, s.Objects); e == nil {
			a.DurationMS = d
		}
	}
	if e = s.Cut.CreateAsset(ctx, p.OwnerUsername, a); e != nil {
		return screenplay.MediaItem{}, e
	}
	if s.Assets != nil {
		if e = s.Assets.CreateAsset(ctx, p.OwnerUsername, aicutObjectAsset(a)); e != nil {
			return screenplay.MediaItem{}, e
		}
	}
	item := screenplay.MediaItem{ID: id, AssetID: id, Name: a.Name, DurationMS: a.DurationMS, Kind: kind, Status: "succeeded", TargetID: target, CreatedAt: time.Now().UTC()}
	return item, s.append(ctx, p, item)
}
func mediaItem(st screenplay.MediaState, id string) (screenplay.MediaItem, error) {
	for _, item := range st.Items {
		if item.ID == id && item.Status == "succeeded" && item.AssetID != "" {
			return item, nil
		}
	}
	return screenplay.MediaItem{}, fmt.Errorf("所选素材不存在或未完成")
}
func (s *ScreenplayMediaWorkbench) action(ctx context.Context, p screenplay.Project, in screenplayMediaAction) (any, error) {
	st, e := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if e != nil {
		return nil, e
	}
	switch in.Action {
	case "voice_design_drafts":
		return s.saveVoiceDesignDrafts(ctx, p, in)
	case "voice_analyze":
		return s.analyzeVoices(ctx, p, in)
	case "voice_design":
		return s.designVoice(ctx, p, in)
	case "split_grid":
		return s.splitGrid(ctx, p, in, st)
	case "select":
		item, e := mediaItem(st, in.ItemID)
		if e != nil {
			return nil, e
		}
		if item.Kind != "image" {
			return nil, fmt.Errorf("参考图必须为图片")
		}
		if e = s.validTarget(ctx, p, in.TargetID); e != nil {
			return nil, e
		}
		if in.TargetID == "" {
			return nil, fmt.Errorf("请选择目标资产或镜头")
		}
		return s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(v *screenplay.MediaState) error { v.Selected[in.TargetID] = item.ID; return nil })
	case "retry_import":
		var target screenplay.MediaItem
		for _, item := range st.Items {
			if item.ID == in.ItemID {
				target = item
				break
			}
		}
		if target.Status == "failed" && target.Kind == "export" && target.ProviderID != "" && s.Exports != nil {
			job, e := s.Exports.Job(ctx, p.OwnerUsername, p.ID, target.ProviderID)
			if e != nil {
				return nil, e
			}
			return s.persistExport(ctx, p, job, true)
		}
		if target.Status != "failed" || target.Kind != "video" || target.ProviderID == "" {
			return nil, fmt.Errorf("请选择导入失败的已生成镜头")
		}
		for _, item := range st.Items {
			if item.ProviderID == target.ProviderID && item.Status == "succeeded" {
				return item, nil
			}
		}
		s.syncProduced(ctx, p, target.ProviderID)
		latest, e := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
		if e != nil {
			return nil, e
		}
		for i := len(latest.Items) - 1; i >= 0; i-- {
			item := latest.Items[i]
			if item.ProviderID == target.ProviderID {
				if item.Status == "succeeded" {
					return item, nil
				}
				return nil, fmt.Errorf("%s", item.Error)
			}
		}
		return nil, fmt.Errorf("导入仍在进行，请刷新查看")
	case "voice":
		if e = s.validTarget(ctx, p, in.TargetID); e != nil {
			return nil, e
		}
		if in.TargetID == "" || strings.TrimSpace(in.VoiceID) == "" {
			return nil, fmt.Errorf("角色和音色ID不能为空")
		}
		return s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(v *screenplay.MediaState) error { v.Voices[in.TargetID] = in.VoiceID; return nil })
	case "tts", "image", "asset_process":
		return s.generate(ctx, p, in, st)
	case "timeline":
		return s.timeline(ctx, p, in, st)
	case "subtitles":
		if e = s.ensureCut(ctx, p); e != nil {
			return nil, e
		}
		tl, e := s.Cut.Timeline(ctx, p.OwnerUsername, p.ID)
		if e != nil {
			return nil, e
		}
		tracks := []aicut.Track{}
		for _, track := range tl.Tracks {
			if track.ID != "screenplay-subtitle" {
				tracks = append(tracks, track)
			}
		}
		track := aicut.Track{ID: "screenplay-subtitle", Kind: aicut.TrackSubtitle, Name: "字幕", Order: len(tracks), Clips: []aicut.Clip{}}
		for _, sub := range in.Subtitles {
			track.Clips = append(track.Clips, aicut.Clip{ID: firstText(sub.ID, newAgentID("subtitle")), Kind: aicut.ClipSubtitle, Text: sub.Text, Name: "字幕", StartMS: sub.StartMS, DurationMS: sub.DurationMS})
		}
		tl.Tracks = append(tracks, track)
		return s.Cut.SaveTimeline(ctx, p.OwnerUsername, p.ID, tl, in.ExpectedRevision, time.Now().UTC())
	case "asr":
		if s.ASR == nil {
			return nil, fmt.Errorf("字幕识别服务未配置")
		}
		item, e := mediaItem(st, in.ItemID)
		if e != nil {
			return nil, e
		}
		if e = s.ensureCut(ctx, p); e != nil {
			return nil, e
		}
		lockName := "screenplay-media-asr:" + p.OwnerUsername + ":" + p.ID + ":" + item.AssetID
		token := newAgentID("asr-lock")
		if e = s.Projects.AcquireLock(ctx, lockName, token, time.Minute); e != nil {
			return nil, e
		}
		defer s.Projects.ReleaseLock(context.WithoutCancel(ctx), lockName, token)
		tl, e := s.Cut.Timeline(ctx, p.OwnerUsername, p.ID)
		if e != nil {
			return nil, e
		}
		if tl.Revision != in.ExpectedRevision {
			return nil, aicut.ErrRevisionConflict
		}
		return s.ASR.Submit(ctx, p.OwnerUsername, aicut.ASRSubmit{ProjectID: p.ID, AssetID: item.AssetID, ModelID: in.ModelID, ExpectedRevision: in.ExpectedRevision})
	case "render":
		if s.Exports == nil {
			return nil, fmt.Errorf("渲染服务未配置")
		}
		if e = s.ensureCut(ctx, p); e != nil {
			return nil, e
		}
		project, e := s.Cut.Project(ctx, p.OwnerUsername, p.ID)
		if e != nil {
			return nil, e
		}
		tl, e := s.Cut.Timeline(ctx, p.OwnerUsername, p.ID)
		if e != nil {
			return nil, e
		}
		if tl.Revision != in.ExpectedRevision {
			return nil, aicut.ErrRevisionConflict
		}
		format := in.Format
		if format == "" {
			format = "mp4"
		}
		if format != "mp4" {
			return nil, fmt.Errorf("漫剧成片目前仅支持MP4，请在AI-CUT使用其他格式")
		}
		return s.Exports.Submit(ctx, p.OwnerUsername, project, tl, format, aicut.ExportOptions{Resolution: "1080p", FrameRate: 30, Quality: "standard", Codec: "h264"})
	default:
		return nil, fmt.Errorf("不支持的媒体操作")
	}
}
func (s *ScreenplayMediaWorkbench) generate(ctx context.Context, p screenplay.Project, in screenplayMediaAction, st screenplay.MediaState) (any, error) {
	imageWork := in.Action == "image" || in.Action == "asset_process"
	imageSize, imageResolution := "", ""
	imageAspect := screenplay.EffectiveCreativeConfig(p).AspectRatio
	if e := s.validTarget(ctx, p, in.TargetID); e != nil {
		return nil, e
	}
	if imageWork && (s.Images == nil || s.Assets == nil || s.Objects == nil || s.Cut == nil) {
		return nil, fmt.Errorf("图片服务未配置")
	}
	if in.Action == "tts" && (strings.TrimSpace(in.Text) == "" || len([]rune(in.Text)) > 5000) {
		return nil, fmt.Errorf("配音文本须为1到5000字")
	}
	if imageWork {
		if in.Action == "asset_process" {
			if len(in.ReferenceItemIDs) == 0 {
				if selected := st.Selected[in.TargetID]; selected != "" {
					in.ReferenceItemIDs = []string{selected}
				}
			}
			if len(in.ReferenceItemIDs) == 0 {
				return nil, fmt.Errorf("%w: 资产处理须选择当前项目的参考图片", screenplay.ErrInvalidArgument)
			}
		}
		if p.CreativeConfig != nil && strings.TrimSpace(in.ModelID) == "" {
			in.ModelID = p.CreativeConfig.ImageModelID
			if len(in.ReferenceItemIDs) > 0 && in.Action != "asset_process" {
				in.ModelID = firstText(p.CreativeConfig.ImageEditModelID, in.ModelID)
			}
		}
		if strings.TrimSpace(in.ModelID) == "" {
			return nil, fmt.Errorf("请选择图片模型")
		}
		var err error
		if in.Action == "asset_process" {
			in.Prompt, err = s.creativeAssetProcessPrompt(ctx, p, in.TargetID, in.Prompt)
		} else {
			in.Prompt, err = s.creativeImagePrompt(ctx, p, in.TargetID, in.Prompt)
		}
		if err != nil {
			return nil, err
		}
		imageSize, imageResolution, err = screenplayImageDimensions(p, imageAspect)
		if err != nil {
			return nil, err
		}
	}
	refs := []string{}
	for _, id := range in.ReferenceItemIDs {
		item, e := mediaItem(st, id)
		if e != nil {
			return nil, e
		}
		if item.Kind != "image" {
			return nil, fmt.Errorf("参考素材须为图片")
		}
		refs = append(refs, item.AssetID)
	}
	attempt := screenplay.MediaItem{ID: newAgentID("sp-attempt"), Kind: in.Action, Status: "running", TargetID: in.TargetID, CreatedAt: time.Now().UTC()}
	if in.IdempotencyKey != "" {
		if len(in.IdempotencyKey) > 256 {
			return nil, fmt.Errorf("请求标识过长")
		}
		attempt.ID = fmt.Sprintf("sp-attempt-%x", sha256.Sum256([]byte(in.Action+":"+in.IdempotencyKey)))
		for _, old := range st.Items {
			if old.ID == attempt.ID {
				if old.Status == "succeeded" {
					return mediaItem(st, old.ProviderID)
				}
				return nil, fmt.Errorf("此请求已提交，当前状态：%s；%s，请勿重复提交", old.Status, old.Error)
			}
		}
	}
	if _, e := s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(v *screenplay.MediaState) error {
		for _, old := range v.Items {
			if old.ID == attempt.ID {
				return fmt.Errorf("此请求已经提交，请刷新结果")
			}
		}
		v.Items = append(v.Items, attempt)
		return nil
	}); e != nil {
		return nil, e
	}
	var result any
	var runErr error
	if in.Action == "tts" {
		voice := firstText(in.VoiceID, st.Voices[in.TargetID])
		if voice == "" {
			runErr = fmt.Errorf("请先设置角色或旁白音色ID")
		} else {
			speed := in.Speed
			if speed == 0 {
				speed = 1
			}
			if speed < .5 || speed > 2 {
				runErr = fmt.Errorf("语速须在0.5到2之间")
			} else {
				var raw []byte
				var ct string
				raw, ct, runErr = requestAICutTTSBytes(ctx, p.OwnerUsername, ttsRequest{Input: in.Text, Voice: voice, Speed: speed, Style: "general"}, s.Catalog)
				if runErr == nil {
					if !strings.HasPrefix(ct, "audio/") || len(raw) == 0 {
						runErr = fmt.Errorf("配音服务没有返回音频")
					} else {
						result, runErr = s.saveBytes(ctx, p, "audio", "配音"+aicutTTSExtension(ct), ct, bytes.NewReader(raw), in.TargetID)
					}
				}
			}
		}
	} else {
		var urls []string
		referenceURLs := []string{}
		for _, id := range refs {
			if s.ReferencesProvider == nil {
				runErr = fmt.Errorf("参考图片签名服务未配置")
				break
			}
			a, e := s.Assets.Asset(ctx, p.OwnerUsername, p.ID, id)
			if e != nil {
				runErr = e
				break
			}
			u, e := s.ReferencesProvider.ReferenceURL(ctx, a, time.Hour)
			if e != nil {
				runErr = e
				break
			}
			referenceURLs = append(referenceURLs, u)
		}
		if runErr == nil {
			urls, _, runErr = s.Images(ctx, p.OwnerUsername, AgentImageGenerationInput{ProjectID: p.ID, ModelID: in.ModelID, Prompt: in.Prompt, Count: 1, AspectRatio: imageAspect, Size: imageSize, Resolution: imageResolution, ReferenceAssetIDs: refs, ReferenceImages: referenceURLs})
			if ctx.Err() != nil {
				runErr = ctx.Err()
			}
		}
		if runErr == nil {
			if len(urls) == 0 {
				runErr = fmt.Errorf("图片服务未返回结果")
			} else {
				ing := newAgentMediaAssetIngestor(s.Assets, s.Objects, nil, time.Now)
				a, e := ing.SaveImageResult(ctx, p.OwnerUsername, p.ID, newAgentID("sp-media"), urls[0])
				runErr = e
				if e == nil {
					if e = s.ensureCut(ctx, p); e == nil {
						e = s.Cut.CreateAsset(ctx, p.OwnerUsername, aicut.Asset{ID: a.ID, ProjectID: p.ID, Kind: a.Kind, Name: a.Name, ContentType: a.ContentType, ObjectKey: a.ObjectKey, ByteSize: a.ByteSize, CreatedAt: a.CreatedAt})
					}
					runErr = e
					if e == nil {
						item := screenplay.MediaItem{ID: a.ID, AssetID: a.ID, Kind: "image", Status: "succeeded", TargetID: in.TargetID, CreatedAt: a.CreatedAt}
						runErr = s.append(ctx, p, item)
						result = item
					}
				}
			}
		}
	}
	_, saveErr := s.Media.MutateMedia(context.WithoutCancel(ctx), p.OwnerUsername, p.ID, func(v *screenplay.MediaState) error {
		for i := range v.Items {
			if v.Items[i].ID == attempt.ID {
				v.Items[i].Status = "succeeded"
				if item, ok := result.(screenplay.MediaItem); ok {
					v.Items[i].ProviderID = item.ID
				}
				if runErr != nil {
					v.Items[i].Status = "failed"
					v.Items[i].Error = runErr.Error()
				}
			}
		}
		return nil
	})
	if saveErr != nil {
		return nil, saveErr
	}
	return result, runErr
}
func (s *ScreenplayMediaWorkbench) timeline(ctx context.Context, p screenplay.Project, in screenplayMediaAction, st screenplay.MediaState) (aicut.Timeline, error) {
	if e := s.ensureCut(ctx, p); e != nil {
		return aicut.Timeline{}, e
	}
	tracks := []aicut.Track{{ID: "screenplay-video", Kind: aicut.TrackVideo, Name: "画面", Order: 0, Clips: []aicut.Clip{}}, {ID: "screenplay-audio", Kind: aicut.TrackAudio, Name: "配音", Order: 1, Clips: []aicut.Clip{}}, {ID: "screenplay-subtitle", Kind: aicut.TrackSubtitle, Name: "字幕", Order: 2, Clips: []aicut.Clip{}}}
	sort.SliceStable(in.Clips, func(i, j int) bool { return in.Clips[i].Order < in.Clips[j].Order })
	var cursor int64
	for _, c := range in.Clips {
		item, e := mediaItem(st, c.ItemID)
		if e != nil {
			return aicut.Timeline{}, e
		}
		a, e := s.Cut.Asset(ctx, p.OwnerUsername, p.ID, item.AssetID)
		if e != nil {
			return aicut.Timeline{}, e
		}
		if c.DurationMS <= 0 || c.SourceStartMS < 0 || c.StartMS < 0 {
			return aicut.Timeline{}, fmt.Errorf("镜头时长或起点无效")
		}
		if a.DurationMS > 0 && c.SourceStartMS+c.DurationMS > a.DurationMS && a.Kind != "image" {
			return aicut.Timeline{}, fmt.Errorf("裁切超出源素材时长")
		}
		scale := c.Scale
		if scale == 0 {
			scale = 1
		}
		clip := aicut.Clip{ID: newAgentID("clip"), AssetID: a.ID, Kind: aicut.ClipKind(a.Kind), Name: a.Name, StartMS: cursor, DurationMS: c.DurationMS, SourceStartMS: c.SourceStartMS, SourceDurationMS: a.DurationMS, CanvasTransform: &aicut.CanvasTransform{ScaleX: scale, ScaleY: scale, Opacity: 1}}
		idx := 0
		if a.Kind == "audio" {
			idx = 1
			clip.StartMS = c.StartMS
		} else {
			cursor += c.DurationMS
		}
		tracks[idx].Clips = append(tracks[idx].Clips, clip)
	}
	for _, sub := range in.Subtitles {
		tracks[2].Clips = append(tracks[2].Clips, aicut.Clip{ID: firstText(sub.ID, newAgentID("subtitle")), Kind: aicut.ClipSubtitle, Text: sub.Text, Name: "字幕", StartMS: sub.StartMS, DurationMS: sub.DurationMS})
	}
	return s.Cut.SaveTimeline(ctx, p.OwnerUsername, p.ID, aicut.Timeline{Tracks: tracks}, in.ExpectedRevision, time.Now().UTC())
}
func (s *ScreenplayMediaWorkbench) References(ctx context.Context, p screenplay.Project, shot screenplay.StoryboardShot) ([]string, error) {
	st, e := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if e != nil {
		return nil, e
	}
	targets := []string{shot.ID}
	assets, e := s.Projects.ListAssets(ctx, p.OwnerUsername, p.ID, "")
	if e != nil {
		return nil, e
	}
	for _, a := range assets {
		if a.Name != "" && (strings.Contains(shot.Characters, a.Name) || strings.Contains(shot.Scene, a.Name) || strings.Contains(shot.Props, a.Name)) {
			targets = append(targets, a.ID)
		}
	}
	urls := []string{}
	seen := map[string]bool{}
	for _, target := range targets {
		id := st.Selected[target]
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		item, e := mediaItem(st, id)
		if e != nil {
			return nil, e
		}
		if s.ReferencesProvider == nil || s.Cut == nil {
			return nil, fmt.Errorf("参考图片存储签名服务未配置")
		}
		a, e := s.Cut.Asset(ctx, p.OwnerUsername, p.ID, item.AssetID)
		if e != nil {
			return nil, e
		}
		u, e := s.ReferencesProvider.ReferenceURL(ctx, aicutObjectAsset(a), time.Hour)
		if e != nil {
			return nil, e
		}
		urls = append(urls, u)
	}
	return urls, nil
}
func (s *ScreenplayMediaWorkbench) packageMedia(w http.ResponseWriter, r *http.Request, p screenplay.Project) error {
	if s.Cut == nil || s.Objects == nil {
		return fmt.Errorf("媒体存储未配置")
	}
	tl, e := s.Cut.Timeline(r.Context(), p.OwnerUsername, p.ID)
	if e != nil {
		return e
	}
	assets, e := s.Cut.ListAssets(r.Context(), p.OwnerUsername, p.ID)
	if e != nil {
		return e
	}
	var buf bytes.Buffer
	z := zip.NewWriter(&buf)
	manifest := map[string]string{}
	dimensions := map[string][2]int{}
	var total int64
	for _, a := range assets {
		body, e := s.Objects.Open(r.Context(), aicutObjectAsset(a))
		if e != nil {
			return e
		}
		name := "media/" + a.ID + path.Ext(a.Name)
		entry, e := z.Create(name)
		if e != nil {
			body.Close()
			return e
		}
		var n int64
		n, e = io.Copy(entry, io.LimitReader(body, (512<<20)-total+1))
		total += n
		body.Close()
		if e != nil {
			return e
		}
		manifest[a.ID] = name
		if r.URL.Query().Get("format") == "jianying" && a.Kind != "audio" {
			size, e := probeScreenplayDimensions(r.Context(), a, s.Objects)
			if e != nil {
				return e
			}
			dimensions[a.ID] = size
		}
		if total > 512<<20 || buf.Len() > 512<<20 {
			return fmt.Errorf("资料包超过512MB，请分集导出")
		}
	}
	entry, _ := z.Create("timeline.json")
	if e = json.NewEncoder(entry).Encode(map[string]any{"project": p, "timeline": tl, "media": manifest, "format": "v88-media-package-v1", "assets": assets}); e != nil {
		return e
	}
	if r.URL.Query().Get("format") == "jianying" {
		cutProject, e := s.Cut.Project(r.Context(), p.OwnerUsername, p.ID)
		if e != nil {
			return e
		}
		draft, meta, e := screenplayJianyingDraft(cutProject, tl, assets, manifest, dimensions)
		if e != nil {
			return e
		}
		for name, content := range map[string]any{"draft_content.json": draft, "draft_meta_info.json": meta} {
			entry, e := z.Create(name)
			if e != nil {
				return e
			}
			if e = json.NewEncoder(entry).Encode(content); e != nil {
				return e
			}
		}
		entry, e := z.Create("连接素材地址.ps1")
		if e != nil {
			return e
		}
		if _, e = io.WriteString(entry, screenplayJianyingRelinkScript); e != nil {
			return e
		}
		entry, e = z.Create("使用说明.txt")
		if e != nil {
			return e
		}
		if _, e = io.WriteString(entry, "剪映兼容草稿包：含真实媒体、画面/音频轨道、裁切、缩放和字幕。\r\n解压到固定目录后，用 PowerShell 运行“连接素材地址.ps1”，连接本机绝对素材地址，再将此目录放入剪映草稿目录后打开。\r\n本版本尚未在本机剪映界面验证导入；新版剪映可能需要重新定位素材。复杂特效请使用成片导出。\r\n格式参考：https://github.com/GuanYixuan/pyJianYingDraft\r\n"); e != nil {
			return e
		}
	}
	if e = z.Close(); e != nil {
		return e
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="screenplay-media.zip"`)
	_, e = w.Write(buf.Bytes())
	return e
}

// Finalize imports the existing provider results once and creates a real,
// editable AI-CUT project. It does not trigger another paid generation.
func (s *ScreenplayMediaWorkbench) Finalize(ctx context.Context, p screenplay.Project) ([]string, error) {
	if s.Production == nil || s.Assets == nil || s.Objects == nil {
		return nil, fmt.Errorf("成片交付存储未配置")
	}
	if e := s.ensureCut(ctx, p); e != nil {
		return nil, e
	}
	tasks, e := s.Production.ListShotTasks(ctx, p.OwnerUsername, p.ID)
	if e != nil {
		return nil, e
	}
	if len(tasks) == 0 {
		return nil, screenplay.ErrClipsReady
	}
	// Old request versions stay in history; only the newest version of each
	// currently active storyboard shot may enter a deliverable.
	boards, e := s.Projects.ListStoryboards(ctx, p.OwnerUsername, p.ID, "")
	if e != nil {
		return nil, e
	}
	current := map[string]bool{}
	selectedBoards := map[int]screenplay.Storyboard{}
	for _, board := range boards {
		old, exists := selectedBoards[board.EpisodeNo]
		if !exists || (board.Engine == screenplay.EngineXiaoyunqueV3 && old.Engine != screenplay.EngineXiaoyunqueV3) {
			selectedBoards[board.EpisodeNo] = board
		}
	}
	for _, board := range selectedBoards {
		for _, shot := range board.Shots {
			current[board.ID+":"+fmt.Sprint(shot.Seq)+":"+shot.ID] = true
		}
	}
	latest := map[string]screenplay.ShotTask{}
	for _, task := range tasks {
		key := task.StoryboardID + ":" + fmt.Sprint(task.Seq)
		if !current[key+":"+task.ShotID] {
			continue
		}
		old, ok := latest[key]
		if !ok || task.UpdatedAt.After(old.UpdatedAt) {
			latest[key] = task
		}
	}
	tasks = tasks[:0]
	for _, task := range latest {
		tasks = append(tasks, task)
	}
	if len(tasks) != len(current) {
		return nil, fmt.Errorf("当前分镜仍有镜头未完成，不能交付")
	}
	if len(tasks) == 0 {
		return nil, fmt.Errorf("当前分镜没有已完成的视频")
	}
	sort.SliceStable(tasks, func(i, j int) bool {
		if tasks[i].Episode != tasks[j].Episode {
			return tasks[i].Episode < tasks[j].Episode
		}
		return tasks[i].Seq < tasks[j].Seq
	})
	st, e := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if e != nil {
		return nil, e
	}
	ing := newAgentMediaAssetIngestor(s.Assets, s.Objects, nil, time.Now)
	clips := []screenplayMediaClip{}
	currentByShot := map[string]screenplay.MediaItem{}
	for _, task := range tasks {
		if task.Status != "succeeded" || task.Artifact == "" {
			return nil, fmt.Errorf("镜头尚未完成，不能交付")
		}
		var item screenplay.MediaItem
		for _, old := range st.Items {
			if old.ProviderID == task.TaskID && old.Status == "succeeded" {
				item = old
				break
			}
		}
		if item.ID == "" {
			id := newAgentID("sp-video")
			var a agentworkspace.Asset
			const localPrefix = "/api/shuihuo-production/local-executor-artifacts/"
			if strings.HasPrefix(task.Artifact, localPrefix) {
				a, e = ing.SaveLocal(ctx, p.OwnerUsername, p.ID, id, strings.TrimPrefix(task.Artifact, localPrefix), s.Executors, s.Files)
			} else {
				a, e = ing.SaveRemote(ctx, p.OwnerUsername, p.ID, id, task.Artifact)
			}
			if e != nil {
				return nil, e
			}
			cut := aicut.Asset{ID: a.ID, ProjectID: p.ID, Kind: a.Kind, Name: a.Name, ContentType: a.ContentType, ObjectKey: a.ObjectKey, ByteSize: a.ByteSize, CreatedAt: a.CreatedAt}
			cut.DurationMS, e = probeAICutAssetDuration(ctx, cut, s.Objects)
			if e != nil || cut.DurationMS <= 0 {
				return nil, fmt.Errorf("视频时长读取失败，无法创建真实时间线")
			}
			if e = s.Cut.CreateAsset(ctx, p.OwnerUsername, cut); e != nil {
				return nil, e
			}
			item = screenplay.MediaItem{ID: a.ID, AssetID: a.ID, Kind: "video", Status: "succeeded", TargetID: task.ShotID, ShotID: task.ShotID, ProviderID: task.TaskID, CreatedAt: time.Now().UTC()}
			if e = s.append(ctx, p, item); e != nil {
				return nil, e
			}
			st.Items = append(st.Items, item)
		}
		a, e := s.Cut.Asset(ctx, p.OwnerUsername, p.ID, item.AssetID)
		if e != nil {
			return nil, e
		}
		clips = append(clips, screenplayMediaClip{ItemID: item.ID, DurationMS: a.DurationMS, Order: len(clips)})
		currentByShot[task.ShotID] = item
	}
	tl, e := s.Cut.Timeline(ctx, p.OwnerUsername, p.ID)
	if e != nil {
		return nil, e
	}
	if len(tl.Tracks) == 0 {
		if _, e = s.timeline(ctx, p, screenplayMediaAction{Clips: clips, ExpectedRevision: tl.Revision}, st); e != nil {
			return nil, e
		}
	} else {
		byAsset := map[string]screenplay.MediaItem{}
		for _, item := range st.Items {
			if item.ProviderID != "" && item.AssetID != "" {
				byAsset[item.AssetID] = item
			}
		}
		found := map[string]bool{}
		changed := false
		for ti := range tl.Tracks {
			for ci := range tl.Tracks[ti].Clips {
				clip := &tl.Tracks[ti].Clips[ci]
				old, exists := byAsset[clip.AssetID]
				if !exists {
					continue
				}
				latest, exists := currentByShot[old.ShotID]
				if !exists {
					return nil, fmt.Errorf("当前时间线包含已移除的分镜，请在媒体工作台重新应用当前镜头")
				}
				found[old.ShotID] = true
				if latest.AssetID == clip.AssetID {
					continue
				}
				a, e := s.Cut.Asset(ctx, p.OwnerUsername, p.ID, latest.AssetID)
				if e != nil {
					return nil, e
				}
				if clip.SourceStartMS+clip.DurationMS > a.DurationMS {
					return nil, fmt.Errorf("重建镜头短于已保存的裁切范围，请在媒体工作台调整时长")
				}
				clip.AssetID = a.ID
				clip.SourceDurationMS = a.DurationMS
				changed = true
			}
		}
		if len(found) != len(currentByShot) {
			return nil, fmt.Errorf("当前时间线未包含全部当前镜头，请在媒体工作台重新应用镜头")
		}
		if changed {
			tl, e = s.Cut.SaveTimeline(ctx, p.OwnerUsername, p.ID, tl, tl.Revision, time.Now().UTC())
			if e != nil {
				return nil, e
			}
		}
	}
	if s.Exports == nil {
		return nil, fmt.Errorf("镜头与剪辑工程已就绪，成片渲染服务未配置")
	}
	tl, e = s.Cut.Timeline(ctx, p.OwnerUsername, p.ID)
	if e != nil {
		return nil, e
	}
	jobs, e := s.Exports.List(ctx, p.OwnerUsername, p.ID)
	if e != nil {
		return nil, e
	}
	var job aicut.ExportJob
	for _, j := range jobs {
		if j.Format == "mp4" && j.ExpectedRevision == tl.Revision {
			job = j
			break
		}
	}
	if job.ID == "" {
		project, e := s.Cut.Project(ctx, p.OwnerUsername, p.ID)
		if e != nil {
			return nil, e
		}
		job, e = s.Exports.Submit(ctx, p.OwnerUsername, project, tl, "mp4", aicut.ExportOptions{Resolution: "1080p", FrameRate: 30, Quality: "standard", Codec: "h264"})
		if e != nil {
			return nil, e
		}
	}
	for {
		switch job.State {
		case aicut.ExportSucceeded:
			if _, e = s.persistExport(ctx, p, job, false); e != nil {
				return nil, e
			}
			return []string{"/api/comic-screenplays/projects/" + p.ID + "/media/exports/" + job.ID, "/api/comic-screenplays/projects/" + p.ID + "/media/package"}, nil
		case aicut.ExportFailed, aicut.ExportCanceled:
			return nil, fmt.Errorf("成片渲染失败：%s", job.Error)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
		job, e = s.Exports.Job(ctx, p.OwnerUsername, p.ID, job.ID)
		if e != nil {
			return nil, e
		}
	}
}

// ReferenceIDs is the stable identity used for production request deduplication;
// signed URLs are deliberately excluded because their expiry changes on resume.
func (s *ScreenplayMediaWorkbench) ReferenceIDs(ctx context.Context, p screenplay.Project, shot screenplay.StoryboardShot) ([]string, error) {
	st, e := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if e != nil {
		return nil, e
	}
	targets := []string{shot.ID}
	assets, e := s.Projects.ListAssets(ctx, p.OwnerUsername, p.ID, "")
	if e != nil {
		return nil, e
	}
	for _, a := range assets {
		if a.Name != "" && (strings.Contains(shot.Characters, a.Name) || strings.Contains(shot.Scene, a.Name) || strings.Contains(shot.Props, a.Name)) {
			targets = append(targets, a.ID)
		}
	}
	ids := []string{}
	seen := map[string]bool{}
	for _, target := range targets {
		selected := st.Selected[target]
		if selected == "" || seen[selected] {
			continue
		}
		item, e := mediaItem(st, selected)
		if e != nil {
			return nil, e
		}
		seen[selected] = true
		ids = append(ids, item.AssetID)
	}
	return ids, nil
}

// syncProduced exposes partial completed shot media before the whole batch has
// finished. Import errors are recorded once and never resubmit a provider task.
func (s *ScreenplayMediaWorkbench) syncProduced(ctx context.Context, p screenplay.Project, retryTaskIDs ...string) {
	if s.Production == nil || s.Objects == nil || s.Assets == nil || s.Cut == nil {
		return
	}
	token := newAgentID("media-sync")
	lock := "screenplay-media-sync:" + p.OwnerUsername + ":" + p.ID
	if s.Projects.AcquireLock(ctx, lock, token, 5*time.Minute) != nil {
		return
	}
	defer s.Projects.ReleaseLock(context.WithoutCancel(ctx), lock, token)
	tasks, e := s.Production.ListShotTasks(ctx, p.OwnerUsername, p.ID)
	if e != nil {
		return
	}
	st, e := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if e != nil {
		return
	}
	known := map[string]bool{}
	for _, item := range st.Items {
		if item.ProviderID != "" {
			known[item.ProviderID] = true
		}
	}
	for _, id := range retryTaskIDs {
		delete(known, id)
	}
	ing := newAgentMediaAssetIngestor(s.Assets, s.Objects, nil, time.Now)
	imported := 0
	for _, task := range tasks {
		if len(retryTaskIDs) > 0 && task.TaskID != retryTaskIDs[0] {
			continue
		}
		if task.Status != "succeeded" || task.Artifact == "" || known[task.TaskID] {
			continue
		}
		if imported >= 4 {
			break
		}
		imported++
		id := newAgentID("sp-video")
		item := screenplay.MediaItem{ID: id, Kind: "video", ShotID: task.ShotID, TargetID: task.ShotID, ProviderID: task.TaskID, Status: "failed", CreatedAt: time.Now().UTC()}
		var a agentworkspace.Asset
		const prefix = "/api/shuihuo-production/local-executor-artifacts/"
		if strings.HasPrefix(task.Artifact, prefix) {
			a, e = ing.SaveLocal(ctx, p.OwnerUsername, p.ID, id, strings.TrimPrefix(task.Artifact, prefix), s.Executors, s.Files)
		} else {
			a, e = ing.SaveRemote(ctx, p.OwnerUsername, p.ID, id, task.Artifact)
		}
		if e == nil {
			e = s.ensureCut(ctx, p)
		}
		if e == nil {
			asset := aicut.Asset{ID: a.ID, ProjectID: p.ID, Kind: a.Kind, Name: a.Name, ContentType: a.ContentType, ObjectKey: a.ObjectKey, ByteSize: a.ByteSize, CreatedAt: a.CreatedAt}
			asset.DurationMS, e = probeAICutAssetDuration(ctx, asset, s.Objects)
			if e == nil && asset.DurationMS <= 0 {
				e = fmt.Errorf("视频时长无效")
			}
			if e == nil {
				e = s.Cut.CreateAsset(ctx, p.OwnerUsername, asset)
			}
		}
		if e != nil {
			item.Error = "已生成镜头导入失败：" + e.Error()
		} else {
			item.AssetID = a.ID
			item.Status = "succeeded"
		}
		if s.append(context.WithoutCancel(ctx), p, item) != nil {
			return
		}
		known[task.TaskID] = true
	}
}

func writeScreenplayMediaError(w http.ResponseWriter, e error) {
	if errors.Is(e, screenplay.ErrAccessDenied) {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": e.Error()})
		return
	}
	writeScreenplayError(w, e)
}
