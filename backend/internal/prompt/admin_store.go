package prompt

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// AdminStore is the small, durable writer used by the Go admin console.  It
// deliberately keeps the existing JSON shape so the old prompt catalog and
// the new Go service can share the same source of truth during migration.
type AdminStore struct {
	path string
	mu   sync.Mutex
}

type StoreError struct {
	Code    string
	Message string
}

func (e *StoreError) Error() string { return e.Message }

func NewAdminStore(path string) *AdminStore { return &AdminStore{path: strings.TrimSpace(path)} }

var adminIDPattern = regexp.MustCompile(`^[\p{L}\p{N}][\p{L}\p{N}._-]{0,63}$`)

func (s *AdminStore) List(module string) ([]map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items, err := s.readLocked()
	if err != nil {
		return nil, err
	}
	module = strings.TrimSpace(module)
	if module == "" {
		return nil, &StoreError{Code: "invalid", Message: "module is required"}
	}
	result := make([]map[string]any, 0)
	for _, item := range items {
		if stringValue(item["module"]) == module {
			result = append(result, item)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if stringValue(result[i]["id"]) == stringValue(result[j]["id"]) {
			return intValue(result[i]["version"]) < intValue(result[j]["version"])
		}
		return stringValue(result[i]["id"]) < stringValue(result[j]["id"])
	})
	return result, nil
}

func (s *AdminStore) Get(id string, version int) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items, err := s.readLocked()
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if stringValue(item["id"]) == id && intValue(item["version"]) == version {
			return item, nil
		}
	}
	return nil, &StoreError{Code: "not_found", Message: "提示词版本不存在"}
}

func (s *AdminStore) CreateDraft(actor string, input map[string]any) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := validateActor(actor); err != nil {
		return nil, err
	}
	if err := validateDraftInput(input, false); err != nil {
		return nil, err
	}
	items, err := s.readLocked()
	if err != nil {
		return nil, err
	}
	id := stringValue(input["id"])
	version := 0
	for _, item := range items {
		if stringValue(item["id"]) == id && intValue(item["version"]) > version {
			version = intValue(item["version"])
		}
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	item := map[string]any{"id": id, "module": stringValue(input["module"]), "name": strings.TrimSpace(stringValue(input["name"])), "kind": stringValue(input["kind"]), "description": stringValue(input["description"]), "compatibleBaseIds": cloneValue(input["compatibleBaseIds"]), "version": version + 1, "revision": 1, "status": "draft", "body": cloneValue(input["body"]), "protocolLock": cloneValue(input["protocolLock"]), "createdAt": now, "createdBy": actor, "publishedAt": nil, "publishedBy": nil}
	items = append(items, item)
	if err := s.writeLocked(items); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *AdminStore) UpdateDraft(actor, id string, version int, input map[string]any) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := validateActor(actor); err != nil {
		return nil, err
	}
	if err := validateDraftInput(input, true); err != nil {
		return nil, err
	}
	items, err := s.readLocked()
	if err != nil {
		return nil, err
	}
	for i, item := range items {
		if stringValue(item["id"]) != id || intValue(item["version"]) != version {
			continue
		}
		if stringValue(item["status"]) != "draft" {
			return nil, &StoreError{Code: "conflict", Message: "只有草稿可以编辑"}
		}
		if intValue(input["expectedRevision"]) != intValue(item["revision"]) {
			return nil, &StoreError{Code: "conflict", Message: "草稿版本已变化，请刷新后再保存"}
		}
		for _, key := range []string{"module", "name", "kind", "description", "compatibleBaseIds", "body", "protocolLock"} {
			item[key] = cloneValue(input[key])
		}
		item["name"] = strings.TrimSpace(stringValue(item["name"]))
		item["revision"] = intValue(item["revision"]) + 1
		items[i] = item
		if err := s.writeLocked(items); err != nil {
			return nil, err
		}
		return item, nil
	}
	return nil, &StoreError{Code: "not_found", Message: "草稿不存在"}
}

func (s *AdminStore) Publish(actor, id string, version int) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := validateActor(actor); err != nil {
		return nil, err
	}
	items, err := s.readLocked()
	if err != nil {
		return nil, err
	}
	var target map[string]any
	for _, item := range items {
		if stringValue(item["id"]) == id && intValue(item["version"]) == version {
			target = item
			break
		}
	}
	if target == nil {
		return nil, &StoreError{Code: "not_found", Message: "草稿不存在"}
	}
	if stringValue(target["status"]) != "draft" {
		return nil, &StoreError{Code: "conflict", Message: "只有草稿可以发布"}
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, item := range items {
		if stringValue(item["id"]) == id && stringValue(item["status"]) == "published" {
			item["status"] = "archived"
		}
	}
	target["status"], target["publishedAt"], target["publishedBy"] = "published", now, actor
	if err := s.writeLocked(items); err != nil {
		return nil, err
	}
	return target, nil
}

func (s *AdminStore) Rollback(actor, id string, version int) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := validateActor(actor); err != nil {
		return nil, err
	}
	items, err := s.readLocked()
	if err != nil {
		return nil, err
	}
	var target map[string]any
	for _, item := range items {
		if stringValue(item["id"]) == id && intValue(item["version"]) == version {
			target = item
			break
		}
	}
	if target == nil {
		return nil, &StoreError{Code: "not_found", Message: "要回滚的版本不存在"}
	}
	if stringValue(target["status"]) != "archived" {
		return nil, &StoreError{Code: "conflict", Message: "只有已归档版本可以回滚"}
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, item := range items {
		if stringValue(item["id"]) == id && stringValue(item["status"]) == "published" {
			item["status"] = "archived"
		}
	}
	target["status"], target["publishedAt"], target["publishedBy"] = "published", now, actor
	if err := s.writeLocked(items); err != nil {
		return nil, err
	}
	return target, nil
}

func (s *AdminStore) readLocked() ([]map[string]any, error) {
	if s == nil || s.path == "" {
		return nil, errors.New("preset catalog path is not configured")
	}
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}
	var items []map[string]any
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *AdminStore) writeLocked(items []map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".presets-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err = tmp.Write(append(raw, '\n')); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, s.path)
}

func validateActor(actor string) error {
	if !adminIDPattern.MatchString(actor) {
		return &StoreError{Code: "invalid", Message: "操作人不合法"}
	}
	return nil
}

func validateDraftInput(input map[string]any, update bool) error {
	if input == nil || !adminIDPattern.MatchString(stringValue(input["id"])) || !adminIDPattern.MatchString(stringValue(input["module"])) || strings.TrimSpace(stringValue(input["name"])) == "" || len(stringValue(input["name"])) > 160 || (stringValue(input["kind"]) != "base" && stringValue(input["kind"]) != "addon") || len(stringValue(input["description"])) > 1000 || input["body"] == nil || !jsonValue(input["body"]) || !jsonValue(input["protocolLock"]) {
		return &StoreError{Code: "invalid", Message: "提示词草稿内容不合法"}
	}
	if update && intValue(input["expectedRevision"]) < 1 {
		return &StoreError{Code: "invalid", Message: "缺少草稿版本号"}
	}
	ids, ok := input["compatibleBaseIds"].([]any)
	if !ok {
		if _, ok2 := input["compatibleBaseIds"].([]string); !ok2 {
			return &StoreError{Code: "invalid", Message: "兼容基础提示词格式不合法"}
		}
	} else {
		for _, value := range ids {
			if !adminIDPattern.MatchString(stringValue(value)) {
				return &StoreError{Code: "invalid", Message: "兼容基础提示词格式不合法"}
			}
		}
	}
	return nil
}

func jsonValue(value any) bool {
	if value == nil {
		return true
	}
	_, err := json.Marshal(value)
	return err == nil
}
func cloneValue(value any) any {
	if value == nil {
		return nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var out any
	_ = json.Unmarshal(raw, &out)
	return out
}
func stringValue(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
func intValue(value any) int {
	switch n := value.(type) {
	case int:
		return n
	case float64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	}
	return 0
}
