package httpapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/ffmpegbin"
)

var runAICutFFmpeg = ffmpegbin.Run

type aicutExportJob struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"projectId"`
	Format    string    `json:"format"`
	State     string    `json:"state"`
	Progress  int       `json:"progress"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	owner     string
	output    string
	cleanup   func()
	cancel    context.CancelFunc
}

type aicutExportManager struct {
	mu   sync.Mutex
	jobs map[string]aicutExportJob
}

func newAICutExportManager() *aicutExportManager {
	return &aicutExportManager{jobs: make(map[string]aicutExportJob)}
}

func (m *aicutExportManager) create(owner, projectID, format string, now time.Time) (aicutExportJob, context.Context) {
	ctx, cancel := context.WithCancel(context.Background())
	job := aicutExportJob{ID: newAgentID("export"), ProjectID: projectID, Format: format, State: "queued", CreatedAt: now, UpdatedAt: now, owner: owner, cancel: cancel}
	m.mu.Lock()
	m.jobs[job.ID] = job
	m.mu.Unlock()
	return job, ctx
}

func (m *aicutExportManager) get(owner, projectID, id string) (aicutExportJob, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	job, ok := m.jobs[id]
	return job, ok && job.owner == owner && job.ProjectID == projectID
}

func (m *aicutExportManager) update(id, state string, progress int, message, output string, cleanup func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	job, ok := m.jobs[id]
	if !ok {
		if cleanup != nil {
			cleanup()
		}
		return
	}
	job.State, job.Progress, job.Error, job.UpdatedAt = state, max(0, min(progress, 100)), strings.TrimSpace(message), time.Now().UTC()
	if output != "" {
		job.output = output
	}
	if cleanup != nil {
		job.cleanup = cleanup
	}
	m.jobs[id] = job
}

func (m *aicutExportManager) cancel(owner, projectID, id string) (aicutExportJob, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	job, ok := m.jobs[id]
	if !ok || job.owner != owner || job.ProjectID != projectID {
		return aicutExportJob{}, false
	}
	if job.State == "queued" || job.State == "running" {
		job.State = "cancel_requested"
		job.UpdatedAt = time.Now().UTC()
		m.jobs[id] = job
		job.cancel()
	}
	return job, true
}

func (m *aicutExportManager) expire(id string) {
	m.mu.Lock()
	job, ok := m.jobs[id]
	if ok {
		delete(m.jobs, id)
	}
	m.mu.Unlock()
	if ok && job.cleanup != nil {
		job.cleanup()
	}
}

type aicutRenderClip struct {
	clip         aicut.Clip
	asset        aicut.Asset
	path         string
	layerEffects []aicut.Clip
}

func RegisterAICutExportRoutes(mux *http.ServeMux, auth BridgeAuth, store aicut.Store, objects agentworkspace.AssetObjectStore) {
	if mux == nil || store == nil {
		return
	}
	manager := newAICutExportManager()
	authenticated := func(fn func(http.ResponseWriter, *http.Request, string)) http.Handler {
		return auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity, ok := BridgeIdentityFromContext(req.Context())
			if !ok || strings.TrimSpace(identity.Username) == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			fn(w, req, strings.TrimSpace(identity.Username))
		}))
	}
	mux.Handle("POST /api/ai-cut/projects/{id}/export-jobs", authenticated(func(w http.ResponseWriter, req *http.Request, owner string) {
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "素材存储尚未配置"})
			return
		}
		var input struct {
			Format           string `json:"format"`
			ExpectedRevision int64  `json:"expectedRevision"`
		}
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		input.Format = strings.ToLower(strings.TrimSpace(input.Format))
		if input.Format != "mp4" && input.Format != "webm" && input.Format != "mp3" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "仅支持 MP4、WebM 或 MP3 导出"})
			return
		}
		projectID := strings.TrimSpace(req.PathValue("id"))
		project, err := store.Project(req.Context(), owner, projectID)
		if err != nil {
			writeAICutError(w, err)
			return
		}
		timeline, err := store.Timeline(req.Context(), owner, projectID)
		if err != nil {
			writeAICutError(w, err)
			return
		}
		if timeline.Revision != input.ExpectedRevision {
			writeAICutError(w, aicut.ErrRevisionConflict)
			return
		}
		job, renderCtx := manager.create(owner, projectID, input.Format, time.Now().UTC())
		go func() {
			manager.update(job.ID, "running", 10, "", "", nil)
			output, cleanup, renderErr := renderAICutExport(renderCtx, owner, project, timeline, input.Format, store, objects)
			if renderErr != nil {
				if cleanup != nil {
					cleanup()
				}
				if errors.Is(renderCtx.Err(), context.Canceled) {
					manager.update(job.ID, "canceled", 100, "", "", nil)
				} else {
					manager.update(job.ID, "failed", 100, renderErr.Error(), "", nil)
				}
				return
			}
			manager.update(job.ID, "succeeded", 100, "", output, cleanup)
			time.AfterFunc(time.Hour, func() { manager.expire(job.ID) })
		}()
		writeJSON(w, http.StatusAccepted, map[string]any{"job": job})
	}))
	mux.Handle("GET /api/ai-cut/projects/{id}/export-jobs/{jobId}", authenticated(func(w http.ResponseWriter, req *http.Request, owner string) {
		job, ok := manager.get(owner, strings.TrimSpace(req.PathValue("id")), strings.TrimSpace(req.PathValue("jobId")))
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "导出任务不存在或已清理"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"job": job})
	}))
	mux.Handle("POST /api/ai-cut/projects/{id}/export-jobs/{jobId}/cancel", authenticated(func(w http.ResponseWriter, req *http.Request, owner string) {
		job, ok := manager.cancel(owner, strings.TrimSpace(req.PathValue("id")), strings.TrimSpace(req.PathValue("jobId")))
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "导出任务不存在或已清理"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"job": job})
	}))
	mux.Handle("GET /api/ai-cut/projects/{id}/export-jobs/{jobId}/download", authenticated(func(w http.ResponseWriter, req *http.Request, owner string) {
		job, ok := manager.get(owner, strings.TrimSpace(req.PathValue("id")), strings.TrimSpace(req.PathValue("jobId")))
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "导出任务不存在或已清理"})
			return
		}
		if job.State != "succeeded" || job.output == "" {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "导出尚未完成"})
			return
		}
		file, err := os.Open(job.output)
		if err != nil {
			writeJSON(w, http.StatusGone, map[string]string{"error": "导出文件已清理，请重新导出"})
			return
		}
		defer file.Close()
		contentTypes := map[string]string{"mp4": "video/mp4", "webm": "video/webm", "mp3": "audio/mpeg"}
		w.Header().Set("Content-Type", contentTypes[job.Format])
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=ai-cut.%s", job.Format))
		if info, statErr := file.Stat(); statErr == nil {
			w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
		}
		_, _ = io.Copy(w, file)
	}))
	mux.Handle("POST /api/ai-cut/projects/{id}/exports/{format}", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok || strings.TrimSpace(identity.Username) == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "素材存储尚未配置"})
			return
		}
		format := strings.ToLower(strings.TrimSpace(req.PathValue("format")))
		if format != "mp4" && format != "webm" && format != "mp3" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "仅支持 MP4、WebM 或 MP3 导出"})
			return
		}
		var input struct {
			ExpectedRevision int64 `json:"expectedRevision"`
		}
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		owner, projectID := strings.TrimSpace(identity.Username), strings.TrimSpace(req.PathValue("id"))
		project, err := store.Project(req.Context(), owner, projectID)
		if err != nil {
			writeAICutError(w, err)
			return
		}
		timeline, err := store.Timeline(req.Context(), owner, projectID)
		if err != nil {
			writeAICutError(w, err)
			return
		}
		if timeline.Revision != input.ExpectedRevision {
			writeAICutError(w, aicut.ErrRevisionConflict)
			return
		}
		output, cleanup, err := renderAICutExport(req.Context(), owner, project, timeline, format, store, objects)
		if cleanup != nil {
			defer cleanup()
		}
		if err != nil {
			status := http.StatusUnprocessableEntity
			if ffmpegbin.IsMissing(err) {
				status = http.StatusServiceUnavailable
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		file, err := os.Open(output)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "导出文件读取失败"})
			return
		}
		defer file.Close()
		contentTypes := map[string]string{"mp4": "video/mp4", "webm": "video/webm", "mp3": "audio/mpeg"}
		w.Header().Set("Content-Type", contentTypes[format])
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=ai-cut.%s", format))
		if info, statErr := file.Stat(); statErr == nil {
			w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
		}
		_, _ = io.Copy(w, file)
	})))
}

func renderAICutExport(ctx context.Context, owner string, project aicut.Project, timeline aicut.Timeline, format string, store aicut.Store, objects agentworkspace.AssetObjectStore) (string, func(), error) {
	dir, err := os.MkdirTemp("", "qiantie-aicut-render-")
	if err != nil {
		return "", nil, fmt.Errorf("无法创建导出临时目录")
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	clips := make([]aicut.Clip, 0)
	effectLayers := make([]aicut.Clip, 0)
	for _, track := range timeline.Tracks {
		if track.Kind == aicut.TrackEffect && !track.Muted {
			effectLayers = append(effectLayers, track.Clips...)
		}
		include := track.Kind == aicut.TrackAudio
		if format != "mp3" {
			include = include || track.Kind == aicut.TrackVideo
		}
		if include && !track.Muted {
			clips = append(clips, track.Clips...)
		}
	}
	sort.SliceStable(clips, func(i, j int) bool { return clips[i].StartMS < clips[j].StartMS })
	if len(clips) == 0 {
		cleanup()
		if format == "mp3" {
			return "", nil, fmt.Errorf("时间线里没有可导出的音频片段")
		}
		return "", nil, fmt.Errorf("时间线里没有可导出的画面片段")
	}
	prepared := make([]aicutRenderClip, 0, len(clips))
	for index, clip := range clips {
		if strings.TrimSpace(clip.AssetID) == "" {
			cleanup()
			return "", nil, fmt.Errorf("片段“%s”缺少素材", clip.Name)
		}
		asset, assetErr := store.Asset(ctx, owner, project.ID, clip.AssetID)
		if assetErr != nil {
			cleanup()
			return "", nil, fmt.Errorf("片段“%s”的素材不存在", clip.Name)
		}
		if format == "mp3" && asset.Kind != "audio" {
			cleanup()
			return "", nil, fmt.Errorf("片段“%s”不是音频素材", clip.Name)
		}
		if format != "mp3" && asset.Kind != "video" && asset.Kind != "image" && asset.Kind != "audio" {
			cleanup()
			return "", nil, fmt.Errorf("片段“%s”不是可导出媒体素材", clip.Name)
		}
		body, openErr := objects.Open(ctx, aicutObjectAsset(asset))
		if openErr != nil {
			cleanup()
			return "", nil, fmt.Errorf("素材“%s”读取失败", asset.Name)
		}
		ext := filepath.Ext(asset.Name)
		if ext == "" {
			ext = ".bin"
		}
		inputPath := filepath.Join(dir, fmt.Sprintf("input-%03d%s", index, ext))
		outputFile, createErr := os.Create(inputPath)
		if createErr == nil {
			_, createErr = io.Copy(outputFile, body)
			if closeErr := outputFile.Close(); createErr == nil {
				createErr = closeErr
			}
		}
		_ = body.Close()
		if createErr != nil {
			cleanup()
			return "", nil, fmt.Errorf("素材“%s”准备失败", asset.Name)
		}
		item := aicutRenderClip{clip: clip, asset: asset, path: inputPath}
		if asset.Kind == "video" || asset.Kind == "image" {
			clipEnd := clip.StartMS + clip.DurationMS
			for _, effect := range effectLayers {
				if effect.StartMS < clipEnd && effect.StartMS+effect.DurationMS > clip.StartMS {
					item.layerEffects = append(item.layerEffects, effect)
				}
			}
		}
		prepared = append(prepared, item)
	}
	output := filepath.Join(dir, "output."+format)
	subtitlePath := ""
	if format != "mp3" {
		width, height := aicutRenderSize(project.AspectRatio)
		if content := buildAICutASS(timeline, width, height); content != "" {
			subtitlePath = filepath.Join(dir, "subtitles.ass")
			if writeErr := os.WriteFile(subtitlePath, []byte(content), 0o600); writeErr != nil {
				cleanup()
				return "", nil, fmt.Errorf("字幕渲染文件准备失败")
			}
		}
	}
	args, err := buildAICutFFmpegArgs(project.AspectRatio, prepared, format, output, subtitlePath)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	renderCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()
	if err := runAICutFFmpeg(renderCtx, "ffmpeg", args...); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("AI-CUT 渲染失败：%w", err)
	}
	return output, cleanup, nil
}

func buildAICutFFmpegArgs(aspectRatio string, clips []aicutRenderClip, format, output string, subtitlePaths ...string) ([]string, error) {
	if len(clips) == 0 {
		return nil, fmt.Errorf("没有可导出的片段")
	}
	args := []string{"-hide_banner", "-y"}
	filters := make([]string, 0, len(clips)+2)
	videoLabels := make([]string, 0, len(clips))
	videoItems := make([]aicutRenderClip, 0, len(clips))
	audioLabels := make([]string, 0, len(clips))
	visualClips := make([]aicutRenderClip, 0, len(clips))
	for _, item := range clips {
		if item.asset.Kind == "video" || item.asset.Kind == "image" {
			visualClips = append(visualClips, item)
		}
	}
	timelineComposition := false
	if len(visualClips) > 0 {
		expectedStart := int64(0)
		for _, item := range visualClips {
			if item.clip.StartMS != expectedStart {
				timelineComposition = true
				break
			}
			expectedStart = item.clip.StartMS + item.clip.DurationMS
		}
	}
	for i, item := range clips {
		duration := float64(item.clip.DurationMS) / 1000
		if item.asset.Kind == "image" {
			args = append(args, "-loop", "1", "-t", fmt.Sprintf("%.3f", duration), "-i", item.path)
		} else {
			args = append(args, "-i", item.path)
		}
		if item.asset.Kind == "audio" {
			start := float64(item.clip.SourceStartMS) / 1000
			delay := item.clip.StartMS
			volume := item.clip.Volume
			if volume <= 0 {
				volume = 1
			}
			chain := fmt.Sprintf("[%d:a]atrim=start=%.3f:duration=%.3f,asetpts=PTS-STARTPTS,volume=%.3f", i, start, duration, volume)
			if item.clip.FadeInMS > 0 {
				chain += fmt.Sprintf(",afade=t=in:st=0:d=%.3f", float64(item.clip.FadeInMS)/1000)
			}
			if item.clip.FadeOutMS > 0 {
				fadeDuration := float64(item.clip.FadeOutMS) / 1000
				chain += fmt.Sprintf(",afade=t=out:st=%.3f:d=%.3f", max(0, duration-fadeDuration), fadeDuration)
			}
			filters = append(filters, chain+fmt.Sprintf(",adelay=%d|%d[a%d]", delay, delay, i))
			audioLabels = append(audioLabels, fmt.Sprintf("[a%d]", i))
			continue
		}
		if format == "mp3" {
			continue
		}
		width, height := aicutRenderSize(aspectRatio)
		start := float64(item.clip.SourceStartMS) / 1000
		rate := item.clip.PlaybackRate
		if rate <= 0 {
			rate = 1
		}
		setPTS := fmt.Sprintf("(PTS-STARTPTS)/%.3f", rate)
		if timelineComposition {
			setPTS += fmt.Sprintf("+%.3f/TB", float64(item.clip.StartMS)/1000)
		}
		chain := []string{
			fmt.Sprintf("[%d:v]trim=start=%.3f:duration=%.3f", i, start, duration*rate),
			"setpts=" + setPTS,
		}
		chain = append(chain, buildAICutVisualEffectFilters(item.clip.VisualEffect)...)
		if catalogFilters, ok := buildAICutCatalogResourceFilters(item.clip.CatalogResourceID, map[string]float64{"intensity": .5}); ok {
			chain = append(chain, catalogFilters...)
		}
		for _, effect := range item.layerEffects {
			if effectFilters, ok := buildAICutCatalogResourceFilters(effect.ResourceID, effect.Parameters); ok {
				chain = append(chain, effectFilters...)
			}
		}
		chain = append(chain,
			fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease", width, height),
			fmt.Sprintf("pad=%d:%d:(ow-iw)/2:(oh-ih)/2", width, height),
		)
		chain = append(chain, buildAICutCanvasTransformFilters(width, height, item.clip.CanvasTransform)...)
		chain = append(chain, "fps=30", "format=yuv420p"+fmt.Sprintf("[v%d]", i))
		filters = append(filters, strings.Join(chain, ","))
		videoLabels = append(videoLabels, fmt.Sprintf("[v%d]", i))
		videoItems = append(videoItems, item)
	}
	if format == "mp3" {
		if len(audioLabels) == 0 {
			return nil, fmt.Errorf("没有可导出的音频片段")
		}
		filters = append(filters, strings.Join(audioLabels, "")+fmt.Sprintf("amix=inputs=%d:duration=longest:normalize=0[outa]", len(audioLabels)))
		args = append(args, "-filter_complex", strings.Join(filters, ";"), "-map", "[outa]", "-c:a", "libmp3lame", "-q:a", "2", output)
		return args, nil
	}
	if len(videoLabels) == 0 {
		return nil, fmt.Errorf("没有可导出的画面片段")
	}
	subtitlePath := ""
	if len(subtitlePaths) > 0 {
		subtitlePath = strings.TrimSpace(subtitlePaths[0])
	}
	finalVideoLabel := "[outv]"
	if subtitlePath != "" {
		finalVideoLabel = "[basev]"
	}
	if timelineComposition {
		width, height := aicutRenderSize(aspectRatio)
		totalMS := int64(0)
		for _, item := range videoItems {
			totalMS = max(totalMS, item.clip.StartMS+item.clip.DurationMS)
		}
		args = append(args, "-f", "lavfi", "-t", fmt.Sprintf("%.3f", float64(totalMS)/1000), "-i", fmt.Sprintf("color=c=black:s=%dx%d:d=%.3f", width, height, float64(totalMS)/1000))
		currentLabel := fmt.Sprintf("[%d:v]", len(clips))
		for index, label := range videoLabels {
			outputLabel := fmt.Sprintf("[layer%d]", index)
			if index == len(videoLabels)-1 {
				outputLabel = finalVideoLabel
			}
			item := videoItems[index]
			start, end := float64(item.clip.StartMS)/1000, float64(item.clip.StartMS+item.clip.DurationMS)/1000
			filters = append(filters, fmt.Sprintf("%s%soverlay=eof_action=pass:enable='between(t,%.3f,%.3f)'%s", currentLabel, label, start, end, outputLabel))
			currentLabel = outputLabel
		}
	} else if len(videoLabels) == 1 {
		filters = append(filters, videoLabels[0]+"concat=n=1:v=1:a=0"+finalVideoLabel)
	} else {
		currentLabel := videoLabels[0]
		currentDuration := float64(videoItems[0].clip.DurationMS) / 1000
		for index := 1; index < len(videoLabels); index++ {
			outputLabel := fmt.Sprintf("[vx%d]", index)
			if index == len(videoLabels)-1 {
				outputLabel = finalVideoLabel
			}
			item := videoItems[index]
			transitionDuration := float64(item.clip.TransitionMS) / 1000
			if item.clip.Transition != "" && transitionDuration > 0 {
				transition, ok := aicutXFadeTransition(item.clip.Transition)
				if !ok {
					return nil, fmt.Errorf("不支持的转场：%s", item.clip.Transition)
				}
				transitionDuration = min(transitionDuration, currentDuration, float64(item.clip.DurationMS)/1000)
				offset := max(0, currentDuration-transitionDuration)
				filters = append(filters, fmt.Sprintf("%s%sxfade=transition=%s:duration=%.3f:offset=%.3f%s", currentLabel, videoLabels[index], transition, transitionDuration, offset, outputLabel))
				currentDuration += float64(item.clip.DurationMS)/1000 - transitionDuration
			} else {
				filters = append(filters, fmt.Sprintf("%s%sconcat=n=2:v=1:a=0%s", currentLabel, videoLabels[index], outputLabel))
				currentDuration += float64(item.clip.DurationMS) / 1000
			}
			currentLabel = outputLabel
		}
	}
	if subtitlePath != "" {
		filters = append(filters, fmt.Sprintf("[basev]ass='%s'[outv]", escapeAICutFilterPath(subtitlePath)))
	}
	if len(audioLabels) > 0 {
		filters = append(filters, strings.Join(audioLabels, "")+fmt.Sprintf("amix=inputs=%d:duration=longest:normalize=0[outa]", len(audioLabels)))
	}
	args = append(args, "-filter_complex", strings.Join(filters, ";"), "-map", "[outv]")
	if len(audioLabels) > 0 {
		args = append(args, "-map", "[outa]")
	}
	if format == "webm" {
		args = append(args, "-c:v", "libvpx-vp9", "-crf", "32", "-b:v", "0")
		if len(audioLabels) > 0 {
			args = append(args, "-c:a", "libopus")
		}
	} else {
		args = append(args, "-c:v", "libx264", "-preset", "medium", "-crf", "20", "-movflags", "+faststart")
		if len(audioLabels) > 0 {
			args = append(args, "-c:a", "aac")
		}
	}
	if len(audioLabels) > 0 {
		args = append(args, "-shortest")
	}
	args = append(args, output)
	return args, nil
}

func buildAICutVisualEffectFilters(effect *aicut.VisualEffect) []string {
	if effect == nil {
		return nil
	}
	filters := make([]string, 0, 3)
	if effect.Brightness != 0 || effect.Contrast != 1 || effect.Saturation != 1 {
		filters = append(filters, fmt.Sprintf("eq=brightness=%.3f:contrast=%.3f:saturation=%.3f", effect.Brightness, effect.Contrast, effect.Saturation))
	}
	if effect.Blur > 0 {
		filters = append(filters, fmt.Sprintf("gblur=sigma=%.3f", effect.Blur))
	}
	if effect.Grayscale {
		filters = append(filters, "hue=s=0")
	}
	intensity := effect.Intensity
	if intensity <= 0 {
		intensity = .5
	}
	speed := effect.Speed
	if speed <= 0 {
		speed = 1
	}
	switch effect.Preset {
	case "vignette":
		filters = append(filters, fmt.Sprintf("vignette=angle=%.3f", .2+.8*intensity))
	case "glow":
		filters = append(filters, fmt.Sprintf("gblur=sigma=%.3f", .5+2.5*intensity))
	case "shake":
		pixels := 2 + int(12*intensity)
		filters = append(filters, fmt.Sprintf("crop=iw-%d:ih-%d:x='%d+%d*sin(t*%.3f)':y='%d+%d*cos(t*%.3f)'", pixels*2, pixels*2, pixels, pixels, speed*10, pixels, pixels, speed*13))
	case "breathe":
		filters = append(filters, fmt.Sprintf("zoompan=z='min(1+%.4f*(1+sin(on*%.3f)),1.100)':d=1:fps=30", .02*intensity, speed*.12))
	case "flash":
		filters = append(filters, fmt.Sprintf("eq=brightness='%.3f*(1+sin(t*%.3f))':eval=frame", .25*intensity, speed*12))
	case "mosaic":
		block := 4 + int(28*intensity)
		filters = append(filters, fmt.Sprintf("pixelize=w=%d:h=%d", block, block))
	}
	return filters
}

func buildAICutCatalogResourceFilters(resourceID string, parameters map[string]float64) ([]string, bool) {
	if resourceID == "" {
		return nil, false
	}
	intensity := parameters["intensity"]
	if intensity <= 0 {
		intensity = .5
	}
	intensity = min(1, intensity)
	speed := parameters["speed"]
	if speed <= 0 {
		speed = 1
	}
	switch resourceID {
	case "filter-original":
		return nil, true
	case "filter-bright":
		return []string{fmt.Sprintf("eq=brightness=%.3f:saturation=%.3f", .12*intensity, 1+.08*intensity)}, true
	case "filter-clear":
		return []string{fmt.Sprintf("unsharp=5:5:%.3f:5:5:0", .5+intensity)}, true
	case "filter-natural":
		return []string{fmt.Sprintf("eq=contrast=%.3f:saturation=%.3f", 1+.04*intensity, 1+.04*intensity)}, true
	case "filter-warm-sun":
		return []string{fmt.Sprintf("colorbalance=rs=%.3f:gs=%.3f:bs=-%.3f", .12*intensity, .04*intensity, .08*intensity)}, true
	case "filter-cool-blue":
		return []string{fmt.Sprintf("colorbalance=rs=-%.3f:bs=%.3f", .08*intensity, .14*intensity)}, true
	case "filter-cinema":
		return []string{fmt.Sprintf("eq=contrast=%.3f:saturation=%.3f", 1+.18*intensity, 1-.12*intensity), "vignette=angle=0.35"}, true
	case "filter-retro":
		return []string{fmt.Sprintf("colorbalance=rs=%.3f:gs=%.3f:bs=-%.3f", .10*intensity, .05*intensity, .10*intensity), "noise=alls=4:allf=t"}, true
	case "filter-monochrome":
		return []string{"hue=s=0"}, true
	case "filter-high-contrast":
		return []string{fmt.Sprintf("eq=contrast=%.3f", 1+.45*intensity)}, true
	case "filter-low-saturation":
		return []string{fmt.Sprintf("eq=saturation=%.3f", 1-.7*intensity)}, true
	case "filter-teal-orange":
		return []string{fmt.Sprintf("colorbalance=rs=%.3f:gs=-%.3f:bs=%.3f", .12*intensity, .03*intensity, .10*intensity)}, true
	case "filter-night":
		return []string{fmt.Sprintf("eq=brightness=%.3f:gamma=%.3f:saturation=%.3f", .04*intensity, 1+.22*intensity, 1-.15*intensity)}, true
	case "filter-portrait":
		return []string{fmt.Sprintf("eq=brightness=%.3f:saturation=%.3f", .04*intensity, 1+.06*intensity), fmt.Sprintf("gblur=sigma=%.3f", .15+.25*intensity)}, true
	case "filter-vivid":
		return []string{fmt.Sprintf("eq=contrast=%.3f:saturation=%.3f", 1+.12*intensity, 1+.45*intensity)}, true
	case "filter-soft":
		return []string{fmt.Sprintf("eq=contrast=%.3f", 1-.16*intensity), fmt.Sprintf("gblur=sigma=%.3f", .25+.45*intensity)}, true
	case "effect-blur":
		radius := parameters["radius"]
		if radius <= 0 {
			radius = 2 + 8*intensity
		}
		return []string{fmt.Sprintf("gblur=sigma=%.3f", min(20, radius))}, true
	case "effect-vignette":
		return []string{fmt.Sprintf("vignette=angle=%.3f", .2+.8*intensity)}, true
	case "effect-glow":
		return []string{fmt.Sprintf("gblur=sigma=%.3f", .5+2.5*intensity)}, true
	case "effect-grain":
		return []string{fmt.Sprintf("noise=alls=%d:allf=t", 2+int(18*intensity))}, true
	case "effect-shake":
		pixels := 2 + int(12*intensity)
		return []string{fmt.Sprintf("crop=iw-%d:ih-%d:x='%d+%d*sin(t*%.3f)':y='%d+%d*cos(t*%.3f)'", pixels*2, pixels*2, pixels, pixels, speed*10, pixels, pixels, speed*13)}, true
	case "effect-breathe-zoom":
		return []string{fmt.Sprintf("zoompan=z='min(1+%.4f*(1+sin(on*%.3f)),1.100)':d=1:fps=30", .02*intensity, speed*.12)}, true
	case "effect-flash":
		return []string{fmt.Sprintf("eq=brightness='%.3f*(1+sin(t*%.3f))':eval=frame", .25*intensity, speed*12)}, true
	case "effect-glitch":
		return []string{fmt.Sprintf("rgbashift=rh=%d:bv=%d", 1+int(8*intensity), -1-int(8*intensity))}, true
	case "effect-chromatic":
		return []string{fmt.Sprintf("rgbashift=rh=%d:bh=%d", 1+int(6*intensity), -1-int(6*intensity))}, true
	case "effect-mosaic":
		block := 4 + int(28*intensity)
		return []string{fmt.Sprintf("scale=iw/%d:ih/%d,scale=iw*%d:ih*%d:flags=neighbor", block, block, block, block)}, true
	case "effect-fisheye":
		return []string{fmt.Sprintf("lenscorrection=k1=-%.3f:k2=%.3f", .20*intensity, .05*intensity)}, true
	case "effect-sharpen":
		return []string{fmt.Sprintf("unsharp=5:5:%.3f:5:5:0", .5+1.5*intensity)}, true
	default:
		return nil, false
	}
}

func buildAICutCanvasTransformFilters(width, height int, transform *aicut.CanvasTransform) []string {
	if transform == nil {
		return nil
	}
	scaleX, scaleY, opacity := transform.ScaleX, transform.ScaleY, transform.Opacity
	if scaleX == 0 {
		scaleX = 1
	}
	if scaleY == 0 {
		scaleY = 1
	}
	if opacity == 0 {
		opacity = 1
	}
	filters := make([]string, 0, 4)
	if transform.Rotation != 0 {
		filters = append(filters, fmt.Sprintf("rotate=%.6f*PI/180:ow=rotw(iw):oh=roth(ih):c=black@0", transform.Rotation))
	}
	if scaleX != 1 || scaleY != 1 {
		targetWidth, targetHeight := max(1, int(float64(width)*scaleX)), max(1, int(float64(height)*scaleY))
		filters = append(filters, fmt.Sprintf("scale=%d:%d", targetWidth, targetHeight))
		xOffset, yOffset := int(float64(width)*transform.X/2), int(float64(height)*transform.Y/2)
		if targetWidth <= width && targetHeight <= height {
			filters = append(filters, fmt.Sprintf("pad=%d:%d:%d:%d:color=black@0", width, height, max(0, (width-targetWidth)/2+xOffset), max(0, (height-targetHeight)/2+yOffset)))
		} else {
			filters = append(filters, fmt.Sprintf("crop=%d:%d:%d:%d", width, height, max(0, (targetWidth-width)/2-xOffset), max(0, (targetHeight-height)/2-yOffset)))
		}
	}
	if opacity != 1 {
		filters = append(filters, fmt.Sprintf("format=rgba,colorchannelmixer=aa=%.3f", opacity))
	}
	return filters
}

func aicutXFadeTransition(value string) (string, bool) {
	mapping := map[string]string{
		"cross-dissolve": "dissolve", "fade": "fade", "flash-black": "fadeblack", "flash-white": "fadewhite",
		"slide-left": "slideleft", "slide-right": "slideright", "slide-up": "slideup", "slide-down": "slidedown", "zoom": "zoomin", "blur": "hblur",
		"transition-cross-dissolve": "dissolve", "transition-fade": "fade", "transition-flash-black": "fadeblack", "transition-flash-white": "fadewhite",
		"transition-push-left": "slideleft", "transition-push-right": "slideright", "transition-push-up": "slideup", "transition-push-down": "slidedown",
		"transition-wipe-left": "wipeleft", "transition-wipe-right": "wiperight", "transition-wipe-up": "wipeup", "transition-wipe-down": "wipedown",
		"transition-zoom": "zoomin", "transition-blur": "hblur", "transition-glitch": "pixelize", "transition-light-leak": "fadewhite",
	}
	transition, ok := mapping[value]
	return transition, ok
}

func aicutRenderSize(aspectRatio string) (int, int) {
	if aspectRatio == "16:9" {
		return 1920, 1080
	}
	if aspectRatio == "1:1" {
		return 1080, 1080
	}
	return 1080, 1920
}

func buildAICutASS(timeline aicut.Timeline, width, height int) string {
	styles := make([]string, 0)
	events := make([]string, 0)
	styleIndex := 0
	for _, track := range timeline.Tracks {
		if track.Kind != aicut.TrackSubtitle || track.Muted {
			continue
		}
		styleName := fmt.Sprintf("Track%d", styleIndex)
		fontSize, primary, outline, back, borderStyle, outlineWidth := 54, "&H00FFFFFF", "&H00000000", "&H80000000", 1, 3
		styleID := track.Settings.SubtitleStyleID
		if track.Settings.TextResourceID != "" {
			styleID = track.Settings.TextResourceID
		}
		switch styleID {
		case "黄字大标题", "故事黄高亮", "yellow-emphasis", "word-highlight":
			fontSize, primary, outlineWidth = 72, "&H0000FFFF", 4
		case "黑底白字", "短视频撞色", "绿底堆叠", "translucent-box", "text-box":
			borderStyle, outlineWidth = 3, 0
		case "影院白字", "简洁白字", "classic-outline":
			outlineWidth = 2
		case "粗描边", "黑描边", "short-drama", "text-title":
			fontSize = 72
			outlineWidth = 6
		case "dual-color":
			primary, outline, outlineWidth = "&H00FFFFFF", "&H005040E0", 4
		case "text-subtitle":
			fontSize = 42
		case "text-vertical":
			fontSize = 58
		case "text-basic-caption":
			fontSize, outlineWidth = 48, 2
		case "text-bold-title", "text-gradient-title":
			fontSize, outlineWidth = 82, 5
		case "text-outline-caption", "text-comic":
			fontSize, outlineWidth = 64, 6
		case "text-shadow-caption", "text-handwriting":
			fontSize, outlineWidth = 58, 2
		case "text-panel-caption", "text-news-bar", "text-lower-third":
			fontSize, borderStyle, outlineWidth = 52, 3, 0
		case "text-word-highlight":
			fontSize, primary, outlineWidth = 64, "&H0000FFFF", 4
		}
		styles = append(styles, fmt.Sprintf("Style: %s,Microsoft YaHei,%d,%s,&H000000FF,%s,%s,0,0,0,0,100,100,0,0,%d,%d,0,2,40,40,70,1", styleName, fontSize, primary, outline, back, borderStyle, outlineWidth))
		for _, clip := range track.Clips {
			text := strings.TrimSpace(clip.Text)
			if text == "" || clip.DurationMS <= 0 {
				continue
			}
			prefix := ""
			switch track.Settings.SubtitleAnimation {
			case "", "none", "无动效":
			case "pop":
				prefix = "{\\fad(80,120)\\fscx118\\fscy118\\t(0,240,\\fscx100\\fscy100)}"
			case "float-up":
				prefix = fmt.Sprintf("{\\move(%d,%d,%d,%d,0,350)\\fad(100,120)}", width/2, height-35, width/2, height-70)
			case "typewriter":
				prefix = fmt.Sprintf("{\\clip(0,0,0,%d)\\t(0,1200,\\clip(0,0,%d,%d))}", height, width, height)
			case "word-highlight":
				prefix = "{\\fad(100,100)\\1c&H00FFFF&}"
			default:
				prefix = "{\\fad(200,200)}"
			}
			events = append(events, fmt.Sprintf("Dialogue: 0,%s,%s,%s,,0,0,0,,%s%s", aicutASSTime(clip.StartMS), aicutASSTime(clip.StartMS+clip.DurationMS), styleName, prefix, escapeAICutASSText(text)))
		}
		styleIndex++
	}
	if len(events) == 0 {
		return ""
	}
	return fmt.Sprintf("[Script Info]\nScriptType: v4.00+\nPlayResX: %d\nPlayResY: %d\nWrapStyle: 2\n\n[V4+ Styles]\nFormat: Name,Fontname,Fontsize,PrimaryColour,SecondaryColour,OutlineColour,BackColour,Bold,Italic,Underline,StrikeOut,ScaleX,ScaleY,Spacing,Angle,BorderStyle,Outline,Shadow,Alignment,MarginL,MarginR,MarginV,Encoding\n%s\n\n[Events]\nFormat: Layer,Start,End,Style,Name,MarginL,MarginR,MarginV,Effect,Text\n%s\n", width, height, strings.Join(styles, "\n"), strings.Join(events, "\n"))
}

func aicutASSTime(ms int64) string {
	if ms < 0 {
		ms = 0
	}
	hours := ms / 3600000
	minutes := (ms / 60000) % 60
	seconds := (ms / 1000) % 60
	hundredths := (ms % 1000) / 10
	return fmt.Sprintf("%d:%02d:%02d.%02d", hours, minutes, seconds, hundredths)
}

func escapeAICutASSText(value string) string {
	return strings.NewReplacer("\\", "\\\\", "\r\n", "\\N", "\n", "\\N", "\r", "\\N", "{", "\\{", "}", "\\}").Replace(value)
}

func escapeAICutFilterPath(value string) string {
	return strings.NewReplacer("\\", "/", ":", "\\:", "'", "\\'", ",", "\\,").Replace(value)
}
