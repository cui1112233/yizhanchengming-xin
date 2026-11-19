# 小说创作逐段链路排查：2026-11-20

结论：**仍未与洛水全量对齐**。这轮已实际复现短篇3项控制失效；后台24条预设可编辑、生效、冻结旧任务是已完成项，不能替代原版全部动态提示词/状态/审稿规则的迁移。本轮排查，不修改生产业务代码或线上配置。

## 1. 对照基准与证据位置

- CM源码观察基点：`b1b34efd`，分支v88。共享工作区继续有其他任务提交；本文只按实际读取文件判断。
- 本地4001运行：`a0a43bb8`；公网观察版本：`2571b1be`；均为Go，入口`/assets/user-BVnCvltX.js`。源码、本地、线上不是一个SHA，不混称同版实测。
- 原版：`F:/软件/洛水2026/洛水2026`，已打开窗口标题洛水v2.9.4。
- 原生文件：`modules/novels.pyd`，SHA256 `b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d`，Craft本轮重新确认。
- **当前前端已换版**：`dist/index.html:7`是`index-DgTjggsH.js`，SHA256 `7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6`。11-17恢复文件头的`5278a6d1`不是当前版本。
- 本轮重新静态提取23个Novel组件，目录：`_local-archive-20261028/novel-chain-audit-20261120/frontend-current-294/`。`manifest.json`保存源SHA、原偏移、每份恢复片段SHA。恢复片段行号是格式化后的分析文件行号，不冒称原作者源码行号。
- 原生反编译/常量继续使用：`docs/reverse-engineering/luoshui-novel-native-20261119/`，包括`native/`、`readable/`、SPEC/STATE-SPEC/ROUTING-SPEC。不是完整原Python源码。
- Craft/REA本轮真实调用：Craft target_open/overview/feature_find；REA先解析旧恢复目录3文件，再对新提取的Agent设置/短篇仿写2文件解析。新REA父证据`ev_633030bc5555f699daca0f97327732503c6f9bc74a67e9502ef30aa4f5750f9a`，摘要`ev_9c67dbc8167407a38638d0e0c3f8679233c46ee999dc39c93c3134d1230071bb`。2文件解析0错误，其余22文件未进入本次REA分析，图覆盖partial，不能称整包调用图恢复。

## 2. 按步骤登记

| 步骤 | 本轮核查 | 结果 |
| --- | --- | --- |
| 1 版本和目录 | 原生SHA、当前前端入口、CM源码及运行SHA | 已查；发现旧前端对照过期并重新提取 |
| 2 六入口/表单 | 当前原版23组件、CM表单/路由、短篇现有字段 | 静态比对；未完成原版逐按钮运行验收 |
| 3 模型与Agent设置 | 分步模型路由、37维、写作要求、默认继承 | 分步路由和注入存在；新书/短篇继承仍有差异，见D4 |
| 4 提示词发布与冻结 | admin登记、snapshot、解析、请求、旧任务 | 现有已发布正文会实际使用；并非原版全部builder，见D5/D10 |
| 5 蓝图/大纲 | 长篇蓝图与短篇独立大纲、确认后继续 | 有独立代码/测试；原版全部输出合同未迁移 |
| 6 正文写作 | JSON正文、短篇分段、上下文资料 | CM链路存在；原版写前/正文区块尚不同 |
| 7 审稿与修订 | 长篇循环、短篇score/issues分支、控制开关 | **短篇3项失败探针复现**，见D1–D3 |
| 8 状态/伏笔/记忆 | 结算、增量合并、后写输入 | 有CM增量适配；StateValidator/完整字段/跨章输入未齐，见D6–D7 |
| 9 保存/恢复/取消 | 现有领域整包测试；前端导入/账号隔离/失败恢复测试 | 已有测试运行；不等于本轮线上真实模型成功→落库验收 |
| 10 原版UI/同输入产出 | 原版AX、截图与点击；真实模型预算 | AX可读；截图FrameArrived/capture超时、点击几何不可用；付费调用0，保留未验证 |

## 3. 已确认问题及修复验收条件

### D1 高：短篇忽略“最多修订轮数=0”

- 证据：`backend/internal/novelcreation/short_story_pipeline.go:301`仅按issues/字数判断进入修订，未使用Agent.MaxRevisionRounds；长篇`pipeline.go:378`实际读取该值。
- 本轮探针：MaxRevisionRounds=0，审核给出逻辑问题，仍发送ShortBodyRevision，共4次模拟请求并返回completed。
- 修复验收：预算0不发修订；预算N最多N轮；失败/提醒状态明确。历史任务应有版本边界，不能改已有请求指纹。

### D2 高：短篇关闭字数整理仍会因字数偏差调用修订

- 证据：同文件`:301,311`；长篇`pipeline.go:351`会检查EnableLengthNorm，短篇该分支未检查。
- 本轮探针：EnableLengthNorm=false、无审稿问题，仅正文偏短，仍发送ShortBodyRevision，共4次模拟请求。
- 修复验收：明确区分用户关闭的模型整理和必要的输入/成品长度验证；关闭时不因该开关覆盖而增加模型请求；长度不合要求如何提示必须写清。

### D3 高：短篇审核1分仍可成为完成品

- 证据：`short_story_pipeline.go:17`有Score；`:301`只看issues是否非空，`:350`设completed。没有在此流程使用Agent.PassScore；修订后也没有对应的重新审核调用。
- 本轮探针：PassScore=80，模型审核Score=1且Issues=[]，共3次模拟请求，返回completed。
- 修复验收：分数、问题严重性和通过状态使用明确一致合同；无问题但低分不能当通过；修订后重审或显式标为未验收，不能仅有修订文本就通过。

**D1–D3复现文件**：`_local-archive-20261028/novel-chain-audit-20261120/short-control-probe.go.txt`、`short-control-probe.log`。临时测试文件已从backend删除，避免把故意失败的排查探针混入业务测试。现有整包通过与探针3项失败可以同时成立：现有测试没有覆盖这些控制断点。

探针源码和失败日志另存入可提交目录`docs/verification/novel-chain-audit-20261120/`，不依赖临时目录才能追溯。重现方法：在已确认文件不存在的前提下，把go.txt临时复制为`backend/internal/novelcreation/audit20261120_control_probe_test.go`，从backend执行`go test ./internal/novelcreation -run '^TestAudit20261120ShortControls$' -count=1 -v`，保存结果后仅移除这份临时文件。当前预期是3项失败，不把其伪装为通过测试。

### D4 高：原版全局Agent入口与CM按作品设置仍不一致

- 当前原版恢复片段`frontend-current-294/NovelAgentSettings.readable.js:6`仅接modelOptions/isDarkMode，`:250`起读取配置且没有作品参数，`:469`说明设置永久生效。
- CM `frontend/src/user/pages/NovelCreationPage.jsx:353`没有projectId时显示“请先选择作品”；`AgentPanel.jsx:18`从当前作品settings读取，保存也写当前project。
- `backend/internal/novelcreation/short_story.go:191,200,204`新建短篇作品后读取该新作品设置并规范化默认值，没有传入已选作品Agent设置或已确认的账号全局设置快照。
- 影响：用户在别处配置了模型分工/37维/自定义要求，不能推断新建短篇会沿用；没有作品时原版可进入的设置在CM不可用。
- 修复验收：先恢复原版配置作用范围合同，再明确账号默认、作品覆盖、新建任务冻结的优先顺序；新短篇请求中逐项验证实际选中的配置，取消创建不产生假作品。

### D5 高：24条可编辑系统词不等于原版完整动态提示词

- CM `backend/internal/prompt/novelcreation/definitions.go:38`与`pipeline.go:323`仍是JSON章节正文合同。`admin_presets.go:20`登记的是这些当前步骤的默认词及已移植分流词。
- 原版`readable/novels--prompts--inkos--writer_prompts.md:152,159,377,407,413`包含写前检查、正文输出和不同结算格式约束；`SPEC.md:7`引到原生writer builder的输出格式、题材、核心规则等函数体位置。
- 不是“发布没用上”，而是“还没有完整恢复/移植并执行所有原版builder和解析分支”。直接把原词粘到JSON正文框会造成输出协议冲突。
- 修复验收：每个builder形成输入→匹配分支→最终system/user→输出解析样本；原文区块不能混入正文；支持的自家模型接同一协议，不复制原版服务配置。

### D6 高：状态与伏笔只做了CM字段适配，完整验证链未齐

- 原版`STATE-SPEC.md:8–14`对应读取旧状态、Settler、StateValidator、状态增量应用与保存的已恢复函数；`:24`明确完整字段待移植。
- CM `native_settlement.go:10,25`按事实id与伏笔thread等合并；`pipeline.go:465,489`接结算/记忆步骤，未实现上述完整StateValidator协议和原状态字段。
- 缺口包括hookId/候选准入、事实有效期、人物关系矩阵、支线/情感/数值账本等。不能用“旧事实不再丢”代表状态系统全部对齐。
- 修复验收：原始增量、应用结果和验证结果分别留证；重复增量、章节倒退、已解决伏笔重开、冲突与取消均有明确结果；沿用MySQL事务保护，不盲目复制原版中途写文件顺序。

### D7 高：后写校验函数支持的资料，生产入口没有完整喂入

- `local_checks.go:35`可接受ExistingTitles、PreviousChapters、FatigueWords、FirstPerson、Protagonist。
- 实际`native_review_merge.go:18`只传Title；`pipeline.go:435–439`旧标题为空，跨章状态明确为not_checked_missing_full_text。
- 影响：UI打开后写校验不能代表跨章重复、历史标题、人称等所有检查都执行过。
- 修复验收：冻结真实的完整前章/历史标题/人称资料和来源版本；缺资料逐项显示“未检查”，不能拿尾文冒充完整正文；新增开销需有字数范围。

### D8 中：长篇修订择优与提前退出还没按原版补齐

- `pipeline.go:383–423`按轮数循环；`betterAudit:1193`按是否通过、分数更高选择，没有传目标字数或正文长度。
- 原版`SPEC.md:28`引到length_in_range、best及无净改善退出的反编译证据。具体原阈值仍需小函数补证，不能凭NET_IMPROVEMENT_EPSILON字符串自行定值。
- 修复验收：长度更合适但分数相同、分数高但严重失长、无新正文、连续无净改善、取消、回退旧最佳版本均留明确快照和退出理由。

### D9 中：AI味与文风统计算法仍有明确差异

- `local_checks.go:309–322`明确是CM本地表达指标，original_algorithm_verified=false；AgentPanel已如实说明，不是零功能按钮，但不能叫原版AI味检测已实现。
- `styles.go:29–46`用Unicode码点、单换行分段和中英文标点计句，只保留少量统计字段。原版当前bundle的sentenceLengthStdDev附近仍包含段长范围、词汇多样性、常用句式、修辞特征等字段；旧版公式完整记录在`2026-11-17-luoshui-novel-style-statistics.md:27–39`，当前混淆表达仍需逐式复核，旧公式不能直接宣称新版本相同。
- 修复验收：当前版原函数逐式恢复，空文、空行、英文标点、emoji、反复句式等同输入得到相同字段；AI味原算法没恢复前保留差异说明。

### D10 中：实际用了后台新版正文，但调用日志里的版本号含义不清

- `system_presets.go:8`保存发布Version/Revision，`:55`解析时只替换Definition.SystemPrompt；Definition.Version保持输出协议版本，属于历史兼容需要。
- `pipeline.go:737`把调用传入的promptVersion记成ModelCallMetadata.PromptVersion，它仍是代码协议版本，不能直接当管理员发布版本。
- 影响：后台发布V2后实际用了V2，调用日志仍可能显示1，容易再次误判“提示词没生效”。任务冻结正文是实际证据，二者不要混淆。
- 修复验收：分开记录协议版本、后台发布版本/修订号、正文摘要；原生修订还要记录基础词和分流词各自版本。不通过改协议版本来解决日志问题。

## 4. 已有能力与本轮验证边界

- 37项审稿维度/自定义要求实际通过`agent_requirements.go:79`、`prompt/novelcreation/user_requirements.go:15`及`review_dimensions.go:44`进入相关请求；不能一概称全部按钮都是假的。新任务冻结、禁用不注入、按类别注入和修订保留上下文有既有请求级测试。
- 分步骤模型路由实际在`pipeline.go:705–737`选择自家可用模型。短篇步骤有映射；D4是新作品配置继承问题，不是所有路由都没用。
- 已保存任务的幂等、取消、未知结果不盲目重发、章节/状态事务等已有免费测试，本轮小说领域及prompt包整包通过。
- 原版窗口AX能读到首页导航及小说入口，但截图/坐标读取失败，没有把没点击成功的步骤登记为成功。
- 前端测试是本机浏览器+受控API替身，不等于原版软件/线上真实模型对照。命令`node --test src/user/pages/novel-review-dimensions.browser.test.js src/user/pages/novel-creation.browser.test.js src/user/pages/NovelCreationPage.source.test.js`结束：**68项，56通过、12失败、0跳过，退出1**。日志见同目录`frontend-tests.log`；不能记为前端整包通过。
- 本轮新增真实文本/图片/视频生成请求：0。没有新增付费授权，原/CM同输入真实生成质量依旧未验证。

### 前端12项失败如何处理

这是额外验收缺口，不自动认定为12个产品缺陷。受控浏览器保存每例`result.html`、`requests.json`、`screenshot.png`和`protocol.json`，目录在`_local-archive-20261028/novel-task7-browser/`；本轮错误摘要另存`novel-chain-audit-20261120/browser-failure-details.json`。

| 失败区域 | 当前证据与判断 | 下一步 |
| --- | --- | --- |
| Agent表单布局1项 | 失败首先是sidebar width，测试`novel-creation.browser.test.js:211`要求260；实际CSS`:531`覆盖为240。没有执行到后面的八模型/零值保存断言，不能据此说那些保存都失败 | 按当前批准的原版布局更新几何基准，然后单独跑后续保存断言 |
| 草稿隔离、资料自动化、四模式3项 | 失败包含missing input/timeout；测试`:199–218`仍通过旧内嵌参考文本定位输入，当前短篇使用独立弹窗 | 改为真实弹窗操作，再测试跨账号、跨作品、回包及实际DTO；不能简单删除断言 |
| 工具工作区3项 | C5a工具切换超时、内部补建会话超时、无作品限制文案不符 | 分别复核当前无书对话/有书工具状态；与D4全局设置入口差异一起形成合同 |
| 文风维护1项 | timeout style；当前保存/模型维护恢复是否仍通尚未由这项失败证明 | 当前入口重现创建文风、发起维护、丢回包恢复同run，禁止重生成 |
| 定时恢复1项 | missing/disabled button intent schedules，失败发生在测试目标按钮处 | 按当前入口复核计划恢复，验证原key及旧响应不冒充新意图 |
| 原计划编辑3项 | timeout edit original session / original failure / original late | 保留本轮失败，按当前UI验证s2引用、查询失败禁保存、迟到查询不串新作品 |

37维两项浏览器测试本轮通过；上述失败不推翻独立已通过的后台版本冻结/请求注入测试，也不允许反过来用领域包通过掩盖前端未通过。

## 5. 修复顺序与登记

1. **优先D1/D2/D3**：短篇控制与通过状态，先把三项失败探针变为修复后的行为测试；同时补修订后重审和取消边界。
2. **D4/D10**：全局/作品配置继承与调用版本可追溯，确保用户的选择确实进入新任务。
3. **D5**：按原builder的输入/分支/输出合同逐个迁移；每个合入后保存最终请求样本，避免再只抄表单。
4. **D6/D7/D8**：状态校验、完整跨章输入及审稿循环退出/回退。
5. **D9与UI/真实生成验收**：算法证据补全、原版窗口交互恢复，再逐入口正常/空/重复/错误/取消及同输入结果对比。

- [x] 固定当前原件身份，更新前端恢复目录。
- [x] 查请求拼接、分步模型、保存/恢复与短篇控制。
- [x] 3项短篇失败探针实际复现并保留日志，未改业务代码。
- [x] 小说/prompt现有整包测试通过。
- [ ] 原版所有按钮实际点击与视觉验收：截图/几何故障阻断。
- [ ] 原/CM真实模型同输入验收：本轮预算0。
- [ ] D1–D10修复：本轮任务为排查，不能登记为已修好。
