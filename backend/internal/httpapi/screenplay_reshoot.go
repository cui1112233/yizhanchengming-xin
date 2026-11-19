package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
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
	"qiantie/backend/internal/ffmpegbin"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/mergeworker"
	"qiantie/backend/internal/screenplay"
)

// 洛水 2026-10-29 版「片段重拍」：在 20 帧时间轴上框选当前镜头视频的一段，
// 抽首尾帧作为参考图重新生成该片段，再把「原前段 + 新片段 + 原后段」拼回完整视频。
// 这里只重拍被框选的部分，不重新生成整条镜头，也不覆盖原视频素材。

type reshootRequest struct {
	ShotID   string  `json:"shot_id"`
	Episode  int     `json:"episode"`
	Engine   string  `json:"engine"`
	Start    float64 `json:"start"`
	End      float64 `json:"end"`
	Duration float64 `json:"duration"`
	Prompt   string  `json:"prompt"`
	Mode     string  `json:"mode"`
}

type reshootCandidate struct {
	ShotID   string  `json:"shot_id"`
	Seq      int     `json:"seq"`
	Episode  int     `json:"episode"`
	Engine   string  `json:"engine"`
	Duration float64 `json:"duration"`
	ItemID   string  `json:"item_id"`
	VideoURL string  `json:"video_url"`
	Name     string  `json:"name,omitempty"`
}

type reshootCandidatesView struct {
	Episode      int                `json:"episode"`
	Candidates   []reshootCandidate `json:"candidates"`
	Capabilities map[string]any     `json:"capabilities"`
}

type reshootTaskView struct {
	screenplay.ReshootTask
	SourceURL  string `json:"source_url,omitempty"`
	SegmentURL string `json:"segment_url,omitempty"`
	AppliedURL string `json:"applied_url,omitempty"`
}

type reshootFramesView struct {
	ShotID     string   `json:"shot_id"`
	Duration   float64  `json:"duration"`
	Frames     []string `json:"frames"`
	FirstFrame string   `json:"first_frame"`
	LastFrame  string   `json:"last_frame"`
}

var reshootSyncs = struct {
	sync.Mutex
	active map[string]bool
}{active: map[string]bool{}}

func reshootSyncKey(owner, project, shotID string) string {
	return owner + ":" + project + ":" + shotID
}

func registerScreenplayReshootRoutes(mux *http.ServeMux, auth BridgeAuth, s *ScreenplayMediaWorkbench) {
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
	handle("GET /api/comic-screenplays/projects/{id}/reshoot/candidates", "read", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		episode := 0
		if raw := strings.TrimSpace(r.URL.Query().Get("episode")); raw != "" {
			fmt.Sscanf(raw, "%d", &episode)
		}
		engine := strings.TrimSpace(r.URL.Query().Get("engine"))
		view, e := s.reshootCandidates(r.Context(), auth, p, episode, screenplay.StoryboardEngine(engine))
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		writeJSON(w, http.StatusOK, view)
	})
	handle("POST /api/comic-screenplays/projects/{id}/reshoot/frames", "read", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		var in reshootRequest
		if e := decodeReshootBody(r, &in); e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		view, e := s.reshootFrames(r.Context(), p, in)
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		writeJSON(w, http.StatusOK, view)
	})
	handle("POST /api/comic-screenplays/projects/{id}/reshoot/generate", "generate", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		var in reshootRequest
		if e := decodeReshootBody(r, &in); e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		view, e := s.startReshoot(r.Context(), p, in)
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		writeJSON(w, http.StatusOK, view)
	})
	handle("GET /api/comic-screenplays/projects/{id}/reshoot/status", "read", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		shotID := strings.TrimSpace(r.URL.Query().Get("shot_id"))
		view, e := s.reshootStatus(r.Context(), auth, p, shotID)
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		writeJSON(w, http.StatusOK, view)
	})
	handle("POST /api/comic-screenplays/projects/{id}/reshoot/apply", "generate", func(w http.ResponseWriter, r *http.Request, p screenplay.Project) {
		var in reshootRequest
		if e := decodeReshootBody(r, &in); e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		view, e := s.applyReshoot(r.Context(), auth, p, in.ShotID)
		if e != nil {
			writeScreenplayMediaError(w, e)
			return
		}
		writeJSON(w, http.StatusOK, view)
	})
}

func decodeReshootBody(r *http.Request, into any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(into); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: 请求体无效", screenplay.ErrInvalidArgument)
	}
	return nil
}

// findShotBoard 在项目中按镜头编号定位分镜与其所属剧集。
func (s *ScreenplayMediaWorkbench) findShotBoard(ctx context.Context, p screenplay.Project, shotID string) (screenplay.Storyboard, screenplay.StoryboardShot, error) {
	shotID = strings.TrimSpace(shotID)
	if shotID == "" {
		return screenplay.Storyboard{}, screenplay.StoryboardShot{}, fmt.Errorf("%w: 请选择分镜", screenplay.ErrInvalidArgument)
	}
	boards, err := s.Projects.ListStoryboards(ctx, p.OwnerUsername, p.ID, "")
	if err != nil {
		return screenplay.Storyboard{}, screenplay.StoryboardShot{}, err
	}
	for _, board := range boards {
		for _, shot := range board.Shots {
			if shot.ID == shotID {
				return board, shot, nil
			}
		}
	}
	return screenplay.Storyboard{}, screenplay.StoryboardShot{}, fmt.Errorf("%w: 分镜不存在", screenplay.ErrNotFound)
}

// shotVideoItem 取这条分镜最近一次生成成功的视频素材。
func shotVideoItem(state screenplay.MediaState, shotID string) (screenplay.MediaItem, bool) {
	var latest screenplay.MediaItem
	found := false
	for _, item := range state.Items {
		if item.Kind != "video" || item.Status != "succeeded" || item.AssetID == "" {
			continue
		}
		if item.ShotID != shotID {
			continue
		}
		if !found || item.CreatedAt.After(latest.CreatedAt) {
			latest = item
			found = true
		}
	}
	return latest, found
}

func (s *ScreenplayMediaWorkbench) reshootCandidates(ctx context.Context, auth BridgeAuth, p screenplay.Project, episode int, engine screenplay.StoryboardEngine) (reshootCandidatesView, error) {
	boards, err := s.Projects.ListStoryboards(ctx, p.OwnerUsername, p.ID, engine)
	if err != nil {
		return reshootCandidatesView{}, err
	}
	state, err := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if err != nil {
		return reshootCandidatesView{}, err
	}
	view := reshootCandidatesView{Episode: episode, Candidates: []reshootCandidate{}, Capabilities: s.reshootCapabilities()}
	for _, board := range boards {
		if episode > 0 && board.EpisodeNo != episode {
			continue
		}
		for _, shot := range board.Shots {
			item, ok := shotVideoItem(state, shot.ID)
			if !ok {
				continue
			}
			candidate := reshootCandidate{
				ShotID:   shot.ID,
				Seq:      shot.Seq,
				Episode:  board.EpisodeNo,
				Engine:   string(board.Engine),
				Duration: shot.Duration,
				ItemID:   item.ID,
				VideoURL: aicutSignedAssetContentURL(auth, p.OwnerUsername, p.ID, item.AssetID, "content", time.Now()),
				Name:     item.Name,
			}
			view.Candidates = append(view.Candidates, candidate)
		}
	}
	sortReshootCandidates(view.Candidates)
	if len(view.Candidates) > 0 && view.Episode == 0 {
		view.Episode = view.Candidates[0].Episode
	}
	return view, nil
}

func sortReshootCandidates(rows []reshootCandidate) {
	for i := 1; i < len(rows); i++ {
		for j := i; j > 0; j-- {
			left, right := rows[j-1], rows[j]
			ordered := left.Episode < right.Episode || (left.Episode == right.Episode && left.Seq <= right.Seq)
			if ordered {
				break
			}
			rows[j-1], rows[j] = rows[j], rows[j-1]
		}
	}
}

func (s *ScreenplayMediaWorkbench) reshootCapabilities() map[string]any {
	generateReady := s.Video != nil && s.Configs != nil && s.Objects != nil && s.Cut != nil && s.ReferencesProvider != nil
	generateReason := "使用当前账号在设置中保存的视频模型，首尾帧作为参考图"
	if s.Video == nil || s.Configs == nil {
		generateReason = "视频生成通道未配置"
	} else if s.ReferencesProvider == nil {
		generateReason = "当前素材存储不支持给视频模型提供首尾帧参考图"
	}
	available := true
	applyReason := "使用 ffmpeg 把原前段、新片段、原后段拼成完整视频"
	if _, err := mergeworkerProbeFFmpeg(); err != nil {
		available = false
		applyReason = err.Error()
	}
	if s.Objects == nil || s.Cut == nil {
		available = false
		applyReason = "需要素材存储与剪辑服务"
	}
	return map[string]any{
		"frames":   map[string]any{"available": s.Objects != nil && s.Cut != nil, "reason": "抽取 20 帧预览与首尾帧需要素材存储"},
		"generate": map[string]any{"available": generateReady, "reason": generateReason},
		"apply":    map[string]any{"available": available, "reason": applyReason},
		"max_seconds": screenplay.ReshootMaxSeconds,
		"min_seconds": screenplay.ReshootMinSeconds,
	}
}

// mergeworkerProbeFFmpeg 允许测试替换，默认使用项目自带的 ffmpeg 解析。
var mergeworkerProbeFFmpeg = ffmpegbin.ResolveFFmpeg

func (s *ScreenplayMediaWorkbench) shotVideoPath(ctx context.Context, p screenplay.Project, shotID string) (screenplay.MediaItem, string, float64, error) {
	state, err := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if err != nil {
		return screenplay.MediaItem{}, "", 0, err
	}
	item, ok := shotVideoItem(state, shotID)
	if !ok {
		return screenplay.MediaItem{}, "", 0, fmt.Errorf("%w: 这条分镜还没有生成好的视频", screenplay.ErrInvalidState)
	}
	if s.Objects == nil || s.Cut == nil {
		return screenplay.MediaItem{}, "", 0, fmt.Errorf("%w: 素材存储未配置", screenplay.ErrInvalidState)
	}
	asset, err := s.Cut.Asset(ctx, p.OwnerUsername, p.ID, item.AssetID)
	if err != nil {
		return screenplay.MediaItem{}, "", 0, err
	}
	reader, err := s.Objects.Open(ctx, aicutObjectAsset(asset))
	if err != nil {
		return screenplay.MediaItem{}, "", 0, err
	}
	defer reader.Close()
	directory, err := os.MkdirTemp("", "qiantie-reshoot-*")
	if err != nil {
		return screenplay.MediaItem{}, "", 0, err
	}
	target := filepath.Join(directory, "source.mp4")
	file, err := os.Create(target)
	if err != nil {
		os.RemoveAll(directory)
		return screenplay.MediaItem{}, "", 0, err
	}
	_, copyErr := io.Copy(file, reader)
	file.Close()
	if copyErr != nil {
		os.RemoveAll(directory)
		return screenplay.MediaItem{}, "", 0, copyErr
	}
	info, probeErr := mergeworker.ProbeMedia(ctx, target)
	if probeErr != nil {
		// 时长读不出来时仍可播放，但无法框选，交给调用方给出原因。
		return item, target, 0, nil
	}
	return item, target, info.Duration, nil
}

func (s *ScreenplayMediaWorkbench) reshootFrames(ctx context.Context, p screenplay.Project, in reshootRequest) (reshootFramesView, error) {
	if s.Objects == nil || s.Cut == nil {
		return reshootFramesView{}, fmt.Errorf("%w: 素材存储未配置", screenplay.ErrInvalidState)
	}
	_, source, duration, err := s.shotVideoPath(ctx, p, in.ShotID)
	if err != nil {
		return reshootFramesView{}, err
	}
	if duration <= 0 {
		return reshootFramesView{}, fmt.Errorf("%w: 原视频时长未知，无法框选片段", screenplay.ErrInvalidState)
	}
	defer os.RemoveAll(filepath.Dir(source))
	directory := filepath.Join(filepath.Dir(source), "frames")
	runner := &mergeworker.FFmpegRunner{}
	paths, err := runner.ExtractFrames(ctx, source, duration, screenplay.ReshootFrameCount, 480, directory)
	if err != nil {
		return reshootFramesView{}, fmt.Errorf("%w: 抽取预览帧失败：%v", screenplay.ErrInvalidState, err)
	}
	frames := make([]string, 0, len(paths))
	for _, path := range paths {
		encoded, err := os.ReadFile(path)
		if err != nil {
			return reshootFramesView{}, err
		}
		frames = append(frames, "data:image/jpeg;base64,"+base64.StdEncoding.EncodeToString(encoded))
	}
	first, last := "", ""
	if raw, err := os.ReadFile(paths[0]); err == nil {
		first = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(raw)
	}
	if raw, err := os.ReadFile(paths[len(paths)-1]); err == nil {
		last = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(raw)
	}
	return reshootFramesView{ShotID: in.ShotID, Duration: duration, Frames: frames, FirstFrame: first, LastFrame: last}, nil
}

func (s *ScreenplayMediaWorkbench) startReshoot(ctx context.Context, p screenplay.Project, in reshootRequest) (reshootTaskView, error) {
	if s.Video == nil || s.Configs == nil || s.ReferencesProvider == nil {
		return reshootTaskView{}, fmt.Errorf("%w: 视频生成通道或首尾帧参考未配置", screenplay.ErrInvalidState)
	}
	board, shot, err := s.findShotBoard(ctx, p, in.ShotID)
	if err != nil {
		return reshootTaskView{}, err
	}
	_, source, duration, err := s.shotVideoPath(ctx, p, shot.ID)
	if err != nil {
		return reshootTaskView{}, err
	}
	defer os.RemoveAll(filepath.Dir(source))
	if duration <= 0 {
		return reshootTaskView{}, fmt.Errorf("%w: 原视频时长未知，无法重拍", screenplay.ErrInvalidState)
	}
	start, end := in.Start, in.End
	if end > duration {
		end = duration
	}
	if err = screenplay.ValidateReshootRange(start, end); err != nil {
		return reshootTaskView{}, err
	}
	state, err := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if err != nil {
		return reshootTaskView{}, err
	}
	if existing, ok := state.FindReshoot(shot.ID); ok && existing.Status == screenplay.ReshootStatusGenerating {
		return reshootTaskView{}, fmt.Errorf("%w: 这条分镜正在重拍，请等待当前任务", screenplay.ErrInvalidState)
	}
	config, err := s.Configs.GetConfig(ctx, p.OwnerUsername)
	if err != nil {
		return reshootTaskView{}, fmt.Errorf("%w: 读取视频模型配置失败", screenplay.ErrInvalidState)
	}
	modelKey, resolution := screenplayVideoDefaults(config)
	if modelKey == "" {
		return reshootTaskView{}, fmt.Errorf("%w: 请先在设置中保存一个可用的视频模型", screenplay.ErrInvalidState)
	}
	directory := filepath.Join(filepath.Dir(source), "refs")
	runner := &mergeworker.FFmpegRunner{}
	firstPath := filepath.Join(directory, "first.jpg")
	lastPath := filepath.Join(directory, "last.jpg")
	if err = runner.ExtractFrame(ctx, source, start, firstPath); err != nil {
		return reshootTaskView{}, fmt.Errorf("%w: 抽取首帧失败：%v", screenplay.ErrInvalidState, err)
	}
	lastAt := end - 0.05
	if lastAt < start {
		lastAt = start
	}
	if err = runner.ExtractFrame(ctx, source, lastAt, lastPath); err != nil {
		return reshootTaskView{}, fmt.Errorf("%w: 抽取尾帧失败：%v", screenplay.ErrInvalidState, err)
	}
	firstURL, err := s.saveReshootFrame(ctx, p, firstPath, "首帧")
	if err != nil {
		return reshootTaskView{}, err
	}
	lastURL, err := s.saveReshootFrame(ctx, p, lastPath, "尾帧")
	if err != nil {
		return reshootTaskView{}, err
	}
	task := screenplay.ReshootTask{
		ShotID:   shot.ID,
		Seq:      shot.Seq,
		Episode:  board.EpisodeNo,
		Engine:   string(board.Engine),
		Start:    start,
		End:      end,
		Duration: screenplay.NormalizeReshootDuration(in.Duration, start, end),
		Prompt:   strings.TrimSpace(in.Prompt),
		Mode:     screenplay.ReshootModeFrames,
		Status:   screenplay.ReshootStatusGenerating,
		Progress: 8,
	}
	input := scriptVideoLocalInput{
		SourceTaskID:   fmt.Sprintf("screenplay-reshoot:%s:%s:%d", p.ID, shot.ID, time.Now().UnixNano()),
		ModelKey:       modelKey,
		Prompt:         screenplay.ReshootPrompt(task.Prompt),
		Duration:       task.Duration,
		AspectRatio:    p.AspectRatio,
		Resolution:     resolution,
		GenerateAudio:  true,
		GenerationMode: "image_to_video",
		ImageURLs:      []string{firstURL, lastURL},
	}
	submission, err := s.Video.Submit(ctx, p.OwnerUsername, input)
	if err != nil {
		return reshootTaskView{}, fmt.Errorf("%w: 提交重拍任务失败：%v", screenplay.ErrInvalidState, err)
	}
	if strings.TrimSpace(submission.TaskID) == "" {
		return reshootTaskView{}, fmt.Errorf("%w: 供应商未返回任务编号，已停止重复提交", screenplay.ErrInvalidState)
	}
	task.ProviderID = submission.TaskID
	task.Progress = 12
	if _, err = s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(state *screenplay.MediaState) error {
		state.PutReshoot(task)
		return nil
	}); err != nil {
		return reshootTaskView{}, err
	}
	return s.reshootTaskView(ctx, BridgeAuth{}, p, task), nil
}

// saveReshootFrame 把首尾帧写入对象存储并返回视频模型可访问的地址。
func (s *ScreenplayMediaWorkbench) saveReshootFrame(ctx context.Context, p screenplay.Project, path, label string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	id := newAgentID("reshoot-frame")
	key, err := agentworkspace.AgentAssetObjectKey(p.OwnerUsername, p.ID, id, "jpg")
	if err != nil {
		return "", err
	}
	asset := agentworkspace.Asset{ID: id, ProjectID: p.ID, Kind: "image", Name: label, ContentType: "image/jpeg", ObjectKey: key, ByteSize: int64(len(raw)), CreatedAt: time.Now().UTC()}
	if err = s.Objects.Put(ctx, asset, bytes.NewReader(raw)); err != nil {
		return "", err
	}
	url, err := s.ReferencesProvider.ReferenceURL(ctx, asset, 2*time.Hour)
	if err != nil {
		return "", fmt.Errorf("%w: 首尾帧参考地址生成失败：%v", screenplay.ErrInvalidState, err)
	}
	return url, nil
}

func (s *ScreenplayMediaWorkbench) reshootStatus(ctx context.Context, auth BridgeAuth, p screenplay.Project, shotID string) (reshootTaskView, error) {
	shotID = strings.TrimSpace(shotID)
	if shotID == "" {
		return reshootTaskView{}, fmt.Errorf("%w: 请选择分镜", screenplay.ErrInvalidArgument)
	}
	task, err := s.reconcileReshoot(ctx, p, shotID)
	if err != nil {
		return reshootTaskView{}, err
	}
	return s.reshootTaskView(ctx, auth, p, task), nil
}

// reconcileReshoot 在查询状态时顺带向供应商核对一次，完成后立即导入素材库。
// 这样关闭窗口或重启服务后，只要任务号还在就能继续查询，不会重复提交。
func (s *ScreenplayMediaWorkbench) reconcileReshoot(ctx context.Context, p screenplay.Project, shotID string) (screenplay.ReshootTask, error) {
	state, err := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if err != nil {
		return screenplay.ReshootTask{}, err
	}
	task, ok := state.FindReshoot(shotID)
	if !ok {
		return screenplay.ReshootTask{}, fmt.Errorf("%w: 这条分镜还没有重拍记录", screenplay.ErrNotFound)
	}
	if task.Status != screenplay.ReshootStatusGenerating || task.ProviderID == "" {
		return task, nil
	}
	key := reshootSyncKey(p.OwnerUsername, p.ID, shotID)
	reshootSyncs.Lock()
	if reshootSyncs.active[key] {
		reshootSyncs.Unlock()
		return task, nil
	}
	reshootSyncs.active[key] = true
	reshootSyncs.Unlock()
	defer func() {
		reshootSyncs.Lock()
		delete(reshootSyncs.active, key)
		reshootSyncs.Unlock()
	}()
	if s.Video == nil {
		return task, nil
	}
	status, err := s.Video.Poll(ctx, p.OwnerUsername, task.ProviderID)
	if err != nil {
		return task, nil
	}
	switch strings.ToLower(strings.TrimSpace(status.Status)) {
	case "succeeded", "success", "completed", "done":
		artifact := strings.TrimSpace(status.VideoURL)
		if artifact == "" && strings.TrimSpace(status.ArtifactID) != "" {
			artifact = "/api/shuihuo-production/local-executor-artifacts/" + strings.TrimSpace(status.ArtifactID)
		}
		if artifact == "" {
			return s.failReshoot(p, shotID, "生成完成但没有视频地址")
		}
		item, importErr := s.importReshoot(ctx, p, task, artifact)
		if importErr != nil {
			return s.failReshoot(p, shotID, "已生成但导入失败："+importErr.Error())
		}
		updated := task
		updated.Status = screenplay.ReshootStatusDone
		updated.Progress = 100
		updated.ItemID = item.ID
		updated.Error = ""
		_, _ = s.Media.MutateMedia(context.Background(), p.OwnerUsername, p.ID, func(state *screenplay.MediaState) error {
			state.PutReshoot(updated)
			return nil
		})
		return updated, nil
	case "failed", "error", "cancelled":
		return s.failReshoot(p, shotID, firstText(status.Error, "视频通道返回失败"))
	}
	progress := task.Progress + 4
	if progress > 92 {
		progress = 92
	}
	task.Progress = progress
	_, _ = s.Media.MutateMedia(context.Background(), p.OwnerUsername, p.ID, func(state *screenplay.MediaState) error {
		state.PutReshoot(task)
		return nil
	})
	return task, nil
}

func (s *ScreenplayMediaWorkbench) failReshoot(p screenplay.Project, shotID, message string) (screenplay.ReshootTask, error) {
	result := screenplay.ReshootTask{ShotID: shotID}
	_, _ = s.Media.MutateMedia(context.Background(), p.OwnerUsername, p.ID, func(state *screenplay.MediaState) error {
		state.MutateReshoot(shotID, func(task *screenplay.ReshootTask) {
			task.Status = screenplay.ReshootStatusFailed
			task.Error = message
		})
		if found, ok := state.FindReshoot(shotID); ok {
			result = found
		}
		return nil
	})
	return result, nil
}

// importReshoot 把供应商生成的新片段导入素材库，重试不会重复提交生成任务。
func (s *ScreenplayMediaWorkbench) importReshoot(ctx context.Context, p screenplay.Project, task screenplay.ReshootTask, artifact string) (screenplay.MediaItem, error) {
	if s.Assets == nil || s.Objects == nil || s.Cut == nil {
		return screenplay.MediaItem{}, fmt.Errorf("素材存储未配置")
	}
	itemID := newAgentID("sp-reshoot")
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
	name := fmt.Sprintf("重拍片段-第%d镜.mp4", task.Seq)
	item := screenplay.MediaItem{ID: itemID, AssetID: asset.ID, Name: name, Kind: "video", Status: "succeeded", TargetID: task.ShotID, ProviderID: task.ProviderID, DurationMS: cutAsset.DurationMS, CreatedAt: time.Now().UTC()}
	return item, s.append(ctx, p, item)
}

func (s *ScreenplayMediaWorkbench) applyReshoot(ctx context.Context, auth BridgeAuth, p screenplay.Project, shotID string) (reshootTaskView, error) {
	task, err := s.reconcileReshoot(ctx, p, shotID)
	if err != nil {
		return reshootTaskView{}, err
	}
	if task.Status != screenplay.ReshootStatusDone || task.ItemID == "" {
		return reshootTaskView{}, fmt.Errorf("%w: 还没有生成好的重拍片段", screenplay.ErrInvalidState)
	}
	if s.Cut == nil || s.Objects == nil {
		return reshootTaskView{}, fmt.Errorf("%w: 素材存储未配置", screenplay.ErrInvalidState)
	}
	if _, err := mergeworkerProbeFFmpeg(); err != nil {
		return reshootTaskView{}, fmt.Errorf("%w: %v", screenplay.ErrInvalidState, err)
	}
	_, sourcePath, _, err := s.shotVideoPath(ctx, p, shotID)
	if err != nil {
		return reshootTaskView{}, err
	}
	defer os.RemoveAll(filepath.Dir(sourcePath))
	state, err := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if err != nil {
		return reshootTaskView{}, err
	}
	var segment screenplay.MediaItem
	for _, item := range state.Items {
		if item.ID == task.ItemID {
			segment = item
			break
		}
	}
	if segment.AssetID == "" {
		return reshootTaskView{}, fmt.Errorf("%w: 重拍片段素材不存在", screenplay.ErrNotFound)
	}
	segmentAsset, err := s.Cut.Asset(ctx, p.OwnerUsername, p.ID, segment.AssetID)
	if err != nil {
		return reshootTaskView{}, err
	}
	segmentReader, err := s.Objects.Open(ctx, aicutObjectAsset(segmentAsset))
	if err != nil {
		return reshootTaskView{}, err
	}
	segmentPath := filepath.Join(filepath.Dir(sourcePath), "segment.mp4")
	segmentFile, err := os.Create(segmentPath)
	if err != nil {
		segmentReader.Close()
		return reshootTaskView{}, err
	}
	_, copyErr := io.Copy(segmentFile, segmentReader)
	segmentReader.Close()
	segmentFile.Close()
	if copyErr != nil {
		return reshootTaskView{}, copyErr
	}
	outputPath := filepath.Join(filepath.Dir(sourcePath), "applied.mp4")
	runner := &mergeworker.FFmpegRunner{}
	if err = runner.Splice(ctx, sourcePath, task.Start, task.End, segmentPath, outputPath, p.AspectRatio); err != nil {
		return reshootTaskView{}, fmt.Errorf("%w: 拼接失败：%v", screenplay.ErrInvalidState, err)
	}
	name := fmt.Sprintf("片段重拍成片-第%d镜.mp4", task.Seq)
	saved, err := s.saveStreamedVideoFile(ctx, p, outputPath, name, shotID)
	if err != nil {
		return reshootTaskView{}, fmt.Errorf("%w: 成片保存失败：%v", screenplay.ErrInvalidState, err)
	}
	now := time.Now().UTC()
	task.Status = screenplay.ReshootStatusApplied
	task.AppliedAt = &now
	task.AppliedID = saved.ID
	_, err = s.Media.MutateMedia(ctx, p.OwnerUsername, p.ID, func(state *screenplay.MediaState) error {
		state.PutReshoot(task)
		return nil
	})
	if err != nil {
		return reshootTaskView{}, err
	}
	return s.reshootTaskView(ctx, auth, p, task), nil
}

func (s *ScreenplayMediaWorkbench) reshootTaskView(ctx context.Context, auth BridgeAuth, p screenplay.Project, task screenplay.ReshootTask) reshootTaskView {
	view := reshootTaskView{ReshootTask: task}
	state, err := s.Media.LoadMedia(ctx, p.OwnerUsername, p.ID)
	if err != nil {
		return view
	}
	if item, ok := shotVideoItem(state, task.ShotID); ok && item.ID != task.ItemID {
		view.SourceURL = aicutSignedAssetContentURL(auth, p.OwnerUsername, p.ID, item.AssetID, "content", time.Now())
	}
	for _, item := range state.Items {
		if item.AssetID == "" || item.Status != "succeeded" {
			continue
		}
		switch item.ID {
		case task.ItemID:
			view.SegmentURL = aicutSignedAssetContentURL(auth, p.OwnerUsername, p.ID, item.AssetID, "content", time.Now())
		case task.AppliedID:
			view.AppliedURL = aicutSignedAssetContentURL(auth, p.OwnerUsername, p.ID, item.AssetID, "content", time.Now())
		}
	}
	return view
}

