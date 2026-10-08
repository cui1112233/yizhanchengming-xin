# V88 Unified Go Migration Program Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把已批准的 V88 公网迁移设计拆成可顺序交付的工作包，在不复制旧 Node 生产链路的前提下完成 Go + MySQL + Redis + TOS + React/AntD 单二进制系统。

**Architecture:** 以当前 Go API、Cookie Session、CSRF、capability、MySQL 事实表和 Go embed 为唯一主干。迁移按共享底座、创作工作流、生产工作流、账户与管理、最终清理五个门逐步推进；后续阶段只能消费前一阶段已稳定的合同，不另建认证、请求、队列或事实系统。

**Tech Stack:** Go 1.24、MySQL 8、Redis、TOS、React 18、Ant Design 5、Vite 6、Vitest 3、Playwright、Go `embed`。

## Global Constraints

- 只提交 `main`；每次开始和提交前执行 `git fetch origin && git checkout main && git pull --ff-only origin main && git status -sb`。
- 提交信息包含 `[skip ci]`；禁止修改/触发 `.github/workflows`，禁止 force push，禁止部署或改动公网。
- 旧 `cui1112233/-` 的 `origin/v88` 和公网只读；不得合并旧 Git 历史、复制 hash 构建产物、密钥、Token、Cookie 或用户媒体。
- Go 是唯一生产后端；Node 只用于构建/测试 React。禁止 iframe、旧 Node API、`/api/chat`、第二套 Auth/Queue/Worker/Scheduler。
- MySQL 保存业务事实；Redis 只保存队列/租约/锁/短期协调；TOS 只保存有 MySQL 归属记录的媒体对象。
- 浏览器不得保存业务事实；`localStorage` 仅限主题、侧栏折叠等非敏感 UI 偏好，且服务端设置优先。
- 所有写 API 使用 Cookie Session、同源 CSRF、capability 和项目归属校验；用户错误只返回安全摘要与 `request_id`。
- `/issues` 从用户导航、前端实现和 Go HTTP API 移除；不新增管理错误日志 UI。`/agent` 与 `/agent/canvas` 只展示批准的预留说明，退役旧 React/HTTP adapter，但保留 Agent Go 领域包、migration、MySQL/TOS 数据引用，不执行 DROP/DML 或对象删除。
- 每个页面必须覆盖 loading、空态、错误、重试、401/403、刷新恢复和无 Provider/执行器的真实失败。
- 任何 migration 都只做 additive 变更；必须提供用户/团队归属、恢复与删除语义以及 migration contract test。

---

## Work Package Order

1. `2026-10-09-v88-foundation-shell.md`：统一主题、前后台 API 合同、范围收敛、用户壳和 embed 验收门。
2. `2026-10-09-v88-creative-workflows.md`：`/`、`/script`、`/novel-fetch`、`/novel-fetch-workshop`、`/novel-panel`。
3. `2026-10-09-v88-production-workflows.md`：`/batch-factory`、`/shuihuo-production`、`/tts`、`/history`、`/settings`。
4. `2026-10-09-v88-account-admin-cutover.md`：`/member`、`/profile`、`/admin`、旧链路清理与全站验收。

## Program Gates

### Gate 1 — Shared Contract Stable

- [ ] 完成基础计划的所有任务，确认用户/后台请求均经过唯一客户端，`request_id` 在页面可见错误中可追踪。
- [ ] `/issues` 不可从用户 UI 到达且旧 API 安全返回 JSON 404；Agent 两条路由不触发业务 API，旧 `/api/v1/agent/*` 同样安全返回 JSON 404。
- [ ] 前后台使用同一套经过公网核验的颜色、字体、圆角、控件高度 token。
- [ ] `go test ./...`、前后台完整 Vitest、前后台 build 和 embed handler tests 全部通过。

### Gate 2 — Creative Entry Flows Stable

- [ ] 首页所有卡片来自 MySQL 投影或固定功能入口，没有伪造项目数据。
- [ ] 剧本、小说获取、改文、小说面板分别具备独立可恢复事实流。
- [ ] 公网 UI 对照证据记录到 `docs/parity/`，但不把截图或旧 bundle 当作实现依赖。

### Gate 3 — Production Flows Stable

- [ ] 批量工厂、水货生产、TTS 和历史复用同一项目/任务事实，不建立重复历史表。
- [ ] Provider、执行器或 TOS 未配置时返回真实 unavailable/error；不调用付费 Provider 作为单元测试。
- [ ] Redis 中断后可从 MySQL 恢复任务事实，不产生重复执行。

### Gate 4 — Account and Admin Stable

- [ ] 账户、成员/团队、设置和管理能力全部由当前认证用户及 capability 限定。
- [ ] 管理端不泄露 Prompt 正文、Provider 密钥、Session、Cookie、MFA、恢复码或密码。
- [ ] 普通用户访问 `/admin` 由 Go API 返回 403，而不只依赖隐藏菜单。

### Gate 5 — Cutover Evidence Complete

- [ ] 运行 `cd api && go test ./...`。
- [ ] 在 `前台/` 和 `后台/` 分别执行 `npm install --no-package-lock --prefer-offline --no-audit`、`CI=1 npm test -- --run`、`npm run build`。
- [ ] 执行 `./scripts/build-embedded-ui.sh`，确认单二进制 `/api/build-info` 返回当前 Git SHA，HTML/JS 响应包含 `X-YCM-Static-Source: go-embed`。
- [ ] 使用隔离的本机 MySQL/Redis/TOS 前缀完成浏览器路由、登录、CSRF、401/403、刷新恢复和真实失败态验收。
- [ ] 只有替代证据齐全且没有恢复依赖时，才删除未引用的旧兼容代码；不删除旧仓库、生产数据或公网资源。

## Commit Discipline

每个任务最多形成一个小而完整的提交，推荐格式：

```text
test(scope): lock <behavior> [skip ci]
feat(scope): migrate <vertical slice> [skip ci]
refactor(scope): remove verified legacy path [skip ci]
```

如果测试和实现必须原子提交，则合并为一个 `feat`/`fix` 提交。推送后必须记录：

```bash
git rev-parse HEAD
git rev-parse origin/main
git status -sb
```

只有 `HEAD == origin/main` 且工作树没有本任务遗留，阶段才算进入仓库。
