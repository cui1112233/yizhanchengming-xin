//! Tauri IPC 命令(`docs/spec/05-tauri-ipc.md`)。
//!
//! 约定(计划"共享约定"):
//! - spec 的 `account.list` 在 Rust 侧叫 `account_list`(标识符不能含 `.`);
//! - 命令全部 `async`,签名 `State<'_, Arc<AppState>>`,返回 `Result<Value, RelayError>`;
//! - domain 类型是 `snake_case`,出 IPC 前经 [`to_camel`] 递归改 key;
//!   WebView 传进来的 `patch` 经 [`to_snake`] 还原后再反序列化。
//!
//! 每个命令都拆成"纯逻辑函数(`&AppState`)+ `#[tauri::command]` 薄壳",逻辑函数可直接测试。
//! `generate_handler!` 必须用 `commands::<模块>::<命令>` 的完整路径(宏生成的伴生项不随 `pub use` 导出)。

pub mod account;
pub mod app;
pub mod doubao;
pub mod log;
pub mod privacy; // Task 5.2
pub mod task;
pub mod upload;

use serde::Serialize;
use serde_json::{Map, Value};

use crate::error::{RelayError, RelayResult};

/// `snake_case` key → `camelCase`;递归处理对象与数组,值不动。
/// 已是 camelCase / 无下划线 / 以下划线开头的 key 保持原样。
pub fn to_camel(value: Value) -> Value {
    map_keys(value, camel_key)
}

/// `camelCase` key → `snake_case`(WebView 传入的 patch 还原成 domain 命名);递归。
pub fn to_snake(value: Value) -> Value {
    map_keys(value, snake_key)
}

fn map_keys(value: Value, f: fn(&str) -> String) -> Value {
    match value {
        Value::Object(obj) => {
            let mut out = Map::with_capacity(obj.len());
            for (k, v) in obj {
                out.insert(f(&k), map_keys(v, f));
            }
            Value::Object(out)
        }
        Value::Array(items) => Value::Array(items.into_iter().map(|v| map_keys(v, f)).collect()),
        other => other,
    }
}

fn camel_key(key: &str) -> String {
    if !key.contains('_') || key.starts_with('_') {
        return key.to_string();
    }
    let mut out = String::with_capacity(key.len());
    let mut upper_next = false;
    for c in key.chars() {
        if c == '_' {
            upper_next = true;
        } else if upper_next {
            out.extend(c.to_uppercase());
            upper_next = false;
        } else {
            out.push(c);
        }
    }
    out
}

fn snake_key(key: &str) -> String {
    let mut out = String::with_capacity(key.len() + 4);
    for (i, c) in key.chars().enumerate() {
        if c.is_ascii_uppercase() {
            if i > 0 {
                out.push('_');
            }
            out.push(c.to_ascii_lowercase());
        } else {
            out.push(c);
        }
    }
    out
}

/// 把 domain 值序列化并转成 camelCase,作为命令返回值。
pub fn respond<T: Serialize>(value: &T) -> RelayResult<Value> {
    serde_json::to_value(value)
        .map(to_camel)
        .map_err(|e| RelayError::internal(format!("序列化返回值失败:{e}")))
}

#[cfg(test)]
pub(crate) mod testing {
    //! 命令测试共用的假平台与 `AppState` 构造。

    use std::path::{Path, PathBuf};
    use std::sync::Arc;

    use async_trait::async_trait;
    use tokio::sync::{mpsc, watch};
    use tokio_util::sync::CancellationToken;

    use crate::config::Config;
    use crate::doubao::capture::{CaptureEvent, CaptureOptions, CaptureOutcome};
    use crate::doubao::download::Downloaded;
    use crate::doubao::generation::{self, StageFn, SubmissionPlan, SubmissionReport};
    use crate::doubao::{Account, AccountList, DoubaoConnection};
    use crate::error::{RelayError, RelayResult};
    use crate::logging::{now_ms, LogBuffer};
    use crate::platform::{Platform, PlatformError};
    use crate::state::{AppState, CdpLogBuffer};
    use crate::storage::Storage;
    use crate::tasks::{GenerateRequest, GenerationBackend};

    pub struct FakePlatform {
        pub data_dir: PathBuf,
    }

    #[async_trait]
    impl Platform for FakePlatform {
        async fn find_doubao_exe(&self) -> Result<PathBuf, PlatformError> {
            Err(PlatformError::DoubaoNotFound("/nowhere".into()))
        }
        async fn launch_doubao_with_cdp(&self, _: &Path, _: u16) -> Result<(), PlatformError> {
            Ok(())
        }
        async fn kill_doubao(&self) -> Result<(), PlatformError> {
            Ok(())
        }
        async fn is_doubao_running(&self) -> bool {
            false
        }
        fn app_data_dir(&self) -> PathBuf {
            self.data_dir.clone()
        }
        fn log_dir(&self) -> PathBuf {
            self.data_dir.join("logs")
        }
        fn config_path(&self) -> PathBuf {
            self.data_dir.join("config.toml")
        }
        async fn reveal_in_file_manager(&self, _: &Path) -> Result<(), PlatformError> {
            Ok(())
        }
        fn platform_name(&self) -> &'static str {
            "fake"
        }
        fn listening_process_exe(&self, _: u16) -> Option<PathBuf> {
            None
        }
        fn is_doubao_exe_path(&self, _: &Path) -> bool {
            false
        }
        fn doubao_version(&self) -> Option<String> {
            None
        }
    }

    /// 假的豆包后端:固定账号列表(首个为当前账号),其余操作要么挂起到取消、要么直接报错。
    /// 调度循环在测试里不启动,所以任务停在 `pending`,只用来验证提交 / 查询 / 取消等接口。
    pub struct FakeBackend {
        pub accounts: Vec<String>,
    }

    impl FakeBackend {
        pub fn new(accounts: &[&str]) -> Arc<Self> {
            Arc::new(Self {
                accounts: accounts.iter().map(|s| s.to_string()).collect(),
            })
        }
    }

    #[async_trait]
    impl GenerationBackend for FakeBackend {
        async fn accounts(&self, _force: bool) -> RelayResult<AccountList> {
            if self.accounts.is_empty() {
                return Err(RelayError::doubao_not_logged_in());
            }
            Ok(AccountList {
                current: self.accounts.first().cloned(),
                accounts: self
                    .accounts
                    .iter()
                    .enumerate()
                    .map(|(i, n)| Account {
                        name: n.clone(),
                        subscription: Some("体验套餐".into()),
                        label: format!("{n} 体验套餐"),
                        is_current: i == 0,
                        is_ambiguous: false,
                    })
                    .collect(),
                read_at: now_ms(),
            })
        }

        async fn ensure_account(&self, account: &str) -> RelayResult<()> {
            if self.accounts.iter().any(|a| a == account) {
                Ok(())
            } else {
                Err(RelayError::account_not_found(
                    account,
                    self.accounts.clone(),
                ))
            }
        }

        async fn submit(
            &self,
            _request: &GenerateRequest,
            _plan: &SubmissionPlan,
            _reference_paths: &[PathBuf],
            cancel: &CancellationToken,
            _on_stage: StageFn<'_>,
        ) -> RelayResult<SubmissionReport> {
            cancel.cancelled().await;
            Err(generation::cancelled("prepare"))
        }

        async fn capture(
            &self,
            _opts: CaptureOptions,
            cancel: &CancellationToken,
            _confirm_rx: &mut watch::Receiver<bool>,
            _events: mpsc::Sender<CaptureEvent>,
        ) -> RelayResult<CaptureOutcome> {
            cancel.cancelled().await;
            Err(generation::cancelled("capture"))
        }

        async fn download_image(
            &self,
            _url: &str,
            _dest_dir: &Path,
            _index: u32,
        ) -> RelayResult<Downloaded> {
            Err(RelayError::internal("FakeBackend 不下载"))
        }

        async fn download_video(&self, _url: &str, _dest_dir: &Path) -> RelayResult<Downloaded> {
            Err(RelayError::internal("FakeBackend 不下载"))
        }
    }

    /// 临时目录里的 `AppState`:配置已写盘、豆包连接指向一个没人监听的端口、CDP 日志开启,
    /// `Storage` 开在 `dir` 下,调度器后端是真豆包连接(`DoubaoBackend`,任何豆包操作都报 `doubao_*`)。
    pub async fn test_state(dir: &Path) -> Arc<AppState> {
        test_state_with(dir, None).await
    }

    /// 同 [`test_state`],但调度器后端可换成 [`FakeBackend`](豆包连接仍指向空端口)。
    pub async fn test_state_with(
        dir: &Path,
        backend: Option<Arc<dyn GenerationBackend>>,
    ) -> Arc<AppState> {
        let platform: Arc<dyn Platform> = Arc::new(FakePlatform {
            data_dir: dir.to_path_buf(),
        });
        let config_path = platform.config_path();
        let mut config = Config::load_or_create(&config_path).unwrap();
        config.logging.cdp_debug_log = true;
        config.doubao.cdp_port = 1; // 保证连不上
        config.save(&config_path).unwrap();
        let cdp_log = CdpLogBuffer::new(true);
        let doubao = DoubaoConnection::new_with(
            Arc::clone(&platform),
            config.doubao.cdp_address.clone(),
            config.doubao.cdp_port,
            cdp_log.sender(),
            crate::doubao::ConnTiming::default(),
        );
        let storage = Storage::open(dir).await.expect("open storage");
        let backend = backend
            .unwrap_or_else(|| crate::tasks::DoubaoBackend::new(Arc::clone(&doubao)) as Arc<_>);
        AppState::new(
            config,
            config_path,
            platform,
            doubao,
            LogBuffer::new(),
            cdp_log,
            storage,
            backend,
        )
    }

    /// 同步测试用(不能在 tokio 运行时内调用)。
    pub fn test_state_blocking(dir: &Path) -> Arc<AppState> {
        tauri::async_runtime::block_on(test_state(dir))
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    #[test]
    fn to_camel_converts_nested_objects_and_arrays() {
        let v = json!({
            "cdp_connected": true,
            "ui_compat_details": { "account_button_found": false, "nested_list": [ { "read_at": 1 } ] },
            "accounts": [ { "is_current": true, "subscription": null }, { "is_ambiguous": false } ],
            "plain": "value",
            "last_check_at": 123
        });
        let got = to_camel(v);
        assert_eq!(got["cdpConnected"], true);
        assert_eq!(got["uiCompatDetails"]["accountButtonFound"], false);
        assert_eq!(got["uiCompatDetails"]["nestedList"][0]["readAt"], 1);
        assert_eq!(got["accounts"][0]["isCurrent"], true);
        assert!(got["accounts"][0]["subscription"].is_null());
        assert_eq!(got["accounts"][1]["isAmbiguous"], false);
        assert_eq!(got["plain"], "value");
        assert_eq!(got["lastCheckAt"], 123);
        assert!(got.get("cdp_connected").is_none());
    }

    #[test]
    fn to_camel_keeps_camel_and_special_keys_unchanged() {
        let v = json!({ "alreadyCamel": 1, "_private_key": 2, "单字": 3, "a_b_c": 4, "x_": 5 });
        let got = to_camel(v);
        assert_eq!(got["alreadyCamel"], 1);
        assert_eq!(got["_private_key"], 2);
        assert_eq!(got["单字"], 3);
        assert_eq!(got["aBC"], 4);
        assert_eq!(got["x"], 5);
    }

    #[test]
    fn to_camel_leaves_scalars_and_arrays_of_scalars() {
        assert_eq!(to_camel(json!("a_b")), json!("a_b"));
        assert_eq!(to_camel(json!([1, "x_y", null])), json!([1, "x_y", null]));
    }

    #[test]
    fn to_snake_reverses_camel_keys_recursively() {
        let v = json!({
            "server": { "bearerToken": "t", "port": 1 },
            "logging": { "fileMaxSizeMb": 10, "cdpDebugLog": true },
            "list": [ { "autoConnectOnStartup": false } ],
            "already_snake": 1
        });
        let got = to_snake(v);
        assert_eq!(got["server"]["bearer_token"], "t");
        assert_eq!(got["logging"]["file_max_size_mb"], 10);
        assert_eq!(got["logging"]["cdp_debug_log"], true);
        assert_eq!(got["list"][0]["auto_connect_on_startup"], false);
        assert_eq!(got["already_snake"], 1);
    }

    #[test]
    fn camel_snake_roundtrip() {
        for k in ["cdp_port", "ui_compat_check", "a", "last_check_at"] {
            assert_eq!(snake_key(&camel_key(k)), k);
        }
    }

    #[test]
    fn respond_serializes_to_camel() {
        #[derive(Serialize)]
        struct S {
            read_at: i64,
        }
        let v = respond(&S { read_at: 5 }).unwrap();
        assert_eq!(v, json!({ "readAt": 5 }));
    }
}
