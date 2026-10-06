package shuihuo

import (
	"os"
	"strings"
	"testing"
)

func TestB1MigrationOnlyCreatesProjectAndSegmentFacts(t *testing.T) {
	body, err := os.ReadFile("../../db/migrations/00009_shuihuo_project_core.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := string(body)
	for _, required := range []string{"CREATE TABLE shuihuo_projects", "CREATE TABLE shuihuo_segments", "owner_user_id", "team_id", "segmentation_status", "UNIQUE KEY uq_shuihuo_segments_project_position", "ON DELETE CASCADE"} {
		if !strings.Contains(sql, required) {
			t.Fatalf("migration missing %q", required)
		}
	}
	for _, forbidden := range []string{"shuihuo_assets", "shuihuo_asset_images", "shuihuo_media", "shuihuo_model_catalog", "shuihuo_provider", "shuihuo_credentials"} {
		if strings.Contains(sql, forbidden) {
			t.Fatalf("B1 migration must not create %q", forbidden)
		}
	}
}
