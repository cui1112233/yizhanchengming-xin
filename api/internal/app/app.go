package app

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/httpapi"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

// NewHandler wires intake, pipeline and Task 12 generation services to MySQL.
// Prompt/Stage facts never live in React/localStorage.
func NewHandler(db *sql.DB, fetcher intake.Fetcher, classifier intake.Classifier, now pipeline.Clock) http.Handler {
	store := intake.NewMySQLStore(db)
	intakeService := intake.NewService(store, fetcher, classifier)
	pipelineService := pipeline.NewService(store, now)
	generationStore := generation.NewMySQLStore(db)

	var textProvider generation.Provider = generation.UnavailableProvider{}
	if baseURL, key, model := os.Getenv("QIANTIE_TEXT_API_BASE_URL"), os.Getenv("QIANTIE_TEXT_API_KEY"), os.Getenv("QIANTIE_TEXT_MODEL"); baseURL != "" && key != "" && model != "" {
		textProvider = generation.NewHTTPProvider(baseURL, key, model)
	}
	generationService := generation.NewService(generationStore, textProvider, nil)

	return httpapi.NewHandler(httpapi.Dependencies{
		Intakes:       intakeService,
		Reader:        store,
		Pipeline:      pipelineService,
		BatchProjects: store,
		Generation:    generationService,
	})
}
