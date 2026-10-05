-- +goose Up
CREATE TABLE generation_prompts (
    id BIGINT NOT NULL AUTO_INCREMENT,
    prompt_key VARCHAR(128) NOT NULL,
    version INT NOT NULL,
    content MEDIUMTEXT NOT NULL,
    enabled TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_generation_prompts_key_version (prompt_key, version),
    KEY idx_generation_prompts_active (prompt_key, enabled, version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE book_runs (
    id BIGINT NOT NULL AUTO_INCREMENT,
    batch_project_id BIGINT NOT NULL,
    book_id BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    request_id VARCHAR(191) NOT NULL DEFAULT '',
    error_message VARCHAR(1024) NOT NULL DEFAULT '',
    started_at DATETIME(6) NULL,
    finished_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_book_runs_project_book (batch_project_id, book_id, id),
    KEY idx_book_runs_project_status (batch_project_id, status),
    CONSTRAINT fk_book_runs_project FOREIGN KEY (batch_project_id) REFERENCES batch_projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_book_runs_book FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE stage_runs (
    id BIGINT NOT NULL AUTO_INCREMENT,
    book_run_id BIGINT NOT NULL,
    book_id BIGINT NOT NULL,
    stage VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    attempt INT NOT NULL,
    request_id VARCHAR(191) NOT NULL DEFAULT '',
    prompt_key VARCHAR(128) NOT NULL DEFAULT '',
    prompt_version INT NOT NULL DEFAULT 0,
    input_snapshot MEDIUMTEXT NOT NULL,
    output_text MEDIUMTEXT NOT NULL,
    error_message VARCHAR(1024) NOT NULL DEFAULT '',
    started_at DATETIME(6) NULL,
    finished_at DATETIME(6) NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_stage_runs_attempt (book_run_id, stage, attempt),
    KEY idx_stage_runs_book_run_stage (book_run_id, stage, id),
    KEY idx_stage_runs_book_status (book_id, status),
    CONSTRAINT fk_stage_runs_book_run FOREIGN KEY (book_run_id) REFERENCES book_runs(id) ON DELETE CASCADE,
    CONSTRAINT fk_stage_runs_book FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO generation_prompts (prompt_key, version, content, enabled) VALUES
('script.default', 1, '根据小说原文生成结构清晰、可继续进入 Hook/Director 的剧本。保留主要事实、人物和因果，不虚构关键剧情。', 1),
('script.plot_mode', 1, '通用剧情模式：不得改变主剧情；首镜头优先强动作、强冲突、强情绪或明确事件；明确空间与动作；合理使用特写、中景、全景、推镜、拉镜、摇镜、跟拍、景深、环境光、阴影与空间调度；不得强化性暗示，不得色情化未成年人。', 1),
('hook.default', 1, '基于已生成剧本提炼短视频开场 Hook。强化冲突与信息密度，但不得改变主剧情。', 1),
('director.default', 1, '你是短视频导演分镜规划器。将剧本与 Hook 转成可继续用于视频提示词编译的导演输出。保持剧情连续、空间明确、动作可视化。', 1),
('director.h3', 1, '你是 H3 结构化导演。输出稳定结构的导演结果。matchAudio 与 audioDurationSec 仅作为兼容输入字段记录，本阶段不得执行精确音频时长重排。', 1),
('final_prompt.default', 1, '按后端确定性顺序编译最终提示词，不在前端拼装。', 1);

-- +goose Down
DROP TABLE IF EXISTS stage_runs;
DROP TABLE IF EXISTS book_runs;
DROP TABLE IF EXISTS generation_prompts;
