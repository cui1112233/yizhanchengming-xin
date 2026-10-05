package app

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/generation"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/httpapi"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/intake"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/pipeline"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/unifiedsettings"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/video"
)

// NewHandler wires intake, pipeline, generation, unified settings and video services to MySQL.
// React/localStorage never becomes the source of truth for persisted configuration.
func NewHandler(db *sql.DB, fetcher intake.Fetcher, classifier intake.Classifier, now pipeline.Clock) http.Handler {
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
	settingsService := unifiedsettings.NewService(settingsStore, unifiedsettings.StaticDefaults{Config: unifiedsettings.Settings{
		Production: map[string]any{"productionMode": "original", "aiCopyEnabled": false, "aiCopyCount": float64(1)},
		Publishing: map[string]any{"uploadVideoType": "merged", "materialReuse": false},
	}}, settingsStore)

	videoStore := video.NewMySQLStore(db)
	masterKey := []byte(os.Getenv("VIDEO_PROVIDER_MASTER_KEY"))
	providerFactory := video.DefaultProviderFactory{LocalJobs: videoStore}
	videoConfigService := video.NewConfigServiceWithProviders(videoStore, masterKey, providerFactory)
	localExecutorService := video.NewLocalExecutorService(videoStore, nil)

	var artifactStore video.ArtifactStore
	if endpoint, region, bucket, accessKey, secretKey, publicBase := os.Getenv("TOS_ENDPOINT"), os.Getenv("TOS_REGION"), os.Getenv("TOS_BUCKET"), os.Getenv("TOS_ACCESS_KEY"), os.Getenv("TOS_SECRET_KEY"), os.Getenv("TOS_PUBLIC_BASE_URL"); endpoint != "" && region != "" && bucket != "" && accessKey != "" && secretKey != "" && publicBase != "" {
		if uploader, err := video.NewTOSUploader(endpoint, region, accessKey, secretKey); err == nil {
			artifactStore, _ = video.NewArtifactStore(video.ArtifactStoreConfig{Bucket: bucket, PublicBaseURL: publicBase, Uploader: uploader})
		}
	}
	videoService := video.NewService(
		videoStore,
		video.NewGenerationFinalPromptSource(generationStore),
		providerFactory,
		artifactStore,
		masterKey,
	)

	return httpapi.NewHandler(httpapi.Dependencies{
		Intakes:             intakeService,
		Reader:              store,
		Pipeline:            pipelineService,
		BatchProjects:       store,
		BatchProjectDetails: store,
		Generation:          generationService,
		UnifiedSettings:     settingsService,
		Video:               videoService,
		VideoConfig:         videoConfigService,
		VideoLocalExecutor:  localExecutorService,
		VideoStatus:         videoService,
	})
}
