package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/auth"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/batchfactoryv11/external"
	"qiantie/backend/internal/config"
	"qiantie/backend/internal/giantmaterialexecutor"
	"qiantie/backend/internal/giantmaterialresolver"
	"qiantie/backend/internal/httpapi"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
	"qiantie/backend/internal/mergeworker"
	"qiantie/backend/internal/n8catalog"
	"qiantie/backend/internal/n8comic"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/storage"
	"qiantie/backend/internal/videomanagementconnector"
)

type novelFetchRewriteAdapter struct {
	provider *batchfactoryv11.OpenAICompatibleProvider
}

type ownerTextProvider struct {
	fallback batchfactoryv11.DirectorProvider
	resolve  func(context.Context, string) (batchfactoryv11.DirectorProvider, error)
}

func (p ownerTextProvider) Complete(ctx context.Context, request batchfactoryv11.TextCompletionRequest) (string, error) {
	if p.resolve != nil {
		if identity, ok := httpapi.BridgeIdentityFromContext(ctx); ok {
			provider, err := p.resolve(ctx, identity.Username)
			if err != nil {
				return "", err
			}
			// A catalog with no enabled text model can return a typed nil
			// *OpenAICompatibleProvider inside the interface. Treat that as
			// unavailable instead of calling Validate through a nil receiver.
			if provider != nil {
				if configured, ok := provider.(*batchfactoryv11.OpenAICompatibleProvider); !ok || configured != nil {
					return provider.Complete(ctx, request)
				}
			}
		}
	}
	if p.fallback != nil {
		if configured, ok := p.fallback.(*batchfactoryv11.OpenAICompatibleProvider); ok && configured == nil {
			return "", errors.New("text provider is not configured")
		}
		return p.fallback.Complete(ctx, request)
	}
	return "", errors.New("text provider is not configured")
}

func (a novelFetchRewriteAdapter) Complete(ctx context.Context, request novelfetchworkshop.RewriteRequest) (string, error) {
	return a.provider.Complete(ctx, batchfactoryv11.TextCompletionRequest{
		SystemPrompt:        request.SystemPrompt,
		UserPrompt:          request.UserPrompt,
		Temperature:         request.Temperature,
		MaxTokens:           request.MaxTokens,
		DisableJSONResponse: true,
	})
}

// configuredTextProvider reads the owner-scoped model catalog. An explicit
// selection wins; when it is empty, the first enabled text model is the
// account's default. This keeps the Go runtime aligned with the API Config
// page instead of requiring a duplicate environment-variable configuration.
func configuredTextProvider(settings map[string]any, selected string) (*batchfactoryv11.OpenAICompatibleProvider, error) {
	selected = strings.TrimSpace(selected)
	var catalog []any
	switch raw := settings["modelCatalog"].(type) {
	case []any:
		catalog = raw
	case []map[string]any:
		catalog = make([]any, 0, len(raw))
		for _, item := range raw {
			catalog = append(catalog, item)
		}
	}
	for _, raw := range catalog {
		model, ok := raw.(map[string]any)
		if !ok || model["kind"] != "text" || model["enabled"] != true {
			continue
		}
		id := strings.TrimSpace(fmt.Sprint(model["id"]))
		modelID := strings.TrimSpace(fmt.Sprint(model["modelId"]))
		if selected != "" && id != selected && modelID != selected {
			continue
		}
		endpoint := strings.TrimRight(strings.TrimSpace(fmt.Sprint(model["baseUrl"])), "/")
		if !strings.HasSuffix(endpoint, "/chat/completions") {
			endpoint += "/chat/completions"
		}
		provider := &batchfactoryv11.OpenAICompatibleProvider{
			Endpoint: endpoint,
			APIKey:   strings.TrimSpace(fmt.Sprint(model["credential"])),
			Model:    modelID,
		}
		if err := provider.Validate(); err != nil {
			return nil, err
		}
		return provider, nil
	}
	return nil, nil
}

func selectedTextModelID(settings map[string]any) string {
	selected := firstConfiguredString(settings, "textModelId", "text_model_id")
	if aiConfig, ok := settings["aiConfig"].(map[string]any); ok {
		if value := firstConfiguredString(aiConfig, "textModelId", "text_model_id"); value != "" {
			selected = value
		}
	}
	return strings.TrimSpace(selected)
}

func firstConfiguredString(settings map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(fmt.Sprint(settings[key])); value != "" && value != "<nil>" {
			return value
		}
	}
	return ""
}

func textGenderJudge(provider *batchfactoryv11.OpenAICompatibleProvider) novelfetchworkshop.AIGenderJudge {
	if provider == nil {
		return nil
	}
	return func(ctx context.Context, info novelfetchworkshop.BookInfo) (string, error) {
		return provider.Complete(ctx, batchfactoryv11.TextCompletionRequest{
			SystemPrompt:        "你是小说男女频分类助手。只输出 JSON：{\"gender\":\"男频或女频\"}。必须从「男频」「女频」中选择最接近的一项，不要留空，不要输出 JSON 以外的内容。",
			UserPrompt:          "请根据以下分类信息和正文片段判断小说男女频：分类=" + info.Category + "；题材=" + info.Genre + "；正文片段=" + genderTextExcerpt(info.Text),
			Temperature:         0,
			MaxTokens:           256,
			DisableJSONResponse: false,
		})
	}
}

func genderTextExcerpt(text string) string {
	text = strings.TrimSpace(text)
	runes := []rune(text)
	if len(runes) > 1200 {
		runes = runes[:1200]
	}
	return string(runes)
}

func selectGiantMaterialResolver(cfg config.Config) httpapi.GiantMaterialResolver {
	if strings.TrimSpace(cfg.GiantMaterialPluginURL) != "" {
		return giantmaterialresolver.NewHTTPPluginClient(nil, cfg.GiantMaterialPluginURL, cfg.GiantMaterialPluginToken)
	}
	return giantmaterialresolver.NewClient(nil, cfg.QingyuMaterialSelectURL, cfg.QingyuN8AdminToken)
}

func Build(ctx context.Context, cfg config.Config, db *sql.DB, register func(*http.ServeMux)) (http.Handler, error) {
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	if err := prompt.EnsureComicPresets(cfg.PresetPath); err != nil {
		return nil, err
	}
	migrations := append(storage.AppMigrations(), storage.NovelFetchWorkshopMigrations()...)
	migrations = append(migrations, storage.AuthMigrations()...)
	if err := storage.RunMigrations(ctx, db, migrations); err != nil {
		return nil, err
	}
	store := batchfactoryv11.NewReadbackMySQLStore(db)
	authStore := auth.NewMySQLStore(db)
	if err := auth.ImportLegacyAccountsOnce(ctx, authStore, cfg.LegacyAccountsPath); err != nil {
		return nil, err
	}
	novelFetchStore := novelfetchworkshop.NewMySQLStore(db)
	agentWorkspaceStore := agentworkspace.NewMySQLStore(db)
	agentAssetStore := agentworkspace.NewMySQLAssetStore(db)
	var agentAssetObjects agentworkspace.AssetObjectStore
	var legacyAgentAssetObjects agentworkspace.AssetObjectStore
	if cfg.AgentAssetLocalDir != "" {
		legacyAgentAssetObjects = agentworkspace.NewLocalAssetObjectStore(cfg.AgentAssetLocalDir)
	}
	novelFetchStore.LegacyConfigDir = cfg.ScriptHistoryDir
	if cfg.NovelFetchTOSEndpoint != "" || cfg.NovelFetchTOSRegion != "" || cfg.NovelFetchTOSBucket != "" || cfg.NovelFetchTOSAccessKey != "" || cfg.NovelFetchTOSSecretKey != "" {
		bodyStore, bodyStoreErr := novelfetchworkshop.NewBodyTOSStore(novelfetchworkshop.BodyTOSConfig{Endpoint: cfg.NovelFetchTOSEndpoint, Region: cfg.NovelFetchTOSRegion, Bucket: cfg.NovelFetchTOSBucket, AccessKey: cfg.NovelFetchTOSAccessKey, SecretKey: cfg.NovelFetchTOSSecretKey})
		if bodyStoreErr != nil {
			return nil, bodyStoreErr
		}
		novelFetchStore.ObjectBodies = bodyStore
	}
	if cfg.AgentAssetTOSEndpoint != "" || cfg.AgentAssetTOSRegion != "" || cfg.AgentAssetTOSBucket != "" || cfg.AgentAssetTOSAccessKey != "" || cfg.AgentAssetTOSSecretKey != "" {
		store, storeErr := agentworkspace.NewTOSAssetObjectStore(agentworkspace.AssetTOSConfig{Endpoint: cfg.AgentAssetTOSEndpoint, Region: cfg.AgentAssetTOSRegion, Bucket: cfg.AgentAssetTOSBucket, AccessKey: cfg.AgentAssetTOSAccessKey, SecretKey: cfg.AgentAssetTOSSecretKey})
		if storeErr != nil {
			return nil, storeErr
		}
		if legacyAgentAssetObjects != nil {
			agentAssetObjects = agentworkspace.NewMigratingAssetObjectStore(store, store, legacyAgentAssetObjects)
		} else {
			agentAssetObjects = store
		}
	} else {
		agentAssetObjects = legacyAgentAssetObjects
	}
	var director *batchfactoryv11.DirectorService
	var textProvider *batchfactoryv11.OpenAICompatibleProvider
	if cfg.DirectorEndpoint != "" || cfg.DirectorAPIKey != "" || cfg.DirectorModel != "" {
		textProvider = &batchfactoryv11.OpenAICompatibleProvider{Endpoint: cfg.DirectorEndpoint, APIKey: cfg.DirectorAPIKey, Model: cfg.DirectorModel}
		if err := textProvider.Validate(); err != nil {
			return nil, err
		}
	}
	if cfg.Slice >= 2 {
		director = &batchfactoryv11.DirectorService{Store: store, Provider: textProvider}
	}
	var novelFetchAIGenderJudge novelfetchworkshop.AIGenderJudge
	var novelFetchAIGenderJudgeResolver func(context.Context, string) (novelfetchworkshop.AIGenderJudge, error)
	var novelFetchRewrite novelfetchworkshop.RewriteProvider
	var novelFetchRewriteResolver func(context.Context, string) (novelfetchworkshop.RewriteProvider, error)
	if textProvider != nil {
		novelFetchAIGenderJudge = func(ctx context.Context, info novelfetchworkshop.BookInfo) (string, error) {
			return textProvider.Complete(ctx, batchfactoryv11.TextCompletionRequest{
				SystemPrompt:        "你是小说男女频分类助手。只输出 JSON：{\"gender\":\"男频或女频\",\"confidence\":0.0,\"reason\":\"判断依据\"}。confidence 范围为 0 到 1；依据不足时 gender 留空、confidence 设为 0 并说明原因。不要输出 JSON 以外的内容。",
				UserPrompt:          "请根据以下分类信息和正文片段判断小说男女频：分类=" + info.Category + "；题材=" + info.Genre + "；正文片段=" + genderTextExcerpt(info.Text),
				Temperature:         0,
				MaxTokens:           256,
				DisableJSONResponse: false,
			})
		}
		novelFetchRewrite = novelFetchRewriteAdapter{provider: textProvider}
	}
	resolveCatalogTextProvider := func(ctx context.Context, owner string) (*batchfactoryv11.OpenAICompatibleProvider, error) {
		settings, err := novelFetchStore.GetConfig(ctx, owner)
		if err != nil {
			return nil, err
		}
		provider, providerErr := configuredTextProvider(settings, selectedTextModelID(settings))
		if providerErr != nil {
			return nil, providerErr
		}
		// Keep the Go runtime aligned with the visible API Config page even
		// when MySQL contains an older redacted/empty catalog row. The legacy
		// file is local runtime configuration, not a second request path.
		if provider == nil && strings.TrimSpace(cfg.ScriptHistoryDir) != "" {
			raw, readErr := os.ReadFile(filepath.Join(cfg.ScriptHistoryDir, owner, "api-config.json"))
			if readErr == nil {
				var legacy map[string]any
				if json.Unmarshal(raw, &legacy) == nil {
					if catalog, ok := legacy["modelCatalog"]; ok {
						settings["modelCatalog"] = catalog
					}
					selected := selectedTextModelID(settings)
					if selected == "" {
						if value, ok := legacy["textModelId"].(string); ok {
							selected = strings.TrimSpace(value)
						}
					}
					provider, providerErr = configuredTextProvider(settings, selected)
					if providerErr != nil {
						return nil, providerErr
					}
				}
			}
		}
		return provider, nil
	}
	novelFetchRewriteResolver = func(ctx context.Context, owner string) (novelfetchworkshop.RewriteProvider, error) {
		provider, providerErr := resolveCatalogTextProvider(ctx, owner)
		if providerErr != nil {
			return nil, providerErr
		}
		if provider != nil {
			return novelFetchRewriteAdapter{provider: provider}, nil
		}
		return novelFetchRewrite, nil
	}
	novelFetchAIGenderJudgeResolver = func(ctx context.Context, owner string) (novelfetchworkshop.AIGenderJudge, error) {
		provider, providerErr := resolveCatalogTextProvider(ctx, owner)
		if providerErr != nil {
			return nil, providerErr
		}
		if provider != nil {
			return textGenderJudge(provider), nil
		}
		return novelFetchAIGenderJudge, nil
	}
	var novelFetchQueue novelfetchworkshop.ProcessQueue
	if strings.TrimSpace(cfg.RedisURL) != "" {
		redisTransport, redisErr := mergeworker.NewRedisStore(cfg.RedisURL, "qiantie:novel-fetch")
		if redisErr != nil {
			return nil, redisErr
		}
		novelFetchQueue = novelfetchworkshop.NewRedisQueueWithDo("qiantie:novel-fetch", novelfetchworkshop.RedisDoFunc(redisTransport.Do))
	}
	compiler := &batchfactoryv11.PromptCompilerService{Store: store}
	// Explicit video provider settings are persisted in the owner config document
	// so a restart keeps the last saved provider; catalog-driven sync stays in
	// memory because the model catalog is already durable.
	videoRegistry := batchfactoryv11.NewPersistentVideoProviderRegistry(novelFetchStore)
	videoRegistry.Warm(ctx)
	localExecutorService := localexecutor.NewService(localexecutor.NewMySQLStore(db), nil)
	giantMaterialExecutorService := giantmaterialexecutor.NewService(giantmaterialexecutor.NewMySQLStore(db), nil)
	if _, err := giantMaterialExecutorService.ConsolidateExecutors(ctx); err != nil {
		return nil, err
	}
	if err := giantMaterialExecutorService.EnsureIdentityUniqueIndex(ctx); err != nil {
		return nil, err
	}
	giantMaterialResolver := selectGiantMaterialResolver(cfg)
	artifactStore := localartifact.NewStore(cfg.LocalExecutorArtifactDir, cfg.LocalExecutorArtifactMaxBytes)
	automationPresetStore, err := httpapi.NewFileAutomationPresetStore(cfg.AutomationPresetPath, time.Now)
	if err != nil {
		return nil, err
	}
	// A deleted Batch Factory project is removed from its UI immediately, while
	// only its ledgered local merge MP4s are released after their retention
	// deadline. Cleanup also recovers files queued by an earlier Slice 5 run, so
	// it must not depend on whether video production or merging is enabled now.
	// This worker deliberately has no TOS authority.
	if shouldStartLocalArtifactPurger(cfg) {
		batchfactoryv11.LocalArtifactPurger{Repository: store, Artifacts: artifactStore}.Start(time.Hour)
	}
	// Keep a read-only production service available for status queries even
	// before the staged video provider is enabled. Write/submit routes remain
	// gated by the configured slice below.
	production := &batchfactoryv11.ProductionService{Store: store, Compiler: compiler, ProviderRegistry: videoRegistry, Enabled: cfg.ProductionEnabled}
	if cfg.Slice >= 4 {
		var fallbackAdapter batchfactoryv11.ProductionAdapter
		if strings.TrimSpace(cfg.VideoEndpoint) != "" && strings.TrimSpace(cfg.VideoPollEndpoint) != "" && strings.TrimSpace(cfg.VideoAPIKey) != "" && strings.TrimSpace(cfg.VideoModel) != "" {
			adapter := &batchfactoryv11.HTTPVideoAdapter{Endpoint: cfg.VideoEndpoint, PollEndpoint: cfg.VideoPollEndpoint, APIKey: cfg.VideoAPIKey, Model: cfg.VideoModel}
			if err := adapter.Validate(); err != nil {
				return nil, err
			}
			fallbackAdapter = adapter
		}
		localAdapter := batchfactoryv11.NewLocalExecutorVideoAdapter(&localVideoJobClient{service: localExecutorService}, cfg.LocalExecutorPublicBaseURL)
		localAdapter.ArtifactSecret = cfg.BridgeSecret
		modelID := strings.TrimSpace(cfg.VideoModel)
		if modelID == "" {
			modelID = batchfactoryv11.DefaultPersonalVideoModel
		}
		var fallbackPoller batchfactoryv11.ProductionPoller
		if poller, ok := fallbackAdapter.(batchfactoryv11.ProductionPoller); ok {
			fallbackPoller = poller
		}
		production = &batchfactoryv11.ProductionService{
			Store: store, Compiler: compiler, Adapter: fallbackAdapter, Poller: fallbackPoller,
			ProviderRegistry: videoRegistry, LocalExecutor: localAdapter,
			Enabled: cfg.ProductionEnabled, Model: batchfactoryv11.FrozenVideoModel{ID: modelID, MaxDuration: 15},
		}
	}
	var merge *batchfactoryv11.MergeService
	var mergeOutput *mergeworker.TOSObjectStore
	if cfg.Slice >= 5 {
		if cfg.LocalMergeEnabled {
			adapter := batchfactoryv11.NewLocalMergeAdapter(artifactStore)
			if strings.TrimSpace(cfg.MergeTOSEndpoint) != "" || strings.TrimSpace(cfg.MergeTOSBucket) != "" {
				var err error
				mergeOutput, err = mergeworker.NewTOSObjectStore(mergeworker.TOSConfig{Endpoint: cfg.MergeTOSEndpoint, Region: cfg.MergeTOSRegion, Bucket: cfg.MergeTOSBucket, AccessKey: cfg.MergeTOSAccessKey, SecretKey: cfg.MergeTOSSecretKey, PublicBaseURL: cfg.MergeTOSPublicBaseURL})
				if err != nil {
					return nil, err
				}
				adapter.Output = mergeOutput
			}
			merge = &batchfactoryv11.MergeService{Store: store, Adapter: adapter, Poller: adapter, DurationProbe: batchfactoryv11.HTTPVideoDurationProbe{}, Enabled: true}
		} else {
			adapter := &batchfactoryv11.HTTPMergeAdapter{Endpoint: cfg.MergeEndpoint, PollEndpoint: cfg.MergePollEndpoint, APIKey: cfg.MergeAPIKey}
			if err := adapter.Validate(); err != nil {
				return nil, err
			}
			merge = &batchfactoryv11.MergeService{Store: store, Adapter: adapter, Poller: adapter, DurationProbe: batchfactoryv11.HTTPVideoDurationProbe{}, Enabled: cfg.MergeEnabled}
		}
	}
	var externalPublish *external.Service
	if cfg.Slice >= 6 {
		externalStore := external.NewMySQLStore(db)
		provider121 := external.New121Provider(cfg.External121Endpoint, cfg.External121APIKey)
		providerYadi := external.NewYadiProvider(cfg.ExternalYadiEndpoint, cfg.ExternalYadiAPIKey)
		if cfg.External121Enabled {
			if err := provider121.Validate(); err != nil {
				return nil, err
			}
		}
		if cfg.ExternalYadiEnabled {
			if err := providerYadi.Validate(); err != nil {
				return nil, err
			}
		}
		externalPublish = &external.Service{
			Credentials: externalStore, Intents: externalStore, Audits: externalStore,
			BatchReader:      store,
			ProductionReader: production,
			MergeReader:      merge,
			Providers: map[external.Provider]external.SubmissionProvider{
				external.Provider121:  provider121,
				external.ProviderYadi: providerYadi,
			},
			Enabled: map[external.Provider]bool{external.Provider121: cfg.External121Enabled, external.ProviderYadi: cfg.ExternalYadiEnabled},
			Key:     cfg.ExternalCredentialsKey,
		}
	}
	var fallbackTextProvider batchfactoryv11.DirectorProvider
	if textProvider != nil {
		fallbackTextProvider = textProvider
	}
	effectiveTextProvider := ownerTextProvider{fallback: fallbackTextProvider, resolve: func(ctx context.Context, owner string) (batchfactoryv11.DirectorProvider, error) {
		return resolveCatalogTextProvider(ctx, owner)
	}}
	// Slice 2+ director routes must use the same account-scoped model catalog
	// as the rest of the Go workbench. Without this binding, a slice-3 H3
	// request boots successfully but reports an unavailable text provider.
	if director != nil {
		director.Provider = effectiveTextProvider
	}
	targetBaseURL := videoManagementTargetBaseURL()
	n8ComicClient := n8comic.NewClient(n8comic.Config{BaseURL: cfg.N8ComicBaseURL, APIKey: cfg.N8ComicOpenAPIKey})
	n8ProducerDirectory := n8catalog.NewClient(nil, cfg.N8AdminBaseURL, cfg.QingyuN8AdminToken)
	n8ComicWorks := n8catalog.NewWorksClientWithPath(nil, cfg.N8ComicBaseURL, cfg.QingyuN8AdminToken, os.Getenv("QIANTIE_N8_WORKS_PATH"))
	return httpapi.NewRouter(httpapi.RouterOptions{AgentVideoWorkerContext: context.Background(), AgentScriptSplitWorkerContext: context.Background(), AgentImageWorkerContext: context.Background(), AgentAudioWorkerContext: context.Background(), BridgeSecret: cfg.BridgeSecret, ReleaseSHA: cfg.ReleaseSHA, RuntimeManifestPath: os.Getenv("QIANTIE_RUNTIME_MANIFEST_PATH"), Health: databaseHealth(db), Users: storage.BridgeUsers{DB: db}, AuthStore: authStore, SessionResolver: nil, Slice: cfg.Slice, Store: store, Director: director, Compiler: compiler, Production: production, Merge: merge, MergeOutput: mergeOutput, External: externalPublish, NovelFetchStore: novelFetchStore, NovelFetchQueue: novelFetchQueue, NovelFetchAIGenderJudge: novelFetchAIGenderJudge, NovelFetchAIGenderJudgeResolver: novelFetchAIGenderJudgeResolver, NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Endpoint: cfg.NovelFetch121Endpoint, Client: &http.Client{Timeout: cfg.NovelFetch121Timeout}}, NovelFetchRewrite: novelFetchRewrite, NovelFetchRewriteResolver: novelFetchRewriteResolver, LocalExecutors: localExecutorService, GiantMaterialExecutor: giantMaterialExecutorService, GiantMaterialResolver: giantMaterialResolver, LocalArtifacts: artifactStore, LocalExecutorDownloadsDir: cfg.LocalExecutorDownloadsDir, PresetCatalog: prompt.NewCatalog(cfg.PresetPath), AutomationPresetStore: automationPresetStore, AutomationStatusReader: httpapi.NewFileAutomationStatusReader(cfg.AutomationStatePath), PersonalPromptStore: prompt.NewPersonalStore(cfg.PersonalPromptPath), ScriptHistoryStore: httpapi.NewScriptHistoryStore(cfg.ScriptHistoryDir), MemberProfileStore: httpapi.NewMemberProfileStore(cfg.MemberSystemDir), MemberAvatarDir: cfg.MemberAvatarDir, UserDataDir: cfg.ScriptHistoryDir, UserConfigStore: httpapi.NewUserConfigStore(cfg.ScriptHistoryDir), TextProvider: effectiveTextProvider, AgentWorkspace: agentWorkspaceStore, AgentAssets: agentAssetStore, AgentAssetObjects: agentAssetObjects, RedisURL: cfg.RedisURL, RegisterV11: register, VideoManagementConnector: &videomanagementconnector.TargetHTTPService{BaseURL: targetBaseURL}, N8ComicPublisher: &httpapi.BatchFactoryN8ComicPublisher{Store: store, LocalFiles: artifactStore, Client: n8ComicClient, ProducerDirectory: n8ProducerDirectory, WorksDirectory: n8ComicWorks, ProducerAccounts: cfg.N8ComicProducerAccounts, WorkTitleMapPath: os.Getenv("QIANTIE_N8_WORK_TITLE_MAP_PATH")}}), nil
}

func shouldStartLocalArtifactPurger(cfg config.Config) bool {
	return strings.TrimSpace(cfg.LocalExecutorArtifactDir) != ""
}

func videoManagementTargetBaseURL() string {
	raw := strings.TrimRight(strings.TrimSpace(os.Getenv("QIANTIE_VIDEO_CONNECTOR_TARGET_BASE_URL")), "/")
	if raw == "" || strings.EqualFold(raw, "http://two.121w.com") {
		return "https://two.121w.com"
	}
	return raw
}
