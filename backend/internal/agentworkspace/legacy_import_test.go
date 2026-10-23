package agentworkspace

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestImportLegacyFilePreservesAndImportsWholeProjectOnce(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	path := filepath.Join(t.TempDir(), "agent-tasks.json")
	raw := []byte(`{"version":1,"tasks":[{"id":"legacy-1","title":"旧项目","titleMode":"manual","preview":"已分析","createdAt":"2026-10-17T01:00:00Z","updatedAt":"2026-10-17T01:02:00Z","messages":[{"role":"user","content":"做这本书","createdAt":"2026-10-17T01:01:00Z"},{"role":"assistant","content":"已分析","createdAt":"2026-10-17T01:02:00Z"}],"canvas":{"version":1,"revision":3,"viewport":{"x":12,"y":18,"zoom":0.8},"nodes":[{"id":"n1","type":"input","position":{"x":20,"y":30},"title":"小说原文","version":1,"sourceNodeIds":[],"status":"confirmed","summary":"第一章"}],"edges":[]},"executions":[{"id":"e1","toolName":"分析小说","status":"succeeded","createdAt":"2026-10-17T01:01:30Z","updatedAt":"2026-10-17T01:01:40Z"}]}]}`)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	wantHash := fmt.Sprintf("%x", sha256.Sum256(raw))
	now := time.Date(2026, 10, 18, 2, 0, 0, 0, time.UTC)

	result, err := ImportLegacyFile(ctx, store, "alice", path, now)
	if err != nil {
		t.Fatal(err)
	}
	if result.AlreadyImported || result.ImportedProjects != 1 || result.SourceSHA256 != wantHash {
		t.Fatalf("result=%+v", result)
	}
	project, err := store.Project(ctx, "alice", "legacy-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(project.Messages) != 2 || project.Canvas.Revision != 3 || len(project.Canvas.Nodes) != 1 || project.Canvas.Nodes[0].Type != "script" || len(project.Executions) != 1 {
		t.Fatalf("project=%+v", project)
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(raw) {
		t.Fatalf("legacy file changed: err=%v", err)
	}

	again, err := ImportLegacyFile(ctx, store, "alice", path, now.Add(time.Hour))
	if err != nil || !again.AlreadyImported || again.ImportedProjects != 1 {
		t.Fatalf("again=%+v err=%v", again, err)
	}
	projects, _ := store.ListProjects(ctx, "alice")
	if len(projects) != 1 {
		t.Fatalf("projects=%+v", projects)
	}
}

func TestImportLegacyFileRecordsMissingSourceAsCompleted(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	path := filepath.Join(t.TempDir(), "missing.json")
	now := time.Date(2026, 10, 18, 2, 0, 0, 0, time.UTC)

	result, err := ImportLegacyFile(ctx, store, "alice", path, now)
	if err != nil || result.ImportedProjects != 0 || result.SourceSHA256 != emptyLegacySHA256 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	again, err := ImportLegacyFile(ctx, store, "alice", path, now.Add(time.Hour))
	if err != nil || !again.AlreadyImported {
		t.Fatalf("again=%+v err=%v", again, err)
	}
}
