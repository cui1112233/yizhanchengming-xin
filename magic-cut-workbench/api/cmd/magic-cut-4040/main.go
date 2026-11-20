package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"yizhanchengming/magic-cut-workbench/internal/authbridge"
	"yizhanchengming/magic-cut-workbench/internal/config"
	"yizhanchengming/magic-cut-workbench/internal/httpapi"
	"yizhanchengming/magic-cut-workbench/internal/project"
	"yizhanchengming/magic-cut-workbench/web"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	accountResolver, err := authbridge.NewResolver(cfg.AuthBaseURL, &http.Client{Timeout: 15 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	var projectRepository project.Repository
	if cfg.DatabaseDSN != "" {
		db, openErr := sql.Open("mysql", cfg.DatabaseDSN)
		if openErr != nil {
			log.Fatal("open isolated Magic Cut database: ", openErr)
		}
		db.SetMaxOpenConns(12)
		db.SetMaxIdleConns(4)
		db.SetConnMaxLifetime(5 * time.Minute)
		projectRepository = project.NewStore(db, time.Now, project.NewID)
	}

	server := &http.Server{
		Addr: cfg.ListenAddr,
		Handler: httpapi.New(httpapi.Dependencies{
			Frontend:      web.Handler(),
			Auth:          authbridge.New(cfg.AuthBaseURL, &http.Client{Timeout: 15 * time.Second}),
			AuthBootstrap: authbridge.NewBootstrap(cfg.AuthBaseURL, &http.Client{Timeout: 15 * time.Second}),
			Projects:      project.NewHandler(projectRepository, accountResolver.Resolve),
		}),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

	log.Printf("magic-cut workbench listening on %s", cfg.ListenAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
