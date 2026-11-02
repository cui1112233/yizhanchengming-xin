package localexecutor

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"
)

type MemoryStore struct {
	mu           sync.Mutex
	pairings     map[string]PairingRecord
	executors    map[string]ExecutorRecord
	tokenIDs     map[string]string
	jobs         map[string]JobRecord
	sourceJobs   map[string]string
	artifacts    map[string]ArtifactRecord
	jobArtifacts map[string][]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		pairings: make(map[string]PairingRecord), executors: make(map[string]ExecutorRecord), tokenIDs: make(map[string]string),
		jobs: make(map[string]JobRecord), sourceJobs: make(map[string]string),
		artifacts: make(map[string]ArtifactRecord), jobArtifacts: make(map[string][]string),
	}
}

func (s *MemoryStore) CreatePairing(_ context.Context, record PairingRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pairings[secretHashKey(record.CodeHash)] = record
	return nil
}

func (s *MemoryStore) PairExecutor(_ context.Context, codeHash SecretHash, platform string, executor ExecutorRecord, now time.Time) (ExecutorRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := secretHashKey(codeHash)
	pairing, ok := s.pairings[key]
	if !ok || pairing.ConsumedAt != nil || !pairing.ExpiresAt.After(now) || pairing.Platform != platform {
		return ExecutorRecord{}, ErrPairingInvalid
	}
	consumed := now
	pairing.ConsumedAt = &consumed
	s.pairings[key] = pairing
	executor.OwnerUsername = pairing.OwnerUsername
	// Re-pairing the same named machine refreshes its credential instead of
	// creating a second visible device entry.
	for id, existing := range s.executors {
		if existing.OwnerUsername == executor.OwnerUsername && existing.Platform == platform && existing.DeviceName == executor.DeviceName && existing.OS == executor.OS {
			delete(s.tokenIDs, secretHashKey(existing.TokenHash))
			executor.ID = id
			break
		}
	}
	s.executors[executor.ID] = executor
	s.tokenIDs[secretHashKey(executor.TokenHash)] = executor.ID
	return executor, nil
}

func (s *MemoryStore) ExecutorByTokenHash(_ context.Context, tokenHash SecretHash) (ExecutorRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.tokenIDs[secretHashKey(tokenHash)]
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
	if input.DeviceName != "" {
		record.DeviceName = input.DeviceName
	}
	if input.OS != "" {
		record.OS = input.OS
	}
	if input.Version != "" {
		record.Version = input.Version
	}
	record.Accounts = input.Accounts
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
	seen := make(map[string]bool)
	for _, record := range s.executors {
		if record.OwnerUsername == owner {
			key := record.Platform + "\x00" + record.DeviceName + "\x00" + record.OS
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, record)
		}
	}
	return out, nil
}

func (s *MemoryStore) PairingPlaintextSeen(secret string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	raw, _ := json.Marshal(s.pairings)
	return strings.Contains(string(raw), secret)
}

func (s *MemoryStore) ExecutorPlaintextTokenSeen(secret string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	raw, _ := json.Marshal(struct {
		Executors map[string]ExecutorRecord
		Tokens    map[string]string
	}{s.executors, s.tokenIDs})
	return strings.Contains(string(raw), secret)
}
