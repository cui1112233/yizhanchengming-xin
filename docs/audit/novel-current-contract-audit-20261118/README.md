# 小说创作逐项复核（2026-11-18）

结论：没有全量对齐。此次复核原版恢复源码与当前代码，更新旧审计75项，另做免费页面检查；没有修改产品，没有付费生成。

原包：F:/软件/洛水2026/洛水2026/dist/assets/index-C1aA08SG.js；SHA256 5278a6d1b525152a6e9542c9d245bd90c658b84db3d94be9198ee2ac98306d4a。
恢复源码：_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current。
现场4001：c4054404，/assets/user-LoEy2Yrp.js；仓库文档提交8ab7e06f。

## 本次结果

部分实现 23项、合同不同 12项、缺失 25项、已有实现待对照 10项、尚未验证 5项。没有一项被无证据地改成“完全一致”。

旧审计仍保留。已纠正长篇原创、长篇仿写、双列图标和部分上下文设置的过时结论；修过一个切片不代表整项闭环。

## 确认的问题

1. 仿写短篇没有独立13000字表单；点击仅切模式，提交仍走通用章节。
2. 短篇结果为ChapterVersion，缺独立storyId、大纲/成品文件/推荐文案的保存及重开。
3. 番外仍映射fanfiction；正传书籍关联及续篇/前传/平行/IF分支没有对应参数。
4. Agent只有8条可配置模型路径，缺原17分工及37维开关；补充提示词不等于内置维度。
5. 内容检查点仍不等价：逐次模型调用确认与大纲文件编辑后继续是两件事。
6. 新建作品/Agent表单允许10000字，通用后端只收800—8000；8001—10000存在真实边界冲突。独立长篇仿写接口另有范围，不混为一谈。
7. 通用文风管理、资料文件、消息操作、章节审修快捷动作、看板趋势与伏笔仍有缺口。

## 核验边界

已实际打开当前六创作入口和六工具、关闭/返回部分弹层，记录在ui-observations.json。没有逐个点击收费、删除和启动自动任务按钮。
原版已打开窗口的无障碍树可读，显示长篇仿写表单；本轮原版点击返回geometry unavailable，窗口截图超时，Raise/Escape未退出弹窗。原版其余状态以保留恢复源码及用户截图为证，不能称本轮逐按钮实测完成。
已运行7个顶层免费Go测试，含4个模式子例，全部通过；仅证明当前校验/持久确认规则，不证明与原版一致。详见verification.md。

## 逐项记录

|编号|功能|当前判定|当前结论|原版位置|当前代码位置|
|---|---|---|---|---|---|
|E01|长篇原创首次发送与建书确认|部分实现|长篇原创已能无书讨论、返回create_book建议并确认/取消；仅此动作接通，原版其他建议动作及完整书籍资料合同仍缺。|NovelStudioPage.readable.js:420|frontend/src/user/pages/novel-creation/StudioChat.jsx:66|
|E02|短篇原创专用产物流程|合同不同|原版短篇大纲/正文/审核/打包与独立storyId；CM短篇结果是一个ChapterVersion。|NovelConfirmCard.readable.js:37|backend/internal/novelcreation/exports.go:101|
|E03|仿写长篇专用导入窗|部分实现|已存在长篇仿写专用表单，默认5000字/100章、文件选章、多文风及备忘录；不再算缺失。原生私有提示词、真实生成及同状态全部交互未验证。|NovelImportPage.readable.js:130|frontend/src/user/pages/novel-creation/LongImitationDialog.jsx:19|
|E04|长仿写参数与参考规则|部分实现|长篇仿写有独立强类型输入和不可变快照；通用白名单不代表长篇专用接口。其余模式不能沿用这项交付，原生拆解算法仍未知。|NovelImportPage.readable.js:375|backend/internal/novelcreation/long_imitation.go:26|
|E05|仿写短篇专用弹窗|缺失|专用弹窗缺失；不能用Agent设置页或通用参考框代替。|NovelImitationShortDialog.readable.js:181|frontend/src/user/pages/NovelCreationPage.jsx:304|
|E06|短仿写总字数和篇章数|合同不同|原版总字数13000默认、篇章1；CM默认每章2000、章字数最多8000，两者单位与范围不同。|NovelImitationShortDialog.readable.js:153|backend/internal/novelcreation/validation.go:10|
|E07|番外正传关联与四类型|缺失|原版有parentBookId和sequel/prequel/parallel/if分支；CM番外标签实际用fanfiction和一段existingSetting。|NovelImportPage.readable.js:437|frontend/src/user/pages/NovelCreationPage.jsx:326|
|E08|同人与番外语义区分|合同不同|原导入器有fanfic和spinoff两个分支；CM把番外入口映射为fanfiction，模式合同仍叫同人。|NovelImportPage.readable.js:116|frontend/src/user/pages/novel-creation/contracts.js:3|
|E09|续写导入正式章节|部分实现|已存在专用续写导入、目标书/新书、分章、多文风选择和正式保存；不能继续列为完全缺失。原版文件选择和形态仍未同状态验收。|NovelImportPage.readable.js:461|frontend/src/user/pages/novel-creation/ContinuationImportDialog.jsx:36|
|E10|无书会话与正式作品的关系|部分实现|无书讨论仅long_original走StudioChat；短篇/番外仍先要求project，六模式关系没有统一对齐。|NovelStudioSidebar.readable.js:41|frontend/src/user/pages/NovelCreationPage.jsx:83|
|L01|六入口排列与模式图标|部分实现|六入口已改两列带lucide模式图标；这修复了旧单列问题，但未逐个证明与原图标/几何一致。|NovelStudioSidebar.readable.js:250|frontend/src/user/pages/NovelCreationPage.jsx:27|
|L02|作品类型筛选|缺失|原版有全部/类型筛选并保留选择；CM只区分归档与未归档。|NovelStudioSidebar.readable.js:50|frontend/src/user/pages/NovelCreationPage.jsx:297|
|L03|每本书展开嵌套会话|合同不同|原版作品展开显示其会话并能在书下新建；CM扁平作品列表+当前模式会话。|NovelStudioSidebar.readable.js:41|frontend/src/user/pages/NovelCreationPage.jsx:307|
|L04|会话相对时间与运行标记|部分实现|CM有当前运行标记，但无原列表相对更新时间样式与完整组别。|NovelStudioSidebar.readable.js:351|frontend/src/user/pages/NovelCreationPage.jsx:308|
|L05|作品与会话重命名|已有实现待对照|当前有真实重命名入口；交互位置/弹窗及空值边界尚未与原版同状态比较。|NovelStudioSidebar.readable.js:41|frontend/src/user/pages/NovelCreationPage.jsx:308|
|L06|作品快捷导出|部分实现|CM有正文导出，但没有原侧栏同位置动作和全部产物合同。|NovelStudioSidebar.readable.js:41|frontend/src/user/pages/novel-creation/ChaptersPanel.jsx:28|
|L07|删除与可恢复归档|合同不同|CM采用可恢复归档，原版有删除动作。不得擅自永久删除来追求界面一致。|NovelStudioSidebar.readable.js:41|frontend/src/user/pages/NovelCreationPage.jsx:306|
|L08|六工具进入位置与返回|已有实现待对照|已修成左下工具、主区工具页、导入弹窗与独立备忘录弹层；旧报告统一Drawer结论过时。|NovelStudioSidebar.readable.js:41|frontend/src/user/pages/NovelCreationPage.jsx:289|
|C01|聊天发送与快捷键|已有实现待对照|CM有真实发送/Enter/Shift+Enter和中文输入保护；不代表原版会话/动作解析相同。|NovelStudioChat.readable.js:248|frontend/src/user/pages/NovelCreationPage.jsx:324|
|C02|可调上下文长度|部分实现|长篇无书讨论可调4000至64000；正式作品通用会话仍固定120000，未提供同样入口。|NovelStudioChat.readable.js:253|frontend/src/user/pages/novel-creation/ComposerToolbar.jsx:14|
|C03|清除上下文与Ctrl+K|缺失|当前聊天区没有原清上下文按钮/动作；不能以新建会话代替。|NovelStudioChat.readable.js:48|frontend/src/user/pages/NovelCreationPage.jsx:321|
|C04|回复复制|缺失|当前消息只有角色和文本，无复制控件。|NovelStudioChat.readable.js:286|frontend/src/user/pages/NovelCreationPage.jsx:317|
|C05|消息存档|缺失|原消息可存档；当前只有作品/会话归档，不是同一种动作。|NovelStudioChat.readable.js:48|frontend/src/user/pages/NovelCreationPage.jsx:317|
|C06|单条消息及配对回复删除|缺失|原区分用户消息及对应回复/单AI回复；CM消息没有删除入口。|NovelStudioChat.readable.js:48|frontend/src/user/pages/NovelCreationPage.jsx:317|
|C07|流式正文展示|缺失|当前只渲染snapshot.messages及RunCard，没有原delta正文流显示合同。|NovelStudioPage.readable.js:926|frontend/src/user/pages/NovelCreationPage.jsx:317|
|C08|思考块展开与耗时|缺失|参数详情不是原思考过程块，当前没有该折叠/耗时显示。|NovelThinkingBlock.readable.js:49|frontend/src/user/pages/novel-creation/RunCard.jsx:22|
|C09|模型与思考强度选择|部分实现|使用自己的模型目录和提示词式思考要求，符合自家配置要求；原模型分组/能力控制与真实服务能力映射尚未逐项证实。|NovelStudioChat.readable.js:231|frontend/src/user/pages/novel-creation/ComposerToolbar.jsx:9|
|C10|建议动作确认卡|部分实现|StudioChat已有create_book确认/取消卡；原write_next/audit/revise/short_run等建议动作没有整套接入。|NovelConfirmCard.readable.js:33|frontend/src/user/pages/novel-creation/StudioChat.jsx:75|
|C11|快捷写下一章|部分实现|当前可通过普通发送写章，但缺少原同位置快捷动作及相同确认语义。|NovelQuickActions.readable.js:11|frontend/src/user/pages/NovelCreationPage.jsx:304|
|C12|独立手动审稿|缺失|自动生产中审核存在；当前维护只开放文风分析/事实刷新，未见对应手动审稿入口。|NovelQuickActions.readable.js:11|backend/internal/novelcreation/maintenance.go:48|
|C13|独立手动修订|缺失|失败重试不是对已保存章节主动修订；缺同位置选择/执行/版本结果链。|NovelQuickActions.readable.js:11|frontend/src/user/pages/novel-creation/RunCard.jsx:19|
|C14|原工具阶段与日志|部分实现|CM有持久run/百分比/错误，原按tool动态阶段与日志卡没有完全对应。|NovelToolSteps.readable.js:7|frontend/src/user/pages/novel-creation/RunCard.jsx:1|
|C15|停止、失败重试、结果未知|已有实现待对照|当前有停止后续步骤、原任务恢复、结果未知不自动重试；这是已有真实保护，原native细节尚未知。|NovelToolSteps.readable.js:203|frontend/src/user/pages/novel-creation/RunCard.jsx:14|
|V01|已保存章节选择/编辑/版本|已有实现待对照|当前能编辑正文、版本历史与恢复；不应列为无功能，列表/编辑形态仍不同。|NovelStudioBookPanel.readable.js:225|frontend/src/user/pages/novel-creation/ChaptersPanel.jsx:32|
|V02|角色档案列表|缺失|当前右栏仅正文列表，缺原角色主次分类及档案浏览。|NovelStudioBookPanel.readable.js:261|frontend/src/user/pages/novel-creation/ChaptersPanel.jsx:25|
|V03|核心文件与文件浏览|缺失|缺核心文件树/文件详情，现有蓝图弹窗不能覆盖全部文件。|NovelStudioBookPanel.readable.js:287|frontend/src/user/pages/novel-creation/ChaptersPanel.jsx:25|
|V04|世界观摘要与完整设定|缺失|当前项目创建表单的world字段不等于原右栏摘要和完整设定文件。|NovelStudioBookPanel.readable.js:58|frontend/src/user/pages/novel-creation/ChaptersPanel.jsx:25|
|V05|封面资源与预览|缺失|当前小说资料区未提供原封面资源列表/预览与加载失败状态。|NovelCoverPreview.readable.js:100|frontend/src/user/pages/novel-creation/ChaptersPanel.jsx:25|
|V06|生成封面专用动作|缺失|原有封面确认动作和独立图像模型；当前只列文本步骤，未接对应功能。|NovelConfirmCard.readable.js:39|frontend/src/user/pages/novel-creation/AgentPanel.jsx:34|
|V07|短篇章节结构|缺失|当前短篇只是通用章列表的一次结果，没有独立短篇章节结构。|NovelStudioShortPanel.readable.js:193|frontend/src/user/pages/novel-creation/ChaptersPanel.jsx:7|
|V08|短篇大纲查看/编辑|缺失|长篇卷章蓝图不是短篇完整大纲文件；缺专用输出/编辑/继续。|NovelStudioShortPanel.readable.js:153|frontend/src/user/pages/novel-creation/BlueprintEditor.jsx:13|
|V09|短篇成品文件|缺失|原artifacts列表与文件编辑缺失，现有一个ChapterVersion无法承载。|NovelStudioShortPanel.readable.js:223|backend/internal/novelcreation/exports.go:101|
|V10|短篇推荐文案|缺失|缺推荐预览、完整推荐与保存字段。|NovelStudioShortPanel.readable.js:53|frontend/src/user/pages/novel-creation/ChaptersPanel.jsx:25|
|V11|通用资料详情编辑器|缺失|CM只有结构化蓝图和章节编辑，不能打开原任意业务资料文件。|NovelPanelArtifactView.readable.js:63|frontend/src/user/pages/novel-creation/BlueprintEditor.jsx:6|
|V12|编辑章节后真相同步与回放|合同不同|CM修改后标记dirty再显式刷新事实记忆；原保存payload含syncTruth/replayFrom，原算法未知。|NovelStudioBookPanel.readable.js:173|frontend/src/user/pages/novel-creation/ChaptersPanel.jsx:27|
|V13|资料栏显隐/宽度/拖动|部分实现|当前有可拖宽资料栏与按页显隐，不能据此说内容、默认宽度和像素全一致。|NovelStudioPage.readable.js:137|frontend/src/user/pages/NovelCreationPage.jsx:300|
|S01|独立文风新增编辑表单|合同不同|原直接名称/原文/指南编辑；当前复用资料解析与单文风选择，流程不同。|NovelStylePanel.readable.js:49|frontend/src/user/pages/novel-creation/MaterialsPanel.jsx:7|
|S02|文风多选与应用到对话|部分实现|长篇仿写专用输入及续写导入支持多文风；普通创作仍是单styleProfileId，文风工具并未整体改为多选。|NovelStylePanel.readable.js:11|frontend/src/user/pages/novel-creation/MaterialsPanel.jsx:52|
|S03|指南为空时分析策略|合同不同|原UI声明空指南保存时自动分析；CM保存统计，AI分析单独点击。原native费用与算法未知。|NovelStylePanel.readable.js:231|backend/internal/novelcreation/styles.go:49|
|S04|原文风八维/统计指纹|合同不同|原Tln/qln静态合同已恢复，CM仍是runes、单换行分段与简单句长，不能称等价。|NovelImportPage.readable.js:379|backend/internal/novelcreation/styles.go:29|
|S05|上传/粘贴资料真实解析|已有实现待对照|CM有真实TXT/DOCX/PDF上传、解析状态与章节范围；原文件选择/追加和资料用途不同。|NovelImportPage.readable.js:225|frontend/src/user/pages/novel-creation/MaterialsPanel.jsx:31|
|S06|长参考分块与当前截取上限|部分实现|长篇仿写已有独立全参考分块链；普通资料仍明确限制每份12000字，两条路径不同，不能宣称所有长参考均已对齐。|NovelImportPage.readable.js:380|frontend/src/user/pages/novel-creation/MaterialsPanel.jsx:43|
|S07|文风删除/编辑完整管理|部分实现|CM选用和建立文风存在；没有同一完整管理列表中的全部编辑/删除/多选动作。|NovelStylePanel.readable.js:191|frontend/src/user/pages/novel-creation/MaterialsPanel.jsx:52|
|S08|备忘录弹层与选择|部分实现|独立弹层与CRUD已实现；原共享笔记/会话与当前按作品noteIds、正文拼接语义仍不同。|NovelStudioSidebar.readable.js:41|frontend/src/user/pages/novel-creation/NotesPanel.jsx:5|
|A01|原17项模型分工|合同不同|原17配置包含聊天/编排/校验/卷压缩/互动/封面；CM仅8个实际收费步骤。不能添加空选项宣称接通。|NovelAgentSettings.readable.js:212|frontend/src/user/pages/novel-creation/AgentPanel.jsx:35|
|A02|37项审核维度|缺失|自定义补充已补不等于内置37维、不可改说明、全选/全不选及勾选注入。|NovelAgentSettings.readable.js:354|frontend/src/user/pages/novel-creation/AgentRequirements.jsx:9|
|A03|Agent保存范围|合同不同|原无作品/会话限定参数，CM按owner+project；原profile隔离物理实现仍未知。|NovelAgentSettings.readable.js:463|backend/internal/novelcreation/service.go:717|
|A04|补充审稿新增停用删除与注入|已有实现待对照|本批已真实补入reviewPrompts、校验、冻结和审核/修订注入，不再列为全缺。|NovelAgentSettings.readable.js:172|backend/internal/novelcreation/agent_requirements.go:11|
|A05|三类写作要求|已有实现待对照|已补分类与写作/修订注入；六模式本身不同仍会影响产物，不能以此证明模式完整。|NovelAgentSettings.readable.js:127|backend/internal/novelcreation/agent_requirements.go:57|
|A06|审稿循环与升温算法|尚未验证|UI数值与零值已支持；首轮偏移/同分回退/否决项/温度上限未证实与原native一致。|NovelAgentSettings.readable.js:29|backend/internal/novelcreation/pipeline.go:574|
|A07|后写校验与跨章输入|部分实现|15规则本地实现已补，跨章完整正文/标题/人称等生产输入仍不齐；显式未检查不是已通过。|NovelAgentSettings.readable.js:342|backend/internal/novelcreation/local_checks.go:30|
|A08|AI味检测|尚未验证|当前是明确标注CM自有统计，不是已恢复的原分类算法。|NovelAgentSettings.readable.js:346|backend/internal/novelcreation/local_checks.go:310|
|A09|长度归一化|部分实现|已补严格超过45%才触发模型；原计字、路由和扩缩策略还未全部恢复。|NovelAgentSettings.readable.js:336|backend/internal/novelcreation/local_checks.go:112|
|A10|生产二次确认与文件重读|合同不同|原是大纲/蓝图/设定后暂停；CM还有逐模型请求确认，含义不同。|NovelCheckpointCard.readable.js:22|backend/internal/novelcreation/model_confirmation.go:17|
|A11|原系统提示词与路由调用次序|尚未验证|当前后端提示词是本项目独立合同，原native完整提示词/实际角色次序没有全部恢复。不能因标签同名就称已配对。|NovelStudioPage.readable.js:307|backend/internal/prompt/novelcreation/definitions.go:8|
|A12|互动世界相关建议与模型配置|缺失|原建议动作和Agent枚举包含互动世界；当前六入口以外也没有对应完整玩法执行。|NovelConfirmCard.readable.js:38|frontend/src/user/pages/novel-creation/AgentPanel.jsx:34|
|T01|定时写作默认与启停|部分实现|30分钟/每日10/连续失败3的默认方向相符，CM有完整持久任务；原全局调度器状态与CM多计划UI不同。|NovelSchedulerPanel.readable.js:21|frontend/src/user/pages/novel-creation/AutomationPanel.jsx:19|
|T02|批量写作默认/进度/取消|部分实现|默认5章失败3次相符；CM是统一自动任务表单与持久队列，原简洁进度弹窗不同。|NovelBatchDialog.readable.js:20|frontend/src/user/pages/novel-creation/AutomationPanel.jsx:19|
|T03|自动任务引用与冻结|已有实现待对照|CM已有模型/引用/模式快照、原编号恢复和状态控制；这是可靠性实现，不是原入口形态等价。|NovelSchedulerPanel.readable.js:8|frontend/src/user/pages/novel-creation/AutomationPanel.jsx:25|
|D01|总章数/总字数/平均字数|部分实现|CM只有章数/字数/任务数/成功率，缺原平均字数和相同统计口径。|NovelDashboard.readable.js:46|frontend/src/user/pages/NovelCreationPage.jsx:313|
|D02|章节状态分布与字数趋势图|缺失|CM没有原按章节状态分布条与逐章字数趋势。|NovelDashboard.readable.js:54|frontend/src/user/pages/NovelCreationPage.jsx:313|
|D03|伏笔管理与活跃/已解决筛选|缺失|原看板实际挂载伏笔组件，显示起始/推进/解决；CM看板无该列表和状态合同。|NovelHooksPanel.readable.js:40|frontend/src/user/pages/NovelCreationPage.jsx:313|
|Q01|已保存数据与失败恢复保护|已有实现待对照|持久派发、租约、恢复保存和未知结果保护已做，必须保留；原native同等实现尚未证明。|NovelToolSteps.readable.js:203|backend/internal/novelcreation/worker_store.go:383|
|Q02|主题/图标/同状态截图|尚未验证|当前沿用用户要求的玻璃主题；原模式图标/控件几何与全部弹窗没有完成同状态对照。|NovelCreateIcon.readable.js:5|frontend/src/user/pages/novel-creation/AgentPanel.css:3|
|Q03|每个按钮真实效果|尚未验证|本轮枚举声明和静态调用，不是原应用每个按钮都运行过；自动测试通过也不等于原/CM完整对照。|NovelStudioPage.readable.js:5|frontend/src/user/pages/NovelCreationPage.jsx:30|

详细字段和hash见features.json；修复顺序与输入/输出/边界/失败规则见repair-spec.md。原版原生私有提示词及实际模型次序仍未知，不能把CM自编提示词写成已经恢复原版。
