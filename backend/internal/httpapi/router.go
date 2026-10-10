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
	BridgeSecret              string
	ReleaseSHA                string
	Health                    func(context.Context) error
	Now                       func() time.Time
	Users                     BridgeUserResolver
	AuthStore                 auth.Store
	SessionResolver           func(context.Context, string) (BridgeIdentity, error)
	Slice                     int
	RegisterV11               func(*http.ServeMux)
	Store                     batchfactoryv11.Store
	Director                  *batchfactoryv11.DirectorService
	H3AudioProbe              batchfactoryv11.H3AudioDurationProbe
	Compiler                  *batchfactoryv11.PromptCompilerService
	Production                *batchfactoryv11.ProductionService
	Merge                     *batchfactoryv11.MergeService
	External                  *external.Service
	NovelFetchStore           novelfetchworkshop.LifecycleStore
	NovelFetchAIGenderJudge   novelfetchworkshop.AIGenderJudge
	NovelFetchFetcher         *novelfetchworkshop.UpstreamFetcher
	NovelFetchRewrite         novelfetchworkshop.RewriteProvider
	NovelFetchRewriteResolver func(context.Context, string) (novelfetchworkshop.RewriteProvider, error)
	PresetCatalog             *prompt.Catalog
	PersonalPromptStore       *prompt.PersonalStore
	ScriptHistoryStore        *ScriptHistoryStore
	MemberProfileStore        *MemberProfileStore
	MemberAvatarDir           string
	UserDataDir               string
	UserConfigStore           *UserConfigStore
	TextProvider              batchfactoryv11.DirectorProvider
	LocalExecutors            *localexecutor.Service
	GiantMaterialExecutor     *giantmaterialexecutor.Service
	GiantMaterialResolver     GiantMaterialResolver
	LocalArtifacts            *localartifact.Store
	MergeOutput               mergeworker.ObjectStore
}

func NewRouter(options RouterOptions) http.Handler {
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
	}
	if options.Store != nil && options.Director != nil && options.Slice >= 2 {
		registerDirectorRoutes(v11, options.Director, options.Store)
		registerBookStageRoutes(v11, &batchfactoryv11.BookStageService{Store: options.Store, Director: options.Director, Production: options.Production})
	}
	if options.Compiler != nil && options.Slice >= 3 {
		registerCompilerRoutes(v11, options.Compiler)
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
	if options.Store != nil && options.Slice >= 3 {
		registerV12H3Routes(v12, &batchfactoryv11.H3KernelService{Store: options.Store, AudioProbe: options.H3AudioProbe}, options.Director)
	}

	root := http.NewServeMux()
	root.HandleFunc("GET /api/runtime-build-info", func(w http.ResponseWriter, req *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"service": "go-api", "git_sha": options.ReleaseSHA})
	})
	auth := BridgeAuth{Secret: options.BridgeSecret, Now: options.Now, Users: options.Users, AuthStore: options.AuthStore, SessionResolver: options.SessionResolver}
	mfaFile := ""
	if options.MemberProfileStore != nil {
		mfaFile = filepath.Join(options.MemberProfileStore.SystemDir, "mfa.json")
	}
	passkeyFile := ""
	if options.MemberProfileStore != nil {
		passkeyFile = filepath.Join(options.MemberProfileStore.SystemDir, "passkeys.json")
	}
	RegisterAuthRoutes(root, AuthOptions{Store: options.AuthStore, Now: options.Now, SessionResolver: options.SessionResolver, MFAFile: mfaFile, PasskeyFile: passkeyFile})
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
	registerUserConfigRoutes(root, auth, options.UserConfigStore, options.AuthStore)
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
		registerNovelFetchWorkshopRoutes(novelFetch, options.NovelFetchStore, options.NovelFetchAIGenderJudge, options.NovelFetchFetcher, options.NovelFetchRewrite, options.NovelFetchRewriteResolver)
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
		root.Handle("/api/batch-rewrite/knowledge/", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("/api/batch-rewrite/opening/", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
		root.Handle("/api/batch-rewrite/ai/", novelAuth.Middleware(http.HandlerFunc(legacyWorkshopAlias)))
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
