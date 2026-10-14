package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
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
	"qiantie/backend/internal/runtimeaudit"
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
	nodeAuthTarget := strings.TrimSpace(os.Getenv("QIANTIE_NODE_AUTH_BASE_URL"))
	goAuth := goAuthEnabled(strings.TrimSpace(nodeAuthTarget))
	// The Go runtime is self-contained by default. A legacy Node bridge is
	// only created when an explicit target is supplied for a temporary
	// migration check; no request is sent to port 3001 implicitly.
	authProxy := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNotImplemented)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "该功能尚未迁移到 Go 服务"})
	}))
	if nodeAuthTarget != "" {
		proxy, proxyErr := newAuthProxy(nodeAuthTarget)
		if proxyErr != nil {
			log.Fatal(proxyErr)
		}
		authProxy = proxy
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if (req.Method == http.MethodGet || req.Method == http.MethodPost) && strings.HasPrefix(req.URL.Path, "/api/login/invite/") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if !goAuth && (req.URL.Path == "/api/login" || strings.HasPrefix(req.URL.Path, "/api/login/")) {
			if req.Method == http.MethodPost && (req.URL.Path == "/api/login/passkey/options" || req.URL.Path == "/api/login/passkey/verify") {
				apiHandler.ServeHTTP(w, req)
				return
			}
			runtimeaudit.Default().RecordBridge(req.URL.Path)
			authProxy.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && (req.URL.Path == "/api/login/passkey/options" || req.URL.Path == "/api/login/passkey/verify") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if strings.HasPrefix(req.URL.Path, "/api/login/") && req.URL.Path != "/api/login/logout" && req.URL.Path != "/api/login/session" {
			runtimeaudit.Default().RecordBridge(req.URL.Path)
			authProxy.ServeHTTP(w, req)
			return
		}
		// /novel-fetch and its complete batch-rewrite API are served by Go.
		// Unknown endpoints stay on Go and return an explicit response; they are
		// never silently sent to the retired Node service.
		if isGoNovelFetchCompatRequest(req) || isGoAgentReadRequest(req) {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if (req.Method == http.MethodGet && (req.URL.Path == "/api/team-admin/teams" || strings.HasPrefix(req.URL.Path, "/api/team-admin/teams/") && strings.HasSuffix(req.URL.Path, "/members"))) || (req.Method == http.MethodPatch && strings.HasPrefix(req.URL.Path, "/api/team-admin/teams/") && strings.HasSuffix(req.URL.Path, "/governance")) || (req.Method == http.MethodPut && (strings.HasSuffix(req.URL.Path, "/co-managers") || strings.HasSuffix(req.URL.Path, "/api-scopes"))) || (req.Method == http.MethodPost && (strings.HasSuffix(req.URL.Path, "/status") || strings.HasSuffix(req.URL.Path, "/reset-password"))) {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if (req.Method == http.MethodGet && req.URL.Path == "/api/account-recovery/passkeys") || (req.Method == http.MethodDelete && strings.HasPrefix(req.URL.Path, "/api/account-recovery/passkeys/")) {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && req.URL.Path == "/api/account-recovery/passkeys/options" {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && req.URL.Path == "/api/account-recovery/passkeys" {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodGet && req.URL.Path == "/api/member/team/audit" {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodGet && strings.HasPrefix(req.URL.Path, "/api/account-admin/accounts/") && strings.HasSuffix(req.URL.Path, "/transfer-preview") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if (req.Method == http.MethodPost && req.URL.Path == "/api/client-errors") || (req.Method == http.MethodGet && req.URL.Path == "/api/client-errors/mine") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && strings.HasPrefix(req.URL.Path, "/api/account-admin/accounts/") && strings.HasSuffix(req.URL.Path, "/transfer") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if (req.Method == http.MethodGet && (req.URL.Path == "/api/build-info" || req.URL.Path == "/api/novel-panel/build-info" || req.URL.Path == "/api/novel-panel/character-core/health" || req.URL.Path == "/api/novel-panel/clean-core/health")) || isAdminAgentSkillRequest(req) || isGoAdminReadRequest(req) {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if (req.Method == http.MethodGet && req.URL.Path == "/api/novel-fetch-upload/upload-session") || (req.Method == http.MethodPost && (req.URL.Path == "/api/novel-fetch-upload/upload-login" || req.URL.Path == "/api/novel-fetch-upload/upload-batch")) {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && (req.URL.Path == "/api/novel-panel/character-core/parse-slots" || req.URL.Path == "/api/novel-panel/character-core/resolve-scene-cast" || req.URL.Path == "/api/novel-panel/character-core/build-scene-context" || req.URL.Path == "/api/novel-panel/character-core/migrate-project" || req.URL.Path == "/api/novel-panel/character-core/analyze" || req.URL.Path == "/api/novel-panel/character-core/project-lease") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && req.URL.Path == "/api/novel-panel/analyze" {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && strings.HasPrefix(req.URL.Path, "/api/novel-panel/") && (strings.HasSuffix(req.URL.Path, "/optimize-character") || strings.HasSuffix(req.URL.Path, "/optimize-character-copy") || strings.HasSuffix(req.URL.Path, "/optimize-all-characters") || strings.HasSuffix(req.URL.Path, "/optimize-style-copy") || strings.HasSuffix(req.URL.Path, "/instruction-assist")) {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && (req.URL.Path == "/api/novel-panel/outline-scenes" || req.URL.Path == "/api/novel-panel/regenerate-scene-outline" || req.URL.Path == "/api/novel-panel/generate-scene-prompts") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && req.URL.Path == "/api/novel-panel/reference-assets/describe" {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && req.URL.Path == "/api/novel-panel/reference-assets/generate" {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && req.URL.Path == "/api/settings/test" {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && req.URL.Path == "/api/novel-panel/native-clipboard/copy-rich" {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if (req.Method == http.MethodGet || req.Method == http.MethodPost) && (req.URL.Path == "/api/batch-rewrite/web-submit/config" || req.URL.Path == "/api/batch-rewrite/web-submit/environment" || req.URL.Path == "/api/batch-rewrite/web-submit/history") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodGet && req.URL.Path == "/api/batch-rewrite/issues" {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && (req.URL.Path == "/api/batch-rewrite/web-submit/test-visible" || req.URL.Path == "/api/batch-rewrite/web-submit/sync-configs" || req.URL.Path == "/api/batch-rewrite/web-submit/sync-styles" || req.URL.Path == "/api/batch-rewrite/web-submit/preview" || req.URL.Path == "/api/batch-rewrite/web-submit/submit") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodGet && req.URL.Path == "/api/batch-factory/v12/publish-121/organizations" {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && req.URL.Path == "/api/config/models/test" {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && (req.URL.Path == "/api/test/text" || req.URL.Path == "/api/test/image") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && strings.HasPrefix(req.URL.Path, "/api/batch-rewrite/knowledge/") && strings.HasSuffix(req.URL.Path, "/optimize") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && strings.HasPrefix(req.URL.Path, "/api/member/security/mfa/") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && (req.URL.Path == "/api/account-recovery/email/request" || req.URL.Path == "/api/account-recovery/email/verify" || req.URL.Path == "/api/account-recovery/password/request" || req.URL.Path == "/api/account-recovery/password/reset") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && req.URL.Path == "/api/novel-panel/character-core/trace" {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if strings.HasPrefix(req.URL.Path, "/api/novel-panel/history") || strings.HasPrefix(req.URL.Path, "/api/novel-panel/settings") || strings.HasPrefix(req.URL.Path, "/api/novel-panel/runtime-config") || strings.HasPrefix(req.URL.Path, "/api/novel-panel/image-settings") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodGet && strings.HasPrefix(req.URL.Path, "/api/novel-panel/reference-assets/public/") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodGet && req.URL.Path == "/api/novel-panel/reference-assets/legacy" {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if (req.Method == http.MethodGet && (req.URL.Path == "/api/novel-panel/diagnostics" || req.URL.Path == "/api/novel-panel/diagnostics/traces" || req.URL.Path == "/api/novel-panel/diagnostics/self-check" || req.URL.Path == "/api/novel-panel/projects" || req.URL.Path == "/api/novel-panel/draft" || strings.HasPrefix(req.URL.Path, "/api/novel-panel/projects/") || strings.HasPrefix(req.URL.Path, "/api/novel-panel/diagnostics/trace/"))) || (req.Method == http.MethodPut && req.URL.Path == "/api/novel-panel/draft") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if (req.Method == http.MethodPost && req.URL.Path == "/api/novel-panel/projects") || (req.Method == http.MethodDelete && strings.HasPrefix(req.URL.Path, "/api/novel-panel/projects/")) {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && strings.HasPrefix(req.URL.Path, "/api/novel-panel/") && strings.HasSuffix(req.URL.Path, "/export") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if (req.Method == http.MethodGet && strings.HasPrefix(req.URL.Path, "/api/reference-assets/file/")) || (req.Method == http.MethodPost && (req.URL.Path == "/api/reference-assets/upload" || req.URL.Path == "/api/reference-assets/use-source-as-main")) {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if (req.Method == http.MethodGet && (req.URL.Path == "/api/member/me" || req.URL.Path == "/api/member/team" || req.URL.Path == "/api/member/team/meta" || req.URL.Path == "/api/member/team/governance" || req.URL.Path == "/api/member/team/invites" || req.URL.Path == "/api/admin/error-logs" || req.URL.Path == "/api/admin/accounts" || strings.HasPrefix(req.URL.Path, "/api/member/team/members/") && strings.HasSuffix(req.URL.Path, "/usage") || req.URL.Path == "/api/account-recovery/email/status")) || (req.Method == http.MethodPost && (req.URL.Path == "/api/account-admin/accounts" || req.URL.Path == "/api/member/team/members" || req.URL.Path == "/api/member/team/invites" || strings.HasSuffix(req.URL.Path, "/status") || strings.HasSuffix(req.URL.Path, "/reset-password") || strings.HasSuffix(req.URL.Path, "/archive") || strings.HasSuffix(req.URL.Path, "/restore") || strings.HasSuffix(req.URL.Path, "/api"))) || (req.Method == http.MethodPut && strings.HasSuffix(req.URL.Path, "/api-scopes")) || (req.Method == http.MethodPatch && (req.URL.Path == "/api/member/profile" || strings.HasPrefix(req.URL.Path, "/api/account-admin/accounts/") || strings.HasPrefix(req.URL.Path, "/api/member/team/members/") || strings.HasPrefix(req.URL.Path, "/api/member/team/governance/") || strings.HasPrefix(req.URL.Path, "/api/member/team/meta/"))) || (req.Method == http.MethodPost && req.URL.Path == "/api/member/avatar") || (req.Method == http.MethodPost && req.URL.Path == "/api/member/security/password") || (req.Method == http.MethodGet && req.URL.Path == "/api/member/security") || (req.Method == http.MethodPost && req.URL.Path == "/api/member/security/sessions/revoke-others") || (req.Method == http.MethodGet && req.URL.Path == "/api/member/notifications") || (req.Method == http.MethodPost && req.URL.Path == "/api/member/notifications/read-all") || (req.Method == http.MethodPost && strings.HasPrefix(req.URL.Path, "/api/member/notifications/") && strings.HasSuffix(req.URL.Path, "/read")) || ((req.Method == http.MethodGet || req.Method == http.MethodPost) && req.URL.Path == "/api/config") || ((req.Method == http.MethodGet && (req.URL.Path == "/api/batch-rewrite/config" || req.URL.Path == "/api/batch-rewrite/bootstrap" || req.URL.Path == "/api/batch-rewrite/tasks" || strings.HasPrefix(req.URL.Path, "/api/batch-rewrite/tasks/"))) || (req.Method == http.MethodPost && (req.URL.Path == "/api/batch-rewrite/work-form" || req.URL.Path == "/api/batch-rewrite/tasks/batch-delete" || req.URL.Path == "/api/batch-rewrite/tasks/batch-retry" || strings.HasSuffix(req.URL.Path, "/fetch") || strings.HasSuffix(req.URL.Path, "/restore-original") || strings.HasSuffix(req.URL.Path, "/generate-ai")))) || ((req.Method == http.MethodGet && req.URL.Path == "/api/storage/list") || (req.Method == http.MethodPost && (req.URL.Path == "/api/storage/restore" || req.URL.Path == "/api/storage/save-media"))) || (req.Method == http.MethodGet && req.URL.Path == "/api/novel-fetch/saved") || (req.Method == http.MethodPost && (req.URL.Path == "/api/novel-fetch" || req.URL.Path == "/api/novel-fetch/save")) || (req.Method == http.MethodGet && req.URL.Path == "/api/platform-projects") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if strings.HasPrefix(req.URL.Path, "/downloads/local-executor/") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if shouldBridgeLegacyAPIRequest(req, cfg.Slice) {
			runtimeaudit.Default().RecordBridge(req.URL.Path)
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

func isLegacyNovelFetchBridgeRequest(req *http.Request) bool {
	return req != nil && req.URL != nil && strings.HasPrefix(req.URL.Path, "/api/batch-rewrite") && !isGoNovelFetchCompatRequest(req)
}

func isGoAgentReadRequest(req *http.Request) bool {
	if req == nil || req.URL == nil || (req.Method != http.MethodGet && req.Method != http.MethodPost && req.Method != http.MethodPatch && req.Method != http.MethodDelete) {
		return false
	}
	return req.URL.Path == "/api/agent/chat" || req.URL.Path == "/api/agent/tasks" || strings.HasPrefix(req.URL.Path, "/api/agent/tasks/") || req.URL.Path == "/api/agent/skills" || strings.HasPrefix(req.URL.Path, "/api/agent/skills/")
}

func isAdminAgentSkillRequest(req *http.Request) bool {
	if req == nil || req.URL == nil {
		return false
	}
	if req.Method == http.MethodGet {
		return req.URL.Path == "/api/admin/agent-skills" || strings.HasPrefix(req.URL.Path, "/api/admin/agent-skills/")
	}
	if req.Method != http.MethodPost {
		return false
	}
	if req.URL.Path == "/api/admin/agent-skills/draft" {
		return true
	}
	return strings.HasPrefix(req.URL.Path, "/api/admin/agent-skills/") && (strings.HasSuffix(req.URL.Path, "/publish") || strings.HasSuffix(req.URL.Path, "/archive"))
}

func isGoAdminReadRequest(req *http.Request) bool {
	if req == nil || req.URL == nil {
		return false
	}
	path := req.URL.Path
	if path == "/api/admin/accounts" {
		return req.Method == http.MethodGet || req.Method == http.MethodPost
	}
	if strings.HasPrefix(path, "/api/admin/accounts/") {
		return req.Method == http.MethodPost && (strings.HasSuffix(path, "/status") || strings.HasSuffix(path, "/reset-password"))
	}
	if path == "/api/admin/applications" || path == "/api/admin/grants" {
		return req.Method == http.MethodGet
	}
	if strings.HasPrefix(path, "/api/admin/applications/") {
		return req.Method == http.MethodPost && (strings.HasSuffix(path, "/approve") || strings.HasSuffix(path, "/reject"))
	}
	if strings.HasPrefix(path, "/api/admin/grants/") {
		return req.Method == http.MethodDelete
	}
	if path == "/api/admin/grants" {
		return req.Method == http.MethodPost
	}
	if req.Method == http.MethodGet {
		return path == "/api/admin/presets" || path == "/api/admin/preset-slots" || adminPresetVersionPath(path)
	}
	if req.Method == http.MethodPost {
		return path == "/api/admin/presets/draft" || adminPresetActionPath(path, "publish") || adminPresetActionPath(path, "rollback")
	}
	return req.Method == http.MethodPut && adminPresetDraftPath(path)
}

func adminPresetVersionPath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return len(parts) == 5 && parts[0] == "api" && parts[1] == "admin" && parts[2] == "presets" && parts[3] != "" && parts[4] != ""
}

func adminPresetDraftPath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return len(parts) == 6 && parts[0] == "api" && parts[1] == "admin" && parts[2] == "presets" && parts[5] == "draft"
}

func adminPresetActionPath(path, action string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	return len(parts) == 5 && parts[0] == "api" && parts[1] == "admin" && parts[2] == "presets" && parts[4] == action
}

func isGoNovelFetchCompatRequest(req *http.Request) bool {
	if req == nil || req.URL == nil {
		return false
	}
	path := req.URL.Path
	// The complete novel-fetch workbench belongs to the Go runtime now. Keep
	// the whole namespace here (including future/optional endpoints) so an
	// unimplemented route returns a Go response instead of silently falling
	// back to the retired Node service on port 3001.
	if strings.HasPrefix(path, "/api/batch-rewrite") ||
		strings.HasPrefix(path, "/api/novel-fetch-workshop") ||
		strings.HasPrefix(path, "/api/novel-fetch-upload") ||
		strings.HasPrefix(path, "/api/novel-fetch") {
		return true
	}
	// These two routes only read/write the user's Go workshop store. They do
	// not call an upstream provider, so they can safely bypass the old Node
	// bridge while the actual novel-fetch request remains staged.
	if (req.Method == http.MethodGet && path == "/api/novel-fetch/saved") ||
		(req.Method == http.MethodPost && path == "/api/novel-fetch") ||
		(req.Method == http.MethodPost && path == "/api/novel-fetch/save") ||
		(req.Method == http.MethodGet && (path == "/api/batch-rewrite/tasks" || isBatchRewriteTaskDetailPath(path) || isBatchRewriteTaskDiagnosticPath(path) || path == "/api/batch-rewrite/config" || path == "/api/batch-rewrite/bootstrap" || strings.HasPrefix(path, "/api/batch-rewrite/process/"))) ||
		(req.Method == http.MethodPost && (path == "/api/batch-rewrite/config" || path == "/api/batch-rewrite/process" || strings.HasPrefix(path, "/api/batch-rewrite/process/") || path == "/api/batch-rewrite/ai/test" || strings.HasPrefix(path, "/api/batch-rewrite/opening/") || path == "/api/batch-rewrite/rules/ai-suggest" || path == "/api/batch-rewrite/rules/preview")) ||
		(req.Method == http.MethodPost && (path == "/api/batch-rewrite/knowledge/optimize" || path == "/api/batch-rewrite/tasks/apply-rules" || path == "/api/batch-rewrite/tasks/reprocess-sensitive" || path == "/api/batch-rewrite/tasks/batch-ai-count" || path == "/api/batch-rewrite/web-submit/operations" || path == "/api/batch-rewrite/schedules")) ||
		(req.Method == http.MethodPost && (strings.HasSuffix(path, "/tasks/batch-retry") || strings.HasSuffix(path, "/generate-ai") || (strings.Contains(path, "/tasks/") && (strings.HasSuffix(path, "/fetch") || strings.HasSuffix(path, "/restore-original"))))) ||
		(req.Method == http.MethodGet && (path == "/api/batch-rewrite/logs" || path == "/api/batch-rewrite/schedules" || path == "/api/batch-rewrite/batches/current" || strings.HasPrefix(path, "/api/batch-rewrite/web-submit/operations/"))) ||
		((req.Method == http.MethodPatch || req.Method == http.MethodDelete) && strings.HasPrefix(path, "/api/batch-rewrite/schedules/")) ||
		(req.Method == http.MethodGet && (path == "/api/batch-rewrite/knowledge" || strings.HasPrefix(path, "/api/batch-rewrite/knowledge/"))) ||
		(req.Method == http.MethodPost && path == "/api/batch-rewrite/knowledge") ||
		(req.Method == http.MethodDelete && strings.HasPrefix(path, "/api/batch-rewrite/knowledge/")) {
		return true
	}
	// These legacy batch-rewrite routes already have Go compatibility
	// handlers registered in internal/httpapi. Keep them on Go instead of
	// letting the broad legacy bridge send them back to Node on port 3001.
	if isGoBatchRewriteCompatRequest(req) {
		return true
	}
	// The complete namespace is handled above, including future endpoints.
	return false
}

func isGoBatchRewriteCompatRequest(req *http.Request) bool {
	if req == nil || req.URL == nil {
		return false
	}
	path := req.URL.Path
	if req.Method == http.MethodGet && path == "/api/batch-rewrite/issues" {
		return true
	}
	if path == "/api/batch-rewrite/config" || path == "/api/batch-rewrite/bootstrap" {
		return req.Method == http.MethodGet
	}
	if path == "/api/batch-rewrite/tasks/batch-delete" || path == "/api/batch-rewrite/work-form" {
		return req.Method == http.MethodPost
	}
	if strings.HasPrefix(path, "/api/batch-rewrite/web-submit/") {
		switch path {
		case "/api/batch-rewrite/web-submit/config", "/api/batch-rewrite/web-submit/environment", "/api/batch-rewrite/web-submit/history":
			return req.Method == http.MethodGet || req.Method == http.MethodPost
		case "/api/batch-rewrite/web-submit/test-visible", "/api/batch-rewrite/web-submit/sync-configs", "/api/batch-rewrite/web-submit/sync-styles", "/api/batch-rewrite/web-submit/preview", "/api/batch-rewrite/web-submit/submit":
			return req.Method == http.MethodPost
		}
	}
	return false
}

func isBatchRewriteTaskDetailPath(path string) bool {
	const prefix = "/api/batch-rewrite/tasks/"
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	return strings.Trim(strings.TrimPrefix(path, prefix), "/") != "" && !strings.Contains(strings.TrimPrefix(path, prefix), "/")
}

func isBatchRewriteTaskDiagnosticPath(path string) bool {
	const prefix = "/api/batch-rewrite/tasks/"
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	rest := strings.TrimPrefix(path, prefix)
	for _, suffix := range []string{"/sensitive-log", "/rules-trace", "/site-submit-log"} {
		if strings.HasSuffix(rest, suffix) {
			bookID := strings.TrimSuffix(rest, suffix)
			return bookID != "" && !strings.Contains(bookID, "/")
		}
	}
	return false
}

func isLocalScriptVideoRequest(req *http.Request) bool {
	if req == nil || req.URL == nil {
		return false
	}
	if req.Method == http.MethodGet && strings.HasPrefix(req.URL.Path, "/api/script-video/go-local:") {
		return true
	}
	if req.Method != http.MethodPost || req.URL.Path != "/api/script-video" || req.Body == nil {
		return false
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		req.Body = io.NopCloser(bytes.NewReader(body))
		return false
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	var input struct {
		ModelKey string `json:"modelKey"`
	}
	if json.Unmarshal(body, &input) != nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(input.ModelKey), "local-doubao-executor-video")
}

func goAuthEnabled(nodeAuthTarget string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("QIANTIE_GO_AUTH_ENABLED"))) {
	case "1", "true":
		return true
	case "0", "false":
		return strings.TrimSpace(nodeAuthTarget) == ""
	default:
		return strings.TrimSpace(nodeAuthTarget) == ""
	}
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

func shouldBridgeLegacyAPI(path string) bool {
	return shouldBridgeLegacyAPIForSlice(path, -1)
}

// Read-only H3 traces are safe to serve from the Go store while the
// provider-backed H3 writes continue using the legacy bridge.
func shouldBridgeLegacyAPIRequest(req *http.Request, slice int) bool {
	if req != nil && req.URL != nil && req.Method == http.MethodGet &&
		strings.HasPrefix(req.URL.Path, "/api/batch-factory/v12/") &&
		strings.HasSuffix(req.URL.Path, "/h3/trace") {
		return false
	}
	if req == nil || req.URL == nil {
		return false
	}
	return shouldBridgeLegacyAPIForSlice(req.URL.Path, slice)
}

// shouldBridgeLegacyAPIForSlice keeps staged V12 routes honest. A route is
// only kept on Go when the corresponding slice is actually enabled; otherwise
// the existing Node implementation remains the functional fallback instead
// of returning a misleading Go 404.
func shouldBridgeLegacyAPIForSlice(path string, slice int) bool {
	// These catalog endpoints are now implemented by the Go service. The
	// remaining /api/config children below are still audited independently.
	if path == "/api/models" || strings.HasPrefix(path, "/api/models/") || path == "/api/config/models" || strings.HasPrefix(path, "/api/config/models/") {
		return false
	}
	if path == "/api/script/smart-unified-style" {
		return false
	}
	if path == "/api/chat" {
		return false
	}
	// The quick director contract is now implemented by the Go prompt/chat
	// pipeline too. Keep this exact legacy URL on Go so older clients cannot
	// silently jump back to the retired Node process.
	if path == "/api/script/director-pipeline" {
		return false
	}
	// These account-recovery actions are registered by the Go security
	// handlers. The destructive purge action is intentionally not exposed as a
	// Node fallback until its account-store parity is implemented; all
	// user-facing recovery flows stay on the single Go runtime.
	for _, exact := range []string{
		"/api/account-recovery/email/status",
		"/api/account-recovery/email/request",
		"/api/account-recovery/email/verify",
		"/api/account-recovery/password/request",
		"/api/account-recovery/password/reset",
		"/api/account-recovery/passkeys",
		"/api/account-recovery/passkeys/options",
	} {
		if path == exact {
			return false
		}
	}
	if strings.HasPrefix(path, "/api/account-recovery/passkeys/") {
		return false
	}
	// The owner-scoped effective configuration is served by Go. Keep the
	// legacy /api/config URL on that handler; only its still-unmigrated child
	// settings endpoints may remain on the compatibility bridge.
	if path == "/api/config" {
		return false
	}
	if path == "/api/config/models" || strings.HasPrefix(path, "/api/config/models/") {
		return false
	}
	if path == "/api/test" || strings.HasPrefix(path, "/api/test/") {
		return false
	}
	if path == "/api/client-errors" || strings.HasPrefix(path, "/api/client-errors/") {
		return false
	}
	if path == "/api/admin" || strings.HasPrefix(path, "/api/admin/") {
		return false
	}
	if path == "/api/presets" || strings.HasPrefix(path, "/api/presets/") {
		return false
	}
	if path == "/api/admin/presets" || strings.HasPrefix(path, "/api/admin/presets/") {
		return false
	}
	if path == "/api/script-constraint-prompts" || strings.HasPrefix(path, "/api/script-constraint-prompts/") {
		return false
	}
	for _, prefix := range []string{"/api/novel-fetch", "/api/novel-fetch-upload", "/api/batch-rewrite", "/api/web-submit"} {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return false
		}
	}
	for _, prefix := range []string{"/api/platform-projects", "/api/storage", "/api/novel-panel", "/api/member", "/api/account-admin", "/api/team-admin", "/api/agent"} {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return false
		}
	}
	// This prefix already ends with a slash; do not add a second slash when
	// matching script endpoints such as /api/script/director-pipeline.
	if strings.HasPrefix(path, "/api/script/") {
		return true
	}
	// V12 is the public workbench contract. All current V12 routes are handled
	// by Go below (including the former summary/direct-fetch adapters). Keep the
	// compatibility check explicit so a new or misspelled V12 path cannot
	// silently resurrect the retired Node service.
	if strings.HasPrefix(path, "/api/batch-factory/v12/") {
		// This legacy repair endpoint was removed from the Node router. Do not
		// resurrect it through the broad V12 compatibility fallback.
		if strings.HasSuffix(path, "/automation/repair-legacy-overrides") {
			return false
		}
		// Batch Factory's final 121 upload is implemented by the Go publisher;
		// keep it on the native router instead of sending it to the retired Node
		// compatibility response.
		if strings.HasSuffix(path, "/publish-121") {
			return false
		}
		// V12 automation start is owned by the native Go coordinator. Keep the
		// legacy Node controller out of this request so one click cannot create
		// two different jobs or silently fall back to the old runtime.
		if strings.HasSuffix(path, "/automation/start") || strings.HasSuffix(path, "/automation/cancel") || strings.HasSuffix(path, "/automation/pause") || strings.HasSuffix(path, "/automation/resume") || strings.HasSuffix(path, "/automation/retry") {
			return false
		}
		if path == "/api/batch-factory/v12/capabilities" || path == "/api/batch-factory/v12/batches" || path == "/api/batch-factory/v12/batches/summary" || path == "/api/batch-factory/v12/fetch-originals" || path == "/api/batch-factory/v12/image-generation" || strings.HasPrefix(path, "/api/batch-factory/v12/intakes/") || (strings.Contains(path, "/h3/") && (slice < 0 || slice >= 3)) || isGoV12AliasPath(path) {
			return false
		}
		if strings.Contains(path, "/h3/") {
			return true
		}
		// Unknown V12 endpoints must fail on Go rather than being forwarded to
		// Node. This makes the remaining migration gap visible and prevents a
		// second runtime from handling a new batch-factory request by accident.
		return false
	}
	// Every known non-provider API namespace is now routed through Go above.
	// Unknown paths must stay on Go and return an explicit 404/405 instead of
	// reviving the old Node service through a broad namespace fallback.
	return false
}

func isGoV12AliasPath(path string) bool {
	if !strings.HasPrefix(path, "/api/batch-factory/v12/") {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(path, "/api/batch-factory/v12/"), "/")
	if len(parts) == 2 && parts[0] == "publish-121" && parts[1] == "organizations" {
		return true
	}
	if len(parts) >= 2 && parts[0] == "intakes" && len(parts) == 3 && parts[2] == "batches" {
		return true
	}
	if len(parts) == 2 && parts[0] == "intakes" && parts[1] == "manual" {
		return true
	}
	if len(parts) == 2 && parts[0] == "intakes" && parts[1] != "" {
		// The V12 intake detail route is registered directly by the Go batch
		// store. Keeping it on the compatibility bridge caused the novel-fetch
		// handoff to switch runtimes between creation and detail loading.
		return true
	}
	if len(parts) == 5 && parts[0] == "batches" && parts[2] == "intakes" && parts[4] == "books" {
		return true
	}
	if len(parts) == 2 && parts[0] == "batches" {
		return true
	}
	if len(parts) == 3 && parts[0] == "batches" && (parts[2] == "runtime-index" || parts[2] == "settings") {
		return true
	}
	if len(parts) == 3 && parts[0] == "batches" && parts[2] == "status" {
		return true
	}
	if len(parts) == 3 && parts[0] == "batches" && parts[2] == "merge-status" {
		return true
	}
	if len(parts) == 1 && (parts[0] == "config-versions" || parts[0] == "prompts" || parts[0] == "drafts") {
		return true
	}
	if len(parts) == 1 && parts[0] == "automation-presets" {
		return true
	}
	if len(parts) == 1 && parts[0] == "schedules" {
		return true
	}
	if len(parts) == 2 && parts[0] == "schedules" {
		return true
	}
	if len(parts) == 2 && parts[0] == "automation-presets" {
		return true
	}
	if len(parts) == 2 && parts[0] == "video-provider" && parts[1] == "status" {
		return true
	}
	if len(parts) == 3 && parts[0] == "batches" && parts[2] == "automation-presets" {
		return true
	}
	if len(parts) == 3 && parts[0] == "batches" && parts[2] == "automation" {
		return true
	}
	if len(parts) == 4 && parts[0] == "batches" && parts[2] == "automation" && (parts[3] == "start" || parts[3] == "cancel" || parts[3] == "pause" || parts[3] == "resume" || parts[3] == "retry") {
		return true
	}
	if len(parts) == 3 && parts[0] == "batches" && parts[2] == "runtime-summary" {
		return true
	}
	if len(parts) == 3 && parts[0] == "batches" && parts[2] == "change-impact" {
		return true
	}
	if len(parts) == 3 && parts[0] == "batches" && parts[2] == "director" {
		return true
	}
	if len(parts) == 5 && parts[0] == "batches" && parts[2] == "books" && (parts[4] == "director" || parts[4] == "opening-variants") {
		return true
	}
	if len(parts) == 5 && parts[0] == "batches" && parts[2] == "books" && parts[4] == "hook" {
		return true
	}
	if len(parts) == 6 && parts[0] == "batches" && parts[2] == "books" && parts[4] == "smart-unified" && parts[5] == "refresh" {
		return true
	}
	if len(parts) == 6 && parts[0] == "batches" && parts[2] == "books" && parts[4] == "stages" && parts[5] == "retry" {
		return true
	}
	if len(parts) == 6 && parts[0] == "batches" && parts[2] == "books" && parts[4] == "stages" {
		return true
	}
	if len(parts) == 6 && parts[0] == "batches" && parts[2] == "books" && parts[4] == "working-front" && parts[5] == "viral" {
		return true
	}
	if len(parts) == 7 && parts[0] == "batches" && parts[2] == "books" && parts[4] == "hooks" && parts[6] == "approve" {
		return true
	}
	if len(parts) == 5 && parts[0] == "batches" && parts[2] == "books" && parts[4] == "production" {
		return true
	}
	if len(parts) == 3 && parts[0] == "batches" && parts[2] == "production" {
		return true
	}
	if len(parts) == 4 && parts[0] == "batches" && parts[2] == "production" && parts[3] == "cancel" {
		return true
	}
	if len(parts) == 3 && parts[0] == "batches" && parts[2] == "merge" {
		return true
	}
	if len(parts) == 5 && parts[0] == "batches" && parts[2] == "books" && parts[4] == "merge" {
		return true
	}
	if len(parts) == 4 && parts[0] == "batches" && parts[2] == "merge-artifacts" && parts[3] == "migrate-to-tos" {
		return true
	}
	if len(parts) == 4 && parts[0] == "batches" && (parts[2] == "merge-media" || parts[2] == "merge-cover") {
		return true
	}
	if len(parts) == 3 && parts[0] == "batches" && (parts[2] == "classify-fetched-metadata" || parts[2] == "backfill-121-metadata") {
		return true
	}
	if len(parts) >= 3 && parts[0] == "publish" && parts[2] == "credential" {
		return true
	}
	if len(parts) >= 3 && parts[0] == "publish" && parts[2] == "intents" {
		return true
	}
	if len(parts) == 5 && parts[0] == "batches" && parts[2] == "books" && (parts[4] == "metadata" || parts[4] == "source" || parts[4] == "override" || parts[4] == "fetch-original") {
		return true
	}
	if len(parts) == 5 && parts[0] == "batches" && parts[2] == "books" && parts[4] == "stages" {
		return true
	}
	if len(parts) == 7 && parts[0] == "batches" && parts[2] == "books" && parts[4] == "videos" && parts[6] == "override" {
		return true
	}
	if len(parts) == 7 && parts[0] == "batches" && parts[2] == "books" && parts[4] == "videos" && (parts[6] == "effective-settings" || parts[6] == "final-prompt") {
		return true
	}
	if len(parts) == 7 && parts[0] == "batches" && parts[2] == "books" && parts[4] == "videos" && parts[6] == "upload" {
		return true
	}
	if len(parts) == 6 && parts[0] == "batches" && parts[2] == "books" && parts[4] == "merge" && parts[5] == "upload" {
		return true
	}
	if len(parts) == 9 && parts[0] == "batches" && parts[2] == "books" && parts[4] == "videos" && parts[6] == "tasks" {
		return true
	}
	if len(parts) == 2 && parts[0] == "video-provider" && parts[1] == "config" {
		return true
	}
	// Book reference assets are already implemented by the Go V11 store and
	// have identical owner-scoped semantics under the V12 browser contract.
	if len(parts) == 5 && parts[0] == "batches" && parts[2] == "books" && parts[4] == "assets" {
		return true
	}
	if len(parts) == 6 && parts[0] == "batches" && parts[2] == "books" && parts[4] == "assets" {
		return true
	}
	if len(parts) == 7 && parts[0] == "batches" && parts[2] == "books" && parts[4] == "assets" && parts[6] == "images" {
		return true
	}
	if len(parts) == 7 && parts[0] == "batches" && parts[2] == "books" && parts[4] == "assets" && parts[5] == "images" && parts[6] == "generate" {
		return true
	}
	if len(parts) == 8 && parts[0] == "batches" && parts[2] == "books" && parts[4] == "assets" && parts[6] == "images" {
		return true
	}
	if len(parts) == 9 && parts[0] == "batches" && parts[2] == "books" && parts[4] == "assets" && parts[6] == "images" && (parts[8] == "primary" || parts[8] == "content") {
		return true
	}
	return false
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
