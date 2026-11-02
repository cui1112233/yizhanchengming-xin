package localexecutor

import (
	"context"
	"time"
)

func (s *MemoryStore) CreateArtifact(_ context.Context, executorID, jobID string, leaseHash SecretHash, generation int64, input ArtifactRecord, now time.Time) (ArtifactRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[jobID]
	if !ok {
		return ArtifactRecord{}, ErrJobNotFound
	}
	if err := checkLease(job, executorID, leaseHash, generation, now); err != nil {
		return ArtifactRecord{}, err
	}
	if !isAccepted(job) || job.State != JobUploading {
		return ArtifactRecord{}, ErrInvalidJobState
	}
	if existingIDs := s.jobArtifacts[jobID]; len(existingIDs) > 0 {
		for _, existingID := range existingIDs {
			existing := s.artifacts[existingID]
			if sameArtifactContent(existing, input) {
				return existing, nil
			}
		}
	}
	if _, exists := s.artifacts[input.ID]; exists {
		return ArtifactRecord{}, ErrArtifactConflict
	}

	input.JobID = jobID
	input.OwnerUsername = job.OwnerUsername
	input.CreatedAt = now
	s.artifacts[input.ID] = input
	s.jobArtifacts[jobID] = append(s.jobArtifacts[jobID], input.ID)
	return input, nil
}

func (s *MemoryStore) ArtifactForOwner(_ context.Context, owner, id string) (ArtifactRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.artifacts[id]
	if !ok || record.OwnerUsername != owner {
		return ArtifactRecord{}, ErrArtifactNotFound
	}
	return record, nil
}

func (s *MemoryStore) ArtifactForJob(_ context.Context, jobID string) (ArtifactRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	artifactIDs := s.jobArtifacts[jobID]
	if len(artifactIDs) == 0 {
		return ArtifactRecord{}, ErrArtifactNotFound
	}
	record, ok := s.artifacts[artifactIDs[0]]
	if !ok {
		return ArtifactRecord{}, ErrArtifactNotFound
	}
	return record, nil
}

func (s *MemoryStore) ArtifactsForJob(_ context.Context, jobID string) ([]ArtifactRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := s.jobArtifacts[jobID]
	if len(ids) == 0 {
		return nil, ErrArtifactNotFound
	}
	result := make([]ArtifactRecord, 0, len(ids))
	for _, id := range ids {
		record, ok := s.artifacts[id]
		if !ok {
			return nil, ErrArtifactNotFound
		}
		result = append(result, record)
	}
	return result, nil
}

func sameArtifactContent(left, right ArtifactRecord) bool {
	return left.MediaType == right.MediaType && left.ByteSize == right.ByteSize && left.SHA256 == right.SHA256
}
