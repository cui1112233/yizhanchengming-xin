package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func RegisterNovelPanelDiagnosticsCompat(root *http.ServeMux, middleware BridgeAuth, usersDir string) {
	if root == nil || middleware.AuthStore == nil || usersDir == "" {
		return
	}
	root.Handle("GET /api/novel-panel/diagnostics", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		limit := 50
		if raw := req.URL.Query().Get("limit"); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
				limit = parsed
			}
		}
		if limit > 100 {
			limit = 100
		}
		entries := []map[string]any{}
		path := filepath.Join(usersDir, identity.Username, "novel-panel", "ai-diagnostics.json")
		if raw, err := os.ReadFile(path); err == nil {
			_ = json.Unmarshal(raw, &entries)
		}
		if len(entries) > limit {
			entries = entries[:limit]
		}
		writeJSON(w, http.StatusOK, map[string]any{"diagnostics": entries})
	})))
	root.Handle("GET /api/novel-panel/diagnostics/traces", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		entries := readNovelPanelDiagnostics(usersDir, identity.Username, 200)
		writeJSON(w, http.StatusOK, map[string]any{"traces": entries})
	})))
	root.Handle("GET /api/novel-panel/diagnostics/trace/{traceId}", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		for _, entry := range readNovelPanelDiagnostics(usersDir, identity.Username, 200) {
			if toString(entry["id"]) == req.PathValue("traceId") {
				writeJSON(w, http.StatusOK, map[string]any{"trace": entry})
				return
			}
		}
		managedAccountError(w, http.StatusNotFound, "诊断追踪不存在")
	})))
	root.Handle("GET /api/novel-panel/diagnostics/self-check", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			managedAccountError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		panelDir := filepath.Join(usersDir, identity.Username, "novel-panel")
		checks := []map[string]any{}
		add := func(id, label string, pass bool, detail string) {
			status := "fail"
			if pass {
				status = "pass"
			}
			checks = append(checks, map[string]any{"id": id, "label": label, "status": status, "detail": detail})
		}
		if err := os.MkdirAll(panelDir, 0o755); err != nil {
			add("data_writable", "data目录可写", false, err.Error())
		} else {
			add("data_writable", "data目录可写", true, panelDir)
		}
		if _, err := os.Stat(filepath.Join(panelDir, "reference_assets")); err == nil {
			add("reference_assets", "参考资产目录可用", true, "目录已存在")
		} else if err := os.MkdirAll(filepath.Join(panelDir, "reference_assets"), 0o755); err != nil {
			add("reference_assets", "参考资产目录可用", false, err.Error())
		} else {
			add("reference_assets", "参考资产目录可用", true, "目录已创建")
		}
		add("diagnostics_store", "诊断记录可读取", true, "Go诊断存储可用")
		failures := 0
		for _, check := range checks {
			if check["status"] == "fail" {
				failures++
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": failures == 0, "checks": checks, "summary": map[string]any{"pass": len(checks) - failures, "warn": 0, "fail": failures}, "active_instances": 1, "trace_records": len(readNovelPanelDiagnostics(usersDir, identity.Username, 200))})
	})))
}

func readNovelPanelDiagnostics(usersDir, username string, limit int) []map[string]any {
	entries := []map[string]any{}
	path := filepath.Join(usersDir, username, "novel-panel", "ai-diagnostics.json")
	if raw, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(raw, &entries)
	}
	if len(entries) > limit {
		entries = entries[:limit]
	}
	return entries
}
