package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"qiantie/backend/internal/giantmaterialplugin"
)

func main() {
	cfg, err := giantmaterialplugin.ConfigFromEnv(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           giantmaterialplugin.NewQingyuHandler(cfg.PluginToken, cfg.UpstreamEndpoint, cfg.UpstreamToken),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("giant material HTTP plugin listening on %s", cfg.ListenAddr)
	log.Fatal(server.ListenAndServe())
}
