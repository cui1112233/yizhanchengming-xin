package aicut

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

type ASRSubmit struct {
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
	Store           Store
	OpenObject      ASRObjectOpenFunc
	ResolveProvider ASRProviderResolver
	Notifier        ASRNotifier
	Now             func() time.Time
}

type ASRService struct {
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
	return &ASRService{store: options.Store, openObject: options.OpenObject, resolveProvider: options.ResolveProvider, notifier: options.Notifier, now: options.Now, active: map[string]context.CancelFunc{}}
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
		if existing.AssetID == input.AssetID && (existing.State == ASRQueued || existing.State == ASRRunning || existing.State == ASRCancelRequested) {
			return ASRJob{}, errors.New("该素材正在识别字幕，请等待当前任务完成或先取消")
		}
	}
	now := s.now().UTC()
	job := ASRJob{ID: newASRID("aicut-asr"), Owner: owner, ProjectID: input.ProjectID, AssetID: input.AssetID, ModelID: input.ModelID, Language: strings.TrimSpace(input.Language), ExpectedRevision: input.ExpectedRevision, State: ASRQueued, Progress: 0, CreatedAt: now, UpdatedAt: now}
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
	current, err := s.store.Timeline(jobCtx, job.Owner, job.ProjectID)
	if err != nil {
		return fail(err)
	}
	if current.Revision != job.ExpectedRevision {
		return fail(ErrRevisionConflict)
	}
	prepared, trackID, err := ApplySRTCues(current, asset.Name, cues, newASRID)
	if err != nil {
		return fail(err)
	}
	if _, err := s.store.SaveTimeline(jobCtx, job.Owner, job.ProjectID, prepared, job.ExpectedRevision, s.now().UTC()); err != nil {
		return fail(err)
	}
	job = job.WithState(ASRSucceeded, 100, "", s.now())
	job.Imported, job.TrackID = len(cues), trackID
	return s.store.UpdateASRJob(jobCtx, job)
}
