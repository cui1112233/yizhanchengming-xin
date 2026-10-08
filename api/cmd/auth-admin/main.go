package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
	"github.com/go-sql-driver/mysql"
)

const localFixtureUsernamePrefix = "ycm-staging-member-"

func main() {
	logger := observability.NewJSONLogger(os.Stdout)
	slog.SetDefault(logger)
	dsn := os.Getenv("QIANTIE_MYSQL_DSN")
	if dsn == "" {
		authAdminFatal(logger, "required configuration is missing", "read_mysql_dsn", nil)
	}
	action := strings.TrimSpace(os.Getenv("QIANTIE_AUTH_ADMIN_ACTION"))
	if strings.HasPrefix(action, "local-") {
		if err := validateLocalFixtureTarget(os.Getenv("QIANTIE_ENV"), dsn); err != nil {
			authAdminFatal(logger, "local acceptance fixture target rejected", "validate_local_fixture_target", err)
		}
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
	case "local-create-member":
		displayName := strings.TrimSpace(os.Getenv("QIANTIE_AUTH_ADMIN_DISPLAY_NAME"))
		if err := createLocalAcceptanceMember(ctx, db, username, displayName, password); err != nil {
			authAdminFatal(logger, "create local acceptance member failed", "local_create_member", err)
		}
		logger.Info("local acceptance member created", "subsystem", "auth", "operation", "local_create_member")
	case "local-delete-member":
		if err := deleteLocalAcceptanceMember(ctx, db, username); err != nil {
			authAdminFatal(logger, "delete local acceptance member failed", "local_delete_member", err)
		}
		logger.Info("local acceptance member deleted", "subsystem", "auth", "operation", "local_delete_member")
	default:
		authAdminFatal(logger, "invalid QIANTIE_AUTH_ADMIN_ACTION", "validate_action", nil)
	}
}

func validateLocalFixtureTarget(environment, dsn string) error {
	if strings.ToLower(strings.TrimSpace(environment)) != "local" {
		return errors.New("local acceptance fixtures require QIANTIE_ENV=local")
	}
	config, err := mysql.ParseDSN(dsn)
	if err != nil || config.Net != "tcp" || config.DBName != "ycm_staging" {
		return errors.New("local acceptance fixtures require the ycm_staging database over TCP loopback")
	}
	host, _, err := net.SplitHostPort(config.Addr)
	if err != nil || host != "127.0.0.1" {
		return errors.New("local acceptance fixtures require MySQL at 127.0.0.1")
	}
	return nil
}

func createLocalAcceptanceMember(ctx context.Context, db *sql.DB, username, displayName, password string) error {
	username = strings.TrimSpace(username)
	if !validLocalFixtureUsername(username) || len(strings.TrimSpace(password)) < 12 {
		return errors.New("reserved fixture username and password of at least 12 characters are required")
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		displayName = username
	}
	if len([]rune(displayName)) > 191 {
		return errors.New("local acceptance member display name is too long")
	}
	hash, err := authn.HashPassword(password)
	if err != nil {
		return errors.New("hash local acceptance member password")
	}
	result, err := db.ExecContext(ctx, `INSERT INTO auth_users (username, display_name, password_hash, role, active) VALUES (?, ?, ?, 'member', TRUE)`, username, displayName, hash)
	if err != nil {
		return fmt.Errorf("insert local acceptance member: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil || rows != 1 {
		return errors.New("verify local acceptance member creation")
	}
	return nil
}

func deleteLocalAcceptanceMember(ctx context.Context, db *sql.DB, username string) error {
	username = strings.TrimSpace(username)
	if !validLocalFixtureUsername(username) {
		return errors.New("reserved fixture username is required")
	}
	result, err := db.ExecContext(ctx, `DELETE FROM auth_users WHERE username = ? AND role = 'member' AND team_id IS NULL`, username)
	if err != nil {
		return fmt.Errorf("delete local acceptance member: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil || rows != 1 {
		return errors.New("local acceptance member not found or does not match the deletion guard")
	}
	return nil
}

func validLocalFixtureUsername(username string) bool {
	if !strings.HasPrefix(username, localFixtureUsernamePrefix) || len(username) > 191 || len(username) == len(localFixtureUsernamePrefix) {
		return false
	}
	for _, char := range username[len(localFixtureUsernamePrefix):] {
		if !(char >= 'a' && char <= 'z' || char >= '0' && char <= '9' || char == '-') {
			return false
		}
	}
	return true
}

func authAdminFatal(logger *slog.Logger, message, operation string, err error) {
	args := []any{"subsystem", "auth", "operation", operation}
	if err != nil {
		args = append(args, "safe_error", observability.SafeError(err))
	}
	logger.Error(message, args...)
	os.Exit(1)
}
