package agentworkspace

import (
	"context"
	"sort"
	"strings"
	"sync"
)

type MemoryAssetStore struct {
	mu     sync.Mutex
	assets map[string]Asset
}

func NewMemoryAssetStore() *MemoryAssetStore {
	return &MemoryAssetStore{assets: make(map[string]Asset)}
}

func assetKey(owner, projectID, assetID string) string {
	return owner + "\x00" + projectID + "\x00" + assetID
}

func (s *MemoryAssetStore) ListAssets(_ context.Context, owner, projectID, kind string) ([]Asset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]Asset, 0)
	for key, asset := range s.assets {
		if !strings.HasPrefix(key, owner+"\x00"+projectID+"\x00") || (kind != "" && asset.Kind != kind) {
			continue
		}
		items = append(items, asset)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return items, nil
}

func (s *MemoryAssetStore) CreateAsset(_ context.Context, owner string, asset Asset) error {
	if err := ValidateAsset(asset); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := assetKey(owner, asset.ProjectID, asset.ID)
	if _, exists := s.assets[key]; exists {
		return ErrAssetExists
	}
	s.assets[key] = asset
	return nil
}

func (s *MemoryAssetStore) Asset(_ context.Context, owner, projectID, assetID string) (Asset, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	asset, ok := s.assets[assetKey(owner, projectID, assetID)]
	if !ok {
		return Asset{}, ErrAssetNotFound
	}
	return asset, nil
}
