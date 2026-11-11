package httpapi

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
)

type AICutExportService struct {
	store    aicut.Store
	objects  agentworkspace.AssetObjectStore
	notifier aicut.ExportNotifier
	now      func() time.Time
	mu       sync.Mutex
	active   map[string]context.CancelFunc
}

func NewAICutExportService(store aicut.Store, objects agentworkspace.AssetObjectStore, notifier aicut.ExportNotifier, now func() time.Time) *AICutExportService {
	if now == nil {
		now = time.Now
	}
	return &AICutExportService{store: store, objects: objects, notifier: notifier, now: now, active: map[string]context.CancelFunc{}}
}

func (s *AICutExportService) Submit(ctx context.Context, owner string, project aicut.Project, timeline aicut.Timeline, format string, options aicut.ExportOptions) (aicut.ExportJob, error) {
	if s == nil || s.store == nil || s.objects == nil {
		return aicut.ExportJob{}, errors.New("导出服务尚未配置")
	}
	owner, format = strings.TrimSpace(owner), strings.ToLower(strings.TrimSpace(format))
	// Background color belongs to the immutable project snapshot. Keeping it out
	// of the serialized option set makes duplicate detection stable after a
	// process restart while still rendering the exact submitted project state.
	options.BackgroundColor = ""
	if owner == "" || project.ID == "" || timeline.Revision < 0 {
		return aicut.ExportJob{}, errors.New("导出项目无效")
	}
	jobs, err := s.store.ListExportJobs(ctx, owner, project.ID)
	if err != nil {
		return aicut.ExportJob{}, err
	}
	for _, existing := range jobs {
		if existing.ExpectedRevision == timeline.Revision && existing.Format == format && reflect.DeepEqual(existing.Options, options) && (existing.State == aicut.ExportQueued || existing.State == aicut.ExportRunning || existing.State == aicut.ExportCancelRequested) {
			return aicut.ExportJob{}, errors.New("相同导出任务正在排队或渲染，请勿重复提交")
		}
	}
	now := s.now().UTC()
	job := aicut.ExportJob{ID: newAgentID("export"), Owner: owner, ProjectID: project.ID, Format: format, ExpectedRevision: timeline.Revision, State: aicut.ExportQueued, Options: options, ProjectSnapshot: project, TimelineSnapshot: timeline, CreatedAt: now, UpdatedAt: now}
	if err := s.store.CreateExportJob(ctx, owner, job); err != nil {
		return aicut.ExportJob{}, err
	}
	if s.notifier == nil {
		job = job.WithState(aicut.ExportFailed, 100, "Redis 导出队列未配置", "", s.now())
		_ = s.store.UpdateExportJob(ctx, job)
		return job, errors.New(job.Error)
	}
	if err := s.notifier.Notify(ctx, job.ID); err != nil {
		job = job.WithState(aicut.ExportFailed, 100, "Redis 导出队列不可用，请稍后重试", "", s.now())
		_ = s.store.UpdateExportJob(ctx, job)
		return job, errors.New(job.Error)
	}
	return job, nil
}

func (s *AICutExportService) Job(ctx context.Context, owner, projectID, jobID string) (aicut.ExportJob, error) {
	return s.store.ExportJob(ctx, owner, projectID, jobID)
}

func (s *AICutExportService) List(ctx context.Context, owner, projectID string) ([]aicut.ExportJob, error) {
	return s.store.ListExportJobs(ctx, owner, projectID)
}

func (s *AICutExportService) Cancel(ctx context.Context, owner, projectID, jobID string) (aicut.ExportJob, error) {
	job, err := s.store.RequestExportCancel(ctx, owner, projectID, jobID, s.now().UTC())
	if err != nil {
		return aicut.ExportJob{}, err
	}
	s.mu.Lock()
	cancel := s.active[jobID]
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	return job, nil
}

func (s *AICutExportService) ProcessNext(ctx context.Context) error {
	job, err := s.store.ClaimNextExportJob(ctx, s.now().UTC())
	if err != nil {
		return err
	}
	jobCtx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	s.active[job.ID] = cancel
	s.mu.Unlock()
	defer func() {
		cancel()
		s.mu.Lock()
		delete(s.active, job.ID)
		s.mu.Unlock()
	}()
	fail := func(cause error) error {
		state, message := aicut.ExportFailed, cause.Error()
		if errors.Is(cause, context.Canceled) {
			state, message = aicut.ExportCanceled, ""
		}
		job = job.WithState(state, 100, message, "", s.now())
		_ = s.store.UpdateExportJob(context.Background(), job)
		if state == aicut.ExportCanceled {
			return nil
		}
		return cause
	}
	latest, latestErr := s.store.ExportJob(jobCtx, job.Owner, job.ProjectID, job.ID)
	if latestErr != nil {
		return fail(latestErr)
	}
	if latest.State == aicut.ExportCancelRequested || latest.State == aicut.ExportCanceled {
		return fail(context.Canceled)
	}
	renderOptions := job.Options
	renderOptions.BackgroundColor = job.ProjectSnapshot.BackgroundColor
	output, cleanup, renderErr := renderAICutExportWithOptions(jobCtx, job.Owner, job.ProjectSnapshot, job.TimelineSnapshot, job.Format, renderOptions, s.store, s.objects)
	if renderErr != nil {
		if cleanup != nil {
			cleanup()
		}
		return fail(renderErr)
	}
	if jobCtx.Err() != nil {
		if cleanup != nil {
			cleanup()
		}
		return fail(jobCtx.Err())
	}
	job = job.WithState(aicut.ExportSucceeded, 100, "", output, s.now())
	if err := s.store.UpdateExportJob(jobCtx, job); err != nil {
		if cleanup != nil {
			cleanup()
		}
		return err
	}
	if cleanup != nil {
		time.AfterFunc(time.Hour, cleanup)
	} else if output != "" {
		time.AfterFunc(time.Hour, func() { _ = os.Remove(output) })
	}
	return nil
}

func (s *AICutExportService) Run(ctx context.Context, queue aicut.ExportWakeQueue) {
	if s == nil || queue == nil {
		return
	}
	for ctx.Err() == nil {
		_ = queue.Wait(ctx, 5*time.Second)
		for ctx.Err() == nil {
			err := s.ProcessNext(ctx)
			if errors.Is(err, aicut.ErrExportQueueEmpty) {
				break
			}
			if err != nil {
				continue
			}
		}
	}
}

func exportJobErrorStatus(err error) (int, string) {
	if errors.Is(err, aicut.ErrExportJobNotFound) {
		return 404, "导出任务不存在"
	}
	return 500, fmt.Sprintf("导出任务操作失败：%v", err)
}
