package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
)

type IntakeService interface {
	CreateIntake(ctx context.Context, input intake.CreateIntakeInput) (intake.Intake, []intake.Book, error)
	ExecuteIntake(ctx context.Context, intakeID int64, maxText int) (intake.ExecuteResult, error)
}

type IntakeReader interface {
	ListIntakes(ctx context.Context) ([]intake.Intake, error)
	ListBooks(ctx context.Context, intakeID int64) ([]intake.Book, error)
}

type BatchProjectReader interface {
	ListBatchProjects(ctx context.Context) ([]intake.BatchProject, error)
}

type PipelineService interface {
	Create(ctx context.Context, request pipeline.CreateRequest) (pipeline.CreateResult, error)
}

type Dependencies struct {
	Intakes         IntakeService
	Reader          IntakeReader
	BatchProjects   BatchProjectReader
	Pipeline        PipelineService
	Generation      GenerationService
	UnifiedSettings UnifiedSettingsService
}

func NewHandler(values ...Dependencies) http.Handler {
	var deps Dependencies
	if len(values) > 0 {
		deps = values[0]
	}

	api := handler{deps: deps}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /api/v1/intakes", api.createIntake)
	mux.HandleFunc("GET /api/v1/intakes", api.listIntakes)
	mux.HandleFunc("POST /api/v1/intakes/{id}/execute", api.executeIntake)
	mux.HandleFunc("GET /api/v1/intakes/{id}/books", api.listBooks)
	mux.HandleFunc("POST /api/v1/intakes/{id}/batch-projects", api.createBatchProject)
	mux.HandleFunc("GET /api/v1/batch-projects", api.listBatchProjects)

	mux.HandleFunc("GET /api/v1/batch-projects/{id}/settings", api.getUnifiedSettings)
	mux.HandleFunc("PUT /api/v1/batch-projects/{id}/settings/production", api.saveProductionSettings)
	mux.HandleFunc("PUT /api/v1/batch-projects/{id}/settings/publishing", api.savePublishingSettings)
	mux.HandleFunc("GET /api/v1/batch-projects/{id}/version-profile", api.getVersionProfile)
	mux.HandleFunc("PUT /api/v1/batch-projects/{id}/version-profile", api.saveVersionProfile)
	mux.HandleFunc("POST /api/v1/batch-projects/{id}/version-profile/sync-121", api.sync121Settings)
	mux.HandleFunc("POST /api/v1/batch-projects/{id}/version-profile/sync-style-types", api.syncStyleTypes)

	mux.HandleFunc("GET /api/v1/batch-projects/{projectId}/generation", api.projectGeneration)
	mux.HandleFunc("POST /api/v1/batch-projects/{projectId}/generation", api.projectGeneration)
	mux.HandleFunc("GET /api/v1/batch-projects/{projectId}/books/{bookId}/generation", api.bookGeneration)
	mux.HandleFunc("POST /api/v1/batch-projects/{projectId}/books/{bookId}/generation", api.bookGeneration)
	mux.HandleFunc("GET /api/v1/batch-projects/{projectId}/books/{bookId}/audio-measurement", api.audioMeasurement)
	mux.HandleFunc("POST /api/v1/batch-projects/{projectId}/books/{bookId}/audio-measurement", api.audioMeasurement)
	mux.HandleFunc("POST /api/v1/batch-projects/{projectId}/books/{bookId}/generation/stages/{stage}/retry", api.retryGenerationStage)
	mux.HandleFunc("GET /api/v1/batch-projects/{projectId}/books/{bookId}/generation/stages/{stage}", api.generationStage)
	mux.HandleFunc("GET /api/v1/generation/prompts", api.generationPrompts)
	return mux
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
