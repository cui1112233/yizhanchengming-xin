package httpapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/batchfactoryv11/external"
	"qiantie/backend/internal/giantmaterialexecutor"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
	"qiantie/backend/internal/mergeworker"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/runtimeaudit"
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
	BridgeSecret                    string
	ReleaseSHA                      string
	RuntimeManifestPath             string
	RuntimeAudit                    *runtimeaudit.Registry
	Health                          func(context.Context) error
	Now                             func() time.Time
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
	AutomationCoordinator           *BatchAutomationCoordinator
	PersonalPromptStore             *prompt.PersonalStore
	ScriptHistoryStore              *ScriptHistoryStore
	MemberProfileStore              *MemberProfileStore
	MemberAvatarDir                 string
	UserDataDir                     string
	UserConfigStore                 *UserConfigStore
	TextProvider                    batchfactoryv11.DirectorProvider
	AgentTextModelResolver          AgentTextModelResolver
	AgentWorkspace                  agentworkspace.Store
	AgentAssets                     agentworkspace.AssetStore
	AgentAssetObjects               agentworkspace.AssetObjectStore
	RedisURL                        string
	LocalExecutors                  *localexecutor.Service
	GiantMaterialExecutor           *giantmaterialexecutor.Service
	GiantMaterialResolver           GiantMaterialResolver
	LocalArtifacts                  *localartifact.Store
	LocalExecutorDownloadsDir       string
	MergeOutput                     mergeworker.ObjectStore
	VideoManagementConnector        videomanagementconnector.Service
}

func NewRouter(options RouterOptions) http.Handler {
	if options.Production == nil && options.Store != nil {
		// Read-only status is safe before the production slice is enabled. Keep
		// the service disabled for all write/submit operations.
		options.Production = &batchfactoryv11.ProductionService{Store: options.Store, Compiler: options.Compiler}
	}
	if options.AutomationCoordinator == nil && options.Store != nil && options.Director != nil && options.AutomationStatusReader != nil {
		var providerSync func(context.Context, string) error
		if options.Production != nil && options.Production.ProviderRegistry != nil {
			providerSync = func(ctx context.Context, owner string) error {
				return syncShuihuoConfiguredVideoModels(ctx, options.Store, options.Production.ProviderRegistry, owner, options.NovelFetchStore)
			}
		}
		options.AutomationCoordinator = NewBatchAutomationCoordinator(options.Store, options.AutomationPresetStore, &batchfactoryv11.BookStageService{
			Store: options.Store, Director: options.Director, Production: options.Production,
		}, options.Merge, options.External, options.NovelFetchStore, options.H3AudioProbe, options.AutomationStatusReader, options.Now, providerSync)
		options.AutomationCoordinator.SetLocalFiles(options.LocalArtifacts)
		options.AutomationCoordinator.SetSourceFetcher(options.NovelFetchFetcher)
	}
	if options.AutomationCoordinator != nil {
		if err := options.AutomationCoordinator.RecoverInterruptedJobs(); err != nil {
			log.Printf("recover interrupted batch automation jobs: %v", err)
		}
	}
	v11 := http.NewServeMux()
	v12 := http.NewServeMux()
	v11.HandleFunc("GET /api/batch-factory/v11/capabilities", capabilityHandlerForRuntime(options.Slice, options.Production != nil && options.Production.Enabled, options.Merge != nil && options.Merge.Enabled, options.External != nil && options.External.Enabled[external.Provider121], options.External != nil && options.External.Enabled[external.ProviderYadi]))
	// V12 is the public workbench namespace. Keep its capability contract on
	// Go even while a few legacy V12 adapters still bridge to Node.
	v12.HandleFunc("GET /api/batch-factory/v12/capabilities", capabilityHandlerForRuntime(options.Slice, options.Production != nil && options.Production.Enabled, options.Merge != nil && options.Merge.Enabled, options.External != nil && options.External.Enabled[external.Provider121], options.External != nil && options.External.Enabled[external.ProviderYadi]))
	if options.Store != nil && options.Slice >= 1 {
		registerSliceOneRoutes(v11, options.Store, options.PresetCatalog)
		registerV12IntakeRoutes(v12, options.Store)
		registerV12PublishClassificationRoute(v12, options.Store, batchFetchClassification{
			AIGenderJudge: options.NovelFetchAIGenderJudge, AIGenderJudgeResolver: options.NovelFetchAIGenderJudgeResolver,
			RewriteProvider: options.NovelFetchRewrite, RewriteResolver: options.NovelFetchRewriteResolver,
		})
		registerConfiguredImageGenerationRoute(v11, options.Store, options.NovelFetchStore)
		registerConfiguredImageGenerationRoute(v12, options.Store, options.NovelFetchStore)
		registerBookAssetRoutes(v11, options.Store, options.LocalArtifacts)
		registerV12AssetGenerationRoute(v12, options.Store, options.LocalArtifacts, options.NovelFetchStore)
	}
	if options.NovelFetchFetcher != nil {
		registerV12FetchOriginalsRoute(v12, options.NovelFetchFetcher, batchFetchClassification{
			AIGenderJudge: options.NovelFetchAIGenderJudge, AIGenderJudgeResolver: options.NovelFetchAIGenderJudgeResolver,
			RewriteProvider: options.NovelFetchRewrite, RewriteResolver: options.NovelFetchRewriteResolver,
		})
		registerV12SingleFetchRoute(v12, options.Store, options.NovelFetchFetcher, batchFetchClassification{
			AIGenderJudge: options.NovelFetchAIGenderJudge, AIGenderJudgeResolver: options.NovelFetchAIGenderJudgeResolver,
			RewriteProvider: options.NovelFetchRewrite, RewriteResolver: options.NovelFetchRewriteResolver,
		})
	}
	if options.Store != nil && options.Director != nil && options.Slice >= 2 {
		registerDirectorRoutes(v11, options.Director, options.Store)
		stageService := &batchfactoryv11.BookStageService{Store: options.Store, Director: options.Director, Production: options.Production}
		if options.LocalArtifacts != nil {
			stageService.ImageReady = shuihuoImageReadyChecker(options.Store, options.NovelFetchStore)
			stageService.VideoReady = shuihuoVideoReadyChecker(options.Store, options.NovelFetchStore, options.Production)
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
				_, err = generateShuihuoAssetImages(ctx, options.Store, options.LocalArtifacts, owner, batch, assetIDs, "", "9:16", options.NovelFetchStore)
				return err
			}
		}
		registerBookStageRoutes(v11, stageService)
	} else if options.Store != nil {
		registerBookStageSummaryRoute(v11, &batchfactoryv11.BookStageService{Store: options.Store}, "v11")
	}
	if options.Compiler != nil && options.Slice >= 3 {
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
		registerProductionRoutes(v11, options.Production, options.NovelFetchStore)
		if output, ok := options.MergeOutput.(interface {
			PutUploadedMP4(context.Context, string, string) (string, error)
		}); ok {
			registerVideoUploadRoutes(v11, options.Production, options.Merge, options.LocalArtifacts, output)
		}
	}
	// Completed local merge artifacts must remain readable in the ordinary
	// review runtime. This only exposes owner-scoped GET routes; merge writes
	// remain gated by the production slice below.
	if mergeStatus != nil && options.LocalArtifacts != nil {
		var mergeReader mergeworker.ObjectReader
		if candidate, ok := options.MergeOutput.(mergeworker.ObjectReader); ok {
			mergeReader = candidate
		}
		registerMergeRoutes(v11, mergeStatus, options.LocalArtifacts, options.MergeOutput, mergeReader, options.Merge != nil && options.Slice >= 5)
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
		})
	})
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
	if options.Store != nil {
		registerV12SmartUnifiedRefresh(v12, auth, options.Store, options.TextProvider)
	}
	if options.Store != nil {
		registerV12MetadataCompat(v12, auth, options.Store, options.NovelFetchFetcher)
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
	registerVideoProviderStatusRoute(v12, auth, options.UserConfigStore, "v12", options.NovelFetchStore)
	if options.Store != nil {
		runtimeStageService := &batchfactoryv11.BookStageService{Store: options.Store, Director: options.Director, Production: options.Production, ImageReady: shuihuoImageReadyChecker(options.Store, options.NovelFetchStore), VideoReady: shuihuoVideoReadyChecker(options.Store, options.NovelFetchStore, options.Production)}
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
	RegisterAuthRoutes(root, AuthOptions{Store: options.AuthStore, Profiles: options.MemberProfileStore, Now: options.Now, SessionResolver: options.SessionResolver, MFAFile: mfaFile, PasskeyFile: passkeyFile})
	if options.PresetCatalog != nil {
		registerPresetRoutes(root, auth, options.PresetCatalog)
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
	RegisterAdminReadCompat(root, auth, options.AuthStore, options.MemberProfileStore, options.PresetCatalog)
	RegisterAdminGovernanceCompat(root, auth, options.AuthStore, options.MemberProfileStore)
	RegisterApplicationsCompat(root, options.AuthStore, options.MemberProfileStore, options.Now)
	RegisterClientErrorCompat(root, auth, options.MemberProfileStore)
	if options.MemberProfileStore != nil {
		RegisterAccountRecoveryStatusCompat(root, auth, options.MemberProfileStore.SystemDir)
		RegisterAccountPasskeyCompat(root, auth, options.MemberProfileStore, options.AuthStore)
		RegisterAccountPurgeCompat(root, auth, options.AuthStore, options.MemberProfileStore, options.MemberAvatarDir)
	}
	RegisterNovelPanelDiagnosticsCompat(root, auth, options.UserDataDir)
	RegisterNovelPanelSettingsCompat(root, auth, options.UserConfigStore, options.TextProvider, options.NovelFetchStore)
	RegisterNovelPanelAssetCompat(root, auth, options.UserDataDir)
	RegisterNovelPanelCharacterCompat(root, auth)
	RegisterNovelPanelLeaseCompat(root, auth)
	RegisterNovelPanelCharacterTraceCompat(root, auth, options.UserDataDir)
	RegisterNovelPanelCharacterAICompat(root, auth, options.TextProvider)
	RegisterNovelPanelAnalysisCompat(root, auth, options.TextProvider, options.PresetCatalog)
	RegisterNovelPanelTextOpsCompat(root, auth, options.TextProvider)
	RegisterNovelPanelOutlineCompat(root, auth, options.TextProvider, options.PresetCatalog)
	RegisterNovelPanelAssetDescribeCompat(root, auth, options.TextProvider)
	RegisterNovelPanelImageGenerateCompat(root, auth, options.UserConfigStore, options.NovelFetchStore)
	RegisterNovelPanelClipboardCompat(root, auth)
	registerUserConfigRoutes(root, auth, options.UserConfigStore, options.AuthStore, options.NovelFetchStore, options.MemberProfileStore)
	RegisterStorageRoutes(root, auth, options.UserConfigStore, options.ScriptHistoryStore)
	RegisterPlatformProjectsRoute(root, auth, options.ScriptHistoryStore, options.Store)
	RegisterNovelFetchUploadSessionCompat(root, novelCompatAuth, options.UserDataDir, options.VideoManagementConnector)
	RegisterNovelFetchUploadBatchCompat(root, novelCompatAuth, options.UserDataDir, options.NovelFetchStore, options.VideoManagementConnector)
	RegisterWebSubmitTargetCheckCompat(root, novelCompatAuth, options.UserDataDir, options.VideoManagementConnector)
	publish121OrganizationsHandler(root, novelCompatAuth, options.UserDataDir, options.VideoManagementConnector)
	if options.UserConfigStore != nil {
		RegisterAgentWorkspaceCompat(root, auth, options.AgentWorkspace, options.UserConfigStore.BaseDir, options.TextProvider, options.Now)
		RegisterAgentAssetRoutes(root, auth, options.AgentWorkspace, options.AgentAssets, options.AgentAssetObjects, options.Now)
		textModels := options.AgentTextModelResolver
		if textModels == nil && options.NovelFetchStore != nil {
			textModels = func(ctx context.Context, owner, modelID string) (batchfactoryv11.DirectorProvider, error) {
				return resolveAgentTextModel(ctx, options.NovelFetchStore, owner, modelID)
			}
		}
		RegisterAgentNodeGenerationRoutes(root, auth, AgentNodeGenerationServices{Projects: options.AgentWorkspace, TextModels: textModels, Now: options.Now})
		if options.NovelFetchStore != nil && options.AgentWorkspace != nil && options.AgentAssets != nil {
			var references agentworkspace.ReferenceURLProvider
			if options.AgentAssetObjects != nil {
				references, _ = options.AgentAssetObjects.(agentworkspace.ReferenceURLProvider)
			}
			videoResolver := func(ctx context.Context, owner, modelKey string) (batchfactoryv11.VideoProviderConfig, error) {
				config, err := loadShuihuoOwnerConfig(ctx, options.Store, owner, options.NovelFetchStore)
				if err != nil {
					return batchfactoryv11.VideoProviderConfig{}, err
				}
				return scriptVideoModelConfig(ctx, options.Store, owner, config, modelKey)
			}
			RegisterAgentVideoJobRoutes(root, auth, &agentVideoCoordinator{
				projects: options.AgentWorkspace, assets: options.AgentAssets, references: references,
				videos:   &scriptVideoService{resolver: videoResolver, localExecutor: options.LocalExecutors},
				ingestor: newAgentVideoAssetIngestor(options.AgentAssets, options.AgentAssetObjects, nil, options.Now),
				executor: options.LocalExecutors, artifacts: options.LocalArtifacts, now: options.Now,
			})
		}
		systemDir := ""
		if options.MemberProfileStore != nil {
			systemDir = options.MemberProfileStore.SystemDir
		}
		RegisterAgentSkillsCompat(root, auth, systemDir, options.UserConfigStore.BaseDir)
	}
	if options.TextProvider != nil {
		registerScriptSmartUnifiedRoute(root, auth, options.TextProvider)
		registerChatRoute(root, auth, options.TextProvider, options.PresetCatalog, options.PersonalPromptStore)
		registerScriptDirectorPipelineCompat(root, auth, options.TextProvider, options.PresetCatalog, options.PersonalPromptStore)
	}
	if legacyV11Public() {
		root.Handle("/api/batch-factory/v11/", auth.Middleware(v11))
	}
	root.Handle("/api/batch-factory/v12/", auth.Middleware(v12))
	if options.NovelFetchStore != nil {
		RegisterShuihuoAccountAIConfigCompat(root, auth, options.NovelFetchStore)
	}
	registerShuihuoProjectCompatRoutes(root, auth, options.Store, options.LocalArtifacts, options.Director, options.TextProvider, options.Production, options.NovelFetchStore, options.PresetCatalog, options.RedisURL)
	registerTTSRoute(root, auth, options.NovelFetchStore)
	RegisterLocalExecutorRoutes(root, auth, options.LocalExecutors)
	RegisterLocalExecutorJobRoutes(root, auth, options.LocalExecutors)
	RegisterLocalExecutorArtifactRoutes(root, auth, options.LocalExecutors, options.LocalArtifacts)
	if options.NovelFetchStore != nil {
		RegisterScriptVideoRoutes(root, auth, options.LocalExecutors, func(ctx context.Context, owner, modelKey string) (batchfactoryv11.VideoProviderConfig, error) {
			config, err := loadShuihuoOwnerConfig(ctx, options.Store, owner, options.NovelFetchStore)
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
	RegisterGiantMaterialResolverRoutes(root, auth, options.GiantMaterialResolver)
	if options.NovelFetchStore != nil {
		novelFetch := http.NewServeMux()
		registerNovelFetchWorkshopRoutes(novelFetch, options.NovelFetchStore, options.NovelFetchQueue, options.NovelFetchAIGenderJudge, options.NovelFetchAIGenderJudgeResolver, options.NovelFetchFetcher, options.NovelFetchRewrite, options.NovelFetchRewriteResolver, options.UserDataDir, options.VideoManagementConnector)
		registerNovelFetchLegacyWorkflowRoutes(novelFetch, options.NovelFetchStore, options.NovelFetchQueue, options.UserDataDir, func(req *http.Request) (novelfetchworkshop.RewriteProvider, error) {
			if options.NovelFetchRewriteResolver == nil {
				return options.NovelFetchRewrite, nil
			}
			identity, _ := BridgeIdentityFromContext(req.Context())
			return options.NovelFetchRewriteResolver(req.Context(), identity.Username)
		}, options.VideoManagementConnector)
		registerNovelFetchCleanupRoutes(novelFetch, options.NovelFetchStore, options.Now)
		novelAuth := NovelFetchBridgeAuth{Secret: options.BridgeSecret, Now: options.Now, Users: options.Users, AuthStore: options.AuthStore, SessionResolver: options.SessionResolver, IdentityScope: novelScope}
		root.Handle("/api/novel-fetch-workshop/", novelAuth.Middleware(novelFetch))
		RegisterBatchRewriteConfigCompat(root, novelCompatAuth, options.NovelFetchStore, options.UserDataDir, options.VideoManagementConnector)
		RegisterWebSubmitPreviewCompat(root, novelCompatAuth, options.NovelFetchStore)
		RegisterWebSubmitSubmitCompat(root, novelCompatAuth, options.UserDataDir, options.NovelFetchStore, func(ctx context.Context, owner string) (novelfetchworkshop.AIGenderJudge, error) {
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
		root.Handle("POST /api/batch-rewrite/web-submit/operations", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("GET /api/batch-rewrite/web-submit/operations/", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("POST /api/web-submit/operations", novelAuth.Middleware(http.HandlerFunc(legacyWebSubmitAlias)))
		root.Handle("GET /api/web-submit/operations/", novelAuth.Middleware(http.HandlerFunc(legacyWebSubmitAlias)))
		root.Handle("PATCH /api/batch-rewrite/schedules/", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("DELETE /api/batch-rewrite/schedules/", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		registerCanonicalNovelFetchWebSubmitAliases(root, auth)
		registerModelCatalogRoutes(root, novelAuth, options.NovelFetchStore)
	}
	root.HandleFunc("GET /health", func(w http.ResponseWriter, req *http.Request) {
		if options.Health != nil && options.Health(req.Context()) != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	return root
}

// registerCanonicalNovelFetchWebSubmitAliases keeps the public workshop
// namespace on Go while reusing the tested compatibility handlers. The
// canonical request is authenticated with NovelFetchBridgeAuth; the internal
// hop is re-signed only because the compatibility handler uses the legacy
// bridge payload path. No Node service is involved.
func registerCanonicalNovelFetchWebSubmitAliases(root *http.ServeMux, bridge BridgeAuth) {
	if root == nil {
		return
	}
	root.Handle("/api/novel-fetch-workshop/web-submit/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		clone := req.Clone(req.Context())
		legacyPath := strings.Replace(clone.URL.Path, "/api/novel-fetch-workshop/web-submit/", "/api/batch-rewrite/web-submit/", 1)
		clone.URL.Path = legacyPath
		username := clone.Header.Get(HeaderUsername)
		issued := clone.Header.Get(HeaderIssuedAt)
		owner := clone.Header.Get(HeaderIsOwner)
		if username != "" && issued != "" && owner != "" && bridge.Secret != "" {
			clone.Header.Set(HeaderSignature, bridgeSignature(bridge.Secret, canonicalBridgePayload(username, issued, owner, clone.Method, legacyPath)))
		}
		root.ServeHTTP(w, clone)
	}))
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

func shuihuoImageReadyChecker(store batchfactoryv11.Store, catalogStore novelfetchworkshop.Store) func(context.Context, string) (bool, string) {
	return func(ctx context.Context, owner string) (bool, string) {
		config, err := loadShuihuoOwnerConfig(ctx, store, owner, catalogStore)
		if err != nil {
			return false, "图片模型配置读取失败"
		}
		for _, model := range shuihuoCollection(config, "modelCatalog", nil) {
			if shuihuoImageModelReady(model) {
				return true, ""
			}
		}
		return false, "图片模型或图片生成服务未配置"
	}
}

func shuihuoVideoReadyChecker(store batchfactoryv11.Store, catalogStore novelfetchworkshop.Store, production *batchfactoryv11.ProductionService) func(context.Context, string) (bool, string) {
	return func(ctx context.Context, owner string) (bool, string) {
		if production == nil || !production.Enabled {
			return false, "视频生成服务未启用"
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
