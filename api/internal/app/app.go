package app

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/httpapi"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/unifiedsettings"
)

// NewHandler wires production services to MySQL. Authentication is enabled by default;
// tests that only exercise business-store wiring use newHandler(..., false).
func NewHandler(db *sql.DB, fetcher intake.Fetcher, classifier intake.Classifier, now pipeline.Clock) http.Handler {
	return newHandler(db, fetcher, classifier, now, true)
}

func newHandler(db *sql.DB, fetcher intake.Fetcher, classifier intake.Classifier, now pipeline.Clock, authEnabled bool) http.Handler {
	store := intake.NewMySQLStore(db)
	intakeService := intake.NewService(store, fetcher, classifier)
	pipelineService := pipeline.NewService(store, now)
	generationStore := generation.NewMySQLStore(db)
	settingsStore := unifiedsettings.NewMySQLStore(db)

	var textProvider generation.Provider = generation.UnavailableProvider{}
	if baseURL, key, model := os.Getenv("QIANTIE_TEXT_API_BASE_URL"), os.Getenv("QIANTIE_TEXT_API_KEY"), os.Getenv("QIANTIE_TEXT_MODEL"); baseURL != "" && key != "" && model != "" {
		textProvider = generation.NewHTTPProvider(baseURL, key, model)
	}
	generationService := generation.NewService(generationStore, textProvider, nil)
	settingsService := unifiedsettings.NewService(settingsStore, unifiedsettings.StaticDefaults{Config: unifiedsettings.Settings{
		Production: map[string]any{"productionMode": "original", "aiCopyEnabled": false, "aiCopyCount": float64(1)},
		Publishing: map[string]any{"uploadVideoType": "merged", "materialReuse": false},
	}}, settingsStore)

	var authService httpapi.AuthService
	if authEnabled {
		authStore := authn.NewMySQLStore(db)
		authService = authn.NewService(authStore, authn.NewManager(authStore, authn.Options{}))
	}
	secureCookies, _ := strconv.ParseBool(os.Getenv("QIANTIE_COOKIE_SECURE"))

	return httpapi.NewHandler(httpapi.Dependencies{
		Intakes:         intakeService,
		Reader:          store,
		Pipeline:        pipelineService,
		BatchProjects:   store,
		Generation:      generationService,
		UnifiedSettings: settingsService,
		Auth:            authService,
		SecureCookies:   secureCookies,
	})
}
