package app

import (
	"database/sql"
	"net/http"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/httpapi"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/unifiedsettings"
)

// NewHandler wires production services to MySQL. Unified settings are resolved
// server-side so React never becomes the configuration source of truth.
func NewHandler(db *sql.DB, fetcher intake.Fetcher, classifier intake.Classifier, now pipeline.Clock) http.Handler {
	store := intake.NewMySQLStore(db)
	intakeService := intake.NewService(store, fetcher, classifier)
	pipelineService := pipeline.NewService(store, now)
	settingsStore := unifiedsettings.NewMySQLStore(db)
	settingsService := unifiedsettings.NewService(settingsStore, unifiedsettings.StaticDefaults{Config: unifiedsettings.Settings{
		Production: map[string]any{"productionMode": "original", "aiCopyEnabled": false, "aiCopyCount": float64(1)},
		Publishing: map[string]any{"uploadVideoType": "merged", "materialReuse": false},
	}}, settingsStore)

	return httpapi.NewHandler(httpapi.Dependencies{
		Intakes: intakeService, Reader: store, Pipeline: pipelineService, BatchProjects: store,
		UnifiedSettings: settingsService,
	})
}
