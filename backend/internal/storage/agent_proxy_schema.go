package storage

// AgentProxyStatements 是 Agent 代理最小骨架的表结构：工作区、会话、消息，
// 以及暂不启用的技能与定时任务占位表。
func AgentProxyStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS agent_proxy_workspaces (
  id VARCHAR(96) NOT NULL PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  name VARCHAR(191) NOT NULL,
  description VARCHAR(512) NOT NULL DEFAULT '',
  permission_mode VARCHAR(16) NOT NULL DEFAULT 'plan',
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  deleted_at DATETIME(6) NULL,
  KEY idx_agent_proxy_workspaces_owner_updated (owner_username, deleted_at, updated_at)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS agent_proxy_sessions (
  id VARCHAR(96) NOT NULL PRIMARY KEY,
  workspace_id VARCHAR(96) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  title VARCHAR(191) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  permission_mode VARCHAR(16) NOT NULL DEFAULT 'plan',
  last_message_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  deleted_at DATETIME(6) NULL,
  KEY idx_agent_proxy_sessions_owner_workspace (owner_username, workspace_id, deleted_at, updated_at)
) ENGINE=InnoDB`,
		`CREATE TABLE IF NOT EXISTS agent_proxy_messages (
  id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  session_id VARCHAR(96) NOT NULL,
  workspace_id VARCHAR(96) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  role VARCHAR(32) NOT NULL,
  content MEDIUMTEXT NOT NULL,
  metadata_json JSON NULL,
  created_at DATETIME(6) NOT NULL,
  KEY idx_agent_proxy_messages_session_created (owner_username, session_id, created_at, id)
) ENGINE=InnoDB`,
		// 技能表先建结构，第一版不加载技能包。
		`CREATE TABLE IF NOT EXISTS agent_proxy_skills (
  id VARCHAR(96) NOT NULL PRIMARY KEY,
  workspace_id VARCHAR(96) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  skill_key VARCHAR(191) NOT NULL,
  display_name VARCHAR(191) NOT NULL,
  description TEXT NOT NULL,
  source VARCHAR(32) NOT NULL DEFAULT 'workspace',
  status VARCHAR(32) NOT NULL DEFAULT 'draft',
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_agent_proxy_skills_owner_workspace_key (owner_username, workspace_id, skill_key),
  KEY idx_agent_proxy_skills_owner_workspace (owner_username, workspace_id)
) ENGINE=InnoDB`,
		// 定时任务第一版只读列表，不写入、不调度。
		`CREATE TABLE IF NOT EXISTS agent_proxy_scheduled_jobs (
  id VARCHAR(96) NOT NULL PRIMARY KEY,
  workspace_id VARCHAR(96) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  name VARCHAR(191) NOT NULL,
  schedule VARCHAR(64) NOT NULL,
  prompt TEXT NOT NULL,
  permission_mode VARCHAR(16) NOT NULL DEFAULT 'plan',
  enabled TINYINT(1) NOT NULL DEFAULT 1,
  last_run_at DATETIME(6) NULL,
  next_run_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  KEY idx_agent_proxy_scheduled_jobs_owner_workspace (owner_username, workspace_id, enabled, updated_at)
) ENGINE=InnoDB`,
	}
}

func AgentProxyMigrations() []Migration {
	return []Migration{
		{Version: 8800200, SQL: AgentProxyStatements(), CallbackChecksum: "v88-agent-proxy-skeleton-v1"},
	}
}
