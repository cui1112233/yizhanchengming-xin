# 12 · 安全模型

## 威胁模型(Threat Model)

### 资产(Assets)

| 资产 | 位置 | 重要性 |
|---|---|---|
| Bearer Token | config.toml | 高 — 获得等于获得所有豆包能力 |
| 豆包账号登录态 | 豆包自己的 cookie 存储 | 高 — doubao-relay 读不到也不碰 |
| 用户的 prompt 内容 | SQLite + 日志(debug 级别) | 中 — 可能含商业机密 |
| 生成结果文件 | files/ | 中 — 用户的作品 |
| CDP 调试通道 | ws://127.0.0.1:9707 | 高 — 等效于远程控制豆包 |

### 威胁(Threats)

| 威胁 | 说明 | 缓解 |
|---|---|---|
| T1: 本地恶意程序读 token | 其他程序读 config.toml | OS 文件权限;token 可快速轮换 |
| T2: 本地恶意程序打 9707 | 别的程序直接连豆包 CDP | 豆包本身只监听 loopback;doubao-relay 不扩大攻击面 |
| T3: 恶意调用方 DoS | 短时大量请求打爆豆包账号 | 队列深度限制 + 全局并发 |
| T4: Webhook URL 作为 SSRF | 回调恶意 URL 探测内网 | 禁 Private IP 可选,或留给用户自担 |
| T5: 结果文件被篡改 | 生成后被改 | 文件系统 ACL,产品层不管 |
| T6: 偷听 HTTP 流量 | 中间人 | 只 loopback,无中间人 |
| T7: 日志泄露 prompt | 日志被偷 | debug 以上才打 prompt,文件权限 |
| T8: 豆包账号被封 | 使用违反 ToS | 产品层免责;API 返回明确错误码 |

## 认证

### Bearer Token

- 首次启动自动生成 32 字节随机(hex 编码,64 字符)
- 存 `config.toml`
- 用户可在设置页"重新生成"(撤销旧 token)
- MUST 常量时间比较(防时序攻击)

```rust
pub fn verify_token(provided: &str, expected: &str) -> bool {
    use subtle::ConstantTimeEq;
    provided.as_bytes().ct_eq(expected.as_bytes()).into()
}
```

### 认证中间件

```rust
async fn auth_middleware(req: Request, next: Next) -> Result<Response, StatusCode> {
    let token = req.headers()
        .get(axum::http::header::AUTHORIZATION)
        .and_then(|h| h.to_str().ok())
        .and_then(|s| s.strip_prefix("Bearer "));

    match token {
        None => Err(StatusCode::UNAUTHORIZED),
        Some(t) if verify_token(t, &CONFIG.server.bearer_token) => Ok(next.run(req).await),
        _ => Err(StatusCode::UNAUTHORIZED),
    }
}
```

### 豁免端点

`/api/v1/health` 不需要鉴权(方便探活)。

## 网络访问控制

### 绑定地址

**默认且推荐**:`127.0.0.1:8787`,**不监听** `0.0.0.0`。

**不允许**配置成 `0.0.0.0`。如果用户硬要远程访问,只能通过 SSH tunnel 转发。

```rust
#[derive(Deserialize)]
pub struct ServerConfig {
    #[serde(default = "default_host")]
    pub host: String,  // 校验必须是 127.0.0.1 / ::1 / localhost
    pub port: u16,
    pub bearer_token: String,
}

fn validate_host(host: &str) -> Result<(), String> {
    match host {
        "127.0.0.1" | "::1" | "localhost" => Ok(()),
        _ => Err("出于安全考虑,只允许监听 loopback".into()),
    }
}
```

### CORS

**默认关闭**。本地工具不需要跨域。

### Rate limiting

**v0.1 不做 rate limiting**(靠队列深度自然限流)。保留扩展口。

## 敏感数据处理

### Prompt 内容

- SQLite 中存储(明文)
- 日志中:
  - `info` 级别只打 `prompt_length` 和 `prompt_hash`,不打内容
  - `debug` 级别打前 60 字符 + "..."
  - CDP debug 日志中可能完整出现(该日志默认关闭)

### Bearer Token

- SQLite 不存
- 日志 **MUST NOT** 出现(用 `tracing::field::display` 时过滤)
- WebView 的"复制 token"按钮使用后 **MUST** 从剪贴板清除(60 秒超时)

### 账号信息

- 账号名存 SQLite + 日志
- 账号的真实 cookie/session 工具**完全碰不到**(在豆包自己的 Chromium 存储里)

### 结果文件

- 本地文件系统,无加密
- 依赖 OS 文件权限(默认当前用户读写,其他用户无权限)

## Webhook 安全

### 签名

每个 webhook 请求带:

```http
X-Doubao-Relay-Signature: sha256=<hex>
X-Doubao-Relay-Timestamp: 1697433700000
```

签名计算:

```
signature = HMAC-SHA256(key=bearer_token, data=timestamp + "\n" + body)
```

调用方 MUST 校验签名,防止:
- 恶意第三方伪造回调
- 回放攻击(校验 timestamp 在 ±5 分钟内)

### URL 校验

- `callback_url` MUST 是 `http(s)://` scheme
- MUST 不是 loopback(0.0.0.0 / 127.0.0.1 / ::1 / localhost)— 防自环
- MAY 禁 Private IP(10.x / 172.16-31.x / 192.168.x / 169.254.x)— 可配置,默认允许(本地工具用户可能真的要打内网)

### HTTP vs HTTPS

MAY 接受 `http://`(本地工具调用方可能也在内网)。SHOULD 日志 warn 一次。

## 数据驻留

- **所有数据 LOCAL ONLY**。工具 **MUST NOT** 向任何外部服务上报:
  - 使用指标
  - 任务信息
  - 错误报告
  - 用户 ID / 机器 ID
- **例外**:
  - "检查更新"(用户 opt-in):GET github.com/<org>/doubao-relay/releases/latest,不带任何用户标识
  - Webhook callback(用户主动配置)

## 代码签名

### Windows

- 推荐用 EV Code Signing Certificate(避免 SmartScreen)
- `signtool sign /tr http://timestamp.digicert.com /td sha256 /fd sha256 /a <installer>.exe`

### macOS

- MUST 用 Apple Developer ID 签名
- SHOULD notarize(苹果公证),避免 Gatekeeper 拦截
- `codesign --sign "Developer ID Application: <name>" --deep --timestamp --options runtime <app>`
- `xcrun notarytool submit <app>.dmg --apple-id <id> --wait`

**v0.1 可以先不签**,但 Windows 会被 SmartScreen 警告,macOS 会被 Gatekeeper 拦。

## 更新机制

### 检查更新(opt-in)

- 启动后 1 分钟 + 之后每 24 小时
- GET `https://api.github.com/repos/<org>/doubao-relay/releases/latest`
- 对比版本号,新版本在 UI 显示 banner
- **不自动下载**(用户点"下载"再 fetch 二进制)

### 升级流程

- 用户下载新 installer,安装覆盖
- 老版本的 SQLite 自动 migrate
- 配置文件保留(新字段用默认值)

**没有强制更新**。

## 隐私声明

doubao-relay 应在首次启动时显示:

```
doubao-relay 承诺:
✓ 所有数据保存在本机,从不上传任何外部服务
✓ 不收集使用指标、不记录 Prompt 用于分析
✓ 不代理豆包登录,不收集账号密码
✓ 代码开源,任何人可审计

警告:
⚠ 本工具通过模拟客户端操作使用豆包账号。
   使用本工具可能违反豆包用户协议,账号被封等后果由用户承担。
⚠ 请仅使用自己拥有的豆包账号。
```

MUST 明确同意一次才可用。

## 响应式信息披露

### 发现漏洞

- 发现安全问题请发邮件到:`security@<domain>`(spec 签名文档前确定)
- 90 天修复窗口
- 修复后公开披露

### 已知局限

- 无法防范"本机其他进程读 config.toml"(OS 层面问题)
- 无法防范"本机其他进程监听 8787 后窃听 token"(用户若配别的端口可避)
- 无法防范"豆包改版导致读取失败"(靠快速发版)
- Webhook 签名可被持有 token 的攻击者伪造(token 即为 shared secret)

## 审计友好

- 所有网络出入站**对外可见**:
  - HTTP API 请求/响应在 info 日志
  - CDP 命令在 debug 日志(需开启)
  - 工具自身不发任何对外 HTTP(除 check update)
- SQLite 可用 `sqlite3 relay.db` 直接查
- 日志为纯文本 ndjson,可 `jq` 分析
