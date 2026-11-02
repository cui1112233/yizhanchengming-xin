//! `log.recent` / `log.export` / `debug.cdp_log` 命令。

use std::path::Path;
use std::sync::Arc;

use serde_json::{json, Value};
use tauri::State;

use crate::error::{RelayError, RelayResult};
use crate::state::AppState;

use super::respond;

/// spec 05:`limit` 默认 200。
pub const DEFAULT_LIMIT: usize = 200;
/// 单次最多返回的条数(内存缓冲本身就是 2000)。
pub const MAX_LIMIT: usize = 2000;

fn clamp_limit(limit: Option<usize>) -> usize {
    limit.unwrap_or(DEFAULT_LIMIT).clamp(1, MAX_LIMIT)
}

/// `log.recent` → `{entries}`;`level` 为最低级别(`warn` → 只含 warn/error)。
pub fn recent(state: &AppState, limit: Option<usize>, level: Option<&str>) -> RelayResult<Value> {
    let level = level.map(str::trim).filter(|l| !l.is_empty());
    let entries = state.logs.recent(clamp_limit(limit), level);
    respond(&json!({ "entries": entries }))
}

/// `log.export`:把内存缓冲里的日志(最低级别 `level`,最多 [`MAX_LIMIT`] 条)按 ndjson 写到 `path`
/// (日志页"导出",路径来自 `tauri-plugin-dialog` 的 save 对话框)。返回 `{ok, path, count}`。
/// 文件内容是 domain 形态(字段 key 保持 `snake_case`),与 `doubao-relay.*.log` 一致。
pub fn export(state: &AppState, path: &str, level: Option<&str>) -> RelayResult<Value> {
    let path = path.trim();
    if path.is_empty() {
        return Err(RelayError::new(
            "request_malformed",
            "导出路径为空,请在保存对话框里选择文件位置",
        ));
    }
    let level = level.map(str::trim).filter(|l| !l.is_empty());
    let entries = state.logs.recent(MAX_LIMIT, level);
    let mut out = String::with_capacity(entries.len() * 160);
    for e in &entries {
        let line = serde_json::to_string(e)
            .map_err(|e| RelayError::internal(format!("序列化日志失败:{e}")))?;
        out.push_str(&line);
        out.push('\n');
    }
    std::fs::write(Path::new(path), out).map_err(|e| {
        let code = if e.kind() == std::io::ErrorKind::StorageFull {
            "system_disk_full"
        } else {
            "system_io_error"
        };
        RelayError::new(code, "写入日志导出文件失败,请换一个目录重试")
            .with_details(json!({ "path": path, "io": e.to_string() }))
    })?;
    tracing::info!(path, count = entries.len(), "日志已导出");
    Ok(json!({ "ok": true, "path": path, "count": entries.len() }))
}

/// `debug.cdp_log` → `{entries, enabled}`;未开启 `cdp_debug_log` 时 `entries` 为空。
pub fn cdp_log(state: &AppState, limit: Option<usize>) -> RelayResult<Value> {
    let entries = state.cdp_log.recent(clamp_limit(limit));
    respond(&json!({ "entries": entries, "enabled": state.cdp_log.enabled() }))
}

#[tauri::command]
pub async fn log_recent(
    state: State<'_, Arc<AppState>>,
    limit: Option<usize>,
    level: Option<String>,
) -> RelayResult<Value> {
    recent(&state, limit, level.as_deref())
}

#[tauri::command]
pub async fn log_export(
    state: State<'_, Arc<AppState>>,
    path: String,
    level: Option<String>,
) -> RelayResult<Value> {
    export(&state, &path, level.as_deref())
}

#[tauri::command]
pub async fn debug_cdp_log(
    state: State<'_, Arc<AppState>>,
    limit: Option<usize>,
) -> RelayResult<Value> {
    cdp_log(&state, limit)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::cdp::CdpLogEntry;
    use crate::commands::testing::test_state_blocking;
    use crate::logging::{now_ms, LogEntry};

    fn entry(level: &str, msg: &str) -> LogEntry {
        LogEntry {
            ts: now_ms(),
            level: level.into(),
            target: "t".into(),
            message: msg.into(),
            fields: json!({ "task_id": "01A" }),
        }
    }

    #[test]
    fn recent_filters_by_level_and_limit_and_is_camel() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state_blocking(dir.path());
        state.logs.push(entry("info", "i1"));
        state.logs.push(entry("warn", "w1"));
        state.logs.push(entry("error", "e1"));
        state.logs.push(entry("info", "i2"));

        let v = recent(&state, None, None).unwrap();
        assert_eq!(v["entries"].as_array().unwrap().len(), 4);
        assert_eq!(v["entries"][0]["fields"]["taskId"], "01A");

        let v = recent(&state, Some(1), None).unwrap();
        assert_eq!(v["entries"][0]["message"], "i2");

        let v = recent(&state, None, Some("warn")).unwrap();
        let msgs: Vec<&str> = v["entries"]
            .as_array()
            .unwrap()
            .iter()
            .map(|e| e["message"].as_str().unwrap())
            .collect();
        assert_eq!(msgs, ["w1", "e1"]);

        let v = recent(&state, Some(0), Some("")).unwrap();
        assert_eq!(
            v["entries"].as_array().unwrap().len(),
            1,
            "limit 0 夹到 1,空 level 不过滤"
        );
    }

    #[test]
    fn export_writes_ndjson_filtered_by_level() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state_blocking(dir.path());
        state.logs.push(entry("info", "i1"));
        state.logs.push(entry("warn", "w1"));
        state.logs.push(entry("error", "e1"));

        let path = dir.path().join("export").join("logs.ndjson");
        let v = export(&state, path.to_str().unwrap(), Some("warn")).unwrap_err();
        assert_eq!(v.code, "system_io_error", "父目录不存在 → IO 错误");

        std::fs::create_dir_all(path.parent().unwrap()).unwrap();
        let v = export(&state, path.to_str().unwrap(), Some("warn")).unwrap();
        assert_eq!(v["ok"], true);
        assert_eq!(v["count"], 2);
        assert_eq!(v["path"], path.to_str().unwrap());
        let text = std::fs::read_to_string(&path).unwrap();
        let lines: Vec<serde_json::Value> = text
            .lines()
            .map(|l| serde_json::from_str(l).unwrap())
            .collect();
        assert_eq!(lines.len(), 2);
        assert_eq!(lines[0]["message"], "w1");
        assert_eq!(lines[1]["level"], "error");
        assert_eq!(lines[0]["fields"]["task_id"], "01A", "文件保持 snake_case");
        assert!(text.ends_with('\n'));

        let v = export(&state, path.to_str().unwrap(), None).unwrap();
        assert_eq!(v["count"], 3, "不过滤 → 全部;重复导出覆盖");
        assert_eq!(std::fs::read_to_string(&path).unwrap().lines().count(), 3);

        let e = export(&state, "   ", None).unwrap_err();
        assert_eq!(e.code, "request_malformed");
    }

    #[test]
    fn cdp_log_returns_buffered_entries_and_enabled_flag() {
        let dir = tempfile::tempdir().unwrap();
        let state = test_state_blocking(dir.path());
        state.cdp_log.push(CdpLogEntry {
            ts: 1,
            direction: "send".into(),
            method: Some("Target.getTargets".into()),
            params: Some(json!({})),
            result: None,
            event: None,
        });
        let v = cdp_log(&state, None).unwrap();
        assert_eq!(v["enabled"], true);
        assert_eq!(v["entries"][0]["method"], "Target.getTargets");
        assert_eq!(v["entries"][0]["direction"], "send");
        assert!(v["entries"][0].get("result").is_none());
    }
}
