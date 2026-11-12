package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"qiantie/backend/internal/agentproxy"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/aicut"
	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/batchfactoryv11/external"
	"qiantie/backend/internal/ffmpegbin"
	"qiantie/backend/internal/giantmaterialexecutor"
	"qiantie/backend/internal/imagegenerationcache"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
	"qiantie/backend/internal/mergeworker"
	"qiantie/backend/internal/novelcreation"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/releasenotes"
	"qiantie/backend/internal/runtimeaudit"
	"qiantie/backend/internal/screenplay"
	"qiantie/backend/internal/upscale"
	"qiantie/backend/internal/videomanagementconnector"
)

type runtimeManifest struct {
	GitSHA        string `json:"gitSha"`
	FrontendEntry string `json:"frontendEntry"`
}

func readRuntimeManifest(path string) runtimeManifest {
	contents, err := os.ReadFile(path)
	if err != nil {
		return runtimeManifest{}
	}
	var manifest runtimeManifest
	if json.Unmarshal([]byte(strings.TrimPrefix(string(contents), "\ufeff")), &manifest) != nil {
		return runtimeManifest{}
	}
	return manifest
}

type RouterOptions struct {
	AgentVideoWorkerContext       context.Context
	AgentScriptSplitWorkerContext context.Context
	AgentImageWorkerContext       context.Context
	AgentAudioWorkerContext       context.Context
	BridgeSecret                  string
	ReleaseSHA                    string
	RuntimeManifestPath           string
	RuntimeAudit                  *runtimeaudit.Registry
	Health                        func(context.Context) error
	Now                           func() time.Time
	// AuthRateLimits overrides the login/registration/password-reset limits.
	// A zero value falls back to the defaults documented in the config package.
	AuthRateLimits                  AuthRateLimitConfig
	Users                           BridgeUserResolver
	AuthStore                       auth.Store
	SessionResolver                 func(context.Context, string) (BridgeIdentity, error)
	Slice                           int
	RegisterV11                     func(*http.ServeMux)
	Store                           batchfactoryv11.Store
	Director                        *batchfactoryv11.DirectorService
	H3AudioProbe                    batchfactoryv11.H3AudioDurationProbe
	Compiler                        *batchfactoryv11.PromptCompilerService
	Production                      *batchfactoryv11.ProductionService
	Merge                           *batchfactoryv11.MergeService
	External                        *external.Service
	NovelFetchStore                 novelfetchworkshop.LifecycleStore
	NovelFetchQueue                 novelfetchworkshop.ProcessQueue
	NovelFetchAIGenderJudge         novelfetchworkshop.AIGenderJudge
	NovelFetchAIGenderJudgeResolver func(context.Context, string) (novelfetchworkshop.AIGenderJudge, error)
	NovelFetchFetcher               *novelfetchworkshop.UpstreamFetcher
	NovelFetchRewrite               novelfetchworkshop.RewriteProvider
	NovelFetchRewriteResolver       func(context.Context, string) (novelfetchworkshop.RewriteProvider, error)
	PresetCatalog                   *prompt.Catalog
	AutomationPresetStore           AutomationPresetStore
	AutomationStatusReader          *FileAutomationStatusReader
	AutomationStateStore            AutomationStateStore
	AutomationCoordinator           *BatchAutomationCoordinator
	PersonalPromptStore             *prompt.PersonalStore
	ScriptHistoryStore              *ScriptHistoryStore
	MemberProfileStore              *MemberProfileStore
	MemberAvatarDir                 string
	UserDataDir                     string
	UserConfigStore                 *UserConfigStore
	TextProvider                    batchfactoryv11.DirectorProvider
	// TextProviderResolver resolves the owner-scoped provider for an explicit
	// textModelId saved in batch/book settings. Batch stages use it so the
	// workbench text-model selector is honoured instead of ignored.
	TextProviderResolver      func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error)
	AgentTextModelResolver    AgentTextModelResolver
	AgentWorkspace            agentworkspace.Store
	AgentProxy                agentproxy.Store
	AICut                     aicut.Store
	Screenplay                screenplay.Store
	NovelCreation             *novelcreation.Service
	NovelCreationRunStarter   NovelCreationRunStarter
	AICutASR                  *aicut.ASRService
	AICutExport               *AICutExportService
	AICutAgentStore           aicut.AgentStore
	AICutAgentStarter         AICutAgentStarter
	Upscale                   *upscale.Service
	AgentAssets               agentworkspace.AssetStore
	AgentAssetObjects         agentworkspace.AssetObjectStore
	RedisURL                  string
	LocalExecutors            *localexecutor.Service
	GiantMaterialExecutor     *giantmaterialexecutor.Service
	GiantMaterialResolver     GiantMaterialResolver
	LocalArtifacts            *localartifact.Store
	ImageGenerationCache      imagegenerationcache.Repository
	LocalExecutorDownloadsDir string
	MergeOutput               mergeworker.ObjectStore
	VideoManagementConnector  videomanagementconnector.Service
	N8ComicPublisher          *BatchFactoryN8ComicPublisher
	ComicDirectRecordPath     string
}

func managedTextProviderResolver(resolver func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error), resolveOwner func(string) string) func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error) {
	if resolver == nil {
		return nil
	}
	return func(ctx context.Context, owner, modelID string) (batchfactoryv11.DirectorProvider, error) {
		if resolveOwner != nil {
			if resolved := strings.TrimSpace(resolveOwner(owner)); resolved != "" {
				owner = resolved
			}
		}
		return resolver(ctx, owner, modelID)
	}
}

func NewRouter(options RouterOptions) http.Handler {
	resolveRuntimeModelOwner := func(owner string) string {
		return effectiveModelCatalogOwner(BridgeIdentity{Username: owner}, options.MemberProfileStore, options.AuthStore)
	}
	runtimeCatalogStore := runtimeModelCatalogStore(options.NovelFetchStore, resolveRuntimeModelOwner)
	runtimeTextProviderResolver := managedTextProviderResolver(options.TextProviderResolver, resolveRuntimeModelOwner)
	if options.Director != nil && options.Director.PresetCatalog == nil {
		options.Director.PresetCatalog = options.PresetCatalog
	}
	comicGenerationGate := newComicPageGenerationGate()
	if options.Production == nil && options.Store != nil {
		// Read-only status is safe before the production slice is enabled. Keep
		// the service disabled for all write/submit operations.
		options.Production = &batchfactoryv11.ProductionService{Store: options.Store, Compiler: options.Compiler}
	}
	if options.AutomationCoordinator == nil && options.Store != nil && options.AutomationStatusReader != nil {
		var providerSync func(context.Context, string) error
		if options.Production != nil && options.Production.ProviderRegistry != nil {
			providerSync = func(ctx context.Context, owner string) error {
				return syncShuihuoConfiguredVideoModels(ctx, options.Store, options.Production.ProviderRegistry, owner, runtimeCatalogStore)
			}
		}
		automationStages := &batchfactoryv11.BookStageService{
			Store: options.Store, Director: options.Director, Production: options.Production,
			ImageReady:           shuihuoImageReadyChecker(options.Store, runtimeCatalogStore, options.UserConfigStore, resolveRuntimeModelOwner),
			VideoReady:           shuihuoVideoReadyChecker(options.Store, runtimeCatalogStore, options.Production),
			TextProviderResolver: runtimeTextProviderResolver,
		}
		options.AutomationCoordinator = NewBatchAutomationCoordinator(options.Store, options.AutomationPresetStore, automationStages, options.Merge, options.External, runtimeCatalogStore, options.H3AudioProbe, options.AutomationStatusReader, options.Now, providerSync)
		options.AutomationCoordinator.SetLocalFiles(options.LocalArtifacts)
		options.AutomationCoordinator.SetSourceFetcher(options.NovelFetchFetcher)
		options.AutomationCoordinator.SetGiantMaterialResolver(options.GiantMaterialResolver)
		comicGenerator := comicPageGenerator{store: options.Store, files: options.LocalArtifacts, localExecutors: options.LocalExecutors, catalogStore: runtimeCatalogStore, userConfig: options.UserConfigStore, presetCatalog: options.PresetCatalog, generationGate: comicGenerationGate, imageCache: options.ImageGenerationCache, resolveModelOwner: resolveRuntimeModelOwner, rewritePrompt: func(ctx context.Context, owner string, request novelfetchworkshop.RewriteRequest) (string, error) {
			provider := providerForBatchClassification(ctx, owner, batchFetchClassification{RewriteProvider: options.NovelFetchRewrite, RewriteResolver: options.NovelFetchRewriteResolver})
			if provider == nil {
				return "", batchfactoryv11.ErrUnavailable
			}
			return provider.Complete(ctx, request)
		}}
		options.AutomationCoordinator.SetComicBookGenerator(comicGenerator.generateBook)
	}
	if options.AutomationCoordinator != nil {
		options.AutomationCoordinator.SetGiantMaterialResolver(options.GiantMaterialResolver)
		options.AutomationCoordinator.SetAutomationStateStore(options.AutomationStateStore)
		if options.AutomationStatusReader != nil {
			options.AutomationStatusReader.SetStateStore(options.AutomationStateStore)
		}
		// Keep injected coordinators on the router's prompt catalog too.
		options.AutomationCoordinator.promptCatalog = options.PresetCatalog
		options.AutomationCoordinator.SetSourceClassification(batchFetchClassification{
			AIGenderJudge:         options.NovelFetchAIGenderJudge,
			AIGenderJudgeResolver: options.NovelFetchAIGenderJudgeResolver,
			RewriteProvider:       options.NovelFetchRewrite,
			RewriteResolver:       options.NovelFetchRewriteResolver,
		})
		if options.N8ComicPublisher != nil {
			options.AutomationCoordinator.SetComicPublisher(options.N8ComicPublisher.PublishFromSettings)
			options.AutomationCoordinator.SetComicPublishRecordReader(options.N8ComicPublisher.savedRecord)
			options.AutomationCoordinator.SetComicPublishPreflight(options.N8ComicPublisher.ValidateBookSettings)
		}
		log.Printf("[DIAG] AutomationCoordinator reader=%v path=%q", options.AutomationStatusReader != nil, func() string {
			if options.AutomationStatusReader != nil {
				return options.AutomationStatusReader.Path()
			}
			return "(nil)"
		}())
		if err := options.AutomationCoordinator.RecoverInterruptedJobs(); err != nil {
			log.Printf("recover interrupted batch automation jobs: %v", err)
		} else {
			log.Printf("[DIAG] RecoverInterruptedJobs OK (no error)")
		}
		// 启动时默认不自动恢复（只做 RecoverInterruptedJobs 标记）。自动恢复会
		// 把书重新置为 running 并立即重新派发，缺少执行器/发布配置时会产生
		// 一批无效重试；需要显式开启 QIANTIE_AUTOMATION_RESUME_ON_BOOT=1。
		resumeAutomationOnBoot(options.AutomationCoordinator)
		options.AutomationCoordinator.StartScheduledDispatcher()
	}
	// 生产任务心跳超时兜底：启动时清理历史遗留的 running/queued，
	// 避免服务中断后父任务永久停在“制作中”。阈值取 30 分钟。
	if cleaner, ok := options.Store.(batchfactoryv11.StaleProductionJobCleaner); ok {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			cleaned, err := cleaner.FailStaleProductionJobs(ctx, time.Now().Add(-30*time.Minute), "服务中断遗留，任务超时未完成，已自动改判失败")
			if err != nil {
				log.Printf("clean stale production jobs: %v", err)
			} else if cleaned > 0 {
				log.Printf("clean stale production jobs: %d jobs marked failed", cleaned)
			}
		}()
	}
	v11 := http.NewServeMux()
	v12 := http.NewServeMux()
	publish121Available := options.VideoManagementConnector != nil || (options.External != nil && options.External.Enabled[external.Provider121])
	publishYadiAvailable := options.External != nil && options.External.Enabled[external.ProviderYadi]
	// V12 is the public workbench namespace. Keep its capability contract on
	// Go even while a few legacy V12 adapters still bridge to Node.
	if options.Store != nil && options.Slice >= 1 {
		registerSliceOneRoutes(v11, options.Store, options.PresetCatalog)
		registerV12IntakeRoutes(v12, options.Store)
		registerV12PublishClassificationRoute(v12, options.Store, batchFetchClassification{
			AIGenderJudge: options.NovelFetchAIGenderJudge, AIGenderJudgeResolver: options.NovelFetchAIGenderJudgeResolver,
			RewriteProvider: options.NovelFetchRewrite, RewriteResolver: options.NovelFetchRewriteResolver,
		})
		registerConfiguredImageGenerationRoute(v11, options.Store, runtimeCatalogStore)
		registerConfiguredImageGenerationRoute(v12, options.Store, runtimeCatalogStore)
		registerBookAssetRoutes(v11, options.Store, options.LocalArtifacts)
		registerV12AssetGenerationRoute(v12, options.Store, options.LocalArtifacts, options.LocalExecutors, options.PresetCatalog, runtimeCatalogStore, options.UserConfigStore, resolveRuntimeModelOwner, func(ctx context.Context, owner string, request novelfetchworkshop.RewriteRequest) (string, error) {
			provider := providerForBatchClassification(ctx, owner, batchFetchClassification{RewriteProvider: options.NovelFetchRewrite, RewriteResolver: options.NovelFetchRewriteResolver})
			if provider == nil {
				return "", batchfactoryv11.ErrUnavailable
			}
			return provider.Complete(ctx, request)
		})
		registerComicPageRoutes(v12, options.Store, options.LocalArtifacts, options.LocalExecutors, runtimeCatalogStore, options.UserConfigStore, options.PresetCatalog, options.AuthStore, options.MemberProfileStore, options.Director, comicGenerationGate, options.ImageGenerationCache)
	}
	registerComicDirectRoutes(v12, options.N8ComicPublisher, options.ComicDirectRecordPath, options.Now)
	registerN8ComicPublishRoutes(v12, options.N8ComicPublisher)
	if options.NovelFetchFetcher != nil {
		registerV12FetchOriginalsRoute(v12, options.Store, options.NovelFetchStore, options.NovelFetchFetcher, batchFetchClassification{
			AIGenderJudge: options.NovelFetchAIGenderJudge, AIGenderJudgeResolver: options.NovelFetchAIGenderJudgeResolver,
			RewriteProvider: options.NovelFetchRewrite, RewriteResolver: options.NovelFetchRewriteResolver,
		})
		registerV12SingleFetchRoute(v12, options.Store, options.NovelFetchStore, options.NovelFetchFetcher, options.GiantMaterialResolver, batchFetchClassification{
			AIGenderJudge: options.NovelFetchAIGenderJudge, AIGenderJudgeResolver: options.NovelFetchAIGenderJudgeResolver,
			RewriteProvider: options.NovelFetchRewrite, RewriteResolver: options.NovelFetchRewriteResolver,
		})
	}
	if options.Store != nil && options.Director != nil && options.Slice >= 2 {
		registerDirectorRoutes(v11, options.Director, options.Store)
		stageService := &batchfactoryv11.BookStageService{Store: options.Store, Director: options.Director, Production: options.Production, TextProviderResolver: runtimeTextProviderResolver}
		if options.LocalArtifacts != nil {
			stageService.ImageReady = shuihuoImageReadyChecker(options.Store, runtimeCatalogStore, options.UserConfigStore, resolveRuntimeModelOwner)
			stageService.VideoReady = shuihuoVideoReadyChecker(options.Store, runtimeCatalogStore, options.Production)
			stageService.ImageGenerator = func(ctx context.Context, owner, batchID, bookID string) error {
				batch, err := options.Store.GetBatch(ctx, owner, batchID)
				if err != nil {
					return err
				}
				var book batchfactoryv11.Book
				found := false
				for _, candidate := range batch.Books {
					if candidate.ID == bookID {
						book, found = candidate, true
						break
					}
				}
				if !found {
					return batchfactoryv11.ErrNotFound
				}
				assetIDs := make([]string, 0, len(book.AssetRecords))
				for _, asset := range book.AssetRecords {
					assetIDs = append(assetIDs, asset.ID)
				}
				_, err = generateShuihuoAssetImagesWithPresetCatalog(ctx, options.Store, options.LocalArtifacts, owner, batch, assetIDs, "", "9:16", "", options.LocalExecutors, options.PresetCatalog, runtimeCatalogStore)
				return err
			}
		}
		retryComicPages := (comicPageGenerator{store: options.Store, files: options.LocalArtifacts, localExecutors: options.LocalExecutors, catalogStore: runtimeCatalogStore, userConfig: options.UserConfigStore, presetCatalog: options.PresetCatalog, generationGate: comicGenerationGate, imageCache: options.ImageGenerationCache, resolveModelOwner: resolveRuntimeModelOwner, rewritePrompt: func(ctx context.Context, owner string, request novelfetchworkshop.RewriteRequest) (string, error) {
			provider := providerForBatchClassification(ctx, owner, batchFetchClassification{RewriteProvider: options.NovelFetchRewrite, RewriteResolver: options.NovelFetchRewriteResolver})
			if provider == nil {
				return "", batchfactoryv11.ErrUnavailable
			}
			return provider.Complete(ctx, request)
		}}).retryFailedComicPages
		registerBookStageRoutes(v11, stageService, retryComicPages)
	} else if options.Store != nil {
		registerBookStageSummaryRoute(v11, &batchfactoryv11.BookStageService{Store: options.Store}, "v11")
	}
	if options.Compiler != nil {
		// The compiler preview is read-only. Keep the V11 route available even
		// before the staged V11 production slices so an already-open legacy page
		// can refresh its final prompt instead of falling back to stale raw text.
		registerCompilerRoutes(v11, options.Compiler)
	}
	// The V12 workbench uses these endpoints for read-only preview data. Keep
	// them available independently of the staged write/production slices so a
	// local V12 batch does not fall back to the Node compatibility bridge.
	if options.Compiler != nil {
		registerV12CompilerRoutes(v12, options.Compiler)
	}
	if options.Production != nil {
		registerProductionStatusRoute(v12, options.Production, "v12")
	}
	// Merge status is a persisted read and does not require the external merge
	// adapter. Keep it available in the early slices without advertising merge
	// submission as enabled.
	mergeStatus := options.Merge
	if mergeStatus == nil && options.Store != nil {
		mergeStatus = &batchfactoryv11.MergeService{Store: options.Store}
	}
	if mergeStatus != nil {
		registerMergeStatusRoute(v12, mergeStatus, "v12")
	}
	if options.Production != nil && options.Slice >= 4 {
		registerProductionRoutes(v11, options.Production, runtimeCatalogStore)
		if output, ok := options.MergeOutput.(interface {
			PutUploadedMP4(context.Context, string, string) (string, error)
		}); ok {
			registerVideoUploadRoutes(v11, options.Production, options.Merge, options.LocalArtifacts, output)
		}
	}
	// Completed local merge artifacts must remain readable in the ordinary
	// review runtime. This only exposes owner-scoped GET routes; merge writes
	// remain gated by the production slice below.
	if mergeStatus != nil {
		var mergeReader mergeworker.ObjectReader
		if candidate, ok := options.MergeOutput.(mergeworker.ObjectReader); ok {
			mergeReader = candidate
		}
		registerMergeRoutes(v11, mergeStatus, options.LocalArtifacts, options.MergeOutput, mergeReader, options.Merge != nil && options.Slice >= 5)
		registerMergeMediaTokenRoute(v11, "v11", mergeStatus, options.BridgeSecret, options.Now)
		registerMergeMediaTokenRoute(v12, "v12", mergeStatus, options.BridgeSecret, options.Now)
	}
	if options.External != nil && options.Slice >= 6 {
		registerExternalRoutes(v11, options.External)
	}
	if options.RegisterV11 != nil {
		options.RegisterV11(v11)
	}
	if options.Store != nil {
		registerV12H3TraceRoute(v12, &batchfactoryv11.H3KernelService{Store: options.Store, AudioProbe: options.H3AudioProbe})
	}
	if options.Store != nil && options.Slice >= 3 {
		registerV12H3Routes(v12, &batchfactoryv11.H3KernelService{Store: options.Store, AudioProbe: options.H3AudioProbe}, options.Director)
	}
	registerV12V11Aliases(v12, v11)

	if options.RuntimeAudit == nil {
		options.RuntimeAudit = runtimeaudit.Default()
	}
	root := http.NewServeMux()
	root.HandleFunc("GET /api/runtime-build-info", func(w http.ResponseWriter, req *http.Request) {
		audit := options.RuntimeAudit.Snapshot()
		manifest := readRuntimeManifest(options.RuntimeManifestPath)
		writeJSON(w, http.StatusOK, map[string]any{
			"service":              "go-api",
			"runtime":              "go",
			"git_sha":              options.ReleaseSHA,
			"frontend_entry":       manifest.FrontendEntry,
			"runtime_manifest_sha": manifest.GitSHA,
			"bridge_hits":          audit.BridgeHits,
			"bridge_paths":         audit.Paths,
			"media_engine":         ffmpegbin.Current(),
		})
	})
	// Release notes ship inside this binary and are matched against the release
	// manifest, so a stale or mismatched release cannot show the wrong payload.
	registerProductUpdateRoute(root, options.RuntimeManifestPath, releasenotes.Default())
	auth := BridgeAuth{Secret: options.BridgeSecret, Now: options.Now, Users: options.Users, AuthStore: options.AuthStore, SessionResolver: options.SessionResolver}
	novelScope := novelFetchAccountScopeResolver(options.AuthStore, options.MemberProfileStore)
	novelCompatAuth := auth
	novelCompatAuth.IdentityScope = novelScope
	var mergeReader mergeworker.ObjectReader
	if candidate, ok := options.MergeOutput.(mergeworker.ObjectReader); ok {
		mergeReader = candidate
	}
	publish121 := &BatchFactory121Publisher{Store: options.Store, UsersDir: options.UserDataDir, LocalFiles: options.LocalArtifacts, MergedFiles: mergeReader, Connector: options.VideoManagementConnector}
	if options.AutomationCoordinator != nil {
		options.AutomationCoordinator.publish121 = publish121
	}
	registerBatchFactory121Publish(v11, auth, publish121)
	registerBatchFactory121Publish(v12, auth, publish121)
	// Capabilities are derived from the routes that were actually registered on
	// the V11 source mux. This keeps the advertised surface and the real
	// handlers from drifting apart as slices are enabled.
	gates := capabilityGates{
		ProductionEnabled: options.Production != nil && options.Production.Enabled,
		MergeEnabled:      options.Merge != nil && options.Merge.Enabled,
		Publish121:        publish121Available,
		PublishYadi:       publishYadiAvailable,
	}
	capabilities := deriveCapabilities(options.Slice, gates, v11, v12)
	v11.HandleFunc("GET /api/batch-factory/v11/capabilities", capabilityHandler(capabilities))
	v12.HandleFunc("GET /api/batch-factory/v12/capabilities", capabilityHandler(capabilities))
	enabledCapabilities, disabledCapabilities := capabilityBootSummary(capabilities)
	log.Printf("batch-factory capabilities slice=%d enabled=%d disabled=%v", options.Slice, enabledCapabilities, disabledCapabilities)
	if options.Store != nil {
		registerV12SmartUnifiedRefresh(v12, auth, options.Store, options.TextProvider)
	}
	if options.Store != nil {
		registerV12MetadataCompat(v12, auth, options.Store, options.NovelFetchFetcher, batchFetchClassification{
			AIGenderJudge: options.NovelFetchAIGenderJudge, AIGenderJudgeResolver: options.NovelFetchAIGenderJudgeResolver,
			RewriteProvider: options.NovelFetchRewrite, RewriteResolver: options.NovelFetchRewriteResolver,
		})
	}
	if options.AutomationPresetStore != nil {
		registerAutomationPresetRoutes(v12, auth, options.AutomationPresetStore, "v12")
		registerBatchAutomationPresetRoute(v12, auth, options.Store, options.AutomationPresetStore, "v12")
	}
	if options.AutomationStatusReader != nil {
		registerAutomationStatusRoute(v12, auth, options.AutomationStatusReader, "v12")
	}
	if options.AutomationCoordinator != nil {
		registerBatchAutomationRoutes(v12, auth, options.AutomationCoordinator, "v12")
	}
	RegisterBatchFactoryScheduleRoutes(v12, auth, options.UserDataDir, options.Now)
	var videoProviderRegistry batchfactoryv11.VideoProviderRegistry
	if options.Production != nil {
		videoProviderRegistry = options.Production.ProviderRegistry
	}
	registerVideoProviderStatusRoute(v12, auth, options.UserConfigStore, videoProviderRegistry, options.Store, "v12", runtimeCatalogStore)
	if options.Store != nil {
		runtimeStageService := &batchfactoryv11.BookStageService{Store: options.Store, Director: options.Director, Production: options.Production, ImageReady: shuihuoImageReadyChecker(options.Store, runtimeCatalogStore, options.UserConfigStore, resolveRuntimeModelOwner), VideoReady: shuihuoVideoReadyChecker(options.Store, runtimeCatalogStore, options.Production)}
		registerRuntimeSummaryRoute(v12, auth, options.Store, runtimeStageService, options.Production, mergeStatus, options.AutomationStatusReader, "v12")
	}
	mfaFile := ""
	if options.MemberProfileStore != nil {
		mfaFile = filepath.Join(options.MemberProfileStore.SystemDir, "mfa.json")
	}
	passkeyFile := ""
	if options.MemberProfileStore != nil {
		passkeyFile = filepath.Join(options.MemberProfileStore.SystemDir, "passkeys.json")
	}
	rateLimitNow := options.Now
	if rateLimitNow == nil {
		rateLimitNow = time.Now
	}
	authLimiter := newAuthRateLimiter(options.AuthRateLimits, rateLimitNow)
	RegisterAuthRoutes(root, AuthOptions{Store: options.AuthStore, Profiles: options.MemberProfileStore, Now: options.Now, SessionResolver: options.SessionResolver, MFAFile: mfaFile, PasskeyFile: passkeyFile, RateLimiter: authLimiter})
	if options.PresetCatalog != nil {
		registerPresetRoutes(root, auth, options.PresetCatalog, options.LocalArtifacts)
	}
	if options.PersonalPromptStore != nil {
		registerPersonalPromptRoutes(root, auth, options.PersonalPromptStore)
	}
	RegisterScriptHistoryRoutes(root, auth, options.ScriptHistoryStore)
	RegisterMemberProfileRoutes(root, auth, options.MemberProfileStore, options.AuthStore, options.MemberAvatarDir)
	RegisterMFARoutes(root, auth, options.MemberProfileStore, options.AuthStore)
	RegisterAccountAdminReadCompat(root, auth, options.AuthStore, options.MemberProfileStore)
	RegisterAccountTransferCompat(root, auth, options.AuthStore, options.MemberProfileStore)
	RegisterMemberTeamReadCompat(root, auth, options.AuthStore, options.MemberProfileStore)
	RegisterMemberTeamWriteCompat(root, auth, options.AuthStore, options.MemberProfileStore)
	RegisterTeamAdminCompat(root, auth, options.AuthStore, options.MemberProfileStore)
	if options.MemberProfileStore != nil {
		RegisterPublicTeamInviteCompat(root, options.MemberProfileStore.SystemDir, options.AuthStore, options.MemberProfileStore)
	}
	prompt.SetLuoshuiOverrideCatalog(options.PresetCatalog)
	RegisterAdminReadCompat(root, auth, options.AuthStore, options.MemberProfileStore, options.PresetCatalog, options.LocalArtifacts)
	RegisterAdminGovernanceCompat(root, auth, options.AuthStore, options.MemberProfileStore)
	RegisterApplicationsCompat(root, options.AuthStore, options.MemberProfileStore, options.Now, authLimiter)
	RegisterClientErrorCompat(root, auth, options.MemberProfileStore)
	if options.MemberProfileStore != nil {
		RegisterAccountRecoveryStatusCompat(root, auth, options.MemberProfileStore.SystemDir, authLimiter)
		RegisterAccountPasskeyCompat(root, auth, options.MemberProfileStore, options.AuthStore)
		RegisterAccountPurgeCompat(root, auth, options.AuthStore, options.MemberProfileStore, options.MemberAvatarDir)
	}
	RegisterReferenceAssetRoutes(root, auth, options.UserDataDir)
	registerReferenceAssetImageGenerateRoute(root, auth, options.UserConfigStore, runtimeCatalogStore, options.LocalExecutors)
	registerUserConfigRoutes(root, auth, options.UserConfigStore, options.AuthStore, options.NovelFetchStore, options.MemberProfileStore)
	registerAPIIntegrationRoutes(root, auth, options.UserConfigStore, options.AuthStore, options.MemberProfileStore)
	RegisterStorageRoutes(root, auth, options.UserConfigStore, options.ScriptHistoryStore)
	RegisterPlatformProjectsRoute(root, auth, options.ScriptHistoryStore, options.Store)
	RegisterNovelFetchUploadSessionCompat(root, novelCompatAuth, options.UserDataDir, options.VideoManagementConnector)
	RegisterNovelFetchUploadBatchCompat(root, novelCompatAuth, options.UserDataDir, options.NovelFetchStore, options.VideoManagementConnector)
	RegisterWebSubmitTargetCheckCompat(root, novelCompatAuth, options.UserDataDir, options.VideoManagementConnector)
	publish121OrganizationsHandler(root, novelCompatAuth, options.UserDataDir, options.VideoManagementConnector)
	if options.AgentWorkspace != nil {
		RegisterAgentGenerationHistoryRoutes(root, auth, options.AgentWorkspace, options.AgentAssets, options.Now)
		RegisterAgentMaterialLibraryRoutes(root, auth, options.AgentWorkspace, options.AgentAssets, options.Now)
	}
	if options.AgentProxy != nil {
		RegisterAgentProxyRoutes(root, auth, options.AgentProxy, options.Now)
	}
	if options.UserConfigStore != nil {
		RegisterAgentWorkspaceCompat(root, auth, options.AgentWorkspace, options.UserConfigStore.BaseDir, options.TextProvider, options.Now)
		if batchFactoryBridgeStore, ok := options.Store.(agentBatchFactoryBridgeStore); ok {
			if projectionStore, projectionOK := options.AgentWorkspace.(agentworkspace.BatchFactoryProjectionStore); projectionOK {
				var mergeMedia mergeworker.ObjectReader
				if candidate, ok := options.MergeOutput.(mergeworker.ObjectReader); ok {
					mergeMedia = candidate
				}
				projection := &agentworkspace.BatchFactoryProjectionService{
					CanvasStore: projectionStore, BatchFactory: batchFactoryBridgeStore,
					MergeStatus:    mergeStatus,
					AudioRecords:   newBatchFactoryProductionAudioReader(batchFactoryBridgeStore, options.LocalArtifacts),
					PublishResults: newBatchFactoryPublishResults(options.External),
					Now:            options.Now,
				}
				RegisterAgentBatchFactoryBridge(root, auth, options.AgentWorkspace, batchFactoryBridgeStore, projection, options.Now, agentBatchFactoryBridgeMedia{Local: options.LocalArtifacts, Remote: mergeMedia})
			}
		}
		RegisterAgentAssetRoutes(root, auth, options.AgentWorkspace, options.AgentAssets, options.AgentAssetObjects, options.Now)
		RegisterAgentExecutorReferenceAssetRoutes(root, options.LocalExecutors, options.AgentAssets, options.AgentAssetObjects)
		RegisterAgentGenerationTaskRoutes(root, auth, options.AgentWorkspace, options.Now)
		if options.AgentAssets != nil && options.AgentAssetObjects != nil {
			audioServices := AgentAudioServices{
				Projects: options.AgentWorkspace, Assets: options.AgentAssets, Objects: options.AgentAssetObjects,
				Catalog:   runtimeCatalogStore,
				Generator: &configuredAgentAudioGenerator{catalog: runtimeCatalogStore}, Processor: &FFmpegAgentAudioProcessor{}, Waveform: &FFmpegAgentAudioWaveform{}, QueueAudio: true, Now: options.Now,
			}
			RegisterAgentAudioJobRoutes(root, auth, audioServices)
			if options.AgentAudioWorkerContext != nil {
				go newAgentAudioCoordinator(audioServices).Run(options.AgentAudioWorkerContext)
			}
			videoToolServices := AgentVideoToolServices{
				Projects: options.AgentWorkspace, Assets: options.AgentAssets, Objects: options.AgentAssetObjects,
				Processor: &FFmpegAgentVideoProcessor{}, Now: options.Now,
			}
			RegisterAgentVideoToolRoutes(root, auth, videoToolServices)
			if options.AgentVideoWorkerContext != nil {
				go newAgentVideoToolCoordinator(videoToolServices).Run(options.AgentVideoWorkerContext)
			}
		}
		textModels := options.AgentTextModelResolver
		if textModels == nil && options.NovelFetchStore != nil {
			textModels = func(ctx context.Context, owner, modelID string) (batchfactoryv11.DirectorProvider, error) {
				return resolveAgentTextModel(ctx, runtimeCatalogStore, options.UserConfigStore, resolveRuntimeModelOwner(owner), modelID)
			}
		}
		var imageReferences agentworkspace.ReferenceURLProvider
		if options.AgentAssetObjects != nil {
			imageReferences, _ = options.AgentAssetObjects.(agentworkspace.ReferenceURLProvider)
		}
		imageGenerate := newAgentImageGenerator(options.Store, runtimeCatalogStore, options.UserConfigStore, options.LocalExecutors, options.LocalArtifacts, options.AgentAssets, imageReferences)
		textComplete := newAgentTextCompletionExecutor(textModels)
		imageServices := AgentNodeGenerationServices{Projects: options.AgentWorkspace, Catalog: runtimeCatalogStore, TextModels: textModels, TextComplete: textComplete, Assets: options.AgentAssets, Objects: options.AgentAssetObjects, ImageGenerate: imageGenerate, ImageProcess: NewAgentImageProcessor(), QueueImages: true, Now: options.Now}
		RegisterAgentNodeGenerationRoutes(root, auth, imageServices)
		if options.AgentImageWorkerContext != nil {
			go newAgentImageCoordinator(imageServices).Run(options.AgentImageWorkerContext)
		}
		var videoCoordinator *agentVideoCoordinator
		if options.NovelFetchStore != nil && options.AgentWorkspace != nil && options.AgentAssets != nil {
			var references agentworkspace.ReferenceURLProvider
			if options.AgentAssetObjects != nil {
				references, _ = options.AgentAssetObjects.(agentworkspace.ReferenceURLProvider)
			}
			videoResolver := func(ctx context.Context, owner, modelKey string) (batchfactoryv11.VideoProviderConfig, error) {
				config, err := loadShuihuoOwnerConfig(ctx, options.Store, owner, runtimeCatalogStore)
				if err != nil {
					return batchfactoryv11.VideoProviderConfig{}, err
				}
				return scriptVideoModelConfig(ctx, options.Store, owner, config, modelKey)
			}
			videoCoordinator = &agentVideoCoordinator{
				projects: options.AgentWorkspace, catalog: runtimeCatalogStore, assets: options.AgentAssets, references: references,
				videos:   &scriptVideoService{resolver: videoResolver, localExecutor: options.LocalExecutors},
				ingestor: newAgentVideoAssetIngestor(options.AgentAssets, options.AgentAssetObjects, nil, options.Now),
				executor: options.LocalExecutors, artifacts: options.LocalArtifacts, now: options.Now,
			}
			RegisterAgentVideoJobRoutes(root, auth, videoCoordinator)
			if options.AgentVideoWorkerContext != nil {
				go videoCoordinator.Run(options.AgentVideoWorkerContext)
			}
		}
		scriptShotServices := AgentScriptShotServices{Projects: options.AgentWorkspace, TextModels: textModels, TextComplete: textComplete, Models: agentConfiguredModelResolver(runtimeCatalogStore), Dispatch: newAgentShotGenerationDispatcher(root, auth, options.Now, videoCoordinator), Now: options.Now}
		RegisterAgentScriptShotRoutes(root, auth, scriptShotServices)
		if options.AgentScriptSplitWorkerContext != nil {
			go newAgentScriptSplitCoordinator(scriptShotServices).Run(options.AgentScriptSplitWorkerContext)
		}
		systemDir := ""
		if options.MemberProfileStore != nil {
			systemDir = options.MemberProfileStore.SystemDir
		}
		RegisterAgentSkillsCompat(root, auth, systemDir, options.UserConfigStore.BaseDir)
		RegisterAgentChatRoutes(root, auth, options.UserConfigStore.BaseDir, systemDir, runtimeCatalogStore, options.UserConfigStore, textModels, options.Now)
	}
	RegisterAICutRoutes(root, auth, options.AICut, options.Now)
	var screenplayMedia screenplay.BatchMediaRunner
	if options.UserConfigStore != nil {
		videoResolver := func(ctx context.Context, owner, modelKey string) (batchfactoryv11.VideoProviderConfig, error) {
			config, err := options.UserConfigStore.GetConfig(ctx, owner)
			if err != nil {
				return batchfactoryv11.VideoProviderConfig{}, err
			}
			return scriptVideoModelConfig(ctx, options.Store, owner, config, modelKey)
		}
		screenplayMedia = &screenplayVideoRunner{
			service: &scriptVideoService{resolver: videoResolver, localExecutor: options.LocalExecutors},
			configs: options.UserConfigStore,
		}
	}
	RegisterScreenplayRoutes(root, auth, options.Screenplay, runtimeTextProviderResolver, options.Now, screenplayMedia)
	RegisterNovelCreationRoutes(root, auth, options.NovelCreation, options.NovelCreationRunStarter)
	RegisterAICutAssetRoutes(root, auth, options.AICut, options.AgentAssetObjects, options.Now)
	RegisterAICutTTSRoutes(root, auth, options.AICut, options.AgentAssetObjects, runtimeCatalogStore, options.Now)
	RegisterAICutBackgroundRemovalRoutes(root, auth, options.AICut, options.AgentAssetObjects, runtimeCatalogStore, nil, options.Now)
	RegisterAICutVoiceCloneRoutes(root, auth, options.AICut, options.AgentAssetObjects, runtimeCatalogStore, nil, options.Now)
	RegisterAICutExportRoutes(root, auth, options.AICut, options.AgentAssetObjects, options.AICutExport)
	RegisterAICutMediaToolRoutes(root, auth, options.AICut, options.AgentAssetObjects, options.Now)
	RegisterAICutAIRoutes(root, auth, options.AICut, runtimeTextProviderResolver, options.Now)
	if editPlans, ok := options.AICut.(aicut.EditPlanStore); ok {
		RegisterAICutEditPlanRoutes(root, auth, options.AICut, editPlans, runtimeTextProviderResolver, options.Now)
	}
	RegisterAICutASRRoutes(root, auth, options.AICut, options.AICutASR)
	RegisterAICutAgentRoutes(root, auth, options.AICut, options.AICutAgentStore, options.AICutAgentStarter, options.Now)
	RegisterUpscaleRoutes(root, auth, options.Upscale)
	if options.TextProvider != nil {
		pluginTextModels := options.AgentTextModelResolver
		if pluginTextModels == nil && options.NovelFetchStore != nil {
			pluginTextModels = func(ctx context.Context, owner, modelID string) (batchfactoryv11.DirectorProvider, error) {
				return resolveAgentTextModel(ctx, runtimeCatalogStore, options.UserConfigStore, resolveRuntimeModelOwner(owner), modelID)
			}
		}
		registerScriptSmartUnifiedRoute(root, auth, options.TextProvider)
		registerChatRoute(root, auth, options.TextProvider, runtimeTextProviderResolver, options.PresetCatalog, options.PersonalPromptStore)
		registerPluginToolRoutes(root, auth, options.TextProvider, newAgentTextCompletionExecutor(pluginTextModels))
		registerScriptDirectorPipelineCompat(root, auth, options.TextProvider, options.PresetCatalog, options.PersonalPromptStore)
	}
	v11Root := auth.Middleware(adminBatchReadThrough(capabilityNotFoundHandler(v11), options.Store, options.MemberProfileStore, options.AuthStore))
	v12Root := auth.Middleware(adminBatchReadThrough(capabilityNotFoundHandler(v12), options.Store, options.MemberProfileStore, options.AuthStore))
	if legacyV11Public() {
		root.Handle("/api/batch-factory/v11/", v11Root)
	}
	root.Handle("/api/batch-factory/v12/", v12Root)
	// A native <video> element cannot attach the Authorization header, so the
	// signed-token playback path is registered as an exact root route. Requests
	// without a token keep falling through to the normal authenticated handler.
	if mergeStatus != nil {
		registerMergeMediaPublicRoute(root, "v12", v12Root, mergeStatus, options.LocalArtifacts, mergeReader, options.BridgeSecret, options.Now)
		if legacyV11Public() {
			registerMergeMediaPublicRoute(root, "v11", v11Root, mergeStatus, options.LocalArtifacts, mergeReader, options.BridgeSecret, options.Now)
		}
	}
	if options.NovelFetchStore != nil {
		RegisterShuihuoAccountAIConfigCompat(root, auth, options.NovelFetchStore)
	}
	registerShuihuoProjectCompatRoutes(root, auth, options.Store, options.LocalArtifacts, options.Director, options.TextProvider, options.Production, runtimeCatalogStore, options.PresetCatalog, options.RedisURL, options.LocalExecutors, options.UserConfigStore)
	RegisterBatchComicReferenceRoutes(root, auth, options.Store, options.LocalArtifacts, options.LocalExecutors, options.PresetCatalog)
	registerTTSRoute(root, auth, runtimeCatalogStore)
	RegisterLocalExecutorRoutes(root, auth, options.LocalExecutors)
	RegisterLocalExecutorJobRoutes(root, auth, options.LocalExecutors)
	RegisterLocalExecutorArtifactRoutes(root, auth, options.LocalExecutors, options.LocalArtifacts)
	if options.NovelFetchStore != nil {
		RegisterScriptVideoRoutes(root, auth, options.LocalExecutors, func(ctx context.Context, owner, modelKey string) (batchfactoryv11.VideoProviderConfig, error) {
			config, err := loadShuihuoOwnerConfig(ctx, options.Store, owner, runtimeCatalogStore)
			if err != nil {
				return batchfactoryv11.VideoProviderConfig{}, err
			}
			return scriptVideoModelConfig(ctx, options.Store, owner, config, modelKey)
		})
	} else {
		RegisterLocalScriptVideoRoutes(root, auth, options.LocalExecutors)
	}
	RegisterLocalExecutorDownloadRoutes(root, options.LocalExecutorDownloadsDir)
	RegisterGiantMaterialExecutorRoutes(root, auth, options.GiantMaterialExecutor)
	registerExecutorStatusRoute(root, auth, options.LocalExecutors, options.GiantMaterialExecutor)
	RegisterGiantMaterialResolverRoutes(root, auth, options.GiantMaterialResolver)
	if options.NovelFetchStore != nil {
		novelFetch := http.NewServeMux()
		registerNovelFetchWorkshopRoutes(novelFetch, options.NovelFetchStore, options.NovelFetchQueue, options.NovelFetchAIGenderJudge, options.NovelFetchAIGenderJudgeResolver, options.NovelFetchFetcher, options.NovelFetchRewrite, options.NovelFetchRewriteResolver, options.UserDataDir, options.VideoManagementConnector, options.MemberProfileStore, options.AuthStore)
		registerNovelFetchLegacyWorkflowRoutes(novelFetch, options.NovelFetchStore, options.NovelFetchQueue, options.UserDataDir, func(req *http.Request) (novelfetchworkshop.RewriteProvider, error) {
			if options.NovelFetchRewriteResolver == nil {
				return options.NovelFetchRewrite, nil
			}
			identity, _ := BridgeIdentityFromContext(req.Context())
			return options.NovelFetchRewriteResolver(req.Context(), identity.Username)
		}, options.VideoManagementConnector)
		registerNovelFetchCleanupRoutes(novelFetch, options.NovelFetchStore, options.Now)
		novelAuth := NovelFetchBridgeAuth{Secret: options.BridgeSecret, Now: options.Now, Users: options.Users, AuthStore: options.AuthStore, SessionResolver: options.SessionResolver, IdentityScope: novelScope}
		// Legacy web-submit aliases must not mint identity from signed headers.
		// Keep the same store/session resolver but disable the HMAC path.
		sessionOnlyNovelAuth := novelAuth
		sessionOnlyNovelAuth.Secret = ""
		// Web-submit endpoints only trust a real login session. The old
		// /api/web-submit/* aliases used to accept self-signed identity headers.
		sessionOnlyCompatAuth := novelCompatAuth.SessionOnly()
		root.Handle("/api/novel-fetch-workshop/", novelAuth.Middleware(novelFetch))
		RegisterBatchRewriteConfigCompat(root, novelCompatAuth, options.NovelFetchStore, options.UserDataDir, options.VideoManagementConnector)
		RegisterWebSubmitPreviewCompat(root, sessionOnlyCompatAuth, options.NovelFetchStore)
		RegisterWebSubmitSubmitCompat(root, sessionOnlyCompatAuth, options.UserDataDir, options.NovelFetchStore, func(ctx context.Context, owner string) (novelfetchworkshop.AIGenderJudge, error) {
			if options.NovelFetchAIGenderJudgeResolver == nil {
				return options.NovelFetchAIGenderJudge, nil
			}
			return options.NovelFetchAIGenderJudgeResolver(ctx, owner)
		}, func(ctx context.Context, owner string) (novelfetchworkshop.RewriteProvider, error) {
			if options.NovelFetchRewriteResolver == nil {
				return options.NovelFetchRewrite, nil
			}
			return options.NovelFetchRewriteResolver(ctx, owner)
		}, options.VideoManagementConnector)
		root.Handle("/api/batch-rewrite/tasks/", novelAuth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			clone := req.Clone(req.Context())
			clone.URL.Path = strings.Replace(clone.URL.Path, "/api/batch-rewrite/tasks/", "/api/novel-fetch-workshop/tasks/", 1)
			novelFetch.ServeHTTP(w, clone)
		})))
		legacyWorkshopAlias := func(w http.ResponseWriter, req *http.Request) {
			clone := req.Clone(req.Context())
			clone.URL.Path = strings.Replace(clone.URL.Path, "/api/batch-rewrite/", "/api/novel-fetch-workshop/", 1)
			novelFetch.ServeHTTP(w, clone)
		}
		legacyWebSubmitAlias := func(w http.ResponseWriter, req *http.Request) {
			clone := req.Clone(req.Context())
			stripBridgeIdentityHeaders(clone)
			clone.URL.Path = strings.Replace(clone.URL.Path, "/api/web-submit/", "/api/novel-fetch-workshop/web-submit/", 1)
			novelFetch.ServeHTTP(w, clone)
		}
		legacyKnowledgeSummaryAlias := func(w http.ResponseWriter, req *http.Request) {
			clone := req.Clone(req.Context())
			clone.URL.Path = "/api/novel-fetch-workshop/knowledge/summary"
			novelFetch.ServeHTTP(w, clone)
		}
		legacyRulesSuggestAlias := func(w http.ResponseWriter, req *http.Request) {
			clone := req.Clone(req.Context())
			clone.URL.Path = strings.Replace(clone.URL.Path, "/api/batch-rewrite/rules/ai-suggest", "/api/novel-fetch-workshop/rules/suggest", 1)
			novelFetch.ServeHTTP(w, clone)
		}
		legacyRulesPreviewAlias := func(w http.ResponseWriter, req *http.Request) {
			clone := req.Clone(req.Context())
			clone.URL.Path = strings.Replace(clone.URL.Path, "/api/batch-rewrite/rules/preview", "/api/novel-fetch-workshop/rules/preview", 1)
			novelFetch.ServeHTTP(w, clone)
		}
		// These legacy compatibility handlers remain Go-owned, but the workshop
		// auth middleware must stay in front of them because the canonical page
		// signs requests with its workshop session, not the old bridge headers.
		canonicalWorkshopLegacyAlias := func(w http.ResponseWriter, req *http.Request) {
			clone := req.Clone(req.Context())
			clone.URL.Path = strings.Replace(clone.URL.Path, "/api/novel-fetch-workshop/", "/api/batch-rewrite/", 1)
			root.ServeHTTP(w, clone)
		}
		for _, route := range []struct {
			method string
			path   string
		}{
			{http.MethodGet, "/api/novel-fetch-workshop/bootstrap"},
			{http.MethodGet, "/api/novel-fetch-workshop/knowledge"},
			{http.MethodPost, "/api/novel-fetch-workshop/work-form"},
		} {
			root.Handle(route.method+" "+route.path, novelAuth.Middleware(http.HandlerFunc(canonicalWorkshopLegacyAlias)))
		}
		root.Handle("/api/batch-rewrite/knowledge/", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("/api/batch-rewrite/knowledge", novelAuth.Middleware(http.HandlerFunc(legacyKnowledgeSummaryAlias)))
		root.Handle("/api/batch-rewrite/opening/", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("/api/batch-rewrite/ai/", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("/api/batch-rewrite/process", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("/api/batch-rewrite/process/", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("/api/batch-rewrite/rules/ai-suggest", novelAuth.Middleware(http.HandlerFunc(legacyRulesSuggestAlias)))
		root.Handle("/api/batch-rewrite/rules/preview", novelAuth.Middleware(http.HandlerFunc(legacyRulesPreviewAlias)))
		root.Handle("GET /api/batch-rewrite/logs", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("GET /api/batch-rewrite/schedules", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("POST /api/batch-rewrite/schedules", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("GET /api/batch-rewrite/batches/current", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		// The V78 page polls this endpoint for the small queue summary shown
		// above the task table. Keep it on Go as well; without the alias the
		// page reports a misleading 404 even though the task list is available.
		root.Handle("GET /api/batch-rewrite/realtime/status", novelAuth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			identity, ok := BridgeIdentityFromContext(req.Context())
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
			documents, err := options.NovelFetchStore.ListDocuments(req.Context(), identity.Username)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取任务队列失败"})
				return
			}
			counts := map[string]int{"running": 0, "queued": 0, "waiting_retry": 0, "failed": 0, "stopped": 0}
			for _, document := range documents {
				// AI bodies are durable evidence of completion. Do not keep the
				// queue summary stuck on an old in-flight aiStatus after versions
				// have been saved.
				task := legacyNovelFetchTask(document)
				status := strings.ToLower(firstText(task["ai_status"], document.Meta["status"], document.Meta["taskStatus"], document.Meta["aiStatus"]))
				switch {
				case strings.Contains(status, "running") || strings.Contains(status, "processing"):
					counts["running"]++
				case strings.Contains(status, "queued") || strings.Contains(status, "queue"):
					counts["queued"]++
				case strings.Contains(status, "retry"):
					counts["waiting_retry"]++
				case strings.Contains(status, "failed") || strings.Contains(status, "error"):
					counts["failed"]++
				case strings.Contains(status, "stopped") || strings.Contains(status, "cancel"):
					counts["stopped"]++
				}
			}
			writeJSON(w, http.StatusOK, map[string]any{"counts": counts, "updatedAt": time.Now().UTC().Format(time.RFC3339Nano)})
		})))
		root.Handle("POST /api/batch-rewrite/tasks/apply-rules", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("POST /api/batch-rewrite/tasks/reprocess-sensitive", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("POST /api/batch-rewrite/tasks/batch-ai-count", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("POST /api/batch-rewrite/knowledge/optimize", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("POST /api/batch-rewrite/web-submit/operations", sessionOnlyNovelAuth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			clone := req.Clone(req.Context())
			stripBridgeIdentityHeaders(clone)
			legacyWorkshopAlias(w, clone)
		})))
		root.Handle("GET /api/batch-rewrite/web-submit/operations/", sessionOnlyNovelAuth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			clone := req.Clone(req.Context())
			stripBridgeIdentityHeaders(clone)
			legacyWorkshopAlias(w, clone)
		})))
		root.Handle("POST /api/web-submit/operations", sessionOnlyNovelAuth.Middleware(http.HandlerFunc(legacyWebSubmitAlias)))
		root.Handle("GET /api/web-submit/operations/", sessionOnlyNovelAuth.Middleware(http.HandlerFunc(legacyWebSubmitAlias)))
		root.Handle("PATCH /api/batch-rewrite/schedules/", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("DELETE /api/batch-rewrite/schedules/", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		// V78 exposed these process/task paths directly. Keep them as real Go
		// aliases so an older page does not render controls that fall through to
		// a 404 while the canonical page uses the workshop namespace.
		legacyProcessAlias := func(w http.ResponseWriter, req *http.Request) {
			clone := req.Clone(req.Context())
			clone.URL.Path = strings.Replace(clone.URL.Path, "/api/process", "/api/novel-fetch-workshop/process", 1)
			novelFetch.ServeHTTP(w, clone)
		}
		root.Handle("POST /api/process", novelAuth.Middleware(http.HandlerFunc(legacyProcessAlias)))
		root.Handle("POST /api/process/start", novelAuth.Middleware(http.HandlerFunc(legacyProcessAlias)))
		root.Handle("POST /api/process/stop", novelAuth.Middleware(http.HandlerFunc(legacyProcessAlias)))
		root.Handle("POST /api/process/queue/start", novelAuth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// The Go worker's automation switch is the durable replacement for the
			// old in-memory manual queue starter.
			clone := req.Clone(req.Context())
			clone.URL.Path = "/api/novel-fetch-workshop/automation/start"
			novelFetch.ServeHTTP(w, clone)
		})))
		root.Handle("GET /api/process/queue/status", novelAuth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			clone := req.Clone(req.Context())
			clone.URL.Path = "/api/novel-fetch-workshop/realtime/status"
			novelFetch.ServeHTTP(w, clone)
		})))
		legacyTaskAlias := func(w http.ResponseWriter, req *http.Request) {
			clone := req.Clone(req.Context())
			path := clone.URL.Path
			// The old detail tabs used separate summary/content/diagnostic URLs;
			// the Go document endpoint already contains the same durable payload.
			for _, suffix := range []string{"/summary", "/content"} {
				if strings.HasSuffix(path, suffix) {
					path = strings.TrimSuffix(path, suffix)
					break
				}
			}
			if strings.HasSuffix(path, "/rules-trace") || strings.HasSuffix(path, "/sensitive-log") || strings.HasSuffix(path, "/site-submit-log") {
				clone.URL.Path = strings.Replace(path, "/api/tasks", "/api/batch-rewrite/tasks", 1)
				root.ServeHTTP(w, clone)
				return
			}
			clone.URL.Path = strings.Replace(path, "/api/tasks", "/api/novel-fetch-workshop/tasks", 1)
			novelFetch.ServeHTTP(w, clone)
		}
		root.Handle("GET /api/tasks", novelAuth.Middleware(http.HandlerFunc(legacyTaskAlias)))
		root.Handle("GET /api/tasks/", novelAuth.Middleware(http.HandlerFunc(legacyTaskAlias)))
		root.Handle("POST /api/tasks/", novelAuth.Middleware(http.HandlerFunc(legacyTaskAlias)))
		root.Handle("PUT /api/tasks/", novelAuth.Middleware(http.HandlerFunc(legacyTaskAlias)))
		root.Handle("DELETE /api/tasks/", novelAuth.Middleware(http.HandlerFunc(legacyTaskAlias)))
		root.Handle("/api/web-submit/", sessionOnlyNovelAuth.Middleware(http.HandlerFunc(legacyWebSubmitAlias)))
		registerCanonicalNovelFetchWebSubmitAliases(root, auth)
		registerModelCatalogRoutes(root, novelAuth, options.NovelFetchStore, options.UserConfigStore, options.AuthStore, options.MemberProfileStore)
	}
	root.HandleFunc("GET /health", func(w http.ResponseWriter, req *http.Request) {
		if options.Health != nil && options.Health(req.Context()) != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	// Compress JSON responses at the outermost layer so every registered route
	// benefits without touching individual handlers. Non-JSON streams (images,
	// ServeContent Range/304) are passed through untouched.
	return withJSONCompression(root)
}

// registerCanonicalNovelFetchWebSubmitAliases keeps the public workshop
// namespace on Go while reusing the tested compatibility handlers.
//
// The alias used to mint a bridge signature from caller-supplied X-Qiantie-*
// headers, which let anonymous requests impersonate any username. It now
// requires a real Bearer session and never derives identity from headers:
// the HMAC header path is disabled for this alias and every X-Qiantie- identity
// header is stripped before the internal hop.
func registerCanonicalNovelFetchWebSubmitAliases(root *http.ServeMux, bridge BridgeAuth) {
	if root == nil {
		return
	}
	sessionOnly := bridge
	// An empty secret disables the signed-header path in BridgeAuth.authenticate,
	// so only a session token accepted by AuthStore/SessionResolver can pass.
	sessionOnly.Secret = ""
	root.Handle("/api/novel-fetch-workshop/web-submit/", sessionOnly.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		if _, err := SafeUsername(identity.Username); err != nil {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "invalid username"})
			return
		}
		clone := req.Clone(req.Context())
		clone.URL.Path = strings.Replace(clone.URL.Path, "/api/novel-fetch-workshop/web-submit/", "/api/batch-rewrite/web-submit/", 1)
		// The downstream compatibility route authenticates again. Drop headers
		// that could be mistaken for a trusted identity on that hop.
		for _, header := range []string{HeaderUsername, HeaderIsOwner, HeaderIssuedAt, HeaderSignature} {
			clone.Header.Del(header)
		}
		root.ServeHTTP(w, clone)
	})))
}

func shuihuoImageModelReady(model map[string]any) bool {
	if model == nil || !strings.EqualFold(modelString(model, "kind"), "image") {
		return false
	}
	if enabled, exists := model["enabled"].(bool); exists && !enabled {
		return false
	}
	endpoint := modelString(model, "endpoint")
	if endpoint == "" {
		endpoint = modelString(model, "baseUrl")
	}
	if _, err := shuihuoImageGenerationEndpoint(endpoint); err != nil {
		return false
	}
	return shuihuoModelCredential(model) != ""
}

func shuihuoImageReadyChecker(store batchfactoryv11.Store, catalogStore novelfetchworkshop.Store, userConfig *UserConfigStore, resolveOwner func(string) string) func(context.Context, string) (bool, string) {
	return func(ctx context.Context, owner string) (bool, string) {
		config, err := loadShuihuoOwnerConfig(ctx, store, owner, catalogStore)
		if err == nil {
			for _, model := range shuihuoCollection(config, "modelCatalog", nil) {
				if shuihuoImageModelReady(model) {
					return true, ""
				}
			}
		}
		// 出图既可能走旧 modelCatalog，也可能走 API 接口配置里的 custom:/channel:
		// 渠道。渠道必须和真实出图一样按“解析后的模型归属账号”读取，否则成员
		// 自己的 API 配置为空时会被整批判为“图片模型或图片生成服务未配置”，
		// 与成员实际使用的管理员渠道口径不一致。
		channelOwner := owner
		if resolveOwner != nil {
			if resolved := strings.TrimSpace(resolveOwner(owner)); resolved != "" {
				channelOwner = resolved
			}
		}
		if shuihuoImageChannelReady(ctx, userConfig, channelOwner) {
			return true, ""
		}
		if err != nil {
			return false, "图片模型配置读取失败"
		}
		return false, "图片模型或图片生成服务未配置"
	}
}

// shuihuoImageChannelReady 判断 API 接口配置里是否已有可用的图片渠道：
// 预置图片渠道填了 Key，或自定义渠道分到 image 组且地址与 Key 都齐全。
func shuihuoImageChannelReady(ctx context.Context, userConfig *UserConfigStore, owner string) bool {
	if userConfig == nil {
		return false
	}
	config, err := userConfig.GetConfig(ctx, owner)
	if err != nil {
		return false
	}
	stored := apiIntegrationStoredMap(config)
	values := apiIntegrationValues(stored)
	selection := apiIntegrationChannelSelection(stored)
	for _, preset := range integrationChannelPresets {
		if preset.Group != "image" {
			continue
		}
		credential := strings.TrimSpace(fmt.Sprint(values[preset.KeyField]))
		if credential == "" || credential == "<nil>" {
			continue
		}
		if strings.TrimSpace(integrationChannelBaseURL(preset, values)) != "" {
			return true
		}
	}
	for _, raw := range apiIntegrationChannels(stored) {
		channel, ok := raw.(map[string]any)
		if !ok || integrationChannelGroup(channel, selection) != "image" {
			continue
		}
		credential := strings.TrimSpace(fmt.Sprint(channel["apiKey"]))
		baseURL := strings.TrimSpace(fmt.Sprint(channel["baseUrl"]))
		if credential == "" || credential == "<nil>" || baseURL == "" || baseURL == "<nil>" {
			continue
		}
		return true
	}
	return false
}

func shuihuoVideoReadyChecker(store batchfactoryv11.Store, catalogStore novelfetchworkshop.Store, production *batchfactoryv11.ProductionService) func(context.Context, string) (bool, string) {
	return func(ctx context.Context, owner string) (bool, string) {
		if production == nil || !production.Enabled {
			return false, "视频生成服务未启用"
		}
		if production.LocalExecutor != nil {
			if err := production.LocalExecutor.EnsureAvailable(ctx, owner); err == nil {
				return true, ""
			}
		}
		config, err := loadShuihuoOwnerConfig(ctx, store, owner, catalogStore)
		if err != nil {
			return false, "视频模型配置读取失败"
		}
		if shuihuoModelKindConfigured(config, "video") {
			return true, ""
		}
		return false, "视频模型或视频生成服务未配置"
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
