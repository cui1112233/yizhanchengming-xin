package giantmaterialexecutor

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"sort"
	"strings"
	"sync"
	"time"
)

type MemoryStore struct {
	mu              sync.Mutex
	pairings        map[string]PairingRecord
	executors       map[string]ExecutorRecord
	tokens          map[string]string
	jobs            map[string]JobRecord
	jobKeys         map[string]string
	preferences     map[string]PreferenceRecord
	progressChanged map[string]time.Time
	lastFailure     map[string]time.Time
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{pairings: make(map[string]PairingRecord), executors: make(map[string]ExecutorRecord), tokens: make(map[string]string), jobs: make(map[string]JobRecord), jobKeys: make(map[string]string), preferences: make(map[string]PreferenceRecord), progressChanged: make(map[string]time.Time), lastFailure: make(map[string]time.Time)}
}

func (s *MemoryStore) EnsureIdentityUniqueIndex(_ context.Context) error { return nil }

func (s *MemoryStore) ConsolidateDuplicateExecutors(_ context.Context, now time.Time) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	type groupKey struct {
		owner, osName, deviceName string
	}
	groups := make(map[groupKey][]executorCandidate)
	for id, item := range s.executors {
		key := groupKey{item.OwnerUsername, item.OS, item.DeviceName}
		groups[key] = append(groups[key], executorCandidate{
			id: id, version: item.Version, lastSeen: item.LastSeenAt, updatedAt: item.UpdatedAt,
		})
	}
	merged := 0
	for _, candidates := range groups {
		if len(candidates) < 2 {
			continue
		}
		keeping := pickKeeping(candidates)
		for _, item := range candidates {
			if item.id == keeping {
				continue
			}
			for _, job := range s.jobs {
				if job.LeaseExecutorID == item.id {
					job.LeaseExecutorID = keeping
					s.jobs[job.ID] = job
				}
			}
			delete(s.executors, item.id)
			merged++
		}
	}
	return merged, nil
}

func (s *MemoryStore) RecentFailures(_ context.Context, owner string, since time.Time) (map[string]time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]time.Time)
	for executorID, failedAt := range s.lastFailure {
		if failedAt.Before(since) {
			continue
		}
		executor, ok := s.executors[executorID]
		if !ok || executor.OwnerUsername != owner {
			continue
		}
		out[executorID] = failedAt
	}
	return out, nil
}

func (s *MemoryStore) CreatePairing(_ context.Context, record PairingRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pairings[secretHashKey(record.CodeHash)] = record
	return nil
}

// ListPendingPairings 返回最近一条未消费的配对记录（新的在前），并在读取时
// 清理超过保留期的过期记录。已过期的最近记录仍会返回，由调用方判断有效性。
func (s *MemoryStore) ListPendingPairings(_ context.Context, owner, platform string, now time.Time) ([]PairingRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cutoff := now.Add(-pairingCleanupRetention)
	out := make([]PairingRecord, 0, 4)
	for key, record := range s.pairings {
		if record.ExpiresAt.Before(cutoff) {
			delete(s.pairings, key)
			continue
		}
		if record.OwnerUsername != owner || record.Platform != platform || record.ConsumedAt != nil || record.CreatedAt.Before(cutoff) {
			continue
		}
		out = append(out, record)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (s *MemoryStore) PairExecutor(_ context.Context, codeHash SecretHash, platform string, executor ExecutorRecord, now time.Time) (ExecutorRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	pairing, ok := s.pairings[secretHashKey(codeHash)]
	if !ok || pairing.ConsumedAt != nil || pairing.Platform != platform || !pairing.ExpiresAt.After(now) {
		return ExecutorRecord{}, ErrPairingInvalid
	}
	consumed := now
	pairing.ConsumedAt = &consumed
	s.pairings[secretHashKey(codeHash)] = pairing
	executor.OwnerUsername = pairing.OwnerUsername
	var existingID string
	for id, item := range s.executors {
		if item.OwnerUsername != pairing.OwnerUsername || item.OS != executor.OS || item.DeviceName != executor.DeviceName {
			continue
		}
		if existingID == "" || item.UpdatedAt.After(s.executors[existingID].UpdatedAt) {
			existingID = id
		}
	}
	if existingID != "" {
		old := s.executors[existingID]
		delete(s.tokens, secretHashKey(old.TokenHash))
		executor.ID = existingID
		executor.UpdatedAt = now
		executor.LastSeenAt = old.LastSeenAt
		s.executors[existingID] = executor
		s.tokens[secretHashKey(executor.TokenHash)] = existingID
	} else {
		s.executors[executor.ID] = executor
		s.tokens[secretHashKey(executor.TokenHash)] = executor.ID
	}
	return executor, nil
}

func (s *MemoryStore) ExecutorByTokenHash(_ context.Context, hash SecretHash) (ExecutorRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.tokens[secretHashKey(hash)]
	if !ok {
		return ExecutorRecord{}, ErrExecutorUnauthorized
	}
	record, ok := s.executors[id]
	if !ok {
		return ExecutorRecord{}, ErrExecutorUnauthorized
	}
	return record, nil
}

func (s *MemoryStore) UpdateHeartbeat(_ context.Context, id string, input HeartbeatInput, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.executors[id]
	if !ok {
		return ErrExecutorUnauthorized
	}
	if strings.TrimSpace(input.DeviceName) != "" {
		record.DeviceName = strings.TrimSpace(input.DeviceName)
	}
	if strings.TrimSpace(input.OS) != "" {
		record.OS = strings.TrimSpace(input.OS)
	}
	if strings.TrimSpace(input.Version) != "" {
		record.Version = strings.TrimSpace(input.Version)
	}
	seen := now
	record.LastSeenAt = &seen
	record.UpdatedAt = now
	s.executors[id] = record
	return nil
}

func (s *MemoryStore) ListExecutors(_ context.Context, owner string) ([]ExecutorRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ExecutorRecord, 0)
	for _, record := range s.executors {
		if record.OwnerUsername == owner {
			out = append(out, record)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

func (s *MemoryStore) LatestPlatformFailure(_ context.Context, owner, osName string, since time.Time) (*time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var latest *time.Time
	for executorID, failedAt := range s.lastFailure {
		if failedAt.Before(since) {
			continue
		}
		executor, ok := s.executors[executorID]
		if !ok || executor.OwnerUsername != owner || executor.OS != osName {
			continue
		}
		copyTime := failedAt
		if latest == nil || copyTime.After(*latest) {
			latest = &copyTime
		}
	}
	return latest, nil
}

func (s *MemoryStore) GetPreference(_ context.Context, owner string) (PreferenceRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.preferences[owner]
	if !ok {
		return PreferenceRecord{}, ErrPreferenceNotFound
	}
	return record, nil
}

func (s *MemoryStore) SavePreference(_ context.Context, record PreferenceRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.preferences[record.OwnerUsername] = record
	return nil
}

func (s *MemoryStore) FindJobByKey(_ context.Context, owner, key string) (JobRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.jobKeys[owner+"\x00"+key]
	if !ok {
		return JobRecord{}, ErrJobNotFound
	}
	record, ok := s.jobs[id]
	if !ok {
		return JobRecord{}, ErrJobNotFound
	}
	return record, nil
}

func (s *MemoryStore) CreateJob(_ context.Context, record JobRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := record.OwnerUsername + "\x00" + jobKey(record)
	if _, exists := s.jobKeys[key]; exists {
		return ErrJobConflict
	}
	s.jobs[record.ID] = record
	s.jobKeys[key] = record.ID
	return nil
}

func (s *MemoryStore) JobForOwner(_ context.Context, owner, id string) (JobRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.jobs[id]
	if !ok || record.OwnerUsername != owner {
		return JobRecord{}, ErrJobNotFound
	}
	return record, nil
}

func (s *MemoryStore) FailQueuedJob(_ context.Context, owner, id string, failure FailureInput, now time.Time) (JobRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.jobs[id]
	if !ok || record.OwnerUsername != owner {
		return JobRecord{}, ErrJobNotFound
	}
	if record.State == JobQueued {
		record.State = JobFailed
		record.ErrorCode = bounded(failure.Code, 96)
		record.ErrorMessage = bounded(failure.Message, 512)
		record.UpdatedAt = now
		s.jobs[id] = record
	}
	return record, nil
}

func (s *MemoryStore) CancelJob(_ context.Context, owner, id string, now time.Time) (JobRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.jobs[id]
	if !ok || record.OwnerUsername != owner {
		return JobRecord{}, ErrJobNotFound
	}
	if record.State != JobSucceeded && record.State != JobFailed && record.State != JobCancelled {
		record.CancelRequested = true
		record.State = JobCancelled
		record.UpdatedAt = now
		s.jobs[id] = record
	}
	return record, nil
}

func (s *MemoryStore) RequeueJob(_ context.Context, id string, update JobRecord, now time.Time) (JobRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.jobs[id]
	if !ok {
		return JobRecord{}, ErrJobNotFound
	}
	if record.State != JobFailed && record.State != JobCancelled {
		return record, nil
	}
	record.State = JobQueued
	record.CancelRequested = false
	record.Title = update.Title
	record.VideoURL = update.VideoURL
	record.VideoExpiresAt = update.VideoExpiresAt
	record.DurationSeconds = update.DurationSeconds
	record.ContentRangeLines = update.ContentRangeLines
	record.TargetExecutorID = update.TargetExecutorID
	record.LeaseExecutorID = ""
	record.LeaseTokenHash = SecretHash{}
	record.LeaseGeneration++
	record.LeaseExpiresAt = nil
	record.ErrorCode = ""
	record.ErrorMessage = ""
	record.Progress = ProgressInput{}
	record.UpdatedAt = now
	s.jobs[id] = record
	delete(s.progressChanged, id)
	return record, nil
}

func (s *MemoryStore) ClaimJob(_ context.Context, executor ExecutorRecord, leaseHash SecretHash, expires, now time.Time, targetedOnly bool) (JobRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var picked *JobRecord
	for _, candidate := range s.jobs {
		if candidate.OwnerUsername != executor.OwnerUsername || candidate.Platform != PlatformGiantMaterial || candidate.CancelRequested || candidate.State == JobCancelled || candidate.State == JobSucceeded || candidate.State == JobFailed {
			continue
		}
		if targetedOnly {
			// 只领明确点名给本执行器的任务（定向重试），平台偏好门控不拦截。
			if candidate.TargetExecutorID != executor.ID {
				continue
			}
		} else if candidate.TargetExecutorID != "" {
			// 公共池：点名给别人的任务不参与普通领取。
			continue
		}
		stuck := false
		switch candidate.State {
		case JobLeased, JobRunning, JobCleaning, JobUploading:
			if candidate.LeaseExpiresAt != nil && candidate.LeaseExpiresAt.After(now) {
				if pc, ok := s.progressChanged[candidate.ID]; ok && !pc.After(now.Add(-StuckProgressLimit)) {
					stuck = true
				}
			}
		}
		claimable := candidate.State == JobQueued ||
			(candidate.LeaseExpiresAt != nil && !candidate.LeaseExpiresAt.After(now)) ||
			stuck
		if !claimable {
			continue
		}
		copy := candidate
		if picked == nil || copy.CreatedAt.Before(picked.CreatedAt) || (copy.CreatedAt.Equal(picked.CreatedAt) && copy.ID < picked.ID) {
			picked = &copy
		}
	}
	if picked == nil {
		return JobRecord{}, ErrNoClaimableJob
	}
	record := *picked
	record.State = JobLeased
	record.LeaseExecutorID = executor.ID
	record.LeaseTokenHash = leaseHash
	record.LeaseGeneration++
	record.LeaseExpiresAt = &expires
	record.UpdatedAt = now
	record.ErrorCode = ""
	record.ErrorMessage = ""
	s.jobs[record.ID] = record
	s.progressChanged[record.ID] = now
	return record, nil
}

func (s *MemoryStore) RenewJob(_ context.Context, executorID, id string, leaseHash SecretHash, generation int64, expires, now time.Time) (JobRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, err := s.leaseLocked(id, executorID, leaseHash, generation, now)
	if err != nil {
		return JobRecord{}, err
	}
	record.LeaseExpiresAt = &expires
	record.UpdatedAt = now
	s.jobs[id] = record
	return record, nil
}

func (s *MemoryStore) SetProgress(_ context.Context, executorID, id string, leaseHash SecretHash, generation int64, next JobState, progress ProgressInput, now time.Time) (JobRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, err := s.leaseLocked(id, executorID, leaseHash, generation, now)
	if err != nil {
		return JobRecord{}, err
	}
	if err := validJobTransition(record.State, next); err != nil {
		return JobRecord{}, err
	}
	if record.Progress.Completed != progress.Completed || record.Progress.Total != progress.Total || record.Progress.Percent != progress.Percent {
		s.progressChanged[id] = now
	}
	record.State = next
	record.Progress = progress
	record.UpdatedAt = now
	s.jobs[id] = record
	return record, nil
}

func (s *MemoryStore) CompleteJob(_ context.Context, executorID, id string, leaseHash SecretHash, generation int64, result ResultInput, now time.Time) (JobRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, err := s.leaseLocked(id, executorID, leaseHash, generation, now)
	if err != nil {
		return JobRecord{}, err
	}
	if record.State != JobUploading {
		return JobRecord{}, ErrInvalidJobState
	}
	record.State = JobSucceeded
	record.Result = ResultRecord{Text: result.Text, WordCount: result.WordCount, CreatedAt: now}
	record.UpdatedAt = now
	s.jobs[id] = record
	return record, nil
}

func (s *MemoryStore) FailJob(_ context.Context, executorID, id string, leaseHash SecretHash, generation int64, failure FailureInput, now time.Time) (JobRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, err := s.leaseLocked(id, executorID, leaseHash, generation, now)
	if err != nil {
		return JobRecord{}, err
	}
	record.State = JobFailed
	record.ErrorCode = bounded(failure.Code, 96)
	record.ErrorMessage = bounded(failure.Message, 512)
	record.UpdatedAt = now
	s.jobs[id] = record
	s.lastFailure[executorID] = now
	return record, nil
}

func (s *MemoryStore) leaseLocked(id, executorID string, leaseHash SecretHash, generation int64, now time.Time) (JobRecord, error) {
	record, ok := s.jobs[id]
	if !ok {
		return JobRecord{}, ErrJobNotFound
	}
	if record.CancelRequested || record.State == JobCancelled {
		return JobRecord{}, ErrJobCancelled
	}
	if record.LeaseExecutorID != executorID || record.LeaseGeneration != generation || record.LeaseTokenHash != leaseHash || record.LeaseExpiresAt == nil || !record.LeaseExpiresAt.After(now) {
		return JobRecord{}, ErrStaleLease
	}
	return record, nil
}

func secretHashKey(hash SecretHash) string { return base64.RawURLEncoding.EncodeToString(hash[:]) }
func hashSecret(value string) SecretHash   { return sha256.Sum256([]byte(value)) }
func hashPairingCode(value string) SecretHash {
	return sha256.Sum256([]byte(normalizePairingCode(value)))
}

func jobKey(record JobRecord) string {
	return strings.Join([]string{record.MaterialID, record.PlatformBookID, record.ModelVersion}, "\x00")
}

func bounded(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) > max {
		return value[:max]
	}
	return value
}

func (s *MemoryStore) DeleteExecutor(_ context.Context, owner, id string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.executors[id]
	if !ok || record.OwnerUsername != owner {
		return ErrExecutorNotFound
	}
	delete(s.tokens, secretHashKey(record.TokenHash))
	delete(s.lastFailure, id)
	for jobID, job := range s.jobs {
		terminal := false
		switch job.State {
		case JobSucceeded, JobFailed, JobCancelled:
			terminal = true
		}
		if job.LeaseExecutorID == id && !terminal {
			job.State = JobQueued
			job.CancelRequested = false
			job.LeaseExecutorID = ""
			job.LeaseTokenHash = SecretHash{}
			job.LeaseExpiresAt = nil
			job.LeaseGeneration++
			job.Progress = ProgressInput{}
			job.ErrorCode = ""
			job.ErrorMessage = ""
			job.UpdatedAt = now
			s.jobs[jobID] = job
			delete(s.progressChanged, jobID)
		}
	}
	delete(s.executors, id)
	return nil
}
