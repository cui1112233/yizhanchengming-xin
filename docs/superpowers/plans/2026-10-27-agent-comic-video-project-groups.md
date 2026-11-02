# Agent 漫画与视频作品分组实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为 Agent 项目增加固定的漫画/视频类型，并让列表、画布能力、输入框、模型和后端生成接口统一遵守该类型。

**Architecture:** 在 Agent workspace 的项目摘要和详情上增加 `projectType`，旧数据归一化为 `video`。前端用一个纯函数能力表控制项目筛选、节点菜单和编辑器输入；后端在项目创建、批量工厂绑定和视频接口处做同样的类型校验，避免只隐藏 UI。

**Tech Stack:** Go、MySQL/现有 Agent workspace Store、React、Ant Design、Node test、Go test、Go embed。

## Global Constraints

- 保持当前 `v88` 分支，禁止新建分支、reset、clean 或覆盖无关脏改动。
- 旧项目没有 `projectType` 时必须按 `video` 兼容。
- 漫画项目禁止视频生成、视频工具、视频合成和视频上传；视频项目保持现有行为。
- 不发送真实 API Key，不触发付费生成。
- 系统预设提示词仍由后端 prompt 模块提供，前端只读取类型过滤后的模型和字段。

### Task 1: 项目类型纯函数与 API 契约

**Files:**
- Create: `frontend/src/user/pages/agent-project-type.js`
- Test: `frontend/src/user/pages/agent-project-type.test.js`
- Modify: `frontend/src/shared/api/agent.js`
- Test: `frontend/src/shared/api/agent.test.js`

**Interfaces:**
- `normalizeAgentProjectType(value): 'video' | 'comic'`
- `agentProjectTypeLabel(type): string`
- `agentProjectCapabilities(type): { nodeTypes: string[], modelKinds: string[], allowVideoGeneration: boolean, allowAudio: boolean }`
- `createAgentTask({ projectType } = {}): Promise<ProjectSummary>` sends `{ projectType }`.

- [ ] Write tests for defaulting unknown/missing values to `video`, comic labels, and comic capability flags.
- [ ] Run `node --test src/user/pages/agent-project-type.test.js src/shared/api/agent.test.js` and confirm the new assertions fail.
- [ ] Implement the pure functions and update `createAgentTask` to accept and send `projectType` while preserving callers with no argument.
- [ ] Run the focused tests and confirm they pass.

### Task 2: Go project type persistence and compatibility

**Files:**
- Modify: `backend/internal/agentworkspace/types.go`
- Modify: `backend/internal/agentworkspace/store.go`
- Modify: `backend/internal/agentworkspace/memory_store.go`
- Modify: `backend/internal/agentworkspace/mysql_store.go`
- Modify: `backend/internal/storage/agent_workspace_schema.go`
- Test: `backend/internal/agentworkspace/store_test.go`
- Test: `backend/internal/storage/agent_workspace_schema_test.go`

**Interfaces:**
- `ProjectSummary.ProjectType` and `Project.ProjectType` are JSON fields with values `video` or `comic`.
- `NormalizeProjectType(value string) string` returns `video` for empty/unknown values.
- `CreateProject` persists the normalized type and all read paths return it.

- [ ] Add tests that create/read comic projects and that legacy JSON/rows without a type read as `video`.
- [ ] Run the focused Go tests and confirm failure before implementation.
- [ ] Add the type to the in-memory and MySQL project representations using the repository’s existing migration/schema conventions; do not rewrite existing rows destructively.
- [ ] Normalize on input and on read so imported legacy projects remain usable.
- [ ] Run `go test ./backend/internal/agentworkspace ./backend/internal/storage` and confirm pass.

### Task 3: HTTP project creation, list, and batch-factory handoff

**Files:**
- Modify: `backend/internal/httpapi/agent_workspace.go` and the existing Agent project route file found by the current handler registration.
- Modify: `backend/internal/httpapi/agent_batch_factory_bridge.go`
- Test: corresponding `*_test.go` files for project routes and bridge.
- Modify: `frontend/src/user/pages/AgentPageV2.jsx`

**Interfaces:**
- `POST /api/agent/tasks` accepts optional `{ projectType }` and returns normalized `projectType`.
- `GET /api/agent/tasks` and `GET /api/agent/tasks/{id}` return `projectType`.
- Batch-factory Agent binding forwards comic mode as `projectType: 'comic'` and video mode as `video`.

- [ ] Add HTTP tests for creating comic projects, listing their type, and rejecting invalid types.
- [ ] Add a bridge test that a comic batch creates/updates a comic Agent project.
- [ ] Implement request decoding and response projection with `video` compatibility.
- [ ] Update frontend creation calls and task state to retain the returned type.
- [ ] Run focused Go and frontend API tests.

### Task 4: Agent home list and creation UI grouping

**Files:**
- Modify: `frontend/src/user/pages/AgentPageV2.jsx`
- Modify: `frontend/src/user/pages/AgentPageV2` related stylesheet.
- Test: `frontend/src/user/pages/agent-project-groups.source.test.js`

**Interfaces:**
- Project filter state is `all | video | comic`.
- New-project flow sends the selected type and displays a type label on cards.

- [ ] Add source tests for the three filters, type labels, comic/video creation options, and passing `projectType` to `createAgentTask`.
- [ ] Implement the filter tabs and compact card tag without changing existing navigation URLs.
- [ ] Implement creation selection with video as the compatibility default.
- [ ] Run the focused source test.

### Task 5: Canvas capability filtering and matched inputs

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/components/AgentNodeComposer.jsx`
- Modify: `frontend/src/user/components/AgentNodeSettingsPanel.jsx`
- Modify: `frontend/src/user/components/AgentCanvasControls.jsx` or context-menu module.
- Test: `frontend/src/user/pages/agent-project-capabilities.test.js`
- Test: related Agent Canvas source tests.

**Interfaces:**
- `AgentCanvas` receives `projectType` and derives capabilities from `agentProjectCapabilities`.
- Comic projects hide/disable video and audio creation controls, video-only parameter fields, and video actions.
- Video projects preserve current controls unchanged.

- [ ] Add tests for comic allowed nodes (`script`, `text`, `image`), hidden video/audio creation actions, image model/prompt fields, and video compatibility.
- [ ] Implement one capability guard used by top menu, context menu, composer, settings, and quick tools; do not rely on CSS-only hiding.
- [ ] Pass project type through AgentPageV2 into the canvas and keep old projects normalized to video.
- [ ] Run all focused Agent Canvas tests.

### Task 6: Server-side generation guard

**Files:**
- Modify: `backend/internal/httpapi/agent_video_jobs.go`
- Modify: `backend/internal/httpapi/agent_video_processor.go` and any shared video generation route reached by the current router.
- Test: related video route tests.

**Interfaces:**
- Before creating a video job/tool task, load the project and return HTTP 409 with `漫画作品不支持视频生成` when `ProjectType == comic`.

- [ ] Add a failing HTTP test posting a video request for a comic project.
- [ ] Implement the guard at the shared route boundary so retries/tools cannot bypass it.
- [ ] Run focused video route tests and ensure video projects still submit normally.

### Task 7: Build, runtime, and acceptance checks

**Files:**
- Modify only generated embed outputs through the existing build scripts.

- [ ] Run frontend focused tests, `npm run build`, and `go test ./...`.
- [ ] Rebuild and restart the Go embedded local runtime with the existing script.
- [ ] Run `scripts/check-local-go.ps1` and verify runtime build info matches the new commit.
- [ ] In the local browser, create no real media and verify list filters, type labels, comic controls hidden, and video controls preserved.
- [ ] Run `git diff --check` and report unrelated pre-existing test failures separately.
