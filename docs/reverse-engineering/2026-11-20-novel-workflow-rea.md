# 小说创作：继续 REA 核查步骤、使用与功能

本轮为取证与差异登记，不修改业务逻辑，不发付费模型请求，不部署。CM 读取基点为 v88 / 0930d818；共享工作区其他改动不在此次范围。D1–D3 的短篇审核控制已在 70381602 修复，本报告不沿用旧的“仍未修”状态。

## 对照身份及方法

- 原版 `F:/软件/洛水2026/洛水2026/modules/novels.pyd`，SHA256 `b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d`。
- 当前前端 `dist/assets/index-DgTjggsH.js`，SHA256 `7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6`。恢复片段是检查材料，不是原始开发源码。
- 实际调用 REA：AgentSettingsStore.load、pipeline_runner._settings_store、AgentSettingsStore.review_cycle_config，短篇仿写run/_produce，以及当前 NovelStudioSidebar / NovelStudioPage / NovelImportPage 静态分析。MCP 中途出现 Transport closed，后两项改用 REA 6.5.0 命令行，成功返回并保存 JSON；不是用自写字符串扫描冒充 REA。
- 原始完整 CLI 证据在 `_local-archive-20261028/novel-rea-workflow-20261120/`；可提交的伪代码、来源摘要和摘录在 `docs/verification/novel-rea-workflow-20261120/`。前两个 MCP 结果保存为提取视图，不冒称完整导出包。
- Ghidra 伪代码属于反编译结果；旁边的常量注释由已有 pool 绑定及原字节范围生成。静态调用仍有间接目标未解析，不等于全部运行链已复现。

## 1. Agent 设置：入口 → 读取 → 审核配置

| 检查点 | 原版证据 | CM 当前行为 | 判断 |
| --- | --- | --- | --- |
| 无作品打开设置 | NovelStudioPage.readable.js:416–418 直接切换 agent-settings；:1640 起只传模型与主题等属性 | NovelCreationPage.jsx:353 先判断 projectId，无作品显示“请先选择作品” | 明确不一致 |
| 设置作用范围 | NovelAgentSettings.readable.js:244–270 无作品参数加载；:469 说明永久保存 | AgentPanel.jsx:16 读取作品 /settings，:35 写回同一作品 | 全局与作品级不一致 |
| 原生读取 | settings-0.pseudocode.annotated.c:39 引用 FileRepository，:81 引用 merge_agent_settings | CM 从 MySQL 作品设置读取并规范化 | 存储技术可不同，配置作用范围不能因此丢失 |
| 审核配置读取 | review-cycle-config.pseudocode.annotated.c:60 使用 load；:106 默认设置；:122 reviewCycle | short_story.go:191 新建作品，:200 读取新作品设置，:204 规范化并冻结 | 新短篇没有沿用先前配置的来源 |
| 执行器持有设置 | settings-2.pseudocode.annotated.c:37 等出现 AgentSettingsStore 构造绑定 | 当前任务快照冻结实际取到的设置 | 有冻结机制不等于取到了用户先前选择 |

REA ID：load=`ev_19b8180306d335993c1197d95490f7477fb945146a2e8d66f0906f7e22aae591`；settings_store=`ev_a8aaeb2de17b1b3fbf0a531422e4c77962bc3812219a61a2ca1f7713b34f1b9e`；review_cycle_config=`ev_3044484ffce83756e9ee496767b2ac8041915dfac0fb099b57eff5d7853c73e1`。最后一个函数有 70 个未解析调用目标，已知取值不能扩张成所有字段转换都已完全还原。

可验证修复规则：账号默认配置 → 用户明确选择的作品覆盖 → 创建任务时冻结最终配置。无作品也能编辑账号默认；新短篇实际请求应携带这份最终配置。旧任务继续用旧快照；甲账号不能读取乙账号配置。此处为 CM 待实现合同，不能声称原版具有多账号 MySQL 层。

原版自身边界问题：NovelAgentSettings.readable.js:151 用 `Number(maxRetries) || 3`，0 会变成 3；:152–154 也存在类似默认回退。CM 已支持合法 0，保留此修复，不能为“复刻”复制缺陷。

## 2. 短篇仿写：表单真实传递了什么

当前原版 NovelImitationShortDialog.readable.js:142–163：

- 输入：参考原文、可选标题、总字数、章节数、补充要求、读者偏好、灵感标签、关联存档、知识库、所选备忘录。
- 边界：:90–91 用去空白字数检查至少 100；:117–123 字数规范为默认 13000、最高 500000，章节默认 1、最高 50；:128–137 重复文件跳过、读不到文件提示；取消文件选择直接返回。
- 输出：:151–161 生成结构化提交对象，不是直接调用写作模型。备忘录文本实际拼入 referenceText；targetChapters 与 chapters 都是数字，不能误当续写的章节数组。
- 提交：:163 发 submit 后清空表单并关闭。父组件再附加会话和当前模型。表单提交成功不代表后端写作成功。

CM 已对应参考原文至少100个非空白字符（short_story.go:99）、默认13000字和1章（:105–109），并限制最多500000字与50章（:111–112）。超限时CM报错，原版前端会钳制到上限；这是明确的边界行为差异，不能称完全一样。CM 对重复提交保留原请求核对和幂等标识（ShortStoryDialog.jsx:30–32），原版表单关闭/清空也不能替代服务端幂等验收。

原包 imitation_short_runner 的 Constant 251（原文件字节3957779..3958078）自述流程为“deconstruct reference → pick direction → build blueprint → run 7-stage short fiction pipeline”，并说明中间材料按story目录保存供右侧面板显示。这是包内功能说明证据，不能单独当成实际调用次序。原生 run/_produce 的补证详见下方。

CM shortSource（short_story_pipeline.go:60–90）把参考原文分块分析，再合并为 content 字符串；:241–246 将它交给通用短篇大纲。需要继续对照原版专用仿写阶段，不能因为参考原文已进入提示词就认定整条仿写等价。

原生 run 补证：REA `ev_6b44caf2989e697505e613bd996c893c4198175cc8b38313ba4bcf371c5e55ca`，函数 `0x1800cabe0`。`imitation-short-run.pseudocode.annotated.c:179`包含空参考原文错误；:552–858 准备 shorts_root 下的 imitation / outline / reviews / drafts / final；:1118 引用 `_produce` 并执行调用路径；:1202–1234 异常分支写 status.json 的 failed/error；:1399–1597 读取状态及 final/full.md，构造含 storyId、title、finalMarkdownPath、resumed 的返回。具体恢复判定仍须按条件分支验证，不能把所有已存在文件都视为已成功。此函数有130个未解析调用目标，内部阶段次序由下方 `_produce` 补证，不是本函数单独证明。

原版专用提示词已拆成四段输入/输出/边界/失败合同，另见 `docs/specs/novel-creation-alignment/2026-11-20-short-imitation-stage-contracts.md`。明确缺口包括六维拆书、三个方向、七项仿写蓝图及专用编辑意见单，不是再加一个表单选项便能补齐。

### 原生 `_produce` 顺序已补证

REA `ev_a31e4b4aa970e79ea15d8e72a9743752e14d529b17a4c43d3224d844d2ecc52c`，函数 `0x1800cd2d0`，完整返回已存原始 JSON；可提交伪代码为 `docs/verification/novel-rea-workflow-20261120/imitation-short-produce.pseudocode.annotated.c`。

1. :179–221 先检查取消，调用 `_resume_or_run_deconstruction`，成功后再检查取消并调用 `_resume_or_run_directions`。
2. :233–355 读取 `selectedDirection`（tuple Constant84，原字节3953786..3953807）；去空白后为空才调用 `_extract_first_direction`，随后把拆书结果和方向传入 `_resume_or_run_blueprint`。不能声称每次都要求用户人工选三个方向。
3. :728–772 引用 `build_imitation_short_writer_addendum` 和 `_extract_anti_plagiarism`，准备仿写附加资料；helper完整拼接仍待补证，不据名称宣称最终模型请求已捕获。
4. :920–948 调用 `_maybe_pause_after_blueprint`；结果为真时进入返回路径，未暂停才继续。暂停的触发设置及恢复请求还需该helper补证。
5. :949–1134 构建 ShortFictionRunner、增强上下文和偏移事件转发，传入 cancel_checker / review_config / nested 等参数，再调用内部 run（字段来自tuple Constant120，字节3954344..3954400）。
6. :1148–1217 再检查取消、读取内部返回的 finalMarkdownPath；:1912–2038 包含最终文本读取和 `_run_editorial_review` 分支。没有可用最终文本存在跳过分支，不能把“函数有终审”扩成“每条路径均终审且失败会阻塞完成”。终审通过标准与失败状态仍待helper补证。
7. :1265–1362 有状态保存；:1598–1689 返回拆书、方向、蓝图、编辑报告路径等中间材料。原版使用文件，CM可用MySQL实现等价保存，不复制文件架构。

这个函数的174个调用目标未解析，已证明的是以上命名方法调用和局部控制路径；没有证明所有被调函数行为或真实模型输出。由此可确认 CM 的通用分析字符串尚未覆盖专用拆书/方向/蓝图衔接；不能把已接通通用短篇生成当完整仿写对齐。

## 3. 入口与使用步骤：不能只核对菜单文字

当前 Sidebar.readable.js:44 起分别声明 launch-mode、switch-workflow、open-session、open-book、create-book-session、open-import、open-style、open-agent-settings、open-scheduler、open-dashboard、open-memo 等事件。StudioPage.readable.js:289–295 切换模式会清理不匹配会话；:404–418 区分模式切换、导入页与设置页。

| 入口/使用动作 | 本轮追踪到的范围 | 仍未通过的验收 |
| --- | --- | --- |
| 长篇原创 | 原版模式/会话切换；已有 writer builder 原生证据 | 全部 builder → 输出区块 → 保存与状态闭环 |
| 短篇原创 | CM 独立大纲、正文、审核控制及任务快照 | 原版全部专用阶段与相同输入运行结果 |
| 长篇仿写 | 原版独立导入入口与模式，已有初始化绑定 | 方向选择、确认、恢复与每步提示词逐项运行 |
| 短篇仿写 | 当前表单字段、原文校验、备忘录拼接、父提交；原生 run 继续取证 | 分析结果是否结构化进入后续方向/蓝图/审稿，见后续原生补证 |
| 番外/续写 | 当前独立 workflow / 导入入口；不得混用 chapters 数组和章节数量 | 关联正典、导入章节、模式专用状态和失败恢复 |
| Agent 设置 | 原版与 CM 的实际读取/保存作用范围 | 全局继承断点尚未修复 |
| 风格/素材/定时/看板/备忘录 | 原版存在独立事件，部分读取位置已恢复 | 每个弹窗、取消、重复、跨账号及重启后行为未逐项运行 |

当前前端 REA ID=`ev_85d200b39b6596c45ca334bd7e66289ed310c631c1adf0de1256b39a837eea55`。本次选择了 3 个文件，另 21 个清单文件不在本次静态图范围，覆盖 partial。不可据此称“每个按键都有效”。

## 4. 后续修复与验收登记

1. 高优先：实现账号默认 Agent 设置、新作品继承和任务最终快照；分别断言模型分工、37维、自定义写作要求、审核阈值及开关真正进入请求。
2. 高优先：对短篇仿写专用阶段建立输入/输出和恢复点；明确取消后哪些材料已保存，不能把一个通用分析字符串当完整阶段结果。
3. 高优先：继续已登记的 writer 动态 builder、正文区块解析、StateValidator 与后写历史输入，沿用已有 SPEC/STATE-SPEC，不另造规则。
4. 每步测试正常、空值、重复、错误、取消；错误提示要区分未配置、校验失败、模型失败和已取消。离线请求断言与真实模型结果分别登记；本轮付费调用 0。

本轮结论：实际 REA 取证有新增，配置作用范围与短篇继承的断点有源码支撑；整体功能、UI 和运行结果仍未全量对齐。
