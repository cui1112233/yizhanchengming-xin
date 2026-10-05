package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/cui1112233/yizhanchengming-xin/api/internal/authn"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/observability"
	"github.com/cui1112233/yizhanchengming-xin/api/internal/video"
)

type responseStatusWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseStatusWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseStatusWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(body)
}

func (h handler) withObservability(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		requestID, ok := observability.ValidRequestID(r.Header.Values(observability.RequestIDHeader))
		if !ok {
			requestID = observability.NewRequestID()
		}
		ctx := observability.WithRequestID(r.Context(), requestID)
		r = r.WithContext(ctx)
		w.Header().Set(observability.RequestIDHeader, requestID)
		recorder := &responseStatusWriter{ResponseWriter: w}
		logger := h.logger()

		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("http panic recovered",
					"request_id", requestID,
					"subsystem", "http",
					"operation", "serve",
					"safe_error", observability.SanitizeString(toSafeRecovered(recovered)),
				)
				if recorder.status == 0 {
					writeJSON(recorder, http.StatusInternalServerError, map[string]any{
						"code": "INTERNAL_ERROR", "message": "服务暂时不可用", "request_id": requestID,
					})
				}
			}
			status := recorder.status
			if status == 0 {
				status = http.StatusOK
			}
			attrs := []any{
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"duration_ms", time.Since(started).Milliseconds(),
				"remote_class", remoteClass(r.RemoteAddr),
			}
			if userID := observability.UserID(r.Context()); userID > 0 {
				attrs = append(attrs, "user_id", userID)
			}
			logger.Info("http request completed", attrs...)
		}()

		next.ServeHTTP(recorder, r)
	})
}

func (h handler) logger() *slog.Logger {
	if h.deps.Logger != nil {
		return h.deps.Logger
	}
	return observability.NewJSONLogger(io.Discard)
}

func toSafeRecovered(value any) string {
	switch v := value.(type) {
	case error:
		return observability.SafeError(v)
	case string:
		return observability.SanitizeString(v)
	default:
		return "panic"
	}
}

func remoteClass(remote string) string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(remote))
	if err != nil {
		host = strings.TrimSpace(remote)
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return "unknown"
	}
	if ip.IsLoopback() {
		return "loopback"
	}
	if ip.IsPrivate() {
		return "private"
	}
	return "public"
}

func (h handler) readyz(w http.ResponseWriter, r *http.Request) {
	requestID := observability.RequestID(r.Context())
	if !h.deps.AppInitialized {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "not_ready", "code": "APP_NOT_INITIALIZED", "message": "application initialization incomplete", "request_id": requestID,
		})
		return
	}
	if h.deps.Database == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "not_ready", "code": "DATABASE_UNAVAILABLE", "message": "database unavailable", "request_id": requestID,
		})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()
	if err := h.deps.Database.PingContext(ctx); err != nil {
		h.logger().Warn("readiness database check failed",
			"request_id", requestID, "subsystem", "mysql", "operation", "ping", "safe_error", observability.SafeError(err),
		)
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "not_ready", "code": "DATABASE_UNAVAILABLE", "message": "database unavailable", "request_id": requestID,
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ready", "request_id": requestID})
}

type diagnosticsDTO struct {
	RequestID string                       `json:"request_id"`
	APIReady  bool                         `json:"api_ready"`
	Database  databaseDiagnosticsDTO       `json:"database"`
	Runtime   runtimeDiagnosticsDTO        `json:"runtime"`
	Process   observability.ProcessStats   `json:"process"`
	Providers []providerDiagnosticsDTO     `json:"providers"`
	Executors executorFleetDiagnosticsDTO  `json:"executors"`
	Jobs      jobDiagnosticsDTO            `json:"jobs"`
}

type databaseDiagnosticsDTO struct {
	Ready           bool          `json:"ready"`
	OpenConnections int           `json:"open_connections"`
	InUse           int           `json:"in_use"`
	Idle            int           `json:"idle"`
	WaitCount       int64         `json:"wait_count"`
	WaitDuration    time.Duration `json:"wait_duration_ns"`
}

type runtimeDiagnosticsDTO struct {
	Ready  bool   `json:"ready"`
	Status string `json:"status"`
	Redis  string `json:"redis"`
}

type providerDiagnosticsDTO struct {
	Provider           string `json:"provider"`
	Model              string `json:"model"`
	Configured         bool   `json:"configured"`
	Enabled            bool   `json:"enabled"`
	Status             string `json:"status"`
	LatestSafeError    string `json:"latest_safe_error,omitempty"`
	RecentFailureCount int64  `json:"recent_submit_failure_count"`
}

type executorDiagnosticsDTO struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Provider      string    `json:"provider"`
	Model         string    `json:"model"`
	Capabilities  []string  `json:"capabilities"`
	Online        bool      `json:"online"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

type executorFleetDiagnosticsDTO struct {
	OnlineCount  int                      `json:"online_count"`
	OfflineCount int                      `json:"offline_count"`
	Items        []executorDiagnosticsDTO `json:"items"`
}

type jobDiagnosticsDTO struct {
	VideoPending int64 `json:"video_pending"`
	VideoRunning int64 `json:"video_running"`
	VideoFailed  int64 `json:"video_failed"`
	VideoStale   int64 `json:"video_stale"`
	MergePending int64 `json:"merge_pending"`
	MergeRunning int64 `json:"merge_running"`
	MergeFailed  int64 `json:"merge_failed"`
	MergeStale   int64 `json:"merge_stale"`
}

func (h handler) requireOperationsAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.deps.Auth == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期", "request_id": observability.RequestID(r.Context())})
			return
		}
		h.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := authn.CurrentUser(r.Context())
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]any{"code": "AUTH_UNAUTHENTICATED", "message": "登录状态无效或已过期", "request_id": observability.RequestID(r.Context())})
				return
			}
			role := strings.ToLower(strings.TrimSpace(user.Role))
			if role != "admin" && role != "owner" {
				writeJSON(w, http.StatusForbidden, map[string]any{"code": "AUTH_FORBIDDEN", "message": "你没有查看系统诊断的权限", "request_id": observability.RequestID(r.Context())})
				return
			}
			next.ServeHTTP(w, r)
		})).ServeHTTP(w, r)
	})
}

func (h handler) diagnostics(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	requestID := observability.RequestID(r.Context())
	result := diagnosticsDTO{
		RequestID: requestID,
		APIReady: h.deps.AppInitialized && h.deps.Database != nil,
		Runtime: runtimeDiagnosticsDTO{Ready: false, Status: "pending_task9_runtime", Redis: "pending_task9"},
		Process: observability.ReadProcessStats(h.deps.StartedAt, now),
		Providers: []providerDiagnosticsDTO{},
		Executors: executorFleetDiagnosticsDTO{Items: []executorDiagnosticsDTO{}},
	}

	if h.deps.Database != nil {
		stats := h.deps.Database.Stats()
		result.Database = databaseDiagnosticsDTO{
			OpenConnections: stats.OpenConnections,
			InUse: stats.InUse,
			Idle: stats.Idle,
			WaitCount: stats.WaitCount,
			WaitDuration: stats.WaitDuration,
		}
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		result.Database.Ready = h.deps.Database.PingContext(ctx) == nil
		cancel()
		result.APIReady = h.deps.AppInitialized && result.Database.Ready
		result.Jobs = h.collectJobDiagnostics(r.Context(), now)
	}

	if h.deps.VideoConfig != nil {
		pairs := [][2]string{
			{video.ProviderPersonalAPI, video.ModelYD20Mini},
			{video.ProviderYFAISeedance, video.ModelSeedance20Official},
			{video.ProviderAutoDLComfyUI, video.ModelMiniMaxH3Video},
			{video.ProviderDoubaoLocalExecutor, video.ModelDoubaoSeedance},
		}
		for _, pair := range pairs {
			view, err := h.deps.VideoConfig.Status(r.Context(), pair[0], pair[1])
			if err != nil {
				h.logger().Warn("provider diagnostic failed", "request_id", requestID, "subsystem", "video_provider", "operation", "status", "provider", pair[0], "model", pair[1], "safe_error", observability.SafeError(err))
				result.Providers = append(result.Providers, providerDiagnosticsDTO{Provider: pair[0], Model: pair[1], Status: "unavailable", LatestSafeError: "provider diagnostic unavailable"})
				continue
			}
			result.Providers = append(result.Providers, providerDiagnosticsDTO{
				Provider: view.ProviderKey, Model: view.Model, Configured: view.Configured, Enabled: view.Enabled,
				Status: string(view.Status), LatestSafeError: safeProviderDiagnosticError(view.Status, view.Message),
				RecentFailureCount: h.recentProviderFailures(r.Context(), view.ProviderKey, view.Model, now.Add(-time.Hour)),
			})
		}
	}

	if h.deps.VideoLocalExecutor != nil {
		items, err := h.deps.VideoLocalExecutor.List(r.Context())
		if err != nil {
			h.logger().Warn("executor diagnostic failed", "request_id", requestID, "subsystem", "local_executor", "operation", "list", "safe_error", observability.SafeError(err))
		} else {
			for _, item := range items {
				safe := executorDiagnosticsDTO{ID: item.ID, Name: item.Name, Provider: item.ProviderKey, Model: item.Model, Capabilities: append([]string(nil), item.Capabilities...), Online: item.Online, LastHeartbeat: item.LastSeenAt}
				result.Executors.Items = append(result.Executors.Items, safe)
				if item.Online {
					result.Executors.OnlineCount++
				} else {
					result.Executors.OfflineCount++
				}
			}
		}
	}

	writeJSON(w, http.StatusOK, result)
}

func safeProviderDiagnosticError(status video.ProviderAvailability, message string) string {
	if strings.TrimSpace(message) == "" {
		return ""
	}
	switch status {
	case video.ProviderStatusAuthFailed:
		return "provider authentication failed"
	case video.ProviderStatusUnconfigured:
		return "provider unconfigured"
	case video.ProviderStatusUnavailable:
		return "provider unavailable"
	default:
		return "provider diagnostic detail withheld"
	}
}

func (h handler) recentProviderFailures(ctx context.Context, provider, model string, since time.Time) int64 {
	if h.deps.Database == nil {
		return 0
	}
	var count int64
	if err := h.deps.Database.QueryRowContext(ctx, `SELECT COUNT(*) FROM video_production_tasks WHERE provider = ? AND model = ? AND status = 'failed' AND updated_at >= ?`, provider, model, since).Scan(&count); err != nil {
		return 0
	}
	return count
}

func (h handler) collectJobDiagnostics(ctx context.Context, now time.Time) jobDiagnosticsDTO {
	var result jobDiagnosticsDTO
	if h.deps.Database == nil {
		return result
	}
	result.VideoPending = h.countJobs(ctx, `SELECT COUNT(*) FROM video_production_tasks WHERE status = 'queued'`)
	result.VideoRunning = h.countJobs(ctx, `SELECT COUNT(*) FROM video_production_tasks WHERE status = 'running'`)
	result.VideoFailed = h.countJobs(ctx, `SELECT COUNT(*) FROM video_production_tasks WHERE status = 'failed'`)
	result.VideoStale = h.countJobs(ctx, `SELECT COUNT(*) FROM video_production_tasks WHERE status IN ('queued','running') AND updated_at < ?`, now.Add(-30*time.Minute))
	result.MergePending = h.countJobs(ctx, `SELECT COUNT(*) FROM video_merge_jobs WHERE status = 'queued'`)
	result.MergeRunning = h.countJobs(ctx, `SELECT COUNT(*) FROM video_merge_jobs WHERE status = 'running'`)
	result.MergeFailed = h.countJobs(ctx, `SELECT COUNT(*) FROM video_merge_jobs WHERE status = 'failed'`)
	result.MergeStale = h.countJobs(ctx, `SELECT COUNT(*) FROM video_merge_jobs WHERE status IN ('queued','running') AND updated_at < ?`, now.Add(-20*time.Minute))
	return result
}

func (h handler) countJobs(ctx context.Context, query string, args ...any) int64 {
	var count int64
	if h.deps.Database == nil {
		return 0
	}
	if err := h.deps.Database.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0
	}
	return count
}
