# Liblib 节点与生成框对标 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 在现有 Agent Canvas 组件上完成 Liblib 风格的四类节点和底部生成框视觉对标。

**Architecture:** 保留现有节点数据与事件处理，仅统一 JSX 语义层级和 `agent-workspace.css` 的布局规则。音频、视频、图片的媒体状态继续由现有组件提供，CSS 负责稳定尺寸和信息分层。

**Tech Stack:** React, Ant Design, React Flow, CSS, Node test runner, Vite, Go embed runtime.

## Global Constraints

- 只修改 `v88` 工作区源码，不修改后端接口或画布数据结构。
- 前端改动必须通过 `npm.cmd --prefix frontend run build` 并重新执行 `scripts/build-go-local.ps1`。
- 不触发真实生成、不消耗模型额度。
- 保留工作区中与本任务无关的已有改动。

### Task 1: 节点结构与样式断言

**Files:**
- Modify: `frontend/src/user/components/AgentCanvas.jsx`
- Modify: `frontend/src/user/components/AgentNodeComposer.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Test: `frontend/src/user/components/AgentCanvas.director-desk.source.test.js`

- [ ] 读取四类节点的现有 JSX，确认事件入口和媒体状态组件不变。
- [ ] 为节点头部、媒体区、状态区和操作区补齐稳定的语义 class，保留原有回调。
- [ ] 将音频节点的现有音频元素包进波形卡片布局，并保留播放行为。
- [ ] 为视频、图片、文本节点补齐固定尺寸、截断和空状态样式。
- [ ] 为生成框整理素材、提示词、参数和提交按钮的布局层级。
- [ ] 增加源码断言，检查四类节点 class 和生成框结构存在。
- [ ] 运行相关 Node 测试，确认通过。

### Task 2: 构建与运行验证

**Files:**
- Modify: `frontend/dist/*` (generated)
- Modify: `backend/web/dist/*` (generated)

- [ ] 运行 `npm.cmd --prefix frontend run build`。
- [ ] 运行 `scripts/build-go-local.ps1`。
- [ ] 只重启 `backend/qiantie-local.exe` 对应的 4001 进程。
- [ ] 检查 `/health` 和 `/api/runtime-build-info`，确认 Go embed SHA 与当前 `v88` HEAD 一致。
- [ ] 在本地浏览器打开工作流，确认四类节点和底部生成框已加载新样式。
