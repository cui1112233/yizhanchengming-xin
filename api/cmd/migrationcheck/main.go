// migrationcheck is a read-only schema consistency diagnostic. It never runs
// Goose and never issues DDL or DML.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var BuildGitSHA = "development"

type report struct {
	Code                       string   `json:"code"`
	Database                   string   `json:"database,omitempty"`
	BuildSHA                   string   `json:"buildSha"`
	Migrations                 []string `json:"migrations"`
	AppliedVersions            []int64  `json:"appliedVersions,omitempty"`
	WorkspacePreferencesExists bool     `json:"workspacePreferencesExists"`
	WorkspacePreferenceFields  []string `json:"workspacePreferenceFields,omitempty"`
	Classification             string   `json:"classification,omitempty"`
	Message                    string   `json:"message"`
}

func classifyWorkspacePreferences(applied, exists bool, fields []string) string {
	if !applied {
		return "A"
	}
	if !exists {
		return "B"
	}
	need := map[string]bool{"user_id": true, "theme": true, "notifications_enabled": true, "storage_preference": true, "updated_at": true}
	for _, field := range fields {
		delete(need, field)
	}
	if len(need) != 0 {
		return "B"
	}
	return "C"
}

func safeDiagnosticError(string) string { return "migrationcheck_database_unavailable" }

func main() {
	r := report{BuildSHA: BuildGitSHA, Migrations: migrationFiles()}
	dsn := os.Getenv("QIANTIE_MYSQL_DSN")
	if strings.TrimSpace(dsn) == "" {
		r.Code, r.Message = "MIGRATIONCHECK_DSN_MISSING", "QIANTIE_MYSQL_DSN is required"
		emit(r, 2)
		return
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		r.Code, r.Message = "MIGRATIONCHECK_DATABASE_UNAVAILABLE", safeDiagnosticError(err.Error())
		emit(r, 3)
		return
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		r.Code, r.Message = "MIGRATIONCHECK_DATABASE_UNAVAILABLE", safeDiagnosticError(err.Error())
		emit(r, 3)
		return
	}
	_ = db.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&r.Database)
	r.AppliedVersions = appliedVersions(ctx, db)
	r.WorkspacePreferencesExists = tableExists(ctx, db, "user_workspace_preferences")
	if r.WorkspacePreferencesExists {
		r.WorkspacePreferenceFields = tableFields(ctx, db, "user_workspace_preferences")
	}
	r.Classification = classifyWorkspacePreferences(hasVersion(r.AppliedVersions, 13), r.WorkspacePreferencesExists, r.WorkspacePreferenceFields)
	r.Code, r.Message = "MIGRATIONCHECK_OK", map[string]string{"A": "workspace preferences migration is not recorded", "B": "workspace preferences schema drift detected", "C": "workspace preferences schema matches migration 00013"}[r.Classification]
	emit(r, 0)
}

func emit(r report, code int) {
	_ = json.NewEncoder(os.Stdout).Encode(r)
	if code != 0 {
		os.Exit(code)
	}
}
func migrationFiles() []string {
	for _, p := range []string{"db/migrations", "api/db/migrations"} {
		if entries, err := os.ReadDir(p); err == nil {
			out := []string{}
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
					out = append(out, e.Name())
				}
			}
			sort.Strings(out)
			return out
		}
	}
	return []string{}
}
func appliedVersions(ctx context.Context, db *sql.DB) []int64 {
	rows, err := db.QueryContext(ctx, "SELECT version_id FROM goose_db_version WHERE is_applied=1 ORDER BY version_id")
	if err != nil {
		return []int64{}
	}
	defer rows.Close()
	out := []int64{}
	for rows.Next() {
		var v int64
		if rows.Scan(&v) == nil {
			out = append(out, v)
		}
	}
	return out
}
func tableExists(ctx context.Context, db *sql.DB, name string) bool {
	var n int
	return db.QueryRowContext(ctx, "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=?", name).Scan(&n) == nil && n == 1
}
func tableFields(ctx context.Context, db *sql.DB, name string) []string {
	rows, err := db.QueryContext(ctx, "SELECT column_name FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name=? ORDER BY ordinal_position", name)
	if err != nil {
		return []string{}
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var v string
		if rows.Scan(&v) == nil {
			out = append(out, v)
		}
	}
	return out
}
func hasVersion(values []int64, want int64) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
