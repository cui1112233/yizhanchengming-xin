use anyhow::{anyhow, Context};
use base64::Engine;
use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::{path::Path, sync::Arc, time::Duration};

use crate::{
    state::AppState,
    tasks::{ConfirmationMode, GenerateRequest, TaskStatus, TaskType},
};

pub struct V88ExecutorClient {
    base_url: String,
    http: reqwest::Client,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct PairRequest {
    pub code: String,
    #[serde(rename = "deviceName")]
    pub device_name: String,
    pub platform: String,
    pub os: String,
    pub version: String,
}

#[derive(Debug, Deserialize)]
pub struct PairResult {
    #[serde(rename = "executorId")]
    pub executor_id: String,
    pub token: String,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct HeartbeatRequest {
    pub device_name: String,
    pub os: String,
    pub version: String,
    pub accounts: AccountStats,
}

#[derive(Debug, Serialize)]
#[serde(rename_all = "camelCase")]
pub struct AccountStats {
    pub total: usize,
    pub available: usize,
    pub busy: usize,
    pub quota_exhausted: usize,
    pub login_error: usize,
    pub human_verification: usize,
}

#[derive(Debug, Clone, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct V88Job {
    pub id: String,
    #[serde(default)]
    pub source_task_id: String,
    #[serde(default)]
    pub payload: Value,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ClaimResult {
    pub job: V88Job,
    pub lease_token: String,
    pub lease_generation: i64,
}

#[derive(Debug, Clone)]
pub struct Lease {
    pub token: String,
    pub generation: i64,
}

#[derive(Debug, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct ArtifactResult {
    #[serde(alias = "artifactId")]
    pub id: String,
    pub media_type: String,
}

impl HeartbeatRequest {
    pub fn ready(device_name: &str, version: &str, account_count: usize) -> Self {
        Self {
            device_name: device_name.into(),
            os: "windows".into(),
            version: version.into(),
            accounts: AccountStats {
                total: account_count,
                available: account_count,
                busy: 0,
                quota_exhausted: 0,
                login_error: 0,
                human_verification: 0,
            },
        }
    }
}

impl V88Job {
    pub fn reference_images(&self) -> anyhow::Result<Vec<String>> {
        let object = self
            .payload
            .as_object()
            .ok_or_else(|| anyhow!("V88 任务参数格式错误"))?;
        let references = ["referenceImages", "imageUrls", "image_urls"]
            .into_iter()
            .find_map(|key| object.get(key).filter(|value| !value.is_null()));
        match references {
            None | Some(Value::Null) => Ok(Vec::new()),
            Some(Value::Array(items)) => items
                .iter()
                .map(|item| {
                    item.as_str()
                        .map(str::trim)
                        .filter(|url| !url.is_empty())
                        .map(str::to_string)
                        .ok_or_else(|| anyhow!("V88 参考图地址无效"))
                })
                .collect(),
            Some(_) => Err(anyhow!("V88 参考图地址格式错误")),
        }
    }

    pub fn to_generate_request(&self, account: &str) -> anyhow::Result<GenerateRequest> {
        let object = self
            .payload
            .as_object()
            .ok_or_else(|| anyhow!("V88 任务参数格式错误"))?;
        let prompt = value_string(object.get("prompt"));
        if prompt.is_empty() {
            return Err(anyhow!("V88 任务缺少提示词"));
        }
        let declared_type = value_string(object.get("type"));
        let task_type = if declared_type == "image" {
            TaskType::Image
        } else if declared_type == "video"
            || object.contains_key("duration")
            || self.source_task_id.starts_with("script-video:")
            || self.source_task_id.starts_with("bf11:")
        {
            TaskType::Video
        } else {
            TaskType::Image
        };
        let model = value_string(object.get("model"));
        let model_name = match task_type {
            TaskType::Image => match model.as_str() {
                "doubao-image-4.0" | "Seedream 4.0" => "Seedream 4.0",
                "doubao-image-5.0-flash" | "Seedream 5.0 Flash" => "Seedream 5.0 Flash",
                "Seedream 5.0 Pro" => "Seedream 5.0 Pro",
                "doubao-image-4.5" | "Seedream 4.5" => "Seedream 4.5",
                _ => "Seedream 4.5",
            },
            TaskType::Video => match model.as_str() {
                "doubao-video-mini" | "Seedance 2.0 Mini" => "Seedance 2.0 Mini",
                "Seedance 2.0" => "Seedance 2.0",
                "Seedance 2.5" => "Seedance 2.5",
                "doubao-video-fast" | "Seedance 2.0 Fast" => "Seedance 2.0 Fast",
                _ => "Seedance 2.0 Fast",
            },
        };
        let ratio = value_string(object.get("ratio"));
        let ratio = if ratio.is_empty() {
            let aspect_ratio = value_string(object.get("aspectRatio"));
            if aspect_ratio.is_empty() && task_type == TaskType::Image {
                crate::http::routes::openai::size_to_ratio(
                    object.get("size").and_then(Value::as_str),
                )
            } else {
                aspect_ratio
            }
        } else {
            ratio
        };
        let ratio = if ratio.is_empty() {
            if task_type == TaskType::Image {
                "1:1"
            } else {
                "16:9"
            }
            .into()
        } else {
            ratio
        };
        let duration = if task_type == TaskType::Video {
            Some(value_u32(object.get("duration")).unwrap_or(5))
        } else {
            None
        };
        Ok(GenerateRequest {
            account: account.into(),
            task_type,
            model_name: model_name.into(),
            prompt,
            ratio,
            duration,
            reference_files: Vec::new(),
            confirmation_mode: ConfirmationMode::Auto,
            auto_use_membership: true,
            callback_url: None,
        })
    }
}

fn is_executor_reference_path(value: &str) -> bool {
    let value = value.trim();
    if value.contains(['?', '#', '\\']) || value.to_ascii_lowercase().contains("%2e") {
        return false;
    }
    [
        "/api/local-executor/v1/agent-reference-assets/",
        "/api/local-executor/v1/batch-comic-references/",
        "/api/local-executor/v1/batch-comic-style-previews/",
        "/api/local-executor/v1/batch-comic-reference-scenes/",
        "/api/local-executor/v1/batch-comic-materials/",
        "/api/local-executor/v1/batch-book-asset-images/",
    ]
    .iter()
    .any(|prefix| value.starts_with(prefix))
}

fn value_string(value: Option<&Value>) -> String {
    value
        .and_then(Value::as_str)
        .unwrap_or_default()
        .trim()
        .into()
}

fn value_u32(value: Option<&Value>) -> Option<u32> {
    value.and_then(|value| {
        value
            .as_u64()
            .and_then(|number| u32::try_from(number).ok())
            .or_else(|| value.as_str()?.trim().parse().ok())
    })
}

impl PairRequest {
    pub fn new(code: &str, device_name: &str, os: &str, version: &str) -> Self {
        Self {
            code: code.trim().into(),
            device_name: device_name.trim().into(),
            platform: "doubao".into(),
            os: os.into(),
            version: version.into(),
        }
    }
}

impl V88ExecutorClient {
    pub fn new(base_url: &str) -> anyhow::Result<Self> {
        let base_url = base_url.trim().trim_end_matches('/');
        let url = reqwest::Url::parse(base_url).context("V88 地址不合法")?;
        if !matches!(url.scheme(), "http" | "https") {
            return Err(anyhow!("V88 地址必须使用 http 或 https"));
        }
        Ok(Self {
            base_url: base_url.into(),
            http: reqwest::Client::new(),
        })
    }

    pub const fn pair_path() -> &'static str {
        "/api/local-executor/v1/pair"
    }

    pub const fn claim_path() -> &'static str {
        "/api/local-executor/v1/jobs/claim?waitSeconds=25"
    }

    pub async fn pair(&self, request: &PairRequest) -> anyhow::Result<PairResult> {
        let response = self
            .http
            .post(format!("{}{}", self.base_url, Self::pair_path()))
            .json(request)
            .send()
            .await
            .context("连接 V88 失败")?;
        let status = response.status();
        if !status.is_success() {
            return Err(anyhow!("V88 配对失败，HTTP {status}"));
        }
        let result: PairResult = response.json().await.context("V88 配对响应格式错误")?;
        if result.executor_id.is_empty() || result.token.is_empty() {
            return Err(anyhow!("V88 配对响应不完整"));
        }
        Ok(result)
    }

    pub async fn heartbeat(&self, token: &str, request: &HeartbeatRequest) -> anyhow::Result<()> {
        self.auth_json("/api/local-executor/v1/heartbeat", token, request)
            .await
            .map(|_| ())
    }

    pub async fn claim(&self, token: &str) -> anyhow::Result<Option<ClaimResult>> {
        let response = self
            .http
            .post(format!("{}{}", self.base_url, Self::claim_path()))
            .bearer_auth(token)
            .send()
            .await
            .context("领取 V88 任务失败")?;
        if response.status() == reqwest::StatusCode::NO_CONTENT {
            return Ok(None);
        }
        let response = ensure_success(response, "领取 V88 任务").await?;
        let claim: ClaimResult = response.json().await.context("V88 任务格式错误")?;
        if claim.job.id.is_empty() || claim.lease_token.is_empty() || claim.lease_generation < 1 {
            return Err(anyhow!("V88 任务信息不完整"));
        }
        Ok(Some(claim))
    }

    pub async fn renew(&self, token: &str, job_id: &str, lease: &Lease) -> anyhow::Result<()> {
        self.lease_action(token, job_id, "renew", lease, json!({}))
            .await
    }

    pub async fn progress(
        &self,
        token: &str,
        job_id: &str,
        lease: &Lease,
        state: &str,
    ) -> anyhow::Result<()> {
        self.lease_action(token, job_id, "progress", lease, json!({"state":state}))
            .await
    }

    pub async fn acceptance(
        &self,
        token: &str,
        job_id: &str,
        lease: &Lease,
        account_id: &str,
        submission_id: &str,
    ) -> anyhow::Result<()> {
        self.lease_action(
            token,
            job_id,
            "acceptance",
            lease,
            json!({"accountId":account_id,"submissionId":submission_id}),
        )
        .await
    }

    pub async fn fail(
        &self,
        token: &str,
        job_id: &str,
        lease: &Lease,
        code: &str,
        message: &str,
    ) -> anyhow::Result<()> {
        self.lease_action(
            token,
            job_id,
            "fail",
            lease,
            json!({"code":code,"message":message}),
        )
        .await
    }

    pub async fn upload_bytes(
        &self,
        token: &str,
        job_id: &str,
        lease: &Lease,
        media_type: &str,
        bytes: Vec<u8>,
    ) -> anyhow::Result<ArtifactResult> {
        let response = self
            .http
            .post(format!(
                "{}/api/local-executor/v1/jobs/{job_id}/artifact",
                self.base_url
            ))
            .bearer_auth(token)
            .header("Content-Type", media_type)
            .header("X-Lease-Token", &lease.token)
            .header("X-Lease-Generation", lease.generation)
            .body(bytes)
            .send()
            .await
            .context("上传 V88 生成结果失败")?;
        let response = ensure_success(response, "上传 V88 生成结果").await?;
        let result: ArtifactResult = response.json().await.context("V88 上传响应格式错误")?;
        if result.id.is_empty() {
            return Err(anyhow!("V88 上传响应不完整"));
        }
        Ok(result)
    }

    async fn download_reference_asset(
        &self,
        token: &str,
        reference_path: &str,
    ) -> anyhow::Result<Vec<u8>> {
        const MAX_REFERENCE_BYTES: usize = 50 << 20;

        let reference_path = reference_path.trim();
        if !is_executor_reference_path(reference_path) {
            return Err(anyhow!("V88 参考素材地址不合法"));
        }
        let mut response = self
            .http
            .get(format!("{}{}", self.base_url, reference_path))
            .bearer_auth(token)
            .send()
            .await
            .context("读取 V88 参考素材失败")?;
        if !response.status().is_success() {
            return Err(anyhow!("读取 V88 参考素材失败，HTTP {}", response.status()));
        }
        if response
            .content_length()
            .is_some_and(|size| size > MAX_REFERENCE_BYTES as u64)
        {
            return Err(anyhow!("V88 参考素材超过 50MB 上限"));
        }
        let mut bytes = Vec::new();
        while let Some(chunk) = response.chunk().await.context("读取 V88 参考素材失败")? {
            if bytes.len() + chunk.len() > MAX_REFERENCE_BYTES {
                return Err(anyhow!("V88 参考素材超过 50MB 上限"));
            }
            bytes.extend_from_slice(&chunk);
        }
        if bytes.is_empty() {
            return Err(anyhow!("V88 参考素材为空"));
        }
        Ok(bytes)
    }

    pub async fn result(
        &self,
        token: &str,
        job_id: &str,
        lease: &Lease,
        artifact_id: &str,
    ) -> anyhow::Result<()> {
        self.lease_action(
            token,
            job_id,
            "result",
            lease,
            json!({"artifactId":artifact_id}),
        )
        .await
    }

    async fn lease_action(
        &self,
        token: &str,
        job_id: &str,
        action: &str,
        lease: &Lease,
        extra: Value,
    ) -> anyhow::Result<()> {
        let mut body = extra.as_object().cloned().unwrap_or_default();
        body.insert("leaseToken".into(), Value::String(lease.token.clone()));
        body.insert("leaseGeneration".into(), json!(lease.generation));
        self.auth_json(
            &format!("/api/local-executor/v1/jobs/{job_id}/{action}"),
            token,
            &body,
        )
        .await
        .map(|_| ())
    }

    async fn auth_json<T: Serialize + ?Sized>(
        &self,
        path: &str,
        token: &str,
        body: &T,
    ) -> anyhow::Result<reqwest::Response> {
        let response = self
            .http
            .post(format!("{}{}", self.base_url, path))
            .bearer_auth(token)
            .json(body)
            .send()
            .await
            .context("连接 V88 失败")?;
        ensure_success(response, "V88 请求").await
    }
}

async fn ensure_success(
    response: reqwest::Response,
    action: &str,
) -> anyhow::Result<reqwest::Response> {
    if response.status().is_success() {
        return Ok(response);
    }
    let status = response.status();
    let message = response.text().await.unwrap_or_default();
    Err(anyhow!("{action}失败，HTTP {status}: {message}"))
}

fn active_v88_account(accounts: &crate::doubao::AccountList) -> Option<String> {
    let current = accounts.current.as_deref()?.trim();
    if current.is_empty() {
        return None;
    }
    accounts
        .accounts
        .iter()
        .find(|account| {
            account.is_current && !account.is_ambiguous && account.name.trim() == current
        })
        .map(|account| account.name.trim().to_string())
}

fn available_v88_accounts(accounts: &crate::doubao::AccountList) -> Vec<String> {
    accounts
        .accounts
        .iter()
        .filter(|account| !account.is_ambiguous)
        .map(|account| account.name.trim())
        .filter(|name| !name.is_empty())
        .map(str::to_string)
        .fold(Vec::new(), |mut names, name| {
            if !names.iter().any(|item| item == &name) {
                names.push(name);
            }
            names
        })
}

fn next_v88_account(accounts: &[String], last: Option<&str>) -> Option<String> {
    if accounts.is_empty() {
        return None;
    }
    let index = last
        .and_then(|name| accounts.iter().position(|item| item == name))
        .map(|index| (index + 1) % accounts.len())
        .unwrap_or(0);
    accounts.get(index).cloned()
}

pub fn spawn_v88_service(state: Arc<AppState>) {
    tokio::spawn(async move {
        let mut last_heartbeat = tokio::time::Instant::now() - Duration::from_secs(30);
        let mut consecutive_claim_failures = 0_u32;
        // Account discovery opens the native Doubao menu. Read the account
        // pool once while idle, then rotate only between tasks. Never switch
        // an account while a generation is in flight.
        let mut active_account: Option<String> = None;
        let mut account_pool: Vec<String> = Vec::new();
        let mut last_used_account: Option<String> = None;
        loop {
            if state.shutdown.is_cancelled() {
                return;
            }
            let config = state.config().await.v88_executor;
            if config.base_url.trim().is_empty() || config.executor_token.trim().is_empty() {
                sleep_or_shutdown(&state, Duration::from_secs(2)).await;
                continue;
            }
            let client = match V88ExecutorClient::new(&config.base_url) {
                Ok(client) => client,
                Err(error) => {
                    tracing::warn!(error = %error, "V88 地址无效，等待用户重新配对");
                    sleep_or_shutdown(&state, Duration::from_secs(5)).await;
                    continue;
                }
            };
            if account_pool.is_empty() {
                if let Ok(accounts) = state.doubao.accounts(false).await {
                    active_account = active_v88_account(&accounts);
                    account_pool = available_v88_accounts(&accounts);
                }
            }
            if last_heartbeat.elapsed() >= Duration::from_secs(15) {
                let device_name = std::env::var("COMPUTERNAME").unwrap_or_else(|_| "本机".into());
                let heartbeat = HeartbeatRequest::ready(
                    &device_name,
                    env!("CARGO_PKG_VERSION"),
                    account_pool.len(),
                );
                if let Err(error) = client.heartbeat(&config.executor_token, &heartbeat).await {
                    tracing::warn!(error = %error, "V88 心跳失败");
                }
                last_heartbeat = tokio::time::Instant::now();
            }
            if active_account.is_none() {
                sleep_or_shutdown(&state, Duration::from_secs(2)).await;
                continue;
            }
            match client.claim(&config.executor_token).await {
                Ok(Some(claim)) => {
                    consecutive_claim_failures = 0;
                    let target_account =
                        next_v88_account(&account_pool, last_used_account.as_deref())
                            .or_else(|| active_account.clone());
                    if let Some(account_name) = target_account {
                        // 切号发生在任务真正开始前；run_claim 会随后把页面标记为忙，
                        // 因此不会在生图中途切换账号。
                        let switch_result =
                            if active_account.as_deref() != Some(account_name.as_str()) {
                                state.doubao.switch_account(&account_name).await.map(|_| ())
                            } else {
                                Ok(())
                            };
                        if let Err(error) = switch_result {
                            let message = format!("自动切换豆包账号失败: {error}");
                            let lease = Lease {
                                token: claim.lease_token.clone(),
                                generation: claim.lease_generation,
                            };
                            let _ = client
                                .fail(
                                    &config.executor_token,
                                    &claim.job.id,
                                    &lease,
                                    "ACCOUNT_SWITCH_FAILED",
                                    &message,
                                )
                                .await;
                            tracing::warn!(error = %error, account = %account_name, "V88 自动切换账号失败");
                        } else {
                            active_account = Some(account_name.clone());
                            last_used_account = Some(account_name.clone());
                            if let Err(error) = run_claim(
                                &state,
                                &client,
                                &config.executor_token,
                                &account_name,
                                claim,
                            )
                            .await
                            {
                                tracing::warn!(error = %error, account = %account_name, "V88 任务执行失败");
                            }
                        }
                    } else {
                        tracing::warn!("没有可用的豆包账号，任务暂不执行");
                    }
                }
                Ok(None) => consecutive_claim_failures = 0,
                Err(error) => {
                    consecutive_claim_failures = consecutive_claim_failures.saturating_add(1);
                    tracing::warn!(error = %error, "领取 V88 任务失败");
                    sleep_or_shutdown(&state, retry_delay(consecutive_claim_failures)).await;
                }
            }
        }
    });
}

fn retry_delay(failures: u32) -> Duration {
    Duration::from_secs(match failures {
        0 | 1 => 2,
        2 => 5,
        3 => 10,
        _ => 30,
    })
}

async fn run_claim(
    state: &Arc<AppState>,
    client: &V88ExecutorClient,
    executor_token: &str,
    account_name: &str,
    claim: ClaimResult,
) -> anyhow::Result<()> {
    let job_id = claim.job.id.clone();
    let lease = Lease {
        token: claim.lease_token,
        generation: claim.lease_generation,
    };
    let outcome = run_claim_inner(
        state,
        client,
        executor_token,
        account_name,
        &claim.job,
        &lease,
    )
    .await;
    if let Err(error) = &outcome {
        let message = format!("{error:#}");
        let _ = client
            .fail(
                executor_token,
                &job_id,
                &lease,
                "RELAY_TASK_FAILED",
                &message,
            )
            .await;
    }
    outcome
}

async fn run_claim_inner(
    state: &Arc<AppState>,
    client: &V88ExecutorClient,
    executor_token: &str,
    account_name: &str,
    job: &V88Job,
    lease: &Lease,
) -> anyhow::Result<()> {
    client
        .progress(executor_token, &job.id, lease, "preparing")
        .await?;
    let mut request = job.to_generate_request(account_name)?;
    let references = job.reference_images()?;
    if !references.is_empty() {
        let mut resolved_references = Vec::with_capacity(references.len());
        for reference in references {
            if is_executor_reference_path(&reference) {
                let bytes = client
                    .download_reference_asset(executor_token, &reference)
                    .await?;
                resolved_references.push(format!(
                    "data:application/octet-stream;base64,{}",
                    base64::engine::general_purpose::STANDARD.encode(bytes)
                ));
            } else {
                resolved_references.push(reference);
            }
        }
        request.reference_files =
            crate::http::routes::openai::resolve_ref_images(state, &resolved_references).await?;
    }
    client
        .progress(executor_token, &job.id, lease, "submitting")
        .await?;
    // V88 已用启动时确认的账号领取任务。任务从入队校验到结果回传期间都把
    // chat 页面标记为忙，让调度器沿用账号缓存，避免缓存过期后打开账号菜单
    // 干扰紧接着的模型、比例和风格菜单。
    let _page_busy = state.doubao.mark_page_busy();
    let (relay_task_id, _, _) = state
        .submit_task(request)
        .await
        .map_err(|error| anyhow!(error.message))?;
    client
        .acceptance(executor_token, &job.id, lease, account_name, &relay_task_id)
        .await?;
    client
        .progress(executor_token, &job.id, lease, "generating")
        .await?;

    let mut renew = tokio::time::interval(Duration::from_secs(15));
    renew.set_missed_tick_behavior(tokio::time::MissedTickBehavior::Delay);
    let task = loop {
        tokio::select! {
            _ = state.shutdown.cancelled() => return Err(anyhow!("执行器正在退出")),
            _ = renew.tick() => client.renew(executor_token, &job.id, lease).await?,
            _ = tokio::time::sleep(Duration::from_secs(2)) => {
                let task = state.scheduler.get_task(&relay_task_id).await
                    .map_err(|error| anyhow!(error.message))?
                    .ok_or_else(|| anyhow!("Relay 任务不存在"))?;
                match task.status {
                    TaskStatus::Completed => break task,
                    TaskStatus::Failed => {
                        let message = task.error.map(|error| error.message).unwrap_or_else(|| "本地生成失败".into());
                        return Err(anyhow!(message));
                    }
                    TaskStatus::Cancelled => return Err(anyhow!("本地生成任务已取消")),
                    _ => {}
                }
            }
        }
    };
    let results = task
        .result_files
        .filter(|files| !files.is_empty())
        .ok_or_else(|| anyhow!("本地生成完成但没有结果文件"))?;
    client
        .progress(executor_token, &job.id, lease, "downloading")
        .await?;
    let mut artifact_ids = Vec::with_capacity(results.len());
    for result in results {
        let path = state.storage.root.join(Path::new(&result.relative_path));
        let bytes = tokio::fs::read(&path)
            .await
            .with_context(|| format!("读取生成结果失败: {}", path.display()))?;
        let artifact = client
            .upload_bytes(executor_token, &job.id, lease, &result.mime, bytes)
            .await?;
        artifact_ids.push(artifact.id);
    }
    let artifact_id = artifact_ids
        .first()
        .ok_or_else(|| anyhow!("本地生成完成但没有结果文件"))?;
    client
        .result(executor_token, &job.id, lease, artifact_id)
        .await
}

async fn sleep_or_shutdown(state: &AppState, duration: Duration) {
    tokio::select! {
        _ = state.shutdown.cancelled() => {}
        _ = tokio::time::sleep(duration) => {}
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn v88_account_pool_rotates_round_robin() {
        let accounts = vec!["甲".to_string(), "乙".to_string(), "丙".to_string()];
        assert_eq!(next_v88_account(&accounts, None).as_deref(), Some("甲"));
        assert_eq!(
            next_v88_account(&accounts, Some("甲")).as_deref(),
            Some("乙")
        );
        assert_eq!(
            next_v88_account(&accounts, Some("丙")).as_deref(),
            Some("甲")
        );
        assert_eq!(
            next_v88_account(&accounts, Some("不存在")).as_deref(),
            Some("甲")
        );
    }
    use axum::{
        body::Bytes,
        extract::State,
        http::{HeaderMap, StatusCode},
        response::IntoResponse,
        routing::{get, post},
        Json, Router,
    };
    use std::sync::{Arc, Mutex};

    #[test]
    fn claim_uses_long_wait_and_errors_back_off() {
        assert!(V88ExecutorClient::claim_path().contains("waitSeconds=25"));
        assert_eq!(retry_delay(1), Duration::from_secs(2));
        assert_eq!(retry_delay(2), Duration::from_secs(5));
        assert_eq!(retry_delay(3), Duration::from_secs(10));
        assert_eq!(retry_delay(9), Duration::from_secs(30));
    }

    #[test]
    fn active_v88_account_uses_only_a_confirmed_current_account() {
        let accounts = crate::doubao::AccountList {
            current: Some("当前账号".into()),
            accounts: vec![
                crate::doubao::Account {
                    name: "当前账号".into(),
                    subscription: None,
                    label: "当前账号".into(),
                    is_current: true,
                    is_ambiguous: false,
                },
                crate::doubao::Account {
                    name: "同名账号".into(),
                    subscription: None,
                    label: "同名账号 A".into(),
                    is_current: false,
                    is_ambiguous: true,
                },
            ],
            read_at: 1,
        };

        assert_eq!(active_v88_account(&accounts).as_deref(), Some("当前账号"));
        let mut missing = accounts.clone();
        missing.current = Some("不存在".into());
        assert_eq!(active_v88_account(&missing), None);
    }

    type HeartbeatCapture = Arc<Mutex<Option<(HeaderMap, serde_json::Value)>>>;
    type RequestCapture = Arc<Mutex<Vec<(String, HeaderMap, Bytes)>>>;

    #[test]
    fn pair_request_targets_v88_pair_endpoint() {
        let request = PairRequest::new("ABC123", "电脑", "windows", "1.0.0");
        assert_eq!(request.code, "ABC123");
        assert_eq!(
            V88ExecutorClient::pair_path(),
            "/api/local-executor/v1/pair"
        );
    }

    #[tokio::test]
    async fn pair_posts_code_to_v88_and_returns_executor_credentials() {
        let received = Arc::new(Mutex::new(None));
        let app = Router::new()
            .route(
                V88ExecutorClient::pair_path(),
                post(
                    |State(received): State<Arc<Mutex<Option<PairRequest>>>>,
                     Json(request): Json<PairRequest>| async move {
                        *received.lock().unwrap() = Some(request);
                        Json(serde_json::json!({"executorId":"executor_01","token":"v88-token"}))
                    },
                ),
            )
            .with_state(Arc::clone(&received));
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let base_url = format!("http://{}", listener.local_addr().unwrap());
        tokio::spawn(async move { axum::serve(listener, app).await.unwrap() });

        let result = V88ExecutorClient::new(&base_url)
            .unwrap()
            .pair(&PairRequest::new("PAIR-123", "设备", "windows", "1.0.0"))
            .await
            .unwrap();
        assert_eq!(result.executor_id, "executor_01");
        assert_eq!(result.token, "v88-token");
        assert_eq!(received.lock().unwrap().as_ref().unwrap().code, "PAIR-123");
    }

    #[tokio::test]
    async fn heartbeat_uses_executor_bearer_and_reports_account_capacity() {
        let received = Arc::new(Mutex::new(None));
        let app = Router::new()
            .route(
                "/api/local-executor/v1/heartbeat",
                post(
                    |State(received): State<HeartbeatCapture>,
                     headers: HeaderMap,
                     Json(body): Json<serde_json::Value>| async move {
                        *received.lock().unwrap() = Some((headers, body));
                        Json(serde_json::json!({"ok":true,"heartbeatIntervalSeconds":15}))
                    },
                ),
            )
            .with_state(Arc::clone(&received));
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let base_url = format!("http://{}", listener.local_addr().unwrap());
        tokio::spawn(async move { axum::serve(listener, app).await.unwrap() });

        V88ExecutorClient::new(&base_url)
            .unwrap()
            .heartbeat(
                "executor-token",
                &HeartbeatRequest::ready("电脑", "1.1.0", 2),
            )
            .await
            .unwrap();

        let guard = received.lock().unwrap();
        let (headers, body) = guard.as_ref().unwrap();
        assert_eq!(headers["authorization"], "Bearer executor-token");
        assert_eq!(body["accounts"]["total"], 2);
        assert_eq!(body["accounts"]["available"], 2);
    }

    #[tokio::test]
    async fn claim_treats_204_as_empty_and_parses_a_job() {
        let calls = Arc::new(Mutex::new(0usize));
        let app = Router::new()
            .route(
                "/api/local-executor/v1/jobs/claim",
                post(|State(calls): State<Arc<Mutex<usize>>>| async move {
                    let mut count = calls.lock().unwrap();
                    *count += 1;
                    if *count == 1 {
                        return StatusCode::NO_CONTENT.into_response();
                    }
                    (
                        StatusCode::OK,
                        Json(serde_json::json!({
                            "job": {"id":"job-1","sourceTaskId":"settings-test:image","payload":{"type":"image","prompt":"蓝色小鸟","ratio":"1:1","model":"doubao-image-4.5"}},
                            "leaseToken":"lease-1","leaseGeneration":2
                        })),
                    )
                        .into_response()
                }),
            )
            .with_state(calls);
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let base_url = format!("http://{}", listener.local_addr().unwrap());
        tokio::spawn(async move { axum::serve(listener, app).await.unwrap() });
        let client = V88ExecutorClient::new(&base_url).unwrap();

        assert!(client.claim("token").await.unwrap().is_none());
        let claim = client.claim("token").await.unwrap().unwrap();
        assert_eq!(claim.job.id, "job-1");
        assert_eq!(claim.lease_generation, 2);
    }

    #[test]
    fn v88_payload_maps_to_relay_image_and_video_requests() {
        let image = V88Job {
            id: "image-job".into(),
            source_task_id: "settings-test:image".into(),
            payload: serde_json::json!({"type":"image","prompt":"蓝色小鸟","ratio":"1:1","model":"doubao-image-4.5"}),
        };
        let image_request = image.to_generate_request("当前账号").unwrap();
        assert_eq!(image_request.task_type, crate::tasks::TaskType::Image);
        assert_eq!(image_request.model_name, "Seedream 4.5");
        assert_eq!(image_request.duration, None);

        let referenced_image = V88Job {
            id: "referenced-image-job".into(),
            source_task_id: "agent:image".into(),
            payload: serde_json::json!({
                "type":"image",
                "prompt":"沿用参考图风格",
                "referenceImages":["https://assets.example/reference.png?signature=test"]
            }),
        };
        assert_eq!(
            referenced_image.reference_images().unwrap(),
            vec!["https://assets.example/reference.png?signature=test"]
        );

        let video = V88Job {
            id: "video-job".into(),
            source_task_id: "settings-test:video".into(),
            payload: serde_json::json!({"type":"video","prompt":"小鸟轻轻扇动翅膀","aspectRatio":"16:9","duration":5,"model":"doubao-video-fast"}),
        };
        let video_request = video.to_generate_request("当前账号").unwrap();
        assert_eq!(video_request.task_type, crate::tasks::TaskType::Video);
        assert_eq!(video_request.model_name, "Seedance 2.0 Fast");
        assert_eq!(video_request.duration, Some(5));

        for model in [
            "Seedream 5.0 Flash",
            "Seedream 5.0 Pro",
            "Seedream 4.5",
            "Seedream 4.0",
        ] {
            let job = V88Job {
                id: format!("image-{model}"),
                source_task_id: "batch-factory:image".into(),
                payload: serde_json::json!({"type":"image","prompt":"测试","ratio":"1:1","model":model}),
            };
            assert_eq!(
                job.to_generate_request("当前账号").unwrap().model_name,
                model
            );
        }
        for model in [
            "Seedance 2.5",
            "Seedance 2.0",
            "Seedance 2.0 Fast",
            "Seedance 2.0 Mini",
        ] {
            let job = V88Job {
                id: format!("video-{model}"),
                source_task_id: "bf11:batch:book:video".into(),
                payload: serde_json::json!({"type":"video","prompt":"测试","aspectRatio":"16:9","duration":5,"model":model}),
            };
            assert_eq!(
                job.to_generate_request("当前账号").unwrap().model_name,
                model
            );
        }
    }

    #[test]
    fn v88_reference_image_aliases_share_one_relay_contract() {
        for payload in [
            serde_json::json!({"referenceImages": ["https://assets.example/a.png"]}),
            serde_json::json!({"imageUrls": ["https://assets.example/a.png"]}),
            serde_json::json!({"image_urls": ["https://assets.example/a.png"]}),
        ] {
            let job = V88Job {
                id: "reference-alias-job".into(),
                source_task_id: "batch-factory:image".into(),
                payload,
            };
            assert_eq!(
                job.reference_images().unwrap(),
                vec!["https://assets.example/a.png"]
            );
        }
    }

    #[test]
    fn v88_image_size_maps_to_a_doubao_ratio() {
        let job = V88Job {
            id: "image-size-job".into(),
            source_task_id: "batch-factory:image".into(),
            payload: serde_json::json!({
                "type": "image",
                "prompt": "测试",
                "size": "1024x1792",
                "model": "doubao-image-4.5"
            }),
        };

        assert_eq!(job.to_generate_request("当前账号").unwrap().ratio, "9:16");
    }

    #[tokio::test]
    async fn upload_and_result_send_lease_without_exposing_local_api_token() {
        let received = Arc::new(Mutex::new(Vec::<(String, HeaderMap, Bytes)>::new()));
        let app =
            Router::new()
                .route(
                    "/api/local-executor/v1/jobs/job-1/artifact",
                    post(
                        |State(received): State<RequestCapture>,
                         headers: HeaderMap,
                         body: Bytes| async move {
                            received
                                .lock()
                                .unwrap()
                                .push(("artifact".into(), headers, body));
                            Json(serde_json::json!({"id":"artifact-1","mediaType":"image/png"}))
                        },
                    ),
                )
                .route(
                    "/api/local-executor/v1/jobs/job-1/result",
                    post(
                        |State(received): State<RequestCapture>,
                         headers: HeaderMap,
                         body: Bytes| async move {
                            received
                                .lock()
                                .unwrap()
                                .push(("result".into(), headers, body));
                            Json(serde_json::json!({"ok":true}))
                        },
                    ),
                )
                .with_state(Arc::clone(&received));
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let base_url = format!("http://{}", listener.local_addr().unwrap());
        tokio::spawn(async move { axum::serve(listener, app).await.unwrap() });
        let client = V88ExecutorClient::new(&base_url).unwrap();
        let lease = Lease {
            token: "lease-token".into(),
            generation: 3,
        };

        let artifact = client
            .upload_bytes(
                "executor-token",
                "job-1",
                &lease,
                "image/png",
                b"png-data".to_vec(),
            )
            .await
            .unwrap();
        client
            .result("executor-token", "job-1", &lease, &artifact.id)
            .await
            .unwrap();

        let guard = received.lock().unwrap();
        assert_eq!(guard.len(), 2);
        assert_eq!(guard[0].1["authorization"], "Bearer executor-token");
        assert_eq!(guard[0].1["x-lease-token"], "lease-token");
        assert_eq!(guard[0].1["content-type"], "image/png");
        assert!(!String::from_utf8_lossy(&guard[1].2).contains("local-api"));
    }

    #[tokio::test]
    async fn download_reference_asset_uses_executor_authorization() {
        let reference = get(|headers: HeaderMap| async move {
            if headers
                .get("authorization")
                .and_then(|value| value.to_str().ok())
                != Some("Bearer executor-token")
            {
                return (StatusCode::UNAUTHORIZED, Vec::new());
            }
            (StatusCode::OK, b"reference-image".to_vec())
        });
        let app = Router::new()
            .route(
                "/api/local-executor/v1/agent-reference-assets/project-1/asset-1",
                reference.clone(),
            )
            .route(
                "/api/local-executor/v1/batch-comic-reference-scenes/scene-1",
                reference,
            );
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let base_url = format!("http://{}", listener.local_addr().unwrap());
        tokio::spawn(async move { axum::serve(listener, app).await.unwrap() });
        let client = V88ExecutorClient::new(&base_url).unwrap();

        let content = client
            .download_reference_asset(
                "executor-token",
                "/api/local-executor/v1/agent-reference-assets/project-1/asset-1",
            )
            .await
            .unwrap();

        assert_eq!(content, b"reference-image");

        let comic_content = client
            .download_reference_asset(
                "executor-token",
                "/api/local-executor/v1/batch-comic-reference-scenes/scene-1",
            )
            .await
            .unwrap();
        assert_eq!(comic_content, b"reference-image");
    }

    #[test]
    fn executor_reference_path_accepts_comic_assets_but_rejects_other_api_routes() {
        for path in [
            "/api/local-executor/v1/agent-reference-assets/project-1/asset-1",
            "/api/local-executor/v1/batch-comic-references/batch-1/asset-1",
            "/api/local-executor/v1/batch-comic-reference-scenes/scene-1",
            "/api/local-executor/v1/batch-comic-materials/material-1",
            "/api/local-executor/v1/batch-book-asset-images/batch-1/book-1/asset-1/image-1",
        ] {
            assert!(
                is_executor_reference_path(path),
                "expected allowed path: {path}"
            );
        }
        for path in [
            "/api/local-executor/v1/jobs/claim",
            "/api/local-executor/v1/batch-comic-reference-scenes/scene-1?token=leak",
            "/api/admin/presets",
            "https://example.com/reference.png",
        ] {
            assert!(
                !is_executor_reference_path(path),
                "expected rejected path: {path}"
            );
        }
    }
}
