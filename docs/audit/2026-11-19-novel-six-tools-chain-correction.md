# 六工具对齐纠错：REA证据、现场结果与未完成项

2026-11-19。范围为用户截图左下六工具；本轮重新分析和实查，没有修改产品代码，也没有启动调度、保存新任务或调用付费模型。4001本轮观察版本11aab727，按钮贴底修复不等于功能复刻完成。

## REA分析在哪里

原版实际文件：F:/软件/洛水2026/洛水2026/dist/assets/index-C1aA08SG.js。
本轮重新计算SHA256：5278a6d1b525152a6e9542c9d245bd90c658b84db3d94be9198ee2ac98306d4a，和留存副本一致。

恢复目录：F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/。
该目录的manifest.json列出原包UTF16/UTF8范围、原始压缩行号与各恢复文件hash。本轮六份重点恢复文件hash与manifest完全相符。它们是为阅读提取、格式化的前端片段，不是完整的原软件开发源码，也不是已经恢复的原生后端算法。

本轮REA调用记录：_local-archive-20261028/novel-tools-audit-20261119/rea-results.json。
- analyze_javascript_application直接读取实际原包，1GiB限制下解析耗尽内存，parsed_javascript_files=0，不能计作整包分析成功。证据ev_f923fa93e22df3294600be72b2d70c3753a9e5db2ba1b26bb1a4234c4460d707。
- 转为定向读取已核对来源的NovelSchedulerPanel、NovelStudioSidebar、NovelStudioPage三份片段；成功解析3文件、19051语法节点、16条静态发现。证据ev_2359d343a1e39d14330553dabf2ea928175738f66ae7a6f089e5dfbb9f883769。其他26文件未分析、图有数量上限、动态绑定仍未知；这不是完整后端调用链。
- Craft Anything本轮实际调用了investigation/session、brief/plan、target_open、finding_record等；原Scheduler函数分析记录ev_8b6769ee50c7635593bfc86c8ae61d61d0555f9d77d9399ce76bba37c18a8c0c、原Agent记录ev_e7f456c2f7b104d46a9413412cf5ea0386bae98d6da64a018e203421f9486311可与恢复文件hash对应。

## 现场检查边界

原版窗口6882506、洛水v2.9.4、页面127.0.0.1:47496。本轮实际打开Agent设置和定时写作，原版作品数0。CUA坐标输入报告geometry unavailable，之后使用用户先前明确授权的Windows无障碍Invoke接口打开已观察到的按钮；没有点击启动或保存。
CM在无选中作品时依次点开写作风格、Agent设置、数据看板、定时写作，四者均显示“请先选择作品后使用此工具”。随后选择已有验收作品，实际逐个打开六工具，记录实际控件与错误提示；未创建任务、未保存修改。
AX、截图、REA与hash证据均在_local-archive-20261028/novel-tools-audit-20261119/。

## 已确认的具体问题

|工具|原版证据|CM实际接入|判定|
|写作风格|NovelStylePanel.readable.js:5、49、231：独立风格编辑、参考原文与可自动分析的指南；StudioPage:1604不要求bookId|NovelCreationPage.jsx:341把styleMode传给MaterialsPanel；MaterialsPanel.jsx:7、15、52同时显示文档/上传/文风编号，且无书拦截|有实现，但复用了素材页，入口范围和编辑交互不一致|
|导入素材|StudioPage.readable.js:1650：专用导入弹窗及引用；Sidebar:377独立open-import-modal|NovelCreationPage.jsx:318、386；MaterialsPanel真实文档接口，当前4001明确显示未配置TOS并禁用上传，只能粘贴|有接口和错误反馈；上传能力未具备，不能登记全部可用，原选章/范围仍需逐项对照|
|Agent设置|NovelAgentSettings.readable.js:6、349—354：不要求bookId；内容确认暂停；37维；原现场显示审稿维度37项|AgentPanel.jsx:35仅8条路由；:47把同名requireSecondaryConfirm显示为每次后续模型调用前确认；:61保存当前作品；pipeline.go:651把它传入请求前确认|作用范围、关键开关含义和选项不一致，不只是UI差异|
|定时写作|NovelSchedulerPanel.readable.js:21默认30/10/3，:66启动，:90—97只提交bookId/sessionId/model/intervalSeconds/dailyChapterCap/maxConsecutiveFailures，:158后为调度器状态页|AutomationPanel.jsx的save提交通用ScheduleInput、模型、写作指令、引用等；NovelCreationPage:341无书拦截|通用任务系统真实存在，但接入了不同表单和规则，不能称原版调度器复刻|
|数据看板|NovelDashboard.readable.js:15—30按章节算平均与状态；:40—61有状态分布及字数趋势；StudioPage:1638后还有伏笔区|NovelCreationPage.jsx:341仅章数、总字数、任务数、成功率四个数|统计代码存在；主要维度、趋势和伏笔没有对应，页面不一致|
|备忘录|StudioPage.readable.js:1711以novel_会话或default作为session-key，并有show-all；独立浮层|NotesPanel.jsx:9、13、14调用当前作品/notes，当前弹层有标题内容和增改删按钮|保存API存在，但原会话/default/全部范围与现作品范围不一致；全量保存链未在本轮实测|

以上的行号来自恢复文件或当前工作区，不冒充原开发源码行号；原压缩包多数内容在第41行，精确范围见manifest。

## 为什么越写越错

先前将已经存在的通用素材、任务、作品设置页挂到相似名称下，再把界面局部修正和自身测试通过当成推进复刻。这样测试能证明CM内部自洽，仍然无法证明原版入口、参数、设置作用范围和操作顺序相同。最近的短篇专用流程补了真实存储与检查点，但新增暂停框没有修正原Agent同名开关的错误含义。

不能笼统说六按钮全是空按钮；也不能因为有HTTP处理函数就宣称都已复刻、都可运行。当前正确判定是：有实际实现，存在错接、能力配置缺失和语义不同。

## 下一步必须按此顺序修复

1. 先拆开“内容生成后的暂停确认”和“模型调用前确认”两种独立规则，核查已有保存值和进行中任务。保留既有防止未知结果重复计费的保护；不要把旧值直接改义后让任务意外自动生成。内容确认应读已编辑的新版本。
2. 明确账户设置、会话设置、作品设置的范围，原版可打开的工具不应被统一挡成先建书；允许打开不代表允许无书启动，本轮未测试启动边界。
3. 定时页先恢复原调度器的三参数/状态/启动暂停恢复停止/累计与失败反馈，再适配我们现有MySQL任务系统。不能只给通用表单换标题。
4. 独立风格编辑器、37维真实开关与阶段调用、看板状态趋势及伏笔、备忘录作用范围分别有接口输入/输出/失败规则，再逐个实现。
5. 每个控件必须验：点击入口→实际字段→请求参数→保存数据→刷新读回→正常/空/重复/错误/取消。不跑的付费操作明确未验证，不能拿页面截图当成生成成功。

本轮仅纠正证据与下一步修复合同，不声称已修好上述业务差异。
