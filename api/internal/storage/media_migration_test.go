package storage

import (
	"os"
	"strings"
	"testing"
)

func TestMediaMigrationDefinesCanonicalTOSAssets(t *testing.T) {
	body, err := os.ReadFile("../../migrations/00004_media_assets.sql")
	if err != nil { t.Fatal(err) }
	sql := string(body)
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS media_assets",
		"owner",
		"media_type",
		"tos_bucket",
		"tos_key",
		"mime_type",
		"size_bytes",
		"width_px",
		"height_px",
		"duration_ms",
		"source_task_id",
		"metadata_json",
	} {
		if !strings.Contains(sql, required) { t.Fatalf("migration missing %s", required) }
	}
	if !strings.Contains(sql, "source_task_id VARCHAR(64)") {
		t.Fatal("source_task_id must match agent_tasks.id VARCHAR(64)")
	}
	if strings.Contains(sql, "local_path") || strings.Contains(sql, "file_path") {
		t.Fatal("canonical media assets must not persist local file paths")
	}
}
