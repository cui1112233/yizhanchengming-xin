# 小说面板模块迁移审计与集成说明（2026-10-07）

## 1. 本轮边界

本轮只在独立模块内补齐小说面板，不注册全局路由，不申请或占用 Goose 迁移编号，不修改共享账号权限、模型配置、队列、TOS、依赖或 Docker，也不部署公网、不调用付费模型。

目标仓库起点：`cui1112233/yizhanchengming-xin@19df4c4ce8f446df856ec0f0eb1558fd913de51d`。

固定模块分支：`feat/novel-panel-module-20261007`。

任务 A 的最终可复制协作基线（共享接口/类型、文件归属、数据库迁移编号）在本轮开工时尚未形成，因此本轮严格停留在模块内实现 + 合成测试 + 集成契约说明。

## 2. 公网与旧源码基线

### 公网

当前公网地址 `http://115.190.156.223:3000` 在本执行环境中无法直接建立连接，因此没有把“页面打不开/抓不到”误判为功能不存在，也没有声称已完成公网逐按钮验收。

旧仓 `CURRENT_VERSION.md` 明确规定：公网运行 SHA 必须以 `/api/build-info` 或正式发布记录核验，不能假设 `v88 HEAD == production`。本轮无法访问该接口，所以“公网当前 SHA”列为未验证项。

此前历史线索为 `27fa2e12f9401d378d1e6d6298a9d50a27581512`；本轮重新确认当前 `v88` HEAD 已前进到 `ab02baa92a6f059f0b39c7bb57cd61872da77da3`，且其父提交正是 `27fa2e12...`。因此本轮把 `27fa2e12...` 仅作为“公网候选源码对照点”，不把它冒充为已重新验证的公网运行 SHA。

### 旧源码小说面板证据

在 `cui1112233/-@27fa2e12...` 中确认小说面板不是占位页，而是独立工作台与后端域：

- `frontend/src/user/pages/NovelPanelPage.jsx`：主站嵌入工作台并通过受限 bridge 访问 `/api/novel-panel/*`、TTS、模型目录。
- `lib/novel-panel/contracts.js`：项目输入与 JSON 安全校验。
- `lib/novel-panel/project-store.js`：项目保存。
- `lib/novel-panel/history-store.js`：历史记录、摘要与恢复快照语义。
- `lib/novel-panel/premium-store.js`：精品模式配置。
- `lib/novel-panel/quality-gate.js`：分镜质量门。
- `routes/novel-panel.js`：文本模型选择、人物/CharacterCore、分镜、设置、异常与并发控制。
- `public/novel-panel/workbench/`：原工作台 UI、人物核心、正式名单、casting、scene context 等。
- `prompts/小说面板人物场景提取.md`、`prompts/快速导演分镜-*.md`：人物提取与不同分镜密度/案例学习语义。

旧实现的关键行为已经迁移成新模块的领域规则，而不是照搬 Node/静态工作台：

1. 强制人物名单支持换行、逗号、中文逗号、顿号、分号、竖线，并且括号内部的分隔符不拆人。
2. 人物名单为正式人物卡权威集合；重算名单时保留已编辑的人物外形、备注、共享资产引用。
3. 每个非空原文段必须被至少一个分镜覆盖，`sourceIndex` 指向非空原文顺序，`sourceBasis` 必须来自对应原文。
4. 分镜必须有画面语义；非法/不完整分镜不能覆盖已保存版本。
5. 保存采用修订号冲突检测，恢复历史会生成新的可审计恢复记录，而不是直接覆写历史。
6. 普通/精品带图模式、内容类型、统一风格、密度、案例学习和指令设置都进入同一 Workspace 快照。

## 3. 新仓本轮实现

### 后端：`api/internal/novelpanel`

- `model.go`
  - 普通模式 `normal`
  - 精品带图模式 `premium_illustrated`
  - 精简/标准/较细三档密度
  - 文本处理、人物卡、人物关系、分镜、案例学习、指令设置、历史记录模型
  - `StoryboardRequest` 只负责把小说面板上下文交给共享生成服务，不定义第二套模型供应商接口
- `text.go`
  - CRLF 统一、行首尾清理、连续空行折叠、全角空格处理
  - 保留原文与处理文本两份数据
  - 非空原文段建立稳定的 `sourceIndex`
- `characters.go`
  - 强制人物名单解析、去重、括号提示、稳定人物 ID
  - 正式名单覆盖集合，同时保留人物卡编辑字段
  - 人物关系端点/自关联/重复关系校验
- `storyboard.go`
  - 可编辑分镜骨架
  - 较细密度按句界拆分
  - 全文覆盖、原文依据、画面必填、时长与重复 ID 校验
- `store.go`
  - `Store` 是模块持久化边界
  - `MemoryStore` 用于模块合成测试
  - 保存和历史写入是同一 Store 操作，避免业务层出现“工作区已保存但历史没写”的半状态
  - 乐观修订号避免陈旧页面覆盖新内容
- `service.go`
  - 统一规范化与校验
  - 保存、历史列表、恢复
  - 构造共享分镜服务输入，但不发起模型调用

### 前端：`前台/src/novel-panel`

`NovelPanelWorkbench.jsx` 提供完整模块工作台：

- 普通 / 精品带图模式
- 整段原文输入与文本处理预览
- 内容类型与统一风格
- 人物强制名单与人物卡编辑
- 人物关系编辑
- 分镜/画面逐镜编辑
- 精简/标准/较细密度
- 案例学习
- 生成规则、必须覆盖细节、镜头节奏、负面指令
- 保存说明、保存记录与历史恢复
- 配音、人物/场景/道具资产、图片入口全部采用 `sharedServices` 注入，不出现 API Key/Provider 管理
- “请求共享 Director 分镜”采用 `api.requestStoryboard` 注入，不复制剧本模块的生成实现

本轮没有把工作台挂进 `App.jsx`，因为那属于全局路由/总集成所有权。

## 4. 与剧本/共享能力的重叠边界

| 能力 | 小说面板本轮所有权 | 应由总集成/共享模块所有 |
| --- | --- | --- |
| 原文编辑、文本处理快照 | 是 | 小说获取负责来源抓取，不在面板重复抓取 |
| 强制人物名单、人物卡编辑、人物关系 | 是（面板语义与状态） | 人物/场景/道具资产的真实资产记录与生成 |
| 分镜逐镜编辑、全文覆盖规则、sourceBasis | 是 | SCRIPT/HOOK/DIRECTOR/FINAL_PROMPT 的模型生成链路 |
| 密度、案例学习、指令设置 | 是（作为生成上下文） | 提示词目录、模型选择和实际推理调用 |
| 普通/精品带图、内容类型、统一风格 | 是 | 图片 Provider、图片任务、TOS 资产落库 |
| 配音入口 | 仅共享入口 | Task 13/共享音频服务 |
| 图片入口 | 仅共享入口 | Task 14/共享图片与资产服务 |
| 模型/API Key/供应商 | 否 | 统一后台模型配置 |
| 队列、锁、任务执行 | 否 | 共享 Redis/任务运行时 |

建议总集成明确：小说面板只拥有“创作工作区状态 + 校验 + 编辑体验”；任何模型执行和媒资生命周期都保持单一共享事实源。

## 5. 待总集成接入位置

以下位置只记录契约，本分支没有改动：

1. `api/internal/httpapi/server.go`
   - 注册小说面板 GET/PUT、history list/restore 路由。
   - 路由继续复用现有鉴权、同源检查、项目归属与 capability 中间件。
2. `api/db/migrations/`
   - 需要任务 A 分配正式迁移编号后再增加持久化表；本分支没有占号。
   - 建议最小事实源：
     - `novel_panel_workspaces(project_id PK, revision, workspace_json, created_at, updated_at)`
     - `novel_panel_history(id PK, project_id, revision, note, summary_json, snapshot_json, created_at)`
   - Workspace + History 必须同一 MySQL 事务保存。
3. MySQL Store
   - 用 `novelpanel.Store` 实现 MySQL 适配器；`MemoryStore` 只用于合成测试，不能作为生产事实源。
4. `前台/src/App.jsx`
   - 注册小说面板路由/菜单，并注入 API adapter 与 sharedServices。
5. Task 12 generation
   - 将 `novelpanel.StoryboardRequest` 映射为共享 Director 输入；不要新建小说面板专属 LLM client。
6. Task 13 / Task 14
   - 把配音、资产、图片 UI 按钮接到现有共享服务；不要在小说面板重复 Provider 配置。

## 6. 测试范围

### 已执行（模块独立）

Go 包测试：`go test ./internal/novelpanel -count=1`

覆盖：

- CRLF/全角空格/空行处理且不重排文本
- 强制人物名单括号内分隔符、去重、稳定 ID
- 强制名单更新时人物卡编辑字段保留、名单外人物被剔除
- 未知人物关系、自关联拦截
- 密度分镜骨架与原文索引
- 全文分镜覆盖、原文依据、画面必填
- 保存/历史/恢复
- 陈旧 revision 冲突不覆盖
- 非法分镜保存失败后旧记录保持不变
- 精品带图模式统一风格必填
- 共享 Director 请求数据携带人物、关系、密度、案例学习和指令

### 分支 CI 待验证

推送后由现有 `ci.yml` 在 `feat/**` 分支执行：

- `go test ./...`
- `go build ./...`
- 全仓前端单测/构建（若 CI 对前端有对应 job）
- 现有 Goose 回归（本分支未增加迁移）

## 7. 当前明确未验证项

1. 公网 `/api/build-info` 与当前真实运行 SHA：执行环境无法连接公网地址。
2. 公网小说面板逐按钮/逐字段视觉一致性：无法直接打开公网工作台，本轮以旧源码与用户给定范围做语义对照。
3. MySQL 重启后的小说面板保存恢复：任务 A 尚未分配迁移编号，本分支禁止新增迁移，因此生产 Store 尚未接入。
4. 全局路由、菜单、权限与项目归属：属于总集成文件，本分支未改。
5. 真实 SCRIPT/HOOK/DIRECTOR/FINAL_PROMPT、TTS、图片、TOS 调用：按要求不调用付费模型，也不复制共享能力。
6. 真实跨模块 E2E：需要任务 A 将路由、MySQL Store、共享 generation/assets/audio adapter 集成后执行。

## 8. 对标差距结论

模块内部的编辑语义、数据校验、保存/恢复生命周期已补齐；距离“公网可替换”的剩余差距全部集中在总集成拥有的四类共享接线：MySQL 持久化迁移、全局路由/权限、Task 12 Director、Task 13/14 音频/资产/图片。未把这些依赖伪装成已完成。
