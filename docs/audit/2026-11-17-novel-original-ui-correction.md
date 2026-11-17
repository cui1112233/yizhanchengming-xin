# 原版小说 UI 与调用合同纠偏审计（只读）

补证说明：本报告保留只读审计时的U编号；并行完成的`2026-11-17-novel-original-agent-correction.md`随后静态解出了17项路由和5个桥接名。其中U1的普通发送/动作确认已分别确定为`novels_send_message`/`novels_confirm_action`，内容继续为`novels_confirm_pipeline_checkpoint`。此补证不代表U2–U5或原生算法已恢复，也不把原路由表推断成实际调用次数。

日期：2026-11-17。结论：**当前通用聊天工作台不等于洛水当前版；仅保留六个名称和三栏宽度，不能宣称一比一。当前 UI 按用户反馈判定不通过，整体验收 HOLD。** 直接差异见下表的入口、表单、DTO和结果文件合同；本次不提出替代布局，也不修改产品代码。

## 证据边界与引用方法

- 下文 `O/` 的绝对根目录为 `F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/frontend-current/`；其后均为**精确文件名与格式化片段行号**，不是原作者源码行号。`N`精确指 `F:/脚本测试/一战晟铭-v88/_local-archive-20261028/luoshui-novel-audit-20261117/recovered/native-current/api-identifier-candidates.tsv`。`frontend/`、`backend/`相对当前仓库根目录。恢复方式与限制见 `O/manifest.json:1–10`、`docs/reverse-engineering/2026-11-17-luoshui-novel-current-baseline.md:5–26`。
- 本次只读重新计算：23个片段与manifest哈希全部一致；实际安装 `dist/assets/index-C1aA08SG.js` SHA256仍为 `5278a6d1b525152a6e9542c9d245bd90c658b84db3d94be9198ee2ac98306d4a`，`modules/novels.pyd`仍为 `0471b2a77ee1aad06f4dd30e58b2515ed89b3a20d780ad5ae0f64ab3e4552853`，与上述基准一致。没有运行这些文件或操作原窗口。
- `S`=片段可直接证明的静态 UI/前端请求合同；`D`=从已读代码路径得出的差异；`U`=外部绑定/原生实现未确认。根提供的当前 v2.9.3 原窗口“仿写短篇”AX观察与该专用组件字段顺序相符，但本审计未重新操作/抓屏；没有同状态截图，不将AX当像素证明。原生完整函数体、真实模型结果与费用均未恢复/验证；基准本身也如此限定（docs/reverse-engineering/2026-11-17-luoshui-novel-current-baseline.md:18–24,67–69）。

## 一、六个侧栏入口：实际点击 → 输入 → 请求 → 结果 → 当前缺口

| 模式 | 原版入口/表单（S） | 原版前端请求及结果链（S，桥接名未解码处标U） | 当前实现的具体差异（D） |
|---|---|---|---|
| 长篇原创 `book-create` | 侧栏点击只 `switch-workflow`，**不打开导入弹窗**；切换到聊天，模式不同则清空当前会话展示。空态是自然语言创作提示，非必填书名/题材/模型的建项目表单。`O/NovelStudioSidebar.readable.js:65–74,255–257`；`O/NovelStudioPage.readable.js:281–284`；`O/NovelStudioChat.readable.js:345–350,610–620`。 | 首次发送可先建无bookId会话；发送 `sessionId,message,model,knowledgeBookIds,thinkingMode,codexReasoningEffort,claudeReasoningEffort,contextLength?`。返回 `action` 进入确认卡；`create_book`→`architect`，确认执行可返回 `book + switchToBook + session` 再切正式书籍会话。`O/NovelStudioPage.readable.js:309–380,420–460,494–499,641–678,685–689`；`O/NovelConfirmCard.readable.js:29–44`。聊天桥接属性完整拼接未解码（U1）。 | 当前无project时发送直接开通用“新建小说作品”；必填title/genre/model，默认每章2000，另world/protagonist/coreConflict，是新增前置流程。`frontend/src/user/pages/NovelCreationPage.jsx:121–132,235–236`。Go提前要求作品关联，所有运行准备 FoundationPlan→ChapterPlan；不能把原聊天确认建书路径说成已复刻。`backend/internal/novelcreation/service.go:92–132`；`backend/internal/novelcreation/pipeline.go:213–275`。原版此入口的目标字数/章数提取默认值在当前片段内**未知**，不得套13000。 |
| 短篇原创 `short` | 同样只切换聊天，未出现专用新建项目弹窗；六项中 `short` 明确排除 `open-import`。`O/NovelStudioSidebar.readable.js:70–74,255–257`。 | 自然语言发送合同同上；确认卡有 `short_run`，映射 `short_fiction_run`。短篇检查点UI声明“大纲生成/审核/修订→正文/审核/修订→打包”；返回 `shortFiction/storyId` 加载独立短篇结果并显示 ShortPanel。`O/NovelConfirmCard.readable.js:33–37`；`O/NovelStudioPage.readable.js:494–499,567–580,689–699,813–831,1596–1604`。完整原生选参/阶段实现未知（U1/U6）。 | 当前只有 `short_original` 名称及一段mode rule，复用同一章链、同一项目章节序号；短篇导出实质查该run的一个ChapterVersion。没有相同的短篇大纲/正文两轮审核与成品文件组织。`backend/internal/novelcreation/pipeline.go:251–305,453–461,1103–1143`；`backend/internal/novelcreation/service.go:120–124`；`backend/internal/novelcreation/exports.go:101–128`。不能将“独立短篇导出可用”等同独立短篇产品流程相同。原版短篇原创的固定表单/13000默认值无此片段证据。 |
| 仿写长篇 `imitation` | 点击 `switch-workflow` + `open-import('imitation')`，打开仿写创作专用导入窗，字段按第二节A；参考原文至少100个非空白字符。`O/NovelStudioSidebar.readable.js:75–78,255–257`；`O/NovelStudioPage.readable.js:385–389,1522–1525,1648–1659`；`O/NovelImportPage.readable.js:302–304,511–690`。 | 表单组装 `sessionId?,title,storyIdea,referenceText,genre:'other',knowledgeBookIds[]`，加 `creationMode:'styleImitation',selectedDirection?,targetWordCount(每章),targetChapters,readerPreferences?,supplementaryInfo?,styleProfile,styleDimensions,referenceChunks?`。先emit `submit-start`、清表单并关窗，父页面补sessionId/model并进入 `architect`；桥接属性U2。`O/NovelImportPage.readable.js:315–382`；`O/NovelStudioPage.readable.js:488–492,513–540`。 | 当前点击只换activeMode；参考只有一段内联文本，目标字数5000/目标100章、选定仿写方向、readerPreferences、storyIdea、补充信息、统计指纹/维度及50k分块都没有对应正式字段；还允许仅文风提交，反而不同于原版可见校验。`frontend/src/user/pages/NovelCreationPage.jsx:131–132,204,217`；`backend/internal/httpapi/novel_creation_routes.go:126–144`；`backend/internal/novelcreation/validation.go:84–118`；`backend/internal/novelcreation/automation_input.go:119–130`。 |
| 仿写短篇 `imitation-short` | 是**专用弹窗**：点击额外 `open-imitation-short`，标题/副标题及字段顺序见第二节B。`O/NovelStudioSidebar.readable.js:79–82,255–256`；`O/NovelStudioPage.readable.js:1525,1659–1666`；`O/NovelImitationShortDialog.readable.js:181–338`。 | emit `referenceText,title?,targetWordCount,targetChapters,chapters,supplementaryInfo?,readerPreferences?,inspirationTags?,styleName?,linkedSaveId?,knowledgeBookIds?[]`；memos实际拼入referenceText。父补sessionId/model，显式以 `imitation_short_run`显示11阶段，可返回blueprint检查点，最终storyId→短篇实体/文件。`O/NovelImitationShortDialog.readable.js:143–164`；`O/NovelStudioPage.readable.js:133,838–916`。动态桥接名U3。 | 当前没有该弹窗和13000/1章字段合同，通用必填项目先行；丢失目标章数、读者偏好、灵感多选、可选关联存档和多文风。Go 800–8000限制不能表达原版13000总目标，更不能以改Agent每章值替代总目标。当前`modeUsesCheckpoint`也不为short_imitation首轮启用蓝图检查点；原版允许该专用流程返回蓝图。`frontend/src/user/pages/NovelCreationPage.jsx:217,235–236`；`backend/internal/novelcreation/validation.go:10–12,125–133,221–234`；`backend/internal/novelcreation/types.go:89–108`；`backend/internal/novelcreation/service.go:132`。 |
| 番外创作 `spinoff` | 点击打开番外导入窗。顺序为正传书籍必选→书名必填→番外类型默认sequel（后传/前传/平行线/If线）→番外方向→参考文本可选→文风多选。`O/NovelStudioSidebar.readable.js:83–86,255–257`；`O/NovelImportPage.readable.js:39–46,305–307,692–729,752–769`。 | 通用body `{sessionId?,title,storyIdea,referenceText,genre:'other',knowledgeBookIds[]}` 加 `parentBookId,spinoffType`；本组件直接await桥接，loading后返回创建book并emit created，父刷新/转书籍。`O/NovelImportPage.readable.js:315–322,384–438,465–468`；`O/NovelStudioPage.readable.js:1648–1658`。动态方法U4；原生如何读取正传上下文未恢复。 | 当前侧栏“番外”对应Go `fanfiction`，但Go modeContract名字还是“同人创作”，仅要求existingSetting，没parentBookId/spinoffType关系与新作品创建合同。`frontend/src/user/pages/novel-creation/contracts.js:2–7`；`backend/internal/novelcreation/pipeline.go:460`；`backend/internal/novelcreation/validation.go:86`。不能把手填世界规则/复制当前作品设定当作原正传关联番外流程。 |
| 续写创作 `chapters` | 点击打开续写导入窗。顺序为导入到书籍（可选不选则新建）→仅新建时显示新书书名→章节正文→文风多选。已有目标书时不强制新书名；正文必填。`O/NovelStudioSidebar.readable.js:87–90,255–257`；`O/NovelImportPage.readable.js:308–309,698–707,752–769`。 | 本地H按“第…章”标题拆为 `{title,content,wordCount}` 数组，无标题兜底第1章；调用body加 `bookId?,chapters[]`。成功明确显示“章节已导入”，将返回book交给created路径，并非本表单直接运行写正文。`O/NovelImportPage.readable.js:196–217,438–468`。桥接方法U5；原生写库/上下文重建细节未知。 | 当前是 `previousText + 创作要求` 立即POST普通生成run；可复制最后章或文档选段，但没有“把导入前文章节正式落到已有/新书”这一步、目标书选择和章节数组合同。Document资料记录不等于正式小说前文章节。`frontend/src/user/pages/NovelCreationPage.jsx:121–151,217`；`backend/internal/novelcreation/validation.go:87`；`backend/internal/novelcreation/types.go:89–108`；`backend/internal/novelcreation/pipeline.go:251–305`。 |

说明：原ImportPage还含 `fanfic`同人子流程（canon/au/ooc/cp），但它不是侧栏六项里的第七项；不能把它与 `spinoff`番外混称或擅自增设侧栏入口。`O/NovelStudioSidebar.readable.js:65–90`；`O/NovelImportPage.readable.js:110–127,305–306,411–414,730–751`。当前Go“fanfiction=同人”与前端“fanfiction=番外”的混淆是实锤差异，非待猜测项。

## 二、必须逐字段保留的两种仿写表单

### A. 仿写长篇（顺序即渲染顺序）

| 序 | 原UI字段/值域/交互 | 实际payload对应与证据 |
|---|---|---|
| 1–3 | 暂定标题可空；参考原文必填≥100非空白字符；上传docx/txt，读取内容填入可编辑原文；按path跳过重复。多章默认最近5章，3/5/10与自选，子窗全选/清空/取消/确定，确认后按**源章节序号排序**组成原文。不是按点击先后拼接。 | `title,referenceText`；UI `O/NovelImportPage.readable.js:513–574`；解析/读取 `:196–279`；选择提交 `:281–295,790–843`；校验 `:302–304`。本片段未显示文件数/字节上限，不能把Go的8文件/20MiB/12k参考限额写成原版值。 |
| 4 | 文风统计指纹：平均句长、句长波动、段落长度、词汇多样性、高频句首、修辞特征；原文≥100后600ms计算。UI说明纯算法，不耗AI；实际Tln/qln实现未在组件内。 | `styleProfile=R||Tln(text)`、`styleDimensions=qln`；`O/NovelImportPage.readable.js:367–380,471–479,574–588`。八维具体定义U6。 |
| 5–6 | 每章字数默认5000，1000–50000，step500，快捷3000/4000/5000/6000/8000/10000；目标章节数默认100，1–500，step10，快捷50/100/150/200/300。 | `targetWordCount,targetChapters`；默认/快捷 `O/NovelImportPage.readable.js:53–97`；UI范围 `:588–631`；DTO `:374–375`。Go每章800–8000且没有目标总章数，不能直接同名对接（`backend/internal/novelcreation/validation.go:10–12,221–234`）。 |
| 7–10 | 读者偏好自由填+题材标签追加去重；原创故事想法可选；仿写方向可选，空时UI说默认拆书方向一；补充信息可选。 | `readerPreferences,storyIdea,selectedDirection,supplementaryInfo`；标签12项 `O/NovelImportPage.readable.js:61`；UI `:631–670`；DTO `:315–322,372–377`。 |
| 11–12 | 文风设置多选卡片；关联备忘录多选并显示数量。 | 文风ID数组 `knowledgeBookIds=c`，备忘录`h`逐条把plain拼入referenceText，未以独立memoIds发送。`O/NovelImportPage.readable.js:310–322,670–690`。 |

提交/关闭：仿写分支先组上述payload（原文按50k分块，超过一块才带referenceChunks）→emit submit-start→清所有表单→关窗，父组件负责开始/失败显示；并非本窗等生成成功才关闭。其他导入分支 `l=true` 后await桥接，成功重置/关闭、finally清loading。关闭本身只是emit close/back，没有本组件的取消后台生成调用。`O/NovelImportPage.readable.js:297–300,323–382,384–468,775–790`；`O/NovelImportModal.readable.js:31–42,48–58`。

### B. 仿写短篇（原窗口AX已由根确认，同字段有静态证据）

| 序 | 原UI字段/值域/交互 | 实际payload对应与证据 |
|---|---|---|
| 1–3 | 暂定标题可空→参考原文≥100非空白字符→docx/txt可选文件；文件读取后填入原文，可继续编辑。 | `title?,referenceText`；`O/NovelImitationShortDialog.readable.js:90–143,192–225`。删除附件行仅splice附件列表，不能推断同时从原文删其文字（`:219–225`）。 |
| 4–5 | 目标字数默认13000；1000–500000、step500；快捷10000/11000/12000/13000/15000/20000。读者偏好自由填，5个标签追加去重（轻松搞笑、热血爽文、智斗权谋、悬疑惊悚、甜宠治愈）。 | `targetWordCount`（提交函数也夹值）；`readerPreferences`，同时把偏好放 `styleName`。`O/NovelImitationShortDialog.readable.js:35–39,65–88,115–121,153–159,226–267`。**13000是这个短篇目标，不是每章Agent字数。** |
| 6–8 | 目标章节默认1，1–50、step1；补充信息可选textarea3行；灵感标签多选/再次点击取消，22个标签含女频/男频/丧尸等。 | `targetChapters`与兼容别名`chapters`同时发送；`supplementaryInfo,inspirationTags`（选中值用“、”连接为字符串）。`O/NovelImitationShortDialog.readable.js:36–41,88,119–121,153–158,267–297`。 |
| 9–11 | 文风设置多选；关联备忘录多选；关联存档可空单选（books列表）。 | `knowledgeBookIds[]?`；memo正文拼入referenceText；`linkedSaveId?`是独立可选关联，不是强制当前projectId。`O/NovelImitationShortDialog.readable.js:43–64,89–93,145–162,297–324`。 |
| 页脚 | 取消/开始仿写；原文不足100时开始禁用。取消只关窗；submit emit后清字段、恢复13000/1并关窗。局部`r`作为disabled/mask guard存在，但组件里没见设置为true，不能据此保证本窗提交锁。 | `O/NovelImitationShortDialog.readable.js:28–40,111–113,143–164,167–169,326–338`。父页面显示运行/错误/检查点，详见下一节。 |

重要纠偏：此前把“只选文风也能仿写”写成原版合同，是把非imitation分支的通用placeholder套到了仿写分支。**两种仿写当前前端静态代码都要求原文≥100**；`O/NovelImportPage.readable.js:302–304,752–757`、`O/NovelImitationShortDialog.readable.js:143`。同人也有单独必填原文校验（ImportPage:306），即使通用标签称可选。两处最少100校验都发生在追加备忘录正文之前，不能拿所选备注代替必填原文；不能以这条提示推断原生可接受空参考。

## 三、提交后的阶段、检查点与右侧资料，不能折叠成同一章节Tabs

| 合同 | 原版直接证据（S） | 当前差异（D） |
|---|---|---|
| 仿写短篇阶段 | UI预置11阶段：拆书分析→仿写方向→创作蓝图→短篇大纲生成/审核/修订→正文/审核/修订→编辑审核→打包。父ct补sessionId/model；失败error；awaitingConfirmation后保留检查点；storyId结果载入短篇。`O/NovelStudioPage.readable.js:133,838–916`。这是前端声明/事件消费，不是已恢复的11次模型调用证据。 | Go全部模式共享 FoundationPlan→ChapterPlan→单章正文→可选整理→章节审核/修订→事实/记忆；模式差异主要是modeContract/meta。没有对应拆书/方向/短篇大纲三阶段审核/编辑审核/成品文件数据合同。`backend/internal/novelcreation/pipeline.go:213–334,340–446,453–461,1103–1143`。 |
| 原创短篇阶段 | 无独立表单证据；短篇检查点预置7阶段，大纲前三步完成后继续正文与打包。`O/NovelStudioPage.readable.js:567–580`。 | `short_original`也进入同一章链，短篇读取/导出返回一个ChapterVersion，而非独立shortStory/artifacts。`backend/internal/novelcreation/exports.go:101–128`。 |
| 中间确认 | tool/checkpoint/bookId/storyId/sessionId/message +当前模型/思考参数发确认；可能再次awaitingConfirmation。右侧可修改生成文件后继续；标题区分blueprint/foundation。`O/NovelStudioPage.readable.js:542–580,587–637`；`O/NovelCheckpointCard.readable.js:18–38`。 | 当前蓝图是结构化字段Modal；另有每次后续模型调用opaque token gate。这些可靠保护不代表原大纲/文件检查点已对齐。`frontend/src/user/pages/novel-creation/BlueprintEditor.jsx:1–17`；`backend/internal/novelcreation/worker.go:531–562`。 |
| 右侧模式/有数据时显示 | 长篇 BookPanel：章节、角色档案、核心文件、封面资源、世界观；ShortPanel：章节结构、完整大纲、成品文件、推荐文案；能进入文件浏览/编辑。不同于始终显示同一章列表。`O/NovelStudioBookPanel.readable.js:29–50,211–354`；`O/NovelStudioShortPanel.readable.js:33–62,148–159,186–269`；父显隐 `O/NovelStudioPage.readable.js:157–176,1566–1604`。 | 当前右侧“当前资料+正文/进度Tabs”统一两类，没有原对应文件树/角色/世界观/短篇成品。`frontend/src/user/pages/NovelCreationPage.jsx:228–233`；`frontend/src/user/pages/novel-creation/ChaptersPanel.jsx:1–37`。 |
| 流与进度 | 消费chat.stream.start/done、tool:start/log/progress/checkpoint/end，按session/tool过滤，动态阶段可能覆盖预置。`O/NovelStudioPage.readable.js:944–1019`。 | 当前游标轮询可靠，但事件结构及UI不同；不能把Go run.step百分比称为原阶段完整等价。`backend/internal/novelcreation/pipeline.go:667–689`；`frontend/src/user/pages/novel-creation/RunCard.jsx:1–20`。 |

## 四、共享工具/布局的已证实误差

- **位置与分组**：原“开始创作”下连续六项；我的作品可展开每书会话，另有会话记录；六工具在左侧栏下部。我们增设“原创创作/参考创作/衍生创作”三组，并把六工具挪到右资料栏，是自行设计。`O/NovelStudioSidebar.readable.js:248–262,272–328,333–382` 对比 `frontend/src/user/pages/NovelCreationPage.jsx:204–207,228`。原右栏宽38%/360–800有依据（StudioPage:135–150），但宽度正确不能证明整个页面相同。
- **工具不是统一Drawer**：原风格、Agent设置切换主工作区，有“返回对话”；导入走modal，定时/看板主区域，备忘录独立popup。当前统一520宽Drawer不是原布局/状态合同。`O/NovelStudioPage.readable.js:391–405,1604–1666,1690–1711` 对比 `frontend/src/user/pages/NovelCreationPage.jsx:240–245`。
- **多文风**：原选中数组能在风格页多选后“应用”到聊天，聊天传knowledgeBookIds；风格对象含sourceText/styleGuide，可编辑/更新，UI说明styleGuide空时保存自动分析。当前单styleProfileId下拉+编号导入+纯统计/单独AI分析是另一个流程，knowledgeIds文档引用不能冒充多文风。`O/NovelStylePanel.readable.js:11–17,51–87,193–194,210–239,253–281`；`O/NovelStudioPage.readable.js:430–438,475–478,1604–1613` 对比 `frontend/src/user/pages/novel-creation/MaterialsPanel.jsx:49–57,73`；`backend/internal/novelcreation/automation_input.go:38–52`、`backend/internal/novelcreation/styles.go:49–68`。原分析具体算法/调用费用未知，不能由文案强推Native body。
- **Agent语义**：原reviewCycle显示默认3/80/0.15/0.65，数值范围0–5/0–100/0–0.5/0–2；启用长度归一化仅文案承诺偏离>45%时处理，后写校验和AI味检测明确标“零LLM开销”，二次确认是大纲/蓝图/基础设定生成后暂停。当前Go三个optional开关各直接触发p.callJSON，且二次确认按每个后续外部调用gate，语义与费用边界不同。`O/NovelAgentSettings.readable.js:26–34,304–354` 对比 `backend/internal/novelcreation/pipeline.go:316–327,558`、`backend/internal/novelcreation/worker.go:531–562`。原37审稿维度、自定义审稿/customWritePrompts也有UI和保存数组，不是当前十路由表单的同义物（原Agent:123–143,172–210,354–435；当前 `AgentPanel.jsx:1–43`）。原路由枚举`crn`外部绑定未解码，不把Go十阶段称为原路由全集。
- **字数/引用上限**：Go项目与Agent统一800–8000，前端却允许Agent/新项目到10000；原长仿写可50000每章、短仿写13000默认且上限500000总目标，没有现有DTO可直接承载；资料每份12k是当前Go安全设计，不是原50k referenceChunks合同。`backend/internal/novelcreation/validation.go:10–20,38–39,221–234`、`pipeline.go:842–843`；当前 `AgentPanel.jsx:31`、`NovelCreationPage.jsx:236`、`MaterialsPanel.jsx:40–43`；原ImportPage:366–380,588–631；ImitationShortDialog:115–121,226–279。
- **备注/关联存档**：原表单关联memos是多选shared notebook记录plain拼入参考；短仿写关联存档可选linkedSaveId，与当前必须先选项目、项目级noteIds和单styleID不同。`O/NovelImitationShortDialog.readable.js:43–64,145–162,308–324`；`O/NovelImportPage.readable.js:310–322,681–690`；备忘录session-key在 `O/NovelStudioPage.readable.js:1702–1710`。当前notes快照可保留追溯价值，但不能据此省掉原选择UI与字段关系（`backend/internal/novelcreation/automation_input.go:96–116`）。
- **自动写作不能证明六模式复刻**：原定时输入bookId/sessionId/model/intervalSeconds/dailyCap/failureCap；批量count/failureCap，当前Go的完整引用持久化和幂等改进是可靠性增强而非原版入口证据。`O/NovelSchedulerPanel.readable.js:85–91,121–142`；`O/NovelBatchDialog.readable.js:67–84`；当前 `backend/internal/novelcreation/automation_creation.go:8–72`。

## 五、精确桥接绑定与原生未知项（不猜函数体）

| 编号 | 已知精确调用点/参数 | 未解决边界 |
|---|---|---|
| U1 | 正常聊天 `O/NovelStudioPage.readable.js:430–458`：`can()`对象上拼接属性后`.call(t,e)`，参数s已完整可读；执行确认`:651–678`仅传session/model/reasoning。 | 拼接属性依赖han/gan未在23组件内恢复；不能把疑似novels_send_message作为本次确定绑定。原生候选同名字符串只在 `N:63–64`，候选文件开头明示不是定义/调用图。聊天如何抽取原创长短篇参数未知。 |
| U2 | 长仿写父调用 `O/NovelStudioPage.readable.js:519–536`；导入表单payload `O/NovelImportPage.readable.js:315–382`；tool标签architect可证。 | 属性四段动态拼接；Native候选novels_init_imitation（N:49–50）不构成该属性已解码或函数体已恢复。 |
| U3 | 短仿写 `O/NovelStudioPage.readable.js:863–885`，参数`{...e,sessionId,model}`；toolName imitation_short_run。 | 动态属性未解码；Native候选novels_run_imitation_short（N:51–52）仅为候选。11阶段UI不等于Native已证明11次调用。 |
| U4/U5 | 番外 `O/NovelImportPage.readable.js:414–438`；续写`:441–463`，原始对象与parentBookId/spinoffType或bookId/chapters明确。 | Native候选novels_init_spinoff / novels_import_chapters仅在N:55–58；桥接方法完整拼接与原生落盘/上下文流程未知。 |
| U6 | 文风`Tln/qln`调用 `O/NovelImportPage.readable.js:367–380,471–479`；Agent路由枚举 `crn`在原Agent:279。 | 未定位这些packed helper完整定义；只记录当前调用点、可见统计输出字段及UI文案，不生成新的8维算法、路由全集、私有提示词或原生执行保证。 |

## 六、应保留的已修后端能力与此次仍缺的原语义分开

现有持久派发/事务/租约fence、成功结果save_pending保存恢复、上游unknown禁止普通重试、账号归属、稳定创建key及编辑原会话保护，是独立可靠性能力，不因用户拒绝UI而丢弃；也不能以这些测试通过替代原UI验收。实际实现位置：`backend/internal/novelcreation/worker_store.go:19–21,107–109,262–280,332`；`worker.go:103–148,503–567,603,679–723`；`automation_creation.go:8–72`；`frontend/src/user/pages/novel-creation/AutomationPanel.jsx:21–77`。原生如何实现这些并未在本次证明。

仍缺的是上面逐项确定的原入口分流、原专用表单/默认值/多选、原输入DTO、长短篇独立产物/检查点，以及相同状态的原版画面。**仅改CSS或给旧通用输入再加几个可见字段都不能证明处理链对齐。** 该结论是六行映射与当前HTTP允许字段/Go白名单共同推导：`backend/internal/httpapi/novel_creation_routes.go:126–144`、`validation.go:84–118`、`pipeline.go:213–334,1103–1143`；不得在没有对应合同前以现有通用pipeline自定补足。

本次唯一写入为此审计Markdown；未修改/提交产品代码，未操作原窗口或当前运行包，未重启4001，未调用真实模型/TOS/数据库。原生截图证明、动态桥接完整解码、六模式真实生成与原生全函数体仍未验收。本报告供根重写已确认范围内的纠偏规格与实施计划，不表示任何新设计被用户批准。
