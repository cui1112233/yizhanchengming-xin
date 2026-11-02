//! SQLite 访问层(`docs/spec/07-data-model.md`)。
//!
//! - 连接时设置 `journal_mode=WAL`、`foreign_keys=ON`、`busy_timeout=5s`;
//!   迁移用 `sqlx::migrate!("../migrations")` 编译期嵌入。
//! - 全部 `sqlx::query()` 运行时绑定,不用 `query!` 宏。
//! - `Task.progress` / `Task.queue_position` 是运行时态,不入库,读出来恒为 `None`。
//! - sqlx 错误统一映射为 `system_database_error`(details.db)。

use std::path::{Path, PathBuf};
use std::time::Duration;

use anyhow::Context;
use serde_json::{json, Value};
use sqlx::sqlite::{
    SqliteConnectOptions, SqliteJournalMode, SqlitePoolOptions, SqliteRow, SqliteSynchronous,
};
use sqlx::{Row, SqlitePool};

use super::files;
use crate::error::{RelayError, RelayResult};
use crate::tasks::{
    ConfirmationMode, GenerateRequest, ResultFile, StageEntry, Task, TaskError, TaskStage,
    TaskStatus, TaskSummary, TaskType, WatermarkSource,
};

/// 非终态状态列表(08-task-lifecycle "恢复策略")。
const NON_TERMINAL_STATUSES: &str = "('pending','preparing','generating','waiting_user_confirm')";
const RUNNING_NON_TERMINAL_STATUSES: &str = "('preparing','generating','waiting_user_confirm')";

const TASK_COLUMNS: &str =
    "id, account, type, model_name, prompt, ratio, duration, reference_files, \
confirmation_mode, auto_use_membership, callback_url, status, stage, stage_message, error_code, \
error_message, error_details, conversation_id, submission_id, local_message_id, watermark_source, \
result_files, created_at, started_at, completed_at, expires_at";

/// 持久化入口:SQLite 连接池 + 应用数据根目录。
#[derive(Debug, Clone)]
pub struct Storage {
    pub(crate) pool: SqlitePool,
    /// 应用数据目录(`relay.db` / `uploads/` / `results/` / `logs/` 都在它下面)。
    pub root: PathBuf,
}

/// `uploads` 表一行。
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Upload {
    pub id: String,
    pub mime: String,
    pub bytes: u64,
    pub path: PathBuf,
    pub created_at: i64,
    pub expires_at: i64,
}

/// `webhook_log` 表一行(`id` 自增,插入时不需要)。
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct WebhookLog {
    pub task_id: String,
    /// `task.completed` / `task.failed` / `task.cancelled`。
    pub event: String,
    pub url: String,
    /// 1-based。
    pub attempt: u32,
    pub http_status: Option<u16>,
    pub response_body: Option<String>,
    pub error_message: Option<String>,
    pub sent_at: i64,
}

impl From<sqlx::Error> for RelayError {
    fn from(e: sqlx::Error) -> Self {
        RelayError::new(
            "system_database_error",
            "数据库读写失败,请检查磁盘空间与目录权限后重启 doubao-relay",
        )
        .with_details(json!({ "db": e.to_string() }))
    }
}

fn task_not_found(id: &str) -> RelayError {
    RelayError::new("task_not_found", "任务不存在,请确认任务 ID 后重试")
        .with_details(json!({ "task_id": id }))
}

fn corrupt(column: &str, value: impl Into<Value>) -> RelayError {
    RelayError::new(
        "system_database_error",
        "数据库中存在无法识别的字段值,请重启 doubao-relay 或清理数据目录",
    )
    .with_details(json!({ "column": column, "value": value.into() }))
}

fn json_string<T: serde::Serialize>(v: &T) -> RelayResult<String> {
    serde_json::to_string(v).map_err(|e| RelayError::internal(format!("序列化字段失败:{e}")))
}

fn parse_json<T: serde::de::DeserializeOwned>(column: &str, text: &str) -> RelayResult<T> {
    serde_json::from_str(text).map_err(|_| corrupt(column, text))
}

impl Storage {
    /// 建目录(uploads/ results/ logs/)→ 打开/创建 `relay.db`(WAL)→ 跑迁移。
    pub async fn open(root: &Path) -> anyhow::Result<Storage> {
        files::ensure_layout(root)
            .with_context(|| format!("创建数据目录失败:{}", root.display()))?;
        let db_path = files::db_path(root);
        let options = SqliteConnectOptions::new()
            .filename(&db_path)
            .create_if_missing(true)
            .journal_mode(SqliteJournalMode::Wal)
            .synchronous(SqliteSynchronous::Normal)
            .foreign_keys(true)
            .busy_timeout(Duration::from_secs(5));
        let pool = SqlitePoolOptions::new()
            .max_connections(4)
            .connect_with(options)
            .await
            .with_context(|| format!("打开数据库失败:{}", db_path.display()))?;
        sqlx::migrate!("../migrations")
            .run(&pool)
            .await
            .context("数据库迁移失败")?;
        Ok(Storage {
            pool,
            root: root.to_path_buf(),
        })
    }

    // ---- 路径 ----

    /// `<root>/results/{task_id}`。
    pub fn result_dir(&self, task_id: &str) -> PathBuf {
        files::result_dir(&self.root, task_id)
    }

    /// `<root>/uploads/{id}.{ext}`。
    pub fn upload_path(&self, id: &str, ext: &str) -> PathBuf {
        files::upload_path(&self.root, id, ext)
    }

    // ---- tasks ----

    /// 插入任务行 + 其 `stage_history`(同一事务)。
    pub async fn insert_task(&self, t: &Task) -> RelayResult<()> {
        let mut tx = self.pool.begin().await?;
        let sql = format!(
            "INSERT INTO tasks ({TASK_COLUMNS}) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
        );
        let mut q = sqlx::query(&sql).bind(&t.id);
        q = bind_task_fields(q, t)?;
        q.execute(&mut *tx).await?;
        for entry in &t.stage_history {
            sqlx::query(
                "INSERT INTO task_stage_history (task_id, stage, message, ts) VALUES (?, ?, ?, ?)",
            )
            .bind(&t.id)
            .bind(&entry.stage)
            .bind(&entry.message)
            .bind(entry.ts)
            .execute(&mut *tx)
            .await?;
        }
        tx.commit().await?;
        Ok(())
    }

    /// 读任务(含按 ts 升序的 `stage_history`);不存在返回 `None`。
    pub async fn get_task(&self, id: &str) -> RelayResult<Option<Task>> {
        let sql = format!("SELECT {TASK_COLUMNS} FROM tasks WHERE id = ?");
        let Some(row) = sqlx::query(&sql)
            .bind(id)
            .fetch_optional(&self.pool)
            .await?
        else {
            return Ok(None);
        };
        let mut task = row_to_task(&row)?;
        task.stage_history = self.stage_history(id).await?;
        Ok(Some(task))
    }

    /// 全字段 UPDATE `tasks` 行(不动 `task_stage_history`,阶段用 `append_stage`)。
    /// 行不存在 → `task_not_found`。
    pub async fn update_task(&self, t: &Task) -> RelayResult<()> {
        let sql = "UPDATE tasks SET account = ?, type = ?, model_name = ?, prompt = ?, ratio = ?, \
duration = ?, reference_files = ?, confirmation_mode = ?, auto_use_membership = ?, callback_url = ?, \
status = ?, stage = ?, stage_message = ?, error_code = ?, error_message = ?, error_details = ?, \
conversation_id = ?, submission_id = ?, local_message_id = ?, watermark_source = ?, result_files = ?, \
created_at = ?, started_at = ?, completed_at = ?, expires_at = ? WHERE id = ?";
        let q = bind_task_fields(sqlx::query(sql), t)?.bind(&t.id);
        let res = q.execute(&self.pool).await?;
        if res.rows_affected() == 0 {
            return Err(task_not_found(&t.id));
        }
        Ok(())
    }

    /// 追加一条阶段历史。
    pub async fn append_stage(
        &self,
        id: &str,
        stage: &str,
        message: Option<&str>,
        ts: i64,
    ) -> RelayResult<()> {
        sqlx::query(
            "INSERT INTO task_stage_history (task_id, stage, message, ts) VALUES (?, ?, ?, ?)",
        )
        .bind(id)
        .bind(stage)
        .bind(message)
        .bind(ts)
        .execute(&self.pool)
        .await?;
        Ok(())
    }

    /// 按 `id DESC` 分页;`cursor` 为上一页最后一个 id,返回的 `next_cursor` 为本页最后 id(还有下一页时)。
    /// `status` 为 `None` 或空切片表示不过滤;`limit` 至少按 1 处理。
    pub async fn list_tasks(
        &self,
        status: Option<&[TaskStatus]>,
        account: Option<&str>,
        limit: u32,
        cursor: Option<&str>,
    ) -> RelayResult<(Vec<TaskSummary>, Option<String>)> {
        let limit = limit.max(1) as i64;
        let statuses = status.unwrap_or(&[]);
        let mut clauses: Vec<String> = Vec::new();
        if !statuses.is_empty() {
            let marks = vec!["?"; statuses.len()].join(", ");
            clauses.push(format!("status IN ({marks})"));
        }
        if account.is_some() {
            clauses.push("account = ?".into());
        }
        if cursor.is_some() {
            clauses.push("id < ?".into());
        }
        let mut sql = String::from(
            "SELECT id, account, type, status, stage, created_at, completed_at FROM tasks",
        );
        if !clauses.is_empty() {
            sql.push_str(" WHERE ");
            sql.push_str(&clauses.join(" AND "));
        }
        sql.push_str(" ORDER BY id DESC LIMIT ?");

        let mut q = sqlx::query(&sql);
        for s in statuses {
            q = q.bind(s.as_str());
        }
        if let Some(a) = account {
            q = q.bind(a);
        }
        if let Some(c) = cursor {
            q = q.bind(c);
        }
        q = q.bind(limit + 1);
        let rows = q.fetch_all(&self.pool).await?;

        let mut items = Vec::with_capacity(rows.len());
        for row in &rows {
            items.push(row_to_summary(row)?);
        }
        let next = if items.len() as i64 > limit {
            items.truncate(limit as usize);
            items.last().map(|s| s.id.clone())
        } else {
            None
        };
        Ok((items, next))
    }

    /// 把所有非终态任务置为 `failed`(error.code = `code`,`completed_at = now`),返回被改动任务的 id
    /// (调用方据此补发终态 Webhook)。启动恢复用 `relay_restart`,豆包崩溃用 `doubao_crashed`。
    pub async fn fail_non_terminal(
        &self,
        code: &str,
        message: &str,
        now: i64,
    ) -> RelayResult<Vec<String>> {
        let sql = format!(
            "UPDATE tasks SET status = 'failed', error_code = ?, error_message = ?, error_details = NULL, \
completed_at = ? WHERE status IN {NON_TERMINAL_STATUSES} RETURNING id"
        );
        let ids = sqlx::query_scalar::<_, String>(&sql)
            .bind(code)
            .bind(message)
            .bind(now)
            .fetch_all(&self.pool)
            .await?;
        Ok(ids)
    }

    /// 启动恢复时只结束上次真正运行中的任务,保留 pending 任务交给调度器恢复。
    pub async fn fail_running_non_terminal(
        &self,
        code: &str,
        message: &str,
        now: i64,
    ) -> RelayResult<Vec<String>> {
        let sql = format!(
            "UPDATE tasks SET status = 'failed', error_code = ?, error_message = ?, error_details = NULL, \
completed_at = ? WHERE status IN {RUNNING_NON_TERMINAL_STATUSES} RETURNING id"
        );
        let ids = sqlx::query_scalar::<_, String>(&sql)
            .bind(code)
            .bind(message)
            .bind(now)
            .fetch_all(&self.pool)
            .await?;
        Ok(ids)
    }

    async fn stage_history(&self, id: &str) -> RelayResult<Vec<StageEntry>> {
        let rows = sqlx::query(
            "SELECT stage, message, ts FROM task_stage_history WHERE task_id = ? ORDER BY ts ASC, id ASC",
        )
        .bind(id)
        .fetch_all(&self.pool)
        .await?;
        rows.iter()
            .map(|r| {
                Ok(StageEntry {
                    stage: r.try_get("stage")?,
                    message: r.try_get("message")?,
                    ts: r.try_get("ts")?,
                })
            })
            .collect()
    }

    // ---- uploads ----

    pub async fn insert_upload(&self, u: &Upload) -> RelayResult<()> {
        sqlx::query(
            "INSERT INTO uploads (id, mime, bytes, path, created_at, expires_at) VALUES (?, ?, ?, ?, ?, ?)",
        )
        .bind(&u.id)
        .bind(&u.mime)
        .bind(u.bytes as i64)
        .bind(u.path.to_string_lossy().into_owned())
        .bind(u.created_at)
        .bind(u.expires_at)
        .execute(&self.pool)
        .await?;
        Ok(())
    }

    pub async fn get_upload(&self, id: &str) -> RelayResult<Option<Upload>> {
        let row = sqlx::query(
            "SELECT id, mime, bytes, path, created_at, expires_at FROM uploads WHERE id = ?",
        )
        .bind(id)
        .fetch_optional(&self.pool)
        .await?;
        row.map(|r| {
            let bytes: i64 = r.try_get("bytes")?;
            let path: String = r.try_get("path")?;
            Ok(Upload {
                id: r.try_get("id")?,
                mime: r.try_get("mime")?,
                bytes: bytes.max(0) as u64,
                path: PathBuf::from(path),
                created_at: r.try_get("created_at")?,
                expires_at: r.try_get("expires_at")?,
            })
        })
        .transpose()
    }

    // ---- webhook_log ----

    pub async fn insert_webhook_log(&self, l: &WebhookLog) -> RelayResult<()> {
        sqlx::query(
            "INSERT INTO webhook_log (task_id, event, url, attempt, http_status, response_body, error_message, sent_at) \
VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
        )
        .bind(&l.task_id)
        .bind(&l.event)
        .bind(&l.url)
        .bind(l.attempt as i64)
        .bind(l.http_status.map(|s| s as i64))
        .bind(&l.response_body)
        .bind(&l.error_message)
        .bind(l.sent_at)
        .execute(&self.pool)
        .await?;
        Ok(())
    }

    // ---- metadata ----

    pub async fn get_meta(&self, key: &str) -> RelayResult<Option<String>> {
        let row = sqlx::query("SELECT value FROM metadata WHERE key = ?")
            .bind(key)
            .fetch_optional(&self.pool)
            .await?;
        row.map(|r| r.try_get::<String, _>("value").map_err(RelayError::from))
            .transpose()
    }

    pub async fn set_meta(&self, key: &str, value: &str, now: i64) -> RelayResult<()> {
        sqlx::query(
            "INSERT INTO metadata (key, value, updated_at) VALUES (?, ?, ?) \
ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at",
        )
        .bind(key)
        .bind(value)
        .bind(now)
        .execute(&self.pool)
        .await?;
        Ok(())
    }
}

type SqliteQuery<'q> = sqlx::query::Query<'q, sqlx::Sqlite, sqlx::sqlite::SqliteArguments<'q>>;

/// 按 `TASK_COLUMNS` 中 `id` 之后的顺序绑定 25 个字段。
fn bind_task_fields<'q>(q: SqliteQuery<'q>, t: &'q Task) -> RelayResult<SqliteQuery<'q>> {
    let r = &t.request;
    let reference_files = json_string(&r.reference_files)?;
    let error_details = t
        .error
        .as_ref()
        .and_then(|e| e.details.as_ref())
        .map(json_string)
        .transpose()?;
    let result_files = t.result_files.as_ref().map(json_string).transpose()?;
    Ok(q.bind(&r.account)
        .bind(r.task_type.as_str())
        .bind(&r.model_name)
        .bind(&r.prompt)
        .bind(&r.ratio)
        .bind(r.duration.map(i64::from))
        .bind(reference_files)
        .bind(r.confirmation_mode.as_str())
        .bind(i64::from(r.auto_use_membership))
        .bind(&r.callback_url)
        .bind(t.status.as_str())
        .bind(t.stage.map(|s| s.as_str()))
        .bind(&t.stage_message)
        .bind(t.error.as_ref().map(|e| e.code.as_str()))
        .bind(t.error.as_ref().map(|e| e.message.as_str()))
        .bind(error_details)
        .bind(&t.conversation_id)
        .bind(&t.submission_id)
        .bind(&t.local_message_id)
        .bind(t.watermark_source.map(|w| w.as_str()))
        .bind(result_files)
        .bind(t.created_at)
        .bind(t.started_at)
        .bind(t.completed_at)
        .bind(t.expires_at))
}

fn row_to_task(row: &SqliteRow) -> RelayResult<Task> {
    let task_type: String = row.try_get("type")?;
    let task_type = TaskType::parse(&task_type).ok_or_else(|| corrupt("type", task_type))?;
    let mode: String = row.try_get("confirmation_mode")?;
    let confirmation_mode =
        ConfirmationMode::parse(&mode).ok_or_else(|| corrupt("confirmation_mode", mode))?;
    let status: String = row.try_get("status")?;
    let status = TaskStatus::parse(&status).ok_or_else(|| corrupt("status", status))?;
    let stage: Option<String> = row.try_get("stage")?;
    let stage = stage
        .map(|s| TaskStage::parse(&s).ok_or_else(|| corrupt("stage", s)))
        .transpose()?;
    let watermark: Option<String> = row.try_get("watermark_source")?;
    let watermark_source = watermark
        .map(|w| WatermarkSource::parse(&w).ok_or_else(|| corrupt("watermark_source", w)))
        .transpose()?;
    let reference_files: String = row.try_get("reference_files")?;
    let reference_files: Vec<String> = parse_json("reference_files", &reference_files)?;
    let duration: Option<i64> = row.try_get("duration")?;
    let auto_use_membership: i64 = row.try_get("auto_use_membership")?;

    let error_code: Option<String> = row.try_get("error_code")?;
    let error = match error_code {
        Some(code) => {
            let message: Option<String> = row.try_get("error_message")?;
            let details: Option<String> = row.try_get("error_details")?;
            let details = details
                .map(|d| parse_json::<Value>("error_details", &d))
                .transpose()?;
            Some(TaskError {
                code,
                message: message.unwrap_or_default(),
                details,
            })
        }
        None => None,
    };
    let result_files: Option<String> = row.try_get("result_files")?;
    let result_files = result_files
        .map(|r| parse_json::<Vec<ResultFile>>("result_files", &r))
        .transpose()?;

    Ok(Task {
        id: row.try_get("id")?,
        request: GenerateRequest {
            account: row.try_get("account")?,
            task_type,
            model_name: row.try_get("model_name")?,
            prompt: row.try_get("prompt")?,
            ratio: row.try_get("ratio")?,
            duration: duration.map(|d| d.clamp(0, i64::from(u32::MAX)) as u32),
            reference_files,
            confirmation_mode,
            auto_use_membership: auto_use_membership != 0,
            callback_url: row.try_get("callback_url")?,
        },
        status,
        stage,
        stage_message: row.try_get("stage_message")?,
        progress: None,
        error,
        conversation_id: row.try_get("conversation_id")?,
        submission_id: row.try_get("submission_id")?,
        local_message_id: row.try_get("local_message_id")?,
        watermark_source,
        result_files,
        created_at: row.try_get("created_at")?,
        started_at: row.try_get("started_at")?,
        completed_at: row.try_get("completed_at")?,
        expires_at: row.try_get("expires_at")?,
        stage_history: Vec::new(),
        queue_position: None,
    })
}

fn row_to_summary(row: &SqliteRow) -> RelayResult<TaskSummary> {
    let task_type: String = row.try_get("type")?;
    let task_type = TaskType::parse(&task_type).ok_or_else(|| corrupt("type", task_type))?;
    let status: String = row.try_get("status")?;
    let status = TaskStatus::parse(&status).ok_or_else(|| corrupt("status", status))?;
    let stage: Option<String> = row.try_get("stage")?;
    let stage = stage
        .map(|s| TaskStage::parse(&s).ok_or_else(|| corrupt("stage", s)))
        .transpose()?;
    Ok(TaskSummary {
        id: row.try_get("id")?,
        account: row.try_get("account")?,
        task_type,
        status,
        stage,
        created_at: row.try_get("created_at")?,
        completed_at: row.try_get("completed_at")?,
    })
}

#[cfg(test)]
pub(crate) mod tests {
    use super::*;
    use crate::tasks::{ConfirmationMode, TaskType};

    pub(crate) async fn open_tmp() -> (tempfile::TempDir, Storage) {
        let dir = tempfile::tempdir().unwrap();
        let storage = Storage::open(dir.path()).await.unwrap();
        (dir, storage)
    }

    pub(crate) fn request(account: &str) -> GenerateRequest {
        GenerateRequest {
            account: account.into(),
            task_type: TaskType::Image,
            model_name: "Seedream 5.0 Pro".into(),
            prompt: "一只蓝色小鸟".into(),
            ratio: "16:9".into(),
            duration: None,
            reference_files: vec![],
            confirmation_mode: ConfirmationMode::Auto,
            auto_use_membership: true,
            callback_url: None,
        }
    }

    /// 最小 pending 任务;`id` 决定分页顺序(ULID 字典序)。
    pub(crate) fn pending_task(id: &str, account: &str, created_at: i64) -> Task {
        Task {
            id: id.into(),
            request: request(account),
            status: TaskStatus::Pending,
            stage: None,
            stage_message: None,
            progress: None,
            error: None,
            conversation_id: None,
            submission_id: None,
            local_message_id: None,
            watermark_source: None,
            result_files: None,
            created_at,
            started_at: None,
            completed_at: None,
            expires_at: created_at + 24 * 3_600_000,
            stage_history: vec![StageEntry {
                stage: "pending".into(),
                message: None,
                ts: created_at,
            }],
            queue_position: None,
        }
    }

    fn full_task() -> Task {
        Task {
            id: "01HQZ9ABCDEFGHJKMNPQRSTVWX".into(),
            request: GenerateRequest {
                account: "张三".into(),
                task_type: TaskType::Video,
                model_name: "Seedance 2.5".into(),
                prompt: "海边日落\n第二行".into(),
                ratio: "9:16".into(),
                duration: Some(10),
                reference_files: vec!["01HQZ8AAAAAAAAAAAAAAAAAAAA".into()],
                confirmation_mode: ConfirmationMode::Manual,
                auto_use_membership: false,
                callback_url: Some("https://example.com/hook".into()),
            },
            status: TaskStatus::Failed,
            stage: Some(TaskStage::Refused),
            stage_message: Some("豆包拒绝生成".into()),
            progress: None,
            error: Some(TaskError {
                code: "generation_doubao_refused".into(),
                message: "豆包拒绝生成,请调整提示词后重试".into(),
                details: Some(json!({ "reason": "policy", "n": 17 })),
            }),
            conversation_id: Some("12345678901234567890".into()),
            submission_id: Some("98765432109876543210".into()),
            local_message_id: Some("11111111111111111111".into()),
            watermark_source: Some(WatermarkSource::LrUnwatermarked),
            result_files: Some(vec![ResultFile {
                index: 0,
                filename: "video_1.mp4".into(),
                relative_path: "results/01HQZ9ABCDEFGHJKMNPQRSTVWX/video_1.mp4".into(),
                mime: "video/mp4".into(),
                bytes: 12_345_678,
                width: Some(1080),
                height: Some(1920),
                duration_ms: Some(10_000),
                source_url: None,
            }]),
            created_at: 1_697_433_600_000,
            started_at: Some(1_697_433_610_000),
            completed_at: Some(1_697_433_700_000),
            expires_at: 1_697_520_000_000,
            stage_history: vec![
                StageEntry {
                    stage: "pending".into(),
                    message: None,
                    ts: 1_697_433_600_000,
                },
                StageEntry {
                    stage: "preparing".into(),
                    message: None,
                    ts: 1_697_433_610_000,
                },
                StageEntry {
                    stage: "refused".into(),
                    message: Some("豆包拒绝生成".into()),
                    ts: 1_697_433_700_000,
                },
            ],
            queue_position: None,
        }
    }

    #[tokio::test]
    async fn open_creates_layout_and_migrates_idempotently() {
        let (dir, storage) = open_tmp().await;
        for d in ["uploads", "results", "logs"] {
            assert!(dir.path().join(d).is_dir(), "{d}");
        }
        assert!(dir.path().join("relay.db").is_file());
        assert_eq!(storage.root, dir.path());
        assert_eq!(
            storage.result_dir("01HQZ9"),
            dir.path().join("results").join("01HQZ9")
        );
        assert_eq!(
            storage.upload_path("01HQZ8", "png"),
            dir.path().join("uploads").join("01HQZ8.png")
        );
        // 第二次打开同一目录:迁移已应用,不报错
        let again = Storage::open(dir.path()).await.unwrap();
        assert_eq!(again.get_task("missing").await.unwrap(), None);
        // WAL 已启用
        let mode: String = sqlx::query_scalar("PRAGMA journal_mode")
            .fetch_one(&storage.pool)
            .await
            .unwrap();
        assert_eq!(mode.to_lowercase(), "wal");
        let fk: i64 = sqlx::query_scalar("PRAGMA foreign_keys")
            .fetch_one(&storage.pool)
            .await
            .unwrap();
        assert_eq!(fk, 1);
    }

    #[tokio::test]
    async fn insert_get_round_trips_all_fields_and_history_order() {
        let (_dir, storage) = open_tmp().await;
        let t = full_task();
        storage.insert_task(&t).await.unwrap();
        let got = storage.get_task(&t.id).await.unwrap().unwrap();
        assert_eq!(got, t);
        assert_eq!(got.stage_history.len(), 3);
        assert_eq!(got.stage_history[2].stage, "refused");
        assert_eq!(got.request.reference_files, t.request.reference_files);
        assert_eq!(got.error.unwrap().details.unwrap()["n"], 17);
        assert_eq!(storage.get_task("nope").await.unwrap(), None);
        // 重复插入同一 id → 数据库错误
        let err = storage.insert_task(&t).await.unwrap_err();
        assert_eq!(err.code, "system_database_error");
        assert_eq!(err.http_status(), 500);
    }

    #[tokio::test]
    async fn insert_minimal_task_reads_back_with_none_fields() {
        let (_dir, storage) = open_tmp().await;
        let t = pending_task("01HQZ9AAAAAAAAAAAAAAAAAAAA", "张三", 1_000);
        storage.insert_task(&t).await.unwrap();
        let got = storage.get_task(&t.id).await.unwrap().unwrap();
        assert_eq!(got, t);
        assert_eq!(got.error, None);
        assert_eq!(got.result_files, None);
        assert_eq!(got.request.duration, None);
        assert_eq!(got.request.callback_url, None);
        assert!(got.request.auto_use_membership);
    }

    #[tokio::test]
    async fn update_task_overwrites_all_columns() {
        let (_dir, storage) = open_tmp().await;
        let mut t = pending_task("01HQZ9AAAAAAAAAAAAAAAAAAAA", "张三", 1_000);
        storage.insert_task(&t).await.unwrap();

        t.status = TaskStatus::Completed;
        t.stage = Some(TaskStage::Result);
        t.stage_message = Some("已识别到结果".into());
        t.conversation_id = Some("1".into());
        t.submission_id = Some("2".into());
        t.local_message_id = Some("3".into());
        t.watermark_source = Some(WatermarkSource::ConversationOriginal);
        t.result_files = Some(vec![ResultFile {
            index: 0,
            filename: "image_1.png".into(),
            relative_path: "results/01HQZ9AAAAAAAAAAAAAAAAAAAA/image_1.png".into(),
            mime: "image/png".into(),
            bytes: 10,
            width: Some(1),
            height: Some(1),
            duration_ms: None,
            source_url: None,
        }]);
        t.started_at = Some(1_100);
        t.completed_at = Some(1_200);
        t.expires_at = 9_999;
        t.request.auto_use_membership = false;
        t.request.reference_files = vec!["u1".into(), "u2".into()];
        storage.update_task(&t).await.unwrap();
        let got = storage.get_task(&t.id).await.unwrap().unwrap();
        assert_eq!(got, t);

        // 再把 result_files / error 清掉也能写回 NULL
        t.result_files = None;
        t.error = Some(TaskError {
            code: "x".into(),
            message: "y".into(),
            details: None,
        });
        storage.update_task(&t).await.unwrap();
        let got = storage.get_task(&t.id).await.unwrap().unwrap();
        assert_eq!(got.result_files, None);
        assert_eq!(got.error.as_ref().unwrap().details, None);
        assert_eq!(got, t);

        let missing = pending_task("01HQZ9ZZZZZZZZZZZZZZZZZZZZ", "李四", 1);
        let err = storage.update_task(&missing).await.unwrap_err();
        assert_eq!(err.code, "task_not_found");
        assert_eq!(err.http_status(), 404);
    }

    #[tokio::test]
    async fn append_stage_keeps_ts_order() {
        let (_dir, storage) = open_tmp().await;
        let t = pending_task("01HQZ9AAAAAAAAAAAAAAAAAAAA", "张三", 1_000);
        storage.insert_task(&t).await.unwrap();
        storage
            .append_stage(&t.id, "preparing", None, 2_000)
            .await
            .unwrap();
        storage
            .append_stage(&t.id, "checking", Some("豆包确认参数"), 3_000)
            .await
            .unwrap();
        let got = storage.get_task(&t.id).await.unwrap().unwrap();
        let stages: Vec<&str> = got.stage_history.iter().map(|e| e.stage.as_str()).collect();
        assert_eq!(stages, ["pending", "preparing", "checking"]);
        assert_eq!(
            got.stage_history[2].message.as_deref(),
            Some("豆包确认参数")
        );
        assert_eq!(got.stage_history[1].message, None);
        // 不存在的任务:外键拒绝
        let err = storage
            .append_stage("missing", "preparing", None, 1)
            .await
            .unwrap_err();
        assert_eq!(err.code, "system_database_error");
    }

    #[tokio::test]
    async fn list_tasks_paginates_desc_by_id_with_cursor_and_filters() {
        let (_dir, storage) = open_tmp().await;
        let ids = [
            "01HQZ9AAAAAAAAAAAAAAAAAAA1",
            "01HQZ9AAAAAAAAAAAAAAAAAAA2",
            "01HQZ9AAAAAAAAAAAAAAAAAAA3",
            "01HQZ9AAAAAAAAAAAAAAAAAAA4",
            "01HQZ9AAAAAAAAAAAAAAAAAAA5",
        ];
        for (i, id) in ids.iter().enumerate() {
            let mut t = pending_task(id, if i % 2 == 0 { "张三" } else { "李四" }, i as i64);
            if i == 4 {
                t.status = TaskStatus::Completed;
                t.completed_at = Some(99);
            }
            if i == 3 {
                t.status = TaskStatus::Generating;
                t.stage = Some(TaskStage::Rendering);
            }
            storage.insert_task(&t).await.unwrap();
        }

        let (page1, cur1) = storage.list_tasks(None, None, 2, None).await.unwrap();
        assert_eq!(
            page1.iter().map(|s| s.id.as_str()).collect::<Vec<_>>(),
            [ids[4], ids[3]]
        );
        assert_eq!(cur1.as_deref(), Some(ids[3]));
        assert_eq!(page1[0].status, TaskStatus::Completed);
        assert_eq!(page1[0].completed_at, Some(99));
        assert_eq!(page1[1].stage, Some(TaskStage::Rendering));
        assert_eq!(page1[1].task_type, TaskType::Image);

        let (page2, cur2) = storage
            .list_tasks(None, None, 2, cur1.as_deref())
            .await
            .unwrap();
        assert_eq!(
            page2.iter().map(|s| s.id.as_str()).collect::<Vec<_>>(),
            [ids[2], ids[1]]
        );
        assert_eq!(cur2.as_deref(), Some(ids[1]));

        let (page3, cur3) = storage
            .list_tasks(None, None, 2, cur2.as_deref())
            .await
            .unwrap();
        assert_eq!(
            page3.iter().map(|s| s.id.as_str()).collect::<Vec<_>>(),
            [ids[0]]
        );
        assert_eq!(cur3, None);

        // 正好填满一页时不给 cursor
        let (full, cur) = storage.list_tasks(None, None, 5, None).await.unwrap();
        assert_eq!(full.len(), 5);
        assert_eq!(cur, None);

        // 状态过滤(多个)
        let (filtered, _) = storage
            .list_tasks(
                Some(&[TaskStatus::Completed, TaskStatus::Generating]),
                None,
                10,
                None,
            )
            .await
            .unwrap();
        assert_eq!(
            filtered.iter().map(|s| s.id.as_str()).collect::<Vec<_>>(),
            [ids[4], ids[3]]
        );
        // 空状态切片 = 不过滤
        let (all, _) = storage.list_tasks(Some(&[]), None, 10, None).await.unwrap();
        assert_eq!(all.len(), 5);
        // 账号过滤 + 状态过滤组合
        let (acc, _) = storage
            .list_tasks(Some(&[TaskStatus::Pending]), Some("张三"), 10, None)
            .await
            .unwrap();
        assert_eq!(
            acc.iter().map(|s| s.id.as_str()).collect::<Vec<_>>(),
            [ids[2], ids[0]]
        );
        assert!(acc.iter().all(|s| s.account == "张三"));
        let (none, cur) = storage
            .list_tasks(None, Some("王五"), 10, None)
            .await
            .unwrap();
        assert!(none.is_empty());
        assert_eq!(cur, None);
        // limit 0 按 1 处理
        let (one, cur) = storage.list_tasks(None, None, 0, None).await.unwrap();
        assert_eq!(one.len(), 1);
        assert_eq!(cur.as_deref(), Some(ids[4]));
    }

    #[tokio::test]
    async fn fail_non_terminal_only_changes_non_terminal_rows() {
        let (_dir, storage) = open_tmp().await;
        let statuses = [
            TaskStatus::Pending,
            TaskStatus::Preparing,
            TaskStatus::Generating,
            TaskStatus::WaitingUserConfirm,
            TaskStatus::Completed,
            TaskStatus::Failed,
            TaskStatus::Cancelled,
        ];
        for (i, st) in statuses.iter().enumerate() {
            let mut t = pending_task(&format!("01HQZ9AAAAAAAAAAAAAAAAAAA{i}"), "张三", 1);
            t.status = *st;
            if *st == TaskStatus::Failed {
                t.error = Some(TaskError {
                    code: "generation_cancelled".into(),
                    message: "m".into(),
                    details: None,
                });
                t.completed_at = Some(5);
            }
            storage.insert_task(&t).await.unwrap();
        }
        let n = storage
            .fail_non_terminal("relay_restart", "中转程序已重启,任务作废,请重新提交", 777)
            .await
            .unwrap();
        let mut n = n;
        n.sort();
        assert_eq!(
            n,
            (0..4)
                .map(|i| format!("01HQZ9AAAAAAAAAAAAAAAAAAA{i}"))
                .collect::<Vec<_>>()
        );
        for (i, st) in statuses.iter().enumerate() {
            let got = storage
                .get_task(&format!("01HQZ9AAAAAAAAAAAAAAAAAAA{i}"))
                .await
                .unwrap()
                .unwrap();
            if st.is_terminal() {
                assert_eq!(got.status, *st, "终态不应被改动");
                if *st == TaskStatus::Failed {
                    assert_eq!(got.error.unwrap().code, "generation_cancelled");
                    assert_eq!(got.completed_at, Some(5));
                } else {
                    assert_eq!(got.error, None);
                }
            } else {
                assert_eq!(got.status, TaskStatus::Failed);
                let e = got.error.unwrap();
                assert_eq!(e.code, "relay_restart");
                assert!(e.message.contains("重启"));
                assert_eq!(e.details, None);
                assert_eq!(got.completed_at, Some(777));
            }
        }
        // 再跑一次:没有非终态了
        let n = storage
            .fail_non_terminal("doubao_crashed", "x", 778)
            .await
            .unwrap();
        assert!(n.is_empty());
    }

    #[tokio::test]
    async fn upload_round_trip() {
        let (dir, storage) = open_tmp().await;
        let u = Upload {
            id: "01HQZ8AAAAAAAAAAAAAAAAAAAA".into(),
            mime: "image/png".into(),
            bytes: 4_294_967_296,
            path: storage.upload_path("01HQZ8AAAAAAAAAAAAAAAAAAAA", "png"),
            created_at: 1_000,
            expires_at: 4_600_000,
        };
        storage.insert_upload(&u).await.unwrap();
        let got = storage.get_upload(&u.id).await.unwrap().unwrap();
        assert_eq!(got, u);
        assert!(got.path.starts_with(dir.path()));
        assert_eq!(storage.get_upload("missing").await.unwrap(), None);
        assert_eq!(
            storage.insert_upload(&u).await.unwrap_err().code,
            "system_database_error"
        );
    }

    #[tokio::test]
    async fn webhook_log_insert_requires_existing_task() {
        let (_dir, storage) = open_tmp().await;
        let t = pending_task("01HQZ9AAAAAAAAAAAAAAAAAAAA", "张三", 1);
        storage.insert_task(&t).await.unwrap();
        let log = WebhookLog {
            task_id: t.id.clone(),
            event: "task.completed".into(),
            url: "https://example.com/hook".into(),
            attempt: 1,
            http_status: Some(200),
            response_body: Some("ok".into()),
            error_message: None,
            sent_at: 2,
        };
        storage.insert_webhook_log(&log).await.unwrap();
        storage
            .insert_webhook_log(&WebhookLog {
                attempt: 2,
                http_status: None,
                response_body: None,
                error_message: Some("timeout".into()),
                ..log.clone()
            })
            .await
            .unwrap();
        let rows = sqlx::query("SELECT attempt, http_status, error_message FROM webhook_log WHERE task_id = ? ORDER BY id")
            .bind(&t.id)
            .fetch_all(&storage.pool)
            .await
            .unwrap();
        assert_eq!(rows.len(), 2);
        assert_eq!(rows[0].get::<i64, _>("attempt"), 1);
        assert_eq!(rows[0].get::<Option<i64>, _>("http_status"), Some(200));
        assert_eq!(
            rows[1].get::<Option<String>, _>("error_message").as_deref(),
            Some("timeout")
        );
        let err = storage
            .insert_webhook_log(&WebhookLog {
                task_id: "missing".into(),
                ..log
            })
            .await
            .unwrap_err();
        assert_eq!(err.code, "system_database_error");
    }

    #[tokio::test]
    async fn meta_get_set_upserts() {
        let (_dir, storage) = open_tmp().await;
        assert_eq!(storage.get_meta("last_startup_at").await.unwrap(), None);
        storage.set_meta("last_startup_at", "1", 10).await.unwrap();
        assert_eq!(
            storage
                .get_meta("last_startup_at")
                .await
                .unwrap()
                .as_deref(),
            Some("1")
        );
        storage.set_meta("last_startup_at", "2", 20).await.unwrap();
        assert_eq!(
            storage
                .get_meta("last_startup_at")
                .await
                .unwrap()
                .as_deref(),
            Some("2")
        );
        let updated: i64 =
            sqlx::query_scalar("SELECT updated_at FROM metadata WHERE key = 'last_startup_at'")
                .fetch_one(&storage.pool)
                .await
                .unwrap();
        assert_eq!(updated, 20);
        let count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM metadata")
            .fetch_one(&storage.pool)
            .await
            .unwrap();
        assert_eq!(count, 1);
    }

    #[tokio::test]
    async fn corrupt_enum_value_maps_to_database_error() {
        let (_dir, storage) = open_tmp().await;
        let t = pending_task("01HQZ9AAAAAAAAAAAAAAAAAAAA", "张三", 1);
        storage.insert_task(&t).await.unwrap();
        sqlx::query("UPDATE tasks SET status = 'bogus' WHERE id = ?")
            .bind(&t.id)
            .execute(&storage.pool)
            .await
            .unwrap();
        let err = storage.get_task(&t.id).await.unwrap_err();
        assert_eq!(err.code, "system_database_error");
        assert_eq!(err.details.unwrap()["value"], "bogus");
    }
}
