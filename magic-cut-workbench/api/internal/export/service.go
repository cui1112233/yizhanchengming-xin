package export

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"yizhanchengming/magic-cut-workbench/internal/material"
	worktask "yizhanchengming/magic-cut-workbench/internal/task"
	"yizhanchengming/magic-cut-workbench/internal/timeline"
)

type CreateInput struct {
	ProjectID  string `json:"projectId"`
	OutputName string `json:"outputName"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	FrameRate  int    `json:"frameRate"`
	VideoCodec string `json:"videoCodec"`
	AudioCodec string `json:"audioCodec"`
	Container  string `json:"container"`
}

type PlanInput struct {
	MaterialID    string
	Path          string
	SourceStartMs int64
	DurationMs    int64
	Speed         float64
}

type RenderPlan struct {
	Command    string
	Args       []string
	Inputs     []PlanInput
	OutputPath string
	DurationMs int64
}

type Result struct {
	OutputPath string `json:"outputPath"`
	SizeBytes  int64  `json:"sizeBytes"`
	SHA256     string `json:"sha256"`
}

type Export struct {
	ID         string            `json:"id"`
	ProjectID  string            `json:"projectId"`
	TaskID     string            `json:"taskId"`
	Owner      string            `json:"-"`
	OutputName string            `json:"outputName"`
	StorageKey string            `json:"-"`
	Metadata   material.Metadata `json:"metadata"`
	SHA256     string            `json:"sha256,omitempty"`
	CreatedAt  time.Time         `json:"createdAt"`
}

type TimelineSource interface {
	Get(context.Context, string, string) (timeline.Timeline, error)
}

type MaterialSource interface {
	List(context.Context, string, string) ([]material.Material, error)
}

type Repository interface {
	Create(context.Context, Export, worktask.Task) error
	Start(context.Context, string, time.Time) (int, error)
	Progress(context.Context, string, float64, time.Time) error
	Succeed(context.Context, string, string, Result, material.Metadata, time.Time) error
	Fail(context.Context, string, worktask.Status, string, string, time.Time) error
	TaskStatus(context.Context, string, string) (worktask.Status, error)
	Get(context.Context, string, string) (Export, error)
}

type Renderer interface {
	Run(context.Context, RenderPlan, func(float64)) (Result, error)
}

type ServiceOptions struct {
	StorageRoot string
	Timelines   TimelineSource
	Materials   MaterialSource
	Repository  Repository
	Renderer    Renderer
	Probe       material.Probe
	NewExportID func() string
	NewTaskID   func() string
	Now         func() time.Time
	Start       func(func())
}

type Service struct {
	options ServiceOptions
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
	jobs    map[string]exportJob
}

type exportJob struct {
	owner    string
	exportID string
	plan     RenderPlan
}

func NewService(options ServiceOptions) *Service {
	if options.Now == nil {
		options.Now = time.Now
	}
	if options.Start == nil {
		options.Start = func(work func()) { go work() }
	}
	return &Service{options: options, cancels: map[string]context.CancelFunc{}, jobs: map[string]exportJob{}}
}

func (s *Service) Create(ctx context.Context, owner string, input CreateInput) (Export, worktask.Task, error) {
	owner = strings.TrimSpace(owner)
	input.ProjectID = strings.TrimSpace(input.ProjectID)
	input.OutputName = strings.TrimSpace(input.OutputName)
	if input.OutputName == "" || input.OutputName == "." || input.OutputName == ".." || filepath.Base(input.OutputName) != input.OutputName || strings.ContainsAny(input.OutputName, `/\\`) {
		return Export{}, worktask.Task{}, &ServiceError{Code: "EXPORT_NAME_INVALID", Err: errors.New("输出名称不能包含路径")}
	}
	input.OutputName = strings.TrimSuffix(input.OutputName, filepath.Ext(input.OutputName))
	if owner == "" || input.ProjectID == "" {
		return Export{}, worktask.Task{}, &ServiceError{Code: "EXPORT_INPUT_INVALID", Err: errors.New("账号和工程不能为空")}
	}
	if s == nil || strings.TrimSpace(s.options.StorageRoot) == "" || s.options.Timelines == nil || s.options.Materials == nil || s.options.Repository == nil || s.options.Renderer == nil || s.options.Probe == nil || s.options.NewExportID == nil || s.options.NewTaskID == nil {
		return Export{}, worktask.Task{}, &ServiceError{Code: "EXPORT_SERVICE_UNAVAILABLE", Err: errors.New("导出服务尚未连接")}
	}
	value, err := s.options.Timelines.Get(ctx, owner, input.ProjectID)
	if err != nil {
		return Export{}, worktask.Task{}, err
	}
	items, err := s.options.Materials.List(ctx, owner, input.ProjectID)
	if err != nil {
		return Export{}, worktask.Task{}, err
	}
	exportID, taskID := s.options.NewExportID(), s.options.NewTaskID()
	outputPath := filepath.Join(s.options.StorageRoot, "exports", safeSegment(owner), safeSegment(input.ProjectID), safeSegment(exportID), input.OutputName+".mp4")
	if !insideRoot(s.options.StorageRoot, outputPath) {
		return Export{}, worktask.Task{}, &ServiceError{Code: "EXPORT_PATH_INVALID", Err: errors.New("导出路径超出存储目录")}
	}
	plan, err := BuildRenderPlan(value, items, input, outputPath)
	if err != nil {
		return Export{}, worktask.Task{}, err
	}
	now := s.options.Now().UTC()
	created := Export{ID: exportID, ProjectID: input.ProjectID, TaskID: taskID, Owner: owner, OutputName: input.OutputName, StorageKey: outputPath, CreatedAt: now}
	queued := worktask.Task{ID: taskID, ProjectID: input.ProjectID, Owner: owner, Type: "export_video", Status: worktask.StatusPending, Progress: 0, Input: map[string]any{"exportId": exportID, "outputName": input.OutputName}, CreatedAt: now, UpdatedAt: now}
	if err := s.options.Repository.Create(ctx, created, queued); err != nil {
		return Export{}, worktask.Task{}, err
	}
	s.mu.Lock()
	s.jobs[taskID] = exportJob{owner: owner, exportID: exportID, plan: plan}
	s.mu.Unlock()
	s.options.Start(func() { s.run(exportID, taskID, plan) })
	return created, queued, nil
}

func (s *Service) run(exportID, taskID string, plan RenderPlan) {
	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	s.cancels[taskID] = cancel
	s.mu.Unlock()
	defer func() {
		cancel()
		s.mu.Lock()
		delete(s.cancels, taskID)
		s.mu.Unlock()
	}()
	now := s.options.Now().UTC()
	if _, err := s.options.Repository.Start(ctx, taskID, now); err != nil {
		return
	}
	result, err := s.options.Renderer.Run(ctx, plan, func(progress float64) {
		_ = s.options.Repository.Progress(context.Background(), taskID, progress, s.options.Now().UTC())
	})
	if err != nil {
		status := worktask.StatusFailed
		code := ErrorCode(err)
		if code == "EXPORT_CANCELED" || errors.Is(ctx.Err(), context.Canceled) {
			status = worktask.StatusCanceled
			code = "EXPORT_CANCELED"
		}
		_ = s.options.Repository.Fail(context.Background(), taskID, status, code, err.Error(), s.options.Now().UTC())
		return
	}
	metadata, err := s.options.Probe.Probe(context.Background(), result.OutputPath)
	if err != nil {
		_ = s.options.Repository.Fail(context.Background(), taskID, worktask.StatusFailed, "EXPORT_PROBE_FAILED", "导出文件校验失败", s.options.Now().UTC())
		return
	}
	_ = s.options.Repository.Succeed(context.Background(), exportID, taskID, result, metadata, s.options.Now().UTC())
}

func (s *Service) Cancel(taskID string) bool {
	s.mu.Lock()
	cancel := s.cancels[taskID]
	s.mu.Unlock()
	if cancel == nil {
		return false
	}
	cancel()
	return true
}

func (s *Service) CancelTask(_ context.Context, owner, taskID string) (bool, error) {
	s.mu.Lock()
	job, exists := s.jobs[taskID]
	cancel := s.cancels[taskID]
	s.mu.Unlock()
	if !exists || job.owner != owner {
		return false, nil
	}
	if cancel == nil {
		return false, nil
	}
	cancel()
	return true, nil
}

func (s *Service) RetryTask(ctx context.Context, owner, taskID string) (bool, error) {
	s.mu.Lock()
	job, exists := s.jobs[taskID]
	_, running := s.cancels[taskID]
	s.mu.Unlock()
	if !exists || job.owner != owner {
		return false, nil
	}
	if running {
		return true, &ServiceError{Code: "EXPORT_ALREADY_RUNNING", Err: errors.New("导出任务正在执行")}
	}
	status, err := s.options.Repository.TaskStatus(ctx, owner, taskID)
	if err != nil {
		return true, err
	}
	if status != worktask.StatusFailed && status != worktask.StatusCanceled {
		return true, &ServiceError{Code: "EXPORT_NOT_RETRYABLE", Err: errors.New("只有失败或已取消的导出可以重试")}
	}
	s.options.Start(func() { s.run(job.exportID, taskID, job.plan) })
	return true, nil
}

func (s *Service) File(ctx context.Context, owner, exportID string) (Export, error) {
	return s.options.Repository.Get(ctx, owner, exportID)
}

type ServiceError struct {
	Code string
	Err  error
}

func (e *ServiceError) Error() string {
	if e.Err == nil {
		return e.Code
	}
	return e.Err.Error()
}

func (e *ServiceError) Unwrap() error { return e.Err }

func ErrorCode(err error) string {
	var typed *ServiceError
	if errors.As(err, &typed) {
		return typed.Code
	}
	return ""
}

func BuildRenderPlan(value timeline.Timeline, materials []material.Material, input CreateInput, outputPath string) (RenderPlan, error) {
	if input.Container != "mp4" || input.VideoCodec != "h264" || input.AudioCodec != "aac" {
		return RenderPlan{}, &ServiceError{Code: "EXPORT_FORMAT_NOT_SUPPORTED", Err: errors.New("当前仅支持 H.264 + AAC 的 MP4 导出")}
	}
	if input.Width < 16 || input.Height < 16 || input.Width > 7680 || input.Height > 4320 || input.FrameRate < 1 || input.FrameRate > 120 {
		return RenderPlan{}, &ServiceError{Code: "EXPORT_PARAMS_INVALID", Err: errors.New("导出分辨率或帧率无效")}
	}
	if strings.ToLower(filepath.Ext(outputPath)) != ".mp4" {
		return RenderPlan{}, &ServiceError{Code: "EXPORT_PATH_INVALID", Err: errors.New("导出路径必须是 MP4 文件")}
	}
	byID := make(map[string]material.Material, len(materials))
	for _, item := range materials {
		byID[item.ID] = item
	}
	clips := make([]timeline.Clip, 0)
	for _, track := range value.Tracks {
		if track.Kind == "video" {
			clips = append(clips, track.Clips...)
		}
	}
	sort.SliceStable(clips, func(i, j int) bool { return clips[i].StartMs < clips[j].StartMs })
	if len(clips) == 0 {
		return RenderPlan{}, &ServiceError{Code: "EXPORT_TIMELINE_EMPTY", Err: errors.New("时间线中没有可导出的视频")}
	}
	plan := RenderPlan{Command: "ffmpeg", OutputPath: outputPath}
	args := []string{"-hide_banner", "-loglevel", "error", "-nostats"}
	filters := make([]string, 0, len(clips)+2)
	labels := make([]string, 0, len(clips))
	audioSegmentLabels := make([]string, 0, len(clips))
	for index, clip := range clips {
		item, ok := byID[clip.MaterialID]
		if !ok || item.MediaType != "video" || strings.TrimSpace(item.StorageKey) == "" {
			return RenderPlan{}, &ServiceError{Code: "EXPORT_MATERIAL_MISSING", Err: fmt.Errorf("时间线素材不可用：%s", clip.MaterialID)}
		}
		if clip.DurationMs <= 0 || clip.SourceStartMs < 0 || clip.Speed <= 0 {
			return RenderPlan{}, &ServiceError{Code: "EXPORT_TIMELINE_INVALID", Err: fmt.Errorf("片段参数无效：%s", clip.ID)}
		}
		sourceDuration := float64(clip.DurationMs) * clip.Speed / 1000
		args = append(args, "-ss", seconds(float64(clip.SourceStartMs)/1000), "-t", seconds(sourceDuration), "-i", item.StorageKey)
		label := fmt.Sprintf("v%d", index)
		filters = append(filters, fmt.Sprintf("[%d:v:0]setpts=(PTS-STARTPTS)/%s,scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,fps=%d,setsar=1,format=yuv420p[%s]", index, number(clip.Speed), input.Width, input.Height, input.Width, input.Height, input.FrameRate, label))
		labels = append(labels, "["+label+"]")
		audioLabel := fmt.Sprintf("va%d", index)
		if item.Metadata.AudioCodec == "" {
			filters = append(filters, fmt.Sprintf("anullsrc=channel_layout=stereo:sample_rate=48000,atrim=duration=%s,asetpts=PTS-STARTPTS[%s]", seconds(float64(clip.DurationMs)/1000), audioLabel))
		} else {
			chain := []string{
				fmt.Sprintf("[%d:a:0]atrim=duration=%s", index, seconds(sourceDuration)),
				"asetpts=PTS-STARTPTS",
				"aresample=48000",
			}
			chain = append(chain, atempoFilters(clip.Speed)...)
			chain = append(chain, "volume="+number(clipVolume(clip, 1)))
			filters = append(filters, strings.Join(chain, ",")+"["+audioLabel+"]")
		}
		audioSegmentLabels = append(audioSegmentLabels, "["+audioLabel+"]")
		plan.Inputs = append(plan.Inputs, PlanInput{MaterialID: clip.MaterialID, Path: item.StorageKey, SourceStartMs: clip.SourceStartMs, DurationMs: clip.DurationMs, Speed: clip.Speed})
		plan.DurationMs += clip.DurationMs
	}
	if len(labels) == 1 {
		filters = append(filters, labels[0]+"null[vout]")
	} else {
		filters = append(filters, strings.Join(labels, "")+fmt.Sprintf("concat=n=%d:v=1:a=0[vout]", len(labels)))
	}
	if len(audioSegmentLabels) == 1 {
		filters = append(filters, audioSegmentLabels[0]+"anull[baseaudio]")
	} else {
		filters = append(filters, strings.Join(audioSegmentLabels, "")+fmt.Sprintf("concat=n=%d:v=0:a=1[baseaudio]", len(audioSegmentLabels)))
	}
	audioClips := make([]timeline.Clip, 0)
	for _, track := range value.Tracks {
		if track.Kind == "audio" {
			audioClips = append(audioClips, track.Clips...)
		}
	}
	sort.SliceStable(audioClips, func(i, j int) bool { return audioClips[i].StartMs < audioClips[j].StartMs })
	overlayLabels := make([]string, 0, len(audioClips))
	for _, clip := range audioClips {
		item, ok := byID[clip.MaterialID]
		if !ok || item.MediaType != "audio" || strings.TrimSpace(item.StorageKey) == "" {
			return RenderPlan{}, &ServiceError{Code: "EXPORT_MATERIAL_MISSING", Err: fmt.Errorf("时间线素材不可用：%s", clip.MaterialID)}
		}
		if clip.DurationMs <= 0 || clip.SourceStartMs < 0 || clip.StartMs < 0 || clip.Speed <= 0 {
			return RenderPlan{}, &ServiceError{Code: "EXPORT_TIMELINE_INVALID", Err: fmt.Errorf("片段参数无效：%s", clip.ID)}
		}
		inputIndex := len(plan.Inputs)
		sourceDuration := float64(clip.DurationMs) * clip.Speed / 1000
		args = append(args, "-ss", seconds(float64(clip.SourceStartMs)/1000), "-t", seconds(sourceDuration), "-i", item.StorageKey)
		label := fmt.Sprintf("audio%d", len(overlayLabels))
		chain := []string{
			fmt.Sprintf("[%d:a:0]atrim=duration=%s", inputIndex, seconds(sourceDuration)),
			"asetpts=PTS-STARTPTS",
			"aresample=48000",
		}
		chain = append(chain, atempoFilters(clip.Speed)...)
		chain = append(chain,
			"volume="+number(clipVolume(clip, .7)),
			fmt.Sprintf("adelay=%d|%d", clip.StartMs, clip.StartMs),
			"apad",
			"atrim=duration="+seconds(float64(plan.DurationMs)/1000),
		)
		filters = append(filters, strings.Join(chain, ",")+"["+label+"]")
		overlayLabels = append(overlayLabels, "["+label+"]")
		plan.Inputs = append(plan.Inputs, PlanInput{MaterialID: clip.MaterialID, Path: item.StorageKey, SourceStartMs: clip.SourceStartMs, DurationMs: clip.DurationMs, Speed: clip.Speed})
	}
	if len(overlayLabels) == 0 {
		filters = append(filters, "[baseaudio]anull[aout]")
	} else {
		filters = append(filters, "[baseaudio]"+strings.Join(overlayLabels, "")+fmt.Sprintf("amix=inputs=%d:duration=first:dropout_transition=0:normalize=0,atrim=duration=%s[aout]", len(overlayLabels)+1, seconds(float64(plan.DurationMs)/1000)))
	}
	args = append(args,
		"-filter_complex", strings.Join(filters, ";"),
		"-map", "[vout]", "-map", "[aout]",
		"-c:v", "libx264", "-c:a", "aac", "-r", strconv.Itoa(input.FrameRate),
		"-movflags", "+faststart", "-shortest", "-progress", "pipe:1", "-y", outputPath,
	)
	plan.Args = args
	return plan, nil
}

func seconds(value float64) string { return strconv.FormatFloat(value, 'f', 3, 64) }
func number(value float64) string  { return strconv.FormatFloat(value, 'f', -1, 64) }

func clipVolume(clip timeline.Clip, fallback float64) float64 {
	if clip.Volume == nil {
		return fallback
	}
	return *clip.Volume
}

func atempoFilters(speed float64) []string {
	if speed == 1 {
		return nil
	}
	filters := make([]string, 0, 3)
	for speed > 2 {
		filters = append(filters, "atempo=2")
		speed /= 2
	}
	for speed < .5 {
		filters = append(filters, "atempo=0.5")
		speed /= .5
	}
	if speed != 1 {
		filters = append(filters, "atempo="+number(speed))
	}
	return filters
}

type FFmpeg struct {
	Path string
}

func (runner FFmpeg) Run(ctx context.Context, plan RenderPlan, onProgress func(float64)) (Result, error) {
	command := runner.Path
	if command == "" {
		command = plan.Command
	}
	if command == "" || len(plan.Args) == 0 || strings.TrimSpace(plan.OutputPath) == "" || plan.DurationMs <= 0 {
		return Result{}, &ServiceError{Code: "EXPORT_PLAN_INVALID", Err: errors.New("导出计划不完整")}
	}
	if err := os.MkdirAll(filepath.Dir(plan.OutputPath), 0o755); err != nil {
		return Result{}, &ServiceError{Code: "EXPORT_OUTPUT_UNAVAILABLE", Err: err}
	}
	process := exec.CommandContext(ctx, command, plan.Args...)
	stdout, err := process.StdoutPipe()
	if err != nil {
		return Result{}, err
	}
	var stderr bytes.Buffer
	process.Stderr = &stderr
	if err := process.Start(); err != nil {
		return Result{}, &ServiceError{Code: "EXPORT_FFMPEG_UNAVAILABLE", Err: err}
	}
	readProgress(stdout, plan.DurationMs, onProgress)
	if err := process.Wait(); err != nil {
		_ = os.Remove(plan.OutputPath)
		if errors.Is(ctx.Err(), context.Canceled) {
			return Result{}, &ServiceError{Code: "EXPORT_CANCELED", Err: errors.New("导出已取消")}
		}
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return Result{}, &ServiceError{Code: "EXPORT_FFMPEG_FAILED", Err: errors.New(message)}
	}
	file, err := os.Open(plan.OutputPath)
	if err != nil {
		return Result{}, err
	}
	defer file.Close()
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return Result{}, err
	}
	if onProgress != nil {
		onProgress(1)
	}
	return Result{OutputPath: plan.OutputPath, SizeBytes: size, SHA256: fmt.Sprintf("%x", hash.Sum(nil))}, nil
}

func readProgress(reader io.Reader, durationMs int64, onProgress func(float64)) {
	if onProgress == nil || durationMs <= 0 {
		_, _ = io.Copy(io.Discard, reader)
		return
	}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "out_time_us=") || strings.HasPrefix(line, "out_time_ms=") {
			parts := strings.SplitN(line, "=", 2)
			micros, _ := strconv.ParseInt(parts[1], 10, 64)
			progress := float64(micros) / float64(durationMs*1000)
			if progress < 0 {
				progress = 0
			}
			if progress > 0.99 {
				progress = 0.99
			}
			onProgress(progress)
		}
	}
}

func safeSegment(value string) string {
	value = strings.TrimSpace(value)
	value = strings.NewReplacer("/", "_", "\\", "_", ":", "_", "..", "_").Replace(value)
	value = strings.Trim(value, ". ")
	if value == "" {
		return "unknown"
	}
	return value
}

func insideRoot(root, candidate string) bool {
	rootPath, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	candidatePath, err := filepath.Abs(candidate)
	if err != nil {
		return false
	}
	relative, err := filepath.Rel(rootPath, candidatePath)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
