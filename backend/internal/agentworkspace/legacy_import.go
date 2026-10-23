package agentworkspace

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

var emptyLegacySHA256 = fmt.Sprintf("%x", sha256.Sum256(nil))

type LegacyImportResult struct {
	AlreadyImported  bool
	ImportedProjects int
	SourceSHA256     string
}

func ImportLegacyFile(ctx context.Context, store Store, owner, path string, now time.Time) (LegacyImportResult, error) {
	owner = strings.TrimSpace(owner)
	path = strings.TrimSpace(path)
	if store == nil || owner == "" || path == "" {
		return LegacyImportResult{}, errors.New("agent legacy import input is incomplete")
	}
	if record, ok, err := store.LegacyImport(ctx, owner, path); err != nil {
		return LegacyImportResult{}, err
	} else if ok {
		return LegacyImportResult{AlreadyImported: true, ImportedProjects: record.ImportedProjects, SourceSHA256: record.SourceSHA256}, nil
	}

	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		raw = nil
	} else if err != nil {
		return LegacyImportResult{}, err
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(raw))
	projects := []Project{}
	if len(raw) > 0 {
		var document struct {
			Tasks []Project `json:"tasks"`
		}
		if err := json.Unmarshal(raw, &document); err != nil {
			return LegacyImportResult{}, fmt.Errorf("decode agent legacy file: %w", err)
		}
		projects = document.Tasks
	}

	for index := range projects {
		project := normalizeLegacyProject(projects[index], now, index)
		if err := store.ImportProject(ctx, owner, project); err != nil && !errors.Is(err, ErrProjectExists) {
			return LegacyImportResult{}, err
		}
	}
	record := LegacyImport{OwnerUsername: owner, SourcePath: path, SourceSHA256: digest, ImportedProjects: len(projects), ImportedAt: now}
	if err := store.RecordLegacyImport(ctx, record); err != nil {
		return LegacyImportResult{}, err
	}
	return LegacyImportResult{ImportedProjects: len(projects), SourceSHA256: digest}, nil
}

func normalizeLegacyProject(project Project, now time.Time, index int) Project {
	project = normalizeProject(project)
	for nodeIndex := range project.Canvas.Nodes {
		if project.Canvas.Nodes[nodeIndex].Type == "input" {
			project.Canvas.Nodes[nodeIndex].Type = "script"
		}
	}
	if project.ID == "" {
		project.ID = fmt.Sprintf("agent-import-%d-%d", now.UnixNano(), index)
	}
	if project.CreatedAt.IsZero() {
		project.CreatedAt = now
	}
	if project.UpdatedAt.IsZero() {
		project.UpdatedAt = project.CreatedAt
	}
	for messageIndex := range project.Messages {
		if project.Messages[messageIndex].CreatedAt.IsZero() {
			project.Messages[messageIndex].CreatedAt = project.CreatedAt.Add(time.Duration(messageIndex) * time.Nanosecond)
		}
	}
	for executionIndex := range project.Executions {
		execution := &project.Executions[executionIndex]
		if execution.ID == "" {
			execution.ID = fmt.Sprintf("%s-import-execution-%d", project.ID, executionIndex)
		}
		if execution.ToolName == "" {
			execution.ToolName = "历史工具"
		}
		if execution.Status == "" {
			execution.Status = "completed"
		}
		if execution.CreatedAt.IsZero() {
			execution.CreatedAt = project.CreatedAt
		}
		if execution.UpdatedAt.IsZero() {
			execution.UpdatedAt = execution.CreatedAt
		}
	}
	return project
}
