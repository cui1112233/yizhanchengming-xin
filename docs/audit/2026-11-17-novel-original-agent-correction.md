# 原版 Agent 与执行链纠偏审计（只读）

日期：2026-11-17。只读检查原安装包前端、已有原生标识符证据及当前Go源码；唯一新增文件为本报告。没有修改产品、提交、模型调用、UI操作、运行包重启或清理临时目录。当前运行包保留、源码及整体交付HOLD由主流程管理；本报告不代表对当前运行包重新验收。

## 证据边界与引用

本次重新计算原bundle SHA256，结果为`5278a6d1b525152a6e9542c9d245bd90c658b84db3d94be9198ee2ac98306d4a`，与提取清单一致。清单明确可读片段是检查材料，运行等价性未验证，不是原开发源码。原生证据明确只是ASCII标识符匹配，不是函数体或调用图。见`F/manifest.json:2–10`、`N/api-identifier-candidates.json:2–3`。

以下路径缩写用于逐项定位，行号为本次读取文件的实际行号：

- **B**：`_local-archive-20261028/luoshui-novel-audit-20261117/installed-original/index-C1aA08SG.js`。
- **F/**：`_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/`。
- **N/**：`_local-archive-20261028/luoshui-novel-audit-20261117/recovered/native-current/`。
- **G/**：`backend/internal/novelcreation/`；**P**：`backend/internal/prompt/novelcreation/definitions.go`。

**证实**表示原前端字段、静态解码后的字面值、桥接参数或当前Go分支直接可见；**推断**表示据界面承诺推导的业务意图；**未知**表示没有原生函数体或真实提供方轨迹，不能补成算法事实。未复制原版私有提示词，未宣称恢复完整native代码。

## 先给结论

| 项目 | 原版证据 | 当前Go证据 | 结论性质 |
| --- | --- | --- | --- |
| 后写校验 | “确定性规则检查（零 LLM 开销）”，`F/NovelAgentSettings.readable.js:336–342` | 进入独立提示词并`callJSON`，替换正文，`G/pipeline.go:316–332`、`P:145` | **证实偏离**：当前新增收费调用及正文改写，不能称为原版校验 |
| AI 味检测 | “段落等长/套话密度/公式化转折（零 LLM 开销）”，`F/NovelAgentSettings.readable.js:343–348` | 独立LLM整理并替换正文，`G/pipeline.go:321–332`、`P:146` | **证实偏离**：本地检测被改成LLM润色 |
| 字数归一化 | “草稿字数偏离目标 >45% 时自动扩写/压缩”，`F/NovelAgentSettings.readable.js:330–336` | 开关为true就无条件调用，无偏离判定，`G/pipeline.go:316–327` | **证实控制条件偏离**；原生计字法、触发后模型/轮数仍未知 |
| 生产二次确认 | “大纲/蓝图/基础设定生成后暂停，可在右侧修改文件后再继续”，`F/NovelAgentSettings.readable.js:349–354` | 首次外呼后每个新call key都gate，包括repair，`G/pipeline.go:558`、`G/worker.go:530–562` | **证实交互偏离**：把内容里程碑确认变成逐外呼批准 |
| Agent配置范围 | 无作品/会话props，无作用域参数；“本地数据目录，永久生效”，`F/NovelAgentSettings.readable.js:6–14,122–144,219–243,461–469` | 按owner+project读写`novel_project_settings`，`G/mysql_store.go:785–788,831–841` | **证实前端合同不同**；原生是否另按登录profile分区、物理文件路径未知 |
| 模型路由目录 | 17个原始key/label，B:41，见下表 | 当前10个key，含自行增加的`post-write/ai-tells/length-normalize`，`G/validation.go:394–398` | **证实不能直接对号**；必须保留原词义，不能用新key表替代原目录声称一致 |

这些是业务语义偏离，不是把文字换回原词就能消除。特别是“后写校验/AI味检测”不能再因为同名开关存在或单元测试通过而算原版链路对齐。原生具体实现未知不影响原界面“零LLM”承诺和当前源码实际外呼之间的直接矛盾。

## 原17项路由：已静态解码，不从Go反推

原`crn`按顺序引用`Fon,Von,Uon,Kon,Jon,Hon,Xon,$on,Yon,ern,trn,nrn,arn,orn,rrn,lrn,srn`。纯Python复现其字符串表解码：`urn`有300项，旋转149；分别用已知属性锚点`key`和`label`核对，再解出赋值。未eval/import bundle、未启动原程序。原赋值与decoder均在**B:41**，Python code-point范围`8786954..8794589`，`crn`声明在其后；这个范围不是UTF16偏移，不与manifest的UTF16范围混用。

| 顺序 | 原key | 原label | 直接证据 |
| --- | --- | --- | --- |
| 1 | `chat` | 对话助手 | B:41，Fon |
| 2 | `planner` | 章节规划 | B:41，Von |
| 3 | `composer` | 上下文编排 | B:41，Uon |
| 4 | `architect` | 创建作品 / 设定生成 | B:41，Kon |
| 5 | `writer` | 章节写作 | B:41，Jon |
| 6 | `settler` | 真相沉淀 | B:41，Hon |
| 7 | `state_validator` | 真相校验 | B:41，Xon |
| 8 | `consolidator` | 卷级压缩 | B:41，$on |
| 9 | `auditor` | 审稿 | B:41，Yon |
| 10 | `reviser` | 章节修订 | B:41，ern |
| 11 | `short_fiction` | 短篇生成 | B:41，trn |
| 12 | `play` | 互动世界 | B:41，nrn |
| 13 | `play_parser` | 互动·行动解析 | B:41，arn |
| 14 | `play_engine` | 互动·世界推演 | B:41，orn |
| 15 | `play_narrator` | 互动·场景叙事 | B:41，rrn |
| 16 | `play_state` | 互动·状态提取 | B:41，lrn |
| 17 | `cover` | 封面配图 | B:41，srn |

非`cover`项使用传入的对话模型候选；`cover`走单独出图候选，原提示明确“与对话LLM无关”，且有Codex出图能力探测。路由空值为“跟随对话默认模型”，新增缺失key补空字符串。**这些证明16个文本模型配置入口和1个图像入口，不证明每个阶段每次必调LLM、调用先后或调用次数**。尤其`state_validator`本身是模型路由项，不能只因名字有“校验”就归为后写本地检查。见`F/NovelAgentSettings.readable.js:36–45,212–221,273–302`。

当前Go的`foundation-plan/chapter-plan/chapter-draft/chapter-audit/chapter-revision/truth-settlement`与部分原名存在职责相似，但不能宣称逐项等价；当前`memory-index`也不能据名称自动当作原`composer/consolidator`。现表缺少原`chat/composer/state_validator/consolidator/cover`等入口，同时增加原目录没有的三项可选处理路由。证据：原上表B:41与`G/validation.go:394–398`、`G/pipeline.go:573–584`；职责映射尚为**待确认**，此处不提出新UI。

## 字段、默认、范围与实际保存

Agent原初值为3/80/0.15/0.65及三个true、一个false；加载及保存返回后均重新读取`data.reviewCycle`，使用`??`保留返回的0。数字控件及文字如下。见`F/NovelAgentSettings.readable.js:26–35,153–162,252–261,302–354`。

| 原文案 | 原字段路径 | 初始/缺省 | 控件范围、步长 | 已证实含义及未知 |
| --- | --- | --- | --- | --- |
| 最大修订轮次 | `reviewCycle.maxRetries` | 3 | 0–5；未显式step | 写下一章未达分尝试修订，可能回退更高分版本；native实际循环边界未知 |
| 通过分数 | `reviewCycle.passScore` | 80 | 0–100；未显式step | 审稿通过分输入；native如何汇总37维、是否有否决项未知 |
| 升温步长 | `reviewCycle.temperatureStep` | 0.15 | 0–0.5，step 0.05 | 位于写章审稿循环；首轮是否+step、哪些角色升温、封顶方法未知 |
| 基础温度 | `reviewCycle.baseTemperature` | 0.65 | 0–2，step 0.1 | native实际传参、提供方不支持时处理未知 |
| 字数归一化 | `reviewCycle.enableLengthNorm` | true | bool | 界面承诺偏离目标>45%时才自动扩写/压缩；计字分母/阈值边界/native执行未知 |
| 后写校验 | `reviewCycle.enablePostWrite` | true | bool | 确定性规则，零LLM；具体规则集合/修正或报告形式未知 |
| AI 味检测 | `reviewCycle.enableAiTells` | true | bool | 本地检测三类特征，零LLM；统计公式、字典和阈值未知 |
| 生产二次确认 | `reviewCycle.requireSecondaryConfirm` | false | bool | 生成大纲/蓝图/基础设定后的内容检查点；实际各模式触发条件待native验证 |

“最大尝试数/MaxAttempts”不是这份原版Agent面板的字段名，也不是原控件标签。当前bundle中没有明文`maxAttempts`或`max_attempts`；这只能证明本次静态材料未找到这些明文，不能证明native内部绝无其他参数。不得把`maxRetries=3`直接说成“总共3次模型调用”，也不能将结构修复、传输重试和内容修订混算。原帮助文字是“写下一章时自动审稿-修订循环。未达通过分将尝试修订，修订后仍可能回退到更高分版本。”见`F/NovelAgentSettings.readable.js:302–308`。

**原保存前端有0值缺陷，不能隐去**：实际保存用`Number(v)||default`后夹范围，因此四个数值的0分别会变为3、80、0.15、0.65；负值先truthy后可夹到0。加载采用`??`，又可显示原生返回的0。原输入允许0与保存行为矛盾是可直接复现的表达式事实；native直接API能否保留0未知。当前Go的presence-aware显式0保留属于不同且有意的行为，不应伪称字节级复刻。见`F/NovelAgentSettings.readable.js:135–138,155–158,254–257`与`G/validation.go:371–380`。

### 真实桥接名和payload

再以纯字符串算法复现`ran`：`uan`有475项，旋转219；用`window.pywebview.api`中`pyweb`/`api`两个锚点交叉核对。得到下列方法名，与N中相同候选名相互支持；**方法名已解出，native函数体仍未知**。原decoder及bridge选择在B:41；对照`N/api-identifier-candidates.tsv:25–26,59–66`。

1. `novels_get_agent_settings()`：零参数。返回`success`与`data.modelRoutes/reviewPrompts/customWritePrompts/reviewCycle`。打开Agent页执行加载，保存成功后也以返回数据重置前端值。见`F/NovelAgentSettings.readable.js:219–262,145–163`。
2. `novels_save_agent_settings(payload)`：payload仅有`modelRoutes`、`reviewPrompts`、`customWritePrompts`、嵌套`reviewCycle`。没有`projectId/bookId/sessionId`。见同文件`:122–144`。
3. 原父页给Agent组件仅传`model-options`与主题、back事件，没有当前作品/会话ID。footer“设置保存在本地数据目录，永久生效。”这是**无作品/会话限定的本地持久Agent配置**合同证据；不是浏览器localStorage里临时保存本表。原物理文件、配置函数是否叫`SaveConfig`、是否按隐含登录账号/profile隔离均**未知**，不能从无ID参数推成跨账号共享。见`F/NovelStudioPage.readable.js:1613–1618`、`F/NovelAgentSettings.readable.js:461–469`。

当前Go把字段摊平进`AgentSettings`，用`maxRevisionRounds`而非`reviewCycle.maxRetries`，并加入原Agent面板没有的`targetChapterWords/thinkingLevel`字段；原Agent没有这两个控件不代表原产品其他界面没有字数/思考设置。当前项目ID及MySQL作用域是直接可证实的结构差异。见`G/types.go:136–149`、`G/store.go:78–85`、`G/mysql_store.go:785–788,831–841`。

### 37维、补充审稿与写作要求

原版明确有“审稿维度（37项）”，内置说明不可改字，可开关、全选/全不选；另有可新增/编辑/删除的自定义审稿项。保存维度项包含`id,dimensionId,title,content,enabled,kind:"dimension"`，自定义项为`id,title,content,enabled,kind:"custom"`。文案声明勾选项参与审稿、自定义补充在维度列表之后。没有原生实现，尚不能验证native是否严格按此注入，但**不能以“没有功能证据”抹去该UI和payload合同**。见`F/NovelAgentSettings.readable.js:69–96,168–182,354–386`。

另有`customWritePrompts`，字段`id,category,title,content,enabled`；类别为`long_fiction/short_fiction/imitation`，标签“长篇正文/短篇小说/仿写创作”。文案声明勾选后注入对应写作Agent系统提示。见`F/NovelAgentSettings.readable.js:127–133,190–209,415–417`。本报告不提取或复制这些内置说明/私有系统提示词原文。

当前Go的`AgentSettings`没有这两个列表，`ProjectSettings.Rules []string`也没有维度ID、类别、enabled等语义；“统一规则”不能算还原了原列表。见`G/types.go:136–149`与`G/store.go:78–85`。先前Task5 brief将它们排除是当时实施范围事实，并不等于原版不存在；本次字面对齐规格必须列明这项真实差异。

## 动作确认与生产检查点是两套合同

### 1. 对话建议动作，再确认执行

普通发送实际调用`novels_send_message`，参数为`sessionId,message,model,knowledgeBookIds,thinkingMode:true,codexReasoningEffort,claudeReasoningEffort`，可补`contextLength`。响应`.data.action`进入pending action，加载会话也从`.pendingAction`恢复。它不是在前端无条件直接写正式章节。见`F/NovelStudioPage.readable.js:286–307,420–460`。

动作卡列出`create_book/write_next/audit_chapter/revise_chapter/short_run/play_start/generate_cover`等确认动作，也有`style_imitation/fanfic_init/spinoff_create/continuation_import`打开专用入口的动作。按钮会区分“打开入口/继续执行/已打开/已执行”，不是每个建议都执行模型。见`F/NovelConfirmCard.readable.js:21–39,69–77,88–105`。

确认动作实际为`novels_confirm_action({sessionId,model,codexReasoningEffort,claudeReasoningEffort})`，并据响应区分assistedAction、切换作品、awaitingConfirmation、审稿结果、章节结果。这里的`model`读取**当前对话选择**，不是只把初始模型回传。见`F/NovelStudioPage.readable.js:645–705`。卡片取消在当前前端`Je()`清pending action显示；重试`Ze()`会重新读会话并执行pending action。native幂等/是否清除持久建议的具体语义未知，不能因前端没有token便推断原版会重复收费。见同文件`:708–725`。

### 2. 大纲/蓝图/基础设定产出后确认继续

- `tool:checkpoint`事件进入`Fe()`；state保存`sessionId,tool,checkpoint,storyId,bookId,message`。`outline`将前三阶段标已完成，`blueprint`标到索引2；短篇缺少阶段列表时采用“生成短篇大纲→审阅大纲→修订大纲→撰写正文→审阅正文→修订正文→打包输出”的显示列表。这是前端进度显示，不等于这些阶段各调用一次LLM。见`F/NovelStudioPage.readable.js:540–581,1012`。
- 卡片标题根据checkpoint显示“确认创作蓝图/确认基础设定/确认继续生成”，帮助为可在右侧修改本地文件，主按钮“确认继续”，次按钮“稍后”。“稍后”仅清除当前前端checkpoint显示；独立取消操作另走`novels_cancel_operation`。见`F/NovelCheckpointCard.readable.js:21–39`、`F/NovelStudioPage.readable.js:642–644,711–721`。
- 继续调用已静态解出的`novels_confirm_pipeline_checkpoint`，参数为`sessionId,tool,checkpoint,storyId,bookId,message,model,codexReasoningEffort,claudeReasoningEffort`。响应可再次`awaitingConfirmation`并替换checkpoint，否则刷新会话/作品/短篇。参数中没有每个外呼的call key或confirmationId，也没有内联编辑后的正文/大纲。见`F/NovelStudioPage.readable.js:583–640`。
- 右侧artifact编辑器先调用独立文件保存桥接。短篇保存传`storyId,key,content`，作品文件保存传`bookId,key,content`；章节保存还传`syncTruth:true,replayFrom:true`，并展示`replayedFrom/syncedTruth`结果。**证实**是“先保存文件，后按内容检查点继续”；**推断**native继续应使用已保存资料，但其确切重读、重新审核、真相重算算法没有函数体证据。见`F/NovelStudioShortPanel.readable.js:111–136`、`F/NovelStudioBookPanel.readable.js:167–194`、`F/NovelPanelArtifactView.readable.js:29–43,60–65`。
- 仿写短篇入口单独识别`awaitingConfirmation`并将tool置为`imitation_short_run`，默认提示“创作蓝图已生成，可在右侧修改后确认继续”。这直接证明不只长篇可能有内容检查点。见`F/NovelStudioPage.readable.js:889–898`。

当前Go还有独立差异：长篇初始foundation检查点由`modeUsesCheckpoint`决定，并不看`RequireSecondaryConfirm`，短篇默认straight-through；而同一开关又在journal中控制每个后续外呼。因此“关闭生产二次确认”与“初始蓝图仍被强制确认”在当前代码可同时发生。原开关false时各模式native实际执行尚未验证，但当前把它接到逐调用gate、不是内容checkpoint，已可直接判定不等同。见`G/service.go:126–132,270`、`G/validation.go:122–130`、`G/pipeline.go:246–247,558`。

当前内容蓝图确认接受可选JSON文档，创建`confirm:<sourceRun>`唯一后继run，并冻结原模型；逐调用确认则使用`confirmationKind:"next_model_call"`和token，原run返回queued。原前端没有这套逐调用生产交互。**内部token、幂等与持久运行可以保留，不能拿内部可靠性机制替代原用户检查点含义。**证据：`G/service.go:600–638`、`G/model_confirmation.go:9–22,113–132`与原继续payload（`F/NovelStudioPage.readable.js:619–640`）。

## 当前审核循环：哪些相似，哪些不能认作原版算法

原帮助确认了未达分会尝试修订、可能回退更高分版本；当前Go确有有界round、复审和best版本比较，属于可见方向相似。原native是否“首稿+maxRetries轮”、是否无issues就停止、相同分如何选择、多少阶段共用此设置，均未知。当前Go实际是首稿审核后最多`MaxRevisionRounds`次修订/复审，issues空时即停止；未通过保留正文为`succeeded_with_warnings`。见原`F/NovelAgentSettings.readable.js:302–308`与`G/pipeline.go:349–402,1044–1055`。

当前还强制`passed == score>=threshold && 无high/critical问题`，不一致走结构修复；新审核提示词把该规则写入v2。这是当前可证实的独立合同，**原前端没有足够证据证明原native也用同样否决、schema或结构修复政策**，不能写成恢复所得。当前所有正常步骤用基础温度、修订第1轮即`base+step`并封顶2，也只是当前算法，原UI只证明数值和“升温步长”标签。见`G/pipeline.go:495–501,876–913`、`P:148–154`；原控件见`F/NovelAgentSettings.readable.js:316–329`。

原审稿UI从native`passed`或嵌套`audit.passed`决定通过/警告，没有在前端重算37维得分；手动审稿、修订有独立桥接入口和状态。当前不能仅凭分数默认相同宣称审稿算法一致。见`F/NovelStudioPage.readable.js:700–705,1150–1230`。

## 可供修订规格采用的最小事实合同（不是编码/UI方案）

1. 保留原Agent字段层级与原17项key/label的可追溯对照，明确文本/图像路由。对未证实的native执行频次/角色映射标未知，不擅自增加可收费节点。依据：B:41路由表与`F/NovelAgentSettings.readable.js:273–302`。
2. `enablePostWrite/enableAiTells`的业务含义必须是确定性/本地检查且零LLM；`enableLengthNorm`保留原承诺的条件性触发，具体计字/规则算法待补证，不用新的LLM润色冒充。依据：`F/NovelAgentSettings.readable.js:330–348`。
3. 区分“对话动作确认”与“生产内容检查点”；二次确认绑定outline/blueprint/foundation及已保存可编辑资料，不是批准每次模型请求。原生重复继续的幂等细节未知，但这不妨碍保留当前内部token防重复机制。依据：`F/NovelConfirmCard.readable.js:21–39`、`F/NovelStudioPage.readable.js:540–640`。
4. Agent保存应明确原本“无作品/会话参数、本地长期生效”的共享范围证据，以及账号/profile未知边界，不能默认为每作品专属；37维、自定义审稿、三类写作要求作为真实合同差异列入规格。依据：`F/NovelAgentSettings.readable.js:122–144,354–381,415–417,461–469`。
5. 不把“修订轮次”改成总尝试数，不把未恢复的升温公式/37维算法写成事实。对原保存吞0的缺陷应在修订规格中明确处理决定，不能既悄悄变更又声称完全一致。依据：同文件`:135–138,302–329`。
6. 纠偏必须与持久运行版本分开：当前已经冻结`agent_contract_version`及项目输入，journal在解析新provider前复用completed原响应，unknown不自动外呼。将来修订语义应有清楚的运行来源/版本边界，**不得为“纠正流程”重跑已完成LLM、改变历史请求指纹或为本地检查新增收费调用**；未发送旧计划需要如何停留/迁移由修订规格明确，不应自动推断用户授权额外费用。现有权限隔离、租约/持久调用记录、结果未知保护继续保留。依据：`G/service.go:140–154`、`G/worker.go:492–516,566–603`、`G/pipeline.go:558–598`。

## 仍未知，不能拿实现猜原版

- 原native后写规则、AI味指标的具体公式/词表、是否只告警或局部修正。
- 字数>45%的精确计数口径、是否所有创作模式适用、扩写/压缩路由与轮次。
- 37维完整评分汇总、否决项、修订回退/同分选择、`maxRetries`真实循环及原生结构修复策略。
- 温度步长的适用角色、首轮偏移、上限、模型不支持参数时策略。
- 原Agent持久化具体文件/原生SaveConfig实现、账号/profile隔离、运行中改配置如何影响下一阶段。
- 16个文本路由是否每次实际调用、组合顺序、`composer/state_validator/consolidator`和当前Go记忆/事实模块是否真正等价。
- 原生产checkpoint继续时文件重读/复审、取消/稍后/重试的native持久状态及幂等语义。

未知项的边界来自`F/manifest.json:5–10`与`N/api-identifier-candidates.json:2–3`；本报告没有执行原native或真实提供方，不能补足这些事实。以上已经证实的零LLM和内容检查点合同足以要求更正当前方向，但不支持继续“自己设计”未见过的原生算法或新界面。
