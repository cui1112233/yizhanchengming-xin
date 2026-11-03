# V88 启动与配置说明（新手版）

这份文档只讲两件事：**怎么把 V88 在本地跑起来**，以及**各项功能需要先配什么**。所有参数都来自当前代码和 `.env.v88-review.example`；代码里没有的，不会凭空补。

先记住三个结论：

1. 本地 review 脚本默认端口是 **4001**；直接运行 Go 二进制且不设变量时，默认端口是 **4000**。
2. 本地启动脚本默认只开到 **Slice 3**。能看、能建批次、能生成提示词和漫画页；**不能提交视频生产、不能合成、不能对外发布**。
3. 启动最少的必配项只有两个：review 脚本填 `QIANTIE_V88_MYSQL_PASSWORD` 和 `QIANTIE_V88_BRIDGE_SECRET`；直接跑 Go 则用 `QIANTIE_MYSQL_DSN` 和 `QIANTIE_BRIDGE_SECRET`。Redis 没启动时 API 仍能起，但队列类任务不可用。

---

## 1. 启动前准备

### 1.1 必须有的东西

| 依赖 | 用途 | 没有会怎样 |
| --- | --- | --- |
| Go 1.23 或兼容版本 | 编译并运行主服务 | 无法启动 Go |
| Node.js | 只在本地构建 React | 前端页面不会更新 |
| MySQL | 业务数据、任务、配置 | 服务直接启动失败 |
| Redis | 队列、短时状态、锁 | API 能起，但排队任务不推进 |
| FFmpeg | 音视频处理和合成 | `start-local-go-review.ps1` 会直接报错 |

### 1.2 复制本地配置文件

```powershell
Set-Location F:\脚本测试\一战晟铭-v88
Copy-Item .env.v88-review.example .env.v88-review
```

打开 `.env.v88-review`，至少填写这两项：

```text
QIANTIE_V88_MYSQL_PASSWORD=你的数据库密码
QIANTIE_V88_BRIDGE_SECRET=你的桥接密钥
```

说明：

- `start-local-go-review.ps1` 启动前会检查这两项非空，缺一项就直接退出。
- 使用 `-UseYcmMysql` 时，实际数据库密码会从本机 Docker 容器 `ycm-mysql` 读取；`.env` 里的密码仍要填一个非空值，否则脚本仍会报“缺少数据库密码”。
- 使用 `-UseReviewMysql` 时，连接的是隔离 review MySQL，库名 `qiantie_v88`；脚本会打印 `Using MySQL 用户@地址:端口/库`，常见是 `127.0.0.1:3307`。这时 `.env` 里的密码必须和该容器一致。
- 两种模式都会先执行 `docker inspect ycm-mysql` 读取 root 密码；本机没有这个容器或 Docker 不可用时，脚本会在读取密码阶段直接退出。没有 `ycm-mysql` 时，请改用 `start-local-go.ps1` 并手动设置 `QIANTIE_MYSQL_DSN`。
- 示例文件里这些键，`start-local-go-review.ps1` **不会读取**，它们原本面向隔离 review 环境：
  - `QIANTIE_V88_HOST_PORT=13188`：隔离环境对外端口；
  - `QIANTIE_V88_SLICE=4`：隔离环境把 Slice 开到 4。
  仓库里当前没有对应的 `docker-compose` 文件，这些键的完整用途**待确认**。
- 豆包 Relay、N8 漫画发布所需的键，可以手动追加到 `.env.v88-review`，脚本会读取：
  - `QIANTIE_DOUBAO_RELAY_BASE_URL`
  - `QIANTIE_DOUBAO_RELAY_BEARER_TOKEN`（脚本要求两者都非空才启用 Relay；固定账号 `QIANTIE_DOUBAO_RELAY_FIXED_ACCOUNT` 会被脚本清掉）
  - `QIANTIE_V88_N8_COMIC_OPEN_API_KEY`
  - `QIANTIE_V88_N8_COMIC_PRODUCER_ACCOUNTS`
  - `QIANTIE_V88_N8_ADMIN_TOKEN`
  - `QIANTIE_V88_MERGE_TOS_*`（会映射给小说正文和 Agent 资产使用）

### 1.3 直接运行 Go 时的必配环境变量

不走 review 脚本、直接启动 `backend/qiantie-local.exe` 时，至少设置：

```powershell
$env:QIANTIE_MYSQL_DSN = "user:password@tcp(127.0.0.1:3306)/qiantie?parseTime=true&multiStatements=true"
$env:QIANTIE_BRIDGE_SECRET = "你的桥接密钥"
$env:QIANTIE_REDIS_URL = "redis://127.0.0.1:6379"
```

| 变量 | 是否必填 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `QIANTIE_MYSQL_DSN` | 必填 | 无 | 为空时服务直接启动失败 |
| `QIANTIE_BRIDGE_SECRET` | 必填 | 无 | 为空时服务直接启动失败 |
| `QIANTIE_GO_LISTEN_ADDR` | 选填 | `:4000` | 监听地址；本地脚本会覆盖成 `:4001` |
| `QIANTIE_REDIS_URL` | 推荐 | 空 | 为空时队列不可用，API 仍可启动 |
| `QIANTIE_RELEASE_SHA` | 选填 | 空 | 本地脚本会自动写入当前 Git SHA |
| `QIANTIE_LOCAL_FFMPEG_DIR` | 选填 | `.tools\ffmpeg` | review 脚本用它找 `ffmpeg.exe` |
| `QIANTIE_BATCH_FACTORY_V11_DYNAMIC_TEXT_PROVIDER` | Slice 2+ 需要 | `0` | 设为 `1` 时从个人中心模型目录读取文本模型；否则还要配 `..._TEXT_ENDPOINT`、`..._TEXT_API_KEY`、`..._TEXT_MODEL` |

本地的两个启动脚本会自动把 Slice 设为 `3`，并自动设置 `QIANTIE_BATCH_FACTORY_V11_DYNAMIC_TEXT_PROVIDER=1`。如果你直接运行 Go 二进制并手动把 Slice 调到 2 以上，必须自己补文本模型配置，否则配置校验会让服务启动失败。

`scripts/start-local-go.ps1` 会显式校验 Slice：未设置、`0`、非数字或负数都会被强制提升为 `3` 并打印警告。原因见 [README.md](../README.md) 的启动说明：代码默认 Slice 为 `0`，而 Slice 小于 `1` 时 V12 批次路由不会注册，`GET /api/batch-factory/v12/batches/summary` 会返回 404，页面表现为“作品全部消失”。直接裸跑二进制时同样至少要把 Slice 设为 `1`。

临时端口验证可以用现有二进制、不重建前端：

```powershell
.\scripts\start-local-go.ps1 -SkipBuild -ListenAddr ':4002'
```

启动日志会打印 `本地启动配置：slice=... production=... merge=...`，先确认 slice 至少为 `1` 再排查其它问题。

如果要用巨量素材解析，直接运行 Go 时还要设置下面这一组；`start-local-go-review.ps1` 不会从 `.env` 读取它们：

```text
QIANTIE_GIANT_MATERIAL_PLUGIN_URL=http://127.0.0.1:18180
QIANTIE_GIANT_MATERIAL_PLUGIN_TOKEN=插件令牌
QINGYU_N8_ADMIN_TOKEN=青语后台令牌
QINGYU_MATERIAL_SELECT_URL=https://n8.hnqingyuwen.top/center-api/material/video/select
```

只配 `QINGYU_N8_ADMIN_TOKEN` 也能让主服务直接请求青语；配了 `QIANTIE_GIANT_MATERIAL_PLUGIN_URL` 时，主服务改用插件。

---

## 2. 本地启动

### 2.1 推荐：review 脚本（会自动构建）

第一次先装前端依赖：

```powershell
npm --prefix frontend install
```

然后启动：

```powershell
.\scripts\start-local-go-review.ps1 -UseYcmMysql
```

脚本会做这些事：

1. 读取 `.env.v88-review`；
2. 连接本机 `ycm-mysql` 容器，默认库为 `qiantie_go_local`；
3. 检查 FFmpeg；
4. 前端源码或构建产物变化时，自动重新构建前端和 Go；
5. 以 `:4001` 启动 Go。

如果本机用的是隔离 review MySQL，而不是 `ycm-mysql`：

```powershell
.\scripts\start-local-go-review.ps1 -UseReviewMysql
```

### 2.2 手动构建并启动

```powershell
npm --prefix frontend run build
.\scripts\build-go-local.ps1
.\scripts\start-local-go.ps1 -ListenAddr ':4001'
```

只想重新整理前端静态文件、不重新编译 Go：

```powershell
.\scripts\build-go-local.ps1 -SkipGoBuild
```

### 2.3 启动成功的判断标准

另开一个终端执行：

```powershell
.\scripts\check-local-go.ps1
```

看到这些才算成功：

- `/health` 返回 `PASS (200)`；
- `/`、`/novel-fetch`、`/shuihuo-production`、`/settings` 返回 `200`；
- `/api/runtime-build-info` 返回的 `git_sha` 与当前 `git rev-parse --short HEAD` 一致；
- 启动终端打印 `qiantie Go V12 listening on :4001`。

浏览器打开：

```text
http://127.0.0.1:4001
```

停止服务：在启动终端按 `Ctrl+C`。

### 2.4 端口对照

| 场景 | 地址 |
| --- | --- |
| 本地 review 脚本 | `http://127.0.0.1:4001` |
| Go 裸启动默认 | `http://127.0.0.1:4000` |
| Vite 前端开发 | `http://127.0.0.1:5173` |
| 巨量素材执行器本机页面 | `http://127.0.0.1:17861/setup` |

不要同时让 Node 和 Go 占用同一个端口。

---

## 3. 默认 Slice 3：能做什么、不能做什么

Slice 是批量工厂的能力开关，不是页面开关。

| Slice | 打开的能力 |
| --- | --- |
| 0 | 几乎只有基础 API |
| 1 | 批次读取、创建、设置、快照、覆盖编辑；小说获取、资产和漫画页接口 |
| 2 | 导演分镜、钩子检查、爆款工作前台 |
| 3（本地默认） | 在 1、2 基础上增加最终提示词编译预览、H3 追踪 |
| 4 | 视频生产提交与取消 |
| 5 | 视频合成 |
| 6 | 对外发布（121、Yadi 等） |

注意：

- `start-local-go.ps1` 和 `start-local-go-review.ps1` 在未显式设置时，会把 `QIANTIE_BATCH_FACTORY_V11_SLICE` 设为 `3`。
- 直接运行 Go 二进制且不设这个变量时，配置层默认是 `0`，不是 3。
- 本地默认 Slice 3 **可以**：小说获取、批量工厂建批次、设置覆盖、导演分镜、提示词编译预览、图片生成、漫画页生成。
- 本地默认 Slice 3 **不可以**：提交视频生产、视频合成、121/Yadi 对外发布。这些视频和外部发布写接口不会注册，通常表现为裸 `404`。
- 漫画的 N8 发布是独立链路，不走视频 Slice 4/5，也不受这两个开关影响；只要 N8 配置齐全就能用。

### 3.1 开启视频生产

```powershell
.\scripts\start-local-go-review.ps1 -UseYcmMysql -EnableBatchProduction
```

等价的启动变量：

```text
QIANTIE_BATCH_FACTORY_V11_SLICE=4
QIANTIE_BATCH_FACTORY_V11_PRODUCTION_ENABLED=1
```

只打开 Slice 4 还不够，具体账号还要在「批量工厂 → 设置 → 视频生成通道」里保存可用的视频通道和密钥。

### 3.2 开启视频生产和本地合成

```powershell
.\scripts\start-local-go-review.ps1 -UseYcmMysql -EnableBatchProduction -EnableLocalMerge
```

等价的启动变量：

```text
QIANTIE_BATCH_FACTORY_V11_SLICE=5
QIANTIE_BATCH_FACTORY_V11_PRODUCTION_ENABLED=1
QIANTIE_BATCH_FACTORY_V11_MERGE_ENABLED=1
QIANTIE_BATCH_FACTORY_V11_LOCAL_MERGE_ENABLED=1
```

规则：

- `-EnableLocalMerge` 必须同时带 `-EnableBatchProduction`，否则脚本直接报错。
- 配置层要求 Slice 4 必须同时有 `PRODUCTION_ENABLED=1`；Slice 5 必须有明确的合成配置。
- 合成产物需要对象存储时，再配置 `QIANTIE_BATCH_FACTORY_V11_MERGE_TOS_*`；只做本地合成时可以只保留本地产物。

---

## 4. 两套执行器：豆包 local 与巨量 giant

### 4.1 它们各自负责什么

| | 豆包本地执行器（local） | 巨量素材执行器（giant） |
| --- | --- | --- |
| 负责 | 图片生成、视频生成 | OCR 读取巨量素材的滚屏正文 |
| 运行位置 | 你自己的 Windows / macOS 电脑 | 你自己的 Windows / macOS 电脑 |
| 是否需要登录豆包账号 | 是，账号只在执行器本机登录 | 否 |
| 是否需要 Python | 否 | 是，OCR 依赖和模型在首次任务时下载 |
| 配对方式 | 网页生成一次性配对码，10 分钟内有效 | 同样是一次性配对码 |
| 本机端口 | 由客户端管理 | `127.0.0.1:17861` |

简单记：

- 要生成图片/视频 → 装豆包执行器；
- 要读巨量素材正文 → 装巨量素材执行器；
- 两者都可以不装，但对应功能就用不了。

### 4.2 安装并配对豆包本地执行器

1. 打开网页「设置 → 执行器状态」。
2. 在「一战晟铭执行器」区域下载客户端：
   - Windows：`一战晟铭执行器-安装版.exe`
   - macOS：`doubao-relay_1.0.0_universal.dmg`
3. 点击「生成配对码」。
4. 在客户端里输入配对码。配对码 10 分钟内有效，只能用一次。
5. 等待状态变成「在线」。
6. 点「测试图片生成」或「测试视频生成」。出现预览图/视频，才算真的通了。

注意：安装包来自 `data/downloads` 或后端内置下载目录。文件缺失时下载接口返回 `503`，不是客户端坏了。

### 4.3 安装并配对巨量素材执行器

1. 打开网页「设置 → 执行器状态」。
2. 下载 Windows ZIP 或 macOS 便携包。
3. 解压到一个固定目录，例如：

   ```text
   C:\Users\<你的用户名>\AppData\Local\YizhanShengming\GiantMaterialExecutor
   ```

4. 回到网页「巨量素材执行器」区域，点击「生成配对码」。
5. 双击 `GiantMaterialExecutor.exe`。首次启动会自动打开：

   ```text
   http://127.0.0.1:17861/setup
   ```

6. 在页面里填写：
   - 控制服务地址：本机就是 `http://127.0.0.1:4001`；跨电脑时填 Windows 能访问的地址；
   - 网页上生成的配对码。
7. 点击「绑定执行器」。绑定成功后凭证保存在当前 Windows 用户目录，重启会自动连接。
8. 回到网页点「刷新状态」，显示「在线」才算配好。
9. 首次 OCR 任务会下载运行包和模型，需要本机有可用的 `python` 命令。

安全要求：不要执行器端口映射到公网；它只监听本机。

### 4.4 状态卡怎么看

统一状态接口：

```text
GET /api/shuihuo-production/executors/status
```

它一次返回 `local` 和 `giant` 两组，每组有 `state` 和 `devices`。

| 状态 | 含义 | 下一步 |
| --- | --- | --- |
| `not_installed` / 未安装 | 没有任何已配对设备 | 下载并安装客户端 |
| `needs_pairing` / 需配对 | 刚生成了配对码，但设备还没绑定 | 在客户端输入配对码 |
| `offline` / 离线 | 有已配对设备，但都不在线 | 打开客户端，检查网络和登录 |
| `online` / 在线 | 至少一台设备在线 | 可以接任务 |

页面会显示成「在线：N 台在线 / 共 M 台已配对」。看到「在线」再点测试，判断标准最可靠。

状态接口还新增了 `pairing` 字段，用来表示“待配对”记录：

- 只要存在待配对记录就会返回；已经过期的记录也会保留，并标记 `expired=true`；
- 只有“记录未过期、且没有任何已配对设备”时，`state` 才会从 `not_installed` 覆盖成 `needs_pairing`；
- `maskedCode` 固定是 `****-****`，状态接口不回传明文配对码；
- `fingerprint` 是配对码的哈希指纹，用于核对网页生成的是哪一次配对；
- `expiresAt` 是失效时间，默认 10 分钟；过期后需要重新生成。

所以刷新状态卡时只能看到打码和指纹；明文配对码只在点击「生成配对码」的那一次显示，要立即输入客户端。

---

## 5. 图片生成的三条通道

图片执行路径按后端实际逻辑分三种。

| 通道 | 什么时候用 | 在哪里配 | 配好的判断 |
| --- | --- | --- | --- |
| API 模型目录 | 有可用的图片 API | 「管理员 → 水货生产模型」（`/admin/shuihuo-models`） | 添加并启用图片模型，填服务地址和服务端密钥；生成一张资产图成功 |
| 豆包本地执行器 | 用自己电脑上的豆包账号出图 | 「批量工厂 → 设置 → 图片生成通道」选「豆包本地执行器」 | 豆包执行器显示在线；「设置 → 执行器状态 → 测试图片生成」出图 |
| 豆包 Relay | 服务端通过 Relay 调豆包出图 | 服务器环境变量，不在网页填 | 设置 `QIANTIE_DOUBAO_RELAY_BASE_URL` 和 `QIANTIE_DOUBAO_RELAY_BEARER_TOKEN` 后重启；启动日志出现 `Doubao relay image path enabled` |

下拉里默认还会看到「AutoDL ComfyUI」。图片侧它不是独立通道，仍走 API 模型目录的服务地址和密钥；没有对应模型时不要选。

### 5.1 API 模型目录

1. 用管理员账号打开 `/admin/shuihuo-models`。
2. 新增图片模型，确认：
   - 类型是图片；
   - 状态是启用；
   - 填了服务地址（`endpoint` 或 `baseUrl`）；
   - 填了服务端密钥（`credential`）。
3. 如果当前生产账号不是管理员，还要用该账号打开「个人中心 → API 配置」（`/api-config`），添加并启用自己的图片模型。管理员页面只同步管理员账号自己的模型目录。
4. 回到「批量工厂 → 设置 → 图片生成通道」，选择「个人中心 API」。
5. 选择具体的图片模型。
6. 生成一张资产图，成功即配置完成。

失败时会返回错误码 `IMAGE_MODEL_NOT_CONFIGURED`。

### 5.2 豆包本地执行器

1. 先按第 4.2 节把豆包执行器配对到在线。
2. 在「批量工厂 → 设置 → 图片生成通道」选择「豆包本地执行器」。
3. 选择本地图片模型，选项是 Seedream 4.0 / 4.5 / 5.0 Flash / 5.0 Pro。
4. 在「设置 → 执行器状态」点「测试图片生成」。

这条通道不需要 API 密钥，账号在本地执行器里登录。

### 5.3 豆包 Relay

Relay 不是网页里的第三个下拉项，而是本地执行器通道的服务端替代路径。

1. 在 `.env.v88-review` 或服务器环境里加：

   ```text
   QIANTIE_DOUBAO_RELAY_BASE_URL=http://你的Relay地址
   QIANTIE_DOUBAO_RELAY_BEARER_TOKEN=你的Relay令牌
   ```

2. 可选固定账号：

   ```text
   QIANTIE_DOUBAO_RELAY_FIXED_ACCOUNT=账号名
   ```

   不填时由 Relay 自己按已登录账号轮询。注意：`start-local-go-review.ps1` 会主动清掉这个变量，只在手工启动或服务器环境里保留时才生效。

3. 网页里仍然选择「豆包本地执行器」通道。
4. 重启服务。看到 `Doubao relay image path enabled at ...` 说明 Relay 已接管。
5. 再生成一张图验证。

优先级：Relay 是否接管由 `QIANTIE_DOUBAO_RELAY_BASE_URL` 决定；实际使用时应同时设置 `QIANTIE_DOUBAO_RELAY_BEARER_TOKEN`。`start-local-go-review.ps1` 只在两者都非空时启用 Relay，否则走已配对的本地客户端。

---

## 6. 视频生成的四个 provider

| provider | 页面显示 | 前置条件 | 默认模型 | 备注 |
| --- | --- | --- | --- | --- |
| `personal_api` | 个人中心 API / API 配置模型 | 保存 API 密钥，或管理员模型目录里有对应凭据 | `yd2.0-mini` | 默认通道 |
| `doubao_local_executor` | 本地执行器模型 | 豆包执行器已配对且在线 | `doubao-seedance`（页面本地选项为 Seedance 2.0/2.5 系列） | 不需要 API 密钥 |
| `autodl_comfyui` | AutoDL ComfyUI / API 配置模型（AutoDL H3） | API 密钥；workflow ID 可选，填了必须合法 | `minimax-h3-video` | 默认地址使用 `{workflow}` 占位；不填 workflow 时按有无参考图自动选默认模板 |
| `yfai_seedance` | 梵客视频（部分位置写作梵客 Seedance） | API 密钥 | `seedance-2-0-official` | 默认创建地址 `https://yf.token6688.com` |

### 6.1 在哪里选

按覆盖层级从大到小：

1. 「批量工厂 → 设置 → 视频引擎（视频生成通道）」：整个批次默认；
2. 「书设置 → 视频生成」：单本书覆盖；
3. 单个视频/分镜覆盖；
4. 「管理员 → 水货生产模型」：只负责提供模型目录里的凭据和模型定义，不是通道开关。

保存后，通道配置会写入 MySQL 的账号配置里，重启不丢。状态查询走统一的服务端读取，不再由前端自行拼装。

### 6.2 各通道怎么配好

- `personal_api`
  1. 在视频通道选择 API 配置模型；
  2. 填 API 密钥，或先在管理员模型目录里配好 `yd2-mini-video` 的凭据；
  3. 保存后提交一条视频任务验证。

- `doubao_local_executor`
  1. 豆包执行器在线；
  2. 选择本地执行器模型；
  3. 先点「设置 → 执行器状态 → 测试视频生成」。
  4. 源码里仍有“本地视频下载回传在下一阶段接通”的旧提示；当前版本是否已完整回传，**待确认**，以测试视频能否出现可播放结果为准。

- `autodl_comfyui`
  1. 填 API 密钥；
  2. workflow ID 可以不填：无参考图时默认 `minimax_h3_lightx2v_no_pic`，有参考图时默认 `minimax_h3_lightx2v_v5_15s`；
  3. 如果要指定 workflow，填的值必须合法，否则提交时报 `AutoDL H3 workflow id is invalid`；
  4. 默认创建地址是 `https://autodl.art/api/v1/comfyui/comfyui_workflow/{workflow}`，一般不用改。

- `yfai_seedance`
  1. 填 API 密钥；
  2. 默认模型 `seedance-2-0-official`，一般不需要改地址；
  3. 保存后提交验证。

缺少通道或密钥时返回错误码 `VIDEO_PROVIDER_NOT_CONFIGURED`。

---

## 7. 书城的三种内容来源

| 来源 | 前置条件 | 适合什么场景 |
| --- | --- | --- |
| 121 书城（上游接口） | 能访问 121 上游；有平台 ID / Book ID | 标准的小说获取 |
| 巨量素材 | 插件解析地址与令牌；正文用 OCR 执行器或“优先直接获取原文” | 只有巨量素材 ID，没有书城直连 |
| 手工 / 小说获取 intake | 手工粘贴书目，或在 `/novel-fetch` 提交 | 小批量、测试、补录 |

### 7.1 121 书城

- 默认上游地址：`https://txt.121w.com/api.php`。
- 可用 `QIANTIE_NOVEL_FETCH_121_ENDPOINT` 覆盖；超时可调 `QIANTIE_NOVEL_FETCH_121_TIMEOUT_SECONDS`，默认 20 秒。
- 当前代码只配置了地址和超时，没有看到额外的 API Key 参数；上游是否需要账号或白名单，**待确认**。
- 判断配好：在 `/novel-fetch` 提交一本书，能看到原文、处理版本和任务日志。

### 7.2 巨量素材

需要两层条件：

1. **解析素材**（拿到素材信息和对应书城书目）
   - 配置 `QIANTIE_GIANT_MATERIAL_PLUGIN_URL` + `QIANTIE_GIANT_MATERIAL_PLUGIN_TOKEN`；
   - 或者走青语 N8：`QINGYU_N8_ADMIN_TOKEN` + `QINGYU_MATERIAL_SELECT_URL`（默认 `https://n8.hnqingyuwen.top/center-api/material/video/select`）。

   仓库内也有独立的插件进程 `backend/cmd/giant-material-plugin`，默认监听 `127.0.0.1:18180`。启动插件前必须设置 `QIANTIE_GIANT_MATERIAL_PLUGIN_TOKEN` 和 `QINGYU_N8_ADMIN_TOKEN`；主服务的 `QIANTIE_GIANT_MATERIAL_PLUGIN_URL` 要指向这个地址。
2. **拿正文**（二选一）
   - OCR：巨量素材执行器已配对在线；
   - 优先直接获取原文：系统用素材里的平台和 Book ID 去 121 取正文。直取失败会保留失败原因，可重试，也可以改回 OCR。

操作入口：批量工厂「新建批量 → 巨量获取」，输入 10—25 位数字素材 ID，每行或用逗号分隔一条。

判断配好：输入 ID 后能解析出书名/Book ID；正文出现或进入 OCR 队列并最终回填。

### 7.3 手工 / 小说获取 intake

- 在 `/novel-fetch` 手工提交，或在批量工厂手工粘贴小说列表。
- 手工列表每行一本，可填 ID、书名、男女频、风格、标签、推荐理由、评级。
- 文本模型不是启动必配项，但 AI 男女频判断、AI 改写、分类建议都需要它；没有文本模型时这些步骤会退回人工或失败。
- 判断配好：手工书能进入批次，且需要 AI 的步骤能正常返回。

小说获取的两个固定语义不要配错：

- 原文完整保留，只处理 `maxTxt` 范围内的内容；
- 页面显示处理数/原文数（例如 `4000/27831`），分类完成显示「分类信息已就绪」，执行中显示「正在执行中…」。

---

## 8. 漫画生成：批量工厂里的独立模式

漫画不是“批量出图”的别名，它是批量工厂里独立的 `productionMode=comic` 模式，有独立资产、提示词、页面和发布链路。

### 8.1 怎么开

1. 新建批次时，在「新建批量 · 统一配置」里找到「生产模式」；已有批次则打开「批量工厂 → 设置」。
2. 把「生产模式」切到「漫画生成」（另一个选项是「视频生成」）。
3. 保存后退出再进入，模式会保留。
4. 为每本书配置图片模型；漫画模式下不要求视频模型。

### 8.2 需要什么

- 文本模型：用于漫画资产提取、漫画提示词生成；
- 图片模型：用于资产参考图和漫画页；
- 图片生成通道可用（API 模型目录、豆包本地执行器或 Relay 三者之一）；
- 漫画提示词预设：包括漫画资产、风格和约束；
- 每页必须同时有「分格导演描述」和「本页基础出图描述」才能生成；
- 图片并发来自统一配置里的 `imageConcurrency`。

### 8.3 怎么跑、怎么判断成功

典型顺序：

1. 提取漫画资产；
2. 生成/确认漫画提示词；
3. 生成漫画页；
4. 需要发布时再走 N8 发布。

判断标准：

- 资产图生成成功；
- 每页有主图；
- 下载「最终提示词 + 资产图片 ZIP」能看到完整结果。

限制与注意：

- 单次漫画页生成最多处理前 **35 页**；
- 漫画模式不使用视频 provider，也不依赖合成开关。Slice 3 下漫画页接口可用；
- N8 发布是可选链路，需要：
  - 本地 `.env.v88-review` 写法：`QIANTIE_V88_N8_COMIC_OPEN_API_KEY`、`QIANTIE_V88_N8_COMIC_PRODUCER_ACCOUNTS`、`QIANTIE_V88_N8_ADMIN_TOKEN`
  - 服务器或直接运行 Go 的写法：`QIANTIE_N8_COMIC_OPEN_API_KEY`、`QIANTIE_N8_COMIC_PRODUCER_ACCOUNTS`、`QINGYU_N8_ADMIN_TOKEN`
  - 可选覆盖 `QIANTIE_N8_COMIC_BASE_URL`（默认 `https://n8-view.hnqingyuwen.top/center-api`）
- 没配 N8 密钥时，页面会显示“N8 漫画发布未配置”，但不影响本地生成漫画。

---

## 9. 四类错误码

接口错误响应里会新增 `code` 字段；旧的 `error` 文案仍然保留。

| 错误码 | 含义 | 处理办法 |
| --- | --- | --- |
| `IMAGE_MODEL_NOT_CONFIGURED` | 没有可用图片模型，或模型缺服务地址/密钥 | 到 `/admin/shuihuo-models` 添加并启用图片模型，填写服务地址和服务端密钥；非管理员账号还要到 `/api-config` 启用模型；如果用本地执行器，改成豆包本地通道并确认执行器在线 |
| `VIDEO_PROVIDER_NOT_CONFIGURED` | 视频通道没选或没密钥 | 到「批量工厂 → 设置 → 视频引擎（错误文案里叫视频生成通道）」选择通道并填写密钥后保存；本地执行器通道则先让执行器在线 |
| `EXECUTOR_NOT_READY` | 执行器未安装、未配对或不在线 | 到「设置 → 执行器状态」生成配对码并安装客户端；巨量素材也可以改用「优先直接获取原文」 |
| `GIANT_PLUGIN_NOT_CONFIGURED` | 巨量素材解析服务没配 | 在服务器设置 `QIANTIE_GIANT_MATERIAL_PLUGIN_URL` / `QIANTIE_GIANT_MATERIAL_PLUGIN_TOKEN`，或配置青语 N8 令牌后重启 |

补充：

- 巨量素材执行器离线时，前端还可能直接报 `GIANT_EXECUTOR_OFFLINE`，处理思路和 `EXECUTOR_NOT_READY` 相同。
- 可能会看到 `TEXT_MODEL_REQUIRED`：正文已登记，但没有可用文本模型，AI 判断无法完成。

---

## 10. 常见坑

### 10.1 4001 上还有旧进程

症状：启动脚本报端口被占用，或页面显示的是上一次构建的旧界面。

排查：

```powershell
Get-NetTCPConnection -LocalPort 4001 -State Listen
```

处理：

- 先停掉旧终端里的 Go 进程；
- 确认没有旧的 Node 应用进程占用同一端口；
- 启动脚本本身会调用 `verify-no-node-runtime.ps1` 阻止 Node 与 Go 混跑；
- 若仍怀疑旧进程，执行：

  ```powershell
  .\scripts\verify-no-node-runtime.ps1
  ```

### 10.2 Redis 没启动

症状：页面能打开，但任务一直排队、日志不推进；小说获取任务卡住。

原因：review 脚本把 Redis 固定为：

```text
QIANTIE_REDIS_URL=redis://127.0.0.1:6379
```

处理：

1. 启动本机 Redis；
2. 确认 `6379` 可连接；
3. 重启 Go 服务。

直接运行 Go 时，如果 `QIANTIE_REDIS_URL` 留空，API 仍能启动，但队列功能不可用。

### 10.3 模型目录为空

症状：

- 图片生成直接报 `IMAGE_MODEL_NOT_CONFIGURED`；
- 「图片生成通道」或模型下拉里没有可选项。

处理：

1. 打开 `/admin/shuihuo-models`；
2. 添加模型；
3. 确认「启用」；
4. 图片模型必须有服务地址和密钥；
5. 刷新批量工厂设置页重新选择。

### 10.4 巨量插件缺 token

症状：输入巨量素材 ID 后提示 `GIANT_PLUGIN_NOT_CONFIGURED`，素材解析不出来。

处理：

1. 配置 `QIANTIE_GIANT_MATERIAL_PLUGIN_URL` 和 `QIANTIE_GIANT_MATERIAL_PLUGIN_TOKEN`；或
2. 配置 `QINGYU_N8_ADMIN_TOKEN`；
3. 重启 Go 服务；
4. 重新解析素材。

如果只是正文读取失败，但素材信息已解析出来，可以直接切「优先直接获取原文」，不必等 OCR 执行器。

### 10.5 其他容易踩的点

- **FFmpeg 找不到**：review 脚本会直接报 `FFmpeg not found`。把 `ffmpeg.exe` 放到 `.tools\ffmpeg`，或设置 `QIANTIE_LOCAL_FFMPEG_DIR`。
- **安装包下载 503**：说明 `data/downloads` 里没有对应安装包，不是网络问题。具体发布文件**待确认**。
- **端口用了 13188**：那是示例文件的隔离环境端口，不是本地 review 脚本端口；本地默认用 4001。
- **改了页面但没生效**：确认先构建了前端；`start-local-go.ps1` 会自动检测源码变化并重建。
- **Slice 开太高但配置不全**：Slice 4 必须有 `PRODUCTION_ENABLED=1`；Slice 5 必须有合成配置，否则服务启动直接失败。

---

## 11. 本文标注“待确认”的地方

| 项目 | 为什么待确认 |
| --- | --- |
| `.env.v88-review.example` 中大部分 `QIANTIE_V88_*` 键的用途 | 本地启动脚本只读取其中一小部分；仓库里没有对应的 `docker-compose` 文件 |
| 121 上游是否需要账号、白名单或额外授权 | 代码只暴露地址和超时参数，没有 API Key 配置 |
| AutoDL 默认 workflow 在你的账号是否可用 | 代码有默认模板 `minimax_h3_lightx2v_no_pic` / `minimax_h3_lightx2v_v5_15s`，但线上账号权限和模板可用性需实测 |
| 豆包本地执行器的视频下载回传是否已在当前版本完整接通 | 源码仍有阶段性提示，需用「测试视频生成」实测 |
| `data/downloads` 里实际有哪些安装包 | 仓库不保证安装包存在；缺失时下载接口返回 503 |
| 巨量素材插件服务的具体版本和返回字段 | 插件代码在仓库内，但和主服务分开发布；线上版本与返回字段需以实际部署为准 |
