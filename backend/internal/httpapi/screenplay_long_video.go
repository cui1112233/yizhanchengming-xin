package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/mergeworker"
	"qiantie/backend/internal/screenplay"
)

// 洛水 2026-10-29 版「超长视频」：目标总时长 100/200/250/300 秒，按推荐时长累积分组，
// 每组不超过单次生成上限（30 秒），逐组交给已保存的视频模型生成，再按顺序拼成成片。
type longVideoRequest struct {
	Episode       int                           `json:"episode"`
	Engine        string                        `json:"engine"`
	TargetSeconds float64                       `json:"target_seconds"`
	CapSeconds    float64                       `json:"cap_seconds"`
	Indexes       []int                         `json:"indexes"`
	Sections      []screenplay.LongVideoSection `json:"sections"`
}

type longVideoSectionView struct {
	screenplay.LongVideoSection
	ShotNumbers string `json:"shot_numbers"`
	VideoURL    string `json:"video_url,omitempty"`
}

type longVideoView struct {
	Episode       int                    `json:"episode"`
	TargetSeconds float64                `json:"target_seconds"`
	CapSeconds    float64                `json:"cap_seconds"`
	Sections      []longVideoSectionView `json:"sections"`
	TotalSeconds  float64                `json:"total_seconds"`
	FinalStatus   string                 `json:"final_status"`
	FinalItemID   string                 `json:"final_item_id,omitempty"`
	FinalURL      string                 `json:"final_url,omitempty"`
	FinalError    string                 `json:"final_error,omitempty"`
	Running       bool                   `json:"running"`
	Capabilities  map[string]any         `json:"capabilities"`
}

var longVideoRuns = struct {
	sync.Mutex
	active map[string]bool
}{active: map[string]bool{}}

func longVideoRunKey(owner, project string, episode int) string {
	return fmt.Sprintf("%s:%s:%d", owner, project, episode)
}

func claimLongVideoRun(key string) bool {
	longVideoRuns.Lock()
	defer longVideoRuns.Unlock()
	if longVideoRuns.active[key] {
		return false
	}
	longVideoRuns.active[key] = true
	return true
}

func releaseLongVideoRun(key string) {
	longVideoRuns.Lock()
	defer longVideoRuns.Unlock()
	delete(longVideoRuns.active, key)
}

func longVideoRunning(key string) bool {
	longVideoRuns.Lock()
	defer longVideoRuns.Unlock()
	return longVideoRuns.active[key]
}

func registerScreenplayLongVideoRoutes(mux *http.ServeMux, auth BridgeAuth, s *ScreenplayMediaWorkbench) {
	if mux == nil || s == nil || s.Media == nil || s.Projects == nil {
		return
	}
	handle := func(pattern, permission string, fn func(http.ResponseWriter, *http.Request, screenplay.Project)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := BridgeIdentityFromContext(r.Context())
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			p, e := s.resolve(r.Context(), identity.Username, r.PathValue("id"), permission)
			if e != nil {
				writeScreenplayMediaError(w, e)
				return
			}
			fn(w, r, p)
		})))
	}
	handle("POST /api/comic-screenplays/projects/{id}/long-video/plan", "read", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		var in longVideoRequest
		if e := decodeLongVideoBody(r, &in); e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		view, e := s.planLongVideo(r.Context(), p, in)
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		writeJSON(w, http.StatusOK, view)
	})
	handle("POST /api/comic-screenplays/projects/{id}/long-video/generate", "generate", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		var in longVideoRequest
		if e := decodeLongVideoBody(r, &in); e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		view, e := s.startLongVideo(r.Context(), p, in)
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		writeJSON(w, http.StatusOK, view)
	})
	handle("GET /api/comic-screenplays/projects/{id}/long-video/status", "read", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		episode := 0
		if raw := strings.TrimSpace(r.URL.Query().Get("episode")); raw != "" {
			fmt.Sscanf(raw, "%d", &episode)
		}
		view, e := s.longVideoStatus(r.Context(), auth, p, episode)
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		writeJSON(w, http.StatusOK, view)
	})
	handle("POST /api/comic-screenplays/projects/{id}/long-video/combine", "generate", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		var in longVideoRequest
		if e := decodeLongVideoBody(r, &in); e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		view, e := s.startLongVideoCombine(r.Context(), auth, p, in.Episode)
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		writeJSON(w, http.StatusOK, view)
	})
}

func decodeLongVideoBody(r *http.Request, into any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(into); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: 请求体无效", screenplay.ErrInvalidArgument)
	}
	return nil
}

func (s *ScreenplayMediaWorkbench) longVideoBoard(ctx context.Context, p screenplay.Project, episode int, engine screenplay.StoryboardEngine) (screenplay.Storyboard, error) {
	if episode <= 0 {
		return screenplay.Storyboard{}, fmt.Errorf("%w: 请选择剧集", screenplay.ErrInvalidArgument)
	}
	if engine != "" {
		return s.Projects.GetStoryboard(ctx, p.OwnerUsername, p.ID, episode, engine)
	}
	boards, err := s.Projects.ListStoryboards(ctx, p.OwnerUsername, p.ID, "")
	if err != nil {
		return screenplay.Storyboard{}, err
	}
	for _, board := range boards {
		if board.EpisodeNo == episode && len(board.Shots) > 0 {
			return board, nil
		}
	}
	return screenplay.Storyboard{}, fmt.Errorf("%w: 第 %d 集还没有可用分镜", screenplay.ErrInvalidState, episode)
}

func (s *ScreenplayMediaWorkbench) planLongVideo(ctx context.Context, p screenplay.Project, in longVideoRequest) (longVideoView, error) {
	board, err := s.longVideoBoard(ctx, p, in.Episode, screenplay.StoryboardEngine(in.Engine))
	if err != nil {
		return longVideoView{}, err
	}
	sections := screenplay.PlanLongVideoSections(board.Shots, in.TargetSeconds, in.CapSeconds)
	if len(sections) == 0 {
		return longVideoView{}, fmt.Errorf("%w: 本集没有可编排的超长视频分镜", screenplay.ErrInvalidState)
	}
	if screenplay.LongVideoPlannedSeconds(sections) <= 0 {
		return longVideoView{}, fmt.Errorf("%w: 分镜缺少推荐时长，请先补全分镜时长", screenplay.ErrInvalidState)
	}
	return longVideoView{
		Episode:       in.Episode,
		TargetSeconds: screenplay.NormalizeLongVideoTarget(in.TargetSeconds),
		CapSeconds:    screenplay.NormalizeLongVideoCap(in.CapSeconds),
		Sections:      s.longVideoSectionViews(nil, sections, BridgeAuth{}, p, false),
		TotalSeconds:  screenplay.LongVideoPlannedSeconds(sections),
		FinalStatus:   screenplay.LongVideoFinalStatusIdle,
		Running:       longVideoRunning(longVideoRunKey(p.OwnerUsername, p.ID, in.Episode)),
		Capabilities:  s.longVideoCapabilities(),
	}, nil
}

func (s *ScreenplayMediaWorkbench) startLongVideo(ctx context.Context, p screenplay.Project, in longVideoRequest) (longVideoView, error) {
	if s.Video == nil || s.Configs == nil {
		return longVideoView{}, fmt.Errorf("%w: 视频生成通道未配置", screenplay.ErrInvalidState)
	}
	if in.Episode <= 0 {
		return longVideoView{}, fmt.Errorf("%w: 请选择剧集", screenplay.ErrInvalidArgument)
	}
	batch, err := s.prepareLongVideoBatch(ctx, p, in)
	if err != nil {
		return longVideoView{}, err
	}
	indexes := in.Indexes
	if len(indexes) == 0 {
		indexes = make([]int, 0, len(batch.Sections))
		for _, section := range batch.Sections {
			if section.Status != screenplay.LongVideoStatusDone || section.ItemID == "" {
				indexes = append(indexes, section.Index)
			}
		}
	}
	if len(indexes) == 0 {
		return s.longVideoStatus(ctx, BridgeAuth{}, p, in.Episode)
	}
	key := longVideoRunKey(p.OwnerUsername, p.ID, in.Episode)
	if !claimLongVideoRun(key) {
		return longVideoView{}, fmt.Errorf("%w: 本集超长视频正在生成，请等待当前任务", screenplay.ErrInvalidState)
	}
	go func() {
		defer releaseLongVideoRun(key)
		s.runLongVideoBatch(p, in.Episode, indexes)
	}()
	return s.longVideoStatus(ctx, BridgeAuth{}, p, in.Episode)
}

// prepareLongVideoBatch 以服务端分镜为准重建分组，只接受客户端对提示词的修改。
func (s *ScreenplayMediaWorkbench) prepareLongVideoBatch(ctx context.Context, p screenplay.Project, in longVideoRequest) (screenplay.LongVideoBatch, error) {
	board, err := s.longVideoBoard(ctx, p, in.Episode, screenplay.StoryboardEngine(in.Engine))
	if err != nil {
		return screenplay.LongVideoBatch{}, err
	}
	planned := screenplay.PlanLongVideoSections(board.Shots, in.TargetSeconds, in.CapSeconds)
	if len(planned) == 0 {
		return screenplay.LongVideoBatch{}, fmt.Errorf("%w: 本集没有可编排的超长视频分镜", screenplay.ErrInvalidState)
	}
	if screenplay.LongVideoPlannedSeconds(planned) <= 0 {
		return screenplay.LongVideoBatch{}, fmt.Errorf("%w: 分镜缺少推荐时长，请先补全分镜时长", screenplay.ErrInvalidState)
	}
	st, err := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if err != nil {
		return screenplay.LongVideoBatch{}, err
	}
	batch := screenplay.LongVideoBatch{
		ID:            newAgentID("lv-batch"),
		Episode:       in.Episode,
		TargetSeconds: screenplay.NormalizeLongVideoTarget(in.TargetSeconds),
		CapSeconds:    screenplay.NormalizeLongVideoCap(in.CapSeconds),
		FinalStatus:   screenplay.LongVideoFinalStatusIdle,
		Sections:      planned,
	}
	if existing, ok := st.FindLongVideoBatch(in.Episode); ok && sameLongVideoPlan(existing.Sections, planned) {
		batch = existing
		batch.TargetSeconds = screenplay.NormalizeLongVideoTarget(in.TargetSeconds)
		batch.CapSeconds = screenplay.NormalizeLongVideoCap(in.CapSeconds)
	}
	for i := range batch.Sections {
		if batch.Sections[i].Status == "" {
			batch.Sections[i].Status = screenplay.LongVideoStatusIdle
		}
		if batch.Sections[i].Prompt == "" {
			batch.Sections[i].Prompt = planned[i].Prompt
		}
		if edited := findLongVideoSection(in.Sections, batch.Sections[i].Index); edited != nil {
			if prompt := strings.TrimSpace(edited.Prompt); prompt != "" {
				batch.Sections[i].Prompt = prompt
			}
		}
	}
	_, err = s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(state *screenplay.MediaState) error {
		state.PutLongVideoBatch(batch)
		return nil
	})
	if err != nil {
		return screenplay.LongVideoBatch{}, err
	}
	return batch, nil
}

func sameLongVideoPlan(existing, planned []screenplay.LongVideoSection) bool {
	if len(existing) != len(planned) {
		return false
	}
	for i := range planned {
		if len(existing[i].ShotIDs) != len(planned[i].ShotIDs) {
			return false
		}
		for j := range planned[i].ShotIDs {
			if existing[i].ShotIDs[j] != planned[i].ShotIDs[j] {
				return false
			}
		}
	}
	return true
}

func findLongVideoSection(sections []screenplay.LongVideoSection, index int) *screenplay.LongVideoSection {
	for i := range sections {
		if sections[i].Index == index {
			return &sections[i]
		}
	}
	return nil
}

func (s *ScreenplayMediaWorkbench) longVideoBatch(ctx context.Context, p screenplay.Project, episode int) (screenplay.LongVideoBatch, bool) {
	st, err := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if err != nil {
		return screenplay.LongVideoBatch{}, false
	}
	return st.FindLongVideoBatch(episode)
}

func (s *ScreenplayMediaWorkbench) updateLongVideoSection(p screenplay.Project, episode, index int, mutate func(*screenplay.LongVideoSection)) {
	ctx := context.Background()
	_, err := s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(state *screenplay.MediaState) error {
		state.MutateLongVideoSection(episode, index, mutate)
		return nil
	})
	_ = err
}

func (s *ScreenplayMediaWorkbench) failLongVideoSections(p screenplay.Project, episode int, indexes []int, message string) {
	for _, index := range indexes {
		s.updateLongVideoSection(p, episode, index, func(section *screenplay.LongVideoSection) {
			section.Status = screenplay.LongVideoStatusFailed
			section.Error = message
		})
	}
}

func (s *ScreenplayMediaWorkbench) runLongVideoBatch(p screenplay.Project, episode int, indexes []int) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Minute)
	defer cancel()
	config, err := s.Configs.GetConfig(ctx, p.OwnerUsername)
	if err != nil {
		s.failLongVideoSections(p, episode, indexes, "读取视频模型配置失败："+err.Error())
		return
	}
	modelKey, resolution := screenplayVideoDefaults(config)
	if modelKey == "" {
		s.failLongVideoSections(p, episode, indexes, "请先在设置中保存一个可用的视频模型")
		return
	}
	board, err := s.longVideoBoard(ctx, p, episode, "")
	if err != nil {
		s.failLongVideoSections(p, episode, indexes, err.Error())
		return
	}
	for _, index := range indexes {
		if ctx.Err() != nil {
			return
		}
		batch, ok := s.longVideoBatch(ctx, p, episode)
		if !ok {
			return
		}
		section := findLongVideoSection(batch.Sections, index)
		if section == nil {
			continue
		}
		if section.Status == screenplay.LongVideoStatusDone && section.ItemID != "" {
			continue
		}
		current := *section
		taskID := strings.TrimSpace(current.ProviderID)
		if taskID == "" {
			input, err := s.longVideoInput(ctx, p, board, current, modelKey, resolution)
			if err != nil {
				s.failLongVideoSection(p, episode, index, err.Error())
				continue
			}
			s.updateLongVideoSection(p, episode, index, func(section *screenplay.LongVideoSection) {
				section.Status = screenplay.LongVideoStatusGenerating
				section.Progress = 5
				section.Error = ""
			})
			submission, err := s.Video.Submit(ctx, p.OwnerUsername, input)
			if err != nil {
				s.failLongVideoSection(p, episode, index, "提交失败："+err.Error())
				continue
			}
			if strings.TrimSpace(submission.TaskID) == "" {
				s.failLongVideoSection(p, episode, index, "供应商未返回任务编号，已停止重复提交")
				continue
			}
			taskID = submission.TaskID
			s.updateLongVideoSection(p, episode, index, func(section *screenplay.LongVideoSection) {
				section.ProviderID = taskID
				section.Progress = 12
			})
		}
		s.updateLongVideoSection(p, episode, index, func(section *screenplay.LongVideoSection) {
			section.Status = screenplay.LongVideoStatusGenerating
			if section.Progress < 12 {
				section.Progress = 12
			}
		})
		artifact, err := s.pollLongVideo(ctx, p, episode, index, taskID)
		if err != nil {
			s.failLongVideoSection(p, episode, index, err.Error())
			continue
		}
		item, err := s.importLongVideo(ctx, p, current, taskID, artifact)
		if err != nil {
			s.failLongVideoSection(p, episode, index, "已生成但导入失败："+err.Error())
			continue
		}
		s.updateLongVideoSection(p, episode, index, func(section *screenplay.LongVideoSection) {
			section.Status = screenplay.LongVideoStatusDone
			section.Progress = 100
			section.ItemID = item.ID
			section.Error = ""
		})
	}
}

func (s *ScreenplayMediaWorkbench) failLongVideoSection(p screenplay.Project, episode, index int, message string) {
	s.updateLongVideoSection(p, episode, index, func(section *screenplay.LongVideoSection) {
		section.Status = screenplay.LongVideoStatusFailed
		section.Error = message
	})
}

func (s *ScreenplayMediaWorkbench) longVideoInput(ctx context.Context, p screenplay.Project, board screenplay.Storyboard, section screenplay.LongVideoSection, modelKey, resolution string) (scriptVideoLocalInput, error) {
	prompt := strings.TrimSpace(section.Prompt)
	if prompt == "" {
		return scriptVideoLocalInput{}, fmt.Errorf("第 %d 组缺少视频提示词", section.Index)
	}
	if section.TotalSeconds <= 0 {
		return scriptVideoLocalInput{}, fmt.Errorf("第 %d 组缺少推荐时长", section.Index)
	}
	input := scriptVideoLocalInput{
		SourceTaskID:  fmt.Sprintf("screenplay-long-video:%s:%d:%d", p.ID, section.Index, len(section.ShotIDs)),
		ModelKey:      modelKey,
		Prompt:        prompt,
		Duration:      section.TotalSeconds,
		AspectRatio:   p.AspectRatio,
		Resolution:    resolution,
		GenerateAudio: true,
	}
	config := screenplay.CreativeConfig{}
	if p.CreativeConfig != nil {
		config = *p.CreativeConfig
	}
	if err := screenplay.ValidateCreativeVideoConfig(config); err != nil {
		return input, err
	}
	input.GenerationMode = config.VideoGenerationMode
	if input.GenerationMode != "text_to_video" {
		references, err := s.longVideoReferences(ctx, p, board, section)
		if err != nil {
			return input, err
		}
		input.ImageURLs = references
	}
	if input.GenerationMode == "image_to_video" && len(input.ImageURLs) == 0 {
		return input, fmt.Errorf("图生视频缺少可用参考图片：请在镜头图片中选定参考图")
	}
	return input, nil
}

// longVideoReferences 合并组内各分镜已选定的参考图，最多四张。
func (s *ScreenplayMediaWorkbench) longVideoReferences(ctx context.Context, p screenplay.Project, board screenplay.Storyboard, section screenplay.LongVideoSection) ([]string, error) {
	if s.ReferencesProvider == nil {
		return nil, nil
	}
	wanted := map[string]bool{}
	for _, id := range section.ShotIDs {
		wanted[id] = true
	}
	seen := map[string]bool{}
	urls := make([]string, 0, 4)
	for _, shot := range board.Shots {
		if !wanted[shot.ID] {
			continue
		}
		references, err := s.References(ctx, p, shot)
		if err != nil {
			return nil, err
		}
		for _, url := range references {
			if url == "" || seen[url] {
				continue
			}
			seen[url] = true
			urls = append(urls, url)
			if len(urls) >= 4 {
				return urls, nil
			}
		}
	}
	return urls, nil
}

func (s *ScreenplayMediaWorkbench) pollLongVideo(ctx context.Context, p screenplay.Project, episode, index int, taskID string) (string, error) {
	deadline := time.Now().Add(20 * time.Minute)
	interval := 3 * time.Second
	progress := 15
	for {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		status, err := s.Video.Poll(ctx, p.OwnerUsername, taskID)
		if err != nil {
			return "", fmt.Errorf("查询生成任务失败（继续时会查询原任务）：%w", err)
		}
		switch strings.ToLower(strings.TrimSpace(status.Status)) {
		case "succeeded", "success", "completed", "done":
			artifact := strings.TrimSpace(status.VideoURL)
			if artifact == "" && strings.TrimSpace(status.ArtifactID) != "" {
				artifact = "/api/shuihuo-production/local-executor-artifacts/" + strings.TrimSpace(status.ArtifactID)
			}
			if artifact == "" {
				return "", fmt.Errorf("生成完成但没有视频地址")
			}
			return artifact, nil
		case "failed", "error", "cancelled":
			return "", fmt.Errorf("生成失败：%s", firstText(status.Error, "视频通道返回失败"))
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("生成超时，可在重新打开后继续查询原任务")
		}
		if progress < 90 {
			progress += 5
			s.updateLongVideoSection(p, episode, index, func(section *screenplay.LongVideoSection) {
				if section.Progress < progress {
					section.Progress = progress
				}
			})
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(interval):
		}
	}
}

// importLongVideo 把供应商结果导入素材库，失败也不会重复提交。
func (s *ScreenplayMediaWorkbench) importLongVideo(ctx context.Context, p screenplay.Project, section screenplay.LongVideoSection, taskID, artifact string) (screenplay.MediaItem, error) {
	if s.Assets == nil || s.Objects == nil || s.Cut == nil {
		return screenplay.MediaItem{}, fmt.Errorf("素材存储未配置")
	}
	itemID := newAgentID("sp-lv")
	ingestor := newAgentMediaAssetIngestor(s.Assets, s.Objects, nil, time.Now)
	var asset agentworkspace.Asset
	var err error
	const prefix = "/api/shuihuo-production/local-executor-artifacts/"
	if strings.HasPrefix(artifact, prefix) {
		asset, err = ingestor.SaveLocal(ctx, p.OwnerUsername, p.ID, itemID, strings.TrimPrefix(artifact, prefix), s.Executors, s.Files)
	} else {
		asset, err = ingestor.SaveRemote(ctx, p.OwnerUsername, p.ID, itemID, artifact)
	}
	if err != nil {
		return screenplay.MediaItem{}, err
	}
	if err = s.ensureCut(ctx, p); err != nil {
		return screenplay.MediaItem{}, err
	}
	cutAsset := aicut.Asset{ID: asset.ID, ProjectID: p.ID, Kind: asset.Kind, Name: asset.Name, ContentType: asset.ContentType, ObjectKey: asset.ObjectKey, ByteSize: asset.ByteSize, CreatedAt: asset.CreatedAt}
	if duration, probeErr := probeAICutAssetDuration(ctx, cutAsset, s.Objects); probeErr == nil && duration > 0 {
		cutAsset.DurationMS = duration
	}
	if err = s.Cut.CreateAsset(ctx, p.OwnerUsername, cutAsset); err != nil {
		return screenplay.MediaItem{}, err
	}
	item := screenplay.MediaItem{ID: itemID, AssetID: asset.ID, Name: section.Label, Kind: "video", Status: "succeeded", TargetID: section.ID, ProviderID: taskID, DurationMS: cutAsset.DurationMS, CreatedAt: time.Now().UTC()}
	return item, s.append(ctx, p, item)
}

func (s *ScreenplayMediaWorkbench) longVideoStatus(ctx context.Context, auth BridgeAuth, p screenplay.Project, episode int) (longVideoView, error) {
	batch, ok := s.longVideoBatch(ctx, p, episode)
	if !ok {
		return longVideoView{}, fmt.Errorf("%w: 第 %d 集还没有编排超长视频", screenplay.ErrNotFound, episode)
	}
	st, err := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if err != nil {
		return longVideoView{}, err
	}
	view := longVideoView{
		Episode:       batch.Episode,
		TargetSeconds: batch.TargetSeconds,
		CapSeconds:    batch.CapSeconds,
		TotalSeconds:  screenplay.LongVideoPlannedSeconds(batch.Sections),
		FinalStatus:   firstText(batch.FinalStatus, screenplay.LongVideoFinalStatusIdle),
		FinalItemID:   batch.FinalItemID,
		FinalError:    batch.FinalError,
		Running:       longVideoRunning(longVideoRunKey(p.OwnerUsername, p.ID, episode)),
		Capabilities:  s.longVideoCapabilities(),
		Sections:      s.longVideoSectionViews(&st, batch.Sections, auth, p, true),
	}
	for _, item := range st.Items {
		if item.ID == batch.FinalItemID && item.AssetID != "" && item.Status == "succeeded" {
			if auth.Secret != "" {
				view.FinalURL = aicutSignedAssetContentURL(auth, p.OwnerUsername, p.ID, item.AssetID, "content", time.Now())
			} else {
				view.FinalURL = aicutAssetContentURL(p.ID, item.AssetID)
			}
			break
		}
	}
	return view, nil
}

func (s *ScreenplayMediaWorkbench) longVideoSectionViews(st *screenplay.MediaState, sections []screenplay.LongVideoSection, auth BridgeAuth, p screenplay.Project, withURLs bool) []longVideoSectionView {
	views := make([]longVideoSectionView, 0, len(sections))
	for _, section := range sections {
		view := longVideoSectionView{LongVideoSection: section}
		if section.Status == "" {
			view.Status = screenplay.LongVideoStatusIdle
		}
		parts := make([]string, 0, len(section.ShotSeqs))
		for _, seq := range section.ShotSeqs {
			parts = append(parts, fmt.Sprintf("第%d镜", seq))
		}
		view.ShotNumbers = strings.Join(parts, "、")
		if withURLs && st != nil && section.ItemID != "" {
			for _, item := range st.Items {
				if item.ID != section.ItemID || item.AssetID == "" || item.Status != "succeeded" {
					continue
				}
				if auth.Secret != "" {
					view.VideoURL = aicutSignedAssetContentURL(auth, p.OwnerUsername, p.ID, item.AssetID, "content", time.Now())
				} else {
					view.VideoURL = aicutAssetContentURL(p.ID, item.AssetID)
				}
				break
			}
		}
		views = append(views, view)
	}
	return views
}

func (s *ScreenplayMediaWorkbench) longVideoCapabilities() map[string]any {
	return map[string]any{
		"generate": map[string]any{"available": s.Video != nil && s.Configs != nil, "reason": "使用当前账号在设置中保存的视频模型"},
		"combine":  map[string]any{"available": s.Cut != nil && s.Objects != nil, "reason": "需要剪辑与素材存储"},
	}
}

func (s *ScreenplayMediaWorkbench) startLongVideoCombine(ctx context.Context, auth BridgeAuth, p screenplay.Project, episode int) (longVideoView, error) {
	if s.Cut == nil || s.Objects == nil {
		return longVideoView{}, fmt.Errorf("%w: 剪辑与素材存储未配置", screenplay.ErrInvalidState)
	}
	batch, ok := s.longVideoBatch(ctx, p, episode)
	if !ok {
		return longVideoView{}, fmt.Errorf("%w: 第 %d 集还没有编排超长视频", screenplay.ErrNotFound, episode)
	}
	for _, section := range batch.Sections {
		if section.Status != screenplay.LongVideoStatusDone || section.ItemID == "" {
			return longVideoView{}, fmt.Errorf("%w: 还有分组未生成完成，暂不能合并成片", screenplay.ErrInvalidState)
		}
	}
	combineKey := "combine:" + longVideoRunKey(p.OwnerUsername, p.ID, episode)
	if !claimLongVideoRun(combineKey) {
		return longVideoView{}, fmt.Errorf("%w: 成片正在合并", screenplay.ErrInvalidState)
	}
	_, err := s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(state *screenplay.MediaState) error {
		return markLongVideoCombining(state, episode)
	})
	if err != nil {
		releaseLongVideoRun(combineKey)
		return longVideoView{}, err
	}
	go func() {
		defer releaseLongVideoRun(combineKey)
		s.combineLongVideoBatch(p, episode, combineKey)
	}()
	return s.longVideoStatus(ctx, auth, p, episode)
}

func markLongVideoCombining(state *screenplay.MediaState, episode int) error {
	batch, ok := state.FindLongVideoBatch(episode)
	if !ok {
		return fmt.Errorf("%w: 编排记录不存在", screenplay.ErrNotFound)
	}
	batch.FinalStatus = screenplay.LongVideoFinalStatusCombining
	batch.FinalError = ""
	state.PutLongVideoBatch(batch)
	return nil
}

func (s *ScreenplayMediaWorkbench) combineLongVideoBatch(p screenplay.Project, episode int, combineKey string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	fail := func(message string) {
		_, _ = s.Media.MutateMedia(context.Background(), p.OwnerUsername, p.ID, func(state *screenplay.MediaState) error {
			batch, ok := state.FindLongVideoBatch(episode)
			if !ok {
				return nil
			}
			batch.FinalStatus = screenplay.LongVideoFinalStatusFailed
			batch.FinalError = message
			state.PutLongVideoBatch(batch)
			return nil
		})
	}
	batch, ok := s.longVideoBatch(ctx, p, episode)
	if !ok {
		return
	}
	directory, err := os.MkdirTemp("", "qiantie-long-video-*")
	if err != nil {
		fail("临时目录创建失败：" + err.Error())
		return
	}
	defer os.RemoveAll(directory)
	st, err := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if err != nil {
		fail("读取媒体状态失败：" + err.Error())
		return
	}
	inputs := make([]string, 0, len(batch.Sections))
	for _, section := range batch.Sections {
		if section.ItemID == "" {
			fail("还有分组没有成片")
			return
		}
		var assetID string
		for _, item := range st.Items {
			if item.ID == section.ItemID && item.Status == "succeeded" {
				assetID = item.AssetID
				break
			}
		}
		if assetID == "" {
			fail(fmt.Sprintf("%s 的片段不存在", section.Label))
			return
		}
		asset, err := s.Cut.Asset(ctx, p.OwnerUsername, p.ID, assetID)
		if err != nil {
			fail("片段读取失败：" + err.Error())
			return
		}
		reader, err := s.Objects.Open(ctx, aicutObjectAsset(asset))
		if err != nil {
			fail("片段对象读取失败：" + err.Error())
			return
		}
		target := filepath.Join(directory, fmt.Sprintf("seg-%03d.mp4", len(inputs)))
		file, err := os.Create(target)
		if err != nil {
			reader.Close()
			fail("临时文件创建失败：" + err.Error())
			return
		}
		_, copyErr := io.Copy(file, reader)
		reader.Close()
		file.Close()
		if copyErr != nil {
			fail("片段下载失败：" + copyErr.Error())
			return
		}
		inputs = append(inputs, target)
	}
	output := filepath.Join(directory, "final.mp4")
	if err := (&mergeworker.FFmpegRunner{}).Merge(ctx, inputs, output, 0); err != nil {
		fail("合并失败：" + err.Error())
		return
	}
	item, err := s.saveLongVideoFile(ctx, p, output, fmt.Sprintf("超长视频成片-第%d集.mp4", episode))
	if err != nil {
		fail("成片保存失败：" + err.Error())
		return
	}
	_, err = s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(state *screenplay.MediaState) error {
		batch, ok := state.FindLongVideoBatch(episode)
		if !ok {
			return nil
		}
		batch.FinalStatus = screenplay.LongVideoFinalStatusDone
		batch.FinalItemID = item.ID
		batch.FinalError = ""
		state.PutLongVideoBatch(batch)
		return nil
	})
	if err != nil {
		fail("成片结果保存失败：" + err.Error())
	}
}

// saveLongVideoFile 以流式方式保存成片，不受普通素材 128MB 上限限制。
func (s *ScreenplayMediaWorkbench) saveLongVideoFile(ctx context.Context, p screenplay.Project, path, name string) (screenplay.MediaItem, error) {
	return s.saveStreamedVideoFile(ctx, p, path, name, "long-video")
}

func (s *ScreenplayMediaWorkbench) saveStreamedVideoFile(ctx context.Context, p screenplay.Project, path, name, targetID string) (screenplay.MediaItem, error) {
	file, err := os.Open(path)
	if err != nil {
		return screenplay.MediaItem{}, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return screenplay.MediaItem{}, err
	}
	if err = s.ensureCut(ctx, p); err != nil {
		return screenplay.MediaItem{}, err
	}
	id := newAgentID("sp-long-video")
	key, err := agentworkspace.AgentAssetObjectKey(p.OwnerUsername, p.ID, id, "mp4")
	if err != nil {
		return screenplay.MediaItem{}, err
	}
	key = strings.Replace(key, "agent-assets/", "aicut-assets/", 1)
	asset := aicut.Asset{ID: id, ProjectID: p.ID, Kind: "video", Name: name, ContentType: "video/mp4", ObjectKey: key, ByteSize: info.Size(), CreatedAt: time.Now().UTC()}
	if err = s.Objects.Put(ctx, aicutObjectAsset(asset), file); err != nil {
		return screenplay.MediaItem{}, err
	}
	if duration, probeErr := probeAICutAssetDuration(ctx, asset, s.Objects); probeErr == nil {
		asset.DurationMS = duration
	}
	if err = s.Cut.CreateAsset(ctx, p.OwnerUsername, asset); err != nil {
		return screenplay.MediaItem{}, err
	}
	if s.Assets != nil {
		_ = s.Assets.CreateAsset(ctx, p.OwnerUsername, aicutObjectAsset(asset))
	}
	item := screenplay.MediaItem{ID: id, AssetID: id, Name: name, DurationMS: asset.DurationMS, Kind: "video", Status: "succeeded", TargetID: targetID, CreatedAt: time.Now().UTC()}
	if targetID != "" && targetID != "long-video" {
		item.ShotID = targetID
	}
	return item, s.append(ctx, p, item)
}
