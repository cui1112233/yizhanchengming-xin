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

	one21 "github.com/cui1112233/yizhanchengming-xin/api/internal/integrations/121"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/integrations/metadataai"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/queue"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/storage"
	workerpkg "github.com/cui1112233/yizhanchengming-xin/api/internal/worker"
)

type workerConfig struct {
	MySQLDSN      string
	RedisURL      string
	QueueKey      string
	FetchEndpoint string
	AIBaseURL     string
	AIModel       string
	AIAPIKey      string
}

func configFromEnv() (workerConfig, error) {
	cfg := workerConfig{
		MySQLDSN:      strings.TrimSpace(os.Getenv("MYSQL_DSN")),
		RedisURL:      strings.TrimSpace(os.Getenv("REDIS_URL")),
		QueueKey:      strings.TrimSpace(os.Getenv("PIPELINE_QUEUE_KEY")),
		FetchEndpoint: strings.TrimSpace(os.Getenv("121_FETCH_ENDPOINT")),
		AIBaseURL:     strings.TrimSpace(os.Getenv("AI_BASE_URL")),
		AIModel:       strings.TrimSpace(os.Getenv("AI_MODEL")),
		AIAPIKey:      strings.TrimSpace(os.Getenv("AI_API_KEY")),
	}
	if cfg.MySQLDSN == "" {
		return workerConfig{}, errors.New("MYSQL_DSN is required")
	}
	if cfg.RedisURL == "" {
		return workerConfig{}, errors.New("REDIS_URL is required")
	}
	if cfg.QueueKey == "" {
		cfg.QueueKey = "qiantie:pipeline:ready"
	}
	if cfg.FetchEndpoint == "" {
		cfg.FetchEndpoint = "https://txt.121w.com/api.php"
	}

	aiConfigured := cfg.AIBaseURL != "" || cfg.AIModel != "" || cfg.AIAPIKey != ""
	aiComplete := cfg.AIBaseURL != "" && cfg.AIModel != "" && cfg.AIAPIKey != ""
	if aiConfigured && !aiComplete {
		return workerConfig{}, errors.New("AI_BASE_URL, AI_MODEL and AI_API_KEY must be configured together")
	}
	return cfg, nil
}

func main() {
	cfg, err := configFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("mysql", cfg.MySQLDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(16)
	db.SetMaxIdleConns(16)
	db.SetConnMaxIdleTime(10 * time.Minute)
	db.SetConnMaxLifetime(time.Hour)

	pingCtx, cancelPing := context.WithTimeout(context.Background(), 15*time.Second)
	if err := db.PingContext(pingCtx); err != nil {
		cancelPing()
		log.Fatal(err)
	}
	cancelPing()

	source, err := queue.NewRedisSource(cfg.RedisURL, cfg.QueueKey)
	if err != nil {
		log.Fatal(err)
	}
	jobQueue, err := queue.NewRedisQueue(cfg.RedisURL, cfg.QueueKey)
	if err != nil {
		log.Fatal(err)
	}

	var classifier workerpkg.MetadataClassifier
	if cfg.AIBaseURL != "" {
		classifier, err = metadataai.New(metadataai.Config{
			BaseURL: cfg.AIBaseURL,
			Model:   cfg.AIModel,
			APIKey:  cfg.AIAPIKey,
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	jobStore := storage.NewSQLJobStore(db)
	progress := storage.NewSQLProgressStore(db)
	books := workerpkg.NewSQLIntakeRepository(db)
	batches := workerpkg.NewSQLBatchCreator(db)
	executor := workerpkg.IntakeExecutor{
		Books:      books,
		Fetcher:    one21.NewClient(cfg.FetchEndpoint),
		Classifier: classifier,
		Batches:    batches,
	}
	runner := workerpkg.Runner{
		Source:   source,
		Jobs:     jobStore,
		Executor: executor,
		Progress: progress,
		Queue:    jobQueue,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log.Printf("unified Go worker started; queue=%s ai_fallback=%t", cfg.QueueKey, classifier != nil)
	for ctx.Err() == nil {
		err := runner.RunOnce(ctx)
		if err == nil {
			continue
		}
		if errors.Is(err, context.Canceled) {
			break
		}
		if errors.Is(err, queue.ErrQueueEmpty) {
			continue
		}
		log.Printf("pipeline execution error: %v", err)
		time.Sleep(time.Second)
	}
}
