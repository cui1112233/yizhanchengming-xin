//! 账号读取(spec 01 F2.1/F2.2;`docs/analysis/03-account-module.md` "DOM 通道" 的 Rust 版)。
//!
//! 纯解析(`parse_label` / `parse_entries`)与 CDP 流程(`ensure_injected` / `read_accounts`)分开,
//! 解析部分与 `scripts/accounts.js` 的 `parseLabel` 等价,便于单测。
//!
//! macOS 2.32.5 实测(`docs/analysis/10-live-probe-macos.md` Task 1.2 / 1.7):
//! - 账号行靠 `data-testid="avatar_menu_account_row"`(`readEntries()` 的 `isAccountRow`)判定,
//!   没有自定义头像的账号行里**没有 `<img>`**,所以不能像计划/沈浪那样用 `hasAvatar` 当必要条件;
//! - 只有 1 个账号时菜单里没有账号行,此时列表退化为 `[current]`;
//! - 侧栏折叠时账号按钮在屏外(`left<0`),`openMenu()` 的合成 pointer 事件仍能打开菜单,不必先展开侧栏。

use std::sync::LazyLock;
use std::time::Duration;

use regex::Regex;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};

use crate::cdp::{now_ms, CdpSession};
use crate::error::{RelayError, RelayResult};

use super::{ACCOUNTS_SCRIPT, ACCOUNTS_SCRIPT_VERSION};

/// 一个已登录账号(domain 类型,`snake_case`;IPC 层再转 camelCase)。
#[derive(Clone, Debug, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "snake_case")]
pub struct Account {
    pub name: String,
    pub subscription: Option<String>,
    /// 菜单行完整文本(含套餐),用于内部匹配。
    pub label: String,
    pub is_current: bool,
    pub is_ambiguous: bool,
}

/// `account.list` 的返回。
#[derive(Clone, Debug, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub struct AccountList {
    pub current: Option<String>,
    pub accounts: Vec<Account>,
    /// Unix epoch 毫秒。
    pub read_at: i64,
}

/// `accounts.js` `readEntries().entries[]` 的一条(页面返回 camelCase)。
#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct RawEntry {
    pub label: String,
    pub lines: Vec<String>,
    pub has_avatar: bool,
    pub has_popup: bool,
    #[serde(default)]
    pub disabled: bool,
    /// 页面按 `data-testid="avatar_menu_account_row"`(或无子菜单+有 `<img>` 兜底)判定的账号行。
    #[serde(default)]
    pub is_account_row: bool,
    #[serde(default)]
    pub test_id: String,
}

/// `accounts.js` `current()` 的返回。
#[derive(Debug, Deserialize)]
pub struct CurrentProbe {
    /// `logged_in` | `not_logged_in` | `unknown`。
    pub status: String,
    pub name: String,
    pub subscription: String,
    /// 账号按钮原始文本(可能两行),单账号退化时用作 `label`。
    #[serde(default)]
    pub raw: String,
}

/// 菜单里不是账号的条目(沈浪白名单,照抄)。
pub const EXCLUDED_LABELS: &[&str] = &[
    "设置",
    "外观设置",
    "收藏夹",
    "豆包官网",
    "检查更新",
    "帮助与反馈",
    "更多账号",
    "添加账号",
    "退出登录",
    "豆包订阅",
    "豆包订阅升级",
    "免费用豆包",
    "免费领额度",
    "免费领取额度",
];

/// 额度条目前缀(`当前时段额度用量\n0%`)。
const QUOTA_PREFIX: &str = "当前时段额度用量";
/// 账号名长度上限(字符数)。
const MAX_NAME_CHARS: usize = 120;

/// `read_accounts` 各阶段等待时间(测试用小值)。
#[derive(Clone, Copy, Debug)]
pub struct ReadTiming {
    /// `expandSidebar()` 后再 `probe()` 前。
    pub after_expand: Duration,
    /// `openMenu()` 后。
    pub after_open: Duration,
    /// 第一次 `hoverMore()` 后。
    pub after_hover: Duration,
    /// 每轮 `hoverMore()` → `readEntries()` 之间。
    pub per_attempt: Duration,
    /// 读取轮数上限。
    pub max_attempts: usize,
    /// `closeMenu()` 之后再返回前的等待。真机 2.32.5 实测:Escape 关闭后 **0ms 内再 `openMenu()`
    /// 菜单不会出现**(4/4),间隔 ≥150ms 才 100% 打开;`switch_account` 读完列表立刻点账号正是这条路径。
    pub after_close: Duration,
}

impl Default for ReadTiming {
    fn default() -> Self {
        Self {
            after_expand: Duration::from_millis(400),
            after_open: Duration::from_millis(650),
            after_hover: Duration::from_millis(650),
            per_attempt: Duration::from_millis(400),
            max_attempts: 8,
            after_close: Duration::from_millis(300),
        }
    }
}

/// 账号切换的节奏(沈浪 `operation=switch`:菜单 100ms × ≤40 轮抢时间点击;验证窗口 20s / 300ms)。
#[derive(Clone, Copy, Debug)]
pub struct SwitchTiming {
    /// 菜单里每轮 `hoverMore()` → `switchTo()` 之间。
    pub menu_poll: Duration,
    /// 菜单轮数上限。
    pub menu_max_attempts: usize,
    /// 切换验证轮询间隔。
    pub verify_poll: Duration,
    /// 切换验证窗口(spec F2.3:20s)。
    pub verify_timeout: Duration,
    /// 定位账号按钮(展开侧栏)时沿用的读取节奏。
    pub read: ReadTiming,
}

impl Default for SwitchTiming {
    fn default() -> Self {
        Self {
            menu_poll: Duration::from_millis(100),
            menu_max_attempts: 40,
            verify_poll: Duration::from_millis(300),
            verify_timeout: Duration::from_secs(20),
            read: ReadTiming::default(),
        }
    }
}

static SUB_RE: LazyLock<Regex> =
    LazyLock::new(|| Regex::new(r"^[一-龥A-Za-z0-9]{1,24}套餐$").unwrap());
static INLINE_SUB_RE: LazyLock<Regex> =
    LazyLock::new(|| Regex::new(r"^(.+?)\s+([一-龥A-Za-z0-9]{1,24}套餐)$").unwrap());

fn split_lines(raw: &str) -> Vec<&str> {
    raw.lines()
        .map(str::trim)
        .filter(|l| !l.is_empty())
        .collect()
}

/// 与 `accounts.js` 的 `parseLabel` 等价:
/// `"张三\n体验套餐"` / `"张三 体验套餐"` → (张三, Some(体验套餐));`"李四 字节科技"` → (李四 字节科技, None)。
pub fn parse_label(raw: &str) -> (String, Option<String>) {
    let lines = split_lines(raw);
    let Some(first) = lines.first() else {
        return (String::new(), None);
    };
    if let Some(sub) = lines.iter().skip(1).find(|l| SUB_RE.is_match(l)) {
        return (first.to_string(), Some(sub.to_string()));
    }
    if let Some(m) = INLINE_SUB_RE.captures(first) {
        return (m[1].trim().to_string(), Some(m[2].to_string()));
    }
    (first.to_string(), None)
}

fn is_excluded(label: &str) -> bool {
    EXCLUDED_LABELS.contains(&label)
}

/// 把 `readEntries()` 的条目过滤成账号列表:
/// - 只保留账号行(`is_account_row` 或 `has_avatar`),排除子菜单项、白名单词、额度条目、name 长度不在 1..=120;
/// - 同 `label` 去重;同 `name` 不同 `label` → 全部 `is_ambiguous`;`is_current = name == current.name`;
/// - 结果为空且 `current.status == "logged_in"` → 只含 current 的单元素(单账号时菜单没有账号行)。
pub fn parse_entries(entries: &[RawEntry], current: Option<&CurrentProbe>) -> Vec<Account> {
    let current = current.filter(|c| c.status == "logged_in" && !c.name.trim().is_empty());
    let mut accounts: Vec<Account> = Vec::new();
    for e in entries {
        if !(e.is_account_row || e.has_avatar) || e.has_popup {
            continue;
        }
        let label = e.label.trim();
        if label.is_empty() || is_excluded(label) || label.starts_with(QUOTA_PREFIX) {
            continue;
        }
        let (name, subscription) = parse_label(label);
        let len = name.chars().count();
        if len == 0 || len > MAX_NAME_CHARS || is_excluded(&name) {
            continue;
        }
        if accounts.iter().any(|a| a.label == label) {
            continue;
        }
        accounts.push(Account {
            is_current: current.is_some_and(|c| c.name == name),
            name,
            subscription,
            label: label.to_string(),
            is_ambiguous: false,
        });
    }
    for i in 0..accounts.len() {
        let dup = accounts
            .iter()
            .enumerate()
            .any(|(j, a)| j != i && a.name == accounts[i].name);
        accounts[i].is_ambiguous = dup;
    }
    if accounts.is_empty() {
        if let Some(c) = current {
            let name = c.name.trim().to_string();
            let subscription = Some(c.subscription.trim().to_string()).filter(|s| !s.is_empty());
            let raw = c.raw.trim();
            let label = if raw.is_empty() {
                match &subscription {
                    Some(s) => format!("{name} {s}"),
                    None => name.clone(),
                }
            } else {
                raw.to_string()
            };
            accounts.push(Account {
                name,
                subscription,
                label,
                is_current: true,
                is_ambiguous: false,
            });
        }
    }
    accounts
}

// ---------------------------------------------------------------------------
// CDP 流程
// ---------------------------------------------------------------------------

const API: &str = "window.__doubaoRelayAccounts";
/// 账号按钮的定位描述(进 `failed_selectors`)。
const ACCOUNT_BUTTON_SELECTOR: &str = r#"button[aria-haspopup="menu"] (左下角账号按钮)"#;
const MENU_ITEM_SELECTOR: &str = r#"[role="menuitem"]"#;

fn call(method: &str) -> String {
    format!("{API}.{method}()")
}

/// 带注入守卫的表达式:`__doubaoRelayAccounts` 不存在或版本不符时先注入脚本,再执行 `tail`
/// (`tail` 是一条 `return ...;` 语句)。页面 reload 后脚本丢失也能一次 evaluate 完成。
fn guarded_expression(tail: &str) -> String {
    format!(
        "(() => {{ if (!{API} || {API}.version !== {ver:?}) {{\n{script}\n}} {tail} }})()",
        ver = ACCOUNTS_SCRIPT_VERSION,
        script = ACCOUNTS_SCRIPT
    )
}

/// 注入表达式:返回注入后的版本。
fn inject_expression() -> String {
    guarded_expression(&format!("return {API} ? {API}.version : null;"))
}

/// 注入守卫 + 调用无参方法(切换验证循环在页面 reload 后用它读 `current()`)。
pub fn injected_call(method: &str) -> String {
    guarded_expression(&format!("return {API}.{method}();"))
}

/// `switchTo(name)` 表达式(账号名经 JSON 转义)。
fn switch_to_expression(name: &str) -> String {
    format!(
        "{API}.switchTo({})",
        serde_json::to_string(name).unwrap_or_else(|_| "\"\"".into())
    )
}

/// 确保 `accounts.js` 已注入且版本与 Rust 侧常量一致。
pub async fn ensure_injected(session: &CdpSession) -> RelayResult<()> {
    let version = session.evaluate(&inject_expression()).await?;
    match version.as_str() {
        Some(v) if v == ACCOUNTS_SCRIPT_VERSION => Ok(()),
        other => Err(RelayError::internal(format!(
            "账号脚本注入后版本不一致,期望 {ACCOUNTS_SCRIPT_VERSION},请重启 doubao-relay 后重试"
        ))
        .with_details(json!({ "actual": other }))),
    }
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct Probe {
    button_found: bool,
    #[serde(default)]
    login_visible: bool,
    #[serde(default)]
    haspopup_count: usize,
    #[serde(default)]
    candidates: Value,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct ScriptAction {
    #[serde(default)]
    action: String,
    #[serde(default)]
    still_open: bool,
    #[serde(default)]
    count: usize,
    #[serde(default)]
    x: Option<f64>,
    #[serde(default)]
    y: Option<f64>,
    /// `switchTo` 为 `ambiguous` 时同名各行的原文(11-error-codes `account_ambiguous.details.matches`)。
    #[serde(default)]
    labels: Vec<String>,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
struct ReadEntriesResult {
    #[serde(default)]
    menu_item_count: usize,
    #[serde(default)]
    logout_visible: bool,
    #[serde(default)]
    entries: Vec<RawEntry>,
}

/// 读账号列表(默认节奏)。流程:注入 → 定位账号按钮(必要时展开侧栏)→ `current()` →
/// 打开菜单 → hover "更多账号" → 循环读取直到条目数稳定 → 关闭菜单 → 解析。
pub async fn read_accounts(session: &CdpSession) -> RelayResult<AccountList> {
    read_accounts_with(session, ReadTiming::default()).await
}

/// 当前账号已由页面按钮直接确认,但后台可信点击被豆包拦截而无法展开菜单时,
/// 允许把它作为唯一可用账号返回。其他 UI 错误仍照常失败,不能伪造可切换账号列表。
fn can_fall_back_to_current(error: &RelayError, current: &CurrentProbe) -> bool {
    current.status == "logged_in"
        && !current.name.trim().is_empty()
        && error.code == "doubao_ui_changed"
}

/// 同 [`read_accounts`],可指定各阶段等待时间(测试用)。
pub async fn read_accounts_with(
    session: &CdpSession,
    timing: ReadTiming,
) -> RelayResult<AccountList> {
    ensure_injected(session).await?;
    locate_account_button(session, timing).await?;
    let current: CurrentProbe = session.evaluate_as(&call("current")).await?;
    if current.status == "not_logged_in" {
        return Err(RelayError::doubao_not_logged_in());
    }
    let entries = {
        let read = read_menu_entries(session, timing).await;
        if let Err(e) = close_menu(session).await {
            tracing::warn!(error = %e, "关闭账号菜单失败");
        }
        // 关闭后立刻再打开会被豆包忽略(见 `ReadTiming::after_close`),这里等菜单退出动画结束
        tokio::time::sleep(timing.after_close).await;
        match read {
            Ok(entries) => entries,
            Err(error) if can_fall_back_to_current(&error, &current) => {
                tracing::warn!(
                    current = %current.name,
                    "豆包后台不允许打开账号菜单,仅沿用已确认的当前账号"
                );
                Vec::new()
            }
            Err(error) => return Err(error),
        }
    };
    let mut accounts = parse_entries(&entries, Some(&current));
    // 豆包新版可能把“体验者”显示在页面当前账号按钮中，却只在菜单里
    // 返回一个无法实际切换的标签。当前页面账号才是可执行账号；把它补回列表，
    // 这样 API 可以直接复用当前会话，不会被不存在的菜单标签卡住。
    if current.status == "logged_in"
        && !current.name.trim().is_empty()
        && !accounts.iter().any(|a| a.name == current.name.trim())
    {
        accounts.insert(
            0,
            Account {
                name: current.name.trim().to_string(),
                label: current.name.trim().to_string(),
                subscription: None,
                is_current: true,
                is_ambiguous: false,
            },
        );
    }
    // 豆包新版在单账号模式下可能让 `current()` 保留旧账号名，
    // 但菜单里唯一的账号才是实际可用账号。此时不能把旧 probe 当成真当前账号，
    // 否则下游会对唯一账号重复执行切换并最终超时。
    let current_name = if accounts.len() == 1 {
        Some(accounts[0].name.clone())
    } else if current.status == "logged_in" && !current.name.trim().is_empty() {
        Some(current.name.trim().to_string())
    } else {
        accounts
            .iter()
            .find(|a| a.is_current)
            .map(|a| a.name.clone())
    };
    tracing::info!(
        count = accounts.len(),
        current = current_name.as_deref().unwrap_or(""),
        raw_entries = entries.len(),
        "账号列表读取完成"
    );
    Ok(AccountList {
        current: current_name,
        accounts,
        read_at: now_ms(),
    })
}

/// `probe()`;没找到按钮就 `expandSidebar()` 再试一次。
/// 仍没有:可见登录按钮 → `doubao_not_logged_in`,否则 `doubao_ui_changed`(details 带候选明细)。
async fn locate_account_button(session: &CdpSession, timing: ReadTiming) -> RelayResult<()> {
    let mut probe: Probe = session.evaluate_as(&call("probe")).await?;
    if !probe.button_found {
        let expand: ScriptAction = session.evaluate_as(&call("expandSidebar")).await?;
        tracing::debug!(action = %expand.action, count = expand.count, "账号按钮未找到,尝试展开侧栏");
        tokio::time::sleep(timing.after_expand).await;
        probe = session.evaluate_as(&call("probe")).await?;
    }
    if probe.button_found {
        return Ok(());
    }
    if probe.login_visible {
        return Err(RelayError::doubao_not_logged_in());
    }
    tracing::warn!(
        haspopup_count = probe.haspopup_count,
        "账号按钮未找到且无登录按钮,判定豆包界面变化"
    );
    Err(
        RelayError::doubao_ui_changed(vec![ACCOUNT_BUTTON_SELECTOR]).with_details(json!({
            "failed_selectors": [ACCOUNT_BUTTON_SELECTOR],
            "haspopup_count": probe.haspopup_count,
            "candidates": probe.candidates,
        })),
    )
}

/// 打开菜单并读取条目直到数量稳定(`attempt >= 3 && stable >= 2`)。
async fn read_menu_entries(session: &CdpSession, timing: ReadTiming) -> RelayResult<Vec<RawEntry>> {
    let open = open_menu(session).await?;
    tokio::time::sleep(timing.after_open).await;
    let more: ScriptAction = session.evaluate_as(&call("hoverMore")).await?;
    tracing::debug!(open = %open.action, more = %more.action, more_count = more.count, "账号菜单已打开");
    tokio::time::sleep(timing.after_hover).await;

    let mut last: Option<ReadEntriesResult> = None;
    let mut prev_count: Option<usize> = None;
    let mut stable = 0usize;
    for attempt in 0..timing.max_attempts.max(1) {
        let _: Value = session.evaluate(&call("hoverMore")).await?;
        tokio::time::sleep(timing.per_attempt).await;
        let read: ReadEntriesResult = session.evaluate_as(&call("readEntries")).await?;
        let count = read.entries.len();
        if prev_count == Some(count) {
            stable += 1;
        } else {
            stable = 0;
        }
        prev_count = Some(count);
        last = Some(read);
        if attempt >= 3 && stable >= 2 {
            break;
        }
    }
    let read = last.expect("至少读取一轮");
    if read.menu_item_count == 0 && !read.logout_visible {
        tracing::warn!(open = %open.action, "账号菜单打开后没有任何菜单项");
        return Err(
            RelayError::doubao_ui_changed(vec![MENU_ITEM_SELECTOR]).with_details(json!({
                "failed_selectors": [MENU_ITEM_SELECTOR],
                "open_action": open.action,
            })),
        );
    }
    Ok(read.entries)
}

/// 打开账号菜单。新版豆包忽略页面脚本合成的点击，因此坐标必须由 CDP 发送可信鼠标事件。
async fn open_menu(session: &CdpSession) -> RelayResult<ScriptAction> {
    let mut open: ScriptAction = session.evaluate_as(&call("openMenu")).await?;
    if open.action == "button_not_found" {
        return Err(RelayError::doubao_ui_changed(vec![ACCOUNT_BUTTON_SELECTOR]));
    }
    if open.action != "needs_native_click" {
        return Ok(open);
    }
    let (Some(x), Some(y)) = (open.x, open.y) else {
        return Err(RelayError::doubao_ui_changed(vec![ACCOUNT_BUTTON_SELECTOR]));
    };
    for (kind, buttons) in [("mouseMoved", 0), ("mousePressed", 1), ("mouseReleased", 0)] {
        let mut params = json!({ "type": kind, "x": x, "y": y });
        if kind != "mouseMoved" {
            params["button"] = json!("left");
            params["buttons"] = json!(buttons);
            params["clickCount"] = json!(1);
        }
        session.send("Input.dispatchMouseEvent", params).await?;
    }
    open.action = "native_click".into();
    Ok(open)
}

/// 关闭菜单:先让页面脚本关闭,再无条件发送一次真实 Escape。
/// 豆包菜单可能延迟挂载,页面脚本会在菜单出现前误报已关闭,所以不能只信 DOM 瞬时状态。
async fn close_menu(session: &CdpSession) -> RelayResult<()> {
    let r: ScriptAction = session.evaluate_as(&call("closeMenu")).await?;
    if r.still_open {
        tracing::warn!(action = %r.action, "合成 Escape 未能关闭账号菜单,改用 Input.dispatchKeyEvent");
    }
    for kind in ["rawKeyDown", "keyUp"] {
        session
            .send(
                "Input.dispatchKeyEvent",
                json!({
                    "type": kind,
                    "key": "Escape",
                    "code": "Escape",
                    "windowsVirtualKeyCode": 27,
                    "nativeVirtualKeyCode": 27,
                }),
            )
            .await?;
    }
    tokio::time::sleep(Duration::from_millis(200)).await;
    let again: ScriptAction = session.evaluate_as(&call("closeMenu")).await?;
    if again.still_open {
        tracing::warn!(action = %again.action, "账号菜单仍处于打开状态");
        if let (Some(x), Some(y)) = (again.x, again.y) {
            for (kind, buttons) in [("mouseMoved", 0), ("mousePressed", 1), ("mouseReleased", 0)] {
                let mut params = json!({ "type": kind, "x": x, "y": y });
                if kind != "mouseMoved" {
                    params["button"] = json!("left");
                    params["buttons"] = json!(buttons);
                    params["clickCount"] = json!(1);
                }
                session.send("Input.dispatchMouseEvent", params).await?;
            }
            tokio::time::sleep(Duration::from_millis(200)).await;
            let confirmed: ScriptAction = session.evaluate_as(&call("closeMenu")).await?;
            if confirmed.still_open {
                tracing::warn!(action = %confirmed.action, "原生点击后账号菜单仍处于打开状态");
            }
        }
    }
    Ok(())
}

/// 读当前账号(带注入守卫,页面 reload 后也能用)。返回原始 `CdpError` 供切换验证循环判断是否要重连。
pub async fn probe_current(session: &CdpSession) -> Result<CurrentProbe, crate::cdp::CdpError> {
    session.evaluate_as(&injected_call("current")).await
}

/// 在账号菜单里点击目标账号(只负责"点到",不验证切换结果):
/// 注入 → 定位按钮 → `openMenu` → 快节奏循环 `hoverMore` + `switchTo(target)` 直到 `clicked`;
/// `ambiguous` → `account_ambiguous`;轮数用尽仍未点到 → 关菜单并返回 `account_switch_failed`。
pub async fn click_account(
    session: &CdpSession,
    target: &str,
    timing: SwitchTiming,
) -> RelayResult<()> {
    ensure_injected(session).await?;
    locate_account_button(session, timing.read).await?;
    open_menu(session).await?;
    let mut last_action = String::new();
    for attempt in 0..timing.menu_max_attempts.max(1) {
        tokio::time::sleep(timing.menu_poll).await;
        let _: Value = session.evaluate(&call("hoverMore")).await?;
        let r: ScriptAction = session.evaluate_as(&switch_to_expression(target)).await?;
        match r.action.as_str() {
            "clicked" => {
                tracing::info!(target, attempt, "已在账号菜单点击目标账号");
                return Ok(());
            }
            "ambiguous" => {
                if let Err(e) = close_menu(session).await {
                    tracing::warn!(error = %e, "关闭账号菜单失败");
                }
                let matches = if r.labels.is_empty() {
                    vec![target.to_string(); r.count.max(2)]
                } else {
                    r.labels
                };
                return Err(RelayError::account_ambiguous(target, matches));
            }
            other => {
                last_action = other.to_string();
            }
        }
    }
    tracing::warn!(target, last_action = %last_action, "账号菜单里未能点击到目标账号");
    if let Err(e) = close_menu(session).await {
        tracing::warn!(error = %e, "关闭账号菜单失败");
    }
    Err(
        RelayError::account_switch_failed("account_row_not_clickable").with_details(json!({
            "reason": "account_row_not_clickable",
            "last_action": last_action,
            "attempts": timing.menu_max_attempts,
        })),
    )
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::sync::{Arc, Mutex};

    fn entry(label: &str, has_avatar: bool, has_popup: bool, is_account_row: bool) -> RawEntry {
        RawEntry {
            label: label.to_string(),
            lines: split_lines(label).into_iter().map(String::from).collect(),
            has_avatar,
            has_popup,
            disabled: false,
            is_account_row,
            test_id: String::new(),
        }
    }

    fn current(name: &str, sub: &str) -> CurrentProbe {
        CurrentProbe {
            status: "logged_in".into(),
            name: name.into(),
            subscription: sub.into(),
            raw: if sub.is_empty() {
                name.into()
            } else {
                format!("{name}\n{sub}")
            },
        }
    }

    // ---- parse_label ----

    #[test]
    fn parse_label_single_line_with_subscription() {
        assert_eq!(
            parse_label("张三 体验套餐"),
            ("张三".into(), Some("体验套餐".into()))
        );
    }

    #[test]
    fn parse_label_two_lines_with_subscription() {
        assert_eq!(
            parse_label("张三\n高级套餐"),
            ("张三".into(), Some("高级套餐".into()))
        );
        assert_eq!(
            parse_label("  张三 \r\n 体验套餐 \n"),
            ("张三".into(), Some("体验套餐".into()))
        );
    }

    #[test]
    fn parse_label_without_subscription_keeps_whole_name() {
        assert_eq!(parse_label("李四 字节科技"), ("李四 字节科技".into(), None));
        assert_eq!(parse_label("用户225205"), ("用户225205".into(), None));
        assert_eq!(
            parse_label("李四 字节科技 标准套餐"),
            ("李四 字节科技".into(), Some("标准套餐".into()))
        );
    }

    #[test]
    fn parse_label_empty() {
        assert_eq!(parse_label(""), (String::new(), None));
        assert_eq!(parse_label(" \n "), (String::new(), None));
    }

    #[test]
    fn parse_label_ignores_non_subscription_second_line() {
        assert_eq!(
            parse_label("当前时段额度用量\n0%"),
            ("当前时段额度用量".into(), None)
        );
        assert_eq!(parse_label("张三\n已加入 2024 年"), ("张三".into(), None));
        // 第二行以"套餐"结尾但含标点/空格,不算套餐
        assert_eq!(parse_label("张三\n这不是 套餐"), ("张三".into(), None));
    }

    // ---- parse_entries ----

    /// 真机 2 账号时 `readEntries()` 的 11 条(docs/analysis/10 Task 1.2)。
    fn live_two_accounts() -> Vec<RawEntry> {
        vec![
            entry("设置", false, false, false),
            entry("外观设置", false, true, false),
            entry("收藏夹", false, false, false),
            entry("豆包官网", false, false, false),
            entry("检查更新", false, false, false),
            entry("帮助与反馈", false, false, false),
            entry("当前时段额度用量\n0%", false, false, false),
            entry("她是翻哥\n体验套餐", true, false, true),
            entry("用户225205", false, false, true),
            entry("添加账号", false, false, false),
            entry("退出登录", false, false, false),
        ]
    }

    #[test]
    fn parse_entries_filters_excluded_quota_and_popup_rows() {
        let cur = current("她是翻哥", "体验套餐");
        let list = parse_entries(&live_two_accounts(), Some(&cur));
        assert_eq!(list.len(), 2, "{list:?}");
        assert_eq!(list[0].name, "她是翻哥");
        assert_eq!(list[0].subscription.as_deref(), Some("体验套餐"));
        assert_eq!(list[0].label, "她是翻哥\n体验套餐");
        assert!(list[0].is_current);
        assert!(!list[0].is_ambiguous);
        assert_eq!(list[1].name, "用户225205");
        assert_eq!(list[1].subscription, None);
        assert!(!list[1].is_current);
        // 无头像账号行(无 <img>)不能被丢掉
        assert!(!list[1].is_ambiguous);
    }

    #[test]
    fn parse_entries_excluded_label_even_with_avatar_is_dropped() {
        let entries = vec![
            entry("添加账号", true, false, true),
            entry("更多账号", true, true, true),
            entry("退出登录", true, false, false),
            entry("张三 体验套餐", true, false, true),
        ];
        let list = parse_entries(&entries, None);
        assert_eq!(list.len(), 1);
        assert_eq!(list[0].name, "张三");
    }

    #[test]
    fn parse_entries_does_not_treat_free_doubao_action_as_an_account() {
        let cur = current("用户945344", "");
        let entries = vec![entry("免费用豆包", true, false, true)];
        let list = parse_entries(&entries, Some(&cur));
        assert_eq!(list.len(), 1);
        assert_eq!(list[0].name, "用户945344");
        assert!(list[0].is_current);
    }

    #[test]
    fn parse_entries_has_popup_is_dropped() {
        let entries = vec![entry("张三 体验套餐", true, true, true)];
        assert!(parse_entries(&entries, None).is_empty());
    }

    #[test]
    fn parse_entries_non_account_rows_without_avatar_are_dropped() {
        let entries = vec![entry("王五 标准套餐", false, false, false)];
        assert!(parse_entries(&entries, None).is_empty());
        // 没 testid 但有 <img>(老结构)仍算账号
        let entries = vec![entry("王五 标准套餐", true, false, false)];
        assert_eq!(parse_entries(&entries, None).len(), 1);
    }

    #[test]
    fn parse_entries_same_name_different_label_marks_ambiguous() {
        let entries = vec![
            entry("张三 体验套餐", true, false, true),
            entry("张三 高级套餐", false, false, true),
            entry("李四", false, false, true),
        ];
        let cur = current("张三", "体验套餐");
        let list = parse_entries(&entries, Some(&cur));
        assert_eq!(list.len(), 3);
        assert!(list[0].is_ambiguous && list[1].is_ambiguous);
        assert!(!list[2].is_ambiguous);
        // 同名两条都按 name 匹配到 current
        assert!(list[0].is_current && list[1].is_current);
        assert!(!list[2].is_current);
    }

    #[test]
    fn parse_entries_same_label_is_deduplicated() {
        let entries = vec![
            entry("张三 体验套餐", true, false, true),
            entry("张三 体验套餐", true, false, true),
        ];
        let list = parse_entries(&entries, None);
        assert_eq!(list.len(), 1);
        assert!(!list[0].is_ambiguous);
    }

    #[test]
    fn parse_entries_current_marks_is_current_only_when_logged_in() {
        let entries = vec![entry("张三 体验套餐", true, false, true)];
        let mut cur = current("张三", "体验套餐");
        assert!(parse_entries(&entries, Some(&cur))[0].is_current);
        cur.status = "unknown".into();
        assert!(!parse_entries(&entries, Some(&cur))[0].is_current);
        assert!(!parse_entries(&entries, None)[0].is_current);
    }

    #[test]
    fn parse_entries_empty_with_logged_in_current_returns_single() {
        let cur = current("她是翻哥", "体验套餐");
        let list = parse_entries(&[], Some(&cur));
        assert_eq!(
            list,
            vec![Account {
                name: "她是翻哥".into(),
                subscription: Some("体验套餐".into()),
                label: "她是翻哥\n体验套餐".into(),
                is_current: true,
                is_ambiguous: false,
            }]
        );
        // 真机单账号菜单:只有非账号条目,同样退化为 [current]
        let only_menu: Vec<RawEntry> = live_two_accounts()
            .into_iter()
            .filter(|e| !e.is_account_row)
            .collect();
        let list = parse_entries(&only_menu, Some(&cur));
        assert_eq!(list.len(), 1);
        assert_eq!(list[0].name, "她是翻哥");
    }

    #[test]
    fn parse_entries_single_fallback_label_without_raw() {
        let cur = CurrentProbe {
            status: "logged_in".into(),
            name: "李四".into(),
            subscription: String::new(),
            raw: String::new(),
        };
        let list = parse_entries(&[], Some(&cur));
        assert_eq!(list.len(), 1);
        assert_eq!(list[0].label, "李四");
        assert_eq!(list[0].subscription, None);
    }

    #[test]
    fn parse_entries_empty_without_logged_in_current_is_empty() {
        let mut cur = current("张三", "");
        cur.status = "not_logged_in".into();
        assert!(parse_entries(&[], Some(&cur)).is_empty());
        assert!(parse_entries(&[], None).is_empty());
    }

    #[test]
    fn parse_entries_name_length_bounds() {
        let long = "a".repeat(121);
        let ok = "a".repeat(120);
        let entries = vec![
            entry(&long, true, false, true),
            entry(&ok, true, false, true),
            entry("  ", true, false, true),
        ];
        let list = parse_entries(&entries, None);
        assert_eq!(list.len(), 1);
        assert_eq!(list[0].name, ok);
    }

    #[test]
    fn raw_entry_deserializes_camel_case_with_defaults() {
        let e: RawEntry = serde_json::from_value(json!({
            "label": "张三\n体验套餐", "lines": ["张三", "体验套餐"], "hasAvatar": true, "hasPopup": false
        }))
        .unwrap();
        assert!(e.has_avatar && !e.has_popup && !e.disabled && !e.is_account_row);
        assert_eq!(e.test_id, "");
        let e: RawEntry = serde_json::from_value(json!({
            "label": "x", "lines": ["x"], "hasAvatar": false, "hasPopup": false,
            "disabled": true, "isAccountRow": true, "testId": "avatar_menu_account_row"
        }))
        .unwrap();
        assert!(e.disabled && e.is_account_row);
        assert_eq!(e.test_id, "avatar_menu_account_row");
    }

    #[test]
    fn account_list_serializes_snake_case() {
        let v = serde_json::to_value(AccountList {
            current: Some("张三".into()),
            accounts: vec![Account {
                name: "张三".into(),
                subscription: None,
                label: "张三".into(),
                is_current: true,
                is_ambiguous: false,
            }],
            read_at: 5,
        })
        .unwrap();
        assert_eq!(v["read_at"], 5);
        assert_eq!(v["accounts"][0]["is_current"], true);
        assert_eq!(v["accounts"][0]["is_ambiguous"], false);
        assert!(v["accounts"][0]["subscription"].is_null());
    }

    #[test]
    fn inject_expression_embeds_script_and_version() {
        let e = inject_expression();
        assert!(e.starts_with("(() => {"));
        assert!(e.contains(&format!("const VERSION = '{ACCOUNTS_SCRIPT_VERSION}'")));
        assert!(e.contains(&format!("version !== \"{ACCOUNTS_SCRIPT_VERSION}\"")));
        assert!(!e
            .trim_start_matches("(() => {")
            .trim_start()
            .starts_with("await"));
    }

    // ---- CDP 流程(mock page) ----

    type Handler = Arc<Mutex<Box<dyn FnMut(&str) -> Value + Send>>>;

    /// mock 豆包页面:`Runtime.evaluate` 按表达式交给 `handler`,其它命令回 `{}`;记录收到的 method/表达式。
    async fn spawn_mock_page(handler: Handler) -> (Arc<CdpSession>, Arc<Mutex<Vec<String>>>) {
        use futures_util::{SinkExt, StreamExt};
        use tokio_tungstenite::tungstenite::Message;
        let log: Arc<Mutex<Vec<String>>> = Arc::new(Mutex::new(Vec::new()));
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        let log2 = Arc::clone(&log);
        tokio::spawn(async move {
            let (sock, _) = listener.accept().await.unwrap();
            let mut ws = tokio_tungstenite::accept_async(sock).await.unwrap();
            while let Some(Ok(Message::Text(txt))) = ws.next().await {
                let req: Value = serde_json::from_str(&txt).unwrap();
                let method = req["method"].as_str().unwrap_or("").to_string();
                let result = if method == "Runtime.evaluate" {
                    let expr = req["params"]["expression"].as_str().unwrap_or("");
                    let short = if expr.contains("const VERSION") {
                        "<inject>".to_string()
                    } else {
                        expr.to_string()
                    };
                    log2.lock().unwrap().push(short);
                    let value = (handler.lock().unwrap())(expr);
                    json!({ "result": { "type": "object", "value": value } })
                } else {
                    log2.lock().unwrap().push(format!(
                        "{method} {}",
                        req["params"]["type"].as_str().unwrap_or("")
                    ));
                    json!({})
                };
                let resp = json!({ "id": req["id"], "result": result });
                if ws
                    .send(Message::Text(resp.to_string().into()))
                    .await
                    .is_err()
                {
                    break;
                }
            }
        });
        let session = CdpSession::connect(&format!("ws://127.0.0.1:{port}/devtools/page/X"), "X")
            .await
            .unwrap();
        (session, log)
    }

    fn fast() -> ReadTiming {
        ReadTiming {
            after_expand: Duration::from_millis(5),
            after_open: Duration::from_millis(5),
            after_hover: Duration::from_millis(5),
            per_attempt: Duration::from_millis(5),
            max_attempts: 8,
            after_close: Duration::from_millis(5),
        }
    }

    #[test]
    fn read_timing_default_matches_live_probe() {
        let t = ReadTiming::default();
        assert_eq!(t.after_open, Duration::from_millis(650));
        assert_eq!(t.after_hover, Duration::from_millis(650));
        assert_eq!(t.per_attempt, Duration::from_millis(400));
        assert_eq!(t.max_attempts, 8);
        // 真机:关闭后 ≥150ms 才能再次打开,默认留 300ms 裕量
        assert!(t.after_close >= Duration::from_millis(150));
    }

    fn live_entries_json() -> Value {
        json!([
            {"label":"设置","lines":["设置"],"hasAvatar":false,"hasPopup":false,"disabled":false,"isAccountRow":false,"testId":"profile_menu_setting"},
            {"label":"外观设置","lines":["外观设置"],"hasAvatar":false,"hasPopup":true,"disabled":false,"isAccountRow":false,"testId":""},
            {"label":"收藏夹","lines":["收藏夹"],"hasAvatar":false,"hasPopup":false,"disabled":false,"isAccountRow":false,"testId":""},
            {"label":"豆包官网","lines":["豆包官网"],"hasAvatar":false,"hasPopup":false,"disabled":false,"isAccountRow":false,"testId":"doubao_product_website"},
            {"label":"检查更新","lines":["检查更新"],"hasAvatar":false,"hasPopup":false,"disabled":false,"isAccountRow":false,"testId":""},
            {"label":"帮助与反馈","lines":["帮助与反馈"],"hasAvatar":false,"hasPopup":false,"disabled":false,"isAccountRow":false,"testId":"profile_menu_feedback"},
            {"label":"当前时段额度用量\n0%","lines":["当前时段额度用量","0%"],"hasAvatar":false,"hasPopup":false,"disabled":false,"isAccountRow":false,"testId":""},
            {"label":"她是翻哥\n体验套餐","lines":["她是翻哥","体验套餐"],"hasAvatar":true,"hasPopup":false,"disabled":false,"isAccountRow":true,"testId":"avatar_menu_account_row"},
            {"label":"用户225205","lines":["用户225205"],"hasAvatar":false,"hasPopup":false,"disabled":false,"isAccountRow":true,"testId":"avatar_menu_account_row"},
            {"label":"添加账号","lines":["添加账号"],"hasAvatar":false,"hasPopup":false,"disabled":false,"isAccountRow":false,"testId":"avatar_menu_add_account"},
            {"label":"退出登录","lines":["退出登录"],"hasAvatar":false,"hasPopup":false,"disabled":false,"isAccountRow":false,"testId":"setup_logout"}
        ])
    }

    fn probe_json(found: bool, login: bool) -> Value {
        json!({"version": ACCOUNTS_SCRIPT_VERSION, "buttonFound": found, "buttonText": "", "expanded": "false",
            "buttonHasImg": found, "loginVisible": login, "hasInput": true, "readyState": "complete",
            "visibility": "visible", "innerWidth": 1512, "innerHeight": 879, "haspopupCount": 28,
            "candidates": [{"text": "x", "hasImg": true, "hasSub": true, "rect": [9, 823, 231, 48]}]})
    }

    fn current_json(status: &str, name: &str, sub: &str) -> Value {
        let raw = if sub.is_empty() {
            name.to_string()
        } else {
            format!("{name}\n{sub}")
        };
        json!({"status": status, "name": name, "subscription": sub, "raw": raw})
    }

    /// 标准 mock:按表达式路由;`entries_fn` 控制每次 `readEntries()` 的返回。
    fn handler(
        probe: Value,
        current: Value,
        mut entries_fn: impl FnMut(usize) -> Value + Send + 'static,
        mut close_fn: impl FnMut(usize) -> Value + Send + 'static,
    ) -> Handler {
        let mut reads = 0usize;
        let mut closes = 0usize;
        Arc::new(Mutex::new(Box::new(move |expr: &str| {
            if expr.contains("const VERSION") {
                return Value::String(ACCOUNTS_SCRIPT_VERSION.into());
            }
            match expr {
                e if e == call("probe") => probe.clone(),
                e if e == call("current") => current.clone(),
                e if e == call("expandSidebar") => json!({"action": "none", "count": 0}),
                e if e == call("openMenu") => {
                    json!({"action": "needs_native_click", "x": 424.0, "y": 835.0})
                }
                e if e == call("hoverMore") => json!({"action": "none", "count": 0}),
                e if e == call("readEntries") => {
                    reads += 1;
                    entries_fn(reads)
                }
                e if e == call("closeMenu") => {
                    closes += 1;
                    close_fn(closes)
                }
                other => panic!("unexpected expression: {other}"),
            }
        })))
    }

    fn entries_result(entries: Value) -> Value {
        let n = entries.as_array().map(|a| a.len()).unwrap_or(0);
        json!({"menuItemCount": n, "logoutVisible": n > 0, "entries": entries})
    }

    #[tokio::test]
    async fn read_accounts_happy_path_two_accounts() {
        let h = handler(
            probe_json(true, false),
            current_json("logged_in", "她是翻哥", "体验套餐"),
            |_| entries_result(live_entries_json()),
            |_| json!({"action": "escape", "stillOpen": false}),
        );
        let (session, log) = spawn_mock_page(h).await;
        let list = read_accounts_with(&session, fast()).await.unwrap();
        assert_eq!(list.current.as_deref(), Some("她是翻哥"));
        assert_eq!(list.accounts.len(), 2, "{:?}", list.accounts);
        assert!(list.accounts[0].is_current);
        assert_eq!(list.accounts[1].name, "用户225205");
        assert!(list.read_at > 0);
        let log = log.lock().unwrap().clone();
        assert_eq!(log[0], "<inject>");
        assert_eq!(log[1], call("probe"));
        assert_eq!(log[2], call("current"));
        assert_eq!(log[3], call("openMenu"));
        assert_eq!(
            &log[4..7],
            [
                "Input.dispatchMouseEvent mouseMoved",
                "Input.dispatchMouseEvent mousePressed",
                "Input.dispatchMouseEvent mouseReleased",
            ]
        );
        assert_eq!(log[7], call("hoverMore"));
        // 稳定判定:attempt>=3 && stable>=2 → 恰好 4 轮 hoverMore+readEntries
        let reads = log.iter().filter(|e| **e == call("readEntries")).count();
        assert_eq!(reads, 4, "{log:?}");
        assert_eq!(log.last().unwrap(), &call("closeMenu"));
        assert!(log.contains(&"Input.dispatchKeyEvent rawKeyDown".into()));
        assert!(log.contains(&"Input.dispatchKeyEvent keyUp".into()));
        session.close().await;
    }

    /// 真机:关菜单后立刻再开会被忽略,`read_accounts_with` 必须在 `closeMenu` 之后等 `after_close` 再返回。
    #[tokio::test]
    async fn read_accounts_waits_after_close_before_returning() {
        let h = handler(
            probe_json(true, false),
            current_json("logged_in", "她是翻哥", "体验套餐"),
            |_| entries_result(live_entries_json()),
            |_| json!({"action": "escape", "stillOpen": false}),
        );
        let (session, log) = spawn_mock_page(h).await;
        let mut timing = fast();
        timing.after_close = Duration::from_millis(120);
        let started = std::time::Instant::now();
        read_accounts_with(&session, timing).await.unwrap();
        let elapsed = started.elapsed();
        assert!(elapsed >= Duration::from_millis(120), "{elapsed:?}");
        assert_eq!(log.lock().unwrap().last().unwrap(), &call("closeMenu"));
        session.close().await;
    }

    #[tokio::test]
    async fn read_accounts_always_sends_real_escape_even_when_dom_reports_menu_closed() {
        let h = handler(
            probe_json(true, false),
            current_json("logged_in", "她是翻哥", "体验套餐"),
            |_| entries_result(live_entries_json()),
            |_| json!({"action": "escape", "stillOpen": false}),
        );
        let (session, log) = spawn_mock_page(h).await;
        read_accounts_with(&session, fast()).await.unwrap();
        let log = log.lock().unwrap().clone();
        let keys: Vec<&str> = log
            .iter()
            .filter(|line| line.starts_with("Input.dispatchKeyEvent"))
            .map(String::as_str)
            .collect();
        assert_eq!(
            keys,
            [
                "Input.dispatchKeyEvent rawKeyDown",
                "Input.dispatchKeyEvent keyUp"
            ]
        );
        assert_eq!(log.last().unwrap(), &call("closeMenu"));
        session.close().await;
    }

    #[tokio::test]
    async fn read_accounts_waits_until_entry_count_stable() {
        // 前 3 次读取条目数递增,之后稳定 → 应读 3 + 3 轮(第 6 轮 attempt=5,stable=2)
        let h = handler(
            probe_json(true, false),
            current_json("logged_in", "她是翻哥", "体验套餐"),
            |n| {
                let all = live_entries_json();
                let arr = all.as_array().unwrap();
                let take = match n {
                    1 => 3,
                    2 => 6,
                    _ => arr.len(),
                };
                entries_result(Value::Array(arr[..take].to_vec()))
            },
            |_| json!({"action": "escape", "stillOpen": false}),
        );
        let (session, log) = spawn_mock_page(h).await;
        let list = read_accounts_with(&session, fast()).await.unwrap();
        assert_eq!(list.accounts.len(), 2);
        let reads = log
            .lock()
            .unwrap()
            .iter()
            .filter(|e| **e == call("readEntries"))
            .count();
        assert_eq!(reads, 5);
        session.close().await;
    }

    #[tokio::test]
    async fn read_accounts_gives_up_after_max_attempts_when_never_stable() {
        let h = handler(
            probe_json(true, false),
            current_json("logged_in", "她是翻哥", "体验套餐"),
            |n| {
                let all = live_entries_json();
                let arr = all.as_array().unwrap();
                entries_result(Value::Array(arr[..(n % arr.len()).max(1)].to_vec()))
            },
            |_| json!({"action": "escape", "stillOpen": false}),
        );
        let (session, log) = spawn_mock_page(h).await;
        let _ = read_accounts_with(&session, fast()).await.unwrap();
        let reads = log
            .lock()
            .unwrap()
            .iter()
            .filter(|e| **e == call("readEntries"))
            .count();
        assert_eq!(reads, 8);
        session.close().await;
    }

    #[tokio::test]
    async fn read_accounts_single_account_falls_back_to_current() {
        let only_menu: Vec<Value> = live_entries_json()
            .as_array()
            .unwrap()
            .iter()
            .filter(|e| e["isAccountRow"] != true)
            .cloned()
            .collect();
        let h = handler(
            probe_json(true, false),
            current_json("logged_in", "她是翻哥", "体验套餐"),
            move |_| entries_result(Value::Array(only_menu.clone())),
            |_| json!({"action": "escape", "stillOpen": false}),
        );
        let (session, _) = spawn_mock_page(h).await;
        let list = read_accounts_with(&session, fast()).await.unwrap();
        assert_eq!(list.current.as_deref(), Some("她是翻哥"));
        assert_eq!(list.accounts.len(), 1);
        assert_eq!(list.accounts[0].label, "她是翻哥\n体验套餐");
        assert!(list.accounts[0].is_current);
        session.close().await;
    }

    #[tokio::test]
    async fn read_accounts_not_logged_in_when_login_button_visible() {
        let h = handler(
            probe_json(false, true),
            current_json("not_logged_in", "", ""),
            |_| unreachable!("未登录不应读菜单"),
            |_| unreachable!(),
        );
        let (session, log) = spawn_mock_page(h).await;
        let err = read_accounts_with(&session, fast()).await.unwrap_err();
        assert_eq!(err.code, "doubao_not_logged_in");
        let log = log.lock().unwrap().clone();
        // 先 expandSidebar 再 probe 一次,没有 openMenu
        assert!(log.contains(&call("expandSidebar")), "{log:?}");
        assert_eq!(log.iter().filter(|e| **e == call("probe")).count(), 2);
        assert!(!log.contains(&call("openMenu")));
        session.close().await;
    }

    #[tokio::test]
    async fn read_accounts_not_logged_in_when_current_says_so() {
        let h = handler(
            probe_json(true, false),
            current_json("not_logged_in", "", ""),
            |_| unreachable!(),
            |_| unreachable!(),
        );
        let (session, log) = spawn_mock_page(h).await;
        let err = read_accounts_with(&session, fast()).await.unwrap_err();
        assert_eq!(err.code, "doubao_not_logged_in");
        assert!(!log.lock().unwrap().contains(&call("openMenu")));
        session.close().await;
    }

    #[tokio::test]
    async fn read_accounts_ui_changed_when_button_missing_without_login() {
        let h = handler(
            probe_json(false, false),
            current_json("unknown", "", ""),
            |_| unreachable!(),
            |_| unreachable!(),
        );
        let (session, _) = spawn_mock_page(h).await;
        let err = read_accounts_with(&session, fast()).await.unwrap_err();
        assert_eq!(err.code, "doubao_ui_changed");
        let d = err.details.unwrap();
        assert_eq!(d["failed_selectors"][0], ACCOUNT_BUTTON_SELECTOR);
        assert_eq!(d["haspopup_count"], 28);
        assert!(d["candidates"].is_array());
        session.close().await;
    }

    #[tokio::test]
    async fn read_accounts_uses_current_account_when_background_menu_never_opens() {
        let h = handler(
            probe_json(true, false),
            current_json("logged_in", "她是翻哥", "体验套餐"),
            |_| json!({"menuItemCount": 0, "logoutVisible": false, "entries": []}),
            |_| json!({"action": "none", "stillOpen": false}),
        );
        let (session, log) = spawn_mock_page(h).await;
        let list = read_accounts_with(&session, fast()).await.unwrap();
        assert_eq!(list.current.as_deref(), Some("她是翻哥"));
        assert_eq!(list.accounts.len(), 1);
        assert_eq!(list.accounts[0].name, "她是翻哥");
        assert!(list.accounts[0].is_current);
        // 菜单失败也要尝试收尾,且不能让当前账号读取失败。
        assert!(log.lock().unwrap().contains(&call("closeMenu")));
        session.close().await;
    }

    #[tokio::test]
    async fn read_accounts_falls_back_to_key_event_when_menu_stays_open() {
        let h = handler(
            probe_json(true, false),
            current_json("logged_in", "她是翻哥", "体验套餐"),
            |_| entries_result(live_entries_json()),
            |n| {
                if n == 1 {
                    json!({"action": "escape_then_pointer", "stillOpen": true})
                } else {
                    json!({"action": "none", "stillOpen": false})
                }
            },
        );
        let (session, log) = spawn_mock_page(h).await;
        let list = read_accounts_with(&session, fast()).await.unwrap();
        assert_eq!(list.accounts.len(), 2);
        let log = log.lock().unwrap().clone();
        let n = log.len();
        assert_eq!(log[n - 4], call("closeMenu"));
        assert_eq!(log[n - 3], "Input.dispatchKeyEvent rawKeyDown");
        assert_eq!(log[n - 2], "Input.dispatchKeyEvent keyUp");
        assert_eq!(log[n - 1], call("closeMenu"));
        session.close().await;
    }

    #[tokio::test]
    async fn read_accounts_uses_native_click_when_menu_stays_open_after_escape() {
        let h = handler(
            probe_json(true, false),
            current_json("logged_in", "她是翻哥", "体验套餐"),
            |_| entries_result(live_entries_json()),
            |n| match n {
                1 | 2 => {
                    json!({"action": "escape_then_click", "stillOpen": true, "x": 424.0, "y": 835.0})
                }
                _ => json!({"action": "native_click", "stillOpen": false}),
            },
        );
        let (session, log) = spawn_mock_page(h).await;
        let list = read_accounts_with(&session, fast()).await.unwrap();
        assert_eq!(list.accounts.len(), 2);
        let log = log.lock().unwrap().clone();
        let native_clicks = log
            .iter()
            .filter(|entry| entry.as_str() == "Input.dispatchMouseEvent mouseMoved")
            .count();
        assert_eq!(
            native_clicks, 2,
            "第一次是打开账号菜单，第二次必须是关闭仍遮挡页面的菜单：{log:?}"
        );
        let native_click = log
            .iter()
            .rposition(|entry| entry == "Input.dispatchMouseEvent mouseMoved")
            .expect("菜单仍打开时必须通过 CDP 原生点击关闭");
        assert_eq!(
            &log[native_click..native_click + 3],
            [
                "Input.dispatchMouseEvent mouseMoved",
                "Input.dispatchMouseEvent mousePressed",
                "Input.dispatchMouseEvent mouseReleased",
            ]
        );
        assert_eq!(log.last().unwrap(), &call("closeMenu"));
        session.close().await;
    }

    #[tokio::test]
    async fn ensure_injected_rejects_version_mismatch() {
        let h: Handler = Arc::new(Mutex::new(Box::new(|_: &str| {
            Value::String("0.0.1".into())
        })));
        let (session, _) = spawn_mock_page(h).await;
        let err = ensure_injected(&session).await.unwrap_err();
        assert_eq!(err.code, "system_internal_error");
        assert_eq!(err.details.unwrap()["actual"], "0.0.1");
        session.close().await;
    }

    #[tokio::test]
    async fn read_accounts_maps_closed_session_to_cdp_disconnected() {
        use futures_util::StreamExt;
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        tokio::spawn(async move {
            let (sock, _) = listener.accept().await.unwrap();
            let mut ws = tokio_tungstenite::accept_async(sock).await.unwrap();
            let _ = ws.next().await;
            let _ = ws.close(None).await;
        });
        let session = CdpSession::connect(&format!("ws://127.0.0.1:{port}/devtools/page/X"), "X")
            .await
            .unwrap();
        let err = read_accounts_with(&session, fast()).await.unwrap_err();
        assert_eq!(err.code, "doubao_cdp_disconnected");
    }

    // ---- 切换:injected_call / click_account ----

    #[test]
    fn injected_call_guards_injection_and_calls_method() {
        let e = injected_call("current");
        assert!(e.starts_with("(() => {"));
        assert!(e.contains(&format!("const VERSION = '{ACCOUNTS_SCRIPT_VERSION}'")));
        assert!(e.ends_with("return window.__doubaoRelayAccounts.current(); })()"));
        assert_eq!(
            switch_to_expression("张三 \"x\""),
            r#"window.__doubaoRelayAccounts.switchTo("张三 \"x\"")"#
        );
    }

    fn fast_switch() -> SwitchTiming {
        SwitchTiming {
            menu_poll: Duration::from_millis(2),
            menu_max_attempts: 5,
            verify_poll: Duration::from_millis(5),
            verify_timeout: Duration::from_millis(200),
            read: fast(),
        }
    }

    /// 切换 mock:`switchTo` 的返回由 `switch_fn(第几次)` 决定;其它方法按标准回答。
    fn switch_handler(mut switch_fn: impl FnMut(usize, &str) -> Value + Send + 'static) -> Handler {
        let mut switches = 0usize;
        Arc::new(Mutex::new(Box::new(move |expr: &str| {
            if expr.contains("const VERSION") {
                return Value::String(ACCOUNTS_SCRIPT_VERSION.into());
            }
            if let Some(rest) = expr.strip_prefix("window.__doubaoRelayAccounts.switchTo(") {
                let name: String =
                    serde_json::from_str(rest.trim_end_matches(')')).expect("JSON 转义的账号名");
                switches += 1;
                return switch_fn(switches, &name);
            }
            match expr {
                e if e == call("probe") => probe_json(true, false),
                e if e == call("expandSidebar") => json!({"action": "none", "count": 0}),
                e if e == call("openMenu") => {
                    json!({"action": "needs_native_click", "x": 424.0, "y": 835.0})
                }
                e if e == call("hoverMore") => json!({"action": "none", "count": 0}),
                e if e == call("closeMenu") => json!({"action": "escape", "stillOpen": false}),
                other => panic!("unexpected expression: {other}"),
            }
        })))
    }

    #[tokio::test]
    async fn click_account_retries_until_clicked() {
        // 前两轮账号行还没渲染出来(not_found),第三轮点到
        let h = switch_handler(|n, name| {
            assert_eq!(name, "用户225205");
            if n < 3 {
                json!({"action": "not_found", "count": 0})
            } else {
                json!({"action": "clicked", "name": name})
            }
        });
        let (session, log) = spawn_mock_page(h).await;
        click_account(&session, "用户225205", fast_switch())
            .await
            .unwrap();
        let log = log.lock().unwrap().clone();
        assert_eq!(log[0], "<inject>");
        assert_eq!(log[1], call("probe"));
        assert_eq!(log[2], call("openMenu"));
        assert_eq!(
            &log[3..6],
            [
                "Input.dispatchMouseEvent mouseMoved",
                "Input.dispatchMouseEvent mousePressed",
                "Input.dispatchMouseEvent mouseReleased",
            ]
        );
        let switches = log.iter().filter(|e| e.contains("switchTo(")).count();
        assert_eq!(switches, 3, "{log:?}");
        // 点到之后不再关菜单(页面会 reload)
        assert!(!log.contains(&call("closeMenu")));
        session.close().await;
    }

    #[tokio::test]
    async fn click_account_fails_after_max_attempts_and_closes_menu() {
        let h = switch_handler(|_, _| json!({"action": "not_found", "count": 0}));
        let (session, log) = spawn_mock_page(h).await;
        let err = click_account(&session, "李四", fast_switch())
            .await
            .unwrap_err();
        assert_eq!(err.code, "account_switch_failed");
        let d = err.details.unwrap();
        assert_eq!(d["reason"], "account_row_not_clickable");
        assert_eq!(d["last_action"], "not_found");
        assert_eq!(d["attempts"], 5);
        let log = log.lock().unwrap().clone();
        assert_eq!(log.iter().filter(|e| e.contains("switchTo(")).count(), 5);
        assert_eq!(log.last().unwrap(), &call("closeMenu"));
        session.close().await;
    }

    #[tokio::test]
    async fn click_account_ambiguous_maps_to_error() {
        let h = switch_handler(
            |_, _| json!({"action": "ambiguous", "count": 2, "labels": ["张三\n体验套餐", "张三\n专业版"]}),
        );
        let (session, log) = spawn_mock_page(h).await;
        let err = click_account(&session, "张三", fast_switch())
            .await
            .unwrap_err();
        assert_eq!(err.code, "account_ambiguous");
        let d = err.details.unwrap();
        assert_eq!(d["requested"], "张三");
        assert_eq!(d["matches"], json!(["张三\n体验套餐", "张三\n专业版"]));
        assert!(log.lock().unwrap().contains(&call("closeMenu")));
        session.close().await;
    }

    #[tokio::test]
    async fn click_account_ambiguous_without_labels_falls_back_to_count() {
        let h = switch_handler(|_, _| json!({"action": "ambiguous", "count": 3}));
        let (session, _) = spawn_mock_page(h).await;
        let err = click_account(&session, "张三", fast_switch())
            .await
            .unwrap_err();
        assert_eq!(err.code, "account_ambiguous");
        assert_eq!(err.details.unwrap()["matches"].as_array().unwrap().len(), 3);
        session.close().await;
    }

    #[tokio::test]
    async fn probe_current_injects_then_reads() {
        let h: Handler = Arc::new(Mutex::new(Box::new(|expr: &str| {
            assert!(expr.contains("const VERSION"), "必须带注入守卫");
            assert!(expr.ends_with(".current(); })()"));
            current_json("logged_in", "她是翻哥", "体验套餐")
        })));
        let (session, _) = spawn_mock_page(h).await;
        let c = probe_current(&session).await.unwrap();
        assert_eq!(c.status, "logged_in");
        assert_eq!(c.name, "她是翻哥");
        session.close().await;
    }

    /// 需要真豆包:`cargo test live_ -- --ignored --nocapture`。只读菜单,不点击任何账号。
    /// 持 `LIVE_MENU_LOCK`:与 `connection::live_*` 并行时不能同时操作同一个账号菜单。
    #[tokio::test]
    #[ignore]
    async fn live_read_accounts() {
        use crate::cdp::{list_targets, select_pages};
        let _menu = crate::doubao::LIVE_MENU_LOCK.lock().await;
        let targets = list_targets("127.0.0.1", 9707).await.unwrap();
        let (chat, _) = select_pages(&targets, 9707);
        let chat = chat.first().expect("需要 doubao-chat page");
        let session = CdpSession::connect(&chat.ws_url, &chat.id).await.unwrap();
        ensure_injected(&session).await.unwrap();
        let cur = session.evaluate(&call("current")).await.unwrap();
        println!("current(): {cur}");
        if cur["status"] != "logged_in" {
            // Task 1.2 审查要求:定位失败时打印 probe() 全量(含 candidates)便于排查默认头像按钮
            let probe = session.evaluate(&call("probe")).await.unwrap();
            println!("probe(): {probe}");
        }
        let started = std::time::Instant::now();
        let list = read_accounts(&session).await;
        println!(
            "read_accounts ({}ms): {list:?}",
            started.elapsed().as_millis()
        );
        tokio::time::sleep(Duration::from_millis(350)).await;
        let after = session
            .evaluate(
                "(() => { const m = document.querySelector('[role=menu][data-testid=chat_header_menu]'); if (!m) return false; const r = m.getBoundingClientRect(); return r.width > 0 && r.height > 0; })()",
            )
            .await
            .unwrap();
        println!("menu visible after read: {after}");
        session.close().await;
        let list = list.unwrap();
        assert!(!list.accounts.is_empty());
        assert!(list.current.is_some());
        assert_eq!(after, false, "读取后菜单应已关闭");
    }
}
