# Agent Canvas 左上角项目与视图入口 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 Agent Canvas 的项目导航与工作流/故事板切换收拢到左上角的紧凑工具条，并复用现有项目能力。

**Architecture:** 只调整 `AgentWorkspaceHeader` 的展示和入口，把已有的项目列表、创建、删除、返回首页、视图 URL 同步回调传入头部。项目没有独立的“副本”数据模型，因此不新增虚假的副本选择器；当前项目名称仍对应现有 Agent 任务。样式集中在已有的工作区 CSS 中，避免影响画布节点和侧栏。

**Tech Stack:** React、Ant Design、Lucide React、现有 Go API、Node 内置测试。

## Global Constraints

- 复用 `frontend/public/assets/brand-logo-white.png` 与 `BrandLogo`，不复制参考网站资源。
- 工作流和故事板继续使用 `workspaceViewForActiveView` 与 `syncWorkspaceViewUrl` 管理状态和 `view` 参数。
- Logo 菜单固定为“回到主页、全部项目、创建新项目、删除项目”；删除必须二次确认。
- 不新增项目/副本后端模型、接口或数据库字段。
- 仅本地修改、测试和验证；未经用户再次明确授权，不提交、推送或发布。

---

### Task 1: 固化左上角头部结构的回归测试

**Files:**
- Modify: `frontend/src/user/pages/agent-workspace-header.source.test.js`

**Interfaces:**
- Consumes: `AgentWorkspaceHeader` 源码。
- Produces: 对品牌 Logo 菜单、图标视图切换和回调名称的源码回归保护。

- [ ] **Step 1: 写出会失败的头部布局测试**

```js
test('共享头部使用品牌 Logo 菜单和图标化视图切换', () => {
  const header = read('../components/AgentWorkspaceHeader.jsx');
  for (const label of ['回到主页', '全部项目', '创建新项目', '删除项目']) {
    assert.match(header, new RegExp(label));
  }
  assert.match(header, /<BrandLogo/);
  assert.match(header, /<Workflow/);
  assert.match(header, /<PanelsTopLeft/);
  assert.doesNotMatch(header, /<Segmented/);
});
```

- [ ] **Step 2: 运行测试，确认它先失败**

Run: `npm --prefix frontend test -- agent-workspace-header.source.test.js`

Expected: FAIL，因为当前头部仍使用 `Segmented` 文本切换，没有 Logo 项目菜单。

- [ ] **Step 3: 补充父页面回调的回归测试**

```js
test('画布头部复用已有项目导航、创建和删除能力', () => {
  const page = read('./AgentPageV2.jsx');
  assert.match(page, /onBackHome=/);
  assert.match(page, /onShowProjects=/);
  assert.match(page, /onCreateProject=/);
  assert.match(page, /onDeleteProject=/);
  assert.match(page, /Modal\.confirm/);
});
```

- [ ] **Step 4: 重新运行测试，确认它仍然失败**

Run: `npm --prefix frontend test -- agent-workspace-header.source.test.js`

Expected: FAIL，因为父页面尚未传入左上角菜单需要的项目操作回调。

### Task 2: 实现紧凑的品牌项目工具条

**Files:**
- Modify: `frontend/src/user/components/AgentWorkspaceHeader.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Consumes: `projectTitle`, `activeView`, `onViewChange`，以及新增的 `onBackHome`, `onShowProjects`, `onCreateProject`, `onDeleteProject` 回调。
- Produces: 左上角 Logo 菜单、现有项目名称按钮、工作流图标按钮与故事板图标按钮。

- [ ] **Step 1: 改写头部组件的导入和参数**

```jsx
import { Dropdown, Tooltip } from 'antd';
import { PanelsTopLeft, Workflow } from 'lucide-react';
import { BrandLogo } from '../../shared/components/BrandLogo';

export function AgentWorkspaceHeader({
  projectTitle,
  activeView,
  onBackHome,
  onShowProjects,
  onCreateProject,
  onDeleteProject,
  // retain existing drawer, conversation, save and rename props
}) {
```

- [ ] **Step 2: 将原文字分段器替换为图标按钮组**

```jsx
<div className="agent-canvas-view-switch" role="group" aria-label="项目视图">
  <Tooltip title="工作流">
    <button type="button" aria-label="工作流" aria-pressed={activeView === 'workflow'}
      className={activeView === 'workflow' ? 'is-active' : ''}
      onClick={() => onViewChange('workflow')}><Workflow size={15} /></button>
  </Tooltip>
  <Tooltip title="故事板">
    <button type="button" aria-label="故事板" aria-pressed={activeView === 'storyboard'}
      className={activeView === 'storyboard' ? 'is-active' : ''}
      onClick={() => onViewChange('storyboard')}><PanelsTopLeft size={15} /></button>
  </Tooltip>
</div>
```

- [ ] **Step 3: 在项目标题前加入 Logo 下拉菜单**

```jsx
<Dropdown menu={{ items: [
  { key: 'home', label: '回到主页', onClick: onBackHome },
  { key: 'projects', label: '全部项目', onClick: onShowProjects },
  { type: 'divider' },
  { key: 'new', label: '创建新项目', onClick: onCreateProject },
  { key: 'delete', label: '删除项目', danger: true, disabled: !projectTitle, onClick: onDeleteProject }
] }} trigger={['click']}>
  <button type="button" className="agent-canvas-brand-menu" aria-label="项目菜单"><BrandLogo /></button>
</Dropdown>
```

- [ ] **Step 4: 调整 CSS，让左侧区域变成紧凑的横向工具条**

```css
.agent-page .agent-canvas-toolbar__project { gap: 5px; overflow: hidden; }
.agent-page .agent-canvas-brand-menu,
.agent-page .agent-canvas-view-switch { border: 1px solid var(--agent-border-strong); border-radius: 6px; background: var(--agent-elevated); }
.agent-page .agent-canvas-view-switch { display: inline-flex; padding: 3px; }
.agent-page .agent-canvas-view-switch button { display: grid; width: 26px; height: 26px; place-items: center; border: 0; border-radius: 4px; background: transparent; color: var(--agent-muted); }
.agent-page .agent-canvas-view-switch button.is-active { background: var(--agent-hover); color: var(--agent-text); }
```

- [ ] **Step 5: 添加窄屏保护样式**

```css
@media (max-width: 560px) {
  .agent-page .agent-canvas-toolbar__project { gap: 4px; }
  .agent-page .agent-canvas-project-title { min-width: 0; max-width: min(42vw, 190px); }
  .agent-page .agent-canvas-toolbar__project > .agent-canvas-toolbar-button { display: none; }
}
```

- [ ] **Step 6: 运行头部测试，确认组件和样式契约通过**

Run: `npm --prefix frontend test -- agent-workspace-header.source.test.js`

Expected: PASS。

### Task 3: 将菜单接入既有项目操作并验证状态不回退

**Files:**
- Modify: `frontend/src/user/pages/AgentPageV2.jsx`
- Modify: `frontend/src/user/pages/agent-workspace-header.source.test.js`

**Interfaces:**
- Consumes: `syncTaskUrl`, `createTask`, `deleteCurrentTask`, `updateWorkspaceView`, `workspaceViewForActiveView`。
- Produces: 可用的 Logo 菜单回调；删除项目的二次确认；不改变现有视图 URL 与画布保存队列。

- [ ] **Step 1: 在父页面实现四个菜单回调**

```jsx
function returnToAgentHome() {
  activateTask(null, { historyMode: 'push' });
}

function showAllAgentProjects() {
  updateWorkspaceView({ drawerOpen: true, drawerTab: 'nodes' });
}

function confirmDeleteCurrentTask() {
  if (!activeTaskId) return;
  Modal.confirm({
    title: '删除当前项目？',
    content: '删除后无法恢复，不会影响其他项目。',
    okText: '删除',
    cancelText: '取消',
    okButtonProps: { danger: true },
    onOk: deleteCurrentTask
  });
}
```

- [ ] **Step 2: 将回调传给头部，同时保留原有视图切换处理**

```jsx
<AgentWorkspaceHeader
  projectTitle={activeTask?.title}
  activeView={viewState.activeView}
  onBackHome={returnToAgentHome}
  onShowProjects={showAllAgentProjects}
  onCreateProject={() => createTask()}
  onDeleteProject={confirmDeleteCurrentTask}
  onViewChange={activeView => {
    const nextView = workspaceViewForActiveView(viewStateRef.current, activeView);
    updateWorkspaceView({ activeView: nextView.activeView, conversationOpen: nextView.conversationOpen });
  }}
/>
```

- [ ] **Step 3: 运行针对性测试**

Run: `npm --prefix frontend test -- agent-workspace-header.source.test.js agent-workspace-view.test.js agent-workspace.layout.test.js`

Expected: PASS，且 `workspaceViewForActiveView` 的故事板对话区恢复规则不变。

- [ ] **Step 4: 构建前端**

Run: `npm --prefix frontend run build`

Expected: exit code 0。

### Task 4: 本地真实页面验收

**Files:**
- No source changes expected.

**Interfaces:**
- Consumes: 已重建并重启的本地 Go 嵌入式运行时。
- Produces: 已登录本地页面上的真实可见验收结果。

- [ ] **Step 1: 重建并重启本地 Go 页面服务**

Run: `scripts/build-go-local.ps1`，然后以现有本地审查服务方式重启，并执行 `scripts/check-local-go.ps1`。

Expected: `http://localhost:4001` 可访问，并使用新的嵌入式前端资源。

- [ ] **Step 2: 在已有登录状态下打开画布项目**

Open: `http://localhost:4001/agent/canvas?taskId=<现有任务编号>&view=workflow`

Expected: 左上角从左到右显示品牌 Logo、项目名、工作流图标、故事板图标；无横向溢出。

- [ ] **Step 3: 验证工作流和故事板切换**

Click: 两个图标按钮各一次。

Expected: 当前按钮状态正确改变，地址中的 `view` 分别为 `workflow` 与 `storyboard`，刷新后仍显示对应视图。

- [ ] **Step 4: 验证 Logo 菜单和删除保护**

Click: Logo，再点击“全部项目”、“创建新项目”和“删除项目”。

Expected: 菜单四项可见；全部项目打开已有项目栏；创建项目调用现有创建流程；删除先显示确认框，点取消后当前项目不变。

- [ ] **Step 5: 验证窄屏**

Viewport: 390px 宽度。

Expected: Logo 和两个视图图标保持可见，项目名截断但不越界，画布工具和对话区仍能使用。
