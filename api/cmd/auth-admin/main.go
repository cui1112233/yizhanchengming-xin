package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	logger := observability.NewJSONLogger(os.Stdout)
	slog.SetDefault(logger)
	dsn := os.Getenv("QIANTIE_MYSQL_DSN")
	if dsn == "" {
		authAdminFatal(logger, "required configuration is missing", "read_mysql_dsn", nil)
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		authAdminFatal(logger, "open mysql failed", "open_mysql", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		authAdminFatal(logger, "mysql check failed", "ping_mysql", err)
	}

	store := authn.NewMySQLStore(db)
	action := strings.TrimSpace(os.Getenv("QIANTIE_AUTH_ADMIN_ACTION"))
	username := strings.TrimSpace(os.Getenv("QIANTIE_AUTH_ADMIN_USERNAME"))
	password := os.Getenv("QIANTIE_AUTH_ADMIN_PASSWORD")

	switch action {
	case "bootstrap":
		created, err := store.EnsureInitialAdmin(ctx, username, password)
		if err != nil {
			authAdminFatal(logger, "bootstrap administrator failed", "bootstrap", err)
		}
		logger.Info("auth administrator bootstrap completed", "subsystem", "auth", "operation", "bootstrap", "created", created)
	case "set-password":
		if err := store.SetUserPassword(ctx, username, password); err != nil {
			authAdminFatal(logger, "set administrator password failed", "set_password", err)
		}
		logger.Info("auth administrator password updated", "subsystem", "auth", "operation", "set_password")
	case "disable":
		if err := store.SetUserActive(ctx, username, false); err != nil {
			authAdminFatal(logger, "disable user failed", "disable", err)
		}
		logger.Info("auth user disabled", "subsystem", "auth", "operation", "disable")
	case "enable":
		if err := store.SetUserActive(ctx, username, true); err != nil {
			authAdminFatal(logger, "enable user failed", "enable", err)
		}
		logger.Info("auth user enabled", "subsystem", "auth", "operation", "enable")
	default:
		authAdminFatal(logger, "invalid QIANTIE_AUTH_ADMIN_ACTION", "validate_action", nil)
	}
}

func authAdminFatal(logger *slog.Logger, message, operation string, err error) {
	args := []any{"subsystem", "auth", "operation", operation}
	if err != nil {
		args = append(args, "safe_error", observability.SafeError(err))
	}
	logger.Error(message, args...)
	os.Exit(1)
}
