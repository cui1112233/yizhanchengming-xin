package storage

import "qiantie/backend/internal/aicut"

func AICutMigrations() []Migration {
	out := []Migration{{
		Version: 8800018,
		SQL: []string{
			`CREATE TABLE IF NOT EXISTS aicut_projects (
  id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  title VARCHAR(255) NOT NULL,
  aspect_ratio VARCHAR(16) NOT NULL DEFAULT '9:16',
  revision BIGINT NOT NULL DEFAULT 0,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  deleted_at DATETIME(6) NULL,
  PRIMARY KEY (id, owner_username),
  KEY idx_aicut_projects_owner_updated (owner_username, deleted_at, updated_at)
) ENGINE=InnoDB`,
			`CREATE TABLE IF NOT EXISTS aicut_assets (
  id VARCHAR(64) NOT NULL,
  project_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  kind VARCHAR(24) NOT NULL,
  name VARCHAR(255) NOT NULL,
  content_type VARCHAR(128) NOT NULL,
  object_key VARCHAR(1024) NOT NULL,
  byte_size BIGINT NOT NULL DEFAULT 0,
  duration_ms BIGINT NOT NULL DEFAULT 0,
  created_at DATETIME(6) NOT NULL,
  PRIMARY KEY (id, project_id, owner_username),
  KEY idx_aicut_assets_project (project_id, owner_username, created_at),
  CONSTRAINT fk_aicut_assets_project FOREIGN KEY (project_id, owner_username) REFERENCES aicut_projects(id, owner_username) ON DELETE CASCADE
) ENGINE=InnoDB`,
			`CREATE TABLE IF NOT EXISTS aicut_tracks (
  id VARCHAR(64) NOT NULL,
  project_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  kind VARCHAR(24) NOT NULL,
  name VARCHAR(255) NOT NULL,
  sort_index INT NOT NULL DEFAULT 0,
  muted BOOLEAN NOT NULL DEFAULT FALSE,
  locked BOOLEAN NOT NULL DEFAULT FALSE,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  PRIMARY KEY (id, project_id, owner_username),
  KEY idx_aicut_tracks_project (project_id, owner_username, sort_index),
  CONSTRAINT fk_aicut_tracks_project FOREIGN KEY (project_id, owner_username) REFERENCES aicut_projects(id, owner_username) ON DELETE CASCADE
) ENGINE=InnoDB`,
			`CREATE TABLE IF NOT EXISTS aicut_clips (
  id VARCHAR(64) NOT NULL,
  track_id VARCHAR(64) NOT NULL,
  project_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  asset_id VARCHAR(64) NULL,
  kind VARCHAR(24) NOT NULL,
  name VARCHAR(255) NOT NULL,
  start_ms BIGINT NOT NULL,
  duration_ms BIGINT NOT NULL,
  source_start_ms BIGINT NOT NULL DEFAULT 0,
  source_duration_ms BIGINT NOT NULL DEFAULT 0,
  text_content LONGTEXT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  PRIMARY KEY (id, project_id, owner_username),
  KEY idx_aicut_clips_track (track_id, project_id, owner_username, start_ms),
  CONSTRAINT fk_aicut_clips_track FOREIGN KEY (track_id, project_id, owner_username) REFERENCES aicut_tracks(id, project_id, owner_username) ON DELETE CASCADE
) ENGINE=InnoDB`,
		},
	}, {
		Version:          8800019,
		CallbackChecksum: "aicut-asr-jobs-v1",
		SQL: []string{`CREATE TABLE IF NOT EXISTS aicut_asr_jobs (
  id VARCHAR(64) NOT NULL,
  project_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  asset_id VARCHAR(64) NOT NULL,
  model_id VARCHAR(191) NOT NULL,
  language VARCHAR(32) NOT NULL DEFAULT '',
  expected_revision BIGINT NOT NULL,
  state VARCHAR(32) NOT NULL,
  progress INT NOT NULL DEFAULT 0,
  error_message TEXT NULL,
  imported_count INT NOT NULL DEFAULT 0,
  track_id VARCHAR(64) NOT NULL DEFAULT '',
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  PRIMARY KEY (id, project_id, owner_username),
  KEY idx_aicut_asr_claim (state, created_at),
  KEY idx_aicut_asr_project (project_id, owner_username, created_at),
  CONSTRAINT fk_aicut_asr_project FOREIGN KEY (project_id, owner_username) REFERENCES aicut_projects(id, owner_username) ON DELETE CASCADE
) ENGINE=InnoDB`},
	}, {
		Version:          8800020,
		CallbackChecksum: "aicut-editor-features-v1",
		SQL: []string{
			`ALTER TABLE aicut_projects ADD COLUMN script_text LONGTEXT NULL AFTER revision`,
			`ALTER TABLE aicut_tracks ADD COLUMN settings_json JSON NULL AFTER locked`,
			`ALTER TABLE aicut_clips ADD COLUMN transition_type VARCHAR(32) NOT NULL DEFAULT '' AFTER text_content`,
			`ALTER TABLE aicut_clips ADD COLUMN transition_ms BIGINT NOT NULL DEFAULT 0 AFTER transition_type`,
			`ALTER TABLE aicut_clips ADD COLUMN review_status VARCHAR(16) NOT NULL DEFAULT '' AFTER transition_ms`,
			`ALTER TABLE aicut_clips ADD COLUMN review_reason TEXT NULL AFTER review_status`,
			`ALTER TABLE aicut_clips ADD COLUMN playback_rate DECIMAL(6,3) NOT NULL DEFAULT 1.000 AFTER review_reason`,
		},
	}, {
		Version:          8800021,
		CallbackChecksum: "aicut-audio-envelope-v1",
		SQL: []string{
			`ALTER TABLE aicut_clips ADD COLUMN volume DECIMAL(6,3) NOT NULL DEFAULT 1.000 AFTER playback_rate`,
			`ALTER TABLE aicut_clips ADD COLUMN fade_in_ms BIGINT NOT NULL DEFAULT 0 AFTER volume`,
			`ALTER TABLE aicut_clips ADD COLUMN fade_out_ms BIGINT NOT NULL DEFAULT 0 AFTER fade_in_ms`,
		},
	}, {
		Version:          8800022,
		CallbackChecksum: "aicut-timeline-history-v1",
		SQL: []string{
			`CREATE TABLE IF NOT EXISTS aicut_timeline_history (
  project_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  sequence_no BIGINT NOT NULL,
  timeline_json JSON NOT NULL,
  created_at DATETIME(6) NOT NULL,
  PRIMARY KEY (project_id, owner_username, sequence_no),
  CONSTRAINT fk_aicut_timeline_history_project FOREIGN KEY (project_id, owner_username) REFERENCES aicut_projects(id, owner_username) ON DELETE CASCADE
) ENGINE=InnoDB`,
			`CREATE TABLE IF NOT EXISTS aicut_timeline_history_state (
  project_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  cursor_no BIGINT NOT NULL DEFAULT 0,
  max_no BIGINT NOT NULL DEFAULT 0,
  updated_at DATETIME(6) NOT NULL,
  PRIMARY KEY (project_id, owner_username),
  CONSTRAINT fk_aicut_timeline_history_state_project FOREIGN KEY (project_id, owner_username) REFERENCES aicut_projects(id, owner_username) ON DELETE CASCADE
) ENGINE=InnoDB`,
		},
	}, {
		Version:          8800023,
		CallbackChecksum: "aicut-agent-runtime-v1",
		SQL: []string{
			`CREATE TABLE IF NOT EXISTS aicut_agent_runs (
  id VARCHAR(64) NOT NULL,
  client_request_id VARCHAR(128) NOT NULL,
  project_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  instruction TEXT NOT NULL,
  mode VARCHAR(16) NOT NULL,
  model_id VARCHAR(191) NOT NULL DEFAULT '',
  active_skill_id VARCHAR(191) NOT NULL DEFAULT '',
  expected_revision BIGINT NOT NULL,
  state VARCHAR(32) NOT NULL,
  turn_count INT NOT NULL DEFAULT 0,
  error_message TEXT NULL,
  cancel_requested_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uq_aicut_agent_request (owner_username, project_id, client_request_id),
  KEY idx_aicut_agent_claim (state, created_at),
  CONSTRAINT fk_aicut_agent_run_project FOREIGN KEY (project_id, owner_username) REFERENCES aicut_projects(id, owner_username) ON DELETE CASCADE
) ENGINE=InnoDB`,
			`CREATE TABLE IF NOT EXISTS aicut_agent_messages (
  id VARCHAR(64) NOT NULL,
  run_id VARCHAR(64) NOT NULL,
  project_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  role VARCHAR(16) NOT NULL,
  content LONGTEXT NOT NULL,
  tool_name VARCHAR(128) NOT NULL DEFAULT '',
  tool_call_id VARCHAR(128) NOT NULL DEFAULT '',
  created_at DATETIME(6) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_aicut_agent_messages_project (owner_username, project_id, created_at),
  CONSTRAINT fk_aicut_agent_message_run FOREIGN KEY (run_id) REFERENCES aicut_agent_runs(id) ON DELETE CASCADE
) ENGINE=InnoDB`,
			`CREATE TABLE IF NOT EXISTS aicut_agent_events (
  run_id VARCHAR(64) NOT NULL,
  project_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  sequence_no BIGINT NOT NULL,
  event_type VARCHAR(32) NOT NULL,
  message TEXT NULL,
  tool_name VARCHAR(128) NOT NULL DEFAULT '',
  tool_call_id VARCHAR(128) NOT NULL DEFAULT '',
  revision BIGINT NOT NULL DEFAULT 0,
  created_at DATETIME(6) NOT NULL,
  PRIMARY KEY (run_id, sequence_no),
  KEY idx_aicut_agent_events_project (owner_username, project_id, run_id, sequence_no),
  CONSTRAINT fk_aicut_agent_event_run FOREIGN KEY (run_id) REFERENCES aicut_agent_runs(id) ON DELETE CASCADE
) ENGINE=InnoDB`,
			`CREATE TABLE IF NOT EXISTS aicut_agent_tool_results (
  run_id VARCHAR(64) NOT NULL,
  project_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  tool_call_id VARCHAR(128) NOT NULL,
  tool_name VARCHAR(128) NOT NULL,
  ok BOOLEAN NOT NULL,
  message TEXT NOT NULL,
  revision BIGINT NOT NULL DEFAULT 0,
  data_json JSON NULL,
  created_at DATETIME(6) NOT NULL,
  PRIMARY KEY (run_id, tool_call_id),
  KEY idx_aicut_agent_tool_project (owner_username, project_id, run_id),
  CONSTRAINT fk_aicut_agent_tool_run FOREIGN KEY (run_id) REFERENCES aicut_agent_runs(id) ON DELETE CASCADE
) ENGINE=InnoDB`,
		},
	}, {
		Version:          8800024,
		CallbackChecksum: "aicut-visual-effect-v1",
		SQL: []string{
			`ALTER TABLE aicut_clips ADD COLUMN visual_effect_json JSON NULL AFTER fade_out_ms`,
		},
	}, {
		Version:          8800025,
		CallbackChecksum: "aicut-edit-plans-v1",
		SQL: []string{
			`CREATE TABLE IF NOT EXISTS aicut_edit_plans (
  id VARCHAR(64) NOT NULL,
  client_request_id VARCHAR(128) NOT NULL,
  project_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  base_revision BIGINT NOT NULL,
  model_id VARCHAR(191) NOT NULL,
  state VARCHAR(32) NOT NULL,
  request_json JSON NOT NULL,
  summary_json JSON NOT NULL,
  warnings_json JSON NOT NULL,
  error_message TEXT NULL,
  applied_revision BIGINT NOT NULL DEFAULT 0,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uq_aicut_edit_plan_request (owner_username, project_id, client_request_id),
  KEY idx_aicut_edit_plan_project (owner_username, project_id, created_at),
  CONSTRAINT fk_aicut_edit_plan_project FOREIGN KEY (project_id, owner_username) REFERENCES aicut_projects(id, owner_username) ON DELETE CASCADE
) ENGINE=InnoDB`,
			`CREATE TABLE IF NOT EXISTS aicut_edit_plan_operations (
  plan_id VARCHAR(64) NOT NULL,
  operation_id VARCHAR(96) NOT NULL,
  ordinal_no INT NOT NULL,
  operation_json JSON NOT NULL,
  PRIMARY KEY (plan_id, operation_id),
  KEY idx_aicut_edit_plan_operation_order (plan_id, ordinal_no),
  CONSTRAINT fk_aicut_edit_plan_operation_plan FOREIGN KEY (plan_id) REFERENCES aicut_edit_plans(id) ON DELETE CASCADE
) ENGINE=InnoDB`,
		},
	}, {
		Version:          8800026,
		CallbackChecksum: "aicut-link-groups-v1",
		SQL: []string{
			`ALTER TABLE aicut_clips ADD COLUMN link_group_id VARCHAR(64) NOT NULL DEFAULT '' AFTER review_reason`,
		},
	}, {
		Version:          8800027,
		CallbackChecksum: "aicut-markers-keyframes-v1",
		SQL: []string{
			`ALTER TABLE aicut_projects ADD COLUMN markers_json JSON NULL AFTER script_text`,
			`ALTER TABLE aicut_clips ADD COLUMN editor_json JSON NULL AFTER visual_effect_json`,
		},
	}, {
		Version:          8800028,
		CallbackChecksum: "aicut-project-canvas-background-v1",
		SQL: []string{
			`ALTER TABLE aicut_projects ADD COLUMN background_mode VARCHAR(16) NOT NULL DEFAULT 'color' AFTER aspect_ratio, ADD COLUMN background_color VARCHAR(7) NOT NULL DEFAULT '#000000' AFTER background_mode`,
		},
	}, {
		Version:          8800029,
		CallbackChecksum: "aicut-export-jobs-v1",
		SQL: []string{`CREATE TABLE IF NOT EXISTS aicut_export_jobs (
  id VARCHAR(64) NOT NULL,
  project_id VARCHAR(64) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  format VARCHAR(16) NOT NULL,
  expected_revision BIGINT NOT NULL,
  state VARCHAR(32) NOT NULL,
  progress INT NOT NULL DEFAULT 0,
  error_message TEXT NULL,
  output_path VARCHAR(2048) NOT NULL DEFAULT '',
  options_json JSON NOT NULL,
  project_snapshot_json JSON NOT NULL,
  timeline_snapshot_json JSON NOT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_aicut_export_claim (state, created_at),
  KEY idx_aicut_export_project (project_id, owner_username, created_at),
  CONSTRAINT fk_aicut_export_project FOREIGN KEY (project_id, owner_username) REFERENCES aicut_projects(id, owner_username) ON DELETE CASCADE
) ENGINE=InnoDB`},
	}, {
		Version:          8800030,
		CallbackChecksum: "aicut-named-timelines-v1",
		SQL:              aicut.TimelineCollectionSchemaStatements(),
	}, {
		Version:          8800031,
		CallbackChecksum: "aicut-asr-timeline-target-v1",
		SQL:              []string{`ALTER TABLE aicut_asr_jobs ADD COLUMN timeline_id VARCHAR(64) NOT NULL DEFAULT '' AFTER project_id`},
	}, {
		Version:          8800032,
		CallbackChecksum: "aicut-agent-timeline-target-v1",
		SQL:              []string{`ALTER TABLE aicut_agent_runs ADD COLUMN timeline_id VARCHAR(64) NOT NULL DEFAULT '' AFTER project_id`},
	}}
	return append(append(out, AICutAssetFolderMigrations()...), AICutPersonalTemplateMigrations()...)
}
