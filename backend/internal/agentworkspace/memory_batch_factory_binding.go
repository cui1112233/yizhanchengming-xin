package agentworkspace

import (
	"context"
	"strings"
	"time"
)

func batchFactoryBindingKey(owner, batchID string) string {
	return strings.TrimSpace(owner) + "\x00" + strings.TrimSpace(batchID)
}

func batchFactoryProjectKey(owner, projectID string) string {
	return strings.TrimSpace(owner) + "\x00" + strings.TrimSpace(projectID)
}

func batchFactoryNodeOriginKey(owner, projectID, nodeID string) string {
	return strings.TrimSpace(owner) + "\x00" + strings.TrimSpace(projectID) + "\x00" + strings.TrimSpace(nodeID)
}

func (s *MemoryStore) BindBatchFactoryProject(_ context.Context, binding BatchFactoryBinding) (BatchFactoryBinding, error) {
	binding, err := normalizeBatchFactoryBinding(binding, time.Now().UTC())
	if err != nil {
		return BatchFactoryBinding{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, found := s.batchFactoryBindings[batchFactoryBindingKey(binding.Owner, binding.BatchID)]; found {
		if existing.ProjectID != binding.ProjectID {
			return BatchFactoryBinding{}, ErrBatchFactoryBindingConflict
		}
		return existing, nil
	}
	if existingBatchID, found := s.batchFactoryProjects[batchFactoryProjectKey(binding.Owner, binding.ProjectID)]; found && existingBatchID != binding.BatchID {
		return BatchFactoryBinding{}, ErrBatchFactoryBindingConflict
	}
	s.batchFactoryBindings[batchFactoryBindingKey(binding.Owner, binding.BatchID)] = binding
	s.batchFactoryProjects[batchFactoryProjectKey(binding.Owner, binding.ProjectID)] = binding.BatchID
	return binding, nil
}

func (s *MemoryStore) BatchFactoryBinding(_ context.Context, owner, batchID string) (BatchFactoryBinding, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	binding, found := s.batchFactoryBindings[batchFactoryBindingKey(owner, batchID)]
	if !found {
		return BatchFactoryBinding{}, ErrBatchFactoryBindingNotFound
	}
	return binding, nil
}

func (s *MemoryStore) BindBatchFactoryNode(_ context.Context, owner string, origin BatchFactoryNodeOrigin) error {
	origin, err := normalizeBatchFactoryNodeOrigin(owner, origin)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	binding, found := s.batchFactoryBindings[batchFactoryBindingKey(owner, origin.BatchID)]
	if !found || binding.ProjectID != origin.ProjectID {
		return ErrBatchFactoryBindingNotFound
	}
	s.batchFactoryNodeOrigins[batchFactoryNodeOriginKey(owner, origin.ProjectID, origin.NodeID)] = origin
	return nil
}

func (s *MemoryStore) BatchFactoryNodeOrigins(_ context.Context, owner, projectID, bookID string) ([]BatchFactoryNodeOrigin, error) {
	owner, projectID, bookID = strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(bookID)
	if owner == "" || projectID == "" || bookID == "" {
		return nil, ErrBatchFactoryNodeOriginInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	prefix := owner + "\x00" + projectID + "\x00"
	origins := make([]BatchFactoryNodeOrigin, 0)
	for key, origin := range s.batchFactoryNodeOrigins {
		if strings.HasPrefix(key, prefix) && origin.BookID == bookID {
			origins = append(origins, origin)
		}
	}
	return origins, nil
}

func (s *MemoryStore) RemoveBatchFactoryNodeBinding(_ context.Context, owner, projectID, nodeID string) error {
	owner, projectID, nodeID = strings.TrimSpace(owner), strings.TrimSpace(projectID), strings.TrimSpace(nodeID)
	if owner == "" || projectID == "" || nodeID == "" {
		return ErrBatchFactoryNodeOriginInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.batchFactoryNodeOrigins, batchFactoryNodeOriginKey(owner, projectID, nodeID))
	return nil
}

var _ BatchFactoryBindingStore = (*MemoryStore)(nil)
