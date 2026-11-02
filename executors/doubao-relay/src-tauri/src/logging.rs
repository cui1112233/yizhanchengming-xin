//! 日志初始化:stdout(人类可读)+ 按天滚动的 ndjson 文件 + 内存环形缓冲(供 `log.recent` / `log_entry` 事件)。
//!
//! 约定:message 中文、字段 key 英文;`info` 级别不得出现 bearer_token 与 prompt 内容(调用方负责)。

use std::collections::VecDeque;
use std::fmt;
use std::path::Path;
use std::sync::{Arc, Mutex, OnceLock};
use std::time::{SystemTime, UNIX_EPOCH};

use serde::{Deserialize, Serialize};
use serde_json::{Map, Value};
use tokio::sync::broadcast;
use tracing::field::{Field, Visit};
use tracing::{Event, Level, Subscriber};
use tracing_appender::non_blocking::WorkerGuard;
use tracing_appender::rolling::{RollingFileAppender, Rotation};
use tracing_subscriber::filter::LevelFilter;
use tracing_subscriber::layer::{Context, SubscriberExt};
use tracing_subscriber::registry::LookupSpan;
use tracing_subscriber::util::SubscriberInitExt;
use tracing_subscriber::{reload, EnvFilter, Layer, Registry};

/// 内存缓冲容量。
pub const BUFFER_CAPACITY: usize = 2000;
/// 广播通道容量(慢订阅者会 Lagged,UI 侧自行重新拉 `recent`)。
const BROADCAST_CAPACITY: usize = 1024;
/// 日志文件名前缀,实际文件为 `doubao-relay.YYYY-MM-DD.log`。
const FILE_PREFIX: &str = "doubao-relay";
/// 保留的日志文件个数(按天滚动 → 天数),与 spec 默认 `file_retention_days = 7` 一致。
const FILE_RETENTION_DAYS: usize = 7;

/// 文件日志的后台写线程句柄;[`flush`] 取出并 drop 它以写完缓冲(关闭流程最后一步)。
static GUARD: std::sync::Mutex<Option<WorkerGuard>> = std::sync::Mutex::new(None);
static BUFFER: OnceLock<LogBuffer> = OnceLock::new();
/// 过滤层的热更新句柄(`set_level`)。
static FILTER_HANDLE: OnceLock<reload::Handle<EnvFilter, Registry>> = OnceLock::new();

/// 一条日志(与 `docs/spec/05-tauri-ipc.md` `log.recent` 的 entry 同形)。
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LogEntry {
    /// Unix epoch 毫秒。
    pub ts: i64,
    /// `error` / `warn` / `info` / `debug` / `trace`。
    pub level: String,
    /// tracing target。
    pub target: String,
    pub message: String,
    /// 除 `message` 之外的结构化字段。
    pub fields: Value,
}

/// 固定容量的内存环形缓冲 + 广播。`Clone` 共享同一份数据。
#[derive(Clone)]
pub struct LogBuffer(Arc<Inner>);

struct Inner {
    entries: Mutex<VecDeque<LogEntry>>,
    tx: broadcast::Sender<LogEntry>,
}

impl Default for LogBuffer {
    fn default() -> Self {
        Self::new()
    }
}

impl LogBuffer {
    pub fn new() -> Self {
        let (tx, _rx) = broadcast::channel(BROADCAST_CAPACITY);
        Self(Arc::new(Inner {
            entries: Mutex::new(VecDeque::with_capacity(BUFFER_CAPACITY)),
            tx,
        }))
    }

    /// 写入一条(满了丢最旧),并广播给订阅者。
    pub fn push(&self, entry: LogEntry) {
        {
            let mut q = self.0.entries.lock().unwrap_or_else(|e| e.into_inner());
            if q.len() >= BUFFER_CAPACITY {
                q.pop_front();
            }
            q.push_back(entry.clone());
        }
        let _ = self.0.tx.send(entry);
    }

    /// 最近 `limit` 条(按时间先后排列),`min_level` 为最低级别(`Some("warn")` → 只含 warn/error)。
    pub fn recent(&self, limit: usize, min_level: Option<&str>) -> Vec<LogEntry> {
        let min = min_level.map(level_rank).unwrap_or(usize::MAX);
        let q = self.0.entries.lock().unwrap_or_else(|e| e.into_inner());
        let mut out: Vec<LogEntry> = q
            .iter()
            .rev()
            .filter(|e| level_rank(&e.level) <= min)
            .take(limit)
            .cloned()
            .collect();
        out.reverse();
        out
    }

    pub fn subscribe(&self) -> broadcast::Receiver<LogEntry> {
        self.0.tx.subscribe()
    }

    pub fn len(&self) -> usize {
        self.0
            .entries
            .lock()
            .unwrap_or_else(|e| e.into_inner())
            .len()
    }

    pub fn is_empty(&self) -> bool {
        self.len() == 0
    }
}

/// 级别严重度排名:error=0 … trace=4;未知级别视为最不严重。
fn level_rank(level: &str) -> usize {
    match level.to_ascii_lowercase().as_str() {
        "error" => 0,
        "warn" | "warning" => 1,
        "info" => 2,
        "debug" => 3,
        "trace" => 4,
        _ => usize::MAX,
    }
}

fn level_name(level: &Level) -> &'static str {
    match *level {
        Level::ERROR => "error",
        Level::WARN => "warn",
        Level::INFO => "info",
        Level::DEBUG => "debug",
        Level::TRACE => "trace",
    }
}

pub fn now_ms() -> i64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d| d.as_millis() as i64)
        .unwrap_or(0)
}

/// 把 tracing 事件写入 `LogBuffer` 的 Layer。
struct BufferLayer {
    buffer: LogBuffer,
}

struct FieldCollector {
    message: String,
    fields: Map<String, Value>,
}

impl Visit for FieldCollector {
    fn record_debug(&mut self, field: &Field, value: &dyn fmt::Debug) {
        if field.name() == "message" {
            self.message = format!("{value:?}");
        } else {
            self.fields.insert(
                field.name().to_string(),
                Value::String(format!("{value:?}")),
            );
        }
    }

    fn record_str(&mut self, field: &Field, value: &str) {
        if field.name() == "message" {
            self.message = value.to_string();
        } else {
            self.fields
                .insert(field.name().to_string(), Value::String(value.to_string()));
        }
    }

    fn record_i64(&mut self, field: &Field, value: i64) {
        self.fields
            .insert(field.name().to_string(), Value::from(value));
    }

    fn record_u64(&mut self, field: &Field, value: u64) {
        self.fields
            .insert(field.name().to_string(), Value::from(value));
    }

    fn record_f64(&mut self, field: &Field, value: f64) {
        self.fields
            .insert(field.name().to_string(), Value::from(value));
    }

    fn record_bool(&mut self, field: &Field, value: bool) {
        self.fields
            .insert(field.name().to_string(), Value::from(value));
    }

    fn record_error(&mut self, field: &Field, value: &(dyn std::error::Error + 'static)) {
        self.fields
            .insert(field.name().to_string(), Value::String(value.to_string()));
    }
}

impl<S> Layer<S> for BufferLayer
where
    S: Subscriber + for<'a> LookupSpan<'a>,
{
    fn on_event(&self, event: &Event<'_>, _ctx: Context<'_, S>) {
        let mut collector = FieldCollector {
            message: String::new(),
            fields: Map::new(),
        };
        event.record(&mut collector);
        let meta = event.metadata();
        self.buffer.push(LogEntry {
            ts: now_ms(),
            level: level_name(meta.level()).to_string(),
            target: meta.target().to_string(),
            message: collector.message,
            fields: Value::Object(collector.fields),
        });
    }
}

/// 初始化全局 tracing:stdout + `log_dir/doubao-relay.YYYY-MM-DD.log`(ndjson,按天滚动,保留 7 个)+ 内存缓冲。
///
/// `level` 为默认级别(`error`/`warn`/`info`/`debug`/`trace`),环境变量 `RUST_LOG` 存在时优先。
/// 重复调用(例如测试进程里)返回首次创建的同一个 `LogBuffer`。
pub fn init(level: &str, log_dir: &Path) -> anyhow::Result<LogBuffer> {
    if let Some(existing) = BUFFER.get() {
        return Ok(existing.clone());
    }

    std::fs::create_dir_all(log_dir)
        .map_err(|e| anyhow::anyhow!("创建日志目录失败 {}: {e}", log_dir.display()))?;

    let filter = EnvFilter::try_from_default_env()
        .or_else(|_| EnvFilter::try_new(level))
        .unwrap_or_else(|_| EnvFilter::new("info"));
    let (filter, filter_handle) = reload::Layer::new(filter);

    let file_appender = RollingFileAppender::builder()
        .rotation(Rotation::DAILY)
        .filename_prefix(FILE_PREFIX)
        .filename_suffix("log")
        .max_log_files(FILE_RETENTION_DAYS)
        .build(log_dir)
        .map_err(|e| anyhow::anyhow!("创建日志文件失败 {}: {e}", log_dir.display()))?;
    let (file_writer, guard) = tracing_appender::non_blocking(file_appender);

    let buffer = LogBuffer::new();

    let stdout_layer = tracing_subscriber::fmt::layer()
        .with_target(true)
        .with_writer(std::io::stdout);
    let file_layer = tracing_subscriber::fmt::layer()
        .json()
        .with_current_span(false)
        .with_span_list(false)
        .with_writer(file_writer);
    let buffer_layer = BufferLayer {
        buffer: buffer.clone(),
    };

    tracing_subscriber::registry()
        .with(filter)
        .with(stdout_layer)
        .with(file_layer)
        .with(buffer_layer)
        .try_init()
        .map_err(|e| anyhow::anyhow!("初始化日志失败: {e}"))?;

    *GUARD.lock().unwrap_or_else(|p| p.into_inner()) = Some(guard);
    let _ = FILTER_HANDLE.set(filter_handle);
    let buffer = BUFFER.get_or_init(|| buffer).clone();
    Ok(buffer)
}

/// 关闭流程最后一步:drop 文件日志的 `WorkerGuard`,阻塞到缓冲里的日志全部写盘。
/// 之后文件日志不再写入(stdout / 内存缓冲不受影响);重复调用无副作用。
pub fn flush() {
    let guard = GUARD.lock().unwrap_or_else(|p| p.into_inner()).take();
    drop(guard);
}

/// 热更新全局日志级别(设置页"日志级别"保存后立即生效,不必重启)。
///
/// `level` 必须是 `error`/`warn`/`info`/`debug`/`trace`;会覆盖启动时 `RUST_LOG` 的设置。
/// 日志尚未 `init`(测试进程)时只校验级别、不做别的。
pub fn set_level(level: &str) -> anyhow::Result<()> {
    let parsed: LevelFilter = level
        .trim()
        .parse()
        .map_err(|_| anyhow::anyhow!("日志级别不合法:{level}"))?;
    let Some(handle) = FILTER_HANDLE.get() else {
        return Ok(());
    };
    handle
        .reload(EnvFilter::new(parsed.to_string().to_ascii_lowercase()))
        .map_err(|e| anyhow::anyhow!("热更新日志级别失败:{e}"))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn entry(level: &str, msg: &str) -> LogEntry {
        LogEntry {
            ts: now_ms(),
            level: level.into(),
            target: "t".into(),
            message: msg.into(),
            fields: Value::Object(Map::new()),
        }
    }

    #[test]
    fn recent_returns_latest_in_chronological_order() {
        let buf = LogBuffer::new();
        for i in 0..5 {
            buf.push(entry("info", &format!("m{i}")));
        }
        let got = buf.recent(3, None);
        let msgs: Vec<&str> = got.iter().map(|e| e.message.as_str()).collect();
        assert_eq!(msgs, ["m2", "m3", "m4"]);
    }

    #[test]
    fn recent_filters_by_min_level() {
        let buf = LogBuffer::new();
        buf.push(entry("debug", "d"));
        buf.push(entry("info", "i"));
        buf.push(entry("warn", "w"));
        buf.push(entry("error", "e"));
        let warn_up: Vec<String> = buf
            .recent(10, Some("warn"))
            .into_iter()
            .map(|e| e.message)
            .collect();
        assert_eq!(warn_up, ["w", "e"]);
        let info_up: Vec<String> = buf
            .recent(10, Some("info"))
            .into_iter()
            .map(|e| e.message)
            .collect();
        assert_eq!(info_up, ["i", "w", "e"]);
        assert_eq!(buf.recent(10, Some("debug")).len(), 4);
        assert_eq!(buf.recent(10, None).len(), 4);
    }

    #[test]
    fn buffer_is_bounded() {
        let buf = LogBuffer::new();
        for i in 0..(BUFFER_CAPACITY + 10) {
            buf.push(entry("info", &i.to_string()));
        }
        assert_eq!(buf.len(), BUFFER_CAPACITY);
        let first = buf.recent(BUFFER_CAPACITY, None).remove(0);
        assert_eq!(first.message, "10");
    }

    #[test]
    fn subscribe_receives_pushed_entries() {
        let buf = LogBuffer::new();
        let mut rx = buf.subscribe();
        buf.push(entry("info", "hello"));
        let got = rx.try_recv().unwrap();
        assert_eq!(got.message, "hello");
    }

    #[test]
    fn log_entry_serializes_with_spec_keys() {
        let e = entry("info", "x");
        let v = serde_json::to_value(&e).unwrap();
        let obj = v.as_object().unwrap();
        for k in ["ts", "level", "target", "message", "fields"] {
            assert!(obj.contains_key(k), "missing {k}");
        }
    }

    #[test]
    fn init_captures_tracing_events_into_buffer() {
        let dir = tempfile::tempdir().unwrap();
        let buf = init("debug", dir.path()).unwrap();
        // CI/开发机可能预设 RUST_LOG=warn；测试要验证 debug/info 层本身，
        // 初始化后显式恢复到 debug，避免环境变量污染断言。
        set_level("debug").unwrap();
        let mut rx = buf.subscribe();
        tracing::info!(target: "t", "x");
        tracing::warn!(target: "t", task_id = "01ABC", count = 3u64, "带字段的日志");
        let recent = buf.recent(10, None);
        assert!(
            recent
                .iter()
                .any(|e| e.message == "x" && e.target == "t" && e.level == "info"),
            "recent 应包含 message x: {recent:?}"
        );
        let with_fields = recent.iter().find(|e| e.message == "带字段的日志").unwrap();
        assert_eq!(with_fields.level, "warn");
        assert_eq!(with_fields.fields["task_id"], "01ABC");
        assert_eq!(with_fields.fields["count"], 3);
        assert!(with_fields.ts > 0);
        // 广播同样收到
        let mut received_x = false;
        while let Ok(entry) = rx.try_recv() {
            if entry.message == "x" && entry.target == "t" {
                received_x = true;
                break;
            }
        }
        assert!(received_x, "广播应收到本测试写入的 x");
        // 再次 init 得到同一个缓冲
        let again = init("info", dir.path()).unwrap();
        assert!(Arc::ptr_eq(&again.0, &buf.0));
        // 热更新级别:warn 之后 info 不再进缓冲,warn 仍进;改回 debug 后恢复
        set_level("warn").unwrap();
        tracing::info!(target: "t", "热更新后的 info");
        tracing::warn!(target: "t", "热更新后的 warn");
        let recent = buf.recent(10, None);
        assert!(!recent.iter().any(|e| e.message == "热更新后的 info"));
        assert!(recent.iter().any(|e| e.message == "热更新后的 warn"));
        set_level("debug").unwrap();
        tracing::debug!(target: "t", "改回 debug");
        assert!(buf
            .recent(10, None)
            .iter()
            .any(|e| e.message == "改回 debug"));
        assert!(
            set_level("verbose").is_err(),
            "非法级别应报错且不改当前过滤"
        );
        // 收尾:别让本进程里其它测试(sqlx 等)的 debug 日志刷屏
        set_level("info").unwrap();
        // 日志目录下出现了按天滚动的文件
        let names: Vec<String> = std::fs::read_dir(dir.path())
            .unwrap()
            .map(|e| e.unwrap().file_name().to_string_lossy().into_owned())
            .collect();
        assert!(
            names
                .iter()
                .any(|n| n.starts_with("doubao-relay.") && n.ends_with(".log")),
            "log files: {names:?}"
        );
    }
}
