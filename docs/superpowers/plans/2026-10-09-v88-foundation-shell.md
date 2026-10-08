# V88 Foundation and Shell Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 先收敛用户范围、前后台请求合同、主题字体、用户壳和单二进制验收门，为后续逐页迁移提供唯一基础。

**Architecture:** 保留现有 `httpapi.withObservability`、`observability`、Cookie Session、同源 CSRF 和 capability；只补遗漏，不另建中间件。用户和管理 React 各自只有一个 API 模块，二者消费相同 `{code,message,request_id}` 错误合同。主题值由服务端设置主导，本地仅作首次绘制回退。

**Tech Stack:** Go `net/http`/`slog`/`embed`、React 18、Ant Design 5、Vitest、Testing Library。

## Global Constraints

- 遵守总计划全部约束；本阶段不新增业务表、不删除 Agent Go 模型或 MySQL 数据。
- 只迁移公网经核验的字体栈、色值、间距和合法静态资产；CSS 必须限制在 `.user-shell`、`.admin-shell` 或页面根节点。
- 本阶段不实现 Agent 功能，不实现用户/管理错误日志页面，不修改 TASKS 中尚未验收的完成状态。
- 写代码前先同步 `main`；共享工作树不干净时使用新的临时克隆，禁止覆盖他人未提交文件。

---

## Task 1: Lock the Approved Route Scope

**Files:**
- Modify: `前台/src/RouterApp.test.jsx`
- Modify: `前台/src/UserShell.test.jsx`
- Modify: `前台/src/HomePage.test.jsx`
- Create: `前台/src/AgentReservedPage.test.jsx`
- Create: `前台/src/AgentReservedPage.jsx`
- Modify: `前台/src/RouterApp.jsx`
- Modify: `前台/src/UserShell.jsx`
- Modify: `前台/src/HomePage.jsx`

- [ ] 添加失败测试：主/移动/侧栏导航都没有“问题日志”，访问 `/issues` 命中 404 页面且不调用 `listIssues`。
- [ ] 添加失败测试：`/agent` 与 `/agent/canvas` 都渲染同一预留组件，正文明确“Agent 将重新设计，当前不可用”，不请求 Agent API。
- [ ] 添加失败测试：首页 Agent 入口保留产品位置，但状态标签为“待重新设计”，点击只进入 `/agent` 预留页。
- [ ] 运行定向测试并确认先失败：

```bash
cd 前台
CI=1 npm test -- --run src/RouterApp.test.jsx src/UserShell.test.jsx src/HomePage.test.jsx src/AgentReservedPage.test.jsx
```

- [ ] 新建纯展示组件；组件只接收 `onNavigate`，不导入 `api.js`：

```jsx
export default function AgentReservedPage({ onNavigate }) {
  return <Result status="info" title="Agent 工作区待重新设计" subTitle="当前版本不执行 Agent 任务，也不会创建项目或调用 Provider。" extra={<Button onClick={() => onNavigate('/')}>返回首页</Button>} />
}
```

- [ ] 从 `RouterApp.jsx` 移除 `IssuesPage`、`AgentStudioPage`、`AgentCanvasPage` 的运行时导入；两条 Agent 路由映射到预留组件，`/issues` 不再是合法用户路由。
- [ ] 从 `UserShell.jsx` 删除 `/issues` 导航项；保留 Agent 导航项并确保活动态同时覆盖 `/agent/canvas`。
- [ ] 保留 `AgentStudioPage.jsx`、`AgentCanvasPage.jsx`、`api/internal/agentstudio/` 和对应 migration，不做破坏性删除。
- [ ] 重跑定向测试，期望全部通过。
- [ ] 提交：`feat(scope): reserve agent and remove user issues entry [skip ci]`。

## Task 2: Create One Admin API Client Contract

**Files:**
- Create: `后台/src/api.js`
- Create: `后台/src/api.test.js`
- Modify: `后台/src/main.jsx`
- Modify: `后台/src/main.test.jsx`

- [ ] 为 `requestJSON` 添加失败测试，覆盖 `credentials: include`、JSON body、响应头/响应体 `request_id`、401、403、非 JSON 错误和网络错误。
- [ ] 为管理端写请求添加失败测试，确认继续依赖服务端同源 CSRF，不创建 Bearer token 或浏览器 token 存储。
- [ ] 运行 `cd 后台 && CI=1 npm test -- --run src/api.test.js src/main.test.jsx` 并确认失败。
- [ ] 在 `后台/src/api.js` 导出与用户端同形的错误类型：

```js
export class APIError extends Error {
  constructor(message, { status = 0, code = 'API_ERROR', requestId = '' } = {}) {
    super(message)
    Object.assign(this, { name: 'APIError', status, code, requestId })
  }
}

export async function requestJSON(path, options = {}) {
  const headers = new Headers(options.headers || {})
  if (options.body !== undefined) headers.set('Content-Type', 'application/json')
  const response = await fetch(path, { ...options, headers, credentials: 'include' })
  const payload = await response.json().catch(() => null)
  if (response.ok) return payload
  throw new APIError(payload?.message || `请求失败（HTTP ${response.status}）`, {
    status: response.status,
    code: payload?.code,
    requestId: response.headers.get('X-Request-ID') || payload?.request_id || '',
  })
}
```

- [ ] 将 `后台/src/main.jsx` 内联 `adminFetch` 全部替换为 `requestJSON`；错误 Alert/Message 在可用时显示 `请求编号：<requestId>`，不展示 payload、栈或敏感配置。
- [ ] 使用 `rg -n "fetch\\(" 后台/src` 验证只有 `后台/src/api.js` 可以直接调用 `fetch`。
- [ ] 重跑后台定向测试，期望全部通过。
- [ ] 提交：`refactor(admin): centralize authenticated api client [skip ci]`。

## Task 3: Enforce the Safe Error and Request-ID Contract

**Files:**
- Modify: `api/internal/httpapi/observability_integration_test.go`
- Modify: `api/internal/httpapi/server_test.go`
- Modify: `api/internal/httpapi/safe_errors.go`
- Modify: `前台/src/api.observability.test.js`
- Modify: `前台/src/ui/PageState.jsx`
- Modify: `前台/src/ui/PageState.test.jsx`

- [ ] 添加 Go 失败测试：未知 API、panic、认证失败、权限失败和服务错误都返回单个有效 `X-Request-ID`；JSON 错误包含同值 `request_id`。
- [ ] 添加 Go 失败测试：响应和结构化日志都不出现测试用密码、Cookie、Bearer token、Provider key 或 DSN。
- [ ] 添加前台失败测试：`APIError.requestId` 传给通用错误态，错误态显示安全消息、请求编号、重试按钮，不显示 payload。
- [ ] 运行：

```bash
cd api
go test ./internal/httpapi ./internal/observability
cd ../前台
CI=1 npm test -- --run src/api.observability.test.js src/ui/PageState.test.jsx
```

- [ ] 只对测试证明的缺口修改 `safe_errors.go`/路由错误写入；继续复用 `withObservability` 和 `observability.NewJSONLogger`。
- [ ] 统一通用前台错误视图接口：

```jsx
<PageState kind="error" message={error.message} requestId={error.requestId} onRetry={reload} />
```

- [ ] 用 `rg -n "err\\.Error\\(\\)|error\\.Error\\(\\)" api/internal/httpapi` 审核 handler，所有用户响应必须经过安全映射，不能直接回传内部错误。
- [ ] 重跑上述测试，期望全部通过。
- [ ] 提交：`fix(observability): standardize safe request errors [skip ci]`。

## Task 4: Centralize Public Theme and Typography

**Files:**
- Create: `ui/theme-contract.js`
- Create: `前台/src/theme.js`
- Create: `前台/src/theme.test.js`
- Modify: `前台/src/main.jsx`
- Modify: `前台/src/app.css`
- Create: `后台/src/theme.js`
- Create: `后台/src/theme.test.js`
- Modify: `后台/src/main.jsx`
- Modify: `后台/src/admin.css`
- Modify: `docs/parity/v88-reusable-assets-and-css.md`

- [ ] 从现有公网审计中写出经核验的字体栈、主色、背景、边框、圆角、控件高度和断点证据；不从压缩 bundle 猜字体文件。
- [ ] 为前后台主题模块添加失败测试，断言共享值完全相同；后台只允许表格密度差异。
- [ ] 运行 `CI=1 npm test -- --run src/theme.test.js`（分别在 `前台/`、`后台/`）并确认失败。
- [ ] 在两个 `theme.js` 导出同名不可变合同：

```js
// ui/theme-contract.js
export const typography = Object.freeze({ fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif' })
export const baseToken = Object.freeze({ colorPrimary: '#5b62d9', borderRadius: 10, controlHeight: 40, fontSize: 14 })

// 前台/src/theme.js and 后台/src/theme.js
import { theme as antdTheme } from 'antd'
import { baseToken, typography } from '../../ui/theme-contract.js'
export function themeConfig(mode) {
  return {
    algorithm: mode === 'dark' ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
    token: { ...baseToken, fontFamily: typography.fontFamily },
  }
}
```

- [ ] 前后台都从 `ui/theme-contract.js` 读取基础值，`main.jsx` 只通过 `themeConfig(theme)` 配置 AntD；根 CSS 使用同一字体栈和 CSS variables，不在页面重新定义全局字体。
- [ ] 用 `rg -n "font-family|colorPrimary|borderRadius|controlHeight" 前台/src 后台/src` 审核例外并在 CSS 中限制作用域。
- [ ] 重跑两端主题测试和已有 `main.test.jsx`，期望通过。
- [ ] 提交：`style(theme): align shared public tokens and typography [skip ci]`。

## Task 5: Make Server Theme Authoritative Without Business Browser State

**Files:**
- Modify: `前台/src/main.test.jsx`
- Modify: `前台/src/main.jsx`
- Modify: `前台/src/SettingsPage.test.jsx`
- Modify: `前台/src/SettingsPage.jsx`
- Modify: `前台/src/api.js`

- [ ] 添加失败测试：认证后服务端主题覆盖本地回退；用户切换主题调用 `saveWorkspaceSettings`，保存失败时回滚并显示安全错误。
- [ ] 添加失败测试：本地存储只包含 `yizhan-theme`/布局键，不写项目、任务、历史、账号或 API 配置。
- [ ] 运行 `cd 前台 && CI=1 npm test -- --run src/main.test.jsx src/SettingsPage.test.jsx` 并确认失败。
- [ ] 将切换动作变成先更新 UI、再保存服务端；失败时恢复上一个值。服务端返回是刷新后的权威值。
- [ ] `SettingsPage` 明确区分“已选择配置”和“实际生效配置”；不可用执行器显示服务端 reason/code，不伪造在线。
- [ ] 重跑定向测试，期望通过。
- [ ] 提交：`feat(settings): make server preference authoritative [skip ci]`。

## Task 6: Lock the Single-Binary Release Gate

**Files:**
- Modify: `api/internal/webui/handler_test.go`
- Modify: `api/cmd/server/main_test.go` (create if absent)
- Modify: `scripts/build-embedded-ui.sh`
- Create: `docs/operations/local-embed-acceptance.md`

- [ ] 添加 handler 失败测试：用户/后台 SPA 刷新走各自 embed，静态响应有 `X-YCM-Static-Source: go-embed`，API 永不回退为 SPA。
- [ ] 添加 build-info 测试：`/api/build-info` JSON 的 `gitSha` 与 ldflags 注入值一致。
- [ ] 审核构建脚本：只执行两个 Vite build、复制到 `api/internal/webui/dist/{user,admin}`、Go build；不启动 Node server、不访问公网、不调用 Actions。
- [ ] 运行：

```bash
cd api
go test ./internal/webui ./cmd/server
cd ..
./scripts/build-embedded-ui.sh
test -x api/.staging-bin/ycm-server
```

- [ ] 在操作文档记录隔离环境变量、migration 前置条件、启动/停止、`curl -i /`、`curl -i /admin/`、`curl /api/build-info` 和 401/403/CSRF 验收命令；不得写真实密码或密钥。
- [ ] 提交：`build(embed): enforce single binary acceptance gate [skip ci]`。

## Task 7: Full Foundation Verification

- [ ] 同步最新 `main`；若远端前进，先提交已验证改动，再 `git pull --rebase origin main`，只处理本计划文件冲突。
- [ ] 运行完整 Go 测试：`cd api && go test ./...`。
- [ ] 运行前台：`npm install --no-package-lock --prefer-offline --no-audit`、`CI=1 npm test -- --run`、`npm run build`。
- [ ] 运行后台同样三条命令。
- [ ] 运行 `./scripts/build-embedded-ui.sh` 并记录最终退出码、测试总数和构建产物。
- [ ] 检查：`git diff --check`、`git status -sb`、`rg -n "localStorage|sessionStorage" 前台/src 后台/src`、`rg -n "fetch\\(" 前台/src 后台/src`。
- [ ] 更新 `TASKS.md`，只勾选经过本阶段验证的条目；Agent 功能、公网登录、真实 Provider、截图验收保持未完成。
- [ ] 最终提交：`chore(scope): record foundation acceptance [skip ci]`。
- [ ] 推送 `main`，确认 `git rev-parse HEAD` 等于 `git rev-parse origin/main`。
