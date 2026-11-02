# Go 账号与登录 MySQL 迁移实施计划

> **给执行代理：**按任务顺序实施。每步先写失败测试、确认失败、再写最小实现；不创建分支，不推送，不提交未经用户确认的改动。

**目标：**让 Go + MySQL 成为账号、密码和基础会话的唯一来源，同时保持 React 前端现有登录和账号管理接口不变。

**架构：**Go 在现有 MySQL 迁移体系中创建账号、会话和一次性导入记录；首次运行只读导入 Node 的旧 `accounts.json`，随后不再读取它。Node 保留原 URL 和未迁移的 MFA、Passkey、团队功能，但账号密码登录及账号变更请求改为本地 Go 桥接。

**技术栈：**Go、MySQL、现有 `storage.Migrations`、Node/Express、React（接口不改）、scrypt 兼容验证、SHA-256 会话 Token 哈希。

## 全局约束

- 唯一维护分支是 `v88`；不创建新分支，不推送公网。
- 不读取、输出或提交 `.env.v88-review`、本地密钥、Token、Cookie、MySQL 真实数据。
- MySQL 是迁移完成后的唯一账号来源；Node 不得继续写 `accounts.json` 或 `sessions.json`。
- 旧 `accounts.json` 只允许首次启动时只读导入；写入完成标记后绝不重复读取；旧文件不删除。
- 不迁移旧 Node Token；切换后用户使用原密码重新登录。
- 不保存明文密码或原始会话 Token；密码和 Token 不得进入日志、审计详情或 HTTP 响应。
- 账号密码接口保持：`POST /api/login`、`POST /api/login/logout`、`GET /api/login/session`。
- MFA、Passkey、团队邀请、成员资料、团队角色、本轮不迁移、不删除、不改前端页面。
- Node → Go 只允许 `http://127.0.0.1:4000`、`http://localhost:4000` 或 Docker `http://backend:4000`，并使用现有桥接签名；禁止公网目标和默认桥接密钥。

---

### Task 1：创建账号与会话 MySQL 迁移

**文件：**
- 修改：`backend/internal/storage/*schema*.go` 或现有迁移清单定义文件
- 修改：`backend/internal/storage/migrations_test.go`
- 创建：`backend/internal/auth/mysql_schema_test.go`

**产生的表：**

```sql
accounts(username VARCHAR(32) PRIMARY KEY, password_hash TEXT NOT NULL,
 password_algorithm VARCHAR(32) NOT NULL, is_owner BOOLEAN NOT NULL,
 active BOOLEAN NOT NULL, created_at DATETIME(6) NOT NULL, updated_at DATETIME(6) NOT NULL)

auth_sessions(token_hash CHAR(64) PRIMARY KEY, username VARCHAR(32) NOT NULL,
 expires_at DATETIME(6) NOT NULL, issued_at DATETIME(6) NOT NULL,
 browser VARCHAR(80) NULL, os VARCHAR(80) NULL, ip_hint VARCHAR(64) NULL,
 INDEX auth_sessions_username_expires(username, expires_at),
 FOREIGN KEY (username) REFERENCES accounts(username))

auth_migrations(migration_key VARCHAR(128) PRIMARY KEY,
 completed_at DATETIME(6) NOT NULL, source_summary JSON NOT NULL)
```

- [ ] 写失败测试：迁移计划包含递增的新版本，包含三张表、`accounts.username` 主键、`auth_sessions.token_hash` 主键、会话账号索引和外键。
- [ ] 运行：`go test ./internal/storage -run AuthSchema -count=1`；预期失败，因为迁移尚不存在。
- [ ] 在现有 `storage.Migrations()` 体系追加一条新迁移；不得修改已记录迁移的 SQL。
- [ ] 再运行定向测试和 `go test ./internal/storage -count=1`；预期通过。

### Task 2：实现 Go 账号、密码与会话存储

**文件：**
- 创建：`backend/internal/auth/types.go`
- 创建：`backend/internal/auth/password.go`
- 创建：`backend/internal/auth/store.go`
- 创建：`backend/internal/auth/mysql_store.go`
- 创建：`backend/internal/auth/password_test.go`
- 创建：`backend/internal/auth/mysql_store_test.go`

**接口：**

```go
type Account struct { Username string; PasswordHash string; PasswordAlgorithm string; IsOwner bool; Active bool; CreatedAt time.Time; UpdatedAt time.Time }
type Session struct { Token string; Username string; ExpiresAt time.Time; IssuedAt time.Time; Browser string; OS string; IPHint string }
type Store interface {
  FindAccount(ctx context.Context, username string) (Account, error)
  CreateAccount(ctx context.Context, account Account) error
  UpdatePassword(ctx context.Context, username, hash, algorithm string) error
  SetActive(ctx context.Context, username string, active bool) error
  CreateSession(ctx context.Context, session Session) error
  FindSession(ctx context.Context, rawToken string, now time.Time) (Session, error)
  RevokeSession(ctx context.Context, rawToken string) error
}
```

- [ ] 写失败测试：Go 可验证旧 `scrypt$<salt hex>$<hash hex>` 格式；错误密码失败；新密码格式不会返回或记录明文。
- [ ] 写失败测试：创建会话时数据库接收 SHA-256 Token 哈希，不能接收原 Token；读取过期会话返回不存在。
- [ ] 写失败测试：禁用账号后，其已有会话不能通过账号校验。
- [ ] 运行 `go test ./internal/auth -count=1`，确认失败。
- [ ] 实现用户名、密码长度和状态校验；复用常量时间比较；新密码使用明确的 Go 算法标识，旧 scrypt 能兼容验证。
- [ ] 实现 MySQL 参数化查询、会话哈希、会话撤销和过期清理；不将 password/token 放到错误文本。
- [ ] 运行 `go test ./internal/auth -count=1`，确认通过。

### Task 3：实现一次性旧账号自动导入

**文件：**
- 创建：`backend/internal/auth/legacy_import.go`
- 创建：`backend/internal/auth/legacy_import_test.go`
- 修改：`backend/internal/config/config.go`
- 修改：`backend/internal/config/config_test.go`
- 修改：`backend/internal/app/app.go`

**接口：**

```go
type LegacyAccountFile struct { Accounts []LegacyAccount `json:"accounts"` }
type LegacyAccount struct { Username string `json:"username"`; PasswordHash string `json:"passwordHash"`; IsOwner bool `json:"isOwner"`; Active bool `json:"active"`; CreatedAt string `json:"createdAt"`; UpdatedAt string `json:"updatedAt"` }
func ImportLegacyAccountsOnce(ctx context.Context, store Store, sourcePath string) error
```

- [ ] 写失败测试：有效旧 JSON 只导入账号名、旧 scrypt 哈希、owner、active、时间；不导入会话。
- [ ] 写失败测试：同一迁移键第二次运行不再次打开或导入旧文件。
- [ ] 写失败测试：导入中出现无效记录时整个事务回滚，且不写完成标记。
- [ ] 运行 `go test ./internal/auth -run LegacyImport -count=1`，确认失败。
- [ ] 增加仅用于本地路径的 `QIANTIE_LEGACY_ACCOUNTS_PATH` 配置；路径为空时不执行导入，不得猜测生产文件位置。
- [ ] 实现账号文件严格 JSON 校验、单事务导入和 `auth_migrations` 完成摘要；摘要只记录数量和来源文件名，不能记录哈希或密码。
- [ ] 在 `app.Build` 完成迁移后调用导入；导入失败时 Go 启动失败，不能带着半迁移状态运行。
- [ ] 运行定向 auth/config/app 测试，确认通过。

### Task 4：提供 Go 基础登录 HTTP 接口

**文件：**
- 创建：`backend/internal/httpapi/auth.go`
- 创建：`backend/internal/httpapi/auth_test.go`
- 修改：`backend/internal/httpapi/router.go`
- 修改：`backend/internal/app/app.go`

**接口：**

```text
POST /api/login              {username,password,remember} → {token,username,active,isOwner,effectivePermissions,...}
POST /api/login/logout       Bearer token → 204
GET  /api/login/session      Bearer token → 当前账号基础信息
```

- [ ] 写失败测试：正确密码登录返回随机 Token；错误密码、未知账号、停用账号统一返回 401 和“用户名或密码错误”。
- [ ] 写失败测试：`remember=true` 持久会话有明确过期时间；`remember=false` 使用内存会话并在 Go 重启后失效。
- [ ] 写失败测试：当前用户、退出登录、已退出 Token、停用账号 Token 的行为正确。
- [ ] 写失败测试：响应和错误内容不包含密码、password_hash、token_hash。
- [ ] 运行 `go test ./internal/httpapi -run Auth -count=1`，确认失败。
- [ ] 实现带限长 JSON 解析、账号密码验证、Token 生成、会话创建、Bearer 解析、注销和会话查询。
- [ ] 先返回基础兼容字段：`token`、`username`、`active`、`isOwner`、`effectivePermissions`；未迁移的资料字段由 Node 兼容层补充。
- [ ] 运行 `go test ./internal/httpapi ./internal/auth -count=1`，确认通过。

### Task 5：实现 Go 账号创建、改密、启用禁用接口

**文件：**
- 修改：`backend/internal/httpapi/auth.go`
- 修改：`backend/internal/httpapi/auth_test.go`
- 创建或修改：`backend/internal/auth/service.go`
- 创建或修改：`backend/internal/auth/service_test.go`

**接口：**

```text
POST  /api/account-admin/accounts
PATCH /api/account-admin/accounts/{username}
```

- [ ] 写失败测试：仅 owner 可创建账号、改密码、启用和禁用；普通账号返回 403。
- [ ] 写失败测试：新账号由 Go 密码格式保存；禁用账号不能登录；改密后旧密码失效、新密码生效。
- [ ] 写失败测试：不支持的团队字段（role、boundTo、quota 等）不能假装写成功，返回明确“该字段仍由 Node 团队模块管理”。
- [ ] 运行 auth/httpapi 定向测试，确认失败。
- [ ] 实现最小账号管理接口和 owner 鉴权；只处理本轮账号字段，团队字段不迁移。
- [ ] 运行定向测试，确认通过。

### Task 6：把 Node 原登录与账号接口改为 Go 转发

**文件：**
- 修改：`routes/auth.js`
- 修改：`routes/account-admin.js`
- 修改：`middleware/auth.js`
- 修改：`app.js`
- 创建或修改：`test/go-auth-bridge.test.js`
- 创建或修改：`test/account-admin-go-bridge.test.js`

- [ ] 写失败测试：`POST /api/login`、`POST /api/login/logout`、`GET /api/login/session` 使用现有 Node URL 但请求本地 Go；桥接目标只允许 127.0.0.1、localhost、backend。
- [ ] 写失败测试：Node 不再调用 `accountStore.verifyPassword`、`createPersistentSession`、`sessions.json` 完成账号密码登录。
- [ ] 写失败测试：账号创建、改密、启用禁用转发 Go；Node 团队资料字段仍保持现有逻辑且不会覆盖 Go 的账号状态。
- [ ] 运行 Node 测试，确认失败。
- [ ] 实现桥接签名和响应映射：Node 只补充尚未迁移的展示资料；不能覆盖 Go 的 username、active、isOwner、会话判断。
- [ ] 保留 MFA、Passkey、邀请等未迁移路由；它们若依赖旧账号源，应明确返回“迁移期间暂不可用”，不得读写第二套账号数据。
- [ ] 运行 Go 定向测试、Node 定向测试。

### Task 7：本地切换验证和回归边界检查

**文件：**
- 修改：不创建业务文档；只按现有测试与启动配置验证

- [ ] 使用独立空 MySQL 测试库，配置 `QIANTIE_LEGACY_ACCOUNTS_PATH` 指向测试夹具，启动 Go，验证只导入一次。
- [ ] 验证旧 Node Token 返回 401；使用旧账号原密码通过 Node 原 URL 登录；使用同一 Token 查询 session；退出后 token 失效。
- [ ] 验证新账号创建、改密、禁用/启用均经 Go + MySQL 生效。
- [ ] 验证 MFA、Passkey、邀请路由未被删除；若无法安全与 Go 账号源并存，返回明确迁移期状态而不是写入旧账号文件。
- [ ] 运行 `go test ./...`（在 `backend/`）；运行相关 Node 测试；运行 `git diff --check`。
- [ ] 检查 `git status --short`：不得包含 `.env.v88-review`、`.local-review/`、真实 MySQL 数据、Token、构建产物。
- [ ] 用户确认后才单独提交本轮源码和测试；未确认不得提交或推送。
