package httpapi

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/screenplay"
	"strings"
	"time"
)

type positioningRequest struct {
	screenplay.PositioningState
	ExpectedRevision  *int64 `json:"expected_revision"`
	IdempotencyKey    string `json:"idempotencyKey"`
	MarkedImageItemID string `json:"marked_image_item_id"`
	AllowMissingRefs  bool   `json:"allow_missing_refs"`
	Model             string `json:"model"`
	AnalyzeMode       string `json:"analyze_mode"`
	ShowNames         *bool  `json:"show_character_names"`
	EditPrompt        string `json:"edit_prompt"`
	MediaItemID       string `json:"media_item_id"`
}

func positioningState(st screenplay.MediaState, p screenplay.Project) screenplay.PositioningState {
	if st.Positioning != nil {
		v := *st.Positioning
		v.PrefixText = firstText(v.PrefixText, prompt.ScreenplayPositioningPrefix)
		if v.ImageModel == "" && p.CreativeConfig != nil {
			v.ImageModel = firstText(p.CreativeConfig.ImageEditModelID, p.CreativeConfig.ImageModelID)
		}
		return v
	}
	model := ""
	if p.CreativeConfig != nil {
		model = firstText(p.CreativeConfig.ImageEditModelID, p.CreativeConfig.ImageModelID)
	}
	return screenplay.PositioningState{Items: []screenplay.PositioningItem{}, PrefixText: prompt.ScreenplayPositioningPrefix, ImageModel: model, ShowCharacterNames: true}
}

func positioningExpected(in positioningRequest, st screenplay.PositioningState) error {
	if in.ExpectedRevision == nil {
		return fmt.Errorf("%w: 请传入expected_revision", screenplay.ErrInvalidArgument)
	}
	if *in.ExpectedRevision != st.Revision {
		return screenplay.ErrRevisionConflict
	}
	return nil
}

func RegisterScreenplayPositioningRoutes(mux *http.ServeMux, auth BridgeAuth, s *ScreenplayMediaWorkbench) {
	if mux == nil || s == nil || s.Projects == nil || s.Media == nil {
		return
	}
	handle := func(pattern, permission string, fn func(http.ResponseWriter, *http.Request, screenplay.Project)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			actor, ok := BridgeIdentityFromContext(r.Context())
			if !ok {
				writeJSON(w, 401, map[string]string{"error": "unauthorized"})
				return
			}
			p, e := s.resolve(r.Context(), actor.Username, r.PathValue("id"), permission)
			if e != nil {
				writeScreenplayMediaError(w, e)
				return
			}
			fn(w, r, p)
		})))
	}
	output := func(ctx context.Context, p screenplay.Project) (map[string]any, error) {
		st, e := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
		if e != nil {
			return nil, e
		}
		v := positioningState(st, p)
		for i := range v.Items {
			if v.Items[i].ImageMediaItemID != "" {
				if m, e := mediaItem(st, v.Items[i].ImageMediaItemID); e == nil && m.AssetID != "" {
					v.Items[i].ImagePath = aicutSignedAssetContentURL(auth, p.OwnerUsername, p.ID, m.AssetID, "content", time.Now())
				}
			}
		}
		return map[string]any{"state": v, "capabilities": map[string]bool{"image": s.Images != nil && s.Assets != nil && s.Objects != nil && s.Cut != nil, "analyze": s.PositioningText != nil}}, nil
	}
	handle("GET /api/comic-screenplays/projects/{id}/positioning", "read", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		out, e := output(r.Context(), p)
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		writeJSON(w, 200, out)
	})
	handle("PUT /api/comic-screenplays/projects/{id}/positioning", "edit", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		var in positioningRequest
		if e := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<20)).Decode(&in); e != nil {
			writeJSON(w, 400, map[string]string{"error": "站位数据格式错误"})
			return
		}
		if in.ShowNames != nil {
			in.PositioningState.ShowCharacterNames = *in.ShowNames
		}
		if e := screenplay.ValidatePositioning(&in.PositioningState); e != nil {
			writeJSON(w, 400, map[string]string{"error": e.Error()})
			return
		}
		_, e := s.Media.MutateMedia(r.Context(), p.OwnerUsername, p.ID, func(st *screenplay.MediaState) error {
			old := positioningState(*st, p)
			if e := positioningExpected(in, old); e != nil {
				return e
			}
			// Image ownership and request status are server controlled. Editing a running row cancels its late result.
			for i := range in.Items {
				item := &in.Items[i]
				item.ImageMediaItemID = ""
				item.Status = ""
				item.Error = ""
				item.RequestKey = ""
				for _, o := range old.Items {
					if o.ID == item.ID {
						item.ImageMediaItemID = o.ImageMediaItemID
						if o.Status != "running" {
							item.Status = o.Status
							item.Error = o.Error
							item.RequestKey = o.RequestKey
						}
						break
					}
				}
			}
			in.Revision = old.Revision + 1
			st.Positioning = &in.PositioningState
			return nil
		})
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		out, e := output(r.Context(), p)
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		writeJSON(w, 200, out)
	})
	handle("POST /api/comic-screenplays/projects/{id}/positioning/{itemID}/image", "generate", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		var in positioningRequest
		if e := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); e != nil {
			writeJSON(w, 400, map[string]string{"error": "站位生图参数错误"})
			return
		}
		item, missing, e := s.generatePositioning(r.Context(), p, r.PathValue("itemID"), in)
		if e != nil {
			if errors.Is(e, context.Canceled) {
				writeJSON(w, 499, map[string]any{"error": "站位生图已取消", "cancelled": true})
				return
			}
			if len(missing) > 0 {
				writeJSON(w, 409, map[string]any{"error": e.Error(), "missing_refs": missing})
				return
			}
			writeScreenplayMediaError(w, e)
			return
		}
		out, e := output(r.Context(), p)
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		state := out["state"].(screenplay.PositioningState)
		for _, v := range state.Items {
			if v.ID == item.ID {
				item = v
				break
			}
		}
		out["item"] = item
		out["missing_refs"] = missing
		writeJSON(w, 200, out)
	})
	handle("POST /api/comic-screenplays/projects/{id}/positioning/analyze", "generate", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		var in positioningRequest
		if e := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); e != nil {
			writeJSON(w, 400, map[string]string{"error": "站位分析参数错误"})
			return
		}
		if e := s.analyzePositioning(r.Context(), p, in); e != nil {
			if errors.Is(e, context.Canceled) {
				writeJSON(w, 499, map[string]any{"error": "站位分析已取消", "cancelled": true})
				return
			}
			writeScreenplayMediaError(w, e)
			return
		}
		out, e := output(r.Context(), p)
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		out["cancelled"] = false
		writeJSON(w, 200, out)
	})
	handle("POST /api/comic-screenplays/projects/{id}/positioning/{itemID}/image/upload", "edit", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		var in positioningRequest
		if e := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); e != nil {
			writeJSON(w, 400, map[string]string{"error": "图片绑定参数错误"})
			return
		}
		e := s.bindPositioningImage(r.Context(), p, r.PathValue("itemID"), in)
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		out, e := output(r.Context(), p)
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		for _, item := range out["state"].(screenplay.PositioningState).Items {
			if item.ID == r.PathValue("itemID") {
				out["item"] = item
				break
			}
		}
		writeJSON(w, 200, out)
	})
	handle("POST /api/comic-screenplays/projects/{id}/positioning/{itemID}/image/edit", "generate", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		var in positioningRequest
		if e := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&in); e != nil {
			writeJSON(w, 400, map[string]string{"error": "站位改图参数错误"})
			return
		}
		if strings.TrimSpace(in.EditPrompt) == "" {
			writeJSON(w, 400, map[string]string{"error": "请填写改图要求"})
			return
		}
		item, missing, e := s.generatePositioning(r.Context(), p, r.PathValue("itemID"), in)
		if e != nil {
			if errors.Is(e, context.Canceled) {
				writeJSON(w, 499, map[string]any{"error": "站位改图已取消", "cancelled": true})
				return
			}
			if len(missing) > 0 {
				writeJSON(w, 409, map[string]any{"error": e.Error(), "missing_refs": missing})
				return
			}
			writeScreenplayMediaError(w, e)
			return
		}
		out, e := output(r.Context(), p)
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		for _, v := range out["state"].(screenplay.PositioningState).Items {
			if v.ID == item.ID {
				item = v
				break
			}
		}
		out["item"] = item
		out["missing_refs"] = missing
		writeJSON(w, 200, out)
	})
}

func (s *ScreenplayMediaWorkbench) bindPositioningImage(ctx context.Context, p screenplay.Project, id string, in positioningRequest) error {
	_, e := s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(st *screenplay.MediaState) error {
		v := positioningState(*st, p)
		if e := positioningExpected(in, v); e != nil {
			return e
		}
		m, e := mediaItem(*st, in.MediaItemID)
		if e != nil {
			return fmt.Errorf("%w: 上传的图片不属于当前项目或未保存", screenplay.ErrInvalidArgument)
		}
		if m.Kind != "image" {
			return fmt.Errorf("%w: 请选择已上传的图片", screenplay.ErrInvalidArgument)
		}
		for i := range v.Items {
			if v.Items[i].ID == id {
				v.Items[i].ImageMediaItemID = m.ID
				v.Items[i].ImagePath = ""
				v.Items[i].Status = "succeeded"
				v.Items[i].Error = ""
				v.Items[i].RequestKey = ""
				v.Revision++
				st.Positioning = &v
				return nil
			}
		}
		return screenplay.ErrNotFound
	})
	return e
}

func (s *ScreenplayMediaWorkbench) positioningReferences(ctx context.Context, p screenplay.Project, st screenplay.MediaState, item screenplay.PositioningItem, marked string) ([]string, []string, error) {
	refs := []string{}
	missing := []string{}
	seen := map[string]bool{}
	add := func(m screenplay.MediaItem) error {
		if m.Kind != "image" || m.Status != "succeeded" || m.AssetID == "" {
			return fmt.Errorf("参考素材必须是已保存的图片")
		}
		if !seen[m.AssetID] {
			seen[m.AssetID] = true
			refs = append(refs, m.AssetID)
		}
		return nil
	}
	if marked != "" {
		m, e := mediaItem(st, marked)
		if e != nil {
			return nil, nil, e
		}
		if e = add(m); e != nil {
			return nil, nil, e
		}
	}
	assets, e := s.Projects.ListAssets(ctx, p.OwnerUsername, p.ID, "")
	if e != nil {
		return nil, nil, e
	}
	names := append([]string{}, item.Characters...)
	types := make([]screenplay.AssetType, len(names))
	for i := range types {
		types[i] = screenplay.AssetCharacter
	}
	if strings.TrimSpace(item.SceneName) != "" {
		names = append(names, item.SceneName)
		types = append(types, screenplay.AssetScene)
	}
	for i, name := range names {
		var latest screenplay.MediaItem
		for _, a := range assets {
			if a.Type != types[i] || a.Name != name {
				continue
			}
			if selected := st.Selected[a.ID]; selected != "" {
				bound, err := mediaItem(st, selected)
				if err != nil {
					return nil, nil, fmt.Errorf("%w: %s选定的参考图片已失效", screenplay.ErrInvalidArgument, name)
				}
				if bound.Kind != "image" {
					return nil, nil, fmt.Errorf("%w: %s选定的参考素材不是图片", screenplay.ErrInvalidArgument, name)
				}
				latest = bound
				break
			}
			for _, m := range st.Items {
				if m.TargetID == a.ID && m.Kind == "image" && m.Status == "succeeded" && m.AssetID != "" && (latest.ID == "" || !m.CreatedAt.Before(latest.CreatedAt)) {
					latest = m
				}
			}
		}
		if latest.ID == "" {
			missing = append(missing, name)
			continue
		}
		if e = add(latest); e != nil {
			return nil, nil, e
		}
	}
	return refs, missing, nil
}

func (s *ScreenplayMediaWorkbench) generatePositioning(ctx context.Context, p screenplay.Project, id string, in positioningRequest) (screenplay.PositioningItem, []string, error) {
	var zero screenplay.PositioningItem
	if s.Images == nil || s.Assets == nil || s.Objects == nil || s.Cut == nil {
		return zero, nil, fmt.Errorf("%w: 图片服务未配置", screenplay.ErrInvalidArgument)
	}
	if strings.TrimSpace(in.IdempotencyKey) == "" || len(in.IdempotencyKey) > 256 {
		return zero, nil, fmt.Errorf("%w: 请提供1到256字符请求标识", screenplay.ErrInvalidArgument)
	}
	st, e := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if e != nil {
		return zero, nil, e
	}
	v := positioningState(st, p)
	purpose := "generate"
	if strings.TrimSpace(in.EditPrompt) != "" {
		purpose = "edit"
	}
	key := fmt.Sprintf("positioning-%x", sha256.Sum256([]byte(purpose+":"+id+":"+in.IdempotencyKey)))
	var item screenplay.PositioningItem
	found := false
	for _, it := range v.Items {
		if it.ID == id {
			item = it
			found = true
			break
		}
	}
	if !found {
		return zero, nil, screenplay.ErrNotFound
	}
	if item.RequestKey == key {
		if item.Status == "succeeded" {
			return item, []string{}, nil
		}
		return zero, nil, fmt.Errorf("%w: 此请求已经提交，状态：%s", screenplay.ErrInvalidState, item.Status)
	}
	if e = positioningExpected(in, v); e != nil {
		return zero, nil, e
	}
	if item.Status == "running" {
		return zero, nil, fmt.Errorf("%w: 此站位正在生成", screenplay.ErrInvalidState)
	}
	if strings.TrimSpace(item.PositionDetail) == "" {
		return zero, nil, fmt.Errorf("%w: 请先填写站位描述", screenplay.ErrInvalidArgument)
	}
	if purpose == "edit" {
		if len([]rune(in.EditPrompt)) > 10000 {
			return zero, nil, fmt.Errorf("%w: 改图要求过长", screenplay.ErrInvalidArgument)
		}
		if in.MarkedImageItemID == "" {
			in.MarkedImageItemID = item.ImageMediaItemID
		}
		if in.MarkedImageItemID == "" {
			return zero, nil, fmt.Errorf("%w: 请先上传或生成站位图片", screenplay.ErrInvalidArgument)
		}
	}
	refs, missing, e := s.positioningReferences(ctx, p, st, item, in.MarkedImageItemID)
	if e != nil {
		return zero, nil, e
	}
	if len(missing) > 0 && !in.AllowMissingRefs {
		return zero, missing, fmt.Errorf("缺少参考图片：%s。确认继续后将只使用已有参考图", strings.Join(missing, "、"))
	}
	model := firstText(in.ImageModel, v.ImageModel)
	if model == "" && p.CreativeConfig != nil {
		model = firstText(p.CreativeConfig.ImageEditModelID, p.CreativeConfig.ImageModelID)
	}
	if model == "" {
		return zero, nil, fmt.Errorf("%w: 请选择图片模型", screenplay.ErrInvalidArgument)
	}
	prefix := firstText(in.PrefixText, v.PrefixText, prompt.ScreenplayPositioningPrefix)
	show := v.ShowCharacterNames
	if in.ShowNames != nil {
		show = *in.ShowNames
	}
	text := prefix + "\n场景：" + item.SceneName + "\n角色（与参考图顺序一致）：" + strings.Join(item.Characters, "、") + "\n站位：" + item.PositionDetail
	text = screenplayCorePrefix(p, text)
	text += "\n备注：" + item.Remark
	if purpose == "edit" {
		text += "\n本次为站位图片编辑。第一张图片是当前站位图或用户标注图，只按以下修改要求编辑，保留其他内容：\n" + in.EditPrompt
	}
	if doc, err := s.Projects.ActiveDocument(ctx, p.OwnerUsername, p.ID); err == nil && strings.TrimSpace(doc.Content) != "" {
		text += "\n所属集完整剧本（仅供位置与动作理解，人物服装仍以参考图为准）：\n" + screenplay.PositioningEpisodeText(doc.Content, item.Episode)
	} else if err != nil && !errors.Is(err, screenplay.ErrNoActiveScript) && !errors.Is(err, screenplay.ErrNotFound) {
		return zero, nil, err
	} else if strings.TrimSpace(p.SourceText) != "" {
		text += "\n项目原文参考（尚无激活剧本）：\n" + screenplay.PositioningEpisodeText(p.SourceText, item.Episode)
	}
	if in.MarkedImageItemID != "" {
		text += "\n第一张参考图是站位画面参考（用户标注图或当前站位图），其后依次为角色参考图、场景参考图。"
	}
	if show {
		text += "\n本次角色名开关已开启：覆盖前缀中禁止姓名文字这一条，仅在对应人物附近标注其角色名，其他禁令保持。"
	} else {
		text += "\n不要添加角色名文字。"
	}
	urls := []string{}
	refNames := map[string]string{}
	if assets, err := s.Projects.ListAssets(ctx, p.OwnerUsername, p.ID, ""); err != nil {
		return zero, nil, err
	} else {
		for _, a := range assets {
			for _, m := range st.Items {
				if m.TargetID == a.ID {
					refNames[m.AssetID] = a.Name
				}
			}
		}
	}
	for i, ref := range refs {
		label := refNames[ref]
		if in.MarkedImageItemID != "" {
			if m, err := mediaItem(st, in.MarkedImageItemID); err == nil && m.AssetID == ref {
				label = "用户标注图"
			}
		}
		if label == "" {
			label = "参考图片"
		}
		text += fmt.Sprintf("\n图片%d：%s", i+1, label)
	}
	for _, ref := range refs {
		if s.ReferencesProvider == nil {
			return zero, nil, fmt.Errorf("参考图片签名服务未配置")
		}
		a, e := s.Assets.Asset(ctx, p.OwnerUsername, p.ID, ref)
		if e != nil {
			return zero, nil, e
		}
		u, e := s.ReferencesProvider.ReferenceURL(ctx, a, time.Hour)
		if e != nil {
			return zero, nil, e
		}
		urls = append(urls, u)
	}
	_, e = s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(st *screenplay.MediaState) error {
		current := positioningState(*st, p)
		if e := positioningExpected(in, current); e != nil {
			return e
		}
		for i := range current.Items {
			if current.Items[i].ID == id {
				current.Items[i].Status = "running"
				current.Items[i].Error = ""
				current.Items[i].RequestKey = key
				current.Revision++
				st.Positioning = &current
				return nil
			}
		}
		return screenplay.ErrNotFound
	})
	if e != nil {
		return zero, nil, e
	}
	finish := func(status string, runErr error, media *screenplay.MediaItem) error {
		_, e := s.Media.MutateMedia(context.WithoutCancel(ctx), p.OwnerUsername, p.ID, func(st *screenplay.MediaState) error {
			if media != nil && ctx.Err() != nil {
				return ctx.Err()
			}
			current := positioningState(*st, p)
			for i := range current.Items {
				it := &current.Items[i]
				if it.ID == id && it.RequestKey == key && it.Status == "running" {
					it.Status = status
					if runErr != nil {
						it.Error = runErr.Error()
					}
					if media != nil {
						it.ImageMediaItemID = media.ID
						st.Items = append(st.Items, *media)
					}
					current.Revision++
					st.Positioning = &current
					item = *it
					return nil
				}
			}
			return screenplay.ErrRevisionConflict
		})
		return e
	}
	size, resolution, sizeErr := screenplayImageDimensions(p, "16:9")
	if sizeErr != nil {
		_ = finish("failed", sizeErr, nil)
		return zero, nil, sizeErr
	}
	result, _, runErr := s.Images(ctx, p.OwnerUsername, AgentImageGenerationInput{ProjectID: p.ID, ModelID: model, Prompt: text, Count: 1, AspectRatio: "16:9", Size: size, Resolution: resolution, ReferenceAssetIDs: refs, ReferenceImages: urls})
	if ctx.Err() != nil {
		runErr = ctx.Err()
	}
	if runErr == nil && len(result) == 0 {
		runErr = fmt.Errorf("图片服务未返回结果")
	}
	if runErr != nil {
		status := "failed"
		if errors.Is(runErr, context.Canceled) {
			status = "cancelled"
		}
		_ = finish(status, runErr, nil)
		return zero, nil, runErr
	}
	ing := newAgentMediaAssetIngestor(s.Assets, s.Objects, nil, time.Now)
	a, runErr := ing.SaveImageResult(ctx, p.OwnerUsername, p.ID, newAgentID("sp-positioning"), result[0])
	if runErr == nil {
		runErr = ctx.Err()
	}
	if runErr == nil {
		runErr = s.ensureCut(ctx, p)
	}
	if runErr == nil {
		runErr = s.Cut.CreateAsset(ctx, p.OwnerUsername, aicut.Asset{ID: a.ID, ProjectID: p.ID, Kind: a.Kind, Name: a.Name, ContentType: a.ContentType, ObjectKey: a.ObjectKey, ByteSize: a.ByteSize, CreatedAt: a.CreatedAt})
	}
	if runErr == nil {
		runErr = ctx.Err()
	}
	if runErr != nil {
		_ = finish("failed", runErr, nil)
		return zero, nil, runErr
	}
	media := screenplay.MediaItem{ID: a.ID, AssetID: a.ID, Kind: "image", Status: "succeeded", TargetID: "positioning:" + id, CreatedAt: a.CreatedAt}
	if e = finish("succeeded", nil, &media); e != nil {
		if ctx.Err() != nil {
			_ = finish("cancelled", ctx.Err(), nil)
		}
		return zero, nil, e
	}
	return item, missing, nil
}

func (s *ScreenplayMediaWorkbench) analyzePositioning(ctx context.Context, p screenplay.Project, in positioningRequest) error {
	if s.PositioningText == nil {
		return fmt.Errorf("%w: 站位分析服务未配置", screenplay.ErrInvalidArgument)
	}
	model := firstText(in.Model, p.ModelID)
	if model == "" {
		return fmt.Errorf("%w: 请选择文本模型", screenplay.ErrInvalidArgument)
	}
	if in.AnalyzeMode != "complex" && in.AnalyzeMode != "global" {
		return fmt.Errorf("%w: 分析模式必须为complex或global", screenplay.ErrInvalidArgument)
	}
	st, e := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if e != nil {
		return e
	}
	v := positioningState(st, p)
	if e = positioningExpected(in, v); e != nil {
		return e
	}
	source := p.SourceText
	origin := "项目原文"
	doc, e := s.Projects.ActiveDocument(ctx, p.OwnerUsername, p.ID)
	if e == nil && strings.TrimSpace(doc.Content) != "" {
		source = doc.Content
		origin = "当前激活剧本"
	} else if e != nil && !errors.Is(e, screenplay.ErrNoActiveScript) && !errors.Is(e, screenplay.ErrNotFound) {
		return e
	}
	if strings.TrimSpace(source) == "" {
		return screenplay.ErrEmptyContent
	}
	assets, e := s.Projects.ListAssets(ctx, p.OwnerUsername, p.ID, "")
	if e != nil {
		return e
	}
	names := []string{}
	for _, a := range assets {
		names = append(names, string(a.Type)+":"+a.Name)
	}
	system := prompt.ScreenplayPositioningAnalysis + "\n"
	if in.AnalyzeMode == "complex" {
		system += "只提取需要明确空间关系、多人互动或调度复杂的站位；普通单人场景无需条目。"
	} else {
		system += "覆盖剧本中全部有角色出现的场景，按剧情顺序组织站位。"
	}
	response, e := s.PositioningText(ctx, p.OwnerUsername, screenplay.TextRequest{Model: model, SystemPrompt: system, UserPrompt: "来源：" + origin + "\n本项目资产名称（必须精确引用）：\n" + strings.Join(names, "\n") + "\n正文：\n" + source, JSONResponse: true, MaxOutputTokens: 16384, Temperature: .2})
	if e != nil {
		return e
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	var decoded struct {
		Items []screenplay.PositioningItem `json:"items"`
	}
	if e = json.Unmarshal([]byte(strings.TrimSpace(response.Content)), &decoded); e != nil || decoded.Items == nil {
		return fmt.Errorf("%w: 站位分析必须返回含items数组的JSON", screenplay.ErrAIInvalidJSON)
	}
	v.Items = decoded.Items
	knownCharacters, knownScenes := map[string]bool{}, map[string]bool{}
	for _, a := range assets {
		if a.Type == screenplay.AssetCharacter {
			knownCharacters[a.Name] = true
		}
		if a.Type == screenplay.AssetScene {
			knownScenes[a.Name] = true
		}
	}
	for i := range v.Items {
		it := &v.Items[i]
		if strings.TrimSpace(it.PositionDetail) == "" || (strings.TrimSpace(it.SceneName) == "" && strings.TrimSpace(it.SceneCode) == "") {
			return fmt.Errorf("%w: 分析条目缺少站位详细或场景", screenplay.ErrArtifactInvalid)
		}
		for _, name := range it.Characters {
			if !knownCharacters[name] {
				return fmt.Errorf("%w: 分析返回未绑定的角色%s", screenplay.ErrArtifactInvalid, name)
			}
		}
		if it.SceneName != "" && !knownScenes[it.SceneName] {
			return fmt.Errorf("%w: 分析返回未绑定的场景%s", screenplay.ErrArtifactInvalid, it.SceneName)
		}
		if it.ID == "" {
			it.ID = newAgentID("position")
		}
		it.ImagePath = ""
		it.ImageMediaItemID = ""
		it.Status = ""
		it.Error = ""
		it.RequestKey = ""
	}
	if e = screenplay.ValidatePositioning(&v); e != nil {
		return fmt.Errorf("%w: %v", screenplay.ErrArtifactInvalid, e)
	}
	if in.ImageModel != "" {
		v.ImageModel = in.ImageModel
	}
	_, e = s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(st *screenplay.MediaState) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if e := positioningExpected(in, positioningState(*st, p)); e != nil {
			return e
		}
		v.Revision++
		st.Positioning = &v
		return nil
	})
	return e
}
