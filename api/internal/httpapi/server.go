package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/video"
)

type IntakeService interface { CreateIntake(context.Context, intake.CreateIntakeInput) (intake.Intake, []intake.Book, error); ExecuteIntake(context.Context, int64, int) (intake.ExecuteResult, error) }
type IntakeReader interface { ListIntakes(context.Context) ([]intake.Intake, error); ListBooks(context.Context, int64) ([]intake.Book, error) }
type BatchProjectReader interface { ListBatchProjects(context.Context) ([]intake.BatchProject, error) }
type BatchProjectDetailReader interface { GetBatchProject(context.Context, int64) (intake.BatchProject, error); ListBooks(context.Context, int64) ([]intake.Book, error) }
type PipelineService interface { Create(context.Context, pipeline.CreateRequest) (pipeline.CreateResult, error) }
type VideoLocalExecutorService interface {
	Register(context.Context, video.LocalExecutorRegistrationInput) (video.LocalExecutorRegistrationResult, error)
	Identity(context.Context, string) (video.LocalExecutorIdentity, error)
	Heartbeat(context.Context, string, video.LocalExecutorHeartbeatInput) error
	List(context.Context) ([]video.LocalExecutorIdentity, error)
	CompleteTask(context.Context, string, string, video.LocalExecutorCompleteInput) error
	FailTask(context.Context, string, string, video.LocalExecutorFailInput) error
}
type VideoStatusService interface { ProjectStatus(context.Context, int64) (video.ProjectVideoStatus, error) }
type VideoMergeService interface {
	StartFromProductionTasks(context.Context, video.MergeProductionStartRequest) (video.MergeResult, error)
	Get(context.Context, int64) (video.MergeJob, []video.MergeAttempt, error)
	RetryAttempt(context.Context, int64) (video.MergeResult, error)
}

type Dependencies struct {
	Intakes IntakeService; Reader IntakeReader; BatchProjects BatchProjectReader; BatchProjectDetails BatchProjectDetailReader
	Pipeline PipelineService; Generation GenerationService; UnifiedSettings UnifiedSettingsService
	Video VideoService; VideoConfig VideoConfigService; VideoLocalExecutor VideoLocalExecutorService; VideoStatus VideoStatusService; VideoMerge VideoMergeService
}

func NewHandler(values ...Dependencies) http.Handler {
	var deps Dependencies
	if len(values) > 0 { deps = values[0] }
	api := handler{deps: deps}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, http.StatusOK, map[string]string{"status":"ok"}) })
	mux.HandleFunc("POST /api/v1/intakes", api.createIntake); mux.HandleFunc("GET /api/v1/intakes", api.listIntakes)
	mux.HandleFunc("POST /api/v1/intakes/{id}/execute", api.executeIntake); mux.HandleFunc("GET /api/v1/intakes/{id}/books", api.listBooks)
	mux.HandleFunc("POST /api/v1/intakes/{id}/batch-projects", api.createBatchProject); mux.HandleFunc("GET /api/v1/batch-projects", api.listBatchProjects); mux.HandleFunc("GET /api/v1/batch-projects/{id}", api.getBatchProject)
	mux.HandleFunc("GET /api/v1/batch-projects/{id}/settings", api.getUnifiedSettings); mux.HandleFunc("PUT /api/v1/batch-projects/{id}/settings/production", api.saveProductionSettings); mux.HandleFunc("PUT /api/v1/batch-projects/{id}/settings/publishing", api.savePublishingSettings)
	mux.HandleFunc("GET /api/v1/batch-projects/{id}/version-profile", api.getVersionProfile); mux.HandleFunc("PUT /api/v1/batch-projects/{id}/version-profile", api.saveVersionProfile); mux.HandleFunc("POST /api/v1/batch-projects/{id}/version-profile/sync-121", api.sync121Settings); mux.HandleFunc("POST /api/v1/batch-projects/{id}/version-profile/sync-style-types", api.syncStyleTypes)
	mux.HandleFunc("GET /api/v1/batch-projects/{projectId}/generation", api.projectGeneration); mux.HandleFunc("POST /api/v1/batch-projects/{projectId}/generation", api.projectGeneration); mux.HandleFunc("GET /api/v1/batch-projects/{projectId}/books/{bookId}/generation", api.bookGeneration); mux.HandleFunc("POST /api/v1/batch-projects/{projectId}/books/{bookId}/generation", api.bookGeneration)
	mux.HandleFunc("POST /api/v1/batch-projects/{projectId}/books/{bookId}/generation/stages/{stage}/retry", api.retryGenerationStage); mux.HandleFunc("GET /api/v1/batch-projects/{projectId}/books/{bookId}/generation/stages/{stage}", api.generationStage); mux.HandleFunc("GET /api/v1/generation/prompts", api.generationPrompts)
	mux.HandleFunc("GET /api/v1/video-providers/{provider}/models/{model}", api.getVideoProviderConfig); mux.HandleFunc("PUT /api/v1/video-providers/{provider}/models/{model}", api.putVideoProviderConfig); mux.HandleFunc("GET /api/v1/video-providers/{provider}/models/{model}/status", api.getVideoProviderStatus)
	mux.HandleFunc("GET /api/v1/batch-projects/{projectId}/video", api.projectVideoStatus); mux.HandleFunc("POST /api/v1/batch-projects/{projectId}/books/{bookId}/video", api.startVideo); mux.HandleFunc("POST /api/v1/video-tasks/{taskId}/poll", api.pollVideoTask); mux.HandleFunc("POST /api/v1/video-tasks/{taskId}/cancel", api.cancelVideoTask); mux.HandleFunc("POST /api/v1/video-tasks/{taskId}/retry", api.retryVideoTask)
	mux.HandleFunc("POST /api/v1/video/local-executors/register", api.registerVideoLocalExecutor); mux.HandleFunc("GET /api/v1/video/local-executors", api.listVideoLocalExecutors); mux.HandleFunc("GET /api/v1/video/local-executors/me", api.getVideoLocalExecutorIdentity); mux.HandleFunc("POST /api/v1/video/local-executors/heartbeat", api.heartbeatVideoLocalExecutor); mux.HandleFunc("POST /api/v1/video/local-executor-tasks/{taskId}/complete", api.completeVideoLocalExecutorTask); mux.HandleFunc("POST /api/v1/video/local-executor-tasks/{taskId}/fail", api.failVideoLocalExecutorTask)
	mux.HandleFunc("POST /api/v1/batch-projects/{projectId}/books/{bookId}/merge", api.startVideoMerge); mux.HandleFunc("GET /api/v1/video-merge-jobs/{jobId}", api.getVideoMerge); mux.HandleFunc("POST /api/v1/video-merge-attempts/{attemptId}/retry", api.retryVideoMerge)
	return mux
}

func writeJSON(w http.ResponseWriter, status int, value any) { w.Header().Set("Content-Type", "application/json; charset=utf-8"); w.WriteHeader(status); _ = json.NewEncoder(w).Encode(value) }
