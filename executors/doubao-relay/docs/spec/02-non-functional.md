# 02 · 非功能需求

## N1 · 性能预算

### N1.1 · 响应时间

| 操作 | 50 分位 | 95 分位 | 99 分位 |
|---|---|---|---|
| `GET /tasks/{id}`(查询) | < 50 ms | < 200 ms | < 500 ms |
| `POST /generate`(入队) | < 200 ms | < 1 s | < 3 s |
| `POST /uploads`(1MB 图) | < 500 ms | < 2 s | < 5 s |
| `GET /accounts`(账号列表) | < 2 s | < 5 s | < 10 s |
| 生图端到端(单任务) | 30-90 s | < 180 s | < 300 s |
| 生视频端到端(单任务) | 2-8 min | < 15 min | < 20 min |

### N1.2 · 吞吐

单账号串行:同一时间 1 个任务。
全局默认并发:3 个任务(可配置)。
队列深度:默认 50 个 pending 任务。超出返回 `queue_full`。

### N1.3 · 资源占用

| 指标 | 空闲 | 1 任务进行中 | 3 任务并发 |
|---|---|---|---|
| CPU | < 5% | < 20% | < 40% |
| 内存(RSS) | < 100 MB | < 200 MB | < 400 MB |
| 磁盘 | 20 MB(安装)+ 结果文件 | | |
| 网络 | 0 | 几 MB(图)/几十 MB(视频) | |

**豆包本身的资源不算在内**。豆包是独立 Chromium 应用,占几百 MB 内存属正常。

## N2 · 兼容性

### N2.1 · 操作系统

- Windows 10 1809+(含 Server 2019+)
- Windows 11
- macOS 12 Monterey+(含 Apple Silicon)

**不支持**:Windows 7/8/8.1, Windows Server 2016 及以下, macOS 11 及以下, Linux。

### N2.2 · 豆包版本

- 目标版本:豆包桌面端 v3.0+(2025 年及以后)
- MUST 在启动时跑 UI 兼容性自检:关键 selector(账号按钮、输入框、发送按钮)全部命中才允许服务对外。
- 不兼容时 MUST 在 UI 显著位置提示"豆包版本不兼容,需升级 doubao-relay"。

### N2.3 · WebView 内核

- Windows:WebView2(自带,随 Windows 11 预装;Windows 10 需 Edge 升级到最新版)
- macOS:WKWebView(系统自带)

## N3 · 可用性

### N3.1 · 启动时间
从双击图标到 UI 可用 < 3 秒。
从 UI 启动豆包到 CDP 就绪 < 20 秒。

### N3.2 · 错误消息
所有面向用户的错误消息 MUST 中文,附最多 1 句"下一步建议"。例:
> ❌ 豆包未登录
> 请先在豆包客户端里登录账号,然后点"重连豆包"。

### N3.3 · 容错
豆包崩溃时,中转程序 MUST:
1. 不同时崩溃
2. UI 显示"豆包已退出,等待重启"
3. 当前进行中任务置于 `failed` 状态,错误码 `doubao_crashed`
4. 新请求返回 503

## N4 · 可维护性

### N4.1 · 代码规范
- Rust 代码遵循 `rustfmt` 默认规则 + `cargo clippy -- -D warnings` 零警告
- WebView 侧用 TypeScript + `prettier`
- 每个 PR MUST 带单元测试或集成测试

### N4.2 · 配置集中
所有配置项 MUST 走单一配置文件 `~/.../doubao-relay/config.toml`,UI 和 CLI 都读这里。

### N4.3 · 版本升级
升级时 MUST:
- 保留现有任务数据
- 自动迁移 SQLite schema(用 sqlx migrate)
- 保留配置文件(新增项用默认值填充)

## N5 · 可观测性

### N5.1 · 日志级别
- `error`: 任务失败、CDP 断开、致命错误
- `warn`: 用户干扰(切会话等)、重试、兼容性降级
- `info`: 任务状态变化、HTTP 请求摘要
- `debug`: CDP 命令/事件、JS 注入结果

默认级别 `info`,UI 可调。

### N5.2 · 指标
MAY 在 `/metrics` 暴露 Prometheus 风格指标(默认关闭):
- `doubao_relay_tasks_total{status=..., account=...}`
- `doubao_relay_task_duration_seconds{type=...}`
- `doubao_relay_cdp_disconnects_total`
- `doubao_relay_queue_depth`

## N6 · 安全

见 `12-security.md` 详细讨论。概述:
- 对外 HTTP 服务默认只监听 `127.0.0.1`,不开 0.0.0.0
- Bearer Token 鉴权,首启自动生成 32 字节随机 token
- 不记录 prompt 到日志(除非 debug 级别)
- 不上报任何数据到外部
- SQLite 不加密,但依赖 OS 文件系统权限

## N7 · 本地化

- UI 和日志:中文(简体)
- API 响应消息:中文
- 代码注释和日志 key:英文(便于国际化扩展)

**v0.1 不做多语言切换**。

## N8 · 可打包性

- Windows:单文件 `.msi` 或 `.exe` installer,< 20 MB(不含 WebView2)
- macOS:`.dmg`,universal binary(Intel + ARM),< 30 MB
- MUST 代码签名(可选但强烈推荐,避免 SmartScreen/Gatekeeper 拦)
- 卸载时清理:程序文件 / 快捷方式 / 开机自启项;**保留**用户数据(SQLite 和配置)

## N9 · 可审计性

**所有网络出入站 MUST 本地可见**:
- 对外 HTTP 请求/响应 MUST 可在日志里查到
- CDP 发给豆包的所有命令 MUST debug 日志可见
- 工具自身 **不发任何对外 HTTP 请求**(update check 除外,且 opt-in)

**用户任何时候都可以关掉中转程序,豆包不受影响**。
