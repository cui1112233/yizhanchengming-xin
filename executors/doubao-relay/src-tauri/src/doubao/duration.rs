//! 视频时长设置(Task 3.2):`prepare` 返回 `{needsDurationInput:true}` 后,按 analysis/04 "Step 2"
//! 的混合策略把"视频生成"技能弹层里的时长滑块调到目标秒数,最后 `verifyVideoSettings` 核对。
//!
//! 循环(≤ 40 轮,与沈浪 `doubao-generate.ps1` 一致):
//! 1. `durationDragPlan` → `matched:true` 结束;
//! 2. 上一轮按过方向键:`duration` 变化不等于 ±1 → 键盘无效(`keyboard_disabled`),之后只拖;
//! 3. 第 0 轮且有 `endpointKey`(目标是滑块端点)→ `Input.dispatchKeyEvent` Home / End;
//! 4. 有 `fineTuneKey` 且键盘未判无效,并且已经粗拖过或离目标 ≤ [`FINE_TUNE_MAX_GAP`] 秒 → 方向键微调一格;
//! 5. 否则拖拽:`mouseMoved` 起点 → `mousePressed` → 12 步 `mouseMoved` → `mouseReleased`;拖满 8 次停。
//!
//! 真机事实(`docs/analysis/10-live-probe-macos.md` Task 2.1 / 3.2):滑块是 `[role=slider]` span,
//! `aria-valuenow` 是刻度索引(秒 = 索引 + 4);`durationDragPlan.duration` 是**秒**(脚本内部换算过),
//! `durationState.duration` 单位不统一,这里不用它。`durationDragPlan` 每次调用都会 `focus()` 滑块,
//! 只有滑块拿到焦点时才给出 `endpointKey` / `fineTuneKey`,所以键盘事件总是落在滑块上。
//!
//! 第 4 步与沈浪不同:PS 只在"至少拖过一次"之后才用方向键;这里离目标 ≤ 2 秒时直接用方向键
//! (少一次拖拽,真机拖拽偶尔差 1 格),键盘无效时照样回退到拖拽。

use std::time::Duration;

use serde::Deserialize;
use serde_json::{json, Value};
use tokio_util::sync::CancellationToken;

use crate::cdp::{CdpError, CdpSession};
use crate::doubao::generation::{
    cancelled, evaluate_step, map_preparing, preparing_failed, preparing_timeout, step_expression,
};
use crate::error::{RelayError, RelayResult};

/// 最多调用 `durationDragPlan` 的轮数(PS:`$attempt -lt 40`)。
pub const MAX_ATTEMPTS: usize = 40;
/// 最多拖拽次数(PS:`$dragAttempts -ge 8` 停)。
pub const MAX_DRAG_ATTEMPTS: usize = 8;
/// 拖拽分几步移动(PS:12 步)。
pub const DRAG_STEPS: usize = 12;
/// 离目标不超过这么多秒时,未拖过也直接用方向键微调。
pub const FINE_TUNE_MAX_GAP: i64 = 2;

/// 各步等待时间(默认值取自 PS),测试用小值。
#[derive(Clone, Debug)]
pub struct DurationTiming {
    /// `durationDragPlan` 的求值超时(PS `Step` 默认 20s)。
    pub step_timeout: Duration,
    /// `verifyVideoSettings` 的求值超时(PS 10s)。
    pub verify_timeout: Duration,
    /// Home / End 之后等待(250ms)。
    pub after_endpoint_key: Duration,
    /// 方向键之后等待(180ms)。
    pub after_fine_tune_key: Duration,
    /// 拖拽每步之间(20ms)。
    pub drag_step: Duration,
    /// 松开鼠标之后等待(300ms)。
    pub after_drag: Duration,
}

impl Default for DurationTiming {
    fn default() -> Self {
        Self {
            step_timeout: Duration::from_secs(20),
            verify_timeout: Duration::from_secs(10),
            after_endpoint_key: Duration::from_millis(250),
            after_fine_tune_key: Duration::from_millis(180),
            drag_step: Duration::from_millis(20),
            after_drag: Duration::from_millis(300),
        }
    }
}

/// `durationDragPlan` 的返回(未匹配时带拖拽坐标与建议按键)。
#[derive(Clone, Debug, Default, Deserialize, PartialEq)]
#[serde(rename_all = "camelCase")]
pub struct DragPlan {
    #[serde(default)]
    pub matched: bool,
    /// 当前滑块对应的秒数。
    #[serde(default)]
    pub duration: Option<f64>,
    #[serde(default)]
    pub start_x: Option<f64>,
    #[serde(default)]
    pub end_x: Option<f64>,
    #[serde(default)]
    pub y: Option<f64>,
    /// `"End"` / `"Home"`(目标是滑块端点且滑块有焦点)。
    #[serde(default)]
    pub endpoint_key: Option<String>,
    /// `"ArrowRight"` / `"ArrowLeft"`(滑块有焦点)。
    #[serde(default)]
    pub fine_tune_key: Option<String>,
}

/// 一轮循环要做的动作(纯函数 [`next_action`] 的结果,便于单测)。
#[derive(Clone, Debug, PartialEq)]
pub enum DurationAction {
    Done,
    Key(String),
    FineTune(String),
    Drag { start_x: f64, end_x: f64, y: f64 },
    GiveUp,
}

/// 循环状态。
#[derive(Clone, Debug, Default)]
pub struct DurationLoop {
    pub drag_attempts: usize,
    pub keyboard_disabled: bool,
    /// 上一轮按方向键前的秒数与按键(用于检测键盘是否生效)。
    pub keyboard_before: Option<(i64, String)>,
}

impl DurationLoop {
    /// 根据本轮 plan 决定动作,并更新键盘检测状态(PS 的分支顺序)。
    pub fn next_action(
        &mut self,
        attempt: usize,
        target_secs: u32,
        plan: &DragPlan,
    ) -> DurationAction {
        if plan.matched {
            return DurationAction::Done;
        }
        let current = plan.duration.map(|d| d.round() as i64);
        if let Some((before, key)) = self.keyboard_before.take() {
            let expected = if key == "ArrowRight" { 1 } else { -1 };
            if current.map(|c| c - before) != Some(expected) {
                self.keyboard_disabled = true;
                tracing::warn!(key = %key, before, after = ?current, "方向键未改变视频时长,之后只用拖拽");
            }
        }
        if attempt == 0 {
            if let Some(key) = plan.endpoint_key.as_deref().filter(|k| is_endpoint_key(k)) {
                return DurationAction::Key(key.to_string());
            }
        }
        if let (false, Some(key), Some(cur)) = (
            self.keyboard_disabled,
            plan.fine_tune_key.as_deref().filter(|k| is_arrow_key(k)),
            current,
        ) {
            let gap = (i64::from(target_secs) - cur).abs();
            if self.drag_attempts > 0 || gap <= FINE_TUNE_MAX_GAP {
                self.keyboard_before = Some((cur, key.to_string()));
                return DurationAction::FineTune(key.to_string());
            }
        }
        if self.drag_attempts >= MAX_DRAG_ATTEMPTS {
            return DurationAction::GiveUp;
        }
        match (plan.start_x, plan.end_x, plan.y) {
            (Some(start_x), Some(end_x), Some(y)) => {
                self.drag_attempts += 1;
                DurationAction::Drag { start_x, end_x, y }
            }
            _ => DurationAction::GiveUp,
        }
    }
}

fn is_endpoint_key(k: &str) -> bool {
    matches!(k, "Home" | "End")
}

fn is_arrow_key(k: &str) -> bool {
    matches!(k, "ArrowLeft" | "ArrowRight")
}

/// Windows 虚拟键码(CDP `windowsVirtualKeyCode` / `nativeVirtualKeyCode`,PS 同值)。
pub fn virtual_key_code(key: &str) -> i64 {
    match key {
        "End" => 35,
        "Home" => 36,
        "ArrowLeft" => 37,
        "ArrowRight" => 39,
        _ => 0,
    }
}

/// `generation_preparing_failed`:时长没调到目标。
pub fn duration_not_applied(target_secs: u32, actual: Option<f64>) -> RelayError {
    let actual_text = actual
        .map(|a| format!("{}", a.round() as i64))
        .unwrap_or_else(|| "未知".into());
    RelayError::new(
        "generation_preparing_failed",
        format!("视频时长未生效:期望{target_secs}秒,实际{actual_text}秒,请检查豆包时长滑块后重试"),
    )
    .with_details(json!({
        "reason": "duration_not_applied",
        "js_exception": Value::Null,
        "expected": target_secs,
        "actual": actual,
    }))
}

/// 把视频时长设到 `target_secs`(默认时间参数)。见 [`set_duration_with`]。
pub async fn set_duration(
    session: &CdpSession,
    tracking_key: &str,
    request_json: &Value,
    target_secs: u32,
    cancel: &CancellationToken,
) -> RelayResult<()> {
    set_duration_with(
        session,
        tracking_key,
        request_json,
        target_secs,
        cancel,
        &DurationTiming::default(),
    )
    .await
}

/// 见模块文档。结束时 `durationDragPlan` 仍未匹配 → `generation_preparing_failed`
/// (`reason:"duration_not_applied"`);`verifyVideoSettings` 抛异常 → `generation_preparing_failed`
/// (`reason:"verifyVideoSettings"`),返回的时长与目标不符 → `reason:"duration_mismatch"`。
pub async fn set_duration_with(
    session: &CdpSession,
    tracking_key: &str,
    request_json: &Value,
    target_secs: u32,
    cancel: &CancellationToken,
    timing: &DurationTiming,
) -> RelayResult<()> {
    let plan_expr = step_expression(tracking_key, "durationDragPlan", request_json);
    let mut state = DurationLoop::default();
    let started = std::time::Instant::now();
    let mut matched = false;
    let mut keys = 0usize;
    for attempt in 0..MAX_ATTEMPTS {
        let plan = drag_plan(session, &plan_expr, timing, cancel).await?;
        let action = state.next_action(attempt, target_secs, &plan);
        tracing::debug!(attempt, current = ?plan.duration, action = ?action, "视频时长调整");
        match action {
            DurationAction::Done => {
                matched = true;
                break;
            }
            DurationAction::Key(key) => {
                keys += 1;
                press_key(session, &key, timing).await?;
                sleep_or_cancel(timing.after_endpoint_key, cancel, "duration").await?;
            }
            DurationAction::FineTune(key) => {
                keys += 1;
                press_key(session, &key, timing).await?;
                sleep_or_cancel(timing.after_fine_tune_key, cancel, "duration").await?;
            }
            DurationAction::Drag { start_x, end_x, y } => {
                drag(session, start_x, end_x, y, timing, cancel).await?;
                sleep_or_cancel(timing.after_drag, cancel, "duration").await?;
            }
            DurationAction::GiveUp => break,
        }
    }
    if !matched {
        // 循环用尽或放弃:再读一次(最后一个动作之后的状态),仍不匹配才报错(PS 同)。
        let last = drag_plan(session, &plan_expr, timing, cancel).await?;
        if !last.matched {
            tracing::warn!(
                target_secs,
                actual = ?last.duration,
                drag_attempts = state.drag_attempts,
                keys,
                keyboard_disabled = state.keyboard_disabled,
                "视频时长未调到目标"
            );
            return Err(duration_not_applied(target_secs, last.duration));
        }
    }

    let verified = evaluate_step(
        session,
        &step_expression(tracking_key, "verifyVideoSettings", request_json),
        timing.verify_timeout,
        cancel,
    )
    .await
    .map_err(|e| map_preparing("verifyVideoSettings", e))?;
    if let Some(got) = verified.get("duration").filter(|d| !d.is_null()) {
        if got.as_f64().map(|g| g.round() as i64) != Some(i64::from(target_secs)) {
            return Err(
                preparing_failed("duration_mismatch", None).with_details(json!({
                    "reason": "duration_mismatch",
                    "js_exception": Value::Null,
                    "expected": target_secs,
                    "actual": got,
                    "verify": verified,
                })),
            );
        }
    }
    tracing::info!(
        target_secs,
        drag_attempts = state.drag_attempts,
        keys,
        keyboard_disabled = state.keyboard_disabled,
        elapsed_ms = started.elapsed().as_millis() as u64,
        "视频时长已设置"
    );
    Ok(())
}

async fn drag_plan(
    session: &CdpSession,
    expr: &str,
    timing: &DurationTiming,
    cancel: &CancellationToken,
) -> RelayResult<DragPlan> {
    let v = evaluate_step(session, expr, timing.step_timeout, cancel)
        .await
        .map_err(|e| map_preparing("durationDragPlan", e))?;
    Ok(DragPlan::from_value(&v))
}

impl DragPlan {
    /// 逐字段宽松解析:某个字段类型不对只丢这个字段,不连带丢掉 `matched:true`。
    /// 数字字段接受数字或数字字符串,`matched` 接受布尔或 `"true"`。
    pub fn from_value(v: &Value) -> Self {
        let num = |k: &str| -> Option<f64> {
            match v.get(k)? {
                Value::Number(n) => n.as_f64(),
                Value::String(s) => s.trim().parse::<f64>().ok().filter(|f| f.is_finite()),
                _ => None,
            }
        };
        let text = |k: &str| -> Option<String> {
            v.get(k)
                .and_then(Value::as_str)
                .filter(|s| !s.is_empty())
                .map(str::to_string)
        };
        let matched = match v.get("matched") {
            Some(Value::Bool(b)) => *b,
            Some(Value::String(s)) => s.eq_ignore_ascii_case("true"),
            _ => false,
        };
        Self {
            matched,
            duration: num("duration"),
            start_x: num("startX"),
            end_x: num("endX"),
            y: num("y"),
            endpoint_key: text("endpointKey"),
            fine_tune_key: text("fineTuneKey"),
        }
    }
}

async fn sleep_or_cancel(d: Duration, cancel: &CancellationToken, step: &str) -> RelayResult<()> {
    tokio::select! {
        _ = cancel.cancelled() => Err(cancelled(step)),
        _ = tokio::time::sleep(d) => Ok(()),
    }
}

/// `Input.*` 命令:Protocol 错误 → `generation_preparing_failed{reason:"input_dispatch_failed"}`,
/// 超时 → `generation_preparing_timeout{step:"duration"}`,断线照常映射。
async fn input(
    session: &CdpSession,
    method: &str,
    params: Value,
    timing: &DurationTiming,
) -> RelayResult<()> {
    match session
        .send_timeout(method, params, timing.step_timeout)
        .await
    {
        Ok(_) => Ok(()),
        Err(CdpError::Protocol { code, message }) => {
            Err(
                preparing_failed("input_dispatch_failed", None).with_details(json!({
                    "reason": "input_dispatch_failed",
                    "js_exception": Value::Null,
                    "method": method,
                    "code": code,
                    "message": message,
                })),
            )
        }
        Err(CdpError::Timeout(_)) => Err(preparing_timeout("duration")),
        Err(e) => Err(e.into()),
    }
}

/// keyDown + keyUp(PS 同参数)。
async fn press_key(session: &CdpSession, key: &str, timing: &DurationTiming) -> RelayResult<()> {
    let code = virtual_key_code(key);
    for ty in ["keyDown", "keyUp"] {
        input(
            session,
            "Input.dispatchKeyEvent",
            json!({
                "type": ty,
                "key": key,
                "code": key,
                "windowsVirtualKeyCode": code,
                "nativeVirtualKeyCode": code,
            }),
            timing,
        )
        .await?;
    }
    Ok(())
}

/// 鼠标拖拽:起点 → 按下 → 12 步移动 → 松开。中途取消或出错也会尽力松开鼠标。
async fn drag(
    session: &CdpSession,
    start_x: f64,
    end_x: f64,
    y: f64,
    timing: &DurationTiming,
    cancel: &CancellationToken,
) -> RelayResult<()> {
    input(
        session,
        "Input.dispatchMouseEvent",
        json!({ "type": "mouseMoved", "x": start_x, "y": y }),
        timing,
    )
    .await?;
    input(
        session,
        "Input.dispatchMouseEvent",
        json!({ "type": "mousePressed", "x": start_x, "y": y, "button": "left", "buttons": 1, "clickCount": 1 }),
        timing,
    )
    .await?;
    let mut outcome: RelayResult<()> = Ok(());
    for step in 1..=DRAG_STEPS {
        let x = start_x + (end_x - start_x) * step as f64 / DRAG_STEPS as f64;
        if let Err(e) = input(
            session,
            "Input.dispatchMouseEvent",
            json!({ "type": "mouseMoved", "x": x, "y": y, "button": "left", "buttons": 1 }),
            timing,
        )
        .await
        {
            outcome = Err(e);
            break;
        }
        if let Err(e) = sleep_or_cancel(timing.drag_step, cancel, "duration").await {
            outcome = Err(e);
            break;
        }
    }
    let released = input(
        session,
        "Input.dispatchMouseEvent",
        json!({ "type": "mouseReleased", "x": end_x, "y": y, "button": "left", "buttons": 0, "clickCount": 1 }),
        timing,
    )
    .await;
    outcome.and(released)
}

/// 测试用 mock 豆包页面(duration / upload / generation 共用):按 `(method, params)` 交给 handler,
/// 记录全部命令;`Runtime.evaluate` 记为 `eval:<step>`(`<key>.run("<step>"` 中的 step)。
#[cfg(test)]
pub(crate) mod mock_page {
    use std::sync::{Arc, Mutex};

    use futures_util::{SinkExt, StreamExt};
    use serde_json::{json, Value};
    use tokio_tungstenite::tungstenite::Message;

    use crate::cdp::CdpSession;

    /// 对一条命令的回应:先推 `events`,再回 `result` 或 `error`。
    #[derive(Default)]
    pub(crate) struct Reply {
        pub result: Value,
        pub error: Option<String>,
        pub events: Vec<Value>,
        /// 不回应(模拟页面卡住,调用方只能靠超时)。
        pub hang: bool,
    }

    impl Reply {
        /// 普通命令的 `result`。
        pub fn raw(result: Value) -> Self {
            Self {
                result,
                ..Default::default()
            }
        }
        /// `Runtime.evaluate` 返回值。
        pub fn value(v: Value) -> Self {
            Self::raw(json!({ "result": { "type": "object", "value": v } }))
        }
        /// `Runtime.evaluate` 抛 JS 异常。
        pub fn throw(msg: &str) -> Self {
            Self::raw(json!({ "exceptionDetails": {
                "text": "Uncaught",
                "exception": { "description": format!("Error: {msg}\n    at <anonymous>:1:1") }
            }}))
        }
        /// CDP 协议错误。
        pub fn protocol_error(msg: &str) -> Self {
            Self {
                error: Some(msg.to_string()),
                ..Default::default()
            }
        }
        /// 永不回应。
        pub fn hang() -> Self {
            Self {
                hang: true,
                ..Default::default()
            }
        }
        pub fn with_events(mut self, events: Vec<Value>) -> Self {
            self.events = events;
            self
        }
    }

    pub(crate) type Handler = Box<dyn FnMut(&str, &Value) -> Reply + Send>;
    pub(crate) type Log = Arc<Mutex<Vec<(String, Value)>>>;

    /// `Promise.resolve(<key>.run("<step>", …))` 中的 step;其它表达式给 `other`。
    pub(crate) fn step_of(expr: &str) -> String {
        expr.split(".run(\"")
            .nth(1)
            .and_then(|rest| rest.split('"').next())
            .map(str::to_string)
            .unwrap_or_else(|| "other".into())
    }

    /// 日志里的短名:`eval:<step>` 或 CDP 方法名。
    pub(crate) fn names(log: &Log) -> Vec<String> {
        log.lock().unwrap().iter().map(|(n, _)| n.clone()).collect()
    }

    pub(crate) async fn spawn(mut handler: Handler) -> (Arc<CdpSession>, Log) {
        let log: Log = Arc::new(Mutex::new(Vec::new()));
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        let log2 = Arc::clone(&log);
        tokio::spawn(async move {
            let (sock, _) = listener.accept().await.unwrap();
            let mut ws = tokio_tungstenite::accept_async(sock).await.unwrap();
            while let Some(Ok(Message::Text(txt))) = ws.next().await {
                let req: Value = serde_json::from_str(&txt).unwrap();
                let method = req["method"].as_str().unwrap_or("").to_string();
                let params = req["params"].clone();
                let name = if method == "Runtime.evaluate" {
                    format!(
                        "eval:{}",
                        step_of(params["expression"].as_str().unwrap_or(""))
                    )
                } else {
                    method.clone()
                };
                log2.lock().unwrap().push((name, params.clone()));
                let reply = handler(&method, &params);
                if reply.hang {
                    continue;
                }
                for ev in reply.events {
                    if ws.send(Message::Text(ev.to_string().into())).await.is_err() {
                        return;
                    }
                }
                let resp = match reply.error {
                    Some(msg) => {
                        json!({ "id": req["id"], "error": { "code": -32000, "message": msg } })
                    }
                    None => json!({ "id": req["id"], "result": reply.result }),
                };
                if ws
                    .send(Message::Text(resp.to_string().into()))
                    .await
                    .is_err()
                {
                    break;
                }
            }
        });
        let session = CdpSession::connect(&format!("ws://127.0.0.1:{port}/devtools/page/M"), "M")
            .await
            .unwrap();
        (session, log)
    }
}

#[cfg(test)]
mod tests {
    use super::mock_page::{names, spawn, step_of, Reply};
    use super::*;
    use std::collections::VecDeque;
    use std::sync::{Arc, Mutex};

    const KEY: &str = "shenlangDoubaoTask_01test";

    fn fast() -> DurationTiming {
        DurationTiming {
            step_timeout: Duration::from_millis(500),
            verify_timeout: Duration::from_millis(500),
            after_endpoint_key: Duration::from_millis(1),
            after_fine_tune_key: Duration::from_millis(1),
            drag_step: Duration::from_millis(1),
            after_drag: Duration::from_millis(1),
        }
    }

    fn req() -> Value {
        json!({ "type": "video", "modelName": "Seedance 2.5", "duration": 10, "ratio": "16:9" })
    }

    fn plan(v: Value) -> DragPlan {
        serde_json::from_value(v).unwrap()
    }

    /// handler:`durationDragPlan` 依次回 `plans`(用完后重复最后一个),`verifyVideoSettings` 回 `verify`。
    fn scripted(plans: Vec<Value>, verify: Reply) -> super::mock_page::Handler {
        let plans = Arc::new(Mutex::new(VecDeque::from(plans)));
        let mut verify = Some(verify);
        Box::new(move |method, params| {
            if method != "Runtime.evaluate" {
                return Reply::raw(json!({}));
            }
            match step_of(params["expression"].as_str().unwrap()).as_str() {
                "durationDragPlan" => {
                    let mut q = plans.lock().unwrap();
                    let v = if q.len() > 1 {
                        q.pop_front().unwrap()
                    } else {
                        q.front().cloned().unwrap()
                    };
                    Reply::value(v)
                }
                "verifyVideoSettings" => verify.take().unwrap_or_else(|| Reply::value(Value::Null)),
                _ => Reply::value(Value::Null),
            }
        })
    }

    #[test]
    fn drag_plan_parses_script_shape() {
        let p = plan(json!({
            "startX": 744.7, "endX": 769.0, "y": 767.4, "matched": false,
            "duration": 10, "endpointKey": null, "fineTuneKey": "ArrowRight"
        }));
        assert!(!p.matched);
        assert_eq!(p.duration, Some(10.0));
        assert_eq!(p.start_x, Some(744.7));
        assert_eq!(p.end_x, Some(769.0));
        assert_eq!(p.y, Some(767.4));
        assert_eq!(p.endpoint_key, None);
        assert_eq!(p.fine_tune_key.as_deref(), Some("ArrowRight"));
        assert!(plan(json!({ "matched": true, "duration": 10 })).matched);
        // 生产路径用宽松解析,结果与 serde 一致
        let shape = json!({
            "startX": 744.7, "endX": 769.0, "y": 767.4, "matched": false,
            "duration": 10, "endpointKey": null, "fineTuneKey": "ArrowRight"
        });
        assert_eq!(DragPlan::from_value(&shape), plan(shape));
    }

    #[test]
    fn drag_plan_bad_field_does_not_hide_matched() {
        // duration 是字符串、startX 是对象:serde 会整体失败,宽松解析只丢坏字段
        let v = json!({ "matched": true, "duration": "7", "startX": {}, "y": "abc" });
        assert!(serde_json::from_value::<DragPlan>(v.clone()).is_err());
        let p = DragPlan::from_value(&v);
        assert!(p.matched);
        assert_eq!(p.duration, Some(7.0));
        assert_eq!(p.start_x, None);
        assert_eq!(p.y, None);
        assert_eq!(DragPlan::from_value(&Value::Null), DragPlan::default());
        assert!(DragPlan::from_value(&json!({ "matched": "true" })).matched);
    }

    #[test]
    fn next_action_follows_ps_branch_order() {
        let far = json!({ "matched": false, "duration": 10, "startX": 1.0, "endX": 2.0, "y": 3.0,
                          "fineTuneKey": "ArrowRight", "endpointKey": "End" });
        // 第 0 轮有 endpointKey → 先按端点键
        let mut s = DurationLoop::default();
        assert_eq!(
            s.next_action(0, 30, &plan(far.clone())),
            DurationAction::Key("End".into())
        );
        // 第 1 轮,离目标 20 秒、未拖过 → 拖
        assert_eq!(
            s.next_action(1, 30, &plan(far.clone())),
            DurationAction::Drag {
                start_x: 1.0,
                end_x: 2.0,
                y: 3.0
            }
        );
        // 拖过之后有 fineTuneKey → 方向键
        assert_eq!(
            s.next_action(2, 30, &plan(far.clone())),
            DurationAction::FineTune("ArrowRight".into())
        );
        // 方向键后时长 +1 → 键盘有效,继续微调
        let after = json!({ "matched": false, "duration": 11, "startX": 1.0, "endX": 2.0, "y": 3.0,
                            "fineTuneKey": "ArrowRight" });
        assert_eq!(
            s.next_action(3, 30, &plan(after.clone())),
            DurationAction::FineTune("ArrowRight".into())
        );
        assert!(!s.keyboard_disabled);
        // 方向键后时长没变 → 键盘无效,改拖
        assert_eq!(
            s.next_action(4, 30, &plan(after.clone())),
            DurationAction::Drag {
                start_x: 1.0,
                end_x: 2.0,
                y: 3.0
            }
        );
        assert!(s.keyboard_disabled);
        assert_eq!(s.drag_attempts, 2);
        // 匹配 → Done
        assert_eq!(
            s.next_action(5, 30, &plan(json!({ "matched": true }))),
            DurationAction::Done
        );
    }

    #[test]
    fn next_action_fine_tunes_small_gap_without_drag_and_gives_up_after_eight_drags() {
        let near = json!({ "matched": false, "duration": 9, "startX": 1.0, "endX": 2.0, "y": 3.0,
                           "fineTuneKey": "ArrowRight" });
        let mut s = DurationLoop::default();
        assert_eq!(
            s.next_action(0, 10, &plan(near)),
            DurationAction::FineTune("ArrowRight".into())
        );
        // 未聚焦(无按键建议)且拖满 8 次 → 放弃
        let nokeys =
            json!({ "matched": false, "duration": 4, "startX": 1.0, "endX": 2.0, "y": 3.0 });
        let mut s = DurationLoop::default();
        for i in 0..MAX_DRAG_ATTEMPTS {
            assert!(matches!(
                s.next_action(i, 10, &plan(nokeys.clone())),
                DurationAction::Drag { .. }
            ));
        }
        assert_eq!(s.next_action(8, 10, &plan(nokeys)), DurationAction::GiveUp);
        // 没有坐标也放弃
        let mut s = DurationLoop::default();
        assert_eq!(
            s.next_action(1, 10, &plan(json!({ "matched": false, "duration": 4 }))),
            DurationAction::GiveUp
        );
        // 非法按键名不发送
        let mut s = DurationLoop::default();
        assert_eq!(
            s.next_action(0, 10, &plan(json!({ "matched": false, "duration": 9, "endpointKey": "Delete", "fineTuneKey": "Enter" }))),
            DurationAction::GiveUp
        );
    }

    #[test]
    fn virtual_key_codes_match_ps() {
        assert_eq!(virtual_key_code("End"), 35);
        assert_eq!(virtual_key_code("Home"), 36);
        assert_eq!(virtual_key_code("ArrowLeft"), 37);
        assert_eq!(virtual_key_code("ArrowRight"), 39);
    }

    /// 计划 Step 1 的场景:End → ArrowRight → matched,最后 verifyVideoSettings。
    #[tokio::test]
    async fn end_then_arrow_right_then_matched_then_verify() {
        let (session, log) = spawn(scripted(
            vec![
                json!({ "matched": false, "endpointKey": "End", "duration": 4 }),
                json!({ "matched": false, "fineTuneKey": "ArrowRight", "duration": 9,
                        "startX": 700.0, "endX": 760.0, "y": 767.0 }),
                json!({ "matched": true, "duration": 10 }),
            ],
            Reply::value(json!({ "mode": "video", "model": "Seedance 2.5", "ratio": "16:9", "duration": 10 })),
        ))
        .await;
        set_duration_with(
            &session,
            KEY,
            &req(),
            10,
            &CancellationToken::new(),
            &fast(),
        )
        .await
        .unwrap();
        let log = log.lock().unwrap().clone();
        let keys: Vec<(String, String)> = log
            .iter()
            .filter(|(n, _)| n == "Input.dispatchKeyEvent")
            .map(|(_, p)| {
                (
                    p["type"].as_str().unwrap().into(),
                    p["key"].as_str().unwrap().into(),
                )
            })
            .collect();
        assert_eq!(
            keys,
            [
                ("keyDown".to_string(), "End".to_string()),
                ("keyUp".into(), "End".into()),
                ("keyDown".into(), "ArrowRight".into()),
                ("keyUp".into(), "ArrowRight".into()),
            ]
        );
        let end = log
            .iter()
            .find(|(n, p)| n == "Input.dispatchKeyEvent" && p["key"] == "End")
            .unwrap();
        assert_eq!(end.1["windowsVirtualKeyCode"], 35);
        assert_eq!(end.1["code"], "End");
        let right = log
            .iter()
            .find(|(n, p)| n == "Input.dispatchKeyEvent" && p["key"] == "ArrowRight")
            .unwrap();
        assert_eq!(right.1["nativeVirtualKeyCode"], 39);
        // 没有拖拽;最后一步是 verifyVideoSettings,且请求 JSON 原样传入
        assert!(!log.iter().any(|(n, _)| n == "Input.dispatchMouseEvent"));
        let last = log.last().unwrap();
        assert_eq!(last.0, "eval:verifyVideoSettings");
        assert!(last.1["expression"]
            .as_str()
            .unwrap()
            .contains("\"duration\":10"));
        assert!(last.1["expression"]
            .as_str()
            .unwrap()
            .starts_with(&format!("Promise.resolve({KEY}.run(")));
        session.close().await;
    }

    #[tokio::test]
    async fn drag_sends_twelve_moves_between_press_and_release() {
        let (session, log) = spawn(scripted(
            vec![
                json!({ "matched": false, "duration": 10, "startX": 700.0, "endX": 820.0, "y": 767.0 }),
                json!({ "matched": true, "duration": 20 }),
            ],
            Reply::value(json!({ "mode": "video", "duration": 20 })),
        ))
        .await;
        set_duration_with(
            &session,
            KEY,
            &req(),
            20,
            &CancellationToken::new(),
            &fast(),
        )
        .await
        .unwrap();
        let mouse: Vec<Value> = log
            .lock()
            .unwrap()
            .iter()
            .filter(|(n, _)| n == "Input.dispatchMouseEvent")
            .map(|(_, p)| p.clone())
            .collect();
        assert_eq!(mouse.len(), 1 + 1 + DRAG_STEPS + 1);
        assert_eq!(mouse[0]["type"], "mouseMoved");
        assert_eq!(mouse[0]["x"], 700.0);
        assert_eq!(mouse[1]["type"], "mousePressed");
        assert_eq!(mouse[1]["buttons"], 1);
        for m in &mouse[2..2 + DRAG_STEPS] {
            assert_eq!(m["type"], "mouseMoved");
            assert_eq!(m["buttons"], 1);
        }
        assert_eq!(mouse[2]["x"], 710.0);
        assert_eq!(mouse[1 + DRAG_STEPS]["x"], 820.0);
        let rel = mouse.last().unwrap();
        assert_eq!(rel["type"], "mouseReleased");
        assert_eq!(rel["x"], 820.0);
        assert_eq!(rel["buttons"], 0);
        session.close().await;
    }

    #[tokio::test]
    async fn not_matched_after_eight_drags_is_preparing_failed_with_message() {
        let (session, log) = spawn(scripted(
            vec![json!({ "matched": false, "duration": 7, "startX": 1.0, "endX": 2.0, "y": 3.0 })],
            Reply::value(Value::Null),
        ))
        .await;
        let err = set_duration_with(
            &session,
            KEY,
            &req(),
            12,
            &CancellationToken::new(),
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "generation_preparing_failed");
        assert!(err.message.contains("期望12秒"), "{}", err.message);
        assert!(err.message.contains("实际7秒"), "{}", err.message);
        let d = err.details.unwrap();
        assert_eq!(d["reason"], "duration_not_applied");
        assert_eq!(d["expected"], 12);
        let names = names(&log);
        let presses = names
            .iter()
            .filter(|n| *n == "Input.dispatchMouseEvent")
            .count();
        assert_eq!(presses, MAX_DRAG_ATTEMPTS * (DRAG_STEPS + 3));
        assert!(!names.iter().any(|n| n == "eval:verifyVideoSettings"));
        session.close().await;
    }

    #[tokio::test]
    async fn verify_exception_or_mismatch_is_preparing_failed() {
        let (session, _log) = spawn(scripted(
            vec![json!({ "matched": true, "duration": 10 })],
            Reply::throw("视频参数核对失败，任务未发送"),
        ))
        .await;
        let err = set_duration_with(
            &session,
            KEY,
            &req(),
            10,
            &CancellationToken::new(),
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "generation_preparing_failed");
        let d = err.details.unwrap();
        assert_eq!(d["reason"], "verifyVideoSettings");
        assert_eq!(d["js_exception"], "视频参数核对失败，任务未发送");
        session.close().await;

        let (session, _log) = spawn(scripted(
            vec![json!({ "matched": true, "duration": 10 })],
            Reply::value(json!({ "mode": "video", "duration": 9 })),
        ))
        .await;
        let err = set_duration_with(
            &session,
            KEY,
            &req(),
            10,
            &CancellationToken::new(),
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.details.unwrap()["reason"], "duration_mismatch");
        session.close().await;
    }

    #[tokio::test]
    async fn drag_plan_exception_and_cancel() {
        let (session, _log) = spawn(Box::new(|method, _| {
            if method == "Runtime.evaluate" {
                Reply::throw("未识别视频时长滑块刻度")
            } else {
                Reply::raw(json!({}))
            }
        }))
        .await;
        let err = set_duration_with(
            &session,
            KEY,
            &req(),
            10,
            &CancellationToken::new(),
            &fast(),
        )
        .await
        .unwrap_err();
        assert_eq!(err.code, "generation_preparing_failed");
        assert_eq!(err.details.unwrap()["reason"], "durationDragPlan");
        session.close().await;

        // 拖拽中取消:立刻返回 cancelled,且仍然松开鼠标
        let (session, log) = spawn(scripted(
            vec![json!({ "matched": false, "duration": 4, "startX": 1.0, "endX": 2.0, "y": 3.0 })],
            Reply::value(Value::Null),
        ))
        .await;
        let cancel = CancellationToken::new();
        let timing = DurationTiming {
            drag_step: Duration::from_millis(40),
            ..fast()
        };
        let c2 = cancel.clone();
        tokio::spawn(async move {
            tokio::time::sleep(Duration::from_millis(60)).await;
            c2.cancel();
        });
        let err = set_duration_with(&session, KEY, &req(), 10, &cancel, &timing)
            .await
            .unwrap_err();
        assert_eq!(err.code, "generation_cancelled");
        let log = log.lock().unwrap().clone();
        let last_mouse = log
            .iter()
            .rev()
            .find(|(n, _)| n == "Input.dispatchMouseEvent")
            .unwrap();
        assert_eq!(last_mouse.1["type"], "mouseReleased");
        session.close().await;
    }

    #[tokio::test]
    async fn input_protocol_error_is_preparing_failed() {
        let (session, _log) = spawn(Box::new(move |method, _| match method {
            "Runtime.evaluate" => {
                Reply::value(json!({ "matched": false, "endpointKey": "Home", "duration": 10 }))
            }
            "Input.dispatchKeyEvent" => Reply::protocol_error("Input is not available"),
            _ => Reply::raw(json!({})),
        }))
        .await;
        let err = set_duration_with(&session, KEY, &req(), 4, &CancellationToken::new(), &fast())
            .await
            .unwrap_err();
        assert_eq!(err.code, "generation_preparing_failed");
        let d = err.details.unwrap();
        assert_eq!(d["reason"], "input_dispatch_failed");
        assert_eq!(d["method"], "Input.dispatchKeyEvent");
        session.close().await;
    }
}
