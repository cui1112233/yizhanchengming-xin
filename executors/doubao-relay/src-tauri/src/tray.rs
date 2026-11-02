//! 系统托盘(spec 01 F7.5)、关闭窗口最小化到托盘、macOS 应用菜单(Cmd+Q 走关闭流程)。
//!
//! - 托盘:应用图标 + 菜单"显示主窗口 / 退出";提示文字带运行中任务数(随任务事件更新);
//!   双击托盘图标(Windows)也显示主窗口;
//! - 关闭按钮:`ui.minimize_to_tray_on_close=true` → 拦截关闭、隐藏窗口(HTTP 服务继续),
//!   否则照常关闭 → `RunEvent::ExitRequested` → 关闭流程;
//! - macOS:默认菜单的"Quit"是 `NSApp terminate:`,直接 `applicationWillTerminate` 退出、不可拦截,
//!   所以换成自定义"退出 doubao-relay"(Cmd+Q)菜单项,与托盘"退出"同走 [`crate::shutdown::request_quit`];
//!   点 Dock 图标(`RunEvent::Reopen`)重新显示主窗口。
//!
//! 托盘 / 菜单只在生产 `run()` 里创建(mock runtime 测试不碰系统托盘)。

use std::sync::Arc;

use tauri::menu::{Menu, MenuEvent, MenuItem, PredefinedMenuItem};
use tauri::tray::{MouseButton, TrayIconBuilder, TrayIconEvent};
use tauri::{AppHandle, Manager, Runtime, Window, WindowEvent};
use tokio::sync::broadcast::error::RecvError;

use crate::shutdown::{close_action, request_quit, CloseAction};
use crate::state::AppState;

pub const TRAY_ID: &str = "main";
pub const MAIN_WINDOW: &str = "main";
pub const MENU_SHOW: &str = "relay_show_main";
pub const MENU_QUIT: &str = "relay_quit";

/// 托盘提示文字。
pub fn tooltip(running: usize) -> String {
    if running == 0 {
        "一战晟铭执行器 · 空闲".to_string()
    } else {
        format!("一战晟铭执行器 · 运行中 {running} 个任务")
    }
}

/// 显示并聚焦主窗口(托盘"显示主窗口"、双击托盘、Dock 图标)。
pub fn show_main_window<R: Runtime>(app: &AppHandle<R>) {
    if let Some(w) = app.get_webview_window(MAIN_WINDOW) {
        let _ = w.unminimize();
        let _ = w.show();
        let _ = w.set_focus();
    }
}

/// 全局菜单事件(托盘菜单与 macOS 应用菜单共用)。
pub fn handle_menu_event<R: Runtime>(app: &AppHandle<R>, event: MenuEvent) {
    match event.id().as_ref() {
        MENU_SHOW => show_main_window(app),
        MENU_QUIT => request_quit(app),
        _ => {}
    }
}

/// 窗口事件:关闭按钮按配置隐藏到托盘。
pub fn handle_window_event<R: Runtime>(window: &Window<R>, event: &WindowEvent) {
    let WindowEvent::CloseRequested { api, .. } = event else {
        return;
    };
    let Some(state) = window.try_state::<Arc<AppState>>() else {
        return;
    };
    // 关闭流程已开始:放行
    if state.quit.is_started() {
        return;
    }
    // 事件回调在主线程上,不能 await;拿不到读锁(正好在保存配置)按默认 true 处理
    let ui = state
        .config
        .try_read()
        .map(|c| c.ui.clone())
        .unwrap_or_default();
    if close_action(&ui) == CloseAction::HideToTray {
        api.prevent_close();
        let _ = window.hide();
        tracing::info!("主窗口已隐藏到托盘,HTTP 服务继续运行");
    }
}

/// 创建托盘图标与菜单,并按任务事件刷新提示文字。需在 `setup` 里调用。
pub fn setup_tray<R: Runtime>(app: &AppHandle<R>, state: &Arc<AppState>) -> tauri::Result<()> {
    let show = MenuItem::with_id(app, MENU_SHOW, "显示主窗口", true, None::<&str>)?;
    let quit = MenuItem::with_id(app, MENU_QUIT, "退出", true, None::<&str>)?;
    let sep = PredefinedMenuItem::separator(app)?;
    let menu = Menu::with_items(app, &[&show, &sep, &quit])?;
    let mut builder = TrayIconBuilder::with_id(TRAY_ID)
        .menu(&menu)
        .tooltip(tooltip(0))
        .on_tray_icon_event(|tray, event| {
            if let TrayIconEvent::DoubleClick {
                button: MouseButton::Left,
                ..
            } = event
            {
                show_main_window(tray.app_handle());
            }
        });
    if let Some(icon) = app.default_window_icon() {
        builder = builder.icon(icon.clone());
    }
    builder.build(app)?;

    // 提示文字:任务事件到来时按运行中任务数刷新
    let app = app.clone();
    let scheduler = Arc::clone(&state.scheduler);
    let mut rx = scheduler.subscribe();
    tauri::async_runtime::spawn(async move {
        let mut last = usize::MAX;
        while let Ok(_) | Err(RecvError::Lagged(_)) = rx.recv().await {
            let n = scheduler.running_task_ids(None).len();
            if n != last {
                last = n;
                if let Some(tray) = app.tray_by_id(TRAY_ID) {
                    let _ = tray.set_tooltip(Some(tooltip(n)));
                }
            }
        }
    });
    Ok(())
}

/// macOS 应用菜单:系统默认菜单去掉不可拦截的"Quit",换成走关闭流程的"退出 doubao-relay"(Cmd+Q);
/// 保留编辑菜单(WKWebView 里的复制 / 粘贴快捷键依赖它)和窗口菜单(Cmd+W = 关闭按钮,按配置隐藏到托盘)。
#[cfg(target_os = "macos")]
pub fn app_menu<R: Runtime>(app: &AppHandle<R>) -> tauri::Result<Menu<R>> {
    use tauri::menu::Submenu;
    let name = app.package_info().name.clone();
    let quit = MenuItem::with_id(
        app,
        MENU_QUIT,
        format!("退出 {name}"),
        true,
        Some("CmdOrCtrl+Q"),
    )?;
    let app_menu = Submenu::with_items(
        app,
        &name,
        true,
        &[
            &PredefinedMenuItem::about(app, None, None)?,
            &PredefinedMenuItem::separator(app)?,
            &PredefinedMenuItem::hide(app, None)?,
            &PredefinedMenuItem::hide_others(app, None)?,
            &PredefinedMenuItem::show_all(app, None)?,
            &PredefinedMenuItem::separator(app)?,
            &quit,
        ],
    )?;
    let edit = Submenu::with_items(
        app,
        "编辑",
        true,
        &[
            &PredefinedMenuItem::undo(app, None)?,
            &PredefinedMenuItem::redo(app, None)?,
            &PredefinedMenuItem::separator(app)?,
            &PredefinedMenuItem::cut(app, None)?,
            &PredefinedMenuItem::copy(app, None)?,
            &PredefinedMenuItem::paste(app, None)?,
            &PredefinedMenuItem::select_all(app, None)?,
        ],
    )?;
    let window = Submenu::with_items(
        app,
        "窗口",
        true,
        &[
            &PredefinedMenuItem::minimize(app, None)?,
            &PredefinedMenuItem::separator(app)?,
            &PredefinedMenuItem::close_window(app, None)?,
        ],
    )?;
    Menu::with_items(app, &[&app_menu, &edit, &window])
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn tooltip_shows_running_count() {
        assert_eq!(tooltip(0), "一战晟铭执行器 · 空闲");
        assert_eq!(tooltip(2), "一战晟铭执行器 · 运行中 2 个任务");
    }
}
