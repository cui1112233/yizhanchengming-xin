package app

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
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
	"github.com/cui1112233/yizhanchengming-xin/api/internal/novelpanel"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/objectkey"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/publishing"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/shuihuo"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/task9runtime"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/taskruntime"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/unifiedsettings"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/video"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/workshop"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/workspace"
)

func NewHandler(db *sql.DB, fetcher intake.Fetcher, classifier intake.Classifier, now pipeline.Clock) http.Handler {
	return newHandler(db, fetcher, classifier, now, true)
}

func newHandler(db *sql.DB, fetcher intake.Fetcher, classifier intake.Classifier, now pipeline.Clock, authEnabled bool) http.Handler {
	return newHandlerWithRuntime(db, fetcher, classifier, now, authEnabled, runtimeWiring{})
}

type runtimeWiring struct {
	generation  httpapi.GenerationRuntimeService
	coordinator task9runtime.RuntimeCoordinator
	diagnostics httpapi.RuntimeDiagnosticsReader
}

type RuntimeLifecycle interface {
	Start(context.Context) error
	Ready() bool
	Status() RuntimeStatus
	Wait() error
	Close() error
}

type Application struct {
	Handler http.Handler
	Runtime RuntimeLifecycle
}

func NewApplication(db *sql.DB, fetcher intake.Fetcher, classifier intake.Classifier, now pipeline.Clock) (*Application, error) {
	runtimeStore := task9runtime.NewMySQLStore(db)
	generationStore := generation.NewMySQLStore(db)
	textProvider, providerConfigured := textGenerationProvider()
	redisAddr := runtimeRedisAddress()
	options := runtimeLifecycleOptions{UnavailableReason: "not_configured", RedisConfigured: redisAddr != ""}
	var coordinator task9runtime.RuntimeCoordinator
	if redisAddr != "" {
		prefix := runtimeRedisPrefix(os.Getenv("QIANTIE_REDIS_PREFIX"), "task9")
		queue, queueErr := taskruntime.NewRedisQueue(redisAddr, prefix)
		if queueErr != nil {
			options.UnavailableReason = "redis_unavailable"
		} else {
			options.Closers = append(options.Closers, io.Closer(queue))
			leases, leaseErr := taskruntime.NewRedisLeaseStore(redisAddr, prefix)
			if leaseErr != nil {
				_ = queue.Close()
				options.Closers = nil
				options.UnavailableReason = "redis_unavailable"
			} else {
				options.Closers = append(options.Closers, io.Closer(leases))
				if !providerConfigured {
					options.UnavailableReason = "provider_not_configured"
				} else {
					nowFn := time.Now
					if now != nil {
						nowFn = now
					}
					owner := runtimeOwner()
					coordinator = task9runtime.NewQueueCoordinator(queue)
					options.Configured = true
					options.Queue = queue
					options.HealthCheck = func(ctx context.Context) error {
						_, err := queue.ReclaimExpired(ctx, time.Now(), 1)
						return err
					}
					options.Worker = task9runtime.NewWorker(runtimeStore, leases, generation.NewRuntimeExecutor(generationStore, textProvider, nowFn), owner, 30*time.Second, nowFn)
					options.Scheduler = task9runtime.NewScheduler(runtimeStore, coordinator, nowFn)
					options.Recovery = task9runtime.NewRecovery(runtimeStore, coordinator)
				}
			}
		}
	}
	lifecycle := newGenerationRuntimeLifecycle(options)
	generationRuntime := task9runtime.NewGenerationAdmissionService(runtimeStore, coordinator, lifecycle)
	handler := newHandlerWithRuntime(db, fetcher, classifier, now, true, runtimeWiring{
		generation:  generationRuntime,
		coordinator: coordinator,
		diagnostics: lifecycle,
	})
	return &Application{Handler: handler, Runtime: lifecycle}, nil
}

func newHandlerWithRuntime(db *sql.DB, fetcher intake.Fetcher, classifier intake.Classifier, now pipeline.Clock, authEnabled bool, runtime runtimeWiring) http.Handler {
	startedAt := time.Now().UTC()
	logger := slog.Default()
	store := intake.NewMySQLStore(db)
	recentStore := workspace.NewMySQLStore(db)
	intakeService := intake.NewService(store, fetcher, classifier)
	pipelineService := pipeline.NewService(store, now)
	generationStore := generation.NewMySQLStore(db)
	settingsStore := unifiedsettings.NewMySQLStore(db)

	textProvider, _ := textGenerationProvider()
	generationService := generation.NewService(generationStore, textProvider, nil)
	workshopService := workshop.NewService(workshop.NewMySQLStore(db), store, generationService)
	novelPanelService := novelpanel.NewService(novelpanel.NewMySQLStore(db))
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
	// Test-only handlers deliberately omit authentication. Do not attach the
	// ownership checker in that mode: it correctly requires an authenticated
	// actor, while the store-wiring tests exercise the unauthenticated fixture.
	// Production handlers always enable both auth and ownership checks together.
	var batchProjectAccess httpapi.BatchProjectAccessChecker
	if authEnabled {
		batchProjectAccess = publishingStore
	}
	allowedOrigins := make([]string, 0)
	for _, value := range strings.Split(os.Getenv("QIANTIE_ALLOWED_ORIGINS"), ",") {
		if value = strings.TrimSpace(value); value != "" {
			allowedOrigins = append(allowedOrigins, value)
		}
	}

	videoStore := video.NewMySQLStore(db)
	shuihuoStore := shuihuo.NewMySQLStore(db)
	masterKey := []byte(os.Getenv("VIDEO_PROVIDER_MASTER_KEY"))
	providerFactory := video.DefaultProviderFactory{LocalJobs: videoStore}
	videoConfigService := video.NewConfigServiceWithProviders(videoStore, masterKey, providerFactory)
	localExecutorService := video.NewLocalExecutorService(videoStore, nil)
	observedLocalExecutor := observedLocalExecutorService{next: localExecutorService, logger: logger}

	var artifactStore video.ArtifactStore
	var fileArtifactStore video.FileArtifactStore
	var tosUploader *video.TOSUploader
	var tosObjects shuihuo.ObjectStore
	var tosPrefix objectkey.Prefix
	tosBucket := ""
	if err := configureTOSKeyPrefix(os.Getenv("TOS_KEY_PREFIX"), func(prefix objectkey.Prefix) {
		tosPrefix = prefix
		if endpoint, region, bucket, accessKey, secretKey, publicBase := os.Getenv("TOS_ENDPOINT"), os.Getenv("TOS_REGION"), os.Getenv("TOS_BUCKET"), os.Getenv("TOS_ACCESS_KEY"), os.Getenv("TOS_SECRET_KEY"), os.Getenv("TOS_PUBLIC_BASE_URL"); endpoint != "" && region != "" && bucket != "" && accessKey != "" && secretKey != "" {
			if uploader, err := video.NewTOSUploader(endpoint, region, accessKey, secretKey); err == nil {
				tosUploader, tosObjects, tosBucket = uploader, uploader, bucket
				if publicBase != "" {
					if durableArtifacts, err := video.NewArtifactStore(video.ArtifactStoreConfig{Bucket: bucket, PublicBaseURL: publicBase, Uploader: uploader, KeyPrefix: prefix}); err == nil {
						artifactStore = durableArtifacts
						fileArtifactStore = durableArtifacts
					}
				}
			}
		}
	}); err != nil {
		logger.Error("TOS storage disabled", "subsystem", "tos", "safe_error", err.Error())
	}
	shuihuoMediaService := shuihuo.NewService(shuihuoStore, tosObjects)
	shuihuoMediaService.SetBucket(tosBucket)
	shuihuoMediaService.SetKeyPrefix(tosPrefix)
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
	redisAddr := runtimeRedisAddress()
	runtimeCoordinator := runtime.coordinator
	// Image/TTS credentials remain server-only. Each adapter uses the same small
	// HTTP contract and writes its durable output back through the existing TOS
	// media service; no browser provider configuration is exposed.
	imageProvider := shuihuo.NewHTTPProvider(shuihuo.HTTPProviderConfig{BaseURL: os.Getenv("SHUIHUO_IMAGE_PROVIDER_BASE_URL"), APIKey: os.Getenv("SHUIHUO_IMAGE_PROVIDER_API_KEY"), ImageModel: os.Getenv("SHUIHUO_IMAGE_PROVIDER_MODEL")})
	ttsProvider := shuihuo.NewHTTPProvider(shuihuo.HTTPProviderConfig{BaseURL: os.Getenv("SHUIHUO_TTS_PROVIDER_BASE_URL"), APIKey: os.Getenv("SHUIHUO_TTS_PROVIDER_API_KEY"), TTSModel: os.Getenv("SHUIHUO_TTS_PROVIDER_MODEL")})
	var mediaQueue taskruntime.Queue
	if redisAddr != "" {
		if queue, err := taskruntime.NewRedisQueue(redisAddr, runtimeRedisPrefix(os.Getenv("QIANTIE_REDIS_PREFIX"), "shuihuo-media")); err == nil {
			mediaQueue = queue
		}
	}
	mediaProvider := shuihuo.ProviderSet{Image: imageProvider, Audio: ttsProvider}
	// A provider output is not a successful media result until it is durable in
	// TOS. Keep the truthful executor_unavailable state when storage is absent.
	if tosUploader != nil && tosBucket != "" {
		shuihuoMediaService.SetMediaExecutor(mediaProvider, mediaQueue)
	}
	if mediaQueue != nil && (mediaProvider.Available(shuihuo.MediaImage) || mediaProvider.Available(shuihuo.MediaAudio)) && tosUploader != nil && tosBucket != "" {
		owner := "api"
		if hostname, err := os.Hostname(); err == nil && hostname != "" {
			owner = fmt.Sprintf("api-%s", hostname)
		}
		go func() {
			if err := shuihuo.NewWorker(shuihuoMediaService, mediaQueue, owner).Run(context.Background()); err != nil {
				logger.Error("shuihuo media worker stopped", "subsystem", "shuihuo_media", "safe_error", err.Error())
			}
		}()
	}
	runtimeService := task9runtime.NewRetryService(runtimeStore, runtimeCoordinator)
	// Production injects the single supervised Task9 lifecycle above. The
	// compatibility constructor deliberately stays fail-closed; it is used by
	// tests and must never infer that a worker exists merely from environment
	// variables. Read-only exact Run polling remains available in either mode.
	var generationRuntime httpapi.GenerationRuntimeService = runtime.generation
	if generationRuntime == nil {
		generationRuntime = task9runtime.NewGenerationAdmissionService(runtimeStore, runtimeCoordinator, task9runtime.StaticRuntimeReadiness(false))
	}

	deps := httpapi.Dependencies{
		Intakes:                     intakeService,
		Reader:                      store,
		Pipeline:                    pipelineService,
		BatchProjects:               store,
		WorkspaceRecent:             recentStore,
		WorkspaceHistory:            recentStore,
		BatchProjectDetails:         store,
		ScriptBooks:                 store,
		ScriptStoryboards:           generationService,
		Generation:                  observedGeneration,
		GenerationRuntime:           generationRuntime,
		RuntimeDiagnostics:          runtime.diagnostics,
		AdminPrompts:                generationStore,
		Workshop:                    workshopService,
		NovelPanel:                  novelPanelService,
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
		ShuihuoMedia:                shuihuoMediaService,
		IntakeAccess:                store,
		BatchProjectAccess:          batchProjectAccess,
		BatchProjectLifecycle:       store,
		VideoResourceProjects:       videoStore,
		VideoExecutorBootstrapToken: strings.TrimSpace(os.Getenv("VIDEO_LOCAL_EXECUTOR_BOOTSTRAP_TOKEN")),
		Database:                    db,
		Logger:                      logger,
		StartedAt:                   startedAt,
		AppInitialized:              true,
	}
	return httpapi.NewHandlerWithRuntime(deps, runtimeService)
}

func textGenerationProvider() (generation.Provider, bool) {
	baseURL := strings.TrimSpace(os.Getenv("QIANTIE_TEXT_API_BASE_URL"))
	key := strings.TrimSpace(os.Getenv("QIANTIE_TEXT_API_KEY"))
	model := strings.TrimSpace(os.Getenv("QIANTIE_TEXT_MODEL"))
	if baseURL == "" || key == "" || model == "" {
		return generation.UnavailableProvider{}, false
	}
	return generation.NewHTTPProvider(baseURL, key, model), true
}

func runtimeRedisAddress() string {
	for _, key := range []string{"REDIS_ADDR", "QIANTIE_REDIS_ADDR", "TASK9_REDIS_ADDR"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func runtimeOwner() string {
	hostname := "unknown"
	if value, err := os.Hostname(); err == nil && strings.TrimSpace(value) != "" {
		hostname = strings.TrimSpace(value)
	}
	return fmt.Sprintf("api-%s-pid%d", hostname, os.Getpid())
}

func runtimeRedisPrefix(namespace, name string) string {
	namespace = strings.Trim(strings.TrimSpace(namespace), ":")
	if namespace == "" {
		return name
	}
	return namespace + ":" + name
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
