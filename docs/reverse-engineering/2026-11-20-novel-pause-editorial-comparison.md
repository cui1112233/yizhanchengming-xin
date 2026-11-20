# 洛水小说创作继续对照：确认、文件编辑与终审

实施后续：上述排查后已完成新协议蓝图暂停、文件编辑、质量状态分离与当前模型确认，见[实施回执](../verification/2026-11-20-novel-blueprint-resume-repair.md)。本文保留排查时的原始差异，不代表这些项目仍全部缺失；原版恢复与Writer补证见[补充证据](../verification/novel-resume-evidence-20261120/FINDINGS.md)。

日期：2026-11-20。范围：洛水 v2.9.4 的短篇仿写流程，对照 CM 的 cf174fa4 实现；本轮只取证、登记差异，没有修改业务实现，也没有消耗模型额度。

## 证据基线

- 原版目录：`F:/软件/洛水2026/洛水2026/`。
- `modules/novels.pyd` SHA256：`b5b22086c64851e0e94fc70a509cddbaf4ec144ac5beedc64e3a9a8caa40e57d`。
- `dist/assets/index-DgTjggsH.js` SHA256：`7c2520565126cf49615a968a6bceefc5cfcf3e032f9858a9570adeab546336a6`。
- 当前前端恢复件：`_local-archive-20261028/novel-chain-audit-20261120/frontend-current-294/`。本文前端行号是格式化恢复件的行号，不是原开发源码行号。
- 本轮调用了 Craft Anything 的调查、会话、计划和假设记录。REA MCP 原生函数查询返回 `Transport closed`，因此改用已安装的 REA CLI 6.5.0 / Ghidra 12.1.4 进行相同目标的只读分析。没有把失败的 MCP 查询当成成功证据。
- 原始 REA JSON：`_local-archive-20261028/novel-rea-followup-20261120/`；可审阅伪代码、前端摘录和索引：`docs/verification/novel-rea-followup-20261120/`。
- 原生伪代码来自编译产物，不是恢复出的可运行 Python 源码。注释中的常量绑定由已恢复模块池地址、常量索引和文件字节范围推导；间接调用仍有未知项。

## 1. 生产二次确认的含义和暂停位置不同：已确认

原版设置页面 `NovelAgentSettings.readable.js:355–361` 将 `requireSecondaryConfirm` 说明为“大纲/蓝图/基础设定生成后暂停，可在右侧修改文件后再继续”。它不是“每一次模型调用前询问”。

原生 `_maybe_pause_after_blueprint`，地址 `0x1800d76c0`：

- REA：`ev_1549b301da78ef4f43576fdc8444f8d9e68a2b3f98df2772c592f6f6e545963b`。
- `pause.pseudocode.annotated.c:58–99`：读取 `require_secondary_confirm`，为假时返回 None。
- `:122–178`：调用 `is_confirmed` 检查 `blueprint`；已确认时返回 None。
- `:332–478`：保存运行载荷，并以 `blueprint` 写入等待状态。
- `:504–547`：构造“创作蓝图已生成，可在右侧查看或修改后确认继续生成正文”，返回 `awaitingConfirmation=true`、`checkpoint=blueprint`。
- 上轮 `_produce` 证据 `docs/verification/novel-rea-workflow-20261120/imitation-short-produce.pseudocode.annotated.c:918–950`：此 helper 返回真值即返回结果，否则才继续进入 `ShortFictionRunner`。这说明暂停位于外层仿写蓝图之后、内部短篇生产之前。

CM：`frontend/src/user/pages/novel-creation/AgentPanel.jsx:59–64` 分成“大纲/蓝图生成后暂停”和“每次后续模型调用前确认”。`backend/internal/novelcreation/short_story_pipeline.go:330–367` 在通用短篇大纲生成、审核及可能修订之后暂停；不是原版外层专用仿写蓝图的暂停点。`short_imitation_pipeline.go:295–312` 保存专用蓝图后直接返回供后续流程使用，没有这个暂停。

**影响**：用户点同名设置，停下来的阶段、可修改的资料和后续调用数量不同。已有 CM 逐次调用保护有实际用途，不能直接改字段含义而改变旧任务保护。

## 2. 原版右侧文件编辑不只是预览：已确认

- `NovelStudioShortPanel.readable.js:80–119`：加载选中文件后设置 `readOnly:false`，带作品 ID 和文件 key。
- `:120–161`：保存调用传递作品 ID、文件 key、编辑内容；失败展示后端 message 或“保存失败”，成功更新内容、提示“已保存”并通知父页面。
- `NovelPanelArtifactView.readable.js:29–47`：开始编辑、取消恢复草稿、保存发送草稿，收到新内容时按编辑状态处理。
- `:55–108`：编辑、保存、取消、textarea 和 saving 状态；null 文件显示“文件不存在”，空字符串不是文件不存在。
- 上述三个面板的 REA 静态分析证据：`ev_9deadbcfa15aa57fbce81c92c595a44b138505ce3401545a027e88610a0fb372`。这里只分析所选三文件，不能覆盖全部页面。

CM：`frontend/src/user/pages/novel-creation/ShortStoryPanel.jsx:24–29` 只为通用短篇大纲提供编辑，其他阶段文件及正文预览只有关闭、下载。**专用蓝图编辑和阶段文件编辑未对应**。这也造成上面的“修改文件后继续”没有完整对应。

边界：原版保存后的文件具体如何重新进入每个 builder，仍需继续追踪后端读取和恢复函数。前端保存接口调用成立，不等于已证明所有后续阶段都会读取新文件。

## 3. 编辑终审与任务成功被混成一件事：CM 已确认有额外规则

原生 `_run_editorial_review`，地址 `0x1800d31d0`：

- REA：`ev_2d67b205f91ed552f02c9da30ea9c71e03f58e6781a1e5b5b4eba4fb08bdaaca`。
- `editorial.pseudocode.annotated.c:47–190`：通过 auditor、专用 system/user builder 发起调用。
- `:206–278`：将返回内容写入 `reviews/editorial.md` 并返回该内容；该 helper 中没有观察到按评级改变任务状态的分支。
- 原生 `_produce:2028–2050` 调用 helper 成功后，经 `LAB_1800cf447`（`:1239`）进入收尾，`:1276–1298` 构造 `status=completed`。这条可见路径没有解析“可过审/需修改/打回”再决定 completed。

CM：`backend/internal/novelcreation/short_imitation_pipeline.go:322–358` 要求结构化 rating、标题、简介、修改清单；保存报告后，只要 rating 不是“可过审”就返回错误。因此“稿件生成完成但编辑建议修改”会表现为任务失败。

结论强度：**CM 的评级门槛是明确实现；原版可见 helper 和调用路径表现为报告保存后完成。** Ghidra 对 helper 的间接调用仍有 28 项目标未知，不能扩大为原版所有模式永远不做评级检查。没有真实模型调用作运行对照。

## 4. 确认继续时的模型来源：前端差异已确认，原版后端采用规则待确认

- `NovelStudioPage.readable.js:635–655` 的继续请求包含当前选择的 `model` 和按当前模型计算的 `codexReasoningEffort`、`claudeReasoningEffort`。
- `NovelCheckpointCard.readable.js:14–41` 提供“确认继续”“稍后”，处理中禁用重复操作。
- CM `backend/internal/novelcreation/short_story.go:332–367` 验证版本、等待状态、大纲，创建继续任务时沿用 `source.ModelID`、`source.ThinkingLevel`。

这证明请求合同不同，**还不能证明原版恢复后实际切换模型**。需定位原版 resume handler 和模型路由，不能先把 CM 改为任意当前模型。

## 5. 可验证的修复规格

以下是对照后的候选实施要求，未在本轮实现。与原版尚未确认的部分明确标出，不拿 CM 自定规则冒充原版行为。

| 项目 | 输入 | 输出 | 边界 | 失败提示 |
| --- | --- | --- | --- | --- |
| 仿写蓝图暂停 | 专用蓝图、生产确认开关、确认记录、任务版本 | 蓝图保存后返回等待状态；确认后继续内部生产 | 关闭开关或蓝图已确认不重复暂停；旧任务逐次调用保护继续保留；新协议单独版本化 | 保存失败不得假装等待成功；过期版本明确要求刷新 |
| 阶段文件编辑 | 作品 ID、受允许文件 key、内容、版本 | 保存并回读同一版本，随后对应阶段读取新内容 | 不允许任意路径；运行中编辑如何生效需明示；取消不保存；重复保存不产生重复任务；空内容规则按文件类型验证 | 读取不存在、无权限、冲突、保存失败分别显示，保留草稿 |
| 终审结果显示 | 已保存正文、编辑报告、报告评级 | 分别显示“生成完成”和“编辑需修改”等质量结论 | 传输失败/报告无效仍报错；有修改意见不能丢失正文；不要自动重生成全文；原版其他模式评级规则待取证 | 明确是生成失败还是编辑意见，提供报告和正文入口 |
| 确认继续 | checkpoint、任务版本、模型/思考选择 | 恢复到准确阶段，重复点击只创建一次后续任务 | 原版后端是否接受新模型待证；配置消失不得静默换模型；现有任务快照不可污染 | 模型失效、版本变化、已确认、恢复失败分别说明 |
| Writer 附加约束 | 仿写补充提示词、消息 role、原 writer system | 只注入原版允许的写作请求 | 原版 `_inject` 谓词继续分析；不得把约束加到拆书/审稿后冒称一致 | 缺必要预设须明确停止，不能用空文本继续 |

## 验证边界与后续顺序

1. 本轮原生与前端静态查询成功，不代表 UI 像素或真实模型输出验收。
2. 通过桌面工具读到了“洛水 v2.9.4”及“小说创作”入口；点击返回 `coordinate input geometry is unavailable`，重新观察窗口为 minimized。因此本轮没有完成原版小说页点击验收，也没有截图对比通过。
3. 不把原版用户剧本/作品内容写入证据；仅保存脱敏窗口观察记录。
4. Writer `_inject`（`0x1800d5c20`）本轮确实发起了 REA CLI 查询，但 330000ms 后返回 `provider_timeout`，没有得到函数证据；失败已纳入索引，不能把它标为已完成。下一步先核查 Writer 注入条件、恢复读取已编辑文件和模型路由，再按新任务版本修复蓝图暂停和阶段文件编辑；之后分开验证正常、空值、重复、错误、取消。真实模型测试需单独有额度授权。
5. 未执行部署、未改业务代码；不能据此宣称“已一比一对齐”。
