//! 持久化层:SQLite(`db`)+ 结果/上传文件布局(`files`)+ 定时清理(`gc`)。
//!
//! 对应 `docs/spec/07-data-model.md`。所有 SQL 用 `sqlx::query()` 运行时绑定,
//! 不用 `query!` 宏(避免编译期 `DATABASE_URL`)。

pub mod db;
pub mod files;
pub mod gc;

pub use db::{Storage, Upload, WebhookLog};
pub use gc::GcReport;
