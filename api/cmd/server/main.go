package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/app"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/provider121"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	logger := observability.NewJSONLogger(os.Stdout)
	slog.SetDefault(logger)

	dsn := os.Getenv("QIANTIE_MYSQL_DSN")
	if dsn == "" {
		fatal(logger, "required configuration is missing", "bootstrap", "read_mysql_dsn", nil)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fatal(logger, "open mysql failed", "mysql", "open", err)
	}
	defer db.Close()

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := db.PingContext(pingCtx); err != nil {
		pingCancel()
		fatal(logger, "mysql startup check failed", "mysql", "ping", err)
	}
	pingCancel()

	bootstrapCtx, bootstrapCancel := context.WithTimeout(context.Background(), 15*time.Second)
	created, err := authn.NewMySQLStore(db).EnsureInitialAdmin(
		bootstrapCtx,
		strings.TrimSpace(os.Getenv("QIANTIE_BOOTSTRAP_ADMIN_USERNAME")),
		os.Getenv("QIANTIE_BOOTSTRAP_ADMIN_PASSWORD"),
	)
	bootstrapCancel()
	if err != nil {
		fatal(logger, "initialize auth administrator failed", "auth", "bootstrap_admin", err)
	}
	if created {
		logger.Info("initial auth administrator created", "subsystem", "auth", "operation", "bootstrap_admin")
	}

	fetcher, err := provider121.NewClient(provider121.Config{})
	if err != nil {
		fatal(logger, "create 121 client failed", "novel_fetch", "create_client", err)
	}

	addr := os.Getenv("QIANTIE_GO_LISTEN_ADDR")
	if addr == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		addr = ":" + port
	}

	logger.Info("api listening", "subsystem", "http", "operation", "listen", "address", observability.SanitizeString(addr))
	if err := http.ListenAndServe(addr, app.NewHandler(db, fetcher, nil, nil)); err != nil {
		fatal(logger, "http server stopped", "http", "listen", err)
	}
}

func fatal(logger *slog.Logger, message, subsystem, operation string, err error) {
	args := []any{"subsystem", subsystem, "operation", operation}
	if err != nil {
		args = append(args, "safe_error", observability.SafeError(err))
	}
	logger.Error(message, args...)
	os.Exit(1)
}
