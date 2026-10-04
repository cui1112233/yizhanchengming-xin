package videogen

import (
	"context"
	"errors"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/providerconfig"
)

type ProviderRecordSource interface {
	Resolve(context.Context, string, string, string) (providerconfig.Resolved, error)
}

type ConfigStoreSource struct { Store ProviderRecordSource }

func (s ConfigStoreSource) Resolve(ctx context.Context, owner, provider string) (Config, error) {
	if s.Store == nil { return Config{}, errors.New("video provider config store unavailable") }
	record, err := s.Store.Resolve(ctx, owner, providerconfig.KindVideo, NormalizeProvider(provider))
	if err != nil { return Config{}, err }
	if !record.Enabled { return Config{}, errors.New("video provider config is disabled") }
	return Config{
		Provider: record.Provider,
		APIKey: record.APIKey,
		Model: record.Model,
		CreateURL: record.CreateURL,
		TasksURL: record.TasksURL,
		ResultURL: record.ResultURL,
	}, nil
}
