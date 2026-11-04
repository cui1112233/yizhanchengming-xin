# 2026-11-04 V88「假成功」专项审计（只读）

> 本轮只读审计：未修改任何源码、未重启 4001、未改 `backend/web/dist`、未 commit。
> 行号以当前磁盘为准；工作区存在大量并行未提交改动，结论仅代表当前读到的版本。

## 一、审计口径与方法

**判定核心**：一次「成功」是否有回读证据。

- 有回读：重新 GET 详情/列表、检查返回体业务字段（`ok`、`status`、`failed_groups`、`CarouselID`）、确认文件/记录/上游回执真实存在。
- 无回读：只要 `await api()` 没抛异常就 `message.success`，或只看 HTTP 2xx。
- 乐观更新（UI 即时反馈 + 已有轮询/回读兜底）计 P2，不按 P0 处理。

**扫描面**：

- 前端 `frontend/src`：307 行 `message.success` / `toast.success` / `notification.success`，逐条结合上下文判断。
- 后端 `backend/internal/httpapi`：165 条 POST/PUT/PATCH/DELETE 写路由注册，重点检查上游降级、`writeStoreError` default、上传/发布/回读。
- 专项核验 6 个已知样本在**前端展示层、后端写入层、状态判定层**是否仍有残留。

## 二、结论

- 共记录 **50 条**：**P0 5 条**、**P1 8 条**、**P2 24 条**、**已修对照 13 条**。
- P0 全部集中在**视频站登录/提交**：前端 `handleWebSubmit` 只要 HTTP 请求 resolve 就弹成功，而后端在部分/全部失败时**仍返回 HTTP 200 + `ok:false`**，两者叠加造成「登录可用/自动提交完成」的假成功。
- 6 个已知样本中，头像、出图 409、Slice=0、N8 发布 `upload_incomplete`、121 单本上传回读均已在代码层修复；仍有 **3 处残留**：视频站前端 `handleWebSubmit`、历史 `uploaded` 被当成完成态、批量上传 121 无 `GetSubmission` 回读。

### 最危险的 3 处

1. **视频站「自动提交完成 / 登录可用」不看业务 `ok`**：`NovelFetchPage.jsx:1359` 的通用包装器 + `web_submit_submit_compat.go:46` 的 200+`ok:false`。用户看到提交成功，实际有失败分组；看到登录可用，实际会话已过期（目标站 302→`login.php` 后返回 200 登录页）。
2. **121 批量上传无回读**：`novel_fetch_upload_batch_compat.go:120-130` 只要 `Submit` 不报错就写 `status:"ok"` 并返回顶层 `ok:true`，没有 `GetSubmission` 确认目标站真的存在该书。
3. **发布/上传完成态被高估**：`batchFactoryBookState.js:126,136` 把 `uploaded` 直接判为「已回读 / completed」，历史未建合集的记录会显示「完成 / 上传成功」；`BatchFactoryNovelList.jsx:1952` 发布成功也不校验 `status/CarouselID`。

## 三、P0 清单

| 编号 | 层 | 文件:行 | 当前判定依据 | 用户会看到的假象 | 建议的回读方式 |
| --- | --- | --- | --- | --- | --- |
| FS-01 | 前端 | `frontend/src/user/pages/NovelFetchPage.jsx:1359-1361` | `handleWebSubmit` 只 `catch` HTTP 异常，请求 resolve 即 `message.success(success)` | 后端返回 200+`ok:false`（部分/全部失败）时仍弹「自动提交完成」 | 读取 `result.ok / failed_groups / success_groups`，失败>0 时 warning 并列出失败分组 |
| FS-02 | 前端 | `NovelFetchPage.jsx:1394` | 「提交全部」走 FS-01 包装器，未检查 `ok` | 「自动提交完成」但多组实际未提交 | 同 FS-01，并要求 `failed_groups === 0` 才 success |
| FS-03 | 前端 | `NovelFetchPage.jsx:1378` | `testWebSubmitVisible` 返回 200+`ok:false` 时包装器仍成功 | 「视频管理系统登录可用」但会话已失效（正是 302→`login.php` 200 登录页场景） | 检查 `result.ok` 与 `checks[].ok`，任一 false 时显示会话失效并阻止后续提交 |
| FS-04 | 后端 | `backend/internal/httpapi/web_submit_submit_compat.go:46` | 成功响应固定 `HTTP 200`，业务结果放 `ok = (failed==0)` | 前端/调用方只看 2xx，误判提交完成 | 部分失败返回 `207/409` 或强制契约：客户端必须读 `ok`；同时写入失败审计记录 |
| FS-05 | 前端 | `NovelFetchPage.jsx:1406,1414` | 重试/核对同样走 FS-01 包装器 | 「版本已重新提交」「核对完成」但实际仍失败/未确认 | 检查 `result.ok / failed_groups`，未确认时提示「已提交，等待回读」 |

## 四、P1 清单

| 编号 | 层 | 文件:行 | 当前判定依据 | 用户会看到的假象 | 建议的回读方式 |
| --- | --- | --- | --- | --- | --- |
| FS-06 | 前端 | `NovelFetchPage.jsx:1417-1426` | 登录后只判断 `session?.status` 用于同步配置，随后无条件成功 | `uploadSession` 若为 `missing/expired`，仍弹「视频管理系统已登录」 | 非 `ready` 时提示未登录并保留登录框 |
| FS-07 | 前端 | `NovelFetchPage.jsx:1465-1469` | `await syncWebSubmitStyles()` 后直接成功，不看 `style_sync.new_count` | 目标站返回 0 条风格仍提示「风格目录已同步」 | 校验 `new_count>0` 或展示 `added/removed` 数量 |
| FS-08 | 后端 | `novel_fetch_upload_batch_compat.go:120-130` | `Submit` 返回即写 `status:"ok"`、顶层 `ok:true`；无 `GetSubmission` 回读，单本失败也只写 `error` | 批量上传显示成功，目标站未确认入库 | 复用 `submissionConfirmedForBook/GetSubmission`，确认后写 `confirmed`，未确认写 `upload_pending_confirm` |
| FS-09 | 前端 | `shuihuo/BatchFactoryNovelList.jsx:3407-3413` | `fetchBookOriginal` → `refreshBatch` → 直接成功，未校验刷新后的 `sourceText` | 「已获取并写入当前小说正文」但正文仍为空 | 刷新后读取 `book.sourceText`，长度≥阈值才成功 |
| FS-10 | 前端 | `shuihuo/batchFactoryBookState.js:126,136` | `publishStatus/websiteSubmitStatus` 含 `uploaded` 即判「已回读/completed」 | 历史 `uploaded`（图传了、合集没建/未回读）显示「完成 / 上传成功」 | 只认 `confirmed/published`；`uploaded` 改为待确认，并与 `upload_incomplete` 对齐 |
| FS-11 | 后端 | `batch_rewrite_compat.go:337` | 环境检查硬编码 `"上传接口": ok:true` | 环境检测面板永远显示上传接口通过 | 实际探测 `/zbooklist_upload.php`，至少返回 `skipped` 并带原因 |
| FS-12 | 后端 | `shuihuo_projects.go:499-501` | `/api/shuihuo-production/health` 硬编码 `database.ready=true`、`storage.ready=true`，无 ping | DB/存储已宕机但生成前检查显示正常，点击后才失败 | 用 `store.Ping` / 文件系统探针，失败返回 `ready:false` + 原因 |
| FS-13 | 前端 | `BatchFactoryNovelList.jsx:1942-1952` | 发布只 `await publishComicCollection`，成功文案不校验 `published.status/CarouselID` | 若后端返回 `uploaded/upload_incomplete`，仍弹「漫画合集已发布」 | 校验 `status==='published' && carouselId>0`，否则 warning「已上传，合集待确认」 |

## 五、P2 清单（乐观更新/缺二次回读）

| 编号 | 层 | 文件:行 | 说明 |
| --- | --- | --- | --- |
| FS-14 | 前端 | `NovelFetchPage.jsx:1364` | 同步组织归属成功后不展示条数；0 个配置档也提示成功（后端空列表已 502，属于展示层弱化） |
| FS-15 | 前端 | `SettingsPage.jsx:262` | `saveConfig` 用返回值收口，未再 GET；「工作台设置/本地存储已保存」依赖服务端返回体 |
| FS-16 | 前端 | `ProfilePage.jsx:80-82` | 资料保存用 `result.member` 收口，无独立二次读取；响应即权威，风险低 |
| FS-17 | 前端 | `NovelFetchPage.jsx:1483-1487` | 默认配置档保存后只改本地 state，未回读 `webSubmitConfig` |
| FS-18 | 前端 | `NovelFetchPage.jsx:1639-1643` | 配置保存用返回的 `savedSettings` 回填，未再 GET |
| FS-19 | 前端 | `NovelFetchPage.jsx:986-990` | 全自动开关用返回 `status` 收口并刷新任务，属于可接受乐观更新 |
| FS-20 | 前端 | `NovelFetchPage.jsx:1170-1171` | 定时任务保存后 `loadSchedules()` 有回读，消息先于回读结果 |
| FS-21 | 前端 | `NovelFetchPage.jsx:1238` | 规则保存成功提示未校验规则是否全部落盘 |
| FS-22 | 前端 | `BatchFactoryNovelList.jsx:2247-2249` | 素材上传后提示「已设为当前分镜主版本」，未校验 `isPrimary` 是否真的回写 |
| FS-23 | 前端 | `BatchFactoryNovelList.jsx:2268-2270` | 成片上传后提示「可在合成成片中查看」，未回读 merge 列表 |
| FS-24 | 前端 | `BatchFactoryNovelList.jsx:419` | 选择资产预设后提示「资产提示词已保存」，未回读资产记录 |
| FS-25 | 前端 | `BatchFactoryNovelList.jsx:1797` | 分页完成提示成功，但页面仍写「等待重新生成」 |
| FS-26 | 前端 | `BatchFactoryNovelList.jsx:3926` | 阶段动作成功文案依赖 run 返回值，刷新异步进行；失败态可能被成功文案短暂覆盖 |
| FS-27 | 前端 | `BatchFactoryNovelList.jsx:4438` | 恢复自动化只信返回的 automation 状态，预检 blocked 时文案仍可能像「已继续」 |
| FS-28 | 前端 | `BatchFactoryNovelList.jsx:3403` | 删除提示「121 内容未受影响」，未回读目标站 |
| FS-29 | 前端 | `ScriptPage.jsx:930` | 剧本配置保存用返回 `saved.scriptDefaults` 回填，无二次 GET |
| FS-30 | 前端 | `PresetLibraryPage.jsx:432` | 发布预设后先弹成功再 `load()`，未等回读确认生效版本 |
| FS-31 | 后端 | `agent_workspace_compat.go:114,121` | 删除任务/画布节点返回 `ok:true`，未回读确认已不存在 |
| FS-32 | 后端 | `script_history.go:225,249` | 删除历史返回 `ok:true`，未回读确认记录已删除 |
| FS-33 | 后端 | `batch_factory_v11_publish_121.go:183-185` | `websiteSubmitStatus` 已为 `uploaded` 时直接 skipped，完全信任历史回读；历史脏数据会被永久跳过 |
| FS-34 | 前端 | `shared/apiErrorCodes.json` | 未镜像 `CAPABILITY_NOT_ENABLED`；Slice 未开启时前端只能展示后端原始错误，无引导 |
| FS-35 | 前端 | `NovelFetchPage.jsx:1381-1383` | 提交预览生成成功，未检查预览里的 blocking 项 |
| FS-36 | 前端 | `AssetsView.jsx:302` | 「智能预设已更新资产并分配到分镜」未校验分配结果数量 |
| FS-37 | 前端 | `ProfilePage.jsx:95` | 「验证邮件已发送」只代表接口受理，无投递回执 |

## 六、6 个已知样本核验

| 样本 | 当前结论 | 证据 | 残留 |
| --- | --- | --- | --- |
| 1. 头像「已更新」但仍是字母占位 | **已修** | `ProfilePage.jsx:111-116` 上传后 `verifyAvatarImage(avatarUrl)` 真实加载图片，失败走错误分支；后端 `member_profile.go` 头像 URL 带版本号 | 其他头像展示面（`UserLayout`/`MemberCenterPage`）依赖 `qiantie:profile-updated` 事件刷新，事件丢失时会显示旧图；属于展示层 P2 |
| 2. 视频站显示「已登录」实际会话失效 | **部分残留** | 后端 `target_http.go` `CheckSession/LooksLikeLoginPage` 已识别 200 登录页；`novel_fetch_upload_compat.go:40-44`、`web_submit_compat.go:66-71` 已接入 | **前端** `handleWebSubmit` 不读 `ok`（FS-03）、登录 toast 无条件（FS-06）、风格同步无回读（FS-07） |
| 3. 出图 409 被当失败/内容审核 | **已修** | `batchFactoryComicMode.js:46-51,96` 将 409 归类 `conflict` 并合并提示+轮询；`BatchFactoryNovelList.jsx:4048` 用 `waitForComicPageResult` 回读主图；`ScriptPage.jsx:1784` 走同一契约 | 无 |
| 4. 发布 `uploaded` 假完成 | **后端已修，展示层残留** | `batch_factory_n8_comic_publish.go:645-653` 部分成功写 `upload_incomplete` 并返回错误（handler 502）；`savedRecord` 只对 `published + CarouselID>0` 短路 | `batchFactoryBookState.js:126,136` 仍把历史 `uploaded` 判完成（FS-10）；发布成功文案不校验 status（FS-13） |
| 5. Slice=0 纯文本 404 被当「作品没了」 | **已修** | `batch_factory_v11_capabilities.go:174-182` 未注册路由返回 `CAPABILITY_NOT_ENABLED + requiredSlice`；`router.go:330` 安装 gate | 前端未识别该 code（FS-34），只能显示原始错误，不再出现「作品没了」 |
| 6. 上传/生成完成只看 200 | **大部分已修** | `BatchFactoryNovelList.jsx:3014` 按 `status === 'confirmed'` 区分「已回读/等待回读」；单本 121 上传 `batch_factory_v11_publish_121.go:189-204,349-360` 有 `GetSubmission`；资产/出图有轮询回读 | 批量上传接口无回读（FS-08）；获取正文无回读（FS-09） |

## 七、同源模式总结（建议纳入代码规范）

1. **HTTP 2xx ≠ 业务成功**：凡是返回体含 `ok/status/failed_groups` 的接口，前端必须读取并分支；不允许「await 没抛异常就 success」。后端批量操作遇部分失败不得用 200+`ok:false` 单独承载，至少要有明确的契约字段与前端强制校验。
2. **上游回执必须二次回读**：目标站/上游返回 200 不代表数据已入库。上传/发布类写操作统一遵循：提交 → `GetSubmission`/详情回读 → 确认存在 → 才写 `confirmed/uploaded`，否则写 `*_pending_confirm`。
3. **状态判定不得只看单一信号**：不允许只看状态码、只看字符串包含、只看本地缓存/内存 state。至少两个独立信号（如 `status` + `CarouselID/imageIDs` + 列表回读）一致才算完成。
4. **展示层不得自行放宽完成态**：`uploaded`、`submitted`、`queued` 是中间态，禁止在 CSS/状态映射层把它们归入 `completed/succeeded`；完成态只允许 `confirmed/published/succeeded`。
5. **环境/健康检查不得硬编码**：所有 `ready:true/ok:true` 必须来自真实探针；无法探测时返回 `skipped + 原因`，不得伪装通过。

## 八、未覆盖范围

- 未逐条执行 165 条写路由的运行时验证，只做静态审计；未启动第二个实例、未调用生产接口。
- 未覆盖 `backend/internal` 下非 `httpapi` 的 service 内部写路径（如存储层、worker）与 `frontend` 构建产物 `backend/web/dist`。
- 未覆盖 WebSocket/SSE 推送状态、定时任务调度器内部状态、Redis 队列消费侧。
- 未做历史数据清洗：`uploaded` 历史记录、`comic_pages` 旧编译提示词均需另行批准后处理。
- 头像相关改动归另一路，本轮仅记录未改代码。

## 九、清单文件

- 本报告：`docs/verification/2026-11-04-fake-success-audit.md`
- 机器可读清单：`docs/verification/2026-11-04-fake-success-audit.csv`
