-- doubao-relay 初始 schema(docs/spec/07-data-model.md 五张表 + 索引)。
-- PRAGMA journal_mode=WAL 与 foreign_keys=ON 在连接时由 storage::db 设置,不写在迁移里。

CREATE TABLE IF NOT EXISTS tasks (
    id              TEXT PRIMARY KEY,              -- ULID
    account         TEXT NOT NULL,                 -- 账号 name
    type            TEXT NOT NULL,                 -- 'image' | 'video'
    model_name      TEXT NOT NULL,
    prompt          TEXT NOT NULL,
    ratio           TEXT NOT NULL,
    duration        INTEGER,                       -- 视频独有
    reference_files TEXT NOT NULL DEFAULT '[]',    -- JSON array of upload_ids
    confirmation_mode TEXT NOT NULL DEFAULT 'auto',
    auto_use_membership INTEGER NOT NULL DEFAULT 1,
    callback_url    TEXT,

    status          TEXT NOT NULL,
    stage           TEXT,
    stage_message   TEXT,                          -- 当前阶段的用户可读消息
    error_code      TEXT,                          -- 失败时的错误码
    error_message   TEXT,
    error_details   TEXT,                          -- JSON

    conversation_id TEXT,                          -- 绑定的豆包会话 ID
    submission_id   TEXT,                          -- 绑定的豆包消息 ID
    local_message_id TEXT,                         -- 本地消息 ID(submissionIdentity)

    watermark_source TEXT,                         -- 'conversation_original' / 'lr_unwatermarked' / 'fallback_decrypted'
    result_files    TEXT,                          -- JSON array

    created_at      INTEGER NOT NULL,              -- Unix ms
    started_at      INTEGER,                       -- 开始 preparing 的时间
    completed_at    INTEGER,                       -- 终态时间
    expires_at      INTEGER NOT NULL               -- 结果文件过期时间
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_account ON tasks(account);
CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at DESC);

CREATE TABLE IF NOT EXISTS task_stage_history (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id         TEXT NOT NULL,
    stage           TEXT NOT NULL,                 -- pending/preparing/checking/...
    message         TEXT,
    ts              INTEGER NOT NULL,              -- Unix ms
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_stage_history_task ON task_stage_history(task_id, ts);

CREATE TABLE IF NOT EXISTS uploads (
    id              TEXT PRIMARY KEY,              -- ULID
    mime            TEXT NOT NULL,
    bytes           INTEGER NOT NULL,
    path            TEXT NOT NULL,                 -- 文件实际路径
    created_at      INTEGER NOT NULL,
    expires_at      INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_uploads_expires ON uploads(expires_at);

CREATE TABLE IF NOT EXISTS webhook_log (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id         TEXT NOT NULL,
    event           TEXT NOT NULL,                 -- 'task.completed' / 'task.failed' / ...
    url             TEXT NOT NULL,
    attempt         INTEGER NOT NULL,              -- 1-based
    http_status     INTEGER,
    response_body   TEXT,
    error_message   TEXT,
    sent_at         INTEGER NOT NULL,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_webhook_task ON webhook_log(task_id);

CREATE TABLE IF NOT EXISTS metadata (
    key             TEXT PRIMARY KEY,
    value           TEXT NOT NULL,
    updated_at      INTEGER NOT NULL
);
