# V88 Admin 后台 Go 化：第一批设计

## 目标与边界

本设计仅覆盖第一批：后台 capability、审计、`/admin` 外壳，以及直接管理既有 Go 提示词事实源。实施位置是 `cui1112233/yizhanchengming-xin` 的干净 `main` 工作目录；旧 V88 / Node 管理后台只作为 UI 与行为参考，不作为运行服务或数据来源。

第一批不实现模型、技能、错误日志浏览和成员管理页面。它们在第二批复用本设计的 capability、审计、API 与 UI 壳。

系统提示词正文不进入 React，不进入新的 SQL seed，也不返回给无 `admin.prompt.view` 权限的调用者。Go `generation` prompt 模块是默认种子的唯一代码来源；MySQL 是运行时版本、草稿、发布和恢复的唯一事实源。

## 当前事实基线

| 边界 | 当前文件或表 | 当前行为 | 第一批处理 |
|---|---|---|---|
| 登录身份 | `api/internal/authn/mysql_store.go`、`api/internal/authn/session.go` | 从 `auth_users` 与 `auth_user_capabilities` 装载用户与显式 capability | 增加统一的有效 capability 计算，不改变 cookie 会话协议 |
| HTTP 鉴权 | `api/internal/httpapi/permissions.go`、`api/internal/httpapi/auth_handlers.go` | `userHasCapability` 因 `admin` / `owner` 角色字符串直接放行 | 删除 HTTP 层角色旁路，所有 Admin API 只检查有效 capability |
| 初始用户 | `api/internal/authn/bootstrap.go` | 首次初始化创建 legacy `admin` | 后续新建 bootstrap 创建 `owner`；已存在 `admin` 保持不改写 |
| 提示词模型 | `api/internal/generation/model.go`、`api/internal/generation/prompts.go` | `DefaultPrompts` 提供 Go 默认正文，`PromptResolver` 从 MySQL 取启用的最高版本 | Go 默认值用于“缺失时初始化”；运行时只解析已发布记录 |
| 提示词 MySQL | `generation_prompts`，定义于 `api/db/migrations/00002_task12_generation.sql` | `id`、`prompt_key`、`version`、`content`、`enabled`、时间戳；`uq_generation_prompts_key_version` | 原表扩展为可区分草稿/已发布/归档，并保留所有版本 |
| 现存演进 | `api/db/migrations/00006_task13_match_audio.sql` | 同一 `prompt_key` 已存在多个 `enabled=1` 历史版本，解析依赖最高版本 | 新迁移先把每个 key 的最高已启用版本规范为唯一已发布版本，其余保留为已归档 |
| 后台前端 | `后台/src/main.jsx`、`后台/index.html`、`后台/package.json` | AntD 占位页；尚未接入认证、路由、菜单或 API | 保持独立 React + AntD 构建，接入 Go 同源 API，产物挂到 `/admin` |
| 用户前端 | `前台/src/api.js`、`前台/src/UserShell.jsx` | 已有带 cookie 的请求、刷新恢复和用户工作区 | Task 2 在 UserShell 迁移时消费 capability 契约；Task 7 不修改用户端入口 |

## 角色兼容与有效能力

标准角色为 `owner`、`dev`、`manager`、`member`。角色只参与 `roleDefaultCapabilities(role)` 的计算；路由中间件不得读取 `Role` 决定允许或拒绝。

```text
effectiveCapabilities = roleDefaultCapabilities(role) ∪ auth_user_capabilities
```

`authn.User.Capabilities` 在认证、刷新与 `GET /api/auth/current-user` 前就应已经是去重、排序后的有效集合。HTTP 与 React 都只消费该集合。

| 数据库 role | 前台展示 | 默认后台 capability | 兼容策略 |
|---|---|---|---|
| `owner` | 所有者 | 十一个 `admin.*` | 标准所有者 |
| `dev` | 开发者 | 十一个 `admin.*` | 标准开发者 |
| `admin` | 所有者 | 十一个 `admin.*` | legacy `admin` 是 `owner` 的能力别名；本阶段不写回数据库 |
| `manager` | 管理员 | 无 | 只能通过 `auth_user_capabilities` 获得明确单项权限 |
| `member` | 普通成员 | 无 | 即使访问 `/admin` 或 API 也无后台权限 |
| 其他值 | 未知角色 | 无 | 仍可登录，但后台入口隐藏且 Admin API 返回 403 |

新 bootstrap 用户在 `api/internal/authn/bootstrap.go` 创建为 `owner`。本阶段不迁移、不删除、不批量更新任何 legacy `admin` 行；未来的数据迁移必须先独立审计现存用户、会话与回滚影响。

## Capability 清单

| Capability | 第一批用途 | 第二批预留用途 |
|---|---|---|
| `admin.dashboard.view` | 后台首页 | — |
| `admin.prompt.view` | 列表、版本与正文读取 | — |
| `admin.prompt.edit` | 创建、保存草稿 | — |
| `admin.prompt.publish` | 发布、恢复 | — |
| `admin.model.view` | — | 模型配置只读及脱敏状态 |
| `admin.model.manage` | — | 模型配置修改 |
| `admin.skill.view` | — | 技能列表与版本读取 |
| `admin.skill.manage` | — | 技能草稿、发布、归档 |
| `admin.audit.view` | — | 审计与脱敏错误日志读取 |
| `admin.member.view` | — | 授权对象与当前授权读取 |
| `admin.member.manage` | — | capability 授予、撤销与成员状态操作 |

后台入口资格为“有效集合中存在任意 `admin.` 前缀 capability”。每一个页面、菜单项、表格操作和 API 仍要求上表中的具体 capability；入口可见性永远不是服务端授权依据。

## MySQL 迁移设计

实施开始时，先以最新 `main` 的 `api/db/migrations` 目录确认下一个未占用编号；实施前必须重新执行 `git fetch`、`git rebase` 并再次确认编号。该 Goose migration 命名为“已确认的下一个编号 + `_task7_admin_governance.sql`”，只定义结构与状态转换，不内置系统提示词正文。

```sql
-- +goose Up
ALTER TABLE generation_prompts
  ADD COLUMN lifecycle ENUM('draft','published','archived') NOT NULL DEFAULT 'archived' AFTER enabled,
  ADD COLUMN seed_source ENUM('go_default','admin','restore','legacy') NOT NULL DEFAULT 'legacy' AFTER lifecycle,
  ADD COLUMN content_sha256 CHAR(64) NOT NULL DEFAULT '' AFTER content,
  ADD COLUMN published_at DATETIME(6) NULL AFTER updated_at,
  ADD COLUMN published_by_user_id BIGINT UNSIGNED NULL AFTER published_at,
  ADD COLUMN active_prompt_key VARCHAR(128)
    GENERATED ALWAYS AS (CASE WHEN lifecycle = 'published' AND enabled = 1 THEN prompt_key ELSE NULL END) STORED;

-- Normalize lifecycle while active_prompt_key is NULL for all legacy rows:
UPDATE generation_prompts
  SET lifecycle = 'archived',
      seed_source = 'legacy',
      content_sha256 = SHA2(content, 256);

UPDATE generation_prompts AS candidate
JOIN (
  SELECT prompt_key, MAX(version) AS version
  FROM generation_prompts
  WHERE enabled = 1
  GROUP BY prompt_key
) AS winner ON winner.prompt_key = candidate.prompt_key AND winner.version = candidate.version
SET candidate.lifecycle = 'published',
    candidate.published_at = candidate.updated_at;

UPDATE generation_prompts
  SET enabled = 0
  WHERE lifecycle <> 'published';

ALTER TABLE generation_prompts
  ADD KEY idx_generation_prompts_lifecycle (prompt_key, lifecycle, version),
  ADD UNIQUE KEY uq_generation_prompts_one_published (active_prompt_key),
  ADD CONSTRAINT fk_generation_prompts_publisher
    FOREIGN KEY (published_by_user_id) REFERENCES auth_users(id) ON DELETE SET NULL;

CREATE TABLE admin_audit_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  actor_user_id BIGINT UNSIGNED NOT NULL,
  capability VARCHAR(128) NOT NULL,
  action VARCHAR(64) NOT NULL,
  resource_type VARCHAR(64) NOT NULL,
  resource_id VARCHAR(191) NOT NULL,
  prompt_version_id BIGINT NULL,
  request_id VARCHAR(191) NOT NULL DEFAULT '',
  result VARCHAR(32) NOT NULL,
  summary_json JSON NOT NULL,
  content_sha256 CHAR(64) NOT NULL DEFAULT '',
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  KEY idx_admin_audit_actor_created (actor_user_id, created_at, id),
  KEY idx_admin_audit_resource_created (resource_type, resource_id, created_at, id),
  KEY idx_admin_audit_request_id (request_id),
  CONSTRAINT fk_admin_audit_actor FOREIGN KEY (actor_user_id) REFERENCES auth_users(id) ON DELETE RESTRICT,
  CONSTRAINT fk_admin_audit_prompt_version FOREIGN KEY (prompt_version_id) REFERENCES generation_prompts(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- +goose Down
DROP TABLE IF EXISTS admin_audit_logs;
ALTER TABLE generation_prompts
  DROP FOREIGN KEY fk_generation_prompts_publisher,
  DROP INDEX uq_generation_prompts_one_published,
  DROP INDEX idx_generation_prompts_lifecycle,
  DROP COLUMN active_prompt_key,
  DROP COLUMN published_by_user_id,
  DROP COLUMN published_at,
  DROP COLUMN content_sha256,
  DROP COLUMN seed_source,
  DROP COLUMN lifecycle;
```

The migration does not delete prompt rows. The corresponding Go startup initializer writes a default only when a registered key has no record, with `lifecycle='published'`, `seed_source='go_default'`, SHA-256 and a version chosen under a transaction. Historical migrations remain immutable release history; no new SQL migration will contain a default prompt body.

Publishing and restoring run in one transaction with a `SELECT ... FOR UPDATE` on the key's version rows. Publishing a draft atomically archives the prior published row, marks the target published, sets `published_by_user_id`, and inserts its audit row. Restoring never updates the selected historical row: it creates the next version with copied content, `seed_source='restore'`, publishes that new row, archives the prior published row, and inserts its audit row in the same transaction.

## Admin API contract

All paths are Go routes under `/api/v1/admin`. They require the existing HttpOnly cookie session. Read routes require the listed capability. Write routes require existing same-origin protection and the listed capability. A missing session returns `401 AUTH_UNAUTHENTICATED`; an authenticated caller without the exact capability returns `403 ADMIN_CAPABILITY_REQUIRED`; malformed input returns `400 ADMIN_INVALID_REQUEST`; absent resources return `404 ADMIN_NOT_FOUND`; an invalid state transition returns `409 ADMIN_PROMPT_STATE_CONFLICT`; transaction failure returns `500 ADMIN_OPERATION_FAILED` without sensitive details. `admin_audit_logs.request_id` is generated from the Go request context or a trusted server-side injector; no Admin request body may supply it.

| Method and path | Capability | Request | Success response |
|---|---|---|---|
| `GET /api/v1/admin/capabilities` | any `admin.*` | none | `{ "capabilities": ["admin.prompt.view"] }` from current effective set |
| `GET /api/v1/admin/prompts` | `admin.prompt.view` | optional `key` | `{ "prompts": [{ "id": 21, "key": "director.default", "version": 3, "lifecycle": "published", "seedSource": "admin", "contentSha256": "…", "publishedAt": "…" }] }` |
| `GET /api/v1/admin/prompts/{key}/versions/{version}` | `admin.prompt.view` | none | one prompt version including `content`; no secrets, provider response or session data |
| `POST /api/v1/admin/prompts/{key}/drafts` | `admin.prompt.edit` | `{ "content": "…" }` | `201 { "prompt": { "id": 22, "key": "…", "version": 4, "lifecycle": "draft" } }` |
| `PUT /api/v1/admin/prompts/{key}/drafts/{version}` | `admin.prompt.edit` | `{ "content": "…" }` | updated draft metadata and content hash |
| `POST /api/v1/admin/prompts/{key}/versions/{version}/publish` | `admin.prompt.publish` | empty body | `{ "prompt": { "id": 22, "lifecycle": "published" } }` |
| `POST /api/v1/admin/prompts/{key}/versions/{version}/restore` | `admin.prompt.publish` | empty body | `201` with a newly created published version; response identifies both source and new versions |

Only the detail endpoint with `admin.prompt.view` returns a prompt body. The list endpoint returns metadata and content hash only. The existing public `GET /api/v1/generation/prompts` must be narrowed to safe metadata or replaced by a non-body selection endpoint before Admin implementation exposes controlled bodies; it must never become an unauthenticated source of system prompt text.

## Audit redaction contract

`admin_audit_logs.summary_json` is built from an allowlist, never by serializing a request, a database row or an upstream error. Allowed fields are action, resource type/ID, prompt key, previous/new lifecycle, source version, new version, capability, request ID, result and a SHA-256 content hash.

The following values are excluded before persistence and excluded from every Admin response: prompt content, password or password hash, Cookie, session token/hash, access token, refresh token, API key, provider credential, recovery code, DSN, Authorization header, request body copied verbatim, provider raw response and raw stack trace. Failure audits record a stable error code and an allowlisted operation name, not an error string from MySQL or a provider.

## React `/admin` design

`后台` becomes a React + AntD single-page Admin application served at `/admin`. The Go web UI packaging path will embed the built Admin asset tree under `/admin` in addition to the existing user UI asset tree; an `/admin/*` browser refresh returns the Admin entry document while `/api/*` remains API-only.

The shell calls `GET /api/auth/current-user`, then `GET /api/v1/admin/capabilities`. It never infers authorization from a role string. It has these states:

| State | Rendered result |
|---|---|
| signed out | sign-in guidance; API still returns 401 until authenticated |
| signed in with no `admin.*` | explicit “无管理后台权限” page; no admin navigation |
| signed in with one or more capabilities | shell plus only the menus allowed by those capabilities |
| API returns 403 after a permission change | inline “权限已变更” state, refresh current-user/capabilities, retain no stale data |
| load, empty or failure | AntD loading, empty and retry states; failures show stable safe messages only |

First-batch menu mapping is: 首页 (`admin.dashboard.view`) and 提示词库 (`admin.prompt.view`). 提示词库 shows versions; 草稿按钮 requires `admin.prompt.edit`; 发布与恢复按钮 require `admin.prompt.publish`. Future disabled-by-absence menu mappings are 模型配置 (`admin.model.view`), 技能 (`admin.skill.view`), 审计/错误日志 (`admin.audit.view`) and 成员与权限 (`admin.member.view`). Task 7 owns only the `/admin` application shell and its menus; Task 2 later decides whether and how `UserShell` exposes an entry by consuming this capability contract.

## Test design and acceptance evidence

Tests are written before product code and verified red before each implementation unit.

1. `authn` unit tests: `owner`, `dev` and legacy `admin` receive the complete eleven-capability set; `manager` has none until an explicit capability is merged; `member` has none; duplicate grants are removed; unknown roles grant none.
2. HTTP middleware tests: an authenticated `member` calling every first-batch Admin route receives `403 ADMIN_CAPABILITY_REQUIRED`; a `manager` with only `admin.prompt.view` can list/read but receives 403 on draft/publish; `owner` succeeds because its effective capability set contains the requirement, not because middleware reads its role.
3. MySQL migration contract tests: existing multiple enabled prompt versions normalize to one published version per key; all rows survive; the unique generated key rejects a second published row; audit foreign keys and indexes exist.
4. Store transaction tests: draft is not returned by runtime resolution; publish changes exactly one active version and creates one redacted audit record; restore creates a higher new version while leaving the source immutable; failed audit insertion rolls back the prompt state change.
5. Seed tests: missing registered keys are inserted from `generation.DefaultPrompts()` with `go_default`; an existing administrator-created or restored record is never overwritten.
6. Handler contract tests: request/response schemas, capability checks, same-origin write protection, stable error codes, no client-provided request ID field, server-context request ID audit attribution, and no prompt content in list responses.
7. Audit redaction tests: adversarial prompt body and request values containing `Authorization`, cookie-like strings, tokens, passwords, DSNs and provider response text do not appear in `summary_json`, API responses or error messages; only the SHA-256 is stored.
8. React tests: no capability renders the no-permission page; a view-only manager sees the prompt menu without edit/publish controls; refresh reloads effective capabilities and prompt data; a 403 invalidates the visible admin state; no system prompt literal exists in `后台/src`.
9. Build gates: `go test ./...` from `api`, `npm test && npm run build` from `前台`, and `npm run build` from `后台`; source scan verifies no Admin API implementation imports a Node Admin route.

## Second-batch boundary

Second batch implements model configuration, skills, audit/error-log read views and member capability management using the already-established Admin shell, effective capability mechanism and append-only audit store. Model responses expose only configured/non-configured state, provider name and masked identifier; they never expose credentials. Error views consume only redacted persisted events. Member management grants or revokes individual `admin.*` capabilities, records both actor and target in `admin_audit_logs`, then immediately recomputes the target user's effective capability set for every active session or revokes the target's active sessions. A browser refresh alone is not permission propagation or revocation evidence.
