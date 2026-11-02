//! 参考图上传(Task 3.2),按 analysis/04 "Step 3-4" 与沈浪 `doubao-generate.ps1`:
//!
//! `Page.enable` + `Page.setInterceptFileChooserDialog{enabled:true}` → 对每个文件依次:
//! 订阅事件 → `run('fileInput')` → 若期间收到 `Page.fileChooserOpened` 就用它的 `backendNodeId`,
//! 否则 `DOM.getDocument` + `DOM.querySelector('input[data-shenlang-reference-upload="true"]')`
//! → `DOM.setFileInputFiles` → `request.uploadedCount = i+1` → `run('uploadReady')`(70s)。
//! 结束(成功或失败)都 `setInterceptFileChooserDialog{enabled:false}`。
//!
//! 真机事实(`docs/analysis/10-live-probe-macos.md` Task 2.1 / 3.2):`fileInput` 直接给输入区里已存在的
//! `input[type=file]`(`accept=".jpg,.png,.jpeg,.webp,.apng"`,multiple)打标记,**不点按钮**,
//! 不会触发 `Page.fileChooserOpened`,所以主路径是 `DOM.querySelector` + `DOM.setFileInputFiles`,
//! `fileChooserOpened` 只是兜底。沈浪在 Windows 上还有 UIAutomation 操作原生"打开"对话框的第三条
//! 路径,relay 未实现(两个平台的豆包 2.32.x 都是页面内 input;入口变化时报 `upload_entry_changed`)。

use std::path::PathBuf;
use std::time::Duration;

use serde_json::{json, Value};
use tokio::sync::broadcast;
use tokio_util::sync::CancellationToken;

use crate::cdp::session::{extract_js_exception, CdpEvent};
use crate::cdp::{CdpError, CdpSession};
use crate::doubao::generation::{
    cancelled, evaluate_step, map_preparing, preparing_failed, preparing_timeout, step_expression,
};
use crate::error::{RelayError, RelayResult};

/// `fileInput` 打了标记的参考图 input。
pub const REFERENCE_INPUT_SELECTOR: &str = r#"input[data-shenlang-reference-upload="true"]"#;

/// 输入区里的参考图卡片(豆包 2.32.5 观察值,`aria-label` 为文件名):图片卡片与视频卡片。
///
/// 真机(Task 3.2 复审):视频技能接受 mp4 参考视频,显示为 `attachment-video-card`
/// (在 `video-attachment-scroll-container` 里),删除按钮同为 `attachment-delete-btn`。
pub const ATTACHMENT_CARD_SELECTORS: [&str; 2] = [
    r#"[data-testid="attachment-image-card"]"#,
    r#"[data-testid="attachment-video-card"]"#,
];
/// 卡片里的删除按钮(`svg`,只在悬停时有尺寸,派发点击事件即可删除)。
pub const ATTACHMENT_DELETE_SELECTOR: &str = r#"[data-testid="attachment-delete-btn"]"#;

/// `prepare` 前 / 重试前删除残留卡片的等待上限(JS 侧最多 3 轮 × 1s)。
pub const REMOVE_TIMEOUT: Duration = Duration::from_secs(5);
/// 失败后清理(删卡片、关拦截、清输入框)的单步上限:要能放进 runner 取消后的 5s 宽限期。
/// 超时只是不再等,页面里的删除脚本照样跑完。
pub const CLEANUP_TIMEOUT: Duration = Duration::from_secs(2);

/// 删除输入区里所有参考图 / 参考视频卡片(每张点删除、等卡片消失,最多 3 轮),返回 `{before, after}`。
///
/// 真机(Task 3.2):参考图属于**当前技能的草稿**,退出技能后卡片仍显示、但在普通对话态删除不生效
/// (再进技能又出现);必须在添加它的那个技能里删除。残留卡片会让下一次 `prepare` 失败
/// (普通对话态 12s 后 `新图片对话未准备好`,图像技能 `豆包输入框已有参考图片，请先移除后重试。`,
/// 视频技能 `视频输入区已有内容，任务未发送`)。
pub fn remove_attachments_expression() -> String {
    let cards = serde_json::to_string(&ATTACHMENT_CARD_SELECTORS).unwrap_or_else(|_| "[]".into());
    format!(
        r#"(async () => {{
  const CARDS = {cards};
  const DEL = {del:?};
  const CARD_SEL = CARDS.join(',');
  const DEL_SEL = CARDS.map(c => c + ' ' + DEL).join(',');
  const sleep = ms => new Promise(r => setTimeout(r, ms));
  const press = el => {{
    const p = {{ bubbles: true, cancelable: true, button: 0 }};
    if (typeof PointerEvent !== 'undefined') el.dispatchEvent(new PointerEvent('pointerdown', {{ ...p, buttons: 1, pointerId: 1, pointerType: 'mouse', isPrimary: true }}));
    el.dispatchEvent(new MouseEvent('mousedown', {{ ...p, buttons: 1 }}));
    if (typeof PointerEvent !== 'undefined') el.dispatchEvent(new PointerEvent('pointerup', {{ ...p, buttons: 0, pointerId: 1, pointerType: 'mouse', isPrimary: true }}));
    el.dispatchEvent(new MouseEvent('mouseup', {{ ...p, buttons: 0 }}));
    el.dispatchEvent(new MouseEvent('click', {{ ...p, buttons: 0 }}));
  }};
  const cards = () => document.querySelectorAll(CARD_SEL).length;
  const before = cards();
  if (!before) return {{ before: 0, after: 0 }};
  for (let round = 0; round < 3 && cards(); round++) {{
    for (const b of [...document.querySelectorAll(DEL_SEL)]) {{ press(b); await sleep(120); }}
    const t0 = Date.now();
    while (cards() && Date.now() - t0 < 1000) await sleep(50);
  }}
  return {{ before, after: cards() }};
}})()"#,
        del = ATTACHMENT_DELETE_SELECTOR,
    )
}

/// 尽力删除输入区里的参考图卡片(最多等 [`REMOVE_TIMEOUT`]);返回删除前的张数(读不到为 0)。
pub async fn remove_attachments(session: &CdpSession) -> u64 {
    remove_attachments_within(session, REMOVE_TIMEOUT).await
}

/// 失败后的清理版本(最多等 [`CLEANUP_TIMEOUT`])。
pub async fn cleanup_attachments(session: &CdpSession) -> u64 {
    remove_attachments_within(session, CLEANUP_TIMEOUT).await
}

/// 见 [`remove_attachments`]。失败只记日志。
pub async fn remove_attachments_within(session: &CdpSession, timeout: Duration) -> u64 {
    let result = session
        .send_timeout(
            "Runtime.evaluate",
            json!({
                "expression": remove_attachments_expression(),
                "returnByValue": true,
                "awaitPromise": true,
            }),
            timeout,
        )
        .await;
    let value = match result {
        Ok(r) => match r.get("exceptionDetails") {
            Some(details) => Err(CdpError::JsException(extract_js_exception(details))),
            None => Ok(r
                .get("result")
                .and_then(|r| r.get("value"))
                .cloned()
                .unwrap_or(Value::Null)),
        },
        Err(e) => Err(e),
    };
    match value {
        Ok(v) => {
            let before = v.get("before").and_then(Value::as_u64).unwrap_or(0);
            let after = v.get("after").and_then(Value::as_u64).unwrap_or(0);
            if before > 0 {
                if after == 0 {
                    tracing::info!(removed = before, "已删除豆包输入区的参考图");
                } else {
                    tracing::warn!(before, after, "豆包输入区的参考图未能全部删除");
                }
            }
            before
        }
        Err(e) => {
            tracing::warn!(error = %e, "删除豆包输入区参考图失败");
            0
        }
    }
}

/// 各步超时(默认值取自 PS),测试用小值。
#[derive(Clone, Debug)]
pub struct UploadTiming {
    /// `fileInput` 与 CDP 命令的超时(PS `Step` 默认 20s)。
    pub step_timeout: Duration,
    /// `uploadReady`(PS 70s:等豆包把这张图传完)。
    pub ready_timeout: Duration,
}

impl Default for UploadTiming {
    fn default() -> Self {
        Self {
            step_timeout: Duration::from_secs(20),
            ready_timeout: Duration::from_secs(70),
        }
    }
}

/// `generation_preparing_failed{reason:"upload_entry_changed"}`:既没有 chooser 事件也找不到标记的 input。
pub fn upload_entry_changed() -> RelayError {
    RelayError::new(
        "generation_preparing_failed",
        "参考图上传入口已变化,任务未发送,请升级 doubao-relay 或去掉参考图后重试",
    )
    .with_details(json!({
        "reason": "upload_entry_changed",
        "js_exception": Value::Null,
        "failed_selectors": [REFERENCE_INPUT_SELECTOR],
    }))
}

/// 上传参考图(默认时间参数)。见 [`upload_references_with`]。
pub async fn upload_references(
    session: &CdpSession,
    tracking_key: &str,
    request_json: &mut Value,
    files: &[PathBuf],
    cancel: &CancellationToken,
    on_progress: &(dyn Fn(usize, usize) + Send + Sync),
) -> RelayResult<()> {
    upload_references_with(
        session,
        tracking_key,
        request_json,
        files,
        cancel,
        on_progress,
        &UploadTiming::default(),
    )
    .await
}

/// 见模块文档。`files` 为空时什么也不做。成功后 `request_json.uploadedCount == files.len()`,
/// 之后的 `fill` / `checkSend` / `send` 必须用这份请求(脚本据此核对附件数)。
pub async fn upload_references_with(
    session: &CdpSession,
    tracking_key: &str,
    request_json: &mut Value,
    files: &[PathBuf],
    cancel: &CancellationToken,
    on_progress: &(dyn Fn(usize, usize) + Send + Sync),
    timing: &UploadTiming,
) -> RelayResult<()> {
    if files.is_empty() {
        return Ok(());
    }
    if cancel.is_cancelled() {
        return Err(cancelled("upload"));
    }
    command_soft(session, "Page.enable", json!({}), timing).await?;
    command_soft(
        session,
        "Page.setInterceptFileChooserDialog",
        json!({ "enabled": true }),
        timing,
    )
    .await?;
    let started = std::time::Instant::now();
    let result = upload_all(
        session,
        tracking_key,
        request_json,
        files,
        cancel,
        on_progress,
        timing,
    )
    .await;
    if let Err(e) = session
        .send_timeout(
            "Page.setInterceptFileChooserDialog",
            json!({ "enabled": false }),
            timing.step_timeout.min(CLEANUP_TIMEOUT),
        )
        .await
    {
        tracing::warn!(error = %e, "关闭文件选择框拦截失败");
    }
    match &result {
        Ok(()) => tracing::info!(
            files = files.len(),
            elapsed_ms = started.elapsed().as_millis() as u64,
            "参考图已全部上传"
        ),
        Err(e) => tracing::warn!(code = e.code, error = %e, files = files.len(), "参考图上传失败"),
    }
    result
}

async fn upload_all(
    session: &CdpSession,
    tracking_key: &str,
    request_json: &mut Value,
    files: &[PathBuf],
    cancel: &CancellationToken,
    on_progress: &(dyn Fn(usize, usize) + Send + Sync),
    timing: &UploadTiming,
) -> RelayResult<()> {
    let total = files.len();
    for (index, file) in files.iter().enumerate() {
        if cancel.is_cancelled() {
            return Err(cancelled("upload"));
        }
        let path = file.to_string_lossy().into_owned();
        open_upload_entry(session, timing).await;
        // 订阅要在 fileInput 之前:chooser 事件在 evaluate 的响应之前到达。
        let mut rx = session.subscribe();
        evaluate_step(
            session,
            &step_expression(tracking_key, "fileInput", request_json),
            timing.step_timeout,
            cancel,
        )
        .await
        .map_err(|e| map_preparing("fileInput", e))?;
        let chooser = take_chooser_backend_node(&mut rx);

        let target = match chooser {
            Some(backend) => json!({ "backendNodeId": backend }),
            None => match marked_input_node(session, timing).await? {
                Some(node) => json!({ "nodeId": node }),
                None => return Err(upload_entry_changed()),
            },
        };
        let mut params = target.clone();
        params["files"] = json!([path]);
        command(session, "DOM.setFileInputFiles", params, timing).await?;
        tracing::debug!(index = index + 1, total, via = %target, "已把参考图交给豆包");

        request_json["uploadedCount"] = json!(index + 1);
        evaluate_step(
            session,
            &step_expression(tracking_key, "uploadReady", request_json),
            timing.ready_timeout,
            cancel,
        )
        .await
        .map_err(|e| map_preparing("uploadReady", e))?;
        on_progress(index + 1, total);
    }
    Ok(())
}

/// 取走接收器里已到达的 `Page.fileChooserOpened`,返回其 `backendNodeId`(没有则 None)。
fn take_chooser_backend_node(rx: &mut broadcast::Receiver<CdpEvent>) -> Option<i64> {
    let mut found = None;
    loop {
        match rx.try_recv() {
            Ok(ev) if ev.method == "Page.fileChooserOpened" => {
                if let Some(id) = ev.params.get("backendNodeId").and_then(Value::as_i64) {
                    found = Some(id);
                }
            }
            Ok(_) => {}
            Err(broadcast::error::TryRecvError::Lagged(_)) => {}
            Err(_) => return found,
        }
    }
}

/// `DOM.getDocument{depth:0}` → `DOM.querySelector(REFERENCE_INPUT_SELECTOR)`;nodeId 为 0 → None。
/// 豆包 2.32.6 起技能输入区不再常驻 `input[type=file]`:点"+"(`upload_file_button`,
/// `aria-haspopup=menu`)→ 菜单项"添加图片"(图像技能)/"上传素材"(视频技能)后才生成(真机实测,见 docs/analysis/10)。
/// 页面刷新后 JS 合成的 pointer 事件打不开这个菜单,必须用 CDP `Input.dispatchMouseEvent`(可信事件)点"+"。
/// 第一步:已有 file input / 菜单是否已打开 / "+" 的中心坐标。
pub fn upload_entry_probe_expression() -> &'static str {
    r#"(() => {
  const visible = e => { const r = e.getBoundingClientRect(); return r.width > 0 && r.height > 0; };
  if (document.querySelectorAll('input[type=file]').length > 0) return { existing: true };
  const menuOpen = [...document.querySelectorAll('[role=menuitem]')].filter(visible).some(m => /^(添加图片|上传素材)$/.test((m.innerText || '').trim()));
  const btn = [...document.querySelectorAll('[data-testid=upload_file_button]')].find(visible);
  if (!btn) return { existing: false, menuOpen, rect: null };
  const r = btn.getBoundingClientRect();
  return { existing: false, menuOpen, rect: { x: r.x + r.width / 2, y: r.y + r.height / 2 } };
})()"#
}

/// 第二步:等菜单项"添加图片"出现 → 点击 → 等 file input 生成 → Escape 收起菜单。
/// 调用前文件选择框拦截已开启,点"添加图片"不会弹出系统对话框。
pub fn upload_entry_pick_expression() -> &'static str {
    r#"(async () => {
  const visible = e => { const r = e.getBoundingClientRect(); return r.width > 0 && r.height > 0; };
  const sleep = ms => new Promise(r => setTimeout(r, ms));
  const waitFor = async (f, ms) => { const end = Date.now() + ms; while (Date.now() < end) { const v = f(); if (v) return v; await sleep(50); } return null; };
  const findItem = () => [...document.querySelectorAll('[role=menuitem]')].filter(visible).find(m => /^(添加图片|上传素材)$/.test((m.innerText || '').trim()));
  let item = await waitFor(findItem, 600);
  // 窗口不在前台时第一下可信点击只激活窗口(真机 2.32.6 实测),之后再补发 pointer 事件才能打开菜单
  for (let attempt = 0; !item && attempt < 3; attempt++) {
    const btn = [...document.querySelectorAll('[data-testid=upload_file_button]')].find(visible);
    if (!btn) break;
    const r = btn.getBoundingClientRect();
    const o = { bubbles: true, cancelable: true, pointerId: 1, pointerType: 'mouse', isPrimary: true, button: 0, clientX: r.x + r.width / 2, clientY: r.y + r.height / 2 };
    btn.dispatchEvent(new PointerEvent('pointerdown', { ...o, buttons: 1 }));
    btn.dispatchEvent(new PointerEvent('pointerup', { ...o, buttons: 0 }));
    item = await waitFor(findItem, 1000);
  }
  if (!item) return { action: 'no_menu_item' };
  item.click();
  const ok = await waitFor(() => document.querySelectorAll('input[type=file]').length > 0, 3000);
  document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', code: 'Escape', keyCode: 27, bubbles: true }));
  return { action: ok ? 'opened' : 'no_input' };
})()"#
}

async fn eval_value(session: &CdpSession, expr: &str, timeout: Duration) -> Option<Value> {
    let r = session
        .send_timeout(
            "Runtime.evaluate",
            json!({ "expression": expr, "returnByValue": true, "awaitPromise": true }),
            timeout,
        )
        .await
        .ok()?;
    r.pointer("/result/value").cloned()
}

/// 已有 file input 时什么也不做;任何一步失败只记日志,交给 `fileInput` 步骤报原有错误。
async fn open_upload_entry(session: &CdpSession, timing: &UploadTiming) {
    let Some(probe) = eval_value(
        session,
        upload_entry_probe_expression(),
        timing.step_timeout,
    )
    .await
    else {
        tracing::warn!("读取参考图上传入口状态失败,继续按原流程");
        return;
    };
    if probe.get("existing").and_then(Value::as_bool) == Some(true) {
        tracing::debug!("参考图上传框已存在");
        return;
    }
    let menu_open = probe.get("menuOpen").and_then(Value::as_bool) == Some(true);
    if !menu_open {
        let (Some(x), Some(y)) = (
            probe.pointer("/rect/x").and_then(Value::as_f64),
            probe.pointer("/rect/y").and_then(Value::as_f64),
        ) else {
            tracing::warn!("未找到参考图上传按钮(upload_file_button)");
            return;
        };
        for (kind, buttons) in [("mouseMoved", 0), ("mousePressed", 1), ("mouseReleased", 0)] {
            let mut p = json!({ "type": kind, "x": x, "y": y });
            if kind != "mouseMoved" {
                p["button"] = json!("left");
                p["buttons"] = json!(buttons);
                p["clickCount"] = json!(1);
            }
            if let Err(e) = session
                .send_timeout("Input.dispatchMouseEvent", p, timing.step_timeout)
                .await
            {
                tracing::warn!(error = %e, "点击参考图上传按钮失败");
                return;
            }
        }
    }
    let action = eval_value(session, upload_entry_pick_expression(), timing.step_timeout)
        .await
        .and_then(|v| v.get("action").and_then(Value::as_str).map(str::to_string))
        .unwrap_or_else(|| "unknown".into());
    if action == "opened" {
        tracing::debug!("已打开参考图上传入口");
    } else {
        tracing::warn!(action = %action, "未能打开参考图上传入口");
    }
}

async fn marked_input_node(
    session: &CdpSession,
    timing: &UploadTiming,
) -> RelayResult<Option<i64>> {
    let doc = command(session, "DOM.getDocument", json!({ "depth": 0 }), timing).await?;
    let Some(root) = doc
        .get("root")
        .and_then(|r| r.get("nodeId"))
        .and_then(Value::as_i64)
    else {
        return Ok(None);
    };
    let found = command(
        session,
        "DOM.querySelector",
        json!({ "nodeId": root, "selector": REFERENCE_INPUT_SELECTOR }),
        timing,
    )
    .await?;
    Ok(found
        .get("nodeId")
        .and_then(Value::as_i64)
        .filter(|id| *id > 0))
}

/// CDP 命令:Protocol 错误 → `generation_preparing_failed{reason:"cdp_command_failed"}`,
/// 超时 → `generation_preparing_timeout{step:"upload"}`,断线照常映射。
async fn command(
    session: &CdpSession,
    method: &str,
    params: Value,
    timing: &UploadTiming,
) -> RelayResult<Value> {
    match session
        .send_timeout(method, params, timing.step_timeout)
        .await
    {
        Ok(v) => Ok(v),
        Err(CdpError::Protocol { code, message }) => {
            Err(
                preparing_failed("cdp_command_failed", None).with_details(json!({
                    "reason": "cdp_command_failed",
                    "js_exception": Value::Null,
                    "method": method,
                    "code": code,
                    "message": message,
                })),
            )
        }
        Err(CdpError::Timeout(_)) => Err(preparing_timeout("upload")),
        Err(e) => Err(e.into()),
    }
}

/// 可选的准备命令(`Page.enable` / 开拦截):Protocol 错误只 warn(主路径不依赖 chooser 事件)。
async fn command_soft(
    session: &CdpSession,
    method: &str,
    params: Value,
    timing: &UploadTiming,
) -> RelayResult<()> {
    match command(session, method, params, timing).await {
        Ok(_) => Ok(()),
        Err(e) if e.code == "generation_preparing_failed" => {
            tracing::warn!(method, error = %e, "CDP 命令被拒绝,忽略");
            Ok(())
        }
        Err(e) => Err(e),
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::doubao::duration::mock_page::{names, spawn, step_of, Handler, Log, Reply};
    use std::sync::{Arc, Mutex};

    const KEY: &str = "shenlangDoubaoTask_01test";

    fn fast() -> UploadTiming {
        UploadTiming {
            step_timeout: Duration::from_millis(500),
            ready_timeout: Duration::from_millis(500),
        }
    }

    fn req(files: &[&str]) -> Value {
        json!({ "type": "image", "modelName": "Seedream 5.0 Pro", "referenceFiles": files, "uploadedCount": 0 })
    }

    /// handler:`fileInput` 时可选推 chooser 事件;DOM 命令按 `marked` 回 nodeId。
    fn handler(chooser: bool, marked: i64, ready: Option<&'static str>) -> Handler {
        Box::new(move |method, params| match method {
            "Runtime.evaluate" => match step_of(params["expression"].as_str().unwrap()).as_str() {
                "fileInput" => {
                    let r = Reply::value(json!({ "multiple": true }));
                    if chooser {
                        r.with_events(vec![
                            json!({ "method": "Network.dataReceived", "params": {} }),
                            json!({ "method": "Page.fileChooserOpened",
                                    "params": { "frameId": "F", "mode": "selectMultiple", "backendNodeId": 4242 } }),
                        ])
                    } else {
                        r
                    }
                }
                "uploadReady" => match ready {
                    Some(msg) => Reply::throw(msg),
                    None => Reply::value(json!({ "ready": true })),
                },
                _ => Reply::value(Value::Null),
            },
            "DOM.getDocument" => Reply::raw(json!({ "root": { "nodeId": 1, "backendNodeId": 1 } })),
            "DOM.querySelector" => Reply::raw(json!({ "nodeId": marked })),
            _ => Reply::raw(json!({})),
        })
    }

    type Progress = Arc<Mutex<Vec<(usize, usize)>>>;

    fn progress() -> (Progress, impl Fn(usize, usize) + Send + Sync) {
        let p: Progress = Arc::new(Mutex::new(Vec::new()));
        let p2 = Arc::clone(&p);
        (p, move |a, b| p2.lock().unwrap().push((a, b)))
    }

    fn params_of(log: &Log, name: &str) -> Vec<Value> {
        log.lock()
            .unwrap()
            .iter()
            .filter(|(n, _)| n == name)
            .map(|(_, p)| p.clone())
            .collect()
    }

    #[tokio::test]
    async fn file_chooser_event_uses_backend_node_id() {
        let (session, log) = spawn(handler(true, 0, None)).await;
        let files = [PathBuf::from("/tmp/ref-a.png")];
        let mut request = req(&["/tmp/ref-a.png"]);
        let (prog, on_progress) = progress();
        upload_references_with(
            &session,
            KEY,
            &mut request,
            &files,
            &CancellationToken::new(),
            &on_progress,
            &fast(),
        )
        .await
        .unwrap();
        let set = params_of(&log, "DOM.setFileInputFiles");
        assert_eq!(set.len(), 1);
        assert_eq!(set[0]["backendNodeId"], 4242);
        assert_eq!(set[0]["files"], json!(["/tmp/ref-a.png"]));
        assert!(set[0].get("nodeId").is_none());
        // chooser 命中时不走 DOM.querySelector
        assert!(!names(&log).iter().any(|n| n == "DOM.querySelector"));
        assert_eq!(request["uploadedCount"], 1);
        assert_eq!(*prog.lock().unwrap(), vec![(1, 1)]);
        let names = names(&log);
        assert_eq!(names.first().unwrap(), "Page.enable");
        let intercept = params_of(&log, "Page.setInterceptFileChooserDialog");
        assert_eq!(
            intercept,
            vec![json!({ "enabled": true }), json!({ "enabled": false })]
        );
        assert_eq!(names.last().unwrap(), "Page.setInterceptFileChooserDialog");
        session.close().await;
    }

    /// 豆包 2.32.6:图像/视频技能里没有常驻的 `input[type=file]`,要先点"+"(`upload_file_button`)
    /// → "添加图片" 才生成。每张图的 `fileInput` 之前都要先确保入口已打开。
    #[tokio::test]
    async fn upload_menu_is_opened_before_each_file_input() {
        let (session, log) = spawn(handler(false, 77, None)).await;
        let files = [PathBuf::from("/tmp/a.png"), PathBuf::from("/tmp/b.png")];
        let mut request = req(&["/tmp/a.png", "/tmp/b.png"]);
        let (_prog, on_progress) = progress();
        upload_references_with(
            &session,
            KEY,
            &mut request,
            &files,
            &CancellationToken::new(),
            &on_progress,
            &fast(),
        )
        .await
        .unwrap();
        let evals: Vec<String> = log
            .lock()
            .unwrap()
            .iter()
            .filter(|(n, _)| n.starts_with("eval:"))
            .map(|(_, p)| p["expression"].as_str().unwrap().to_string())
            .collect();
        let order: Vec<&str> = evals
            .iter()
            .filter_map(|e| {
                if e.contains("upload_file_button") {
                    Some("open")
                } else if e.contains("\"fileInput\"") {
                    Some("fileInput")
                } else {
                    None
                }
            })
            .collect();
        assert_eq!(
            order,
            ["open", "fileInput", "open", "fileInput"],
            "{evals:?}"
        );
        let open = evals
            .iter()
            .find(|e| e.contains("upload_file_button"))
            .unwrap();
        assert!(open.contains("添加图片"), "{open}");
        session.close().await;
    }

    /// 入口没打开时用可信鼠标事件点 "+" 的中心坐标,再点"添加图片"。
    #[tokio::test]
    async fn upload_button_is_clicked_with_trusted_mouse_events() {
        let mut inner = handler(false, 77, None);
        let (session, log) = spawn(Box::new(move |method, params| {
            if method == "Runtime.evaluate"
                && params["expression"]
                    .as_str()
                    .unwrap()
                    .contains("upload_file_button")
            {
                return Reply::value(
                    json!({ "existing": false, "menuOpen": false, "rect": { "x": 424.0, "y": 835.0 } }),
                );
            }
            inner(method, params)
        }))
        .await;
        let files = [PathBuf::from("/tmp/a.png")];
        let mut request = req(&["/tmp/a.png"]);
        let (_prog, on_progress) = progress();
        upload_references_with(
            &session,
            KEY,
            &mut request,
            &files,
            &CancellationToken::new(),
            &on_progress,
            &fast(),
        )
        .await
        .unwrap();
        let mouse = params_of(&log, "Input.dispatchMouseEvent");
        let kinds: Vec<&str> = mouse.iter().map(|p| p["type"].as_str().unwrap()).collect();
        assert_eq!(kinds, ["mouseMoved", "mousePressed", "mouseReleased"]);
        assert_eq!(
            (mouse[1]["x"].as_f64(), mouse[1]["y"].as_f64()),
            (Some(424.0), Some(835.0))
        );
        assert_eq!(mouse[1]["button"], "left");
        let picked = log.lock().unwrap().iter().any(|(_, p)| {
            p["expression"]
                .as_str()
                .is_some_and(|e| e.contains("添加图片") && e.contains("item.click()"))
        });
        assert!(picked);
        session.close().await;
    }

    #[tokio::test]
    async fn marked_input_is_main_path_and_files_go_one_by_one() {
        let (session, log) = spawn(handler(false, 77, None)).await;
        let files = [PathBuf::from("/tmp/a.png"), PathBuf::from("/tmp/b.jpg")];
        let mut request = req(&["/tmp/a.png", "/tmp/b.jpg"]);
        let (prog, on_progress) = progress();
        upload_references_with(
            &session,
            KEY,
            &mut request,
            &files,
            &CancellationToken::new(),
            &on_progress,
            &fast(),
        )
        .await
        .unwrap();
        let seq: Vec<String> = names(&log)
            .into_iter()
            .filter(|n| {
                (n.starts_with("eval:") && n != "eval:other") || n == "DOM.setFileInputFiles"
            })
            .collect();
        assert_eq!(
            seq,
            [
                "eval:fileInput",
                "DOM.setFileInputFiles",
                "eval:uploadReady",
                "eval:fileInput",
                "DOM.setFileInputFiles",
                "eval:uploadReady"
            ]
        );
        let set = params_of(&log, "DOM.setFileInputFiles");
        assert_eq!(set[0], json!({ "nodeId": 77, "files": ["/tmp/a.png"] }));
        assert_eq!(set[1], json!({ "nodeId": 77, "files": ["/tmp/b.jpg"] }));
        let q = params_of(&log, "DOM.querySelector");
        assert_eq!(q[0]["selector"], REFERENCE_INPUT_SELECTOR);
        assert_eq!(q[0]["nodeId"], 1);
        // uploadReady 带递增的 uploadedCount
        let ready: Vec<String> = log
            .lock()
            .unwrap()
            .iter()
            .filter(|(n, _)| n == "eval:uploadReady")
            .map(|(_, p)| p["expression"].as_str().unwrap().to_string())
            .collect();
        assert!(ready[0].contains("\"uploadedCount\":1"), "{}", ready[0]);
        assert!(ready[1].contains("\"uploadedCount\":2"), "{}", ready[1]);
        assert!(ready[0].contains(&format!("{KEY}.run(\"uploadReady\"")));
        assert_eq!(request["uploadedCount"], 2);
        assert_eq!(*prog.lock().unwrap(), vec![(1, 2), (2, 2)]);
        session.close().await;
    }

    #[tokio::test]
    async fn missing_entry_is_upload_entry_changed_and_interception_is_disabled() {
        let (session, log) = spawn(handler(false, 0, None)).await;
        let files = [PathBuf::from("/tmp/a.png")];
        let mut request = req(&["/tmp/a.png"]);
        let (prog, on_progress) = progress();
        let err = upload_references_with(
            &session,
            KEY,
            &mut request,
            &files,
            &CancellationToken::new(),
            &on_progress,
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "generation_preparing_failed");
        let d = err.details.unwrap();
        assert_eq!(d["reason"], "upload_entry_changed");
        assert_eq!(d["failed_selectors"][0], REFERENCE_INPUT_SELECTOR);
        assert!(!names(&log).iter().any(|n| n == "DOM.setFileInputFiles"));
        assert_eq!(
            params_of(&log, "Page.setInterceptFileChooserDialog")
                .last()
                .unwrap(),
            &json!({ "enabled": false })
        );
        assert!(prog.lock().unwrap().is_empty());
        session.close().await;
    }

    #[tokio::test]
    async fn upload_ready_exception_is_preparing_failed() {
        let (session, log) = spawn(handler(
            false,
            5,
            Some("参考图尚未全部上传完成，未发送，请检查豆包上传状态。"),
        ))
        .await;
        let files = [PathBuf::from("/tmp/a.png")];
        let mut request = req(&["/tmp/a.png"]);
        let (_p, on_progress) = progress();
        let err = upload_references_with(
            &session,
            KEY,
            &mut request,
            &files,
            &CancellationToken::new(),
            &on_progress,
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "generation_preparing_failed");
        let d = err.details.unwrap();
        assert_eq!(d["reason"], "uploadReady");
        assert!(d["js_exception"]
            .as_str()
            .unwrap()
            .contains("参考图尚未全部上传完成"));
        assert_eq!(
            names(&log).last().unwrap(),
            "Page.setInterceptFileChooserDialog"
        );
        session.close().await;
    }

    #[tokio::test]
    async fn empty_files_send_nothing_and_cancel_still_disables_interception() {
        let (session, log) = spawn(handler(false, 5, None)).await;
        let mut request = req(&[]);
        let (_p, on_progress) = progress();
        upload_references_with(
            &session,
            KEY,
            &mut request,
            &[],
            &CancellationToken::new(),
            &on_progress,
            &fast(),
        )
        .await
        .unwrap();
        assert!(log.lock().unwrap().is_empty());

        // fileInput 期间取消 → cancelled,拦截照样关掉
        let cancel = CancellationToken::new();
        let c2 = cancel.clone();
        let (session2, log2) = spawn(Box::new(move |method, _| {
            if method == "Page.setInterceptFileChooserDialog" {
                c2.cancel();
            }
            Reply::raw(json!({}))
        }))
        .await;
        let err = upload_references_with(
            &session2,
            KEY,
            &mut req(&["/tmp/a.png"]),
            &[PathBuf::from("/tmp/a.png")],
            &cancel,
            &on_progress,
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "generation_cancelled");
        assert_eq!(
            params_of(&log2, "Page.setInterceptFileChooserDialog"),
            vec![json!({ "enabled": true }), json!({ "enabled": false })]
        );
        session.close().await;
        session2.close().await;
    }

    #[test]
    fn remove_attachments_expression_targets_delete_buttons_inside_cards() {
        let e = remove_attachments_expression();
        assert!(
            e.contains(
                r#"const CARDS = ["[data-testid=\"attachment-image-card\"]","[data-testid=\"attachment-video-card\"]"];"#
            ),
            "{e}"
        );
        assert!(
            e.contains(r#"const DEL = "[data-testid=\"attachment-delete-btn\"]";"#),
            "{e}"
        );
        assert!(e.contains("CARDS.map(c => c + ' ' + DEL).join(',')"));
        assert!(e.contains("querySelectorAll(DEL_SEL)"));
        assert!(e.contains("querySelectorAll(CARD_SEL)"));
        assert!(e.starts_with("(async () => {"));
    }

    #[tokio::test]
    async fn remove_attachments_reports_count_and_never_fails() {
        let (session, _log) = spawn(Box::new(|_, _| {
            Reply::value(json!({ "before": 2, "after": 0 }))
        }))
        .await;
        assert_eq!(remove_attachments(&session).await, 2);
        session.close().await;
        let (session, _log) = spawn(Box::new(|_, _| Reply::throw("boom"))).await;
        assert_eq!(remove_attachments(&session).await, 0);
        session.close().await;
    }

    #[tokio::test]
    async fn cleanup_gives_up_within_runner_grace_when_page_hangs() {
        // 页面卡住:删卡片的 evaluate 与关拦截都不回应,清理仍要在 runner 的 5s 宽限期内结束
        let (session, _log) = spawn(Box::new(|_, _| Reply::hang())).await;
        let t0 = std::time::Instant::now();
        assert_eq!(cleanup_attachments(&session).await, 0);
        let took = t0.elapsed();
        assert!(
            took >= CLEANUP_TIMEOUT && took < Duration::from_millis(2_800),
            "{took:?}"
        );
        session.close().await;

        // uploadReady 失败后关拦截用 min(step_timeout, CLEANUP_TIMEOUT),不是 20s
        let (session, log) = spawn(Box::new(|method, params| match method {
            "Page.setInterceptFileChooserDialog" if params["enabled"] == false => Reply::hang(),
            "Runtime.evaluate" => match step_of(params["expression"].as_str().unwrap()).as_str() {
                "uploadReady" => Reply::throw("参考图尚未全部上传完成"),
                _ => Reply::value(json!({})),
            },
            "DOM.getDocument" => Reply::raw(json!({ "root": { "nodeId": 1 } })),
            "DOM.querySelector" => Reply::raw(json!({ "nodeId": 9 })),
            _ => Reply::raw(json!({})),
        }))
        .await;
        let (_p, on_progress) = progress();
        let slow = UploadTiming {
            step_timeout: Duration::from_secs(20),
            ready_timeout: Duration::from_secs(20),
        };
        let t0 = std::time::Instant::now();
        let err = upload_references_with(
            &session,
            KEY,
            &mut req(&["/tmp/a.png"]),
            &[PathBuf::from("/tmp/a.png")],
            &CancellationToken::new(),
            &on_progress,
            &slow,
        )
        .await
        .unwrap_err();
        assert_eq!(err.details.unwrap()["reason"], "uploadReady");
        assert!(
            t0.elapsed() < Duration::from_millis(2_800),
            "{:?}",
            t0.elapsed()
        );
        assert_eq!(
            params_of(&log, "Page.setInterceptFileChooserDialog")
                .last()
                .unwrap(),
            &json!({ "enabled": false })
        );
        session.close().await;
    }

    #[test]
    fn remove_attachments_expression_deletes_image_and_video_cards_in_a_fake_dom() {
        // 用 node 跑表达式:假 DOM 里各有 1 张图片卡片与视频卡片,点删除按钮即移除卡片
        let script = format!(
            r#"
const cards = [{{ kind: 'image', alive: true }}, {{ kind: 'video', alive: true }}];
const sel = k => `[data-testid="attachment-${{k}}-card"]`;
const del = sel => `${{sel}} [data-testid="attachment-delete-btn"]`;
globalThis.MouseEvent = class {{ constructor(t) {{ this.type = t; }} }};
globalThis.document = {{
  querySelectorAll(q) {{
    const parts = q.split(',');
    const out = [];
    for (const c of cards) {{
      if (!c.alive) continue;
      if (parts.includes(sel(c.kind))) out.push(c);
      if (parts.includes(del(sel(c.kind)))) out.push({{ dispatchEvent(e) {{ if (e.type === 'click') c.alive = false; }} }});
    }}
    return out;
  }}
}};
Promise.resolve({expr}).then(r => console.log(JSON.stringify(r)));
"#,
            expr = remove_attachments_expression()
        );
        let out = match std::process::Command::new("node")
            .arg("-e")
            .arg(&script)
            .output()
        {
            Ok(o) => o,
            Err(_) => return, // 没有 node 的环境跳过
        };
        assert!(
            out.status.success(),
            "{}",
            String::from_utf8_lossy(&out.stderr)
        );
        let v: Value = serde_json::from_slice(&out.stdout).unwrap();
        assert_eq!(v, json!({ "before": 2, "after": 0 }));
    }

    #[tokio::test]
    async fn dom_protocol_error_is_preparing_failed() {
        let (session, _log) = spawn(Box::new(|method, params| match method {
            "Runtime.evaluate" => {
                let _ = step_of(params["expression"].as_str().unwrap());
                Reply::value(json!({}))
            }
            "DOM.getDocument" => Reply::raw(json!({ "root": { "nodeId": 1 } })),
            "DOM.querySelector" => Reply::raw(json!({ "nodeId": 9 })),
            "DOM.setFileInputFiles" => Reply::protocol_error("Node is not a file input element"),
            _ => Reply::raw(json!({})),
        }))
        .await;
        let (_p, on_progress) = progress();
        let err = upload_references_with(
            &session,
            KEY,
            &mut req(&["/tmp/a.png"]),
            &[PathBuf::from("/tmp/a.png")],
            &CancellationToken::new(),
            &on_progress,
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "generation_preparing_failed");
        let d = err.details.unwrap();
        assert_eq!(d["reason"], "cdp_command_failed");
        assert_eq!(d["method"], "DOM.setFileInputFiles");
        session.close().await;
    }
}
