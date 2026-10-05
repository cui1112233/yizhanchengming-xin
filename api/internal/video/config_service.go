package video

import (
	"context"
	"errors"
	"strings"
)

type ProviderConfigStore interface {
	GetProviderConfig(context.Context, string, string) (ProviderConfig, error)
	UpsertProviderConfig(context.Context, ProviderConfig) error
}

type ProviderConfigInput struct {
	ProviderKey string `json:"providerKey"`
	Model       string `json:"model"`
	CreateURL   string `json:"createUrl"`
	TasksURL    string `json:"tasksUrl"`
	ResultURL   string `json:"resultUrl"`
	Secret      string `json:"secret"`
	Enabled     bool   `json:"enabled"`
}

type ConfigService struct {
	store     ProviderConfigStore
	masterKey []byte
	providers ProviderFactory
}

func NewConfigService(store ProviderConfigStore, masterKey []byte) *ConfigService {
	return NewConfigServiceWithProviders(store, masterKey, nil)
}

func NewConfigServiceWithProviders(store ProviderConfigStore, masterKey []byte, providers ProviderFactory) *ConfigService {
	return &ConfigService{store: store, masterKey: append([]byte(nil), masterKey...), providers: providers}
}

func (s *ConfigService) Get(ctx context.Context, provider, model string) (ProviderConfigView, error) {
	cfg, err := s.store.GetProviderConfig(ctx, provider, model)
	if err != nil {
		return ProviderConfigView{}, err
	}
	return cfg.View(), nil
}

func (s *ConfigService) Save(ctx context.Context, input ProviderConfigInput) (ProviderConfigView, error) {
	input.ProviderKey = strings.TrimSpace(input.ProviderKey)
	input.Model = strings.TrimSpace(input.Model)
	mapped, ok := ProviderForModel(input.Model)
	if !ok || mapped != input.ProviderKey {
		return ProviderConfigView{}, providerError(ErrorProviderUnavailable, "provider/model mapping mismatch", nil)
	}
	cfg := ProviderConfig{
		ProviderKey: input.ProviderKey,
		Model: input.Model,
		CreateURL: strings.TrimSpace(input.CreateURL),
		TasksURL: strings.TrimSpace(input.TasksURL),
		ResultURL: strings.TrimSpace(input.ResultURL),
		Enabled: input.Enabled,
	}
	if strings.TrimSpace(input.Secret) != "" {
		ciphertext, nonce, err := EncryptSecret(s.masterKey, strings.TrimSpace(input.Secret))
		if err != nil {
			return ProviderConfigView{}, err
		}
		cfg.EncryptedSecret, cfg.SecretNonce = ciphertext, nonce
	} else {
		existing, err := s.store.GetProviderConfig(ctx, input.ProviderKey, input.Model)
		if err != nil {
			var providerErr *ProviderError
			if !errors.As(err, &providerErr) || providerErr.Code != ErrorProviderUnconfigured {
				return ProviderConfigView{}, err
			}
			return ProviderConfigView{}, providerError(ErrorProviderUnconfigured, "provider secret is required", nil)
		}
		cfg.EncryptedSecret, cfg.SecretNonce = existing.EncryptedSecret, existing.SecretNonce
	}
	if err := s.store.UpsertProviderConfig(ctx, cfg); err != nil {
		return ProviderConfigView{}, err
	}
	stored, err := s.store.GetProviderConfig(ctx, input.ProviderKey, input.Model)
	if err != nil {
		return ProviderConfigView{}, err
	}
	return stored.View(), nil
}

func (s *ConfigService) Status(ctx context.Context, provider, model string) (ProviderStatusView, error) {
	base := ProviderStatusView{ProviderKey: strings.TrimSpace(provider), Model: strings.TrimSpace(model), Status: ProviderStatusUnconfigured}
	if s == nil || s.store == nil {
		base.Status = ProviderStatusUnavailable
		base.Message = "provider configuration service unavailable"
		return base, nil
	}
	cfg, err := s.store.GetProviderConfig(ctx, base.ProviderKey, base.Model)
	if err != nil {
		var providerErr *ProviderError
		if errors.As(err, &providerErr) && providerErr.Code == ErrorProviderUnconfigured {
			return base, nil
		}
		return ProviderStatusView{}, err
	}
	base.Configured = len(cfg.EncryptedSecret) > 0 && len(cfg.SecretNonce) > 0
	base.Enabled = cfg.Enabled
	if !cfg.Enabled || !base.Configured {
		return base, nil
	}
	if s.providers == nil {
		base.Status = ProviderStatusUnavailable
		base.Message = "provider adapter unavailable"
		return base, nil
	}
	secret, err := DecryptSecret(s.masterKey, cfg.EncryptedSecret, cfg.SecretNonce)
	if err != nil {
		base.Status = ProviderStatusUnavailable
		base.Message = "provider credential unavailable"
		return base, nil
	}
	adapter, err := s.providers.Build(cfg, secret)
	if err != nil {
		base.Status, base.Message = availabilityFromError(err)
		return base, nil
	}
	prober, ok := adapter.(ProviderProber)
	if !ok {
		base.Status = ProviderStatusAvailable
		return base, nil
	}
	if err := prober.Probe(ctx); err != nil {
		base.Status, base.Message = availabilityFromError(err)
		return base, nil
	}
	base.Status = ProviderStatusAvailable
	return base, nil
}

func availabilityFromError(err error) (ProviderAvailability, string) {
	var providerErr *ProviderError
	if errors.As(err, &providerErr) {
		switch providerErr.Code {
		case ErrorProviderAuthFailed:
			return ProviderStatusAuthFailed, truncateError(providerErr.Message)
		case ErrorProviderUnconfigured:
			return ProviderStatusUnconfigured, truncateError(providerErr.Message)
		default:
			return ProviderStatusUnavailable, truncateError(providerErr.Message)
		}
	}
	return ProviderStatusUnavailable, "provider unavailable"
}
