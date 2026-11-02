//! 豆包侧核心服务:启动/重启豆包(`bootstrap`)、UI 兼容性自检(`compat`)、账号读取(`accounts`),
//! 后续 Task 追加会话管理、生成/捕获管线。
//!
//! 注入豆包 `doubao-chat` 页面的脚本以 `include_str!` 嵌入(`scripts/`),经 `Runtime.evaluate` 注入。

pub mod accounts;
pub mod bootstrap;
pub mod capture;
pub mod compat;
pub mod connection;
pub mod download;
pub mod duration;
pub mod feedback;
pub mod generation;
pub mod upload;
pub mod video_decoder;

pub use accounts::{
    click_account, ensure_injected, parse_entries, parse_label, probe_current, read_accounts,
    Account, AccountList, CurrentProbe, RawEntry, SwitchTiming, EXCLUDED_LABELS,
};
pub use bootstrap::{ensure_doubao_started, wait_for_cdp, StartupAction};
pub use capture::{
    capture, capture_with, extract_image_originals, CaptureEvent, CaptureOptions, CaptureOutcome,
    CaptureTiming, FallbackVideo,
};
pub use compat::{ui_compat_check, UiCompat};
pub use connection::{ConnEvent, ConnTiming, DoubaoConnection, DoubaoStatus, SwitchResult};
pub use duration::{set_duration, DurationTiming};
pub use feedback::{
    map_feedback_stage, refusal_error, should_auto_confirm, FeedbackState,
    RENDERING_UNVERIFIED_LIMIT,
};
pub use generation::{
    build_js_request, generation_bootstrap, is_generation_request, reader_bootstrap,
    step_expression, submit, SubmissionIdentity, SubmissionPlan, SubmissionReport, SubmitTiming,
};
pub use upload::{remove_attachments, upload_references, UploadTiming};

/// 账号操作脚本(`window.__doubaoRelayAccounts`),自带版本守卫,重复注入是 no-op。
pub const ACCOUNTS_SCRIPT: &str = include_str!("scripts/accounts.js");

/// 与 `accounts.js` 里 `VERSION` 一致;注入后可据此断言脚本版本。
pub const ACCOUNTS_SCRIPT_VERSION: &str = "1.0.1";

/// CDP 只连本机回环地址。
pub const CDP_HOST: &str = "127.0.0.1";

/// 进程级锁:所有要操作真豆包账号菜单的 `live_` 测试先 `lock().await` 再跑。
/// `cargo test live_ -- --ignored` 默认多线程并行,`accounts::live_read_accounts` 与
/// `connection::live_*` 同时开/关同一个菜单会互相干扰(`doubao_ui_changed` / `already_open`)。
#[cfg(test)]
pub(crate) static LIVE_MENU_LOCK: tokio::sync::Mutex<()> = tokio::sync::Mutex::const_new(());

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn accounts_script_version_constant_matches_script() {
        assert!(
            ACCOUNTS_SCRIPT.contains(&format!("const VERSION = '{ACCOUNTS_SCRIPT_VERSION}'")),
            "ACCOUNTS_SCRIPT_VERSION 与 accounts.js 的 VERSION 不一致"
        );
    }

    #[test]
    fn accounts_script_closes_a_visible_menu_when_aria_expanded_is_stale() {
        assert!(
            ACCOUNTS_SCRIPT.contains("if (b && menuOpen()) {")
                && ACCOUNTS_SCRIPT.contains("b.click();"),
            "账号菜单仍可见时必须直接点击账号按钮关闭，不能在关闭前再补一组会反向切换菜单的指针事件"
        );
    }

    #[test]
    fn accounts_script_returns_coordinates_for_a_native_menu_click() {
        assert!(
            ACCOUNTS_SCRIPT.contains("openMenu() { const b = findAccountButton();")
                && ACCOUNTS_SCRIPT.contains("if (menuOpen()) return { action: 'already_open' };")
                && ACCOUNTS_SCRIPT.contains("return { action: 'needs_native_click', x: r.x + r.width / 2, y: r.y + r.height / 2 };"),
            "账号菜单打开必须把坐标交给 CDP 的可信鼠标事件，不能在页面脚本内伪造点击"
        );
    }

    #[test]
    fn accounts_script_ignores_a_radix_menu_after_it_has_closed() {
        assert!(
            ACCOUNTS_SCRIPT.contains("m.getAttribute('data-state') !== 'closed'"),
            "Radix 关闭动画会保留菜单节点，不能仅按矩形尺寸把已关闭菜单当作仍打开"
        );
    }

    #[test]
    fn accounts_script_ignores_items_left_inside_a_closed_radix_menu() {
        assert!(
            ACCOUNTS_SCRIPT.contains("el.closest(MENU_ROOT)")
                && ACCOUNTS_SCRIPT.contains("getAttribute('data-state') !== 'closed'"),
            "关闭动画遗留的菜单项也不能继续触发账号菜单打开状态"
        );
    }
}
