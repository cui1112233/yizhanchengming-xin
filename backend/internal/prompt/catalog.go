package prompt

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"strings"
)

// Catalog reads the published prompt catalog written by the existing system
// store. It is deliberately read-only: Node remains the writer until the
// prompt data has a durable Go migration.
type Catalog struct {
	path string
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
	var records []presetRecord
	if err := json.Unmarshal(raw, &records); err != nil {
		return nil, err
	}
	return records, nil
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
	result := PublicPreset{ID: record.ID, Module: record.Module, Name: record.Name, Kind: record.Kind, Description: record.Description, CompatibleBaseIDs: append([]string(nil), record.CompatibleBaseIDs...), Version: record.Version, Revision: revision, Status: record.Status, Slot: slot}
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
