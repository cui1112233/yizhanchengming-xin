package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/app"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/provider121"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/webui"
	_ "github.com/go-sql-driver/mysql"
)

// BuildGitSHA is injected by scripts/build-embedded-ui.sh at package time.
var BuildGitSHA = "development"

func buildInfo() webui.BuildInfo {
	return webui.BuildInfo{GitSHA: BuildGitSHA}
}

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
	application, err := app.NewApplication(db, fetcher, nil, nil)
	if err != nil {
		fatal(logger, "initialize application failed", "app", "initialize", err)
	}
	ui := webui.NewHandlerWithAdmin(application.Handler, webui.EmbeddedFiles(), webui.EmbeddedAdminFiles(), buildInfo())
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fatal(logger, "listen failed", "http", "listen", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	server := &http.Server{Addr: addr, Handler: ui, ReadHeaderTimeout: 10 * time.Second}
	if err := serve(ctx, server, listener, application.Runtime); err != nil {
		fatal(logger, "http server stopped", "http", "listen", err)
	}
}

func serve(ctx context.Context, server *http.Server, listener net.Listener, runtime app.RuntimeLifecycle) error {
	if err := runtime.Start(ctx); err != nil {
		_ = listener.Close()
		return err
	}
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()

	select {
	case err := <-serveErr:
		closeErr := runtime.Close()
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		return errors.Join(err, closeErr)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		shutdownErr := server.Shutdown(shutdownCtx)
		cancel()
		closeErr := runtime.Close()
		serverErr := <-serveErr
		if errors.Is(serverErr, http.ErrServerClosed) {
			serverErr = nil
		}
		return errors.Join(shutdownErr, closeErr, serverErr)
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
