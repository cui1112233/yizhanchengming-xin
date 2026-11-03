package httpapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const defaultN8WorkTitleMapPath = "data/n8-work-title-map.json"

// n8WorkTitleMap 缓存已经确认过的“书名 -> N8 作品 ID”。
// 一本书只要成功匹配过一次，之后无论 N8 如何调整作品目录接口，
// 发布都直接命中本地记录，不再依赖外部服务。
type n8WorkTitleMap struct {
	mu      sync.Mutex
	path    string
	entries map[string][]int64
	loaded  bool
}

type n8WorkTitleMapFile struct {
	Version int                `json:"version"`
	Titles  map[string][]int64 `json:"titles"`
}

func newN8WorkTitleMap(path string) *n8WorkTitleMap {
	path = strings.TrimSpace(path)
	if path == "" {
		path = defaultN8WorkTitleMapPath
	}
	return &n8WorkTitleMap{path: path, entries: make(map[string][]int64)}
}

// lookup 按规范化后的书名读取已缓存的 N8 作品 ID。
func (m *n8WorkTitleMap) lookup(normalizedTitle string) []int64 {
	if m == nil || normalizedTitle == "" {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureLoaded()
	ids, ok := m.entries[normalizedTitle]
	if !ok || len(ids) == 0 {
		return nil
	}
	return append([]int64(nil), ids...)
}

// store 写入匹配结果。Windows 上 os.Rename 会覆盖目标文件，
// 因此先写临时文件再改名可以避免写到一半的文件被读到。
func (m *n8WorkTitleMap) store(normalizedTitle string, ids []int64) {
	if m == nil || normalizedTitle == "" || len(ids) == 0 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensureLoaded()
	cleaned := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		cleaned = append(cleaned, id)
	}
	if len(cleaned) == 0 {
		return
	}
	sort.Slice(cleaned, func(i, j int) bool { return cleaned[i] < cleaned[j] })
	if existing, ok := m.entries[normalizedTitle]; ok && sameInt64Slice(existing, cleaned) {
		return
	}
	m.entries[normalizedTitle] = cleaned
	m.persist()
}

func (m *n8WorkTitleMap) ensureLoaded() {
	if m.loaded {
		return
	}
	m.loaded = true
	data, err := os.ReadFile(m.path)
	if err != nil {
		return
	}
	var file n8WorkTitleMapFile
	if err := json.Unmarshal(data, &file); err != nil {
		return
	}
	for key, ids := range file.Titles {
		key = strings.TrimSpace(key)
		if key == "" || len(ids) == 0 {
			continue
		}
		m.entries[key] = append([]int64(nil), ids...)
	}
}

func (m *n8WorkTitleMap) persist() {
	file := n8WorkTitleMapFile{Version: 1, Titles: m.entries}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return
	}
	if dir := filepath.Dir(m.path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return
		}
	}
	tmp := m.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return
	}
	_ = os.Rename(tmp, m.path)
}

func sameInt64Slice(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
