package aicut

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
	"time"
)

type ASRSubmit struct {
	TimelineID       string
	ProjectID        string
	AssetID          string
	ModelID          string
	Language         string
	ExpectedRevision int64
}

type ASRProviderResolver func(context.Context, string, string) (ASRProvider, error)
type ASRObjectOpenFunc func(context.Context, Asset) (io.ReadCloser, error)

type ASRNotifier interface {
	Notify(context.Context, string) error
}

type ASRServiceOptions struct {
	Timelines       *TimelineCollectionStore
	Store           Store
	OpenObject      ASRObjectOpenFunc
	ResolveProvider ASRProviderResolver
	Notifier        ASRNotifier
	Now             func() time.Time
}

type ASRService struct {
	timelines       *TimelineCollectionStore
	store           Store
	openObject      ASRObjectOpenFunc
	resolveProvider ASRProviderResolver
	notifier        ASRNotifier
	now             func() time.Time
	mu              sync.Mutex
	active          map[string]context.CancelFunc
}

func NewASRService(options ASRServiceOptions) *ASRService {
	if options.Now == nil {
		options.Now = time.Now
	}
	return &ASRService{store: options.Store, timelines: options.Timelines, openObject: options.OpenObject, resolveProvider: options.ResolveProvider, notifier: options.Notifier, now: options.Now, active: map[string]context.CancelFunc{}}
}

func newASRID(prefix string) string {
	raw := make([]byte, 10)
	if _, err := rand.Read(raw); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return prefix + "-" + hex.EncodeToString(raw)
}

func (s *ASRService) Submit(ctx context.Context, owner string, input ASRSubmit) (ASRJob, error) {
	if s == nil || s.store == nil {
		return ASRJob{}, errors.New("ASR 服务尚未配置")
	}
	owner, input.ProjectID, input.AssetID, input.ModelID = strings.TrimSpace(owner), strings.TrimSpace(input.ProjectID), strings.TrimSpace(input.AssetID), strings.TrimSpace(input.ModelID)
	if owner == "" || input.ProjectID == "" || input.AssetID == "" || input.ModelID == "" || input.ExpectedRevision < 0 {
		return ASRJob{}, errors.New("请选择语音素材和 ASR 模型")
	}
	input.TimelineID = strings.TrimSpace(input.TimelineID)
	if input.TimelineID == DefaultTimelineID {
		input.TimelineID = ""
	}
	if input.TimelineID != "" {
		v, err := s.timeline(ctx, owner, input.ProjectID, input.TimelineID)
		if err != nil {
			return ASRJob{}, err
		}
		if v.Revision != input.ExpectedRevision {
			return ASRJob{}, ErrRevisionConflict
		}
	}
	asset, err := s.store.Asset(ctx, owner, input.ProjectID, input.AssetID)
	if err != nil {
		return ASRJob{}, err
	}
	if asset.Kind != "audio" && asset.Kind != "video" {
		return ASRJob{}, errors.New("语音转字幕仅支持音频或视频素材")
	}
	jobs, err := s.store.ListASRJobs(ctx, owner, input.ProjectID)
	if err != nil {
		return ASRJob{}, err
	}
	for _, existing := range jobs {
		if existing.AssetID == input.AssetID && canonicalASRTimeline(existing.TimelineID) == canonicalASRTimeline(input.TimelineID) && (existing.State == ASRQueued || existing.State == ASRRunning || existing.State == ASRCancelRequested) {
			return ASRJob{}, errors.New("该素材正在识别字幕，请等待当前任务完成或先取消")
		}
	}
	now := s.now().UTC()
	job := ASRJob{ID: newASRID("aicut-asr"), Owner: owner, ProjectID: input.ProjectID, TimelineID: input.TimelineID, AssetID: input.AssetID, ModelID: input.ModelID, Language: strings.TrimSpace(input.Language), ExpectedRevision: input.ExpectedRevision, State: ASRQueued, Progress: 0, CreatedAt: now, UpdatedAt: now}
	if err := s.store.CreateASRJob(ctx, owner, job); err != nil {
		return ASRJob{}, err
	}
	if s.notifier != nil {
		if err := s.notifier.Notify(ctx, job.ID); err != nil {
			job = job.WithState(ASRFailed, 100, "ASR 队列不可用，请稍后重试", s.now())
			_ = s.store.UpdateASRJob(ctx, job)
			return job, errors.New(job.Error)
		}
	}
	return job, nil
}

func (s *ASRService) Cancel(ctx context.Context, owner, projectID, jobID string) (ASRJob, error) {
	return s.CancelForTimeline(ctx, owner, projectID, jobID, DefaultTimelineID)
}
func (s *ASRService) CancelForTimeline(ctx context.Context, owner, projectID, jobID, timelineID string) (ASRJob, error) {
	if _, err := s.JobForTimeline(ctx, owner, projectID, jobID, timelineID); err != nil {
		return ASRJob{}, err
	}
	job, err := s.store.RequestASRCancel(ctx, owner, projectID, jobID, s.now().UTC())
	if err != nil {
		return ASRJob{}, err
	}
	s.mu.Lock()
	cancel := s.active[jobID]
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	return job, nil
}

func (s *ASRService) ProcessNext(ctx context.Context) error {
	job, err := s.store.ClaimNextASRJob(ctx, s.now().UTC())
	if err != nil {
		return err
	}
	jobCtx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	s.active[job.ID] = cancel
	s.mu.Unlock()
	defer func() { cancel(); s.mu.Lock(); delete(s.active, job.ID); s.mu.Unlock() }()
	fail := func(cause error) error {
		state := ASRFailed
		message := cause.Error()
		if errors.Is(cause, context.Canceled) {
			state, message = ASRCanceled, ""
		}
		job = job.WithState(state, 100, message, s.now())
		_ = s.store.UpdateASRJob(context.Background(), job)
		if state == ASRCanceled {
			return nil
		}
		return cause
	}
	asset, err := s.store.Asset(jobCtx, job.Owner, job.ProjectID, job.AssetID)
	if err != nil {
		return fail(fmt.Errorf("读取语音素材记录失败：%w", err))
	}
	if s.openObject == nil {
		return fail(errors.New("素材存储尚未配置"))
	}
	body, err := s.openObject(jobCtx, asset)
	if err != nil {
		return fail(fmt.Errorf("读取语音素材失败：%w", err))
	}
	defer body.Close()
	if s.resolveProvider == nil {
		return fail(errors.New("ASR 模型未配置，请先到 API 配置启用语音识别模型"))
	}
	provider, err := s.resolveProvider(jobCtx, job.Owner, job.ModelID)
	if err != nil || provider == nil {
		if err == nil {
			err = errors.New("模型不存在")
		}
		return fail(fmt.Errorf("ASR 模型不可用：%w", err))
	}
	job = job.WithState(ASRRunning, 20, "", s.now())
	_ = s.store.UpdateASRJob(jobCtx, job)
	cues, err := provider.Transcribe(jobCtx, ASRInput{Filename: asset.Name, ContentType: asset.ContentType, Language: job.Language, Body: body})
	if err != nil {
		return fail(err)
	}
	latest, latestErr := s.store.ASRJob(context.Background(), job.Owner, job.ProjectID, job.ID)
	if errors.Is(jobCtx.Err(), context.Canceled) || (latestErr == nil && (latest.State == ASRCancelRequested || latest.State == ASRCanceled)) {
		return fail(context.Canceled)
	}
	current, err := s.timeline(jobCtx, job.Owner, job.ProjectID, job.TimelineID)
	if err != nil {
		return fail(err)
	}
	if current.Revision != job.ExpectedRevision {
		return fail(ErrRevisionConflict)
	}
	mapped, err := mapASRCuesToTimeline(current, asset.ID, cues)
	if err != nil {
		return fail(err)
	}
	prepared, trackID, err := ApplySRTCues(current, asset.Name, mapped, newASRID)
	if err != nil {
		return fail(err)
	}
	if _, err := s.saveTimeline(jobCtx, job.Owner, job.ProjectID, job.TimelineID, prepared, job.ExpectedRevision, s.now().UTC()); err != nil {
		return fail(err)
	}
	job = job.WithState(ASRSucceeded, 100, "", s.now())
	job.Imported, job.TrackID = len(mapped), trackID
	return s.store.UpdateASRJob(jobCtx, job)
}

func canonicalASRTimeline(id string) string {
	if strings.TrimSpace(id) == "" {
		return DefaultTimelineID
	}
	return strings.TrimSpace(id)
}
func (s *ASRService) timeline(ctx context.Context, o, p, id string) (Timeline, error) {
	if canonicalASRTimeline(id) == DefaultTimelineID {
		return s.store.Timeline(ctx, o, p)
	}
	if s.timelines == nil {
		return Timeline{}, ErrTimelineJobUnsupported
	}
	return s.timelines.Timeline(ctx, o, p, id)
}
func (s *ASRService) saveTimeline(ctx context.Context, o, p, id string, v Timeline, rev int64, now time.Time) (Timeline, error) {
	if canonicalASRTimeline(id) == DefaultTimelineID {
		return s.store.SaveTimeline(ctx, o, p, v, rev, now)
	}
	if s.timelines == nil {
		return Timeline{}, ErrTimelineJobUnsupported
	}
	return s.timelines.Save(ctx, o, p, id, v, rev, now)
}
func (s *ASRService) JobForTimeline(ctx context.Context, o, p, jobID, id string) (ASRJob, error) {
	job, e := s.store.ASRJob(ctx, o, p, jobID)
	if e != nil {
		return ASRJob{}, e
	}
	if canonicalASRTimeline(job.TimelineID) != canonicalASRTimeline(id) {
		return ASRJob{}, ErrASRJobNotFound
	}
	return job, nil
}
func (s *ASRService) ListForTimeline(ctx context.Context, o, p, id string) ([]ASRJob, error) {
	jobs, e := s.store.ListASRJobs(ctx, o, p)
	if e != nil {
		return nil, e
	}
	out := make([]ASRJob, 0, len(jobs))
	for _, job := range jobs {
		if canonicalASRTimeline(job.TimelineID) == canonicalASRTimeline(id) {
			out = append(out, job)
		}
	}
	return out, nil
}

// ASR times refer to the original media. Each placed instance needs its own
// clipped and speed-adjusted subtitles. Unplaced media retains source-time import.
func mapASRCuesToTimeline(timeline Timeline, assetID string, cues []SRTCue) ([]SRTCue, error) {
	mapped := []SRTCue{}
	matched := false
	for _, track := range timeline.Tracks {
		for _, clip := range track.Clips {
			if clip.AssetID != assetID || (clip.Kind != ClipAudio && clip.Kind != ClipVideo) {
				continue
			}
			matched = true
			if len(clip.SpeedCurve) > 0 || clip.Reversed || clip.FreezeSourceMS != nil || clip.CompoundID != "" {
				return nil, fmt.Errorf("片段 %s 含曲线变速、倒放、定格或复合片段，暂不支持自动对齐字幕，请先处理后重试", clip.Name)
			}
			rate := clip.PlaybackRate
			if rate <= 0 {
				rate = 1
			}
			sourceStart := float64(clip.SourceStartMS)
			sourceEnd := sourceStart + float64(clip.DurationMS)*rate
			for _, cue := range cues {
				start := math.Max(float64(cue.StartMS), sourceStart)
				end := math.Min(float64(cue.EndMS), sourceEnd)
				if end <= start {
					continue
				}
				a := clip.StartMS + int64(math.Round((start-sourceStart)/rate))
				b := clip.StartMS + int64(math.Round((end-sourceStart)/rate))
				if b > a {
					mapped = append(mapped, SRTCue{StartMS: a, EndMS: b, Text: cue.Text})
				}
			}
		}
	}
	if !matched {
		return cues, nil
	}
	if len(mapped) == 0 {
		return nil, errors.New("识别字幕不在该素材的可见剪辑范围内，未写入时间线")
	}
	return mapped, nil
}
