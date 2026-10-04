package videogen

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

var ErrProviderUnavailable = errors.New("video provider unavailable")

type ConfigSource interface {
	Resolve(context.Context, string, string) (Config, error)
}

type Factory struct {
	Configs     ConfigSource
	HTTPClient  *http.Client
	ValidateURL URLValidator
	Local       Provider
}

func (f Factory) ForModel(ctx context.Context, owner, model, fallback string) (Provider, string, error) {
	if f.Configs == nil { return nil, "", errors.New("video provider config source unavailable") }
	owner = strings.TrimSpace(owner)
	if owner == "" { return nil, "", errors.New("owner is required") }
	providerName := ProviderForModel(model, fallback)
	cfg, err := f.Configs.Resolve(ctx, owner, providerName)
	if err != nil { return nil, providerName, err }
	cfg.Provider = providerName
	cfg, err = NormalizeConfig(cfg)
	if err != nil { return nil, providerName, err }
	if !ModelMatchesProviderModel(model, cfg.Model, providerName) {
		return nil, providerName, errors.New("selected model does not match configured video provider")
	}
	switch providerName {
	case ProviderPersonalAPI:
		return PersonalProvider{Config:cfg, Client:f.HTTPClient, ValidateURL:f.ValidateURL}, providerName, nil
	case ProviderYFAISeedance:
		return YFAIProvider{Config:cfg, Client:f.HTTPClient, ValidateURL:f.ValidateURL}, providerName, nil
	case ProviderAutoDLH3:
		return AutoDLProvider{Config:cfg, Client:f.HTTPClient, ValidateURL:f.ValidateURL}, providerName, nil
	case ProviderDoubaoLocal:
		if f.Local == nil { return nil, providerName, ErrProviderUnavailable }
		return f.Local, providerName, nil
	default:
		return nil, providerName, ErrProviderUnavailable
	}
}
