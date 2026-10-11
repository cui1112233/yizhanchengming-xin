package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var scheduleUsernamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{3,32}$`)

type batchFactoryScheduleStore struct {
	root string
	now  func() time.Time
	mu   sync.Mutex
}

func newBatchFactoryScheduleStore(root string, now func() time.Time) *batchFactoryScheduleStore {
	if now == nil {
		now = time.Now
	}
	return &batchFactoryScheduleStore{root: strings.TrimSpace(root), now: now}
}

func (s *batchFactoryScheduleStore) file(username string) (string, error) {
	if s == nil || s.root == "" || !scheduleUsernamePattern.MatchString(username) {
		return "", errors.New("用户名无效")
	}
	return filepath.Join(s.root, username, "batch-factory-v11", "schedules.json"), nil
}

func (s *batchFactoryScheduleStore) read(username string) ([]map[string]any, error) {
	path, err := s.file(username)
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []map[string]any{}, nil
	}
	if err != nil {
		return nil, err
	}
	var schedules []map[string]any
	if err := json.Unmarshal(raw, &schedules); err != nil {
		return nil, err
	}
	if schedules == nil {
		schedules = []map[string]any{}
	}
	sort.SliceStable(schedules, func(i, j int) bool {
		return fmt.Sprint(schedules[i]["runAt"]) < fmt.Sprint(schedules[j]["runAt"])
	})
	return schedules, nil
}

func (s *batchFactoryScheduleStore) write(username string, schedules []map[string]any) error {
	path, err := s.file(username)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(schedules, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(raw, '\n'), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func scheduleID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err == nil {
		return hex.EncodeToString(bytes[:])
	}
	return fmt.Sprintf("schedule-%d", time.Now().UnixNano())
}

func scheduleTime(value any) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(fmt.Sprint(value)))
	return parsed, err == nil
}

func RegisterBatchFactoryScheduleRoutes(root *http.ServeMux, middleware BridgeAuth, usersDir string, now func() time.Time) {
	if root == nil || strings.TrimSpace(usersDir) == "" {
		return
	}
	store := newBatchFactoryScheduleStore(usersDir, now)
	root.Handle("GET /api/batch-factory/v12/schedules", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		store.mu.Lock()
		schedules, err := store.read(identity.Username)
		store.mu.Unlock()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取计划失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"schedules": schedules})
	})))
	root.Handle("POST /api/batch-factory/v12/schedules", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var input map[string]any
		if err := decodeNovelFetchJSON(w, req, &input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		runAt, valid := scheduleTime(input["runAt"])
		if !valid || !runAt.After(store.now()) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "runAt 必须是未来时间"})
			return
		}
		batchID, requestID := strings.TrimSpace(fmt.Sprint(input["batchId"])), strings.TrimSpace(fmt.Sprint(input["requestId"]))
		if batchID == "<nil>" || requestID == "<nil>" || batchID == "" || requestID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "batchId 和 requestId 不能为空"})
			return
		}
		provider := strings.TrimSpace(fmt.Sprint(input["provider"]))
		if provider == "" || provider == "<nil>" {
			provider = "personal_api"
		}
		snapshot, _ := input["inputSnapshot"].(map[string]any)
		if snapshot == nil {
			snapshot = map[string]any{}
		}
		bookID := strings.TrimSpace(fmt.Sprint(input["bookId"]))
		if bookID == "<nil>" {
			bookID = ""
		}
		nowText := store.now().UTC().Format(time.RFC3339Nano)
		item := map[string]any{"id": scheduleID(), "username": identity.Username, "batchId": batchID, "bookId": bookID, "requestId": requestID, "provider": provider, "runAt": runAt.UTC().Format(time.RFC3339Nano), "inputSnapshot": snapshot, "status": "scheduled", "createdAt": nowText, "updatedAt": nowText, "errorMessage": ""}
		store.mu.Lock()
		schedules, err := store.read(identity.Username)
		if err == nil {
			schedules = append(schedules, item)
			err = store.write(identity.Username, schedules)
		}
		store.mu.Unlock()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存计划失败"})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"schedule": item})
	})))
	root.Handle("PATCH /api/batch-factory/v12/schedules/{id}", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var patch map[string]any
		if err := decodeNovelFetchJSON(w, req, &patch); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请求内容无效"})
			return
		}
		store.mu.Lock()
		defer store.mu.Unlock()
		schedules, err := store.read(identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取计划失败"})
			return
		}
		index := -1
		for i, item := range schedules {
			if fmt.Sprint(item["id"]) == req.PathValue("id") {
				index = i
				break
			}
		}
		if index < 0 {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "schedule not found"})
			return
		}
		item := schedules[index]
		if fmt.Sprint(item["status"]) == "scheduled" {
			if raw, exists := patch["runAt"]; exists {
				runAt, valid := scheduleTime(raw)
				if !valid || !runAt.After(store.now()) {
					writeJSON(w, http.StatusBadRequest, map[string]string{"error": "runAt 必须是未来时间"})
					return
				}
				item["runAt"] = runAt.UTC().Format(time.RFC3339Nano)
			}
			item["updatedAt"] = store.now().UTC().Format(time.RFC3339Nano)
			schedules[index] = item
			if err := store.write(identity.Username, schedules); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "保存计划失败"})
				return
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"schedule": item})
	})))
	root.Handle("DELETE /api/batch-factory/v12/schedules/{id}", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		identity, ok := BridgeIdentityFromContext(req.Context())
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		store.mu.Lock()
		defer store.mu.Unlock()
		schedules, err := store.read(identity.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "读取计划失败"})
			return
		}
		index := -1
		for i, item := range schedules {
			if fmt.Sprint(item["id"]) == req.PathValue("id") {
				index = i
				if fmt.Sprint(item["status"]) == "running" {
					writeJSON(w, http.StatusConflict, map[string]string{"error": "running schedule cannot be deleted"})
					return
				}
				break
			}
		}
		if index < 0 {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "schedule not found"})
			return
		}
		schedules = append(schedules[:index], schedules[index+1:]...)
		if err := store.write(identity.Username, schedules); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "删除计划失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": req.PathValue("id")})
	})))
}
