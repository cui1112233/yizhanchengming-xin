package httpapi

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/ffmpegbin"
	"qiantie/backend/internal/prompt"
)

type aiSegmentClip struct {
	Index      int    `json:"index"`
	StartUS    int64  `json:"startUs"`
	EndUS      int64  `json:"endUs"`
	Subtitle   string `json:"subtitle,omitempty"`
	SourceID   string `json:"sourceId,omitempty"`
	SourceName string `json:"sourceName,omitempty"`
}

type aiSegmentTextTrack struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type aiSegmentSuggestion struct {
	EndClip int    `json:"end_clip"`
	Reason  string `json:"reason"`
}

type aiSegmentEpisode struct {
	Index           int     `json:"index"`
	StartClip       int     `json:"startClip"`
	EndClip         int     `json:"endClip"`
	StartSeconds    float64 `json:"startSeconds"`
	EndSeconds      float64 `json:"endSeconds"`
	DurationSeconds float64 `json:"durationSeconds"`
	Reason          string  `json:"reason,omitempty"`
}

type aiSegmentConstraints struct {
	FrontCount int     `json:"frontCount"`
	FrontMin   float64 `json:"frontMin"`
	FrontMax   float64 `json:"frontMax"`
	BackMin    float64 `json:"backMin"`
	BackTarget float64 `json:"backTarget"`
	BackMax    float64 `json:"backMax"`
}

type aiSegmentAnalyzeInput struct {
	Mode          string               `json:"mode"`
	ModelID       string               `json:"modelId"`
	Script        string               `json:"script"`
	SRT           string               `json:"srt"`
	DraftContent  string               `json:"draftContent"`
	TextTrackID   string               `json:"textTrackId"`
	Clips         []aiSegmentClip      `json:"clips"`
	Constraints   aiSegmentConstraints `json:"constraints"`
	SystemPrompt  string               `json:"systemPrompt,omitempty"`
	OutputDir     string               `json:"outputDir,omitempty"`
	VideoPath     string               `json:"videoPath,omitempty"`
	OutputFPS     int                  `json:"fps,omitempty"`
	OutputQuality string               `json:"resolution,omitempty"`
}

func defaultAISegmentConstraints() aiSegmentConstraints {
	return aiSegmentConstraints{FrontCount: 2, FrontMin: 90, FrontMax: 180, BackMin: 45, BackTarget: 60, BackMax: 90}
}

func normalizeAISegmentConstraints(value aiSegmentConstraints) aiSegmentConstraints {
	defaults := defaultAISegmentConstraints()
	if value == (aiSegmentConstraints{}) {
		return defaults
	}
	if value.FrontCount < 0 || value.FrontCount > 20 {
		value.FrontCount = defaults.FrontCount
	}
	if value.FrontMin <= 0 {
		value.FrontMin = defaults.FrontMin
	}
	if value.FrontMax < value.FrontMin {
		value.FrontMax = defaults.FrontMax
	}
	if value.BackMin <= 0 {
		value.BackMin = defaults.BackMin
	}
	if value.BackTarget < value.BackMin {
		value.BackTarget = defaults.BackTarget
	}
	if value.BackMax < value.BackTarget {
		value.BackMax = defaults.BackMax
	}
	return value
}

func registerAISegmentRoutes(root *http.ServeMux, auth BridgeAuth, provider batchfactoryv11.DirectorProvider, execute AgentTextCompletionExecutor) {
	root.Handle("POST /api/plugins/ai-segment/analyze", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input aiSegmentAnalyzeInput
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<20))
		if err := decoder.Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求参数无效"})
			return
		}
		input.Mode = strings.TrimSpace(input.Mode)
		if input.Mode == "" {
			input.Mode = "subtitle"
		}
		clips, tracks, warnings, err := resolveAISegmentClips(input)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		constraints := normalizeAISegmentConstraints(input.Constraints)
		var suggestions []aiSegmentSuggestion
		if input.Mode == "ai" {
			if strings.TrimSpace(input.Script) == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请先提供剧本文稿"})
				return
			}
			if strings.TrimSpace(input.ModelID) == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择模型"})
				return
			}
			script := strings.TrimSpace(input.Script)
			if len([]rune(script)) > 120000 {
				script = string([]rune(script)[:120000])
				warnings = append(warnings, "剧本文稿超过 120000 字符，分析时已截断后部")
			}
			identity, _ := BridgeIdentityFromContext(r.Context())
			suggestions, err = requestAISegmentSuggestions(r.Context(), identity.Username, input.ModelID, script, clips, constraints, provider, execute)
			if err != nil {
				writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
				return
			}
		}
		episodes, planWarnings := planAISegmentEpisodes(clips, suggestions, constraints)
		warnings = append(warnings, planWarnings...)
		subtitleCount := 0
		for _, clip := range clips {
			if strings.TrimSpace(clip.Subtitle) != "" {
				subtitleCount++
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"episodes": episodes, "clips": clips, "textTracks": tracks, "clipCount": len(clips),
			"subtitleCount": subtitleCount, "totalSeconds": roundSeconds(float64(clips[len(clips)-1].EndUS-clips[0].StartUS) / 1e6),
			"warnings": aiSegmentUniqueStrings(warnings),
		})
	})))
	root.Handle("POST /api/plugins/ai-segment/export", auth.Middleware(http.HandlerFunc(exportAISegmentVideo)))
}

func exportAISegmentVideo(w http.ResponseWriter, r *http.Request) {
	if !limitRequestBody(w, r, maxAISegmentUploadBytes+maxMultipartOverheadBytes) {
		return
	}
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		if isBodyTooLarge(err) {
			writePayloadTooLarge(w, maxAISegmentUploadBytes+maxMultipartOverheadBytes)
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "上传内容无法读取"})
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}
	file, header, err := r.FormFile("video")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请选择视频文件"})
		return
	}
	defer file.Close()
	var episodes []aiSegmentEpisode
	if json.Unmarshal([]byte(r.FormValue("episodes")), &episodes) != nil || len(episodes) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "分集结果无效，请先重新分析"})
		return
	}
	for index, episode := range episodes {
		if episode.EndSeconds <= episode.StartSeconds || episode.StartSeconds < 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("第 %d 集时间范围无效", index+1)})
			return
		}
	}
	tempDir, err := os.MkdirTemp("", "yizhan-ai-segment-")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "无法创建视频处理临时目录"})
		return
	}
	defer os.RemoveAll(tempDir)
	extension := strings.ToLower(filepath.Ext(header.Filename))
	if extension == "" || len(extension) > 10 {
		extension = ".mp4"
	}
	inputPath := filepath.Join(tempDir, "input"+extension)
	inputFile, err := os.Create(inputPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "无法保存待处理视频"})
		return
	}
	_, copyErr := io.Copy(inputFile, io.LimitReader(file, 20<<30))
	closeErr := inputFile.Close()
	if copyErr != nil || closeErr != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "视频上传未完成"})
		return
	}
	resolution := strings.TrimSpace(r.FormValue("resolution"))
	fps, _ := strconv.Atoi(r.FormValue("fps"))
	outputs := make([]string, 0, len(episodes))
	for index, episode := range episodes {
		outputPath := filepath.Join(tempDir, fmt.Sprintf("episode-%03d.mp4", index+1))
		args := buildAISegmentFFmpegArgs(inputPath, outputPath, episode, resolution, fps)
		command, _, commandErr := ffmpegbin.Command(r.Context(), "ffmpeg", args...)
		if commandErr != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": commandErr.Error()})
			return
		}
		output, runErr := command.CombinedOutput()
		if runErr != nil {
			message := strings.TrimSpace(string(output))
			if len(message) > 500 {
				message = message[len(message)-500:]
			}
			if message == "" {
				message = runErr.Error()
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("第 %d 集切割失败：%s", index+1, message)})
			return
		}
		outputs = append(outputs, outputPath)
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="ai-segment.zip"`)
	archive := zip.NewWriter(w)
	defer archive.Close()
	for index, outputPath := range outputs {
		entry, createErr := archive.Create(fmt.Sprintf("第%03d集.mp4", index+1))
		if createErr != nil {
			return
		}
		outputFile, openErr := os.Open(outputPath)
		if openErr != nil {
			return
		}
		_, _ = io.Copy(entry, outputFile)
		_ = outputFile.Close()
	}
}

func buildAISegmentFFmpegArgs(inputPath, outputPath string, episode aiSegmentEpisode, resolution string, fps int) []string {
	duration := episode.EndSeconds - episode.StartSeconds
	args := []string{"-y", "-ss", strconv.FormatFloat(episode.StartSeconds, 'f', 3, 64), "-i", inputPath, "-t", strconv.FormatFloat(duration, 'f', 3, 64), "-map", "0:v?", "-map", "0:a?"}
	height := 0
	switch resolution {
	case "1080p":
		height = 1080
	case "720p":
		height = 720
	}
	if height == 0 && fps <= 0 {
		args = append(args, "-c", "copy")
	} else {
		filters := []string{}
		if height > 0 {
			filters = append(filters, fmt.Sprintf("scale=-2:%d", height))
		}
		if fps > 0 {
			filters = append(filters, fmt.Sprintf("fps=%d", fps))
		}
		if len(filters) > 0 {
			args = append(args, "-vf", strings.Join(filters, ","))
		}
		args = append(args, "-c:v", "libx264", "-preset", "medium", "-crf", "20", "-c:a", "aac")
	}
	return append(args, "-movflags", "+faststart", outputPath)
}

func resolveAISegmentClips(input aiSegmentAnalyzeInput) ([]aiSegmentClip, []aiSegmentTextTrack, []string, error) {
	if len(input.Clips) > 0 {
		clips, err := normalizeAISegmentClips(input.Clips)
		return clips, nil, nil, err
	}
	switch input.Mode {
	case "subtitle":
		clips, err := parseAISegmentSRT(input.SRT)
		return clips, nil, nil, err
	case "main-track", "text-track", "ai":
		if strings.TrimSpace(input.DraftContent) == "" {
			return nil, nil, nil, errors.New("请选择剪映草稿")
		}
		clips, tracks, err := parseAISegmentDraft([]byte(input.DraftContent), input.TextTrackID)
		if err != nil {
			return nil, nil, nil, err
		}
		if input.Mode == "text-track" && strings.TrimSpace(input.TextTrackID) == "" {
			return nil, tracks, nil, errors.New("请选择包含字幕的文本轨道")
		}
		warnings := []string{}
		hasSubtitle := false
		for _, clip := range clips {
			hasSubtitle = hasSubtitle || strings.TrimSpace(clip.Subtitle) != ""
		}
		if !hasSubtitle {
			warnings = append(warnings, "没有读取到字幕，将只按主轨片段和剧本分析")
		}
		return clips, tracks, warnings, nil
	default:
		return nil, nil, nil, errors.New("分段模式无效")
	}
}

var srtTimingPattern = regexp.MustCompile(`(?m)^\s*(\d{2}):(\d{2}):(\d{2})[,.](\d{3})\s*-->\s*(\d{2}):(\d{2}):(\d{2})[,.](\d{3})[^\r\n]*\r?\n([\s\S]*?)(?:\r?\n\s*\r?\n|\z)`)

func parseAISegmentSRT(value string) ([]aiSegmentClip, error) {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\ufeff", "")) + "\n\n"
	matches := srtTimingPattern.FindAllStringSubmatch(value, -1)
	clips := make([]aiSegmentClip, 0, len(matches))
	for _, match := range matches {
		start := srtTimeUS(match[1:5])
		end := srtTimeUS(match[5:9])
		if end <= start {
			continue
		}
		text := strings.TrimSpace(strings.Join(strings.Fields(match[9]), " "))
		clips = append(clips, aiSegmentClip{Index: len(clips) + 1, StartUS: start, EndUS: end, Subtitle: text})
	}
	if len(clips) == 0 {
		return nil, errors.New("没有读取到有效字幕")
	}
	return normalizeAISegmentClips(clips)
}

func srtTimeUS(parts []string) int64 {
	values := make([]int64, 4)
	for i, part := range parts {
		values[i], _ = strconv.ParseInt(part, 10, 64)
	}
	return (((values[0]*60+values[1])*60+values[2])*1000 + values[3]) * 1000
}

func parseAISegmentDraft(raw []byte, selectedTextTrack string) ([]aiSegmentClip, []aiSegmentTextTrack, error) {
	var root map[string]any
	if json.Unmarshal(raw, &root) != nil {
		return nil, nil, errors.New("剪映草稿无法解析")
	}
	tracks, _ := root["tracks"].([]any)
	if len(tracks) == 0 {
		return nil, nil, errors.New("主视频轨道没有有效片段")
	}
	materials := draftTextMaterials(root)
	var videoSegments []any
	type timedText struct {
		start, end int64
		text       string
	}
	textByTrack := map[string][]timedText{}
	trackOptions := []aiSegmentTextTrack{}
	for trackIndex, rawTrack := range tracks {
		track, _ := rawTrack.(map[string]any)
		kind := strings.ToLower(aiSegmentStringValue(track["type"]))
		segments, _ := track["segments"].([]any)
		if kind == "video" && len(videoSegments) == 0 && len(segments) > 0 {
			videoSegments = segments
		}
		if kind != "text" && kind != "subtitle" {
			continue
		}
		trackID := aiSegmentStringValue(track["id"])
		if trackID == "" {
			trackID = fmt.Sprintf("text-track-%d", trackIndex+1)
		}
		name := aiSegmentStringValue(track["name"])
		if name == "" {
			name = fmt.Sprintf("文本轨道 %d", len(trackOptions)+1)
		}
		for _, rawSegment := range segments {
			segment, _ := rawSegment.(map[string]any)
			start, end, ok := draftSegmentRange(segment)
			if !ok {
				continue
			}
			materialID := aiSegmentFirstString(segment, "material_id", "materialId")
			text := materials[materialID]
			if text == "" {
				text = aiSegmentFirstString(segment, "content", "text")
			}
			textByTrack[trackID] = append(textByTrack[trackID], timedText{start: start, end: end, text: strings.TrimSpace(text)})
		}
		trackOptions = append(trackOptions, aiSegmentTextTrack{ID: trackID, Name: name, Count: len(textByTrack[trackID])})
	}
	if len(videoSegments) == 0 {
		return nil, trackOptions, errors.New("主视频轨道没有有效片段")
	}
	selectedTexts := []timedText{}
	if selectedTextTrack != "" {
		selectedTexts = textByTrack[selectedTextTrack]
	} else {
		for _, option := range trackOptions {
			selectedTexts = append(selectedTexts, textByTrack[option.ID]...)
		}
	}
	clips := make([]aiSegmentClip, 0, len(videoSegments))
	for _, rawSegment := range videoSegments {
		segment, _ := rawSegment.(map[string]any)
		start, end, ok := draftSegmentRange(segment)
		if !ok {
			continue
		}
		texts := []string{}
		for _, item := range selectedTexts {
			if item.end > start && item.start < end && item.text != "" {
				texts = append(texts, item.text)
			}
		}
		clips = append(clips, aiSegmentClip{Index: len(clips) + 1, StartUS: start, EndUS: end, Subtitle: strings.Join(aiSegmentUniqueStrings(texts), " ")})
	}
	if len(clips) == 0 {
		return nil, trackOptions, errors.New("主视频轨道没有有效片段")
	}
	normalized, err := normalizeAISegmentClips(clips)
	return normalized, trackOptions, err
}

func draftTextMaterials(root map[string]any) map[string]string {
	result := map[string]string{}
	materials, _ := root["materials"].(map[string]any)
	for _, key := range []string{"texts", "text_templates"} {
		items, _ := materials[key].([]any)
		for _, raw := range items {
			item, _ := raw.(map[string]any)
			id := aiSegmentStringValue(item["id"])
			if id != "" {
				result[id] = aiSegmentFirstString(item, "content", "text")
			}
		}
	}
	return result
}

func draftSegmentRange(segment map[string]any) (int64, int64, bool) {
	var value map[string]any
	for _, key := range []string{"target_timerange", "targetTimerange", "timerange", "time_range"} {
		if item, ok := segment[key].(map[string]any); ok {
			value = item
			break
		}
	}
	if value == nil {
		return 0, 0, false
	}
	start := int64(numberValue(value["start"]))
	duration := int64(numberValue(value["duration"]))
	if duration <= 0 {
		return 0, 0, false
	}
	return start, start + duration, true
}

func normalizeAISegmentClips(clips []aiSegmentClip) ([]aiSegmentClip, error) {
	result := append([]aiSegmentClip(nil), clips...)
	sort.SliceStable(result, func(i, j int) bool { return result[i].StartUS < result[j].StartUS })
	for index := range result {
		if result[index].EndUS <= result[index].StartUS {
			return nil, errors.New("片段时间范围无效")
		}
		result[index].Index = index + 1
		if index > 0 && result[index].StartUS < result[index-1].EndUS {
			return nil, errors.New("主轨片段时间发生重叠")
		}
	}
	if len(result) == 0 {
		return nil, errors.New("主视频轨道没有有效片段")
	}
	return result, nil
}

func requestAISegmentSuggestions(ctx context.Context, username, modelID, script string, clips []aiSegmentClip, constraints aiSegmentConstraints, provider batchfactoryv11.DirectorProvider, execute AgentTextCompletionExecutor) ([]aiSegmentSuggestion, error) {
	request := batchfactoryv11.TextCompletionRequest{
		SystemPrompt: prompt.AISegmentSystemPrompt(),
		UserPrompt:   prompt.AISegmentUserPrompt(script, aiSegmentPromptClips(clips), constraints.FrontCount, constraints.FrontMin, constraints.FrontMax, constraints.BackMin, constraints.BackTarget, constraints.BackMax),
		MaxTokens:    6000,
		Temperature:  0.2,
	}
	for attempt := 0; attempt < 2; attempt++ {
		current := request
		if attempt > 0 {
			current.UserPrompt += "\n\n前一次结果无法解析。此次只返回一个 JSON 对象，不要附加说明。"
		}
		var raw string
		var err error
		if execute != nil {
			raw, err = execute(ctx, username, modelID, current)
		} else if provider != nil {
			raw, err = provider.Complete(ctx, current)
		} else {
			return nil, errors.New("文本模型执行器未配置")
		}
		if err != nil {
			return nil, errors.New("文本模型请求失败")
		}
		if suggestions, ok := parseAISegmentSuggestions(raw); ok {
			return suggestions, nil
		}
	}
	return nil, errors.New("模型没有返回有效切点")
}

func aiSegmentPromptClips(clips []aiSegmentClip) []prompt.AISegmentClip {
	result := make([]prompt.AISegmentClip, 0, len(clips))
	for _, clip := range clips {
		result = append(result, prompt.AISegmentClip{Index: clip.Index, StartSeconds: float64(clip.StartUS) / 1e6, EndSeconds: float64(clip.EndUS) / 1e6, Subtitle: clip.Subtitle})
	}
	return result
}

func parseAISegmentSuggestions(raw string) ([]aiSegmentSuggestion, bool) {
	raw = strings.TrimSpace(raw)
	start, end := strings.Index(raw, "{"), strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return nil, false
	}
	var value struct {
		Episodes []aiSegmentSuggestion `json:"episodes"`
	}
	if json.Unmarshal([]byte(raw[start:end+1]), &value) != nil || len(value.Episodes) == 0 {
		return nil, false
	}
	return value.Episodes, true
}

func planAISegmentEpisodes(clips []aiSegmentClip, suggested []aiSegmentSuggestion, constraints aiSegmentConstraints) ([]aiSegmentEpisode, []string) {
	constraints = normalizeAISegmentConstraints(constraints)
	if len(clips) == 0 {
		return nil, []string{"没有主轨道片段"}
	}
	reasons := map[int]string{}
	for _, item := range suggested {
		if item.EndClip > 0 && item.EndClip <= len(clips) {
			if _, exists := reasons[item.EndClip]; !exists {
				reasons[item.EndClip] = strings.TrimSpace(item.Reason)
			}
		}
	}
	episodes := []aiSegmentEpisode{}
	warnings := []string{}
	start := 0
	for start < len(clips) {
		episodeIndex := len(episodes) + 1
		isFront := episodeIndex <= constraints.FrontCount
		minSeconds, targetSeconds, maxSeconds := constraints.BackMin, constraints.BackTarget, constraints.BackMax
		if isFront {
			minSeconds, targetSeconds, maxSeconds = constraints.FrontMin, (constraints.FrontMin+constraints.FrontMax)/2, constraints.FrontMax
		}
		if clipDuration(clips[start], clips[start]) > maxSeconds {
			warnings = append(warnings, fmt.Sprintf("第 %d 个单个主轨片段时长超过 %.0f 秒，已整段保留", start+1, maxSeconds))
			episode := buildAISegmentEpisode(clips, start, start, reasons[start+1])
			episode.Index = len(episodes) + 1
			episodes = append(episodes, episode)
			start++
			continue
		}
		chosen := -1
		bestDistance := math.MaxFloat64
		for end := start; end < len(clips); end++ {
			duration := clipDuration(clips[start], clips[end])
			if duration > maxSeconds {
				break
			}
			if _, ok := reasons[end+1]; ok && duration >= minSeconds {
				distance := math.Abs(duration - targetSeconds)
				if distance < bestDistance {
					chosen, bestDistance = end, distance
				}
			}
		}
		if chosen < 0 {
			for end := start; end < len(clips); end++ {
				duration := clipDuration(clips[start], clips[end])
				if duration <= maxSeconds {
					distance := math.Abs(duration - targetSeconds)
					if distance < bestDistance {
						chosen, bestDistance = end, distance
					}
				}
			}
		}
		if chosen < start {
			chosen = start
		}
		if chosen < len(clips)-1 {
			remaining := clipDuration(clips[chosen+1], clips[len(clips)-1])
			if remaining < constraints.BackMin {
				chosen = len(clips) - 1
				warnings = append(warnings, "片尾不足最短时长，已并入上一集")
			}
		}
		episode := buildAISegmentEpisode(clips, start, chosen, reasons[chosen+1])
		episode.Index = len(episodes) + 1
		episodes = append(episodes, episode)
		start = chosen + 1
	}
	return episodes, aiSegmentUniqueStrings(warnings)
}

func buildAISegmentEpisode(clips []aiSegmentClip, start, end int, reason string) aiSegmentEpisode {
	return aiSegmentEpisode{
		StartClip: start + 1, EndClip: end + 1,
		StartSeconds: roundSeconds(float64(clips[start].StartUS) / 1e6), EndSeconds: roundSeconds(float64(clips[end].EndUS) / 1e6),
		DurationSeconds: roundSeconds(clipDuration(clips[start], clips[end])), Reason: reason,
	}
}

func clipDuration(start, end aiSegmentClip) float64 { return float64(end.EndUS-start.StartUS) / 1e6 }

func roundSeconds(value float64) float64 { return math.Round(value*1000) / 1000 }

func numberValue(value any) float64 {
	switch item := value.(type) {
	case float64:
		return item
	case json.Number:
		result, _ := item.Float64()
		return result
	case string:
		result, _ := strconv.ParseFloat(item, 64)
		return result
	default:
		return 0
	}
}

func aiSegmentStringValue(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func aiSegmentFirstString(value map[string]any, keys ...string) string {
	for _, key := range keys {
		if result := aiSegmentStringValue(value[key]); result != "" {
			return result
		}
	}
	return ""
}

func aiSegmentUniqueStrings(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}
