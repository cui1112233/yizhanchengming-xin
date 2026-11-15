package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	"qiantie/backend/internal/config"
	"qiantie/backend/internal/giantmaterialexecutor"
	"qiantie/backend/internal/giantmaterialresolver"
	"qiantie/backend/internal/httpapi"
	"qiantie/backend/internal/imagecapacity"
	"qiantie/backend/internal/imagegenerationcache"
	"qiantie/backend/internal/localartifact"
	"qiantie/backend/internal/localexecutor"
	"qiantie/backend/internal/mergeworker"
	"qiantie/backend/internal/n8catalog"
	"qiantie/backend/internal/n8comic"
	"qiantie/backend/internal/novelcreation"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/productionmedia"
	"qiantie/backend/internal/projectretention"
	"qiantie/backend/internal/prompt"
	"qiantie/backend/internal/safehttp"
	"qiantie/backend/internal/safepath"
	"qiantie/backend/internal/screenplay"
	"qiantie/backend/internal/storage"
	"qiantie/backend/internal/upscale"
	"qiantie/backend/internal/videomanagementconnector"
)

type novelFetchRewriteAdapter struct {
	provider *batchfactoryv11.OpenAICompatibleProvider
}

func (a novelFetchRewriteAdapter) ModelIdentity() string {
	if a.provider == nil {
		return ""
	}
	return strings.TrimSpace(a.provider.Model)
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

// resolveNovelFetchTextProvider 解析批量测书使用的文本模型。
// selection 为空表示使用账号默认模型：先尝试保存的默认模型，失效时
// 回退到目录里第一个启用的文本模型；显式选择不做回退，避免静默换模型。
func resolveNovelFetchTextProvider(settings map[string]any, selection string) (*batchfactoryv11.OpenAICompatibleProvider, error) {
	selection = strings.TrimSpace(selection)
	if selection != "" {
		return configuredTextProvider(settings, selection)
	}
	provider, err := configuredTextProvider(settings, selectedTextModelID(settings))
	if err != nil {
		return nil, err
	}
	if provider != nil {
		return provider, nil
	}
	// 保存的默认模型可能已被管理员下架或被历史配置写死；账号仍应能
	// 使用目录里第一个可用文本模型，否则所有未指定模型的任务都会失败。
	return configuredTextProvider(settings, "")
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
	if err := prompt.EnsureComicPresets(cfg.PresetPath, cfg.LocalExecutorArtifactDir); err != nil {
		return nil, err
	}
	migrations := append(storage.AppMigrations(), storage.NovelFetchWorkshopMigrations()...)
	migrations = append(migrations, storage.AgentProxyMigrations()...)
	migrations = append(migrations, storage.AuthMigrations()...)
	migrations = append(migrations, storage.ScreenplayMigrations()...)
	migrations = append(migrations, storage.NovelCreationMigrations()...)
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
	agentProxyStore := agentproxy.NewMySQLStore(db)
	aiCutStore := aicut.NewMySQLStore(db)
	aiCutTimelines := aicut.NewMySQLTimelineCollectionStore(db, aiCutStore)
	screenplayStore := screenplay.NewMySQLStore(db)
	novelCreationStore := novelcreation.NewMySQLStore(db)
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
	var userConfigStore *httpapi.UserConfigStore
	resolveCatalogTextProviderWithSelection := func(ctx context.Context, owner, selection string) (*batchfactoryv11.OpenAICompatibleProvider, error) {
		settings, err := novelFetchStore.GetConfig(ctx, owner)
		if err != nil {
			return nil, err
		}
		provider, providerErr := resolveNovelFetchTextProvider(settings, selection)
		if providerErr != nil {
			return nil, providerErr
		}
		// Keep the Go runtime aligned with the visible API Config page even
		// when MySQL contains an older redacted/empty catalog row. The legacy
		// file is local runtime configuration, not a second request path.
		// Explicit per-task selections must never be swapped for the legacy
		// account model: a bad selection has to fail visibly instead.
		if provider == nil && strings.TrimSpace(selection) == "" && strings.TrimSpace(cfg.ScriptHistoryDir) != "" {
			legacyPath, pathErr := safepath.UserPath(cfg.ScriptHistoryDir, owner, "api-config.json")
			if pathErr != nil {
				return nil, pathErr
			}
			raw, readErr := os.ReadFile(legacyPath)
			if readErr == nil {
				var legacy map[string]any
				if json.Unmarshal(raw, &legacy) == nil {
					if catalog, ok := legacy["modelCatalog"]; ok {
						settings["modelCatalog"] = catalog
					}
					provider, providerErr = resolveNovelFetchTextProvider(settings, "")
					if providerErr != nil {
						return nil, providerErr
					}
				}
			}
		}
		return provider, nil
	}
	resolveCatalogTextProvider := func(ctx context.Context, owner string) (*batchfactoryv11.OpenAICompatibleProvider, error) {
		return resolveCatalogTextProviderWithSelection(ctx, owner, "")
	}
	// textProviderForModel honours an explicit textModelId saved on a batch or
	// book. The workbench selector used to be dropped by the stage service and
	// silently fell back to the account default.
	textProviderForModel := func(ctx context.Context, owner, modelID string) (batchfactoryv11.DirectorProvider, error) {
		modelID = strings.TrimSpace(modelID)
		if modelID == "" {
			return nil, nil
		}
		if provider, matched, err := httpapi.ResolveIntegrationTextProvider(ctx, userConfigStore, owner, modelID); matched {
			return provider, err
		}
		provider, err := resolveCatalogTextProviderWithSelection(ctx, owner, modelID)
		if err != nil {
			return nil, err
		}
		if provider == nil {
			return nil, batchfactoryv11.NewConfigRequiredError("选择的文本模型不可用：" + modelID)
		}
		return provider, nil
	}
	var novelCreationRuntime novelcreation.Runtime = novelcreation.NewMemoryRuntime()
	if strings.TrimSpace(cfg.RedisURL) != "" {
		redisTransport, redisErr := mergeworker.NewRedisStore(cfg.RedisURL, "qiantie:novel-creation")
		if redisErr != nil {
			return nil, redisErr
		}
		novelCreationRuntime = novelcreation.NewRedisRuntime("qiantie:novel-creation", novelcreation.RedisDoFunc(redisTransport.Do))
	}
	novelCreationService := novelcreation.NewService(novelCreationStore, novelCreationRuntime, novelcreation.NewPipeline(textProviderForModel))
	novelCreationRunStarter := func(_ context.Context, owner, runID string) {
		go func() {
			runCtx, cancel := context.WithTimeout(context.Background(), 45*time.Minute)
			defer cancel()
			if _, runErr := novelCreationService.Run(runCtx, owner, runID); runErr != nil {
				fmt.Printf("novel creation run %s failed: %v\n", runID, runErr)
			}
		}()
	}
	// The scheduler is a process-lifetime worker: Build's ctx is a startup
	// timeout, not a server lifecycle context.
	novelcreation.StartScheduler(context.Background(), novelCreationService, novelCreationRunStarter, time.Minute)
	novelcreation.StartBatchWorker(context.Background(), novelCreationService, novelCreationRunStarter, 5*time.Second)
	novelFetchRewriteResolver = func(ctx context.Context, owner string) (novelfetchworkshop.RewriteProvider, error) {
		provider, providerErr := resolveCatalogTextProviderWithSelection(ctx, owner, novelfetchworkshop.ProcessTextModel(ctx))
		if providerErr != nil {
			return nil, providerErr
		}
		if provider != nil {
			return novelFetchRewriteAdapter{provider: provider}, nil
		}
		return novelFetchRewrite, nil
	}
	novelFetchAIGenderJudgeResolver = func(ctx context.Context, owner string) (novelfetchworkshop.AIGenderJudge, error) {
		provider, providerErr := resolveCatalogTextProviderWithSelection(ctx, owner, novelfetchworkshop.ProcessTextModel(ctx))
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
	imageGenerationCache := imagegenerationcache.NewMySQLRepository(db)
	imageCapacityPolicies := imagecapacity.NewMySQLRepository(db)
	var imageCapacityRedis func(context.Context, ...string) (any, error)
	if strings.TrimSpace(cfg.RedisURL) != "" {
		transport, err := mergeworker.NewRedisStore(cfg.RedisURL, "qiantie:imagecapacity")
		if err != nil {
			return nil, err
		}
		imageCapacityRedis = transport.Do
	}
	imageCapacityController := imagecapacity.NewController(imageCapacityPolicies, imageCapacityRedis, "qiantie:imagecapacity")
	safehttp.SetManagedOutboundDecorator(imageCapacityController.WrapTransportWithTimeout)
	imagecapacity.SetEnabled(true)
	userConfigStore = httpapi.NewUserConfigStore(cfg.ScriptHistoryDir)
	var aiCutASRService *aicut.ASRService
	if strings.TrimSpace(cfg.RedisURL) != "" && agentAssetObjects != nil {
		redisTransport, redisErr := mergeworker.NewRedisStore(cfg.RedisURL, "qiantie:aicut:asr")
		if redisErr != nil {
			return nil, redisErr
		}
		queue := aicut.NewRedisASRQueue("qiantie:aicut:asr", aicut.RedisDoFunc(redisTransport.Do))
		aiCutASRService = aicut.NewASRService(aicut.ASRServiceOptions{
			Timelines: aiCutTimelines,
			Store:     aiCutStore,
			Notifier:  queue,
			OpenObject: func(ctx context.Context, asset aicut.Asset) (io.ReadCloser, error) {
				return agentAssetObjects.Open(ctx, agentworkspace.Asset{ID: asset.ID, ProjectID: asset.ProjectID, Kind: asset.Kind, Name: asset.Name, ContentType: asset.ContentType, ObjectKey: asset.ObjectKey, ByteSize: asset.ByteSize, CreatedAt: asset.CreatedAt})
			},
			ResolveProvider: func(ctx context.Context, owner, modelID string) (aicut.ASRProvider, error) {
				settings, err := novelFetchStore.GetConfig(ctx, owner)
				if err != nil {
					return nil, err
				}
				return aicut.ASRProviderFromCatalog(settings, modelID, nil)
			},
		})
		go aiCutASRService.Run(context.Background(), queue)
	}
	var aiCutExportService *httpapi.AICutExportService
	if strings.TrimSpace(cfg.RedisURL) != "" && agentAssetObjects != nil {
		redisTransport, redisErr := mergeworker.NewRedisStore(cfg.RedisURL, "qiantie:aicut:export")
		if redisErr != nil {
			return nil, redisErr
		}
		queue := aicut.NewRedisExportQueue("qiantie:aicut:export", aicut.RedisDoFunc(redisTransport.Do))
		aiCutExportService = httpapi.NewAICutExportService(aiCutStore, agentAssetObjects, queue, time.Now, aiCutTimelines)
		go aiCutExportService.Run(context.Background(), queue)
	}
	agentTools := aicut.NewAgentToolExecutor(aiCutStore, aiCutStore, time.Now, aiCutTimelines)
	agentOralTextProvider := func(ctx context.Context, owner, modelID string) (batchfactoryv11.DirectorProvider, error) {
		if strings.TrimSpace(modelID) == "" {
			return resolveCatalogTextProvider(ctx, owner)
		}
		return textProviderForModel(ctx, owner, modelID)
	}
	agentTools.ConfigureCompositeTools(aicut.AgentCompositeTools{
		AnalyzeSilence: func(ctx context.Context, owner, projectID, assetID string, options aicut.SilenceOptions) (aicut.SilenceAnalysis, error) {
			return httpapi.RunAICutAnalyzeSilence(ctx, aiCutStore, agentAssetObjects, owner, projectID, assetID, options)
		},
		ListASRModels: func(ctx context.Context, owner string) ([]aicut.AgentASRModel, error) {
			settings, err := novelFetchStore.GetConfig(ctx, owner)
			if err != nil {
				return nil, err
			}
			return aicut.AgentASRModelsFromCatalog(settings), nil
		},
		SmartSplitForTimeline: func(ctx context.Context, owner, projectID, timelineID, assetID string, threshold float64, minIntervalMS, expectedRevision int64) (aicut.Timeline, int, string, error) {
			if agentAssetObjects == nil {
				return aicut.Timeline{}, 0, "", errors.New("素材存储尚未配置")
			}
			scoped, err := aiCutTimelines.ScopedStore(ctx, owner, projectID, timelineID)
			if err != nil {
				return aicut.Timeline{}, 0, "", err
			}
			return httpapi.RunAICutSmartSplit(ctx, scoped, agentAssetObjects, time.Now, owner, projectID, assetID, threshold, minIntervalMS, expectedRevision)
		},
		OralCleanForTimeline: func(ctx context.Context, owner, projectID, timelineID, modelID string, maxPauseMS, expectedRevision int64) (aicut.Timeline, aicut.FillerTrimSummary, string, error) {
			scoped, err := aiCutTimelines.ScopedStore(ctx, owner, projectID, timelineID)
			if err != nil {
				return aicut.Timeline{}, aicut.FillerTrimSummary{}, "", err
			}
			return httpapi.RunAICutOralClean(ctx, scoped, httpapi.AICutTextProviderResolver(agentOralTextProvider), time.Now, owner, projectID, modelID, maxPauseMS, expectedRevision)
		},
		SubmitASR: func(ctx context.Context, owner string, input aicut.ASRSubmit) (aicut.ASRJob, error) {
			if aiCutASRService == nil {
				return aicut.ASRJob{}, errors.New("ASR 服务尚未配置，请先配置 Redis、素材存储和 ASR 模型")
			}
			return aiCutASRService.Submit(ctx, owner, input)
		},
		SmartSplit: func(ctx context.Context, owner, projectID, assetID string, threshold float64, minIntervalMS, expectedRevision int64) (aicut.Timeline, int, string, error) {
			if agentAssetObjects == nil {
				return aicut.Timeline{}, 0, "", errors.New("素材存储尚未配置")
			}
			return httpapi.RunAICutSmartSplit(ctx, aiCutStore, agentAssetObjects, time.Now, owner, projectID, assetID, threshold, minIntervalMS, expectedRevision)
		},
		ExtractAudio: func(ctx context.Context, owner, projectID, assetID string) (aicut.Asset, error) {
			if agentAssetObjects == nil {
				return aicut.Asset{}, errors.New("素材存储尚未配置")
			}
			return httpapi.RunAICutExtractAudio(ctx, aiCutStore, agentAssetObjects, time.Now, owner, projectID, assetID)
		},
		OralClean: func(ctx context.Context, owner, projectID, modelID string, maxPauseMS, expectedRevision int64) (aicut.Timeline, aicut.FillerTrimSummary, string, error) {
			return httpapi.RunAICutOralClean(ctx, aiCutStore, httpapi.AICutTextProviderResolver(agentOralTextProvider), time.Now, owner, projectID, modelID, maxPauseMS, expectedRevision)
		},
	})
	agentService := aicut.NewAgentService(aiCutStore, agentTools, func(ctx context.Context, owner, modelID string) (aicut.AgentModel, error) {
		var provider batchfactoryv11.DirectorProvider
		var err error
		if strings.TrimSpace(modelID) == "" {
			provider, err = resolveCatalogTextProvider(ctx, owner)
		} else {
			provider, err = textProviderForModel(ctx, owner, modelID)
		}
		if err != nil || provider == nil {
			return nil, err
		}
		return aicut.AgentModelFunc(func(callCtx context.Context, request aicut.AgentModelRequest) (string, error) {
			return provider.Complete(callCtx, batchfactoryv11.TextCompletionRequest{SystemPrompt: request.SystemPrompt, UserPrompt: request.UserPrompt, Temperature: 0.1, MaxTokens: 4096, DisableJSONResponse: true})
		}), nil
	}, func(mode aicut.AgentMode) string {
		return prompt.BuildAICutAgentSystemPrompt(string(mode))
	}, time.Now)
	var aiCutAgentStarter httpapi.AICutAgentStarter
	agentService.ConfigureSkills(prompt.ResolveAICutAgentSkill)
	if strings.TrimSpace(cfg.RedisURL) != "" {
		redisTransport, redisErr := mergeworker.NewRedisStore(cfg.RedisURL, "qiantie:aicut:agent")
		if redisErr != nil {
			return nil, redisErr
		}
		queue := aicut.NewRedisAgentQueue("qiantie:aicut:agent", aicut.RedisDoFunc(redisTransport.Do))
		aiCutAgentStarter = queue.Notify
		go agentService.RunQueue(context.Background(), queue)
	} else {
		aiCutAgentStarter = func(_ context.Context, runID string) error {
			go func() { _, _ = agentService.Run(context.Background(), runID) }()
			return nil
		}
	}
	// 磁盘写保护与是否配置 TOS 无关：写满时必须在落盘前拒绝，否则错误会
	// 退化成普通 IO 失败，页面只会显示“执行中”，既看不到磁盘原因也无法重试。
	usage := func() (float64, error) { return productionmedia.DiskUsagePercent(artifactStore.Root()) }
	artifactStore.SetWriteGuard(productionmedia.NewDiskWriteGuard(usage))
	// 回源超时：本地缓存缺失时从 TOS 恢复必须有上限，避免慢下载长期占用请求
	// goroutine；大文件场景可用环境变量调高（内部上限 10 分钟）。
	if raw := strings.TrimSpace(os.Getenv("QIANTIE_ARTIFACT_RESTORE_TIMEOUT")); raw != "" {
		if timeout, err := time.ParseDuration(raw); err == nil && timeout > 0 {
			artifactStore.SetRestoreTimeout(timeout)
		} else {
			log.Printf("ignore invalid QIANTIE_ARTIFACT_RESTORE_TIMEOUT=%q", raw)
		}
	}
	var retentionRemote batchfactoryv11.RemoteObjectDeleter
	retentionDays := func(owner string) int {
		settings, err := userConfigStore.GetConfig(context.Background(), owner)
		if err != nil {
			return 7
		}
		switch value := settings["productionRetentionDays"].(type) {
		case int:
			return productionmedia.NormalizeRetentionDays(value)
		case float64:
			return productionmedia.NormalizeRetentionDays(int(value))
		default:
			return 7
		}
	}
	productionMediaEnabled := strings.TrimSpace(cfg.ProductionMediaTOSEndpoint) != "" || strings.TrimSpace(cfg.ProductionMediaTOSBucket) != ""
	if productionMediaEnabled {
		if err := os.MkdirAll(artifactStore.Root(), 0o750); err != nil {
			return nil, fmt.Errorf("prepare production media cache: %w", err)
		}
		remote, mediaErr := productionmedia.NewTOSStore(productionmedia.TOSConfig{
			Endpoint: cfg.ProductionMediaTOSEndpoint, Region: cfg.ProductionMediaTOSRegion,
			Bucket: cfg.ProductionMediaTOSBucket, AccessKey: cfg.ProductionMediaTOSAccessKey, SecretKey: cfg.ProductionMediaTOSSecretKey,
		})
		if mediaErr != nil {
			return nil, mediaErr
		}
		retentionRemote = remote
		repository := productionmedia.NewMySQLRepository(db)
		resolver := &productionmedia.Resolver{Repository: repository, Local: artifactStore, Remote: remote}
		artifactStore.SetMissingRestorer(resolver)

		scanner := productionmedia.Scanner{DB: db, Repository: repository, Local: artifactStore, Prefix: cfg.ProductionMediaTOSPrefix, RetentionDays: retentionDays}
		migrator := productionmedia.Migrator{Repository: repository, Local: artifactStore, Remote: remote, MaxPerRun: 200}
		cleanup := productionmedia.CleanupSupervisor{Purger: productionmedia.Purger{Repository: repository, Local: artifactStore, MaxPerRun: 200}, Usage: usage}
		scanner.Start(time.Minute)
		migrator.Start(time.Minute)
		// Retention still controls normal deletion. The short supervisor cadence
		// only makes the pressure path react before a busy production host fills;
		// it can delete exclusively TOS-verified local copies.
		cleanup.Start(time.Minute)

	}
	// Expiry must also run on local-only installations without TOS. 全站统一：
	// 每天 00:00 重算保留期，到期只归档、不自动删除。
	projectRetention := projectretention.NewService(db, store.MySQLStore, artifactStore, retentionRemote, retentionDays)
	projectRetention.CanvasAssets = agentAssetStore
	projectRetention.CanvasObjects = agentAssetObjects
	projectRetention.Grace = projectretention.DefaultArchiveGrace
	projectRetention.StartDaily(0, 0)
	// 启动后补跑一次：立即归档已到期作品，并清理归档超过 24 小时的作品，
	// 避免升级/重启后仍能看到早已到期的旧作品，或让归档文件长期占用空间。
	go func() {
		time.Sleep(5 * time.Second)
		if err := projectRetention.RunOnce(context.Background()); err != nil {
			log.Printf("project retention startup sweep failed: %v", err)
		}
	}()
	// 批量测书删除历史文档时正文会转入可回收状态；这里每天兜底清理一次，
	// 避免 TOS 对象和正文行长期无人回收。
	startNovelFetchBodyCleanup(ctx, novelFetchStore)
	// 自动化预设走 MySQL 版本化存储（HEAD + 不可变修订），旧的 JSON 文件在
	// 启动时一次性导入，之后只读 MySQL，避免单机文件与多实例/重部署漂移。
	automationPresetStore, err := httpapi.NewMySQLAutomationPresetStore(db, cfg.AutomationPresetPath, time.Now)
	if err != nil {
		return nil, err
	}
	// A deleted Batch Factory project is removed from its UI immediately, while
	// only its ledgered local merge MP4s are released after their retention
	// deadline. Cleanup also recovers files queued by an earlier Slice 5 run, so
	// it must not depend on whether video production or merging is enabled now.
	// This worker deliberately has no TOS authority.
	if shouldStartLocalArtifactPurger(cfg) && !productionMediaEnabled {
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
	upscaleService := upscale.NewService(upscale.NewClient(cfg.VolcanoMediaKitBaseURL, cfg.VolcanoMediaKitAPIKey), time.Now)
	return httpapi.NewRouter(httpapi.RouterOptions{AgentVideoWorkerContext: context.Background(), AgentScriptSplitWorkerContext: context.Background(), AgentImageWorkerContext: context.Background(), AgentAudioWorkerContext: context.Background(), BridgeSecret: cfg.BridgeSecret, ReleaseSHA: cfg.ReleaseSHA, RuntimeManifestPath: os.Getenv("QIANTIE_RUNTIME_MANIFEST_PATH"), Health: databaseHealth(db), AuthRateLimits: httpapi.AuthRateLimitConfig{LoginWindow: cfg.AuthRateLimitLoginWindow, LoginMaxFailures: cfg.AuthRateLimitLoginMaxFailures, LoginLockout: cfg.AuthRateLimitLoginLockout, RegisterWindow: cfg.AuthRateLimitRegisterWindow, RegisterMax: cfg.AuthRateLimitRegisterMax, RegisterLockout: cfg.AuthRateLimitRegisterLockout, PasswordResetWindow: cfg.AuthRateLimitPasswordResetWindow, PasswordResetMax: cfg.AuthRateLimitPasswordResetMax, PasswordResetLockout: cfg.AuthRateLimitPasswordResetLockout}, Users: storage.BridgeUsers{DB: db}, AuthStore: authStore, SessionResolver: nil, Slice: cfg.Slice, Store: store, Director: director, Compiler: compiler, Production: production, Merge: merge, MergeOutput: mergeOutput, External: externalPublish, NovelFetchStore: novelFetchStore, NovelFetchQueue: novelFetchQueue, NovelFetchAIGenderJudge: novelFetchAIGenderJudge, NovelFetchAIGenderJudgeResolver: novelFetchAIGenderJudgeResolver, NovelFetchFetcher: &novelfetchworkshop.UpstreamFetcher{Endpoint: cfg.NovelFetch121Endpoint, Client: safehttp.NewClientWithTimeout(cfg.NovelFetch121Timeout)}, NovelFetchRewrite: novelFetchRewrite, NovelFetchRewriteResolver: novelFetchRewriteResolver, LocalExecutors: localExecutorService, GiantMaterialExecutor: giantMaterialExecutorService, GiantMaterialResolver: giantMaterialResolver, LocalArtifacts: artifactStore, ImageGenerationCache: imageGenerationCache, LocalExecutorDownloadsDir: cfg.LocalExecutorDownloadsDir, PresetCatalog: prompt.NewCatalog(cfg.PresetPath), AutomationPresetStore: automationPresetStore, AutomationStatusReader: httpapi.NewFileAutomationStatusReader(cfg.AutomationStatePath), AutomationStateStore: httpapi.NewMySQLAutomationStateStore(db), PersonalPromptStore: prompt.NewPersonalStore(cfg.PersonalPromptPath), ScriptHistoryStore: httpapi.NewScriptHistoryStore(cfg.ScriptHistoryDir), MemberProfileStore: httpapi.NewMemberProfileStore(cfg.MemberSystemDir), MemberAvatarDir: cfg.MemberAvatarDir, UserDataDir: cfg.ScriptHistoryDir, UserConfigStore: userConfigStore, TextProvider: effectiveTextProvider, TextProviderResolver: textProviderForModel, AgentWorkspace: agentWorkspaceStore, AgentProxy: agentProxyStore, AICut: aiCutStore, AICutPersonalTemplates: aicut.NewMySQLPersonalTemplateStore(db), AICutTimelines: aiCutTimelines, Screenplay: screenplayStore, NovelCreation: novelCreationService, ProjectRetention: projectRetention, NovelCreationRunStarter: novelCreationRunStarter, AICutASR: aiCutASRService, AICutExport: aiCutExportService, AICutAgentStore: aiCutStore, AICutAgentStarter: aiCutAgentStarter, Upscale: upscaleService, AgentAssets: agentAssetStore, AgentAssetObjects: agentAssetObjects, RedisURL: cfg.RedisURL, ImageCapacity: imageCapacityController, RegisterV11: register, VideoManagementConnector: &videomanagementconnector.TargetHTTPService{BaseURL: targetBaseURL}, N8ComicPublisher: &httpapi.BatchFactoryN8ComicPublisher{Store: store, LocalFiles: artifactStore, Client: n8ComicClient, ProducerDirectory: n8ProducerDirectory, WorksDirectory: n8ComicWorks, ProducerAccounts: cfg.N8ComicProducerAccounts, WorkTitleMapPath: os.Getenv("QIANTIE_N8_WORK_TITLE_MAP_PATH")}, ComicDirectRecordPath: comicDirectRecordPath(cfg)}), nil
}

func shouldStartLocalArtifactPurger(cfg config.Config) bool {
	return strings.TrimSpace(cfg.LocalExecutorArtifactDir) != ""
}

// comicDirectRecordPath 把合集直传的幂等记录放到和其他运行数据同一目录，
// 避免进程工作目录变化导致记录漂移。
func comicDirectRecordPath(cfg config.Config) string {
	if value := strings.TrimSpace(os.Getenv("QIANTIE_COMIC_DIRECT_RECORD_PATH")); value != "" {
		return value
	}
	if dir := strings.TrimSpace(cfg.ScriptHistoryDir); dir != "" {
		return filepath.Join(filepath.Dir(dir), "comic-direct-collections.json")
	}
	return ""
}

func videoManagementTargetBaseURL() string {
	raw := strings.TrimRight(strings.TrimSpace(os.Getenv("QIANTIE_VIDEO_CONNECTOR_TARGET_BASE_URL")), "/")
	if raw == "" || strings.EqualFold(raw, "http://two.121w.com") {
		return "https://two.121w.com"
	}
	return raw
}

// startNovelFetchBodyCleanup 定时兜底回收批量测书的过期正文。
// 删除历史文档只把正文标记为 releasable，真正的 TOS 对象删除依赖这里。
func startNovelFetchBodyCleanup(ctx context.Context, store *novelfetchworkshop.MySQLStore) {
	if store == nil {
		return
	}
	const batchLimit = 500
	const maxBatchesPerSweep = 20
	go func() {
		// 启动后尽快回收一次，避免服务重启把已过期正文又拖一天。
		timer := time.NewTimer(time.Minute)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				// 单轮只取 500 条会让积压正文按天排队；这里连续消费到不足
				// 一批为止，同时限制单次轮询总量，避免长期占用数据库连接。
				for batch := 0; batch < maxBatchesPerSweep && ctx.Err() == nil; batch++ {
					result, err := store.CleanupExpiredBodies(ctx, time.Now().UTC(), batchLimit)
					if err != nil {
						log.Printf("novel fetch body cleanup failed: %v", err)
						break
					}
					if result.Deleted > 0 {
						log.Printf("novel fetch body cleanup deleted %d expired bodies", result.Deleted)
					}
					if result.Deleted < batchLimit {
						break
					}
				}
				timer.Reset(6 * time.Hour)
			}
		}
	}()
}
