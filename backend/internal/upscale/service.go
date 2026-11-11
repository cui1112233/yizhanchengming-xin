package upscale

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// API 是本服务依赖的火山接口集合，便于单测替换。
type API interface {
	Configured() bool
	RequestUploadURL(ctx context.Context) (UploadTarget, error)
	UploadMedia(ctx context.Context, target UploadTarget, body io.Reader, size int64, contentType string) error
	SubmitEnhance(ctx context.Context, videoURL, resolution string) (string, error)
	QueryTask(ctx context.Context, taskID string) (Task, error)
	Terminate(ctx context.Context, taskIDs []string) (TerminateOutcome, error)
}

// Service 负责「上传 -> 提交 -> 查询 -> 取消」的完整超分流程。
type Service struct {
	API   API
	Jobs  *JobStore
	Now   func() time.Time
	NewID func() string
}

func NewService(api API, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{API: api, Jobs: NewJobStore(), Now: now, NewID: newJobID}
}

// Configured 用于前端判断是否展示「未配置火山凭证」。
func (s *Service) Configured() bool {
	return s != nil && s.API != nil && s.API.Configured()
}

// Submit 上传本地视频并提交超分任务。
func (s *Service) Submit(ctx context.Context, owner string, media io.Reader, size int64, contentType, fileName, resolution string) (Job, error) {
	if !s.Configured() {
		return Job{}, ErrNotConfigured
	}
	target := normalizeResolution(resolution)
	if target == "" {
		return Job{}, ErrInvalidResolution
	}
	now := s.Now()
	job := Job{
		ID:         s.newID(),
		Owner:      strings.TrimSpace(owner),
		Status:     StatusQueued,
		Resolution: target,
		FileName:   strings.TrimSpace(fileName),
		Progress:   intPointer(0),
		Message:    "正在上传视频到火山",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.Jobs.Create(job); err != nil {
		return Job{}, err
	}
	upload, err := s.API.RequestUploadURL(ctx)
	if err != nil {
		return s.fail(job, err)
	}
	if err := s.API.UploadMedia(ctx, upload, media, size, contentType); err != nil {
		return s.fail(job, err)
	}
	taskID, err := s.API.SubmitEnhance(ctx, NormalizeFileID(upload.FileID), target)
	if err != nil {
		return s.fail(job, err)
	}
	return s.Jobs.Update(job.Owner, job.ID, func(item *Job) {
		item.TaskID = taskID
		item.Status = StatusRunning
		item.Progress = nil
		item.Message = "火山处理中…"
		item.UpdatedAt = s.Now()
	})
}

// Job 返回本地记录；running 状态会顺带刷新上游状态。
func (s *Service) Job(ctx context.Context, owner, id string) (Job, error) {
	job, err := s.Jobs.Get(owner, id)
	if err != nil {
		return Job{}, err
	}
	if job.Status != StatusRunning || job.TaskID == "" {
		return job, nil
	}
	task, err := s.API.QueryTask(ctx, job.TaskID)
	if err != nil {
		return job, err
	}
	return s.applyTask(job, task)
}

// Cancel 取消任务。已完成/失败/已取消的任务原样返回，作为幂等处理。
func (s *Service) Cancel(ctx context.Context, owner, id string) (Job, error) {
	job, err := s.Jobs.Get(owner, id)
	if err != nil {
		return Job{}, err
	}
	switch job.Status {
	case StatusCompleted, StatusFailed, StatusCancelled:
		return job, nil
	}
	if job.TaskID == "" {
		return s.Jobs.Update(job.Owner, job.ID, func(item *Job) {
			item.Status = StatusCancelled
			item.Progress = nil
			item.Message = "已取消"
			item.UpdatedAt = s.Now()
		})
	}
	outcome, err := s.API.Terminate(ctx, []string{job.TaskID})
	if err != nil {
		return job, err
	}
	if message, failed := outcome.Failed[job.TaskID]; failed {
		return job, fmt.Errorf("火山暂不支持取消该任务：%s", message)
	}
	terminated := false
	for _, taskID := range outcome.Terminated {
		if taskID == job.TaskID {
			terminated = true
			break
		}
	}
	if !terminated {
		return job, errors.New("火山未确认取消结果，请稍后刷新状态")
	}
	return s.Jobs.Update(job.Owner, job.ID, func(item *Job) {
		item.Status = StatusCancelled
		item.Progress = nil
		item.Message = "已取消"
		item.UpdatedAt = s.Now()
	})
}

func (s *Service) applyTask(job Job, task Task) (Job, error) {
	status := strings.ToLower(strings.TrimSpace(task.Status))
	return s.Jobs.Update(job.Owner, job.ID, func(item *Job) {
		item.UpdatedAt = s.Now()
		switch status {
		case "completed":
			item.Status = StatusCompleted
			item.Progress = intPointer(100)
			item.ResultURL = task.VideoURL
			item.Message = "超分完成"
			item.Error = ""
			if task.VideoURL == "" {
				item.Message = "任务完成但火山未返回视频地址"
			}
		case "failed":
			item.Status = StatusFailed
			item.Progress = nil
			item.Error = taskErrorMessage(task)
			item.Message = "超分失败"
		default:
			item.Status = StatusRunning
			item.Progress = nil
			item.Message = "火山处理中…"
		}
	})
}

func (s *Service) fail(job Job, cause error) (Job, error) {
	updated, updateErr := s.Jobs.Update(job.Owner, job.ID, func(item *Job) {
		item.Status = StatusFailed
		item.Progress = nil
		item.Error = cause.Error()
		item.Message = "超分失败"
		item.UpdatedAt = s.Now()
	})
	if updateErr != nil {
		return Job{}, cause
	}
	return updated, cause
}

func (s *Service) newID() string {
	if s.NewID != nil {
		return s.NewID()
	}
	return newJobID()
}

func newJobID() string {
	buffer := make([]byte, 8)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("upscale-%d", time.Now().UnixNano())
	}
	return "upscale-" + hex.EncodeToString(buffer)
}

func normalizeResolution(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1080p":
		return "1080p"
	case "2k":
		return "2k"
	default:
		return ""
	}
}

func taskErrorMessage(task Task) string {
	if message := strings.TrimSpace(task.ErrorMessage); message != "" {
		if code := strings.TrimSpace(task.ErrorCode); code != "" {
			return message + "（" + code + "）"
		}
		return message
	}
	if code := strings.TrimSpace(task.ErrorCode); code != "" {
		return code
	}
	return "火山任务执行失败"
}

func intPointer(value int) *int {
	return &value
}
