package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/publishing"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/video"
)

type IntakeService interface {
	CreateIntake(ctx context.Context, input intake.CreateIntakeInput) (intake.Intake, []intake.Book, error)
	ExecuteIntake(ctx context.Context, intakeID int64, maxText int) (intake.ExecuteResult, error)
}

type IntakeReader interface {
	ListIntakes(context.Context) ([]intake.Intake, error)
	ListBooks(context.Context, int64) ([]intake.Book, error)
}

type BatchProjectReader interface {
	ListBatchProjects(context.Context) ([]intake.BatchProject, error)
}

type BatchProjectDetailReader interface {
	GetBatchProject(ctx context.Context, id int64) (intake.BatchProject, error)
	ListBooks(ctx context.Context, intakeID int64) ([]intake.Book, error)
}

type PipelineService interface {
	Create(context.Context, pipeline.CreateRequest) (pipeline.CreateResult, error)
}

type AuthService interface {
	Login(context.Context, string, string) (authn.Credentials, authn.User, error)
	AuthenticateAccess(context.Context, string) (authn.User, error)
	Refresh(context.Context, string) (authn.Credentials, authn.User, error)
	Logout(context.Context, string, string) error
}

type PublishingService interface {
	CreateAccount(context.Context, authn.User, publishing.CreateAccountInput) (publishing.Account, error)
	ListAccounts(context.Context, authn.User) ([]publishing.Account, error)
	ClaimBatchProject(context.Context, authn.User, int64) error
	CreateIntent(context.Context, authn.User, publishing.CreateIntentInput) (publishing.Intent, error)
	GetIntent(context.Context, authn.User, int64) (publishing.Intent, error)
	ListAudits(context.Context, authn.User, int64) ([]publishing.Audit, error)
}

type VideoLocalExecutorService interface {
	Register(context.Context, video.LocalExecutorRegistrationInput) (video.LocalExecutorRegistrationResult, error)
	Identity(context.Context, string) (video.LocalExecutorIdentity, error)
	Heartbeat(context.Context, string, video.LocalExecutorHeartbeatInput) error
	List(context.Context) ([]video.LocalExecutorIdentity, error)
	CompleteTask(context.Context, string, string, video.LocalExecutorCompleteInput) error
	FailTask(context.Context, string, string, video.LocalExecutorFailInput) error
}

type VideoStatusService interface {
	ProjectStatus(context.Context, int64) (video.ProjectVideoStatus, error)
}

type VideoMergeService interface {
	StartFromProductionTasks(context.Context, video.MergeProductionStartRequest) (video.MergeResult, error)
	Get(context.Context, int64) (video.MergeJob, []video.MergeAttempt, error)
	RetryAttempt(context.Context, int64) (video.MergeResult, error)
}

type Dependencies struct {
	Intakes                     IntakeService
	Reader                      IntakeReader
	BatchProjects               BatchProjectReader
	BatchProjectDetails         BatchProjectDetailReader
	Pipeline                    PipelineService
	Generation                  GenerationService
	UnifiedSettings             UnifiedSettingsService
	Auth                        AuthService
	Publishing                  PublishingService
	LoginLimiter                *LoginRateLimiter
	SecureCookies               bool
	AllowedOrigins              []string
	Video                       VideoService
	VideoConfig                 VideoConfigService
	VideoLocalExecutor          VideoLocalExecutorService
	VideoStatus                 VideoStatusService
	VideoMerge                  VideoMergeService
	BatchProjectAccess          BatchProjectAccessChecker
	VideoResourceProjects       VideoResourceProjectResolver
	VideoExecutorBootstrapToken string
}

func NewHandler(values ...Dependencies) http.Handler {
	var deps Dependencies
	if len(values) > 0 {
		deps = values[0]
	}
	if deps.Auth != nil && deps.LoginLimiter == nil {
		deps.LoginLimiter = NewLoginRateLimiter(LoginRateLimitOptions{})
	}
	api := handler{deps: deps}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.Handle("POST /api/auth/login", api.requireSameOrigin(http.HandlerFunc(api.login)))
	mux.Handle("POST /api/auth/refresh", api.requireSameOrigin(http.HandlerFunc(api.refreshAuth)))
	mux.Handle("GET /api/auth/current-user", api.requireAuth(http.HandlerFunc(api.currentUser)))
	mux.Handle("POST /api/auth/logout", api.requireSameOrigin(http.HandlerFunc(api.logout)))

	// Stage 1: Shuihuo/Batch Factory intake and project access.
	mux.Handle("POST /api/v1/intakes", api.requireSameOrigin(api.requireCapability(CapabilityBatchConfigure, http.HandlerFunc(api.createIntake))))
	mux.Handle("GET /api/v1/intakes", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.listIntakes)))
	mux.Handle("POST /api/v1/intakes/{id}/execute", api.requireSameOrigin(api.requireCapability(CapabilityBatchExecute, http.HandlerFunc(api.executeIntake))))
	mux.Handle("GET /api/v1/intakes/{id}/books", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.listBooks)))
	mux.Handle("POST /api/v1/intakes/{id}/batch-projects", api.requireSameOrigin(api.requireCapability(CapabilityBatchExecute, http.HandlerFunc(api.createOwnedBatchProject))))
	mux.Handle("GET /api/v1/batch-projects", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.listBatchProjects)))
	mux.Handle("GET /api/v1/batch-projects/{id}", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.getBatchProject)))

	// Stage 2: unified settings/version profile.
	mux.Handle("GET /api/v1/batch-projects/{id}/settings", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.getUnifiedSettings)))
	mux.Handle("PUT /api/v1/batch-projects/{id}/settings/production", api.requireSameOrigin(api.requireCapability(CapabilityBatchConfigure, http.HandlerFunc(api.saveProductionSettings))))
	mux.Handle("PUT /api/v1/batch-projects/{id}/settings/publishing", api.requireSameOrigin(api.requireCapability(CapabilityPublishConfigure, http.HandlerFunc(api.savePublishingSettings))))
	mux.Handle("GET /api/v1/batch-projects/{id}/version-profile", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.getVersionProfile)))
	mux.Handle("PUT /api/v1/batch-projects/{id}/version-profile", api.requireSameOrigin(api.requireCapability(CapabilityBatchConfigure, http.HandlerFunc(api.saveVersionProfile))))
	mux.Handle("POST /api/v1/batch-projects/{id}/version-profile/sync-121", api.requireSameOrigin(api.requireCapability(CapabilityBatchConfigure, http.HandlerFunc(api.sync121Settings))))
	mux.Handle("POST /api/v1/batch-projects/{id}/version-profile/sync-style-types", api.requireSameOrigin(api.requireCapability(CapabilityBatchConfigure, http.HandlerFunc(api.syncStyleTypes))))

	// Stage 3: generation reads and mutations.
	mux.Handle("GET /api/v1/batch-projects/{projectId}/generation", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.projectGeneration)))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/generation", api.requireSameOrigin(api.requireCapability(CapabilityBatchExecute, http.HandlerFunc(api.projectGeneration))))
	mux.Handle("GET /api/v1/batch-projects/{projectId}/books/{bookId}/generation", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.bookGeneration)))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/generation", api.requireSameOrigin(api.requireCapability(CapabilityBatchExecute, http.HandlerFunc(api.bookGeneration))))
	mux.Handle("GET /api/v1/batch-projects/{projectId}/books/{bookId}/audio-measurement", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.audioMeasurement)))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/audio-measurement", api.requireSameOrigin(api.requireCapability(CapabilityBatchExecute, http.HandlerFunc(api.audioMeasurement))))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/generation/stages/{stage}/retry", api.requireSameOrigin(api.requireCapability(CapabilityBatchExecute, http.HandlerFunc(api.retryGenerationStage))))
	mux.Handle("GET /api/v1/batch-projects/{projectId}/books/{bookId}/generation/stages/{stage}", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.generationStage)))

	// Only immutable system presets are public. Private/project prompts must be registered behind auth.
	mux.HandleFunc("GET /api/v1/generation/prompts", api.generationPrompts)

	// Stage 4: publishing authorization. Browser requests create server-side intent only.
	mux.Handle("POST /api/v1/publishing/accounts", api.requireSameOrigin(api.requireCapability(CapabilityPublishAccountConfigure, http.HandlerFunc(api.createPublishingAccount))))
	mux.Handle("GET /api/v1/publishing/accounts", api.requireCapability(CapabilityPublishAccountConfigure, http.HandlerFunc(api.listPublishingAccounts)))
	mux.Handle("POST /api/v1/publishing/intents", api.requireSameOrigin(api.requireCapability(CapabilityPublishExecute, http.HandlerFunc(api.createPublishIntent))))
	mux.Handle("GET /api/v1/publishing/intents/{id}", api.requireCapability(CapabilityPublishExecute, http.HandlerFunc(api.getPublishIntent)))
	mux.Handle("GET /api/v1/publishing/audits", api.requireCapability(CapabilityPublishAuditView, http.HandlerFunc(api.listPublishAudits)))

	// Task 14 browser-facing VIDEO APIs reuse Task 15 authentication, capability,
	// project ownership and same-origin semantics. Provider secrets are
	// configuration-only and therefore require batch.configure.
	mux.Handle("GET /api/v1/video-providers/{provider}/models/{model}", api.requireCapability(CapabilityBatchConfigure, http.HandlerFunc(api.getVideoProviderConfig)))
	mux.Handle("PUT /api/v1/video-providers/{provider}/models/{model}", api.requireSameOrigin(api.requireCapability(CapabilityBatchConfigure, http.HandlerFunc(api.putVideoProviderConfig))))
	mux.Handle("GET /api/v1/video-providers/{provider}/models/{model}/status", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.getVideoProviderStatus)))
	mux.Handle("GET /api/v1/batch-projects/{projectId}/video", api.requireCapability(CapabilityBatchView, api.requireVideoProjectAccess("projectId", http.HandlerFunc(api.projectVideoStatus))))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/video", api.requireSameOrigin(api.requireCapability(CapabilityBatchExecute, api.requireVideoProjectAccess("projectId", http.HandlerFunc(api.startVideo)))))
	mux.Handle("POST /api/v1/video-tasks/{taskId}/poll", api.requireSameOrigin(api.requireCapability(CapabilityBatchExecute, api.requireVideoTaskAccess(http.HandlerFunc(api.pollVideoTask)))))
	mux.Handle("POST /api/v1/video-tasks/{taskId}/cancel", api.requireSameOrigin(api.requireCapability(CapabilityBatchExecute, api.requireVideoTaskAccess(http.HandlerFunc(api.cancelVideoTask)))))
	mux.Handle("POST /api/v1/video-tasks/{taskId}/retry", api.requireSameOrigin(api.requireCapability(CapabilityBatchExecute, api.requireVideoTaskAccess(http.HandlerFunc(api.retryVideoTask)))))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/merge", api.requireSameOrigin(api.requireCapability(CapabilityBatchExecute, api.requireVideoProjectAccess("projectId", http.HandlerFunc(api.startVideoMerge)))))
	mux.Handle("GET /api/v1/video-merge-jobs/{jobId}", api.requireCapability(CapabilityBatchView, api.requireVideoMergeJobAccess(http.HandlerFunc(api.getVideoMerge))))
	mux.Handle("POST /api/v1/video-merge-attempts/{attemptId}/retry", api.requireSameOrigin(api.requireCapability(CapabilityBatchExecute, api.requireVideoMergeAttemptAccess(http.HandlerFunc(api.retryVideoMerge)))))

	// Local executor traffic has a distinct service identity. Registration uses a
	// server-side bootstrap token; heartbeat/identity/complete/fail continue to
	// authenticate with the executor's one-time-issued Bearer credential.
	mux.Handle("POST /api/v1/video/local-executors/register", api.requireVideoExecutorBootstrap(http.HandlerFunc(api.registerVideoLocalExecutor)))
	mux.Handle("GET /api/v1/video/local-executors", api.requireCapability(CapabilityBatchConfigure, http.HandlerFunc(api.listVideoLocalExecutors)))
	mux.HandleFunc("GET /api/v1/video/local-executors/me", api.getVideoLocalExecutorIdentity)
	mux.HandleFunc("POST /api/v1/video/local-executors/heartbeat", api.heartbeatVideoLocalExecutor)
	mux.HandleFunc("POST /api/v1/video/local-executor-tasks/{taskId}/complete", api.completeVideoLocalExecutorTask)
	mux.HandleFunc("POST /api/v1/video/local-executor-tasks/{taskId}/fail", api.failVideoLocalExecutorTask)
	return mux
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
