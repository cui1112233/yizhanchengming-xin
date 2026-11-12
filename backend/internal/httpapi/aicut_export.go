package httpapi

import (
	"context"
	"fmt"
	"io"
	"math"
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

var aicutHardwareProbeCache = struct {
	sync.Mutex
	values map[string]bool
}{values: map[string]bool{}}

type aicutExportOptions = aicut.ExportOptions

type aicutRenderClip struct {
	clip              aicut.Clip
	asset             aicut.Asset
	path              string
	mattePath         string
	matteInputIndex   int
	stabilizationPath string
	hslLUTPath        string
	layerEffects      []aicut.Clip
}

func aicutPersonSegmentationSignature(clip aicut.Clip) string {
	return aicut.PersonSegmentationSignature(clip)
}

func normalizeAICutExportOptions(format string, input aicutExportOptions) (aicutExportOptions, error) {
	if format == "mp3" {
		return aicutExportOptions{}, nil
	}
	input.Resolution = strings.ToLower(strings.TrimSpace(input.Resolution))
	input.Quality = strings.ToLower(strings.TrimSpace(input.Quality))
	input.Codec = strings.ToLower(strings.TrimSpace(input.Codec))
	input.HardwareAcceleration = strings.ToLower(strings.TrimSpace(input.HardwareAcceleration))
	input.BackgroundColor = strings.TrimSpace(input.BackgroundColor)
	if input.BackgroundColor == "" {
		input.BackgroundColor = "#000000"
	}
	if input.Resolution == "" {
		input.Resolution = "1080p"
	}
	if input.FrameRate == 0 {
		input.FrameRate = 30
	}
	if input.Quality == "" {
		input.Quality = "standard"
	}
	if input.Codec == "" {
		if format == "mov" {
			input.Codec = "prores4444"
		} else if format == "webm" {
			input.Codec = "vp9"
		} else {
			input.Codec = "h264"
		}
	}
	if input.HardwareAcceleration == "" {
		input.HardwareAcceleration = "auto"
	}
	if input.BitDepth == 0 {
		if input.Codec == "prores4444" {
			input.BitDepth = 12
		} else {
			input.BitDepth = 8
		}
	}
	if input.Resolution != "720p" && input.Resolution != "1080p" && input.Resolution != "4k" {
		return aicutExportOptions{}, fmt.Errorf("导出分辨率仅支持 720P、1080P 或 4K")
	}
	switch input.FrameRate {
	case 24, 25, 30, 50, 60:
	default:
		return aicutExportOptions{}, fmt.Errorf("导出帧率仅支持 24、25、30、50 或 60 fps")
	}
	if input.Quality != "draft" && input.Quality != "standard" && input.Quality != "high" {
		return aicutExportOptions{}, fmt.Errorf("导出质量仅支持快速预览、标准或高质量")
	}
	if format == "mp4" && input.Codec != "h264" && input.Codec != "hevc" && input.Codec != "av1" {
		return aicutExportOptions{}, fmt.Errorf("MP4 编码器仅支持 H.264、HEVC 或 AV1")
	}
	if format == "webm" && input.Codec != "vp9" && input.Codec != "av1" {
		return aicutExportOptions{}, fmt.Errorf("WebM 编码器仅支持 VP9 或 AV1")
	}
	if format == "mov" && input.Codec != "prores4444" {
		return aicutExportOptions{}, fmt.Errorf("MOV 透明导出仅支持 ProRes 4444")
	}
	if input.BitDepth != 8 && input.BitDepth != 10 && input.BitDepth != 12 {
		return aicutExportOptions{}, fmt.Errorf("导出色深仅支持 8-bit、10-bit 或 ProRes 4444 的 12-bit")
	}
	if input.BitDepth == 12 && input.Codec != "prores4444" {
		return aicutExportOptions{}, fmt.Errorf("12-bit 仅用于 ProRes 4444")
	}
	if input.Codec == "h264" && input.BitDepth == 10 {
		return aicutExportOptions{}, fmt.Errorf("H.264 暂不提供 10-bit，请选择 HEVC 或 AV1")
	}
	if input.Alpha && !((format == "webm" && input.Codec == "vp9") || (format == "mov" && input.Codec == "prores4444")) {
		return aicutExportOptions{}, fmt.Errorf("透明导出仅支持 WebM VP9 或 MOV ProRes 4444")
	}
	if input.BitrateMbps < 0 || input.BitrateMbps > 100 {
		return aicutExportOptions{}, fmt.Errorf("目标码率必须为自动或 1–100 Mbps")
	}
	if input.HardwareAcceleration != "auto" && input.HardwareAcceleration != "software" {
		return aicutExportOptions{}, fmt.Errorf("硬件加速仅支持自动选择或软件编码")
	}
	return input, nil
}

func aicutHardwareEncoderCandidates(codec string) []string {
	switch strings.ToLower(strings.TrimSpace(codec)) {
	case "h264":
		return []string{"h264_nvenc", "h264_qsv", "h264_amf"}
	case "hevc":
		return []string{"hevc_nvenc", "hevc_qsv", "hevc_amf"}
	case "av1":
		return []string{"av1_nvenc", "av1_qsv", "av1_amf"}
	default:
		return nil
	}
}

func selectAICutHardwareEncoder(codec string, bitDepth int, available map[string]bool) string {
	if bitDepth != 8 && bitDepth != 10 {
		return ""
	}
	for _, encoder := range aicutHardwareEncoderCandidates(codec) {
		if available[encoder] {
			return encoder
		}
	}
	return ""
}

func probeAICutHardwareEncoder(ctx context.Context, encoder string, bitDepth int) bool {
	key := fmt.Sprintf("%s/%d", encoder, bitDepth)
	aicutHardwareProbeCache.Lock()
	value, cached := aicutHardwareProbeCache.values[key]
	aicutHardwareProbeCache.Unlock()
	if cached {
		return value
	}
	pixelFormat := "yuv420p"
	if bitDepth == 10 {
		pixelFormat = "p010le"
	}
	probeCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	err := runAICutFFmpeg(probeCtx, "ffmpeg", "-hide_banner", "-loglevel", "error", "-f", "lavfi", "-i", "color=c=black:s=64x64:d=0.05", "-frames:v", "1", "-an", "-pix_fmt", pixelFormat, "-c:v", encoder, "-f", "null", "-")
	aicutHardwareProbeCache.Lock()
	aicutHardwareProbeCache.values[key] = err == nil
	aicutHardwareProbeCache.Unlock()
	return err == nil
}

func resolveAICutHardwareEncoder(ctx context.Context, options aicutExportOptions) string {
	if options.HardwareAcceleration != "auto" || options.Alpha {
		return ""
	}
	available := make(map[string]bool)
	for _, encoder := range aicutHardwareEncoderCandidates(options.Codec) {
		available[encoder] = probeAICutHardwareEncoder(ctx, encoder, options.BitDepth)
	}
	return selectAICutHardwareEncoder(options.Codec, options.BitDepth, available)
}

func RegisterAICutExportRoutes(mux *http.ServeMux, auth BridgeAuth, store aicut.Store, objects agentworkspace.AssetObjectStore, service *AICutExportService) {
	if mux == nil || store == nil {
		return
	}
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
		if service == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "导出队列尚未配置，请检查 Redis 和素材存储"})
			return
		}
		var input struct {
			Format               string `json:"format"`
			ExpectedRevision     int64  `json:"expectedRevision"`
			Resolution           string `json:"resolution"`
			FrameRate            int    `json:"frameRate"`
			Quality              string `json:"quality"`
			Codec                string `json:"codec"`
			BitDepth             int    `json:"bitDepth"`
			BitrateMbps          int    `json:"bitrateMbps"`
			Alpha                bool   `json:"alpha"`
			HardwareAcceleration string `json:"hardwareAcceleration"`
		}
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		input.Format = strings.ToLower(strings.TrimSpace(input.Format))
		if input.Format != "mp4" && input.Format != "webm" && input.Format != "mov" && input.Format != "mp3" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "仅支持 MP4、WebM、MOV 或 MP3 导出"})
			return
		}
		options, optionsErr := normalizeAICutExportOptions(input.Format, aicutExportOptions{Resolution: input.Resolution, FrameRate: input.FrameRate, Quality: input.Quality, Codec: input.Codec, BitDepth: input.BitDepth, BitrateMbps: input.BitrateMbps, Alpha: input.Alpha, HardwareAcceleration: input.HardwareAcceleration})
		if optionsErr != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": optionsErr.Error()})
			return
		}
		projectID := strings.TrimSpace(req.PathValue("id"))
		project, err := store.Project(req.Context(), owner, projectID)
		if err != nil {
			writeAICutError(w, err)
			return
		}
		if options.Alpha && project.BackgroundMode != "transparent" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请先把画布背景切换为透明，再导出 Alpha 成片"})
			return
		}
		options.BackgroundColor = project.BackgroundColor
		timeline, err := store.Timeline(req.Context(), owner, projectID)
		if err != nil {
			writeAICutError(w, err)
			return
		}
		if timeline.Revision != input.ExpectedRevision {
			writeAICutError(w, aicut.ErrRevisionConflict)
			return
		}
		job, submitErr := service.Submit(req.Context(), owner, project, timeline, input.Format, options)
		if submitErr != nil {
			status := http.StatusUnprocessableEntity
			if strings.Contains(submitErr.Error(), "重复提交") || strings.Contains(submitErr.Error(), "正在排队") {
				status = http.StatusConflict
			} else if strings.Contains(submitErr.Error(), "Redis") || strings.Contains(submitErr.Error(), "尚未配置") {
				status = http.StatusServiceUnavailable
			}
			writeJSON(w, status, map[string]string{"error": submitErr.Error()})
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"job": job})
	}))
	mux.Handle("GET /api/ai-cut/projects/{id}/export-jobs", authenticated(func(w http.ResponseWriter, req *http.Request, owner string) {
		if service == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "导出队列尚未配置"})
			return
		}
		jobs, err := service.List(req.Context(), owner, strings.TrimSpace(req.PathValue("id")))
		if err != nil {
			writeAICutError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"jobs": jobs})
	}))
	mux.Handle("GET /api/ai-cut/projects/{id}/export-jobs/{jobId}", authenticated(func(w http.ResponseWriter, req *http.Request, owner string) {
		if service == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "导出队列尚未配置"})
			return
		}
		job, err := service.Job(req.Context(), owner, strings.TrimSpace(req.PathValue("id")), strings.TrimSpace(req.PathValue("jobId")))
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "导出任务不存在或已清理"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"job": job})
	}))
	mux.Handle("POST /api/ai-cut/projects/{id}/export-jobs/{jobId}/cancel", authenticated(func(w http.ResponseWriter, req *http.Request, owner string) {
		if service == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "导出队列尚未配置"})
			return
		}
		job, err := service.Cancel(req.Context(), owner, strings.TrimSpace(req.PathValue("id")), strings.TrimSpace(req.PathValue("jobId")))
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "导出任务不存在或已清理"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"job": job})
	}))
	mux.Handle("GET /api/ai-cut/projects/{id}/export-jobs/{jobId}/download", authenticated(func(w http.ResponseWriter, req *http.Request, owner string) {
		if service == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "导出队列尚未配置"})
			return
		}
		job, err := service.Job(req.Context(), owner, strings.TrimSpace(req.PathValue("id")), strings.TrimSpace(req.PathValue("jobId")))
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "导出任务不存在或已清理"})
			return
		}
		if job.State != aicut.ExportSucceeded || job.OutputPath == "" {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "导出尚未完成"})
			return
		}
		file, err := os.Open(job.OutputPath)
		if err != nil {
			writeJSON(w, http.StatusGone, map[string]string{"error": "导出文件已清理，请重新导出"})
			return
		}
		defer file.Close()
		contentTypes := map[string]string{"mp4": "video/mp4", "webm": "video/webm", "mov": "video/quicktime", "mp3": "audio/mpeg"}
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
			ExpectedRevision     int64  `json:"expectedRevision"`
			Resolution           string `json:"resolution"`
			FrameRate            int    `json:"frameRate"`
			Quality              string `json:"quality"`
			Codec                string `json:"codec"`
			BitDepth             int    `json:"bitDepth"`
			BitrateMbps          int    `json:"bitrateMbps"`
			HardwareAcceleration string `json:"hardwareAcceleration"`
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
		options, optionsErr := normalizeAICutExportOptions(format, aicutExportOptions{Resolution: input.Resolution, FrameRate: input.FrameRate, Quality: input.Quality, Codec: input.Codec, BitDepth: input.BitDepth, BitrateMbps: input.BitrateMbps, HardwareAcceleration: input.HardwareAcceleration})
		if optionsErr != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": optionsErr.Error()})
			return
		}
		output, cleanup, err := renderAICutExportWithOptions(req.Context(), owner, project, timeline, format, options, store, objects)
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
	options, err := normalizeAICutExportOptions(format, aicutExportOptions{})
	if err != nil {
		return "", nil, err
	}
	return renderAICutExportWithOptions(ctx, owner, project, timeline, format, options, store, objects)
}

func renderAICutExportWithOptions(ctx context.Context, owner string, project aicut.Project, timeline aicut.Timeline, format string, options aicutExportOptions, store aicut.Store, objects agentworkspace.AssetObjectStore) (string, func(), error) {
	options, err := normalizeAICutExportOptions(format, options)
	if err != nil {
		return "", nil, err
	}
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
		body, openErr := objects.Open(ctx, aicutExportSourceObject(asset))
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
		if person := clip.VisualProcessing; person != nil && person.PersonSegmentation.Enabled {
			settings := person.PersonSegmentation
			if strings.TrimSpace(settings.Signature) != aicutPersonSegmentationSignature(clip) {
				cleanup()
				return "", nil, fmt.Errorf("片段“%s”的人物蒙版已过期，请重新生成", clip.Name)
			}
			matteAsset, matteErr := store.Asset(ctx, owner, project.ID, settings.MatteAssetID)
			if matteErr != nil || matteAsset.Kind != "video" {
				cleanup()
				return "", nil, fmt.Errorf("片段“%s”的人物蒙版素材不存在", clip.Name)
			}
			matteBody, openMatteErr := objects.Open(ctx, aicutExportSourceObject(matteAsset))
			if openMatteErr != nil {
				cleanup()
				return "", nil, fmt.Errorf("人物蒙版“%s”读取失败", matteAsset.Name)
			}
			matteExt := filepath.Ext(matteAsset.Name)
			if matteExt == "" {
				matteExt = ".webm"
			}
			item.mattePath = filepath.Join(dir, fmt.Sprintf("matte-%03d%s", index, matteExt))
			matteFile, matteCreateErr := os.Create(item.mattePath)
			if matteCreateErr == nil {
				_, matteCreateErr = io.Copy(matteFile, matteBody)
				if closeErr := matteFile.Close(); matteCreateErr == nil {
					matteCreateErr = closeErr
				}
			}
			_ = matteBody.Close()
			if matteCreateErr != nil {
				cleanup()
				return "", nil, fmt.Errorf("人物蒙版“%s”准备失败", matteAsset.Name)
			}
		}
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
		var fontErr error
		timeline, _, fontErr = prepareAICutExportFonts(ctx, owner, project.ID, dir, timeline, store, objects)
		if fontErr != nil {
			cleanup()
			return "", nil, fontErr
		}
		width, height := aicutRenderSizeFor(project.AspectRatio, options.Resolution)
		if content := buildAICutASS(timeline, width, height); content != "" {
			subtitlePath = filepath.Join(dir, "subtitles.ass")
			if writeErr := os.WriteFile(subtitlePath, []byte(content), 0o600); writeErr != nil {
				cleanup()
				return "", nil, fmt.Errorf("字幕渲染文件准备失败")
			}
		}
	}
	if format != "mp3" {
		if lutErr := prepareAICutHSLLUTs(dir, prepared); lutErr != nil {
			cleanup()
			return "", nil, lutErr
		}
		if stabilizeErr := prepareAICutStabilizationTransforms(ctx, dir, prepared); stabilizeErr != nil {
			cleanup()
			return "", nil, stabilizeErr
		}
	}
	options.ResolvedVideoEncoder = resolveAICutHardwareEncoder(ctx, options)
	args, err := buildAICutFFmpegArgsWithOptions(project.AspectRatio, prepared, format, output, options, subtitlePath)
	if err != nil {
		cleanup()
		return "", nil, err
	}
	renderCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()
	if renderErr := runAICutFFmpeg(renderCtx, "ffmpeg", args...); renderErr != nil {
		if options.ResolvedVideoEncoder != "" && options.HardwareAcceleration == "auto" {
			_ = os.Remove(output)
			options.ResolvedVideoEncoder = ""
			softwareArgs, argsErr := buildAICutFFmpegArgsWithOptions(project.AspectRatio, prepared, format, output, options, subtitlePath)
			if argsErr == nil {
				renderErr = runAICutFFmpeg(renderCtx, "ffmpeg", softwareArgs...)
			}
		}
		if renderErr != nil {
			cleanup()
			return "", nil, fmt.Errorf("AI-CUT 渲染失败：%w", renderErr)
		}
	}
	return output, cleanup, nil
}

// aicutExportSourceObject is deliberately the original object. Proxy media is
// only for low-memory browser preview and must never reduce final export quality.
func aicutExportSourceObject(asset aicut.Asset) agentworkspace.Asset {
	return aicutObjectAsset(asset)
}

func buildAICutFFmpegArgs(aspectRatio string, clips []aicutRenderClip, format, output string, subtitlePaths ...string) ([]string, error) {
	options, err := normalizeAICutExportOptions(format, aicutExportOptions{})
	if err != nil {
		return nil, err
	}
	subtitlePath := ""
	if len(subtitlePaths) > 0 {
		subtitlePath = strings.TrimSpace(subtitlePaths[0])
	}
	return buildAICutFFmpegArgsWithOptions(aspectRatio, clips, format, output, options, subtitlePath)
}

func buildAICutFFmpegArgsWithOptions(aspectRatio string, clips []aicutRenderClip, format, output string, options aicutExportOptions, subtitlePath string) ([]string, error) {
	options, err := normalizeAICutExportOptions(format, options)
	if err != nil {
		return nil, err
	}
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
	timelineComposition := options.Alpha
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
	matteCount := 0
	for index := range clips {
		if clips[index].clip.VisualProcessing != nil && clips[index].clip.VisualProcessing.PersonSegmentation.Enabled {
			timelineComposition = true
			if strings.TrimSpace(clips[index].mattePath) == "" {
				return nil, fmt.Errorf("片段“%s”的人物蒙版素材不存在", clips[index].clip.Name)
			}
			if clips[index].clip.VisualProcessing.PersonSegmentation.Signature != aicutPersonSegmentationSignature(clips[index].clip) {
				return nil, fmt.Errorf("片段“%s”的人物蒙版已过期，请重新生成", clips[index].clip.Name)
			}
			clips[index].matteInputIndex = len(clips) + matteCount
			matteCount++
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
			volumeFilter := fmt.Sprintf("volume=%.3f", volume)
			if expression, ok := buildAICutKeyframeExpression(item.clip, "volume", volume, "t"); ok {
				volumeFilter = fmt.Sprintf("volume='%s':eval=frame", expression)
			}
			chain := fmt.Sprintf("[%d:a]atrim=start=%.3f:duration=%.3f,asetpts=PTS-STARTPTS,%s", i, start, duration, volumeFilter)
			if processing := item.clip.AudioProcessing; processing != nil {
				if processing.Denoise > 0 {
					chain += fmt.Sprintf(",afftdn=nr=%.2f:nf=-50", 6+processing.Denoise*18)
				}
				if processing.EQLowDB != 0 {
					chain += fmt.Sprintf(",equalizer=f=120:t=q:w=1:g=%.2f", processing.EQLowDB)
				}
				if processing.EQMidDB != 0 {
					chain += fmt.Sprintf(",equalizer=f=1000:t=q:w=1:g=%.2f", processing.EQMidDB)
				}
				if processing.EQHighDB != 0 {
					chain += fmt.Sprintf(",equalizer=f=8000:t=q:w=1:g=%.2f", processing.EQHighDB)
				}
				switch processing.ChannelMode {
				case "mono":
					chain += ",pan=mono|c0=0.5*c0+0.5*c1"
				case "left":
					chain += ",pan=stereo|c0=c0|c1=c0"
				case "right":
					chain += ",pan=stereo|c0=c1|c1=c1"
				case "swap":
					chain += ",pan=stereo|c0=c1|c1=c0"
				}
				for _, filter := range buildAICutVoiceEffectFilters(processing.VoiceEffect) {
					chain += "," + filter
				}
				for _, filter := range buildAICutPitchFilters(processing.PitchSemitones) {
					chain += "," + filter
				}
				if processing.LoudnessEnabled {
					chain += fmt.Sprintf(",loudnorm=I=%.1f:TP=-1.5:LRA=11", processing.TargetLUFS)
				}
				for _, filter := range buildAICutBalanceFilters(processing.Pan) {
					chain += "," + filter
				}
			}
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
		width, height := aicutRenderSizeFor(aspectRatio, options.Resolution)
		start := float64(item.clip.SourceStartMS) / 1000
		rate := item.clip.PlaybackRate
		if rate <= 0 {
			rate = 1
		}
		inputLabel := fmt.Sprintf("[%d:v]", i)
		if item.clip.FreezeSourceMS == nil && item.asset.Kind == "video" && len(item.clip.SpeedCurve) >= 2 {
			points := item.clip.SpeedCurve
			branchLabels := make([]string, 0, len(points)-1)
			for segmentIndex := 0; segmentIndex < len(points)-1; segmentIndex++ {
				branchLabels = append(branchLabels, fmt.Sprintf("[scsrc%d_%d]", i, segmentIndex))
			}
			filters = append(filters, fmt.Sprintf("[%d:v]split=%d%s", i, len(branchLabels), strings.Join(branchLabels, "")))
			segmentLabels := make([]string, 0, len(points)-1)
			sourceCursor := start
			for segmentIndex := 0; segmentIndex < len(points)-1; segmentIndex++ {
				left, right := points[segmentIndex], points[segmentIndex+1]
				outputDuration := duration * (right.Position - left.Position)
				segmentRate := (left.Rate + right.Rate) / 2
				sourceDuration := outputDuration * segmentRate
				label := fmt.Sprintf("[scv%d_%d]", i, segmentIndex)
				filters = append(filters, fmt.Sprintf("[scsrc%d_%d]trim=start=%.3f:duration=%.3f,setpts=(PTS-STARTPTS)/%.3f%s", i, segmentIndex, sourceCursor, sourceDuration, segmentRate, label))
				segmentLabels = append(segmentLabels, label)
				sourceCursor += sourceDuration
			}
			inputLabel = fmt.Sprintf("[scv%d]", i)
			filters = append(filters, strings.Join(segmentLabels, "")+fmt.Sprintf("concat=n=%d:v=1:a=0%s", len(segmentLabels), inputLabel))
			rate = 1
		}
		chain := make([]string, 0, 8)
		if item.clip.FreezeSourceMS != nil {
			frameDuration := min(duration, 1.0/30.0)
			chain = append(chain,
				fmt.Sprintf("%strim=start=%.6f:duration=%.6f", inputLabel, float64(*item.clip.FreezeSourceMS)/1000, frameDuration),
				"setpts=PTS-STARTPTS",
			)
			if padding := max(0, duration-frameDuration); padding > 0 {
				chain = append(chain, fmt.Sprintf("tpad=stop_mode=clone:stop_duration=%.6f", padding))
			}
			if timelineComposition {
				chain = append(chain, fmt.Sprintf("setpts=PTS+%.3f/TB", float64(item.clip.StartMS)/1000))
			}
		} else {
			if len(item.clip.SpeedCurve) < 2 || item.asset.Kind != "video" {
				chain = append(chain, fmt.Sprintf("%strim=start=%.3f:duration=%.3f", inputLabel, start, duration*rate))
			} else if item.clip.Reversed {
				chain = append(chain, inputLabel+"reverse")
			} else {
				chain = append(chain, inputLabel+"setpts=PTS")
			}
			if item.clip.Reversed && (len(item.clip.SpeedCurve) < 2 || item.asset.Kind != "video") {
				chain = append(chain, "reverse")
			}
			setPTS := fmt.Sprintf("(PTS-STARTPTS)/%.3f", rate)
			if timelineComposition {
				setPTS += fmt.Sprintf("+%.3f/TB", float64(item.clip.StartMS)/1000)
			}
			chain = append(chain, "setpts="+setPTS)
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
		if item.clip.VisualProcessing != nil && aicut.HasColorGradeAdjustments(item.clip.VisualProcessing.Color) && item.hslLUTPath == "" {
			return nil, fmt.Errorf("HSL 调色文件尚未准备，无法导出")
		}
		chain = append(chain, buildAICutVisualProcessingFilters(item.clip.VisualProcessing, timelineComposition, item.hslLUTPath)...)
		chain = append(chain, buildAICutStabilizationFilters(item)...)
		chain = append(chain, buildAICutFrameInterpolationFilters(item.clip.VisualProcessing, options.FrameRate)...)
		padColor := aicutFFmpegCanvasColor(options.BackgroundColor)
		if options.Alpha {
			padColor = "black@0.0"
		}
		chain = append(chain,
			fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease", width, height),
			fmt.Sprintf("pad=%d:%d:(ow-iw)/2:(oh-ih)/2:color=%s", width, height, padColor),
		)
		chain = append(chain, buildAICutObjectRemovalFilters(item.clip.VisualProcessing, width, height)...)
		if person := item.clip.VisualProcessing; person != nil && person.PersonSegmentation.Enabled {
			settings := person.PersonSegmentation
			colorLabel := fmt.Sprintf("[personcolor%d]", i)
			maskLabel := fmt.Sprintf("[personmask%d]", i)
			filters = append(filters, strings.Join(append(chain, "format=rgba"+colorLabel), ","))
			maskPTS := "PTS-STARTPTS"
			if timelineComposition {
				maskPTS += fmt.Sprintf("+%.3f/TB", float64(item.clip.StartMS)/1000)
			}
			lower := max(0, settings.Threshold-settings.Feather/2) * 255
			upper := min(1, settings.Threshold+settings.Feather/2) * 255
			lut := fmt.Sprintf("lut=y='if(gte(val,%.3f),255,0)'", settings.Threshold*255)
			if settings.Feather > 0 {
				lut = fmt.Sprintf("lut=y='if(lte(val,%.3f),0,if(gte(val,%.3f),255,(val-%.3f)/%.3f*255))'", lower, upper, lower, max(.001, upper-lower))
			}
			blurRadius := max(.1, settings.Feather*20)
			filters = append(filters, fmt.Sprintf("[%d:v]trim=start=0:duration=%.3f,setpts=%s,scale=%d:%d,format=gray,boxblur=luma_radius=%.3f:luma_power=1,%s%s", item.matteInputIndex, duration, maskPTS, width, height, blurRadius, lut, maskLabel))
			chain = []string{fmt.Sprintf("%s%salphamerge", colorLabel, maskLabel)}
		}
		chain = append(chain, buildAICutAnimatedCanvasTransformFilters(width, height, item.clip, timelineComposition)...)
		pixelFormat := "format=yuv420p"
		if options.BitDepth == 10 {
			pixelFormat = "format=yuv420p10le"
		}
		if timelineComposition {
			pixelFormat = "format=rgba"
		}
		chain = append(chain, fmt.Sprintf("fps=%d", options.FrameRate), pixelFormat+fmt.Sprintf("[v%d]", i))
		filters = append(filters, strings.Join(chain, ","))
		videoLabels = append(videoLabels, fmt.Sprintf("[v%d]", i))
		videoItems = append(videoItems, item)
	}
	for _, item := range clips {
		if strings.TrimSpace(item.mattePath) != "" {
			args = append(args, "-i", item.mattePath)
		}
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
	finalVideoLabel := "[outv]"
	if subtitlePath != "" {
		finalVideoLabel = "[basev]"
	}
	if timelineComposition {
		width, height := aicutRenderSizeFor(aspectRatio, options.Resolution)
		totalMS := int64(0)
		for _, item := range videoItems {
			totalMS = max(totalMS, item.clip.StartMS+item.clip.DurationMS)
		}
		baseColor := aicutFFmpegCanvasColor(options.BackgroundColor)
		if options.Alpha {
			baseColor = "black@0.0"
		}
		args = append(args, "-f", "lavfi", "-t", fmt.Sprintf("%.3f", float64(totalMS)/1000), "-i", fmt.Sprintf("color=c=%s:s=%dx%d:r=%d:d=%.3f,format=rgba", baseColor, width, height, options.FrameRate, float64(totalMS)/1000))
		currentLabel := fmt.Sprintf("[%d:v]", len(clips)+matteCount)
		for index, label := range videoLabels {
			outputLabel := fmt.Sprintf("[layer%d]", index)
			if index == len(videoLabels)-1 {
				outputLabel = "[composed]"
			}
			item := videoItems[index]
			start, end := float64(item.clip.StartMS)/1000, float64(item.clip.StartMS+item.clip.DurationMS)/1000
			if blendMode := aicutFFmpegBlendMode(item.clip.BlendMode); blendMode != "" {
				baseBlend, baseOverlay := fmt.Sprintf("[blendbase%d]", index), fmt.Sprintf("[overlaybase%d]", index)
				topBlend, topAlpha := fmt.Sprintf("[blendtop%d]", index), fmt.Sprintf("[alphatop%d]", index)
				mixed, mixedRGB := fmt.Sprintf("[mixed%d]", index), fmt.Sprintf("[mixedrgb%d]", index)
				alphaMask, mixedAlpha := fmt.Sprintf("[alphamask%d]", index), fmt.Sprintf("[mixedalpha%d]", index)
				filters = append(filters,
					fmt.Sprintf("%ssplit=2%s%s", currentLabel, baseBlend, baseOverlay),
					fmt.Sprintf("%ssplit=2%s%s", label, topBlend, topAlpha),
					fmt.Sprintf("%s%sblend=all_mode=%s:eof_action=pass:enable='between(t,%.3f,%.3f)'%s", baseBlend, topBlend, blendMode, start, end, mixed),
					fmt.Sprintf("%salphaextract%s", topAlpha, alphaMask),
					fmt.Sprintf("%sformat=rgb24%s", mixed, mixedRGB),
					fmt.Sprintf("%s%salphamerge%s", mixedRGB, alphaMask, mixedAlpha),
					fmt.Sprintf("%s%soverlay=eof_action=pass:enable='between(t,%.3f,%.3f)'%s", baseOverlay, mixedAlpha, start, end, outputLabel),
				)
			} else {
				filters = append(filters, fmt.Sprintf("%s%soverlay=eof_action=pass:enable='between(t,%.3f,%.3f)'%s", currentLabel, label, start, end, outputLabel))
			}
			currentLabel = outputLabel
		}
		finalFormat := "yuv420p"
		if options.Alpha {
			finalFormat = "rgba"
		}
		filters = append(filters, "[composed]format="+finalFormat+finalVideoLabel)
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
		assFilter := fmt.Sprintf("ass='%s'", escapeAICutFilterPath(subtitlePath))
		fontsDir := filepath.Join(filepath.Dir(subtitlePath), "fonts")
		if info, statErr := os.Stat(fontsDir); statErr == nil && info.IsDir() {
			assFilter += fmt.Sprintf(":fontsdir='%s'", escapeAICutFilterPath(fontsDir))
		}
		filters = append(filters, "[basev]"+assFilter+"[outv]")
	}
	if len(audioLabels) > 0 {
		filters = append(filters, strings.Join(audioLabels, "")+fmt.Sprintf("amix=inputs=%d:duration=longest:normalize=0[outa]", len(audioLabels)))
	}
	args = append(args, "-filter_complex", strings.Join(filters, ";"), "-map", "[outv]")
	if len(audioLabels) > 0 {
		args = append(args, "-map", "[outa]")
	}
	pixelFormat := "yuv420p"
	if options.BitDepth == 10 {
		pixelFormat = "yuv420p10le"
	}
	if options.BitDepth == 10 && options.ResolvedVideoEncoder != "" {
		pixelFormat = "p010le"
	}
	if options.Alpha && options.Codec == "vp9" {
		pixelFormat = "yuva420p"
	}
	if options.Alpha && options.Codec == "prores4444" {
		pixelFormat = "yuva444p12le"
	}
	if options.ResolvedVideoEncoder != "" {
		encoder := options.ResolvedVideoEncoder
		qualityValue := map[string]string{"draft": "28", "standard": "23", "high": "18"}[options.Quality]
		args = append(args, "-c:v", encoder)
		switch {
		case strings.HasSuffix(encoder, "_nvenc"):
			preset := map[string]string{"draft": "p2", "standard": "p4", "high": "p6"}[options.Quality]
			args = append(args, "-preset", preset, "-cq", qualityValue)
		case strings.HasSuffix(encoder, "_qsv"):
			preset := map[string]string{"draft": "veryfast", "standard": "medium", "high": "slow"}[options.Quality]
			args = append(args, "-preset", preset, "-global_quality", qualityValue)
		case strings.HasSuffix(encoder, "_amf"):
			quality := map[string]string{"draft": "speed", "standard": "balanced", "high": "quality"}[options.Quality]
			args = append(args, "-quality", quality, "-qp_i", qualityValue, "-qp_p", qualityValue)
		}
	} else {
		switch options.Codec {
		case "vp9":
			crf := map[string]string{"draft": "38", "standard": "32", "high": "26"}[options.Quality]
			args = append(args, "-c:v", "libvpx-vp9", "-crf", crf)
			if options.BitrateMbps == 0 {
				args = append(args, "-b:v", "0")
			}
			if options.Alpha {
				args = append(args, "-auto-alt-ref", "0")
			}
		case "prores4444":
			args = append(args, "-c:v", "prores_ks", "-profile:v", "4")
		case "av1":
			crf := map[string]string{"draft": "40", "standard": "32", "high": "24"}[options.Quality]
			cpuUsed := map[string]string{"draft": "8", "standard": "6", "high": "4"}[options.Quality]
			args = append(args, "-c:v", "libaom-av1", "-crf", crf, "-cpu-used", cpuUsed, "-row-mt", "1")
			if options.BitrateMbps == 0 {
				args = append(args, "-b:v", "0")
			}
		case "hevc":
			crf := map[string]string{"draft": "30", "standard": "25", "high": "20"}[options.Quality]
			preset := map[string]string{"draft": "veryfast", "standard": "medium", "high": "slow"}[options.Quality]
			args = append(args, "-c:v", "libx265", "-preset", preset, "-crf", crf, "-tag:v", "hvc1")
		default:
			crf := map[string]string{"draft": "28", "standard": "23", "high": "18"}[options.Quality]
			preset := map[string]string{"draft": "veryfast", "standard": "medium", "high": "slow"}[options.Quality]
			args = append(args, "-c:v", "libx264", "-preset", preset, "-crf", crf)
		}
	}
	args = append(args, "-pix_fmt", pixelFormat)
	if options.BitrateMbps > 0 {
		bitrate := fmt.Sprintf("%dM", options.BitrateMbps)
		args = append(args, "-b:v", bitrate, "-maxrate", bitrate, "-bufsize", fmt.Sprintf("%dM", options.BitrateMbps*2))
	}
	if format == "webm" {
		if len(audioLabels) > 0 {
			args = append(args, "-c:a", "libopus")
		}
	} else if format == "mp4" {
		args = append(args, "-movflags", "+faststart")
		if len(audioLabels) > 0 {
			args = append(args, "-c:a", "aac")
		}
	} else if format == "mov" && len(audioLabels) > 0 {
		args = append(args, "-c:a", "pcm_s16le")
	}
	if len(audioLabels) > 0 {
		args = append(args, "-shortest")
	}
	args = append(args, output)
	return args, nil
}

func buildAICutObjectRemovalFilters(processing *aicut.VisualProcessing, width, height int) []string {
	if processing == nil || !processing.ObjectRemoval.Enabled || width < 1 || height < 1 {
		return nil
	}
	region := processing.ObjectRemoval
	regionWidth := max(1, min(width, int(math.Round(region.Width*float64(width)))))
	regionHeight := max(1, min(height, int(math.Round(region.Height*float64(height)))))
	x := int(math.Round((region.X - region.Width/2) * float64(width)))
	y := int(math.Round((region.Y - region.Height/2) * float64(height)))
	x = max(0, min(width-regionWidth, x))
	y = max(0, min(height-regionHeight, y))
	return []string{fmt.Sprintf("delogo=x=%d:y=%d:w=%d:h=%d:show=0", x, y, regionWidth, regionHeight)}
}

// Linear stereo balance: center is unchanged, positive attenuates only left,
// negative attenuates only right. Mono is first duplicated at unity gain.
func buildAICutBalanceFilters(pan float64) []string {
	if pan == 0 {
		return nil
	}
	return []string{
		"pan=stereo|c0=FL+FC|c1=FR+FC",
		fmt.Sprintf("pan=stereo|c0=%.9f*c0|c1=%.9f*c1", 1-max(0, pan), 1+min(0, pan)),
	}
}

func buildAICutVoiceEffectFilters(effect string) []string {
	switch strings.ToLower(strings.TrimSpace(effect)) {
	case "", "natural":
		return nil
	case "enhance":
		return []string{
			"highpass=f=80",
			"lowpass=f=14000",
			"equalizer=f=2500:t=q:w=1:g=4",
			"acompressor=threshold=0.125:ratio=3:attack=10:release=150:makeup=1.25",
		}
	case "bright":
		return []string{"equalizer=f=2500:t=q:w=1:g=3", "treble=g=4:f=6000"}
	case "bass":
		return []string{"bass=g=8:f=120", "lowpass=f=15000"}
	case "treble":
		return []string{"highpass=f=90", "treble=g=7:f=6000"}
	case "telephone":
		return []string{
			"highpass=f=300",
			"lowpass=f=3400",
			"equalizer=f=1800:t=q:w=1:g=3",
			"acompressor=threshold=0.18:ratio=3:attack=10:release=120:makeup=1.15",
		}
	case "radio":
		return []string{"highpass=f=250", "lowpass=f=4500", "acrusher=bits=12:mix=0.25", "acompressor=threshold=0.18:ratio=3:attack=10:release=120:makeup=1.15"}
	case "megaphone":
		return []string{"highpass=f=400", "lowpass=f=5000", "equalizer=f=2200:t=q:w=1:g=4", "acrusher=bits=10:mix=0.38", "acompressor=threshold=0.16:ratio=4:attack=8:release=100:makeup=1.2"}
	case "robot":
		return []string{
			"highpass=f=120",
			"lowpass=f=6500",
			"chorus=0.6:0.9:50:0.4:0.25:2",
			"aecho=0.8:0.6:20:0.25",
		}
	case "chipmunk":
		return []string{
			"aresample=44100",
			"asetrate=59535",
			"aresample=44100",
			"atempo=0.740741",
		}
	case "deep":
		return []string{"aresample=44100", "asetrate=35280", "aresample=44100", "atempo=1.25", "bass=g=5:f=120"}
	case "sweet":
		return []string{"equalizer=f=2500:t=q:w=1:g=5", "treble=g=3:f=6500", "acompressor=threshold=0.16:ratio=2.5:attack=12:release=140:makeup=1.15"}
	case "whisper":
		return []string{"highpass=f=500", "lowpass=f=12500", "treble=g=6:f=7000", "acompressor=threshold=0.2:ratio=3:attack=5:release=90:makeup=1.1"}
	case "distorted":
		return []string{"acrusher=bits=7:mix=0.65", "acompressor=threshold=0.2:ratio=4:attack=5:release=100:makeup=1.1"}
	case "lofi":
		return []string{"highpass=f=120", "lowpass=f=5200", "acrusher=bits=12:mix=0.2", "aecho=0.8:0.5:35:0.12"}
	case "echo":
		return []string{"aecho=0.8:0.6:220:0.35"}
	case "room":
		return []string{"aecho=0.82:0.55:28|47:0.20|0.12"}
	case "hall":
		return []string{"aecho=0.80:0.62:55|90:0.28|0.18"}
	case "plate":
		return []string{"treble=g=2:f=6000", "aecho=0.82:0.58:38|72:0.24|0.16"}
	case "cathedral":
		return []string{"aecho=0.78:0.68:95|180:0.42|0.24"}
	case "stadium":
		return []string{"aecho=0.78:0.65:125|240:0.34|0.22"}
	case "underwater":
		return []string{"lowpass=f=1100", "chorus=0.7:0.8:60:0.35:0.4:1.2"}
	case "vinyl":
		return []string{"highpass=f=120", "lowpass=f=7800", "acrusher=bits=14:mix=0.12", "aecho=0.85:0.35:18:0.08"}
	default:
		return nil
	}
}

func buildAICutPitchFilters(semitones float64) []string {
	if math.Abs(semitones) < 0.000001 {
		return nil
	}
	ratio := math.Pow(2, semitones/12)
	return []string{
		fmt.Sprintf("asetrate=44100*%.6f", ratio),
		"aresample=44100",
		fmt.Sprintf("atempo=%.6f", 1/ratio),
	}
}

func aicutFFmpegCanvasColor(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" || value == "#000000" {
		return "black"
	}
	return strings.TrimPrefix(value, "#")
}

func aicutFFmpegBlendMode(value string) string {
	switch value {
	case "multiply", "screen", "overlay", "difference":
		return value
	case "soft-light":
		return "softlight"
	case "add":
		return "addition"
	default:
		return ""
	}
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
	case "filter-golden-hour":
		return []string{fmt.Sprintf("colorbalance=rs=%.3f:gs=%.3f:bs=-%.3f", .16*intensity, .08*intensity, .10*intensity), fmt.Sprintf("eq=brightness=%.3f:contrast=%.3f:saturation=%.3f", .05*intensity, 1+.08*intensity, 1+.30*intensity)}, true
	case "filter-noir":
		return []string{"hue=s=0", fmt.Sprintf("eq=brightness=-%.3f:contrast=%.3f:gamma=%.3f", .03*intensity, 1+.50*intensity, 1-.10*intensity)}, true
	case "filter-vintage":
		return []string{fmt.Sprintf("colorbalance=rs=%.3f:gs=%.3f:bs=-%.3f", .11*intensity, .05*intensity, .09*intensity), fmt.Sprintf("eq=contrast=%.3f:saturation=%.3f", 1-.10*intensity, 1-.28*intensity), fmt.Sprintf("noise=alls=%d:allf=t", 2+int(5*intensity))}, true
	case "filter-cyberpunk":
		return []string{fmt.Sprintf("colorbalance=rs=%.3f:gs=-%.3f:bs=%.3f", .16*intensity, .10*intensity, .22*intensity), fmt.Sprintf("eq=contrast=%.3f:saturation=%.3f", 1+.32*intensity, 1+.75*intensity)}, true
	case "filter-vaporwave":
		return []string{fmt.Sprintf("colorbalance=rs=%.3f:gs=-%.3f:bs=%.3f", .20*intensity, .08*intensity, .18*intensity), fmt.Sprintf("eq=brightness=%.3f:gamma=%.3f:saturation=%.3f", .05*intensity, 1+.10*intensity, 1+.60*intensity)}, true
	case "filter-mint-haze":
		return []string{fmt.Sprintf("colorbalance=rs=-%.3f:gs=%.3f:bs=%.3f", .06*intensity, .10*intensity, .08*intensity), fmt.Sprintf("eq=brightness=%.3f:contrast=%.3f:saturation=%.3f", .07*intensity, 1-.14*intensity, 1-.12*intensity)}, true
	case "filter-bleach-bypass":
		return []string{fmt.Sprintf("eq=brightness=-%.3f:contrast=%.3f:saturation=%.3f", .02*intensity, 1+.48*intensity, 1-.55*intensity), fmt.Sprintf("unsharp=5:5:%.3f:5:5:0", .35+.55*intensity)}, true
	case "filter-winter-cool":
		return []string{fmt.Sprintf("colorbalance=rs=-%.3f:gs=%.3f:bs=%.3f", .10*intensity, .02*intensity, .18*intensity), fmt.Sprintf("eq=contrast=%.3f:saturation=%.3f", 1+.10*intensity, 1-.25*intensity)}, true
	case "filter-rosy":
		return []string{fmt.Sprintf("colorbalance=rs=%.3f:gs=-%.3f:bs=%.3f", .14*intensity, .03*intensity, .05*intensity), fmt.Sprintf("eq=brightness=%.3f:saturation=%.3f", .05*intensity, 1+.35*intensity)}, true
	case "filter-film":
		return []string{fmt.Sprintf("eq=brightness=-%.3f:contrast=%.3f:saturation=%.3f", .01*intensity, 1+.20*intensity, 1-.12*intensity), fmt.Sprintf("noise=alls=%d:allf=t", 2+int(7*intensity)), fmt.Sprintf("vignette=angle=%.3f", .12+.18*intensity)}, true
	case "filter-super-8":
		return []string{fmt.Sprintf("colorbalance=rs=%.3f:gs=%.3f:bs=-%.3f", .10*intensity, .03*intensity, .08*intensity), fmt.Sprintf("eq=contrast=%.3f:saturation=%.3f", 1+.28*intensity, 1+.18*intensity), fmt.Sprintf("noise=alls=%d:allf=t", 4+int(10*intensity)), fmt.Sprintf("vignette=angle=%.3f", .18+.22*intensity)}, true
	case "filter-instant-print":
		return []string{fmt.Sprintf("colorbalance=rs=%.3f:gs=%.3f:bs=-%.3f", .07*intensity, .04*intensity, .04*intensity), fmt.Sprintf("eq=brightness=%.3f:contrast=%.3f:saturation=%.3f", .08*intensity, 1-.16*intensity, 1-.18*intensity), fmt.Sprintf("vignette=angle=%.3f", .08+.10*intensity)}, true
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

func buildAICutVisualProcessingFilters(processing *aicut.VisualProcessing, preserveAlpha bool, hslLUTPath ...string) []string {
	if processing == nil {
		return nil
	}
	filters := make([]string, 0, 6)
	geometry := processing.Geometry
	if geometry.CropLeft != 0 || geometry.CropTop != 0 || geometry.CropRight != 0 || geometry.CropBottom != 0 {
		filters = append(filters, fmt.Sprintf(
			"crop=iw*%.6f:ih*%.6f:iw*%.6f:ih*%.6f",
			1-geometry.CropLeft-geometry.CropRight,
			1-geometry.CropTop-geometry.CropBottom,
			geometry.CropLeft,
			geometry.CropTop,
		))
	}
	if geometry.FlipHorizontal {
		filters = append(filters, "hflip")
	}
	if geometry.FlipVertical {
		filters = append(filters, "vflip")
	}
	usesAlpha := processing.ChromaKey.Enabled || processing.ChromaKey.RemoveBlack
	if usesAlpha {
		key := strings.TrimPrefix(processing.ChromaKey.Color, "#")
		if processing.ChromaKey.RemoveBlack {
			key = "000000"
		}
		keyRGB, _ := strconv.ParseUint(key, 16, 32)
		red, green, blue := (keyRGB>>16)&255, (keyRGB>>8)&255, keyRGB&255
		similarity, blend := processing.ChromaKey.Similarity, processing.ChromaKey.Blend
		if math.IsNaN(similarity) || math.IsInf(similarity, 0) {
			similarity = .2
		}
		if math.IsNaN(blend) || math.IsInf(blend, 0) {
			blend = .08
		}
		similarity, blend = max(.01, min(1, similarity)), max(.001, min(1, blend))
		// Match the preview: determine the matte from source RGB before grading,
		// and multiply existing alpha instead of replacing partially transparent pixels.
		distance := fmt.Sprintf("sqrt(pow(r(X,Y)-%d,2)+pow(g(X,Y)-%d,2)+pow(b(X,Y)-%d,2))/255", red, green, blue)
		ratio := fmt.Sprintf("clip(((%s)-%.9f)/%.9f,0,1)", distance, similarity, blend)
		filters = append(filters, fmt.Sprintf("format=gbrap,geq=r='r(X,Y)':g='g(X,Y)':b='b(X,Y)':a='st(0,%s);alpha(X,Y)*ld(0)*ld(0)*(3-2*ld(0))'", ratio))
	}
	color := processing.Color
	gamma := color.Gamma
	if gamma == 0 {
		gamma = 1
	}
	if color.Exposure != 0 || gamma != 1 || color.Highlights != 0 || color.Shadows != 0 {
		brightness := color.Exposure / 4
		contrast := 1 + color.Highlights*.18 - color.Shadows*.12
		filters = append(filters, fmt.Sprintf("eq=brightness=%.3f:contrast=%.3f:gamma=%.3f", brightness, contrast, gamma))
	}
	if color.Temperature != 0 || color.Tint != 0 || color.Highlights != 0 || color.Shadows != 0 {
		filters = append(filters, fmt.Sprintf("colorbalance=rs=%.3f:gs=%.3f:bs=%.3f:gm=%.3f:rh=%.3f:gh=%.3f:bh=%.3f", .10*color.Temperature+.05*color.Shadows, .02*color.Temperature, -.10*color.Temperature+.05*color.Shadows, .10*color.Tint, .07*color.Highlights, .04*color.Highlights, -.05*color.Highlights))
	}
	if len(color.CurvePoints) >= 2 {
		custom := len(color.CurvePoints) != 5
		parts := make([]string, 0, len(color.CurvePoints))
		for _, point := range color.CurvePoints {
			if math.Abs(point.X-point.Y) > .000001 {
				custom = true
			}
			parts = append(parts, fmt.Sprintf("%.6f/%.6f", point.X, point.Y))
		}
		if custom {
			filters = append(filters, fmt.Sprintf("curves=all='%s'", strings.Join(parts, " ")))
		}
	} else if len(color.Curve) == 5 {
		defaultCurve := []float64{0, .25, .5, .75, 1}
		custom := false
		for index, point := range color.Curve {
			if math.Abs(point-defaultCurve[index]) > .000001 {
				custom = true
				break
			}
		}
		if custom {
			filters = append(filters, fmt.Sprintf("curves=all='0/%.6f .25/%.6f .5/%.6f .75/%.6f 1/%.6f'", color.Curve[0], color.Curve[1], color.Curve[2], color.Curve[3], color.Curve[4]))
		}
	}
	if aicut.HasColorGradeAdjustments(color) && len(hslLUTPath) > 0 && hslLUTPath[0] != "" {
		filters = append(filters, fmt.Sprintf("lut3d=file='%s':interp=trilinear", escapeAICutFilterPath(hslLUTPath[0])))
	}
	details := processing.Details
	if details.Sharpen > 0 {
		filters = append(filters, fmt.Sprintf("unsharp=5:5:%.3f:5:5:0", .25+1.25*details.Sharpen))
	}
	if details.Vignette > 0 {
		filters = append(filters, fmt.Sprintf("vignette=angle=%.3f", .2+.8*details.Vignette))
	}
	if details.Grain > 0 {
		filters = append(filters, fmt.Sprintf("noise=alls=%d:allf=t+u", 2+int(math.Round(18*details.Grain))))
	}
	mask := processing.Mask
	if mask.Type == "rectangle" || mask.Type == "ellipse" || mask.Type == "polygon" || mask.Type == "bezier" {
		condition := "1"
		if mask.Type == "rectangle" {
			left, right := max(0, mask.X-mask.Width/2), min(1, mask.X+mask.Width/2)
			top, bottom := max(0, mask.Y-mask.Height/2), min(1, mask.Y+mask.Height/2)
			if mask.Feather > 0 {
				distance := fmt.Sprintf("min(min(X/W-%.6f,%.6f-X/W),min(Y/H-%.6f,%.6f-Y/H))", left, right, top, bottom)
				condition = fmt.Sprintf("clip(((%s)+%.6f)/(%.6f),0,1)", distance, mask.Feather, mask.Feather*2)
			} else {
				condition = fmt.Sprintf("between(X/W,%.6f,%.6f)*between(Y/H,%.6f,%.6f)", left, right, top, bottom)
			}
		} else if mask.Type == "ellipse" {
			rx, ry := max(.025, mask.Width/2), max(.025, mask.Height/2)
			ellipseDistance := fmt.Sprintf("sqrt(((X/W-%.6f)*(X/W-%.6f))/(%.6f*%.6f)+((Y/H-%.6f)*(Y/H-%.6f))/(%.6f*%.6f))", mask.X, mask.X, rx, rx, mask.Y, mask.Y, ry, ry)
			if mask.Feather > 0 {
				condition = fmt.Sprintf("clip(((1-(%s))*%.6f+%.6f)/(%.6f),0,1)", ellipseDistance, min(rx, ry), mask.Feather, mask.Feather*2)
			} else {
				condition = fmt.Sprintf("lte(%s,1)", ellipseDistance)
			}
		} else if mask.Type == "polygon" {
			condition = buildAICutPolygonMaskCondition(mask.Points, mask.Feather)
		} else {
			condition = buildAICutPolygonMaskCondition(aicut.SampleVisualBezierMask(mask.BezierPoints), mask.Feather)
		}
		if mask.Invert {
			condition = "1-(" + condition + ")"
		}
		filters = append(filters, fmt.Sprintf("format=rgba,geq=r='r(X,Y)':g='g(X,Y)':b='b(X,Y)':a='alpha(X,Y)*(%s)'", condition))
		usesAlpha = true
	}
	if usesAlpha && !preserveAlpha {
		filters = append(filters, "geq=r='r(X,Y)*alpha(X,Y)/255':g='g(X,Y)*alpha(X,Y)/255':b='b(X,Y)*alpha(X,Y)/255':a='alpha(X,Y)'")
	}
	return filters
}

func buildAICutFrameInterpolationFilters(processing *aicut.VisualProcessing, targetFPS int) []string {
	if processing == nil || processing.FrameInterpolation == "" || processing.FrameInterpolation == "off" || targetFPS <= 0 {
		return nil
	}
	switch processing.FrameInterpolation {
	case "blend":
		return []string{fmt.Sprintf("minterpolate=fps=%d:mi_mode=blend", targetFPS)}
	case "optical-flow":
		return []string{fmt.Sprintf("minterpolate=fps=%d:mi_mode=mci:mc_mode=aobmc:me_mode=bidir:vsbmc=1", targetFPS)}
	default:
		return nil
	}
}

func prepareAICutStabilizationTransforms(ctx context.Context, dir string, clips []aicutRenderClip) error {
	for index := range clips {
		item := &clips[index]
		settings := item.clip.VisualProcessing
		if item.asset.Kind != "video" || settings == nil || !settings.Stabilization.Enabled {
			continue
		}
		transformPath := filepath.Join(dir, fmt.Sprintf("stabilization-%03d.trf", index))
		tripod := 0
		if settings.Stabilization.Tripod {
			tripod = 1
		}
		duration := float64(item.clip.DurationMS) / 1000
		if item.clip.PlaybackRate > 0 {
			duration *= item.clip.PlaybackRate
		}
		if duration <= 0 {
			duration = .1
		}
		analysisCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		err := runAICutFFmpeg(analysisCtx, "ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-ss", fmt.Sprintf("%.3f", float64(item.clip.SourceStartMS)/1000), "-t", fmt.Sprintf("%.3f", duration), "-i", item.path, "-vf", fmt.Sprintf("vidstabdetect=result='%s':shakiness=5:accuracy=15:tripod=%d", escapeAICutFilterPath(transformPath), tripod), "-an", "-f", "null", "-")
		cancel()
		if err != nil {
			return fmt.Errorf("片段“%s”防抖分析失败：%w", item.clip.Name, err)
		}
		if info, statErr := os.Stat(transformPath); statErr != nil || info.Size() == 0 {
			return fmt.Errorf("片段“%s”防抖分析没有生成有效数据", item.clip.Name)
		}
		item.stabilizationPath = transformPath
	}
	return nil
}

func buildAICutStabilizationFilters(item aicutRenderClip) []string {
	settings := item.clip.VisualProcessing
	if settings == nil || !settings.Stabilization.Enabled || strings.TrimSpace(item.stabilizationPath) == "" {
		return nil
	}
	tripod := 0
	if settings.Stabilization.Tripod {
		tripod = 1
	}
	return []string{fmt.Sprintf("vidstabtransform=input='%s':smoothing=%d:zoom=%.3f:optzoom=0:crop=black:tripod=%d", escapeAICutFilterPath(item.stabilizationPath), settings.Stabilization.Smoothness, settings.Stabilization.Zoom, tripod)}
}

func buildAICutPolygonMaskCondition(points []aicut.VisualMaskPoint, feather float64) string {
	if len(points) < 3 {
		return "1"
	}
	crossings := make([]string, 0, len(points))
	distances := make([]string, 0, len(points))
	for index, start := range points {
		end := points[(index+1)%len(points)]
		dx, dy := end.X-start.X, end.Y-start.Y
		if math.Abs(dy) > .0000001 {
			minimumY, maximumY := min(start.Y, end.Y), max(start.Y, end.Y)
			intersectionX := fmt.Sprintf("%.6f+(Y/H-%.6f)*%.6f/%.6f", start.X, start.Y, dx, dy)
			crossings = append(crossings, fmt.Sprintf("gte(Y/H,%.6f)*lt(Y/H,%.6f)*lt(X/W,(%s))", minimumY, maximumY, intersectionX))
		}
		denominator := dx*dx + dy*dy
		if denominator < .0000001 {
			continue
		}
		ratio := fmt.Sprintf("clip(((X/W-%.6f)*%.6f+(Y/H-%.6f)*%.6f)/%.9f,0,1)", start.X, dx, start.Y, dy, denominator)
		closestX := fmt.Sprintf("(%.6f+%.6f*(%s))", start.X, dx, ratio)
		closestY := fmt.Sprintf("(%.6f+%.6f*(%s))", start.Y, dy, ratio)
		distanceX := fmt.Sprintf("(X/W-(%s))", closestX)
		distanceY := fmt.Sprintf("(Y/H-(%s))", closestY)
		distances = append(distances, fmt.Sprintf("sqrt((%s)*(%s)+(%s)*(%s))", distanceX, distanceX, distanceY, distanceY))
	}
	if len(crossings) == 0 {
		return "0"
	}
	inside := fmt.Sprintf("mod(%s,2)", strings.Join(crossings, "+"))
	if feather <= 0 || len(distances) == 0 {
		return inside
	}
	minimumDistance := distances[len(distances)-1]
	for index := len(distances) - 2; index >= 0; index-- {
		minimumDistance = fmt.Sprintf("min(%s,%s)", distances[index], minimumDistance)
	}
	signedDistance := fmt.Sprintf("if(gt(%s,0.5),%s,-(%s))", inside, minimumDistance, minimumDistance)
	return fmt.Sprintf("clip(((%s)+%.6f)/%.6f,0,1)", signedDistance, feather, feather*2)
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

func aicutCanvasTransformValues(transform *aicut.CanvasTransform) map[string]float64 {
	values := map[string]float64{"x": 0, "y": 0, "scaleX": 1, "scaleY": 1, "rotation": 0, "opacity": 1}
	if transform == nil {
		return values
	}
	values["x"], values["y"], values["rotation"] = transform.X, transform.Y, transform.Rotation
	if transform.ScaleX != 0 {
		values["scaleX"] = transform.ScaleX
	}
	if transform.ScaleY != 0 {
		values["scaleY"] = transform.ScaleY
	}
	values["opacity"] = transform.Opacity
	return values
}

func aicutKeyframeField(values aicut.KeyframeValues, field string) (*float64, bool) {
	var value *float64
	switch field {
	case "x":
		value = values.X
	case "y":
		value = values.Y
	case "scaleX":
		value = values.ScaleX
	case "scaleY":
		value = values.ScaleY
	case "rotation":
		value = values.Rotation
	case "opacity":
		value = values.Opacity
	case "volume":
		value = values.Volume
	}
	return value, value != nil
}

type aicutEasingSample struct {
	x float64
	y float64
}

func aicutCubicCoordinate(control1, control2, ratio float64) float64 {
	inverse := 1 - ratio
	return 3*inverse*inverse*ratio*control1 + 3*inverse*ratio*ratio*control2 + ratio*ratio*ratio
}

func aicutKeyframeEasingSamples(easing *aicut.KeyframeEasing) []aicutEasingSample {
	normalized := aicut.NormalizeKeyframeEasing(easing)
	if normalized.Type == "hold" {
		return []aicutEasingSample{{x: 0, y: 0}, {x: 1, y: 1}}
	}
	const steps = 8
	samples := make([]aicutEasingSample, 0, steps+1)
	for index := 0; index <= steps; index++ {
		ratio := float64(index) / steps
		samples = append(samples, aicutEasingSample{x: aicutCubicCoordinate(normalized.X1, normalized.X2, ratio), y: aicutCubicCoordinate(normalized.Y1, normalized.Y2, ratio)})
	}
	return samples
}

func aicutKeyframeEasingExpression(easing *aicut.KeyframeEasing, progressExpression string) string {
	normalized := aicut.NormalizeKeyframeEasing(easing)
	if normalized.Type == "linear" {
		return progressExpression
	}
	if normalized.Type == "hold" {
		return fmt.Sprintf("if(lt(%s,1),0,1)", progressExpression)
	}
	samples := aicutKeyframeEasingSamples(&normalized)
	expression := "1"
	for index := len(samples) - 2; index >= 0; index-- {
		left, right := samples[index], samples[index+1]
		span := right.x - left.x
		segment := fmt.Sprintf("%.6f", right.y)
		if span > 0 {
			segment = fmt.Sprintf("%.6f+(%.6f-%.6f)*(%s-%.6f)/%.6f", left.y, right.y, left.y, progressExpression, left.x, span)
		}
		expression = fmt.Sprintf("if(lt(%s,%.6f),%s,%s)", progressExpression, right.x, segment, expression)
	}
	return expression
}

func buildAICutKeyframeExpression(clip aicut.Clip, field string, base float64, timeExpression string) (string, bool) {
	frames := append([]aicut.ClipKeyframe(nil), clip.Keyframes...)
	sort.SliceStable(frames, func(i, j int) bool { return frames[i].OffsetMS < frames[j].OffsetMS })
	hasField := false
	for _, frame := range frames {
		if _, ok := aicutKeyframeField(frame.Values, field); ok {
			hasField = true
			break
		}
	}
	if !hasField || len(frames) == 0 {
		return "", false
	}
	values := make([]float64, len(frames))
	for index, frame := range frames {
		values[index] = base
		if value, ok := aicutKeyframeField(frame.Values, field); ok {
			values[index] = *value
		}
	}
	if len(frames) == 1 {
		return fmt.Sprintf("%.6f", values[0]), true
	}
	expression := fmt.Sprintf("%.6f", values[len(values)-1])
	for index := len(frames) - 2; index >= 0; index-- {
		leftTime := float64(frames[index].OffsetMS) / 1000
		rightTime := float64(frames[index+1].OffsetMS) / 1000
		leftValue, rightValue := values[index], values[index+1]
		segment := fmt.Sprintf("%.6f", rightValue)
		if rightTime > leftTime {
			linearProgress := fmt.Sprintf("(%s-%.6f)/%.6f", timeExpression, leftTime, rightTime-leftTime)
			easedProgress := aicutKeyframeEasingExpression(frames[index].Easing, linearProgress)
			segment = fmt.Sprintf("%.6f+(%.6f-%.6f)*(%s)", leftValue, rightValue, leftValue, easedProgress)
		}
		expression = fmt.Sprintf("if(lt(%s,%.6f),%s,%s)", timeExpression, rightTime, segment, expression)
	}
	firstTime := float64(frames[0].OffsetMS) / 1000
	return fmt.Sprintf("if(lte(%s,%.6f),%.6f,%s)", timeExpression, firstTime, values[0], expression), true
}

func aicutEvenDimension(value float64) int {
	dimension := max(2, int(math.Ceil(value)))
	if dimension%2 != 0 {
		dimension++
	}
	return dimension
}

func buildAICutAnimatedCanvasTransformFilters(width, height int, clip aicut.Clip, timelineComposition bool) []string {
	visualFields := []string{"x", "y", "scaleX", "scaleY", "rotation", "opacity"}
	hasVisualKeyframe := false
	for _, frame := range clip.Keyframes {
		for _, field := range visualFields {
			if _, ok := aicutKeyframeField(frame.Values, field); ok {
				hasVisualKeyframe = true
				break
			}
		}
	}
	if !hasVisualKeyframe {
		return buildAICutCanvasTransformFilters(width, height, clip.CanvasTransform)
	}
	base := aicutCanvasTransformValues(clip.CanvasTransform)
	timeExpression := "t"
	if timelineComposition && clip.StartMS > 0 {
		timeExpression = fmt.Sprintf("(t-%.6f)", float64(clip.StartMS)/1000)
	}
	expressions := make(map[string]string, len(visualFields))
	for _, field := range visualFields {
		fieldTimeExpression := timeExpression
		if field == "opacity" {
			fieldTimeExpression = "T"
			if timelineComposition && clip.StartMS > 0 {
				fieldTimeExpression = fmt.Sprintf("(T-%.6f)", float64(clip.StartMS)/1000)
			}
		}
		expression, ok := buildAICutKeyframeExpression(clip, field, base[field], fieldTimeExpression)
		if !ok {
			expression = fmt.Sprintf("%.6f", base[field])
		}
		expressions[field] = expression
	}
	maxScaleX, maxScaleY := max(1.0, base["scaleX"]), max(1.0, base["scaleY"])
	for _, frame := range clip.Keyframes {
		if value, ok := aicutKeyframeField(frame.Values, "scaleX"); ok {
			maxScaleX = max(maxScaleX, *value)
		}
		if value, ok := aicutKeyframeField(frame.Values, "scaleY"); ok {
			maxScaleY = max(maxScaleY, *value)
		}
	}
	maxWidth, maxHeight := aicutEvenDimension(float64(width)*maxScaleX), aicutEvenDimension(float64(height)*maxScaleY)
	diagonal := aicutEvenDimension(math.Hypot(float64(maxWidth), float64(maxHeight)))
	return []string{
		fmt.Sprintf("scale=w='max(2,trunc(iw*(%s)/2)*2)':h='max(2,trunc(ih*(%s)/2)*2)':eval=frame", expressions["scaleX"], expressions["scaleY"]),
		fmt.Sprintf("pad=%d:%d:(ow-iw)/2:(oh-ih)/2:color=black@0", maxWidth, maxHeight),
		fmt.Sprintf("rotate='(%s)*PI/180':ow=%d:oh=%d:c=black@0", expressions["rotation"], diagonal, diagonal),
		fmt.Sprintf("crop=%d:%d:x='max(0,min(iw-ow,(iw-ow)/2-(%s)*%d/2))':y='max(0,min(ih-oh,(ih-oh)/2-(%s)*%d/2))'", width, height, expressions["x"], width, expressions["y"], height),
		fmt.Sprintf("format=rgba,geq=r='r(X,Y)':g='g(X,Y)':b='b(X,Y)':a='alpha(X,Y)*(%s)'", expressions["opacity"]),
	}
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
	return aicutRenderSizeFor(aspectRatio, "1080p")
}

func aicutRenderSizeFor(aspectRatio, resolution string) (int, int) {
	shortEdge := 1080
	longEdge := 1920
	if resolution == "720p" {
		shortEdge = 720
		longEdge = 1280
	} else if resolution == "4k" {
		shortEdge = 2160
		longEdge = 3840
	}
	if aspectRatio == "16:9" {
		return longEdge, shortEdge
	}
	if aspectRatio == "1:1" {
		return shortEdge, shortEdge
	}
	return shortEdge, longEdge
}

func aicutASSColor(value, fallback string) string {
	value = strings.TrimSpace(value)
	if len(value) != 7 && len(value) != 9 {
		return fallback
	}
	red, redErr := strconv.ParseUint(value[1:3], 16, 8)
	green, greenErr := strconv.ParseUint(value[3:5], 16, 8)
	blue, blueErr := strconv.ParseUint(value[5:7], 16, 8)
	if redErr != nil || greenErr != nil || blueErr != nil {
		return fallback
	}
	alpha := uint64(0)
	if len(value) == 9 {
		opacity, err := strconv.ParseUint(value[7:9], 16, 8)
		if err != nil {
			return fallback
		}
		alpha = 255 - opacity
	}
	return fmt.Sprintf("&H%02X%02X%02X%02X", alpha, blue, green, red)
}

func aicutTextBackgroundVisible(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) == 7 {
		return validAICutHexColor(value)
	}
	if len(value) != 9 || !validAICutHexColor(value) {
		return false
	}
	alpha, err := strconv.ParseUint(value[7:9], 16, 8)
	return err == nil && alpha > 0
}

func validAICutHexColor(value string) bool {
	if (len(value) != 7 && len(value) != 9) || value[0] != '#' {
		return false
	}
	_, err := strconv.ParseUint(value[1:], 16, 64)
	return err == nil
}

func aicutKaraokeText(value string, durationMS int64) string {
	characters := []rune(value)
	if len(characters) == 0 {
		return ""
	}
	centiseconds := max(int64(1), durationMS/10/int64(len(characters)))
	var builder strings.Builder
	for _, character := range characters {
		builder.WriteString(fmt.Sprintf("{\\k%d}%s", centiseconds, escapeAICutASSText(string(character))))
	}
	return builder.String()
}

func aicutTextStyleRunAt(settings aicut.TrackSettings, index int) *aicut.TextStyleRun {
	var matched *aicut.TextStyleRun
	for runIndex := range settings.TextRuns {
		run := &settings.TextRuns[runIndex]
		if index >= run.Start && index < run.End {
			matched = run
		}
	}
	return matched
}

func aicutASSRunTag(settings aicut.TrackSettings, index int, primary string, fontSize int) string {
	if len(settings.TextRuns) == 0 {
		return ""
	}
	color, size, bold, italic := primary, fontSize, 0, 0
	if run := aicutTextStyleRunAt(settings, index); run != nil {
		color = aicutASSColor(run.TextColor, primary)
		if run.FontSize > 0 {
			size = int(math.Round(run.FontSize))
		}
		if run.Bold {
			bold = 1
		}
		if run.Italic {
			italic = 1
		}
	}
	return fmt.Sprintf("{\\c%s&\\fs%d\\b%d\\i%d}", color, size, bold, italic)
}

func aicutASSRichText(value string, settings aicut.TrackSettings, primary string, fontSize int, karaokeDurationMS int64) string {
	characters := []rune(value)
	centiseconds := int64(0)
	if karaokeDurationMS > 0 && len(characters) > 0 {
		centiseconds = max(int64(1), karaokeDurationMS/10/int64(len(characters)))
	}
	var builder strings.Builder
	for index, character := range characters {
		builder.WriteString(aicutASSRunTag(settings, index, primary, fontSize))
		if centiseconds > 0 {
			builder.WriteString(fmt.Sprintf("{\\k%d}", centiseconds))
		}
		builder.WriteString(escapeAICutASSText(string(character)))
	}
	return builder.String()
}

func buildAICutCurvedASSEvents(clip aicut.Clip, settings aicut.TrackSettings, styleName, prefix, primary string, width, height, fontSize int) []string {
	text := strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ").Replace(strings.TrimSpace(clip.Text))
	characters := []rune(text)
	if len(characters) == 0 {
		return nil
	}
	curve := math.Max(-180, math.Min(180, settings.TextCurve))
	advance := float64(fontSize)*0.62 + settings.LetterSpacing
	centerX := float64(width) / 2
	baseY := float64(height - 70)
	arcHeight := math.Abs(curve) / 180 * float64(fontSize) * 1.8
	events := make([]string, 0, len(characters))
	for index, character := range characters {
		position := 0.0
		if len(characters) > 1 {
			position = float64(index)/float64(len(characters)-1)*2 - 1
		}
		x := centerX + (float64(index)-float64(len(characters)-1)/2)*advance
		y := baseY
		if curve != 0 {
			y -= math.Copysign((1-position*position)*arcHeight, curve)
		}
		rotation := curve * position * .5
		glyphPrefix := fmt.Sprintf("{\\an5\\pos(%.2f,%.2f)\\frz%.2f}", x, y, rotation)
		if settings.SubtitleAnimation == "typewriter" {
			delay := int(math.Round(float64(index) / float64(max(1, len(characters))) * float64(min(1200, int(clip.DurationMS)))))
			glyphPrefix += fmt.Sprintf("{\\alpha&HFF&\\t(%d,%d,\\alpha&H00&)}", delay, delay+40)
		}
		events = append(events, fmt.Sprintf("Dialogue: 0,%s,%s,%s,,0,0,0,,%s%s%s%s", aicutASSTime(clip.StartMS), aicutASSTime(clip.StartMS+clip.DurationMS), styleName, prefix, glyphPrefix, aicutASSRunTag(settings, index, primary, fontSize), escapeAICutASSText(string(character))))
	}
	return events
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
		fontName := "Microsoft YaHei"
		fontSize, primary, outline, back, borderStyle, outlineWidth := 54, "&H00FFFFFF", "&H00000000", "&H80000000", 1, 3
		spacing, scaleY, alignment := 0.0, 100, 2
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
		case "text-basic-caption":
			fontSize, outlineWidth = 48, 2
		case "text-bold-title":
			fontName, fontSize, outlineWidth = "SimHei", 82, 5
		case "text-gradient-title":
			fontName, fontSize, primary, outline, outlineWidth = "SimHei", 82, "&H00FF8BB7", "&H004F1625", 5
		case "text-outline-caption":
			fontSize, outlineWidth = 64, 6
		case "text-comic":
			fontName, fontSize, primary, outline, outlineWidth = "SimHei", 64, "&H0039E2FF", "&H00122D7C", 6
		case "text-shadow-caption":
			fontSize, back, outlineWidth = 58, "&H33000000", 2
		case "text-handwriting":
			fontName, fontSize, primary, outline, outlineWidth = "KaiTi", 58, "&H00EDF7FF", "&H00122D7C", 2
		case "text-panel-caption":
			fontSize, back, borderStyle, outlineWidth = 52, "&H47000000", 3, 0
		case "text-news-bar":
			fontName, fontSize, back, borderStyle, outlineWidth, alignment = "SimHei", 52, "&H22EF5E15", 3, 0, 1
		case "text-lower-third":
			fontSize, back, borderStyle, outlineWidth, alignment = 48, "&H26182711", 3, 0, 1
		case "text-word-highlight":
			fontSize, primary, outlineWidth = 64, "&H0000FFFF", 4
		case "text-vertical":
			fontName, fontSize, outlineWidth = "KaiTi", 58, 3
		}
		settings := track.Settings
		if settings.FontFamily != "" {
			fontName = settings.FontFamily
		}
		if settings.FontSize > 0 {
			fontSize = int(math.Round(settings.FontSize))
		}
		primary = aicutASSColor(settings.TextColor, primary)
		outline = aicutASSColor(settings.StrokeColor, outline)
		back = aicutASSColor(settings.BackgroundColor, back)
		if settings.StrokeWidth > 0 || settings.StrokeColor != "" {
			outlineWidth = int(math.Round(settings.StrokeWidth))
		}
		if aicutTextBackgroundVisible(settings.BackgroundColor) {
			borderStyle = 3
			outlineWidth = max(outlineWidth, 8)
		}
		spacing = settings.LetterSpacing
		if settings.LineHeight > 0 {
			scaleY = int(math.Round(settings.LineHeight / 1.3 * 100))
		}
		switch settings.TextAlign {
		case "left":
			alignment = 1
		case "right":
			alignment = 3
		}
		styles = append(styles, fmt.Sprintf("Style: %s,%s,%d,%s,&H000000FF,%s,%s,0,0,0,0,100,%d,%.2f,0,%d,%d,0,%d,40,40,70,1", styleName, fontName, fontSize, primary, outline, back, scaleY, spacing, borderStyle, outlineWidth, alignment))
		for _, clip := range track.Clips {
			text := strings.TrimSpace(clip.Text)
			if text == "" || clip.DurationMS <= 0 {
				continue
			}
			prefix := ""
			renderedText := escapeAICutASSText(text)
			if styleID == "text-vertical" {
				vertical := make([]string, 0, len([]rune(text)))
				for _, character := range []rune(text) {
					vertical = append(vertical, escapeAICutASSText(string(character)))
				}
				renderedText = strings.Join(vertical, "\\N")
			}
			switch track.Settings.SubtitleAnimation {
			case "", "none", "无动效":
			case "pop":
				prefix = "{\\fad(80,120)\\fscx118\\fscy118\\t(0,240,\\fscx100\\fscy100)}"
			case "float-up":
				prefix = fmt.Sprintf("{\\move(%d,%d,%d,%d,0,350)\\fad(100,120)}", width/2, height-35, width/2, height-70)
			case "typewriter":
				prefix = fmt.Sprintf("{\\clip(0,0,0,%d)\\t(0,1200,\\clip(0,0,%d,%d))}", height, width, height)
			case "word-highlight":
				prefix = "{\\fad(100,100)}"
				if len(settings.TextRuns) == 0 {
					renderedText = aicutKaraokeText(text, clip.DurationMS)
				}
			default:
				prefix = "{\\fad(200,200)}"
			}
			if settings.ShadowOffsetX != 0 || settings.ShadowOffsetY != 0 || settings.ShadowBlur != 0 {
				shadowColor := aicutASSColor(settings.ShadowColor, "&H80000000")
				prefix = fmt.Sprintf("{\\4c%s&\\xshad%.2f\\yshad%.2f\\blur%.2f}", shadowColor, settings.ShadowOffsetX, settings.ShadowOffsetY, settings.ShadowBlur) + prefix
			}
			if settings.LetterSpacing != 0 {
				prefix = fmt.Sprintf("{\\fsp%.2f}", settings.LetterSpacing) + prefix
			}
			if settings.TextCurve != 0 {
				events = append(events, buildAICutCurvedASSEvents(clip, settings, styleName, prefix, primary, width, height, fontSize)...)
				continue
			}
			if len(settings.TextRuns) > 0 {
				karaokeDuration := int64(0)
				if settings.SubtitleAnimation == "word-highlight" {
					karaokeDuration = clip.DurationMS
				}
				renderedText = aicutASSRichText(text, settings, primary, fontSize, karaokeDuration)
			}
			events = append(events, fmt.Sprintf("Dialogue: 0,%s,%s,%s,,0,0,0,,%s%s", aicutASSTime(clip.StartMS), aicutASSTime(clip.StartMS+clip.DurationMS), styleName, prefix, renderedText))
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
