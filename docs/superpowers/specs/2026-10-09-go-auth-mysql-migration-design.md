# Go 账号与登录 MySQL 迁移设计

## 目标

让 MySQL 成为唯一账号来源，Go 接管账号密码登录、账号管理和会话。React 前端接口地址保持不变，Node 在过渡期只转发账号相关请求，最终可删除 Node 的账号模块。

## 本轮范围

迁移到 Go：

- 账号创建；
- 账号启用、禁用；
- 密码修改；
- 账号密码登录；
- 退出登录；
- 查询当前登录用户；
- 创建、校验和注销登录会话；
- 首次启动时导入旧 `accounts.json`。

本轮不迁移、不删除：

- MFA 动态码；
- Passkey 登录；
- 团队邀请；
- 团队成员资料、角色和 API 授权；
- 前端页面；
- 公网部署。

## 数据结构

### accounts

| 字段 | 用途 |
|---|---|
| username | 唯一登录账号 |
| password_hash | 密码校验值，不保存明文密码 |
| password_algorithm | 密码校验值格式，例如旧 `scrypt` 或新的 Go 格式 |
| is_owner | 是否所有者账号 |
| active | 是否允许登录 |
| created_at | 创建时间 |
| updated_at | 更新时间 |

### auth_sessions

| 字段 | 用途 |
|---|---|
| token_hash | 登录 Token 的 SHA-256 哈希；原 Token 不存数据库 |
| username | 会话所属账号 |
| expires_at | 过期时间 |
| issued_at | 签发时间 |
| browser | 浏览器名称 |
| os | 系统名称 |
| ip_hint | 脱敏 IP 提示 |

### auth_migrations

| 字段 | 用途 |
|---|---|
| migration_key | 固定迁移标识，防止重复导入 |
| completed_at | 完成时间 |
| source_summary | 只保存导入数量等摘要，不能保存密码校验值或 Token |

## 首次启动导入

1. Go 启动时检查 `auth_migrations` 是否已有旧账号导入记录。
2. 没有记录时，读取配置指定的旧 `accounts.json`。
3. 仅导入账号名、旧 scrypt 密码校验值、所有者标记、启用状态和原创建/更新时间。
4. 在同一个数据库事务中写入账号和迁移记录。
5. 迁移记录写入成功后，后续启动绝不再读取旧文件。
6. 旧文件不删除，只保留为只读核对备份。
7. 导入失败时不写迁移完成记录，下一次启动可安全重试。

旧 Node 会话和 Token 不导入。迁移切换后，所有用户都必须使用原密码重新登录。

## 密码规则

- 密码明文只存在于本次登录或改密请求内，绝不写日志或数据库。
- 旧账号的 `scrypt$...` 校验值可由 Go 兼容验证，确保用户无需重设密码。
- 密码登录成功后，若当前仍是旧格式，可在同一账户记录中升级为 Go 的目标密码格式；升级失败不应阻止本次已经验证成功的登录，但必须记录安全错误。
- 新建账号和密码修改直接使用 Go 目标密码格式。
- 密码校验必须使用常量时间比对。

## 接口兼容

前端现有地址不变：

```text
POST /api/login
POST /api/login/logout
GET  /api/login/session
```

账号管理接口继续使用现有路径。过渡期请求路径为：

```text
React → Node 原路径 → Go → MySQL
```

Node 不再创建、修改或读取账号文件中的账号数据；它只验证现有前端认证格式并转发给 Go。Go 返回的登录字段必须兼容当前前端：

```text
token
username
displayName
avatarUrl
role
boundTo
monthlyTokenLimit
apiEnabled
active
isOwner
effectivePermissions
```

本轮尚未迁移的团队资料字段，过渡期由 Node 读取后与 Go 的基础账号会话结果合并；Node 不得覆盖 Go 的 `username`、`active`、`isOwner` 和权限结论。

## 安全与会话

- 登录失败统一返回“用户名或密码错误”，不暴露账号是否存在或是否禁用。
- 禁用账号不能创建新会话；已有会话在下次校验时也应失效。
- `remember=true` 写入持久会话，`remember=false` 的会话仅存在当前 Go 运行期间。
- 退出登录只撤销当前 Token。
- 会话校验每次都检查账号仍然启用。
- Go 与 Node 过渡桥接继续使用现有本地桥接签名，不能使用公网地址或默认密钥。

## 测试和验收

必须先写失败测试，再写最小实现。至少验证：

1. MySQL 表和唯一约束正确；
2. 旧账号只导入一次，重复启动不会重复导入；
3. 旧 scrypt 密码可登录，错误密码失败；
4. 新密码不会以明文、Token 不会以原文写入 MySQL；
5. 启用、禁用、改密和创建账号行为正确；
6. 登录、查会话、退出登录沿用现有前端接口字段；
7. 旧 Node Token 在切换后不能使用；
8. Node 转发只访问本地 Go，不能访问公网项目服务器；
9. MFA、Passkey、团队邀请路由不因本轮迁移被删除或修改；
10. Go、Node 定向测试和全量回归通过。

## 完成边界

本轮完成后，账号密码登录的账号来源、密码校验和会话来源都在 Go + MySQL。Node 仍会保留 MFA、Passkey、团队邀请与团队资料兼容层，直到这些功能也单独迁移并验证完成。