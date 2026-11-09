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

test('agent skill center and composer contracts are present', () => {
  for (const token of [
    'Skills', 'MCP', 'Codex', '搜索技能', '全部开启', '导入 Skill',
    'renderSkillCenter', 'renderSkillDetail', 'renderContextPicker',
    'TOGGLE_SKILL', 'ADD_CONTEXT', 'REMOVE_CONTEXT', 'SET_PERMISSION_MODE'
  ]) assert.match(html, new RegExp(token));
  for (const label of ['计划模式', '自动审批', '完全运行']) assert.match(html, new RegExp(label));
});

test('agent session, workspace and run state machine are represented', () => {
  for (const token of [
    'renderAgentSession', 'renderAutomationPage', 'renderAgentWorkspace', 'renderRunTimeline', 'advanceRun',
    'START_RUN', 'APPROVE_RUN', 'CANCEL_RUN', 'FAIL_RUN', 'RETRY_RUN',
    '画布', '故事板', '素材', '任务', '等待确认', '继续执行', '从失败处重试'
  ]) assert.match(html, new RegExp(token));
  for (const status of ['idle', 'planning', 'awaiting_approval', 'running', 'succeeded', 'failed', 'cancelled']) assert.match(html, new RegExp(status));
});

test('Agent proxy hands completed work to the existing Agent workspace', () => {
  for (const token of [
    'Agent工作区', 'Agent代理负责计划与调度', '一战晟铭原有生产工作台',
    '发送到Agent工作区', 'SEND_TO_AGENT_WORKSPACE', 'renderAgentWorkspace'
  ]) assert.match(html, new RegExp(token));
  assert.doesNotMatch(html, /打开第 1 集故事板/);
});

test('main workspace exposes a clickable AI-CUT production flow', () => {
  for (const token of [
    'AI剪辑视频', 'AI-CUT', 'renderAiCutPage', '导入视频', '智能拆分',
    '字幕纠错', 'BGM', '剪映草稿', 'START_AI_CUT_DEMO', 'NEXT_AI_CUT_STAGE'
  ]) assert.match(html, new RegExp(token));
});

test('plugin center is separate from Agent skills and has managed state', () => {
  for (const token of [
    '扩展与资源', '插件中心', 'renderPluginCenter', 'enabledPluginIds',
    'TOGGLE_PLUGIN', 'INSTALL_PLUGIN', '插件提供工具，Agent技能负责编排调用'
  ]) assert.match(html, new RegExp(token));
});

test('main navigation omits unwanted community and support entries', () => {
  assert.match(html, /扩展与资源/);
  for (const removed of ['社区与服务', '社区', '反馈', '在线客服']) {
    assert.doesNotMatch(html, new RegExp(removed));
  }
});

test('grouped sidebar uses one scroll surface instead of nested nav scrollbars', () => {
  assert.match(html, /class="sidebar-content"/);
  assert.match(html, /\.sidebar-content\s*\{[^}]*overflow-y:\s*auto/s);
  assert.match(html, /\.sidebar-content \.side-nav\s*\{[^}]*overflow:\s*visible/s);
});

test('Duet exposes channel production and controlled relay interactions', () => {
  for (const token of [
    'CM Duet', '剧本生产线', '资产工坊', '分镜台', '员工', '接力看板', '定时任务',
    'renderDuetWorkspace', 'renderDuetTaskPanel', 'renderRelayBoard', 'renderDuetSchedules',
    'CREATE_DUET_TASK', 'HANDOFF_DUET_TASK', 'SUPPLEMENT_DUET_TASK',
    'REVIEW_DUET_TASK', 'REWORK_DUET_TASK', 'TOGGLE_DUET_SCHEDULE'
  ]) assert.match(html, new RegExp(token));
});

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
