import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const read = relativePath => fs.readFileSync(path.resolve(dirname, relativePath), 'utf8');

test('工作区侧栏提供节点素材搜索和拖动改宽', () => {
  const drawer = read('../components/AgentWorkspaceDrawer.jsx');
  assert.match(drawer, /画布节点/);
  assert.match(drawer, /项目素材/);
  assert.match(drawer, /setPointerCapture/);
  assert.match(drawer, /onWidthChange/);
  assert.match(drawer, /draggable/);
});

test('页面读取项目素材并把定位与创建动作交给真实画布', () => {
  const page = read('./AgentPageV2.jsx');
  assert.match(page, /listAgentAssets\(activeTaskId\)/);
  assert.match(page, /canvasRef\.current\?\.locateNode/);
  assert.match(page, /canvasRef\.current\?\.createFromAsset/);
});

test('抽屉和显示开关只更新本地工作区状态，不触发画布保存', () => {
  const page = read('./AgentPageV2.jsx');
  const workspaceUpdate = page.match(/function updateWorkspaceView\(patch\) \{([\s\S]*?)\n  \}/);
  assert.ok(workspaceUpdate, 'workspace view update handler should exist');
  assert.match(workspaceUpdate[1], /setViewState\(nextViewState\)/);
  assert.doesNotMatch(workspaceUpdate[1], /canvasSaveQueueRef\.current\?\.enqueue/);
});

test('项目抽屉提供真实生成历史并复用当前项目执行记录', () => {
  const drawer = read('../components/AgentWorkspaceDrawer.jsx');
  const canvas = read('../components/AgentCanvas.jsx');
  const page = read('./AgentPageV2.jsx');
  assert.match(drawer, /生成历史/);
  assert.match(drawer, /executions = \[\]/);
  assert.match(drawer, /execution\.status/);
  assert.doesNotMatch(canvas, /label="生成历史"/);
  assert.doesNotMatch(canvas, /onOpenHistory/);
  assert.match(page, /executions=\{executions\}/);
  assert.match(page, /activeTab=\{viewState\.drawerTab\}/);
});

test('结果库按镜头显示版本标签并可定位对应工作流节点', () => {
  const drawer = read('../components/AgentWorkspaceDrawer.jsx');
  const page = read('./AgentPageV2.jsx');
  assert.match(drawer, /groupResultLibrary/);
  assert.match(drawer, /lineageLabel/);
  assert.match(drawer, /onLocateNode\?\.\(item\.nodeId\)/);
  assert.match(page, /listAgentGenerations\(activeTaskId\)/);
  assert.match(page, /generations=\{generations\}/);
	assert.match(page, /shots=\{storyboardShots\}/);
});

test('结果库可将非视频历史结果设为当前并先同步画布', () => {
	const drawer = read('../components/AgentWorkspaceDrawer.jsx');
	const page = read('./AgentPageV2.jsx');
	assert.match(drawer, /onActivateGenerationResult/);
	assert.match(drawer, /设为当前/);
	assert.match(page, /activateAgentGenerationResult/);
	assert.match(page, /flushRequired\(\)/);
	assert.match(page, /onActivateGenerationResult=\{handleGenerationResultActivate\}/);
});

test('项目菜单统一打开抽屉后可切换当前项目素材库', () => {
  const header = read('../components/AgentWorkspaceHeader.jsx');
  const drawer = read('../components/AgentWorkspaceDrawer.jsx');
  const page = read('./AgentPageV2.jsx');
  assert.match(header, /key: 'projects'/);
  assert.match(page, /showAllAgentProjects/);
  assert.match(drawer, /key: 'assets'/);
});

test('项目素材抽屉为图片和视频显示真实缩略图，其他媒介保留类型图标', () => {
  const drawer = read('../components/AgentWorkspaceDrawer.jsx');
  const css = read('../../shared/styles/agent-workspace.css');
  assert.match(drawer, /AuthenticatedAgentImage/);
  assert.match(drawer, /AuthenticatedAgentVideo/);
  assert.match(drawer, /agent-workspace-assets__preview/);
  assert.match(css, /agent-workspace-assets__preview/);
});

test('工作流画布优先，节点抽屉默认收起并从项目菜单打开', () => {
  const view = read('./agent-workspace-view.js');
  const canvas = read('../components/AgentCanvas.jsx');
  const page = read('./AgentPageV2.jsx');
  assert.match(view, /drawerOpen:\s*false/);
  assert.doesNotMatch(canvas, /label="画布节点"/);
  assert.doesNotMatch(canvas, /onOpenNodes/);
  assert.match(page, /showAllAgentProjects/);
  assert.match(page, /drawerTab: 'nodes'/);
});

test('切换工作流或故事板会把计算后的抽屉状态一并写回', () => {
  const page = read('./AgentPageV2.jsx');
  assert.match(page, /drawerOpen:\s*nextView\.drawerOpen/);
});

test('隐藏的连线目标仍然能够接住一次连续的拖线操作', () => {
  const css = read('../../shared/styles/agent-workspace.css');
  const targetHandleRule = css.match(/\.agent-page \.react-flow__handle\.agent-canvas-node-target \{[^}]*\}/)?.[0] || '';
  assert.match(targetHandleRule, /pointer-events:\s*auto/);
  assert.doesNotMatch(targetHandleRule, /pointer-events:\s*none/);
});

test('画布顶部栏使用无状态的原生悬浮说明，避免工具提示反复重排', () => {
  const header = read('../components/AgentWorkspaceHeader.jsx');
  assert.doesNotMatch(header, /Tooltip/);
  assert.match(header, /title="工作流"/);
  assert.match(header, /title="故事板"/);
});
