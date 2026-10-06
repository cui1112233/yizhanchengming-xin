# 小说获取最小完整流程：A/B 共享接入说明

更新时间：2026-10-07

本轮只打通“登录与权限 → 提交 Book ID → 121 正文获取 → MySQL 保存/查询 → 刷新恢复 → 单书失败重试”的最小完整流程。B 的多版本改文、知识库、处理记录、定时 Runtime、批量工厂交接和发布意图范围不缩减，继续在固定 B 分支推进；A 不复制这些业务实现。

## 1. B 当前接口 / 持久化缺口

| 环节 | B 分支现状 | A 共享主线处理 |
| --- | --- | --- |
| 登录与权限 | B 模块 handler 独立存在，未挂全局 Auth / owner-team | 复用 `authn.User`、capability、same-origin；新增 Intake owner/team ownership |
| 提交 Book ID | B 有 `CreateBatchInput` / `CreateBatch` | 最小闭环继续使用现有 `POST /api/v1/intakes`，不建第二套 batch 表 |
| 真实正文 | B 只有 `Fetcher` interface | 生产仍复用 `provider121.Client`，由 `cmd/server -> app.NewHandler` 注入 |
| MySQL | B 目前只有 Store contract / 测试替身 | 复用已有 `intakes` + `books`；正文事实是 `books.original_text` |
| 查询 / 刷新恢复 | B 有模块 readback，但无 durable Store | 使用受权限保护的 Intake/Books API；URL 只保存 intake id 提示，不保存正文/状态 |
| 失败重试 | B 有 Retry 业务设计 | 新增现有 Intake 单书 retry；正文已落库时不重新抓 121 |
| 文本改写 | B 有 TextModel boundary | 不在本轮伪接；等待 A 的共享文本模型 contract |
| 定时/队列 | B 有 Dispatcher boundary | 不在本轮另建 Redis Runtime；后续复用现有 taskruntime |
| TOS | B 尚未接 | 本轮正文仍在 MySQL；不新建 B 私有 TOS client |
| 发布 | B 只创建 Intent | 保持边界，不执行真实发布 |

## 2. 权威持久化

现有事实表继续作为最小闭环权威来源：

- `intakes`：一次小说获取批次；
- `books`：Book ID、书城、platformId、标题、正文、元数据、状态、错误；
- `books.original_text`：当前正文持久事实；
- `auth_intake_ownership`：本轮新建，仅记录 Intake owner/team，不复制正文或任务状态。

Migration：

- `00009` 仍保留给 Shuihuo B1；
- `00010_intake_ownership.sql` 分配给 A 的 Intake ownership；
- 分配 `00010` 前已重新检查 B/C/D、Shuihuo、安全、性能和其它仍有领先提交的活动分支，没有发现活动分支占用 `api/db/migrations/00010_*.sql`。

遗留 Intake 没有 ownership 行时：

- admin/owner 可通过 elevated 规则访问；
- 普通成员默认不可见、不可执行、不可重试；
- A 不猜测历史 owner，避免跨用户泄露。

## 3. API 契约

前端统一复用 `前台/src/api.js::requestJSON`。不要另建 Token/localStorage Auth client。

### 创建 Book ID 批次

`POST /api/v1/intakes`

- capability：`batch.configure`
- mutation：要求 same-origin
- auth 开启时：创建成功后写 `auth_intake_ownership`
- 请求：沿用现有 `CreateIntakeInput` JSON
- 返回：沿用现有 `{ intake, books }`

### 读取当前用户可见 Intake

`GET /api/v1/intakes`

- capability：`batch.view`
- 普通用户只返回 owner/team 可访问的 Intake
- admin/owner 可看现有全部 Intake

### 执行真实 121 获取

`POST /api/v1/intakes/{id}/execute`

- capability：`batch.execute`
- ownership：必须通过 Intake owner/team 检查
- 服务：现有 `intake.Service.ExecuteIntake`
- Fetcher：现有 `provider121.Client`
- 成功正文落 `books.original_text`
- 单书失败落 `retryable_failed + error_message`

### 列出书籍 / 刷新恢复

`GET /api/v1/intakes/{id}/books`

- capability：`batch.view`
- ownership：必须通过
- 返回不包含完整正文，适合列表/状态恢复

前端创建 Intake 后把 `?intake=<id>` 写进当前 URL。刷新后只用该 ID 重新请求 MySQL API；URL 不是业务事实源。

### 单书正文详情

`GET /api/v1/intakes/{id}/books/{bookId}`

- capability：`batch.view`
- ownership：必须通过
- 新接口错误格式：`{ code, message, request_id }`
- `{bookId}` 是 MySQL `books.id`（内部数值 ID），不是外部小说 Book ID；外部 ID 仍在响应字段 `bookId` 中。
- 返回：`{ book: { ...列表字段, originalText } }`

### 单书失败重试

`POST /api/v1/intakes/{id}/books/{bookId}/retry`

请求：

```json
{ "maxText": 4000 }
```

- capability：`batch.execute`
- mutation：要求 same-origin
- ownership：必须通过
- 仅允许未成功书；
- 若 `original_text` 为空，重新请求 121；
- 若正文此前已保存、失败发生在分类等后续阶段，不重复请求 121；
- 成功后重新聚合 Intake 状态；
- 已成功书返回 `409 INTAKE_BOOK_NOT_RETRYABLE`；
- 重试仍失败返回明确失败，不返回伪成功。

## 4. B 如何接入

B 不要实现第二套 MySQL Store 作为公网最小流程事实源。B 最新 `前台/src/novelfetch/client.js` 仍默认请求 `/api/v1/novel-fetch`。最小完整流程不要把该 prefix 注册成第二套事实 API；B 的 UI/client 应直接从 `前台/src/api.js` 引用下列共享函数。完整 config/knowledge/history/handoff/submit-intent 能力继续留在 B 模块等待后续集成。

“处理/任务”最小链路映射到：

- `createIntake(...)`
- `executeIntake(intakeId, maxText)`
- `listBooks(intakeId)`
- `getIntakeBook(intakeId, bookId)`
- `retryIntakeBook(intakeId, bookId, maxText)`

对应前端函数均来自 `前台/src/api.js`，底层统一是 `requestJSON`。

B 自己的 `api/internal/novelfetch` 仍保留作为后续完整业务能力设计，不删除；等共享 TextModel / Runtime / Publish / BatchFactory 边界明确后再由 A 集成，不能为了“先跑起来”创建第二套 provider、queue、credential 或 storage。

## 5. 当前接通状态

| 能力 | 接口已定义 | 实现已接通 | 运行已验证 |
| --- | --- | --- | --- |
| Auth + Intake ownership | 是 | 是 | 否 |
| Book ID 提交 | 已有 | 已有 | 本轮未执行 |
| 121 Fetcher 生产注入 | 已有 | 已有 | 真实 121 未执行 |
| MySQL 正文保存 | 已有 | 已有 | 本轮未执行 |
| 单书 MySQL 详情 | 是 | 是 | 本轮未执行 |
| URL 刷新恢复 | 是 | 是 | 本轮未执行 |
| 单书失败重试 | 是 | 是 | 本轮未执行 |
| B 多版本 TextModel | B 已定义 | 尚未接共享实现 | 否 |
| B Dispatcher/定时 Runtime | B 已定义 | 尚未接共享实现 | 否 |
| B TOS 大对象 | 未统一 | 未接 | 否 |

“运行未验证”不等于失败，只表示当前执行环境没有完成对应真实运行验证。

## 6. 本地独立开发环境运行

要求：

- Go 1.23.x
- MySQL 8.4+
- Node 22
- Goose v3.24.3

不要使用生产账号、生产 DSN、真实发布凭据。下面都是本地占位示例。

```bash
cd api

go install github.com/pressly/goose/v3/cmd/goose@v3.24.3

export QIANTIE_MYSQL_DSN='local_user:local_password@tcp(127.0.0.1:3306)/ycm_dev?parseTime=true&multiStatements=true'
export QIANTIE_BOOTSTRAP_ADMIN_USERNAME='local-admin'
export QIANTIE_BOOTSTRAP_ADMIN_PASSWORD='replace-with-a-local-strong-password'
export QIANTIE_ENV='local'

goose -dir db/migrations mysql "$QIANTIE_MYSQL_DSN" up
go test ./internal/intake ./internal/sharedplatform ./internal/httpapi ./internal/app -count=1
go test ./...
go build ./...
go run ./cmd/server
```

用户端：

```bash
cd 前台
npm install --no-audit --no-fund
npm test
npm run build
npm run dev
```

浏览器验证顺序：

1. 本地管理员登录；
2. 打开 `/novel-fetch`；
3. 添加书城、platformId、Book ID；
4. 执行；
5. 查看书籍状态；
6. 用单书详情 API 确认正文来自 MySQL；
7. 刷新页面，确认 URL 中的 intake id 能重新读回 MySQL 状态；
8. 对 `retryable_failed` 单书执行重试；
9. 不执行真实发布，不接生产凭据。

## 7. 本轮测试环境限制

当前 ChatGPT 独立容器具备 Go 1.23.2，但容器 DNS 无法解析 GitHub，不能 clone 完整仓库；连接器也没有把完整 repository materialize 到容器的能力。因此本轮不能把 `go test ./...`、`npm test`、build 标记为通过。

GitHub Actions 已被用户明确禁止。A 分支后续提交必须带 `[skip ci]`，测试应由可执行的本地/独立开发环境完成并回报对应 SHA。
