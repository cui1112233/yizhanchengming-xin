package httpapi

import (
	"context"
	"net/http"
	"qiantie/backend/internal/agentworkspace"
	"qiantie/backend/internal/batchfactoryv11"
	"qiantie/backend/internal/mergeworker"
	"qiantie/backend/internal/novelfetchworkshop"
	"qiantie/backend/internal/screenplay"
	"qiantie/backend/internal/subjectupload"
)

func registerScreenplayWorkflow(root *http.ServeMux, auth BridgeAuth, options RouterOptions, resolver func(context.Context, string, string) (batchfactoryv11.DirectorProvider, error), catalog novelfetchworkshop.Store) {
	if options.Screenplay == nil {
		return
	}
	var authoringRepo screenplay.AuthoringRepository
	if factory, ok := options.Screenplay.(interface {
		AuthoringRepository() screenplay.AuthoringRepository
	}); ok {
		authoringRepo = factory.AuthoringRepository()
	} else {
		authoringRepo = screenplay.NewMemoryAuthoringRepository()
	}
	authoring := &screenplay.AuthoringService{Store: options.Screenplay, Repository: authoringRepo}
	productionRepo, _ := options.Screenplay.(screenplay.ProductionRepository)
	mediaStore, _ := options.Screenplay.(screenplay.MediaStore)
	references, _ := options.AgentAssetObjects.(agentworkspace.ReferenceURLProvider)
	media := &ScreenplayMediaWorkbench{Projects: options.Screenplay, Media: mediaStore, Cut: options.AICut, Objects: options.AgentAssetObjects, Assets: options.AgentAssets, ReferencesProvider: references, Catalog: catalog, ASR: options.AICutASR, Exports: options.AICutExport, Authoring: authoring, Production: productionRepo, Executors: options.LocalExecutors, Files: options.LocalArtifacts}
	media.Images = newAgentImageGenerator(options.Store, catalog, options.UserConfigStore, options.LocalExecutors, options.LocalArtifacts, options.AgentAssets, references)
	media.VoiceDesigner = newScreenplayVoiceDesigner(catalog)
	media.PositioningText = func(ctx context.Context, owner string, req screenplay.TextRequest) (screenplay.TextResponse, error) {
		return (screenplayTextGenerator{owner: owner, resolver: resolver}).Generate(ctx, req)
	}
	var runner screenplay.BatchMediaRunner
	if options.UserConfigStore != nil {
		videoResolver := func(ctx context.Context, owner, modelKey string) (batchfactoryv11.VideoProviderConfig, error) {
			config, err := options.UserConfigStore.GetConfig(ctx, owner)
			if err != nil {
				return batchfactoryv11.VideoProviderConfig{}, err
			}
			return scriptVideoModelConfig(ctx, options.Store, owner, config, modelKey)
		}
		runner = &screenplayVideoRunner{service: &scriptVideoService{resolver: videoResolver, localExecutor: options.LocalExecutors}, configs: options.UserConfigStore, repo: productionRepo, References: media.References, ReferenceIDs: media.ReferenceIDs, Subjects: media.VideoSubjects}
	}
	runtime := &screenplay.ProductionRuntime{Store: options.Screenplay, Repository: productionRepo, Executor: func(owner string) screenplay.BatchExecutor {
		return screenplay.BatchExecutor{Store: options.Screenplay, Generator: screenplayTextGenerator{owner: owner, resolver: resolver}, Media: runner, Finalizer: media}
	}}
	if redis, err := mergeworker.NewRedisStore(options.RedisURL, "qiantie:screenplay:production"); err == nil {
		runtime.RedisDo = redis.Do
	}
	RegisterScreenplayWorkflowRoutes(root, auth, options.Screenplay, resolver, options.Now, options.PresetCatalog, authoring, runner)
	RegisterScreenplayAuthoringRoutes(root, auth, options.Screenplay, authoringRepo, resolver)
	RegisterScreenplayStoryboardPreferenceRoutes(root, auth, options.Screenplay, authoring)
	RegisterScreenplayStoryboardPlanRoutes(root, auth, authoring, resolver)
	RegisterScreenplayStoryboardGenerationRoutes(root, auth, options.Screenplay, authoring)
	registerScreenplayFileRoutes(root, auth, authoring)
	RegisterScreenplayMediaWorkbenchRoutes(root, auth, media)
	RegisterScreenplayPositioningRoutes(root, auth, media)
	RegisterScreenplayVoiceBindingRoutes(root, auth, media)
	RegisterScreenplayProductionRoutes(root, auth, runtime, authoring)
	libraryRepo, _ := options.Screenplay.(screenplay.AssetLibraryRepository)
	if libraryRepo == nil {
		if _, memoryOnly := options.Screenplay.(*screenplay.MemoryStore); memoryOnly {
			libraryRepo = screenplay.NewMemoryAssetLibraryRepository()
		}
	}
	library := &ScreenplayAssetLibrary{Repository: libraryRepo, Media: media}
	uploadRepo, _ := options.Screenplay.(screenplay.SubjectUploadRepository)
	uploads := &ScreenplaySubjectUploadService{Repository: uploadRepo, Library: library, Configs: options.UserConfigStore, Runner: subjectupload.NewClient()}
	if redis, err := mergeworker.NewRedisStore(options.RedisURL, "qiantie:screenplay:subject-upload"); err == nil {
		uploads.RedisDo = subjectUploadRedisDo(redis.Do)
	}
	library.SubjectUploads = uploads
	RegisterScreenplayAssetLibraryRoutes(root, auth, library)
	assetCheckRepo, _ := options.Screenplay.(screenplay.AssetCheckRepository)
	assetChecks := &screenplay.AssetCheckRuntime{Store: options.Screenplay, Repository: assetCheckRepo, Authoring: authoring, BuildSteps: screenplay.BuildAssetCheckSteps, Generator: func(owner string) screenplay.TextGenerator {
		return screenplayTextGenerator{owner: owner, resolver: resolver}
	}}
	if redis, err := mergeworker.NewRedisStore(options.RedisURL, "qiantie:screenplay:asset-check"); err == nil {
		assetChecks.RedisDo = redis.Do
	}
	RegisterScreenplayAssetCheckRoutes(root, auth, assetChecks, authoring)
	if options.AgentImageWorkerContext != nil && assetCheckRepo != nil && assetChecks.RedisDo != nil {
		go assetChecks.Run(options.AgentImageWorkerContext)
	}

	if options.AgentImageWorkerContext != nil && productionRepo != nil && runtime.RedisDo != nil {
		go runtime.Run(options.AgentImageWorkerContext)
	}
}
