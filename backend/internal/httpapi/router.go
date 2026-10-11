package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/batchfactoryv11/external"
	"qiantie/backend/internal/giantmaterialexecutor"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
	"qiantie/backend/internal/mergeworker"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/prompt"
)

type RouterOptions struct {
	BridgeSecret                    string
	ReleaseSHA                      string
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
	NovelFetchAIGenderJudge         novelfetchworkshop.AIGenderJudge
	NovelFetchAIGenderJudgeResolver func(context.Context, string) (novelfetchworkshop.AIGenderJudge, error)
	NovelFetchFetcher               *novelfetchworkshop.UpstreamFetcher
	NovelFetchRewrite               novelfetchworkshop.RewriteProvider
	NovelFetchRewriteResolver       func(context.Context, string) (novelfetchworkshop.RewriteProvider, error)
	PresetCatalog                   *prompt.Catalog
	AutomationPresetStore           AutomationPresetStore
	AutomationStatusReader          *FileAutomationStatusReader
	PersonalPromptStore             *prompt.PersonalStore
	ScriptHistoryStore              *ScriptHistoryStore
	MemberProfileStore              *MemberProfileStore
	MemberAvatarDir                 string
	UserDataDir                     string
	UserConfigStore                 *UserConfigStore
	TextProvider                    batchfactoryv11.DirectorProvider
	LocalExecutors                  *localexecutor.Service
	GiantMaterialExecutor           *giantmaterialexecutor.Service
	GiantMaterialResolver           GiantMaterialResolver
	LocalArtifacts                  *localartifact.Store
	MergeOutput                     mergeworker.ObjectStore
}

func NewRouter(options RouterOptions) http.Handler {
	if options.Production == nil && options.Store != nil {
		// Read-only status is safe before the production slice is enabled. Keep
		// the service disabled for all write/submit operations.
		options.Production = &batchfactoryv11.ProductionService{Store: options.Store, Compiler: options.Compiler}
	}
	v11 := http.NewServeMux()
	v12 := http.NewServeMux()
	v11.HandleFunc("GET /api/batch-factory/v11/capabilities", capabilityHandlerForRuntime(options.Slice, options.Production != nil && options.Production.Enabled, options.Merge != nil && options.Merge.Enabled, options.External != nil && options.External.Enabled[external.Provider121], options.External != nil && options.External.Enabled[external.ProviderYadi]))
	// V12 is the public workbench namespace. Keep its capability contract on
	// Go even while a few legacy V12 adapters still bridge to Node.
	v12.HandleFunc("GET /api/batch-factory/v12/capabilities", capabilityHandlerForRuntime(options.Slice, options.Production != nil && options.Production.Enabled, options.Merge != nil && options.Merge.Enabled, options.External != nil && options.External.Enabled[external.Provider121], options.External != nil && options.External.Enabled[external.ProviderYadi]))
	if options.Store != nil && options.Slice >= 1 {
		registerSliceOneRoutes(v11, options.Store)
		registerV12IntakeRoutes(v12, options.Store)
		registerBookAssetRoutes(v11, options.Store, options.LocalArtifacts)
	}
	if options.NovelFetchFetcher != nil {
		registerV12FetchOriginalsRoute(v12, options.NovelFetchFetcher)
		registerV12SingleFetchRoute(v12, options.Store, options.NovelFetchFetcher)
	}
	if options.Store != nil && options.Director != nil && options.Slice >= 2 {
		registerDirectorRoutes(v11, options.Director, options.Store)
		registerBookStageRoutes(v11, &batchfactoryv11.BookStageService{Store: options.Store, Director: options.Director, Production: options.Production})
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
		registerProductionRoutes(v11, options.Production)
		if output, ok := options.MergeOutput.(interface {
			PutUploadedMP4(context.Context, string, string) (string, error)
		}); ok {
			registerVideoUploadRoutes(v11, options.Production, options.Merge, options.LocalArtifacts, output)
		}
	}
	if options.Merge != nil && options.Slice >= 5 {
		var mergeReader mergeworker.ObjectReader
		if candidate, ok := options.MergeOutput.(mergeworker.ObjectReader); ok {
			mergeReader = candidate
		}
		registerMergeRoutes(v11, options.Merge, options.LocalArtifacts, options.MergeOutput, mergeReader)
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

	root := http.NewServeMux()
	root.HandleFunc("GET /api/runtime-build-info", func(w http.ResponseWriter, req *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"service": "go-api", "git_sha": options.ReleaseSHA})
	})
	auth := BridgeAuth{Secret: options.BridgeSecret, Now: options.Now, Users: options.Users, AuthStore: options.AuthStore, SessionResolver: options.SessionResolver}
	if options.AutomationPresetStore != nil {
		registerAutomationPresetRoutes(v12, auth, options.AutomationPresetStore, "v12")
		registerBatchAutomationPresetRoute(v12, auth, options.Store, options.AutomationPresetStore, "v12")
	}
	if options.AutomationStatusReader != nil {
		registerAutomationStatusRoute(v12, auth, options.AutomationStatusReader, "v12")
	}
	registerVideoProviderStatusRoute(v12, auth, options.UserConfigStore, "v12")
	if options.Store != nil {
		registerRuntimeSummaryRoute(v12, auth, options.Store, options.Production, mergeStatus, options.AutomationStatusReader, "v12")
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
	RegisterAdminReadCompat(root, auth, options.AuthStore, options.MemberProfileStore)
	if options.MemberProfileStore != nil {
		RegisterAccountRecoveryStatusCompat(root, auth, options.MemberProfileStore.SystemDir)
		RegisterAccountPasskeyCompat(root, auth, options.MemberProfileStore, options.AuthStore)
	}
	RegisterNovelPanelDiagnosticsCompat(root, auth, options.UserDataDir)
	RegisterNovelPanelSettingsCompat(root, auth, options.UserConfigStore, options.TextProvider)
	RegisterNovelPanelAssetCompat(root, auth, options.UserDataDir)
	RegisterNovelPanelCharacterCompat(root, auth)
	RegisterNovelPanelLeaseCompat(root, auth)
	RegisterNovelPanelCharacterTraceCompat(root, auth, options.UserDataDir)
	RegisterNovelPanelCharacterAICompat(root, auth, options.TextProvider)
	RegisterNovelPanelAnalysisCompat(root, auth, options.TextProvider, options.PresetCatalog)
	RegisterNovelPanelTextOpsCompat(root, auth, options.TextProvider)
	RegisterNovelPanelOutlineCompat(root, auth, options.TextProvider, options.PresetCatalog)
	RegisterNovelPanelAssetDescribeCompat(root, auth, options.TextProvider)
	RegisterNovelPanelImageGenerateCompat(root, auth, options.UserConfigStore)
	RegisterNovelPanelClipboardCompat(root, auth)
	registerUserConfigRoutes(root, auth, options.UserConfigStore, options.AuthStore, options.MemberProfileStore)
	RegisterStorageRoutes(root, auth, options.UserConfigStore, options.ScriptHistoryStore)
	RegisterPlatformProjectsRoute(root, auth, options.ScriptHistoryStore, options.Store)
	RegisterNovelFetchUploadSessionCompat(root, auth, options.UserDataDir)
	RegisterNovelFetchUploadBatchCompat(root, auth, options.UserDataDir, options.NovelFetchStore)
	RegisterWebSubmitTargetCheckCompat(root, auth, options.UserDataDir)
	if options.UserConfigStore != nil {
		RegisterAgentReadCompat(root, auth, options.UserConfigStore.BaseDir)
		RegisterAgentChatCompat(root, auth, options.UserConfigStore.BaseDir, options.TextProvider)
		systemDir := ""
		if options.MemberProfileStore != nil {
			systemDir = options.MemberProfileStore.SystemDir
		}
		RegisterAgentSkillsCompat(root, auth, systemDir, options.UserConfigStore.BaseDir)
	}
	if options.TextProvider != nil {
		registerScriptSmartUnifiedRoute(root, auth, options.TextProvider)
		registerChatRoute(root, auth, options.TextProvider, options.PresetCatalog, options.PersonalPromptStore)
	}
	root.Handle("/api/batch-factory/v11/", auth.Middleware(v11))
	root.Handle("/api/batch-factory/v12/", auth.Middleware(v12))
	registerShuihuoProjectCompatRoutes(root, auth, options.Store, options.LocalArtifacts, options.Director, options.Production)
	registerTTSRoute(root, auth)
	RegisterLocalExecutorRoutes(root, auth, options.LocalExecutors)
	RegisterLocalExecutorJobRoutes(root, auth, options.LocalExecutors)
	RegisterLocalExecutorArtifactRoutes(root, auth, options.LocalExecutors, options.LocalArtifacts)
	RegisterLocalScriptVideoRoutes(root, auth, options.LocalExecutors)
	RegisterGiantMaterialExecutorRoutes(root, auth, options.GiantMaterialExecutor)
	RegisterGiantMaterialResolverRoutes(root, auth, options.GiantMaterialResolver)
	if options.NovelFetchStore != nil {
		novelFetch := http.NewServeMux()
		registerNovelFetchWorkshopRoutes(novelFetch, options.NovelFetchStore, options.NovelFetchAIGenderJudge, options.NovelFetchAIGenderJudgeResolver, options.NovelFetchFetcher, options.NovelFetchRewrite, options.NovelFetchRewriteResolver)
		registerNovelFetchCleanupRoutes(novelFetch, options.NovelFetchStore, options.Now)
		novelAuth := NovelFetchBridgeAuth{Secret: options.BridgeSecret, Now: options.Now, Users: options.Users, AuthStore: options.AuthStore, SessionResolver: options.SessionResolver}
		root.Handle("/api/novel-fetch-workshop/", novelAuth.Middleware(novelFetch))
		RegisterLegacyNovelFetchRoutes(root, auth, options.NovelFetchStore, options.NovelFetchFetcher, options.UserConfigStore)
		RegisterBatchRewriteConfigCompat(root, auth, options.NovelFetchStore, options.UserDataDir)
		RegisterWebSubmitPreviewCompat(root, auth, options.NovelFetchStore)
		RegisterWebSubmitSubmitCompat(root, auth, options.UserDataDir, options.NovelFetchStore)
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
		root.Handle("/api/batch-rewrite/knowledge/", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("/api/batch-rewrite/knowledge", novelAuth.Middleware(http.HandlerFunc(legacyKnowledgeSummaryAlias)))
		root.Handle("/api/batch-rewrite/opening/", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("/api/batch-rewrite/ai/", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("/api/batch-rewrite/process", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("/api/batch-rewrite/process/", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("/api/batch-rewrite/rules/ai-suggest", novelAuth.Middleware(http.HandlerFunc(legacyRulesSuggestAlias)))
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

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
