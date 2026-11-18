# 内容确认与定时调度器纠错交付

2026-11-19。本轮实施已批准的六工具纠错报告中的前两条错链，不代表整个小说创作已复刻完成。产品提交a56b9a2a、d9c8b335。实际4001运行d9c8b335，入口/assets/user-C-PYiuT7.js，二进制_local-archive-20261028/novel-confirmation-scheduler-delivery-20261119/qiantie-confirmation-scheduler-final.exe。没有公开部署，没有真实模型调用。

## 原版依据及实现边界

原恢复文件位于_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/；manifest记录原包hash与UTF16/UTF8位置。NovelAgentSettings.readable.js:349—354说明内容生成后暂停；NovelSchedulerPanel.readable.js:21、66、90—97、158—217说明30/10/3三参数、状态、启停/恢复。上一轮已在原窗口6882506实查相同界面。本轮沿用该证据，不重新设计原合同，不声称恢复原生私有调度算法或提示词。

## 内容暂停与模型调用确认

- AgentSettings新增可空pauseAfterBlueprint。true表示新蓝图完成后暂停，false表示继续正文；缺字段沿用旧模式的暂停规则。原requireSecondaryConfirm继续保留历史逐次模型调用前确认保护，避免把旧值直接改义后意外发送请求。
- service.go创建新任务时将内容checkpoint写入不可变快照；旧任务不重写。前端两个独立控件分别保存，用户只改其他设置时不为缺省字段偷偷写入新值。继续蓝图沿用既有确认/编辑版本链。
- 两个开关的运行测试通过：仅内容暂停实际调用一次蓝图模型后等待；仅调用确认实际调用一次后在下一调用前等待；都关闭可用6个本地假模型响应贯通并保存正文。
- 兼容边界：旧记录缺字段时保留原先规则，因此不冒充原版全新安装默认值。Agent仍按作品保存，账户全局默认尚未接通；专用短仿写仍使用它已有的独立大纲检查点，不冒充全局设置继承已完成。

## 调度器实际链路

入口：NovelCreationPage的定时写作工具 → SchedulerPanel（默认只显示三参数和状态，可无选书打开）。
启动：POST /api/novel-creation/projects/{projectID}/scheduler/start → Service.StartScheduler → 同一作品事务内写创建回执与Schedule → MySQL定时扫描 → DispatchDueSchedules → Notify/StartWorker → Run → 保存正文 → 保存成功才推进计数。
后台入口在backend/internal/app/novel_startup.go:19、21，实际启动StartWorker与StartScheduler；不是只注册HTTP按钮。

|项目|输入/输出与规则|边界和失败处理|
|---|---|---|
|表单|intervalSeconds、dailyChapterCap、maxConsecutiveFailures；默认1800秒/10/3；会话与模型来自当前上下文|1到1440整分钟、每日1到100、失败1到10；无选书可打开，点击启动明确要求先选长篇，不新建隐藏作品|
|服务适配|已有作品/会话、当前自家文本模型、原会话已存引用；默认续写指令由后端prompt/novelcreation/scheduler.go提供|短篇独立成品流程拒绝逐章调度；引用失效/归档/越权/不匹配明确报错；前端不能注入prompt字段|
|原编号恢复|创建回执和计划在同一事务保存；同编号同参数返回同一计划及其当前状态|同编号不同参数冲突；回包丢失先GET列表按原编号核对；找不到才允许明确原编号重发|
|避免重复生产|专用入口同一作品只允许一个未停止计划；重放回执先于可变校验|已有运行、暂停或失败未停止计划时拒绝再开一个；不静默替换，不清除历史数据|
|暂停/恢复/停止|调用已有真实/schedules/{id}/pause、resume、stop接口|旧逐轮日志、配置修改保留在“历史计划管理”，默认不展开旧创建表单|
|待确认返回|显示waitingReason；“查看创作会话”返回该计划sessionId，父页校验作品关系并恢复对应模式/会话|等待蓝图或模型确认时不只显示运行中；读取失败有重读入口，不冒充已停止或成功|
|计数|到期派发→假模型正文保存→累计/当日各增加1；再扫描不重复生成|新增完整链路测试通过；未实际启动4001付费生产，不证明真实模型质量或原生私有算法等价|

## 验证结果

- 后端novelcreation与prompt包全量PASS；新增内容设置独立/旧快照不改、实际运行阶段分离、专用调度输入与越权、重复活动计划拒绝、到期派发保存计数仅一次等测试PASS。
- HTTP相关小说接口及新专用调度接口PASS；拒绝前端prompt等未知字段；201创建、重放同ID、跨账号404；启动不立即生成。
- 新浏览器7/7 PASS、0跳过：无书打开/无隐藏任务、三参数提交与暂停恢复停止、丢包只核对原编号、两个确认设置独立保存/重开、旧计划管理可达、缺省设置不被静默改写、等待确认原因和回原会话。
- 真实隔离MySQL 8.4再次PASS：独立设置保存后换Store读回、创建回执重放、暂停恢复停止持久化、模型任务数0。复用已确认属于本测试的nonce/schema；运行日志在_local-archive-20261028/novel-parity-20261117/novelparity20261117ac92519bf9fa-*-test.log。
- 现有定向界面回归6项中5过，1项在进入Agent之前因历史断言侧栏260px失败（现有CSS此前已240px）。该测试未修改，不能宣称旧界面回归全部通过；新增隔离Agent测试已独立验证0/false、保存和重开。
- 4001实查：新自动写作调度器标题、已停止、30/10/3、单启动按钮；Agent两个确认控件分别可见并已读取服务设置；无选书点击启动明确提示。实际库保持作品1、计划0、模型任务0，未启动生产、未改用户偏好。
- 构建通过。启动后首次health曾在监听就绪前连接拒绝，随后health与实际包身份检查通过；共享工作区期间另一个本地构建替换过4001，重查进程后才切换到明确候选，最终运行d9c8b335。

证据目录：_local-archive-20261028/novel-confirmation-scheduler-delivery-20261119/，含runtime manifest、构建/浏览器日志、scheduler-live.jpg、agent-live.jpg与AX。受控浏览器每次请求/DOM/截图在_local-archive-20261028/novel-scheduler-browser/；初始失败证据保留，不删除。回包丢失用例初期模拟/api/client-errors误改了假计划状态，修正测试桩并加强真实页面核对按钮后通过；后续检查点导航先失败再补齐。

## 继续未完成的项目

Agent全局设置范围、17分工/37维真实注入、独立文风编辑器、完整状态/趋势/伏笔看板、备忘录原范围、短篇原创建议动作和番外四类型仍待修。本轮仅将A10升级为部分实现，T01继续部分实现，没有登记完全等价项。所有用户和模型配置仍使用CM自身服务。
