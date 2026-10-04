package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/agent"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/agentworker"
	one21 "github.com/cui1112233/yizhanchengming-xin/api/internal/integrations/121"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/integrations/metadataai"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/integrations/tosstore"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/media"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/providerconfig"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/queue"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/storage"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/videogen"
	workerpkg "github.com/cui1112233/yizhanchengming-xin/api/internal/worker"
)

type workerConfig struct {
	MySQLDSN             string
	RedisURL             string
	QueueKey             string
	AgentToolQueueKey    string
	FetchEndpoint        string
	AIBaseURL            string
	AIModel              string
	AIAPIKey             string
	TOSEndpoint          string
	TOSRegion            string
	TOSAccessKey         string
	TOSSecretKey         string
	TOSBucket            string
	ProviderCredentialKey string
}

func configFromEnv() (workerConfig, error) {
	cfg := workerConfig{
		MySQLDSN: strings.TrimSpace(os.Getenv("MYSQL_DSN")),
		RedisURL: strings.TrimSpace(os.Getenv("REDIS_URL")),
		QueueKey: strings.TrimSpace(os.Getenv("PIPELINE_QUEUE_KEY")),
		AgentToolQueueKey: strings.TrimSpace(os.Getenv("AGENT_TOOL_QUEUE_KEY")),
		FetchEndpoint: strings.TrimSpace(os.Getenv("121_FETCH_ENDPOINT")),
		AIBaseURL: strings.TrimSpace(os.Getenv("AI_BASE_URL")),
		AIModel: strings.TrimSpace(os.Getenv("AI_MODEL")),
		AIAPIKey: strings.TrimSpace(os.Getenv("AI_API_KEY")),
		TOSEndpoint: strings.TrimSpace(os.Getenv("TOS_ENDPOINT")),
		TOSRegion: strings.TrimSpace(os.Getenv("TOS_REGION")),
		TOSAccessKey: strings.TrimSpace(os.Getenv("TOS_ACCESS_KEY")),
		TOSSecretKey: strings.TrimSpace(os.Getenv("TOS_SECRET_KEY")),
		TOSBucket: strings.TrimSpace(os.Getenv("TOS_BUCKET")),
		ProviderCredentialKey: os.Getenv("PROVIDER_CREDENTIAL_KEY"),
	}
	if cfg.MySQLDSN == "" { return workerConfig{}, errors.New("MYSQL_DSN is required") }
	if cfg.RedisURL == "" { return workerConfig{}, errors.New("REDIS_URL is required") }
	if cfg.QueueKey == "" { cfg.QueueKey = "qiantie:pipeline:ready" }
	if cfg.AgentToolQueueKey == "" { cfg.AgentToolQueueKey = "qiantie:agent:tools" }
	if cfg.FetchEndpoint == "" { cfg.FetchEndpoint = "https://txt.121w.com/api.php" }

	aiConfigured := cfg.AIBaseURL != "" || cfg.AIModel != "" || cfg.AIAPIKey != ""
	aiComplete := cfg.AIBaseURL != "" && cfg.AIModel != "" && cfg.AIAPIKey != ""
	if aiConfigured && !aiComplete { return workerConfig{}, errors.New("AI_BASE_URL, AI_MODEL and AI_API_KEY must be configured together") }
	tosConfigured := cfg.TOSEndpoint != "" || cfg.TOSRegion != "" || cfg.TOSAccessKey != "" || cfg.TOSSecretKey != "" || cfg.TOSBucket != ""
	tosComplete := cfg.TOSEndpoint != "" && cfg.TOSRegion != "" && cfg.TOSAccessKey != "" && cfg.TOSSecretKey != "" && cfg.TOSBucket != ""
	if tosConfigured && !tosComplete { return workerConfig{}, errors.New("TOS_ENDPOINT, TOS_REGION, TOS_ACCESS_KEY, TOS_SECRET_KEY and TOS_BUCKET must be configured together") }
	if cfg.ProviderCredentialKey != "" && len([]byte(cfg.ProviderCredentialKey)) != 32 {
		return workerConfig{}, errors.New("PROVIDER_CREDENTIAL_KEY must be exactly 32 bytes")
	}
	return cfg, nil
}

func main() {
	cfg, err := configFromEnv()
	if err != nil { log.Fatal(err) }

	db, err := sql.Open("mysql", cfg.MySQLDSN)
	if err != nil { log.Fatal(err) }
	defer db.Close()
	db.SetMaxOpenConns(24)
	db.SetMaxIdleConns(24)
	db.SetConnMaxIdleTime(10 * time.Minute)
	db.SetConnMaxLifetime(time.Hour)
	pingCtx, cancelPing := context.WithTimeout(context.Background(), 15*time.Second)
	if err := db.PingContext(pingCtx); err != nil { cancelPing(); log.Fatal(err) }
	cancelPing()

	pipelineSource, err := queue.NewRedisSource(cfg.RedisURL, cfg.QueueKey)
	if err != nil { log.Fatal(err) }
	jobQueue, err := queue.NewRedisQueue(cfg.RedisURL, cfg.QueueKey)
	if err != nil { log.Fatal(err) }
	agentSource, err := queue.NewAgentToolSource(cfg.RedisURL, cfg.AgentToolQueueKey)
	if err != nil { log.Fatal(err) }

	var classifier workerpkg.MetadataClassifier
	if cfg.AIBaseURL != "" {
		classifier, err = metadataai.New(metadataai.Config{BaseURL:cfg.AIBaseURL,Model:cfg.AIModel,APIKey:cfg.AIAPIKey})
		if err != nil { log.Fatal(err) }
	}
	pipelineRunner := workerpkg.Runner{
		Source: pipelineSource,
		Jobs: storage.NewSQLJobStore(db),
		Executor: workerpkg.IntakeExecutor{Books:workerpkg.NewSQLIntakeRepository(db),Fetcher:one21.NewClient(cfg.FetchEndpoint),Classifier:classifier,Batches:workerpkg.NewSQLBatchCreator(db)},
		Progress: storage.NewSQLProgressStore(db),
		Queue: jobQueue,
	}

	registry := agent.NewToolRegistry()
	videoEnabled := false
	if cfg.TOSEndpoint != "" && cfg.ProviderCredentialKey != "" {
		tosClient, err := tosstore.New(tosstore.Config{Endpoint:cfg.TOSEndpoint,Region:cfg.TOSRegion,AccessKey:cfg.TOSAccessKey,SecretKey:cfg.TOSSecretKey,Bucket:cfg.TOSBucket})
		if err != nil { log.Fatal(err) }
		cipher, err := providerconfig.NewCipher([]byte(cfg.ProviderCredentialKey))
		if err != nil { log.Fatal(err) }
		assetStore := media.NewSQLStore(db)
		mediaResolver := media.Service{Assets:assetStore,Signer:tosClient}
		ingestor := media.IngestService{Uploader:tosClient,Assets:assetStore}
		providerStore := providerconfig.NewSQLStore(db,cipher)
		factory := videogen.Factory{Configs:videogen.ConfigStoreSource{Store:providerStore}}
		videoService := videogen.Service{Providers:factory,References:mediaResolver,Downloader:videogen.HTTPDownloader{},Ingestor:ingestor}
		if err := registry.Register(agent.ToolVideoGenerate, agent.VideoGenerateTool{Service:videoService}); err != nil { log.Fatal(err) }
		videoEnabled = true
	}
	agentRunner := agentworker.Runner{Source:agentSource,Store:agent.NewSQLStore(db),Executor:registry}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log.Printf("unified Go worker started; pipeline_queue=%s agent_queue=%s ai_fallback=%t agent_video=%t",cfg.QueueKey,cfg.AgentToolQueueKey,classifier!=nil,videoEnabled)
	go runPipelineLoop(ctx,pipelineRunner)
	runAgentLoop(ctx,agentRunner)
}

func runPipelineLoop(ctx context.Context, runner workerpkg.Runner) {
	for ctx.Err()==nil {
		err:=runner.RunOnce(ctx)
		if err==nil { continue }
		if errors.Is(err,context.Canceled) { return }
		if errors.Is(err,queue.ErrQueueEmpty) { continue }
		log.Printf("pipeline execution error: %v",err)
		time.Sleep(time.Second)
	}
}

func runAgentLoop(ctx context.Context, runner agentworker.Runner) {
	for ctx.Err()==nil {
		err:=runner.RunOnce(ctx)
		if err==nil { continue }
		if errors.Is(err,context.Canceled) { return }
		if errors.Is(err,queue.ErrQueueEmpty) { continue }
		log.Printf("agent tool execution error: %v",err)
		time.Sleep(time.Second)
	}
}
