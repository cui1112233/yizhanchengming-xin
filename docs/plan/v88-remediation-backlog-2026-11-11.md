# V88 整改任务表（2026-11-11）

> 依据：四轮排查（业务正确性、前端界面行为、主题与按钮、安全）+ 图片积压专项的结论汇总。
> 用途：按批次推进，每批一次验收，验收通过才进下一批。
> 说明：工作区当前有多个会话并行改动，`git diff` 里混有未完成中间态。本文按现状登记，不评价对错；文件路径以实际存在的为准，行号可能漂移。
> 约束：本文件只做跟踪，不代表任何一项目前已上线。

---

## 1. 一页概览

| 批次 | 主题 | 状态 | 依赖 | 验收方式 |
|---|---|---|---|---|
| 批次一 | 已改完待验收（错误文案 / 重复扣钱 / 主题第一步 / 安全三条） | 部分进行中 | 无（当前工作区） | 构建 + 重启 4001 后按《批次一验收清单》逐条点 |
| 批次二 | 安全收尾 | 未开始 | 批次一验收通过 | 安全用例 + 登录/注册/重置链路回归 |
| 批次三 | 存量数据与遗留 | 未开始 | 批次二（清理动作需授权） | 库内计数核对 + 测试退出失败状态 |
| 批次四 | 错误统一收口 | 未开始 | 批次一（文案改动已落地） | 抽样 20 个页面触发错误，检查主文案与原文分离 |
| 批次五 | 主题与按钮 | 未开始 | 批次一主题第一步通过 | 逐页截图对比 + 组件单测 |
| 批次六 | 设计债（8 条，6-8 待产品拍板） | 未开始 | 图片止血线（1-18~1-21）停手；6-3 依赖批次四 4-1 | 抽公共层 + 删废弃文件 + 逐条手工回归，见各条验收 |

**计数：共拆 53 条任务**（批次一 25、批次二 6、批次三 6、批次四 4、批次五 4、批次六 8）。

批次内部两类特殊标记：
- `[进行中]`：对应改动仍在别的会话手上，本批验收前必须等它落地。
- `[待授权]`：需要用户明确点头才能执行（删数据、调并发、对外发布）。

---

## 2. 批次一 · 已改完待验收

### 2.1 错误文案三轮（5 条）

#### 1-1 批量工厂兜底文案中文化 `[已完成]`
- **改什么**：前端不再直接把后端错误串弹给用户；批量工厂的兜底文案统一走中文映射，未命中时给通用中文提示，原文放次要位置。
- **涉及文件**：`frontend/src/user/pages/shuihuo/batchFactoryErrorParse.js`、`batchFactoryErrorParse.test.js`、`batchFactoryVisibleError.test.js`、`shuihuo/TaskDrawer.jsx`
- **为什么改**：对应「批量工厂兜底文案」排查结论——英文哨兵前缀（`invalid input`、`temporary upstream failure` 等）直接出现在界面上。
- **验收标准**：制造一次上游 4xx，界面主文案为中文，无英文前缀；展开后能看到上游原文。

#### 1-2 漫画页提示词相关文案 `[已完成]`
- **改什么**：漫画页提示词缺失 / 生成失败的提示改为可读中文，并区分「提示词为空」「模型返回为空」「上游拒绝」三种情况。
- **涉及文件**：`backend/internal/batchfactoryv11/comic_prompt_service.go`、`comic_prompt_service_test.go`、`backend/internal/httpapi/batch_factory_comic_pages.go`、`batch_factory_comic_pages_test.go`
- **为什么改**：对应「漫画页提示词」排查结论。
- **验收标准**：清空提示词后触发生成 → 中文提示；返回空 → 提示「模型返回为空」，不是裸错误码。

#### 1-3 图片通道 4xx 文案 `[已完成]`
- **改什么**：图片通道返回 4xx 时，主文案给结论（通道拒绝 / 鉴权失败 / 参数不接受），上游响应体不直接当主文案。
- **涉及文件**：`backend/internal/httpapi/shuihuo_projects.go`、`batch_factory_v12_image.go`、`shuihuo_image_attempt_guard_test.go`、`shuihuo_image_chat_channel_retry_test.go`
- **为什么改**：对应「图片通道 4xx」结论——`image provider returned HTTP 400` 这类拼接串直接上屏。
- **验收标准**：用错误 Key 触发一次出图 → 主文案中文且指出是通道鉴权问题；展开可见上游原文。

#### 1-4 文本模型请求失败文案 `[已完成]`
- **改什么**：文本模型调用失败按「未配置执行器 / 模型返回为空 / 上游错误」分类给中文文案。
- **涉及文件**：`backend/internal/httpapi/model_catalog.go`、`model_catalog_test.go`、`chat.go`、`chat_test.go`、`tts.go`
- **为什么改**：对应「文本模型请求失败」结论。
- **验收标准**：断网或填错 Key，封面提示词 / 对话 / 配音三条链路各触发一次，均给中文分类文案。

#### 1-5 状态行中文化 `[已完成]`
- **改什么**：批量工厂任务抽屉、书本状态、自动化状态行的英文状态串改为中文展示。
- **涉及文件**：`frontend/src/user/pages/shuihuo/TaskDrawer.jsx`、`batchFactoryBookState.js`、`backend/internal/httpapi/automation_status.go`、`automation_status_test.go`
- **为什么改**：对应「状态行中文」结论。
- **验收标准**：任务抽屉里所有状态徽标为中文；无 `pending/running/failed` 裸串。

### 2.2 前端重复扣钱与丢成品（6 条）

#### 1-6 API 配置 Key 半截落库 `[已完成]`
- **改什么**：Key 输入改为「失焦 / 显式确认后整体提交」，避免输入中途按字符触发保存；保存前做完整性校验。
- **涉及文件**：`frontend/src/user/pages/ApiConfigPage.jsx`、`api-config-preset-state.js`、`frontend/src/shared/api/modelCatalog.js`、`backend/internal/httpapi/user_config.go`
- **为什么改**：对应「API 配置 Key 半截落库」——边输边存导致库里是半截 Key，生成时才报鉴权失败。
- **验收标准**：输入完整 Key 过程中刷新页面，库里不应出现前缀片段；保存后重启回读一致。

#### 1-7 配音重复下单与失败清空旧音频 `[已完成]`
- **改什么**：下单按钮加防重；合成失败保留原有音频，不用失败结果覆盖旧文件。
- **涉及文件**：`frontend/src/user/pages/TtsPage.jsx`、`frontend/src/shared/api/tts.js`、`backend/internal/httpapi/tts.go`、`tts_test.go`
- **为什么改**：对应「配音重复下单与失败清空旧音频」——连点重复计费，失败后旧成品丢失。
- **验收标准**：连点 5 次只产生 1 条任务；人为让上游失败，旧音频仍可播放。

#### 1-8 小说抓取等待与误判 `[已完成]`
- **改什么**：抓取任务等待态与失败态区分；轮询有上限，超时给明确结论而不是一直转圈。
- **涉及文件**：`frontend/src/user/pages/NovelFetchPage.jsx`、`backend/internal/httpapi/novel_fetch_workshop.go`、`novel_fetch_workshop_test.go`
- **为什么改**：对应「小说抓取等待与误判」。
- **验收标准**：正常抓取 1 本成功；停掉上游后 1 个轮询周期内显示失败原因，不无限等待。

#### 1-9 封面无限轮询 `[已完成]`
- **改什么**：封面生成轮询加终止条件（成功 / 失败 / 超时），组件卸载时停表。
- **涉及文件**：`frontend/src/user/components/CoverGenerationWorkbench.jsx`、`frontend/src/shared/api/coverGeneration.js`
- **为什么改**：对应「封面无限轮询」——失败后仍持续打接口。
- **验收标准**：失败后 30 秒内停止请求（DevTools Network 不再新增）；切页后无残留请求。

#### 1-10 素材参考跨账号可见 `[已完成]`
- **改什么**：素材参考查询/下载补 owner 过滤，禁止按 ID 越权读取他人素材。
- **涉及文件**：`backend/internal/httpapi/reference_assets_settings.go`、`reference_assets_image.go`
- **为什么改**：对应「素材参考跨账号可见」安全结论。
- **验收标准**：A 账号的素材 ID 在 B 账号请求下返回 403/404，不返回图片内容。

#### 1-11 AI 剪辑草案重复应用 `[已完成]`
- **改什么**：草案应用加幂等键/版本校验，重复点击或重复回包只应用一次。
- **涉及文件**：`backend/internal/aicut/types.go`、`backend/internal/httpapi/aicut_export.go`、`aicut_export_test.go`、`frontend/src/user/components/AICutAutoEditDrawer.jsx`
- **为什么改**：对应「AI 剪辑草案重复应用」——同一草案被叠加两次，时间轴错乱。
- **验收标准**：同一草案连续应用两次，时间轴只变化一次；刷新后仍一致。

### 2.3 主题第一步（3 条）

#### 1-12 antd 深浅色算法 `[已完成]`
- **改什么**：antd 主题算法跟随 V88 主题切换，深色/浅色不再各自硬编码一套 token。
- **涉及文件**：`frontend/src/shared/styles/theme.js`、`theme-catalog.js`、`ThemePicker.jsx`、`frontend/src/user/App.jsx`、`main.jsx`
- **为什么改**：对应「主题」排查第一步。
- **验收标准**：切换深/浅色，antd 组件（表格、弹窗、按钮）底色与文字同步变化，无白块残留。

#### 1-13 滚动条样式 `[已完成]`
- **改什么**：滚动条颜色/宽度走主题变量。
- **涉及文件**：`frontend/src/shared/styles/global.css`
- **验收标准**：深浅色下滚动条均可辨识，不出现浅底白条。

#### 1-14 斑马纹与焦点描边 `[已完成]`
- **改什么**：表格斑马纹、键盘焦点描边使用主题变量，深色下对比度达标。
- **涉及文件**：`frontend/src/shared/styles/global.css`、`theme-css-contract.test.js`
- **验收标准**：Tab 键遍历按钮/输入框，焦点环在深色下可见；表格斑马纹不刺眼。

### 2.4 安全三条（3 条）

#### 1-15 web-submit 自签身份 `[已完成]`
- **改什么**：web-submit 不再信任请求自带的身份字段，身份只从已认证会话取。
- **涉及文件**：`backend/internal/httpapi/web_submit_compat.go`、`web_submit_submit_compat.go`、`web_submit_preview_compat.go`、`web_submit_compat_test.go`
- **为什么改**：对应「web-submit 自签身份」安全结论。
- **验收标准**：伪造 `username` 字段提交 → 写入的是当前登录用户或直接 401，不落到伪造身份名下。

#### 1-16 Agent 工作区任意路径 `[已完成]`
- **改什么**：Agent 工作区文件读写限定在用户工作目录内，拒绝 `..` 与绝对路径逃逸。
- **涉及文件**：`backend/internal/httpapi/agent_workspace_compat.go`、`agent_proxy_routes.go`、`backend/internal/storage/agent_proxy_schema.go`
- **为什么改**：对应「Agent 工作区任意路径」安全结论。
- **验收标准**：传 `../../etc/passwd` 或 `C:\Windows\win.ini` → 400/403，不返回文件内容。

#### 1-17 找回密码 Host 头 `[已完成]`
- **改什么**：重置链接的 Host 不再取自请求头，改用配置的站点地址白名单。
- **涉及文件**：`backend/internal/httpapi/account_recovery_compat.go`、`backend/internal/config/config.go`
- **为什么改**：对应「找回密码 Host 头」——可被伪造 Host 骗取重置 token。
- **验收标准**：带伪造 `Host: evil.com` 请求重置，邮件链接域名仍为配置域名。

### 2.5 图片止血四条 `[进行中]`

> 这四条对应图片积压专项，改动仍在别的会话手上；本批验收前需确认已落地。

#### 1-18 并发槽位不再被单请求长期占死 `[进行中]`
- **改什么**：出图并发协调器区分「排队等待」与「上游请求」两段计时，超时释放槽位。
- **涉及文件**：`backend/internal/httpapi/shuihuo_image_adaptive_concurrency.go`、`shuihuo_image_semaphore_test.go`、`shuihuo_projects.go`
- **验收标准**：人为让一路请求挂起，其余队列在等待上限内推进，不出现全员卡死。

#### 1-19 页级任务超时后落终态 `[进行中]`
- **改什么**：页级 detached 生成超时后写失败终态并释放账本，不再留在「生成中」。
- **涉及文件**：`backend/internal/httpapi/batch_factory_comic_pages.go`、`backend/internal/batchfactoryv11/comic_pages_mysql.go`、`comic_pages_memory.go`
- **验收标准**：超时页在列表里变为失败并可重试，不再是「生成中」。

#### 1-20 上游 4xx 不重试 `[进行中]`
- **改什么**：4xx（鉴权/参数）直接失败，只有 5xx/超时才进入重试阶梯。
- **涉及文件**：`backend/internal/httpapi/shuihuo_projects.go`、`shuihuo_image_attempt_guard_test.go`
- **验收标准**：错误 Key 触发出图，上游只被调用 1 次（看日志计数），不放大成多次计费。

#### 1-21 积压任务可观测 `[进行中]`
- **改什么**：暴露排队数、在跑数、超时数，便于判断积压而不是靠日志翻。
- **涉及文件**：`backend/internal/httpapi/shuihuo_image_adaptive_concurrency.go`、`automation_status.go`
- **验收标准**：压入 10 个任务后，状态接口能读到排队/在跑/失败三个数。

### 2.6 错误与状态四条 `[进行中]`

#### 1-22 失败分类只认结构化字段 `[进行中]`
- **改什么**：失败分类优先读结构化错误码，英文关键词匹配降级为兜底。
- **涉及文件**：`backend/internal/httpapi/automation_failure.go`、`automation_failure_test.go`
- **验收标准**：构造 `insufficient balance` 类错误，分类结果稳定为「额度不足」，不随文案大小写变化。

#### 1-23 终止类错误不再进重试阶梯 `[进行中]`
- **改什么**：配置错误/鉴权错误/额度不足标记为终止，直接落失败，不跑满重试。
- **涉及文件**：`backend/internal/httpapi/automation_failure.go`、`batch_factory_automation.go`
- **验收标准**：额度不足时单页只尝试 1 次，日志无第 2、3 次上游调用。

#### 1-24 状态推进幂等 `[进行中]`
- **改什么**：同一页重复回包不重复推进状态，避免「已完成 → 生成中」回退。
- **涉及文件**：`backend/internal/httpapi/automation_status.go`、`batch_factory_automation.go`、`batchfactoryv11/stage_service.go`
- **验收标准**：重复触发一次补页，已完成页状态不变。

#### 1-25 状态与错误展示分离 `[进行中]`
- **改什么**：界面「状态」列与「错误」列分开，失败页不再整行显示长错误串。
- **涉及文件**：`frontend/src/user/pages/shuihuo/TaskDrawer.jsx`、`batchFactoryBookState.js`、`backend/internal/httpapi/automation_status.go`
- **验收标准**：失败页状态显示「失败」，错误在详情里可展开查看。

---

## 3. 批次二 · 安全收尾（6 条）

### 2-1 剩余六处用户名拼路径
- **改什么**：把用户名直接拼进文件路径的六处改为 ID 映射 + 白名单校验（对应 1-16 的同源问题，只是剩余点位）。
- **涉及文件**：`backend/internal/httpapi/agent_workspace_compat.go`、`aicut_assets.go`、`aicut_fonts.go`、`localartifact` 调用点（以全仓 `filepath.Join(..., owner...)` 检索为准）
- **验收标准**：用户名含 `../`、`\`、空格、中文时，文件仍落在该用户目录内。
- **依赖**：无。

### 2-2 旧 web-submit 别名仍认签名身份
- **改什么**：`/api/batch-rewrite/web-submit/*` 与 `/api/novel-fetch-workshop/web-submit/*` 两个入口统一走同一身份解析，别名不允许绕过认证。
- **涉及文件**：`backend/internal/httpapi/web_submit_*_compat.go`、`batch_rewrite_compat.go`、`router.go`
- **验收标准**：两个入口分别伪造身份，均被拒绝；正常会话均成功。
- **依赖**：1-15。

### 2-3 出站请求统一安全客户端
- **改什么**：全仓出站 HTTP 走同一个 client：内网地址黑名单（127/10/172.16/192.168/169.254/::1）、重定向逐跳校验、超时统一。
- **涉及文件**：`backend/internal/httpapi` 下 20+ 处 `&http.Client{}`、`backend/internal/videomanagementconnector/target_http.go`、`backend/internal/localexecutor`
- **验收标准**：把渠道 Base URL 指向 `http://127.0.0.1:4001` 或 `http://169.254.169.254` 被拒绝；正常公网渠道可用。
- **代价提示**：这是本批唯一会触碰所有出站链路的改动，建议单独一次构建验证。

### 2-4 密钥加密存储并清理历史明文
- **改什么**：API Key / 渠道凭证落库前加密（现有 `target_credential_store.go` 基础上统一），并提供一次性清理历史明文的任务。
- **涉及文件**：`backend/internal/httpapi/target_credential_store.go`、`user_config.go`、`api_integrations.go`、`backend/internal/storage/auth_schema.go`
- **验收标准**：库内看不到明文 Key；旧数据被重写为密文后接口仍可正常调用。
- **依赖**：`[待授权]` 历史明文清理属于写库操作，执行前需用户确认时间窗。

### 2-5 登录/注册/重置邮件限流
- **改什么**：按 IP + 账号双维度限流，防撞库与邮件轰炸。
- **涉及文件**：`backend/internal/httpapi/account_recovery_compat.go`、`account_passkey_compat.go`、`mfa_compat.go`
- **验收标准**：同一账号 1 分钟内第 6 次重置请求返回 429；正常用户不受影响。

### 2-6 请求体与解压上限
- **改什么**：全局请求体上限 + 压缩包解压体积/文件数上限，防 zip bomb 与超大上传。
- **涉及文件**：`backend/internal/httpapi/router.go`（中间件）、`novel_fetch_upload_compat.go`、`aicut_assets.go`
- **验收标准**：上传 100MB 文件被拒；构造高压缩比 zip 被拒且不撑爆磁盘。

---

## 4. 批次三 · 存量数据与遗留（6 条）

### 3-1 卡在「生成中」的 6 页 `[待授权]`
- **改什么**：先只读导出这 6 页的 id/批次/时间/最后错误，确认是孤儿任务还是仍在跑；确认后统一置失败并允许重试。
- **涉及文件**：`backend/internal/batchfactoryv11/comic_pages_mysql.go`、`backend/internal/httpapi/batch_factory_comic_pages.go`
- **验收标准**：处理前后计数对比（6 → 0）；对应页可重新生成。
- **依赖**：用户授权写库。

### 3-2 28 页无主图 `[待授权]`
- **改什么**：判定是「生成成功但未写主图标记」还是「图片丢失」；前者补标记，后者置失败重跑。
- **涉及文件**：`comic_pages_mysql.go`、`comic_collections` 相关查询、`shuihuo_projects.go`
- **验收标准**：28 页全部有明确归属（有主图或失败可重试），无悬挂引用。

### 3-3 91 张资产图如何处置 `[待授权]`
- **改什么**：区分孤儿文件（DB 无引用）与重复文件（同 asset 多版本），给出保留/删除清单，先出清单再执行。
- **涉及文件**：`backend/internal/httpapi/shuihuo_projects.go`、`backend/internal/localartifact`、`book_asset_images` 查询
- **验收标准**：清单经用户确认后执行；执行后孤儿数为 0，引用完整性校验通过。

### 3-4 错误字段 200 字节截断
- **改什么**：确认错误列长度，评估加长到 TEXT 或改为「摘要 + 详情」两列；加长需迁移。
- **涉及文件**：`backend/internal/storage/batch_factory_v11_schema.go`、`batch_factory_v12_schema.go`、`migrations.go`
- **验收标准**：长错误串完整落库，界面展开可读全文；迁移可回滚。

### 3-5 两条长期失败测试（H3 配置 / 图片并发）
- **改什么**：分别判定是「断言过期」还是「实现真有问题」；过期就改断言并注明原因，真问题就修实现。不允许长期红着。
- **涉及文件**：`backend/internal/batchfactoryv11/*h3*_test.go`、`backend/internal/httpapi/shuihuo_image_*_test.go`
- **验收标准**：`go test ./...` 全绿；若保留跳过，必须有 `t.Skip` 理由与跟踪条目。

### 3-6 两条被带坏的前端测试（音频模型 key / 画幅档位）
- **改什么**：恢复断言或更新期望值，确认是别的改动连带影响。
- **涉及文件**：`frontend/src/user/pages/aicut-workbench.layout.test.js`、`frontend/src/shared/imageAspectRatios.test.js`、`frontend/src/shared/api/modelCatalog.test.js`
- **验收标准**：对应测试文件单独跑全绿；`npm test` 无新增红。

---

## 5. 批次四 · 错误统一收口（4 条）

### 4-1 统一错误转换函数
- **改什么**：后端所有错误返回前过同一个 `toClientError`：主文案取中文映射表，原始信息进 `detail`/`raw` 字段，前端只渲染主文案。
- **涉及文件**：`backend/internal/httpapi/client_errors_compat.go`、`router.go`、各 handler 的 `writeJSON(...Status..., map[string]string{"error": ...})`
- **验收标准**：全仓不再有 handler 把上游响应体拼进 `error` 字段；新增错误码只需改映射表一处。

### 4-2 前端 20+ 页面中文兜底
- **改什么**：所有直接 `message.error(err.message)` 的地方改为「中文兜底 + 可选展开原文」。
- **涉及文件**：`frontend/src/user/pages/*.jsx`（重点：水货生产、小说获取、AI-CUT、个人中心、Agent）、`frontend/src/shared/api/client.js`
- **验收标准**：随机触发 20 个页面的失败，界面无英文/裸 HTTP 状态码。
- **依赖**：4-1。

### 4-3 Agent 对话失败不再伪装正常
- **改什么**：Agent 对话上游失败时标记消息为失败态并允许重试，不再把错误文本当正常助手回复。
- **涉及文件**：`backend/internal/httpapi/chat.go`、`agent_chat_routes.go`、`frontend/src/user/pages/AgentChatPage.jsx`
- **验收标准**：断网后发消息，气泡显示「发送失败/重试」，不是一段错误说明当回答。

### 4-4 漫画页分页失败不保存错误原文（对应 1-25 的界面缺口）
- **现象**：漫画页分页生成失败时，服务端只落 `status=generation_failed / generationState=failed`，**不保存错误原文**；列表行只显示「生成失败」，任务抽屉 `TaskDrawer.jsx` 把错误平铺成一段不能折叠，「展开原始信息」只存在于书籍级/自动化级的 `BatchFactoryErrorNotice`。
- **要做的**：分页失败时持久化结构化错误（主文案 + 原始信息 + 状态码），列表点开能看到中文原因并可展开原文；任务抽屉也支持展开。
- **验收标准**：制造一次真实失败后，在页面和任务抽屉里都能看到中文原因，点开能看原文，且原文不进主文案。

---

## 6. 批次五 · 主题与按钮（4 条）

### 5-1 删除 111 条 body 类名浮层规则
- **改什么**：把依赖 `body.xxx` 类名撑浮层的 111 条规则换成基于根节点主题标记（`data-theme`）的方案，浮层挂到主题根。
- **涉及文件**：`frontend/src/shared/styles/global.css`、`theme-catalog.js`、各页面 `.css`（以 `body.` 前缀检索为准）
- **验收标准**：切换主题后所有浮层（下拉、Tooltip、Modal、Drawer）颜色正确；删除后无「白底白字」。

### 5-2 逐页迁移写死颜色
- **改什么**：按页替换硬编码色值为主题 token。规模：水货生产 814 处、个人中心 340 处、小说获取 172 处、AI-CUT 93 处。
- **涉及文件**：`frontend/src/user/pages/shuihuo/*`、`MemberCenterPage.jsx`、`NovelFetchPage.jsx`、`aicut/editor.css`
- **验收标准**：深色主题下四页无浅色残留块；新增颜色检查规则防回归。

### 5-3 统一按钮组件
- **改什么**：四变体（主/次/危险/文字）× 三尺寸，内置 loading 与防重复点击，替换散落 `button` 与外链样式。
- **涉及文件**：新增 `frontend/src/shared/components/Button*`、`theme.js`；替换各页面按钮
- **验收标准**：同一页面按钮高度/圆角/间距一致；连点只触发一次。

### 5-4 统一动词与主操作位置
- **改什么**：同一动作统一措辞（保存/提交/生成/应用），主操作固定在一致位置。
- **涉及文件**：各页面操作区（水货生产、小说获取、插件中心、AI-CUT）
- **验收标准**：文案表里每个动作唯一写法；主操作在弹窗右下、页面右上。

---

## 7. 批次六 · 设计债（8 条）

> 依据：前端结构审查（下方 8 条按审查给出的优先级排序）+ 后端设计债核对。仓库 `docs/` 下未找到独立的后端设计债文档，后端部分按代码逐条补证，证据位置写在各条「涉及文件」里；未取到的结论不臆造。
> 标记：`[小手术]` 可立即做、回归面小；`[中重构]` 要单独开窗口；`[待产品拍板]` 决策前不动代码。
> **并行约束（必读）**：**6-1（生成任务管理）和 6-4（批次设置）都会改 `frontend/src/user/pages` 下的文件**（`shuihuo/BatchFactoryNovelList.jsx`、`ScriptPage.jsx`、`NovelFetchPage.jsx` 等），必须等「图片止血」那条线（1-18~1-21）停手、文件冻结后再单独开窗口，禁止并行改同一文件。6-3 是跨页文字替换，同样按文件分工避让 6-1/6-4。

### 6-1 生成任务管理缺失 `[小手术]`（最先做）
- **改什么**：抽出唯一的生成任务管理（提交 → 轮询带退避 → 终态：成功/失败/超时/取消，取消走 AbortController），先替换 3 份逐字复制的 `pollGeneration`；**不改 UI、不改文案**。调用方第二梯队（漫剧页、素材图、批量资产、上传升级、小说抓取）等本条验收通过后再逐个接入。
- **涉及文件**：`frontend/src/user/components/CoverGenerationWorkbench.jsx`、`CharacterTurnaroundWorkbench.jsx`、`AssetReferenceWorkbench.jsx`（三份 `pollGeneration` 逐字复制，grep 7 行 / 3 文件）；新增 `frontend/src/shared/generation/*`；后续调用方 `user/pages/shuihuo/BatchFactoryNovelList.jsx`、`ScriptPage.jsx`、`NovelFetchPage.jsx`、`batch-factory-v11/BatchFactoryV11UiPage.jsx`、`user/components/AgentCanvas.jsx`、`UpscaleWorkbench.jsx`、`shuihuo/BatchFactoryGiantMaterialPendingProgress.jsx`。
- **为什么改**：全前端「发起出图 → 轮询 → 展示」有 12 处以上独立实现，超时、取消、失败文案三套行为并存；已修好的取消/超时只覆盖部分入口，同一种失败在不同页面表现不一致。
- **验收标准**：① `grep -rn "pollGeneration" frontend/src` 只剩共享实现，三个 workbench 不再各自递归 `setTimeout`；② 三个页面 × 三场景（成功 / 上游失败 / 手动取消）表现一致，取消后 60 秒内 Network 不再新增请求，且 `signal.aborted` 为真；③ 三份原有测试与新增共享层测试全绿。
- **后端侧（本轮补证，本条不改行为）**：任务态至少三套——`assetImageGenerationTasks` 文件型 JSON 任务（`backend/internal/httpapi/batch_factory_v12_asset_generation.go`，`Status:"processing"` → `finish` 落终态，322 行）、漫剧页状态（`backend/internal/batchfactoryv11/comic_pages_memory.go` 的 `prompt_ready` / `image_ready` / `generation_failed`）、自动化状态（`httpapi/automation_status.go`）。字段命名与终态集合不一致，先登记进共享层适配表，不在本条合并后端模型。
- **风险/依赖**：**必须等图片止血线停手**（同一批 `user/pages` 文件）；替换期间不改请求参数与超时阈值，保持行为等价；不做构建、不重启 4001。

### 6-2 已废弃实现没删 `[小手术]`
- **改什么**：删除无路由引用的 `frontend/src/user/pages/batch-factory-v11/`（68 个文件）、`batch-factory/`（3 个文件）、`BatchFactoryWorkbenchPage.jsx`、`AgentPage.jsx`；同步删除只为源码字符串断言存在的 `.source.test.js`；移除 `frontend/public/batch-rewrite/` 旧静态页并重建 dist。
- **涉及文件**：上述目录/文件；路由现状见 `frontend/src/user/App.jsx`（`/batch-factory` → `BatchFactoryPageV30`、`/agent` → `AgentPageV2`，均未引用待删文件）；旧静态页 `frontend/public/batch-rewrite/`（`index.html` 31KB、`app.js` 185KB 等，随 build 复制进 `frontend/dist/batch-rewrite/`）；后端 `/api/batch-rewrite/*` 别名在 `backend/internal/httpapi/router.go:658-709` 与 `batch_rewrite_compat.go`。
- **为什么改**：这些文件只被同名 `.source.test.js` 读源码断言「续命」，实际路由早已切走；旧静态页仍随 dist 发布、可直接访问，和新实现并存，改一处要改两套。
- **验收标准**：① `grep -rn "batch-factory-v11\|BatchFactoryWorkbenchPage\|AgentPage.jsx" frontend/src --include=*.jsx --include=*.js` 只剩「允许保留清单」（见风险）内的命中；② `npm run build` 后 `frontend/dist` 无 `batch-rewrite` 目录，直接访问 `/batch-rewrite` 返回应用 404 或重定向到 `/novel-fetch`；③ `/batch-factory`、`/agent`、`/novel-fetch` 三个路由打开正常；④ `npm test` 全绿（删掉的源码断言测试一并消失，不是改成 skip）。
- **风险/依赖**：`batch-factory-v11/videoProviderCapabilities.json` 仍被管理端 `admin/pages/ShuihuoModelCatalogPage.jsx` 和 `shared/api/videoProviderCapabilities.test.js` 引用，须先迁到 `shared` 再删；`shared/styles/contrast-contract.test.js` 仍读 v11 的 5 个 css，删 css 要同步改契约；后端 `/api/batch-rewrite/*` 别名**不在本条删除范围**（前端删页不影响它，单独评估）。

### 6-3 错误展示没有统一出口 `[小手术]`
- **改什么**：收敛成唯一出口：`shared/api/client.js` 统一把后端错误转成中文主文案 + `raw/detail`，页面只渲染主文案；`apiErrorMessage` 作为唯一文案函数全量替换 `message.error(err.message)` 一类直出；`qiantie:api-error` 只保留一个展示者，页面级调用用 `suppressGlobalError` 显式声明，避免同一次失败弹两次或一次不弹。
- **涉及文件**：`frontend/src/shared/api/client.js:39`（事件派发）、`frontend/src/shared/apiErrors.js`、`frontend/src/shared/layouts/UserLayout.jsx:276-277`（唯一监听者）；重点替换页：`NovelFetchPage.jsx`（53）、`shuihuo/BatchFactoryNovelList.jsx`（89）、`AICutPage.jsx`（40）、`batch-factory-v11/BatchFactoryV11UiPage.jsx`（27）、`ComicScreenplayPage.jsx`（18）。
- **为什么改**：本轮按行扫描 `message.error` 命中 561 行 / 81 个文件（审查口径 1405 处），统一文案函数 `apiErrorMessage` 仅 25 行命中 / 7 个文件；全局错误事件只有 `UserLayout` 一个监听者，页面自己又弹一次，所以同一次失败可能弹两次也可能只弹一次。
- **验收标准**：① `grep -rn "message.error(" frontend/src --include=*.jsx` 的命中只允许出现在共享封装内部与白名单文件；② 抽 20 个页面触发失败，主文案为中文、无英文哨兵前缀，同一次失败只出现 1 条提示；③ 断网与上游 4xx 各抽 5 个页面各验一次；④ `apiErrors` 相关测试全绿。
- **风险/依赖**：依赖批次四 4-1（后端错误结构）先落地，否则前端拿不到稳定的 `detail/raw`；跨文件纯文字替换，仍按文件分工避让 6-1/6-4。

### 6-4 批次设置多写入入口 `[中重构]`
- **改什么**：收敛为「一个设置模型 + 一个读取出口 + 一个写入出口」：批次级只保留 `BatchFactoryUnifiedSettingsModal`，书本级只保留 `BatchFactoryBookSettingsModal`，`BatchFactoryV11ScopedSettings` 删除或降级只读；三份手写 normalize 合并为一份（含历史 `{patch:{...}}` 外层包裹兼容）；服务端保留 batch / book / video 三个 scope，但共用同一套 patch 校验与 revision 冲突返回。
- **涉及文件**：前端 `shuihuo/BatchFactoryUnifiedSettingsModal.jsx`（762 行）、`BatchFactoryBookSettingsModal.jsx`（797 行）、`batch-factory-v11/BatchFactoryV11ScopedSettings.jsx`（312 行）、`shuihuo/BatchFactoryNovelList.jsx`（5081/5095 等 11 处设置写入）、`user/components/AgentBatchFactoryPromptSelector.jsx`、`ScriptPage.jsx:1760/2148`、`AgentCanvas.jsx:1230`、`shared/api/batchFactoryV12.js:359/367`；后端 `httpapi/batch_factory_v11_slice1.go:346`（batch settings）、`:413`（book override）、`:414`（video override）、`batchfactoryv11/settings.go`（`normalizeLegacyNestedPatch`、`ApplySparseUpdate`）、`merge.go`、`mysql_store.go`。
- **为什么改**：设置写入入口本轮已核实不少于 8 处（审查口径 7 个），同一套字段在三个组件里各带一份 normalize；历史上把 `{"patch":{...}}` 当设置对象落库（`settings.go:16-40` 的兼容注释即证据），只能靠读取端兜底。草稿也分散：drafts API 至少 5 类 kind（`working-front-content`、`working-front-viral-candidate`、`asset-prompt`、`asset-image`、`visual-image`）外加 3 处 localStorage（`asset-styles`、漫剧发布目录缓存、巨量素材 nonce），对应审查口径的「6 套本地 draft」。
- **验收标准**：① `grep -rn "normalizeLegacyPatch\|normalizeLegacyNestedPatch" frontend/src` 只剩 1 处实现；② `grep -rn "saveBatchSettings(\|saveBookOverride(" frontend/src --include=*.jsx` 只剩共享服务层与唯一表单提交点，数量从 ≥8 降到 2-3 并附清单；③ 改一次生产模式只产生 1 次写入（看 Network），保存后 `settingsState.revision` 单调 +1，库内不再出现 `{"patch":{...}}` 嵌套；④ 两端并发保存返回 409，前端提示「配置已被更新，请刷新」，不静默覆盖。
- **风险/依赖**：**必须等图片止血线停手**（同改 `BatchFactoryNovelList.jsx` / `ScriptPage.jsx`）；动库前先只读抽样 20 个 batch 的 settings JSON 确认无 `patch` 包裹残留；不改 schema、不清理历史数据；与 6-5 同文件，必须串行。

### 6-5 BatchFactoryNovelList.jsx 巨石 `[中重构]`
- **改什么**：按文件内已有天然边界拆 5 个文件：`AssetEditor`（342 行起）、`PromptPanel`（1155 行起）、`ComicMediaPanel`（1883 行起）、`MediaVersionPanel`（2245 行起）、`UploadNetwork`（3249 行起）；纯搬运不改行为，父组件只保留状态编排与数据获取。
- **涉及文件**：`shuihuo/BatchFactoryNovelList.jsx`（实测 6293 行、约 193 个 `useState`）、`shuihuo/batchFactoryNovelListHelpers.js`；约 20 个只读源码路径的 `.source.test.js` 需同步改断言路径。
- **为什么改**：单文件 6.3k 行、193 个 useState，评审和改动成本最高；五个区块之间没有共享局部状态，天然可切。
- **验收标准**：① 父文件降到 2500 行以内且不再包含这 5 个组件定义；② 5 个新文件可被单独 import，无循环依赖；③ `npm test` 全绿（源码断言测试改读新路径）；④ 手工回归五个入口（素材编辑 / 提示词 / 漫剧页 / 合并版本 / 上传网络）与拆分前一致，附截图对比。
- **风险/依赖**：搬运期禁止顺手改逻辑；与 6-4 同文件必须串行；拆分期间不做构建发布。

### 6-6 AICutPage.jsx 单组件 `[中重构]`（排在 6-5 之后）
- **改什么**：先把 119 个 `useState` 归约为「项目 / 时间轴 / 选中对象 / 播放 / 导出」五组状态（useReducer 或分组 hook），再评估拆文件；本轮不要求拆文件。
- **涉及文件**：`frontend/src/user/pages/AICutPage.jsx`（实测 5713 行）、`user/pages/aicut/*`、`shared/api/aicut*`。
- **为什么改**：业务封闭但状态网状，直接拆文件会把隐式耦合打散成跨文件 props；先归约再拆风险最低。
- **验收标准**：① `useState` 数量降到 40 以下，或归约为不超过 6 个 reducer/hook 分组；② 新增时间轴核心链路单测（新增 / 删除 / 裁剪 / 撤销）；③ 手工回归「导入素材 → 自动剪辑 → 导出草稿」全程无回归，撤销重做可用。
- **风险/依赖**：与图片止血线不同文件，不阻塞；工作量大，排 6-5 之后。

### 6-7 空态/加载态/表格外壳各 2-4 套 `[小手术]`
- **改什么**：只统一空态、加载态、错误态三个基础组件（`V88EmptyState` 可作基线），替换各页面自写实现；表格外壳只统一 loading / 空数据 / 分页外壳，不做视觉改版。
- **涉及文件**：`frontend/src/shared/components/V88EmptyState.jsx`、新增 LoadingState / ErrorState；命中较多页：`NovelFetchPage.jsx`（14）、`user/components/AgentWorkspaceDrawer.jsx`（10）、`DraftToolWorkbench.jsx`（10）、`AICutPage.jsx`（9）、`AssetReferenceWorkbench.jsx`（6）、`ComicScreenplayPage.jsx`（5）。
- **为什么改**：同一语义 2-4 套实现，空态文案与留白不一致，个别自写空态在深色主题下是白块。
- **验收标准**：① `grep -rn "<Empty\|<Skeleton" frontend/src --include=*.jsx` 只出现在共享组件与白名单；② 抽 10 个列表页断网 / 空数据各看一次，文案、留白、深色底色一致；③ 截图 diff 只允许空态区域变化，布局尺寸不动。
- **风险/依赖**：纯展示层，可与 6-1/6-4 并行，但同文件按分工避让。

### 6-8 漫剧三条产品线 `[待产品拍板]`
- **改什么**：先定主产品线再谈合并。三个入口：`/comic-screenplay`（`ComicScreenplayPage.jsx`）、私人定制（`PrivateComicWorkspace.jsx`）、批量工厂 `ComicMediaPanel`（`BatchFactoryNovelList.jsx` 内 1883 行起）。后端两套模型：`backend/internal/screenplay/*`（20 个文件，自有 model / state / store / mysql）与 `batchfactoryv11` 的 ComicPage（`comic_pages_mysql.go` / `comic_pages_memory.go`，状态 `prompt_ready` / `image_ready` / `generation_failed`），另有 `httpapi/private_customization_comic.go`（129 行）做桥接。
- **为什么改**：三处都能「做漫剧」，数据模型两套，用户资产不互通，同一需求要改三遍。
- **验收标准**（本批只到决策）：产出一页决策结论——主产品线是谁、另外两条定位（保留 / 只读 / 下线）、资产迁移方向；经用户确认后才拆实现条目。未确认前本条保持「待产品拍板」，**不改代码**。
- **风险/依赖**：依赖 D-7；不决策就动任一条都会造成二次迁移。

---

## 8. 一批的验收流程

1. **构建**：`cd frontend && npm run build` → 重建 `backend/qiantie-local.exe`（按项目现有流程，不在本文件展开）。
2. **重启**：停计划任务 → 确认进程全退 → 确认 4001 端口无监听 → 启动 → 核对服务端 assets 与本地 `dist` 的 SHA256 一致 → `/` 返回 200。
3. **逐条核对**：按《[批次一验收清单](v88-batch1-acceptance-2026-11-11.md)》逐条执行，记录 PASS / FAIL / BLOCKED。
4. **不通过就地修完**：同一批次内修完重测，不跨批合并。
5. **收尾**：本批全部 PASS 后，更新本文件的「状态」列，再进下一批。
6. **红线**：全程不 commit、不部署、不消耗模型额度；涉及写库的条目等用户单独授权。

---

## 9. 待用户拍板

| 编号 | 事项 | 影响 | 需要什么 |
|---|---|---|---|
| D-1 | 线上发布时机 | 批次一包含安全三条，越晚发布风险越大 | 指定发布窗口 |
| D-2 | 图片并发取值调整 | 直接影响积压恢复速度与上游限流风险 | 给定并发上限（当前默认值需确认） |
| D-3 | 存量数据清理授权 | 3-1 / 3-2 / 3-3 都是写库操作 | 授权范围 + 时间窗 + 是否先备份 |
| D-4 | 漫剧部分完成是否独立成状态 | 影响状态机新增与前端展示（对应设计债「状态散落」） | 是否新增 `partial` 类状态 |
| D-5 | 批次二 2-3 出站安全客户端是否一次全量切换 | 一次全量切换最快但回归面最大 | 全量 or 按渠道灰度 |
| D-6 | 批次三 3-4 错误列加长是否走迁移 | 加长列在线上大表上有锁表风险 | 是否允许在线 DDL |
| D-7 | 漫剧主产品线（6-8） | 决定两套后端模型是否合并、三条前端入口去留 | 指定主入口 + 另外两条定位 |
| D-8 | 批次六开工窗口 | 6-1 / 6-4 与图片止血线改同一批 `user/pages` 文件 | 图片止血线停手后给一个独占窗口 |
