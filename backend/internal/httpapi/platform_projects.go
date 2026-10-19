package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"qiantie/backend/internal/batchfactoryv11"
)

func platformLimit(raw string) int {
	limit, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 100
	}
	if limit < 1 {
		return 1
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func platformTime(value any) string {
	text := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(toString(value)), ""))
	if text == "" {
		return ""
	}
	if _, err := time.Parse(time.RFC3339, text); err != nil {
		return ""
	}
	return text
}

func toString(value any) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}

func listNovelPanelHistory(root, username string) []map[string]any {
	projectsDir := filepath.Join(root, username, "novel-panel", "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return []map[string]any{}
	}
	result := make([]map[string]any, 0)
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(projectsDir, entry.Name()))
		if err != nil {
			continue
		}
		var project map[string]any
		if json.Unmarshal(raw, &project) != nil {
			continue
		}
		id := strings.TrimSpace(toString(project["id"]))
		if id == "" {
			id = strings.TrimSuffix(entry.Name(), ".json")
		}
		name := strings.TrimSpace(toString(project["name"]))
		if name == "" {
			name = "未命名小说面板项目"
		}
		created := platformTime(project["created_at"])
		updated := platformTime(project["updated_at"])
		result = append(result, map[string]any{"id": "novel-panel:" + id, "sourceId": id, "kind": "novel-panel", "typeLabel": "小说面板", "name": name, "summary": "小说分析、人物场景与分镜项目", "createdAt": created, "updatedAt": updated, "route": "/novel-panel?project=" + id})
	}
	return result
}

func RegisterPlatformProjectsRoute(root *http.ServeMux, middleware BridgeAuth, history *ScriptHistoryStore, store batchfactoryv11.Store) {
	if root == nil || history == nil || store == nil {
		return
	}
	root.Handle("GET /api/platform-projects", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		entries := make([]map[string]any, 0)
		history.mu.Lock()
		data, _, err := history.load(identity.Username)
		history.mu.Unlock()
		if err == nil {
			if rows, ok := data["entries"].([]any); ok {
				for _, raw := range rows {
					row, ok := raw.(map[string]any)
					if !ok {
						continue
					}
					id := strings.TrimSpace(toString(row["id"]))
					if id == "" {
						continue
					}
					name := strings.TrimSpace(toString(row["formatName"]))
					if name == "" {
						name = strings.TrimSpace(toString(row["format"]))
					}
					if name == "" {
						name = "剧本生成记录"
					}
					entries = append(entries, map[string]any{"id": "script-history:" + id, "sourceId": id, "kind": "script-history", "typeLabel": "剧本生成", "name": name, "summary": toString(row["preview"]), "updatedAt": platformTime(row["createdAt"]), "route": "/history?entry=script-history:" + id})
				}
			}
		}
		entries = append(entries, listNovelPanelHistory(history.Root, identity.Username)...)
		batches, batchErr := store.ListBatches(req.Context(), identity.Username)
		unavailable := []string{}
		if batchErr != nil {
			unavailable = append(unavailable, "shuihuo-production")
		} else {
			for _, batch := range batches {
				created := batch.CreatedAt.UTC().Format(time.RFC3339)
				updated := batch.UpdatedAt.UTC().Format(time.RFC3339)
				entries = append(entries, map[string]any{"id": "shuihuo-production:" + batch.ID, "sourceId": batch.ID, "kind": "shuihuo-production", "typeLabel": "水货生产", "name": batch.Title, "summary": "水货生产项目", "createdAt": created, "updatedAt": updated, "route": "/batch-factory?batch=" + batch.ID})
			}
		}
		sort.SliceStable(entries, func(i, j int) bool {
			return toString(entries[i]["updatedAt"]) > toString(entries[j]["updatedAt"])
		})
		limit := platformLimit(req.URL.Query().Get("limit"))
		if len(entries) > limit {
			entries = entries[:limit]
		}
		writeJSON(w, http.StatusOK, map[string]any{"entries": entries, "unavailableSources": unavailable})
	})))
}
