package storage

import (
	"strings"
	"testing"
)

func TestAICutMigrationsCreatePersistedEditingDomain(t *testing.T) {
	migrations := AICutMigrations()
	if len(migrations) != 15 {
		t.Fatalf("migrations=%d, want 15", len(migrations))
	}
	joined := strings.Join(append(append([]string{}, migrations[0].SQL...), migrations[1].SQL...), "\n")
	for _, table := range []string{"aicut_projects", "aicut_assets", "aicut_tracks", "aicut_clips", "aicut_asr_jobs"} {
		if !strings.Contains(joined, "CREATE TABLE IF NOT EXISTS "+table) {
			t.Fatalf("missing %s schema", table)
		}
	}
	for _, required := range []string{"owner_username", "revision", "source_start_ms", "source_duration_ms", "deleted_at"} {
		if !strings.Contains(joined, required) {
			t.Fatalf("missing %s column", required)
		}
	}
	featureSQL := strings.Join(migrations[2].SQL, "\n")
	for _, required := range []string{"script_text", "settings_json", "transition_type", "transition_ms", "review_status", "review_reason", "playback_rate"} {
		if !strings.Contains(featureSQL, required) {
			t.Fatalf("missing persisted AI-CUT feature column %s", required)
		}
	}
	audioSQL := strings.Join(migrations[3].SQL, "\n")
	for _, required := range []string{"volume", "fade_in_ms", "fade_out_ms"} {
		if !strings.Contains(audioSQL, required) {
			t.Fatalf("missing persisted AI-CUT audio column %s", required)
		}
	}
	historySQL := strings.Join(migrations[4].SQL, "\n")
	for _, required := range []string{"aicut_timeline_history", "aicut_timeline_history_state", "timeline_json", "cursor_no", "max_no"} {
		if !strings.Contains(historySQL, required) {
			t.Fatalf("missing persisted AI-CUT history field %s", required)
		}
	}
	agentSQL := strings.Join(migrations[5].SQL, "\n")
	for _, required := range []string{"aicut_agent_runs", "aicut_agent_messages", "aicut_agent_events", "aicut_agent_tool_results", "client_request_id", "sequence_no", "cancel_requested_at", "tool_call_id"} {
		if !strings.Contains(agentSQL, required) {
			t.Fatalf("missing persisted AI-CUT Agent field %s", required)
		}
	}
	if effectSQL := strings.Join(migrations[6].SQL, "\n"); !strings.Contains(effectSQL, "visual_effect_json") {
		t.Fatal("missing persisted AI-CUT visual effect field")
	}
	planSQL := strings.Join(migrations[7].SQL, "\n")
	for _, required := range []string{"aicut_edit_plans", "aicut_edit_plan_operations", "client_request_id", "base_revision", "state", "request_json", "summary_json", "warnings_json", "applied_revision"} {
		if !strings.Contains(planSQL, required) {
			t.Fatalf("missing persisted AI-CUT edit plan field %s", required)
		}
	}
	if linkSQL := strings.Join(migrations[8].SQL, "\n"); !strings.Contains(linkSQL, "link_group_id") {
		t.Fatal("missing persisted AI-CUT link group field")
	}
	automationSQL := strings.Join(migrations[9].SQL, "\n")
	for _, required := range []string{"markers_json", "editor_json"} {
		if !strings.Contains(automationSQL, required) {
			t.Fatalf("missing persisted AI-CUT automation field %s", required)
		}
	}
	canvasSQL := strings.Join(migrations[10].SQL, "\n")
	for _, required := range []string{"background_mode", "background_color"} {
		if !strings.Contains(canvasSQL, required) {
			t.Fatalf("missing persisted AI-CUT canvas field %s", required)
		}
	}
	exportSQL := strings.Join(migrations[11].SQL, "\n")
	for _, required := range []string{"aicut_export_jobs", "expected_revision", "state", "progress", "options_json", "project_snapshot_json", "timeline_snapshot_json", "output_path"} {
		if !strings.Contains(exportSQL, required) {
			t.Fatalf("missing persisted AI-CUT export field %s", required)
		}
	}
}

func TestAppMigrationsIncludeAICutSchema(t *testing.T) {
	wantVersion := AICutMigrations()[0].Version
	for _, migration := range AppMigrations() {
		if migration.Version == wantVersion {
			return
		}
	}
	t.Fatalf("AppMigrations missing AI-CUT version %d", wantVersion)
}
