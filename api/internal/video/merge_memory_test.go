package video

import (
	"context"
	"errors"
	"os"
	"sync"
)

type memoryMergeStore struct {
	mu          sync.Mutex
	jobs        map[int64]MergeJob
	attempts    map[int64]MergeAttempt
	nextJobID   int64
	nextAttempt int64
}

func newMemoryMergeStore() *memoryMergeStore {
	return &memoryMergeStore{jobs: map[int64]MergeJob{}, attempts: map[int64]MergeAttempt{}, nextJobID: 1, nextAttempt: 1}
}

func (s *memoryMergeStore) CreateMergeJob(_ context.Context, job MergeJob) (MergeJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job.ID = s.nextJobID
	s.nextJobID++
	s.jobs[job.ID] = job
	return job, nil
}
func (s *memoryMergeStore) GetMergeJob(_ context.Context, id int64) (MergeJob, error) {
	s.mu.Lock(); defer s.mu.Unlock()
	job, ok := s.jobs[id]
	if !ok { return MergeJob{}, errors.New("merge job not found") }
	return job, nil
}
func (s *memoryMergeStore) UpdateMergeJob(_ context.Context, job MergeJob) error {
	s.mu.Lock(); defer s.mu.Unlock(); s.jobs[job.ID] = job; return nil
}
func (s *memoryMergeStore) CreateMergeAttempt(_ context.Context, attempt MergeAttempt) (MergeAttempt, error) {
	s.mu.Lock(); defer s.mu.Unlock()
	attempt.ID = s.nextAttempt
	s.nextAttempt++
	attempt.Inputs = copyMergeInputs(attempt.Inputs)
	s.attempts[attempt.ID] = attempt
	return attempt, nil
}
func (s *memoryMergeStore) GetMergeAttempt(_ context.Context, id int64) (MergeAttempt, error) {
	s.mu.Lock(); defer s.mu.Unlock()
	attempt, ok := s.attempts[id]
	if !ok { return MergeAttempt{}, errors.New("merge attempt not found") }
	attempt.Inputs = copyMergeInputs(attempt.Inputs)
	return attempt, nil
}
func (s *memoryMergeStore) UpdateMergeAttempt(_ context.Context, attempt MergeAttempt) error {
	s.mu.Lock(); defer s.mu.Unlock(); attempt.Inputs = copyMergeInputs(attempt.Inputs); s.attempts[attempt.ID] = attempt; return nil
}
func (s *memoryMergeStore) ListMergeAttempts(_ context.Context, jobID int64) ([]MergeAttempt, error) {
	s.mu.Lock(); defer s.mu.Unlock()
	out := []MergeAttempt{}
	for _, attempt := range s.attempts {
		if attempt.MergeJobID == jobID { attempt.Inputs = copyMergeInputs(attempt.Inputs); out = append(out, attempt) }
	}
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Attempt < out[i].Attempt { out[i], out[j] = out[j], out[i] }
		}
	}
	return out, nil
}

type recordingMergeExecutor struct {
	err              error
	artifact         Artifact
	videoSubmitCalls int
}
func (e *recordingMergeExecutor) Execute(context.Context, MergeExecutionRequest) (Artifact, error) {
	if e.err != nil { return Artifact{}, e.err }
	return e.artifact, nil
}

type fakeCommandRunner struct {
	run func(context.Context, string, ...string) ([]byte, error)
}
func (r *fakeCommandRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	if r.run == nil { return nil, nil }
	return r.run(ctx, name, args...)
}

type fakeMediaDownloader struct{}
func (*fakeMediaDownloader) Download(_ context.Context, _ string, destination string) error {
	return os.WriteFile(destination, []byte("video"), 0o600)
}

type fakeFileArtifactStore struct {
	artifact Artifact
	calls    int
}
func (s *fakeFileArtifactStore) PersistFile(context.Context, string, string) (Artifact, error) {
	s.calls++
	return s.artifact, nil
}
