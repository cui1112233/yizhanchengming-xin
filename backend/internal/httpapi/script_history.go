package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ScriptHistoryStore keeps the legacy /api/history file contract during the
// migration. It deliberately uses the same data/users/<user>/outputs files as
// Node so old history remains readable from either port.
type ScriptHistoryStore struct {
	Root string
	mu   sync.Mutex
}

func NewScriptHistoryStore(root string) *ScriptHistoryStore {
	return &ScriptHistoryStore{Root: strings.TrimSpace(root)}
}

func (s *ScriptHistoryStore) userDir(username string) (string, error) {
	if s == nil || s.Root == "" {
		return "", fmt.Errorf("script history store is not configured")
	}
	return UserScopedPath(s.Root, username, "outputs")
}

func validScriptHistoryID(id string) bool {
	return id != "" && id != "." && id != ".." && !strings.ContainsAny(id, `/\\`)
}

func (s *ScriptHistoryStore) load(username string) (map[string]any, string, error) {
	dir, err := s.userDir(username)
	if err != nil {
		return nil, "", err
	}
	path := filepath.Join(dir, "index.json")
	data := map[string]any{"entries": []any{}}
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return data, dir, nil
	}
	if err != nil {
		return nil, "", err
	}
	if len(raw) > 0 && json.Unmarshal(raw, &data) != nil {
		return nil, "", fmt.Errorf("invalid history index")
	}
	if _, ok := data["entries"].([]any); !ok {
		data["entries"] = []any{}
	}
	return data, dir, nil
}

func saveScriptHistory(dir string, data map[string]any) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	tmp := filepath.Join(dir, "index.json.tmp")
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dir, "index.json"))
}

func historyEntryMap(value any) map[string]any {
	entry, _ := value.(map[string]any)
	return entry
}

func normalizedVideoTasks(value any) map[string]any {
	result := map[string]any{}
	input, ok := value.(map[string]any)
	if !ok {
		return result
	}
	for key, raw := range input {
		if _, err := fmt.Sscanf(key, "%d", new(int)); err != nil {
			continue
		}
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		taskID, _ := item["taskId"].(string)
		status, _ := item["status"].(string)
		if strings.TrimSpace(taskID) == "" || !map[string]bool{"processing": true, "succeeded": true, "failed": true}[status] {
			continue
		}
		clean := map[string]any{"taskId": strings.TrimSpace(taskID), "status": status}
		for _, field := range []string{"prompt", "error", "videoUrl"} {
			if value, ok := item[field].(string); ok && strings.TrimSpace(value) != "" {
				clean[field] = strings.TrimSpace(value)
			}
		}
		result[key] = clean
	}
	return result
}

func RegisterScriptHistoryRoutes(root *http.ServeMux, auth BridgeAuth, store *ScriptHistoryStore) {
	if root == nil || store == nil {
		return
	}
	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		store.mu.Lock()
		defer store.mu.Unlock()
		data, dir, err := store.load(identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "history unavailable"})
			return
		}
		id := req.PathValue("id")
		if req.Method == http.MethodGet && id == "" {
			writeJSON(w, http.StatusOK, data)
			return
		}
		if req.Method == http.MethodGet {
			if !validScriptHistoryID(id) {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
				return
			}
			entry := findHistoryEntry(data, id)
			if entry == nil {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "record not found"})
				return
			}
			text, readErr := os.ReadFile(filepath.Join(dir, id+".txt"))
			if readErr != nil {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "record not found"})
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = w.Write(stripHistoryHeader(string(text)))
			return
		}
		if req.Method == http.MethodPost && id == "" {
			var input map[string]any
			if json.NewDecoder(http.MaxBytesReader(w, req.Body, 8<<20)).Decode(&input) != nil {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
				return
			}
			value, _ := input["id"].(string)
			output, _ := input["output"].(string)
			if !validScriptHistoryID(value) || strings.TrimSpace(output) == "" {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id and output are required"})
				return
			}
			entry := normalizeHistoryEntry(input, value, output)
			entries := data["entries"].([]any)
			filtered := make([]any, 0, len(entries)+1)
			filtered = append(filtered, entry)
			for _, current := range entries {
				if historyEntryMap(current)["id"] != value {
					filtered = append(filtered, current)
				}
			}
			if len(filtered) > 50 {
				filtered = filtered[:50]
			}
			data["entries"] = filtered
			if err := os.MkdirAll(dir, 0o755); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "save failed"})
				return
			}
			header := fmt.Sprintf("格式：%v\n模式：%v\n时长：%v\n生成时间：%s\n%s\n\n", input["formatName"], input["mode"], input["duration"], time.Now().UTC().Format(time.RFC3339), strings.Repeat("=", 40))
			if err := os.WriteFile(filepath.Join(dir, value+".txt"), []byte(header+output), 0o600); err != nil || saveScriptHistory(dir, data) != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "save failed"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "id": value})
			return
		}
		if req.Method == http.MethodPatch {
			if !validScriptHistoryID(id) {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
				return
			}
			entry := findHistoryEntry(data, id)
			if entry == nil {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "record not found"})
				return
			}
			var patch map[string]any
			if !limitRequestBody(w, req, maxJSONBodyBytes) {
				return
			}
			if err := json.NewDecoder(req.Body).Decode(&patch); err != nil {
				if isBodyTooLarge(err) {
					writePayloadTooLarge(w, maxJSONBodyBytes)
					return
				}
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
				return
			}
			if value, ok := patch["videoTasks"]; ok {
				entry["videoTasks"] = normalizedVideoTasks(value)
			}
			if value, ok := patch["videoTaskHistory"]; ok {
				entry["videoTaskHistory"] = value
			}
			if err := saveScriptHistory(dir, data); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "save failed"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "entry": entry})
			return
		}
		if req.Method == http.MethodDelete {
			if id == "" {
				data["entries"] = []any{}
				_ = saveScriptHistory(dir, data)
				_ = clearHistoryFiles(dir)
				writeJSON(w, http.StatusOK, map[string]any{"ok": true})
				return
			}
			if !validScriptHistoryID(id) {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
				return
			}
			if findHistoryEntry(data, id) == nil {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "record not found"})
				return
			}
			entries := data["entries"].([]any)
			kept := make([]any, 0, len(entries))
			for _, current := range entries {
				if historyEntryMap(current)["id"] != id {
					kept = append(kept, current)
				}
			}
			data["entries"] = kept
			_ = os.Remove(filepath.Join(dir, id+".txt"))
			if err := saveScriptHistory(dir, data); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "delete failed"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	root.Handle("GET /api/history", handler)
	root.Handle("GET /api/history/{id}", handler)
	root.Handle("POST /api/history", handler)
	root.Handle("PATCH /api/history/{id}", handler)
	root.Handle("PATCH /api/history/{id}/note", handler)
	root.Handle("PUT /api/history/{id}", handler)
	root.Handle("DELETE /api/history", handler)
	root.Handle("DELETE /api/history/{id}", handler)
}

func findHistoryEntry(data map[string]any, id string) map[string]any {
	for _, value := range data["entries"].([]any) {
		entry := historyEntryMap(value)
		if entry["id"] == id {
			return entry
		}
	}
	return nil
}
func stripHistoryHeader(value string) []byte {
	marker := "\n" + strings.Repeat("=", 40) + "\n\n"
	if index := strings.Index(value, marker); index >= 0 {
		return []byte(value[index+len(marker):])
	}
	return []byte(value)
}
func clearHistoryFiles(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".txt") {
			_ = os.Remove(filepath.Join(dir, entry.Name()))
		}
	}
	return nil
}
func normalizeHistoryEntry(input map[string]any, id, output string) map[string]any {
	preview := strings.ReplaceAll(output, "\n", " ")
	if len(preview) > 40 {
		preview = preview[:40]
	}
	entry := map[string]any{"id": id, "filename": id + ".txt", "format": input["format"], "formatName": input["formatName"], "mode": input["mode"], "duration": input["duration"], "preview": preview, "output": output, "novelText": input["novelText"], "extractInfo": input["extractInfo"], "constraints": input["constraints"], "textModelId": input["textModelId"], "imageModelId": input["imageModelId"], "videoTasks": normalizedVideoTasks(input["videoTasks"]), "videoTaskHistory": input["videoTaskHistory"], "createdAt": time.Now().UTC().Format(time.RFC3339)}
	return entry
}
