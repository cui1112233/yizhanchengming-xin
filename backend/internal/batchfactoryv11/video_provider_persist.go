package batchfactoryv11

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// OwnerConfigStore is the durable per-owner settings document already used by
// the model catalog. Reusing it keeps video provider configuration on the same
// account config record instead of introducing another storage mechanism.
type OwnerConfigStore interface {
	GetConfig(context.Context, string) (map[string]any, error)
	PutConfig(context.Context, string, map[string]any) error
}

// OwnerConfigOwnerLister is implemented by the account config stores that can
// enumerate owners so persisted provider settings can be warmed at startup.
type OwnerConfigOwnerLister interface {
	ListConfigOwners(context.Context) ([]string, error)
}

// VideoProviderConfigPersister is implemented by registries that can store an
// explicitly saved provider configuration. Catalog-driven synchronization keeps
// using Put (in-memory) because the model catalog is already the durable source.
type VideoProviderConfigPersister interface {
	SaveConfig(context.Context, string, VideoProviderConfig) error
}

// PersistentVideoProviderRegistry keeps the fast in-memory registry used at
// runtime, while explicit provider settings are written to and hydrated from the
// owner config document. A restart therefore keeps the last saved provider.
type PersistentVideoProviderRegistry struct {
	mu      sync.Mutex
	cache   *MemoryVideoProviderRegistry
	configs OwnerConfigStore
}

func NewPersistentVideoProviderRegistry(configs OwnerConfigStore) *PersistentVideoProviderRegistry {
	return &PersistentVideoProviderRegistry{
		cache:   NewMemoryVideoProviderRegistry(),
		configs: configs,
	}
}

// Warm hydrates every owner's persisted provider config at startup. Failures are
// ignored by the caller because the registry also hydrates lazily on first use.
func (r *PersistentVideoProviderRegistry) Warm(ctx context.Context) {
	if r == nil || r.configs == nil {
		return
	}
	lister, ok := r.configs.(OwnerConfigOwnerLister)
	if !ok {
		return
	}
	owners, err := lister.ListConfigOwners(ctx)
	if err != nil {
		return
	}
	for _, owner := range owners {
		_ = r.hydrate(ctx, owner)
	}
}

func (r *PersistentVideoProviderRegistry) Put(_ context.Context, owner string, cfg VideoProviderConfig) error {
	if r == nil || r.cache == nil {
		return ErrUnavailable
	}
	return r.cache.Put(context.Background(), owner, cfg)
}

func (r *PersistentVideoProviderRegistry) SaveConfig(ctx context.Context, owner string, cfg VideoProviderConfig) error {
	if r == nil || r.configs == nil {
		return ErrUnavailable
	}
	normalized, err := normalizeVideoProviderConfig(cfg)
	if err != nil {
		return err
	}
	if err := r.persist(ctx, owner, normalized); err != nil {
		return err
	}
	r.cache.putNormalized(strings.TrimSpace(owner), normalized)
	return nil
}

func (r *PersistentVideoProviderRegistry) Resolve(ctx context.Context, owner, provider string) (VideoProviderConfig, error) {
	if r == nil || r.cache == nil {
		return VideoProviderConfig{}, ErrUnavailable
	}
	cfg, err := r.cache.Resolve(ctx, owner, provider)
	if err == nil || !errors.Is(err, ErrUnavailable) {
		return cfg, err
	}
	if hydrateErr := r.hydrate(ctx, owner); hydrateErr != nil {
		return VideoProviderConfig{}, hydrateErr
	}
	return r.cache.Resolve(ctx, owner, provider)
}

func (r *PersistentVideoProviderRegistry) View(ctx context.Context, owner, provider string) (VideoProviderConfigView, error) {
	cfg, err := r.Resolve(ctx, owner, provider)
	if err != nil {
		return VideoProviderConfigView{Provider: normalizeVideoProvider(provider)}, err
	}
	return VideoProviderConfigView{
		Provider:   cfg.Provider,
		Model:      cfg.Model,
		Configured: strings.TrimSpace(cfg.APIKey) != "" || cfg.Provider == VideoProviderDoubaoLocal,
	}, nil
}

func (r *PersistentVideoProviderRegistry) persist(ctx context.Context, owner string, cfg VideoProviderConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return fmt.Errorf("%w: owner is required", ErrInvalid)
	}
	config, err := r.configs.GetConfig(ctx, owner)
	if err != nil {
		return fmt.Errorf("读取账号配置失败: %w", err)
	}
	if config == nil {
		config = map[string]any{}
	}
	providers, _ := config["videoProviders"].(map[string]any)
	if providers == nil {
		providers = map[string]any{}
	}
	entry := map[string]any{
		"apiKey":     cfg.APIKey,
		"model":      cfg.Model,
		"createUrl":  cfg.CreateURL,
		"tasksUrl":   cfg.TasksURL,
		"resultUrl":  cfg.ResultURL,
		"workflowId": cfg.WorkflowID,
	}
	if existing, ok := providers[cfg.Provider].(map[string]any); ok && configEntriesEqual(existing, entry) {
		return nil
	}
	providers[cfg.Provider] = entry
	config["videoProviders"] = providers
	return r.configs.PutConfig(ctx, owner, config)
}

// hydrate loads persisted provider settings into the in-memory cache. Catalog
// synchronization may already have written a fresher value, so existing entries
// are never overwritten.
func (r *PersistentVideoProviderRegistry) hydrate(ctx context.Context, owner string) error {
	if r.configs == nil {
		return nil
	}
	owner = strings.TrimSpace(owner)
	if owner == "" {
		return nil
	}
	config, err := r.configs.GetConfig(ctx, owner)
	if err != nil {
		return fmt.Errorf("读取视频通道配置失败: %w", err)
	}
	providers, _ := config["videoProviders"].(map[string]any)
	for name, raw := range providers {
		cfg := persistedVideoProviderConfig(name, raw)
		if cfg.Provider == "" || r.cache.has(owner, cfg.Provider) {
			continue
		}
		r.cache.putNormalized(owner, cfg)
	}
	return nil
}

func persistedVideoProviderConfig(name string, raw any) VideoProviderConfig {
	entry, _ := raw.(map[string]any)
	return VideoProviderConfig{
		Provider:   normalizeVideoProvider(name),
		APIKey:     configString(entry, "apiKey"),
		Model:      configString(entry, "model"),
		CreateURL:  configString(entry, "createUrl"),
		TasksURL:   configString(entry, "tasksUrl"),
		ResultURL:  configString(entry, "resultUrl"),
		WorkflowID: configString(entry, "workflowId"),
	}
}

func configString(value map[string]any, key string) string {
	text, _ := value[key].(string)
	return strings.TrimSpace(text)
}

// configNumber accepts the numeric shapes that appear in hand-built maps and
// JSON-decoded account configs alike.
func configNumber(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	default:
		return 0
	}
}

func configEntriesEqual(left, right map[string]any) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}
