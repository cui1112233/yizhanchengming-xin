package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/novelpanel"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/publishing"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/shuihuo"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/task9runtime"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/video"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/workshop"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/workspace"
)

type IntakeService interface {
	CreateIntake(ctx context.Context, input intake.CreateIntakeInput) (intake.Intake, []intake.Book, error)
	ExecuteIntake(ctx context.Context, intakeID int64, maxText int) (intake.ExecuteResult, error)
	RestoreBook(ctx context.Context, intakeID, bookID int64, maxText int) (intake.Book, error)
}

type IntakeReader interface {
	ListIntakes(context.Context) ([]intake.Intake, error)
	ListBooks(context.Context, int64) ([]intake.Book, error)
}

type WorkshopService interface {
	Snapshot(context.Context, int64) (workshop.Snapshot, error)
	Save(context.Context, int64, json.RawMessage) (json.RawMessage, error)
}
type NovelPanelService interface {
	GetWorkspace(context.Context, int64) (novelpanel.Workspace, error)
	Save(context.Context, novelpanel.SaveRequest) (novelpanel.SaveResult, error)
	ListHistory(context.Context, int64, int) ([]novelpanel.HistoryRecord, error)
	Restore(context.Context, novelpanel.RestoreRequest) (novelpanel.SaveResult, error)
	PrepareStoryboardRequest(novelpanel.Workspace) (novelpanel.StoryboardRequest, error)
}

type BatchProjectReader interface {
	ListBatchProjects(context.Context, intake.BatchProjectListQuery) (intake.BatchProjectPage, error)
}

type WorkspaceRecentReader interface {
	ListRecent(context.Context, int64, int64, int) ([]workspace.RecentItem, error)
	ListRecentElevated(context.Context, int) ([]workspace.RecentItem, error)
}

type WorkspaceHistoryReader interface {
	ListHistory(context.Context, workspace.HistoryQuery) (workspace.HistoryPage, error)
}

type BatchProjectDetailReader interface {
	GetBatchProject(ctx context.Context, id int64) (intake.BatchProject, error)
	ListBooks(ctx context.Context, intakeID int64) ([]intake.Book, error)
}

type BatchProjectLifecycle interface {
	ArchiveBatchProject(context.Context, int64, int64) error
	RestoreBatchProject(context.Context, int64) error
	IsBatchProjectArchived(context.Context, int64) (bool, error)
	IsIntakeBatchProjectArchived(context.Context, int64) (bool, error)
}

type ScriptBookEditor interface {
	UpdateBookOriginalText(context.Context, int64, int64, string) (intake.Book, error)
}
type ScriptStoryboardService interface {
	Storyboard(context.Context, int64, int64) (generation.StoryboardDocument, error)
	SaveStoryboardCard(context.Context, int64, int64, generation.SaveStoryboardCardRequest) (generation.StoryboardDocument, error)
	DeleteStoryboardCard(context.Context, int64, int64, int64, int) (generation.StoryboardDocument, error)
	ReorderStoryboard(context.Context, int64, int64, []int64, int) (generation.StoryboardDocument, error)
	RecompileStoryboard(context.Context, int64, int64, string) (generation.BookGenerationResult, error)
}

type GenerationRuntimeService interface {
	Ready() bool
	AdmitGeneration(context.Context, task9runtime.GenerationRequest) (task9runtime.AdmissionResult, error)
	GenerationRun(context.Context, int64, int64) (task9runtime.GenerationRunStatus, error)
}

type RuntimeDiagnosticsStatus struct {
	Ready      bool
	Status     string
	Redis      string
	ReasonCode string
}

type RuntimeDiagnosticsReader interface {
	RuntimeDiagnostics() RuntimeDiagnosticsStatus
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
	CreatePairingIntent(context.Context, int64) (video.LocalExecutorPairingResult, error)
	RedeemPairingIntent(context.Context, string, video.LocalExecutorRegistrationInput) (video.LocalExecutorRegistrationResult, error)
	ListForOwner(context.Context, int64) ([]video.LocalExecutorIdentity, error)
	UnbindForOwner(context.Context, int64, string) error
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
type ShuihuoMediaService interface {
	CreateSegment(context.Context, shuihuo.CreateSegmentInput) (shuihuo.Segment, error)
	UpdateSegment(context.Context, int64, int64, int64, shuihuo.UpdateSegmentInput) (shuihuo.Segment, error)
	ListSegments(context.Context, int64, int64) ([]shuihuo.Segment, error)
	CreateAsset(context.Context, shuihuo.CreateAssetInput) (shuihuo.Asset, error)
	ListAssets(context.Context, int64, int64, int64) ([]shuihuo.Asset, error)
	CreateMediaTask(context.Context, shuihuo.CreateMediaTaskInput) (shuihuo.MediaTask, error)
	ListMediaTasks(context.Context, int64, int64) ([]shuihuo.MediaTask, error)
	ReorderSegments(context.Context, int64, int64, []int64) ([]shuihuo.Segment, error)
	ListCandidates(context.Context, int64, int64, int64) ([]shuihuo.Candidate, error)
	SelectCandidate(context.Context, int64, int64, int64, int64) (shuihuo.Candidate, error)
	RetryMediaTask(context.Context, int64, int64, int64) (shuihuo.MediaTask, error)
	UploadAsset(context.Context, shuihuo.UploadAssetInput) (shuihuo.Asset, error)
	OpenAsset(context.Context, int64, int64, int64) (shuihuo.Asset, io.ReadCloser, error)
}

type Dependencies struct {
	Intakes                     IntakeService
	Reader                      IntakeReader
	BatchProjects               BatchProjectReader
	WorkspaceRecent             WorkspaceRecentReader
	WorkspaceHistory            WorkspaceHistoryReader
	BatchProjectDetails         BatchProjectDetailReader
	ScriptBooks                 ScriptBookEditor
	ScriptStoryboards           ScriptStoryboardService
	Pipeline                    PipelineService
	Generation                  GenerationService
	GenerationRuntime           GenerationRuntimeService
	RuntimeDiagnostics          RuntimeDiagnosticsReader
	AdminPrompts                AdminPromptService
	Workshop                    WorkshopService
	NovelPanel                  NovelPanelService
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
	ShuihuoMedia                ShuihuoMediaService
	IntakeAccess                IntakeAccessChecker
	BatchProjectAccess          BatchProjectAccessChecker
	BatchProjectLifecycle       BatchProjectLifecycle
	VideoResourceProjects       VideoResourceProjectResolver
	VideoExecutorBootstrapToken string
	Database                    *sql.DB
	Logger                      *slog.Logger
	StartedAt                   time.Time
	AppInitialized              bool
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
	batchMutation := func(capability, pathKey string, next http.Handler) http.Handler {
		return api.requireSameOrigin(api.requireCapability(capability, api.requireBatchProjectAccess(pathKey, api.requireActiveBatchProject(pathKey, next))))
	}
	intakeMutation := func(capability string, next http.Handler) http.Handler {
		return api.requireSameOrigin(api.requireCapability(capability, api.requireIntakeAccess("id", api.requireActiveIntakeBatchProject("id", next))))
	}
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", api.readyz)
	mux.Handle("GET /api/v1/diagnostics", api.requireOperationsAdmin(http.HandlerFunc(api.diagnostics)))
	mux.Handle("GET /api/v1/history", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.listWorkspaceHistory)))
	mux.Handle("GET /api/v1/workspace/recent", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.listWorkspaceRecent)))
	mux.Handle("GET /api/v1/workspace/settings", api.requireAuth(http.HandlerFunc(api.workspaceSettings)))
	mux.Handle("PUT /api/v1/workspace/settings", api.requireSameOrigin(api.requireAuth(http.HandlerFunc(api.workspaceSettings))))
	mux.Handle("GET /api/v1/account/profile", api.requireAuth(http.HandlerFunc(api.accountProfile)))
	mux.Handle("PUT /api/v1/account/profile", api.requireSameOrigin(api.requireAuth(http.HandlerFunc(api.accountProfile))))
	mux.Handle("GET /api/v1/member-center", api.requireAuth(http.HandlerFunc(api.memberCenter)))

	mux.Handle("POST /api/auth/login", api.requireSameOrigin(http.HandlerFunc(api.login)))
	mux.Handle("POST /api/auth/refresh", api.requireSameOrigin(http.HandlerFunc(api.refreshAuth)))
	mux.Handle("GET /api/auth/current-user", api.requireAuth(http.HandlerFunc(api.currentUser)))
	mux.Handle("POST /api/auth/logout", api.requireSameOrigin(http.HandlerFunc(api.logout)))

	// Stage 1: Shuihuo/Batch Factory intake and project access.
	mux.Handle("POST /api/v1/intakes", api.requireSameOrigin(api.requireCapability(CapabilityBatchConfigure, http.HandlerFunc(api.createIntake))))
	mux.Handle("GET /api/v1/intakes", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.listIntakes)))
	mux.Handle("POST /api/v1/intakes/{id}/execute", intakeMutation(CapabilityBatchExecute, http.HandlerFunc(api.executeIntake)))
	mux.Handle("GET /api/v1/intakes/{id}/books", api.requireCapability(CapabilityBatchView, api.requireIntakeAccess("id", http.HandlerFunc(api.listBooks))))
	mux.Handle("GET /api/v1/intakes/{id}/workshop", api.requireCapability(CapabilityBatchView, api.requireIntakeAccess("id", http.HandlerFunc(api.getWorkshop))))
	mux.Handle("PUT /api/v1/intakes/{id}/workshop", intakeMutation(CapabilityBatchConfigure, http.HandlerFunc(api.saveWorkshop)))
	mux.Handle("POST /api/v1/intakes/{id}/books/{bookId}/restore", intakeMutation(CapabilityBatchExecute, http.HandlerFunc(api.restoreWorkshopBook)))
	mux.Handle("POST /api/v1/intakes/{id}/batch-projects", intakeMutation(CapabilityBatchExecute, http.HandlerFunc(api.createOwnedBatchProject)))
	mux.Handle("GET /api/v1/batch-projects", api.requireCapability(CapabilityBatchView, http.HandlerFunc(api.listBatchProjects)))
	// SECURITY: every object route carrying {id}/{projectId} below must put
	// requireBatchProjectAccess inside its capability check, so callers cannot
	// use ownership responses as an oracle and business services never see a
	// foreign project. Resource-ID routes resolve their project separately.
	mux.Handle("GET /api/v1/batch-projects/{id}", api.requireCapability(CapabilityBatchView, api.requireBatchProjectAccess("id", http.HandlerFunc(api.getBatchProject))))
	mux.Handle("POST /api/v1/batch-projects/{id}/archive", api.requireSameOrigin(api.requireCapability(CapabilityBatchConfigure, api.requireBatchProjectAccess("id", http.HandlerFunc(api.archiveBatchProject)))))
	mux.Handle("POST /api/v1/batch-projects/{id}/restore", api.requireSameOrigin(api.requireCapability(CapabilityBatchConfigure, api.requireBatchProjectAccess("id", http.HandlerFunc(api.restoreBatchProject)))))
	mux.Handle("PUT /api/v1/batch-projects/{projectId}/books/{bookId}/original-text", batchMutation(CapabilityBatchConfigure, "projectId", http.HandlerFunc(api.saveScriptOriginalText)))
	mux.Handle("GET /api/v1/batch-projects/{projectId}/books/{bookId}/storyboard", api.requireCapability(CapabilityBatchView, api.requireBatchProjectAccess("projectId", http.HandlerFunc(api.scriptStoryboard))))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/storyboard/cards", batchMutation(CapabilityBatchConfigure, "projectId", http.HandlerFunc(api.scriptStoryboard)))
	mux.Handle("PUT /api/v1/batch-projects/{projectId}/books/{bookId}/storyboard/cards/{cardId}", batchMutation(CapabilityBatchConfigure, "projectId", http.HandlerFunc(api.scriptStoryboard)))
	mux.Handle("DELETE /api/v1/batch-projects/{projectId}/books/{bookId}/storyboard/cards/{cardId}", batchMutation(CapabilityBatchConfigure, "projectId", http.HandlerFunc(api.scriptStoryboard)))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/storyboard/reorder", batchMutation(CapabilityBatchConfigure, "projectId", http.HandlerFunc(api.scriptStoryboard)))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/storyboard/recompile", batchMutation(CapabilityBatchExecute, "projectId", http.HandlerFunc(api.scriptStoryboard)))
	mux.Handle("GET /api/v1/batch-projects/{id}/novel-panel", api.requireCapability(CapabilityBatchView, api.requireBatchProjectAccess("id", http.HandlerFunc(api.getNovelPanel))))
	mux.Handle("PUT /api/v1/batch-projects/{id}/novel-panel", batchMutation(CapabilityBatchConfigure, "id", http.HandlerFunc(api.saveNovelPanel)))
	mux.Handle("GET /api/v1/batch-projects/{id}/novel-panel/history", api.requireCapability(CapabilityBatchView, api.requireBatchProjectAccess("id", http.HandlerFunc(api.listNovelPanelHistory))))
	mux.Handle("POST /api/v1/batch-projects/{id}/novel-panel/history/{historyId}/restore", batchMutation(CapabilityBatchConfigure, "id", http.HandlerFunc(api.restoreNovelPanelHistory)))
	mux.Handle("GET /api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/segments", api.requireCapability(CapabilityBatchView, api.requireBatchProjectAccess("projectId", http.HandlerFunc(api.shuihuoSegments))))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/segments", batchMutation(CapabilityBatchConfigure, "projectId", http.HandlerFunc(api.shuihuoSegments)))
	mux.Handle("PUT /api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/segments/{segmentId}", batchMutation(CapabilityBatchConfigure, "projectId", http.HandlerFunc(api.updateShuihuoSegment)))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/segments/reorder", batchMutation(CapabilityBatchConfigure, "projectId", http.HandlerFunc(api.reorderShuihuoSegments)))
	mux.Handle("GET /api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/assets", api.requireCapability(CapabilityBatchView, api.requireBatchProjectAccess("projectId", http.HandlerFunc(api.shuihuoAssets))))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/assets/upload", batchMutation(CapabilityBatchConfigure, "projectId", http.HandlerFunc(api.uploadShuihuoAsset)))
	mux.Handle("GET /api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/assets/{assetId}/content", api.requireCapability(CapabilityBatchView, api.requireBatchProjectAccess("projectId", http.HandlerFunc(api.readShuihuoAsset))))
	mux.Handle("GET /api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/media-tasks", api.requireCapability(CapabilityBatchView, api.requireBatchProjectAccess("projectId", http.HandlerFunc(api.shuihuoMediaTasks))))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/media-tasks", batchMutation(CapabilityBatchExecute, "projectId", http.HandlerFunc(api.shuihuoMediaTasks)))
	mux.Handle("GET /api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/media-tasks/{taskId}/candidates", api.requireCapability(CapabilityBatchView, api.requireBatchProjectAccess("projectId", http.HandlerFunc(api.listShuihuoCandidates))))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/media-tasks/{taskId}/candidates/{candidateId}/select", batchMutation(CapabilityBatchConfigure, "projectId", http.HandlerFunc(api.selectShuihuoCandidate)))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/shuihuo/media-tasks/{taskId}/retry", batchMutation(CapabilityBatchExecute, "projectId", http.HandlerFunc(api.retryShuihuoMediaTask)))

	// Stage 2: unified settings/version profile.
	mux.Handle("GET /api/v1/batch-projects/{id}/settings", api.requireCapability(CapabilityBatchView, api.requireBatchProjectAccess("id", http.HandlerFunc(api.getUnifiedSettings))))
	mux.Handle("PUT /api/v1/batch-projects/{id}/settings/production", batchMutation(CapabilityBatchConfigure, "id", http.HandlerFunc(api.saveProductionSettings)))
	mux.Handle("PUT /api/v1/batch-projects/{id}/settings/publishing", batchMutation(CapabilityPublishConfigure, "id", http.HandlerFunc(api.savePublishingSettings)))
	mux.Handle("GET /api/v1/batch-projects/{id}/version-profile", api.requireCapability(CapabilityBatchView, api.requireBatchProjectAccess("id", http.HandlerFunc(api.getVersionProfile))))
	mux.Handle("PUT /api/v1/batch-projects/{id}/version-profile", batchMutation(CapabilityBatchConfigure, "id", http.HandlerFunc(api.saveVersionProfile)))
	mux.Handle("POST /api/v1/batch-projects/{id}/version-profile/sync-121", batchMutation(CapabilityBatchConfigure, "id", http.HandlerFunc(api.sync121Settings)))
	mux.Handle("POST /api/v1/batch-projects/{id}/version-profile/sync-style-types", batchMutation(CapabilityBatchConfigure, "id", http.HandlerFunc(api.syncStyleTypes)))

	// Stage 3: generation reads and mutations.
	mux.Handle("GET /api/v1/batch-projects/{projectId}/generation", api.requireCapability(CapabilityBatchView, api.requireBatchProjectAccess("projectId", http.HandlerFunc(api.projectGeneration))))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/generation", batchMutation(CapabilityBatchExecute, "projectId", http.HandlerFunc(api.projectGeneration)))
	mux.Handle("GET /api/v1/batch-projects/{projectId}/generation/runs/{runId}", api.requireCapability(CapabilityBatchView, api.requireBatchProjectAccess("projectId", http.HandlerFunc(api.generationRun))))
	mux.Handle("GET /api/v1/batch-projects/{projectId}/books/{bookId}/generation", api.requireCapability(CapabilityBatchView, api.requireBatchProjectAccess("projectId", http.HandlerFunc(api.bookGeneration))))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/generation", batchMutation(CapabilityBatchExecute, "projectId", http.HandlerFunc(api.bookGeneration)))
	mux.Handle("GET /api/v1/batch-projects/{projectId}/books/{bookId}/audio-measurement", api.requireCapability(CapabilityBatchView, api.requireBatchProjectAccess("projectId", http.HandlerFunc(api.audioMeasurement))))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/audio-measurement", batchMutation(CapabilityBatchExecute, "projectId", http.HandlerFunc(api.audioMeasurement)))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/generation/stages/{stage}/retry", batchMutation(CapabilityBatchExecute, "projectId", http.HandlerFunc(api.retryGenerationStage)))
	mux.Handle("GET /api/v1/batch-projects/{projectId}/books/{bookId}/generation/stages/{stage}", api.requireCapability(CapabilityBatchView, api.requireBatchProjectAccess("projectId", http.HandlerFunc(api.generationStage))))

	// Only immutable system presets are public. Private/project prompts must be registered behind auth.
	mux.HandleFunc("GET /api/v1/generation/prompts", api.generationPrompts)

	// Admin prompt governance is independently protected by effective admin
	// capabilities. Mutation routes additionally require same-origin CSRF.
	mux.Handle("GET /api/v1/admin/capabilities", api.requireAnyAdminCapability(http.HandlerFunc(api.adminCapabilities)))
	mux.Handle("GET /api/v1/admin/prompts", api.requireAdminCapability(authn.CapabilityAdminPromptView, http.HandlerFunc(api.adminPromptList)))
	mux.Handle("GET /api/v1/admin/prompts/{key}/versions/{version}", api.requireAdminCapability(authn.CapabilityAdminPromptView, http.HandlerFunc(api.adminPromptDetail)))
	mux.Handle("POST /api/v1/admin/prompts/{key}/drafts", api.requireAdminCapability(authn.CapabilityAdminPromptEdit, api.requireSameOrigin(http.HandlerFunc(api.adminPromptCreateDraft))))
	mux.Handle("PUT /api/v1/admin/prompts/{key}/drafts/{version}", api.requireAdminCapability(authn.CapabilityAdminPromptEdit, api.requireSameOrigin(http.HandlerFunc(api.adminPromptUpdateDraft))))
	mux.Handle("POST /api/v1/admin/prompts/{key}/versions/{version}/publish", api.requireAdminCapability(authn.CapabilityAdminPromptPublish, api.requireSameOrigin(http.HandlerFunc(api.adminPromptPublish))))
	mux.Handle("POST /api/v1/admin/prompts/{key}/versions/{version}/restore", api.requireAdminCapability(authn.CapabilityAdminPromptPublish, api.requireSameOrigin(http.HandlerFunc(api.adminPromptRestore))))

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
	mux.Handle("GET /api/v1/batch-projects/{projectId}/video", api.requireCapability(CapabilityBatchView, api.requireBatchProjectAccess("projectId", http.HandlerFunc(api.projectVideoStatus))))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/video", batchMutation(CapabilityBatchExecute, "projectId", http.HandlerFunc(api.startVideo)))
	mux.Handle("POST /api/v1/video-tasks/{taskId}/poll", api.requireSameOrigin(api.requireCapability(CapabilityBatchExecute, api.requireVideoTaskAccess(api.requireActiveResolvedBatchProject(http.HandlerFunc(api.pollVideoTask))))))
	mux.Handle("POST /api/v1/video-tasks/{taskId}/cancel", api.requireSameOrigin(api.requireCapability(CapabilityBatchExecute, api.requireVideoTaskAccess(http.HandlerFunc(api.cancelVideoTask)))))
	mux.Handle("POST /api/v1/video-tasks/{taskId}/retry", api.requireSameOrigin(api.requireCapability(CapabilityBatchExecute, api.requireVideoTaskAccess(api.requireActiveResolvedBatchProject(http.HandlerFunc(api.retryVideoTask))))))
	mux.Handle("POST /api/v1/batch-projects/{projectId}/books/{bookId}/merge", batchMutation(CapabilityBatchExecute, "projectId", http.HandlerFunc(api.startVideoMerge)))
	mux.Handle("GET /api/v1/video-merge-jobs/{jobId}", api.requireCapability(CapabilityBatchView, api.requireVideoMergeJobAccess(http.HandlerFunc(api.getVideoMerge))))
	mux.Handle("POST /api/v1/video-merge-attempts/{attemptId}/retry", api.requireSameOrigin(api.requireCapability(CapabilityBatchExecute, api.requireVideoMergeAttemptAccess(api.requireActiveResolvedBatchProject(http.HandlerFunc(api.retryVideoMerge))))))

	// Local executor traffic has a distinct service identity. Registration uses a
	// server-side bootstrap token; heartbeat/identity/complete/fail continue to
	// authenticate with the executor's one-time-issued Bearer credential.
	mux.Handle("POST /api/v1/video/local-executors/register", api.requireVideoExecutorBootstrap(http.HandlerFunc(api.registerVideoLocalExecutor)))
	mux.Handle("POST /api/v1/video/local-executors/pairings", api.requireSameOrigin(api.requireCapability(CapabilityBatchConfigure, http.HandlerFunc(api.createVideoLocalExecutorPairing))))
	mux.HandleFunc("POST /api/v1/video/local-executors/redeem", api.redeemVideoLocalExecutorPairing)
	mux.Handle("GET /api/v1/video/local-executors", api.requireCapability(CapabilityBatchConfigure, http.HandlerFunc(api.listVideoLocalExecutors)))
	mux.Handle("DELETE /api/v1/video/local-executors/{executorId}", api.requireSameOrigin(api.requireCapability(CapabilityBatchConfigure, http.HandlerFunc(api.unbindVideoLocalExecutor))))
	mux.HandleFunc("GET /api/v1/video/local-executors/me", api.getVideoLocalExecutorIdentity)
	mux.HandleFunc("POST /api/v1/video/local-executors/heartbeat", api.heartbeatVideoLocalExecutor)
	mux.HandleFunc("POST /api/v1/video/local-executor-tasks/{taskId}/complete", api.completeVideoLocalExecutorTask)
	mux.HandleFunc("POST /api/v1/video/local-executor-tasks/{taskId}/fail", api.failVideoLocalExecutorTask)
	return api.withObservability(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pattern := mux.Handler(r)
		if pattern == "" && (r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/")) {
			// Keep ServeMux's 404/405 decision and Allow header; replace only
			// its default plain-text body with the shared API error envelope.
			mux.ServeHTTP(&apiRouteErrorWriter{ResponseWriter: w}, r)
			return
		}
		mux.ServeHTTP(w, r)
	}))
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	if status >= http.StatusBadRequest {
		value = errorEnvelope(w, status, value)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
