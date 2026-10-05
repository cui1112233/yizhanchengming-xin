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
}

func NewConfigService(store ProviderConfigStore, masterKey []byte) *ConfigService {
	return &ConfigService{store: store, masterKey: append([]byte(nil), masterKey...)}
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
