package main

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"qiantie/backend/internal/ffmpegbin"
	"qiantie/backend/internal/mergeworker"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		if err := healthcheck(); err != nil {
			log.Print(err)
			os.Exit(1)
		}
		return
	}
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	listenAddr := env("QIANTIE_MERGE_WORKER_LISTEN_ADDR", ":8790")
	serviceKey := required("QIANTIE_MERGE_WORKER_SECRET")
	redisURL := required("QIANTIE_MERGE_REDIS_URL")
	if serviceKey == "" || redisURL == "" {
		return fmt.Errorf("merge worker secret and Redis URL are required")
	}
	store, err := mergeworker.NewRedisStore(redisURL, env("QIANTIE_MERGE_REDIS_PREFIX", "qiantie:bf11:merge"))
	if err != nil {
		return err
	}
	output, err := mergeworker.NewTOSObjectStore(mergeworker.TOSConfig{
		Endpoint:      required("QIANTIE_MERGE_TOS_ENDPOINT"),
		Region:        required("QIANTIE_MERGE_TOS_REGION"),
		Bucket:        required("QIANTIE_MERGE_TOS_BUCKET"),
		AccessKey:     required("QIANTIE_MERGE_TOS_ACCESS_KEY"),
		SecretKey:     required("QIANTIE_MERGE_TOS_SECRET_KEY"),
		PublicBaseURL: required("QIANTIE_MERGE_TOS_PUBLIC_BASE_URL"),
	})
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ffmpegbin.LogStartup()
	worker := &mergeworker.Worker{
		Store:      store,
		Queue:      store,
		Downloader: &mergeworker.Downloader{},
		Merger:     &mergeworker.FFmpegRunner{},
		Output:     output,
		WorkRoot:   env("QIANTIE_MERGE_WORK_ROOT", "/tmp/qiantie-merge"),
	}
	go func() {
		if err := worker.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("merge worker loop stopped: %v", err)
			stop()
		}
	}()

	api := mergeworker.NewHandler(serviceKey, store, store)
	mux := http.NewServeMux()
	mux.Handle("/v1/merge", api)
	mux.Handle("/v1/merge/", api)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if subtleHeaderMatch(r.Header.Get("Authorization"), serviceKey) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"ok"}`))
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	})
	server := &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		log.Printf("BF11 merge worker listening on %s", listenAddr)
		errCh <- server.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func healthcheck() error {
	secret := required("QIANTIE_MERGE_WORKER_SECRET")
	if secret == "" {
		return fmt.Errorf("merge worker healthcheck secret is required")
	}
	client := &http.Client{Timeout: 3 * time.Second}
	request, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:8790/healthz", nil)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+secret)
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("merge worker healthcheck failed")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("merge worker healthcheck returned HTTP %d", response.StatusCode)
	}
	return nil
}

func required(name string) string { return strings.TrimSpace(os.Getenv(name)) }
func env(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func subtleHeaderMatch(header, secret string) bool {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	candidate := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return len(candidate) == len(secret) && subtle.ConstantTimeCompare([]byte(candidate), []byte(secret)) == 1
}
