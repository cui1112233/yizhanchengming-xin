package app

import (
	"database/sql"
	"encoding/base64"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/httpapi"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/publishing"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/task9runtime"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/taskruntime"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/unifiedsettings"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/video"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/workshop"
)

func NewHandler(db *sql.DB, fetcher intake.Fetcher, classifier intake.Classifier, now pipeline.Clock) http.Handler {
	return newHandler(db, fetcher, classifier, now, true)
}

func newHandler(db *sql.DB, fetcher intake.Fetcher, classifier intake.Classifier, now pipeline.Clock, authEnabled bool) http.Handler {
	startedAt := time.Now().UTC()
	logger := slog.Default()
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
	workshopService := workshop.NewService(workshop.NewMySQLStore(db), store, generationService)
	observedGeneration := observedGenerationService{next: generationService, logger: logger}
	settingsService := unifiedsettings.NewService(settingsStore, unifiedsettings.StaticDefaults{Config: unifiedsettings.Settings{
		Production: map[string]any{"productionMode": "original", "aiCopyEnabled": false, "aiCopyCount": float64(1)},
		Publishing: map[string]any{"uploadVideoType": "merged", "materialReuse": false},
	}}, settingsStore)

	var authService httpapi.AuthService
	if authEnabled {
		authStore := authn.NewMySQLStore(db)
		authService = authn.NewService(authStore, authn.NewManager(authStore, authn.Options{}))
	}
	publishingStore := publishing.NewMySQLStore(db)
	publishingService := publishing.NewService(publishingStore, publishing.Options{CredentialKey: publishingCredentialKey()})
	observedPublishing := observedPublishingService{next: publishingService, logger: logger}
	allowedOrigins := make([]string, 0)
	for _, value := range strings.Split(os.Getenv("QIANTIE_ALLOWED_ORIGINS"), ",") {
		if value = strings.TrimSpace(value); value != "" {
			allowedOrigins = append(allowedOrigins, value)
		}
	}

	videoStore := video.NewMySQLStore(db)
	masterKey := []byte(os.Getenv("VIDEO_PROVIDER_MASTER_KEY"))
	providerFactory := video.DefaultProviderFactory{LocalJobs: videoStore}
	videoConfigService := video.NewConfigServiceWithProviders(videoStore, masterKey, providerFactory)
	localExecutorService := video.NewLocalExecutorService(videoStore, nil)
	observedLocalExecutor := observedLocalExecutorService{next: localExecutorService, logger: logger}

	var artifactStore video.ArtifactStore
	var fileArtifactStore video.FileArtifactStore
	if endpoint, region, bucket, accessKey, secretKey, publicBase := os.Getenv("TOS_ENDPOINT"), os.Getenv("TOS_REGION"), os.Getenv("TOS_BUCKET"), os.Getenv("TOS_ACCESS_KEY"), os.Getenv("TOS_SECRET_KEY"), os.Getenv("TOS_PUBLIC_BASE_URL"); endpoint != "" && region != "" && bucket != "" && accessKey != "" && secretKey != "" && publicBase != "" {
		if uploader, err := video.NewTOSUploader(endpoint, region, accessKey, secretKey); err == nil {
			if durableArtifacts, err := video.NewArtifactStore(video.ArtifactStoreConfig{Bucket: bucket, PublicBaseURL: publicBase, Uploader: uploader}); err == nil {
				artifactStore = durableArtifacts
				fileArtifactStore = durableArtifacts
			}
		}
	}
	videoService := video.NewService(videoStore, video.NewGenerationFinalPromptSource(generationStore), providerFactory, artifactStore, masterKey)
	observedVideo := &observedVideoService{next: videoService, logger: logger}
	mergeExecutor := video.NewFFmpegExecutor(video.FFmpegExecutorConfig{
		Binary:    os.Getenv("FFMPEG_BINARY"),
		TempRoot:  os.Getenv("VIDEO_MERGE_TEMP_ROOT"),
		Artifacts: fileArtifactStore,
	})
	mergeService := video.NewMergeService(videoStore, mergeExecutor)
	observedMerge := observedMergeService{next: mergeService, logger: logger}

	// Browser Runtime mutations reuse the same durable BookRun store and Redis
	// queue contract as Scheduler/Worker. If Redis is unavailable the handler
	// remains bootable but retry returns an explicit queue-unavailable response;
	// no new BookRun attempt is created without successful coordination.
	runtimeStore := task9runtime.NewMySQLStore(db)
	var runtimeCoordinator task9runtime.RuntimeCoordinator
	redisAddr := strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	if redisAddr == "" {
		redisAddr = strings.TrimSpace(os.Getenv("QIANTIE_REDIS_ADDR"))
	}
	if redisAddr == "" {
		redisAddr = strings.TrimSpace(os.Getenv("TASK9_REDIS_ADDR"))
	}
	if redisAddr != "" {
		if queue, err := taskruntime.NewRedisQueue(redisAddr, "task9"); err == nil {
			runtimeCoordinator = task9runtime.NewQueueCoordinator(queue)
		}
	}
	runtimeService := task9runtime.NewRetryService(runtimeStore, runtimeCoordinator)

	deps := httpapi.Dependencies{
		Intakes:                     intakeService,
		Reader:                      store,
		Pipeline:                    pipelineService,
		BatchProjects:               store,
		BatchProjectDetails:         store,
		Generation:                  observedGeneration,
		Workshop:                    workshopService,
		UnifiedSettings:             settingsService,
		Auth:                        authService,
		Publishing:                  observedPublishing,
		SecureCookies:               secureCookiesEnabled(),
		AllowedOrigins:              allowedOrigins,
		Video:                       observedVideo,
		VideoConfig:                 videoConfigService,
		VideoLocalExecutor:          observedLocalExecutor,
		VideoStatus:                 videoService,
		VideoMerge:                  observedMerge,
		BatchProjectAccess:          publishingStore,
		VideoResourceProjects:       videoStore,
		VideoExecutorBootstrapToken: strings.TrimSpace(os.Getenv("VIDEO_LOCAL_EXECUTOR_BOOTSTRAP_TOKEN")),
		Database:                    db,
		Logger:                      logger,
		StartedAt:                   startedAt,
		AppInitialized:              true,
	}
	return httpapi.NewHandlerWithRuntime(deps, runtimeService)
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
