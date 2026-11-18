package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	webSubmitManifestVersion     = 1
	webSubmitManifestStaleSecond = 360
)

// webSubmitManifest 在第一次上传前把提交计划落盘，并逐项记录状态。
// 崩溃后可以从 executing/unknown 区分“已派发但未确认”和“从未派发”，
// 避免重试时重复上传。
type webSubmitManifest struct {
	mu    sync.Mutex
	path  string
	runID string
	data  map[string]any
	index map[string]map[string]any
}

func newWebSubmitRunID() string {
	suffix := make([]byte, 4)
	_, _ = rand.Read(suffix)
	return fmt.Sprintf("submit_%d_%s", time.Now().UnixMilli(), hex.EncodeToString(suffix))
}

func webSubmitManifestKey(id, version string) string {
	return strings.TrimSpace(id) + "\x00" + strings.TrimSpace(version)
}

func webSubmitManifestDir(usersDir, owner string) (string, error) {
	return UserScopedPath(usersDir, owner, "web-submit", "manifests")
}

func newWebSubmitManifest(usersDir, owner, runID, reason string, plan map[string]any) (*webSubmitManifest, error) {
	dir, err := webSubmitManifestDir(usersDir, owner)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	index := map[string]map[string]any{}
	items := []any{}
	add := func(id, version, state, reasonCode, message string) {
		id, version = strings.TrimSpace(id), strings.TrimSpace(version)
		if id == "" || version == "" {
			return
		}
		key := webSubmitManifestKey(id, version)
		row := index[key]
		if row == nil {
			row = map[string]any{"id": id, "version": version}
			index[key] = row
			items = append(items, row)
		}
		row["state"] = state
		if reasonCode != "" {
			row["reason"] = reasonCode
		} else {
			delete(row, "reason")
		}
		if message != "" {
			row["error"] = message
		} else {
			delete(row, "error")
		}
		row["updated_at"] = now
	}
	for _, rawGroup := range arrayOrEmpty(plan["groups"]) {
		group := mapOrEmpty(rawGroup)
		for _, rawItem := range arrayOrEmpty(group["items"]) {
			item := mapOrEmpty(rawItem)
			add(toString(item["id"]), toString(item["version"]), "ready", "", "")
		}
	}
	for _, rawItem := range arrayOrEmpty(plan["skipped"]) {
		item := mapOrEmpty(rawItem)
		status := firstText(item["status"])
		state := "skipped"
		if item["deferred"] == true {
			state = "deferred"
		}
		if status == "failed" || status == "error" {
			state = "failed"
		}
		add(firstText(item["id"]), firstText(item["version"]), state, firstText(item["reason"], status), firstText(item["error"]))
	}
	for _, rawItem := range arrayOrEmpty(plan["missing_items"]) {
		item := mapOrEmpty(rawItem)
		add(firstText(item["id"]), firstText(item["version"]), "failed", "missing", "提交计划未包含该版本")
	}
	manifest := &webSubmitManifest{
		path:  filepath.Join(dir, runID+".json"),
		runID: runID,
		index: index,
		data: map[string]any{
			"manifest_version":  webSubmitManifestVersion,
			"submission_run_id": runID,
			"owner":             owner,
			"reason":            strings.TrimSpace(reason),
			"status":            "planned",
			"requested_count":   len(arrayOrEmpty(plan["requested_items"])),
			"planned_count":     len(arrayOrEmpty(plan["planned_items"])),
			"missing_count":     len(arrayOrEmpty(plan["missing_items"])),
			"deferred_count":    len(arrayOrEmpty(plan["deferred_items"])),
			"items":             items,
			"created_at":        now,
			"updated_at":        now,
		},
	}
	if err := manifest.flush(); err != nil {
		return nil, err
	}
	return manifest, nil
}

func (m *webSubmitManifest) setItem(id, version, state, reasonCode, message string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := webSubmitManifestKey(id, version)
	row := m.index[key]
	if row == nil {
		row = map[string]any{"id": strings.TrimSpace(id), "version": strings.TrimSpace(version)}
		m.index[key] = row
		m.data["items"] = append(arrayOrEmpty(m.data["items"]), row)
	}
	row["state"] = state
	if reasonCode != "" {
		row["reason"] = reasonCode
	}
	if message != "" {
		row["error"] = message
	}
	row["updated_at"] = time.Now().UTC().Format(time.RFC3339Nano)
	_ = m.flush()
}

// finish 写入本轮提交的终态；若进程在上传中途崩溃，文件会停留在 executing。
func (m *webSubmitManifest) finish(status string) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	counts := map[string]int{}
	for _, item := range m.index {
		counts[toString(item["state"])]++
	}
	deferred := 0
	for _, state := range []string{"deferred", "waiting_classify", "waiting_config", "waiting_generate", "waiting_original", "unknown", "executing"} {
		deferred += counts[state]
	}
	m.data["status"] = status
	m.data["submitted_count"] = counts["submitted"]
	m.data["unknown_count"] = counts["unknown"]
	m.data["failed_count"] = counts["failed"]
	m.data["deferred_count"] = deferred
	m.data["updated_at"] = time.Now().UTC().Format(time.RFC3339Nano)
	_ = m.flush()
}

func (m *webSubmitManifest) flush() error {
	return writeWebSubmitManifest(m.path, m.data)
}

func writeWebSubmitManifest(path string, data map[string]any) error {
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	temp := path + ".tmp"
	if err := os.WriteFile(temp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(temp, path)
}

// recoverStaleWebSubmitManifests 在拿到提交锁后调用：把超过 360 秒仍停留在
// planned/executing 的旧清单收尾，executing 项转为 unknown。
// 下一次提交只会做远端回读，不会重复派发。
func recoverStaleWebSubmitManifests(usersDir, owner string) int {
	dir, err := webSubmitManifestDir(usersDir, owner)
	if err != nil {
		return 0
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	now := time.Now().UTC()
	recovered := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		data := map[string]any{}
		if json.Unmarshal(raw, &data) != nil {
			continue
		}
		status := strings.TrimSpace(toString(data["status"]))
		if status != "planned" && status != "executing" {
			continue
		}
		updated, err := time.Parse(time.RFC3339Nano, firstText(data["updated_at"], data["created_at"]))
		if err != nil || now.Sub(updated) < webSubmitManifestStaleSecond*time.Second {
			continue
		}
		unknown := 0
		for _, rawItem := range arrayOrEmpty(data["items"]) {
			item := mapOrEmpty(rawItem)
			if toString(item["state"]) != "executing" {
				continue
			}
			item["state"] = "unknown"
			item["reason"] = "recovered_after_crash"
			item["error"] = "提交进程中断，结果未知；下次提交只做远端核对"
			item["updated_at"] = now.Format(time.RFC3339Nano)
			unknown++
		}
		data["status"] = "recovered"
		data["executing_count"] = 0
		data["unknown_count"] = intValue(data["unknown_count"]) + unknown
		data["updated_at"] = now.Format(time.RFC3339Nano)
		if writeWebSubmitManifest(path, data) == nil {
			recovered++
		}
	}
	return recovered
}

// listWebSubmitManifests 返回最近的提交清单，按更新时间倒序，供历史/排障使用。
func listWebSubmitManifests(usersDir, owner string, limit int) []any {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	dir, err := webSubmitManifestDir(usersDir, owner)
	if err != nil {
		return []any{}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []any{}
	}
	type candidate struct {
		path    string
		updated time.Time
	}
	candidates := make([]candidate, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		candidates = append(candidates, candidate{path: filepath.Join(dir, entry.Name()), updated: info.ModTime()})
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].updated.After(candidates[j].updated) })
	out := make([]any, 0, limit)
	for _, item := range candidates {
		if len(out) >= limit {
			break
		}
		raw, err := os.ReadFile(item.path)
		if err != nil {
			continue
		}
		data := map[string]any{}
		if json.Unmarshal(raw, &data) != nil {
			continue
		}
		out = append(out, data)
	}
	return out
}
