# V88 Account Admin and Cutover Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 完成真实账户/团队资料和 capability 管理面，并在证据充分后清理旧兼容入口，形成可验收的单二进制候选。

**Architecture:** 账户继续使用现有 `authn` 和 workspace profile MySQL 表；后台通过 effective capabilities 控制页面与每条 Go API。Prompt、Provider/模型元数据和技能目录各自保持清晰模块，敏感配置只在服务端。最终清理以引用、数据和回滚证据为门槛。

**Tech Stack:** Go/MySQL、React/AntD、Vitest、Go auth/capability tests、Playwright、Go embed。

## Global Constraints

- 先完成生产流程 Gate 3；前端隐藏入口不能代替后端 403。
- 本阶段不新增错误日志 UI、不实现 Agent、不泄露 MFA/恢复码/密码/Session/Cookie/Provider 密钥。
- 旧代码只在无路由、无数据迁移、无恢复、无部署依赖且测试覆盖替代链路后删除。

---

## Task 1: Complete Profile and Member Facts

**Files:**
- Modify: `api/internal/authn/service_test.go`
- Modify: `api/internal/authn/mysql_store_test.go`
- Modify: `api/internal/httpapi/account_handlers_test.go`
- Modify: `api/internal/httpapi/account_handlers.go`
- Modify: `前台/src/AccountCenterPage.jsx`
- Modify: `前台/src/AccountCenterPage.test.jsx`

- [ ] 测试个人资料更新、当前角色、effective capabilities、团队成员列表和当前用户归属。
- [ ] 对没有现成事实源的团队字段先定义最小模型；若现有表足够则禁止新 migration。
- [ ] 安全区域只显示能力与状态摘要；任何密码变更必须单独验证当前密码并永不回显哈希/MFA/恢复码。
- [ ] `/profile` 与 `/member` 复用同一账户领域但呈现不同 section；刷新从 MySQL 恢复。
- [ ] 运行 auth/httpapi/前台定向测试并提交：`feat(account): complete owned profile and membership [skip ci]`。

## Task 2: Admin Provider Model Catalog Without Secrets

**Files:**
- Create: `api/db/migrations/00020_admin_provider_catalog.sql`
- Create: `api/internal/admincatalog/model.go`
- Create: `api/internal/admincatalog/mysql_store.go`
- Create: `api/internal/admincatalog/service.go`
- Create: `api/internal/admincatalog/service_test.go`
- Create: `api/internal/httpapi/admin_catalog_handlers.go`
- Create: `api/internal/httpapi/admin_catalog_handlers_test.go`
- Modify: `api/internal/httpapi/server.go`
- Modify: `后台/src/api.js`
- Modify: `后台/src/main.jsx`
- Modify: `后台/src/main.test.jsx`

- [ ] migration 定义 Provider/模型的非敏感目录与启用状态；凭据只保存加密引用/环境配置标识，不保存可回显明文。
- [ ] service/handler 测试覆盖 view/edit capability、CSRF、审计摘要和响应字段白名单。
- [ ] 后台新增模型/Provider 页面，明确“已配置”“已启用”“实际可用”三种状态；不可用显示安全原因。
- [ ] 使用 `rg -n "key|secret|token|password|cookie" 后台/src api/internal/admincatalog` 审核前端 payload 与日志。
- [ ] 运行 Go/后台测试并提交：`feat(admin): add safe provider model governance [skip ci]`。

## Task 3: Admin Skill Catalog Without Agent Runtime

**Files:**
- Create: `api/internal/adminskill/model.go`
- Create: `api/internal/adminskill/service.go`
- Create: `api/internal/adminskill/service_test.go`
- Create: `api/internal/httpapi/admin_skill_handlers.go`
- Create: `api/internal/httpapi/admin_skill_handlers_test.go`
- Modify: `api/internal/httpapi/server.go`
- Modify: `后台/src/api.js`
- Modify: `后台/src/main.jsx`
- Modify: `后台/src/main.test.jsx`

- [ ] 定义只读技能目录合同：标识、名称、版本、启用状态、来源、安全摘要；不执行 Agent、不暴露完整系统 Prompt。
- [ ] 测试 `admin.skill.view/edit` capability、Cookie/CSRF、审计与非管理员 403。
- [ ] 后台实现列表、筛选和允许的启停；运行时不可用必须如实显示。
- [ ] 保持 `/agent` 用户功能仍为预留页，管理目录不等于 Agent 业务完成。
- [ ] 运行 Go/后台测试并提交：`feat(admin): add governed skill catalog [skip ci]`。

## Task 4: Admin Navigation and Capability Matrix

**Files:**
- Modify: `api/internal/authn/capabilities.go`
- Modify: `api/internal/authn/capabilities_test.go`
- Modify: `api/internal/httpapi/admin_permissions_test.go`
- Modify: `后台/src/main.jsx`
- Modify: `后台/src/main.test.jsx`
- Modify: `后台/src/admin.css`

- [ ] 写矩阵测试：普通用户、成员管理员、运营管理员、owner 对 prompt/model/skill/member 页面和 API 的实际权限。
- [ ] 后台菜单只展示拥有 view capability 的模块，但所有 API 仍独立 require capability。
- [ ] Dashboard 删除“错误日志将在后续接入”等错误承诺；只展示已接入模块和真实 unavailable 状态。
- [ ] 对齐公网后台的布局密度、Sidebar、Header、表格、Drawer 和响应式，不复制普通用户不可见数据。
- [ ] 运行 capability/httpapi/后台测试并提交：`feat(admin): enforce effective capability navigation [skip ci]`。

## Task 5: Evidence-Based Legacy Cleanup

**Files:**
- Modify only files proven unreachable by repository-wide reference checks.
- Modify: `docs/parity/v88-public-page-matrix.md`
- Create: `docs/parity/v88-cutover-evidence.md`

- [ ] 用 `rg` 列出旧 Node URL、`/api/chat`、Bearer token、V11/V12、直接 Provider 调用、业务 localStorage 和未引用旧组件。
- [ ] 对每个候选记录替代 Go API、MySQL 事实、前端消费者、测试和恢复方案；任一列缺失则保留代码并记录原因。
- [ ] 先删测试中的旧合同并确认失败，再删除真正无引用的兼容实现；不得删除旧仓库或线上资源。
- [ ] 运行受影响模块测试、完整测试和 build；提交粒度按模块拆分，信息使用 `refactor(scope): remove verified legacy path [skip ci]`。

## Task 6: Full Local Embed Acceptance

**Files:**
- Modify: `docs/parity/v88-cutover-evidence.md`
- Modify: `TASKS.md`

- [ ] 从最新 `origin/main` 创建干净验收 checkout，记录 Git SHA 和工作树状态。
- [ ] 仅使用本机独立 MySQL 数据库、Redis prefix、TOS staging prefix、日志目录和非生产端口；对该数据库执行 migration。
- [ ] 完整运行 Go、前台、后台测试和 build，执行 `./scripts/build-embedded-ui.sh`。
- [ ] 启动单个 Go 二进制，验证 `/api/build-info` 等于当前 SHA，用户/后台静态资源响应来自 `go-embed` 而非 Vite。
- [ ] 登录后逐页验收 `/`、`/script`、`/novel-fetch`、`/novel-fetch-workshop`、`/novel-panel`、`/batch-factory`、`/shuihuo-production`、`/tts`、`/history`、`/settings`、`/member`、`/profile`、`/admin`。
- [ ] 单独验收 `/agent`、`/agent/canvas` 的预留说明；验证 `/issues` 返回 404/不可导航。
- [ ] 验证 Cookie Session、CSRF、401、403、刷新恢复、跨用户隔离、Provider/TOS/执行器未配置的真实失败。
- [ ] 更新 `TASKS.md`：真实完成才勾选；公网部署、真实付费 Provider 和公网截图验收保持未完成。
- [ ] 提交 `chore(release): record local embed acceptance [skip ci]`，推送并确认 `HEAD == origin/main`。
