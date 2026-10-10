package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"qiantie/backend/internal/app"
	"qiantie/backend/internal/config"
	webassets "qiantie/backend/web"
)

func main() {
	cfg, err := config.FromEnv()
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		if err := runHealthcheck(healthURL(cfg.ListenAddr)); err != nil {
			log.Fatal(err)
		}
		return
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	apiHandler, err := app.Build(ctx, cfg, db, nil)
	if err != nil {
		log.Fatal(err)
	}
	staticHandler := webassets.Handler()
	authProxy, err := newAuthProxy(envOrDefault("QIANTIE_NODE_AUTH_BASE_URL", "http://127.0.0.1:3001"))
	if err != nil {
		log.Fatal(err)
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/api/login" || strings.HasPrefix(req.URL.Path, "/api/login/") {
			authProxy.ServeHTTP(w, req)
			return
		}
		if req.URL.Path == "/health" || strings.HasPrefix(req.URL.Path, "/api/") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		staticHandler.ServeHTTP(w, req)
	})
	server := &http.Server{Addr: cfg.ListenAddr, Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	log.Printf("qiantie Go V11 listening on %s", cfg.ListenAddr)
	log.Fatal(server.ListenAndServe())
}

func newAuthProxy(rawTarget string) (http.Handler, error) {
	target, err := url.Parse(strings.TrimSpace(rawTarget))
	if err != nil || target.Scheme != "http" || target.Host == "" {
		return nil, fmt.Errorf("invalid local auth bridge target")
	}
	host := strings.ToLower(target.Hostname())
	if host != "127.0.0.1" && host != "localhost" && host != "::1" && host != "backend" {
		return nil, fmt.Errorf("auth bridge target must be local")
	}
	return httputil.NewSingleHostReverseProxy(target), nil
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func databaseDSN(raw string) (string, error) {
	cfg, err := mysql.ParseDSN(raw)
	if err != nil {
		return "", fmt.Errorf("parse MySQL DSN: %w", err)
	}
	// database/sql otherwise prepares and closes a server-side statement for
	// every parameterized read. The public workbench is read-heavy, so inline
	// parameters to keep MySQL from spending CPU on matching PREPARE/EXECUTE
	// churn. Existing DSN options (charset, parseTime, timeouts) are preserved.
	cfg.InterpolateParams = true
	return cfg.FormatDSN(), nil
}

func configureDatabasePool(db *sql.DB) {
	// Runtime status refreshes arrive in short bursts. Keep those connections
	// reusable instead of closing back to database/sql's default of two idle
	// connections and reopening/authenticating them on every poll.
	db.SetMaxOpenConns(24)
	db.SetMaxIdleConns(24)
	db.SetConnMaxIdleTime(10 * time.Minute)
	db.SetConnMaxLifetime(time.Hour)
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
