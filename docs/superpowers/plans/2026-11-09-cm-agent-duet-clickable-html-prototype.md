# CM Agent 与 Duet 可点击 HTML 原型 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付一个无需构建和联网、打开即可点击的一战晟铭 / Agent代理 / Duet 单文件 HTML 原型，完整演示已批准的入口、状态和业务链路。

**Architecture:** 运行时只有 `docs/prototypes/cm-agent-duet/index.html`，CSS、演示数据和 JavaScript 全部内嵌。页面由一个可序列化状态对象驱动，使用事件委托更新状态并重新渲染；`localStorage` 仅保存原型偏好和演示状态。独立 Node 源码契约测试只读取 HTML，不是运行时依赖；最终使用本地静态服务器和浏览器逐项点击验收。

**Tech Stack:** HTML5、CSS 自定义属性、原生 JavaScript、`localStorage`、Node.js 内置 `node:test`、浏览器手动验收。

## Global Constraints

- 运行时交付物必须是单个 `docs/prototypes/cm-agent-duet/index.html`，不得依赖 npm 构建、CDN、网络字体或外部脚本。
- 左上角固定提供“一战晟铭”“Agent代理”“Duet”三个工作模式；不得新增“内容创作”模式。
- 所有用户可见参考品牌名必须转换为 CM、小Y、CM Duet；原型页面和演示数据不得出现未获豁免的参考品牌名。
- 原型不得请求真实 API、连接模型、MCP、Codex、数据库、TOS 或本地执行器，也不得消耗额度。
- 系统提示词不得写入 HTML；技能详情只展示用途、输入输出和工具需求。
- 原型不得修改 `localhost:4001` 正式页面，不进入 Go embed。
- 所有主要按钮必须有可见反馈；不允许只有外观、点击无响应的空按钮。
- 状态机固定为 `idle → planning → awaiting_approval → running → succeeded / failed / cancelled`。
- 至少支持深夜黑、明亮白、雪山蓝三个主题，并在 1366×768 与 1920×1080 下保持主操作可见。
- 不新建 Git 分支，不推送远端，不部署公网。

---

## File Map

- Create: `docs/prototypes/cm-agent-duet/index.html` — 唯一运行时文件，包含结构、样式、演示数据、状态、渲染和全部交互。
- Create: `docs/prototypes/cm-agent-duet/prototype.source.test.mjs` — Node 内置测试，检查单文件约束、品牌、入口、状态机和关键事件契约。
- Modify: `docs/superpowers/plans/2026-11-09-cm-agent-duet-clickable-html-prototype.md` — 实施时勾选已完成步骤。

## Shared Runtime Interfaces

所有任务都使用以下稳定接口，不得在后续任务中另起一套状态：

```js
const STORAGE_KEY = 'cm-agent-duet-prototype-v1';

const defaultState = {
  mode: 'brand',
  theme: 'dark',
  pages: { brand: 'home', agent: 'new-session', duet: 'channels' },
  activeProjectId: 'project-water-margin',
  activeChannelId: 'channel-screenplay',
  activeWorkspaceView: 'canvas',
  skillTab: 'skills',
  enabledSkillIds: ['cm-creative-advisor', 'storyboard-director'],
  selectedContext: [],
  permissionMode: 'auto',
  run: { status: 'idle', step: 0, error: '' },
  automations: [],
  duetTasks: []
};

function loadState() {}
function saveState() {}
function patchState(patch) {}
function renderApp() {}
function dispatchAction(action, payload = {}) {}

window.__CM_PROTOTYPE__ = {
  getState: () => structuredClone(state),
  reset: () => resetState(),
  dispatch: (action, payload) => dispatchAction(action, payload)
};
```

所有可点击元素使用 `data-action`，需要参数时增加 `data-id`、`data-mode`、`data-page` 或 `data-value`。根节点固定为 `#app`，通知区固定为 `#toast-region`。

---

### Task 1: 单文件外壳、模式切换、主题与持久化

**Files:**
- Create: `docs/prototypes/cm-agent-duet/index.html`
- Create: `docs/prototypes/cm-agent-duet/prototype.source.test.mjs`

**Interfaces:**
- Consumes: 无。
- Produces: `defaultState`、`loadState()`、`saveState()`、`patchState()`、`renderApp()`、`dispatchAction()`、`window.__CM_PROTOTYPE__`，供后续所有任务扩展。

- [x] **Step 1: 写模式、主题和单文件契约的失败测试**

```js
import test from 'node:test';
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';

const html = await readFile(new URL('./index.html', import.meta.url), 'utf8');

test('prototype is self-contained and exposes the three approved modes', () => {
  assert.match(html, /<style>[\s\S]*<\/style>/);
  assert.match(html, /<script>[\s\S]*<\/script>/);
  assert.doesNotMatch(html, /<script[^>]+src=/);
  assert.doesNotMatch(html, /<link[^>]+stylesheet/);
  for (const label of ['一战晟铭', 'Agent代理', 'Duet']) assert.match(html, new RegExp(label));
  assert.doesNotMatch(html, />\s*内容创作\s*</);
});

test('prototype publishes state inspection and persistence contracts', () => {
  for (const token of ['cm-agent-duet-prototype-v1', 'loadState', 'saveState', 'patchState', 'renderApp', 'dispatchAction', '__CM_PROTOTYPE__']) {
    assert.match(html, new RegExp(token));
  }
});

test('prototype includes three approved themes without external assets', () => {
  for (const theme of ['dark', 'light', 'snow']) assert.match(html, new RegExp(`data-theme=["']?${theme}|theme:\\s*["']${theme}`));
  assert.doesNotMatch(html, /https?:\/\//);
});
```

- [x] **Step 2: 运行测试并确认因 HTML 不存在而失败**

Run: `node --test docs/prototypes/cm-agent-duet/prototype.source.test.mjs`

Expected: FAIL，错误包含 `ENOENT` 和 `index.html`。

- [x] **Step 3: 实现页面外壳、统一状态和三模式导航**

在 `index.html` 中实现：

```html
<div id="app"></div>
<div id="toast-region" role="status" aria-live="polite"></div>
<script>
const STORAGE_KEY = 'cm-agent-duet-prototype-v1';
const modes = [
  { id: 'brand', label: '一战晟铭' },
  { id: 'agent', label: 'Agent代理' },
  { id: 'duet', label: 'Duet' }
];
function patchState(patch) {
  state = { ...state, ...patch };
  saveState();
  renderApp();
}
document.addEventListener('click', event => {
  const target = event.target.closest('[data-action]');
  if (target) dispatchAction(target.dataset.action, target.dataset);
});
</script>
```

模式切换必须替换整条侧栏内容；三个模式分别记住 `pages.brand`、`pages.agent` 和 `pages.duet`。主题切换通过 `document.documentElement.dataset.theme` 和 CSS 语义变量驱动。

- [x] **Step 4: 运行源测试并确认通过**

Run: `node --test docs/prototypes/cm-agent-duet/prototype.source.test.mjs`

Expected: 3 tests PASS。

- [x] **Step 5: 浏览器冒烟检查模式、主题和刷新恢复**

Run: `python -m http.server 4178 --directory docs/prototypes/cm-agent-duet`

检查 `http://127.0.0.1:4178/`：依次切换三个模式和三个主题，刷新后仍停留在刷新前模式与主题。

- [x] **Step 6: 提交外壳**

```bash
git add docs/prototypes/cm-agent-duet/index.html docs/prototypes/cm-agent-duet/prototype.source.test.mjs
git commit -m "feat: add CM workspace prototype shell"
```

---

### Task 2: Agent 技能中心、选择器和权限模式

**Files:**
- Modify: `docs/prototypes/cm-agent-duet/index.html`
- Modify: `docs/prototypes/cm-agent-duet/prototype.source.test.mjs`

**Interfaces:**
- Consumes: `state.skillTab`、`state.enabledSkillIds`、`state.selectedContext`、`state.permissionMode`、`dispatchAction()`。
- Produces: `renderSkillCenter()`、`renderSkillDetail()`、`renderContextPicker(kind)` 和动作 `SET_SKILL_TAB`、`TOGGLE_SKILL`、`TOGGLE_ALL_SKILLS`、`OPEN_SKILL_DETAIL`、`ADD_CONTEXT`、`REMOVE_CONTEXT`、`SET_PERMISSION_MODE`。

- [x] **Step 1: 增加技能和选择器契约的失败测试**

```js
test('agent skill center and composer contracts are present', () => {
  for (const token of [
    'Skills', 'MCP', 'Codex', '搜索技能', '全部开启', '导入 Skill',
    'renderSkillCenter', 'renderSkillDetail', 'renderContextPicker',
    'TOGGLE_SKILL', 'ADD_CONTEXT', 'REMOVE_CONTEXT', 'SET_PERMISSION_MODE'
  ]) assert.match(html, new RegExp(token));
  for (const label of ['计划模式', '自动审批', '完全运行']) assert.match(html, new RegExp(label));
});
```

- [x] **Step 2: 运行测试并确认失败**

Run: `node --test docs/prototypes/cm-agent-duet/prototype.source.test.mjs`

Expected: 新测试 FAIL，缺少 `renderSkillCenter`。

- [x] **Step 3: 实现技能演示数据与管理页面**

内嵌至少 8 个技能，覆盖平台、个人、项目三种来源和可用、缺配置、无权限三种环境状态。每项结构固定为：

```js
{
  id: 'cm-creative-advisor',
  name: 'CM 创作顾问',
  version: '1.6.0',
  summary: '整理创作目标并生成可执行计划',
  source: '平台',
  author: 'CM',
  category: '写小说',
  environment: 'ready',
  inputs: ['创作目标', '项目资料'],
  outputs: ['创作计划', '修改建议'],
  tools: ['读取项目资料']
}
```

实现 Skills/MCP/Codex 标签、搜索、分类、全部启停、单项启停、详情抽屉、导入弹窗和管理执行器弹窗。导入按钮只展示“演示模式，不会上传文件”。

- [x] **Step 4: 实现 `/`、`@` 和权限联动**

`renderContextPicker('slash')` 只显示已启用且环境可用的技能，并显示 MCP、自动任务、思考强度；`renderContextPicker('mention')` 显示技能、自动任务、MCP、文件、素材、画布节点和故事板镜头。关闭技能后重新打开 `/`，该技能必须消失。

- [x] **Step 5: 运行测试并浏览器验收联动**

Run: `node --test docs/prototypes/cm-agent-duet/prototype.source.test.mjs`

Expected: all tests PASS。

浏览器：关闭“CM 创作顾问”→返回新会话→打开 `/`，确认不再显示；重新开启后恢复。切换三档权限，说明文案同步变化。

- [x] **Step 6: 提交技能中心**

```bash
git add docs/prototypes/cm-agent-duet/index.html docs/prototypes/cm-agent-duet/prototype.source.test.mjs
git commit -m "feat: add clickable Agent skill center"
```

---

### Task 3: Agent 会话、自动任务、项目四视图和运行状态机

**Files:**
- Modify: `docs/prototypes/cm-agent-duet/index.html`
- Modify: `docs/prototypes/cm-agent-duet/prototype.source.test.mjs`

**Interfaces:**
- Consumes: `state.pages.agent`、`state.activeProjectId`、`state.activeWorkspaceView`、`state.run`、`state.automations`、上下文和权限接口。
- Produces: `renderAgentSession()`、`renderAutomationPage()`、`renderProjectWorkspace()`、`renderRunTimeline()`、`advanceRun()` 和动作 `START_RUN`、`APPROVE_RUN`、`CANCEL_RUN`、`FAIL_RUN`、`RETRY_RUN`、`SET_WORKSPACE_VIEW`、`CREATE_AUTOMATION`、`TOGGLE_AUTOMATION`、`RUN_AUTOMATION`。

- [x] **Step 1: 增加 Agent 完整链路的失败测试**

```js
test('agent session, workspace and run state machine are represented', () => {
  for (const token of [
    'renderAgentSession', 'renderAutomationPage', 'renderProjectWorkspace', 'renderRunTimeline', 'advanceRun',
    'START_RUN', 'APPROVE_RUN', 'CANCEL_RUN', 'FAIL_RUN', 'RETRY_RUN',
    '画布', '故事板', '素材', '任务', '等待确认', '继续执行', '从失败处重试'
  ]) assert.match(html, new RegExp(token));
  for (const status of ['idle', 'planning', 'awaiting_approval', 'running', 'succeeded', 'failed', 'cancelled']) {
    assert.match(html, new RegExp(status));
  }
});
```

- [x] **Step 2: 运行测试并确认失败**

Run: `node --test docs/prototypes/cm-agent-duet/prototype.source.test.mjs`

Expected: 新测试 FAIL，缺少 `renderAgentSession`。

- [x] **Step 3: 实现新会话和可控运行状态机**

实现以下明确行为：

```js
function startRun() {
  patchState({ run: { status: 'planning', step: 1, error: '' } });
  scheduleRunAdvance();
}
function advanceRun() {
  if (state.permissionMode === 'plan') return patchRun('succeeded', 2);
  if (state.permissionMode === 'auto') return patchRun('awaiting_approval', 3);
  return patchRun('running', 4);
}
```

完全运行在短计时器后进入 `succeeded`；自动审批等待点击“继续执行”；计划模式停在计划结果。取消、故障演示和重试必须保留已经完成的步骤。

- [x] **Step 4: 实现自动任务和项目四视图**

自动任务弹窗创建包含名称、工作区、目标、权限、技能、周期和状态的演示记录。项目工作区四个视图共用同一个顶部工具条，工具条始终显示技能、权限、附件、`/` 与 `@`。

- [x] **Step 5: 实现产物定位**

执行成功后产生稳定演示产物：

```js
{ id: 'artifact-storyboard-01', kind: 'storyboard', title: '第 1 集故事板', targetView: 'storyboard' }
```

点击产物调用 `SET_WORKSPACE_VIEW` 并高亮对应镜头、节点或素材卡片。

- [x] **Step 6: 运行测试并浏览器验收 Agent 主链**

Run: `node --test docs/prototypes/cm-agent-duet/prototype.source.test.mjs`

Expected: all tests PASS。

浏览器依次验收计划模式、自动审批和完全运行；再验收取消、故障演示、重试、自动任务创建和四视图工具条一致性。

- [x] **Step 7: 提交 Agent 主链**

```bash
git add docs/prototypes/cm-agent-duet/index.html docs/prototypes/cm-agent-duet/prototype.source.test.mjs
git commit -m "feat: simulate Agent projects and runs"
```

---

### Task 4: Duet 频道、员工、任务、审核返工和接力

**Files:**
- Modify: `docs/prototypes/cm-agent-duet/index.html`
- Modify: `docs/prototypes/cm-agent-duet/prototype.source.test.mjs`

**Interfaces:**
- Consumes: `state.pages.duet`、`state.activeProjectId`、`state.activeChannelId`、`state.duetTasks` 和共享项目/产物 ID。
- Produces: `renderDuetWorkspace()`、`renderDuetTaskPanel()`、`renderRelayBoard()`、`renderDuetSchedules()` 和动作 `SET_DUET_CHANNEL`、`CREATE_DUET_TASK`、`HANDOFF_DUET_TASK`、`SUPPLEMENT_DUET_TASK`、`REVIEW_DUET_TASK`、`REWORK_DUET_TASK`、`TOGGLE_DUET_SCHEDULE`。

- [ ] **Step 1: 增加 Duet 业务链的失败测试**

```js
test('Duet exposes channel production and controlled relay interactions', () => {
  for (const token of [
    'CM Duet', '剧本生产线', '资产工坊', '分镜台', '员工', '接力看板', '定时任务',
    'renderDuetWorkspace', 'renderDuetTaskPanel', 'renderRelayBoard', 'renderDuetSchedules',
    'CREATE_DUET_TASK', 'HANDOFF_DUET_TASK', 'SUPPLEMENT_DUET_TASK',
    'REVIEW_DUET_TASK', 'REWORK_DUET_TASK', 'TOGGLE_DUET_SCHEDULE'
  ]) assert.match(html, new RegExp(token));
});
```

- [ ] **Step 2: 运行测试并确认失败**

Run: `node --test docs/prototypes/cm-agent-duet/prototype.source.test.mjs`

Expected: 新测试 FAIL，缺少 `renderDuetWorkspace`。

- [ ] **Step 3: 实现频道与员工视图**

内嵌 CM Duet、编剧、审校、资产和分镜员工。员工只包含演示职责、状态、能力和当前任务，不包含系统提示词。频道和 Agent 项目共用 `project-water-margin`。

- [ ] **Step 4: 实现任务、转交、补充、审核与返工**

Duet 任务结构固定为：

```js
{
  id: 'duet-task-screenplay-01',
  projectId: 'project-water-margin',
  channelId: 'channel-screenplay',
  employeeId: 'employee-writer',
  title: '改编第 1—3 集剧本',
  status: 'running',
  round: 1,
  maxRounds: 3,
  notes: [],
  artifactIds: []
}
```

审核通过推进接力；退回返工增加 `round`；达到 `maxRounds` 后进入等待用户决定，不再自动推进。

- [ ] **Step 5: 实现接力看板和定时任务**

接力步骤固定演示“写剧本 → 审剧本 → 出资产 → 审资产 → 写分镜 → 审分镜”。每一步可显示待办、执行、通过、返工和失败。定时任务可新建、暂停和恢复，并保存到 `localStorage`。

- [ ] **Step 6: 运行测试并浏览器验收 Duet 链路**

Run: `node --test docs/prototypes/cm-agent-duet/prototype.source.test.mjs`

Expected: all tests PASS。

浏览器：新建任务→转交员工→补充要求→退回返工→审核通过→查看接力下一步→暂停和恢复定时任务。

- [ ] **Step 7: 提交 Duet 主链**

```bash
git add docs/prototypes/cm-agent-duet/index.html docs/prototypes/cm-agent-duet/prototype.source.test.mjs
git commit -m "feat: add clickable CM Duet workflow"
```

---

### Task 5: 响应式、可访问性、品牌检查与最终浏览器验收

**Files:**
- Modify: `docs/prototypes/cm-agent-duet/index.html`
- Modify: `docs/prototypes/cm-agent-duet/prototype.source.test.mjs`
- Modify: `docs/superpowers/plans/2026-11-09-cm-agent-duet-clickable-html-prototype.md`

**Interfaces:**
- Consumes: 所有前述渲染函数和动作。
- Produces: 可交付原型、完整测试证据和已勾选计划。

- [ ] **Step 1: 增加品牌、无网络和响应式契约测试**

```js
test('prototype has no forbidden product strings or network calls', () => {
  for (const forbidden of ['洛水', '小洛', '洛Duet']) assert.doesNotMatch(html, new RegExp(forbidden));
  for (const networkCall of ['fetch(', 'XMLHttpRequest', 'WebSocket(', 'EventSource(']) {
    assert.doesNotMatch(html, new RegExp(networkCall.replace('(', '\\(')));
  }
});

test('prototype includes responsive and accessibility contracts', () => {
  assert.match(html, /@media\s*\(max-width:\s*1366px\)/);
  assert.match(html, /aria-live="polite"/);
  assert.match(html, /aria-label=/);
  assert.match(html, /Escape/);
});
```

- [ ] **Step 2: 运行测试并确认缺失项失败**

Run: `node --test docs/prototypes/cm-agent-duet/prototype.source.test.mjs`

Expected: 响应式或 Escape 契约测试 FAIL。

- [ ] **Step 3: 完成窄屏、键盘和弹层行为**

在 1366px 以下压缩侧栏和间距；在 1100px 以下将右侧执行区改为抽屉。为下拉、标签、抽屉和弹窗设置 `aria-label`、焦点样式和 Esc 关闭。危险确认弹窗不允许点击遮罩直接确认。

- [ ] **Step 4: 运行完整源测试和仓库格式检查**

Run: `node --test docs/prototypes/cm-agent-duet/prototype.source.test.mjs`

Expected: all tests PASS。

Run: `git diff --check`

Expected: no output，exit code 0。

- [ ] **Step 5: 启动静态服务器并完成两个视口的浏览器验收**

Run: `python -m http.server 4178 --directory docs/prototypes/cm-agent-duet`

在 1920×1080 和 1366×768 逐项执行设计规格第 10 节的 22 条验收。特别确认：

- 左上角模式切换替换整条导航；
- Agent 故事板仍有技能和权限入口；
- 技能启停与 `/` 联动；
- 三档权限产生不同停顿行为；
- Duet 审核、返工和最大轮次可见；
- 刷新恢复当前模式、项目、主题和任务；
- 浏览器网络面板没有原型发起的业务请求。

- [ ] **Step 6: 复核文件和工作树**

Run: `git status --short`

Expected: 只显示本任务尚未提交的 `index.html`、测试文件和计划勾选更新；不得包含无关文件。

- [ ] **Step 7: 提交最终验收修正**

```bash
git add docs/prototypes/cm-agent-duet/index.html docs/prototypes/cm-agent-duet/prototype.source.test.mjs docs/superpowers/plans/2026-11-09-cm-agent-duet-clickable-html-prototype.md
git commit -m "test: verify CM Agent and Duet prototype"
```

- [ ] **Step 8: 最终交付说明**

交付时提供：

- 可点击的绝对文件链接；
- 本地打开地址；
- 通过的测试命令；
- 浏览器验收范围；
- 明确说明这是交互原型，不是已接后端的正式功能。
