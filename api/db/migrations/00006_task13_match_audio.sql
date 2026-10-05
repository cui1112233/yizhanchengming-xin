-- +goose Up
ALTER TABLE stage_runs
    ADD COLUMN validation_result VARCHAR(2048) NOT NULL DEFAULT '' AFTER error_message;

CREATE TABLE audio_measurements (
    id BIGINT NOT NULL AUTO_INCREMENT,
    batch_project_id BIGINT NOT NULL,
    book_id BIGINT NOT NULL,
    audio_asset TEXT NOT NULL,
    asset_hash CHAR(64) NOT NULL,
    duration_ms BIGINT UNSIGNED NOT NULL,
    measured_at DATETIME(6) NOT NULL,
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_audio_measurement_asset (batch_project_id, book_id, asset_hash),
    KEY idx_audio_measurement_latest (batch_project_id, book_id, measured_at, id),
    CONSTRAINT fk_audio_measurement_project FOREIGN KEY (batch_project_id) REFERENCES batch_projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_audio_measurement_book FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO generation_prompts (prompt_key, version, content, enabled) VALUES
('director.default', 2, '你是短视频导演分镜规划器。将剧本与 Hook 转为可用于视频提示词编译的导演 JSON 输出。matchAudio=false 时保持原 Director 行为，不强制使用音频时长。matchAudio=true 时，后端提供的 audioDurationSec 是硬性总时长：第一镜 start 必须为 0.00，后一个镜头 start 必须等于前一镜 end，不得有 gap 或 overlap，最后一镜 end 必须严格等于 audioDurationSec；时间单位为秒并保留两位小数。shotDurationLimitSec 是单镜头硬上限；选择 15 秒表示每镜头 <=15 秒，不是固定 15 秒切片。按对白、旁白、停顿、动作复杂度、情绪节奏与镜头信息量分配现有镜头时长。禁止为了凑时长新增人物、剧情、对白、旁白或无关镜头。保持剧情连续、空间明确、动作可视化。', 1),
('director.h3', 2, '你是 H3 结构化导演。必须只输出可解析 JSON，schema_version 固定为 "h3-director/v1"，并提供 director_cards。matchAudio=false 时保持原 H3 Director 行为，不强制使用音频时长。matchAudio=true 时，后端提供的 audioDurationSec 是硬性总时长：第一镜 start=0.00，连续镜头不得有 gap 或 overlap，最后一镜 end 必须严格等于 audioDurationSec；时间单位为秒并保留两位小数。shotDurationLimitSec 是单镜头硬上限；15 秒表示 <=15 秒，不是固定 15 秒切片。按对白、旁白、停顿、动作复杂度、情绪节奏与信息量分配现有镜头时长，禁止为了补足时长新增人物、剧情、对白、旁白或无关镜头。保持剧情连续、空间明确、动作可视化。', 1);

-- +goose Down
DELETE FROM generation_prompts WHERE prompt_key IN ('director.default','director.h3') AND version=2;
DROP TABLE IF EXISTS audio_measurements;
ALTER TABLE stage_runs DROP COLUMN validation_result;
