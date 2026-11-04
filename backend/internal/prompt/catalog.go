package prompt

import (
	"encoding/json"
	"errors"
	"hash/fnv"
	"os"
	"sort"
	"strings"
	"sync"
)

// Catalog reads the published prompt catalog written by the existing system
// store. Go owns the catalog writes; the file-backed data remains a migration
// and read fallback for installations that have not moved to MySQL yet.
type Catalog struct {
	path string

	// The published presets file is several hundred kilobytes and every
	// /api/presets request used to unmarshal it again. Keep the decoded records
	// keyed by the file content hash so an administrator write is picked up on
	// the next read even when the rewrite has the same size and mtime.
	cacheMu     sync.RWMutex
	cache       []presetRecord
	cacheDigest uint64
	cacheValid  bool
}

type presetRecord struct {
	ID                string   `json:"id"`
	Module            string   `json:"module"`
	Name              string   `json:"name"`
	Kind              string   `json:"kind"`
	Description       string   `json:"description"`
	CompatibleBaseIDs []string `json:"compatibleBaseIds"`
	Version           int      `json:"version"`
	Revision          int      `json:"revision"`
	Status            string   `json:"status"`
	Body              string   `json:"body"`
	Categories        []string `json:"categories,omitempty"`
	PreviewAssetID    string   `json:"previewAssetId,omitempty"`
	ProtocolLock      struct {
		Format    string `json:"format"`
		Slot      string `json:"slot"`
		Category  string `json:"category"`
		Operation string `json:"operation"`
	} `json:"protocolLock"`
}

type PublicPreset struct {
	ID                 string   `json:"id"`
	Module             string   `json:"module"`
	Name               string   `json:"name"`
	Kind               string   `json:"kind"`
	Description        string   `json:"description"`
	CompatibleBaseIDs  []string `json:"compatibleBaseIds"`
	Version            int      `json:"version"`
	Revision           int      `json:"revision"`
	Status             string   `json:"status"`
	Slot               *string  `json:"slot"`
	ConstraintCategory string   `json:"constraintCategory,omitempty"`
	ExtractionPreset   bool     `json:"extractionPreset,omitempty"`
	ProcessOperation   string   `json:"processOperation,omitempty"`
	Categories         []string `json:"categories,omitempty"`
	PreviewURL         string   `json:"previewUrl,omitempty"`
	ReferenceURL       string   `json:"referenceUrl,omitempty"`
}

// PublishedPresetBody is the server-side snapshot used when a workbench saves
// a preset selection. The browser receives catalog metadata only; the body is
// resolved here and frozen into the durable job configuration.
type PublishedPresetBody struct {
	ID                 string
	Module             string
	Name               string
	Version            int
	Revision           int
	Body               string
	Slot               string
	ConstraintCategory string
}

func NewCatalog(path string) *Catalog { return &Catalog{path: strings.TrimSpace(path)} }

func (c *Catalog) Path() string {
	if c == nil {
		return ""
	}
	return c.path
}

func (c *Catalog) load() ([]presetRecord, error) {
	if c == nil || c.path == "" {
		return nil, errors.New("preset catalog path is not configured")
	}
	raw, err := os.ReadFile(c.path)
	if err != nil {
		return nil, err
	}
	// 每次读取原始字节并比对内容哈希：mtime 在某些文件系统上的精度不足以
	// 区分同尺寸的快速重写（管理员保存后立即再读），仅凭 size+mtime 会
	// 返回过期缓存。读取几百 KB 很便宜，真正昂贵的是重复 json.Unmarshal，
	// 所以缓存以内容哈希为键，命中时跳过解码。
	sum := fnv.New64a()
	_, _ = sum.Write(raw)
	digest := sum.Sum64()
	c.cacheMu.RLock()
	if c.cacheValid && c.cacheDigest == digest {
		records := c.cache
		c.cacheMu.RUnlock()
		return clonePresetRecords(records), nil
	}
	c.cacheMu.RUnlock()
	var loaded []presetRecord
	if err := json.Unmarshal(raw, &loaded); err != nil {
		return nil, err
	}
	c.cacheMu.Lock()
	c.cache = loaded
	c.cacheDigest = digest
	c.cacheValid = true
	c.cacheMu.Unlock()
	return clonePresetRecords(loaded), nil
}

// clonePresetRecords returns a deep copy of the mutable record fields so a
// caller cannot mutate the shared cache. The structs themselves are small and
// are copied by value; only the slice fields need their own backing arrays.
func clonePresetRecords(records []presetRecord) []presetRecord {
	if records == nil {
		return nil
	}
	out := make([]presetRecord, len(records))
	copy(out, records)
	for i := range out {
		if records[i].CompatibleBaseIDs != nil {
			out[i].CompatibleBaseIDs = append([]string(nil), records[i].CompatibleBaseIDs...)
		}
		if records[i].Categories != nil {
			out[i].Categories = append([]string(nil), records[i].Categories...)
		}
	}
	return out
}

func publicPreset(record presetRecord) PublicPreset {
	revision := record.Revision
	if revision < 1 {
		revision = 1
	}
	var slot *string
	if record.ProtocolLock.Slot != "" {
		value := record.ProtocolLock.Slot
		slot = &value
	}
	result := PublicPreset{ID: record.ID, Module: record.Module, Name: record.Name, Kind: record.Kind, Description: record.Description, CompatibleBaseIDs: append([]string(nil), record.CompatibleBaseIDs...), Version: record.Version, Revision: revision, Status: record.Status, Slot: slot, Categories: append([]string(nil), record.Categories...)}
	if strings.TrimSpace(record.PreviewAssetID) != "" && (record.ProtocolLock.Slot == "batch.comic-style" || record.ProtocolLock.Slot == "batch.comic-reference-scene") {
		result.PreviewURL = "/api/presets/comic-material/" + record.ID + "/preview"
		result.ReferenceURL = "/api/local-executor/v1/batch-comic-materials/" + record.ID
	} else if record.ProtocolLock.Slot == "batch.comic-style" {
		result.PreviewURL = "/api/presets/comic-style/" + record.ID + "/preview"
		result.ReferenceURL = "/api/local-executor/v1/batch-comic-style-previews/" + record.ID
	}
	if record.ProtocolLock.Slot == "batch.comic-reference-scene" {
		if result.PreviewURL == "" {
			result.PreviewURL = "/api/presets/comic-reference-scene/" + record.ID + "/preview"
			result.ReferenceURL = "/api/local-executor/v1/batch-comic-reference-scenes/" + record.ID
		}
	}
	switch record.ProtocolLock.Format {
	case "constraint":
		result.ConstraintCategory = record.ProtocolLock.Category
	case "extract":
		result.ExtractionPreset = true
	case "novel-fetch-process":
		result.ProcessOperation = record.ProtocolLock.Operation
	}
	return result
}

// PublishedComicMaterialAsset returns the privately stored preview image for
// an administrator-managed comic style or scene. Built-in images remain in
// the embedded fallback library and intentionally return no asset ID here.
func (c *Catalog) PublishedComicMaterialAsset(id string) (string, bool, error) {
	record, found, err := c.PublishedByID(id)
	if err != nil || !found {
		return "", false, err
	}
	if record.ProtocolLock.Slot != "batch.comic-style" && record.ProtocolLock.Slot != "batch.comic-reference-scene" {
		return "", false, nil
	}
	assetID := strings.TrimSpace(record.PreviewAssetID)
	return assetID, assetID != "", nil
}

func (c *Catalog) Published(module string) ([]PublicPreset, error) {
	result, err := c.All(module)
	if err != nil {
		return nil, err
	}
	filtered := result[:0]
	for _, item := range result {
		if item.Status == "published" {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}

// All returns every version for an admin view, including drafts and archived
// records. Public callers should continue using Published.
func (c *Catalog) All(module string) ([]PublicPreset, error) {
	records, err := c.load()
	if err != nil {
		return nil, err
	}
	module = strings.TrimSpace(module)
	if module == "" {
		return nil, errors.New("module is required")
	}
	result := make([]PublicPreset, 0)
	for _, record := range records {
		if record.Module == module {
			result = append(result, publicPreset(record))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].ID == result[j].ID {
			return result[i].Version < result[j].Version
		}
		return result[i].ID < result[j].ID
	})
	return result, nil
}

func (c *Catalog) PublishedByID(id string) (presetRecord, bool, error) {
	records, err := c.load()
	if err != nil {
		return presetRecord{}, false, err
	}
	for _, record := range records {
		if record.Status == "published" && record.ID == id {
			return record, true, nil
		}
	}
	return presetRecord{}, false, nil
}

// PublishedBody returns one published preset body for trusted backend callers.
// It must not be used to expose prompt bodies through the public catalog API.
func (c *Catalog) PublishedBody(module, id string) (PublishedPresetBody, bool, error) {
	record, found, err := c.PublishedByID(id)
	if err != nil || !found || record.Module != strings.TrimSpace(module) {
		return PublishedPresetBody{}, false, err
	}
	revision := record.Revision
	if revision < 1 {
		revision = 1
	}
	return PublishedPresetBody{
		ID:                 record.ID,
		Module:             record.Module,
		Name:               record.Name,
		Version:            record.Version,
		Revision:           revision,
		Body:               record.Body,
		Slot:               record.ProtocolLock.Slot,
		ConstraintCategory: record.ProtocolLock.Category,
	}, true, nil
}

func (c *Catalog) ConstraintText(ids []string) (map[string]string, error) {
	texts := make(map[string]string)
	for _, id := range ids {
		record, found, err := c.PublishedByID(id)
		if err != nil {
			return nil, err
		}
		if found && record.Module == "script" && record.Kind == "addon" && record.ProtocolLock.Format == "constraint" {
			texts[id] = record.Body
		}
	}
	return texts, nil
}

func (c *Catalog) Resolve(module string, ids []string) ([]PublicPreset, error) {
	if module == "" || len(ids) == 0 {
		return nil, errors.New("invalid preset selection")
	}
	seen := make(map[string]bool)
	selected := make([]presetRecord, 0, len(ids))
	for _, id := range ids {
		if seen[id] {
			return nil, errors.New("preset selection contains duplicate ids")
		}
		seen[id] = true
		record, found, err := c.PublishedByID(id)
		if err != nil {
			return nil, err
		}
		if !found || record.Module != module {
			return nil, errors.New("preset is not published")
		}
		selected = append(selected, record)
	}
	baseCount := 0
	baseID := ""
	for _, record := range selected {
		if record.Kind == "base" {
			baseCount++
			baseID = record.ID
		}
	}
	if baseCount != 1 {
		return nil, errors.New("preset selection requires one published base")
	}
	for _, record := range selected {
		if record.Kind == "addon" && !contains(record.CompatibleBaseIDs, baseID) {
			return nil, errors.New("addon is not compatible with selected base")
		}
		if record.Kind == "addon" && record.ProtocolLock.Slot != "" {
			for _, base := range selected {
				if base.Kind == "base" && base.ProtocolLock.Slot != "" && record.ProtocolLock.Slot != base.ProtocolLock.Slot {
					return nil, errors.New("addon must use the same ownership slot as selected base")
				}
			}
		}
	}
	result := make([]PublicPreset, 0, len(selected))
	for _, record := range selected {
		if record.Kind == "base" {
			result = append(result, publicPreset(record))
		}
	}
	for _, record := range selected {
		if record.Kind == "addon" {
			result = append(result, publicPreset(record))
		}
	}
	return result, nil
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
