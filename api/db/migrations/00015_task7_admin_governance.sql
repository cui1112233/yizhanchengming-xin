-- +goose Up
ALTER TABLE generation_prompts
  ADD COLUMN lifecycle ENUM('draft','published','archived') NOT NULL DEFAULT 'archived' AFTER enabled,
  ADD COLUMN seed_source ENUM('go_default','admin','restore','legacy') NOT NULL DEFAULT 'legacy' AFTER lifecycle,
  ADD COLUMN content_sha256 CHAR(64) NOT NULL DEFAULT '' AFTER content,
  ADD COLUMN published_at DATETIME(6) NULL AFTER updated_at,
  ADD COLUMN published_by_user_id BIGINT UNSIGNED NULL AFTER published_at,
  ADD COLUMN active_prompt_key VARCHAR(128) GENERATED ALWAYS AS (CASE WHEN lifecycle = 'published' AND enabled = 1 THEN prompt_key ELSE NULL END) STORED;

UPDATE generation_prompts SET lifecycle='archived', seed_source='legacy', content_sha256=SHA2(content,256);
UPDATE generation_prompts candidate JOIN (SELECT prompt_key, MAX(version) version FROM generation_prompts WHERE enabled=1 GROUP BY prompt_key) winner ON winner.prompt_key=candidate.prompt_key AND winner.version=candidate.version SET candidate.lifecycle='published', candidate.published_at=candidate.updated_at;
UPDATE generation_prompts SET enabled=0 WHERE lifecycle <> 'published';
ALTER TABLE generation_prompts ADD KEY idx_generation_prompts_lifecycle (prompt_key,lifecycle,version), ADD UNIQUE KEY uq_generation_prompts_one_published (active_prompt_key), ADD CONSTRAINT fk_generation_prompts_publisher FOREIGN KEY (published_by_user_id) REFERENCES auth_users(id) ON DELETE SET NULL;

CREATE TABLE admin_audit_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  actor_user_id BIGINT UNSIGNED NOT NULL,
  capability VARCHAR(128) NOT NULL,
  action VARCHAR(64) NOT NULL,
  resource_type VARCHAR(64) NOT NULL,
  resource_id VARCHAR(191) NOT NULL,
  prompt_version_id BIGINT NULL,
  request_id VARCHAR(191) NOT NULL DEFAULT '',
  result VARCHAR(32) NOT NULL,
  summary_json JSON NOT NULL,
  content_sha256 CHAR(64) NOT NULL DEFAULT '',
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id), KEY idx_admin_audit_actor_created (actor_user_id,created_at,id), KEY idx_admin_audit_resource_created (resource_type,resource_id,created_at,id), KEY idx_admin_audit_request_id (request_id),
  CONSTRAINT fk_admin_audit_actor FOREIGN KEY (actor_user_id) REFERENCES auth_users(id) ON DELETE RESTRICT,
  CONSTRAINT fk_admin_audit_prompt_version FOREIGN KEY (prompt_version_id) REFERENCES generation_prompts(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- +goose Down
DROP TABLE IF EXISTS admin_audit_logs;
ALTER TABLE generation_prompts DROP FOREIGN KEY fk_generation_prompts_publisher, DROP INDEX uq_generation_prompts_one_published, DROP INDEX idx_generation_prompts_lifecycle, DROP COLUMN active_prompt_key, DROP COLUMN published_by_user_id, DROP COLUMN published_at, DROP COLUMN content_sha256, DROP COLUMN seed_source, DROP COLUMN lifecycle;
