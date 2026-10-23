package storage

func AgentWorkspaceStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS agent_projects (
  id VARCHAR(96) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  title VARCHAR(191) NOT NULL,
  title_mode VARCHAR(32) NOT NULL,
  preview VARCHAR(512) NOT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  deleted_at DATETIME(6) NULL,
  KEY idx_agent_projects_owner_updated (owner_username, deleted_at, updated_at)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS agent_messages (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  project_id VARCHAR(96) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  role VARCHAR(32) NOT NULL,
  content MEDIUMTEXT NOT NULL,
  context_json JSON NULL,
  created_at DATETIME(6) NOT NULL,
  KEY idx_agent_messages_project_created (owner_username, project_id, created_at, id)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS agent_canvases (
  project_id VARCHAR(96) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  canvas_version INT NOT NULL,
  revision BIGINT NOT NULL,
  viewport_json JSON NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  KEY idx_agent_canvases_owner (owner_username, project_id)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS agent_canvas_nodes (
  id VARCHAR(96) NOT NULL,
  project_id VARCHAR(96) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  node_type VARCHAR(64) NOT NULL,
  title VARCHAR(191) NOT NULL,
  summary TEXT NOT NULL,
  position_x DOUBLE NOT NULL,
  position_y DOUBLE NOT NULL,
  current_version BIGINT NOT NULL,
  status VARCHAR(32) NOT NULL,
  source_node_ids_json JSON NULL,
  asset_ref_json JSON NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  PRIMARY KEY (project_id, id),
  KEY idx_agent_canvas_nodes_owner_project (owner_username, project_id)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS agent_canvas_edges (
  id VARCHAR(96) NOT NULL,
  project_id VARCHAR(96) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  source_node_id VARCHAR(96) NOT NULL,
  target_node_id VARCHAR(96) NOT NULL,
  relation_type VARCHAR(64) NOT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  PRIMARY KEY (project_id, id),
  KEY idx_agent_canvas_edges_owner_project (owner_username, project_id)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS agent_node_versions (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  project_id VARCHAR(96) NOT NULL,
  node_id VARCHAR(96) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  version BIGINT NOT NULL,
  content_json JSON NOT NULL,
  source_execution_id VARCHAR(96) NULL,
  created_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_agent_node_versions (project_id, node_id, version),
  KEY idx_agent_node_versions_owner_project (owner_username, project_id)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS agent_executions (
  id VARCHAR(96) NOT NULL PRIMARY KEY,
  project_id VARCHAR(96) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  tool_name VARCHAR(96) NOT NULL,
  status VARCHAR(32) NOT NULL,
  detail_json JSON NULL,
  idempotency_key VARCHAR(191) NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_agent_executions_idempotency (owner_username, idempotency_key),
  KEY idx_agent_executions_project_updated (owner_username, project_id, updated_at)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS agent_legacy_imports (
  owner_username VARCHAR(191) NOT NULL,
  source_path VARCHAR(512) NOT NULL,
  source_sha256 CHAR(64) NOT NULL,
  imported_projects INT NOT NULL,
  imported_at DATETIME(6) NOT NULL,
  PRIMARY KEY (owner_username, source_path)
) ENGINE=InnoDB`,
	}
}

func AgentWorkspaceMigrations() []Migration {
	return []Migration{
		{Version: 8800001, SQL: AgentWorkspaceStatements(), CallbackChecksum: "v88-agent-workspace-foundation-v1"},
		{
			Version: 8800002,
			SQL: []string{
				`ALTER TABLE agent_canvas_nodes ADD COLUMN group_id VARCHAR(96) NULL AFTER asset_ref_json`,
			},
			CallbackChecksum: "v88-agent-canvas-groups-v1",
		},
		{
			Version: 8800003,
			SQL: []string{
				`ALTER TABLE agent_canvases ADD COLUMN appearance_json JSON NULL AFTER viewport_json`,
			},
			CallbackChecksum: "v88-agent-canvas-appearance-v1",
		},
		{
			Version: 8800004,
			SQL: []string{
				`CREATE TABLE IF NOT EXISTS agent_assets (
  id VARCHAR(96) NOT NULL,
  project_id VARCHAR(96) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  kind VARCHAR(16) NOT NULL,
  name VARCHAR(191) NOT NULL,
  content_type VARCHAR(127) NOT NULL,
  object_key VARCHAR(512) NOT NULL,
  byte_size BIGINT NOT NULL,
  created_at DATETIME(6) NOT NULL,
  PRIMARY KEY (project_id, id),
  KEY idx_agent_assets_owner_project_created (owner_username, project_id, created_at, id)
) ENGINE=InnoDB`,
			},
			CallbackChecksum: "v88-agent-project-assets-v1",
		},
		{
			Version: 8800005,
			SQL: []string{
				`ALTER TABLE agent_canvas_nodes ADD COLUMN config_json JSON NULL AFTER asset_ref_json`,
			},
			CallbackChecksum: "v88-agent-canvas-node-config-v1",
		},
		{
			Version: 8800006,
			SQL: []string{
				`CREATE TABLE IF NOT EXISTS agent_video_jobs (
  id VARCHAR(96) NOT NULL PRIMARY KEY,
  project_id VARCHAR(96) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  node_id VARCHAR(96) NOT NULL,
  node_version BIGINT NOT NULL,
  state VARCHAR(32) NOT NULL,
  model_key VARCHAR(96) NOT NULL,
  provider_task_id VARCHAR(191) NULL,
  settings_json JSON NOT NULL,
  reference_asset_ids_json JSON NOT NULL,
  result_asset_id VARCHAR(96) NULL,
  error_message TEXT NULL,
  idempotency_key VARCHAR(191) NOT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_agent_video_jobs_owner_idempotency (owner_username, idempotency_key),
  KEY idx_agent_video_jobs_project_node_updated (owner_username, project_id, node_id, updated_at)
) ENGINE=InnoDB`,
			},
			CallbackChecksum: "v88-agent-video-jobs-v1",
		},
		{
			Version: 8800007,
			SQL: []string{
				`ALTER TABLE agent_canvases ADD COLUMN view_state_json JSON NULL AFTER appearance_json`,
			},
			CallbackChecksum: "v88-agent-workspace-view-state-v1",
		},
		{
			Version: 8800008,
			SQL: []string{
				`CREATE TABLE IF NOT EXISTS agent_generation_tasks (
  id VARCHAR(96) NOT NULL PRIMARY KEY,
  project_id VARCHAR(96) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  node_id VARCHAR(96) NOT NULL,
  node_version BIGINT NOT NULL,
  kind VARCHAR(32) NOT NULL,
  provider VARCHAR(96) NOT NULL,
  model_id VARCHAR(191) NOT NULL,
  mode VARCHAR(64) NULL,
  state VARCHAR(32) NOT NULL,
  settings_json JSON NOT NULL,
  reference_asset_ids_json JSON NOT NULL,
  selected_result_id VARCHAR(96) NULL,
  idempotency_key VARCHAR(191) NOT NULL,
  retry_of_task_id VARCHAR(96) NULL,
  error_message TEXT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_agent_generation_tasks_owner_project_idempotency (owner_username, project_id, idempotency_key),
  KEY idx_agent_generation_tasks_owner_project_updated (owner_username, project_id, updated_at, id)
) ENGINE=InnoDB`,
				`CREATE TABLE IF NOT EXISTS agent_generation_results (
  id VARCHAR(96) NOT NULL PRIMARY KEY,
  task_id VARCHAR(96) NOT NULL,
  project_id VARCHAR(96) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  asset_id VARCHAR(96) NOT NULL,
  kind VARCHAR(32) NOT NULL,
  asset_json JSON NULL,
  created_at DATETIME(6) NOT NULL,
  KEY idx_agent_generation_results_task (owner_username, project_id, task_id, created_at, id)
) ENGINE=InnoDB`,
			},
			CallbackChecksum: "v88-agent-generation-tasks-v1",
		},
		{
			Version: 8800009,
			SQL: []string{
				`ALTER TABLE agent_canvases ADD COLUMN canvas_json JSON NULL AFTER revision`,
			},
			CallbackChecksum: "v88-agent-canvas-json-snapshot-v1",
		},
		{
			Version: 8800010,
			SQL: []string{
				`UPDATE agent_canvases c
SET canvas_json = JSON_OBJECT(
  'version', c.canvas_version,
  'revision', c.revision,
  'viewport', COALESCE(c.viewport_json, JSON_OBJECT()),
  'viewState', COALESCE(c.view_state_json, JSON_OBJECT()),
  'appearance', COALESCE(c.appearance_json, JSON_OBJECT()),
  'nodes', COALESCE((SELECT JSON_ARRAYAGG(JSON_OBJECT(
    'id', n.id, 'type', n.node_type,
    'position', JSON_OBJECT('x', n.position_x, 'y', n.position_y),
    'title', n.title, 'version', n.current_version, 'status', n.status,
    'summary', n.summary, 'sourceNodeIds', COALESCE(n.source_node_ids_json, JSON_ARRAY()),
    'assetRef', n.asset_ref_json, 'config', n.config_json, 'groupId', n.group_id
  )) FROM agent_canvas_nodes n WHERE n.project_id = c.project_id AND n.owner_username = c.owner_username), JSON_ARRAY()),
  'edges', COALESCE((SELECT JSON_ARRAYAGG(JSON_OBJECT(
    'id', e.id, 'source', e.source_node_id, 'target', e.target_node_id, 'relation', e.relation_type
  )) FROM agent_canvas_edges e WHERE e.project_id = c.project_id AND e.owner_username = c.owner_username), JSON_ARRAY())
)
WHERE c.canvas_json IS NULL`,
			},
			CallbackChecksum: "v88-agent-canvas-json-backfill-v1",
		},
		{
			Version: 8800011,
			SQL: []string{
				`CREATE TABLE IF NOT EXISTS agent_script_shots (
  id VARCHAR(96) NOT NULL,
  project_id VARCHAR(96) NOT NULL,
  script_node_id VARCHAR(96) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  sequence INT NOT NULL,
  scene TEXT NOT NULL,
  characters TEXT NOT NULL,
  dialogue TEXT NOT NULL,
  narration TEXT NOT NULL,
  image_prompt TEXT NOT NULL,
  video_prompt TEXT NOT NULL,
  tts_text TEXT NOT NULL,
  revision BIGINT NOT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  PRIMARY KEY (project_id, id),
  UNIQUE KEY uq_agent_script_shots_sequence (project_id, script_node_id, sequence),
  KEY idx_agent_script_shots_owner_script_sequence (owner_username, project_id, script_node_id, sequence)
) ENGINE=InnoDB`,
				`CREATE TABLE IF NOT EXISTS agent_shot_generation_links (
  project_id VARCHAR(96) NOT NULL,
  script_node_id VARCHAR(96) NOT NULL,
  shot_id VARCHAR(96) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  node_id VARCHAR(96) NOT NULL,
  generation_id VARCHAR(96) NOT NULL,
  attempt INT NOT NULL,
  active BOOLEAN NOT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_agent_shot_generation_links (project_id, shot_id, node_id, generation_id),
  KEY idx_agent_shot_generation_links_owner_shot (owner_username, project_id, shot_id, attempt)
) ENGINE=InnoDB`,
			},
			CallbackChecksum: "v88-agent-script-shots-v1",
		},
		{
			Version: 8800012,
			SQL: []string{
				`CREATE TABLE IF NOT EXISTS agent_batch_factory_bindings (
  owner_username VARCHAR(191) NOT NULL,
  batch_id VARCHAR(96) NOT NULL,
  project_id VARCHAR(96) NOT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_agent_batch_factory_bindings_owner_batch (owner_username, batch_id),
  UNIQUE KEY uq_agent_batch_factory_bindings_owner_project (owner_username, project_id)
) ENGINE=InnoDB`,
				`CREATE TABLE IF NOT EXISTS agent_batch_factory_node_origins (
  project_id VARCHAR(96) NOT NULL,
  node_id VARCHAR(96) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  batch_id VARCHAR(96) NOT NULL,
  book_id VARCHAR(96) NOT NULL,
  resource_type VARCHAR(32) NOT NULL,
  resource_id VARCHAR(191) NOT NULL,
  version_id VARCHAR(191) NULL,
  PRIMARY KEY (project_id, node_id),
  KEY idx_agent_batch_factory_node_origins_owner_book (owner_username, book_id, project_id)
) ENGINE=InnoDB`,
			},
			CallbackChecksum: "v88-agent-batch-factory-bindings-v1",
		},
	}
}
