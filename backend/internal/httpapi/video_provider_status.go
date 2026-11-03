package httpapi

import (
	"context"
	"net/http"

	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/novelfetchworkshop"
)

type videoProviderConfigReader interface {
	GetConfig(context.Context, string) (map[string]any, error)
}

// registerVideoProviderStatusRoute reports the same effective setting that the
// production submit path uses. When the runtime registry is available it is
// synchronized from the account config and read through View, so the settings
// drawer and video submission cannot disagree. Early-slice setups without a
// production service keep the read-only account-config fallback.
func registerVideoProviderStatusRoute(mux *http.ServeMux, auth BridgeAuth, reader videoProviderConfigReader, registry batchfactoryv11.VideoProviderRegistry, store batchfactoryv11.Store, version string, catalogStores ...novelfetchworkshop.Store) {
	if mux == nil || reader == nil {
		return
	}
	mux.Handle("GET /api/batch-factory/"+version+"/video-provider/status", auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		owner, ok := bridgeOwner(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		config, err := reader.GetConfig(r.Context(), owner)
		if err != nil {
			writeStoreError(w, batchfactoryv11.ErrUnavailable)
			return
		}
		if len(catalogStores) > 0 {
			mergeManagedModelCatalog(r.Context(), config, catalogStores[0], owner)
		}
		provider := batchfactoryv11.NormalizeVideoProviderForHTTP(r.URL.Query().Get("provider"))
		if registry != nil {
			// Feed the registry from the exact same configuration the submit
			// path reads, then report the registry's view. The sync step only
			// needs the batch store; View also works from persisted settings.
			if store != nil {
				_ = syncShuihuoConfiguredVideoModels(r.Context(), store, registry, owner, catalogStores...)
			}
			if view, viewErr := registry.View(r.Context(), owner, provider); viewErr == nil && view.Configured {
				if view.Model == "" {
					view.Model = batchfactoryv11.DefaultVideoModelFor(provider)
				}
				writeJSON(w, http.StatusOK, view)
				return
			}
		}
		writeJSON(w, http.StatusOK, videoProviderStatus(config, provider))
	})))
}

// videoProviderStatus is the read-only fallback used before a production
// service exists. The derivation itself lives in batchfactoryv11 so the
// registry hydration and this response share one priority order.
func videoProviderStatus(config map[string]any, rawProvider string) map[string]any {
	cfg := batchfactoryv11.ResolveVideoProviderSetting(config, rawProvider)
	return map[string]any{"provider": cfg.Provider, "model": cfg.Model, "configured": cfg.IsConfigured()}
}
