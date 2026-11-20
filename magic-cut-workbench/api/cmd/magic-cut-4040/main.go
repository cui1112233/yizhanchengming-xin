package main

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"yizhanchengming/magic-cut-workbench/internal/authbridge"
	"yizhanchengming/magic-cut-workbench/internal/config"
	workexport "yizhanchengming/magic-cut-workbench/internal/export"
	"yizhanchengming/magic-cut-workbench/internal/httpapi"
	"yizhanchengming/magic-cut-workbench/internal/material"
	"yizhanchengming/magic-cut-workbench/internal/project"
	worktask "yizhanchengming/magic-cut-workbench/internal/task"
	"yizhanchengming/magic-cut-workbench/internal/timeline"
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
	var db *sql.DB
	var projectRepository project.Repository
	var materialHandler http.Handler
	var timelineHandler http.Handler
	var taskHandler http.Handler
	var exportHandler http.Handler
	if cfg.DatabaseDSN != "" {
		openedDB, openErr := sql.Open("mysql", cfg.DatabaseDSN)
		if openErr != nil {
			log.Fatal("open isolated Magic Cut database: ", openErr)
		}
		db = openedDB
		db.SetMaxOpenConns(12)
		db.SetMaxIdleConns(4)
		db.SetConnMaxLifetime(5 * time.Minute)
		projectRepository = project.NewStore(db, time.Now, project.NewID)
		materialRepository := material.NewStore(db)
		materialService := material.NewService(material.ServiceOptions{
			StorageRoot:   cfg.StorageRoot,
			MaxBytes:      10 << 30,
			Probe:         material.FFprobe{},
			Repository:    materialRepository,
			NewMaterialID: material.NewID,
			NewTaskID:     material.NewTaskID,
			Now:           time.Now,
		})
		materialHandler = material.NewHandler(materialService, materialRepository, accountResolver.Resolve)
		timelineRepository := timeline.NewStore(db, time.Now)
		timelineHandler = timeline.NewHandler(timelineRepository, accountResolver.Resolve)
		exportService := workexport.NewService(workexport.ServiceOptions{
			StorageRoot: cfg.StorageRoot,
			Timelines:   timelineRepository,
			Materials:   materialRepository,
			Repository:  workexport.NewSQLRepository(db),
			Renderer:    workexport.FFmpeg{},
			Probe:       material.FFprobe{},
			NewExportID: func() string { return "mce_" + strings.TrimPrefix(material.NewTaskID(), "mct_") },
			NewTaskID:   material.NewTaskID,
			Now:         time.Now,
		})
		taskHandler = worktask.NewHandlerWithActions(worktask.NewSQLStore(db, time.Now), accountResolver.Resolve, exportService)
		exportHandler = workexport.NewHandler(exportService, accountResolver.Resolve)
	}

	server := &http.Server{
		Addr: cfg.ListenAddr,
		Handler: httpapi.New(httpapi.Dependencies{
			Frontend:      web.Handler(),
			Auth:          authbridge.New(cfg.AuthBaseURL, &http.Client{Timeout: 15 * time.Second}),
			AuthBootstrap: authbridge.NewBootstrap(cfg.AuthBaseURL, &http.Client{Timeout: 15 * time.Second}),
			Projects:      project.NewHandler(projectRepository, accountResolver.Resolve),
			Materials:     materialHandler,
			Timeline:      timelineHandler,
			Tasks:         taskHandler,
			Exports:       exportHandler,
		}),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       20 * time.Minute,
		WriteTimeout:      20 * time.Minute,
		IdleTimeout:       90 * time.Second,
	}

	log.Printf("magic-cut workbench listening on %s", cfg.ListenAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
