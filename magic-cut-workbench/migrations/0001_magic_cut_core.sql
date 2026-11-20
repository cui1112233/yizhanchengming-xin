CREATE TABLE IF NOT EXISTS magic_cut_projects (
  id VARCHAR(40) PRIMARY KEY,
  owner_username VARCHAR(191) NOT NULL,
  name VARCHAR(255) NOT NULL,
  width INT NOT NULL,
  height INT NOT NULL,
  frame_rate INT NOT NULL,
  idempotency_key VARCHAR(80) NOT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_magic_cut_project_create (owner_username, idempotency_key),
  KEY idx_magic_cut_projects_owner_updated (owner_username, updated_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS magic_cut_project_versions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  project_id VARCHAR(40) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  revision BIGINT UNSIGNED NOT NULL,
  payload_json JSON NOT NULL,
  created_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_magic_cut_project_revision (project_id, revision),
  KEY idx_magic_cut_project_versions_owner (owner_username, project_id),
  CONSTRAINT fk_magic_cut_version_project FOREIGN KEY (project_id) REFERENCES magic_cut_projects(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS magic_cut_materials (
  id VARCHAR(40) PRIMARY KEY,
  project_id VARCHAR(40) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  original_name VARCHAR(512) NOT NULL,
  media_type VARCHAR(32) NOT NULL,
  storage_key VARCHAR(1024) NOT NULL,
  size_bytes BIGINT UNSIGNED NOT NULL,
  metadata_json JSON NULL,
  status VARCHAR(32) NOT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  KEY idx_magic_cut_materials_owner_project (owner_username, project_id),
  CONSTRAINT fk_magic_cut_material_project FOREIGN KEY (project_id) REFERENCES magic_cut_projects(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS magic_cut_timeline_snapshots (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  project_id VARCHAR(40) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  revision BIGINT UNSIGNED NOT NULL,
  timeline_json JSON NOT NULL,
  created_at DATETIME(6) NOT NULL,
  UNIQUE KEY uq_magic_cut_timeline_revision (project_id, revision),
  KEY idx_magic_cut_timeline_owner_project (owner_username, project_id),
  CONSTRAINT fk_magic_cut_timeline_project FOREIGN KEY (project_id) REFERENCES magic_cut_projects(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS magic_cut_tasks (
  id VARCHAR(40) PRIMARY KEY,
  project_id VARCHAR(40) NULL,
  owner_username VARCHAR(191) NOT NULL,
  task_type VARCHAR(64) NOT NULL,
  status VARCHAR(32) NOT NULL,
  progress DECIMAL(6,3) NOT NULL DEFAULT 0,
  input_json JSON NULL,
  output_json JSON NULL,
  error_code VARCHAR(80) NULL,
  error_message TEXT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  KEY idx_magic_cut_tasks_owner_status (owner_username, status, updated_at),
  CONSTRAINT fk_magic_cut_task_project FOREIGN KEY (project_id) REFERENCES magic_cut_projects(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS magic_cut_task_attempts (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  task_id VARCHAR(40) NOT NULL,
  attempt_number INT UNSIGNED NOT NULL,
  status VARCHAR(32) NOT NULL,
  started_at DATETIME(6) NULL,
  finished_at DATETIME(6) NULL,
  error_code VARCHAR(80) NULL,
  error_message TEXT NULL,
  UNIQUE KEY uq_magic_cut_task_attempt (task_id, attempt_number),
  CONSTRAINT fk_magic_cut_attempt_task FOREIGN KEY (task_id) REFERENCES magic_cut_tasks(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS magic_cut_exports (
  id VARCHAR(40) PRIMARY KEY,
  project_id VARCHAR(40) NOT NULL,
  task_id VARCHAR(40) NOT NULL,
  owner_username VARCHAR(191) NOT NULL,
  output_name VARCHAR(512) NOT NULL,
  storage_key VARCHAR(1024) NULL,
  media_metadata_json JSON NULL,
  sha256 CHAR(64) NULL,
  created_at DATETIME(6) NOT NULL,
  KEY idx_magic_cut_exports_owner_project (owner_username, project_id, created_at),
  CONSTRAINT fk_magic_cut_export_project FOREIGN KEY (project_id) REFERENCES magic_cut_projects(id) ON DELETE CASCADE,
  CONSTRAINT fk_magic_cut_export_task FOREIGN KEY (task_id) REFERENCES magic_cut_tasks(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
