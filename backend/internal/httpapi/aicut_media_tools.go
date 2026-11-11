package httpapi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/ffmpegbin"
)

var aicutSceneTimePattern = regexp.MustCompile(`pts_time:([0-9]+(?:\.[0-9]+)?)`)

type aicutAudioAnalysis struct {
	DurationMS int64     `json:"durationMs"`
	SampleRate int       `json:"sampleRate"`
	Peaks      []float64 `json:"peaks"`
	BeatsMS    []int64   `json:"beatsMs"`
}

func analyzeAICutPCM(reader io.Reader, sampleRate int, durationMS int64, pointCount int) (aicutAudioAnalysis, error) {
	if reader == nil || sampleRate < 100 || durationMS <= 0 {
		return aicutAudioAnalysis{}, errors.New("音频采样参数无效")
	}
	if pointCount < 16 || pointCount > 1024 {
		return aicutAudioAnalysis{}, errors.New("波形采样点必须在 16 到 1024 之间")
	}
	expectedSamples := int64(math.Ceil(float64(sampleRate) * float64(durationMS) / 1000))
	bucketSize := max(int64(1), int64(math.Ceil(float64(expectedSamples)/float64(pointCount))))
	windowSamples := max(1, sampleRate/50)
	peaks := make([]float64, pointCount)
	energies := make([]float64, 0, int(durationMS/20)+1)
	buffer := make([]byte, 64*1024)
	leftover := byte(0)
	hasLeftover := false
	var sampleIndex int64
	var windowSquare float64
	windowCount := 0
	consume := func(sample int16) {
		amplitude := math.Abs(float64(sample)) / 32768
		bucket := int(sampleIndex / bucketSize)
		if bucket >= pointCount {
			bucket = pointCount - 1
		}
		if amplitude > peaks[bucket] {
			peaks[bucket] = amplitude
		}
		windowSquare += amplitude * amplitude
		windowCount++
		if windowCount >= windowSamples {
			energies = append(energies, math.Sqrt(windowSquare/float64(windowCount)))
			windowSquare, windowCount = 0, 0
		}
		sampleIndex++
	}
	for {
		count, err := reader.Read(buffer)
		if count > 0 {
			start := 0
			if hasLeftover {
				consume(int16(binary.LittleEndian.Uint16([]byte{leftover, buffer[0]})))
				start, hasLeftover = 1, false
			}
			for start+1 < count {
				consume(int16(binary.LittleEndian.Uint16(buffer[start : start+2])))
				start += 2
			}
			if start < count {
				leftover, hasLeftover = buffer[start], true
			}
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return aicutAudioAnalysis{}, fmt.Errorf("读取音频采样失败：%w", err)
		}
	}
	if sampleIndex == 0 {
		return aicutAudioAnalysis{}, errors.New("音频没有可分析的采样")
	}
	if windowCount > 0 {
		energies = append(energies, math.Sqrt(windowSquare/float64(windowCount)))
	}
	beats := make([]int64, 0)
	lastBeat := int64(-250)
	for index := 1; index+1 < len(energies); index++ {
		from := max(0, index-20)
		average := 0.0
		for _, energy := range energies[from:index] {
			average += energy
		}
		average /= float64(max(1, index-from))
		current := energies[index]
		timeMS := int64(index * 20)
		if current >= energies[index-1] && current > energies[index+1] && current >= .05 && current > average*1.8 && timeMS-lastBeat >= 250 {
			beats = append(beats, timeMS)
			lastBeat = timeMS
			if len(beats) >= 2000 {
				break
			}
		}
	}
	return aicutAudioAnalysis{DurationMS: durationMS, SampleRate: sampleRate, Peaks: peaks, BeatsMS: beats}, nil
}

func RegisterAICutMediaToolRoutes(mux *http.ServeMux, auth BridgeAuth, store aicut.Store, objects agentworkspace.AssetObjectStore, now func() time.Time) {
	if mux == nil || store == nil {
		return
	}
	if now == nil {
		now = time.Now
	}
	handle := func(pattern string, fn func(http.ResponseWriter, *http.Request, string)) {
		mux.Handle(pattern, auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity, ok := BridgeIdentityFromContext(req.Context())
			if !ok || strings.TrimSpace(identity.Username) == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			if objects == nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "素材存储尚未配置"})
				return
			}
			fn(w, req, strings.TrimSpace(identity.Username))
		})))
	}

	handle("POST /api/ai-cut/projects/{id}/smart-split", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			AssetID          string  `json:"assetId"`
			Threshold        float64 `json:"threshold"`
			MinIntervalMS    int64   `json:"minIntervalMs"`
			ExpectedRevision int64   `json:"expectedRevision"`
		}
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		projectID := strings.TrimSpace(req.PathValue("id"))
		timeline, splits, message, err := RunAICutSmartSplit(req.Context(), store, objects, now, owner, projectID, strings.TrimSpace(input.AssetID), input.Threshold, input.MinIntervalMS, input.ExpectedRevision)
		if err != nil {
			if errors.Is(err, aicut.ErrRevisionConflict) || errors.Is(err, aicut.ErrProjectNotFound) || errors.Is(err, aicut.ErrAssetNotFound) {
				writeAICutError(w, err)
				return
			}
			status := http.StatusUnprocessableEntity
			if ffmpegbin.IsMissing(err) {
				status = http.StatusServiceUnavailable
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"timeline": timeline, "splitCount": splits, "message": message})
	})

	handle("POST /api/ai-cut/projects/{id}/extract-audio", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			AssetID string `json:"assetId"`
		}
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		projectID := strings.TrimSpace(req.PathValue("id"))
		created, err := RunAICutExtractAudio(req.Context(), store, objects, now, owner, projectID, strings.TrimSpace(input.AssetID))
		if err != nil {
			if errors.Is(err, aicut.ErrProjectNotFound) || errors.Is(err, aicut.ErrAssetNotFound) {
				writeAICutError(w, err)
				return
			}
			status := http.StatusUnprocessableEntity
			if ffmpegbin.IsMissing(err) {
				status = http.StatusServiceUnavailable
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		created.URL = aicutSignedAssetContentURL(auth, owner, projectID, created.ID, "content", now())
		writeJSON(w, http.StatusCreated, map[string]any{"asset": created})
	})

	handle("POST /api/ai-cut/projects/{id}/consolidate", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			ClipIDs          []string `json:"clipIds"`
			ExpectedRevision int64    `json:"expectedRevision"`
		}
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		projectID := strings.TrimSpace(req.PathValue("id"))
		created, err := RunAICutConsolidate(req.Context(), store, objects, now, owner, projectID, input.ClipIDs, input.ExpectedRevision)
		if err != nil {
			if errors.Is(err, aicut.ErrRevisionConflict) || errors.Is(err, aicut.ErrProjectNotFound) || errors.Is(err, aicut.ErrAssetNotFound) {
				writeAICutError(w, err)
				return
			}
			status := http.StatusUnprocessableEntity
			if ffmpegbin.IsMissing(err) {
				status = http.StatusServiceUnavailable
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		created.URL = aicutSignedAssetContentURL(auth, owner, projectID, created.ID, "content", now())
		writeJSON(w, http.StatusCreated, map[string]any{"asset": created})
	})

	handle("GET /api/ai-cut/projects/{id}/audio-analysis", func(w http.ResponseWriter, req *http.Request, owner string) {
		projectID := strings.TrimSpace(req.PathValue("id"))
		assetID := strings.TrimSpace(req.URL.Query().Get("assetId"))
		points := 240
		if raw := strings.TrimSpace(req.URL.Query().Get("points")); raw != "" {
			value, err := strconv.Atoi(raw)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "波形采样点无效"})
				return
			}
			points = value
		}
		if assetID == "" || points < 16 || points > 1024 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择音频或视频素材，波形采样点必须在 16 到 1024 之间"})
			return
		}
		asset, err := store.Asset(req.Context(), owner, projectID, assetID)
		if err != nil {
			writeAICutError(w, err)
			return
		}
		if asset.Kind != "audio" && asset.Kind != "video" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "波形和节拍分析只支持音频或视频素材"})
			return
		}
		etag := fmt.Sprintf("\"aicut-audio-%s-%d-%d-%d\"", asset.ID, asset.ByteSize, asset.DurationMS, points)
		w.Header().Set("Cache-Control", "private, max-age=86400")
		w.Header().Set("ETag", etag)
		if req.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		analysis, err := analyzeAICutAssetAudio(req.Context(), asset, objects, points)
		if err != nil {
			status := http.StatusUnprocessableEntity
			if ffmpegbin.IsMissing(err) {
				status = http.StatusServiceUnavailable
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"analysis": analysis})
	})

	handle("POST /api/ai-cut/projects/{id}/visual-frames", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			AssetID string `json:"assetId"`
			Count   int    `json:"count"`
		}
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		if input.Count == 0 {
			input.Count = 6
		}
		if input.Count < 1 || input.Count > 8 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "抽帧数量必须在 1 到 8 之间"})
			return
		}
		projectID := strings.TrimSpace(req.PathValue("id"))
		asset, err := store.Asset(req.Context(), owner, projectID, strings.TrimSpace(input.AssetID))
		if err != nil {
			writeAICutError(w, err)
			return
		}
		if asset.Kind == "image" {
			writeJSON(w, http.StatusOK, map[string]any{"frames": []map[string]any{{"timeMs": 0, "url": aicutSignedAssetContentURL(auth, owner, projectID, asset.ID, "content", now())}}, "message": "图片素材无需抽帧"})
			return
		}
		if asset.Kind != "video" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "视觉抽帧只支持视频或图片素材"})
			return
		}
		frames, err := extractAICutVisualFrames(req.Context(), asset, objects, input.Count)
		if err != nil {
			status := http.StatusUnprocessableEntity
			if ffmpegbin.IsMissing(err) {
				status = http.StatusServiceUnavailable
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"frames": frames, "message": fmt.Sprintf("已提取 %d 张代表帧", len(frames))})
	})

	handle("POST /api/ai-cut/projects/{id}/tracking-frames", func(w http.ResponseWriter, req *http.Request, owner string) {
		var input struct {
			AssetID    string `json:"assetId"`
			StartMS    int64  `json:"startMs"`
			DurationMS int64  `json:"durationMs"`
			Count      int    `json:"count"`
			Width      int    `json:"width"`
		}
		if !decodeAICutJSON(w, req, &input) {
			return
		}
		if input.Width == 0 {
			input.Width = 192
		}
		if input.DurationMS < 1 || input.DurationMS > 300000 || input.Count < 2 || input.Count > 60 || input.Width < 96 || input.Width > 320 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "跟踪区间最长 5 分钟，帧数必须在 2 到 60 之间，宽度必须在 96 到 320 之间"})
			return
		}
		projectID := strings.TrimSpace(req.PathValue("id"))
		asset, err := store.Asset(req.Context(), owner, projectID, strings.TrimSpace(input.AssetID))
		if err != nil {
			writeAICutError(w, err)
			return
		}
		if asset.Kind != "video" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "区域跟踪只支持视频素材"})
			return
		}
		frames, err := extractAICutTrackingFrames(req.Context(), asset, objects, input.StartMS, input.DurationMS, input.Count, input.Width)
		if err != nil {
			status := http.StatusUnprocessableEntity
			if ffmpegbin.IsMissing(err) {
				status = http.StatusServiceUnavailable
			}
			writeJSON(w, status, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"frames": frames, "message": fmt.Sprintf("已提取 %d 张跟踪帧", len(frames))})
	})
}

// RunAICutSmartSplit is shared by the HTTP workbench and AI-CUT Agent. Its
// defaults match capcut-cli detect-scenes: threshold 0.4 and a 2 second gap.
func RunAICutSmartSplit(ctx context.Context, store aicut.Store, objects agentworkspace.AssetObjectStore, now func() time.Time, owner, projectID, assetID string, threshold float64, minIntervalMS, expectedRevision int64) (aicut.Timeline, int, string, error) {
	if threshold == 0 {
		threshold = 0.4
	}
	if minIntervalMS == 0 {
		minIntervalMS = 2000
	}
	if threshold <= 0 || threshold > 1 || minIntervalMS < 100 || minIntervalMS > 60000 {
		return aicut.Timeline{}, 0, "", errors.New("切镜阈值或最小间隔无效")
	}
	asset, err := store.Asset(ctx, owner, projectID, assetID)
	if err != nil {
		return aicut.Timeline{}, 0, "", err
	}
	if asset.Kind != "video" {
		return aicut.Timeline{}, 0, "", errors.New("智能切镜只支持视频素材")
	}
	current, err := store.Timeline(ctx, owner, projectID)
	if err != nil {
		return aicut.Timeline{}, 0, "", err
	}
	if current.Revision != expectedRevision {
		return current, 0, "", aicut.ErrRevisionConflict
	}
	times, err := detectAICutScenes(ctx, asset, objects, threshold, minIntervalMS)
	if err != nil {
		return aicut.Timeline{}, 0, "", err
	}
	prepared, splits := applyAICutSceneSplits(current, asset.ID, times)
	if splits == 0 {
		return current, 0, "没有检测到满足条件的新镜头", nil
	}
	if now == nil {
		now = time.Now
	}
	saved, err := store.SaveTimeline(ctx, owner, projectID, prepared, expectedRevision, now().UTC())
	if err != nil {
		return aicut.Timeline{}, 0, "", err
	}
	return saved, splits, fmt.Sprintf("已智能切出 %d 个新片段", splits), nil
}

func RunAICutExtractAudio(ctx context.Context, store aicut.Store, objects agentworkspace.AssetObjectStore, now func() time.Time, owner, projectID, assetID string) (aicut.Asset, error) {
	asset, err := store.Asset(ctx, owner, projectID, assetID)
	if err != nil {
		return aicut.Asset{}, err
	}
	if asset.Kind != "video" {
		return aicut.Asset{}, errors.New("只能从视频素材分离音频")
	}
	if now == nil {
		now = time.Now
	}
	created, err := extractAICutAudio(ctx, owner, asset, objects, now().UTC())
	if err != nil {
		return aicut.Asset{}, err
	}
	if err := store.CreateAsset(ctx, owner, created); err != nil {
		_ = objects.Delete(ctx, aicutObjectAsset(created))
		return aicut.Asset{}, err
	}
	return created, nil
}

func prepareAICutConsolidationTimeline(timeline aicut.Timeline, clipIDs []string) (aicut.Timeline, int64, error) {
	if len(clipIDs) < 2 || len(clipIDs) > 50 {
		return aicut.Timeline{}, 0, errors.New("请选择 2 到 50 个视频片段进行合并")
	}
	wanted := make(map[string]struct{}, len(clipIDs))
	for _, rawID := range clipIDs {
		id := strings.TrimSpace(rawID)
		if id == "" {
			return aicut.Timeline{}, 0, errors.New("所选片段编号无效")
		}
		if _, exists := wanted[id]; exists {
			return aicut.Timeline{}, 0, errors.New("所选片段不能重复")
		}
		wanted[id] = struct{}{}
	}
	var selectedTrack *aicut.Track
	selected := make([]aicut.Clip, 0, len(wanted))
	for trackIndex := range timeline.Tracks {
		track := &timeline.Tracks[trackIndex]
		for _, clip := range track.Clips {
			if _, ok := wanted[clip.ID]; !ok {
				continue
			}
			if track.Locked {
				return aicut.Timeline{}, 0, errors.New("锁定轨道中的片段不能合并")
			}
			if track.Kind != aicut.TrackVideo || clip.Kind != aicut.ClipVideo {
				return aicut.Timeline{}, 0, errors.New("渲染合并只支持同一画面轨道中的视频片段")
			}
			if selectedTrack != nil && selectedTrack.ID != track.ID {
				return aicut.Timeline{}, 0, errors.New("请选择同一画面轨道中的视频片段")
			}
			selectedTrack = track
			selected = append(selected, clip)
		}
	}
	if selectedTrack == nil || len(selected) != len(wanted) {
		return aicut.Timeline{}, 0, errors.New("部分所选片段已经不存在，请刷新后重试")
	}
	sort.Slice(selected, func(i, j int) bool {
		if selected[i].StartMS == selected[j].StartMS {
			return selected[i].ID < selected[j].ID
		}
		return selected[i].StartMS < selected[j].StartMS
	})
	startMS, endMS := selected[0].StartMS, selected[0].StartMS+selected[0].DurationMS
	if startMS < 0 || selected[0].DurationMS <= 0 {
		return aicut.Timeline{}, 0, errors.New("所选片段时间范围无效")
	}
	for index := range selected {
		clipEnd := selected[index].StartMS + selected[index].DurationMS
		if selected[index].DurationMS <= 0 || selected[index].StartMS < startMS {
			return aicut.Timeline{}, 0, errors.New("所选片段时间范围无效")
		}
		if clipEnd > endMS {
			endMS = clipEnd
		}
		selected[index].StartMS -= startMS
	}
	durationMS := endMS - startMS
	if durationMS <= 0 || durationMS > 30*60*1000 {
		return aicut.Timeline{}, 0, errors.New("单次渲染合并时长必须在 30 分钟以内")
	}
	track := *selectedTrack
	track.Clips = selected
	track.Locked = false
	return aicut.Timeline{Revision: timeline.Revision, Tracks: []aicut.Track{track}}, durationMS, nil
}

func RunAICutConsolidate(ctx context.Context, store aicut.Store, objects agentworkspace.AssetObjectStore, now func() time.Time, owner, projectID string, clipIDs []string, expectedRevision int64) (aicut.Asset, error) {
	if store == nil || objects == nil {
		return aicut.Asset{}, errors.New("渲染合并服务尚未配置")
	}
	project, err := store.Project(ctx, owner, projectID)
	if err != nil {
		return aicut.Asset{}, err
	}
	timeline, err := store.Timeline(ctx, owner, projectID)
	if err != nil {
		return aicut.Asset{}, err
	}
	if timeline.Revision != expectedRevision {
		return aicut.Asset{}, aicut.ErrRevisionConflict
	}
	prepared, durationMS, err := prepareAICutConsolidationTimeline(timeline, clipIDs)
	if err != nil {
		return aicut.Asset{}, err
	}
	output, cleanup, err := renderAICutExport(ctx, owner, project, prepared, "mp4", store, objects)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		return aicut.Asset{}, fmt.Errorf("片段渲染合并失败：%w", err)
	}
	info, err := os.Stat(output)
	if err != nil || info.Size() == 0 {
		return aicut.Asset{}, errors.New("片段渲染合并没有生成有效视频")
	}
	if now == nil {
		now = time.Now
	}
	createdAt := now().UTC()
	id := newAgentID("aicut-asset")
	objectKey, err := agentworkspace.AgentAssetObjectKey(owner, projectID, id, ".mp4")
	if err != nil {
		return aicut.Asset{}, errors.New("合并素材路径无效")
	}
	objectKey = strings.Replace(objectKey, "agent-assets/", "aicut-assets/", 1)
	created := aicut.Asset{ID: id, ProjectID: projectID, Kind: "video", Name: fmt.Sprintf("合并片段-%s.mp4", createdAt.Format("20060102-150405")), ContentType: "video/mp4", ObjectKey: objectKey, ByteSize: info.Size(), DurationMS: durationMS, CreatedAt: createdAt}
	file, err := os.Open(output)
	if err != nil {
		return aicut.Asset{}, errors.New("合并视频读取失败")
	}
	putErr := objects.Put(ctx, aicutObjectAsset(created), file)
	closeErr := file.Close()
	if putErr != nil || closeErr != nil {
		return aicut.Asset{}, errors.New("合并视频保存失败")
	}
	if err := store.CreateAsset(ctx, owner, created); err != nil {
		_ = objects.Delete(ctx, aicutObjectAsset(created))
		return aicut.Asset{}, err
	}
	return created, nil
}

func analyzeAICutAssetAudio(ctx context.Context, asset aicut.Asset, objects agentworkspace.AssetObjectStore, pointCount int) (aicutAudioAnalysis, error) {
	if asset.Kind != "audio" && asset.Kind != "video" {
		return aicutAudioAnalysis{}, errors.New("波形和节拍分析只支持音频或视频素材")
	}
	durationMS := asset.DurationMS
	if durationMS <= 0 {
		var err error
		durationMS, err = probeAICutAssetDuration(ctx, asset, objects)
		if err != nil || durationMS <= 0 {
			return aicutAudioAnalysis{}, errors.New("无法读取音频时长")
		}
	}
	if durationMS > 2*60*60*1000 {
		return aicutAudioAnalysis{}, errors.New("单次音频分析最长支持 2 小时")
	}
	dir, err := os.MkdirTemp("", "qiantie-aicut-audio-analysis-")
	if err != nil {
		return aicutAudioAnalysis{}, errors.New("无法创建音频分析临时目录")
	}
	defer os.RemoveAll(dir)
	input := filepath.Join(dir, "source"+path.Ext(asset.Name))
	if err := copyAICutObjectToFile(ctx, objects, asset, input); err != nil {
		return aicutAudioAnalysis{}, err
	}
	const sampleRate = 2000
	command, _, err := ffmpegbin.Command(ctx, "ffmpeg", "-hide_banner", "-loglevel", "error", "-i", input, "-map", "0:a:0", "-vn", "-ac", "1", "-ar", strconv.Itoa(sampleRate), "-f", "s16le", "-acodec", "pcm_s16le", "pipe:1")
	if err != nil {
		return aicutAudioAnalysis{}, err
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return aicutAudioAnalysis{}, errors.New("无法读取音频分析输出")
	}
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Start(); err != nil {
		return aicutAudioAnalysis{}, fmt.Errorf("音频分析启动失败：%w", err)
	}
	analysis, analyzeErr := analyzeAICutPCM(stdout, sampleRate, durationMS, pointCount)
	waitErr := command.Wait()
	if waitErr != nil {
		return aicutAudioAnalysis{}, fmt.Errorf("音频分析失败：%s", compactAICutToolOutput(stderr.Bytes()))
	}
	if analyzeErr != nil {
		return aicutAudioAnalysis{}, analyzeErr
	}
	return analysis, nil
}

func aicutVisualFrameTimes(durationMS int64, count int) []int64 {
	if durationMS <= 0 || count <= 0 {
		return nil
	}
	if count > 8 {
		count = 8
	}
	start := float64(durationMS) * 0.1
	end := float64(durationMS) * 0.9
	if count == 1 {
		return []int64{int64(math.Round(float64(durationMS) / 2))}
	}
	result := make([]int64, count)
	step := (end - start) / float64(count-1)
	for index := range result {
		result[index] = int64(math.Round(start + step*float64(index)))
	}
	return result
}

func aicutTrackingFrameTimes(assetDurationMS, startMS, durationMS int64, count int) ([]int64, error) {
	if assetDurationMS <= 0 || startMS < 0 || durationMS <= 0 || durationMS > 300000 || count < 2 || count > 60 {
		return nil, errors.New("跟踪抽帧参数无效")
	}
	if startMS+durationMS > assetDurationMS {
		return nil, errors.New("跟踪区间超出视频素材范围")
	}
	result := make([]int64, count)
	for index := range result {
		result[index] = startMS + int64(math.Round(float64(durationMS)*float64(index)/float64(count-1)))
	}
	return result, nil
}

type aicutVisualFrame struct {
	TimeMS int64  `json:"timeMs"`
	URL    string `json:"url"`
}

func extractAICutVisualFrames(ctx context.Context, asset aicut.Asset, objects agentworkspace.AssetObjectStore, count int) ([]aicutVisualFrame, error) {
	durationMS := asset.DurationMS
	if durationMS <= 0 {
		var err error
		durationMS, err = probeAICutAssetDuration(ctx, asset, objects)
		if err != nil || durationMS <= 0 {
			return nil, fmt.Errorf("无法读取视频时长")
		}
	}
	dir, err := os.MkdirTemp("", "qiantie-aicut-frames-")
	if err != nil {
		return nil, fmt.Errorf("无法创建视觉抽帧临时目录")
	}
	defer os.RemoveAll(dir)
	input := filepath.Join(dir, "source"+path.Ext(asset.Name))
	if err := copyAICutObjectToFile(ctx, objects, asset, input); err != nil {
		return nil, err
	}
	times := aicutVisualFrameTimes(durationMS, count)
	frames := make([]aicutVisualFrame, 0, len(times))
	for index, timeMS := range times {
		output := filepath.Join(dir, fmt.Sprintf("frame-%02d.jpg", index+1))
		if err := ffmpegbin.Run(ctx, "ffmpeg", "-hide_banner", "-y", "-ss", fmt.Sprintf("%.3f", float64(timeMS)/1000), "-i", input, "-frames:v", "1", "-vf", "scale=640:-2:force_original_aspect_ratio=decrease", "-q:v", "3", output); err != nil {
			return nil, fmt.Errorf("视觉抽帧失败：%w", err)
		}
		data, err := os.ReadFile(output)
		if err != nil || len(data) == 0 {
			return nil, fmt.Errorf("视觉抽帧结果读取失败")
		}
		frames = append(frames, aicutVisualFrame{TimeMS: timeMS, URL: "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(data)})
	}
	return frames, nil
}

func extractAICutTrackingFrames(ctx context.Context, asset aicut.Asset, objects agentworkspace.AssetObjectStore, startMS, durationMS int64, count, width int) ([]aicutVisualFrame, error) {
	assetDurationMS := asset.DurationMS
	if assetDurationMS <= 0 {
		var err error
		assetDurationMS, err = probeAICutAssetDuration(ctx, asset, objects)
		if err != nil || assetDurationMS <= 0 {
			return nil, errors.New("无法读取视频时长")
		}
	}
	times, err := aicutTrackingFrameTimes(assetDurationMS, startMS, durationMS, count)
	if err != nil {
		return nil, err
	}
	dir, err := os.MkdirTemp("", "qiantie-aicut-tracking-")
	if err != nil {
		return nil, errors.New("无法创建跟踪抽帧临时目录")
	}
	defer os.RemoveAll(dir)
	input := filepath.Join(dir, "source"+path.Ext(asset.Name))
	if err := copyAICutObjectToFile(ctx, objects, asset, input); err != nil {
		return nil, err
	}
	pattern := filepath.Join(dir, "frame-%03d.jpg")
	fps := float64(count-1) * 1000 / float64(durationMS)
	filter := fmt.Sprintf("fps=%.9f,scale=%d:-2:force_original_aspect_ratio=decrease", fps, width)
	if err := ffmpegbin.Run(ctx, "ffmpeg", "-hide_banner", "-y", "-ss", fmt.Sprintf("%.3f", float64(startMS)/1000), "-t", fmt.Sprintf("%.3f", float64(durationMS)/1000), "-i", input, "-an", "-vf", filter, "-frames:v", strconv.Itoa(count), "-q:v", "5", pattern); err != nil {
		return nil, fmt.Errorf("跟踪抽帧失败：%w", err)
	}
	frames := make([]aicutVisualFrame, 0, count)
	for index, timeMS := range times {
		data, readErr := os.ReadFile(filepath.Join(dir, fmt.Sprintf("frame-%03d.jpg", index+1)))
		if errors.Is(readErr, os.ErrNotExist) {
			break
		}
		if readErr != nil || len(data) == 0 {
			return nil, errors.New("跟踪抽帧结果读取失败")
		}
		frames = append(frames, aicutVisualFrame{TimeMS: timeMS, URL: "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(data)})
	}
	if len(frames) < 2 {
		return nil, errors.New("视频区间内没有足够的可跟踪画面")
	}
	return frames, nil
}

func detectAICutScenes(ctx context.Context, asset aicut.Asset, objects agentworkspace.AssetObjectStore, threshold float64, minIntervalMS int64) ([]int64, error) {
	dir, err := os.MkdirTemp("", "qiantie-aicut-scenes-")
	if err != nil {
		return nil, fmt.Errorf("无法创建切镜临时目录")
	}
	defer os.RemoveAll(dir)
	input := filepath.Join(dir, "source"+path.Ext(asset.Name))
	if err := copyAICutObjectToFile(ctx, objects, asset, input); err != nil {
		return nil, err
	}
	command, _, err := ffmpegbin.Command(ctx, "ffmpeg", "-hide_banner", "-i", input, "-vf", fmt.Sprintf("select='gt(scene,%.3f)',showinfo", threshold), "-an", "-f", "null", "-")
	if err != nil {
		return nil, err
	}
	output, runErr := command.CombinedOutput()
	if runErr != nil {
		return nil, fmt.Errorf("智能切镜分析失败：%s", compactAICutToolOutput(output))
	}
	return parseAICutSceneTimes(output, minIntervalMS), nil
}

func parseAICutSceneTimes(output []byte, minIntervalMS int64) []int64 {
	result := make([]int64, 0)
	last := int64(-minIntervalMS)
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		match := aicutSceneTimePattern.FindStringSubmatch(scanner.Text())
		if len(match) != 2 {
			continue
		}
		seconds, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			continue
		}
		value := int64(seconds * 1000)
		if value >= 100 && value-last >= minIntervalMS {
			result = append(result, value)
			last = value
		}
	}
	return result
}

func applyAICutSceneSplits(timeline aicut.Timeline, assetID string, times []int64) (aicut.Timeline, int) {
	result := timeline
	splits := 0
	for trackIndex := range result.Tracks {
		track := &result.Tracks[trackIndex]
		if track.Kind != aicut.TrackVideo || track.Locked {
			continue
		}
		next := make([]aicut.Clip, 0, len(track.Clips)+len(times))
		for _, clip := range track.Clips {
			if clip.AssetID != assetID || clip.Kind != aicut.ClipVideo {
				next = append(next, clip)
				continue
			}
			boundaries := []int64{clip.SourceStartMS}
			for _, value := range times {
				if value > clip.SourceStartMS && value < clip.SourceStartMS+clip.DurationMS {
					boundaries = append(boundaries, value)
				}
			}
			boundaries = append(boundaries, clip.SourceStartMS+clip.DurationMS)
			sort.Slice(boundaries, func(i, j int) bool { return boundaries[i] < boundaries[j] })
			for index := 0; index < len(boundaries)-1; index++ {
				part := clip
				if index > 0 {
					part.ID = newAgentID("aicut-clip")
					part.Name = fmt.Sprintf("%s（镜头%d）", clip.Name, index+1)
					splits++
				}
				part.SourceStartMS = boundaries[index]
				part.DurationMS = boundaries[index+1] - boundaries[index]
				part.StartMS = clip.StartMS + boundaries[index] - clip.SourceStartMS
				next = append(next, part)
			}
		}
		track.Clips = next
	}
	return result, splits
}

func extractAICutAudio(ctx context.Context, owner string, source aicut.Asset, objects agentworkspace.AssetObjectStore, now time.Time) (aicut.Asset, error) {
	dir, err := os.MkdirTemp("", "qiantie-aicut-audio-")
	if err != nil {
		return aicut.Asset{}, fmt.Errorf("无法创建分离音频临时目录")
	}
	defer os.RemoveAll(dir)
	input, output := filepath.Join(dir, "source"+path.Ext(source.Name)), filepath.Join(dir, "audio.mp3")
	if err := copyAICutObjectToFile(ctx, objects, source, input); err != nil {
		return aicut.Asset{}, err
	}
	if err := ffmpegbin.Run(ctx, "ffmpeg", "-hide_banner", "-y", "-i", input, "-vn", "-c:a", "libmp3lame", "-q:a", "2", output); err != nil {
		return aicut.Asset{}, fmt.Errorf("分离音频失败：%w", err)
	}
	info, err := os.Stat(output)
	if err != nil || info.Size() == 0 {
		return aicut.Asset{}, fmt.Errorf("视频没有可分离的音频轨")
	}
	id := newAgentID("aicut-asset")
	objectKey, err := agentworkspace.AgentAssetObjectKey(owner, source.ProjectID, id, ".mp3")
	if err != nil {
		return aicut.Asset{}, fmt.Errorf("音频素材路径无效")
	}
	objectKey = strings.Replace(objectKey, "agent-assets/", "aicut-assets/", 1)
	created := aicut.Asset{ID: id, ProjectID: source.ProjectID, Kind: "audio", Name: strings.TrimSuffix(source.Name, path.Ext(source.Name)) + "-分离音频.mp3", ContentType: "audio/mpeg", ObjectKey: objectKey, ByteSize: info.Size(), DurationMS: source.DurationMS, CreatedAt: now}
	file, err := os.Open(output)
	if err != nil {
		return aicut.Asset{}, fmt.Errorf("分离音频读取失败")
	}
	defer file.Close()
	if err := objects.Put(ctx, aicutObjectAsset(created), file); err != nil {
		return aicut.Asset{}, fmt.Errorf("分离音频保存失败")
	}
	return created, nil
}

func copyAICutObjectToFile(ctx context.Context, objects agentworkspace.AssetObjectStore, asset aicut.Asset, target string) error {
	body, err := objects.Open(ctx, aicutObjectAsset(asset))
	if err != nil {
		return fmt.Errorf("素材“%s”读取失败", asset.Name)
	}
	defer body.Close()
	file, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("素材“%s”准备失败", asset.Name)
	}
	_, copyErr := io.Copy(file, body)
	closeErr := file.Close()
	if copyErr != nil || closeErr != nil {
		return fmt.Errorf("素材“%s”准备失败", asset.Name)
	}
	return nil
}

func probeAICutAssetDuration(ctx context.Context, asset aicut.Asset, objects agentworkspace.AssetObjectStore) (int64, error) {
	dir, err := os.MkdirTemp("", "qiantie-aicut-probe-")
	if err != nil {
		return 0, err
	}
	defer os.RemoveAll(dir)
	input := filepath.Join(dir, "source"+path.Ext(asset.Name))
	if err := copyAICutObjectToFile(ctx, objects, asset, input); err != nil {
		return 0, err
	}
	command, _, err := ffmpegbin.Command(ctx, "ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", input)
	if err != nil {
		return 0, err
	}
	output, err := command.Output()
	if err != nil {
		return 0, err
	}
	seconds, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil || seconds <= 0 {
		return 0, fmt.Errorf("素材时长无效")
	}
	return int64(seconds*1000 + 0.5), nil
}

func compactAICutToolOutput(output []byte) string {
	value := strings.TrimSpace(string(output))
	if len(value) > 500 {
		value = value[len(value)-500:]
	}
	if value == "" {
		return "FFmpeg 未返回可读错误"
	}
	return value
}
