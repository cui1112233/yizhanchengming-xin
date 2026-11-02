//! 定时清理(07-data-model "定时 GC",每 1h 由后台任务调用,也供 `maintenance.gc()` 手动触发)。
//!
//! 1. `tasks.expires_at < now` 且 `result_files` 非空 → 删结果文件(整个 `results/{id}/`),`result_files` 置 NULL;
//! 2. `tasks.created_at < now - metadata_retention_days` → 删任务行(级联 `task_stage_history` / `webhook_log`)及其结果目录;
//! 3. `uploads.expires_at < now` → 删物理文件与行;
//! 4. 日志文件按天滚动由 `tracing-appender` 自己保留 N 天,这里不处理。
//!
//! 文件删除是 best-effort:单个文件删不掉只记 warn,不中断 GC;数据库错误才返回 Err。

use std::path::Path;

use serde::Serialize;
use sqlx::Row;

use super::db::Storage;
use super::files;
use crate::config::StorageConfig;
use crate::error::RelayResult;
use crate::tasks::ResultFile;

/// 一次 GC 的统计。
#[derive(Debug, Clone, Copy, Default, PartialEq, Eq, Serialize)]
pub struct GcReport {
    /// 删掉的结果文件数(步骤 1 + 步骤 2 中级联删掉的)。
    pub result_files_removed: u64,
    /// 删掉的任务行数(步骤 2)。
    pub tasks_removed: u64,
    /// 删掉的上传记录数(步骤 3)。
    pub uploads_removed: u64,
}

const MS_PER_DAY: i64 = 86_400_000;

impl Storage {
    /// 执行四步 GC(见模块文档)。
    pub async fn gc(&self, now: i64, cfg: &StorageConfig) -> RelayResult<GcReport> {
        let mut report = GcReport::default();

        // 1. 过期结果文件
        let rows = sqlx::query(
            "SELECT id, result_files FROM tasks WHERE expires_at < ? AND result_files IS NOT NULL",
        )
        .bind(now)
        .fetch_all(&self.pool)
        .await?;
        for row in &rows {
            let id: String = row.try_get("id")?;
            let json: String = row.try_get("result_files")?;
            report.result_files_removed += self.remove_result_files(&id, &json);
            sqlx::query("UPDATE tasks SET result_files = NULL WHERE id = ?")
                .bind(&id)
                .execute(&self.pool)
                .await?;
        }

        // 2. 超过元数据保留期的任务(级联 stage_history / webhook_log)
        let cutoff = now - i64::from(cfg.metadata_retention_days) * MS_PER_DAY;
        let rows = sqlx::query("SELECT id, result_files FROM tasks WHERE created_at < ?")
            .bind(cutoff)
            .fetch_all(&self.pool)
            .await?;
        for row in &rows {
            let id: String = row.try_get("id")?;
            let json: Option<String> = row.try_get("result_files")?;
            report.result_files_removed +=
                self.remove_result_files(&id, json.as_deref().unwrap_or("[]"));
            let res = sqlx::query("DELETE FROM tasks WHERE id = ?")
                .bind(&id)
                .execute(&self.pool)
                .await?;
            report.tasks_removed += res.rows_affected();
        }

        // 3. 过期上传
        let rows = sqlx::query("SELECT id, path FROM uploads WHERE expires_at < ?")
            .bind(now)
            .fetch_all(&self.pool)
            .await?;
        for row in &rows {
            let id: String = row.try_get("id")?;
            let path: String = row.try_get("path")?;
            if let Err(e) = files::remove_file_if_exists(Path::new(&path)) {
                tracing::warn!(upload_id = %id, path = %path, error = %e, "删除过期上传文件失败");
            }
            let res = sqlx::query("DELETE FROM uploads WHERE id = ?")
                .bind(&id)
                .execute(&self.pool)
                .await?;
            report.uploads_removed += res.rows_affected();
        }

        tracing::info!(
            result_files_removed = report.result_files_removed,
            tasks_removed = report.tasks_removed,
            uploads_removed = report.uploads_removed,
            "存储清理完成"
        );
        Ok(report)
    }

    /// 删掉 `result_files` 里列出的文件(限 root 下的安全相对路径)和整个 `results/{task_id}/` 目录。
    /// 返回实际删掉的文件数。
    fn remove_result_files(&self, task_id: &str, result_files_json: &str) -> u64 {
        let mut removed = 0;
        let listed: Vec<ResultFile> = match serde_json::from_str(result_files_json) {
            Ok(v) => v,
            Err(e) => {
                tracing::warn!(task_id, error = %e, "result_files 不是合法 JSON,只按目录清理");
                Vec::new()
            }
        };
        for f in &listed {
            let rel = Path::new(&f.relative_path);
            if !files::is_safe_relative(rel) {
                tracing::warn!(task_id, path = %f.relative_path, "result_files 路径不合法,跳过");
                continue;
            }
            match files::remove_file_if_exists(&self.root.join(rel)) {
                Ok(true) => removed += 1,
                Ok(false) => {}
                Err(e) => {
                    tracing::warn!(task_id, path = %f.relative_path, error = %e, "删除结果文件失败")
                }
            }
        }
        match files::remove_dir_if_exists(&self.result_dir(task_id)) {
            Ok(n) => removed += n,
            Err(e) => tracing::warn!(task_id, error = %e, "删除结果目录失败"),
        }
        removed
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::storage::db::tests::{open_tmp, pending_task};
    use crate::storage::{Upload, WebhookLog};
    use crate::tasks::{Task, TaskStatus};

    const NOW: i64 = 1_700_000_000_000;

    /// 已完成任务,结果文件真的写在 results/{id}/ 下。
    fn completed_with_files(storage: &Storage, id: &str, created_at: i64, expires_at: i64) -> Task {
        let dir = storage.result_dir(id);
        std::fs::create_dir_all(&dir).unwrap();
        std::fs::write(dir.join("image_1.png"), b"png1").unwrap();
        std::fs::write(dir.join("image_2.png"), b"png2").unwrap();
        let mut t = pending_task(id, "张三", created_at);
        t.status = TaskStatus::Completed;
        t.completed_at = Some(created_at + 1);
        t.expires_at = expires_at;
        t.result_files = Some(
            (0..2u32)
                .map(|i| ResultFile {
                    index: i,
                    filename: format!("image_{}.png", i + 1),
                    relative_path: format!("results/{id}/image_{}.png", i + 1),
                    mime: "image/png".into(),
                    bytes: 4,
                    width: Some(1),
                    height: Some(1),
                    duration_ms: None,
                    source_url: None,
                })
                .collect(),
        );
        t
    }

    async fn count(storage: &Storage, sql: &str) -> i64 {
        sqlx::query_scalar(sql)
            .fetch_one(&storage.pool)
            .await
            .unwrap()
    }

    #[tokio::test]
    async fn expired_result_files_are_deleted_and_nulled_but_row_kept() {
        let (_dir, storage) = open_tmp().await;
        let expired =
            completed_with_files(&storage, "01HQZ9AAAAAAAAAAAAAAAAAAA1", NOW - 10, NOW - 1);
        let fresh = completed_with_files(&storage, "01HQZ9AAAAAAAAAAAAAAAAAAA2", NOW - 10, NOW + 1);
        storage.insert_task(&expired).await.unwrap();
        storage.insert_task(&fresh).await.unwrap();

        let report = storage.gc(NOW, &StorageConfig::default()).await.unwrap();
        assert_eq!(
            report,
            GcReport {
                result_files_removed: 2,
                tasks_removed: 0,
                uploads_removed: 0
            }
        );
        assert!(!storage.result_dir(&expired.id).exists());
        let got = storage.get_task(&expired.id).await.unwrap().unwrap();
        assert_eq!(got.result_files, None);
        assert_eq!(got.status, TaskStatus::Completed);
        assert_eq!(got.stage_history.len(), 1, "元数据保留");

        assert!(storage.result_dir(&fresh.id).join("image_2.png").is_file());
        let got = storage.get_task(&fresh.id).await.unwrap().unwrap();
        assert_eq!(got.result_files.unwrap().len(), 2);

        // 第二次 GC 无事可做
        let report = storage.gc(NOW, &StorageConfig::default()).await.unwrap();
        assert_eq!(report, GcReport::default());
    }

    #[tokio::test]
    async fn tasks_older_than_retention_cascade_delete_history_webhooks_and_files() {
        let (_dir, storage) = open_tmp().await;
        let cfg = StorageConfig::default(); // 7 天
        let old_created = NOW - 8 * MS_PER_DAY;
        let old =
            completed_with_files(&storage, "01HQZ9AAAAAAAAAAAAAAAAAAA1", old_created, NOW + 1);
        let recent = completed_with_files(
            &storage,
            "01HQZ9AAAAAAAAAAAAAAAAAAA2",
            NOW - 6 * MS_PER_DAY,
            NOW + 1,
        );
        storage.insert_task(&old).await.unwrap();
        storage.insert_task(&recent).await.unwrap();
        storage
            .append_stage(&old.id, "result", None, old_created + 5)
            .await
            .unwrap();
        for id in [&old.id, &recent.id] {
            storage
                .insert_webhook_log(&WebhookLog {
                    task_id: id.clone(),
                    event: "task.completed".into(),
                    url: "https://example.com/h".into(),
                    attempt: 1,
                    http_status: Some(200),
                    response_body: None,
                    error_message: None,
                    sent_at: NOW,
                })
                .await
                .unwrap();
        }
        assert_eq!(
            count(&storage, "SELECT COUNT(*) FROM task_stage_history").await,
            3
        );
        assert_eq!(count(&storage, "SELECT COUNT(*) FROM webhook_log").await, 2);

        let report = storage.gc(NOW, &cfg).await.unwrap();
        assert_eq!(report.tasks_removed, 1);
        assert_eq!(report.result_files_removed, 2, "级联删掉的结果文件也计数");
        assert_eq!(report.uploads_removed, 0);
        assert_eq!(storage.get_task(&old.id).await.unwrap(), None);
        assert!(!storage.result_dir(&old.id).exists());
        assert_eq!(
            count(&storage, "SELECT COUNT(*) FROM task_stage_history").await,
            1
        );
        assert_eq!(count(&storage, "SELECT COUNT(*) FROM webhook_log").await, 1);
        assert!(storage.get_task(&recent.id).await.unwrap().is_some());
        assert!(storage.result_dir(&recent.id).join("image_1.png").is_file());

        // 缩短保留期后 recent 也被删;没有结果目录的任务也能删
        let mut short = cfg.clone();
        short.metadata_retention_days = 1;
        let pending = pending_task("01HQZ9AAAAAAAAAAAAAAAAAAA3", "李四", NOW - 2 * MS_PER_DAY);
        storage.insert_task(&pending).await.unwrap();
        let report = storage.gc(NOW, &short).await.unwrap();
        assert_eq!(report.tasks_removed, 2);
        assert_eq!(count(&storage, "SELECT COUNT(*) FROM tasks").await, 0);
    }

    #[tokio::test]
    async fn expired_uploads_remove_file_and_row_tolerating_missing_file() {
        let (_dir, storage) = open_tmp().await;
        let make = |id: &str, expires_at: i64, write: bool| {
            let path = storage.upload_path(id, "png");
            if write {
                std::fs::write(&path, b"x").unwrap();
            }
            Upload {
                id: id.into(),
                mime: "image/png".into(),
                bytes: 1,
                path,
                created_at: NOW - 10,
                expires_at,
            }
        };
        let expired = make("01HQZ8AAAAAAAAAAAAAAAAAAA1", NOW - 1, true);
        let expired_missing_file = make("01HQZ8AAAAAAAAAAAAAAAAAAA2", NOW - 1, false);
        let fresh = make("01HQZ8AAAAAAAAAAAAAAAAAAA3", NOW + 1, true);
        for u in [&expired, &expired_missing_file, &fresh] {
            storage.insert_upload(u).await.unwrap();
        }

        let report = storage.gc(NOW, &StorageConfig::default()).await.unwrap();
        assert_eq!(report.uploads_removed, 2);
        assert_eq!(report.tasks_removed, 0);
        assert!(!expired.path.exists());
        assert_eq!(storage.get_upload(&expired.id).await.unwrap(), None);
        assert_eq!(
            storage.get_upload(&expired_missing_file.id).await.unwrap(),
            None
        );
        assert!(fresh.path.is_file());
        assert!(storage.get_upload(&fresh.id).await.unwrap().is_some());
    }

    #[tokio::test]
    async fn unsafe_relative_path_is_skipped_but_dir_still_cleaned() {
        let (dir, storage) = open_tmp().await;
        let outside = dir.path().join("outside.txt");
        std::fs::write(&outside, b"keep").unwrap();
        let mut t = completed_with_files(&storage, "01HQZ9AAAAAAAAAAAAAAAAAAA1", NOW - 10, NOW - 1);
        t.result_files.as_mut().unwrap()[0].relative_path = "../outside.txt".into();
        storage.insert_task(&t).await.unwrap();
        let report = storage.gc(NOW, &StorageConfig::default()).await.unwrap();
        assert!(outside.is_file(), "root 之外的文件不能被 GC 碰");
        assert!(!storage.result_dir(&t.id).exists());
        assert_eq!(report.result_files_removed, 2);
        assert_eq!(
            storage.get_task(&t.id).await.unwrap().unwrap().result_files,
            None
        );
    }

    #[test]
    fn gc_report_serializes_snake_case() {
        let v = serde_json::to_value(GcReport {
            result_files_removed: 1,
            tasks_removed: 2,
            uploads_removed: 3,
        })
        .unwrap();
        assert_eq!(v["result_files_removed"], 1);
        assert_eq!(v["tasks_removed"], 2);
        assert_eq!(v["uploads_removed"], 3);
    }
}
