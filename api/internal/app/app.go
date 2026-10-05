package app

import (
	"database/sql"
	"encoding/base64"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/httpapi"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/publishing"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/unifiedsettings"
)

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
	publishingService := publishing.NewService(publishing.NewMySQLStore(db), publishing.Options{CredentialKey: publishingCredentialKey()})
	allowedOrigins := make([]string, 0)
	for _, value := range strings.Split(os.Getenv("QIANTIE_ALLOWED_ORIGINS"), ",") {
		if value = strings.TrimSpace(value); value != "" {
			allowedOrigins = append(allowedOrigins, value)
		}
	}

	return httpapi.NewHandler(httpapi.Dependencies{
		Intakes:             intakeService,
		Reader:              store,
		Pipeline:            pipelineService,
		BatchProjects:       store,
		BatchProjectDetails: store,
		Generation:          generationService,
		UnifiedSettings:     settingsService,
		Auth:                authService,
		Publishing:          publishingService,
		SecureCookies:       secureCookiesEnabled(),
		AllowedOrigins:      allowedOrigins,
	})
}

func secureCookiesEnabled() bool {
	environment := strings.ToLower(strings.TrimSpace(os.Getenv("QIANTIE_ENV")))
	development := environment == "dev" || environment == "development" || environment == "local" || environment == "test"
	if !development {
		return true
	}

	raw := strings.TrimSpace(os.Getenv("QIANTIE_COOKIE_SECURE"))
	if raw == "" {
		return false
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false
	}
	return value
}

func publishingCredentialKey() []byte {
	raw := strings.TrimSpace(os.Getenv("QIANTIE_PUBLISH_CREDENTIAL_KEY_B64"))
	if raw == "" {
		return nil
	}
	value, err := base64.StdEncoding.DecodeString(raw)
	if err != nil || len(value) != 32 {
		return nil
	}
	return value
}
