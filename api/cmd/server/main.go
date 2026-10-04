package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/agent"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/batchfactory"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/httpapi"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/queue"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/storage"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/webui"
)

type serverConfig struct {
	MySQLDSN    string
	RedisURL    string
	ListenAddr  string
	QueueKey    string
	SystemOwner string
	AIBaseURL   string
	AIModel     string
	AIAPIKey    string
}

func configFromEnv() (serverConfig, error) {
	cfg := serverConfig{
		MySQLDSN:    strings.TrimSpace(os.Getenv("MYSQL_DSN")),
		RedisURL:    strings.TrimSpace(os.Getenv("REDIS_URL")),
		ListenAddr:  strings.TrimSpace(os.Getenv("LISTEN_ADDR")),
		QueueKey:    strings.TrimSpace(os.Getenv("PIPELINE_QUEUE_KEY")),
		SystemOwner: strings.TrimSpace(os.Getenv("SYSTEM_OWNER")),
		AIBaseURL:   strings.TrimSpace(os.Getenv("AI_BASE_URL")),
		AIModel:     strings.TrimSpace(os.Getenv("AI_MODEL")),
		AIAPIKey:    strings.TrimSpace(os.Getenv("AI_API_KEY")),
	}
	if cfg.MySQLDSN == "" {
		return serverConfig{}, errors.New("MYSQL_DSN is required")
	}
	if cfg.RedisURL == "" {
		return serverConfig{}, errors.New("REDIS_URL is required")
	}
	if cfg.SystemOwner == "" {
		return serverConfig{}, errors.New("SYSTEM_OWNER is required")
	}
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8080"
	}
	if cfg.QueueKey == "" {
		cfg.QueueKey = "qiantie:pipeline:ready"
	}
	aiConfigured := cfg.AIBaseURL != "" || cfg.AIModel != "" || cfg.AIAPIKey != ""
	aiComplete := cfg.AIBaseURL != "" && cfg.AIModel != "" && cfg.AIAPIKey != ""
	if aiConfigured && !aiComplete {
		return serverConfig{}, errors.New("AI_BASE_URL, AI_MODEL and AI_API_KEY must be configured together")
	}
	return cfg, nil
}

func composeHTTPHandler(api, frontend http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/", api)
	mux.Handle("/healthz", api)
	mux.Handle("/", frontend)
	return mux
}

func newAgentService(db *sql.DB, cfg serverConfig) (*agent.Service, error) {
	store := agent.NewSQLStore(db)
	rules := agent.NewResponder()
	var responder agent.ResponseGenerator = rules
	if cfg.AIBaseURL != "" {
		aiResponder, err := agent.NewAIResponder(agent.AIResponderConfig{
			BaseURL: cfg.AIBaseURL,
			APIKey:  cfg.AIAPIKey,
			Model:   cfg.AIModel,
		})
		if err != nil {
			return nil, err
		}
		responder = agent.NewHybridResponder(rules, aiResponder)
	}
	return &agent.Service{Store: store, Responder: responder}, nil
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
	db.SetMaxOpenConns(24)
	db.SetMaxIdleConns(24)
	db.SetConnMaxIdleTime(10 * time.Minute)
	db.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}

	jobQueue, err := queue.NewRedisQueue(cfg.RedisURL, cfg.QueueKey)
	if err != nil {
		log.Fatal(err)
	}
	starter := &batchfactory.StartService{
		Jobs:  storage.NewSQLJobStore(db),
		Queue: jobQueue,
	}
	intakes := &batchfactory.IntakeService{
		Store: storage.NewSQLIntakeStore(db),
	}
	ownerResolver := func(*http.Request) (string, error) {
		if cfg.SystemOwner == "" {
			return "", errors.New("unauthorized")
		}
		return cfg.SystemOwner, nil
	}
	agentService, err := newAgentService(db, cfg)
	if err != nil {
		log.Fatal(err)
	}
	apiHandler := httpapi.NewRouterWithAgent(starter, intakes, ownerResolver, agentService)
	frontendHandler, err := webui.Handler()
	if err != nil {
		log.Fatal(err)
	}
	handler := composeHTTPHandler(apiHandler, frontendHandler)
	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("unified Go server listening on %s agent_ai=%t", cfg.ListenAddr, cfg.AIBaseURL != "")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
