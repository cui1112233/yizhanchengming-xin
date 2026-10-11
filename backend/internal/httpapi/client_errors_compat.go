package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var clientErrorLogMu sync.Mutex

func RegisterClientErrorCompat(root *http.ServeMux, middleware BridgeAuth, profiles *MemberProfileStore) {
	if root == nil || profiles == nil || strings.TrimSpace(profiles.SystemDir) == "" {
		return
	}
	root.Handle("POST /api/client-errors", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input struct {
			Kind    string `json:"kind"`
			Message string `json:"message"`
			Stack   string `json:"stack"`
			Source  string `json:"source"`
			Path    string `json:"path"`
			Method  string `json:"method"`
			Status  int    `json:"status"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, req.Body, 32<<10)).Decode(&input) != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "错误日志格式无效"})
			return
		}
		entry := map[string]any{
			"id": strconv.FormatInt(time.Now().UnixNano(), 10), "at": time.Now().UTC().Format(time.RFC3339Nano),
			"kind": clipClientError(input.Kind, 120), "message": clipClientError(input.Message, 4000),
			"stack": clipClientError(input.Stack, 4000), "source": clipClientError(input.Source, 1000),
			"path": clipClientError(input.Path, 1000), "method": clipClientError(input.Method, 32),
			"status": input.Status, "username": identity.Username,
		}
		if err := appendClientError(profiles.SystemDir, entry); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "错误日志保存失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})))
	root.Handle("GET /api/client-errors/mine", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		entries := readClientErrors(profiles.SystemDir)
		mine := make([]map[string]any, 0)
		for _, entry := range entries {
			if strings.TrimSpace(toString(entry["username"])) == identity.Username {
				mine = append(mine, entry)
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"entries": limitClientErrors(mine, req.URL.Query().Get("limit"))})
	})))
}

func clipClientError(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) > limit {
		return value[:limit]
	}
	return value
}

func clientErrorPath(systemDir string) string { return filepath.Join(systemDir, "error-logs.json") }

func readClientErrors(systemDir string) []map[string]any {
	clientErrorLogMu.Lock()
	defer clientErrorLogMu.Unlock()
	raw, err := os.ReadFile(clientErrorPath(systemDir))
	if err != nil {
		return []map[string]any{}
	}
	var entries []map[string]any
	if json.Unmarshal(raw, &entries) != nil {
		return []map[string]any{}
	}
	return entries
}

func appendClientError(systemDir string, entry map[string]any) error {
	clientErrorLogMu.Lock()
	defer clientErrorLogMu.Unlock()
	entries := []map[string]any{}
	if raw, err := os.ReadFile(clientErrorPath(systemDir)); err == nil {
		_ = json.Unmarshal(raw, &entries)
	}
	entries = append(entries, entry)
	if len(entries) > 500 {
		entries = entries[len(entries)-500:]
	}
	if err := os.MkdirAll(systemDir, 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	tmp := clientErrorPath(systemDir) + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, clientErrorPath(systemDir))
}

func limitClientErrors(entries []map[string]any, rawLimit string) []map[string]any {
	limit := 100
	if value, err := strconv.Atoi(rawLimit); err == nil && value > 0 {
		limit = value
	}
	if limit > 500 {
		limit = 500
	}
	if len(entries) > limit {
		return entries[len(entries)-limit:]
	}
	return entries
}
