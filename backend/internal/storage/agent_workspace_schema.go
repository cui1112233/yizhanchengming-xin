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
	}
}
