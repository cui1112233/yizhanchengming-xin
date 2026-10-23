import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const css = fs.readFileSync(path.resolve(dirname, '../../shared/styles/agent-workspace.css'), 'utf8');
const pageSource = fs.readFileSync(path.resolve(dirname, './AgentPageV2.jsx'), 'utf8');
const canvasSource = fs.readFileSync(path.resolve(dirname, '../components/AgentCanvas.jsx'), 'utf8');
const controlsSource = fs.readFileSync(path.resolve(dirname, '../components/AgentCanvasControls.jsx'), 'utf8');

function cssBlock(selector) {
  const escapedSelector = selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const match = css.match(new RegExp(`${escapedSelector}\\s*\\{([^}]*)\\}`, 's'));
  assert.ok(match, `missing CSS block for ${selector}`);
  return match[1];
}

test('Agent 创作首页覆盖旧版双列和双行布局', () => {
  const page = cssBlock('.agent-page.utility-workbench');
  assert.match(page, /grid-template-columns:\s*minmax\(0,\s*1fr\)/);
  assert.match(page, /grid-template-rows:\s*minmax\(0,\s*1fr\)/);
});

test('Agent 创作选项可换行并使用清晰的主题按钮颜色', () => {
  const actions = cssBlock('.agent-page .agent-composer-actions-left');
  assert.match(actions, /display:\s*flex/);
  assert.match(actions, /flex-wrap:\s*wrap/);

  const chip = cssBlock('.agent-page .agent-context-chip');
  assert.match(chip, /background:\s*var\(--agent-surface\)/);
  assert.match(chip, /color:\s*var\(--agent-text\)/);
  assert.match(chip, /border:\s*1px solid var\(--agent-border\)/);
});

test('Agent 窄屏对话区使用有界底部面板', () => {
  assert.match(css, /@media\s*\(max-width:\s*560px\)[\s\S]*?\.agent-page \.agent-conversation-panel\s*\{[^}]*top:\s*auto[^}]*height:\s*min\(72dvh,\s*620px\)[^}]*width:\s*100%/);
});

test('Agent 窄屏画布隐藏会遮挡输入按钮的桌面宠物', () => {
  assert.match(css, /@media\s*\(max-width:\s*560px\)[\s\S]*?\.legacy-shell:has\(\.agent-page\.agent-workspace-v3\)\s*>\s*\.cm-penguin-shell\s*\{[^}]*display:\s*none/);
});

test('项目对话空状态提供只填入输入框的创作引导', () => {
  assert.match(pageSource, /function handleConversationStarter\(prompt\)/);
  assert.match(pageSource, /questionInputRef\.current\?\.focus\?\.\(\)/);
  assert.match(pageSource, /agent-conversation-starters/);
  for (const label of ['从画布开始创作', '从素材预设开始脚本原创', '上传故事来改编', '批量优化提示词']) {
    assert.match(pageSource, new RegExp(label));
  }
  assert.match(css, /\.agent-page \.agent-conversation-starters\s*\{/);
  assert.match(css, /\.agent-page \.agent-conversation-starter\s*\{/);
  assert.match(css, /@media\s*\(max-width:\s*560px\)[\s\S]*agent-conversation-starters/);
});

test('项目对话输入区将能力收为紧凑图标工具条', () => {
  assert.match(pageSource, /agent-composer-tools/);
  assert.match(pageSource, /全能创作/);
  assert.match(pageSource, /aria-label="添加附件"/);
  assert.match(pageSource, /aria-label="选择技能"/);
  assert.match(pageSource, /aria-label="选择专家"/);
  assert.match(pageSource, /aria-label="发送消息"/);
  assert.match(pageSource, /找问题/);
  assert.match(css, /\.agent-page \.agent-conversation-panel \.agent-composer-tools\s*\{/);
  assert.match(css, /\.agent-page \.agent-conversation-panel \.agent-composer-send\s*\{/);
});

test('项目对话把输入内容与工具条收在同一个输入框内', () => {
  const composer = cssBlock('.agent-page .agent-conversation-panel .agent-workbench-composer');
  assert.match(composer, /border:\s*1px solid var\(--agent-border-strong\)/);
  assert.match(composer, /border-radius:\s*8px/);
  assert.match(composer, /background:\s*var\(--agent-elevated\)/);
  const actions = cssBlock('.agent-page .agent-conversation-panel .agent-composer-actions');
  assert.match(actions, /border-top:\s*1px solid var\(--agent-border\)/);
  assert.match(actions, /padding-top:\s*7px/);
});

test('进入画布页后会重新读取当前项目画布', () => {
  const canvasLoadEffect = pageSource.match(/Promise\.all\(\[getAgentCanvas\(activeTaskId\)[\s\S]*?\n\s*}, \[([^\]]+)\]\);/);
  assert.ok(canvasLoadEffect, 'missing canvas load effect');
  assert.match(canvasLoadEffect[1], /activeTaskId/);
  assert.match(canvasLoadEffect[1], /isCanvasWorkspace/);
});

test('画布节点列表在已经处于工作流时也会重新触发定位', () => {
  assert.match(pageSource, /const \[workflowFocusRequest, setWorkflowFocusRequest\] = useState\(0\);/);
  assert.match(pageSource, /function openNodeInWorkflow\(nodeId\) \{[\s\S]*?setWorkflowFocusRequest\(current => current \+ 1\);[\s\S]*?updateWorkspaceView\(\{ activeView: 'workflow' \}\);/);
  const focusEffect = pageSource.match(/useEffect\(\(\) => \{[\s\S]*?canvasRef\.current\?\.locateNode\(nodeId\);[\s\S]*?}, \[([^\]]+)\]\);/);
  assert.ok(focusEffect, 'missing workflow node focus effect');
  assert.match(focusEffect[1], /workflowFocusRequest/);
});

test('画布保持稳定的节点渲染器定义，输入时不会重置 React Flow 节点类型', () => {
  assert.match(canvasSource, /const canvasNodeTypes = \{\s*agent: CanvasNodeRenderer\s*\};/);
  assert.doesNotMatch(canvasSource, /const canvasNodeTypes = useMemo\(\(\) => \(\{\s*agent: props =>/);
});

test('节点列表定位会居中目标节点而不把单个节点强制放大到最大缩放', () => {
  const locateNode = canvasSource.match(/function locateNode\(nodeID\) \{[\s\S]*?\n  \}/);
  assert.ok(locateNode, 'missing locateNode');
  assert.match(locateNode[0], /flowInstance\?\.getInternalNode\?\.\(nodeID\)/);
  assert.match(locateNode[0], /flowInstance\?\.setCenter\?/);
  assert.doesNotMatch(locateNode[0], /fitView\(\{ nodes:/);
});

test('视频确认回读只把服务端 revision 同步给发起请求时的任务和保存队列', () => {
  assert.match(pageSource, /const requestedTaskId = activeTaskId/);
  assert.match(pageSource, /const requestedQueue = canvasSaveQueueRef\.current/);
  assert.match(pageSource, /shouldAcceptCanvasRefresh/);
  assert.match(pageSource, /await requestedQueue\.acceptServerCanvas\(nextCanvas\)/);
});

test('画布把全局工具与选中节点操作分层，常用控制收进底部工具组', () => {
  for (const selector of ['.agent-page .agent-canvas-toolrail', '.agent-page .agent-canvas-editbar', '.agent-page .agent-canvas-bottom-dock']) {
    const block = cssBlock(selector);
    assert.match(block, /position:\s*absolute/);
    assert.match(block, /z-index:\s*\d+/);
  }
  assert.match(canvasSource, /\{selectedIds\.length > 0 \? <nav className="agent-canvas-editbar"/);
  assert.match(controlsSource, /<div className="agent-canvas-bottom-dock">/);
  const buttons = cssBlock('.agent-page .agent-canvas :is(.agent-canvas-toolrail, .agent-canvas-editbar, .agent-canvas-zoombar, .agent-canvas-display-controls, .agent-canvas-zoom-controls) .ant-btn');
  assert.match(buttons, /width:\s*36px/);
  assert.match(buttons, /height:\s*36px/);
  assert.match(buttons, /border-radius:\s*[0-8]px/);
});

test('文本节点预览保持可读尺寸，不与媒体节点共用过窄卡片', () => {
  assert.match(css, /\.agent-page \.agent-canvas-node--text\s*\{[^}]*width:\s*280px/);
  assert.match(css, /\.agent-page \.agent-canvas-node-text-preview\s*\{[^}]*min-height:\s*64px/);
});

test('项目抽屉打开后不会盖住左侧全局画布工具', () => {
  const drawer = cssBlock('.agent-page .agent-project-drawer');
  const toolrail = cssBlock('.agent-page .agent-canvas-toolrail');
  assert.match(drawer, /z-index:\s*20/);
  assert.match(toolrail, /z-index:\s*2[1-9]/);
});

test('窄屏编辑工具可横向滚动且缩放栏保持有界', () => {
  assert.match(css, /@media\s*\(max-width:\s*560px\)[\s\S]*?\.agent-page \.agent-canvas-editbar\s*\{[^}]*max-width:\s*calc\(100% - 72px\)[^}]*overflow-x:\s*auto/);
  assert.match(css, /@media\s*\(max-width:\s*560px\)[\s\S]*?\.agent-page \.agent-canvas-bottom-dock\s*\{[^}]*max-width:\s*calc\(100% - 24px\)[^}]*overflow-x:\s*auto/);
  assert.match(css, /@media\s*\(max-width:\s*560px\)[\s\S]*?\.agent-canvas-workspace:not\(\.is-conversation-collapsed\) \.agent-canvas-bottom-dock\s*\{[^}]*bottom:\s*calc\(min\(72dvh,\s*620px\) \+ 12px\)/);
});

test('节点点击明确支持 Ctrl Shift 和 Command 多选', () => {
  assert.match(canvasSource, /onNodeClick=\{handleNodeClick\}/);
  assert.match(canvasSource, /const isMultiSelection = isCanvasMultiSelectionEvent\(event\);/);
  assert.match(canvasSource, /selectCanvasNodes\([\s\S]*?nodesRef\.current,[\s\S]*?node\.id,[\s\S]*?isMultiSelection,[\s\S]*?selectedIds/);
});

test('画布提供持久外观设置和按方向流动的连线', () => {
  for (const label of ['连线颜色', '连线动效', '连线粗细', '背景网点', '辅助线对齐']) {
    assert.match(canvasSource, new RegExp(label));
  }
  assert.match(canvasSource, /label="画布外观"/);
  assert.match(canvasSource, /edgeTypes=\{canvasEdgeTypes\}/);
  assert.match(canvasSource, /edgeAnimated/);
  assert.match(canvasSource, /markerEnd/);
  assert.match(canvasSource, /appearance:\s*normalizeCanvasAppearance\(appearance\)/);
  assert.match(canvasSource, /onAppearanceChange/);
});

test('画布外观包含轻量动效和窄屏安全规则', () => {
  assert.match(css, /\.agent-page \.agent-canvas-edge-flow\s*\{[^}]*animation:/);
  assert.match(css, /@keyframes agent-canvas-edge-flow/);
  assert.match(canvasSource, /const energyStyle = \{ \.\.\.style, strokeWidth: Number\(style\.strokeWidth\) \+ 1 \}/);
  assert.match(css, /\.agent-page \.agent-canvas-edge-flow\s*\{[^}]*stroke-dasharray:\s*10 26[^}]*filter:\s*drop-shadow/);
  assert.match(css, /@keyframes agent-canvas-edge-flow\s*\{\s*to\s*\{\s*stroke-dashoffset:\s*-72/);
  assert.match(css, /\.agent-page \.agent-canvas-appearance-panel\s*\{/);
  assert.match(css, /\.agent-canvas-appearance-popover\s*\{[^}]*--agent-edge-coral:\s*#f07167/);
  assert.match(css, /@media \(max-width: 560px\)[\s\S]*?\.agent-canvas-appearance-popover[^}]*max-width:\s*calc\(100vw - 24px\)/);
  assert.match(css, /@media \(prefers-reduced-motion: reduce\)[\s\S]*?\.agent-canvas-edge-flow\s*\{[^}]*animation:\s*none/);
});
