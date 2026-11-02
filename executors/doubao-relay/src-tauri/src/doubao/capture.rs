//! 捕获管线(Task 2.4 图像路径 + Task 3.3 视频三条无水印路径):`send` 之后到拿到无水印原图 /
//! 原片 URL 之间的"等待 + 识别"循环。
//!
//! 视频(analysis/05 "视频三条无水印获取路径",`watermark_source` 记第一条结果的来源):
//! - **路径 1**:`responseReceived` `video/*` + `send_since` 之后的请求 + 视频 CDN 白名单 → `VideoCandidates`;
//!   含 `lr=unwatermarked` → 原片(`lr_unwatermarked`)。reader `read` 自己经
//!   `/creativity/resource/get_without_watermark` 换到的 `images[{url, unwatermarked:true}]` 同样记 `lr_unwatermarked`。
//! - **路径 2**:feedback stage 为 `result` 且还没有原片 → 每 5s `run('requestVideoDownload')` 点"下载"按钮,
//!   直到 `clicked:true`;豆包随后的 `lr=unwatermarked` 请求由路径 1 接住。
//! - **路径 3**:`read` 的 `fallbacks[{vid, api}]`(已绑定才采信)→ 每个 vid 10s 内不重复 → api 必须
//!   `https://vas-[a-z0-9-]+.snssdk.com` 且无 userinfo → [`FallbackFetcher`] GET(10s、≤8MB、不跟随重定向)
//!   → `video_decoder::decode` → [`fallback_original_url`] → 原片(`fallback_decrypted`);Rust 解不出时注入
//!   原版 `shenlangDoubaoVideoDecoder.decode()` 兜底(warn)。
//!
//! 复刻 `doubao-capture.ps1`(analysis/05)的大循环,`tokio::select!` 驱动:
//! - **Network 事件**(chat + background session):`requestWillBeSent` 分类为历史查询
//!   (`www.doubao.com` `/im/chain/single` | `/im/message/chain_by_block` POST,过滤 initiator 含
//!   `shenlangReadBoundConversation` 的自家重放)或生成请求(`generation::is_generation_request`);
//!   `loadingFinished` → `Network.getResponseBody`;`eventSourceMessageReceived` 缓冲(≤64 条 / 1MB,
//!   供 `streamSubmissionBinding` 绑定新对话);视频任务把 `send_since` 之后的**所有** requestId 记进
//!   `video_requests`(PS `videoRequests` 语义),`responseReceived` 按它匹配 `video/*` 直链 → `VideoCandidates`。
//! - **feedback 轮询**(500ms):图像 `terminalFeedback`,视频 `videoFollowup` / `videoFeedback`
//!   (`feedback::feedback_step`);checking / quota_checking 时按 `should_auto_confirm` 代回复
//!   (图像由 Rust 注入 [`accept_reply_expression`]:`run('fill',{prompt:"继续生成"})` + 点发送按钮,
//!   成功判据是返回 `sent:true`;视频由 `videoFollowup` 自行发送)或发 `NeedsUserConfirm`;
//!   **用户确认是一次性的**:代回复真正发出后 `user_confirmed` 回落,下一个确认 stage
//!   (如 `checking` 之后的 `quota_checking`)重新按 `should_auto_confirm` 判断、再发 `NeedsUserConfirm`;
//!   auto 模式代回复异常 / `Runtime.evaluate` 超时 / 连续 3 次未发出 → 降级 manual 并发 `NeedsUserConfirm`;
//!   `rendering_unverified` 连续 30s → `generation_rendering_not_executed`;拒绝类 stage →
//!   `feedback::refusal_relay_error`。
//! - **历史重放**(1s):`shenlangDoubaoResultReader.read(historyRequest, options)`,options 见
//!   analysis/05 "analyze";得到 `binding` 后发 `Bound`,`images` 只在已绑定 submission 时采信
//!   (reader 的 `collect` 在 `boundSubmissionId` 为空时会收下整个对话里的图,不能用)。
//! - **chat target 唯一性**(2s):`/json/list` 里 doubao-chat page > 1 →
//!   `generation_user_interrupted{interruption:"multiple_chat_pages"}`;= 0(页面导航走 / 豆包重启中)
//!   与 `/json/list` 失败一样计数,连续 3 次 → `doubao_cdp_disconnected{reason:"chat_page_missing"}`;
//!   chat session 断开(页面 reload)则经 `DoubaoConnection::chat_session()` 重连并重新订阅
//!   (background session 一并重取)。
//! - **输入框脏检查**(10s):输入框非空只 `warn`(spec 08 `user_modified_input`)。
//! - **deadline** → `generation_task_timeout`;`cancel` → `generation_cancelled{step:"capture"}`。
//!
//! 直接从生成请求响应 / SSE 里 `extract_image_originals` 的旁路只在该段文本含本任务的
//! `local_message_id` / `message_id` / 已绑定的 `submission_id` 时采信(并行跟踪下防止串任务)。
//!
//! 真机契约(Node 里对混淆 `doubao-result-reader.js` 实测,见 docs/analysis/10 "Task 2.4"):
//! `analyze(body, request, options)` / `read(request, options)` 返回
//! `{ok, images:[], feedback:{message,messageId,submissionId,createdAt,scope}|null,
//! binding:{conversationId,submissionId,createdAt}|null, reason?, nativeShadow}`——**feedback 是消息对象,
//! 没有 `stage`**;`validateRequest` 要求历史请求 postData 里恰好一个 `conversation_id`。
//! `run('terminalFeedback', req)` 只读 `req.type/prompt`,返回 `{stage, message?}`(其 `terminalFeedbackKind`
//! 对"会员额度,是否继续生成"只回 `unknown`);`run('videoFollowup', req)` 返回 `{sent, stage, manualRequired?}`。
//! **`run('acceptFeedback', {boundSubmissionId, reply, prompt, modelName, type})` 是历史反馈的分类钩子**
//! (PS doubao-capture.ps1:341-345):`reply` 为上述消息对象时返回 `{stage, message, messageId, …}`
//! (会员额度确认 → `quota_checking`、额度用完 → `quota_exhausted`……),字符串 `reply` 一律 `unknown`;
//! 它**不** fill、**不**点发送。本模块在已绑定 submission、结果无图时把每条反馈消息(按 messageId 去重)
//! 送它分类,结果按 feedback 处理(source `history`);真正的"继续生成"仍由 [`accept_reply_expression`] 发出。
//! feedback 依赖 `send` 时写在 trackingKey 全局对象里的提交状态,
//! 所以本模块只按 `typeof <key> === 'undefined'` 守卫注入,绝不覆盖已存在的全局对象。
//!
//! 生成请求的 `requestWillBeSent` 在 `submit` 的 identity 窗口里就已发生(capture 订阅更晚),
//! 由 `CaptureOptions.generation_requests`(= `SubmissionReport.generation_requests`)预置
//! `(target_id, request_id)`,之后的 SSE / `loadingFinished` 才能接住。拿到原图后再等
//! `result_settle`(默认 1.5s,无新 URL)才返回,上限 `result_settle_max`(6s),防止多图只交付一部分。

use std::collections::{HashMap, HashSet};
use std::sync::Arc;
use std::time::{Duration, Instant};

use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use tokio::sync::{broadcast, mpsc, watch};
use tokio::task::JoinSet;
use tokio_util::sync::CancellationToken;

use super::connection::DoubaoConnection;
use super::download::has_lr_unwatermarked;
use super::feedback::{
    feedback_step, is_confirmation_stage, is_refusal_stage, map_feedback_stage, not_executed_error,
    parse_feedback, refusal_relay_error, should_auto_confirm, FeedbackState, CONTINUE_REPLY,
    RENDERING_UNVERIFIED_LIMIT, STAGE_CHECKING, STAGE_GENERATING, STAGE_QUOTA_CHECKING,
    STAGE_RENDERING_UNVERIFIED, STAGE_RESULT,
};
use super::generation::{
    cancelled, generation_bootstrap, is_generation_request, reader_bootstrap, ObservedRequest,
    SubmissionIdentity, MAX_POST_DATA_SIZE, READER_GLOBAL,
};
use super::video_decoder::{
    self, is_video_cdn_host, unwatermarked_url, unwatermarked_video_url, video_url,
    VIDEO_DECODER_SCRIPT,
};
use crate::cdp::session::{extract_js_exception, CdpEvent, EVALUATE_TIMEOUT};
use crate::cdp::{list_targets, select_pages, CdpError, CdpSession, CdpTarget};
use crate::error::{RelayError, RelayResult};
use crate::tasks::{ConfirmationMode, TaskStage, TaskType, WatermarkSource};

/// 无水印原图允许的 CDN 域(analysis/05 `Parse-Originals`)。
pub const IMAGE_CDN_HOSTS: &[&str] = &["byteimg.com", "volces.com"];
/// 自家重放请求的 initiator 函数名(reader 内部 `shenlangReadBoundConversation`)。
pub const READER_INITIATOR_FN: &str = "shenlangReadBoundConversation";
/// SSE 缓冲上限(analysis/05)。
pub const SSE_BACKLOG_MAX_ENTRIES: usize = 64;
pub const SSE_BACKLOG_MAX_BYTES: usize = 1_048_576;
/// 响应体超过此大小不解析(analysis/05 `Parse-Body`)。
pub const MAX_BODY_BYTES: usize = 16_777_216;
/// `/json/list` 一次失败(或 0 个 chat page)不致命;连续失败超过此数才按断线处理(心跳另有崩溃判定)。
const TARGET_CHECK_MAX_FAILURES: u32 = 3;
/// 图像代回复连续多少次返回 `sent:false`(发送按钮未就绪等)后视为失败,走降级链。
pub const ACCEPT_MAX_ATTEMPTS: u32 = 3;
/// 代回复(auto)真正发出后 `Stage(Sent, …)` 的阶段消息(Task 4.1)。
pub const AUTO_CONTINUE_MESSAGE: &str = "已自动点继续";
/// 用户确认后代回复真正发出时的阶段消息。
pub const USER_CONTINUE_MESSAGE: &str = "已按用户确认继续";
/// 代回复时找发送按钮的 selector(技能态 `#flow-end-msg-send`,普通态 `chat_input_send_button`,
/// 与 `compat::SEND_BUTTON_SELECTOR` 同源)。
pub const REPLY_SEND_BUTTON_SELECTOR: &str =
    r#"#flow-end-msg-send,[data-testid="chat_input_send_button"],.send-btn-wrapper button"#;
/// 代回复表达式里的标记(测试 / 调试日志按它识别)。
pub const ACCEPT_REPLY_MARKER: &str = "doubao-relay:acceptReply";
/// relay 自己读"最后一条机器人回复"的表达式标记(测试按它归类)。
pub const RELAY_CONFIRM_MARKER: &str = "doubao-relay:relayConfirm";

static RELAY_CONFIRM_RE: std::sync::LazyLock<regex::Regex> = std::sync::LazyLock::new(|| {
    regex::Regex::new(
        r"确认参数|参数确认|确认后.{0,10}开始生成|是否继续生成|免费额度已用完.{0,80}将消耗付费额度",
    )
    .expect("常量正则")
});

/// 机器人回复是否在等用户确认后才开始生成(豆包 2.32.6 视频技能的新话术沈浪脚本认不出)。
pub fn is_relay_confirmation_text(text: &str) -> bool {
    RELAY_CONFIRM_RE.is_match(text)
}

/// 当前页是绑定对话时,返回最后一条消息(若是机器人回复)的文字;兼容旧版
/// `send_message` / `receive_message` 和新版统一的 `message_content` 节点。最后一条是用户消息
/// (含 relay 已代回复)或页面不在绑定对话 → `text: null`。
pub fn relay_confirm_expression(bound_conversation_id: Option<&str>) -> String {
    format!(
        r#"(() => {{ /* {marker} */
  const bound = {bound};
  if (bound && !location.pathname.includes(bound)) return {{ text: null, reason: 'other_conversation' }};
  const modern = [...document.querySelectorAll('[data-testid=message_content]')];
  const legacy = [...document.querySelectorAll('[data-testid=send_message],[data-testid=receive_message]')];
  const all = modern.length ? modern : legacy;
  const last = all[all.length - 1];
  if (!last) return {{ text: null }};
  const testid = last.getAttribute('data-testid');
  const sent = testid === 'send_message'
    || last.hasAttribute('data-send-message-boundary')
    || !!last.querySelector('[data-send-message-status]');
  const received = testid === 'receive_message' || (testid === 'message_content' && !sent);
  if (!received) return {{ text: null }};
  return {{ text: String(last.innerText || '').trim().slice(0, 400) }};
}})()"#,
        marker = RELAY_CONFIRM_MARKER,
        bound = serde_json::to_string(&bound_conversation_id).unwrap_or_else(|_| "null".into()),
    )
}
/// 视频路径 3:fallback API 响应上限(沈浪 `8388608`)。
pub const FALLBACK_API_MAX_BYTES: usize = 8_388_608;
/// 原版解密脚本注册的全局对象。
pub const VIDEO_DECODER_GLOBAL: &str = "shenlangDoubaoVideoDecoder";

/// 捕获参数。
#[derive(Clone, Debug)]
pub struct CaptureOptions {
    pub task_type: TaskType,
    pub prompt: String,
    /// 请求的模型名(`acceptFeedback` 分类钩子的 `modelName`);未知时 None。
    pub model_name: Option<String>,
    /// 视频比例 / 时长(秒)/ 参考素材数:`videoFollowup` 的 `confirmationMatches` 拿它们与"视频生成参数确认"
    /// 核对,auto 代回复时也拼进"以上参数已由用户核对"的补发文本;缺了会被判 `manualRequired`
    /// (UC-20 真机)。图像不用,可为 None / 0。
    pub ratio: Option<String>,
    pub duration: Option<u32>,
    pub reference_count: usize,
    pub identity: SubmissionIdentity,
    /// `submit` 已观察到的生成请求(`SubmissionReport.generation_requests`):capture 订阅晚于
    /// 这些请求的 `requestWillBeSent`,预置后才能接住它们后续的 SSE / `loadingFinished`。
    pub generation_requests: Vec<ObservedRequest>,
    /// 点发送前一刻(`SubmissionReport.send_started_at`),Unix epoch 毫秒;reader 只认之后的消息。
    pub send_since: i64,
    pub confirmation_mode: ConfirmationMode,
    pub auto_use_membership: bool,
    pub tracking_key: String,
    /// 总超时(图 5min / 视频 20min,由调用方按 spec 传入)。
    pub deadline: Duration,
}

/// 循环节奏(默认即 spec / analysis;测试用小值)。
#[derive(Clone, Copy, Debug)]
pub struct CaptureTiming {
    pub feedback_poll: Duration,
    pub history_poll: Duration,
    pub target_check: Duration,
    pub input_check: Duration,
    pub rendering_unverified_limit: Duration,
    /// 页面脚本求值超时(`Runtime.evaluate`)。
    pub evaluate_timeout: Duration,
    /// 其它 CDP 命令超时(`Network.getResponseBody` 等)。
    pub command_timeout: Duration,
    /// 拿到第一批原图后再等多久没有新 URL 才返回(多图结果可能分几次到齐)。
    pub result_settle: Duration,
    /// 从第一批原图起最多等多久(防止持续有新 URL 时无限推迟)。
    pub result_settle_max: Duration,
    /// 视频路径 2:feedback 进入 `result` 后每隔多久点一次"下载"按钮(直到 `clicked:true`)。
    pub download_click_interval: Duration,
    /// 视频路径 3:同一个 vid 多久内不重复请求 fallback API。
    pub fallback_retry: Duration,
    /// 视频路径 3:fallback API 请求超时。
    pub fallback_timeout: Duration,
}

impl Default for CaptureTiming {
    fn default() -> Self {
        Self {
            feedback_poll: Duration::from_millis(500),
            history_poll: Duration::from_secs(1),
            target_check: Duration::from_secs(2),
            input_check: Duration::from_secs(10),
            rendering_unverified_limit: RENDERING_UNVERIFIED_LIMIT,
            evaluate_timeout: EVALUATE_TIMEOUT,
            command_timeout: Duration::from_secs(15),
            result_settle: Duration::from_millis(1500),
            result_settle_max: Duration::from_secs(6),
            download_click_interval: Duration::from_secs(5),
            fallback_retry: Duration::from_secs(10),
            fallback_timeout: Duration::from_secs(10),
        }
    }
}

/// 视频路径 3 的 fallback API 请求(生产用 reqwest,测试注入假实现:真 API 只能是
/// `https://vas-*.snssdk.com`,本地 mock 起不来)。
#[async_trait]
pub trait FallbackFetcher: Send + Sync {
    /// GET `api`,返回响应文本;失败返回原因字符串(`fallback_request_failed:…`)。
    async fn fetch(&self, api: &str, timeout: Duration) -> Result<String, String>;
}

/// 生产实现:不跟随重定向(沈浪 `-MaximumRedirection 0`)、`Accept: application/json,…`、≤8MB。
pub struct HttpFallbackFetcher {
    client: reqwest::Client,
}

impl HttpFallbackFetcher {
    pub fn new() -> Self {
        Self {
            client: super::download::new_client(),
        }
    }
}

impl Default for HttpFallbackFetcher {
    fn default() -> Self {
        Self::new()
    }
}

#[async_trait]
impl FallbackFetcher for HttpFallbackFetcher {
    async fn fetch(&self, api: &str, timeout: Duration) -> Result<String, String> {
        let mut resp = self
            .client
            .get(api)
            .header(reqwest::header::ACCEPT, "application/json,text/plain,*/*")
            .timeout(timeout)
            .send()
            .await
            .map_err(|e| format!("fallback_request_failed:{e}"))?;
        if resp.status().as_u16() != 200 {
            return Err(format!("fallback_request_failed:http_{}", resp.status()));
        }
        if resp
            .content_length()
            .is_some_and(|l| l > FALLBACK_API_MAX_BYTES as u64)
        {
            return Err("fallback_response_too_large".into());
        }
        let mut body: Vec<u8> = Vec::new();
        while let Some(chunk) = resp
            .chunk()
            .await
            .map_err(|e| format!("fallback_request_failed:{e}"))?
        {
            if body.len() + chunk.len() > FALLBACK_API_MAX_BYTES {
                return Err("fallback_response_too_large".into());
            }
            body.extend_from_slice(&chunk);
        }
        Ok(String::from_utf8_lossy(&body).into_owned())
    }
}

/// 捕获过程中的事件(调用方据此更新任务状态 / 推 Tauri 事件)。
#[derive(Clone, Debug, PartialEq)]
pub enum CaptureEvent {
    Stage(TaskStage, Option<String>),
    Bound {
        conversation_id: String,
        submission_id: String,
    },
    Feedback(FeedbackState),
    NeedsUserConfirm,
    Refused {
        kind: String,
        message: String,
    },
    /// 图像无水印原图 URL(去重保序)。
    Originals(Vec<String>),
    /// 视频直链候选(Phase 3 消费)。
    VideoCandidates(Vec<String>),
    /// 视频 fallback token 入口(Phase 3 消费)。
    Fallbacks(Vec<FallbackVideo>),
    /// generating 期间豆包输入框出现了非预期内容(spec 08 `user_modified_input`):只提示、不终止。
    /// 每段不同的内容只发一次;不带原文(日志与事件都不记用户输入)。
    InputModified {
        text_length: usize,
    },
}

#[derive(Clone, Debug, PartialEq, Eq, Serialize, Deserialize)]
pub struct FallbackVideo {
    pub vid: String,
    pub api: String,
}

/// 捕获结果。
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct CaptureOutcome {
    pub originals: Vec<String>,
    pub watermark_source: WatermarkSource,
    pub conversation_id: Option<String>,
    pub submission_id: Option<String>,
}

// ------------------------------------------------------------------
// 错误(11-error-codes `generation_*`)
// ------------------------------------------------------------------

/// `generation_task_timeout`。
pub fn task_timeout_error(deadline: Duration, last_stage: Option<&str>) -> RelayError {
    RelayError::new(
        "generation_task_timeout",
        "豆包在限定时间内未返回结果,请稍后重试或检查豆包页面",
    )
    .with_details(json!({
        "deadline_ms": deadline.as_millis() as u64,
        "last_stage": last_stage,
    }))
}

/// `generation_user_interrupted{interruption}`。
pub fn user_interrupted_error(interruption: &str) -> RelayError {
    RelayError::new(
        "generation_user_interrupted",
        "任务被用户操作打断,请保持豆包只有一个对话窗口后重试",
    )
    .with_details(json!({ "interruption": interruption }))
}

// ------------------------------------------------------------------
// 纯函数
// ------------------------------------------------------------------

/// 解码常见 HTML 实体(`&amp;` `&lt;` `&gt;` `&quot;` `&#39;` `&#NN;` `&#xHH;`)。
pub fn html_decode(s: &str) -> String {
    if !s.contains('&') {
        return s.to_string();
    }
    let mut out = String::with_capacity(s.len());
    let mut rest = s;
    while let Some(pos) = rest.find('&') {
        out.push_str(&rest[..pos]);
        rest = &rest[pos..];
        let Some(end) = rest.find(';') else {
            out.push_str(rest);
            return out;
        };
        let entity = &rest[1..end];
        let decoded = match entity {
            "amp" => Some('&'),
            "lt" => Some('<'),
            "gt" => Some('>'),
            "quot" => Some('"'),
            "apos" => Some('\''),
            _ => entity
                .strip_prefix('#')
                .and_then(|num| {
                    if let Some(hex) = num.strip_prefix('x').or_else(|| num.strip_prefix('X')) {
                        u32::from_str_radix(hex, 16).ok()
                    } else {
                        num.parse::<u32>().ok()
                    }
                })
                .and_then(char::from_u32),
        };
        match decoded {
            Some(c) if end <= 12 => {
                out.push(c);
                rest = &rest[end + 1..];
            }
            _ => {
                out.push('&');
                rest = &rest[1..];
            }
        }
    }
    out.push_str(rest);
    out
}

fn host_in(host: &str, allowed: &[&str]) -> bool {
    let host = host.to_ascii_lowercase();
    allowed
        .iter()
        .any(|d| host == *d || host.ends_with(&format!(".{d}")))
}

/// https 且 host 为 `byteimg.com` / `volces.com`(含子域)。
pub fn is_image_cdn_url(url: &str) -> bool {
    let Ok(u) = url::Url::parse(url) else {
        return false;
    };
    u.scheme() == "https" && u.host_str().is_some_and(|h| host_in(h, IMAGE_CDN_HOSTS))
}

/// 判断 DOM 中的图片是否足够像本次生成的高清原图。
/// 豆包不同版本使用的 CDN 路径不固定,不能只依赖 `/rc_gen_image/` 缩略图前缀。
#[cfg(test)]
fn is_dom_generation_image(url: &str, width: u64, height: u64) -> bool {
    is_image_cdn_url(url) && width >= 500 && height >= 500
}

/// DOM 扫描的第一次结果只建立当前页面基线,后续新出现的地址才属于本次任务。
/// 这样即使豆包没有回传 `submission_id`,也能把本次生成后新增的图片交给当前任务。
fn collect_new_dom_images(
    baseline_ready: &mut bool,
    known: &mut HashSet<String>,
    urls: Vec<String>,
) -> Vec<String> {
    if !*baseline_ready {
        known.extend(urls);
        *baseline_ready = true;
        return Vec::new();
    }
    urls.into_iter().filter(|url| known.insert(url.clone())).collect()
}

/// 视频路径 1(analysis/05,capture.ps1:265-267):`responseReceived` 的 MIME 为 `video/*`,
/// URL 为 https、无 userinfo、host 在视频 CDN 白名单 → 返回该 URL(原样,不重编码)。
pub fn video_candidate_url(url: &str, mime: &str) -> Option<String> {
    if !mime.to_ascii_lowercase().starts_with("video/") {
        return None;
    }
    let u = url::Url::parse(url).ok()?;
    if u.scheme() != "https" || !u.username().is_empty() || u.password().is_some() {
        return None;
    }
    is_video_cdn_host(u.host_str()?).then(|| url.to_string())
}

/// 视频路径 3 的 API 白名单:`https://vas-[a-z0-9-]+.snssdk.com`、无 userinfo(capture.ps1:351)。
pub fn is_fallback_api_url(api: &str) -> bool {
    let Ok(u) = url::Url::parse(api) else {
        return false;
    };
    if u.scheme() != "https" || !u.username().is_empty() || u.password().is_some() {
        return false;
    }
    let Some(host) = u.host_str() else {
        return false;
    };
    let Some(id) = host
        .strip_prefix("vas-")
        .and_then(|h| h.strip_suffix(".snssdk.com"))
    else {
        return false;
    };
    !id.is_empty()
        && id
            .bytes()
            .all(|b| b.is_ascii_lowercase() || b.is_ascii_digit() || b == b'-')
}

/// 解密出的地址 → 无水印原片地址:已带 `lr=unwatermarked`(reader `unwatermarkedURL` 通过)原样用;
/// 是视频 CDN 地址但没有 `lr` → 按计划追加 `lr=unwatermarked`(`video_decoder::unwatermarked_url`);
/// 不是视频 CDN 地址 → None。
pub fn fallback_original_url(decoded: &str) -> Option<String> {
    if let Some(u) = unwatermarked_video_url(decoded) {
        return Some(u);
    }
    video_url(decoded).map(|u| unwatermarked_url(&u))
}

/// 视频路径 3 的 JS 兜底:按守卫注入原版解密脚本,`decode(JSON.parse(body), api)` → `{ok, url?, reason?}`。
pub fn video_decoder_expression(body: &str, api: &str) -> String {
    format!(
        "(() => {{ if (typeof {VIDEO_DECODER_GLOBAL} === 'undefined') {{\n{VIDEO_DECODER_SCRIPT}\n}}\nreturn Promise.resolve({VIDEO_DECODER_GLOBAL}.decode(JSON.parse({}), {})); }})()",
        Value::String(body.to_string()),
        Value::String(api.to_string()),
    )
}

fn collect_originals(value: &Value, depth: usize, out: &mut Vec<String>) {
    if depth > 50 {
        return;
    }
    match value {
        Value::String(s) => {
            let t = s.trim_start();
            if t.starts_with('{') || t.starts_with('[') {
                if let Ok(parsed) = serde_json::from_str::<Value>(s) {
                    collect_originals(&parsed, depth + 1, out);
                }
            }
        }
        Value::Array(items) => {
            for item in items {
                collect_originals(item, depth + 1, out);
            }
        }
        Value::Object(map) => {
            for (k, v) in map {
                if k == "image_ori_raw" {
                    if let Some(url) = v.get("url").and_then(Value::as_str) {
                        let url = html_decode(url);
                        if is_image_cdn_url(&url) && !out.iter().any(|u| u == &url) {
                            out.push(url);
                        }
                    }
                }
                collect_originals(v, depth + 1, out);
            }
        }
        _ => {}
    }
}

/// 递归找 `image_ori_raw.url`(https,host byteimg.com|volces.com),去重保序;
/// 整体不是 JSON 时按 SSE `data: {...}` 逐行解析;字符串字段里的嵌套 JSON 也会展开。
pub fn extract_image_originals(body: &str) -> Vec<String> {
    let mut out = Vec::new();
    if body.is_empty() || body.len() > MAX_BODY_BYTES {
        return out;
    }
    if let Ok(v) = serde_json::from_str::<Value>(body) {
        collect_originals(&v, 0, &mut out);
        return out;
    }
    for line in body.split('\n') {
        let line = line.trim_end_matches('\r');
        let Some(rest) = line.strip_prefix("data:") else {
            continue;
        };
        let rest = rest.trim();
        if rest.is_empty() {
            continue;
        }
        if let Ok(v) = serde_json::from_str::<Value>(rest) {
            collect_originals(&v, 0, &mut out);
        }
    }
    out
}

/// 历史查询请求(analysis/05):https、host `www.doubao.com`、POST、
/// path `/im/chain/single` 或 `/im/message/chain_by_block`(可带尾斜杠)。
pub fn is_history_request(url: &str, method: &str) -> bool {
    if !method.eq_ignore_ascii_case("POST") {
        return false;
    }
    let Ok(u) = url::Url::parse(url) else {
        return false;
    };
    if u.scheme() != "https" || u.host_str() != Some("www.doubao.com") {
        return false;
    }
    let path = u.path().trim_end_matches('/');
    path == "/im/chain/single" || path == "/im/message/chain_by_block"
}

/// 优先级更高的历史请求形态(全消息链)。
pub fn is_chain_single(url: &str) -> bool {
    url::Url::parse(url)
        .map(|u| u.path().trim_end_matches('/') == "/im/chain/single")
        .unwrap_or(false)
}

/// `requestWillBeSent.params.initiator.stack.callFrames[].functionName` 含 reader 自家重放函数。
pub fn is_reader_initiated(params: &Value) -> bool {
    fn walk(stack: &Value, depth: usize) -> bool {
        if depth > 8 {
            return false;
        }
        let frames = stack.get("callFrames").and_then(Value::as_array);
        if frames.is_some_and(|f| {
            f.iter()
                .any(|c| c.get("functionName").and_then(Value::as_str) == Some(READER_INITIATOR_FN))
        }) {
            return true;
        }
        stack
            .get("parent")
            .is_some_and(|parent| walk(parent, depth + 1))
    }
    params
        .get("initiator")
        .and_then(|i| i.get("stack"))
        .is_some_and(|s| walk(s, 0))
}

/// 从历史请求 postData 里扫出所有 `conversation_id`(数字或字符串),去重保序。
pub fn extract_conversation_ids(post_data: &str) -> Vec<String> {
    let mut out: Vec<String> = Vec::new();
    let needle = "\"conversation_id\"";
    let mut rest = post_data;
    while let Some(pos) = rest.find(needle) {
        rest = &rest[pos + needle.len()..];
        let after = rest.trim_start();
        let Some(after) = after.strip_prefix(':') else {
            continue;
        };
        let after = after.trim_start();
        let after = after.strip_prefix('"').unwrap_or(after);
        let digits: String = after.chars().take_while(|c| c.is_ascii_digit()).collect();
        if !digits.is_empty() && digits != "0" && !out.contains(&digits) {
            out.push(digits);
        }
    }
    out
}

/// 捕获到的历史查询请求(重放模板)。
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct HistoryRequest {
    pub url: String,
    pub post_data: String,
    pub headers: Value,
    pub conversation_id: Option<String>,
}

impl HistoryRequest {
    fn to_js(&self) -> Value {
        json!({
            "url": self.url,
            "method": "POST",
            "postData": self.post_data,
            "headers": self.headers,
        })
    }
}

/// `read` / `analyze` 的 options(analysis/05 "analyze")。`requireBinding` 只在已知
/// `boundConversationId` 时为 true(否则 reader 一律返回空,新对话永远绑不上);
/// 图像只在 `binding` 已知时采信 `images`,严格性由 Rust 侧保证。
#[allow(clippy::too_many_arguments)]
pub fn read_options(
    task_type: TaskType,
    prompt: &str,
    identity: &SubmissionIdentity,
    send_since: i64,
    bound_conversation_id: Option<&str>,
    bound_submission_id: Option<&str>,
    bound_submission_created_at: Option<&Value>,
    followup_submission_ids: &[String],
) -> Value {
    json!({
        "nativeShadow": false,
        "followupSubmissionIds": followup_submission_ids,
        "requireBinding": bound_conversation_id.is_some(),
        "requireIdentity": true,
        "expectedLocalMessageId": identity.local_message_id,
        "expectedMessageId": identity.message_id,
        "expectedConversationId": identity.conversation_id,
        "boundSubmissionId": bound_submission_id,
        "boundSubmissionCreatedAt": bound_submission_created_at.cloned().unwrap_or(Value::Null),
        "boundConversationId": bound_conversation_id,
        "prompt": prompt,
        "since": send_since,
        "type": task_type.as_str(),
    })
}

/// `Promise.resolve(shenlangDoubaoResultReader.read(<req>, <options>))`。
pub fn read_expression(history: &HistoryRequest, options: &Value) -> String {
    format!(
        "Promise.resolve({READER_GLOBAL}.read({}, {}))",
        history.to_js(),
        options
    )
}

/// `Promise.resolve(shenlangDoubaoResultReader.analyze(<body>, <req>, <options>))`;body 原样传
/// (reader 内部自己做大整数修复 + `JSON.parse`)。
pub fn analyze_expression(body: &str, history: &HistoryRequest, options: &Value) -> String {
    format!(
        "Promise.resolve({READER_GLOBAL}.analyze({}, {}, {}))",
        Value::String(body.to_string()),
        history.to_js(),
        options
    )
}

/// `shenlangDoubaoResultReader.streamSubmissionBinding(<sse 文本>, {expectedLocalMessageId, expectedMessageId})`。
pub fn stream_binding_expression(body: &str, identity: &SubmissionIdentity) -> String {
    format!(
        "{READER_GLOBAL}.streamSubmissionBinding({}, {})",
        Value::String(body.to_string()),
        json!({
            "expectedLocalMessageId": identity.local_message_id,
            "expectedMessageId": identity.message_id,
        })
    )
}

/// feedback 步骤的请求对象(`terminalFeedback` 只读 type/prompt;`videoFollowup` 另看确认相关字段与
/// `modelName/ratio/duration/referenceFiles`——脚本只用 `referenceFiles.length`,这里给占位名)。
pub fn feedback_payload(
    opts: &CaptureOptions,
    mode: ConfirmationMode,
    auto_membership: bool,
    user_confirmed: bool,
) -> Value {
    let reference_files: Vec<String> = (1..=opts.reference_count)
        .map(|i| format!("reference_{i}"))
        .collect();
    json!({
        "type": opts.task_type.as_str(),
        "prompt": opts.prompt,
        "modelName": opts.model_name,
        "ratio": opts.ratio,
        "duration": opts.duration,
        "referenceFiles": reference_files,
        "trackingKey": opts.tracking_key,
        "parallelTracking": true,
        "confirmationMode": mode.as_str(),
        "autoUseMembership": auto_membership,
        "userConfirmed": user_confirmed,
    })
}

/// feedback 轮询表达式:`bootstrap` 为 true 时先按守卫注入(页面 reload 后全局对象丢失),
/// 全局对象不存在时回 `{stage:"unknown", missing:true}`,否则 `Promise.resolve(<key>.run(step, payload))`。
pub fn feedback_expression(
    tracking_key: &str,
    step: &str,
    payload: &Value,
    bootstrap: bool,
) -> String {
    let guard = format!(
        "typeof {key} === 'undefined' ? ({{ stage: 'unknown', missing: true }}) : Promise.resolve({key}.run({step}, {payload}))",
        key = tracking_key,
        step = Value::String(step.to_string()),
        payload = payload
    );
    if bootstrap {
        format!("{}\n{guard}", generation_bootstrap(tracking_key))
    } else {
        guard
    }
}

/// 视频路径 2(capture.ps1:421-424):`run('requestVideoDownload', payload)`,脚本在 stage 为 `result`
/// 且页面上恰好一个 `<video>` 时点它旁边的"下载"按钮,返回 `{clicked, reason?}`;豆包随后请求的
/// `lr=unwatermarked` 原片由路径 1 的 `responseReceived` 接住。
pub fn request_video_download_expression(
    tracking_key: &str,
    payload: &Value,
    bootstrap: bool,
) -> String {
    feedback_expression(tracking_key, "requestVideoDownload", payload, bootstrap)
}

/// `acceptFeedback` 的请求对象(复刻 doubao-capture.ps1:343):reader `read` / `analyze` 返回的
/// `feedback` 是**消息对象** `{message, messageId, submissionId, createdAt, scope}`(没有 `stage`),
/// 原样作为 `reply` 交给生成脚本分类(`feedbackKind`:会员额度确认 → `quota_checking`、
/// 额度用完 → `quota_exhausted`、视频参数确认 → `checking`……)。`reply` 必须是对象,字符串一律 `unknown`。
pub fn accept_feedback_payload(
    opts: &CaptureOptions,
    bound_submission_id: &str,
    reply: &Value,
) -> Value {
    json!({
        "boundSubmissionId": bound_submission_id,
        "reply": reply,
        "prompt": opts.prompt,
        "modelName": opts.model_name,
        "type": opts.task_type.as_str(),
    })
}

/// 历史反馈分类的去重键:messageId + 消息文本(流式输出中的半句与完整文本是两次分类)。
pub fn classify_key(message_id: &str, reply: &Value) -> String {
    let text = reply.get("message").and_then(Value::as_str).unwrap_or("");
    format!("{message_id}\n{text}")
}

/// `typeof <key>==='undefined' ? {stage:'unknown',missing:true} : Promise.resolve(<key>.run('acceptFeedback', payload))`
/// (`bootstrap` 为 true 时先按守卫注入)。只分类、不发送;真正的"继续生成"仍走 [`accept_reply_expression`]。
pub fn accept_feedback_expression(tracking_key: &str, payload: &Value, bootstrap: bool) -> String {
    feedback_expression(tracking_key, "acceptFeedback", payload, bootstrap)
}

/// 图像代回复:复刻 `videoFollowup` 内部的发送路径——输入框为空(或已是回复文本)时
/// `run('fill', {...payload, prompt: reply})`(沈浪自己的 fill + 核对),再点唯一且可用的发送按钮,
/// 等输入框清空。返回 `{sent:true, cleared, waitedMs}`;发送按钮未就绪 → `{sent:false, pending:true, reason}`;
/// 全局对象丢失 → `{sent:false, missing:true}`;输入区有用户内容 → 抛异常(与沈浪同文案,走降级链)。
/// 不调用脚本的 `acceptFeedback`(那是 native-shadow 消息的分类钩子,不发送)。
pub fn accept_reply_expression(tracking_key: &str, payload: &Value, reply: &str) -> String {
    let mut p = payload.clone();
    p["prompt"] = Value::String(reply.to_string());
    format!(
        r#"(async () => {{ /* {marker} */
  const REPLY = {reply};
  const SEND = {send:?};
  const sleep = ms => new Promise(r => setTimeout(r, ms));
  const visible = el => {{
    if (!el || !el.isConnected) return false;
    const r = el.getBoundingClientRect();
    return r.width > 0 && r.height > 0 && getComputedStyle(el).visibility !== 'hidden' && getComputedStyle(el).display !== 'none';
  }};
  const editor = () => document.querySelector('[contenteditable="true"]');
  const text = () => {{ const ed = editor(); return ed ? String(ed.innerText || ed.textContent || '').replace(/\r\n/g, '\n').trim() : ''; }};
  const norm = s => String(s || '').normalize('NFKC').replace(/[\s​-‍﻿]+/g, '');
  if (typeof {key} === 'undefined') return {{ sent: false, missing: true }};
  if (!editor()) throw new Error('豆包输入框未找到，未发送继续生成');
  const current = text();
  if (current && norm(current) !== norm(REPLY)) throw new Error('豆包输入区已有内容，未发送继续生成');
  if (!current) await {key}.run('fill', {payload});
  const buttons = () => [...document.querySelectorAll(SEND)].filter(visible).filter(b => !b.disabled && b.getAttribute('aria-disabled') !== 'true');
  const t0 = Date.now();
  let btns = buttons();
  while (!btns.length && Date.now() - t0 < 1500) {{ await sleep(50); btns = buttons(); }}
  if (btns.length !== 1) return {{ sent: false, pending: true, reason: btns.length ? 'send_button_ambiguous' : 'send_button_missing', waitedMs: Date.now() - t0 }};
  btns[0].click();
  const t1 = Date.now();
  let cleared = false;
  while (Date.now() - t1 < 2000) {{ if (!text()) {{ cleared = true; break; }} await sleep(50); }}
  return {{ sent: true, cleared, waitedMs: Date.now() - t0 }};
}})()"#,
        marker = ACCEPT_REPLY_MARKER,
        reply = Value::String(reply.to_string()),
        send = REPLY_SEND_BUTTON_SELECTOR,
        key = tracking_key,
        payload = p,
    )
}

/// 输入框内容是否属于"正常":空、等于本任务 prompt、或是 relay 自己代回复的"继续生成"
/// (都按 trim 后比较)。其它内容视为用户改动(Task 4.2)。
pub fn input_is_expected(text: &str, prompt: &str) -> bool {
    let t = text.trim();
    t.is_empty() || t == prompt.trim() || t == CONTINUE_REPLY
}

/// 输入框脏检查:返回 `{found, text}`(text 已 trim)。
pub fn input_check_expression() -> String {
    r#"(() => {
  const ed = document.querySelector('[contenteditable="true"]');
  if (!ed) return { found: false, text: '' };
  return { found: true, text: (ed.innerText || '').trim() };
})()"#
        .to_string()
}

// ------------------------------------------------------------------
// 求值 / 后台作业
// ------------------------------------------------------------------

#[derive(Debug)]
enum EvalError {
    Js(String),
    Cdp(CdpError),
}

impl std::fmt::Display for EvalError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            EvalError::Js(m) => write!(f, "页面脚本异常:{m}"),
            EvalError::Cdp(e) => write!(f, "{e}"),
        }
    }
}

async fn eval(
    session: Arc<CdpSession>,
    expr: String,
    timeout: Duration,
) -> Result<Value, EvalError> {
    let result = session
        .send_timeout(
            "Runtime.evaluate",
            json!({ "expression": expr, "returnByValue": true, "awaitPromise": true }),
            timeout,
        )
        .await
        .map_err(EvalError::Cdp)?;
    if let Some(details) = result.get("exceptionDetails") {
        return Err(EvalError::Js(extract_js_exception(details)));
    }
    Ok(result
        .get("result")
        .and_then(|r| r.get("value"))
        .cloned()
        .unwrap_or(Value::Null))
}

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum BodyKind {
    History,
    Generation,
}

enum Job {
    Feedback(Result<Value, EvalError>),
    /// 历史接口没有被监听到时,从当前豆包消息 DOM 兜底读取已渲染的原图。
    DomImages(Result<Value, EvalError>),
    /// relay 自己读最后一条机器人回复(视频确认话术兜底)。
    RelayConfirm(Result<Value, EvalError>),
    Accept {
        stage: String,
        message: String,
        result: Result<Value, EvalError>,
    },
    /// 历史 feedback 消息经 `acceptFeedback` 分类的结果。
    Classify {
        message_id: String,
        /// 去重键(messageId + 消息文本,见 [`classify_key`])。
        key: String,
        result: Result<Value, EvalError>,
    },
    Read(Result<Value, EvalError>),
    Analyze(Result<Value, EvalError>),
    StreamBinding(Result<Value, EvalError>),
    Body {
        kind: BodyKind,
        request_id: String,
        body: Option<String>,
    },
    PostData {
        request_id: String,
        url: String,
        headers: Value,
        post_data: Option<String>,
    },
    InputCheck(Result<Value, EvalError>),
    Targets(Result<Vec<CdpTarget>, CdpError>),
    /// 视频路径 2:`requestVideoDownload` 的返回。
    DownloadClick(Result<Value, EvalError>),
    /// 视频路径 3:fallback API + 解密的结果(`Ok((解密出的地址, "rust"|"js"))` / `Err(原因)`)。
    Fallback {
        vid: String,
        result: Result<(String, &'static str), String>,
    },
}

/// 视频路径 3 的一次尝试:请求 API → Rust 解密 → 失败(或解出非视频 CDN 地址)时注入原版脚本解密(记 warn)。
async fn run_fallback(
    session: Arc<CdpSession>,
    fetcher: Arc<dyn FallbackFetcher>,
    item: FallbackVideo,
    fetch_timeout: Duration,
    eval_timeout: Duration,
) -> Result<(String, &'static str), String> {
    let body = fetcher.fetch(&item.api, fetch_timeout).await?;
    if body.len() > FALLBACK_API_MAX_BYTES {
        return Err("fallback_response_too_large".into());
    }
    let parsed: Value =
        serde_json::from_str(&body).map_err(|_| "fallback_response_not_json".to_string())?;
    // Rust 解出来但不是视频 CDN 地址(`fallback_original_url` 拒收)也按 Rust 失败处理,
    // 改用页面脚本,避免同一个 vid 每 10s 重试同样的坏结果直到 20min 超时。
    let rust_reason = match video_decoder::decode(&parsed, &item.api) {
        Ok(url) if fallback_original_url(&url).is_some() => return Ok((url, "rust")),
        Ok(_) => "rust_decoded_non_video_url",
        Err(reason) => reason,
    };
    tracing::warn!(vid = %item.vid, reason = rust_reason, "Rust 解密视频 token 失败,改用页面脚本解密");
    match eval(
        session,
        video_decoder_expression(&body, &item.api),
        eval_timeout,
    )
    .await
    {
        Ok(v) if v.get("ok").and_then(Value::as_bool) == Some(true) => v
            .get("url")
            .and_then(Value::as_str)
            .filter(|u| !u.is_empty())
            .map(|u| (u.to_string(), "js"))
            .ok_or_else(|| rust_reason.to_string()),
        Ok(v) => Err(v
            .get("reason")
            .and_then(Value::as_str)
            .unwrap_or(rust_reason)
            .to_string()),
        Err(e) => Err(format!("{rust_reason}:{e}")),
    }
}

/// `Network.getResponseBody`:base64 解码;失败 → None。
async fn fetch_body(
    session: Arc<CdpSession>,
    request_id: String,
    timeout: Duration,
) -> Option<String> {
    match session
        .send_timeout(
            "Network.getResponseBody",
            json!({ "requestId": request_id }),
            timeout,
        )
        .await
    {
        Ok(v) => {
            let body = v.get("body").and_then(Value::as_str)?;
            if v.get("base64Encoded").and_then(Value::as_bool) == Some(true) {
                use base64::Engine;
                let bytes = base64::engine::general_purpose::STANDARD
                    .decode(body)
                    .ok()?;
                Some(String::from_utf8_lossy(&bytes).into_owned())
            } else {
                Some(body.to_string())
            }
        }
        Err(e) => {
            tracing::debug!(error = %e, request_id, "读取响应体失败");
            None
        }
    }
}

async fn fetch_post_data(
    session: Arc<CdpSession>,
    request_id: String,
    timeout: Duration,
) -> Option<String> {
    session
        .send_timeout(
            "Network.getRequestPostData",
            json!({ "requestId": request_id }),
            timeout,
        )
        .await
        .ok()?
        .get("postData")
        .and_then(Value::as_str)
        .map(str::to_string)
}

/// `Network.enable`:Protocol 错误只 warn;断线 → Err。
async fn enable_network(session: &CdpSession) -> RelayResult<()> {
    match session
        .send(
            "Network.enable",
            json!({
                "maxPostDataSize": MAX_POST_DATA_SIZE,
                "maxTotalBufferSize": 16_777_216u64,
                "maxResourceBufferSize": 8_388_608u64,
            }),
        )
        .await
    {
        Ok(_) => Ok(()),
        Err(CdpError::Protocol { code, message }) => {
            tracing::warn!(code, message = %message, target_id = %session.target_id, "Network.enable 被拒绝,忽略");
            Ok(())
        }
        Err(e) => Err(e.into()),
    }
}

/// 取当前 background session 并订阅 + `Network.enable`;没有 / 已关闭 / enable 失败 → `(None, None)`。
async fn attach_background(
    conn: &DoubaoConnection,
) -> (
    Option<Arc<CdpSession>>,
    Option<broadcast::Receiver<CdpEvent>>,
) {
    let Some(bg) = conn.background_session().await else {
        return (None, None);
    };
    let rx = bg.subscribe();
    if let Err(e) = enable_network(&bg).await {
        tracing::warn!(code = e.code, error = %e, "background 页 Network.enable 失败,只监听 chat 页");
        return (None, None);
    }
    (Some(bg), Some(rx))
}

async fn next_event(rx: &mut Option<broadcast::Receiver<CdpEvent>>) -> CdpEvent {
    let Some(rx) = rx.as_mut() else {
        return std::future::pending().await;
    };
    loop {
        match rx.recv().await {
            Ok(ev) => return ev,
            Err(broadcast::error::RecvError::Lagged(n)) => {
                tracing::warn!(skipped = n, "CDP 事件积压,已跳过");
            }
            Err(broadcast::error::RecvError::Closed) => return std::future::pending().await,
        }
    }
}

async fn next_job(jobs: &mut JoinSet<Job>) -> Job {
    if jobs.is_empty() {
        return std::future::pending().await;
    }
    loop {
        match jobs.join_next().await {
            Some(Ok(job)) => return job,
            Some(Err(e)) => tracing::warn!(error = %e, "捕获后台作业异常退出"),
            None => return std::future::pending().await,
        }
    }
}

/// 等用户**新的一次**确认:边沿触发——只在 `confirm_rx` 收到新值且为 true 时返回
/// (runner 每次 `POST /confirm` 都 `send(true)`,`watch` 不比较旧值、仍会通知),
/// 电平已是 true 不算,避免一次确认被反复消费;发送端已丢弃则永远挂起。
async fn confirmed(rx: &mut watch::Receiver<bool>) {
    loop {
        if rx.changed().await.is_err() {
            std::future::pending::<()>().await;
        }
        if *rx.borrow_and_update() {
            return;
        }
    }
}

// ------------------------------------------------------------------
// 循环状态
// ------------------------------------------------------------------

struct State<'a> {
    opts: &'a CaptureOptions,
    timing: CaptureTiming,
    events: &'a mpsc::Sender<CaptureEvent>,
    originals: Vec<String>,
    bound_conversation_id: Option<String>,
    bound_submission_id: Option<String>,
    bound_created_at: Option<Value>,
    followup_ids: Vec<String>,
    bound_emitted: bool,
    history: Option<HistoryRequest>,
    history_pending: HashSet<String>,
    /// 生成请求 `(target_id, request_id)`;预置 `opts.generation_requests`。
    generation_requests: HashSet<(String, String)>,
    /// 视频任务:`send_since` 之后的所有 requestId(PS `videoRequests`),供 `responseReceived` 匹配 `video/*`。
    video_requests: HashSet<String>,
    sse_backlog: Vec<String>,
    sse_bytes: usize,
    sse_consumed: usize,
    mode: ConfirmationMode,
    auto_membership: bool,
    /// 用户本次确认尚未被消费(代回复发出后回落为 false)。
    user_confirmed: bool,
    needs_confirm_emitted: bool,
    /// 已代回复过(`sent:true`)的确认提示:相同 stage 的不同提示也必须分别回复;降级后清空。
    accepted_confirmation: Option<(String, String)>,
    /// 当前 stage 已尝试代回复但未发出(`sent:false`)的次数,达到 [`ACCEPT_MAX_ATTEMPTS`] 走降级链。
    accept_attempts: u32,
    last_feedback: Option<FeedbackState>,
    last_stage_js: Option<String>,
    last_emitted_stage: Option<TaskStage>,
    unverified_since: Option<Instant>,
    reader_ready: bool,
    gen_ready: bool,
    feedback_inflight: bool,
    dom_images_inflight: bool,
    relay_confirm_inflight: bool,
    accept_inflight: bool,
    read_inflight: bool,
    analyze_inflight: bool,
    binding_inflight: bool,
    input_inflight: bool,
    /// 最近一次报告过的"被修改"的输入框内容(相同内容不重复报;恢复正常后清空)。
    input_modified: Option<String>,
    targets_inflight: bool,
    target_failures: u32,
    video_candidates: Vec<String>,
    /// 已送去 `acceptFeedback` 分类的历史 feedback messageId(每条只分类一次)。
    classified_feedback: HashSet<String>,
    classify_inflight: bool,
    /// 历史 feedback 分类出的、尚未代回复的确认 stage(`terminalFeedback` 识别不出会员额度确认,
    /// 用户确认 / 代回复重试都靠它在 feedback 节拍上重新触发)。
    history_confirm: Option<FeedbackState>,
    first_original_at: Option<Instant>,
    last_original_at: Option<Instant>,
    /// 第一条原图 / 原片来自哪条路径(决定 `watermark_source`)。
    original_source: Option<WatermarkSource>,
    /// DOM 兜底首次扫描的已知图片,用于区分历史图与本次新生成图。
    dom_baseline_ready: bool,
    dom_known_images: HashSet<String>,
    /// 视频路径 2:"下载"按钮已点到(`clicked:true`),不再点。
    download_clicked: bool,
    download_inflight: bool,
    next_download_at: Option<Instant>,
    /// 视频路径 3:每个 vid 最近一次尝试的时刻(10s 内不重复)与进行中的 vid。
    fallback_attempts: HashMap<String, Instant>,
    fallback_inflight: HashSet<String>,
    fetcher: Arc<dyn FallbackFetcher>,
}

impl<'a> State<'a> {
    fn new(
        opts: &'a CaptureOptions,
        timing: CaptureTiming,
        events: &'a mpsc::Sender<CaptureEvent>,
        fetcher: Arc<dyn FallbackFetcher>,
    ) -> Self {
        Self {
            opts,
            timing,
            events,
            originals: Vec::new(),
            bound_conversation_id: opts.identity.conversation_id.clone(),
            bound_submission_id: None,
            bound_created_at: None,
            followup_ids: Vec::new(),
            bound_emitted: false,
            history: None,
            history_pending: HashSet::new(),
            generation_requests: opts
                .generation_requests
                .iter()
                .map(|r| (r.target_id.clone(), r.request_id.clone()))
                .collect(),
            video_requests: HashSet::new(),
            sse_backlog: Vec::new(),
            sse_bytes: 0,
            sse_consumed: 0,
            mode: opts.confirmation_mode,
            auto_membership: opts.auto_use_membership,
            user_confirmed: false,
            needs_confirm_emitted: false,
            accepted_confirmation: None,
            accept_attempts: 0,
            last_feedback: None,
            last_stage_js: None,
            last_emitted_stage: None,
            unverified_since: None,
            reader_ready: false,
            gen_ready: false,
            feedback_inflight: false,
            dom_images_inflight: false,
            relay_confirm_inflight: false,
            accept_inflight: false,
            read_inflight: false,
            analyze_inflight: false,
            binding_inflight: false,
            input_inflight: false,
            input_modified: None,
            targets_inflight: false,
            target_failures: 0,
            video_candidates: Vec::new(),
            classified_feedback: HashSet::new(),
            classify_inflight: false,
            history_confirm: None,
            first_original_at: None,
            last_original_at: None,
            original_source: None,
            dom_baseline_ready: false,
            dom_known_images: HashSet::new(),
            download_clicked: false,
            download_inflight: false,
            next_download_at: None,
            fallback_attempts: HashMap::new(),
            fallback_inflight: HashSet::new(),
            fetcher,
        }
    }

    async fn emit(&self, ev: CaptureEvent) {
        if self.events.send(ev).await.is_err() {
            tracing::debug!("捕获事件接收端已关闭");
        }
    }

    /// 原图到齐的判定时刻:最后一次新增 + `result_settle`,但不晚于第一次 + `result_settle_max`。
    fn settle_at(&self) -> Option<Instant> {
        let first = self.first_original_at?;
        let last = self.last_original_at.unwrap_or(first);
        Some((last + self.timing.result_settle).min(first + self.timing.result_settle_max))
    }

    /// 有原图且已过静默窗口才返回结果(多图可能分几次到达,不能拿到第一张就走)。
    fn settled_outcome(&self, now: Instant) -> Option<CaptureOutcome> {
        if self.settle_at().is_some_and(|t| now < t) {
            return None;
        }
        self.outcome()
    }

    fn outcome(&self) -> Option<CaptureOutcome> {
        if self.originals.is_empty() {
            return None;
        }
        Some(CaptureOutcome {
            originals: self.originals.clone(),
            watermark_source: self
                .original_source
                .unwrap_or(WatermarkSource::ConversationOriginal),
            conversation_id: self.bound_conversation_id.clone(),
            submission_id: self.bound_submission_id.clone(),
        })
    }

    /// 页面 reload / 重连后脚本全局对象丢失,下次求值前重新注入。
    fn reset_page_scripts(&mut self) {
        self.reader_ready = false;
        self.gen_ready = false;
    }

    fn identity_markers(&self) -> Vec<String> {
        let mut v: Vec<String> = Vec::new();
        for s in [
            self.opts.identity.local_message_id.as_ref(),
            self.opts.identity.message_id.as_ref(),
            self.bound_submission_id.as_ref(),
        ]
        .into_iter()
        .flatten()
        {
            if s.len() >= 6 && !v.contains(s) {
                v.push(s.clone());
            }
        }
        v
    }

    async fn add_originals(&mut self, urls: Vec<String>, source: &str, kind: WatermarkSource) {
        let mut added = Vec::new();
        for u in urls {
            if !self.originals.contains(&u) {
                self.originals.push(u.clone());
                added.push(u);
            }
        }
        if !added.is_empty() {
            let now = Instant::now();
            self.first_original_at.get_or_insert(now);
            self.last_original_at = Some(now);
            self.original_source.get_or_insert(kind);
            tracing::info!(
                count = self.originals.len(),
                source,
                watermark_source = kind.as_str(),
                "已捕获无水印原图 / 原片地址"
            );
            self.emit(CaptureEvent::Originals(self.originals.clone()))
                .await;
        }
    }

    // ---- Network 事件 ----

    async fn handle_event(
        &mut self,
        session: &Arc<CdpSession>,
        ev: CdpEvent,
        jobs: &mut JoinSet<Job>,
    ) {
        match ev.method.as_str() {
            "Network.requestWillBeSent" => self.on_request(session, &ev.params, jobs).await,
            "Network.loadingFinished" => {
                let Some(request_id) = ev.params.get("requestId").and_then(Value::as_str) else {
                    return;
                };
                let kind = if self.history_pending.remove(request_id) {
                    BodyKind::History
                } else if self.is_generation_request_id(session, request_id) {
                    BodyKind::Generation
                } else {
                    return;
                };
                let s = Arc::clone(session);
                let id = request_id.to_string();
                let timeout = self.timing.command_timeout;
                jobs.spawn(async move {
                    let body = fetch_body(s, id.clone(), timeout).await;
                    Job::Body {
                        kind,
                        request_id: id,
                        body,
                    }
                });
            }
            "Network.eventSourceMessageReceived" => {
                let Some(request_id) = ev.params.get("requestId").and_then(Value::as_str) else {
                    return;
                };
                if !self.is_generation_request_id(session, request_id) {
                    return;
                }
                let data = ev
                    .params
                    .get("data")
                    .and_then(Value::as_str)
                    .unwrap_or("")
                    .to_string();
                if data.is_empty() {
                    return;
                }
                if self.bound_submission_id.is_none()
                    && self.sse_backlog.len() < SSE_BACKLOG_MAX_ENTRIES
                    && self.sse_bytes + data.len() <= SSE_BACKLOG_MAX_BYTES
                {
                    self.sse_bytes += data.len();
                    self.sse_backlog.push(data.clone());
                }
                self.maybe_extract_direct(&data, "sse").await;
            }
            "Network.responseReceived" => {
                if self.opts.task_type != TaskType::Video {
                    return;
                }
                let Some(request_id) = ev.params.get("requestId").and_then(Value::as_str) else {
                    return;
                };
                // PS 语义:只看 send_since 之后发出的请求(CDN 媒体是独立的 GET,不在生成请求集合里)
                if !self.video_requests.contains(request_id) {
                    return;
                }
                let resp = ev.params.get("response");
                let mime = resp
                    .and_then(|r| r.get("mimeType"))
                    .and_then(Value::as_str)
                    .unwrap_or("");
                let url = resp
                    .and_then(|r| r.get("url"))
                    .and_then(Value::as_str)
                    .unwrap_or("");
                let Some(url) = video_candidate_url(url, mime) else {
                    return;
                };
                if self.video_candidates.contains(&url) {
                    return;
                }
                self.video_candidates.push(url.clone());
                self.emit(CaptureEvent::VideoCandidates(self.video_candidates.clone()))
                    .await;
                // 路径 1:带 lr=unwatermarked 的直链就是无水印原片(含路径 2 点"下载"后豆包发出的请求)
                if has_lr_unwatermarked(&url) {
                    self.add_originals(vec![url], "network", WatermarkSource::LrUnwatermarked)
                        .await;
                } else {
                    tracing::debug!("视频直链不带 lr=unwatermarked,只作候选");
                }
            }
            _ => {}
        }
    }

    async fn on_request(
        &mut self,
        session: &Arc<CdpSession>,
        params: &Value,
        jobs: &mut JoinSet<Job>,
    ) {
        if is_reader_initiated(params) {
            return;
        }
        let Some(req) = params.get("request") else {
            return;
        };
        let url = req.get("url").and_then(Value::as_str).unwrap_or("");
        let method = req.get("method").and_then(Value::as_str).unwrap_or("");
        let request_id = params
            .get("requestId")
            .and_then(Value::as_str)
            .unwrap_or("")
            .to_string();
        if let Some(wall) = params.get("wallTime").and_then(Value::as_f64) {
            let wall_ms = (wall * 1000.0) as i64;
            // 视频:send_since 之后的所有请求都记下(analysis/05 `videoRequests`),含 CDN 媒体 GET。
            if self.opts.task_type == TaskType::Video
                && wall_ms >= self.opts.send_since
                && !request_id.is_empty()
            {
                self.video_requests.insert(request_id.clone());
            }
            // 发送之前就发出的请求与本任务无关(容忍 1.5s 时钟差)。
            if wall_ms + 1500 < self.opts.send_since {
                return;
            }
        }
        if is_history_request(url, method) {
            let headers = req.get("headers").cloned().unwrap_or(Value::Null);
            match req.get("postData").and_then(Value::as_str) {
                Some(post_data) => {
                    self.record_history(&request_id, url, headers, post_data.to_string());
                }
                None if req.get("hasPostData").and_then(Value::as_bool) == Some(true) => {
                    let s = Arc::clone(session);
                    let id = request_id.clone();
                    let url = url.to_string();
                    let timeout = self.timing.command_timeout;
                    jobs.spawn(async move {
                        let post_data = fetch_post_data(s, id.clone(), timeout).await;
                        Job::PostData {
                            request_id: id,
                            url,
                            headers,
                            post_data,
                        }
                    });
                }
                None => {}
            }
            return;
        }
        if is_generation_request(url, method) && !request_id.is_empty() {
            self.generation_requests
                .insert((session.target_id.clone(), request_id));
        }
    }

    fn is_generation_request_id(&self, session: &CdpSession, request_id: &str) -> bool {
        self.generation_requests
            .contains(&(session.target_id.clone(), request_id.to_string()))
    }

    /// 记录历史查询模板:`/im/chain/single` 优先;已绑定对话时只接受同一对话的请求。
    fn record_history(&mut self, request_id: &str, url: &str, headers: Value, post_data: String) {
        let ids = extract_conversation_ids(&post_data);
        if ids.len() != 1 {
            tracing::debug!(
                url,
                count = ids.len(),
                "历史请求 postData 里 conversation_id 不唯一,忽略"
            );
            return;
        }
        // reader 自家重放原样复用模板 postData(含同一个 `sequence_id`);真机上它的 initiator 栈被
        // slardar 包装的 fetch 挡住,`is_reader_initiated` 认不出,这里按 postData 完全相同识别,
        // 不再把重放响应体当成新历史再 analyze 一遍(页面自己的请求每次 `sequence_id` 都不同)。
        if self
            .history
            .as_ref()
            .is_some_and(|h| h.post_data == post_data)
        {
            return;
        }
        let conversation_id = ids.into_iter().next();
        if let Some(bound) = &self.bound_conversation_id {
            if conversation_id.as_deref() != Some(bound.as_str()) {
                tracing::debug!(url, "历史请求属于其它对话,忽略");
                return;
            }
        }
        let replace = match &self.history {
            None => true,
            Some(h) => is_chain_single(url) || !is_chain_single(&h.url),
        };
        if replace {
            tracing::info!(conversation_id = ?conversation_id, "记录页面的历史查询模板");
            self.history = Some(HistoryRequest {
                url: url.to_string(),
                post_data,
                headers,
                conversation_id,
            });
        }
        if !request_id.is_empty() {
            self.history_pending.insert(request_id.to_string());
        }
    }

    /// 生成响应 / SSE 的旁路:只在文本含本任务身份标记时才采信其中的 `image_ori_raw`。
    async fn maybe_extract_direct(&mut self, text: &str, source: &str) {
        if self.opts.task_type != TaskType::Image {
            return;
        }
        let markers = self.identity_markers();
        if markers.is_empty() || !markers.iter().any(|m| text.contains(m.as_str())) {
            return;
        }
        let urls = extract_image_originals(text);
        if !urls.is_empty() {
            self.add_originals(urls, source, WatermarkSource::ConversationOriginal)
                .await;
        }
    }

    // ---- 定时作业 ----

    fn spawn_feedback(&mut self, session: &Arc<CdpSession>, jobs: &mut JoinSet<Job>) {
        if self.feedback_inflight {
            return;
        }
        let step = feedback_step(
            self.opts.task_type,
            self.mode,
            self.auto_membership,
            self.user_confirmed,
        );
        let payload = feedback_payload(
            self.opts,
            self.mode,
            self.auto_membership,
            self.user_confirmed,
        );
        let expr = feedback_expression(&self.opts.tracking_key, step, &payload, !self.gen_ready);
        self.gen_ready = true;
        self.feedback_inflight = true;
        let s = Arc::clone(session);
        let timeout = self.timing.evaluate_timeout;
        jobs.spawn(async move { Job::Feedback(eval(s, expr, timeout).await) });

        self.spawn_dom_images(session, jobs);

        if self.opts.task_type == TaskType::Video
            && !self.relay_confirm_inflight
            && self.bound_conversation_id.is_some()
            && self.originals.is_empty()
        {
            let expr = relay_confirm_expression(self.bound_conversation_id.as_deref());
            self.relay_confirm_inflight = true;
            let s = Arc::clone(session);
            jobs.spawn(async move { Job::RelayConfirm(eval(s, expr, timeout).await) });
        }
    }

    /// 豆包新版页面有时只把成图渲染进 DOM,不再触发旧版历史请求。
    /// 第一次只记录页面已有图片,之后只接收新出现的高清 CDN 原图;
    /// 不再要求先有 `submission_id`,修复豆包已出图但提交编号未回传时整队卡住。
    fn spawn_dom_images(&mut self, session: &Arc<CdpSession>, jobs: &mut JoinSet<Job>) {
        if self.opts.task_type != TaskType::Image
            || !self.originals.is_empty()
            || self.dom_images_inflight
        {
            return;
        }
        let expr = r#"(()=>Array.from(document.images).map(i=>({url:i.src||i.currentSrc||'',width:i.naturalWidth||0,height:i.naturalHeight||0})).filter(i=>/^https:\/\/(?:[^./]+\.)*(?:byteimg\.com|volces\.com)\//i.test(i.url)&&i.width>=500&&i.height>=500))()"#;
        self.dom_images_inflight = true;
        let s = Arc::clone(session);
        let timeout = self.timing.evaluate_timeout;
        jobs.spawn(async move { Job::DomImages(eval(s, expr.to_string(), timeout).await) });
    }

    /// 图像代回复(`accept_reply_expression`)。成功与否由 `Job::Accept` 按返回的 `sent` 判定,
    /// 这里不预先标记 `accepted_confirmation`。
    fn spawn_accept(
        &mut self,
        session: &Arc<CdpSession>,
        jobs: &mut JoinSet<Job>,
        stage: &str,
        message: &str,
    ) {
        if self.accept_inflight {
            return;
        }
        let payload = feedback_payload(
            self.opts,
            self.mode,
            self.auto_membership,
            self.user_confirmed,
        );
        let expr = accept_reply_expression(&self.opts.tracking_key, &payload, CONTINUE_REPLY);
        self.accept_inflight = true;
        tracing::info!(
            stage,
            reply = CONTINUE_REPLY,
            attempt = self.accept_attempts + 1,
            "代回复豆包确认提示"
        );
        let s = Arc::clone(session);
        let timeout = self.timing.evaluate_timeout;
        let stage = stage.to_string();
        let message = message.to_string();
        jobs.spawn(async move {
            Job::Accept {
                stage,
                message,
                result: eval(s, expr, timeout).await,
            }
        });
    }

    /// 代回复已真正发出(图像 `sent:true` / 视频 `videoFollowup` 的 `sent`):记下 stage,
    /// 并**消费**本次用户确认——下一个确认 stage 重新按 `should_auto_confirm` 判断 / 再问用户。
    ///
    /// 同时发 `Stage(Sent, "已自动点继续" | "已按用户确认继续")`(Task 4.1:followup 已发出,豆包重新处理中);
    /// `last_emitted_stage` 保持为确认阶段,避免豆包尚未刷新提示时 `checking` 被重复广播(sent ↔ checking 抖动)。
    async fn mark_replied(&mut self, stage: &str, confirmation_message: &str) {
        let message = if self.user_confirmed {
            USER_CONTINUE_MESSAGE
        } else {
            AUTO_CONTINUE_MESSAGE
        };
        // 同一确认 stage 重复报 `sent:true`(videoFollowup 每轮都回)只广播一次
        let confirmation = (stage.to_string(), confirmation_message.to_string());
        if self.accepted_confirmation.as_ref() != Some(&confirmation) {
            tracing::info!(stage, message, "确认提示已代回复");
            self.emit(CaptureEvent::Stage(
                TaskStage::Sent,
                Some(message.to_string()),
            ))
            .await;
        }
        if self
            .history_confirm
            .as_ref()
            .is_some_and(|f| f.stage == stage)
        {
            self.history_confirm = None;
        }
        self.accepted_confirmation = Some(confirmation);
        self.accept_attempts = 0;
        self.user_confirmed = false;
        self.needs_confirm_emitted = false;
    }

    /// 降级链(analysis/05):auto 模式在 `checking` 失败 → manual;自动用会员在 `quota_checking`
    /// 失败 → 关闭自动用会员;降级了就清空代回复标记并发一次 `NeedsUserConfirm`。
    async fn degrade_after_failure(&mut self, stage: &str) -> bool {
        let mut degraded = false;
        if self.mode == ConfirmationMode::Auto && stage == STAGE_CHECKING {
            self.mode = ConfirmationMode::Manual;
            degraded = true;
            tracing::warn!("自动确认失败,降级为手动确认");
        }
        if self.auto_membership && stage == STAGE_QUOTA_CHECKING {
            self.auto_membership = false;
            degraded = true;
            tracing::warn!("自动使用会员配额失败,改为手动确认");
        }
        if degraded {
            self.accepted_confirmation = None;
            self.accept_attempts = 0;
            self.user_confirmed = false;
            self.needs_confirm_emitted = true;
            self.emit(CaptureEvent::NeedsUserConfirm).await;
        }
        degraded
    }

    /// 历史 feedback 消息(reader 形状,无 `stage`)送 `acceptFeedback` 分类;同一条消息的**同一份文本**
    /// 只送一次(bot 回复在流式输出期间被读到时只有半句,分类为 `unknown`;文本长全后必须重新分类,
    /// UC-20 真机:"视频生成参数确认"因此 12 分钟没被识别),只在已绑定 submission 后(PS 同样要求 `boundSubmissionId`)。
    fn spawn_classify(
        &mut self,
        session: &Arc<CdpSession>,
        jobs: &mut JoinSet<Job>,
        message_id: String,
        reply: &Value,
    ) {
        let Some(bound) = self.bound_submission_id.clone() else {
            return;
        };
        let key = classify_key(&message_id, reply);
        if self.classify_inflight || self.classified_feedback.contains(&key) {
            return;
        }
        let payload = accept_feedback_payload(self.opts, &bound, reply);
        let expr = accept_feedback_expression(&self.opts.tracking_key, &payload, !self.gen_ready);
        self.gen_ready = true;
        self.classify_inflight = true;
        self.classified_feedback.insert(key.clone());
        let s = Arc::clone(session);
        let timeout = self.timing.evaluate_timeout;
        jobs.spawn(async move {
            Job::Classify {
                message_id,
                key,
                result: eval(s, expr, timeout).await,
            }
        });
    }

    /// 图像:历史分类出的确认 stage 尚未代回复,且 auto 允许或用户已确认 → 代回复(feedback 节拍上调用;
    /// `terminalFeedback` 对会员额度确认只回 `unknown`,不会自己再触发)。
    fn refeed_history_confirm(&mut self, session: &Arc<CdpSession>, jobs: &mut JoinSet<Job>) {
        if self.opts.task_type != TaskType::Image || self.accept_inflight {
            return;
        }
        let Some(state) = self.history_confirm.clone() else {
            return;
        };
        let message = state.message.as_deref().unwrap_or("");
        if self.accepted_confirmation.as_ref() == Some(&(state.stage.clone(), message.to_string()))
        {
            return;
        }
        let ts = map_feedback_stage(&state.stage).unwrap_or(TaskStage::Checking);
        if self.user_confirmed || should_auto_confirm(&ts, self.mode, self.auto_membership) {
            self.spawn_accept(session, jobs, &state.stage, message);
        }
    }

    /// 视频路径 2:feedback 进入 `result`、还没有原片且没点到过"下载"时,每 `download_click_interval`
    /// 调一次 `requestVideoDownload`(feedback 节拍上调用)。
    fn maybe_request_download(&mut self, session: &Arc<CdpSession>, jobs: &mut JoinSet<Job>) {
        if self.opts.task_type != TaskType::Video
            || self.download_clicked
            || self.download_inflight
            || !self.originals.is_empty()
            || self.last_stage_js.as_deref() != Some(STAGE_RESULT)
        {
            return;
        }
        let now = Instant::now();
        if self.next_download_at.is_some_and(|t| now < t) {
            return;
        }
        let payload = feedback_payload(
            self.opts,
            self.mode,
            self.auto_membership,
            self.user_confirmed,
        );
        let expr =
            request_video_download_expression(&self.opts.tracking_key, &payload, !self.gen_ready);
        self.gen_ready = true;
        self.download_inflight = true;
        self.next_download_at = Some(now + self.timing.download_click_interval);
        let s = Arc::clone(session);
        let timeout = self.timing.evaluate_timeout;
        jobs.spawn(async move { Job::DownloadClick(eval(s, expr, timeout).await) });
    }

    /// 视频路径 3:每个 vid 10s 内不重复;API 不在白名单直接跳过(记一次 warn)。
    fn spawn_fallbacks(
        &mut self,
        session: &Arc<CdpSession>,
        jobs: &mut JoinSet<Job>,
        items: &[FallbackVideo],
    ) {
        let now = Instant::now();
        for item in items {
            if self.fallback_inflight.contains(&item.vid)
                || self
                    .fallback_attempts
                    .get(&item.vid)
                    .is_some_and(|t| now.duration_since(*t) < self.timing.fallback_retry)
            {
                continue;
            }
            self.fallback_attempts.insert(item.vid.clone(), now);
            if !is_fallback_api_url(&item.api) {
                tracing::warn!(vid = %item.vid, "fallback API 地址不在白名单(https://vas-*.snssdk.com),跳过");
                continue;
            }
            self.fallback_inflight.insert(item.vid.clone());
            tracing::info!(vid = %item.vid, "请求视频 fallback API 并本地解密");
            tracing::debug!(vid = %item.vid, api = %item.api, "视频 fallback API 地址");
            let s = Arc::clone(session);
            let fetcher = Arc::clone(&self.fetcher);
            let item = item.clone();
            let fetch_timeout = self.timing.fallback_timeout;
            let eval_timeout = self.timing.evaluate_timeout;
            jobs.spawn(async move {
                let vid = item.vid.clone();
                let result = run_fallback(s, fetcher, item, fetch_timeout, eval_timeout).await;
                Job::Fallback { vid, result }
            });
        }
    }

    fn reader_prefix(&mut self) -> String {
        if self.reader_ready {
            String::new()
        } else {
            self.reader_ready = true;
            format!("{}\n", reader_bootstrap())
        }
    }

    fn current_options(&self) -> Value {
        read_options(
            self.opts.task_type,
            &self.opts.prompt,
            &self.opts.identity,
            self.opts.send_since,
            self.bound_conversation_id.as_deref(),
            self.bound_submission_id.as_deref(),
            self.bound_created_at.as_ref(),
            &self.followup_ids,
        )
    }

    /// 绑定前记下的历史模板可能属于侧栏里别的对话(那时还不知道该认哪个):绑定后丢掉,
    /// 否则会一直重放别的对话的历史,直到超时。
    fn drop_stale_history(&mut self) {
        if let (Some(bound), Some(h)) = (&self.bound_conversation_id, &self.history) {
            if h.conversation_id.as_deref() != Some(bound.as_str()) {
                tracing::info!(
                    bound = %bound,
                    template = ?h.conversation_id,
                    "历史查询模板属于其它对话,丢弃"
                );
                self.history = None;
            }
        }
    }

    fn spawn_history_tick(&mut self, session: &Arc<CdpSession>, jobs: &mut JoinSet<Job>) {
        if self.bound_submission_id.is_none()
            && !self.binding_inflight
            && self.sse_consumed < self.sse_backlog.len()
        {
            let text = self.sse_backlog.join("\n");
            self.sse_consumed = self.sse_backlog.len();
            let expr = format!(
                "{}{}",
                self.reader_prefix(),
                stream_binding_expression(&text, &self.opts.identity)
            );
            self.binding_inflight = true;
            let s = Arc::clone(session);
            let timeout = self.timing.evaluate_timeout;
            jobs.spawn(async move { Job::StreamBinding(eval(s, expr, timeout).await) });
        }
        if self.read_inflight {
            return;
        }
        self.drop_stale_history();
        let Some(history) = self.history.clone() else {
            tracing::trace!("尚未观察到历史查询请求,跳过重放");
            return;
        };
        let options = self.current_options();
        let expr = format!(
            "{}{}",
            self.reader_prefix(),
            read_expression(&history, &options)
        );
        self.read_inflight = true;
        let s = Arc::clone(session);
        let timeout = self.timing.evaluate_timeout;
        jobs.spawn(async move { Job::Read(eval(s, expr, timeout).await) });
    }

    fn spawn_analyze(&mut self, session: &Arc<CdpSession>, jobs: &mut JoinSet<Job>, body: String) {
        if self.analyze_inflight {
            tracing::debug!("上一条历史响应仍在分析,跳过这条");
            return;
        }
        let Some(history) = self.history.clone() else {
            return;
        };
        let options = self.current_options();
        let expr = format!(
            "{}{}",
            self.reader_prefix(),
            analyze_expression(&body, &history, &options)
        );
        self.analyze_inflight = true;
        let s = Arc::clone(session);
        let timeout = self.timing.evaluate_timeout;
        jobs.spawn(async move { Job::Analyze(eval(s, expr, timeout).await) });
    }

    fn spawn_targets(&mut self, host: &str, port: u16, jobs: &mut JoinSet<Job>) {
        if self.targets_inflight {
            return;
        }
        self.targets_inflight = true;
        let host = host.to_string();
        jobs.spawn(async move { Job::Targets(list_targets(&host, port).await) });
    }

    fn spawn_input_check(&mut self, session: &Arc<CdpSession>, jobs: &mut JoinSet<Job>) {
        if self.input_inflight {
            return;
        }
        self.input_inflight = true;
        let s = Arc::clone(session);
        let timeout = self.timing.evaluate_timeout;
        jobs.spawn(
            async move { Job::InputCheck(eval(s, input_check_expression(), timeout).await) },
        );
    }

    // ---- 作业结果 ----

    async fn handle_job(
        &mut self,
        session: &Arc<CdpSession>,
        job: Job,
        jobs: &mut JoinSet<Job>,
        port: u16,
    ) -> RelayResult<()> {
        match job {
            Job::Feedback(result) => {
                self.feedback_inflight = false;
                match result {
                    Ok(v) => {
                        if v.get("missing").and_then(Value::as_bool) == Some(true) {
                            tracing::warn!("生成脚本全局对象不存在,下次轮询重新注入");
                            self.gen_ready = false;
                            return Ok(());
                        }
                        let state = parse_feedback(&v);
                        self.process_feedback(session, state, "poll", jobs).await
                    }
                    Err(e) => {
                        let stage = self.last_stage_js.clone().unwrap_or_default();
                        self.on_script_failure("feedback", &stage, &e).await;
                        Ok(())
                    }
                }
            }
            Job::DomImages(result) => {
                self.dom_images_inflight = false;
                let urls: Vec<String> = match result {
                    Ok(v) => v
                        .as_array()
                        .into_iter()
                        .flatten()
                        .filter_map(|item| item.get("url").and_then(Value::as_str))
                        .map(html_decode)
                        .filter(|url| is_image_cdn_url(url))
                        .collect(),
                    Err(e) => {
                        tracing::debug!(error = ?e, "读取豆包页面成图失败");
                        Vec::new()
                    }
                };
                let new_urls = collect_new_dom_images(
                    &mut self.dom_baseline_ready,
                    &mut self.dom_known_images,
                    urls,
                );
                if !new_urls.is_empty() {
                    tracing::info!(count = new_urls.len(), "从豆包页面 DOM 兜底读到本次新成图");
                    self.add_originals(
                        new_urls,
                        "dom",
                        WatermarkSource::ConversationOriginal,
                    )
                    .await;
                }
                Ok(())
            }
            Job::RelayConfirm(result) => {
                self.relay_confirm_inflight = false;
                let text = match &result {
                    Ok(v) => v
                        .get("text")
                        .and_then(Value::as_str)
                        .unwrap_or("")
                        .trim()
                        .to_string(),
                    Err(e) => {
                        tracing::debug!(error = ?e, "读取最后一条机器人回复失败");
                        String::new()
                    }
                };
                if text.is_empty() || !is_relay_confirmation_text(&text) {
                    return Ok(());
                }
                let state = FeedbackState {
                    stage: STAGE_CHECKING.to_string(),
                    message: Some(text),
                    relay_detected: true,
                    ..FeedbackState::default()
                };
                self.process_feedback(session, state, "relay_dom", jobs)
                    .await
            }
            Job::Accept {
                stage,
                message,
                result,
            } => {
                self.accept_inflight = false;
                match result {
                    Ok(v) => {
                        if v.get("sent").and_then(Value::as_bool) == Some(true) {
                            tracing::info!(stage = %stage, result = %v, "代回复已发出");
                            self.mark_replied(&stage, &message).await;
                            return Ok(());
                        }
                        if v.get("missing").and_then(Value::as_bool) == Some(true) {
                            tracing::warn!("生成脚本全局对象不存在,代回复推迟到重新注入后");
                            self.gen_ready = false;
                            return Ok(());
                        }
                        self.accept_attempts += 1;
                        tracing::warn!(
                            stage = %stage,
                            attempts = self.accept_attempts,
                            result = %v,
                            "代回复未发出(sent!=true)"
                        );
                        if self.accept_attempts >= ACCEPT_MAX_ATTEMPTS
                            && !self.degrade_after_failure(&stage).await
                        {
                            // manual 且用户已确认的情况下仍发不出:让用户再确认一次(重置计数)
                            self.accept_attempts = 0;
                            self.user_confirmed = false;
                            self.needs_confirm_emitted = true;
                            self.emit(CaptureEvent::NeedsUserConfirm).await;
                        }
                        Ok(())
                    }
                    Err(e) => {
                        self.on_script_failure("acceptReply", &stage, &e).await;
                        Ok(())
                    }
                }
            }
            Job::Classify {
                message_id,
                key,
                result,
            } => {
                self.classify_inflight = false;
                match result {
                    Ok(v) => {
                        if v.get("missing").and_then(Value::as_bool) == Some(true) {
                            tracing::warn!("生成脚本全局对象不存在,历史 feedback 稍后重新分类");
                            self.gen_ready = false;
                            self.classified_feedback.remove(&key);
                            return Ok(());
                        }
                        let state = parse_feedback(&v);
                        tracing::debug!(message_id, stage = %state.stage, "历史 feedback 已分类");
                        if is_confirmation_stage(&state.stage) {
                            self.history_confirm = Some(state.clone());
                        } else if !state.is_unknown() {
                            self.history_confirm = None;
                        }
                        self.process_feedback(session, state, "history", jobs).await
                    }
                    Err(e) => {
                        tracing::warn!(message_id, error = %e, "历史 feedback 分类失败");
                        if let EvalError::Cdp(err) = &e {
                            // 断线 / 超时:允许下次重放时重试这条消息
                            self.classified_feedback.remove(&key);
                            if matches!(err, CdpError::Closed) {
                                self.reset_page_scripts();
                            }
                        }
                        Ok(())
                    }
                }
            }
            Job::Read(result) => {
                self.read_inflight = false;
                match result {
                    Ok(v) => self.process_analysis(session, &v, "read", jobs).await,
                    Err(e) => {
                        self.on_reader_failure("read", &e);
                        Ok(())
                    }
                }
            }
            Job::Analyze(result) => {
                self.analyze_inflight = false;
                match result {
                    Ok(v) => self.process_analysis(session, &v, "analyze", jobs).await,
                    Err(e) => {
                        self.on_reader_failure("analyze", &e);
                        Ok(())
                    }
                }
            }
            Job::StreamBinding(result) => {
                self.binding_inflight = false;
                match result {
                    Ok(v) => {
                        if v.is_object() {
                            self.apply_binding(&v).await;
                        }
                        Ok(())
                    }
                    Err(e) => {
                        self.on_reader_failure("streamSubmissionBinding", &e);
                        Ok(())
                    }
                }
            }
            Job::Body {
                kind,
                request_id,
                body,
            } => {
                let Some(body) = body else {
                    return Ok(());
                };
                if body.len() > MAX_BODY_BYTES {
                    tracing::warn!(request_id, len = body.len(), "响应体过大,忽略");
                    return Ok(());
                }
                match kind {
                    BodyKind::History => self.spawn_analyze(session, jobs, body),
                    BodyKind::Generation => {
                        if self.bound_submission_id.is_none()
                            && self.sse_backlog.len() < SSE_BACKLOG_MAX_ENTRIES
                            && self.sse_bytes + body.len() <= SSE_BACKLOG_MAX_BYTES
                        {
                            self.sse_bytes += body.len();
                            self.sse_backlog.push(body.clone());
                        }
                        self.maybe_extract_direct(&body, "generation_response")
                            .await;
                    }
                }
                Ok(())
            }
            Job::PostData {
                request_id,
                url,
                headers,
                post_data,
            } => {
                if let Some(post_data) = post_data {
                    self.record_history(&request_id, &url, headers, post_data);
                }
                Ok(())
            }
            Job::InputCheck(result) => {
                self.input_inflight = false;
                // 代回复进行中输入框里是 relay 自己填的补发文本,不算用户改动
                if let (Ok(v), false) = (result, self.accept_inflight) {
                    let text = v.get("text").and_then(Value::as_str).unwrap_or("");
                    if input_is_expected(text, &self.opts.prompt) {
                        self.input_modified = None;
                    } else {
                        let text_length = text.chars().count();
                        // 同一段改动只 warn / 发事件一次;之后每 10s 的复查降为 debug,长视频任务不刷屏
                        if self.input_modified.as_deref() != Some(text) {
                            tracing::warn!(
                                interruption = "user_modified_input",
                                text_length,
                                "豆包输入框里有用户输入的内容,继续等待结果"
                            );
                            self.input_modified = Some(text.to_string());
                            self.emit(CaptureEvent::InputModified { text_length }).await;
                        } else {
                            tracing::debug!(
                                interruption = "user_modified_input",
                                text_length,
                                "输入框仍是用户改动后的内容"
                            );
                        }
                    }
                }
                Ok(())
            }
            Job::DownloadClick(result) => {
                self.download_inflight = false;
                match result {
                    Ok(v) if v.get("missing").and_then(Value::as_bool) == Some(true) => {
                        tracing::warn!("生成脚本全局对象不存在,下载按钮稍后重试");
                        self.gen_ready = false;
                    }
                    Ok(v) if v.get("clicked").and_then(Value::as_bool) == Some(true) => {
                        tracing::info!("已点击豆包视频下载按钮,等待无水印原片请求");
                        self.download_clicked = true;
                    }
                    Ok(v) => {
                        tracing::debug!(result = %v, "下载按钮未点到,稍后重试");
                    }
                    Err(e) => {
                        tracing::debug!(error = %e, "requestVideoDownload 求值失败");
                        if matches!(e, EvalError::Cdp(CdpError::Closed)) {
                            self.reset_page_scripts();
                        }
                    }
                }
                Ok(())
            }
            Job::Fallback { vid, result } => {
                self.fallback_inflight.remove(&vid);
                match result {
                    Ok((decoded, how)) => match fallback_original_url(&decoded) {
                        Some(url) => {
                            if !has_lr_unwatermarked(&decoded) {
                                tracing::info!(vid = %vid, "解密地址没有 lr 参数,已追加 lr=unwatermarked");
                            }
                            self.add_originals(
                                vec![url],
                                if how == "js" {
                                    "fallback_js"
                                } else {
                                    "fallback"
                                },
                                WatermarkSource::FallbackDecrypted,
                            )
                            .await;
                        }
                        None => {
                            tracing::warn!(vid = %vid, how, "解密出的地址不是视频 CDN 地址,忽略");
                        }
                    },
                    Err(reason) => {
                        tracing::warn!(vid = %vid, reason = %reason, "视频 fallback 解密失败,稍后重试");
                    }
                }
                Ok(())
            }
            Job::Targets(result) => {
                self.targets_inflight = false;
                match result {
                    Ok(targets) => {
                        let (chat, _) = select_pages(&targets, port);
                        if chat.len() > 1 {
                            tracing::warn!(
                                chat_targets = chat.len(),
                                "doubao-chat page 多于 1 个,终止任务"
                            );
                            return Err(user_interrupted_error("multiple_chat_pages")
                                .with_details(json!({
                                    "interruption": "multiple_chat_pages",
                                    "chat_targets": chat.len(),
                                })));
                        }
                        if chat.is_empty() {
                            // 页面导航走 / 豆包重启中:给重连机会,与 /json/list 失败一样计数
                            self.target_failures += 1;
                            tracing::warn!(
                                failures = self.target_failures,
                                "doubao-chat page 暂时不存在,等待其恢复"
                            );
                            if self.target_failures >= TARGET_CHECK_MAX_FAILURES {
                                return Err(RelayError::doubao_cdp_disconnected().with_details(
                                    json!({
                                        "reason": "chat_page_missing",
                                        "chat_targets": 0,
                                        "during": "capture",
                                    }),
                                ));
                            }
                            return Ok(());
                        }
                        self.target_failures = 0;
                        Ok(())
                    }
                    Err(e) => {
                        self.target_failures += 1;
                        tracing::warn!(error = %e, failures = self.target_failures, "查询 /json/list 失败");
                        if self.target_failures >= TARGET_CHECK_MAX_FAILURES {
                            return Err(RelayError::doubao_cdp_disconnected().with_details(
                                json!({ "cdp": e.to_string(), "during": "capture" }),
                            ));
                        }
                        Ok(())
                    }
                }
            }
        }
    }

    /// 脚本求值失败(feedback / 代回复):JS 异常与 `Runtime.evaluate` 超时都按 PS 的
    /// `$event.error || exceptionDetails` 走降级链(auto+checking → manual;自动用会员+quota_checking →
    /// 关自动用会员,并发 `NeedsUserConfirm`);session 关闭只重置注入标记。
    async fn on_script_failure(&mut self, step: &str, stage: &str, e: &EvalError) {
        match e {
            EvalError::Js(msg) => {
                tracing::warn!(step, stage, js_exception = %msg, "feedback 脚本异常");
                self.degrade_after_failure(stage).await;
            }
            EvalError::Cdp(CdpError::Timeout(method)) => {
                tracing::warn!(step, stage, method = %method, "feedback 求值超时");
                self.degrade_after_failure(stage).await;
            }
            EvalError::Cdp(err) => {
                tracing::warn!(step, error = %err, "feedback 求值失败");
                if matches!(err, CdpError::Closed) {
                    self.reset_page_scripts();
                }
            }
        }
    }

    fn on_reader_failure(&mut self, step: &str, e: &EvalError) {
        match e {
            EvalError::Js(msg) => {
                tracing::warn!(step, js_exception = %msg, "结果读取脚本异常");
                // 脚本可能没注入成功(页面 reload),下次重新注入。
                self.reader_ready = false;
            }
            EvalError::Cdp(err) => {
                tracing::warn!(step, error = %err, "结果读取求值失败");
                if matches!(err, CdpError::Closed) {
                    self.reset_page_scripts();
                }
            }
        }
    }

    async fn apply_binding(&mut self, binding: &Value) {
        let pick = |k: &str| -> Option<String> {
            match binding.get(k)? {
                Value::String(s) if !s.is_empty() && s != "0" => Some(s.clone()),
                Value::Number(n) => Some(n.to_string()),
                _ => None,
            }
        };
        let submission_id = pick("submissionId");
        let conversation_id = pick("conversationId").or_else(|| self.bound_conversation_id.clone());
        let (Some(submission_id), Some(conversation_id)) = (submission_id, conversation_id) else {
            return;
        };
        if self.bound_submission_id.is_none() {
            self.bound_submission_id = Some(submission_id.clone());
        }
        if self.bound_conversation_id.is_none() {
            self.bound_conversation_id = Some(conversation_id.clone());
        }
        if self.bound_created_at.is_none() {
            self.bound_created_at = binding.get("createdAt").cloned().filter(|v| !v.is_null());
        }
        if !self.bound_emitted {
            self.bound_emitted = true;
            tracing::info!(conversation_id = %conversation_id, submission_id = %submission_id, "已绑定本次提交");
            self.emit(CaptureEvent::Bound {
                conversation_id,
                submission_id,
            })
            .await;
        }
    }

    /// `read` / `analyze` 的返回。
    async fn process_analysis(
        &mut self,
        session: &Arc<CdpSession>,
        v: &Value,
        source: &str,
        jobs: &mut JoinSet<Job>,
    ) -> RelayResult<()> {
        let Some(obj) = v.as_object() else {
            return Ok(());
        };
        if obj.get("ok").and_then(Value::as_bool) == Some(false) {
            let reason = obj.get("reason").and_then(Value::as_str).unwrap_or("");
            let status = obj.get("status").and_then(Value::as_i64);
            tracing::debug!(source, reason, status, "历史读取未成功");
            return Ok(());
        }
        if let Some(reason) = obj.get("reason").and_then(Value::as_str) {
            tracing::trace!(source, reason, "历史读取无匹配");
        }
        if let Some(binding) = obj.get("binding").filter(|b| b.is_object()) {
            self.apply_binding(binding).await;
        }
        if let Some(ids) = obj.get("followupSubmissionIds").and_then(Value::as_array) {
            let ids: Vec<String> = ids
                .iter()
                .filter_map(|x| match x {
                    Value::String(s) => Some(s.clone()),
                    Value::Number(n) => Some(n.to_string()),
                    _ => None,
                })
                .collect();
            if !ids.is_empty() {
                self.followup_ids = ids;
            }
        }
        let has_images = obj
            .get("images")
            .and_then(Value::as_array)
            .is_some_and(|a| !a.is_empty());
        if let Some(feedback) = obj.get("feedback").filter(|f| f.is_object()) {
            let has_stage = feedback
                .get("stage")
                .and_then(Value::as_str)
                .is_some_and(|s| !s.is_empty());
            if has_stage {
                // 已带 stage 的形状(reader 目前不会这样返回,兼容保留)
                let state = parse_feedback(feedback);
                self.process_feedback(session, state, source, jobs).await?;
            } else if !has_images {
                // reader 真实形状 `{message, messageId, submissionId, createdAt, scope}`:
                // 交给 acceptFeedback 分类(PS doubao-capture.ps1:341-345)
                let message_id = match feedback.get("messageId") {
                    Some(Value::String(s)) if !s.is_empty() => Some(s.clone()),
                    Some(Value::Number(n)) => Some(n.to_string()),
                    _ => None,
                };
                if let Some(message_id) = message_id {
                    self.spawn_classify(session, jobs, message_id, feedback);
                }
            }
        }
        if self.opts.task_type == TaskType::Video {
            self.process_video_analysis(session, obj, source, jobs)
                .await;
        }
        if self.opts.task_type == TaskType::Image {
            let images: Vec<String> = obj
                .get("images")
                .and_then(Value::as_array)
                .map(|a| {
                    a.iter()
                        .filter_map(Value::as_str)
                        .map(html_decode)
                        .filter(|u| is_image_cdn_url(u))
                        .collect()
                })
                .unwrap_or_default();
            if !images.is_empty() {
                if self.bound_submission_id.is_some() {
                    self.add_originals(images, source, WatermarkSource::ConversationOriginal)
                        .await;
                } else {
                    tracing::warn!(count = images.len(), "读到图片但尚未绑定本次提交,不采信");
                }
            }
        }
        Ok(())
    }

    /// 视频的 `read` / `analyze` 返回(reader `resolveVideos`,反混淆实读):
    /// - `{ok:true, images:[], fallbacks:[{vid, api}]}`:消息里的视频带 `fallback_api` → 路径 3;
    /// - `{ok:true, images:[{key, messageId, vid, url, unwatermarked:true, …}]}`:reader 已用
    ///   `/creativity/resource/get_without_watermark` 换到 `lr=unwatermarked` 原片 → `lr_unwatermarked`;
    /// - `{ok:false, reason:"video_id_missing"|"unwatermarked_*"}`:继续等。
    ///
    /// 与图像一样只在已绑定本次 submission 后采信(reader 未绑定时会收下整个对话 `since` 之后的视频)。
    async fn process_video_analysis(
        &mut self,
        session: &Arc<CdpSession>,
        obj: &serde_json::Map<String, Value>,
        source: &str,
        jobs: &mut JoinSet<Job>,
    ) {
        let bound = self.bound_submission_id.is_some();
        if let Some(fallbacks) = obj.get("fallbacks").and_then(Value::as_array) {
            let items: Vec<FallbackVideo> = fallbacks
                .iter()
                .filter_map(|f| {
                    Some(FallbackVideo {
                        vid: f.get("vid")?.as_str()?.to_string(),
                        api: f.get("api")?.as_str()?.to_string(),
                    })
                })
                .filter(|f| !f.vid.is_empty())
                .collect();
            if !items.is_empty() {
                self.emit(CaptureEvent::Fallbacks(items.clone())).await;
                if bound {
                    self.spawn_fallbacks(session, jobs, &items);
                } else {
                    tracing::warn!(
                        count = items.len(),
                        "读到视频 fallback 但尚未绑定本次提交,不采信"
                    );
                }
            }
        }
        let urls: Vec<String> = obj
            .get("images")
            .and_then(Value::as_array)
            .map(|a| {
                a.iter()
                    .filter(|i| i.get("unwatermarked").and_then(Value::as_bool) == Some(true))
                    .filter_map(|i| i.get("url").and_then(Value::as_str))
                    .filter_map(unwatermarked_video_url)
                    .collect()
            })
            .unwrap_or_default();
        if urls.is_empty() {
            return;
        }
        if bound {
            self.add_originals(urls, source, WatermarkSource::LrUnwatermarked)
                .await;
        } else {
            tracing::warn!(
                count = urls.len(),
                "读到无水印视频但尚未绑定本次提交,不采信"
            );
        }
    }

    async fn process_feedback(
        &mut self,
        session: &Arc<CdpSession>,
        state: FeedbackState,
        source: &str,
        jobs: &mut JoinSet<Job>,
    ) -> RelayResult<()> {
        let stage = state.stage.clone();
        let message = state.message.clone().unwrap_or_default();

        // "正在渲染"假开始:30s 内不进入 generating/result → not_executed。
        if stage == STAGE_RENDERING_UNVERIFIED
            && self.originals.is_empty()
            && !matches!(
                self.last_stage_js.as_deref(),
                Some(STAGE_GENERATING) | Some(STAGE_RESULT)
            )
        {
            let since = *self.unverified_since.get_or_insert_with(Instant::now);
            if since.elapsed() >= self.timing.rendering_unverified_limit {
                self.emit(CaptureEvent::Refused {
                    kind: "not_executed".into(),
                    message: message.clone(),
                })
                .await;
                return Err(not_executed_error(&message));
            }
        } else if !state.is_unknown() {
            self.unverified_since = None;
        }

        if state.is_unknown() {
            return Ok(());
        }
        if self.last_feedback.as_ref() != Some(&state) {
            tracing::info!(source, stage = %stage, message_length = message.chars().count(), sent = state.sent, "feedback 阶段");
            self.emit(CaptureEvent::Feedback(state.clone())).await;
            self.last_feedback = Some(state.clone());
        }
        let mapped = map_feedback_stage(&stage);
        if let Some(ts) = mapped {
            if self.last_emitted_stage != Some(ts) {
                self.last_emitted_stage = Some(ts);
                self.emit(CaptureEvent::Stage(ts, state.message.clone()))
                    .await;
            }
        }
        if is_refusal_stage(&stage) {
            self.emit(CaptureEvent::Refused {
                kind: stage.clone(),
                message: message.clone(),
            })
            .await;
            return Err(refusal_relay_error(&stage, &message));
        }
        if is_confirmation_stage(&stage) {
            let ts = mapped.unwrap_or(TaskStage::Checking);
            let auto = should_auto_confirm(&ts, self.mode, self.auto_membership);
            let already_replied =
                self.accepted_confirmation.as_ref() == Some(&(stage.clone(), message.clone()));
            if state.sent {
                // videoFollowup 已自行代回复:本次确认已消费
                self.mark_replied(&stage, &message).await;
            } else if already_replied {
                // 已代回复过这个 stage,等豆包处理;不再问用户
            } else if auto || self.user_confirmed {
                match self.opts.task_type {
                    TaskType::Image => self.spawn_accept(session, jobs, &stage, &message),
                    TaskType::Video if state.relay_detected => {
                        self.spawn_accept(session, jobs, &stage, &message)
                    }
                    TaskType::Video => {
                        // videoFollowup 自行发送;它报 manualRequired(页面上没有可自动确认的提示)
                        // 且本来是 auto → 降级 manual 并问用户
                        if state.manual_required
                            && !self.user_confirmed
                            && !self.degrade_after_failure(&stage).await
                            && !self.needs_confirm_emitted
                        {
                            self.needs_confirm_emitted = true;
                            self.emit(CaptureEvent::NeedsUserConfirm).await;
                        }
                    }
                }
            } else if !self.needs_confirm_emitted {
                self.needs_confirm_emitted = true;
                tracing::info!(stage = %stage, "豆包等待确认,需要用户介入");
                self.emit(CaptureEvent::NeedsUserConfirm).await;
            }
        } else {
            self.needs_confirm_emitted = false;
        }
        self.last_stage_js = Some(stage);
        Ok(())
    }
}

// ------------------------------------------------------------------
// 入口
// ------------------------------------------------------------------

/// 捕获(默认节奏)。见 [`capture_with`]。
pub async fn capture(
    conn: &DoubaoConnection,
    opts: CaptureOptions,
    cancel: &CancellationToken,
    confirm_rx: &mut watch::Receiver<bool>,
    events: mpsc::Sender<CaptureEvent>,
) -> RelayResult<CaptureOutcome> {
    capture_with(
        conn,
        opts,
        cancel,
        confirm_rx,
        events,
        CaptureTiming::default(),
    )
    .await
}

/// 捕获循环(节奏可配)。`confirm_rx` 变为 true 视为用户已确认(manual 模式代回复一次)。
pub async fn capture_with(
    conn: &DoubaoConnection,
    opts: CaptureOptions,
    cancel: &CancellationToken,
    confirm_rx: &mut watch::Receiver<bool>,
    events: mpsc::Sender<CaptureEvent>,
    timing: CaptureTiming,
) -> RelayResult<CaptureOutcome> {
    capture_with_fetcher(
        conn,
        opts,
        cancel,
        confirm_rx,
        events,
        timing,
        Arc::new(HttpFallbackFetcher::new()),
    )
    .await
}

/// [`capture_with`] + 可替换的视频 fallback API 请求实现(测试注入)。
pub async fn capture_with_fetcher(
    conn: &DoubaoConnection,
    opts: CaptureOptions,
    cancel: &CancellationToken,
    confirm_rx: &mut watch::Receiver<bool>,
    events: mpsc::Sender<CaptureEvent>,
    timing: CaptureTiming,
    fetcher: Arc<dyn FallbackFetcher>,
) -> RelayResult<CaptureOutcome> {
    let started = Instant::now();
    let deadline_at = started + opts.deadline;
    tracing::info!(
        tracking_key = %opts.tracking_key,
        task_type = opts.task_type.as_str(),
        local_message_id = opts.identity.local_message_id.as_deref().unwrap_or(""),
        message_id = opts.identity.message_id.as_deref().unwrap_or(""),
        conversation_id = opts.identity.conversation_id.as_deref().unwrap_or(""),
        send_since = opts.send_since,
        deadline_ms = opts.deadline.as_millis() as u64,
        "开始捕获生成结果"
    );
    if cancel.is_cancelled() {
        return Err(cancelled("capture"));
    }
    let host = conn.host().to_string();
    let port = conn.port();

    let mut session = conn.chat_session().await?;
    let mut chat_rx = Some(session.subscribe());
    enable_network(&session).await?;
    let (mut background, mut bg_rx) = attach_background(conn).await;

    let mut st = State::new(&opts, timing, &events, fetcher);
    let mut jobs: JoinSet<Job> = JoinSet::new();
    let now = Instant::now();
    let mut next_feedback = now;
    let mut next_history = now;
    let mut next_targets = now + timing.target_check;
    let mut next_input = now + timing.input_check;

    loop {
        if let Some(outcome) = st.settled_outcome(Instant::now()) {
            if st.last_emitted_stage != Some(TaskStage::Result) {
                st.emit(CaptureEvent::Stage(TaskStage::Result, None)).await;
            }
            tracing::info!(
                tracking_key = %opts.tracking_key,
                originals = outcome.originals.len(),
                elapsed_ms = started.elapsed().as_millis() as u64,
                "捕获完成"
            );
            jobs.abort_all();
            return Ok(outcome);
        }
        if session.is_closed() {
            tracing::warn!("chat session 已断开,尝试重连后继续捕获");
            // Task 6.1:reload / 切账号时 chat page 会短暂消失,按 `reconnect_delay` 重试
            // (最长 `restart_after`,豆包进程退出立刻放弃),重连成功后继续本次捕获。
            session = match conn.await_chat_session(cancel).await {
                Ok(s) => s,
                Err(_) if cancel.is_cancelled() => {
                    jobs.abort_all();
                    return Err(cancelled("capture"));
                }
                Err(e) => {
                    jobs.abort_all();
                    return Err(e);
                }
            };
            chat_rx = Some(session.subscribe());
            enable_network(&session).await?;
            // 豆包整体 reload 时 background 页的 session 也会关闭,一并重取
            (background, bg_rx) = attach_background(conn).await;
            st.reset_page_scripts();
            st.history_pending.clear();
        } else if background.as_ref().is_some_and(|b| b.is_closed()) {
            // 只有 background 断了(chat 仍在):心跳重连后再接上;期间不监听它
            tracing::warn!("background session 已断开,等待重连后重新监听");
            background = None;
            bg_rx = None;
        }
        let settle_at = st.settle_at();
        tokio::select! {
            _ = cancel.cancelled() => {
                jobs.abort_all();
                return Err(cancelled("capture"));
            }
            // 原图静默窗口到期:回到循环顶部判定完成
            _ = tokio::time::sleep_until(settle_at.unwrap_or(deadline_at).into()), if settle_at.is_some() => {}
            _ = confirmed(confirm_rx), if !st.user_confirmed => {
                tracing::info!("收到用户确认");
                st.user_confirmed = true;
                st.accept_attempts = 0;
                st.needs_confirm_emitted = false;
                next_feedback = Instant::now();
            }
            ev = next_event(&mut chat_rx) => {
                st.handle_event(&session, ev, &mut jobs).await;
            }
            ev = next_event(&mut bg_rx) => {
                if let Some(bg) = &background {
                    st.handle_event(bg, ev, &mut jobs).await;
                }
            }
            _ = tokio::time::sleep_until(next_feedback.into()) => {
                st.spawn_feedback(&session, &mut jobs);
                st.refeed_history_confirm(&session, &mut jobs);
                st.maybe_request_download(&session, &mut jobs);
                next_feedback = Instant::now() + timing.feedback_poll;
            }
            _ = tokio::time::sleep_until(next_history.into()) => {
                st.spawn_history_tick(&session, &mut jobs);
                next_history = Instant::now() + timing.history_poll;
            }
            _ = tokio::time::sleep_until(next_targets.into()) => {
                st.spawn_targets(&host, port, &mut jobs);
                next_targets = Instant::now() + timing.target_check;
                if background.is_none() {
                    // background 曾断开:心跳可能已重建,顺便试着接回来
                    let (bg, rx) = attach_background(conn).await;
                    if bg.is_some() {
                        tracing::info!("background session 已重新监听");
                        background = bg;
                        bg_rx = rx;
                    }
                }
            }
            _ = tokio::time::sleep_until(next_input.into()) => {
                st.spawn_input_check(&session, &mut jobs);
                next_input = Instant::now() + timing.input_check;
            }
            _ = tokio::time::sleep_until(deadline_at.into()) => {
                jobs.abort_all();
                if let Some(outcome) = st.outcome() {
                    // 静默窗口内撞上总超时:已拿到的原图照样交付
                    tracing::info!(tracking_key = %opts.tracking_key, originals = outcome.originals.len(), "捕获在静默窗口内到达总超时,按已捕获结果完成");
                    return Ok(outcome);
                }
                tracing::warn!(tracking_key = %opts.tracking_key, last_stage = st.last_stage_js.as_deref().unwrap_or(""), "捕获超时");
                return Err(task_timeout_error(opts.deadline, st.last_stage_js.as_deref()));
            }
            job = next_job(&mut jobs) => {
                if let Err(e) = st.handle_job(&session, job, &mut jobs, port).await {
                    jobs.abort_all();
                    return Err(e);
                }
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::HashMap;
    use std::path::{Path, PathBuf};
    use std::sync::atomic::{AtomicBool, Ordering};
    use std::sync::Mutex as StdMutex;

    use async_trait::async_trait;
    use futures_util::{SinkExt, StreamExt};
    use tokio::io::{AsyncReadExt, AsyncWriteExt};
    use tokio::net::TcpListener;
    use tokio_tungstenite::tungstenite::Message;

    use crate::doubao::connection::ConnTiming;
    use crate::platform::{Platform, PlatformError};

    // ---- 纯函数 ----

    #[test]
    fn extract_image_originals_nested_json() {
        let body = r#"{"data":{"messages":[{"content":"{\"blocks\":[{\"image\":{\"image_ori_raw\":{\"url\":\"https://p3-flow.byteimg.com/a.png?x=1&amp;y=2\"},\"image_raw\":{\"url\":\"https://p3-flow.byteimg.com/wm.png\"}}}]}"},{"image_ori_raw":{"url":"https://img.volces.com/b"}},{"image_ori_raw":{"url":"https://p3-flow.byteimg.com/a.png?x=1&y=2"}},{"image_ori_raw":{"url":"https://evil.com/c"}},{"image_ori_raw":{"url":"http://p3.byteimg.com/d"}}]}}"#;
        let urls = extract_image_originals(body);
        assert_eq!(
            urls,
            vec![
                "https://p3-flow.byteimg.com/a.png?x=1&y=2".to_string(),
                "https://img.volces.com/b".to_string(),
            ]
        );
    }

    #[test]
    fn extract_image_originals_sse_lines() {
        let body = "event: message\r\ndata: {\"image_ori_raw\":{\"url\":\"https://a.byteimg.com/1\"}}\r\n\r\ndata: not json\n\ndata: [{\"x\":{\"image_ori_raw\":{\"url\":\"https://a.byteimg.com/2\"}}}]\n";
        assert_eq!(
            extract_image_originals(body),
            vec![
                "https://a.byteimg.com/1".to_string(),
                "https://a.byteimg.com/2".to_string()
            ]
        );
        assert!(extract_image_originals("").is_empty());
        assert!(extract_image_originals("plain text").is_empty());
        assert!(extract_image_originals(r#"{"image_ori_raw":{"url":""}}"#).is_empty());
        assert!(
            extract_image_originals(r#"{"image_ori_raw":"https://a.byteimg.com/x"}"#).is_empty()
        );
    }

    #[test]
    fn html_decode_entities() {
        assert_eq!(
            html_decode("a&amp;b&lt;c&gt;d&quot;e&#39;f"),
            "a&b<c>d\"e'f"
        );
        assert_eq!(html_decode("&#x4e2d;&#25991;"), "中文");
        assert_eq!(html_decode("no entities"), "no entities");
        assert_eq!(html_decode("&unknown; & x"), "&unknown; & x");
        assert_eq!(html_decode("trailing &amp"), "trailing &amp");
    }

    #[test]
    fn image_cdn_url_whitelist() {
        assert!(is_image_cdn_url("https://p3-flow.byteimg.com/a"));
        assert!(is_image_cdn_url("https://byteimg.com/a"));
        assert!(is_image_cdn_url("https://x.volces.com/a"));
        assert!(!is_image_cdn_url("http://p3.byteimg.com/a"));
        assert!(!is_image_cdn_url("https://byteimg.com.evil.com/a"));
        assert!(!is_image_cdn_url("https://notbyteimg.com/a"));
        assert!(!is_image_cdn_url("nope"));
    }

    #[test]
    fn dom_image_scan_uses_first_snapshot_as_baseline_and_accepts_new_images_without_binding() {
        let mut baseline_ready = false;
        let mut known = HashSet::new();

        assert!(
            collect_new_dom_images(
                &mut baseline_ready,
                &mut known,
                vec!["https://p3-flow.byteimg.com/old".into()]
            )
            .is_empty()
        );
        assert!(baseline_ready);

        assert_eq!(
            collect_new_dom_images(
                &mut baseline_ready,
                &mut known,
                vec![
                    "https://p3-flow.byteimg.com/old".into(),
                    "https://p3-flow.byteimg.com/new".into(),
                    "https://p3-flow.byteimg.com/new".into(),
                ]
            ),
            vec!["https://p3-flow.byteimg.com/new".to_string()]
        );
    }

    #[test]
    fn dom_generation_image_accepts_high_res_cdn_paths_beyond_thumbnail_prefix() {
        assert!(is_dom_generation_image(
            "https://p11-flow-imagex-sign.byteimg.com/tos-cn-i-a9rns2rl98/generated.jpg",
            2848,
            1600,
        ));
        assert!(!is_dom_generation_image(
            "https://p11-flow-imagex-sign.byteimg.com/tos-cn-i-a9rns2rl98/rc_gen_image/thumb.png",
            216,
            384,
        ));
        assert!(!is_dom_generation_image(
            "https://example.com/generated.jpg",
            2848,
            1600,
        ));
    }

    #[test]
    fn history_request_table() {
        assert!(is_history_request(
            "https://www.doubao.com/im/chain/single",
            "POST"
        ));
        assert!(is_history_request(
            "https://www.doubao.com/im/chain/single/?aid=1",
            "post"
        ));
        assert!(is_history_request(
            "https://www.doubao.com/im/message/chain_by_block",
            "POST"
        ));
        assert!(!is_history_request(
            "https://www.doubao.com/im/chain/single",
            "GET"
        ));
        assert!(!is_history_request(
            "https://api.doubao.com/im/chain/single",
            "POST"
        ));
        assert!(!is_history_request(
            "http://www.doubao.com/im/chain/single",
            "POST"
        ));
        assert!(!is_history_request(
            "https://www.doubao.com/im/chain/single/extra",
            "POST"
        ));
        assert!(!is_history_request(
            "https://www.doubao.com/samantha/chat/completion",
            "POST"
        ));
        assert!(is_chain_single("https://www.doubao.com/im/chain/single/"));
        assert!(!is_chain_single(
            "https://www.doubao.com/im/message/chain_by_block"
        ));
    }

    #[test]
    fn reader_initiated_detection() {
        let p = json!({ "initiator": { "type": "script", "stack": { "callFrames": [
            { "functionName": "fetch" }, { "functionName": "shenlangReadBoundConversation" }
        ]}}});
        assert!(is_reader_initiated(&p));
        let p = json!({ "initiator": { "type": "script", "stack": {
            "callFrames": [{ "functionName": "x" }],
            "parent": { "callFrames": [{ "functionName": "shenlangReadBoundConversation" }] }
        }}});
        assert!(is_reader_initiated(&p));
        let p = json!({ "initiator": { "type": "script", "stack": { "callFrames": [{ "functionName": "other" }] } } });
        assert!(!is_reader_initiated(&p));
        assert!(!is_reader_initiated(
            &json!({ "initiator": { "type": "other" } })
        ));
        assert!(!is_reader_initiated(&json!({})));
    }

    #[test]
    fn conversation_ids_from_post_data() {
        assert_eq!(
            extract_conversation_ids(r#"{"conversation_id":"123456789012345678","cursor":0}"#),
            vec!["123456789012345678".to_string()]
        );
        assert_eq!(
            extract_conversation_ids(r#"{"conversation_id": 123, "x": {"conversation_id":"123"}}"#),
            vec!["123".to_string()]
        );
        assert_eq!(
            extract_conversation_ids(r#"{"conversation_id":"1","b":{"conversation_id":2}}"#),
            vec!["1".to_string(), "2".to_string()]
        );
        assert!(extract_conversation_ids(r#"{"conversation_id":"0"}"#).is_empty());
        assert!(extract_conversation_ids(r#"{"conversation_id":null}"#).is_empty());
        assert!(extract_conversation_ids("{}").is_empty());
    }

    fn identity() -> SubmissionIdentity {
        SubmissionIdentity {
            local_message_id: Some("lm-123456".into()),
            message_id: None,
            conversation_id: Some("7001".into()),
        }
    }

    fn opts() -> CaptureOptions {
        CaptureOptions {
            task_type: TaskType::Image,
            prompt: "一只蓝色小鸟".into(),
            model_name: Some("Seedream 5.0 Pro".into()),
            ratio: Some("1:1".into()),
            duration: None,
            reference_count: 0,
            identity: identity(),
            generation_requests: Vec::new(),
            send_since: 1_700_000_000_000,
            confirmation_mode: ConfirmationMode::Auto,
            auto_use_membership: true,
            tracking_key: "shenlangDoubaoTask_abc".into(),
            deadline: Duration::from_secs(300),
        }
    }

    fn history() -> HistoryRequest {
        HistoryRequest {
            url: "https://www.doubao.com/im/chain/single".into(),
            post_data: r#"{"conversation_id":"7001","cursor":0}"#.into(),
            headers: json!({ "content-type": "application/json", "x-tt": "1" }),
            conversation_id: Some("7001".into()),
        }
    }

    #[test]
    fn read_options_fields_follow_analysis_05() {
        let o = read_options(
            TaskType::Image,
            "一只蓝色小鸟",
            &identity(),
            1_700_000_000_000,
            Some("7001"),
            Some("9002"),
            Some(&json!(1700000001)),
            &["9003".to_string()],
        );
        let obj = o.as_object().unwrap();
        assert_eq!(obj["nativeShadow"], false);
        assert_eq!(obj["requireBinding"], true);
        assert_eq!(obj["requireIdentity"], true);
        assert_eq!(obj["expectedLocalMessageId"], "lm-123456");
        assert!(obj["expectedMessageId"].is_null());
        assert_eq!(obj["expectedConversationId"], "7001");
        assert_eq!(obj["boundSubmissionId"], "9002");
        assert_eq!(obj["boundSubmissionCreatedAt"], 1700000001);
        assert_eq!(obj["boundConversationId"], "7001");
        assert_eq!(obj["followupSubmissionIds"], json!(["9003"]));
        assert_eq!(obj["prompt"], "一只蓝色小鸟");
        assert_eq!(obj["since"], 1_700_000_000_000i64);
        assert_eq!(obj["type"], "image");
        assert_eq!(obj.len(), 13);
        // 未绑定对话时 requireBinding=false(否则 reader 永远返回空)
        let o = read_options(TaskType::Video, "p", &identity(), 1, None, None, None, &[]);
        assert_eq!(o["requireBinding"], false);
        assert!(o["boundSubmissionCreatedAt"].is_null());
        assert_eq!(o["type"], "video");
    }

    #[test]
    fn expressions_are_well_formed() {
        let h = history();
        let o = json!({ "since": 1 });
        let r = read_expression(&h, &o);
        assert!(r.starts_with("Promise.resolve(shenlangDoubaoResultReader.read({"));
        assert!(r.contains(r#""method":"POST""#));
        assert!(r.contains(r#""postData":"{\"conversation_id\":\"7001\",\"cursor\":0}""#));
        assert!(r.contains(r#""headers":{"content-type":"application/json","x-tt":"1"}"#));
        assert!(r.ends_with(r#", {"since":1}))"#));

        let a = analyze_expression("{\"a\":1}", &h, &o);
        assert!(
            a.starts_with(r#"Promise.resolve(shenlangDoubaoResultReader.analyze("{\"a\":1}", {"#)
        );
        assert!(a.ends_with(r#", {"since":1}))"#));

        let s = stream_binding_expression("data: {}", &identity());
        assert_eq!(
            s,
            r#"shenlangDoubaoResultReader.streamSubmissionBinding("data: {}", {"expectedLocalMessageId":"lm-123456","expectedMessageId":null})"#
        );

        let payload = feedback_payload(&opts(), ConfirmationMode::Manual, false, true);
        assert_eq!(payload["type"], "image");
        assert_eq!(payload["prompt"], "一只蓝色小鸟");
        assert_eq!(payload["confirmationMode"], "manual");
        assert_eq!(payload["autoUseMembership"], false);
        assert_eq!(payload["userConfirmed"], true);
        assert_eq!(payload["trackingKey"], "shenlangDoubaoTask_abc");
        assert_eq!(payload["parallelTracking"], true);

        let f = feedback_expression(
            "shenlangDoubaoTask_abc",
            "terminalFeedback",
            &payload,
            false,
        );
        assert!(f.starts_with("typeof shenlangDoubaoTask_abc === 'undefined' ? ({ stage: 'unknown', missing: true }) : Promise.resolve(shenlangDoubaoTask_abc.run(\"terminalFeedback\", {"));
        let f = feedback_expression("shenlangDoubaoTask_abc", "terminalFeedback", &payload, true);
        assert!(f.starts_with("if (typeof shenlangDoubaoTask_abc === 'undefined') {"));
        assert!(f.contains("typeof shenlangDoubaoTask_abc === 'undefined' ? ("));

        // 图像代回复:走脚本自己的 fill + 点发送按钮,绝不调用 acceptFeedback(那是分类钩子,不发送)
        let acc = accept_reply_expression("shenlangDoubaoTask_abc", &payload, "继续生成");
        assert!(acc.contains(ACCEPT_REPLY_MARKER));
        assert!(acc.contains("shenlangDoubaoTask_abc.run('fill', {"));
        assert!(acc.contains(r#""prompt":"继续生成""#));
        assert!(acc.contains(r#""userConfirmed":true"#));
        assert!(acc.contains("typeof shenlangDoubaoTask_abc === 'undefined'"));
        // selector 以 JS 字符串字面量(Debug 转义)嵌入
        assert!(acc.contains(&format!("const SEND = {REPLY_SEND_BUTTON_SELECTOR:?};")));
        assert!(acc.contains("chat_input_send_button"));
        assert!(acc.contains("#flow-end-msg-send"));
        assert!(acc.contains("sent: true"));
        assert!(!acc.contains("acceptFeedback"));
        assert!(!acc.contains("一只蓝色小鸟"), "回复文本必须覆盖原 prompt");
        assert!(input_check_expression().contains("contenteditable"));
    }

    #[test]
    fn error_constructors() {
        let e = task_timeout_error(Duration::from_secs(300), Some("generating"));
        assert_eq!(e.code, "generation_task_timeout");
        assert_eq!(e.details.as_ref().unwrap()["deadline_ms"], 300_000);
        assert_eq!(e.details.unwrap()["last_stage"], "generating");
        let e = user_interrupted_error("multiple_chat_pages");
        assert_eq!(e.code, "generation_user_interrupted");
        assert_eq!(e.details.unwrap()["interruption"], "multiple_chat_pages");
        for e in [
            task_timeout_error(Duration::ZERO, None),
            user_interrupted_error("x"),
        ] {
            let last = e.message.chars().last().unwrap();
            assert!(!matches!(
                last,
                '。' | '.' | '!' | '\u{FF01}' | ',' | '\u{FF0C}'
            ));
            assert_eq!(e.http_status(), 500);
        }
    }

    #[test]
    fn timing_defaults_follow_spec() {
        let t = CaptureTiming::default();
        assert_eq!(t.feedback_poll, Duration::from_millis(500));
        assert_eq!(t.history_poll, Duration::from_secs(1));
        assert_eq!(t.target_check, Duration::from_secs(2));
        assert_eq!(t.input_check, Duration::from_secs(10));
        assert_eq!(t.rendering_unverified_limit, Duration::from_secs(30));
        assert_eq!(t.evaluate_timeout, Duration::from_secs(30));
        assert_eq!(t.result_settle, Duration::from_millis(1500));
        assert_eq!(t.result_settle_max, Duration::from_secs(6));
        assert_eq!(t.download_click_interval, Duration::from_secs(5));
        assert_eq!(t.fallback_retry, Duration::from_secs(10));
        assert_eq!(t.fallback_timeout, Duration::from_secs(10));
    }

    // ---- 假平台 + mock 豆包 ----

    struct FakePlatform;

    #[async_trait]
    impl Platform for FakePlatform {
        async fn find_doubao_exe(&self) -> Result<PathBuf, PlatformError> {
            Ok(PathBuf::from(
                "/Applications/Doubao.app/Contents/MacOS/Doubao",
            ))
        }
        async fn launch_doubao_with_cdp(&self, _: &Path, _: u16) -> Result<(), PlatformError> {
            Ok(())
        }
        async fn kill_doubao(&self) -> Result<(), PlatformError> {
            Ok(())
        }
        async fn is_doubao_running(&self) -> bool {
            true
        }
        fn app_data_dir(&self) -> PathBuf {
            PathBuf::from("/tmp/doubao-relay")
        }
        fn log_dir(&self) -> PathBuf {
            PathBuf::from("/tmp/doubao-relay/logs")
        }
        fn config_path(&self) -> PathBuf {
            PathBuf::from("/tmp/doubao-relay/config.toml")
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
            true
        }
        fn doubao_version(&self) -> Option<String> {
            Some("2.32.5".into())
        }
    }

    /// 对一条 `Runtime.evaluate` 的回应;`delay` 让响应晚到(模拟表达式挂起 → `Runtime.evaluate` 超时)。
    #[derive(Default)]
    struct MockReply {
        value: Value,
        exception: Option<String>,
        delay: Option<Duration>,
    }

    fn ok(value: Value) -> MockReply {
        MockReply {
            value,
            exception: None,
            delay: None,
        }
    }

    fn throw(msg: &str) -> MockReply {
        MockReply {
            value: Value::Null,
            exception: Some(msg.into()),
            delay: None,
        }
    }

    fn hang(delay: Duration) -> MockReply {
        MockReply {
            value: json!({ "sent": true }),
            exception: None,
            delay: Some(delay),
        }
    }

    /// 把表达式归类成短名(`kind`)。
    fn classify(expr: &str) -> &'static str {
        if expr.contains("shenlangDoubaoVideoDecoder.decode(JSON.parse(") {
            "videoDecoder"
        } else if expr.contains(".run(\"requestVideoDownload\"") {
            "requestVideoDownload"
        } else if expr.contains(ACCEPT_REPLY_MARKER) {
            "acceptReply"
        } else if expr.contains(RELAY_CONFIRM_MARKER) {
            "relayConfirm"
        } else if expr.contains(".run(\"acceptFeedback\"") {
            "acceptFeedback"
        } else if expr.contains(".run(\"terminalFeedback\"") {
            "terminalFeedback"
        } else if expr.contains(".run(\"videoFollowup\"") {
            "videoFollowup"
        } else if expr.contains(".run(\"videoFeedback\"") {
            "videoFeedback"
        } else if expr.contains("shenlangDoubaoResultReader.read(") {
            "read"
        } else if expr.contains("shenlangDoubaoResultReader.analyze(") {
            "analyze"
        } else if expr.contains("shenlangDoubaoResultReader.streamSubmissionBinding(") {
            "streamSubmissionBinding"
        } else if expr.contains("contenteditable") {
            "inputCheck"
        } else {
            "other"
        }
    }

    type Handler = Arc<StdMutex<Box<dyn FnMut(&str, &str) -> MockReply + Send>>>;
    type Log = Arc<StdMutex<Vec<String>>>;

    struct MockDoubao {
        port: u16,
        targets: Arc<StdMutex<Vec<Value>>>,
        bodies: Arc<StdMutex<HashMap<String, String>>>,
        log: Log,
        inject: broadcast::Sender<Value>,
        close_tx: broadcast::Sender<()>,
    }

    impl MockDoubao {
        async fn start(handler: Handler) -> Self {
            let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
            let port = listener.local_addr().unwrap().port();
            let targets: Arc<StdMutex<Vec<Value>>> = Arc::new(StdMutex::new(Vec::new()));
            let bodies: Arc<StdMutex<HashMap<String, String>>> =
                Arc::new(StdMutex::new(HashMap::new()));
            let log: Log = Arc::new(StdMutex::new(Vec::new()));
            let (inject, _) = broadcast::channel::<Value>(64);
            let (close_tx, _) = broadcast::channel::<()>(4);
            let mock = Self {
                port,
                targets: Arc::clone(&targets),
                bodies: Arc::clone(&bodies),
                log: Arc::clone(&log),
                inject: inject.clone(),
                close_tx: close_tx.clone(),
            };
            tokio::spawn(async move {
                loop {
                    let Ok((sock, _)) = listener.accept().await else {
                        break;
                    };
                    let targets = Arc::clone(&targets);
                    let bodies = Arc::clone(&bodies);
                    let log = Arc::clone(&log);
                    let handler = Arc::clone(&handler);
                    let mut inject_rx = inject.subscribe();
                    let mut close_rx = close_tx.subscribe();
                    tokio::spawn(async move {
                        let mut peek = [0u8; 16];
                        let n = sock.peek(&mut peek).await.unwrap_or(0);
                        if String::from_utf8_lossy(&peek[..n]).starts_with("GET /json") {
                            let mut buf = vec![0u8; 4096];
                            let mut sock = sock;
                            let n = sock.read(&mut buf).await.unwrap_or(0);
                            let req = String::from_utf8_lossy(&buf[..n]).to_string();
                            let body = if req.starts_with("GET /json/version") {
                                r#"{"Browser":"Chrome/147.0.0.0","Protocol-Version":"1.3"}"#
                                    .to_string()
                            } else {
                                Value::Array(targets.lock().unwrap().clone()).to_string()
                            };
                            let resp = format!(
                                "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{}",
                                body.len(),
                                body
                            );
                            let _ = sock.write_all(resp.as_bytes()).await;
                            let _ = sock.shutdown().await;
                            return;
                        }
                        let Ok(mut ws) = tokio_tungstenite::accept_async(sock).await else {
                            return;
                        };
                        // 延迟回应走单独的队列,不阻塞这条连接上的其它命令
                        let (late_tx, mut late_rx) = mpsc::unbounded_channel::<Value>();
                        loop {
                            tokio::select! {
                                msg = ws.next() => {
                                    let Some(Ok(Message::Text(txt))) = msg else { break };
                                    let req: Value = serde_json::from_str(&txt).unwrap();
                                    let method = req["method"].as_str().unwrap_or("").to_string();
                                    let mut delay = None;
                                    let result = match method.as_str() {
                                        "Runtime.evaluate" => {
                                            let expr = req["params"]["expression"].as_str().unwrap_or("");
                                            let kind = classify(expr);
                                            log.lock().unwrap().push(format!("eval:{kind}"));
                                            let reply = (handler.lock().unwrap())(kind, expr);
                                            delay = reply.delay;
                                            match reply.exception {
                                                Some(m) => json!({ "exceptionDetails": { "text": "Uncaught",
                                                    "exception": { "description": format!("Error: {m}\n    at <anonymous>:1:1") } } }),
                                                None => json!({ "result": { "type": "object", "value": reply.value } }),
                                            }
                                        }
                                        "Network.getResponseBody" => {
                                            let id = req["params"]["requestId"].as_str().unwrap_or("").to_string();
                                            log.lock().unwrap().push(format!("Network.getResponseBody {id}"));
                                            match bodies.lock().unwrap().get(&id) {
                                                Some(b) => json!({ "body": b, "base64Encoded": false }),
                                                None => json!({ "body": "", "base64Encoded": false }),
                                            }
                                        }
                                        other => {
                                            log.lock().unwrap().push(other.to_string());
                                            json!({})
                                        }
                                    };
                                    let resp = json!({ "id": req["id"], "result": result });
                                    if let Some(d) = delay {
                                        let tx = late_tx.clone();
                                        tokio::spawn(async move {
                                            tokio::time::sleep(d).await;
                                            let _ = tx.send(resp);
                                        });
                                    } else if ws.send(Message::Text(resp.to_string().into())).await.is_err() { break; }
                                }
                                late = late_rx.recv() => {
                                    if let Some(resp) = late {
                                        if ws.send(Message::Text(resp.to_string().into())).await.is_err() { break; }
                                    }
                                }
                                ev = inject_rx.recv() => {
                                    if let Ok(ev) = ev {
                                        if ws.send(Message::Text(ev.to_string().into())).await.is_err() { break; }
                                    }
                                }
                                _ = close_rx.recv() => {
                                    let _ = ws.close(None).await;
                                    break;
                                }
                            }
                        }
                    });
                }
            });
            mock.set_chat_targets(1);
            mock
        }

        fn chat_target(&self, id: &str) -> Value {
            json!({
                "id": id, "type": "page", "url": "doubao://doubao-chat/chat",
                "webSocketDebuggerUrl": format!("ws://127.0.0.1:{}/devtools/page/{id}", self.port)
            })
        }

        fn background_target(&self, id: &str) -> Value {
            json!({
                "id": id, "type": "page", "url": "doubao://doubao-background/",
                "webSocketDebuggerUrl": format!("ws://127.0.0.1:{}/devtools/page/{id}", self.port)
            })
        }

        fn set_chat_targets(&self, n: usize) {
            let v: Vec<Value> = (0..n)
                .map(|i| self.chat_target(&format!("CHAT{i}")))
                .collect();
            *self.targets.lock().unwrap() = v;
        }

        /// 1 个 chat + 1 个 background page(真机形状)。
        fn set_chat_and_background(&self) {
            *self.targets.lock().unwrap() =
                vec![self.chat_target("CHAT0"), self.background_target("BG0")];
        }

        fn set_body(&self, request_id: &str, body: &str) {
            self.bodies
                .lock()
                .unwrap()
                .insert(request_id.into(), body.into());
        }

        /// 向所有已连接的 page session 推一条 CDP 事件。
        fn push(&self, method: &str, params: Value) {
            let _ = self
                .inject
                .send(json!({ "method": method, "params": params }));
        }

        fn log(&self) -> Vec<String> {
            self.log.lock().unwrap().clone()
        }

        fn count(&self, prefix: &str) -> usize {
            self.log().iter().filter(|l| l.starts_with(prefix)).count()
        }

        fn close_all(&self) {
            let _ = self.close_tx.send(());
        }

        /// 捕获开始(订阅事件)之后再推事件;提前推的事件 broadcast 收不到。
        fn push_later(self: &Arc<Self>, delay: Duration, method: &'static str, params: Value) {
            let m = Arc::clone(self);
            tokio::spawn(async move {
                tokio::time::sleep(delay).await;
                m.push(method, params);
            });
        }

        async fn conn(&self) -> Arc<DoubaoConnection> {
            let timing = ConnTiming {
                page_ready_timeout: Duration::from_millis(100),
                page_ready_poll: Duration::from_millis(10),
                reconnect_delay: Duration::from_millis(20),
                ..ConnTiming::default()
            };
            let conn = DoubaoConnection::new_with(
                Arc::new(FakePlatform),
                "127.0.0.1".into(),
                self.port,
                None,
                timing,
            );
            conn.connect_sessions().await.unwrap();
            conn
        }
    }

    fn handler(f: impl FnMut(&str, &str) -> MockReply + Send + 'static) -> Handler {
        Arc::new(StdMutex::new(Box::new(f)))
    }

    fn fast() -> CaptureTiming {
        CaptureTiming {
            feedback_poll: Duration::from_millis(20),
            history_poll: Duration::from_millis(30),
            target_check: Duration::from_millis(40),
            input_check: Duration::from_millis(60),
            rendering_unverified_limit: Duration::from_millis(80),
            evaluate_timeout: Duration::from_millis(500),
            command_timeout: Duration::from_millis(500),
            result_settle: Duration::from_millis(30),
            result_settle_max: Duration::from_millis(200),
            download_click_interval: Duration::from_millis(40),
            fallback_retry: Duration::from_secs(10),
            fallback_timeout: Duration::from_millis(500),
        }
    }

    fn history_event(request_id: &str, wall_time_ms: i64) -> Value {
        json!({
            "requestId": request_id,
            "wallTime": wall_time_ms as f64 / 1000.0,
            "initiator": { "type": "script", "stack": { "callFrames": [{ "functionName": "fetchHistory" }] } },
            "request": {
                "url": "https://www.doubao.com/im/chain/single",
                "method": "POST",
                "headers": { "content-type": "application/json" },
                "postData": r#"{"conversation_id":"7001","cursor":0}"#,
                "hasPostData": true
            }
        })
    }

    /// reader `read` / `analyze` 的真实形状(Node vm 对 doubao-result-reader.js 实测):
    /// 有图时 feedback 为 null;feedback 是消息对象,**没有 stage**。
    fn result_reply() -> Value {
        json!({
            "ok": true,
            "images": ["https://x.byteimg.com/a"],
            "feedback": null,
            "binding": { "conversationId": "7001", "submissionId": "9002", "createdAt": 1700000000001i64 }
        })
    }

    const QUOTA_TEXT: &str = "该操作将消耗会员额度，是否继续生成？";
    const EXHAUSTED_TEXT: &str = "今日生成额度已用完，请明天再来";
    /// mock 分类器把它归为 checking(真脚本里对应"视频参数确认"类提示)。
    const PARAMS_TEXT: &str = "请确认生成参数，确认后开始生成";

    /// 已绑定 + 一条 bot 反馈消息(`collectFeedback` 真实返回的 5 个字段)。
    fn feedback_reply(message_id: &str, text: &str) -> Value {
        json!({
            "ok": true,
            "images": [],
            "binding": { "conversationId": "7001", "submissionId": "9002", "createdAt": 1700000000001i64 },
            "feedback": {
                "message": text,
                "messageId": message_id,
                "submissionId": "9002",
                "createdAt": 1700000000002i64,
                "scope": "after_submitted_message"
            }
        })
    }

    /// `run('acceptFeedback', {reply})` 的 mock:与 generation-page.js 的 `feedbackKind` 同口径分类;
    /// `reply` 必须是消息对象(字符串一律 unknown,与真脚本一致)。
    fn accept_feedback_mock(expr: &str) -> MockReply {
        let start = expr
            .find(".run(\"acceptFeedback\", ")
            .expect("acceptFeedback 调用");
        let json_part = &expr[start + ".run(\"acceptFeedback\", ".len()..];
        let json_part = json_part.strip_suffix("))").expect("表达式结尾");
        let payload: Value = serde_json::from_str(json_part).expect("payload 是 JSON");
        let reply = &payload["reply"];
        let Some(text) = reply.get("message").and_then(Value::as_str) else {
            return ok(json!({ "stage": "unknown" }));
        };
        let stage = if text.contains("会员额度") && text.contains("是否继续生成") {
            "quota_checking"
        } else if text.contains("额度已用完") {
            "quota_exhausted"
        } else if text.contains("确认后开始生成") {
            "checking"
        } else {
            "unknown"
        };
        ok(json!({
            "stage": stage, "message": text, "model": "",
            "source": "current_conversation_message",
            "messageId": reply["messageId"], "submissionId": reply["submissionId"],
        }))
    }

    /// 运行 capture 并把事件收集到 Vec。
    async fn run_capture(
        conn: &DoubaoConnection,
        o: CaptureOptions,
        cancel: &CancellationToken,
        confirm_rx: &mut watch::Receiver<bool>,
    ) -> (RelayResult<CaptureOutcome>, Vec<CaptureEvent>) {
        let (tx, mut rx) = mpsc::channel(256);
        let collector = tokio::spawn(async move {
            let mut v = Vec::new();
            while let Some(ev) = rx.recv().await {
                v.push(ev);
            }
            v
        });
        let r = capture_with_fetcher(
            conn,
            o,
            cancel,
            confirm_rx,
            tx,
            fast(),
            FakeFetcher::new(|_| Err("offline".into())),
        )
        .await;
        let events = collector.await.unwrap();
        (r, events)
    }

    type FetchReply = Box<dyn Fn(&str) -> Result<String, String> + Send + Sync>;

    /// 视频路径 3 的假 API:记录请求过的地址,按闭包应答。
    struct FakeFetcher {
        calls: StdMutex<Vec<String>>,
        reply: FetchReply,
    }

    impl FakeFetcher {
        fn new(f: impl Fn(&str) -> Result<String, String> + Send + Sync + 'static) -> Arc<Self> {
            Arc::new(Self {
                calls: StdMutex::new(Vec::new()),
                reply: Box::new(f),
            })
        }

        fn calls(&self) -> Vec<String> {
            self.calls.lock().unwrap().clone()
        }
    }

    #[async_trait]
    impl FallbackFetcher for FakeFetcher {
        async fn fetch(&self, api: &str, _timeout: Duration) -> Result<String, String> {
            self.calls.lock().unwrap().push(api.to_string());
            (self.reply)(api)
        }
    }

    /// 带假 fallback API 的 `run_capture`。
    async fn run_capture_fetch(
        conn: &DoubaoConnection,
        o: CaptureOptions,
        fetcher: Arc<FakeFetcher>,
    ) -> (RelayResult<CaptureOutcome>, Vec<CaptureEvent>) {
        let (tx, mut rx) = mpsc::channel(256);
        let collector = tokio::spawn(async move {
            let mut v = Vec::new();
            while let Some(ev) = rx.recv().await {
                v.push(ev);
            }
            v
        });
        let cancel = CancellationToken::new();
        let (_ctx, mut confirm_rx) = watch::channel(false);
        let r = capture_with_fetcher(conn, o, &cancel, &mut confirm_rx, tx, fast(), fetcher).await;
        let events = collector.await.unwrap();
        (r, events)
    }

    #[tokio::test]
    async fn history_body_is_fetched_and_analyzed_into_originals() {
        let h = handler(|kind, _| match kind {
            "analyze" => ok(result_reply()),
            "read" => ok(json!({ "ok": false, "reason": "read_failed", "images": [] })),
            "terminalFeedback" => ok(json!({ "stage": "unknown" })),
            _ => ok(json!({ "found": true, "text": "" })),
        });
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        mock.set_body("req-h1", r#"{"downlink_body":{}}"#);
        let o = opts();
        let since = o.send_since;
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let mock2 = Arc::new(mock);
        let pusher = {
            let m = Arc::clone(&mock2);
            tokio::spawn(async move {
                tokio::time::sleep(Duration::from_millis(60)).await;
                m.push(
                    "Network.requestWillBeSent",
                    history_event("req-h1", since + 100),
                );
                tokio::time::sleep(Duration::from_millis(10)).await;
                m.push("Network.loadingFinished", json!({ "requestId": "req-h1" }));
            })
        };
        let (r, events) = tokio::time::timeout(
            Duration::from_secs(5),
            run_capture(&conn, o, &cancel, &mut confirm_rx),
        )
        .await
        .unwrap();
        pusher.await.unwrap();
        let out = r.unwrap();
        assert_eq!(out.originals, vec!["https://x.byteimg.com/a".to_string()]);
        assert_eq!(out.watermark_source, WatermarkSource::ConversationOriginal);
        assert_eq!(out.conversation_id.as_deref(), Some("7001"));
        assert_eq!(out.submission_id.as_deref(), Some("9002"));
        let log = mock2.log();
        assert!(
            log.iter().any(|l| l == "Network.getResponseBody req-h1"),
            "{log:?}"
        );
        assert!(log.iter().any(|l| l == "eval:analyze"), "{log:?}");
        assert!(log.iter().any(|l| l == "Network.enable"), "{log:?}");
        assert!(events.contains(&CaptureEvent::Bound {
            conversation_id: "7001".into(),
            submission_id: "9002".into()
        }));
        assert!(events.contains(&CaptureEvent::Stage(TaskStage::Result, None)));
        assert!(events.contains(&CaptureEvent::Originals(vec![
            "https://x.byteimg.com/a".into()
        ])));
    }

    #[tokio::test]
    async fn history_replay_read_yields_originals() {
        let h = handler(|kind, expr| match kind {
            "read" => {
                assert!(expr.contains(r#""expectedLocalMessageId":"lm-123456""#));
                assert!(expr.contains(r#""since":1700000000000"#));
                ok(result_reply())
            }
            "terminalFeedback" => ok(json!({ "stage": "generating", "message": "正在生成" })),
            _ => ok(json!({})),
        });
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let o = opts();
        let since = o.send_since;
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        // 历史请求没有 loadingFinished(body 路径不触发),只靠 1s(这里 30ms)重放。
        mock.push_later(
            Duration::from_millis(30),
            "Network.requestWillBeSent",
            history_event("req-h2", since + 5),
        );
        let (r, events) = tokio::time::timeout(
            Duration::from_secs(5),
            run_capture(&conn, o, &cancel, &mut confirm_rx),
        )
        .await
        .unwrap();
        let out = r.unwrap();
        assert_eq!(out.originals, vec!["https://x.byteimg.com/a".to_string()]);
        assert!(mock.count("eval:read") >= 1, "{:?}", mock.log());
        // reader 只注入一次(表达式前缀),之后直接调用
        assert!(events.contains(&CaptureEvent::Stage(
            TaskStage::Rendering,
            Some("正在生成".into())
        )));
        assert!(events
            .iter()
            .any(|e| matches!(e, CaptureEvent::Feedback(f) if f.stage == "generating")));
    }

    #[tokio::test]
    async fn images_without_binding_are_not_trusted() {
        let h = handler(|kind, _| match kind {
            "read" => ok(
                json!({ "ok": true, "images": ["https://x.byteimg.com/other"], "feedback": null, "binding": null }),
            ),
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let mut o = opts();
        o.deadline = Duration::from_millis(250);
        let since = o.send_since;
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        mock.push_later(
            Duration::from_millis(30),
            "Network.requestWillBeSent",
            history_event("req-h3", since + 5),
        );
        let (r, events) = run_capture(&conn, o, &cancel, &mut confirm_rx).await;
        let e = r.unwrap_err();
        assert_eq!(e.code, "generation_task_timeout");
        assert!(mock.count("eval:read") >= 1, "{:?}", mock.log());
        assert!(!events
            .iter()
            .any(|e| matches!(e, CaptureEvent::Originals(_))));
    }

    #[tokio::test]
    async fn quota_exhausted_feedback_fails_with_account_quota_exhausted() {
        let h = handler(|kind, _| match kind {
            "terminalFeedback" => ok(json!({ "stage": "quota_exhausted", "message": "额度不足" })),
            _ => ok(json!({})),
        });
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let (r, events) = run_capture(&conn, opts(), &cancel, &mut confirm_rx).await;
        let e = r.unwrap_err();
        assert_eq!(e.code, "account_quota_exhausted");
        assert_eq!(e.details.unwrap()["doubao_message"], "额度不足");
        assert!(events.contains(&CaptureEvent::Stage(
            TaskStage::Refused,
            Some("额度不足".into())
        )));
        assert!(events.contains(&CaptureEvent::Refused {
            kind: "quota_exhausted".into(),
            message: "额度不足".into()
        }));
    }

    #[tokio::test]
    async fn refused_feedback_maps_to_doubao_refused() {
        let h = handler(|kind, _| match kind {
            "terminalFeedback" => ok(json!({ "stage": "refused", "message": "不支持该内容" })),
            _ => ok(json!({})),
        });
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let (r, _) = run_capture(&conn, opts(), &cancel, &mut confirm_rx).await;
        let e = r.unwrap_err();
        assert_eq!(e.code, "generation_doubao_refused");
        assert_eq!(e.details.unwrap()["reason"], "refused");
    }

    #[tokio::test]
    async fn cancel_returns_within_50ms() {
        let h = handler(|_, _| ok(json!({ "stage": "unknown" })));
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let c2 = cancel.clone();
        let cancelled_at: Arc<StdMutex<Option<Instant>>> = Arc::new(StdMutex::new(None));
        let slot = Arc::clone(&cancelled_at);
        tokio::spawn(async move {
            tokio::time::sleep(Duration::from_millis(50)).await;
            *slot.lock().unwrap() = Some(Instant::now());
            c2.cancel();
        });
        let (r, _) = run_capture(&conn, opts(), &cancel, &mut confirm_rx).await;
        let since_cancel = cancelled_at.lock().unwrap().expect("已取消").elapsed();
        let e = r.unwrap_err();
        assert_eq!(e.code, "generation_cancelled");
        assert_eq!(e.details.unwrap()["step"], "capture");
        // 计划口径:cancel.cancel() 后 50ms 内返回(从取消时刻起算)
        assert!(since_cancel < Duration::from_millis(50), "{since_cancel:?}");
    }

    #[tokio::test]
    async fn deadline_yields_task_timeout() {
        let h = handler(|_, _| ok(json!({ "stage": "generating" })));
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let mut o = opts();
        o.deadline = Duration::from_millis(120);
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let t0 = Instant::now();
        let (r, _) = run_capture(&conn, o, &cancel, &mut confirm_rx).await;
        let e = r.unwrap_err();
        assert_eq!(e.code, "generation_task_timeout");
        assert_eq!(e.details.unwrap()["last_stage"], "generating");
        assert!(t0.elapsed() < Duration::from_secs(2));
    }

    /// manual + 不自动用会员:会员额度确认只出现在历史消息里(`terminalFeedback` 对它回 unknown),
    /// 经 `acceptFeedback` 分类为 quota_checking → NeedsUserConfirm;用户确认后代回复一次。
    #[tokio::test]
    async fn manual_mode_emits_needs_user_confirm_then_replies_after_confirm() {
        let replied = Arc::new(AtomicBool::new(false));
        let r2 = Arc::clone(&replied);
        let h = handler(move |kind, expr| match kind {
            "terminalFeedback" => ok(json!({ "stage": "unknown" })),
            "acceptFeedback" => accept_feedback_mock(expr),
            "acceptReply" => {
                assert!(expr.contains(r#""prompt":"继续生成""#));
                assert!(expr.contains(r#""userConfirmed":true"#));
                r2.store(true, Ordering::SeqCst);
                ok(json!({ "sent": true, "cleared": true }))
            }
            "read" => {
                if r2.load(Ordering::SeqCst) {
                    ok(result_reply())
                } else {
                    ok(feedback_reply("m-1", QUOTA_TEXT))
                }
            }
            _ => ok(json!({})),
        });
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let mut o = opts();
        o.confirmation_mode = ConfirmationMode::Manual;
        o.auto_use_membership = false;
        let since = o.send_since;
        let cancel = CancellationToken::new();
        let (confirm_tx, mut confirm_rx) = watch::channel(false);
        mock.push_later(
            Duration::from_millis(30),
            "Network.requestWillBeSent",
            history_event("req-h4", since + 5),
        );
        let m = Arc::clone(&mock);
        tokio::spawn(async move {
            // 等历史反馈被分类(且之后又重放了几次)再确认
            for _ in 0..400 {
                if m.count("eval:acceptFeedback") >= 1 && m.count("eval:read") >= 4 {
                    break;
                }
                tokio::time::sleep(Duration::from_millis(5)).await;
            }
            assert_eq!(m.count("eval:acceptReply"), 0, "用户确认前不得代回复");
            let _ = confirm_tx.send(true);
        });
        let (r, events) = tokio::time::timeout(
            Duration::from_secs(5),
            run_capture(&conn, o, &cancel, &mut confirm_rx),
        )
        .await
        .unwrap();
        let out = r.unwrap();
        assert_eq!(out.originals.len(), 1);
        let needs = events
            .iter()
            .position(|e| *e == CaptureEvent::NeedsUserConfirm)
            .expect("NeedsUserConfirm");
        let quota = events
            .iter()
            .position(|e| matches!(e, CaptureEvent::Stage(TaskStage::QuotaChecking, _)))
            .expect("Stage(QuotaChecking)");
        assert!(quota <= needs);
        assert_eq!(mock.count("eval:acceptReply"), 1, "{:?}", mock.log());
        // 同一条反馈消息重放多次,只分类一次
        assert_eq!(mock.count("eval:acceptFeedback"), 1, "{:?}", mock.log());
        assert!(replied.load(Ordering::SeqCst));
        assert!(events.contains(&CaptureEvent::Stage(TaskStage::Result, None)));
        // Task 4.1:用户确认后的代回复发出 → Stage(Sent,"已按用户确认继续"),在 NeedsUserConfirm 之后
        let sent = events
            .iter()
            .position(|e| {
                *e == CaptureEvent::Stage(TaskStage::Sent, Some(USER_CONTINUE_MESSAGE.into()))
            })
            .expect("Stage(Sent, 已按用户确认继续)");
        assert!(needs < sent, "{events:?}");
    }

    #[tokio::test]
    async fn auto_mode_replies_once_and_degrades_to_manual_on_exception() {
        let h = handler(|kind, _| match kind {
            "terminalFeedback" => ok(json!({ "stage": "checking", "message": "是否继续生成?" })),
            "acceptReply" => throw("豆包输入区已有内容，未发送继续生成"),
            _ => ok(json!({})),
        });
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let mut o = opts();
        o.deadline = Duration::from_millis(250);
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let (r, events) = run_capture(&conn, o, &cancel, &mut confirm_rx).await;
        assert_eq!(r.unwrap_err().code, "generation_task_timeout");
        // 自动代回复只尝试一次,失败后降级 manual 并发 NeedsUserConfirm,不再重试
        assert_eq!(mock.count("eval:acceptReply"), 1, "{:?}", mock.log());
        assert_eq!(
            events
                .iter()
                .filter(|e| **e == CaptureEvent::NeedsUserConfirm)
                .count(),
            1
        );
    }

    /// 旧契约的固化点:代回复表达式若像 `acceptFeedback` 那样只回 `{stage:"unknown"}`(没发任何消息),
    /// 绝不能被当成"已发出"——连续 3 次 `sent!=true` 后走降级链并问用户,之后不再盲目重试。
    #[tokio::test]
    async fn accept_reply_without_sent_true_is_not_treated_as_sent() {
        let h = handler(|kind, _| match kind {
            "terminalFeedback" => ok(json!({ "stage": "checking", "message": "是否继续生成?" })),
            "acceptReply" => ok(json!({ "stage": "unknown" })),
            _ => ok(json!({})),
        });
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let mut o = opts();
        o.deadline = Duration::from_millis(400);
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let (r, events) = run_capture(&conn, o, &cancel, &mut confirm_rx).await;
        assert_eq!(r.unwrap_err().code, "generation_task_timeout");
        assert_eq!(
            mock.count("eval:acceptReply"),
            ACCEPT_MAX_ATTEMPTS as usize,
            "{:?}",
            mock.log()
        );
        assert_eq!(
            events
                .iter()
                .filter(|e| **e == CaptureEvent::NeedsUserConfirm)
                .count(),
            1
        );
    }

    /// auto 模式下代回复表达式挂起(`Runtime.evaluate` 超时)同样走降级链(PS 对 `$event.error` 也降级)。
    #[tokio::test]
    async fn evaluate_timeout_during_auto_accept_degrades_to_manual() {
        let h = handler(|kind, _| match kind {
            "terminalFeedback" => ok(json!({ "stage": "checking", "message": "是否继续生成?" })),
            "acceptReply" => hang(Duration::from_millis(800)),
            _ => ok(json!({})),
        });
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let mut o = opts();
        o.deadline = Duration::from_millis(1000);
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let t0 = Instant::now();
        let (r, events) = run_capture(&conn, o, &cancel, &mut confirm_rx).await;
        assert_eq!(r.unwrap_err().code, "generation_task_timeout");
        assert!(
            t0.elapsed() >= Duration::from_millis(500),
            "evaluate_timeout 500ms"
        );
        assert_eq!(mock.count("eval:acceptReply"), 1, "{:?}", mock.log());
        assert_eq!(
            events
                .iter()
                .filter(|e| **e == CaptureEvent::NeedsUserConfirm)
                .count(),
            1
        );
    }

    /// manual + auto_use_membership=false:`checking` 确认并代回复后出现 `quota_checking`,
    /// 必须**再**发一次 `NeedsUserConfirm`,且在用户第二次确认前不得代回复(会员配额单独确认)。
    #[tokio::test]
    async fn image_confirmation_is_consumed_per_stage() {
        use std::sync::atomic::AtomicUsize;
        let phase = Arc::new(AtomicUsize::new(0)); // 0 checking → 1 quota_checking → 2 result
        let confirmed2 = Arc::new(AtomicBool::new(false));
        let violation = Arc::new(AtomicBool::new(false));
        let reads_in_phase1 = Arc::new(AtomicUsize::new(0));
        let (p, c2, v, r1) = (
            Arc::clone(&phase),
            Arc::clone(&confirmed2),
            Arc::clone(&violation),
            Arc::clone(&reads_in_phase1),
        );
        let h = handler(move |kind, expr| match kind {
            "read" => match p.load(Ordering::SeqCst) {
                0 => ok(feedback_reply("m-1", PARAMS_TEXT)),
                1 => {
                    r1.fetch_add(1, Ordering::SeqCst);
                    ok(feedback_reply("m-2", QUOTA_TEXT))
                }
                _ => ok(result_reply()),
            },
            "acceptFeedback" => accept_feedback_mock(expr),
            "acceptReply" => {
                let cur = p.load(Ordering::SeqCst);
                if cur == 1 && !c2.load(Ordering::SeqCst) {
                    v.store(true, Ordering::SeqCst);
                }
                if cur == 0 {
                    assert!(expr.contains(r#""userConfirmed":true"#));
                }
                p.fetch_add(1, Ordering::SeqCst);
                ok(json!({ "sent": true }))
            }
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let mut o = opts();
        o.confirmation_mode = ConfirmationMode::Manual;
        o.auto_use_membership = false;
        let since = o.send_since;
        let cancel = CancellationToken::new();
        let (confirm_tx, mut confirm_rx) = watch::channel(false);
        mock.push_later(
            Duration::from_millis(30),
            "Network.requestWillBeSent",
            history_event("req-h7", since + 5),
        );
        let m = Arc::clone(&mock);
        let (p, c2, r1) = (
            Arc::clone(&phase),
            Arc::clone(&confirmed2),
            Arc::clone(&reads_in_phase1),
        );
        tokio::spawn(async move {
            for _ in 0..400 {
                if m.count("eval:read") >= 2 {
                    break;
                }
                tokio::time::sleep(Duration::from_millis(5)).await;
            }
            let _ = confirm_tx.send(true);
            for _ in 0..400 {
                if p.load(Ordering::SeqCst) == 1 && r1.load(Ordering::SeqCst) >= 2 {
                    break;
                }
                tokio::time::sleep(Duration::from_millis(5)).await;
            }
            c2.store(true, Ordering::SeqCst);
            // 同一个 watch 再 send(true):电平未变也必须被当作新的一次确认
            let _ = confirm_tx.send(true);
        });
        let (r, events) = tokio::time::timeout(
            Duration::from_secs(8),
            run_capture(&conn, o, &cancel, &mut confirm_rx),
        )
        .await
        .unwrap();
        let out = r.unwrap();
        assert_eq!(out.originals.len(), 1);
        assert!(
            !violation.load(Ordering::SeqCst),
            "quota_checking 在用户再次确认前被代回复"
        );
        assert_eq!(mock.count("eval:acceptReply"), 2, "{:?}", mock.log());
        // 两条反馈消息(m-1 / m-2)各分类一次
        assert_eq!(mock.count("eval:acceptFeedback"), 2, "{:?}", mock.log());
        assert_eq!(
            events
                .iter()
                .filter(|e| **e == CaptureEvent::NeedsUserConfirm)
                .count(),
            2,
            "{events:?}"
        );
        let stages: Vec<&CaptureEvent> = events
            .iter()
            .filter(|e| matches!(e, CaptureEvent::Stage(..)))
            .collect();
        assert!(stages
            .iter()
            .any(|e| matches!(e, CaptureEvent::Stage(TaskStage::QuotaChecking, _))));
    }

    /// 视频同理:`videoFollowup` 在 `quota_checking` 只要 `userConfirmed` 为 true 就会发送,
    /// 所以第一次确认被消费后 Rust 必须回到 `videoFeedback`(只识别),直到用户再次确认。
    #[tokio::test]
    async fn video_confirmation_is_consumed_per_stage() {
        use std::sync::atomic::AtomicUsize;
        let phase = Arc::new(AtomicUsize::new(0)); // 0 checking → 1 quota_checking → 2 generating
        let confirmed2 = Arc::new(AtomicBool::new(false));
        let violation = Arc::new(AtomicBool::new(false));
        let feedback_in_phase1 = Arc::new(AtomicUsize::new(0));
        let (p, c2, v, f1) = (
            Arc::clone(&phase),
            Arc::clone(&confirmed2),
            Arc::clone(&violation),
            Arc::clone(&feedback_in_phase1),
        );
        let h = handler(move |kind, expr| match kind {
            "videoFeedback" => match p.load(Ordering::SeqCst) {
                0 => ok(json!({ "stage": "checking", "message": "视频生成参数确认" })),
                1 => {
                    f1.fetch_add(1, Ordering::SeqCst);
                    ok(
                        json!({ "stage": "quota_checking", "message": "本次将消耗会员额度,是否继续生成?" }),
                    )
                }
                _ => ok(json!({ "stage": "generating" })),
            },
            "videoFollowup" => {
                assert!(expr.contains(r#""userConfirmed":true"#));
                match p.load(Ordering::SeqCst) {
                    0 => {
                        p.store(1, Ordering::SeqCst);
                        ok(json!({ "sent": true, "stage": "checking" }))
                    }
                    1 => {
                        if !c2.load(Ordering::SeqCst) {
                            v.store(true, Ordering::SeqCst);
                        }
                        p.store(2, Ordering::SeqCst);
                        ok(json!({ "sent": true, "stage": "quota_checking" }))
                    }
                    _ => ok(json!({ "sent": false, "stage": "generating" })),
                }
            }
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let mut o = opts();
        o.task_type = TaskType::Video;
        o.confirmation_mode = ConfirmationMode::Manual;
        o.auto_use_membership = false;
        o.deadline = Duration::from_millis(900);
        let cancel = CancellationToken::new();
        let (confirm_tx, mut confirm_rx) = watch::channel(false);
        let m = Arc::clone(&mock);
        let (p, c2, f1) = (
            Arc::clone(&phase),
            Arc::clone(&confirmed2),
            Arc::clone(&feedback_in_phase1),
        );
        tokio::spawn(async move {
            for _ in 0..200 {
                if m.count("eval:videoFeedback") >= 2 {
                    break;
                }
                tokio::time::sleep(Duration::from_millis(5)).await;
            }
            let _ = confirm_tx.send(true);
            for _ in 0..200 {
                if p.load(Ordering::SeqCst) == 1 && f1.load(Ordering::SeqCst) >= 2 {
                    break;
                }
                tokio::time::sleep(Duration::from_millis(5)).await;
            }
            c2.store(true, Ordering::SeqCst);
            let _ = confirm_tx.send(true);
        });
        let (r, events) = run_capture(&conn, o, &cancel, &mut confirm_rx).await;
        // 视频结果落盘是 Phase 3,这里到 generating 后超时即可
        assert_eq!(r.unwrap_err().code, "generation_task_timeout");
        assert!(
            !violation.load(Ordering::SeqCst),
            "quota_checking 在用户再次确认前被 videoFollowup 发送"
        );
        assert_eq!(phase.load(Ordering::SeqCst), 2, "{:?}", mock.log());
        assert_eq!(mock.count("eval:videoFollowup"), 2, "{:?}", mock.log());
        assert_eq!(
            events
                .iter()
                .filter(|e| **e == CaptureEvent::NeedsUserConfirm)
                .count(),
            2,
            "{events:?}"
        );
    }

    /// `videoFollowup` 在 auto 模式下报 `manualRequired:true`(页面上没有可自动确认的提示)→ 降级 manual 并问用户。
    #[tokio::test]
    async fn video_manual_required_degrades_auto_to_manual() {
        let h = handler(|kind, _| match kind {
            "videoFollowup" => ok(
                json!({ "sent": false, "stage": "checking", "manualRequired": true, "message": "参数确认" }),
            ),
            "videoFeedback" => ok(json!({ "stage": "checking", "message": "参数确认" })),
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let mut o = opts();
        o.task_type = TaskType::Video;
        o.auto_use_membership = false;
        o.deadline = Duration::from_millis(250);
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let (r, events) = run_capture(&conn, o, &cancel, &mut confirm_rx).await;
        assert_eq!(r.unwrap_err().code, "generation_task_timeout");
        assert_eq!(
            events
                .iter()
                .filter(|e| **e == CaptureEvent::NeedsUserConfirm)
                .count(),
            1
        );
        // 降级后改用只识别的 videoFeedback
        assert!(mock.count("eval:videoFeedback") >= 1, "{:?}", mock.log());
    }

    /// 真机 2.32.6 回归:豆包视频技能回"请先确认参数……确认后我再开始生成视频",`videoFollowup` 只读到"我"
    /// 报 unknown;relay 自己读绑定对话最后一条机器人回复,命中确认话术就代回复"继续生成",只回一次。
    #[tokio::test]
    async fn video_param_confirmation_detected_by_relay_is_auto_replied_once() {
        let h = handler(|kind, _| match kind {
            "videoFollowup" => ok(json!({ "sent": false, "stage": "unknown", "message": "我" })),
            "relayConfirm" => ok(json!({
                "text": "我将基于参考图生成一段视频,请先确认参数:模型 Seedance 2.0 Mini……确认后我再开始生成视频。"
            })),
            "acceptReply" => ok(json!({ "sent": true, "cleared": true })),
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let mut o = opts();
        o.task_type = TaskType::Video;
        o.deadline = Duration::from_millis(400);
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let (r, events) = run_capture(&conn, o, &cancel, &mut confirm_rx).await;
        assert_eq!(r.unwrap_err().code, "generation_task_timeout");
        assert_eq!(mock.count("eval:acceptReply"), 1, "{:?}", mock.log());
        assert!(events.contains(&CaptureEvent::Stage(
            TaskStage::Checking,
            Some("我将基于参考图生成一段视频,请先确认参数:模型 Seedance 2.0 Mini……确认后我再开始生成视频。".into())
        )), "{events:?}");
        assert_eq!(sent_stages(&events).len(), 1, "{events:?}");
        assert!(!events.contains(&CaptureEvent::NeedsUserConfirm));
    }

    #[tokio::test]
    async fn video_distinct_confirmation_messages_in_same_stage_are_each_replied() {
        use std::sync::atomic::AtomicUsize;

        let relay_reads = Arc::new(AtomicUsize::new(0));
        let reads = Arc::clone(&relay_reads);
        let h = handler(move |kind, _| match kind {
            "videoFollowup" => ok(json!({ "sent": false, "stage": "unknown" })),
            "relayConfirm" => {
                let read = reads.fetch_add(1, Ordering::SeqCst);
                let text = if read == 0 {
                    "今日视频生成免费额度已用完，本次将消耗付费额度。是否继续生成？"
                } else {
                    "视频生成参数确认：模型 Seedance 2.0 Fast，确认后我再开始生成视频。"
                };
                ok(json!({ "text": text }))
            }
            "acceptReply" => ok(json!({ "sent": true, "cleared": true })),
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let mut o = opts();
        o.task_type = TaskType::Video;
        o.deadline = Duration::from_millis(500);
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);

        let (result, _) = run_capture(&conn, o, &cancel, &mut confirm_rx).await;

        assert_eq!(result.unwrap_err().code, "generation_task_timeout");
        assert_eq!(mock.count("eval:acceptReply"), 2, "{:?}", mock.log());
    }

    #[test]
    fn relay_confirm_text_matching() {
        for t in [
            "我将基于参考图生成一段美女在海边散步的横版氛围视频，请先确认参数：",
            "视频生成参数确认 模型：Seedance 2.0 Mini 确认后我再开始生成视频。",
            "该操作将消耗会员额度，是否继续生成？",
            "本次使用 Seedance 2.0 Fast 生成，大约需要 1-3 分钟。今日视频生成免费额度已用完，本次将消耗付费额度，并使用优先生成通道。",
        ] {
            assert!(is_relay_confirmation_text(t), "{t}");
        }
        for t in [
            "已生成。",
            "正在生成视频,请稍候",
            "我目前不能直接生成图片",
            "",
        ] {
            assert!(!is_relay_confirmation_text(t), "{t}");
        }
    }

    #[test]
    fn relay_confirm_expression_supports_modern_message_content_dom() {
        let expression = relay_confirm_expression(Some("conversation-1"));

        assert!(expression.contains("[data-testid=message_content]"));
        assert!(expression.contains("data-send-message-boundary"));
        assert!(expression.contains("[data-testid=send_message],[data-testid=receive_message]"));
    }

    fn sent_stages(events: &[CaptureEvent]) -> Vec<Option<String>> {
        events
            .iter()
            .filter_map(|e| match e {
                CaptureEvent::Stage(TaskStage::Sent, m) => Some(m.clone()),
                _ => None,
            })
            .collect()
    }

    /// Task 4.1:auto 模式图像代回复真正发出(`sent:true`)后发 `Stage(Sent,"已自动点继续")`,只发一次,
    /// 且在确认阶段之后。
    #[tokio::test]
    async fn image_auto_reply_emits_sent_stage_once() {
        let replied = Arc::new(AtomicBool::new(false));
        let r2 = Arc::clone(&replied);
        let h = handler(move |kind, _| match kind {
            "terminalFeedback" => {
                if r2.load(Ordering::SeqCst) {
                    ok(json!({ "stage": "generating" }))
                } else {
                    ok(json!({ "stage": "checking", "message": "是否继续生成?" }))
                }
            }
            "acceptReply" => {
                r2.store(true, Ordering::SeqCst);
                ok(json!({ "sent": true, "cleared": true }))
            }
            _ => ok(json!({})),
        });
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let mut o = opts();
        o.deadline = Duration::from_millis(400);
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let (r, events) = run_capture(&conn, o, &cancel, &mut confirm_rx).await;
        assert_eq!(r.unwrap_err().code, "generation_task_timeout");
        assert_eq!(mock.count("eval:acceptReply"), 1, "{:?}", mock.log());
        assert_eq!(
            sent_stages(&events),
            vec![Some(AUTO_CONTINUE_MESSAGE.to_string())],
            "{events:?}"
        );
        let checking = events
            .iter()
            .position(|e| matches!(e, CaptureEvent::Stage(TaskStage::Checking, _)))
            .expect("Stage(Checking)");
        let sent = events
            .iter()
            .position(|e| matches!(e, CaptureEvent::Stage(TaskStage::Sent, _)))
            .unwrap();
        assert!(checking < sent);
        assert!(!events.contains(&CaptureEvent::NeedsUserConfirm));
        assert!(events.contains(&CaptureEvent::Stage(TaskStage::Rendering, None)));
    }

    /// Task 4.1:视频 auto 模式 `videoFollowup` 自行发出 followup(`sent:true`)→ `Stage(Sent,"已自动点继续")`。
    #[tokio::test]
    async fn video_auto_followup_emits_sent_stage() {
        let sent = Arc::new(AtomicBool::new(false));
        let s2 = Arc::clone(&sent);
        let h = handler(move |kind, expr| match kind {
            "videoFollowup" => {
                assert!(expr.contains(r#""confirmationMode":"auto""#));
                if s2.swap(true, Ordering::SeqCst) {
                    ok(json!({ "stage": "generating", "sent": false }))
                } else {
                    ok(json!({ "stage": "checking", "sent": true, "message": "视频生成参数确认" }))
                }
            }
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let mut o = opts();
        o.task_type = TaskType::Video;
        o.deadline = Duration::from_millis(300);
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let (r, events) = run_capture(&conn, o, &cancel, &mut confirm_rx).await;
        assert_eq!(r.unwrap_err().code, "generation_task_timeout");
        assert_eq!(
            sent_stages(&events),
            vec![Some(AUTO_CONTINUE_MESSAGE.to_string())],
            "{events:?}"
        );
        assert!(!events.contains(&CaptureEvent::NeedsUserConfirm));
    }

    /// Task 4.1:auto + 自动用会员,`quota_checking` 代回复抛异常 → 关闭自动用会员(后续 payload
    /// `autoUseMembership:false`)并发一次 `NeedsUserConfirm`;用户确认后以 `userConfirmed:true` 再代回复,
    /// 发出后 `Stage(Sent,"已按用户确认继续")`。
    #[tokio::test]
    async fn membership_auto_failure_degrades_then_user_confirm_replies() {
        use std::sync::atomic::AtomicUsize;
        let attempts = Arc::new(AtomicUsize::new(0));
        let replied = Arc::new(AtomicBool::new(false));
        let payloads = Arc::new(StdMutex::new(Vec::<String>::new()));
        let (a2, r2, p2) = (
            Arc::clone(&attempts),
            Arc::clone(&replied),
            Arc::clone(&payloads),
        );
        let h = handler(move |kind, expr| match kind {
            "terminalFeedback" => {
                p2.lock().unwrap().push(expr.to_string());
                if r2.load(Ordering::SeqCst) {
                    ok(json!({ "stage": "generating" }))
                } else {
                    ok(json!({ "stage": "quota_checking", "message": QUOTA_TEXT }))
                }
            }
            "acceptReply" => {
                if a2.fetch_add(1, Ordering::SeqCst) == 0 {
                    assert!(expr.contains(r#""userConfirmed":false"#));
                    throw("豆包输入区已有内容，未发送继续生成")
                } else {
                    assert!(expr.contains(r#""userConfirmed":true"#));
                    r2.store(true, Ordering::SeqCst);
                    ok(json!({ "sent": true, "cleared": true }))
                }
            }
            _ => ok(json!({})),
        });
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let mut o = opts();
        o.deadline = Duration::from_millis(900);
        let cancel = CancellationToken::new();
        let (confirm_tx, mut confirm_rx) = watch::channel(false);
        let m = Arc::clone(&mock);
        let a3 = Arc::clone(&attempts);
        tokio::spawn(async move {
            for _ in 0..200 {
                if a3.load(Ordering::SeqCst) >= 1 && m.count("eval:terminalFeedback") >= 4 {
                    break;
                }
                tokio::time::sleep(Duration::from_millis(5)).await;
            }
            assert_eq!(a3.load(Ordering::SeqCst), 1, "降级后用户确认前不得再代回复");
            let _ = confirm_tx.send(true);
        });
        let (r, events) = run_capture(&conn, o, &cancel, &mut confirm_rx).await;
        assert_eq!(r.unwrap_err().code, "generation_task_timeout");
        assert_eq!(attempts.load(Ordering::SeqCst), 2, "{:?}", mock.log());
        assert_eq!(
            events
                .iter()
                .filter(|e| **e == CaptureEvent::NeedsUserConfirm)
                .count(),
            1,
            "{events:?}"
        );
        assert_eq!(
            sent_stages(&events),
            vec![Some(USER_CONTINUE_MESSAGE.to_string())],
            "{events:?}"
        );
        let p = payloads.lock().unwrap();
        assert!(p[0].contains(r#""autoUseMembership":true"#));
        assert!(
            p.last().unwrap().contains(r#""autoUseMembership":false"#),
            "降级后 auto_use_membership=false"
        );
        // 会员降级不改确认模式
        assert!(p.last().unwrap().contains(r#""confirmationMode":"auto""#));
    }

    #[tokio::test]
    async fn rendering_unverified_for_too_long_is_not_executed() {
        let h = handler(|kind, _| match kind {
            "terminalFeedback" => {
                ok(json!({ "stage": "rendering_unverified", "message": "正在渲染,请稍候" }))
            }
            _ => ok(json!({})),
        });
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let t0 = Instant::now();
        let (r, events) = run_capture(&conn, opts(), &cancel, &mut confirm_rx).await;
        let e = r.unwrap_err();
        assert_eq!(e.code, "generation_rendering_not_executed");
        assert!(t0.elapsed() >= Duration::from_millis(80));
        assert!(events.contains(&CaptureEvent::Stage(
            TaskStage::RenderingUnverified,
            Some("正在渲染,请稍候".into())
        )));
    }

    #[test]
    fn input_expected_is_empty_prompt_or_continue_reply() {
        assert!(input_is_expected("", "一只蓝色小鸟"));
        assert!(input_is_expected("  \n", "一只蓝色小鸟"));
        assert!(input_is_expected("一只蓝色小鸟", " 一只蓝色小鸟 "));
        assert!(input_is_expected("继续生成", "一只蓝色小鸟"));
        assert!(!input_is_expected("一只红色小鸟", "一只蓝色小鸟"));
        assert!(!input_is_expected("x", ""));
    }

    /// 输入框被改:发一次 `InputModified`(同一内容不重复),**不终止**;恢复为空后再改又发一次。
    #[tokio::test]
    async fn modified_input_is_reported_once_per_text_without_interrupting() {
        let texts = Arc::new(StdMutex::new(vec![
            "用户乱敲的字",
            "用户乱敲的字",
            "",
            "一只蓝色小鸟",
            "又改了",
        ]));
        let t2 = Arc::clone(&texts);
        let h = handler(move |kind, _| match kind {
            "inputCheck" => {
                let mut v = t2.lock().unwrap();
                let text = if v.len() > 1 { v.remove(0) } else { v[0] };
                ok(json!({ "found": true, "text": text }))
            }
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let cancel = CancellationToken::new();
        let c2 = cancel.clone();
        tokio::spawn(async move {
            tokio::time::sleep(Duration::from_millis(700)).await;
            c2.cancel();
        });
        let (_tx, mut confirm_rx) = watch::channel(false);
        let (r, events) = run_capture(&conn, opts(), &cancel, &mut confirm_rx).await;
        assert_eq!(
            r.unwrap_err().code,
            "generation_cancelled",
            "改输入框不应终止任务"
        );
        let modified: Vec<usize> = events
            .iter()
            .filter_map(|e| match e {
                CaptureEvent::InputModified { text_length } => Some(*text_length),
                _ => None,
            })
            .collect();
        assert_eq!(modified, vec![6, 3], "{events:?}");
        assert!(texts.lock().unwrap().len() == 1, "输入框检查应跑满 5 轮");
    }

    #[tokio::test]
    async fn multiple_chat_pages_interrupts_task() {
        let h = handler(|_, _| ok(json!({ "stage": "unknown" })));
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        mock.set_chat_targets(2);
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let (r, _) = run_capture(&conn, opts(), &cancel, &mut confirm_rx).await;
        let e = r.unwrap_err();
        assert_eq!(e.code, "generation_user_interrupted");
        let d = e.details.unwrap();
        assert_eq!(d["interruption"], "multiple_chat_pages");
        assert_eq!(d["chat_targets"], 2);
    }

    /// 0 个 chat page(页面导航走 / 豆包重启中)不是"多开",按断线计数,连续 3 次才失败。
    #[tokio::test]
    async fn zero_chat_pages_is_disconnect_not_interruption() {
        let h = handler(|_, _| ok(json!({ "stage": "unknown" })));
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        mock.set_chat_targets(0);
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let t0 = Instant::now();
        let (r, _) = run_capture(&conn, opts(), &cancel, &mut confirm_rx).await;
        let e = r.unwrap_err();
        assert_eq!(e.code, "doubao_cdp_disconnected");
        let d = e.details.unwrap();
        assert_eq!(d["reason"], "chat_page_missing");
        assert_eq!(d["chat_targets"], 0);
        assert_eq!(d["during"], "capture");
        // 3 次 target 检查(40ms 一次)之后才判定,而不是第一次就终止
        assert!(
            t0.elapsed() >= Duration::from_millis(100),
            "{:?}",
            t0.elapsed()
        );
    }

    /// 视频路径 1:CDN 媒体响应是独立的 GET(不在生成请求集合里),按 PS `videoRequests`
    /// 语义只要是 `send_since` 之后发出的请求、`video/*` + 视频 CDN 白名单即为候选;
    /// 带 `lr=unwatermarked` 的候选直接是无水印原片(`lr_unwatermarked`)。
    #[tokio::test]
    async fn video_cdn_response_with_lr_is_the_original() {
        let h = handler(|_, _| ok(json!({ "stage": "unknown" })));
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let mut o = opts();
        o.task_type = TaskType::Video;
        o.deadline = Duration::from_secs(3);
        let since = o.send_since;
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let m = Arc::clone(&mock);
        tokio::spawn(async move {
            tokio::time::sleep(Duration::from_millis(40)).await;
            // 发送之前就发出的媒体请求:不算
            m.push("Network.requestWillBeSent", json!({
                "requestId": "req-old", "wallTime": (since - 5000) as f64 / 1000.0,
                "request": { "url": "https://v3-vdl.doubao.com/old.mp4?lr=unwatermarked", "method": "GET" }
            }));
            m.push("Network.responseReceived", json!({
                "requestId": "req-old",
                "response": { "url": "https://v3-vdl.doubao.com/old.mp4?lr=unwatermarked", "mimeType": "video/mp4" }
            }));
            // 非白名单 host:不算
            m.push(
                "Network.requestWillBeSent",
                json!({
                    "requestId": "req-v2", "wallTime": (since + 3000) as f64 / 1000.0,
                    "request": { "url": "https://evil.example.com/b.mp4?lr=unwatermarked", "method": "GET" }
                }),
            );
            m.push(
                "Network.responseReceived",
                json!({
                    "requestId": "req-v2",
                    "response": { "url": "https://evil.example.com/b.mp4?lr=unwatermarked", "mimeType": "video/mp4" }
                }),
            );
            // 发送之后的 CDN GET + lr=unwatermarked:原片
            m.push("Network.requestWillBeSent", json!({
                "requestId": "req-v1", "wallTime": (since + 3100) as f64 / 1000.0,
                "request": { "url": "https://v3-vdl.doubao.com/a.mp4?lr=unwatermarked", "method": "GET" }
            }));
            m.push("Network.responseReceived", json!({
                "requestId": "req-v1",
                "response": { "url": "https://v3-vdl.doubao.com/a.mp4?lr=unwatermarked", "mimeType": "video/mp4" }
            }));
        });
        let (r, events) = run_capture(&conn, o, &cancel, &mut confirm_rx).await;
        let out = r.unwrap();
        assert_eq!(
            out.originals,
            vec!["https://v3-vdl.doubao.com/a.mp4?lr=unwatermarked".to_string()]
        );
        assert_eq!(out.watermark_source, WatermarkSource::LrUnwatermarked);
        assert_eq!(
            events
                .iter()
                .filter(|e| matches!(e, CaptureEvent::VideoCandidates(_)))
                .collect::<Vec<_>>(),
            vec![&CaptureEvent::VideoCandidates(vec![
                "https://v3-vdl.doubao.com/a.mp4?lr=unwatermarked".into()
            ])]
        );
        assert!(events.contains(&CaptureEvent::Stage(TaskStage::Result, None)));
    }

    /// 不带 `lr=unwatermarked` 的直链(播放器里的带水印版本)只是候选,不能当结果。
    #[tokio::test]
    async fn video_candidate_without_lr_is_not_a_result() {
        let h = handler(|_, _| ok(json!({ "stage": "unknown" })));
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let mut o = opts();
        o.task_type = TaskType::Video;
        o.deadline = Duration::from_millis(300);
        let since = o.send_since;
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let m = Arc::clone(&mock);
        tokio::spawn(async move {
            tokio::time::sleep(Duration::from_millis(40)).await;
            m.push(
                "Network.requestWillBeSent",
                json!({
                    "requestId": "req-w", "wallTime": (since + 100) as f64 / 1000.0,
                    "request": { "url": "https://v9-vdl.doubao.com/w.mp4?a=1", "method": "GET" }
                }),
            );
            m.push("Network.responseReceived", json!({
                "requestId": "req-w",
                "response": { "url": "https://v9-vdl.doubao.com/w.mp4?a=1", "mimeType": "video/mp4" }
            }));
        });
        let (r, events) = run_capture(&conn, o, &cancel, &mut confirm_rx).await;
        assert_eq!(r.unwrap_err().code, "generation_task_timeout");
        assert!(events.contains(&CaptureEvent::VideoCandidates(vec![
            "https://v9-vdl.doubao.com/w.mp4?a=1".into()
        ])));
        assert!(!events
            .iter()
            .any(|e| matches!(e, CaptureEvent::Originals(_))));
    }

    /// 视频路径 2:feedback 到 `result` 后每隔 `download_click_interval` 调 `requestVideoDownload`,
    /// 点到(`clicked:true`)就不再点;之后豆包发出的 `lr=unwatermarked` 请求由路径 1 接住。
    #[tokio::test]
    async fn video_download_button_is_clicked_after_result_until_clicked() {
        let clicks = Arc::new(std::sync::atomic::AtomicUsize::new(0));
        let c = Arc::clone(&clicks);
        let h = handler(move |kind, _| match kind {
            "videoFollowup" => ok(json!({ "stage": "result", "message": "你的视频生成好了" })),
            "requestVideoDownload" => {
                // 前两次按钮还没出来,第三次点到
                if c.fetch_add(1, Ordering::SeqCst) < 2 {
                    ok(json!({ "clicked": false, "reason": "download_button_not_found" }))
                } else {
                    ok(json!({ "clicked": true }))
                }
            }
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let mut o = opts();
        o.task_type = TaskType::Video;
        o.deadline = Duration::from_secs(3);
        let since = o.send_since;
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let m = Arc::clone(&mock);
        let c2 = Arc::clone(&clicks);
        tokio::spawn(async move {
            // 点到之后豆包才去拿无水印原片
            while c2.load(Ordering::SeqCst) < 3 {
                tokio::time::sleep(Duration::from_millis(5)).await;
            }
            tokio::time::sleep(Duration::from_millis(100)).await;
            m.push("Network.requestWillBeSent", json!({
                "requestId": "req-dl", "wallTime": (since + 60_000) as f64 / 1000.0,
                "request": { "url": "https://v26-vdl.doubao.com/dl.mp4?x=1&lr=unwatermarked", "method": "GET" }
            }));
            m.push("Network.responseReceived", json!({
                "requestId": "req-dl",
                "response": { "url": "https://v26-vdl.doubao.com/dl.mp4?x=1&lr=unwatermarked", "mimeType": "video/mp4" }
            }));
        });
        let (r, events) = run_capture(&conn, o, &cancel, &mut confirm_rx).await;
        let out = r.unwrap();
        assert_eq!(
            out.originals,
            vec!["https://v26-vdl.doubao.com/dl.mp4?x=1&lr=unwatermarked".to_string()]
        );
        assert_eq!(out.watermark_source, WatermarkSource::LrUnwatermarked);
        assert_eq!(
            mock.count("eval:requestVideoDownload"),
            3,
            "点到之后不应再点:{:?}",
            mock.log()
        );
        assert!(events.contains(&CaptureEvent::Stage(
            TaskStage::Result,
            Some("你的视频生成好了".into())
        )));
    }

    /// stage 不是 `result` 时不点下载按钮;图像任务永远不点。
    #[tokio::test]
    async fn download_button_is_not_clicked_before_result() {
        let h = handler(|kind, _| match kind {
            "videoFollowup" => ok(
                json!({ "stage": "generating", "message": "本次使用 Seedance 生成,预计等待 3 分钟" }),
            ),
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let mut o = opts();
        o.task_type = TaskType::Video;
        o.deadline = Duration::from_millis(250);
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let (r, _) = run_capture(&conn, o, &cancel, &mut confirm_rx).await;
        assert_eq!(r.unwrap_err().code, "generation_task_timeout");
        assert_eq!(mock.count("eval:requestVideoDownload"), 0);
    }

    /// 已绑定的视频 `read` 返回(reader `resolveVideos` 的真实形状)。
    fn video_read_reply(extra: Value) -> Value {
        let mut v = json!({
            "ok": true,
            "images": [],
            "feedback": null,
            "binding": { "conversationId": "7001", "submissionId": "9002", "createdAt": 1700000000001i64 },
            "followupSubmissionIds": [],
        });
        for (k, val) in extra.as_object().unwrap() {
            v[k] = val.clone();
        }
        v
    }

    /// reader 经 `get_without_watermark` 换到的原片:`images[{url, unwatermarked:true, vid}]` → `lr_unwatermarked`。
    #[tokio::test]
    async fn reader_unwatermarked_video_is_the_original_once_bound() {
        let h = handler(|kind, _| match kind {
            "read" => ok(video_read_reply(json!({ "images": [{
                "key": "k1", "messageId": "9003", "conversationId": "7001", "createdAt": 1700000001000i64,
                "vid": "v0d1", "fallbackAPI": null,
                "url": "https://v3-vdl.doubao.com/o.mp4?lr=unwatermarked&sig=1", "unwatermarked": true
            }, {
                "url": "https://v3-vdl.doubao.com/wm.mp4", "unwatermarked": true
            }] }))),
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let mut o = opts();
        o.task_type = TaskType::Video;
        o.deadline = Duration::from_secs(3);
        mock.push_later(
            Duration::from_millis(30),
            "Network.requestWillBeSent",
            history_event("req-hv", o.send_since + 5),
        );
        let (r, events) = run_capture_fetch(&conn, o, FakeFetcher::new(|_| Err("x".into()))).await;
        let out = r.unwrap();
        assert_eq!(
            out.originals,
            vec!["https://v3-vdl.doubao.com/o.mp4?lr=unwatermarked&sig=1".to_string()]
        );
        assert_eq!(out.watermark_source, WatermarkSource::LrUnwatermarked);
        assert_eq!(out.submission_id.as_deref(), Some("9002"));
        assert!(events
            .iter()
            .any(|e| matches!(e, CaptureEvent::Bound { .. })));
    }

    #[tokio::test]
    async fn reader_video_without_binding_is_not_trusted() {
        let h = handler(|kind, _| match kind {
            "read" => {
                let mut v = video_read_reply(json!({
                    "images": [{ "url": "https://v3-vdl.doubao.com/o.mp4?lr=unwatermarked", "unwatermarked": true }],
                    "fallbacks": [{ "vid": "v1", "api": "https://vas-hl-x.snssdk.com/video/openapi/v1/?vid=v1" }]
                }));
                v["binding"] = Value::Null;
                ok(v)
            }
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let mut o = opts();
        o.task_type = TaskType::Video;
        o.deadline = Duration::from_millis(300);
        mock.push_later(
            Duration::from_millis(30),
            "Network.requestWillBeSent",
            history_event("req-hu", o.send_since + 5),
        );
        let fetcher = FakeFetcher::new(|_| Err("x".into()));
        let (r, _) = run_capture_fetch(&conn, o, Arc::clone(&fetcher)).await;
        assert_eq!(r.unwrap_err().code, "generation_task_timeout");
        assert!(fetcher.calls().is_empty(), "未绑定不请求 fallback API");
    }

    /// 视频路径 3:fallback API → Rust 解密 → 没有 lr 的地址追加 `lr=unwatermarked` → `fallback_decrypted`;
    /// 同一个 vid 10s 内只请求一次;白名单外的 API 不请求。
    #[tokio::test]
    async fn fallback_api_is_fetched_once_and_decrypted_in_rust() {
        use base64::Engine;
        let token = base64::engine::general_purpose::STANDARD
            .encode("https://v26-vdl.doubao.com/real.mp4?a=1&b=2");
        let body = json!({ "data": { "video_info": { "video_list": {
            "video_1": { "main_url": token, "codec_type": "h264", "bitrate": 1000, "vwidth": 1280, "vheight": 720 }
        } } } })
        .to_string();
        let h = handler(|kind, _| match kind {
            "read" => ok(video_read_reply(json!({ "fallbacks": [
                { "vid": "v1", "api": "https://vas-hl-x.snssdk.com/video/openapi/v1/?vid=v1&logo_type=unwatermarked" },
                { "vid": "v2", "api": "http://vas-hl-x.snssdk.com/insecure" },
            ] }))),
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let mut o = opts();
        o.task_type = TaskType::Video;
        o.deadline = Duration::from_secs(3);
        mock.push_later(
            Duration::from_millis(30),
            "Network.requestWillBeSent",
            history_event("req-hf", o.send_since + 5),
        );
        let fetcher = FakeFetcher::new(move |_| Ok(body.clone()));
        let (r, events) = run_capture_fetch(&conn, o, Arc::clone(&fetcher)).await;
        let out = r.unwrap();
        assert_eq!(
            out.originals,
            vec!["https://v26-vdl.doubao.com/real.mp4?a=1&b=2&lr=unwatermarked".to_string()]
        );
        assert_eq!(out.watermark_source, WatermarkSource::FallbackDecrypted);
        assert_eq!(
            fetcher.calls(),
            vec![
                "https://vas-hl-x.snssdk.com/video/openapi/v1/?vid=v1&logo_type=unwatermarked"
                    .to_string()
            ]
        );
        assert_eq!(
            mock.count("eval:videoDecoder"),
            0,
            "Rust 解出来就不用页面脚本"
        );
        assert!(events
            .iter()
            .any(|e| matches!(e, CaptureEvent::Fallbacks(f) if f.len() == 2)));
    }

    /// Rust 解不出(响应里没有 main_url 等)→ 注入原版 `shenlangDoubaoVideoDecoder.decode()` 兜底。
    #[tokio::test]
    async fn fallback_falls_back_to_page_decoder_when_rust_fails() {
        let h = handler(|kind, expr| match kind {
            "read" => ok(video_read_reply(json!({ "fallbacks": [
                { "vid": "v9", "api": "https://vas-a1.snssdk.com/x?vid=v9" },
            ] }))),
            "videoDecoder" => {
                assert!(
                    expr.contains(r#"JSON.parse("{\"weird\":true}")"#),
                    "{}",
                    &expr[expr.len() - 200..]
                );
                assert!(expr.contains(r#""https://vas-a1.snssdk.com/x?vid=v9")"#));
                assert!(expr.contains("typeof shenlangDoubaoVideoDecoder === 'undefined'"));
                ok(
                    json!({ "ok": true, "url": "https://v3-vdl.doubao.com/js.mp4?lr=unwatermarked" }),
                )
            }
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let mut o = opts();
        o.task_type = TaskType::Video;
        o.deadline = Duration::from_secs(3);
        mock.push_later(
            Duration::from_millis(30),
            "Network.requestWillBeSent",
            history_event("req-hj", o.send_since + 5),
        );
        let fetcher = FakeFetcher::new(|_| Ok(r#"{"weird":true}"#.to_string()));
        let (r, _) = run_capture_fetch(&conn, o, fetcher).await;
        let out = r.unwrap();
        assert_eq!(
            out.originals,
            vec!["https://v3-vdl.doubao.com/js.mp4?lr=unwatermarked".to_string()]
        );
        assert_eq!(out.watermark_source, WatermarkSource::FallbackDecrypted);
        assert_eq!(mock.count("eval:videoDecoder"), 1);
    }

    /// Rust 解出的地址不是视频 CDN(`fallback_original_url` 拒收)→ 同样改用页面脚本解密。
    #[tokio::test]
    async fn fallback_uses_page_decoder_when_rust_url_is_not_video_cdn() {
        use base64::Engine;
        let token =
            base64::engine::general_purpose::STANDARD.encode("https://evil.example.com/a.mp4");
        let body = json!({ "data": { "video_info": { "video_list": {
            "video_1": { "main_url": token, "codec_type": "h264", "bitrate": 1000, "vwidth": 1280, "vheight": 720 }
        } } } })
        .to_string();
        let h = handler(|kind, _| match kind {
            "read" => ok(video_read_reply(json!({ "fallbacks": [
                { "vid": "v7", "api": "https://vas-a1.snssdk.com/x?vid=v7" },
            ] }))),
            "videoDecoder" => ok(
                json!({ "ok": true, "url": "https://v3-vdl.doubao.com/js7.mp4?lr=unwatermarked" }),
            ),
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let mut o = opts();
        o.task_type = TaskType::Video;
        o.deadline = Duration::from_secs(3);
        mock.push_later(
            Duration::from_millis(30),
            "Network.requestWillBeSent",
            history_event("req-hn", o.send_since + 5),
        );
        let fetcher = FakeFetcher::new(move |_| Ok(body.clone()));
        let (r, _) = run_capture_fetch(&conn, o, fetcher).await;
        let out = r.unwrap();
        assert_eq!(
            out.originals,
            vec!["https://v3-vdl.doubao.com/js7.mp4?lr=unwatermarked".to_string()]
        );
        assert_eq!(out.watermark_source, WatermarkSource::FallbackDecrypted);
        assert_eq!(mock.count("eval:videoDecoder"), 1);
    }

    #[test]
    fn video_url_whitelists() {
        // 路径 1:MIME video/* + https + 视频 CDN
        for (url, mime) in [
            (
                "https://v3-vdl.doubao.com/a.mp4?lr=unwatermarked",
                "video/mp4",
            ),
            ("https://x.bytevcloudcn.com/a", "VIDEO/MP4"),
            ("https://cdn.doubaovideo.com/a", "video/webm"),
            ("https://a.b.volcvideo.com/a", "video/mp4"),
        ] {
            assert_eq!(
                video_candidate_url(url, mime).as_deref(),
                Some(url),
                "{url}"
            );
        }
        for (url, mime) in [
            ("http://v3-vdl.doubao.com/a.mp4", "video/mp4"),
            ("https://v3-vdl.doubao.com/a.mp4", "application/json"),
            ("https://v3-vdl.doubao.com.evil.com/a.mp4", "video/mp4"),
            ("https://www.doubao.com/a.mp4", "video/mp4"),
            ("https://u:p@v3-vdl.doubao.com/a.mp4", "video/mp4"),
            ("not a url", "video/mp4"),
        ] {
            assert_eq!(video_candidate_url(url, mime), None, "{url}");
        }
        // 路径 3:API 必须 https://vas-[a-z0-9-]+.snssdk.com 且无 userinfo
        assert!(is_fallback_api_url(
            "https://vas-hl-x.snssdk.com/video/openapi/v1/?vid=1"
        ));
        assert!(is_fallback_api_url("https://vas-1a.snssdk.com/"));
        for api in [
            "http://vas-hl-x.snssdk.com/x",
            "https://vas-.snssdk.com/x",
            "https://vas_x.snssdk.com/x",
            "https://vas-X.snssdk.com.evil.com/x",
            "https://x.vas-a.snssdk.com/x",
            "https://api.snssdk.com/x",
            "https://user@vas-a.snssdk.com/x",
            "https://u:p@vas-a.snssdk.com/x",
            "",
        ] {
            assert!(!is_fallback_api_url(api), "{api}");
        }
        // 解密结果 → 原片地址
        assert_eq!(
            fallback_original_url("https://v3-vdl.doubao.com/a.mp4?lr=unwatermarked").as_deref(),
            Some("https://v3-vdl.doubao.com/a.mp4?lr=unwatermarked")
        );
        assert_eq!(
            fallback_original_url("https://v3-vdl.doubao.com/a.mp4?x=1&lr=wm").as_deref(),
            Some("https://v3-vdl.doubao.com/a.mp4?x=1&lr=unwatermarked")
        );
        assert_eq!(
            fallback_original_url("https://evil.com/a.mp4?lr=unwatermarked"),
            None
        );
        assert_eq!(
            fallback_original_url("http://v3-vdl.doubao.com/a.mp4"),
            None
        );
        // 表达式
        let e = request_video_download_expression(
            "shenlangDoubaoTask_abc",
            &json!({ "type": "video" }),
            false,
        );
        assert!(
            e.contains(r#"shenlangDoubaoTask_abc.run("requestVideoDownload", {"type":"video"})"#)
        );
        assert!(e.starts_with("typeof shenlangDoubaoTask_abc === 'undefined'"));
        let d = video_decoder_expression("{\"a\":1}", "https://vas-a.snssdk.com/x");
        assert!(d.contains(VIDEO_DECODER_SCRIPT));
        assert!(d.ends_with(r#"shenlangDoubaoVideoDecoder.decode(JSON.parse("{\"a\":1}"), "https://vas-a.snssdk.com/x")); })()"#));
    }

    #[tokio::test]
    async fn generation_response_with_identity_marker_is_extracted_directly() {
        let h = handler(|kind, _| match kind {
            "streamSubmissionBinding" => {
                ok(json!({ "conversationId": "7001", "submissionId": "9009", "createdAt": 1 }))
            }
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let o = opts();
        let since = o.send_since;
        mock.set_body(
            "req-g1",
            "data: {\"local_message_id\":\"lm-123456\",\"image_ori_raw\":{\"url\":\"https://p.byteimg.com/direct\"}}\n",
        );
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let mock = Arc::new(mock);
        let m = Arc::clone(&mock);
        tokio::spawn(async move {
            tokio::time::sleep(Duration::from_millis(40)).await;
            m.push("Network.requestWillBeSent", json!({
                "requestId": "req-g1",
                "wallTime": (since + 50) as f64 / 1000.0,
                "request": { "url": "https://www.doubao.com/samantha/chat/completion", "method": "POST", "hasPostData": true }
            }));
            // 与本任务无关的 SSE(不含身份标记)不会被采信
            m.push(
                "Network.eventSourceMessageReceived",
                json!({
                    "requestId": "req-g1",
                    "data": "{\"image_ori_raw\":{\"url\":\"https://p.byteimg.com/foreign\"}}"
                }),
            );
            tokio::time::sleep(Duration::from_millis(10)).await;
            m.push("Network.loadingFinished", json!({ "requestId": "req-g1" }));
        });
        let (r, events) = tokio::time::timeout(
            Duration::from_secs(5),
            run_capture(&conn, o, &cancel, &mut confirm_rx),
        )
        .await
        .unwrap();
        let out = r.unwrap();
        assert_eq!(
            out.originals,
            vec!["https://p.byteimg.com/direct".to_string()]
        );
        assert!(mock
            .log()
            .iter()
            .any(|l| l == "Network.getResponseBody req-g1"));
        assert!(!events.iter().any(
            |e| matches!(e, CaptureEvent::Originals(v) if v.iter().any(|u| u.contains("foreign")))
        ));
    }

    #[tokio::test]
    async fn reader_initiated_history_requests_are_ignored() {
        let h = handler(|kind, _| match kind {
            "read" => ok(result_reply()),
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let mut o = opts();
        o.deadline = Duration::from_millis(200);
        let since = o.send_since;
        let mut ev = history_event("req-self", since + 5);
        ev["initiator"]["stack"]["callFrames"] =
            json!([{ "functionName": "shenlangReadBoundConversation" }]);
        mock.push_later(Duration::from_millis(30), "Network.requestWillBeSent", ev);
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let (r, _) = run_capture(&conn, o, &cancel, &mut confirm_rx).await;
        // 没有可用的历史模板 → 从不 read → 超时
        assert_eq!(r.unwrap_err().code, "generation_task_timeout");
        assert_eq!(mock.count("eval:read"), 0, "{:?}", mock.log());
    }

    #[tokio::test]
    async fn chat_session_is_reconnected_after_close() {
        let h = handler(|kind, _| match kind {
            "read" => ok(result_reply()),
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let o = opts();
        let since = o.send_since;
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let mock = Arc::new(mock);
        let m = Arc::clone(&mock);
        tokio::spawn(async move {
            tokio::time::sleep(Duration::from_millis(30)).await;
            m.close_all();
            tokio::time::sleep(Duration::from_millis(120)).await;
            m.push(
                "Network.requestWillBeSent",
                history_event("req-h5", since + 5),
            );
        });
        let (r, _) = tokio::time::timeout(
            Duration::from_secs(5),
            run_capture(&conn, o, &cancel, &mut confirm_rx),
        )
        .await
        .unwrap();
        let out = r.unwrap();
        assert_eq!(out.originals.len(), 1);
        assert!(mock.count("Network.enable") >= 2, "{:?}", mock.log());
    }

    /// Task 6.1:豆包 reload 期间 chat page 短暂消失(`/json/list` 里一个 chat 都没有),
    /// 捕获不能因一次 `chat_session()` 失败就以 `doubao_cdp_disconnected` 结束,
    /// 应按 `reconnect_delay` 重试,新 target 出现后接上并继续拿到结果。
    #[tokio::test]
    async fn capture_waits_for_chat_page_to_return_after_reload() {
        let h = handler(|kind, _| match kind {
            "read" => ok(result_reply()),
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let o = opts();
        let since = o.send_since;
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let mock = Arc::new(mock);
        let m = Arc::clone(&mock);
        tokio::spawn(async move {
            tokio::time::sleep(Duration::from_millis(30)).await;
            m.set_chat_targets(0);
            m.close_all();
            tokio::time::sleep(Duration::from_millis(100)).await;
            *m.targets.lock().unwrap() = vec![m.chat_target("CHAT-NEW")];
            tokio::time::sleep(Duration::from_millis(150)).await;
            m.push(
                "Network.requestWillBeSent",
                history_event("req-h6", since + 6),
            );
        });
        let (r, _) = tokio::time::timeout(
            Duration::from_secs(5),
            run_capture(&conn, o, &cancel, &mut confirm_rx),
        )
        .await
        .unwrap();
        let out = r.unwrap();
        assert_eq!(out.originals.len(), 1);
        assert_eq!(conn.chat_session().await.unwrap().target_id, "CHAT-NEW");
    }

    /// 豆包整体 reload 时 background session 也关闭:chat 重连后必须重取 background、重新订阅 + `Network.enable`,
    /// 之后 background 侧推来的 Network 事件仍能驱动捕获。
    #[tokio::test]
    async fn background_session_is_reattached_after_reload() {
        use std::sync::atomic::AtomicUsize;
        let h = handler(|kind, _| match kind {
            "read" => ok(result_reply()),
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = MockDoubao::start(h).await;
        mock.set_chat_and_background();
        let conn = mock.conn().await;
        assert!(conn.background_session().await.is_some());
        let o = opts();
        let since = o.send_since;
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let mock = Arc::new(mock);
        let enables_before_close = Arc::new(AtomicUsize::new(0));
        let m = Arc::clone(&mock);
        let before = Arc::clone(&enables_before_close);
        tokio::spawn(async move {
            tokio::time::sleep(Duration::from_millis(30)).await;
            before.store(m.count("Network.enable"), Ordering::SeqCst);
            m.close_all();
            tokio::time::sleep(Duration::from_millis(150)).await;
            // push 广播给所有已连接的 session(含重连后的 background)
            m.push(
                "Network.requestWillBeSent",
                history_event("req-h8", since + 5),
            );
        });
        let (r, _) = tokio::time::timeout(
            Duration::from_secs(5),
            run_capture(&conn, o, &cancel, &mut confirm_rx),
        )
        .await
        .unwrap();
        let out = r.unwrap();
        assert_eq!(out.originals.len(), 1);
        let before = enables_before_close.load(Ordering::SeqCst);
        assert!(before >= 2, "{:?}", mock.log());
        // 重连后 chat + background 各再 enable 一次(连接层自身的 enable 另计)
        assert!(
            mock.count("Network.enable") >= before + 2,
            "before={before} log={:?}",
            mock.log()
        );
        assert!(conn.background_session().await.is_some());
    }

    /// 需要真豆包(端口 9707):`cargo test live_capture_expressions_dry_run -- --ignored --nocapture`。
    /// 只读:注入 reader 与一次性 trackingKey 全局对象,把捕获用到的每条表达式各求值一次,
    /// 确认在真实页面里不抛异常、返回形状符合预期;不发送任何消息,不消耗配额。
    #[tokio::test]
    #[ignore]
    async fn live_capture_expressions_dry_run() {
        use crate::doubao::CDP_HOST;
        let targets = list_targets(CDP_HOST, 9707).await.unwrap();
        let (chat, _) = select_pages(&targets, 9707);
        let chat = chat.first().expect("需要 doubao-chat page");
        let session = CdpSession::connect(&chat.ws_url, &chat.id).await.unwrap();
        let key = format!("shenlangDoubaoTask_live{}", crate::cdp::now_ms());
        let o = CaptureOptions {
            tracking_key: key.clone(),
            ..opts()
        };
        let t = CaptureTiming::default();

        // 1. reader 注入 + 空 SSE 的 streamSubmissionBinding(应为 null,不抛)
        let expr = format!(
            "{}\n{}",
            reader_bootstrap(),
            stream_binding_expression("data: {}", &o.identity)
        );
        let v = eval(Arc::clone(&session), expr, t.evaluate_timeout)
            .await
            .unwrap();
        println!("streamSubmissionBinding(empty) => {v}");
        assert!(v.is_null());

        // 2. read:用假的历史请求(reader 的 validateRequest 通过,fetch 由页面发出,
        //    对话不存在 → ok=false),验证表达式与返回形状
        let h = HistoryRequest {
            url: "https://www.doubao.com/im/chain/single".into(),
            post_data: r#"{"conversation_id":"1","cursor":0}"#.into(),
            headers: json!({ "content-type": "application/json" }),
            conversation_id: Some("1".into()),
        };
        let options = read_options(
            TaskType::Image,
            &o.prompt,
            &o.identity,
            o.send_since,
            Some("1"),
            None,
            None,
            &[],
        );
        let v = eval(
            Arc::clone(&session),
            read_expression(&h, &options),
            t.evaluate_timeout,
        )
        .await
        .unwrap();
        println!("read(fake) => {v}");
        assert!(v.is_object());
        assert!(v.get("ok").is_some());
        assert!(v.get("images").is_some());

        // 3. analyze:空 body → {ok:true, images:[], feedback:null, binding:null}
        let v = eval(
            Arc::clone(&session),
            analyze_expression("{}", &h, &options),
            t.evaluate_timeout,
        )
        .await
        .unwrap();
        println!("analyze({{}}) => {v}");
        assert_eq!(v["ok"], true);
        assert_eq!(v["images"], json!([]));

        // 4. feedback 轮询(带注入):没有提交状态 → stage unknown
        let payload = feedback_payload(&o, ConfirmationMode::Auto, true, false);
        let v = eval(
            Arc::clone(&session),
            feedback_expression(&key, "terminalFeedback", &payload, true),
            t.evaluate_timeout,
        )
        .await
        .unwrap();
        println!("terminalFeedback => {v}");
        assert_eq!(v["stage"], "unknown");
        assert!(v.get("missing").is_none());
        let v = eval(
            Arc::clone(&session),
            feedback_expression(&key, "videoFollowup", &payload, false),
            t.evaluate_timeout,
        )
        .await
        .unwrap();
        println!("videoFollowup => {v}");
        assert_eq!(v["stage"], "unknown");

        // 4b. acceptFeedback 只分类不发送:reader 形状的消息对象 → quota_checking / quota_exhausted
        for (text, want) in [
            ("该操作将消耗会员额度，是否继续生成？", "quota_checking"),
            ("今日生成额度已用完，请明天再来", "quota_exhausted"),
        ] {
            let reply = json!({ "message": text, "messageId": "1002", "submissionId": "1001",
                "createdAt": crate::cdp::now_ms(), "scope": "after_submitted_message" });
            let payload = accept_feedback_payload(&o, "1001", &reply);
            let v = eval(
                Arc::clone(&session),
                accept_feedback_expression(&key, &payload, false),
                t.evaluate_timeout,
            )
            .await
            .unwrap();
            println!("acceptFeedback({want}) => {v}");
            assert_eq!(v["stage"], want);
        }

        // 5. 输入框脏检查
        let v = eval(
            Arc::clone(&session),
            input_check_expression(),
            t.evaluate_timeout,
        )
        .await
        .unwrap();
        println!("inputCheck => {v}");
        assert_eq!(v["found"], true);

        // 6. chat target 唯一性
        let (chat, _) = select_pages(&list_targets(CDP_HOST, 9707).await.unwrap(), 9707);
        assert_eq!(chat.len(), 1);

        // 清理一次性全局对象
        let _ = session
            .evaluate(&format!("delete window.{key}; typeof {key}"))
            .await;
        session.close().await;
    }

    #[tokio::test]
    async fn video_task_uses_followup_step_and_requests_fallback_once() {
        let h = handler(|kind, _| match kind {
            "videoFollowup" => ok(json!({ "stage": "checking", "sent": true })),
            "read" => ok(json!({
                "ok": true, "images": [], "feedback": { "stage": "generating" },
                "binding": { "conversationId": "7001", "submissionId": "9002" },
                "fallbacks": [{ "vid": "v1", "api": "https://vas-x.snssdk.com/x" }]
            })),
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let mut o = opts();
        o.task_type = TaskType::Video;
        o.deadline = Duration::from_millis(250);
        let since = o.send_since;
        mock.push_later(
            Duration::from_millis(30),
            "Network.requestWillBeSent",
            history_event("req-h6", since + 5),
        );
        // API 一直失败:10s 内同一个 vid 只请求一次,任务继续等(最终超时)
        let fetcher = FakeFetcher::new(|_| Err("fallback_request_failed:http_500".into()));
        let (r, events) = run_capture_fetch(&conn, o, Arc::clone(&fetcher)).await;
        assert_eq!(r.unwrap_err().code, "generation_task_timeout");
        assert!(mock.count("eval:videoFollowup") >= 1, "{:?}", mock.log());
        assert!(mock.count("eval:read") >= 2, "{:?}", mock.log());
        assert_eq!(mock.count("eval:acceptFeedback"), 0);
        assert_eq!(
            fetcher.calls(),
            vec!["https://vas-x.snssdk.com/x".to_string()]
        );
        assert!(
            events.contains(&CaptureEvent::Fallbacks(vec![FallbackVideo {
                vid: "v1".into(),
                api: "https://vas-x.snssdk.com/x".into()
            }]))
        );
        assert!(events
            .iter()
            .any(|e| matches!(e, CaptureEvent::Feedback(f) if f.stage == "checking" && f.sent)));
        assert!(!events.contains(&CaptureEvent::NeedsUserConfirm));
    }

    #[test]
    fn accept_feedback_expression_passes_reader_object_as_reply() {
        let reply = json!({ "message": QUOTA_TEXT, "messageId": "m-1", "submissionId": "9002",
            "createdAt": 1, "scope": "after_submitted_message" });
        let payload = accept_feedback_payload(&opts(), "9002", &reply);
        assert_eq!(payload["boundSubmissionId"], "9002");
        assert_eq!(payload["reply"], reply, "reply 必须是消息对象原样");
        assert_eq!(payload["prompt"], "一只蓝色小鸟");
        assert_eq!(payload["modelName"], "Seedream 5.0 Pro");
        assert_eq!(payload["type"], "image");
        let e = accept_feedback_expression("shenlangDoubaoTask_abc", &payload, false);
        assert!(e.starts_with("typeof shenlangDoubaoTask_abc === 'undefined' ? ({ stage: 'unknown', missing: true }) : Promise.resolve(shenlangDoubaoTask_abc.run(\"acceptFeedback\", {"));
        assert_eq!(classify(&e), "acceptFeedback");
        let e = accept_feedback_expression("shenlangDoubaoTask_abc", &payload, true);
        assert!(e.starts_with("if (typeof shenlangDoubaoTask_abc === 'undefined') {"));
    }

    /// 审查修复:reader 返回的 feedback 没有 stage,必须经 `acceptFeedback`(reply=消息对象)分类;
    /// auto + 自动用会员 → quota_checking 直接代回复。
    #[tokio::test]
    async fn history_feedback_is_classified_via_accept_feedback_with_reader_object() {
        let replied = Arc::new(AtomicBool::new(false));
        let r2 = Arc::clone(&replied);
        let h = handler(move |kind, expr| match kind {
            "terminalFeedback" => ok(json!({ "stage": "unknown" })),
            "acceptFeedback" => {
                assert!(expr.contains(r#""boundSubmissionId":"9002""#), "{expr}");
                assert!(expr.contains(r#""reply":{"#), "reply 必须是对象: {expr}");
                assert!(expr.contains(r#""messageId":"m-1""#), "{expr}");
                assert!(
                    expr.contains(r#""scope":"after_submitted_message""#),
                    "{expr}"
                );
                assert!(expr.contains(r#""modelName":"Seedream 5.0 Pro""#), "{expr}");
                accept_feedback_mock(expr)
            }
            "acceptReply" => {
                r2.store(true, Ordering::SeqCst);
                ok(json!({ "sent": true }))
            }
            "read" => {
                if r2.load(Ordering::SeqCst) {
                    ok(result_reply())
                } else {
                    ok(feedback_reply("m-1", QUOTA_TEXT))
                }
            }
            _ => ok(json!({})),
        });
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let o = opts(); // auto + auto_use_membership
        let since = o.send_since;
        mock.push_later(
            Duration::from_millis(30),
            "Network.requestWillBeSent",
            history_event("req-c1", since + 5),
        );
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let (r, events) = tokio::time::timeout(
            Duration::from_secs(5),
            run_capture(&conn, o, &cancel, &mut confirm_rx),
        )
        .await
        .unwrap();
        assert_eq!(r.unwrap().originals.len(), 1);
        assert_eq!(mock.count("eval:acceptFeedback"), 1, "{:?}", mock.log());
        assert_eq!(mock.count("eval:acceptReply"), 1, "{:?}", mock.log());
        assert!(events.contains(&CaptureEvent::Stage(
            TaskStage::QuotaChecking,
            Some(QUOTA_TEXT.into())
        )));
        assert!(!events.contains(&CaptureEvent::NeedsUserConfirm));
    }

    /// UC-20 真机:bot 的"视频生成参数确认"在流式输出时先被 `read` 读到半句(分类 `unknown`),
    /// 同一个 messageId 文本长全后必须**重新分类**;分类出 `checking` 后 auto 模式由 `videoFollowup`
    /// 代回复,payload 必须带 `modelName/ratio/duration/referenceFiles`(脚本 `confirmationMatches` 核对用)。
    #[tokio::test]
    async fn streaming_feedback_is_reclassified_and_video_auto_confirm_carries_params() {
        const PARTIAL: &str = "我将为你生成一条橘猫视频。";
        let full = format!(
            "{PARTIAL}视频生成参数确认 模型:Seedance 2.0 Fast 时长:5 秒 比例:16:9 {PARAMS_TEXT}"
        );
        let reads = Arc::new(std::sync::atomic::AtomicUsize::new(0));
        let classified = Arc::new(AtomicBool::new(false));
        let sent = Arc::new(AtomicBool::new(false));
        let (r1, c1, s1) = (
            Arc::clone(&reads),
            Arc::clone(&classified),
            Arc::clone(&sent),
        );
        let full2 = full.clone();
        let h = handler(move |kind, expr| match kind {
            "read" => {
                if s1.load(Ordering::SeqCst) {
                    return ok(video_read_reply(json!({ "images": [{
                        "url": "https://v11-vdl.doubao.com/x/?a=1&lr=unwatermarked", "unwatermarked": true, "vid": "v1"
                    }] })));
                }
                let n = r1.fetch_add(1, Ordering::SeqCst);
                let text = if n < 3 { PARTIAL } else { full2.as_str() };
                ok(json!({
                    "ok": true, "images": [],
                    "binding": { "conversationId": "7001", "submissionId": "9002", "createdAt": 1700000000001i64 },
                    "feedback": { "message": text, "messageId": "m-v", "submissionId": "9002",
                                  "createdAt": 1700000000002i64, "scope": "after_submitted_message" }
                }))
            }
            "acceptFeedback" => {
                let r = accept_feedback_mock(expr);
                if r.value["stage"] == "checking" {
                    c1.store(true, Ordering::SeqCst);
                }
                r
            }
            "videoFollowup" => {
                if !c1.load(Ordering::SeqCst) || s1.load(Ordering::SeqCst) {
                    return ok(json!({ "sent": false, "stage": "unknown" }));
                }
                // 真脚本:参数与确认文案不符(或缺失)→ manualRequired
                if expr.contains(r#""modelName":"Seedance 2.0 Fast""#)
                    && expr.contains(r#""duration":5"#)
                    && expr.contains(r#""ratio":"16:9""#)
                    && expr.contains(r#""referenceFiles":[]"#)
                {
                    s1.store(true, Ordering::SeqCst);
                    ok(json!({ "sent": true, "stage": "checking" }))
                } else {
                    ok(json!({ "sent": false, "stage": "checking", "manualRequired": true }))
                }
            }
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let mut o = opts();
        o.task_type = TaskType::Video;
        o.model_name = Some("Seedance 2.0 Fast".into());
        o.ratio = Some("16:9".into());
        o.duration = Some(5);
        o.deadline = Duration::from_secs(4);
        mock.push_later(
            Duration::from_millis(30),
            "Network.requestWillBeSent",
            history_event("req-sv", o.send_since + 5),
        );
        let (r, events) = run_capture_fetch(&conn, o, FakeFetcher::new(|_| Err("x".into()))).await;
        let out = r.unwrap();
        assert_eq!(out.watermark_source, WatermarkSource::LrUnwatermarked);
        assert!(
            sent.load(Ordering::SeqCst),
            "auto 模式应由 videoFollowup 代回复"
        );
        assert_eq!(
            mock.count("eval:acceptFeedback"),
            2,
            "半句与完整文本各分类一次:{:?}",
            mock.log()
        );
        assert!(events.contains(&CaptureEvent::Stage(
            TaskStage::Checking,
            Some(full.clone())
        )));
        assert!(!events.contains(&CaptureEvent::NeedsUserConfirm));
    }

    /// 真机:reader 自家重放的 initiator 栈只有 slardar 包装层,认不出;postData 与模板完全相同
    /// (同一个 sequence_id)即视为重放,不再取响应体重复 analyze。
    #[tokio::test]
    async fn replay_with_identical_post_data_is_not_reanalyzed() {
        let h = handler(|kind, _| match kind {
            "analyze" | "read" => {
                ok(json!({ "ok": true, "images": [], "feedback": null, "binding": null }))
            }
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let mut o = opts();
        o.deadline = Duration::from_millis(400);
        let since = o.send_since;
        let m = Arc::clone(&mock);
        tokio::spawn(async move {
            tokio::time::sleep(Duration::from_millis(40)).await;
            m.push(
                "Network.requestWillBeSent",
                history_event("req-page", since + 5),
            );
            m.push(
                "Network.loadingFinished",
                json!({ "requestId": "req-page" }),
            );
            tokio::time::sleep(Duration::from_millis(60)).await;
            // slardar 包装后的重放:initiator 只有一帧匿名函数
            let mut replay = history_event("req-replay", since + 900);
            replay["initiator"] = json!({ "type": "script", "stack": { "callFrames": [{ "functionName": "", "url": "chrome://doubao-chat/slardar/fe/sdk-web/" }] } });
            m.push("Network.requestWillBeSent", replay);
            m.push(
                "Network.loadingFinished",
                json!({ "requestId": "req-replay" }),
            );
        });
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let (r, _) = run_capture(&conn, o, &cancel, &mut confirm_rx).await;
        assert_eq!(r.unwrap_err().code, "generation_task_timeout");
        assert_eq!(
            mock.count("Network.getResponseBody req-page"),
            1,
            "{:?}",
            mock.log()
        );
        assert_eq!(
            mock.count("Network.getResponseBody req-replay"),
            0,
            "{:?}",
            mock.log()
        );
    }

    #[test]
    fn feedback_payload_carries_video_params() {
        let mut o = opts();
        o.task_type = TaskType::Video;
        o.model_name = Some("Seedance 2.0 Fast".into());
        o.ratio = Some("16:9".into());
        o.duration = Some(5);
        o.reference_count = 2;
        let p = feedback_payload(&o, ConfirmationMode::Auto, false, false);
        assert_eq!(p["modelName"], "Seedance 2.0 Fast");
        assert_eq!(p["ratio"], "16:9");
        assert_eq!(p["duration"], 5);
        assert_eq!(p["referenceFiles"].as_array().unwrap().len(), 2);
        assert_eq!(p["type"], "video");
        let k1 = classify_key("m1", &json!({ "message": "半句" }));
        let k2 = classify_key("m1", &json!({ "message": "半句完整了" }));
        assert_ne!(k1, k2);
        assert_eq!(
            k1,
            classify_key("m1", &json!({ "message": "半句", "createdAt": 9 }))
        );
    }

    /// 只出现在历史里的"额度已用完"经分类后映射为 account_quota_exhausted。
    #[tokio::test]
    async fn quota_exhausted_in_history_fails_with_account_quota_exhausted() {
        let h = handler(|kind, expr| match kind {
            "terminalFeedback" => ok(json!({ "stage": "unknown" })),
            "acceptFeedback" => accept_feedback_mock(expr),
            "read" => ok(feedback_reply("m-9", EXHAUSTED_TEXT)),
            _ => ok(json!({})),
        });
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let o = opts();
        let since = o.send_since;
        mock.push_later(
            Duration::from_millis(30),
            "Network.requestWillBeSent",
            history_event("req-c2", since + 5),
        );
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let (r, events) = tokio::time::timeout(
            Duration::from_secs(5),
            run_capture(&conn, o, &cancel, &mut confirm_rx),
        )
        .await
        .unwrap();
        let e = r.unwrap_err();
        assert_eq!(e.code, "account_quota_exhausted");
        assert_eq!(e.details.unwrap()["doubao_message"], EXHAUSTED_TEXT);
        assert!(events.contains(&CaptureEvent::Refused {
            kind: "quota_exhausted".into(),
            message: EXHAUSTED_TEXT.into()
        }));
    }

    /// 未绑定 submission 时历史 feedback 不分类(PS 同样要求 boundSubmissionId)。
    #[tokio::test]
    async fn unbound_history_feedback_is_not_classified() {
        let h = handler(|kind, expr| match kind {
            "acceptFeedback" => accept_feedback_mock(expr),
            "read" => {
                let mut v = feedback_reply("m-3", QUOTA_TEXT);
                v["binding"] = Value::Null;
                ok(v)
            }
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = Arc::new(MockDoubao::start(h).await);
        let conn = mock.conn().await;
        let mut o = opts();
        o.deadline = Duration::from_millis(250);
        let since = o.send_since;
        mock.push_later(
            Duration::from_millis(30),
            "Network.requestWillBeSent",
            history_event("req-c3", since + 5),
        );
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let (r, _) = run_capture(&conn, o, &cancel, &mut confirm_rx).await;
        assert_eq!(r.unwrap_err().code, "generation_task_timeout");
        assert!(mock.count("eval:read") >= 1, "{:?}", mock.log());
        assert_eq!(mock.count("eval:acceptFeedback"), 0, "{:?}", mock.log());
    }

    const SSE_BINDING_CHUNK: &str = "{\"local_message_id\":\"lm-123456\",\"message_id\":\"9009\"}";
    const SSE_IMAGE_BODY: &str = "data: {\"local_message_id\":\"lm-123456\",\"image_ori_raw\":{\"url\":\"https://p.byteimg.com/late\"}}\n";

    /// 审查修复:生成请求的 `requestWillBeSent` 在 submit 里(capture 订阅之前)就被消费了;
    /// 靠 `CaptureOptions.generation_requests` 预置后,之后才到的 SSE / `loadingFinished` 仍能绑定并取原图。
    #[tokio::test]
    async fn preloaded_generation_request_receives_later_sse_and_body() {
        let h = handler(|kind, expr| match kind {
            "streamSubmissionBinding" => {
                assert!(expr.contains("9009"), "{expr}");
                ok(json!({ "conversationId": "7001", "submissionId": "9009", "createdAt": 1 }))
            }
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let chat_target = conn.chat_session().await.unwrap().target_id.clone();
        let mut o = opts();
        o.generation_requests = vec![ObservedRequest {
            target_id: chat_target,
            request_id: "req-g2".into(),
        }];
        mock.set_body("req-g2", SSE_IMAGE_BODY);
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let mock = Arc::new(mock);
        let m = Arc::clone(&mock);
        tokio::spawn(async move {
            // 注意:不推 requestWillBeSent(它发生在 capture 之前)
            tokio::time::sleep(Duration::from_millis(40)).await;
            m.push(
                "Network.eventSourceMessageReceived",
                json!({ "requestId": "req-g2", "data": SSE_BINDING_CHUNK }),
            );
            tokio::time::sleep(Duration::from_millis(120)).await;
            m.push("Network.loadingFinished", json!({ "requestId": "req-g2" }));
        });
        let (r, events) = tokio::time::timeout(
            Duration::from_secs(5),
            run_capture(&conn, o, &cancel, &mut confirm_rx),
        )
        .await
        .unwrap();
        let out = r.unwrap();
        assert_eq!(
            out.originals,
            vec!["https://p.byteimg.com/late".to_string()]
        );
        assert_eq!(out.submission_id.as_deref(), Some("9009"));
        assert!(events.contains(&CaptureEvent::Bound {
            conversation_id: "7001".into(),
            submission_id: "9009".into()
        }));
        assert!(mock
            .log()
            .iter()
            .any(|l| l == "Network.getResponseBody req-g2"));
    }

    /// 真机 2.32.6 回归:绑定前页面为侧栏里**别的对话**拉过 `/im/chain/single`,被记成模板;
    /// 绑定后必须丢掉它,等绑定对话的模板,否则一直重放别的对话。
    #[tokio::test]
    async fn stale_history_template_of_other_conversation_is_replaced_after_binding() {
        let reads = Arc::new(StdMutex::new(Vec::<String>::new()));
        let seen = Arc::clone(&reads);
        let h = handler(move |kind, expr| match kind {
            "streamSubmissionBinding" => {
                ok(json!({ "conversationId": "7001", "submissionId": "9009", "createdAt": 1 }))
            }
            "read" => {
                let other = expr.contains("8888");
                seen.lock()
                    .unwrap()
                    .push(if other { "8888" } else { "7001" }.into());
                if other {
                    ok(json!({ "ok": true, "images": [], "feedback": null, "binding": null }))
                } else {
                    ok(result_reply())
                }
            }
            _ => ok(json!({ "stage": "unknown" })),
        });
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let chat_target = conn.chat_session().await.unwrap().target_id.clone();
        let mut o = opts();
        o.identity.conversation_id = None;
        o.generation_requests = vec![ObservedRequest {
            target_id: chat_target,
            request_id: "req-g8".into(),
        }];
        let since = o.send_since;
        let mock = Arc::new(mock);
        let m = Arc::clone(&mock);
        tokio::spawn(async move {
            let mut ev = history_event("req-other", since + 5);
            ev["request"]["postData"] = json!(r#"{"conversation_id":"8888","cursor":0}"#);
            m.push("Network.requestWillBeSent", ev);
            tokio::time::sleep(Duration::from_millis(40)).await;
            m.push(
                "Network.eventSourceMessageReceived",
                json!({ "requestId": "req-g8", "data": SSE_BINDING_CHUNK }),
            );
            // 绑定后页面再为本对话(7001)拉一次历史
            tokio::time::sleep(Duration::from_millis(80)).await;
            m.push(
                "Network.requestWillBeSent",
                history_event("req-mine", since + 50),
            );
        });
        let timing = fast();
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let (tx, mut rx) = mpsc::channel(256);
        tokio::spawn(async move { while rx.recv().await.is_some() {} });
        let r = tokio::time::timeout(
            Duration::from_secs(5),
            capture_with_fetcher(
                &conn,
                o,
                &cancel,
                &mut confirm_rx,
                tx,
                timing,
                FakeFetcher::new(|_| Err("offline".into())),
            ),
        )
        .await
        .unwrap();
        assert_eq!(
            r.unwrap().originals,
            vec!["https://x.byteimg.com/a".to_string()]
        );
        assert!(
            reads.lock().unwrap().iter().any(|c| c == "7001"),
            "{:?}",
            reads.lock().unwrap()
        );
    }

    /// 对照:不预置时,capture 之前发出的生成请求的 SSE / body 一律被丢弃(这正是修复前的缺口)。
    #[tokio::test]
    async fn generation_request_not_preloaded_is_ignored() {
        let h = handler(|_, _| ok(json!({ "stage": "unknown" })));
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let mut o = opts();
        o.deadline = Duration::from_millis(250);
        mock.set_body("req-g3", SSE_IMAGE_BODY);
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let mock = Arc::new(mock);
        mock.push_later(
            Duration::from_millis(40),
            "Network.loadingFinished",
            json!({ "requestId": "req-g3" }),
        );
        let (r, _) = run_capture(&conn, o, &cancel, &mut confirm_rx).await;
        assert_eq!(r.unwrap_err().code, "generation_task_timeout");
        assert!(!mock.log().iter().any(|l| l.contains("req-g3")));
    }

    /// 审查修复(minor):多图分几次到达时,等静默窗口过去再返回,不在第一张就收工。
    #[tokio::test]
    async fn multi_image_results_wait_for_settle_window() {
        let h = handler(|_, _| ok(json!({ "stage": "unknown" })));
        let mock = MockDoubao::start(h).await;
        let conn = mock.conn().await;
        let chat_target = conn.chat_session().await.unwrap().target_id.clone();
        let mut o = opts();
        o.generation_requests = vec![ObservedRequest {
            target_id: chat_target,
            request_id: "req-g4".into(),
        }];
        let cancel = CancellationToken::new();
        let (_tx, mut confirm_rx) = watch::channel(false);
        let mock = Arc::new(mock);
        let chunk = |u: &str| {
            json!({ "requestId": "req-g4",
                "data": format!("{{\"local_message_id\":\"lm-123456\",\"image_ori_raw\":{{\"url\":\"{u}\"}}}}") })
        };
        mock.push_later(
            Duration::from_millis(40),
            "Network.eventSourceMessageReceived",
            chunk("https://p.byteimg.com/1"),
        );
        mock.push_later(
            Duration::from_millis(100),
            "Network.eventSourceMessageReceived",
            chunk("https://p.byteimg.com/2"),
        );
        let timing = CaptureTiming {
            result_settle: Duration::from_millis(150),
            result_settle_max: Duration::from_secs(2),
            ..fast()
        };
        let (tx, mut rx) = mpsc::channel(256);
        let collector = tokio::spawn(async move { while rx.recv().await.is_some() {} });
        let out = tokio::time::timeout(
            Duration::from_secs(5),
            capture_with(&conn, o, &cancel, &mut confirm_rx, tx, timing),
        )
        .await
        .unwrap()
        .unwrap();
        collector.await.unwrap();
        assert_eq!(
            out.originals,
            vec![
                "https://p.byteimg.com/1".to_string(),
                "https://p.byteimg.com/2".to_string()
            ]
        );
    }
}
