package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"qiantie/backend/internal/videomanagementconnector"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	token := strings.TrimSpace(os.Getenv("QIANTIE_VIDEO_CONNECTOR_TOKEN"))
	secret := strings.TrimSpace(os.Getenv("QIANTIE_VIDEO_CONNECTOR_SESSION_KEY"))
	if token == "" || secret == "" {
		return fmt.Errorf("connector token and session key are required")
	}
	root := env("QIANTIE_VIDEO_CONNECTOR_SESSION_DIR", "data/video-management-sessions")
	store, err := videomanagementconnector.NewFileSessionStore(root, []byte(secret))
	if err != nil {
		return err
	}
	service := &videomanagementconnector.TargetHTTPService{BaseURL: env("QIANTIE_VIDEO_CONNECTOR_TARGET_BASE_URL", "https://two.121w.com")}
	server := videomanagementconnector.NewServer(videomanagementconnector.Config{InternalToken: token}, service, store)
	listen := env("QIANTIE_VIDEO_CONNECTOR_LISTEN_ADDR", ":8791")
	httpServer := &http.Server{Addr: listen, Handler: server.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 120 * time.Second, WriteTimeout: 120 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	errCh := make(chan error, 1)
	go func() {
		log.Printf("video-management connector listening on %s", listen)
		errCh <- httpServer.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdown)
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

func env(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
