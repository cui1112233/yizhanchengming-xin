package main

import (
	"log"
	"net/http"
	"time"

	"yizhanchengming/magic-cut-workbench/internal/config"
	"yizhanchengming/magic-cut-workbench/internal/httpapi"
	"yizhanchengming/magic-cut-workbench/web"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           httpapi.New(httpapi.Dependencies{Frontend: web.Handler()}),
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
