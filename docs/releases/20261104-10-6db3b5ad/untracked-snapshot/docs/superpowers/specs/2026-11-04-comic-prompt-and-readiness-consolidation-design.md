# 漫画提示词与状态一致性收口设计方案

- 日期：2026-11-04
- 状态：待评审（只出方案，不改代码、不 commit、不重启 4001、不写库）
- 范围：批量工厂 `comic` 生产模式 + 私人定制漫画（两者共用后端漫画链路）

## 0. 结论先行

1. **A 提示词事实源：以 `comic_pages` 行为唯一事实源，video patch 降级为后端派生投影。**
   两个前端入口只调用现有的 `PUT /api/batch-factory/v12/batches/{batchId}/books/{bookId}/comic-pages/{videoId}`（当前已是 patch 语义）；后端只保留一个写入函数，先写页面行、再尽力同步投影。
2. **B 状态判据：后端算权威字段，前端不重算。**
   新增页级 `promptStatus / imageStatus / generationState` 与书级 `comicReadiness`；"有图"统一为"有主图"，无主图时以最新图兜底并打标。
3. **C 重复实现：抽三类行为，保留两类外观。**
   必抽：在途等待与 409 解释、打包下载与文件名、就绪判定。保留：页面壳/卡片布局、发布面板（后者只有批量工厂有）。
4. **D 35 页：后端唯一常量 + capabilities 下发，默认不放开。**
   现状是同一个 35 存在三份（后端常量、前端常量、前端文案）；先删前端两份。按书可配留作后续可选，本轮不做。
5. **E 历史数据：分"可自动 / 需确认 / 只出报告"三档，全部先 dry-run、全部可回滚。**
   提示词去重可自动（带原文备份）；重复合集只报告不删；20 条"failed 但有图"、2 条状态与图不一致需人工确认。
6. **F 门控：能力声明由真实路由注册派生；未注册路由返回 JSON 错误体（保留 404 状态码）；Slice=3 行为不变，用金标测试锁定 0–6 全表。**

## 1. 边界、约束与现状核对

### 1.1 本轮边界

- 只产出本文档，不产生任何代码、提交、数据库变更或服务重启。
- 其他会话正在修改，本方案只描述、不触碰：`backend/internal/httpapi/batch_factory_automation.go`、`batch_factory_v12_aliases.go`、`router.go`、`error_codes.go`、`frontend/src/user/pages/ScriptPage.jsx`、`PrivateComicWorkspace.jsx`、`shuihuo/BatchFactoryNovelList.jsx`。
- 本文所有"现状"均基于当前工作区代码（含未提交改动），行号可能随并行会话移动，以符号名为准。

### 1.2 审计事实的引用方式

五项审计结论（1430/1428 页重复编译、三个写入者写三个地方、三套就绪判据、两套状态口径、重复实现清单）以及存量数据量（1429 行 page/patch 不一致、237 行 patch-only、10 本重复合集、6 条僵尸 `generating`、20 页 failed 有图、2 页状态与图不一致）直接作为设计输入，不重复验证。

### 1.3 读代码时发现的偏差（需修正审计口径）

| 审计表述 | 代码现状 | 对设计的影响 |
| --- | --- | --- |
| "35 页硬上限只存在于批量工厂前端" | 后端 `comicPageGenerationLimit = 35`（`batch_factory_comic_pages.go:40`）在单页、整本、自动化三处生效；私人定制走同一接口，同样受限 | D 议题从"下沉常量"改为"消除三份重复 + 明确语义" |
| "capabilities 写死 `batch.create: available` 但端点 405" | 工作区已补 `POST /api/batch-factory/v12/batches` 别名（`batch_factory_v12_aliases.go:21`） | F 的真实问题是**静态表与注册无关联**，不是 405 本身 |
| "comic-pages PUT 整行覆盖会清空 snapshotHash" | 当前 PUT 已改为 patch 语义（`PatchComicPage`，`batch_factory_comic_pages.go:1124-1167`），`SnapshotHash` 服务端独占 | A 按 patch 契约设计，不再处理整行覆盖 |
| "提示词重复编译" | `BuildComicImageContract` 已具备"剥离-重编"幂等能力（`comic_image_contract.go:178-237`），但历史数据未清理 | 1.2 防新增与 E 清存量是两件事，必须分开验收 |
| "`snapshot_hash` 变更会带来风险" | `comicPageConfigStale` 只比较原文 digest（`batch_factory_comic_pages.go:1807` 起），快照哈希不参与任何门槛 | E 去重不阻塞合集创建；哈希语义需重新定义为"最近一次生成契约" |

## 2. 议题 A：提示词的唯一事实源

### A.1 推荐方案

**方案①：`comic_pages` 行为唯一事实源，video patch 为后端派生只读投影。**

- 写入：所有提示词写入收敛到一个后端函数 `WriteComicPagePrompt`：剥离编译产物后写页面行，再以 `ExpectedRevision` 尽力同步 `comicPrompt/comicImagePrompt` 投影；投影冲突只记日志，不让页面保存失败。
- 读取：`GET comic-pages` 为权威；`GetBatch` 返回的 video 字段由页面行投影，**页面行缺失时回退到 video patch**，保证切换当天不出现"昨天能点、今天不能点"。
- 前端：两个入口只调 `updateComicPage`；批量工厂现有的 `saveVideoOverride` 提示词镜像调用删除。
- 自动化：`persistComicPromptResults`（`comic_prompt_service.go:490-518`）从"只写 video patch"改为调用同一个写入函数。

### A.2 备选与取舍

| 方案 | 改动面 | 迁移成本 | 回归风险 | 对在跑批次 | 可回滚 |
| --- | --- | --- | --- | --- | --- |
| ① page 行为准 + 投影（推荐） | 后端 1 个写入函数 + 读侧兜底；前端删 1 处镜像写 | 需回填 237 行；需按书暂停自动化 | 中：读侧切换顺序错会短暂不一致 | 低：生成仍读页面行，行为不变 | 高：投影继续写入，开关可回退到 patch 权威 |
| ② video patch 为准，page 行废弃 | 需改造生成、合集校验、Agent 投影全部读 patch；page 行的图片/租约/哈希无处安放 | 高（等于重做漫画存储） | 高 | 高 | 低 |
| ③ 双写 + 同事务一致 | 需要跨 `batchfactoryv11.Store` 与 `ComicPageStore` 的事务方法，改 `SaveSettings` 契约 | 中 | 中：事务边界扩大，锁竞争上升 | 中 | 中：去掉事务即回退 |

取舍说明：② 会让"页级图片、租约、快照哈希"这些天然属于页面行的东西寄居在 video 设置里，改动面最大且没有收益；③ 看似最稳，但两个 Store 接口分别对应不同 SQL，引入跨接口事务会把并发锁范围扩大到设置表，反而增加在跑批次的 409 概率。① 把一致性降级为"投影最终一致"，代价是短暂的 patch 漂移，但对用户可见行为无影响，且回退开关成本最低。

### A.3 迁移路径（237 条 patch-only）

顺序不可颠倒：

1. **只读盘点**：按 owner 列出"video 有双字段、无 page 行"的清单，输出 bookId/videoId/提示词长度。
2. **回填**：按 videoId 建页面行，`status = prompt_ready`（沿用内存实现已用的状态值），`referenceImageUrls` 取空，`snapshotHash` 留空。
3. **读侧切换**：先让 `GetBatch` 用页面行投影（含 patch 回退）；观察一个刷新周期。
4. **停写 patch 前端镜像**：只停前端，后端投影继续写。
5. 回填幂等：以 (batchId, bookId, videoId) 为唯一键，重复执行不新增行。

### A.4 一致性边界

- 页面行 `revision` 是唯一并发令牌：人工编辑与在途生成冲突时返回 409，不允许静默覆盖。
- 人工在生成中改了文本：不取消在途出图；该页打 `promptEditedAfterGeneration` 标，前端显示"提示词已更新，图片仍基于旧提示词"。
- 投影漂移对账：书级读取时若发现 page 与 patch 文本不一致，以 page 为准并异步补投影，同时对账计数进日志（不自动改 patch，避免掩盖 bug）。

### A.5 回滚

- 配置开关 `comic.promptAuthority = page | video`，默认 `page`。
- 回退到 `video` 后：读侧改回 patch、生成改为先取 patch 再取页面行；页面行保留不删。
- 因投影在 ① 中始终写入，回退不需要数据回滚。

## 3. 议题 B：状态判据统一

### B.1 推荐方案

**后端算一个权威字段下发，前端只渲染。** 规则表放在 Go 侧单一函数里，页级与书级共用。

页级派生：

| 字段 | 取值 | 规则 |
| --- | --- | --- |
| `promptStatus` | `missing / comicOnly / ready` | 两字段齐全才是 `ready`（与后端预检一致，兼容 light 版仅需 comicPrompt 的例外仍由同一函数判断） |
| `imageStatus` | `none / candidates / primary` | `primary` = 存在 `is_primary`；无任何主图但有图时取最新一张并置 `primaryFallback=true` |
| `generationState` | `idle / generating / failed / succeeded` | `failed` 且已有图时前端显示"已出图，状态待修正"，不显示失败 |

书级 `comicReadiness`：`pagesTotal / promptReadyCount / imageReadyCount / primaryReadyCount / staleCount / canGenerateImages / canBuildCollection / skippable / blockers[] / limits.maxPages`。

使用点：`GET comic-pages` 响应（页级）、`GET /v12/batches/summary` 与批次详情（书级）、Agent 只读投影（沿用同一函数）。前端本地只允许在请求进行中临时禁用按钮，不得改写计数口径。

### B.2 备选与取舍

| 方案 | 改动面 | 迁移成本 | 回归风险 | 对在跑批次 | 可回滚 |
| --- | --- | --- | --- | --- | --- |
| 后端权威字段（推荐） | 后端 1 个纯函数 + 3 个响应点；前端删三套判据 | 无数据迁移 | 低：字段新增，旧字段保留 | 无 | 高：前端可先双读 |
| 前端抽纯函数 | 前端 2 个入口共用 1 个 js 模块 | 无 | 中：仍有两份数据来源（page 与 patch），函数无法解决事实源问题 | 无 | 高 |
| 保持现状只做对齐 | 只改文案 | 无 | 高：三套判据继续漂移 | 无 | 高 |

取舍说明：前端纯函数只能统一"算法"，无法统一"输入"。三套判据之所以不一致，根源是 A 的三份数据。A 落地后仍建议由后端出字段，避免前端在"页面行缺失回退 patch"时重新实现兜底逻辑。

### B.3 存量"有图无主图"的处理

- 判定上按"有图"计算就绪，**不自动改 `is_primary`**，只在 UI 打"历史图片（未设主图）"标。
- 是否补设主图属于 E 的"需人工确认"档，避免一次读操作产生写副作用。

## 4. 议题 C：重复实现怎么收口

### C.1 分类处置

| 能力 | 现状 | 处置 | 落点 |
| --- | --- | --- | --- |
| 在途等待轮询 | 批量工厂、私人定制两套 | 必抽 | `frontend/src/shared/comic/comicPageWait.js`（纯函数 + hook） |
| 409 解释与是否继续等待 | 一边有一边无 | 必抽，与轮询同一模块 | 同上 |
| 打包下载 + 文件名清洗 | 三套文件名 | 必抽；文件名以服务端 `Content-Disposition` 为准，前端只做兜底清洗 | `frontend/src/shared/api/comicDownload.js` + 后端统一响应头 |
| 就绪判定 | 三套 | 必抽（改为消费 B 的后端字段） | 后端 readiness 函数；前端仅渲染 |
| 图片鉴权加载 | 两套（`AuthenticatedComicPageImage` 与其他） | 先核对，能复用则复用，否则抽一个只读图片组件 | 视核对结果 |
| 页面壳 / 分镜卡片 UI | 两套 | 保留 | 不改 |
| 整本批量触发状态机 | 仅批量工厂 | 保留 | 不改 |
| N8 发布面板 | 仅批量工厂 | 保留 | 不改 |

### C.2 取舍说明

私人定制是"整页纵向编辑 + 逐页下载"，批量工厂是"表格单元格 + 批量触发 + 发布"，交互形态天然不同。若为了统一把卡片 UI 合并，会牺牲私人定制的可编辑密度，回归面覆盖两个页面所有交互，收益不成比例。因此统一只发生在**行为层**（等待、错误语义、下载、就绪），外观层保留两套。

## 5. 议题 D：35 页上限

### D.1 成本估算依据

- 单页出图 2.7–6 分钟，上游单请求 5–6 分钟会出现 RST，因此不能靠提高并发换时间。
- 单书页级 worker 当前为 5（`batch_factory_comic_pages.go:770`），全局出图名额 8，跨书共享。
- 35 页理想耗时约 20–26 分钟，实测 30–90 分钟（排队、重试、内容审核降级）。
- 页数翻到 60：单本上游请求量 +71%，单本占用全局名额的时间接近翻倍，批末回补的等待被放大，且尾部书籍更容易超时。

### D.2 备选与取舍

| 方案 | 改动面 | 迁移成本 | 回归风险 | 对在跑批次 | 可回滚 |
| --- | --- | --- | --- | --- | --- |
| ① 后端唯一常量 + capabilities 下发，默认 35（推荐） | 删前端两份常量与文案拼接；capabilities 增 `limits.maxComicPages` | 无 | 低 | 无 | 高 |
| ② 按书可配 `comicMaxPages`（默认 35，钳制 1–60） | 复用现有 settings patch，前端读取设置 | 中（新增配置项与校验） | 中：用户可把单本推到 60，成本和失败率上升 | 中：老批次默认不变 | 高：清空配置即回退 |
| ③ 取消上限，只靠 8 路配额 | 删三处判断 | 无 | 高：超长书独占名额数小时，批次公平性崩坏 | 高 | 中 |

结论：本轮做 ①，把 35 的语义写成"单本单次批量出图的安全页数"，前端只显示后端下发值。② 作为后续可选，需用户明确接受"更长单本、更慢批次"后再做。③ 否决。

## 6. 议题 E：历史数据收尾

### E.1 处置矩阵

| 类别 | 数量 | 处置 | 自动化程度 | 回滚方式 |
| --- | --- | --- | --- | --- |
| 提示词重复编译 | 1428 页 | 用 `StripCompiledComicImagePrompt` 重写 `image_prompt`，仅当结果为非空且更短 | 自动（抽样人工确认≥20 条后全量） | 原文备份 JSONL，可逐行还原 |
| patch-only 页 | 237 行 | 回填页面行（属 A，不重复计） | 自动 | 删除新增行 |
| 僵尸 `generating` | 6 条 | 按 `updated_at`/租约判定置 `generation_failed` 并释放租约 | 自动（若 1.4 上线后未覆盖历史行） | 记录原状态与原 lease |
| failed 但有图 | 20 页 | 人工确认后置主图、状态改 `succeeded` | 半自动（先出清单） | 记录原状态与原主图 |
| 状态与图不一致 | 2 页 | 人工确认 | 人工 | 同上 |
| 重复合集 | 10 本（1 本 5 个内容完全相同） | 只出报告，不删除、不取消发布 | 只读 | 无需 |

### E.2 关键约束

- `snapshot_hash` **不回写**。它的语义改为"最近一次生成契约哈希"，不是"当前提示词哈希"；`configStale` 只看原文 digest，因此去重不会阻塞合集创建。
- 去重只剥离编译器的固定段（画风名称、参考图规则、页面结构约束、画风锁定、前后缀与负向约束），不得用启发式删句，避免误删用户手工内容。
- 全部操作先 `--dry-run` 输出差异报告（每页 before/after 长度、被删段清单、前后 40 字上下文），再由人工决定全量执行。
- 按书加锁执行，单书处理期间该书的自动化必须处于暂停状态，避免与在途写入互相覆盖。
- 执行入口沿用 `cmd/qiantie` 现有子命令模式（已有 `healthcheck` 先例），新增 `comic-repair` 子命令，不引入新的发布路径。

### E.3 执行顺序

E 必须晚于 A：在双写还在时去重，会被下一次自动化写入再把编译结果写回，等于白做。

## 7. 议题 F：capabilities 与 Slice 门控

### F.1 推荐方案

1. **能力声明由注册结果派生。** 保留 `capabilitySpec{key, minSlice, probeRequest, reason}` 表，但 `Available` 不再由 `slice >= n` 直接决定，而是启动时用 `ServeMux.Handler(probeRequest)` 探测真实注册：命中 pattern 且未落到 NotFound 才算 available。理由文本仍按 `minSlice`/运行态给出。
2. **未注册路由返回 JSON。** 在 v11/v12 子 mux 外层加一个只在 404 时改写响应体的中间件（其余状态原样透传），输出 `{"error":"该能力未开启","code":"CAPABILITY_NOT_ENABLED","feature":"...","requiredSlice":n}`；HTTP 状态码**保持 404**，避免改动按状态码分支的客户端。这里不采用通配路由，因为通配会吞掉 405 语义。
3. **Slice=3 不变。** 用金标测试把 slice 0–6 的能力真值表逐项固定（含 production/merge/publish 现有规则），改造前后必须完全相等。
4. **启动可见性。** 启动摘要打印 slice、已开启能力数量与未开启能力列表，沿用现有启动摘要打印位置。

### F.2 备选与取舍

| 方案 | 改动面 | 迁移成本 | 回归风险 | 对在跑批次 | 可回滚 |
| --- | --- | --- | --- | --- | --- |
| ① 探测派生 + 404 归一（推荐） | `batch_factory_v11_capabilities.go` + router 包装 + 测试 | 无 | 低（状态码不变） | 无 | 高 |
| ② 手工维护 `capability -> 条件表达式` 双份 | 两处同步修改 | 无 | 中：仍会漂移 | 无 | 高 |
| ③ 删除 capabilities，前端探活 | 前端每能力一次请求 | 中 | 中：首屏请求增多、缓存复杂 | 无 | 中 |

取舍说明：② 只是把漂移从"文案"换成"条件表达式"，不解决根因；③ 违背现网"前端先读能力再渲染"的既有结构，改动面大于收益。

## 8. 需要用户拍板的清单

| 编号 | 问题 | 默认建议 | 不选/选反的后果 |
| --- | --- | --- | --- |
| P1 | 历史数据迁移是否接受停服？ | 不停服：按书暂停自动化 + 在线执行 | 停服虽更简单，但会产生 10 分钟以上的全站不可用 |
| P2 | A 是否采用"page 行为准、video patch 降级为投影"？ | 是 | 否则只能选②重做存储或③跨 Store 事务，改动与风险显著上升 |
| P3 | 237 条 patch-only 是否允许自动回填页面行？ | 是（纯新增，可删） | 不回填则前端切到 page 权威后这批书"提示词消失" |
| P4 | 是否允许前端停止写 video patch（后端投影继续写）？ | 是 | 不停则双写继续，1429 行漂移会持续增长 |
| P5 | "有图无主图"是否在读取时自动补主图？ | 否：只打标，修复归 E 人工档 | 自动补会引入读副作用，且改变用户可见的选中图 |
| P6 | 20 条"failed 但有图"是否自动置为成功？ | 否：先出清单人工确认 | 自动置成功会把真实失败伪装成成功，掩盖上游问题 |
| P7 | 10 本重复合集是否允许自动删除？ | 否：只出报告 | 自动删会牵动已发布内容与 N8 侧数据，不可逆 |
| P8 | 提示词去重是否允许改写 `snapshot_hash`？ | 否：不改，仅备份原文 | 改写会制造"看起来刚生成"的假象；不改也不影响 stale 判定 |
| P9 | 35 页是否按书可配？ | 本轮否，保持 35 并由后端下发 | 放开后单本可到 60，批次等待与失败率上升 |
| P10 | 未注册路由是否改为 JSON 错误体（仍 404）？ | 是 | 不改则继续返回裸文本，用户与日志无法区分"未开启"与"路径写错" |
| P11 | 统一轮询参数？ | 5 秒间隔、单次等待上限 20 分钟（与既有补图轮询一致） | 沿用两套参数会继续出现"一个入口自动刷新、一个入口卡住" |
| P12 | 是否允许在真实库上先跑只读 dry-run？ | 是（只读，不写） | 不跑 dry-run 直接全量，误删风险不可控 |

## 9. 分阶段落地切分

已排进执行表的阶段 0.2 / 1.1–1.6 不在此列；下表只覆盖本方案六个议题。

| 阶段 | 内容 | 前置依赖 | 冲突文件（需排他） | 验证 |
| --- | --- | --- | --- | --- |
| S1 | B 后端 readiness + F capabilities 探测与 404 归一 | 无（可与在跑阶段并行，但 router 冲突） | `router.go`、`batch_factory_v11_capabilities.go`、`batch_factory_runtime_summary.go`、`automation_status.go` | 金标测试 slice 0–6；能力 available ⇒ 探测命中；`go build ./...`、相关单测 |
| S2 | A 唯一写入口 + 237 回填 + 读侧兜底 | 1.2 编译幂等已合并；S1 的页级字段可复用 | `comic_prompt_service.go`、`comic_pages_mysql.go`、`batch_factory_comic_pages.go`、`batch_factory_automation.go`、`BatchFactoryNovelList.jsx`、`ScriptPage.jsx`、`PrivateComicWorkspace.jsx`、`batchFactoryV12.js` | 两入口同页文本一致；page/patch 一致率；回填幂等；单测 |
| S3 | C 前端行为抽共用（等待/409/下载） | S2 完成后（下载与就绪依赖新字段） | 上述两个前端页面 + 新增 shared 模块 | `node --test`、两个页面走查 |
| S4 | D 35 常量收口与 capabilities 下发 | S1（capabilities 字段） | `BatchFactoryNovelList.jsx`、capabilities 响应 | 前端不再出现本地上限常量；上限来自接口 |
| S5 | E 历史数据修复：dry-run → 抽样确认 → 全量 | S2（单写入口）、S1（readiness 用于校验） | `cmd/qiantie/main.go` 与新增 repair 包；不动前端 | dry-run 差异报告；去重后标记数归零；合集可创建 |
| S6 | 摘除过渡兜底（patch 读回退、旧 35 文案、404 裸文本） | S1–S5 全部上线并观察 2 周 | 各阶段涉及文件 | 回归清单 + 金标测试仍绿 |

依赖关系要点：

- S2 不能与 S5 并行在同一本书上执行。
- S3 的下载统一若先于 S2 落地，会出现"统一了文件名但仍读旧的提示词来源"，验收口径不一致。
- S4 可提前到 S1 之后独立上线，不依赖 S2。

## 10. 验收标准（可核对）

1. 同一本书在批量工厂与私人定制看到的"提示词就绪 / 已出图 / 可建合集"计数完全一致，且来自同一接口字段。
2. 同一页在两个入口编辑后，另一入口刷新即看到同一文本；`comic_pages` 与 video patch 一致率 ≥ 99.9%，剩余差异为 revision 冲突且有日志。
3. 1428 页去重后，`image_prompt` 中 `【最终画风锁定】` 出现次数为 0；抽样 20 页在相同样式设置下重编，除被删重复段外无其他差异。
4. 全部能力项满足"声明 available 必须真实可路由"，不存在 available 却 404/405 的项。
5. Slice=3 启动后 V12 批次相关路由全部返回 200（沿用"作品消失"事故后的验证清单）。
6. 任何修复脚本均可 dry-run，且都能用备份文件逐行还原。

## 11. 明确不做

- 不合并 `local`/`giant` 两套执行器表。
- 不重构漫画阶段链（`comic_assets → comic_bootstrap → comic_prompt_recovery → comic_resume → comic_images → comic_publish`）。
- 不删除历史分支、不改部署链路、不新增 Node 发布入口。
- 不做跨入口 UI 统一，不引入新依赖（下载沿用现有 zip 能力）。
- 不在本方案内改数据库 schema。

## 12. 附：直接相关的代码位置

- 编译与剥离：`backend/internal/batchfactoryv11/comic_image_contract.go`
- 自动化提示词写入：`backend/internal/batchfactoryv11/comic_prompt_service.go`
- 页面行读写与租约：`backend/internal/batchfactoryv11/comic_pages_mysql.go`、`backend/internal/httpapi/batch_factory_comic_pages.go`
- 合集预检：`backend/internal/httpapi/batch_factory_comic_pages.go` 中 `comicCollectionPreflight`
- 就绪与阶段文案：`frontend/src/user/pages/shuihuo/batchFactoryBookState.js`
- 前端两处写入：`frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`、`frontend/src/user/pages/ScriptPage.jsx`
- 能力表：`backend/internal/httpapi/batch_factory_v11_capabilities.go`
- 路由门控与别名：`backend/internal/httpapi/router.go`、`backend/internal/httpapi/batch_factory_v12_aliases.go`
- 35 页常量：`backend/internal/httpapi/batch_factory_comic_pages.go:40`、`frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx:1646`
