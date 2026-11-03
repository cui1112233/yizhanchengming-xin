package batchfactoryv11

import (
	"context"
	"errors"
	"sync"
	"testing"
)

// fakeOwnerConfigStore mirrors the account config document used by the MySQL
// store so the persistence contract can be tested without a database.
type fakeOwnerConfigStore struct {
	mu      sync.Mutex
	configs map[string]map[string]any
}

func newFakeOwnerConfigStore() *fakeOwnerConfigStore {
	return &fakeOwnerConfigStore{configs: map[string]map[string]any{}}
}

func (s *fakeOwnerConfigStore) GetConfig(_ context.Context, owner string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	config, ok := s.configs[owner]
	if !ok {
		return map[string]any{}, nil
	}
	cloned := map[string]any{}
	for key, value := range config {
		cloned[key] = value
	}
	return cloned, nil
}

func (s *fakeOwnerConfigStore) PutConfig(_ context.Context, owner string, config map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configs[owner] = config
	return nil
}

func (s *fakeOwnerConfigStore) ListConfigOwners(context.Context) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	owners := make([]string, 0, len(s.configs))
	for owner := range s.configs {
		owners = append(owners, owner)
	}
	return owners, nil
}

func TestPersistentVideoProviderRegistrySurvivesRestart(t *testing.T) {
	ctx := context.Background()
	store := newFakeOwnerConfigStore()

	first := NewPersistentVideoProviderRegistry(store)
	if err := first.SaveConfig(ctx, "alice", VideoProviderConfig{Provider: VideoProviderPersonalAPI, APIKey: "yd-key"}); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	// A fresh registry simulates a process restart: only the stored document is
	// available, so the warm/lazy hydration path must restore the provider.
	second := NewPersistentVideoProviderRegistry(store)
	second.Warm(ctx)
	cfg, err := second.Resolve(ctx, "alice", VideoProviderPersonalAPI)
	if err != nil {
		t.Fatalf("Resolve after restart: %v", err)
	}
	if cfg.APIKey != "yd-key" || cfg.Model != DefaultPersonalVideoModel {
		t.Fatalf("unexpected restored config: %+v", cfg)
	}
}

func TestPersistentVideoProviderRegistryHydratesLazily(t *testing.T) {
	ctx := context.Background()
	store := newFakeOwnerConfigStore()
	writer := NewPersistentVideoProviderRegistry(store)
	if err := writer.SaveConfig(ctx, "bob", VideoProviderConfig{Provider: VideoProviderYFAISeedance, APIKey: "yfai-key"}); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	// No Warm call here: Resolve must fall back to reading the stored document.
	reader := NewPersistentVideoProviderRegistry(store)
	if _, err := reader.Resolve(ctx, "bob", VideoProviderYFAISeedance); err != nil {
		t.Fatalf("lazy Resolve: %v", err)
	}
}

func TestPersistentVideoProviderRegistryPutDoesNotWriteConfig(t *testing.T) {
	ctx := context.Background()
	store := newFakeOwnerConfigStore()
	registry := NewPersistentVideoProviderRegistry(store)
	if err := registry.Put(ctx, "carol", VideoProviderConfig{Provider: VideoProviderDoubaoLocal}); err != nil {
		t.Fatalf("Put: %v", err)
	}
	if _, ok := store.configs["carol"]; ok {
		t.Fatal("catalog synchronization must not persist a duplicate provider setting")
	}
	view, err := registry.View(ctx, "carol", VideoProviderDoubaoLocal)
	if err != nil {
		t.Fatalf("View: %v", err)
	}
	if !view.Configured {
		t.Fatal("local doubao provider should be configured without an API key")
	}
}

func TestResolveVideoProviderSettingPrefersExplicitSave(t *testing.T) {
	config := map[string]any{
		"videoProviders": map[string]any{
			VideoProviderPersonalAPI: map[string]any{"apiKey": "saved-key", "model": "saved-model"},
		},
		"modelCatalogVersion": 1,
		"modelCatalog": []any{map[string]any{
			"id": "yd2-mini-video", "kind": "video", "modelId": "catalog-model", "credential": "catalog-key",
		}},
	}
	cfg := ResolveVideoProviderSetting(config, VideoProviderPersonalAPI)
	if cfg.APIKey != "saved-key" || cfg.Model != "saved-model" {
		t.Fatalf("explicit save should win, got %+v", cfg)
	}
}

func TestResolveVideoProviderSettingFallsBackToCatalog(t *testing.T) {
	config := map[string]any{
		"modelCatalogVersion": 1,
		"modelCatalog": []any{map[string]any{
			"id": "yd2-mini-video", "kind": "video", "modelId": "catalog-model", "credential": "catalog-key",
		}},
	}
	cfg := ResolveVideoProviderSetting(config, VideoProviderPersonalAPI)
	if cfg.APIKey != "catalog-key" || cfg.Model != "catalog-model" || !cfg.IsConfigured() {
		t.Fatalf("catalog fallback mismatch: %+v", cfg)
	}
}

func TestResolveVideoProviderSettingDefaults(t *testing.T) {
	cfg := ResolveVideoProviderSetting(nil, VideoProviderAutoDLH3)
	if cfg.Model != AutoDLH3Model || cfg.IsConfigured() {
		t.Fatalf("default mismatch: %+v", cfg)
	}
}

func TestPersistentVideoProviderRegistryUnconfiguredMessage(t *testing.T) {
	registry := NewPersistentVideoProviderRegistry(newFakeOwnerConfigStore())
	_, err := registry.Resolve(context.Background(), "dave", VideoProviderPersonalAPI)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("expected ErrUnavailable, got %v", err)
	}
	if msg := err.Error(); msg != ErrUnavailable.Error()+": 视频生成通道未配置；请到「批量工厂 → 设置 → 视频生成通道」选择通道并填写密钥后保存。" {
		t.Fatalf("unexpected guidance message: %q", msg)
	}
	// 附加错误码不能破坏原有的 errors.Is 语义，前端才能继续按码显示设置入口。
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("error code wrapper broke ErrUnavailable matching: %v", err)
	}
	var carrier interface{ ErrorCode() string }
	if !errors.As(err, &carrier) || carrier.ErrorCode() != ErrorCodeVideoProviderNotConfigured {
		t.Fatalf("missing stable error code: %v", err)
	}
}
