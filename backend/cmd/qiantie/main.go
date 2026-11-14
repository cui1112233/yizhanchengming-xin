package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"qiantie/backend/internal/app"
	"qiantie/backend/internal/config"
	"qiantie/backend/internal/ffmpegbin"
	webassets "qiantie/backend/web"
)

func main() {
	cfg, err := config.FromEnv()
	if err != nil {
		log.Fatal(err)
	}
	// Slice 默认值是 3（见 config.FromEnv）；显式设成 <1 时 V12 批次路由不会
	// 注册，表现为前端“作品全没了”，这里必须留下显式日志。
	if cfg.Slice < 1 {
		log.Printf("WARNING: QIANTIE_BATCH_FACTORY_V11_SLICE=%d，V12 批次路由（/api/batch-factory/v12/batches/summary 等）不会注册；本地请使用 scripts/start-local-go.ps1（最低兜底 1，默认 3）。", cfg.Slice)
	}
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		if err := runHealthcheck(healthURL(cfg.ListenAddr)); err != nil {
			log.Fatal(err)
		}
		return
	}
	ffmpegbin.LogStartup()
	dsn, err := databaseDSN(cfg.MySQLDSN)
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	configureDatabasePool(db)
	// Startup includes durable MySQL DDL. A slow commit can exceed 30 seconds
	// and leave an implicitly committed migration without its ledger entry.
	// Keep a bounded startup budget without interrupting normal recovery.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	apiHandler, err := app.Build(ctx, cfg, db, nil)
	if err != nil {
		log.Fatal(err)
	}
	staticHandler := webassets.Handler()
	handler := newRuntimeHandler(apiHandler, staticHandler)
	server := &http.Server{Addr: cfg.ListenAddr, Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	log.Printf("qiantie Go V12 listening on %s", cfg.ListenAddr)
	log.Fatal(server.ListenAndServe())
}

// newRuntimeHandler is the single live HTTP entry point: API requests and
// health checks are handled by Go, while all other requests are served from
// the embedded web assets. Unknown API paths remain explicit Go 404/405
// responses; there is no Node bridge or slice-based fallback.
func newRuntimeHandler(apiHandler, staticHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req != nil && req.URL != nil && (req.URL.Path == "/health" || strings.HasPrefix(req.URL.Path, "/api/") || strings.HasPrefix(req.URL.Path, "/downloads/local-executor/") || strings.HasPrefix(req.URL.Path, "/user-content/avatars/")) {
			apiHandler.ServeHTTP(w, req)
			return
		}
		staticHandler.ServeHTTP(w, req)
	})
}

func databaseDSN(raw string) (string, error) {
	cfg, err := mysql.ParseDSN(raw)
	if err != nil {
		return "", fmt.Errorf("parse MySQL DSN: %w", err)
	}
	// Avoid a server-side PREPARE/EXECUTE round trip for each parameterized
	// read while preserving all options already present in the DSN.
	cfg.InterpolateParams = true
	return cfg.FormatDSN(), nil
}

func configureDatabasePool(db *sql.DB) {
	// Status polling arrives in short bursts. Reuse connections instead of
	// reopening and authenticating for every refresh.
	// The local MySQL runs in Docker behind the Windows port proxy, which
	// silently drops idle TCP connections. A large long-lived idle pool then
	// hands out dead connections and transactions fail with
	// "driver: bad connection". Keep the idle pool small and short-lived so
	// stale connections are discarded before the proxy kills them.
	db.SetMaxOpenConns(24)
	db.SetMaxIdleConns(4)
	db.SetConnMaxIdleTime(90 * time.Second)
	db.SetConnMaxLifetime(5 * time.Minute)
}

func healthURL(listenAddr string) string {
	addr := strings.TrimSpace(listenAddr)
	switch {
	case strings.HasPrefix(addr, ":"):
		addr = "127.0.0.1" + addr
	case strings.HasPrefix(addr, "0.0.0.0:"):
		addr = "127.0.0.1" + strings.TrimPrefix(addr, "0.0.0.0")
	case strings.HasPrefix(addr, "[::]:"):
		addr = "127.0.0.1:" + strings.TrimPrefix(addr, "[::]:")
	}
	return "http://" + addr + "/health"
}

func runHealthcheck(url string) error {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health endpoint returned %s", resp.Status)
	}
	return nil
}
