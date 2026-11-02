//! 配置文件 `config.toml`,结构与 `docs/spec/07-data-model.md` 一一对应。
//!
//! 所有字段都有默认值:缺字段的旧配置文件能被读取并补齐;`bearer_token` 为空时自动生成并写回。

use std::fs;
use std::path::Path;

use anyhow::Context;
use serde::{Deserialize, Serialize};

const LOOPBACK_HOSTS: &[&str] = &["127.0.0.1", "::1", "localhost"];
const LOG_LEVELS: &[&str] = &["error", "warn", "info", "debug", "trace"];

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize, Default)]
#[serde(default)]
pub struct Config {
    pub server: ServerConfig,
    pub concurrency: ConcurrencyConfig,
    pub storage: StorageConfig,
    pub logging: LoggingConfig,
    pub doubao: DoubaoConfig,
    pub v88_executor: V88ExecutorConfig,
    pub ui: UiConfig,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(default)]
pub struct ServerConfig {
    pub enabled: bool,
    pub host: String,
    pub port: u16,
    pub bearer_token: String,
}

impl Default for ServerConfig {
    fn default() -> Self {
        Self {
            enabled: false,
            host: "127.0.0.1".into(),
            port: 8787,
            bearer_token: String::new(),
        }
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(default)]
pub struct ConcurrencyConfig {
    pub global_max: u32,
    pub per_account_max: u32,
    pub queue_depth_max: u32,
}

impl Default for ConcurrencyConfig {
    fn default() -> Self {
        Self {
            global_max: 3,
            per_account_max: 1,
            queue_depth_max: 50,
        }
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(default)]
pub struct StorageConfig {
    pub result_retention_hours: u32,
    pub metadata_retention_days: u32,
    pub upload_expiry_hours: u32,
}

impl Default for StorageConfig {
    fn default() -> Self {
        Self {
            result_retention_hours: 24,
            metadata_retention_days: 7,
            upload_expiry_hours: 1,
        }
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(default)]
pub struct LoggingConfig {
    pub level: String,
    pub file_max_size_mb: u32,
    pub file_retention_days: u32,
    pub cdp_debug_log: bool,
}

impl Default for LoggingConfig {
    fn default() -> Self {
        Self {
            level: "info".into(),
            file_max_size_mb: 10,
            file_retention_days: 7,
            cdp_debug_log: false,
        }
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(default)]
pub struct DoubaoConfig {
    pub cdp_port: u16,
    pub cdp_address: String,
    pub auto_connect_on_startup: bool,
    pub ui_compat_check_on_startup: bool,
}

impl Default for DoubaoConfig {
    fn default() -> Self {
        Self {
            cdp_port: 9707,
            cdp_address: "127.0.0.1".into(),
            auto_connect_on_startup: true,
            ui_compat_check_on_startup: true,
        }
    }
}

/// V88 配对成功后获得的专用执行器凭证。
///
/// 它与本机 HTTP API 的 `server.bearer_token` 完全独立，不能互相复用。
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize, Default)]
#[serde(default)]
pub struct V88ExecutorConfig {
    pub base_url: String,
    pub executor_id: String,
    pub executor_token: String,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
#[serde(default)]
pub struct UiConfig {
    pub minimize_to_tray_on_close: bool,
    pub start_minimized: bool,
}

impl Default for UiConfig {
    fn default() -> Self {
        Self {
            minimize_to_tray_on_close: true,
            start_minimized: false,
        }
    }
}

impl Config {
    /// 文件不存在:默认配置 + 生成 token + 写回;存在:读取并补默认,token 为空则生成并写回。
    pub fn load_or_create(path: &Path) -> anyhow::Result<Config> {
        let existed = path.exists();
        let mut legacy_server_without_enabled = false;
        let mut config = if existed {
            let text = fs::read_to_string(path)
                .with_context(|| format!("读取配置文件失败:{}", path.display()))?;
            let value = toml::from_str::<toml::Value>(&text)
                .with_context(|| format!("解析配置文件失败:{}", path.display()))?;
            legacy_server_without_enabled = value
                .get("server")
                .and_then(toml::Value::as_table)
                .is_some_and(|server| !server.contains_key("enabled"));
            toml::from_str::<Config>(&text)
                .with_context(|| format!("解析配置文件失败:{}", path.display()))?
        } else {
            Config::default()
        };

        let mut dirty = !existed;
        if legacy_server_without_enabled {
            config.server.enabled = true;
            dirty = true;
        }
        if config.server.bearer_token.trim().is_empty() {
            config.server.bearer_token = Config::generate_token();
            dirty = true;
        }
        if dirty {
            config.save(path)?;
        }
        Ok(config)
    }

    /// 写回 TOML(创建父目录;unix 下文件权限 0600,因为含 bearer_token)。
    pub fn save(&self, path: &Path) -> anyhow::Result<()> {
        if let Some(parent) = path.parent() {
            if !parent.as_os_str().is_empty() {
                fs::create_dir_all(parent)
                    .with_context(|| format!("创建配置目录失败:{}", parent.display()))?;
            }
        }
        let text = toml::to_string_pretty(self).context("序列化配置失败")?;
        let tmp = path.with_extension("toml.tmp");
        fs::write(&tmp, text).with_context(|| format!("写入配置文件失败:{}", tmp.display()))?;
        restrict_permissions(&tmp);
        fs::rename(&tmp, path).with_context(|| format!("替换配置文件失败:{}", path.display()))?;
        Ok(())
    }

    /// 校验:host 必须是 loopback;并发/队列/日志级别取值合理。
    pub fn validate(&self) -> Result<(), String> {
        if !LOOPBACK_HOSTS.contains(&self.server.host.as_str()) {
            return Err(format!(
                "出于安全考虑 server.host 只允许 127.0.0.1 / ::1 / localhost,当前为 {}",
                self.server.host
            ));
        }
        if self.server.port == 0 {
            return Err("server.port 不能为 0".into());
        }
        if self.concurrency.global_max == 0 {
            return Err("concurrency.global_max 至少为 1".into());
        }
        if self.concurrency.per_account_max == 0 {
            return Err("concurrency.per_account_max 至少为 1".into());
        }
        if self.concurrency.queue_depth_max == 0 {
            return Err("concurrency.queue_depth_max 至少为 1".into());
        }
        if !LOG_LEVELS.contains(&self.logging.level.as_str()) {
            return Err(format!(
                "logging.level 只允许 error / warn / info / debug / trace,当前为 {}",
                self.logging.level
            ));
        }
        if self.doubao.cdp_port == 0 {
            return Err("doubao.cdp_port 不能为 0".into());
        }
        Ok(())
    }

    /// 32 字节随机数的 hex 编码(64 字符)。
    pub fn generate_token() -> String {
        let bytes: [u8; 32] = rand::random();
        let mut s = String::with_capacity(64);
        for b in bytes {
            use std::fmt::Write;
            let _ = write!(s, "{b:02x}");
        }
        s
    }
}

#[cfg(unix)]
fn restrict_permissions(path: &Path) {
    use std::os::unix::fs::PermissionsExt;
    let _ = fs::set_permissions(path, fs::Permissions::from_mode(0o600));
}

#[cfg(not(unix))]
fn restrict_permissions(_path: &Path) {}

#[cfg(test)]
mod tests {
    use super::*;

    fn tmp_path(dir: &tempfile::TempDir) -> std::path::PathBuf {
        dir.path().join("config.toml")
    }

    #[test]
    fn creates_file_with_64_char_token_when_missing() {
        let dir = tempfile::tempdir().unwrap();
        let path = tmp_path(&dir);
        assert!(!path.exists());
        let cfg = Config::load_or_create(&path).unwrap();
        assert!(path.exists());
        assert_eq!(cfg.server.bearer_token.len(), 64);
        assert!(cfg
            .server
            .bearer_token
            .chars()
            .all(|c| c.is_ascii_hexdigit()));
        assert_eq!(cfg.server.port, 8787);
        assert_eq!(cfg.server.host, "127.0.0.1");
        assert!(!cfg.server.enabled, "全新安装默认关闭 HTTP API");
        let text = fs::read_to_string(&path).unwrap();
        assert!(text.contains("bearer_token"));
        assert!(text.contains("[doubao]"));
    }

    #[test]
    fn legacy_config_without_enabled_keeps_http_api_enabled() {
        let dir = tempfile::tempdir().unwrap();
        let path = tmp_path(&dir);
        fs::write(
            &path,
            "[server]\nhost=\"127.0.0.1\"\nport=8787\nbearer_token=\"legacy-token\"\n",
        )
        .unwrap();
        let cfg = Config::load_or_create(&path).unwrap();
        assert!(cfg.server.enabled);
        assert!(fs::read_to_string(&path)
            .unwrap()
            .contains("enabled = true"));
    }

    #[test]
    fn explicit_disabled_http_api_stays_disabled() {
        let dir = tempfile::tempdir().unwrap();
        let path = tmp_path(&dir);
        fs::write(
            &path,
            "[server]\nenabled=false\nhost=\"127.0.0.1\"\nport=8787\nbearer_token=\"token\"\n",
        )
        .unwrap();
        let cfg = Config::load_or_create(&path).unwrap();
        assert!(!cfg.server.enabled);
    }

    #[test]
    fn keeps_v88_executor_credential_separate_from_local_api_token() {
        let mut cfg = Config::default();
        cfg.server.bearer_token = "local-relay-token".into();
        cfg.v88_executor.base_url = "https://v88.example.com".into();
        cfg.v88_executor.executor_id = "lex_device".into();
        cfg.v88_executor.executor_token = "v88-device-token".into();
        let saved = toml::to_string(&cfg).unwrap();
        let loaded: Config = toml::from_str(&saved).unwrap();
        assert_eq!(loaded.v88_executor.base_url, "https://v88.example.com");
        assert_eq!(loaded.v88_executor.executor_id, "lex_device");
        assert_eq!(loaded.v88_executor.executor_token, "v88-device-token");
        assert_ne!(
            loaded.v88_executor.executor_token,
            loaded.server.bearer_token
        );
    }

    #[test]
    fn second_load_returns_same_token() {
        let dir = tempfile::tempdir().unwrap();
        let path = tmp_path(&dir);
        let a = Config::load_or_create(&path).unwrap();
        let b = Config::load_or_create(&path).unwrap();
        assert_eq!(a.server.bearer_token, b.server.bearer_token);
        assert_eq!(a, b);
    }

    #[test]
    fn creates_nested_parent_directories() {
        let dir = tempfile::tempdir().unwrap();
        let path = dir.path().join("a").join("b").join("config.toml");
        Config::load_or_create(&path).unwrap();
        assert!(path.exists());
    }

    #[test]
    fn partial_file_gets_defaults_for_other_fields() {
        let dir = tempfile::tempdir().unwrap();
        let path = tmp_path(&dir);
        fs::write(&path, "[server]\nport=9000\n").unwrap();
        let cfg = Config::load_or_create(&path).unwrap();
        assert_eq!(cfg.server.port, 9000);
        assert_eq!(cfg.server.host, "127.0.0.1");
        assert_eq!(cfg.server.bearer_token.len(), 64, "token 应被自动生成");
        assert_eq!(cfg.concurrency, ConcurrencyConfig::default());
        assert_eq!(cfg.storage, StorageConfig::default());
        assert_eq!(cfg.logging, LoggingConfig::default());
        assert_eq!(cfg.doubao, DoubaoConfig::default());
        assert_eq!(cfg.ui, UiConfig::default());
        // 生成的 token 已写回文件
        let again = Config::load_or_create(&path).unwrap();
        assert_eq!(again.server.bearer_token, cfg.server.bearer_token);
        assert_eq!(again.server.port, 9000);
    }

    #[test]
    fn existing_token_is_preserved() {
        let dir = tempfile::tempdir().unwrap();
        let path = tmp_path(&dir);
        fs::write(&path, "[server]\nbearer_token=\"abc\"\n").unwrap();
        let cfg = Config::load_or_create(&path).unwrap();
        assert_eq!(cfg.server.bearer_token, "abc");
    }

    #[test]
    fn invalid_toml_is_an_error() {
        let dir = tempfile::tempdir().unwrap();
        let path = tmp_path(&dir);
        fs::write(&path, "[server\nport=").unwrap();
        assert!(Config::load_or_create(&path).is_err());
    }

    #[test]
    fn validate_rejects_non_loopback_host() {
        let mut cfg = Config::default();
        cfg.server.host = "0.0.0.0".into();
        assert!(cfg.validate().is_err());
        for host in ["127.0.0.1", "::1", "localhost"] {
            cfg.server.host = host.into();
            assert!(cfg.validate().is_ok(), "{host} 应合法");
        }
    }

    #[test]
    fn validate_rejects_bad_level_and_zero_limits() {
        let mut cfg = Config::default();
        cfg.logging.level = "verbose".into();
        assert!(cfg.validate().is_err());
        let mut cfg = Config::default();
        cfg.concurrency.global_max = 0;
        assert!(cfg.validate().is_err());
    }

    #[test]
    fn generate_token_is_random_hex() {
        let a = Config::generate_token();
        let b = Config::generate_token();
        assert_eq!(a.len(), 64);
        assert_ne!(a, b);
        assert!(a.chars().all(|c| c.is_ascii_hexdigit()));
    }

    #[test]
    fn default_matches_spec_values() {
        let cfg = Config::default();
        assert_eq!(cfg.concurrency.global_max, 3);
        assert_eq!(cfg.concurrency.per_account_max, 1);
        assert_eq!(cfg.concurrency.queue_depth_max, 50);
        assert_eq!(cfg.storage.result_retention_hours, 24);
        assert_eq!(cfg.storage.metadata_retention_days, 7);
        assert_eq!(cfg.storage.upload_expiry_hours, 1);
        assert_eq!(cfg.logging.level, "info");
        assert_eq!(cfg.logging.file_max_size_mb, 10);
        assert_eq!(cfg.logging.file_retention_days, 7);
        assert!(!cfg.logging.cdp_debug_log);
        assert_eq!(cfg.doubao.cdp_port, 9707);
        assert_eq!(cfg.doubao.cdp_address, "127.0.0.1");
        assert!(cfg.doubao.auto_connect_on_startup);
        assert!(cfg.doubao.ui_compat_check_on_startup);
        assert!(cfg.ui.minimize_to_tray_on_close);
        assert!(!cfg.ui.start_minimized);
    }

    #[cfg(unix)]
    #[test]
    fn saved_file_is_owner_only() {
        use std::os::unix::fs::PermissionsExt;
        let dir = tempfile::tempdir().unwrap();
        let path = tmp_path(&dir);
        Config::load_or_create(&path).unwrap();
        let mode = fs::metadata(&path).unwrap().permissions().mode() & 0o777;
        assert_eq!(mode, 0o600);
    }
}
