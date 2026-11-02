# 一战晟铭执行器配对与单 EXE 安装包 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将豆包本地执行器升级为“一战晟铭执行器”1.2.0，使用一次性配对、低频长等待领取、隐藏开发者 HTTP API，并交付单个 Windows 安装 EXE。

**Architecture:** V88 Go 后端在现有领取接口上增加可选的长等待参数，并用进程内任务通知避免数据库空轮询；旧客户端不带参数时仍立即返回。Rust/Tauri 执行器保存独立设备凭证，默认关闭本机 HTTP API，开发者连续点击版本号 7 次后才可见本机接口设置。Windows 使用 Tauri NSIS current-user 安装包，React 设置页下载同一个 EXE。

**Tech Stack:** Go 1.x、MySQL、Rust/Tokio、Tauri 2、Preact/TypeScript、Vitest、NSIS

## Global Constraints

- 维护分支为 `v88`，不创建新分支，不推送远程，不发布公网。
- 普通用户只接触 10 分钟有效的一次性配对码。
- 配对码、设备凭证、本机 Bearer Token 三者独立。
- 新安装默认关闭 HTTP API；旧配置升级时保留历史上的 HTTP API 可用状态。
- HTTP API 只能绑定 `127.0.0.1`、`localhost` 或 `::1`。
- Windows 安装包仅为当前用户安装，不要求管理员权限。
- 产品名称统一为“一战晟铭执行器”，版本升级为 `1.2.0`。
- Logo 源文件为 `G:\Backup\Pictures\一战晟铭图标.png`。
- 前端源码修改后必须重新构建并通过 Tauri 打包，禁止用编译产物反向覆盖源码。

---

### Task 1: 后端长等待领取

**Files:**
- Modify: `backend/internal/localexecutor/service.go`
- Modify: `backend/internal/localexecutor/jobs.go`
- Create: `backend/internal/localexecutor/jobs_wait_test.go`
- Modify: `backend/internal/httpapi/local_executor_jobs.go`
- Create: `backend/internal/httpapi/local_executor_long_poll_test.go`

**Interfaces:**
- Produces: `func (s *Service) ClaimJobWait(ctx context.Context, executorToken string, wait time.Duration) (ClaimResult, error)`
- Produces: `GET/POST /api/local-executor/v1/jobs/claim?waitSeconds=25`
- Preserves: 不带 `waitSeconds` 时沿用当前立即领取/204 行为。

- [ ] **Step 1: 写 Service 长等待失败测试**

```go
func TestClaimJobWaitWakesWhenJobIsCreated(t *testing.T) {
    service := NewService(NewMemoryStore(), time.Now)
    paired := pairTestExecutor(t, service, "alice")
    result := make(chan ClaimResult, 1)
    go func() {
        claim, _ := service.ClaimJobWait(context.Background(), paired.Token, time.Second)
        result <- claim
    }()
    time.Sleep(20 * time.Millisecond)
    _, err := service.CreateJob(context.Background(), "alice", CreateJobInput{
        SourceTaskID: "wait-test", Platform: PlatformDoubao,
        Payload: map[string]any{"type": "image", "prompt": "测试"},
    })
    if err != nil { t.Fatal(err) }
    select {
    case claim := <-result:
        if claim.Job.SourceTaskID != "wait-test" { t.Fatalf("claim=%+v", claim) }
    case <-time.After(500 * time.Millisecond):
        t.Fatal("long wait did not wake")
    }
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./backend/internal/localexecutor -run ClaimJobWait -count=1`

Expected: FAIL，提示 `ClaimJobWait` 未定义。

- [ ] **Step 3: 为 Service 增加广播式任务通知**

```go
type Service struct {
    store Store
    jobs JobStore
    now func() time.Time
    jobSignalMu sync.Mutex
    jobSignal chan struct{}
}

func (s *Service) notifyJobCreated() {
    s.jobSignalMu.Lock()
    close(s.jobSignal)
    s.jobSignal = make(chan struct{})
    s.jobSignalMu.Unlock()
}
```

`NewService` 初始化 `jobSignal`；`CreateJob` 成功写入后调用 `notifyJobCreated()`。`ClaimJobWait` 先立即领取，遇到 `ErrNoClaimableJob` 后等待当前 signal、超时或 context 取消；被唤醒后重新领取。

- [ ] **Step 4: 增加 HTTP 兼容测试**

测试不带 `waitSeconds` 立即返回 204；带 `waitSeconds=1` 会等待新任务并返回 200；`waitSeconds=31` 返回 400。

- [ ] **Step 5: 实现 HTTP 参数解析**

```go
wait := time.Duration(0)
if raw := req.URL.Query().Get("waitSeconds"); raw != "" {
    seconds, err := strconv.Atoi(raw)
    if err != nil || seconds < 1 || seconds > 30 {
        writeExecutorError(w, http.StatusBadRequest, "waitSeconds must be between 1 and 30")
        return
    }
    wait = time.Duration(seconds) * time.Second
}
claim, err := service.ClaimJobWait(req.Context(), token, wait)
```

- [ ] **Step 6: 运行后端定向测试**

Run: `go test ./backend/internal/localexecutor ./backend/internal/httpapi -run 'ClaimJobWait|LongPoll|SettingsCanCreateImageAndVideoExecutorTests' -count=1`

Expected: PASS。

### Task 2: 执行器长等待和退避

**Files:**
- Modify: `executors/doubao-relay/src-tauri/src/v88_executor.rs`

**Interfaces:**
- Consumes: `POST /api/local-executor/v1/jobs/claim?waitSeconds=25`
- Produces: `fn retry_delay(failures: u32) -> Duration`，依次返回 2、5、10、30 秒并封顶 30 秒。

- [ ] **Step 1: 写领取 URL 和退避单测**

```rust
#[test]
fn claim_uses_long_wait_and_errors_back_off() {
    assert!(V88ExecutorClient::claim_path().contains("waitSeconds=25"));
    assert_eq!(retry_delay(1), Duration::from_secs(2));
    assert_eq!(retry_delay(2), Duration::from_secs(5));
    assert_eq!(retry_delay(3), Duration::from_secs(10));
    assert_eq!(retry_delay(9), Duration::from_secs(30));
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cargo test --manifest-path executors/doubao-relay/src-tauri/Cargo.toml claim_uses_long_wait_and_errors_back_off`

Expected: FAIL，缺少 `claim_path` 或 `retry_delay`。

- [ ] **Step 3: 修改领取循环**

请求改为 `/api/local-executor/v1/jobs/claim?waitSeconds=25`。收到 204 后直接开始下一次长等待；成功领取后将连续失败次数归零；网络错误后按 `retry_delay` 休眠。未配对或没有可用豆包账号时保留低频本地等待，不请求服务器。

- [ ] **Step 4: 运行 Rust 定向测试**

Run: `cargo test --manifest-path executors/doubao-relay/src-tauri/Cargo.toml v88_executor::tests -- --nocapture`

Expected: PASS。

### Task 3: HTTP API 默认关闭与旧配置迁移

**Files:**
- Modify: `executors/doubao-relay/src-tauri/src/config.rs`
- Modify: `executors/doubao-relay/src-tauri/src/lib.rs`
- Modify: `executors/doubao-relay/src-tauri/src/commands/app.rs`
- Modify: `executors/doubao-relay/src-tauri/src/state.rs`

**Interfaces:**
- Extends: `ServerConfig { enabled: bool, host: String, port: u16, bearer_token: String }`
- Extends: `app_api_status -> { enabled, listening, host, port, error }`
- Preserves: 旧 `config.toml` 没有 `enabled` 时迁移为 `enabled=true`；全新配置为 `enabled=false`。

- [ ] **Step 1: 写配置迁移测试**

```rust
#[test]
fn new_install_disables_http_but_legacy_config_keeps_it_enabled() {
    let new_cfg = Config::load_or_create(&new_path).unwrap();
    assert!(!new_cfg.server.enabled);
    fs::write(&legacy_path, "[server]\nport=8787\nbearer_token=\"abc\"\n").unwrap();
    let legacy = Config::load_or_create(&legacy_path).unwrap();
    assert!(legacy.server.enabled);
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cargo test --manifest-path executors/doubao-relay/src-tauri/Cargo.toml new_install_disables_http_but_legacy_config_keeps_it_enabled`

Expected: FAIL，`enabled` 未定义。

- [ ] **Step 3: 实现配置迁移和启动门控**

新配置默认 `enabled=false`。读取已有文件时先检查 TOML 的 `[server] enabled` 是否存在；不存在则按历史版本始终开启的事实写入 `true`。`spawn_http_server` 在关闭时不绑定端口，并把监听状态保持为 false。

- [ ] **Step 4: 增加明确 API 状态**

`AppState` 保存最近一次 HTTP API 启动错误。绑定成功时清空错误，绑定失败时记录中文原因。`app_api_status` 返回开关、监听、地址和错误；修改 `enabled/host/port` 标记需要重启。

- [ ] **Step 5: 运行配置与命令测试**

Run: `cargo test --manifest-path executors/doubao-relay/src-tauri/Cargo.toml config::tests commands::app::tests -- --nocapture`

Expected: PASS。

### Task 4: 隐藏开发者模式和执行器状态界面

**Files:**
- Modify: `executors/doubao-relay/frontend/src/api.ts`
- Modify: `executors/doubao-relay/frontend/src/pages/settings.tsx`
- Modify: `executors/doubao-relay/frontend/src/pages/settings.css`
- Modify: `executors/doubao-relay/frontend/src/pages/settings.test.ts`
- Modify: `executors/doubao-relay/frontend/src/components/status-bar.tsx`
- Modify: `executors/doubao-relay/frontend/src/view-model.ts`
- Modify: `executors/doubao-relay/frontend/src/view-model.test.ts`

**Interfaces:**
- Produces: `registerVersionClick(now?: number): { unlocked: boolean; remaining: number }`
- Stores: localStorage key `yizhanchengming.executor.developerMode` with value `enabled`。
- Extends: `AppConfig.server.enabled` and `ApiStatus.error`。

- [ ] **Step 1: 写 7 次点击解锁测试**

```ts
it('unlocks developer mode only on the seventh version click', () => {
  resetDeveloperModeForTest();
  for (let i = 1; i < 7; i += 1) expect(registerVersionClick(1000 + i * 100).unlocked).toBe(false);
  expect(registerVersionClick(1700).unlocked).toBe(true);
});
```

- [ ] **Step 2: 运行前端测试确认失败**

Run: `pnpm --dir executors/doubao-relay/frontend test -- --run src/pages/settings.test.ts`

Expected: FAIL，解锁函数未定义。

- [ ] **Step 3: 实现普通设置页和开发者模式**

普通设置页最前展示“连接 V88”，文案使用“一战晟铭执行器”。HTTP API 整组只在开发者模式开启后渲染；版本号连续 7 次点击解锁，从第 3 次开始 toast 剩余次数。开发者区提供 HTTP API 开关、端口、状态、错误、自动选择端口、Token 显示/复制/重新生成和调用示例。

- [ ] **Step 4: 更新状态栏文案**

普通状态栏显示 V88 的未配对/已连接/重连状态和当前任务状态；`● API` 只在开发者模式中显示，HTTP API 关闭不呈现为故障。

- [ ] **Step 5: 运行执行器前端测试与构建**

Run: `pnpm --dir executors/doubao-relay/frontend test -- --run`

Run: `pnpm --dir executors/doubao-relay/frontend build`

Expected: 全部 PASS，TypeScript 无错误。

### Task 5: V88 设置页设备状态与单 EXE 下载

**Files:**
- Modify: `frontend/src/user/pages/SettingsPage.jsx`
- Modify: `frontend/src/shared/styles/settings-workbench.css`
- Modify: `frontend/src/user/pages/SettingsPage.local-executor.source.test.js`
- Modify: `backend/internal/httpapi/local_executor_downloads.go`
- Modify: `backend/internal/httpapi/local_executor_downloads_test.go`

**Interfaces:**
- Produces: `/downloads/local-executor/一战晟铭执行器-安装版.exe`
- Preserves: macOS 1.0.0 DMG 下载地址。

- [ ] **Step 1: 更新下载路由失败测试**

断言 manifest 的 Windows 地址包含 UTF-8 URL 编码后的 `一战晟铭执行器-安装版.exe`，请求该文件返回 PE Content-Type，旧 ZIP 不再出现在普通下载清单中。

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./backend/internal/httpapi -run LocalExecutorDownload -count=1`

Expected: FAIL，manifest 仍返回 ZIP。

- [ ] **Step 3: 更新下载路由和设置页**

Windows 下载按钮改为“下载 Windows 安装版 1.2.0”，地址指向单 EXE。设备区域逐台显示名称、版本、在线/离线、最后在线时间；保留在线数量和图片/视频测试按钮。普通页面不出现端口、Bearer Token 和领取任务。

- [ ] **Step 4: 运行网页前端与后端定向测试**

Run: `node --test frontend/src/user/pages/SettingsPage.local-executor.source.test.js`

Run: `go test ./backend/internal/httpapi -run 'LocalExecutorDownload|SettingsCanCreateImageAndVideoExecutorTests' -count=1`

Expected: PASS。

### Task 6: 品牌、Logo 与 NSIS 安装包

**Files:**
- Modify: `executors/doubao-relay/src-tauri/tauri.conf.json`
- Modify: `executors/doubao-relay/src-tauri/Cargo.toml`
- Modify: `executors/doubao-relay/frontend/package.json`
- Modify: `executors/doubao-relay/package.json`
- Regenerate: `executors/doubao-relay/src-tauri/icons/*`
- Create: `scripts/build-yizhanchengming-executor.ps1`

**Interfaces:**
- Produces: `data/downloads/一战晟铭执行器-安装版.exe`
- Product version: `1.2.0`
- Installer mode: NSIS `currentUser`。

- [ ] **Step 1: 生成多尺寸应用图标**

Run: `pnpm --dir executors/doubao-relay exec tauri icon 'G:\Backup\Pictures\一战晟铭图标.png'`

Expected: `src-tauri/icons/icon.ico` 等资源更新，16/32/256 尺寸可读取且角落透明。

- [ ] **Step 2: 更新产品信息和版本**

`tauri.conf.json` 设置 `productName: "一战晟铭执行器"`、窗口标题、`bundle.targets: ["nsis"]`、`bundle.windows.nsis.installMode: "currentUser"`。Cargo 与两个 package.json 版本统一为 `1.2.0`，保持 `identifier` 不变以支持覆盖升级。

- [ ] **Step 3: 编写可重复打包脚本**

脚本检查 Rust、pnpm 和 Logo；执行前端测试、前端构建、`cargo tauri build --bundles nsis`；找到唯一 NSIS 输出后复制为 `data/downloads/一战晟铭执行器-安装版.exe`；最后输出文件大小和 SHA-256。脚本不得删除用户数据或其他下载文件。

- [ ] **Step 4: 构建单个安装 EXE**

Run: `powershell -ExecutionPolicy Bypass -File scripts/build-yizhanchengming-executor.ps1`

Expected: `data/downloads/一战晟铭执行器-安装版.exe` 存在，文件版本和产品名称为 1.2.0 / 一战晟铭执行器。

### Task 7: 综合验证

**Files:**
- Test only; do not modify unrelated files.

**Interfaces:**
- Verifies all interfaces from Tasks 1-6.

- [ ] **Step 1: 运行 Go 定向回归**

Run: `go test ./backend/internal/localexecutor ./backend/internal/httpapi -count=1`

Expected: PASS。

- [ ] **Step 2: 运行执行器 Rust 回归**

Run: `cargo test --manifest-path executors/doubao-relay/src-tauri/Cargo.toml -- --nocapture`

Expected: PASS。

- [ ] **Step 3: 运行执行器前端回归**

Run: `pnpm --dir executors/doubao-relay/frontend test -- --run`

Run: `pnpm --dir executors/doubao-relay/frontend build`

Expected: PASS。

- [ ] **Step 4: 验证本地 Go 嵌入运行时**

按仓库脚本重新构建并启动本地 Go review 服务，确认 `/settings` 能下载新 EXE，配对码显示 10 分钟有效，设备在线状态正常。

- [ ] **Step 5: 浏览器验收网页流程**

在已登录的 `http://localhost:4001/settings` 中验证生成配对码、在线/离线状态、图片测试和视频测试；确认页面无端口、Bearer Token 或领取任务文案。

- [ ] **Step 6: 执行器 UI 验收**

安装新 EXE，验证品牌和图标；确认普通设置无 HTTP API，连续点击版本号 7 次后出现开发者模式；验证端口占用提示和 HTTP API 开关。

- [ ] **Step 7: 真实图片与视频闭环**

通过网页分别提交一项图片和一项视频，记录 V88 job ID、执行器任务 ID、结果文件路径、媒体尺寸/时长、文件大小和 SHA-256；只有网页展示真实预览并能重新加载时才判定成功。

