# zhangying 登录异常只读排查报告

- 日期：2026-11-05（CST）
- 线上：`https://yzcm.site`（115.190.156.223，nginx 1.24.0，后端 Go，`qiantie.service`）
- 约束：只读。未改库、未重置密码、未重启服务、未改代码、未 commit。
- 线上版本：`/api/runtime-build-info` 返回 `runtime_manifest_sha=32a3c6ad...`、`frontend_entry=/assets/user-Cs6GUvIU.js`，与 13 版部署记录一致。

## 0. 结论先行

1. **当前 V88 主站不存在「登录账号错误」这条文案。** 本地登录只有一个接口 `POST /api/login`，所有账号类失败统一返回 **「用户名或密码错误」**。源码、git 全历史、线上入口及其引用的 84 个 JS chunk 均搜不到「登录账号错误」。所以：
   - 如果用户看到的是原文，它**不是 yzcm.site 当前登录接口返回的**，需要确认实际页面 URL（121 后台、旧入口或旧书签）。
   - 如果只是转述，它对应的就是统一文案「用户名或密码错误」；**不能据此推断“账号查找阶段失败”**，因为该文案把账号不存在、禁用、哈希异常、密码错误等分支全部吞掉了。

2. **线上 `zhangying` 账号存在且启用，密码校验确实执行了，失败点是“密码不匹配”。**
   用故意错误的密码调用线上 `/api/login`，`zhangying` 比不存在的账号稳定多约 **140–150ms**（多轮重复；单次公网抖动可到 13s，已剔除）：
   - `zhangying`：多轮落在 **0.319–0.359s**（中位数约 0.342s）
   - 已知存在的对照账号 `choushiyiguai`：多轮平均约 **0.335s**
   - 大小写变体 `ZhangYing`：多轮平均约 **0.343s**
   - 手机号 `13800000000`：多轮平均约 **0.198s**
   - 随机不存在账号：多轮平均约 **0.191–0.203s**

   代码只在“账号找到且 `active=true`”之后才调用 scrypt（[auth.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/auth.go#L77-L84)、[password.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/auth/password.go#L30-L49)）。多出的约 150ms 正是 scrypt 计算。→ **排除：账号不存在、账号被禁用、哈希格式不可解析、DB 查询失败**。这是时序侧信道的**强推断**，不是直接读线上库。

3. **本地 2026-11-03 全库备份与线上结论一致。** `qiantie_v88.accounts` 中 `zhangying`：
   - `active=1`、`is_owner=0`、`password_algorithm=scrypt-v1`
   - `password_hash` 前缀 `scrypt$`，长度 104（未输出完整哈希）
   - `created_at=2026-09-02 07:06:12`，`updated_at=2026-10-16 12:51:02`
   - `auth_sessions` 有 4 条记录；最近一次 `issued_at=2026-10-16 12:51:50`，`expires_at=2026-11-15 12:51:50`
   - 该备份之后没有更新的成功登录会话（备份本身截至 11-03）

4. **最可能原因排序（详见第 6 节）**
   - **A（强）**：用户提交的密码与 2026-10-16 之后保存的哈希不匹配。账号在、启用、哈希可解析，只是密码字节不对。
   - **B（中）**：用户输入的“账号”不是 username `zhangying`（例如手机号/邮箱/其他字符串）。账号表只有 `username` 一列，手机号实测查不到，统一文案会掩盖这一点。
   - **C（中）**：用户实际在另一个登录入口登录（121 后台/旧系统/旧书签），该入口与 yzcm 主站不是同一套账号。
   - **D（弱）**：账号在 2026-11-03 之后被禁用或删除。时序测试已基本排除“禁用”，只剩“被删除/改库”这种小概率。
   - **E（极弱）**：哈希算法不匹配或近两周代码改坏了登录。全库 20 个账号全部 scrypt，近两周没有相关逻辑改动。

5. **最小且安全的恢复动作**
   - 先确认用户实际访问的 URL，以及输入框里到底填了什么（截图最可靠）。
   - 若确实在 `yzcm.site` 主站登录、输入 `zhangying`：让用户使用 **2026-10-16 之后设置的密码**；忘记则由管理员重置，或走邮箱找回（需已绑定并验证邮箱）。
   - 若用户填的是手机号/邮箱：改为输入 `zhangying`（大小写不敏感）。
   - 若用户在 121/旧系统：使用那边的账号；旧库中 `zhangying` 的密码是占位符 `bridge-managed-account`，不能用于密码登录。
   - **不需要改代码、不需要改库、不需要重启服务。**

---

## 1. 「登录账号错误」文案定位

### 1.1 当前系统只有一个本地登录接口

- 路由：[auth.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/auth.go#L57-L115) 注册 `POST /api/login`。
- 前端登录表单：[UserLayout.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/shared/layouts/UserLayout.jsx#L388-L415)
- 前端错误展示：`message.error(error.message || '登录失败')`（[UserLayout.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/shared/layouts/UserLayout.jsx#L315-L329)），即错误文字由后端返回。

### 1.2 分支与文案

| 分支 | 代码位置 | 实际返回 |
|---|---|---|
| JSON 解析失败 / 账号为空 / 密码为空 | [auth.go:73-75](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/auth.go#L73-L75) | 401 `用户名或密码错误` |
| `FindAccount` 数据库错误（非 NotFound） | [auth.go:78-80](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/auth.go#L78-L80) | 503 `登录服务暂不可用` |
| 账号不存在 / `active=false` / 密码校验失败 | [auth.go:82-84](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/auth.go#L82-L84) | 401 `用户名或密码错误` |
| MFA 未通过 | [auth.go:86-96](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/auth.go#L86-L96) | 428/503，独立文案 |
| `CreateSession` 写库失败 | [auth.go:108-110](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/auth.go#L108-L110) | 401 `用户名或密码错误` |

**结论**：`用户名或密码错误` 至少吞掉了 4 类失败：账号不存在、账号禁用、哈希格式/算法不识别、密码不匹配；另外还吞掉了“会话写入失败”。`VerifyPassword` 只返回 `bool`，没有 error 可区分（[password.go:30-49](file:///f:/脚本测试/一战晟铭-v88/backend/internal/auth/password.go#L30-L49)）。因此单看这句话无法判断卡在哪个阶段。

### 1.3 「登录账号错误」不存在于本系统

已核对：

- 当前源码：无精确匹配。
- `git log --all -S "登录账号错误"` / `-S "登录帐号错误"`：无结果。
- 线上入口 bundle 及其引用的 84 个 JS chunk：无精确匹配。
- 线上 `/api/login` 实测：只返回 `{"error":"用户名或密码错误"}`（37 字节）。
- 121 后台登录页（`http://two.121w.com/tttadmin/login.php`）实测：`zhangying` 与随机不存在账号，错误区显示的都是「用户名或密码错误」，也不是「登录账号错误」。
- 121 后台页面自身只是展示服务端 `message`，没有本地写死该文案。

因此，这条文案要么是用户对「用户名或密码错误」的转述，要么来自本仓库之外的另一个登录页面。

### 1.4 唯一一处“账号…错误”文案

应用内只有 121 上传登录的固定文案：

- [novel_fetch_upload_compat.go:91](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/novel_fetch_upload_compat.go#L88-L92) 返回 `视频管理系统账号或密码错误`。

如果用户是在“小说上传 / 121 视频管理系统登录”卡片里操作，看到的应是这句，也不是「登录账号错误」。

---

## 2. 账号数据（2026-11-03 本地全库备份）

> 备份文件：`.backups/mysql-full-20261103T114757Z.sql.gz`。这是本地离线快照，不能等同于线上实时库；但账号结构、账号名、哈希格式与线上时序结论一致。

### 2.1 `qiantie_v88.accounts` 表结构

```sql
CREATE TABLE `accounts` (
  `username` varchar(32) NOT NULL,
  `password_hash` text NOT NULL,
  `password_algorithm` varchar(32) NOT NULL,
  `is_owner` tinyint(1) NOT NULL,
  `active` tinyint(1) NOT NULL,
  `created_at` datetime(6) NOT NULL,
  `updated_at` datetime(6) NOT NULL,
  PRIMARY KEY (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
```

关键点：

- **没有手机号列、没有邮箱列**；登录标识只有 `username`。
- 没有失败次数、锁定字段、`last_login_at`；登录成功记录在 `auth_sessions`。
- 默认排序规则 `utf8mb4_0900_ai_ci` → 大小写不敏感、重音不敏感。
- 主键唯一，不会出现多条同名账号。

### 2.2 zhangying 行

| 字段 | 值 |
|---|---|
| 表 / 库 | `qiantie_v88.accounts` |
| username | `zhangying` |
| is_owner | 0 |
| active | 1 |
| password_algorithm | `scrypt-v1` |
| password_hash 前缀 | `scrypt$` |
| password_hash 长度 | 104 |
| created_at | 2026-09-02 07:06:12.053 |
| updated_at | 2026-10-16 12:51:02.239 |

`scrypt-v1` 的合法格式为 `scrypt$<32 hex salt>$<64 hex hash>`，总长 104，与 `VerifyPassword` 的解析逻辑完全匹配。

### 2.3 会话记录

`auth_sessions` 中 `zhangying` 共 4 条，最近一条：

- `issued_at = 2026-10-16 12:51:50`
- `expires_at = 2026-11-15 12:51:50`
- `ip_hint = 172.19.0.4:46224`（Docker 网段，提示该快照可能来自本地/评审环境）

含义：2026-10-16 12:51:02 更新密码后，**48 秒后有一次成功登录**。这证明当时的新密码是可用且哈希兼容的；此后再没有新的成功登录会话（备份截至 11-03）。

### 2.4 两套历史账号表

| 库 / 表 | zhangying | password_hash 特征 |
|---|---|---|
| `qiantie_shuihuo_compat.users` | id=38，`is_owner=0`，`is_active=1`，创建 2026-09-17 | 字面量 `bridge-managed-account`（22 字符占位符） |
| `qiantie_shuihuo_legacy.users` | id=182，`is_owner=0`，`is_active=1`，创建 2026-09-28 | 同上，占位符 |
| `qiantie_v88.users` | 备份中该表**无任何数据行** | — |

补注：`qiantie_shuihuo_compat` 的 owner 行（`choushiyiguai`）使用的是 44 字符 base64 风格哈希，不是 scrypt；说明旧系统校验逻辑与 Go 不同。Go 登录不读取这些旧表。

结论：

- yzcm 主站 Go 登录只查 `accounts`（[mysql_store.go:51-67](file:///f:/脚本测试/一战晟铭-v88/backend/internal/auth/mysql_store.go#L51-L67)）。
- 旧 `shuihuo` 库里的 `zhangying` 是 **bridge-managed 占位账号**，没有真实密码；如果用户走的是旧 Node/旧系统登录，密码登录必然失败，且该错误与 yzcm 主站无关。
- `qiantie_v88.users` 是遗留空表，不是当前登录存储。

### 2.5 主站与 `/admin` 是否同一套

是同一套本地账号体系：

- 后端只有 `POST /api/login`（[auth.go:57-115](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/auth.go#L57-L115)）和一个 `accounts` 表（[mysql_store.go:51-67](file:///f:/脚本测试/一战晟铭-v88/backend/internal/auth/mysql_store.go#L51-L67)）。
- `/admin` 是同一个前端 SPA 的页面；管理 API 复用同一个会话 cookie/中间件，没有第二套管理员登录接口或第二张本地账号表。
- `zhangying` 的 `is_owner=0`；owner 是 `choushiyiguai`。所以即使 `zhangying` 登录成功，也不能进入 owner-only 的管理页面（会得到 403/无权限，而不是“账号错误”）。

---

## 3. 账号查找逻辑

`FindAccount`：

```go
SELECT username, password_hash, password_algorithm, is_owner, active, created_at, updated_at
FROM accounts WHERE username = ?
```

见 [mysql_store.go:59](file:///f:/脚本测试/一战晟铭-v88/backend/internal/auth/mysql_store.go#L51-L67)。

- 精确等值查询，没有 `LIKE`、没有 JOIN、没有按状态过滤（`active` 在 handler 里判断）。
- `strings.TrimSpace` 只去掉首尾空白（[auth.go:73/77](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/auth.go#L73-L84)）。
- `validateUsername` 只限制非空、长度 ≤32（[mysql_store.go:44-49](file:///f:/脚本测试/一战晟铭-v88/backend/internal/auth/mysql_store.go#L44-L49)）；超长输入直接 `ErrInvalid`，但仍被包装成 `用户名或密码错误`。
- MySQL `utf8mb4_0900_ai_ci` 使大小写不敏感；实测 `ZhangYing` 同样触发 scrypt（约 0.35s），说明能找到账号。
- 手机号 `13800000000` 实测约 0.21s，等价于“未找到”。

**因此**：

- 输入 `zhangying`、`ZhangYing`、首尾带空格 → 都能命中。
- 输入手机号、邮箱、其他别名 → 查不到；界面只显示统一文案，看起来像“账号错误”。

---

## 4. 近期变更核查

### 4.1 近两周（2026-10-22 至 2026-11-05）

只改了 `auth.go` 两处，均与账号匹配/密码校验无关：

- `0b7f6eda`（2026-11-05）：owner 账号也读取成员资料的 `displayName/avatarUrl`，只影响头像显示（[auth.go:29-46](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/auth.go#L29-L46)）。
- `bf2d9596`（2026-11-02）：`FindAccount` 非 NotFound 错误改为返回 503「登录服务暂不可用」（[auth.go:78-80](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/auth.go#L78-L80)）。这只会让数据库故障更明确，不会造成“账号错误”。

`password.go`、`mysql_store.go`、`auth_schema.go` 近两周没有提交。

### 4.2 13 版（2026-11-05 14:34:01，`32a3c6ad`）

- 部署记录：[deployment-record.json](file:///f:/脚本测试/一战晟铭-v88/docs/releases/20261105-13-32a3c6ad/deployment-record.json)
- 线上 `runtime_manifest_sha=32a3c6ad...`、`frontend_entry=/assets/user-Cs6GUvIU.js`，与 13 版一致。
- 13 版 `changes.diff` 中没有 `auth.go`、`password.go`、`mysql_store.go`、`auth_schema.go`、`accounts`、`QIANTIE_*` 的改动；变更集中在批量工厂/发布映射/漫画等。
- `0b7f6eda` 与 `bf2d9596` 都已包含在 12 版（`c9039a6a`）中；13 版没有再次改动认证代码。

### 4.3 更早的账号体系迁移（2026-10-10 至 10-15）

- `9439ccf4` 2026-10-10：Go 接管基础账号会话
- `cb7b6d37` 2026-10-10：成员改密走 Go
- `20b94a73` 2026-10-10：成员安全走 Go
- `579fe541` 2026-10-10：账号目录读取走 Go
- `ae3c3baf` 2026-10-15：账号清理契约走 Go
- `d9673f8f` 2026-10-15：接受已校验的 public auth checksum（只改 schema/测试）

`zhangying` 的 `updated_at=2026-10-16 12:51` 正好发生在这批迁移之后，且 48 秒后成功登录，说明当时密码是通过新 Go 路径写入并立即可用的。

**结论：没有证据表明某次改动“收紧了账号匹配”或破坏了密码校验。**

---

## 5. 可比对证据：全库哈希格式

`qiantie_v88.accounts` 共 **20** 个账号：

| 算法 | 条数 | hash 前缀 | 异常 |
|---|---|---|---|
| `scrypt-v1` | 2 | `scrypt$` | 0 |
| `scrypt` | 18 | `scrypt$` | 0 |
| 空 / bcrypt / argon2 / pbkdf2 / 明文 | 0 | — | 0 |

- 所有账号都是 scrypt 格式，没有“只有 zhangying 格式特殊”的情况。
- `VerifyPassword` 同时支持 `scrypt`（N=16384）和 `scrypt-v1`（N=32768），见 [password.go:30-49](file:///f:/脚本测试/一战晟铭-v88/backend/internal/auth/password.go#L30-L49)。
- 时序测试显示 zhangying 触发了 scrypt 计算，说明哈希可解析、算法分支命中；不是“格式不认识”。

因此“算法不匹配导致这个账号登不上”可以排除。

---

## 6. 结论：为什么提示「登录账号错误」

### 6.1 对“文案”的解释

- **最强解释**：用户看到的原文大概率是 **「用户名或密码错误」**，两次口述分别转述成了「密码错误」和「登录账号错误」。
- **次强解释**：用户实际在另一个登录入口（121 后台/旧系统/旧书签）操作，该入口不属于 yzcm 主站。当前 yzcm 线上所有 JS 与接口都没有这条文案。
- 无论哪种，**不能由「登录账号错误」推出“账号查找阶段失败”**；yzcm 的代码没有这个独立分支。

### 6.2 对“为什么登不上”的解释

若用户输入的是 `zhangying`（含大小写变体），线上证据显示：

1. 账号找到（否则不会计算 scrypt）
2. 账号 `active`（否则不会计算 scrypt）
3. 哈希格式可解析（否则不会计算 scrypt）
4. 数据库查询正常（否则返回 503「登录服务暂不可用」，而不是 401）
5. **最终密码校验返回 false**

→ **直接原因是密码与 2026-10-16 12:51:02 保存的哈希不匹配。** 这是强推断，不是直接读线上库；但代码分支与时序侧信道相互印证。

若用户输入的不是 `zhangying`（手机号/邮箱/其他别名），则直接原因是 **账号找不到**，并被统一文案掩盖。手机号实测为“未找到”。

### 6.3 证据强度表

| 假设 | 证据 | 强度 | 备注 |
|---|---|---|---|
| 账号存在且启用 | 线上时序 +150ms；备份 `active=1` | **强** | 时序为侧信道推断 |
| 密码不匹配 | 线上错误密码触发 scrypt 后返回 false；10-16 后无成功登录 | **强** | 未拿到用户真实输入 |
| 输入不是 username | 表无手机/邮箱列；手机号时序=未找到 | **中** | 需日志确认实际输入 |
| 文案来自其他入口 | 源码/git/入口 84 chunk/121 实测均无该文案 | **强**（对“不属于本系统”） | 需 URL 才能定位具体入口 |
| 账号被禁用 | 时序显示 scrypt 执行；备份 active=1 | **强**（排除） | — |
| 哈希格式不匹配 | 全库 20 个账号均 scrypt；时序显示 scrypt 执行 | **强**（排除） | — |
| 近两周代码改坏 | `git log` 仅头像/503 两处 | **强**（排除） | — |
| 11-03 后账号被删 | 无法实时读库 | **弱** | 只能靠线上登录取证 |

---

## 7. 现场证据与阻断

### 7.1 从公网可确认的

- `HEAD /`：`HTTP/1.1 200`，`Server: nginx/1.24.0 (Ubuntu)`。
- `POST /api/login`（合法 JSON，账号 `zhangying`，故意错误密码）：`401`，`{"error":"用户名或密码错误"}`。
- `POST /api/login`（合法 JSON，随机不存在账号）：同样 `401` 和同一文案。
- `POST /api/login`（合法 JSON，手机号 `13800000000`）：同样 `401` 和同一文案，但无 scrypt 延迟。
- `POST /api/login/passkey/options`（`zhangying` / 随机账号）：均为 `404 当前账号没有可用 Passkey`（不能据此判断账号是否存在）。
- `GET /api/runtime-build-info`：`runtime=go`、`runtime_manifest_sha=32a3c6ad...`、`frontend_entry=/assets/user-Cs6GUvIU.js`。

### 7.2 无法取得的

- SSH 到 `root@115.190.156.223`：`Permission denied (publickey,password)`（已尝试本机可用密钥，均被拒）。
- 因此**没有拿到** nginx access/error log、`journalctl -u qiantie.service`、线上实时数据库。
- 无法确认用户请求的真实 `username` 字段、UA、来源 IP、精确时间戳，也无法用后端日志区分分支。
- 这是本次排查唯一的硬阻断；在不重启、不改配置、只读的约束下无法绕过。

### 7.3 其他观察

- `/api/runtime-build-info` 的 `git_sha` 字段返回 `6db3b5ad`（2026-11-03 的祖先提交），而 `runtime_manifest_sha` 为 13 版 `32a3c6ad`。这更像是 `QIANTIE_RELEASE_SHA` 环境变量陈旧，与本次登录问题无关；但按仓库发布规则，发布元数据应对齐，建议后续单独核查。

---

## 8. 最小恢复动作（按优先级）

1. **先向用户确认两件事**：实际页面 URL、输入框里的原始账号字符串（截图）。
2. **如果 URL 是 `yzcm.site` 主站，且账号输入是 `zhangying`**：
   - 让用户使用 **2026-10-16 12:51 之后设置的密码**。
   - 若忘记：
     - 由管理员使用 `POST /api/admin/accounts/zhangying/reset-password`（需管理员会话 + 新密码 ≥8，见 [admin_governance_compat.go:78-98](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/admin_governance_compat.go#L78-L98)）；或
     - 使用 `/recover` 邮箱找回（要求邮箱已验证并绑定账号，见 [account_recovery_compat.go:87-135](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/account_recovery_compat.go#L87-L135)）。
   - 如果用户当前浏览器里仍有 10-16 那次登录的会话（有效期至 11-15），可先登录后在 `POST /api/member/security/password` 自助改密（[member_profile.go:595](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/member_profile.go#L595)）。
3. **如果用户输入的是手机号/邮箱**：改为 `zhangying`（大小写不敏感，首尾空格会被 trim）。
4. **如果用户实际在 121/旧系统**：使用该系统的账号；旧库 `zhangying` 的 `bridge-managed-account` 是占位符，不能密码登录。
5. **不要**为此修改登录代码、改库、重置密码或重启服务；当前证据不支持这些动作。

---

## 9. 推断与限制声明

- “账号存在且启用”“哈希可解析”是**时序侧信道推断**，不是直接读线上数据库；依据是代码短路顺序 + 约 150ms 的 scrypt 差异。若线上服务器负载突变，时序可能有噪声；但 8+3 次配对测试方向一致，且对照账号/手机号/随机账号的差异稳定。
- 本地 11-03 备份为离线快照，不能证明线上 11-03 之后没有改库；`auth_sessions.ip_hint=172.19.0.4` 提示它可能来自本地/评审环境。
- 未能取得 nginx/后端日志，因此无法确证用户实际提交的 username、UA、IP，也无法从日志区分“未找到 / 禁用 / 哈希异常 / 密码错”。本文对“密码不匹配”的判断是**强推断**，不是日志确证。
- 本次为只读排查：没有 UPDATE/INSERT/DELETE，没有重置密码，没有重启服务，没有改代码，没有 commit。线上探测仅使用故意错误的密码，未产生成功会话。

---

## 10. 证据索引

- 登录 handler：[auth.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/auth.go#L57-L115)
- 密码校验：[password.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/auth/password.go#L30-L49)
- 账号查找：[mysql_store.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/auth/mysql_store.go#L51-L67)
- 前端登录与错误展示：[UserLayout.jsx](file:///f:/脚本测试/一战晟铭-v88/frontend/src/shared/layouts/UserLayout.jsx#L315-L415)
- 121 上传登录文案：[novel_fetch_upload_compat.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/novel_fetch_upload_compat.go#L88-L92)
- 管理员重置密码：[admin_governance_compat.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/admin_governance_compat.go#L78-L98)
- 邮箱找回：[account_recovery_compat.go](file:///f:/脚本测试/一战晟铭-v88/backend/internal/httpapi/account_recovery_compat.go#L87-L135)
- 13 版部署记录：[deployment-record.json](file:///f:/脚本测试/一战晟铭-v88/docs/releases/20261105-13-32a3c6ad/deployment-record.json)
- 本地全库备份：`.backups/mysql-full-20261103T114757Z.sql.gz`（只读解析，未导出完整哈希）
