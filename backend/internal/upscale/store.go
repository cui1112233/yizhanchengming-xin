package upscale

import (
	"errors"
	"strings"
	"sync"
	"time"
)

var (
	// ErrJobNotFound 表示任务不存在，或不属于当前用户。
	ErrJobNotFound = errors.New("超分任务不存在")
	// ErrDuplicateJob 表示任务 ID 冲突。
	ErrDuplicateJob = errors.New("超分任务已存在")
)

// 任务状态。
const (
	StatusQueued    = "queued"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"
)

// Job 是一次超分任务的本地记录。上游 task_id 只用于轮询与取消。
type Job struct {
	ID         string    `json:"id"`
	Owner      string    `json:"-"`
	Status     string    `json:"status"`
	Resolution string    `json:"resolution"`
	FileName   string    `json:"fileName"`
	Progress   *int      `json:"progress"`
	Message    string    `json:"message,omitempty"`
	Error      string    `json:"error,omitempty"`
	ResultURL  string    `json:"resultUrl,omitempty"`
	TaskID     string    `json:"taskId,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// JobStore 是进程内任务表。重启后历史记录会丢失，但不会影响上游任务本身。
type JobStore struct {
	mu   sync.RWMutex
	jobs map[string]Job
}

func NewJobStore() *JobStore {
	return &JobStore{jobs: map[string]Job{}}
}

func (s *JobStore) Create(job Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.jobs[job.ID]; exists {
		return ErrDuplicateJob
	}
	s.jobs[job.ID] = job
	return nil
}

// Get 按任务 ID 读取，并校验归属，避免跨用户访问。
func (s *JobStore) Get(owner, id string) (Job, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[strings.TrimSpace(id)]
	if !ok || job.Owner != strings.TrimSpace(owner) {
		return Job{}, ErrJobNotFound
	}
	return job, nil
}

func (s *JobStore) Update(owner, id string, mutate func(*Job)) (Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[strings.TrimSpace(id)]
	if !ok || job.Owner != strings.TrimSpace(owner) {
		return Job{}, ErrJobNotFound
	}
	if mutate != nil {
		mutate(&job)
	}
	s.jobs[job.ID] = job
	return job, nil
}
