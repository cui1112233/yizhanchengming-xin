# 小说创作独立重写执行计划

1. REA + Craft Anything 固化证据、假设和不可验证边界。
2. 复用已存在的 novelcreation 领域底座，补 Service、HTTP API、MySQL migration 与 Redis 运行协调。
3. 增加独立 `/novel-creation` 三栏工作台、六模式、统一文本模型、状态轮询、停止和重试。
4. 将模式合同注入后端提示词输入，确保六种模式不只是按钮名称不同。
5. 跑 Go 专项测试、前端源码测试与生产构建。
6. 重建 Go embed 单二进制，启动本地 4001，真实点击验证导航和页面。
7. 本地验收通过后才允许进入线上发布步骤。

## 2026-11-15 REA 全量对齐

### 取证状态

- Craft Anything 的 MCP 工具未连接，本机也没有 `craft-anything-cli`，本轮无法补新的运行取证。
- 以现有 [洛水小说创作 REA 报告](../reverse-engineering/2026-11-12-luoshui-novel-studio-rea.md)
  和 [需求规格](../requirements/2026-11-12-novel-creation-spec.md) 作为对齐合同；装好 Craft Anything 后再补一次
  控件树/接口对照，不在无证据时宣称像素级复刻。
- 模型与密钥只使用本系统自己的配置：文本模型来自 `/api/models?kind=text`，不复制洛水私有域名、令牌和签名。

### 现状盘点

已实现：六模式常量与模式记忆；三栏工作台和 360–800px 右栏拖动；作品/会话/消息/运行/取消/重试
API；事件轮询；章节列表；规划→写作→审核→修订→事实→记忆流水线；13 张 novel_* 表。

相对 REA 仍缺失：

| 缺口 | 现状 | 目标 |
|---|---|---|
| 检查点确认 | 只有 `awaiting_confirmation` 状态常量，流水线不使用，无接口和按钮 | 蓝图生成后暂停，页面确认或要求修改后才写正文 |
| 六模式输入合同 | 新建作品只有通用字段 | 仿写传参考文本、续写传前文、番外传人物/世界规则，后端按模式校验 |
| 写作风格 | 工具面板是占位 | 项目级风格档案，注入提示词 |
| 规则 | 无 | 项目级创作规则/禁区，注入每次运行 |
| 导入素材 / knowledgeIds | 无 | 素材入库并在发送消息时携带 |
| 批量续写 | 无 | 按章连续续写，可停止 |
| 定时写作 | 占位 | 定时创建运行 |
| 数据看板 | 占位 | 字数、章节、运行成功率统计 |
| 备忘录 | 占位 | 项目级笔记 |
| Agent 设置 | 占位 | 思考等级、章节字数、审核轮次等项目设置 |
| 导出 | 无 | 整本 Markdown/TXT 导出 |
| 长篇多章推进 | 每次运行只生成 1 章，`chapter_number` 固定 1 | 按大纲推进章节，支持暂停/继续 |
| 上游错误映射 | 只区分超时/取消/其他 | 429/502/503/EOF/限流映射为可读提示 |

### 阶段

1. P1 核心合同：六模式输入校验与提示词注入；蓝图检查点确认/修改接口和页面动作；写作风格与规则；
   整本导出。
2. P2 工具面板：导入素材与 knowledgeIds、Agent 设置、备忘录、数据看板、定时写作。
3. P3 长篇推进与稳定性：多章连续创作、批量续写、上游错误分类与重试提示。

每阶段验收：Go 专项测试 + 前端源码测试 + 正式构建 + 本地 4001 真实点击；通过后才发布。

### 2026-11-15 P1 完成情况

- 六模式输入合同：`ModeInputFields` / `NormalizeModeInputs` 按模式校验，快照写入 `mode` 与 `mode_inputs`，
  并注入基础设定、章节计划、正文；前端按模式显示参考文本、被续写的前文、既有人物与世界规则输入框。
- 蓝图检查点：长篇原创与仿写长篇首次运行在蓝图后停在 `awaiting_confirmation`；
  `POST /api/novel-creation/runs/{id}/confirm` 支持直接确认或提交修改后的蓝图；蓝图按作品复用，
  后续章节不再重复生成，章节号按已存章节递增。
- 写作风格与规则：新增 `novel_project_settings`（migration 8800301）与
  `GET/PUT /api/novel-creation/projects/{id}/settings`，注入基础设定、章节计划、正文和审核；
  前端“写作风格”面板可保存。
- 整本导出：`GET /api/novel-creation/projects/{id}/export?format=markdown|txt`，右栏章节页提供导出按钮。
- 验证：`go test ./internal/novelcreation ./internal/storage ./internal/httpapi` 通过；前端源码测试 4/4；
  正式构建通过；本地 4011 新二进制 `/health` 正常、migration 已建表、新页面 chunk 含新增文案。
- 未完成：P2 工具面板（导入素材/knowledgeIds、Agent 设置、备忘录、数据看板、定时写作）；
  P3 批量续写、暂停/继续的多章编排、上游错误分类与重试提示。

