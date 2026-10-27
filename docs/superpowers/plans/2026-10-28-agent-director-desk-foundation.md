# Agent 导演台基础能力实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 在现有 Agent 导演支点中补齐场景对象变换、语义机位、分镜视频提示词和视频支点回填。

**Architecture:** 继续使用 `node.data.config.directorProject` 保存导演台状态，由导演台组件通过 `DirectorDeskHostAdapter` 写回 Agent Canvas。导演台不新增后端表、不复制媒体文件；视频支点回填由 Agent Canvas 根据导演台节点与相邻视频节点的边关系完成。

**Tech Stack:** React、Three.js、React Flow、Node.js tests、Go embed 静态资源。

## Global Constraints

- 不新建导演台项目、任务队列或数据库表。
- 不复制媒体文件，持久化只保存项目素材 ID 和导演台 JSON。
- 不触发真实图片或视频生成任务。
- 保留旧 `actors` 和旧 `scene.prompt` 数据的读取兼容。
- 保留截图回传画布并创建图片支点的现有行为。

---

### Task 1: 扩展导演台状态模型并保持旧数据兼容

**Files:**
- Modify: `frontend/src/user/pages/agent-director-desk.js`
- Test: `frontend/src/user/pages/agent-director-desk.test.js`

**Interfaces:**
- `normalizeDirectorProject(value)` 继续返回完整、可保存的导演台状态。
- 新增 `directorObjectTransform(project, selection, patch)`，返回更新后的项目。
- 新增 `directorCameraSemantics(project, patch)`，根据语义机位参数返回更新后的相机。
- 新增 `directorReferenceAssetIds(project)`，返回去重后的场景、人物和道具图片素材 ID。

- [ ] **Step 1: 写失败测试**：覆盖旧 `actors` 转换为 `objects`、旧 `scene.prompt` 转换为 `scenePrompt`、默认旋转和缩放、语义机位参数保存、引用素材 ID 去重。
- [ ] **Step 2: 运行测试确认失败**：`node --test frontend/src/user/pages/agent-director-desk.test.js`。
- [ ] **Step 3: 实现最小状态变换函数**：保持旧字段读取兼容，统一写入 `rendererVersion: '3d-v2'`、`objects`、`scenePrompt`、`shotPrompt` 和 camera 语义字段。
- [ ] **Step 4: 运行测试确认通过**：同一测试命令必须全部通过。
- [ ] **Step 5: 提交**：`git add frontend/src/user/pages/agent-director-desk.js frontend/src/user/pages/agent-director-desk.test.js && git commit -m "feat: extend director desk project state"`。

### Task 2: 增加场景对象变换和道具入口

**Files:**
- Modify: `frontend/src/user/components/director-desk/DirectorDeskApp.jsx`
- Modify: `frontend/src/user/components/director-desk/DirectorScene.jsx`
- Modify: `frontend/src/user/components/director-desk/DirectorToolRail.jsx`
- Modify: `frontend/src/user/components/director-desk/DirectorAssetPanel.jsx`
- Modify: `frontend/src/user/components/director-desk/DirectorInspector.jsx`
- Modify: `frontend/src/user/components/director-desk/director-scene-state.js`
- Test: `frontend/src/user/components/director-desk/director-scene-state.test.js`

**Interfaces:**
- 场景对象统一使用 `{ id, kind: 'actor'|'prop', assetId, position, rotation, scale }`。
- `DirectorScene` 通过 `onProjectChange` 提交对象变换。
- `DirectorAssetPanel` 的角色和道具入口共用项目图片素材，不新增上传路径。

- [ ] **Step 1: 写失败测试**：对象移动、旋转、缩放都只修改选中对象；删除对象后选择回到场景；道具对象可被引用 ID 收集。
- [ ] **Step 2: 运行目标测试确认失败**：`node --test frontend/src/user/components/director-desk/director-scene-state.test.js`。
- [ ] **Step 3: 实现对象变换**：在 Three.js 视口中增加当前对象的变换模式和检查器输入，保存位置、旋转、缩放到 `objects`。
- [ ] **Step 4: 增加道具工具入口**：沿用图片素材列表，创建 `kind: 'prop'` 对象，不改变场景图选择行为。
- [ ] **Step 5: 运行测试和前端源测试**：目标测试及导演台源测试全部通过。
- [ ] **Step 6: 提交**：`git add frontend/src/user/components/director-desk && git commit -m "feat: add director object transforms"`。

### Task 3: 增加语义机位和双提示词编辑

**Files:**
- Modify: `frontend/src/user/components/director-desk/DirectorDeskApp.jsx`
- Modify: `frontend/src/user/components/director-desk/DirectorInspector.jsx`
- Modify: `frontend/src/user/components/director-desk/DirectorBottomDock.jsx`
- Modify: `frontend/src/user/components/director-desk/director-scene-state.js`
- Test: `frontend/src/user/components/director-desk/director-inspector-state.test.js`

**Interfaces:**
- 机位字段固定为 `shotSize`, `focalLength`, `azimuth`, `elevation`, `position`, `target`, `fov`。
- 底部组件接收 `scenePrompt` 和 `shotPrompt`，分别通过 `onScenePromptChange`、`onShotPromptChange`、`onSavePrompts` 保存。

- [ ] **Step 1: 写失败测试**：修改景别、焦段、方位角和俯仰角会产生可预测的 camera patch；两个提示词互不覆盖。
- [ ] **Step 2: 运行测试确认失败**：`node --test frontend/src/user/components/director-desk/director-inspector-state.test.js`。
- [ ] **Step 3: 实现语义机位换算**：以 target 为中心，根据方位角、俯仰角和距离重算 position，并根据焦段更新 fov。
- [ ] **Step 4: 实现双提示词编辑**：场景描述和分镜视频提示词分成两个可折叠输入区，保留保存状态和错误状态。
- [ ] **Step 5: 运行目标测试和构建前静态测试**。
- [ ] **Step 6: 提交**：`git add frontend/src/user/components/director-desk && git commit -m "feat: add director camera semantics and prompts"`。

### Task 4: 将导演台结果回填关联视频支点

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/components/director-desk/DirectorDeskHostAdapter.js`
- Modify: `frontend/src/user/pages/agent-canvas-connections.js`
- Test: `frontend/src/user/components/AgentCanvas.director-desk.source.test.js`
- Test: `frontend/src/user/pages/agent-canvas-connections.test.js`

**Interfaces:**
- `directorReferenceAssetIds(project)` 提供图片引用 ID。
- 关联判断只使用当前画布边：导演节点到视频节点的直接边，避免把无关视频节点覆盖。
- 回填视频节点配置字段：`prompt`, `directorProject`, `camera`, `referenceAssetIds`。

- [ ] **Step 1: 写失败来源测试**：存在导演到视频边时生成回填 patch；无视频目标时只保存导演台；回填错误时产生“视频支点尚未同步”状态。
- [ ] **Step 2: 运行来源测试确认失败**：`node --test frontend/src/user/components/AgentCanvas.director-desk.source.test.js frontend/src/user/pages/agent-canvas-connections.test.js`。
- [ ] **Step 3: 实现回填函数**：在导演台状态写入和关闭前，读取当前节点和边，按直接相邻的视频节点应用配置 patch，不调用生成接口。
- [ ] **Step 4: 实现失败提示**：导演台仍保持打开，保存错误不被吞掉，状态不能显示为“已保存”。
- [ ] **Step 5: 运行目标测试通过并检查没有真实生成请求**。
- [ ] **Step 6: 提交**：`git add frontend/src/user/components/AgentCanvas.jsx frontend/src/user/components/director-desk/DirectorDeskHostAdapter.js frontend/src/user/pages/agent-canvas-connections.js frontend/src/user/components/AgentCanvas.director-desk.source.test.js frontend/src/user/pages/agent-canvas-connections.test.js && git commit -m "feat: sync director desk into video nodes"`。

### Task 5: 构建、嵌入和浏览器验收

**Files:**
- Modify: `frontend/dist/` via the existing build only
- Modify: `backend/web/dist/` via `scripts/build-go-local.ps1`

- [ ] **Step 1: 运行导演台和画布聚焦测试**：`node --test frontend/src/user/components/director-desk/*.test.js frontend/src/user/pages/agent-canvas-connections.test.js`。
- [ ] **Step 2: 构建前端**：在 `frontend` 目录运行 `npm run build`，确认无编译错误。
- [ ] **Step 3: 运行后端全量测试**：在 `backend` 目录运行 `go test ./...`。
- [ ] **Step 4: 更新 Go 内嵌静态资源**：根目录运行 `./scripts/build-go-local.ps1`。
- [ ] **Step 5: 浏览器验收**：打开现有 Agent 项目导演台，添加人物和道具，修改移动/旋转/缩放、机位和两个提示词，关闭并刷新；确认导演台状态保留，关联视频支点出现提示词、机位和引用素材 ID；不点击生成按钮。
- [ ] **Step 6: 提交构建验收记录**：`git status --short` 只核对本任务文件和构建产物，不清理工作区已有改动。

