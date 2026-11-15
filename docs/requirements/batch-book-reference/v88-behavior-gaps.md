# 原系统与批量测书：链路对照和已确认差异

核对日期：2026-11-15。Go 行号为核对时位置，提交后继续修改可能移动。结论只按已读代码和实际测试填写。

**阅读提示：下表是最初调查的历史快照，不是当前完成率。** 定时续做的后续修复见文末；规则与轻量摘要已在 `9a99c60b` 本地验收，见 `docs/acceptance/2026-11-15-batch-book-rules-consistency.md`。本日继续调查新复现6类恢复/派发问题，见 `docs/audit/2026-11-15-batch-book-workflow-reassessment.md`；该轮尚未实施业务修复。

| 功能 | 原系统 | 我们的对应位置 | 结论与后续 |
|---|---|---|---|
| 配置和草稿 | main.py:18429–18476 | backend/internal/httpapi/batch_rewrite_compat.go:44、225；router.go:771 | 存在 bootstrap / work-form 兼容路径；不能因为没有 GET workshop/work-form 就断言功能缺失。需验收刷新恢复和用户隔离 |
| 开始处理 | main.py:18569；2430 | frontend/src/shared/api/novelFetchWorkshop.js:27；backend/internal/httpapi/novel_fetch_workshop.go:2157 | 后台作业入口存在；新入库和已有记录的策略仍需与重试区分 |
| 定时保存 | main.py:15221、15087 | frontend/src/user/pages/NovelFetchPage.jsx:1220；backend/internal/httpapi/novel_fetch_legacy_workflow.go:212 | 原包保存时建立待执行任务，我们只保存卡片，到点才建任务；已确认不同，尚未实现原包此部分 |
| 定时输入校验 | main.py:15221 | novel_fetch_schedule_input.go；legacy_workflow.go:212 | 本轮新增空书单、无有效 ID、错误时间、未知平台校验；测试先失败再通过 |
| 调整时间 | 原包 SchedulePatchRequest 只接收 status（main.py:557），不是时间编辑功能 | frontend/src/user/pages/NovelFetchPage.jsx:1229；legacy_workflow.go:updateSchedule | 我们额外支持时间调整，不能冒充原包同一按钮；本轮补错误时间不覆盖旧值、别名时间生效的测试与修复 |
| 历史损坏时间 | main.py:15320 有到期解析 | novel_fetch_workshop.go:997 | 原实现无效时间直接 continue、卡片一直等待；本轮改为 failed 并显示原因，不派发 |
| 重复书籍 | main.py:14772–14798 按活动 / 终态区别 | novel_fetch_workshop.go:1059、1341 | 我们按是否存在直接跳过，全部存在时标 completed。已确认缺少细分；不能删除判断后直接全量重跑 |
| 输入列解析 | main.py:5149 | frontend/src/user/pages/NovelFetchPage.jsx:1143、1234；novel_fetch_workshop.go:1058、1275 | 定时保存原始 values.bookIds，派发只调用 parseWorkshopDocuments，没有显式传入保存的列顺序。存在与输入预览不一致风险；需用非标准列顺序用例验证，尚不宣称所有格式都失败 |
| 定时重试归并 | main.py:15320 | novel_fetch_schedule_retry_status.go:15 | 已有跨重试作业协调；只在 waiting_retry / retry_tracking 状态启用，其他历史状态还需组合测试 |
| 单书执行和续做 | main.py:16238–16463 | novel_fetch_workshop.go:576 execute；novel_fetch_recovery_snapshot.go:115 | 已有正文和提交指纹恢复逻辑，不需要再造第二个执行器；必须测试缺原文、缺某个 AI、只缺提交等组合 |
| 手动重试 | main.py:11611–11791 | novel_fetch_retry_queue.go:51 | 已有按书籍防重复队列和执行设置恢复；完成条件 / 版本数 / 账号等待需按 E05 核对 |
| 数据与网络 | main.py:6590、8877、14418；本地文件路径函数 | backend/internal/novelfetchworkshop/upstream.go:51；novel_fetch_workshop.go:1838；web_submit_service.go:60 | Go 中有实际上游获取与提交服务，不是全部假按钮；各外部依赖的成功仍须实测。原包可达读写和网络位置已另存 JSON |
| 提交去重 | main.py:12285、15476 | web_submit_service.go:60；web_submit_submit_compat.go:13；novel_fetch_workshop.go:402 | 已有提交服务与后续回查流程。受控链路测试通过不等于真实目标站已成功；未知结果禁止直接重传 |
| 提交日志 | main.py:18814 | batch_rewrite_compat.go:170；novel_fetch_submit_diagnostic.go:26 | 已有诊断适配路径；空白可能来自无记录、字段适配或读取失败，必须以实际任务判断 |
| 摘要与长正文 | main.py:9915、9942 分开；摘要 content_loaded=false | router.go:873–889 将 summary/content 都转完整 document；novel_fetch_workshop.go:2525 hydrate | 已确认接口语义不同：摘要别名也走完整内容加载。属于响应量和性能核对重点，尚未改代码 |
| 规则轨迹 | main.py:18773、18804 | batch_rewrite_compat.go:181；novel_fetch_workshop.go:2815 | 任务轨迹读取存在，独立样本 trace POST 不应因为有任务轨迹接口就标为相同；需要核对请求体和返回结构 |
| 知识库 / 开头词 | main.py:18503–18540 | novel_fetch_workshop.go:2914–3133；router.go:748 | 汇总和分类型入口不同；应检查实际条目更新，而不是强求路径相同 |
| 停止 / 删除 | main.py:18580、18664–18686 | novel_fetch_legacy_workflow.go:198、500；novel_fetch_workshop.go:3198 | 已有停止 / 取消与删除入口；原包彻底删除语义仍需单独核对，不能拿普通 DELETE 充数 |
| 账号管理 | main.py:18355–18424 | 项目统一账号 / 组员体系 | 允许独立技术实现，但切换上下文、管理员模型继承、密钥不可见必须逐条验收 |

## 为什么以前容易“修一个又出现一个”

已读代码至少包含三套入口形式：workshop 主入口、batch-rewrite 兼容入口、旧 /api 路径转发。它们共用部分服务，但请求结构和返回结构不总相同。按路径名或函数名匹配，会漏掉别名，也会把“转发成功”当作“语义一致”。

另外，“书籍有记录”“文案已生成”“整本完成”“版本已提交”“目标站已确认”是不同状态。目前定时去重和摘要转发已经找到这类直接证据。因此后续必须按状态和数据结果核对，而不是继续加一个通用重试或通用同步按钮。

## 当前不能宣布完成的部分

上表保留首次发现时的代码证据。第二轮已修复：保存即建待执行书籍、未完成续做、非标准列/表头定时输入、停止和取消保护，并完成受控上传/回查验收，详见 verification-20261115-schedule-resume.md。

轻量摘要、正文/日志按需加载及规则来源一致性已有本地验收记录；不再列作未实施。仍未全面完成：本轮复现的恢复/派发问题、主动重制的历史实例、独立规则预览轨迹、纯 ID 外部集成、彻底删除及全部动态页面状态。完整入口清单的“未找到”只表示待追踪，不自动等于缺失。
