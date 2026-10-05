# v88 Batch Factory V11 前端盘点

任务：Task 9.1.1 — 列出 V11 所有前端页面与组件。

来源基线：旧仓库 `cui1112233/-` 的 `v88` 分支。

## 结论

旧 v88 的 Batch Factory V11 前端不是单一页面，而是两层入口叠加：

1. `/batch-factory`：当前实际批量工程入口，走 `BatchFactoryWorkbenchPage.jsx`，并依赖 `shuihuo/` 目录中的批次创建、小说列表、统一设置、单书设置等组件。
2. `/batch-factory-preview`：V11 UI/兼容预览入口，走 `BatchFactoryPage.jsx -> BatchFactoryV11UiPage.jsx`；`BatchFactoryPageV11.jsx` 是历史兼容入口，和 `BatchFactoryPage.jsx` 渲染同一权威 V11 页面。

迁移时不能只复制 `batch-factory-v11/` 目录；否则会漏掉公网当前 `/batch-factory` 的真实批次列表、创建批次和单书生产工作台。

---

## 1. 路由与页面入口

| 路由/用途 | v88 文件 | 角色 |
|---|---|---|
| `/batch-factory` | `frontend/src/user/pages/BatchFactoryWorkbenchPage.jsx` | 当前批量工程主页；列出批次、创建批次、打开批次、接收小说获取 Intake、启动自动化 |
| `/batch-factory-preview` | `frontend/src/user/pages/BatchFactoryPage.jsx` | V11 UI 预览入口，直接渲染 `BatchFactoryV11UiPage` |
| 历史 V11 兼容入口 | `frontend/src/user/pages/BatchFactoryPageV11.jsx` | 历史 import 兼容层，同样渲染 `BatchFactoryV11UiPage` |
| 用户端路由表 | `frontend/src/user/App.jsx` | 注册 `/batch-factory` 与 `/batch-factory-preview` |

### 1.1 `/batch-factory` 当前入口的直接 React 依赖

`BatchFactoryWorkbenchPage.jsx` 直接使用：

- `frontend/src/user/pages/shuihuo/BatchFactoryNovelList.jsx`
- `frontend/src/user/pages/shuihuo/BatchFactoryCreateModal.jsx`
- `frontend/src/user/pages/shuihuo/batchFactoryProjects.js`
- `frontend/src/user/pages/shuihuo/batchFactoryNovelFetchHandoff.js`
- `frontend/src/user/pages/BatchFactoryWorkbenchPage.css`
- `frontend/src/user/pages/shuihuo-production.css`

它还直接调用 `shared/api/batchFactoryV11` 中的批次、Intake、自动化 API；这些接口放到 Task 9.1.2 盘点。

---

## 2. `batch-factory-v11/` 专用 React 页面/组件

目录：`frontend/src/user/pages/batch-factory-v11/`

| 文件 | 角色 |
|---|---|
| `BatchFactoryV11UiPage.jsx` | V11 权威 UI 容器；装配 runtime、adapter、批次、生产、合并、设置、最终提示词、外部发布等能力 |
| `BatchFactoryV11Workbench.jsx` | V11 可视化生产工作台；组织 Director、Hook、小说源、资产、状态卡等 |
| `BatchFactoryV11BatchManager.jsx` | 批次管理 UI |
| `BatchFactoryV11ConstraintEditor.jsx` | 约束/提示词约束编辑器 |
| `BatchFactoryV11PublishSettings.jsx` | 发布设置 Drawer/面板 |
| `BatchFactoryV11ScopedSettings.jsx` | 单书设置、视频设置等作用域设置 |
| `BatchFactoryV11SettingsDrawers.jsx` | 生产统一设置等 Drawer 容器 |
| `ChangeImpactNotice.jsx` | 配置/改动影响提示 |
| `DirectorPanel.jsx` | Director 分镜面板 |
| `DirectorRefreshContext.jsx` | Director 刷新上下文 Provider |
| `ExternalPublishPanel.jsx` | 外部发布面板 |
| `FinalPromptPreviewDrawer.jsx` | 最终提示词预览 Drawer |
| `FixedSingleVideoControl.jsx` | 单条固定视频控制组件 |
| `HookReviewPanel.jsx` | Hook 审核面板 |
| `NovelSourceModule.jsx` | 小说源/正文模块 |
| `OverrideCompatibilityDetails.jsx` | override 兼容性详情 |
| `ProductionMediaBoundary.jsx` | 生产媒体边界/媒体输入保护组件 |
| `WorkbenchCard.jsx` | V11 工作台通用卡片 |

### 2.1 `BatchFactoryV11UiPage.jsx` 直接装配关系

它直接装配以下核心 UI：

- `BatchFactoryV11Workbench`
- `ProductionSettingsDrawer`
- `PublishSettingsDrawer`
- `BookSettingsModal`
- `VideoSettingsDrawer`
- `DirectorRefreshProvider`
- `FinalPromptPreviewDrawer`
- `ExternalPublishPanel`

因此这些组件不是孤立页面，而是 V11 主容器运行时的一部分。

### 2.2 `BatchFactoryV11Workbench.jsx` 直接装配关系

它直接使用：

- `DirectorPanel`
- `HookReviewPanel`
- `NovelSourceModule`
- `WorkbenchCard`
- `workspace-layout.js`
- `batchFactoryV11State.js`

---

## 3. `batch-factory-v11/` 前端状态、适配器与布局模块

这些不是 React 页面，但迁移 V11 前端时必须保留其行为语义：

- `batchFactoryV11State.js` — V11 UI action/intake 状态辅助
- `bf11Runtime.js` — V11 页面运行时装配与加载流程
- `bf11UiAdapter.js` — UI 与 API 的适配层
- `directorState.js` — Director 状态转换
- `externalState.js` — 外部发布状态转换
- `changeImpactView.js` — 变更影响视图模型
- `saveFlow.js` — 设置保存流程
- `showcaseData.js` — 展示/示例数据
- `workspace-layout.js` — 工作台布局、拖拽/折叠/最大化状态

> 新主线后续不会机械照搬 Node/前端状态成为业务事实源；MySQL/Go 仍是事实源，这些文件主要用于理解旧交互和状态映射。

---

## 4. `batch-factory-v11/` 样式文件

- `batch-factory-v11-batch-manager.css`
- `batch-factory-v11-detail.css`
- `batch-factory-v11-scoped.css`
- `batch-factory-v11-settings.css`
- `batch-factory-v11-theme.css`
- `batch-factory-v11-workbench.css`

---

## 5. `/batch-factory` 当前真实入口依赖的 `shuihuo/` Batch Factory 组件

目录：`frontend/src/user/pages/shuihuo/`

### 5.1 直接进入当前 V11 主流程的组件

| 文件 | 角色 |
|---|---|
| `BatchFactoryCreateModal.jsx` | 新建批量、书城分组、小说获取、自动化、巨量入口 |
| `BatchFactoryNovelList.jsx` | 当前批次的单书生产主工作台；正文、配置、资产、Director/H3、生产、视频、合并、发布等大量动作集中于此 |
| `BatchFactoryUnifiedSettingsModal.jsx` | 批次统一设置 |
| `BatchFactoryBookSettingsModal.jsx` | 单书级设置 |
| `BatchFactoryGiantMaterialPendingProgress.jsx` | 巨量素材等待/处理中进度 |
| `BatchFactoryGiantMaterialExecutorStatus.jsx` | 巨量本地执行器状态；由创建批量流程直接使用 |

### 5.2 同一批量工厂域中存在的附加 React 组件

这些文件位于同一 `shuihuo/` 批量工厂域，后续对应能力迁移时需要判断是否继续保留：

- `BatchFactoryAiReasoningModal.jsx`
- `BatchFactoryEngineSettingsDrawer.jsx`
- `BatchFactoryGiantMaterialImportModal.jsx`
- `BatchTaskModal.jsx`

其中 `BatchFactoryCreateModal.jsx` 已确认直接引用 `BatchFactoryGiantMaterialExecutorStatus.jsx`；`BatchFactoryNovelList.jsx` 已确认直接引用 `BatchFactoryUnifiedSettingsModal.jsx`、`BatchFactoryBookSettingsModal.jsx`、`BatchFactoryGiantMaterialPendingProgress.jsx` 和 V11 的 `ProductionMediaBoundary.jsx`。

---

## 6. `/batch-factory` 当前真实入口的前端业务辅助模块

这些不是 React 组件，但直接影响当前批量工厂页面行为：

- `batchFactoryAutomationSchedule.js`
- `batchFactoryBookConfigRegions.js`
- `batchFactoryBookState.js`
- `batchFactoryContentRange.js`
- `batchFactoryGiantMaterialImport.js`
- `batchFactoryGiantMaterialQueue.js`
- `batchFactoryH3Constraints.js`
- `batchFactoryManualFetch.js`
- `batchFactoryMergeCover.js`
- `batchFactoryNovelFetchHandoff.js`
- `batchFactoryPlatformOptions.js`
- `batchFactoryPrecompiledStoryboard.js`
- `batchFactoryProjects.js`
- `batchFactoryPublicationMetadata.js`
- `batchFactoryRuntimeLogic.js`
- `batchFactorySmartUnified.js`
- `batchFactoryVideoMeta.js`
- `h3LineAudio.js`
- `h3PromptEditing.js`
- `videoProviderBinding.js`

这些文件会在 Task 9.1.2/9.1.3/9.1.4 继续追踪其 API、状态和 provider 依赖。

---

## 7. 旧 v88 已存在的前端回归测试资产

### 7.1 `batch-factory-v11/` 专用测试

旧目录已经存在大量 source/state/runtime 测试，包括但不限于：

- `BatchFactoryV11UiPage.source.test.js`
- `batch-manager-source.test.js`
- `batchFactoryV11State.test.js`
- `bf11Runtime.test.js`
- `bf11UiAdapter.test.js`
- `config-version-sync.test.js`
- `detail-ui-source.test.js`
- `director-no-effective-settings.test.js`
- `director-revision-refresh-source.test.js`
- `director-ui-source.test.js`
- `directorState.test.js`
- `externalState.test.js`
- `fixed-video-source.test.js`
- `intakeCreateFlow.test.js`
- `live-wiring-source.test.js`
- `novel-fetch-121-safety-source.test.js`
- `novel-fetch-current-batch-source.test.js`
- `novel-fetch-http-login.test.js`
- `novel-fetch-intake-source.test.js`
- `override-compatibility-source.test.js`
- `saveFlow.test.js`
- `scoped-settings-source.test.js`
- `settings-source.test.js`
- `ui-page-source.test.js`
- `v11-ui-source-guard.test.js`
- `workbench-source.test.js`
- `workspace-layout.test.js`

### 7.2 `shuihuo/` 批量工厂相关测试

已确认存在：

- `BatchFactoryBookSettingsModal.source.test.js`
- `BatchFactoryCreateModal.source.test.js`
- `BatchFactoryGiantMaterialExecutorStatus.source.test.js`
- `BatchFactoryGiantMaterialExecutorStatus.test.js`
- `BatchFactoryGiantMaterialFlow.source.test.js`
- `BatchFactoryGiantMaterialImportModal.source.test.js`
- `BatchFactoryGiantMaterialPendingProgress.source.test.js`
- `BatchFactoryH3Trace.source.test.js`
- `BatchFactoryNovelList.assets.source.test.js`
- `BatchFactoryNovelList.source.test.js`
- `BatchFactoryUnifiedSettingsModal.source.test.js`
- `batchFactoryAutomationSchedule.test.js`
- `batchFactoryBookConfigRegions.test.js`
- `batchFactoryBookState.test.js`
- `batchFactoryContentRange.test.js`
- `batchFactoryGiantMaterialImport.test.js`
- `batchFactoryGiantMaterialQueue.test.js`
- `batchFactoryH3Constraints.test.js`
- `batchFactoryManualFetch.test.js`
- `batchFactoryMergeCover.test.js`
- `batchFactoryNovelFetchHandoff.test.js`
- `batchFactoryPlatformOptions.test.js`
- `batchFactoryPrecompiledStoryboard.test.js`
- `batchFactoryProjects.test.js`
- `batchFactoryPublicationMetadata.test.js`
- `batchFactoryRuntimeLogic.test.js`
- `batchFactorySmartUnified.test.js`
- `batchFactoryVideoMeta.test.js`
- `videoProviderBinding.test.js`

后续迁移应优先把这些测试表达的旧行为转成新主线 Go/React 的回归测试，而不是只复制组件代码。

---

## 8. Task 9.1.1 迁移边界结论

本任务只完成“前端页面与组件盘点”，不开始复制 V11 业务代码。

后续顺序：

1. Task 9.1.2：沿 `shared/api/batchFactoryV11`、旧 Node routes 和 Go backend 列出全部历史后端接口。
2. Task 9.1.3：根据接口和旧前端状态文件，列出 Batch / Book / Run / Item / VIDEO 状态字段。
3. Task 9.1.4：专项盘点 `personal_api` / video provider。
4. Task 9.1.5：再对 ECS 实际可用性做“存在 / 可用 / 已坏 / 未验证”分级。
5. Task 9.1.6：形成旧 v88 -> 新 Go 模块迁移映射表。

本清单的目的不是要求新主线原样复制所有旧组件；它是防漏迁基线。后续应按功能职责重新收敛，避免继续维持旧 v88 中两个 Batch Factory UI 入口长期并存的结构。
