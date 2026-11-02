# Agent 项目对话区改版 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 Agent 项目对话区改为带有明确空状态引导、紧凑图标工具条和固定输入区的导演工作台，同时保留原有会话与执行功能。

**Architecture:** 在 `AgentPageV2.jsx` 中将无消息项目的引导内容抽为可点击的对话建议，并复用既有 `question` 状态、输入框 ref 和 `sendQuestion`。项目对话的底部工具条复用既有 `openComposerTool`、模式、附件、技能、专家和找问题状态，仅将文字按钮改为主下拉加图标；样式仅在 `agent-workspace.css` 中扩展项目对话面板，不改变画布、素材区、顶部工具栏或后端接口。

**Tech Stack:** React、Ant Design、lucide-react、CSS、Node test、Vite、Go embed 本地预览。

## Global Constraints

- 只改项目对话面板；工作流顶部工具栏、画布、故事板、素材区和后端接口不变。
- 引导入口只填入并聚焦现有输入框，不能自动发送、创建项目或调用外部接口。
- 原有导演模式、待确认方案、素材引用、节点引用、执行记录和发送行为必须保持。
- 本地 `4001` 预览通过重新构建 React 并打包进 Go 二进制更新；不发布公网。

---

### Task 1: 项目对话空状态引导

**Files:**
- Modify: `frontend/src/user/pages/AgentPageV2.jsx`
- Modify: `frontend/src/user/pages/agent-workspace.layout.test.js`

**Interfaces:**
- Consumes: `question`、`setQuestion`、`questionInputRef` 与现有 `sendQuestion`。
- Produces: `handleConversationStarter(prompt)`，供四个引导入口调用。

- [ ] **Step 1: 写失败的源码结构测试**

```js
assert.match(source, /function handleConversationStarter\(prompt\)/);
assert.match(source, /agent-conversation-starters/);
assert.match(source, /画布创作/);
assert.match(source, /从素材生成/);
assert.match(source, /上传文本改编/);
assert.match(source, /优化提示词/);
```

- [ ] **Step 2: 运行测试确认失败**

Run: `node --test frontend/src/user/pages/agent-workspace.layout.test.js`

Expected: FAIL，缺少项目对话引导结构。

- [ ] **Step 3: 实现最小行为**

```jsx
function handleConversationStarter(prompt) {
  setQuestion(prompt);
  requestAnimationFrame(() => questionInputRef.current?.focus());
}
```

只在 `activeTask && messages.length === 0` 的空状态中渲染四个按钮，点击按钮调用该函数。

- [ ] **Step 4: 运行测试确认通过**

Run: `node --test frontend/src/user/pages/agent-workspace.layout.test.js`

Expected: PASS。

### Task 2: 会话面板视觉与窄屏适配

**Files:**
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Modify: `frontend/src/user/pages/agent-workspace.layout.test.js`

**Interfaces:**
- Consumes: Task 1 的 `agent-conversation-starters`、既有 `agent-workbench-empty` 与 `agent-workbench-composer`。
- Produces: 统一的空状态、入口按钮和底部输入区布局规则。

- [ ] **Step 1: 写失败的源码结构测试**

```js
assert.match(stylesheet, /\.agent-page \.agent-conversation-starters\s*\{/);
assert.match(stylesheet, /\.agent-page \.agent-conversation-starter\s*\{/);
assert.match(stylesheet, /@media \(max-width: 560px\)[\s\S]*agent-conversation-starters/);
```

- [ ] **Step 2: 运行测试确认失败**

Run: `node --test frontend/src/user/pages/agent-workspace.layout.test.js`

Expected: FAIL，缺少引导入口样式与窄屏规则。

- [ ] **Step 3: 实现最小样式**

```css
.agent-page .agent-conversation-starters { display: grid; gap: 8px; }
.agent-page .agent-conversation-starter { display: grid; grid-template-columns: 18px minmax(0, 1fr) 14px; }
```

入口采用现有主题变量，空状态在面板中居中，输入区继续使用现有的底部 flex 布局；窄屏下入口保持单列且不遮挡输入区。

- [ ] **Step 4: 运行测试确认通过**

Run: `node --test frontend/src/user/pages/agent-workspace.layout.test.js`

Expected: PASS。

### Task 3: 紧凑输入工具条

**Files:**
- Modify: `frontend/src/user/pages/AgentPageV2.jsx`
- Modify: `frontend/src/shared/styles/agent-workspace.css`
- Modify: `frontend/src/user/pages/agent-workspace.layout.test.js`

**Interfaces:**
- Consumes: 既有 `openComposerTool(tool)`、`setProblemSearchMode`、`sendQuestion`、`selectedMode`。
- Produces: 只用于项目对话面板的 `agent-composer-tools`，包含 `+`、主模式下拉、附件/技能/专家图标与圆形发送按钮。

- [ ] **Step 1: 写失败的源码结构测试**

```js
assert.match(pageSource, /agent-composer-tools/);
assert.match(pageSource, /全能创作/);
assert.match(pageSource, /aria-label="添加附件"/);
assert.match(pageSource, /aria-label="选择技能"/);
assert.match(pageSource, /aria-label="选择专家"/);
assert.match(pageSource, /aria-label="发送消息"/);
```

- [ ] **Step 2: 运行测试确认失败**

Run: `node --test frontend/src/user/pages/agent-workspace.layout.test.js`

Expected: FAIL，缺少紧凑工具条结构。

- [ ] **Step 3: 实现最小行为**

使用现有 `+` 菜单承载更多能力，并加入“找问题”菜单项；主模式下拉继续调用 `openComposerTool('mode')`；三个图标分别调用 `openComposerTool('file')`、`openComposerTool('skill')` 和 `openComposerTool('expert')`；发送使用现有 `sendQuestion()`。

- [ ] **Step 4: 实现局部样式**

```css
.agent-page .agent-conversation-panel .agent-composer-tools { display: flex; }
.agent-page .agent-conversation-panel .agent-composer-send { border-radius: 50%; }
```

不修改创作首页的 `agent-composer-actions`，窄屏时工具条保持单行横向排列。

- [ ] **Step 5: 运行测试确认通过**

Run: `node --test frontend/src/user/pages/agent-workspace.layout.test.js`

Expected: PASS。

### Task 4: 构建与真实界面验收

**Files:**
- Verify: `frontend/src/user/pages/AgentPageV2.jsx`
- Verify: `frontend/src/shared/styles/agent-workspace.css`

**Interfaces:**
- Consumes: Task 1 与 Task 2 的前端改动。
- Produces: 重新打包的本地 Go 预览服务。

- [ ] **Step 1: 运行相关源码测试**

Run: `node --test frontend/src/user/pages/agent-workspace.layout.test.js frontend/src/shared/layouts/UserLayout.source.test.js`

Expected: PASS。

- [ ] **Step 2: 编译前端**

Run: `npm --prefix frontend run build`

Expected: exit code 0。

- [ ] **Step 3: 重建本地 Go 预览**

Run: `scripts/build-go-local.ps1`，然后用 `scripts/start-local-go-review.ps1 -UseYcmMysql` 启动 `4001`。

Expected: 本地 Go 服务监听 `4001`。

- [ ] **Step 4: 检查服务与实际界面**

Run: `scripts/check-local-go.ps1`，并在已登录的本地浏览器中验证工作流和故事板：无消息项目显示四个入口，点击只填入输入框，有消息项目仍可发送。

Expected: 健康检查通过；不出现顶部工具栏改动、遮挡或溢出。
