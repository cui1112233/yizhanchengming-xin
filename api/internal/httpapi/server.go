package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
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

type AuthService interface {
	Login(ctx context.Context, username, password string) (authn.Credentials, authn.User, error)
	AuthenticateAccess(ctx context.Context, token string) (authn.User, error)
	Refresh(ctx context.Context, token string) (authn.Credentials, authn.User, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
}

type Dependencies struct {
	Intakes         IntakeService
	Reader          IntakeReader
	BatchProjects   BatchProjectReader
	Pipeline        PipelineService
	Generation      GenerationService
	UnifiedSettings UnifiedSettingsService
	Auth            AuthService
	SecureCookies   bool
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

	mux.HandleFunc("POST /api/auth/login", api.login)
	mux.HandleFunc("POST /api/auth/refresh", api.refreshAuth)
	mux.Handle("GET /api/auth/current-user", api.requireAuth(http.HandlerFunc(api.currentUser)))
	mux.HandleFunc("POST /api/auth/logout", api.logout)

	mux.Handle("POST /api/v1/intakes", api.requireCapability(CapabilityBatchConfigure, http.HandlerFunc(api.createIntake)))
	mux.Handle("GET /api/v1/intakes", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.listIntakes)))
	mux.Handle("POST /api/v1/intakes/{id}/execute", api.requireCapability(CapabilityBatchExecute, http.HandlerFunc(api.executeIntake)))
	mux.Handle("GET /api/v1/intakes/{id}/books", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.listBooks)))
	mux.Handle("POST /api/v1/intakes/{id}/batch-projects", api.requireCapability(CapabilityBatchExecute, http.HandlerFunc(api.createBatchProject)))
	mux.Handle("GET /api/v1/batch-projects", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.listBatchProjects)))

	mux.Handle("GET /api/v1/batch-projects/{id}/settings", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.getUnifiedSettings)))
	mux.Handle("PUT /api/v1/batch-projects/{id}/settings/production", api.requireCapability(CapabilityBatchConfigure, http.HandlerFunc(api.saveProductionSettings)))
	mux.Handle("PUT /api/v1/batch-projects/{id}/settings/publishing", api.requireCapability(CapabilityPublishConfigure, http.HandlerFunc(api.savePublishingSettings)))
	mux.Handle("GET /api/v1/batch-projects/{id}/version-profile", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.getVersionProfile)))
	mux.Handle("PUT /api/v1/batch-projects/{id}/version-profile", api.requireCapability(CapabilityBatchConfigure, http.HandlerFunc(api.saveVersionProfile)))
	mux.Handle("POST /api/v1/batch-projects/{id}/version-profile/sync-121", api.requireCapability(CapabilityBatchConfigure, http.HandlerFunc(api.sync121Settings)))
	mux.Handle("POST /api/v1/batch-projects/{id}/version-profile/sync-style-types", api.requireCapability(CapabilityBatchConfigure, http.HandlerFunc(api.syncStyleTypes)))

	mux.Handle("GET /api/v1/batch-projects/{projectId}/generation", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.projectGeneration)))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/generation", api.requireCapability(CapabilityBatchExecute, http.HandlerFunc(api.projectGeneration)))
	mux.Handle("GET /api/v1/batch-projects/{projectId}/books/{bookId}/generation", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.bookGeneration)))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/generation", api.requireCapability(CapabilityBatchExecute, http.HandlerFunc(api.bookGeneration)))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/generation/stages/{stage}/retry", api.requireCapability(CapabilityBatchExecute, http.HandlerFunc(api.retryGenerationStage)))
	mux.Handle("GET /api/v1/batch-projects/{projectId}/books/{bookId}/generation/stages/{stage}", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.generationStage)))

	// System generation presets are public read-only data. User/project-private prompt APIs must use authenticated routes instead of broad whitelisting.
	mux.HandleFunc("GET /api/v1/generation/prompts", api.generationPrompts)
	return mux
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
