# V88 变更摘要（2026-11-11）

> 基线：`HEAD b167c4d7` + 当前未提交工作区（`git status` 共 579 个文件有改动，其中前端源码 79 个）。
> 口径：本文只登记能在工作区代码/测试/文档里核对到的改动；行号按当前工作区记录，并行会话继续改动时可能漂移。
> 证据优先级：运行时代码 > 单测 > 验收记录。标注「验收记录」的条目来自 `docs/plan/v88-batch1-acceptance-2026-11-11.md` 的实测结论。

---

## 一、错误文案中文化（三轮）

| # | 改了什么 | 实现要点 | 涉及文件（行号） | 影响面 |
|---|---|---|---|---|
| 1.1 | 前端脚本协作错误分类改为状态码优先 | 401/403 一律判「登录已过期」；422 直接透传后端校验文案；只有拿不到状态码时才走「密钥/模型配置」关键词匹配 | `frontend/src/shared/pet/scriptCollaboration.js:12-37`；`frontend/src/shared/api/client.js:161` | 私人定制/剧本协作所有请求；会话过期不再误报成模型配置问题 |
| 1.2 | 漫画页提示词文案去哨兵 | 失败文案经 `UserFacingErrorText` 清洗后再拼进结果、再写库；裸哨兵回退成中文兜底 | `backend/internal/batchfactoryv11/comic_prompt_service.go:391,507`；`backend/internal/batchfactoryv11/production.go:367-401` | 提示词提取/补齐的接口回包与历史账本 |
| 1.3 | 图片上游错误统一两行格式 | 主文案为中文结论，第二行固定 `\n上游原文：`；原文压缩空白、去开头冒号、截断 240 字；`shuihuoImageUpstreamError` 保留结构化状态码 | `backend/internal/httpapi/shuihuo_projects.go:4588-4627` | 所有图片出图/资产图/Agent 出图链路的前端展示与日志 |
| 1.4 | 文本模型（Director）HTTP 错误拆分 | 400/401/403/404/422 不重试；429/5xx 有界重试；成人内容、审核、429、401/403、上下文超长、模型不存在、5xx 分别给中文结论 | `backend/internal/batchfactoryv11/director_provider.go:23-85,119-234` | Director/剧本/分镜等所有 OpenAI 兼容文本通道 |
| 1.5 | 历史落库旧串展示时自动清洗 | 清洗函数兼容历史记录中「句子中间夹哨兵」「旧英文 HTTP 串」；裸哨兵映射为可执行中文提示 | `backend/internal/batchfactoryv11/production.go:367-401`；`normalizeLegacyDirectorHTTPMessage` 同文件 383 | 存量失败记录、页账本、任务抽屉回读 |
| 1.6 | 前端按分隔符切第一行做摘要 | `splitBatchFactoryError` 以 `\n上游原文：` 为界，分类与摘要只看第一行，原文只进展开区 | `frontend/src/user/pages/shuihuo/batchFactoryErrorParse.js:19-24` | 批量工厂所有错误提示/Toast/详情 |
| 1.7 | 队列忙/等待取消/等待超时三处文案纯中文 | 「图片队列繁忙，等待 N 分钟…」「本轮已取消…」「本轮已超时…」；底层原因只进 `上游原文：` | `backend/internal/httpapi/shuihuo_projects.go:4088-4135` | 图片并发队列的所有等待分支；不再出现英文 duration/超时串 |
| 1.8 | 资产图任务 error/warning 出口清洗 | 任务回包的 `error` 与批量结果的 `warning` 都过 `UserFacingErrorText` | `backend/internal/httpapi/batch_factory_v12_asset_generation.go:203,313` | 资产图生成任务轮询、批量资产结果 |
| 1.9 | 图片 4xx 区分鉴权失败与「模型不支持出图」 | 401/403 归「通道鉴权失败，请检查 Key/出图权限」；其余非 2xx 归「通道拒绝，模型可能不支持出图」 | `backend/internal/httpapi/shuihuo_projects.go:4604-4612`；调用点含 `batch_factory_v12_image.go:147`、`integration_channel_exec.go:335` | 主出图、v12 出图、Agent/自定义渠道共 5 处入口 |
| 1.10 | 配音链路网络错误分类 | 超时 / 网络不可达 / 已取消三类中文结论，原始错误接 `上游原文：`；默认 TTS 与配置音频模型两条分支都覆盖 | `backend/internal/httpapi/tts.go:59-69,131-134,213-216` | 配音页、AI-CUT 配音、Agent 音频任务 |

**影响面小结**：用户可见错误不再以英文哨兵或上游 JSON 作为主文案；原文统一折叠到「展开原始信息」。遗留见第九节第 3 条。

---

## 二、缺陷清单 5 条

| # | 缺陷 | 修法 | 涉及文件（行号） | 影响面 |
|---|---|---|---|---|
| 2.1 | 批量导演分镜部分失败仍报成功 | 新增 `ErrPartialFailure`；已成功的集/批次保留，整体返回失败并携带 `failed_episodes` / `failed_batches` 与集号/批号 | `backend/internal/screenplay/generator.go:26-29`；`storyboard_service.go:43,195-200`；`assets_service.go:39,191-198` | 批量导演、剧本资产提取；调用方不能再把部分失败当全成功 |
| 2.2 | 漫画缺页被覆盖 | 引入 `comicPagePartialError`（`Completed/Total/Failures`）：有成功页时降级为「部分失败」，只补缺页、不整本判死；已计费未取回不自动重发 | `backend/internal/httpapi/batch_factory_comic_pages.go:121-168`；`batch_factory_automation.go:2213-2227,3525-3538` | 漫画生图续跑、页级账本、自动化状态；避免已完成页被重跑覆盖、重复扣费 |
| 2.3 | 后端英文兜底文案 | 统一清洗 + 中文兜底映射（参数无效/上游不可用/版本冲突/能力不可用/不存在/未配置） | `backend/internal/batchfactoryv11/production.go:367-401` | 批量工厂、提示词、图片任务的所有错误出口 |
| 2.4 | 插件多段生成丢已完成段落 | 生成器返回结构化结果：串行/并发都保留已成功段落与 `failedIndex`，由页面写回；取消走 `AbortController` 不写结果 | `frontend/src/user/pages/pluginGeneration.js:48-102`；用例 `pluginGeneration.test.js:71-100` | 插件中心所有多段/分段生成；失败后无需整批重跑 |
| 2.5 | 登录过期误报密钥问题 | 状态码优先：401/403 → 登录过期并引导重新登录；422 → 透传校验；无状态码才进密钥/模型分支 | `frontend/src/shared/pet/scriptCollaboration.js:12-27`；`routeAccess.js:10` | 剧本协作、Agent 对话、需要登录的插件请求 |

---

## 三、安全

| # | 改了什么 | 实现要点 | 涉及文件（行号） | 影响面 |
|---|---|---|---|---|
| 3.1 | web-submit 别名不再自签身份 | `BridgeAuth.SessionOnly()` 清空 HMAC 密钥，只认 Bearer 会话；旧别名 `/api/batch-rewrite/web-submit/*`、`/api/novel-fetch-workshop/web-submit/*` 同样收紧；转发内部跳转前剥离 `X-Qiantie-*` 身份头 | `backend/internal/httpapi/batch_factory_v11_auth.go:163-179`；`web_submit_submit_compat.go:17-19`；`web_submit_preview_compat.go:19`；`batch_rewrite_compat.go:29`；`router.go:658,685,776-781` | 所有 web-submit 读写入口；伪造用户名/owner 不再落库 |
| 3.2 | 用户名统一校验 | `SafeUsername/ValidUsername/UserScopedPath/WithinDir` 收敛到 `safepath`：拒绝 `..`、路径分隔符、盘符、控制字符；拼接后做软链接解析并校验仍在基目录内 | `backend/internal/httpapi/username.go:11-32`；`backend/internal/safepath` | 所有以用户名拼路径的入口（素材、参考图、Agent 工作区等） |
| 3.3 | Agent 工作区 rootPath 收口 | 创建/更新工作区时 `rootPath` 必须解析到用户目录内；文件读写再做软链接逃逸校验；越界返回 400 | `backend/internal/httpapi/agent_chat_routes.go:101-109,161-184,353`；`backend/internal/agentchat/store.go:249-257`；`agentchat/tools.go:157-215` | Agent 工作区创建、文件树、文件读写 |
| 3.4 | 找回密码链接服务端基址化 | 重置链接域名只读配置站点地址，注释明确禁止从请求 `Host` 推导；默认 `https://yzcm.site` | `backend/internal/httpapi/account_recovery_compat.go:352-369`；`backend/internal/config/config.go` | 找回密码/重置链接；防 Host 头伪造骗取 token |
| 3.5 | API Key 加密存储 | AES-256-GCM；密文前缀 `enc:v1:`；无前缀按历史明文读取、写入时自动升级；解密失败保留密文并告警，不覆盖；`reveal` 遇密文直接 409，不回显密文 | `backend/internal/apisecrets/cipher.go:12,30-31,51-186`；`backend/internal/httpapi/config_secrets.go:37-88,221-245`；`api_integrations.go:555-571`；迁移命令 `backend/cmd/api-secret-migration/main.go:45-51` | 个人中心 API 配置、渠道凭证、模型目录凭据读写；迁移命令仅支持 `-apply` 显式执行，本轮**未执行** |
| 3.6 | 统一出站客户端 | 内网/元数据地址黑名单（127/10/172.16/192.168/169.254/::1 等）、拨号层防 DNS rebinding、重定向逐跳校验且限跳数、只允许 http/https；豆包中继按「协议+主机+端口」显式放行 | `backend/internal/safehttp/safehttp.go:6,42-55,74-111,279-327`；`safehttp/client.go:28-110`；已接入 27 个文件、50 处 `safehttp.NewClient*` | 所有服务端出站 HTTP：文本/图片/视频/发布/抓取/执行器 |
| 3.7 | 请求体与 docx 解压上限 | 全局错误结构 `请求体超过上限`；JSON 16MB、剧本 25MB、插件文档 20MB、视频上传单独上限；docx 条目 4096、单文件 64MB、总解压 256MB，并有流式读取兜底 | `backend/internal/httpapi/body_limits.go:10-61`；`docx_safe.go:10-93`；调用点 `script_history.go:199`、`screenplay_routes.go:312`、`plugin_tools.go:151`、`plugin_ai_segment.go:167`、`batch_factory_v11_video_upload.go:30,86` | 上传/导入/批量配置接口；防超大请求与 zip bomb |

**运行时证据**：`security_hardening_test.go:34-205` 覆盖匿名身份头、签名身份头、rootPath 越界、软链接逃逸；`safehttp_test.go:13-208` 覆盖私网/元数据地址、重定向、显式放行。验收记录 1-15/1-16 已在本地 4001 实测返回 200/400，无越权落库与文件内容回显。

---

## 四、图片积压止血

| # | 改了什么 | 实现要点 | 涉及文件（行号） | 影响面 |
|---|---|---|---|---|
| 4.1 | 已计费未取回转人工、不分页续跑 | 该错误固定为 `needs_manual`：`MaxAttempts=1`、不计重试预算、不再自动续跑；只保留手动重试入口 | `backend/internal/httpapi/automation_failure.go:77-88`；`batch_factory_automation.go:2213-2227,530-540`；用例 `automation_failure_test.go:167-169` | 单页/整本计费保护；避免同一张图重复扣费 |
| 4.2 | 旧批量与 Agent 出图共用同一进程级名额 | 先取进程级信号量，再取任务级信号量，最后取阶段/全局名额；没有任务上下文的单页请求回落到进程级上限 | `backend/internal/httpapi/shuihuo_projects.go:477-505,488`；用例 `shuihuo_generated_image_download_test.go:479-513`、`shuihuo_image_semaphore_test.go:108` | 旧批量、Agent 出图、单页手动生成共享同一出图上限 |
| 4.3 | 20 分钟卡死页回收 | `comicPageLeaseReclaimAfter=20min`；启动先扫一次，之后每 5 分钟扫描；回收时写日志并释放租约 | `backend/internal/httpapi/batch_factory_automation.go:229-315`（阈值 233-242） | 进程崩溃/请求丢弃后卡在「生成中」的漫画页 |
| 4.4 | 正文上限与客户端超时 | 排队 5 分钟、响应头 3 分钟、正文停滞 180 秒、正文总上限 8 分钟；下载客户端 120 秒，避免慢图被误杀 | `backend/internal/httpapi/shuihuo_projects.go:4064-4086,4494-4500` | 慢图/大图下载；减少「已计费但被判失败」 |
| 4.5 | 错误分类改按结构化状态码 | 确定性失败优先读 `automationErrorHTTPStatus`：401/402/403 归鉴权/额度，不再匹配英文关键词 | `backend/internal/httpapi/automation_failure.go:90-103` | 自动化重试预算、终止判定、页级账本 |

**验收记录**：1-18~1-21 代码/单测通过；1-21 数值一致性仍为 BLOCKED（需真实压入 10 个出图任务）。

---

## 五、前端扣钱与丢成品

| # | 改了什么 | 实现要点 | 涉及文件（行号） | 影响面 |
|---|---|---|---|---|
| 5.1 | 密钥字段失焦才提交 | Key 输入只登记 dirty，失焦或显式保存时整体提交；串行队列避免交错写入 | `frontend/src/user/pages/ApiConfigPage.jsx:15,196-229`；`api-config-preset-state.js` | API 配置所有渠道 Key；不再落半截 Key |
| 5.2 | 配音重复下单与清空旧音频 | 同卡在飞直接 `return`；失败分支保留上一版 `audioUrl`，不 revoke、不覆盖 | `frontend/src/user/pages/TtsPage.jsx:190-236`；`frontend/src/shared/api/tts.js` | 配音页连点、失败重试；避免重复计费与成品丢失 |
| 5.3 | 小说抓取等待与停止 | 30 分钟等待上限；`failed` 立即退出；用户可「停止等待」，后台任务不中断；异常中断直接提示 | `frontend/src/user/pages/NovelFetchPage.jsx:612,1116,1147-1165,1938` | 小说获取工作流；不再无限转圈 |
| 5.4 | 封面轮询上限 | 终态即停；`POLL_MAX_ATTEMPTS` 约 10 分钟上限；卸载清理定时器 | `frontend/src/user/components/CoverGenerationWorkbench.jsx:30,291` | 封面生成；失败后不再持续打接口 |
| 5.5 | 素材参考跨账号可见 | 参考素材目录/路径全部由会话身份 `identity.Username` 推导，不再接受请求自带 owner；路径经 `UserScopedPath` 校验 | `backend/internal/httpapi/reference_assets.go:45-50,157-168,212-257`；`reference_assets_image.go:42-188` | 参考图上传、读取、生成引用；跨账号 ID 无法越权读取 |
| 5.6 | AI 剪辑重复应用 | 应用接口校验 `expectedRevision`，冲突返回 `ErrRevisionConflict`；前端应用成功后立即置 `applied` 并禁用按钮 | `backend/internal/httpapi/aicut_export.go:204,244,340,364`；`frontend/src/user/components/AICutAutoEditDrawer.jsx:59,97-98,132-133` | AI-CUT 自动剪辑草案；同一草案不会叠加两次 |

**限制**：1-6/1-7 未做逐字符刷新与真实 TTS 实测（避免写配置、消耗额度）；1-10 跨账号仍为 BLOCKED（缺第二个 owner 账号与素材）。

---

## 六、错误与状态

| # | 改了什么 | 涉及文件（行号） | 影响面 |
|---|---|---|---|
| 6.1 | 剧本资产/分镜部分失败返回失败批次与集号 | `backend/internal/screenplay/assets_service.go:39,191-198`；`storyboard_service.go:43,195-200`；`generator.go:26-29` | 资产提取与分镜生成接口；调用方能看到「第 X 批/第 X 集失败，已保存 N 条」 |
| 6.2 | Agent 对话失败不再伪装正常回复 | `backend/internal/agentchat/runner.go:37-49,50-87`；`store.go` | Agent 对话/画布；失败落 `RunError`，按超时/未配置/鉴权/限流/上游分类 |
| 6.3 | 五处静默空态改为显式失败态 | `frontend/src/admin/pages/ErrorLogPage.jsx:41`；`frontend/src/user/pages/aicut/ExportQueuePanel.jsx:24`（及 `AICutPage.jsx:920-921,2840-2841`）；`frontend/src/user/pages/ApiConfigPage.jsx:34,295`；`frontend/src/user/pages/UsageStatsPage.jsx:38-50,111-135`；`frontend/src/user/pages/shuihuo/TaskDrawer.jsx:92` | 错误日志、导出队列、模型额度、团队用量、任务列表；读取失败不再显示成「没有数据」 |
| 6.4 | 任务抽屉失败信息可展开 | `frontend/src/user/pages/shuihuo/TaskDrawer.jsx:97`；`frontend/src/user/pages/shuihuo/BatchFactoryErrorNotice.jsx:45-105` | 任务抽屉；状态列与错误分离，原文折叠在「展开原始信息」 |

---

## 七、主题第一步

| # | 改了什么 | 涉及文件（行号） | 影响面 |
|---|---|---|---|
| 7.1 | antd 接入深浅色算法 | `algorithm: isDark ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm`，并补齐文字/背景/边框/浮层/链接/错误等 token | `frontend/src/shared/styles/theme.js:9-40` | 表格、Modal、Drawer、下拉、按钮跟随主题 |
| 7.2 | 补浮层与控件 token | 浮层背景、次级表面、禁用底色、spotlight 等按深浅色取值 | `frontend/src/shared/styles/theme.js:12-40`；`theme-catalog.js` | antd 浮层类组件；减少白块/黑底黑字 |
| 7.3 | 滚动条/斑马纹/焦点描边走主题变量 | 深浅两套 `--scrollbar-thumb`、`--table-row-stripe`、`--focus-ring` | `frontend/src/shared/styles/global.css:41-45,98-126,4212-4243`（验收记录 1-13/1-14）；契约测试 `theme-css-contract.test.js` | 长列表滚动、表格斑马纹、键盘焦点环 |

**验收记录**：1-12~1-14 PASS；已实际切换深/浅色并截图，验证后恢复深色。

---

## 八、测试与文档

### 8.1 后端测试

| 类型 | 内容 | 位置 |
|---|---|---|
| 新增安全用例 | 匿名/签名身份头拒绝、rootPath 越界、软链接逃逸、用户名路径逃逸、登录与重置限流、重置链接不用请求 Host | `backend/internal/httpapi/security_hardening_test.go:21-291` |
| 新增/更新出图用例 | 进程级并发、卡死释放、4xx 不重试、结构化分类、账单读失败不重试、部分失败可续跑 | `shuihuo_image_attempt_guard_test.go`、`shuihuo_image_semaphore_test.go`、`shuihuo_generated_image_download_test.go:479-513`、`automation_failure_test.go`、`batch_factory_comic_generation_concurrency_test.go:213-216` |
| 新增/更新文案用例 | 资产图哨兵剥离、漫画页失败文案、TTS 传输错误、Director HTTP 错误 | `batch_factory_v12_asset_generation_test.go:38`、`batch_factory_comic_pages_test.go:51`、`tts_test.go`、`director_provider_test.go` |
| 部分失败用例 | 分镜/资产部分失败返回失败集/批号 | `screenplay/storyboard_service_test.go:190-201`、`screenplay/assets_service_test.go:163-172` |
| 并发策略用例改名 | `TestGPTImageConcurrencyIsPerTaskNotProcessWide` → `TestGPTImageConcurrencyIsProcessWideAcrossTasks`，断言「进程级 + 按图片域名共享、默认 2 路」 | `shuihuo_generated_image_download_test.go:479-513`；验收记录 8.1 |
| 已知长期失败 | `TestModelCatalogH3ConfigRequiresKeyAndSavesWorkflowTogether` 仍红，归批次三 3-5，与本轮无关 | 验收记录 8.1 |

### 8.2 前端测试

| 内容 | 位置 |
|---|---|
| 错误切分/首行摘要/原文进展开区 | `shuihuo/batchFactoryErrorParse.test.js`、`batchFactoryVisibleError.test.js` |
| 插件多段串行/并发部分失败保留已完成段落 | `pluginGeneration.test.js:71-100` |
| 登录过期/422 透传分类 | `shared/pet/scriptCollaboration.test.js:25-26` |
| Key 失焦提交、额度空态、任务抽屉空态、配音空态、小说空态 | `ApiConfigPage.source.test.js`、`api-config-preset-state.test.js`、`TtsPage.layout.test.js`、`novel-fetch-v88.layout.test.js`、`BatchFactoryAutomationBusyFeedback.source.test.js` |
| 主题契约/深浅色 token | `theme-css-contract.test.js`、`theme-catalog.test.js`、`theme-preference.test.js` |
| 错误提示统一与空态断言 | `shuihuo/BatchFactoryUiFixes.source.test.js`、`BatchFactoryPromptMode.source.test.js`、`BatchFactoryNovelList.source.test.js` |

### 8.3 文档

| 文档 | 内容 |
|---|---|
| `docs/plan/v88-remediation-backlog-2026-11-11.md` | 六批 53 条任务表、依赖、验收方式、8 条待拍板 |
| `docs/plan/v88-batch1-acceptance-2026-11-11.md` | 批次一 25 条验收清单与实测结论（22 PASS / 3 BLOCKED） |
| `docs/audit/2026-11-11-dirty-data-inventory.md` | 存量脏数据只读盘点 |
| `docs/v88-startup-and-config.md` | 启动、通道、密钥与错误码说明更新 |
| `docs/releases/20261111-18-b167c4d7/`、`20261111-19-b167c4d7/` | 构建产物、diff、未跟踪文件快照 |

---

## 九、未完成项

| # | 项 | 现状 | 阻塞/条件 |
|---|---|---|---|
| 9.1 | 分页失败错误落库 | 未做：漫画页失败仍只落 `status=generation_failed`，不落结构化错误 | 需给错误列加长/加列，属在线 DDL，待批准（任务表 4-4、D-6） |
| 9.2 | 剩余 9 处错误未收口 | 未做：仍在各 handler 内各自拼错误串，未过统一 `toClientError` | 批次四 4-1/4-2 未开始；需先落后端统一错误结构 |
| 9.3 | 5 处英文 500 文案 | 仍为英文：`history unavailable`、`save failed` ×3、`delete failed` | `backend/internal/httpapi/script_history.go:124,177,182,217,249` |
| 9.4 | 主题与按钮统一 | 只完成第一步（antd 算法 + 滚动条/斑马纹/焦点）；111 条 `body.*` 浮层规则、逐页硬编码色值、统一按钮组件、动词统一均未做 | 批次五 5-1~5-4 未开始 |
| 9.5 | 设计债 6 项 | 未开始 6 项（以任务表批次六清单为准；本轮只完成空态与错误出口的局部改动） | 6-1/6-4 需等图片止血线停手后的独占窗口 |

---

## 十、待用户拍板项

| # | 事项 | 需要什么 | 关联 |
|---|---|---|---|
| 10.1 | 线上发布时机 | 指定发布窗口；批次一含安全三条，越晚风险越大 | D-1 |
| 10.2 | 图片并发取值 | 给定并发上限（当前默认进程级 2 路/按域名共享） | D-2 |
| 10.3 | 存量数据处置 | 授权范围 + 时间窗 + 是否先备份（卡死 6 页、28 页无主图、91 张资产图） | D-3 |
| 10.4 | 漫剧「部分完成」是否独立状态 | 是否新增 `partial` 类状态（影响状态机与前端展示） | D-4 |
| 10.5 | `reveal` 是否改为掩码返回 | 当前仅拦截密文回显，明文仍整段返回；是否统一改为掩码/二次确认 | `api_integrations.go:534-571` |
| 10.6 | 模型目录 `credential` 明文 | 模型目录凭据是否与 API 配置同口径加密迁移；涉及迁移与回滚方案 | `target_credential_store.go`、`cmd/api-secret-migration` |

---

## 十一、本轮边界

- 未 `git commit`、未部署、未推送到线上。
- 本地 4001 已用当轮新构建重启（验收记录：2026-11-11 20:11:52，PID 33616，入口 `assets/user-C2r0yJz_.js`）；此后工作区仍有并行改动，运行实例不包含这些后续改动。
- 未动 4002/4011，未删数据，未改其它账号配置。
- 未跑任何消耗外部额度的真实生成（图片/文本/视频/TTS 均未真实调用上游）。
- 安全密钥迁移命令（`backend/cmd/api-secret-migration`）本轮未执行，库内历史明文是否已升级取决于线上是否单独执行。

---

## 十二、风险提示

1. **当前本地构建是混合产物**：另一个对话框在同一份工作区并行改动（`git status` 579 个文件），本轮构建/重启只覆盖到重启时刻的代码；之后新增改动未进入该二进制。验收结论不能直接外推到当前工作区最新状态。
2. **行号会漂移**：并行会话仍在改同一批 `httpapi`/`pages` 文件，本文行号以撰写时刻为准。
3. **BLOCKED 三条未闭环**：1-10（跨账号素材）、1-21（积压数值）、1-25（真实失败页展示）只有代码/接口证据，缺运行时验证。
4. **存量数据未处理**：卡死页、无主图页、孤儿资产仍按只读盘点记录存在，任何清理都需单独授权。
