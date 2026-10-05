package video

import (
	"context"
	"errors"
	"sync"
	"time"
)

type memoryLocalExecutorStore struct {
	mu        sync.Mutex
	executors map[string]LocalExecutorRecord
	byToken   map[[32]byte]string
	tasks     map[string]LocalExecutorTask
}

func newMemoryLocalExecutorStore() *memoryLocalExecutorStore {
	return &memoryLocalExecutorStore{
		executors: map[string]LocalExecutorRecord{},
		byToken:   map[[32]byte]string{},
		tasks:     map[string]LocalExecutorTask{},
	}
}

func (s *memoryLocalExecutorStore) CreateLocalExecutor(_ context.Context, record LocalExecutorRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.executors[record.ID] = record
	s.byToken[record.TokenHash] = record.ID
	return nil
}

func (s *memoryLocalExecutorStore) GetLocalExecutorByTokenHash(_ context.Context, hash [32]byte) (LocalExecutorRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.byToken[hash]
	if !ok {
		return LocalExecutorRecord{}, ErrLocalExecutorUnauthorized
	}
	return s.executors[id], nil
}

func (s *memoryLocalExecutorStore) UpdateLocalExecutorHeartbeat(_ context.Context, id string, capabilities []string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.executors[id]
	if !ok {
		return ErrLocalExecutorUnauthorized
	}
	record.Capabilities = append([]string(nil), capabilities...)
	record.LastSeenAt = now
	record.UpdatedAt = now
	s.executors[id] = record
	return nil
}

func (s *memoryLocalExecutorStore) ListLocalExecutors(context.Context) ([]LocalExecutorRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]LocalExecutorRecord, 0, len(s.executors))
	for _, record := range s.executors {
		out = append(out, record)
	}
	return out, nil
}

func (s *memoryLocalExecutorStore) CreateLocalExecutorTask(_ context.Context, task LocalExecutorTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.tasks[task.ID]; exists {
		return errors.New("duplicate local executor task")
	}
	s.tasks[task.ID] = task
	return nil
}

func (s *memoryLocalExecutorStore) GetLocalExecutorTask(_ context.Context, id string) (LocalExecutorTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok {
		return LocalExecutorTask{}, ErrLocalExecutorTaskNotFound
	}
	return task, nil
}

func (s *memoryLocalExecutorStore) CompleteLocalExecutorTask(_ context.Context, id, executorID, artifactURL string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok {
		return ErrLocalExecutorTaskNotFound
	}
	task.ExecutorID = executorID
	task.ArtifactURL = artifactURL
	task.Status = TaskSucceeded
	task.ErrorCode = ""
	task.ErrorMessage = ""
	task.UpdatedAt = now
	s.tasks[id] = task
	return nil
}

func (s *memoryLocalExecutorStore) FailLocalExecutorTask(_ context.Context, id, executorID string, code ErrorCode, message string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok {
		return ErrLocalExecutorTaskNotFound
	}
	task.ExecutorID = executorID
	task.Status = TaskFailed
	task.ErrorCode = code
	task.ErrorMessage = message
	task.UpdatedAt = now
	s.tasks[id] = task
	return nil
}

func (s *memoryLocalExecutorStore) CancelLocalExecutorTask(_ context.Context, id string, now time.Time) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok {
		return false, ErrLocalExecutorTaskNotFound
	}
	if task.Status == TaskSucceeded || task.Status == TaskFailed || task.Status == TaskCancelled {
		return false, nil
	}
	task.Status = TaskCancelled
	task.UpdatedAt = now
	s.tasks[id] = task
	return true, nil
}
