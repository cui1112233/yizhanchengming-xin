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

test('导演台在节点列表中提供不依赖画布定位的直接打开入口', () => {
  const drawer = read('../components/AgentWorkspaceDrawer.jsx');
  const page = read('./AgentPageV2.jsx');
  const canvas = read('../components/AgentCanvas.jsx');
  assert.match(drawer, /onOpenDirectorDesk/);
  assert.match(drawer, /打开 3D 导演台/);
  assert.match(page, /canvasRef\.current\?\.openDirectorDesk/);
  assert.match(canvas, /openDirectorDesk\(nodeID\)/);
  assert.match(canvas, /useImperativeHandle\(ref, \(\) => \(\{[\s\S]*?openDirectorDesk/);
});

test('回到批量工厂作品画布时会重新同步真实生产状态', () => {
  const page = read('./AgentPageV2.jsx');
  assert.match(page, /refreshBatchFactoryBookCanvas/);
  assert.match(page, /sync:\s*syncAgentBatchFactoryBook/);
  assert.match(page, /document\.addEventListener\('visibilitychange', refreshWhenVisible\)/);
  assert.match(page, /window\.addEventListener\('focus', refreshWhenVisible\)/);
  assert.match(page, /requestedQueue\.acceptServerCanvas\(nextCanvas\)/);
});

test('抽屉和显示开关只更新本地工作区状态，不触发画布保存', () => {
  const page = read('./AgentPageV2.jsx');
  const workspaceUpdate = page.match(/function updateWorkspaceView\(patch\) \{([\s\S]*?)\n  \}/);
  assert.ok(workspaceUpdate, 'workspace view update handler should exist');
  assert.match(workspaceUpdate[1], /setViewState\(nextViewState\)/);
  assert.doesNotMatch(workspaceUpdate[1], /canvasSaveQueueRef\.current\?\.enqueue/);
});

test('项目抽屉提供真实生成历史并以统一生成任务状态为主', () => {
  const drawer = read('../components/AgentWorkspaceDrawer.jsx');
  const canvas = read('../components/AgentCanvas.jsx');
  const page = read('./AgentPageV2.jsx');
  assert.match(drawer, /生成历史/);
  assert.match(drawer, /executions = \[\]/);
  assert.match(drawer, /const historyEntries = useMemo/);
  assert.match(drawer, /generations\.map/);
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

test('项目素材使用可独立收起的书籍文件夹和固定媒体预览框', () => {
  const drawer = read('../components/AgentWorkspaceDrawer.jsx');
  const css = read('../../shared/styles/agent-workspace.css');
  assert.match(drawer, /collapsedAssetGroups/);
  assert.match(drawer, /aria-expanded=\{!collapsed\}/);
  assert.match(drawer, /agent-workspace-assets__group-header/);
  assert.match(css, /\.agent-page \.agent-workspace-assets__preview\s*\{[\s\S]*?width:\s*48px;[\s\S]*?height:\s*48px;/);
  assert.match(css, /agent-workspace-assets__preview\s+(?:img|video)/);
  assert.match(css, /object-fit:\s*cover/);
});

test('工作流节点使用可收起的文件夹分组并保留定位入口', () => {
  const drawer = read('../components/AgentWorkspaceDrawer.jsx');
  const css = read('../../shared/styles/agent-workspace.css');
  assert.match(drawer, /collapsedNodeGroups/);
  assert.match(drawer, /agent-workspace-tree__group-header/);
  assert.match(drawer, /aria-expanded=\{!collapsed\}/);
  assert.match(drawer, /<Folder size=\{14\}/);
  assert.match(drawer, /onLocateNode\?\.\(node\.id\)/);
  assert.match(css, /agent-workspace-tree__group-header/);
});

test('生成历史每条记录都有独立的内容层级和状态样式', () => {
  const drawer = read('../components/AgentWorkspaceDrawer.jsx');
  const css = read('../../shared/styles/agent-workspace.css');
  assert.match(drawer, /agent-workspace-history__item/);
  assert.match(drawer, /agent-workspace-history__content/);
  assert.match(drawer, /agent-workspace-history__status/);
  assert.match(css, /\.agent-page \.agent-workspace-history\s*\{/);
  assert.match(css, /\.agent-page \.agent-workspace-history__item\s*\{[\s\S]*?display:\s*grid/);
  assert.match(css, /\.agent-page \.agent-workspace-history__content\s*\{[\s\S]*?min-width:\s*0/);
  assert.match(css, /\.agent-page \.agent-workspace-history__status\.is-failed/);
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
  assert.match(header, /title="画布"/);
  assert.match(header, /title="故事板"/);
});
