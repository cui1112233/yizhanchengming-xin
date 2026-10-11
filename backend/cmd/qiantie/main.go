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
		if (req.Method == http.MethodGet || req.Method == http.MethodPost) && strings.HasPrefix(req.URL.Path, "/api/login/invite/") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if !goAuthEnabled() && (req.URL.Path == "/api/login" || strings.HasPrefix(req.URL.Path, "/api/login/")) {
			if req.Method == http.MethodPost && (req.URL.Path == "/api/login/passkey/options" || req.URL.Path == "/api/login/passkey/verify") {
				apiHandler.ServeHTTP(w, req)
				return
			}
			authProxy.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && (req.URL.Path == "/api/login/passkey/options" || req.URL.Path == "/api/login/passkey/verify") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if strings.HasPrefix(req.URL.Path, "/api/login/") && req.URL.Path != "/api/login/logout" && req.URL.Path != "/api/login/session" {
			authProxy.ServeHTTP(w, req)
			return
		}
		if isLocalScriptVideoRequest(req) {
			apiHandler.ServeHTTP(w, req)
			return
		}
		// Go has not implemented the rule-preview transformer yet. Keep the
		// existing Node implementation reachable instead of exposing a 404 on
		// the React novel-fetch workbench.
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
		if req.Method == http.MethodPost && strings.HasPrefix(req.URL.Path, "/api/account-admin/accounts/") && strings.HasSuffix(req.URL.Path, "/transfer") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if (req.Method == http.MethodGet && (req.URL.Path == "/api/build-info" || req.URL.Path == "/api/novel-panel/character-core/health" || req.URL.Path == "/api/novel-panel/clean-core/health")) || isAdminAgentSkillRequest(req) {
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
		if (req.Method == http.MethodGet || req.Method == http.MethodPost) && (req.URL.Path == "/api/batch-rewrite/web-submit/config" || req.URL.Path == "/api/batch-rewrite/web-submit/environment") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && (req.URL.Path == "/api/batch-rewrite/web-submit/test-visible" || req.URL.Path == "/api/batch-rewrite/web-submit/sync-configs" || req.URL.Path == "/api/batch-rewrite/web-submit/sync-styles" || req.URL.Path == "/api/batch-rewrite/web-submit/preview" || req.URL.Path == "/api/batch-rewrite/web-submit/submit") {
			apiHandler.ServeHTTP(w, req)
			return
		}
		if req.Method == http.MethodPost && req.URL.Path == "/api/config/models/test" {
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
		if shouldBridgeLegacyAPIRequest(req, cfg.Slice) {
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

func isGoNovelFetchCompatRequest(req *http.Request) bool {
	if req == nil || req.URL == nil {
		return false
	}
	path := req.URL.Path
	if path == "/api/batch-rewrite/knowledge" || strings.HasPrefix(path, "/api/batch-rewrite/knowledge/") || strings.HasPrefix(path, "/api/batch-rewrite/opening/") || strings.HasPrefix(path, "/api/batch-rewrite/ai/") || strings.HasPrefix(path, "/api/batch-rewrite/process") || path == "/api/batch-rewrite/rules/ai-suggest" {
		return req.Method == http.MethodGet || req.Method == http.MethodPost
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

func goAuthEnabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("QIANTIE_GO_AUTH_ENABLED")), "1") || strings.EqualFold(strings.TrimSpace(os.Getenv("QIANTIE_GO_AUTH_ENABLED")), "true")
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
	// These catalog endpoints are now implemented by the Go service. Keep the
	// broader /api/config bridge for the remaining legacy settings endpoints.
	if path == "/api/models" || strings.HasPrefix(path, "/api/models/") || path == "/api/config/models" || strings.HasPrefix(path, "/api/config/models/") {
		return false
	}
	if path == "/api/script/smart-unified-style" {
		return false
	}
	if path == "/api/chat" {
		return false
	}
	// V12 is the public workbench contract. Its summary, direct-fetch and
	// orchestration adapters still live in Node, while Go owns capabilities,
	// intake handoff, and native H3 routes. Keep only the remaining adapters on
	// the local compatibility bridge so 4001 does not expose false 404s.
	if strings.HasPrefix(path, "/api/batch-factory/v12/") {
		if path == "/api/batch-factory/v12/capabilities" || path == "/api/batch-factory/v12/batches" || path == "/api/batch-factory/v12/batches/summary" || path == "/api/batch-factory/v12/fetch-originals" || strings.HasPrefix(path, "/api/batch-factory/v12/intakes/") || (strings.Contains(path, "/h3/") && (slice < 0 || slice >= 3)) || isGoV12AliasPath(path) {
			return false
		}
		if strings.Contains(path, "/h3/") {
			return true
		}
		return true
	}
	for _, prefix := range []string{
		"/api/config",
		"/api/agent",
		"/api/account-admin",
		"/api/account-recovery",
		"/api/accountAdmin",
		"/api/accountRecovery",
		"/api/admin",
		"/api/batch-rewrite",
		"/api/client-errors",
		"/api/member",
		"/api/modelCatalog",
		"/api/novel-fetch",
		"/api/novel-fetch-upload",
		"/api/novel-panel",
		"/api/platform-projects",
		"/api/script-video",
		"/api/storage",
		"/api/team-admin",
		"/api/teamAdmin",
		"/api/test",
		"/api/models",
		"/api/chat",
		"/api/script/",
	} {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}
	return false
}

func isGoV12AliasPath(path string) bool {
	if !strings.HasPrefix(path, "/api/batch-factory/v12/") {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(path, "/api/batch-factory/v12/"), "/")
	if len(parts) >= 2 && parts[0] == "intakes" && len(parts) == 3 && parts[2] == "batches" {
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
	if len(parts) == 3 && parts[0] == "batches" && parts[2] == "runtime-summary" {
		return true
	}
	if len(parts) == 3 && parts[0] == "batches" && parts[2] == "change-impact" {
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
