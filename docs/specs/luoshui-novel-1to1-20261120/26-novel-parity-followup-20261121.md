# 小说创作本轮对齐登记（2026-11-21）

基线：v88 bc9c9ac6；本轮只修复已证实差异。原版 2.9.5 的16个小说页面状态见 `docs/reference/luoshui-ui-v2.9.5-20261121`。恢复前端源码属于此前归档，版本不能混称同一版本；数值规则以恢复源码和当前可见字段交叉核查。

## 执行顺序

1. [x] 核对16个原版状态、现有 React 入口与 Go 路由，执行基线测试。
2. [x] 复现短篇自填小数直接发送、番外关闭后成功回包留恢复记录。
3. [x] 修复并对同一测试做修改前后对照。修改前实际发送 `[12345.6,2.6]`；番外成功后仍有 intent；修改后两项通过，另补明确拒绝/未知回包/双击测试通过。
4. [x] 检查提示词和引用参数进入实际生成步骤的代码及测试（静态与隔离测试，不代表真实上游验收）。
5. [x] 从提交d8eda3bf独立导出并构建Go embed，4014验证后更新4001；六入口可打开，37项审稿维度加载且保存按钮可用（未点击保存）。
6. [x] 登记本轮结果及仍未验证项，以下缺口不标成完成。

## 已批准对齐范围内的修复规则

- 短篇目标字数：有限数且至少1000，四舍五入，上限500000；否则13000。章节：有限数且至少1，四舍五入，上限50；否则1。原证据：`source/frontend/NovelImitationShortDialog.readable.js:105-112`。界面选项见原版 `states/novel-short-imitation.json`。
- 番外创建提交后关闭表单只是退出查看，不撤销已提交任务。迟到成功仍更新作品事实并清除已完成请求；不抢走用户当前页面。迟到明确拒绝解除恢复锁并保留草稿；未知结果继续保留原请求编号，避免重复创建。现有根因：`SpinoffDialog.jsx` 在清除恢复记录前就因关闭表单而 return。
- 保留我们自己的授权模型、统一主题与后端 prompt 管理；不添加原版平台渠道。

## 验证边界

浏览器测试使用隔离页面和模拟服务返回；Go测试使用测试运行器。它们不能证明上游模型真实生成成功。本轮不消耗模型额度、不启用生产自动任务、不公开部署。

## 16个原版状态逐项映射

以下是已定位的代码，不是16项业务全部等价的结论。原版图像路径统一为 `docs/reference/luoshui-ui-v2.9.5-20261121/screenshots/<状态>.png`。

| 原版状态 | 我们的入口与代码 | 参数去向/输出 | 验证与缺口 |
| --- | --- | --- | --- |
| novel-home | `StudioChat.jsx` 长篇原创 | `/studio-sessions` 讨论→建议→显式确认→作品 | 原创确认/刷新测试通过；真实模型未跑 |
| novel-short | `StudioChat.jsx` 短篇原创 | 独立短篇会话→确认短篇→大纲/正文 | USAGE短篇讨论与确认测试通过 |
| novel-long-imitation | `LongImitationDialog.jsx` | `/long-imitation`→`long_imitation.go`→拆书/方向/蓝图 | 本轮原默认、上传、丢回包、账号切换等9个测试通过 |
| novel-short-imitation | `ShortStoryDialog.jsx` | `/short-stories`→`short_story.go`→全篇协议 | 默认13000/1；小数取整已修；关联文风/偏好/标签进入请求测试通过 |
| novel-spinoff | `SpinoffDialog.jsx` | `/spinoffs`→`spinoff.go:45`→正典/基础设定 | 原版字段和类型已有；迟到成功/拒绝/未知处理已修并测试 |
| novel-spinoff-types | 同上 | sequel/prequel/parallel/if | 现有USAGE类型传参测试通过 |
| novel-continuation | `ContinuationImportDialog.jsx` | `/chapter-imports`→`chapter_import.go:43`→正式章节 | C1b导入/读回/恢复/取消/旧回包测试通过；导入本身不生成 |
| novel-styles | `MaterialsPanel.jsx` | 文风统计、资料选择；AI分析走显式 maintenance | 有现成实现；一项旧场景未通过，不能据此称完全一致 |
| novel-import | 长篇仿写专用弹窗；旧作品资料导入仍独立保留 | TXT/DOCX解析→正文/选章→明确建书 | 原版默认仿写入口已接；历史旧资料场景需按新入口复核 |
| novel-agent | `AgentPanel.jsx` | `/agent-settings`或作品`/settings`→新任务冻结配置 | 账号/作品分开；现有零值测试的旧布局断言失败，待拆分复核 |
| novel-agent-review | `ReviewDimensions.jsx` | 后端37项目录→reviewDimensionIds→审稿系统要求 | 默认29、全选/全不选、保存重开测试通过 |
| novel-agent-review-bottom | 同上 | 32–37项及同人维度与其他维度同样保存 | 目录、选择保存通过；真实审稿结果未生成 |
| novel-agent-prompts | `AgentRequirements.jsx` | reviewPrompts/customWritePrompts→Agent快照 | 新增/停用/分类/保存失败保留等测试通过 |
| novel-scheduler | `SchedulerPanel.jsx` | 三参数→`/scheduler/start`，暂停/恢复/停止各自接口 | 专用调度器浏览器测试通过；旧管理场景未通过，不能混成同一入口 |
| novel-dashboard | `NovelCreationPage.jsx` dashboard分支 | 作品`/dashboard`实际章节统计 | 无作品不伪造数据测试通过；原版全状态未对完 |
| novel-memo | `NotesPanel.jsx` | 作品`/notes`→草稿引用/新任务快照 | 已有编辑与引用；原版无作品备忘录范围仍待核对 |

## 提示词与选项没有停在界面层的证据

- 应用启动接入后台预设目录：`backend/internal/app/app.go:379`。新任务保存已发布提示词版本，后续执行从冻结版本读取：`backend/internal/novelcreation/system_presets.go:20`；缺失或空预设明确停止，不调用该步骤模型。
- 统一目录包含全篇首稿、补写、文字审稿、第二稿、包装、长篇、仿写、正典等：`backend/internal/prompt/novelcreation/admin_presets.go:19`。
- 短篇的读者偏好、补充信息、标签、文风、备忘录和关联作品进入 brief：`backend/internal/novelcreation/short_story_pipeline.go:56`。全篇正文将 brief 加入生成上下文：`short_full_pipeline.go:133`。
- 正文、审稿系统要求与自定义要求组合后送给实际 provider：`short_full_pipeline.go:90`。Go `TestShortFullPipelineKeepsFrozenPromptsCustomRequirementsAndModelRoutes` 验证冻结预设、自定义要求和模型路由的作用。
- `backend/internal/prompt/novelcreation/short_full/provenance.json` 记录此前恢复提示词的目标SHA及适配内容；该SHA不等于本轮2.9.5二进制。还不能称新版全部提示词逐字一致。

## 本轮测试登记

- `go test ./internal/novelcreation ./internal/prompt/novelcreation`：通过；未提供专用DSN的真实MySQL测试会跳过，不是持久化全量验收。
- 新增浏览器回归4项及数值规则单测1项：通过。覆盖小数输入、关闭后成功、关闭后明确拒绝、关闭后未知结果和重复点击。
- 现有四组浏览器基线：86项，77通过、9失败。失败分别涉及旧模式切换、固定侧栏宽度、旧资料输入、旧文风入口、四种既有会话、旧定时创建及3项历史计划编辑。部分仍按旧入口操作（如点击短篇创建入口寻找旧“参考文本”，或在新调度器寻找旧“创建定时任务”），需要逐项迁移测试并复核业务，不删除测试或把它们报为通过。
- 新调度器、短篇全稿/产物、主题检查：23项全部通过。

## 实际4001与新增已确认缺口

- 更新前运行e0dd242a；更新后 `git_sha` 和 `runtime_manifest_sha` 均为d8eda3bf，前端入口 `/assets/user-B4zKc2KL.js`，监听PID35040。`/health`、入口资源与小说路由返回正常。构建从已提交源码导出，未把其他脏修改带入运行包，未重置它们。
- 运行包：`_local-archive-20261028/novel-parity-runtime-d8eda3bf/backend/qiantie-local.exe`；构建清单在同目录根的 `.local-review/runtime-manifest.json`。原4001二进制仍保留，可回退；临时4014已关闭。
- 当前数据库小说任务2条均paused、无定时计划；未点击生成、保存Agent设置、启动调度器或恢复任务。没有本轮真实模型请求。
- 验收截图及DOM记录：`docs/verification/novel-parity-20261121`。
- **写作风格未对齐**：原版无作品也有名称、参考文本、指南、导入/保存和应用；我们无选中作品时显示“请先选择作品后使用此工具”（`NovelCreationPage.jsx:361`）。全局保存文风的既有API存在，但页面尚未按原版接齐。
- **备忘录未对齐**：原版有独立notebook；我们仅按作品提供`NotesPanel`（`NovelCreationPage.jsx:414`）。这不是颜色差异，而是使用范围和资料管理链路差异。
- **导入素材未完全统一**：左侧工具仍打开作品资料导入，原版默认进入仿写创建页。长篇专用入口已有，但不能因此认为导入工具入口相同。
- 原版/CM完整生成结果、取消/错误分支的真实模型A/B仍未进行；审稿算法、文风算法和新增2.9.5提示词的全部逐字一致性仍未证明。

本轮没有完成真实上游生成和原版结果A/B；完整1:1对齐仍未完成。
