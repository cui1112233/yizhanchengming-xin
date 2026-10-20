package agentworkspace

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"
)

func cloneVideoJob(input VideoJob) VideoJob {
	raw, _ := json.Marshal(input)
	var out VideoJob
	_ = json.Unmarshal(raw, &out)
	out.OwnerUsername = input.OwnerUsername
	if out.Settings == nil {
		out.Settings = map[string]any{}
	}
	if out.ReferenceAssetIDs == nil {
		out.ReferenceAssetIDs = []string{}
	}
	return out
}

func (s *MemoryStore) CreateVideoJob(_ context.Context, owner string, input CreateVideoJobInput) (VideoJob, error) {
	input, err := normalizeCreateVideoJob(input)
	if err != nil {
		return VideoJob{}, err
	}
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return VideoJob{}, ErrVideoJobInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	project, exists := s.projects[projectKey(owner, input.ProjectID)]
	if !exists || project.DeletedAt != nil {
		return VideoJob{}, ErrProjectNotFound
	}
	if _, exists := s.videoJobs[videoJobKey(owner, input.ID)]; exists {
		return VideoJob{}, ErrVideoJobExists
	}
	if _, exists := s.videoIdempotency[videoIdempotencyKey(owner, input.IdempotencyKey)]; exists {
		return VideoJob{}, ErrVideoJobExists
	}
	job := VideoJob{
		ID: input.ID, ProjectID: input.ProjectID, OwnerUsername: owner, NodeID: input.NodeID, NodeVersion: input.NodeVersion,
		State: VideoJobPreparing, ModelKey: input.ModelKey, Settings: input.Settings, ReferenceAssetIDs: input.ReferenceAssetIDs,
		IdempotencyKey: input.IdempotencyKey, CreatedAt: input.Now.UTC(), UpdatedAt: input.Now.UTC(),
	}
	s.videoJobs[videoJobKey(owner, job.ID)] = cloneVideoJob(job)
	s.videoIdempotency[videoIdempotencyKey(owner, job.IdempotencyKey)] = job.ID
	return cloneVideoJob(job), nil
}

func (s *MemoryStore) VideoJob(_ context.Context, owner, projectID, id string) (VideoJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, exists := s.videoJobs[videoJobKey(strings.TrimSpace(owner), strings.TrimSpace(id))]
	if !exists || job.ProjectID != strings.TrimSpace(projectID) {
		return VideoJob{}, ErrVideoJobNotFound
	}
	if project, ok := s.projects[projectKey(owner, projectID)]; !ok || project.DeletedAt != nil {
		return VideoJob{}, ErrVideoJobNotFound
	}
	return cloneVideoJob(job), nil
}

func (s *MemoryStore) ListVideoJobs(_ context.Context, owner, projectID string) ([]VideoJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	project, exists := s.projects[projectKey(owner, projectID)]
	if !exists || project.DeletedAt != nil {
		return nil, ErrProjectNotFound
	}
	out := make([]VideoJob, 0)
	for _, job := range s.videoJobs {
		if job.OwnerUsername == owner && job.ProjectID == projectID {
			out = append(out, cloneVideoJob(job))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

func (s *MemoryStore) UpdateVideoJob(_ context.Context, owner string, input VideoJob, now time.Time) (VideoJob, error) {
	owner = strings.TrimSpace(owner)
	key := videoJobKey(owner, strings.TrimSpace(input.ID))
	s.mu.Lock()
	defer s.mu.Unlock()
	current, exists := s.videoJobs[key]
	if !exists || current.ProjectID != strings.TrimSpace(input.ProjectID) {
		return VideoJob{}, ErrVideoJobNotFound
	}
	if !validVideoJobState(input.State) || now.IsZero() {
		return VideoJob{}, ErrVideoJobInvalid
	}
	if current.State == VideoJobApplied || current.State == VideoJobCanceled || current.State == VideoJobAbandoned {
		return VideoJob{}, ErrVideoJobState
	}
	current.State = input.State
	current.ProviderTaskID = strings.TrimSpace(input.ProviderTaskID)
	current.ResultAssetID = strings.TrimSpace(input.ResultAssetID)
	current.ErrorMessage = strings.TrimSpace(input.ErrorMessage)
	if input.Settings != nil {
		current.Settings = input.Settings
	}
	current.UpdatedAt = now.UTC()
	s.videoJobs[key] = cloneVideoJob(current)
	return cloneVideoJob(current), nil
}

func (s *MemoryStore) ConfirmVideoJob(_ context.Context, owner, projectID, id string, now time.Time) (VideoJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := videoJobKey(strings.TrimSpace(owner), strings.TrimSpace(id))
	job, exists := s.videoJobs[key]
	if !exists || job.ProjectID != strings.TrimSpace(projectID) {
		return VideoJob{}, ErrVideoJobNotFound
	}
	if job.State != VideoJobAwaitingConfirmation || strings.TrimSpace(job.ResultAssetID) == "" {
		return VideoJob{}, ErrVideoJobState
	}
	project, exists := s.projects[projectKey(owner, projectID)]
	if !exists || project.DeletedAt != nil {
		return VideoJob{}, ErrProjectNotFound
	}
	found := false
	for index := range project.Canvas.Nodes {
		node := &project.Canvas.Nodes[index]
		if node.ID != job.NodeID {
			continue
		}
		found = true
		if node.Type != "video" || node.Version != job.NodeVersion {
			return VideoJob{}, ErrVideoJobNodeChanged
		}
		node.AssetRef = map[string]any{"id": job.ResultAssetID, "kind": "video"}
		node.Status = "confirmed"
		node.Version++
		break
	}
	if !found {
		return VideoJob{}, ErrVideoJobNodeChanged
	}
	project.Canvas.Revision++
	project.UpdatedAt = now.UTC()
	s.projects[projectKey(owner, projectID)] = project
	job.State = VideoJobApplied
	job.UpdatedAt = now.UTC()
	s.videoJobs[key] = cloneVideoJob(job)
	return cloneVideoJob(job), nil
}

func (s *MemoryStore) RetryVideoJob(ctx context.Context, owner, projectID, id string, input RetryVideoJobInput) (VideoJob, error) {
	current, err := s.VideoJob(ctx, owner, projectID, id)
	if err != nil {
		return VideoJob{}, err
	}
	if !isRetryableVideoJob(current.State) {
		return VideoJob{}, ErrVideoJobState
	}
	return s.CreateVideoJob(ctx, owner, CreateVideoJobInput{
		ID: input.ID, ProjectID: projectID, NodeID: current.NodeID, NodeVersion: current.NodeVersion,
		ModelKey: current.ModelKey, Settings: current.Settings, ReferenceAssetIDs: current.ReferenceAssetIDs,
		IdempotencyKey: input.IdempotencyKey, Now: input.Now,
	})
}
