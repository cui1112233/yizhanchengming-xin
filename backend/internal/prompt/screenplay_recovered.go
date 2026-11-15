package prompt

import (
	"crypto/sha256"
	"embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CreativePromptContract records the output contract actually stated by each
// recovered body. Model capabilities never contain private provider/model IDs.
type CreativePromptContract struct {
	SourceID          int      `json:"source_id"`
	Purpose           string   `json:"purpose"`
	NormalizedDigest  string   `json:"normalized_digest"`
	Source            string   `json:"source"`
	InputSource       string   `json:"input_source"`
	Variables         []string `json:"variables"`
	OutputProtocol    string   `json:"output_protocol"`
	ResponseMode      string   `json:"response_mode"`
	ParserVersion     string   `json:"parser_version"`
	ModelCapabilities []string `json:"model_capabilities"`
	Supported         bool     `json:"supported"`
	UnsupportedReason string   `json:"unsupported_reason,omitempty"`
	RequiredFields    []string `json:"required_fields"`
	ReportMarkers     []string `json:"report_markers,omitempty"`
	MinDurationSec    float64  `json:"min_duration_sec,omitempty"`
	MaxDurationSec    float64  `json:"max_duration_sec,omitempty"`
	FixedDurationSec  float64  `json:"fixed_duration_sec,omitempty"`
}

//go:embed screenplay_recovered_catalog.json screenplay_recovered_dates.json recovered/prompt-*.txt
var recoveredCreativeFiles embed.FS

type recoveredCreativeRecord struct {
	CreativePromptMetadata
	File string `json:"file"`
}

var recoveredCreativeOnce sync.Once
var recoveredCreativeRecords []recoveredCreativeRecord
var recoveredCreativeErr error

func recoveredCreativeEntries() ([]recoveredCreativeRecord, error) {
	recoveredCreativeOnce.Do(func() {
		data, err := recoveredCreativeFiles.ReadFile("screenplay_recovered_catalog.json")
		if err != nil {
			recoveredCreativeErr = err
			return
		}
		if err = json.Unmarshal(data, &recoveredCreativeRecords); err != nil {
			recoveredCreativeErr = err
			return
		}
		dateData, e := recoveredCreativeFiles.ReadFile("screenplay_recovered_dates.json")
		if e != nil {
			recoveredCreativeErr = e
			return
		}
		var dates struct {
			Records []struct {
				SourceID  int    `json:"source_id"`
				CreatedAt string `json:"created_at"`
				UpdatedAt string `json:"updated_at"`
			} `json:"records"`
		}
		if e = json.Unmarshal(dateData, &dates); e != nil {
			recoveredCreativeErr = e
			return
		}
		byID := map[int][2]string{}
		for _, row := range dates.Records {
			_, createdErr := time.Parse("2006-01-02 15:04:05", row.CreatedAt)
			_, updatedErr := time.Parse("2006-01-02 15:04:05", row.UpdatedAt)
			if _, duplicate := byID[row.SourceID]; duplicate || createdErr != nil || updatedErr != nil {
				recoveredCreativeErr = fmt.Errorf("恢复提示词%d日期缺失或重复", row.SourceID)
				return
			}
			byID[row.SourceID] = [2]string{row.CreatedAt, row.UpdatedAt}
		}
		if len(byID) != 63 {
			recoveredCreativeErr = fmt.Errorf("恢复提示词日期%d条，应为63", len(byID))
			return
		}
		for i := range recoveredCreativeRecords {
			row := &recoveredCreativeRecords[i]
			date, ok := byID[row.SourceID]
			if !ok {
				recoveredCreativeErr = fmt.Errorf("恢复提示词%d缺少原日期", row.SourceID)
				return
			}
			row.CreatedAt = date[0]
			row.UpdatedAt = date[1]
			row.CatalogGroup = "original"
		}
		seen := map[int]bool{}
		for _, r := range recoveredCreativeRecords {
			body, e := recoveredCreativeFiles.ReadFile("recovered/" + r.File)
			if e != nil {
				recoveredCreativeErr = e
				return
			}
			if seen[r.SourceID] || strings.TrimSpace(string(body)) == "" || fmt.Sprintf("%x", sha256.Sum256(body)) != r.Digest || fmt.Sprintf("%x", sha256.Sum256([]byte(strings.ReplaceAll(string(body), "\r\n", "\n")))) != r.NormalizedDigest {
				recoveredCreativeErr = fmt.Errorf("恢复提示词%d正文或摘要不一致", r.SourceID)
				return
			}
			seen[r.SourceID] = true
		}
		if len(seen) != 63 {
			recoveredCreativeErr = fmt.Errorf("恢复提示词目录数量%d，应为63", len(seen))
			return
		}
		for i := range recoveredCreativeRecords {
			if err := applyRecoveredProtocolOverrides(&recoveredCreativeRecords[i].CreativePromptMetadata); err != nil {
				recoveredCreativeErr = err
				return
			}
		}
	})
	return recoveredCreativeRecords, recoveredCreativeErr
}

// RecoveredCreativePromptCatalog includes all recovered records, even those
// explicitly unsupported by the current generation flow. It never returns bodies.
func RecoveredCreativePromptCatalog() ([]CreativePromptMetadata, error) {
	records, err := recoveredCreativeEntries()
	if err != nil {
		return nil, err
	}
	out := make([]CreativePromptMetadata, 0, len(records))
	for _, r := range records {
		out = append(out, cloneCreativeMetadata(r.CreativePromptMetadata))
	}
	return out, nil
}

func cloneCreativeContract(c CreativePromptContract) CreativePromptContract {
	c.Variables = append([]string{}, c.Variables...)
	c.ModelCapabilities = append([]string{}, c.ModelCapabilities...)
	c.RequiredFields = append([]string{}, c.RequiredFields...)
	c.ReportMarkers = append([]string{}, c.ReportMarkers...)
	return c
}

func cloneCreativeMetadata(m CreativePromptMetadata) CreativePromptMetadata {
	m.CreativePromptContract = cloneCreativeContract(m.CreativePromptContract)
	return m
}

func recoveredCreativeAlias(id string) string {
	records, err := recoveredCreativeEntries()
	if err == nil {
		for _, r := range records {
			if strconv.Itoa(r.SourceID) == id {
				return r.ID
			}
		}
	}
	return id
}

func resolveRecoveredCreativePrompt(id, purpose string) (CreativePromptSnapshot, bool, error) {
	records, err := recoveredCreativeEntries()
	if err != nil {
		return CreativePromptSnapshot{}, false, err
	}
	for _, r := range records {
		if r.ID != id {
			continue
		}
		if r.Type != purpose {
			return CreativePromptSnapshot{}, true, fmt.Errorf("提示词%s不适用于%s", id, purpose)
		}
		body, err := recoveredCreativeFiles.ReadFile("recovered/" + r.File)
		if err != nil {
			return CreativePromptSnapshot{}, true, err
		}
		return CreativePromptSnapshot{ID: r.ID, Name: r.Name, Type: r.Type, Version: r.Version, Digest: r.Digest, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt, CatalogGroup: r.CatalogGroup, Body: string(body), CreativePromptContract: cloneCreativeContract(r.CreativePromptContract)}, true, nil
	}
	return CreativePromptSnapshot{}, false, nil
}
