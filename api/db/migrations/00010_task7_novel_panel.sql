-- +goose Up
CREATE TABLE novel_panel_workspaces (
  batch_project_id BIGINT NOT NULL PRIMARY KEY,
  revision BIGINT NOT NULL,
  workspace_json JSON NOT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  CONSTRAINT fk_novel_panel_workspace_project FOREIGN KEY (batch_project_id) REFERENCES batch_projects(id) ON DELETE CASCADE
);

CREATE TABLE novel_panel_history (
  id VARCHAR(96) NOT NULL PRIMARY KEY,
  batch_project_id BIGINT NOT NULL,
  revision BIGINT NOT NULL,
  note VARCHAR(1000) NOT NULL DEFAULT '',
  summary_json JSON NOT NULL,
  workspace_json JSON NOT NULL,
  created_at DATETIME(6) NOT NULL,
  KEY idx_novel_panel_history_project_created (batch_project_id, created_at DESC),
  CONSTRAINT fk_novel_panel_history_project FOREIGN KEY (batch_project_id) REFERENCES batch_projects(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS novel_panel_history;
DROP TABLE IF EXISTS novel_panel_workspaces;
