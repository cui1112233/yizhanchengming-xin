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
const headerSource = fs.readFileSync(path.resolve(dirname, '../components/AgentWorkspaceHeader.jsx'), 'utf8');

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
  assert.match(css, /@media\s*\(max-width:\s*560px\)[\s\S]*?\.agent-page \.agent-conversation-panel\s*\{[^}]*bottom:\s*64px[^}]*width:\s*100%[^}]*height:\s*min\(72dvh,\s*620px\)/);
});

test('画布底部圆形入口点击后在中间弹出项目对话，不压缩画布', () => {
  const workspace = cssBlock('.agent-page .agent-canvas-workspace');
  assert.match(workspace, /grid-template-columns:\s*minmax\(0,\s*1fr\)/);
  assert.match(workspace, /grid-template-rows:\s*44px\s+minmax\(0,\s*1fr\)/);

  const conversation = cssBlock('.agent-page .agent-conversation-panel');
  assert.match(conversation, /position:\s*absolute/);
  assert.match(conversation, /z-index:\s*15/);
  assert.match(conversation, /left:\s*50%/);
  assert.match(conversation, /bottom:\s*124px/);
  assert.match(conversation, /width:\s*min\(960px,\s*calc\(100% - 32px\)\)/);
  assert.match(conversation, /transform:\s*translateX\(-50%\)/);
  assert.match(conversation, /border:\s*1px solid var\(--agent-border\)/);

  const collapsed = cssBlock('.agent-page .agent-canvas-workspace.is-conversation-collapsed');
  assert.doesNotMatch(collapsed, /grid-template-rows/);

  const toggle = cssBlock('.agent-page .agent-canvas-conversation-toggle');
  // 圆入口已收进底部工具条（见下文合并工具条用例），不再单独绝对定位。
  assert.doesNotMatch(toggle, /position:\s*absolute/);
  assert.match(toggle, /border-radius:\s*50%/);
  assert.match(controlsSource, /aria-label=\{conversationOpen \? '收起项目对话' : '打开项目对话'\}/);
});

test('Agent 窄屏画布隐藏会遮挡输入按钮的桌面宠物', () => {
  assert.match(css, /@media\s*\(max-width:\s*560px\)[\s\S]*?\.legacy-shell:has\(\.agent-page\.agent-workspace-v3\)\s*>\s*\.cm-penguin-shell\s*\{[^}]*display:\s*none/);
});

test('故事板桌面布局默认使用31比69并提供可拖动主分隔线', () => {
  const storyboard = cssBlock('.agent-page .agent-canvas-workspace.is-storyboard');
  assert.match(storyboard, /grid-template-columns:\s*minmax\(320px,\s*var\(--storyboard-director-split,\s*31%\)\)\s+minmax\(0,\s*1fr\)/);
  assert.match(pageSource, /agent-storyboard-workspace-splitter/);
  assert.match(pageSource, /aria-label="调整 AI 助手面板宽度"/);
  assert.match(pageSource, /onDoubleClick=\{resetStoryboardDirectorSplit\}/);
  const splitter = cssBlock('.agent-page .agent-storyboard-workspace-splitter');
  assert.match(splitter, /cursor:\s*col-resize/);
});

test('故事板隐藏桌宠且素材卡片使用Liblib密度', () => {
  assert.match(css, /\.legacy-shell:has\(\.agent-canvas-workspace\.is-storyboard\)\s*>\s*\.cm-penguin-shell\s*\{[^}]*display:\s*none/);
  assert.match(css, /\.agent-page \.agent-storyboard-section\.is-text \.agent-storyboard-item__select[\s\S]*?min-height:\s*32px/);
  assert.match(css, /\.agent-page \.agent-storyboard-section\.is-image \.agent-storyboard-section__items[\s\S]*?minmax\(88px,\s*1fr\)/);
  assert.match(css, /\.agent-page \.agent-storyboard-section\.is-video \.agent-storyboard-section__items[\s\S]*?minmax\(210px,\s*1fr\)/);
});

test('工作流静态连线使用中性曲线且不绘制箭头', () => {
  assert.doesNotMatch(canvasSource, /MarkerType\.ArrowClosed/);
  assert.doesNotMatch(canvasSource, /markerEnd=\{markerEnd\}/);
  assert.match(css, /--agent-edge-neutral/);
  assert.match(css, /\.react-flow__edge\.selected \.react-flow__edge-path/);
  assert.match(css, /\.react-flow__edge:hover \.react-flow__edge-path/);
});

test('工作流节点使用统一标题栏、类型图标和稳定媒体比例', () => {
  assert.match(canvasSource, /function CanvasNodeTypeIcon/);
  assert.match(canvasSource, /agent-canvas-node-type-icon/);
  assert.match(canvasSource, /agent-canvas-node-drag-handle/);
  assert.match(canvasSource, /agent-canvas-node-card-header-action/);
  assert.match(css, /\.agent-page \.agent-canvas-node-card-header\s*\{[^}]*min-height:\s*28px/);
  assert.match(css, /\.agent-page \.agent-canvas-node--image[^}]*width:\s*320px/);
  assert.match(css, /\.agent-page \.agent-canvas-node-media[^}]*aspect-ratio:\s*16\s*\/\s*9/);
});

test('工作流支点保留大点击区并以小圆点贴在节点边缘', () => {
  const target = cssBlock('.agent-page .agent-canvas-node-target');
  const source = cssBlock('.agent-page .agent-canvas-node-quick-add');
  assert.match(target, /width:\s*24px/);
  assert.match(target, /height:\s*24px/);
  assert.match(source, /width:\s*24px/);
  assert.match(source, /height:\s*24px/);
  assert.match(css, /\.agent-page \.agent-canvas-node-target::after[^}]*width:\s*8px[^}]*height:\s*8px/);
  assert.match(css, /\.agent-page \.agent-canvas-node-quick-add::after[^}]*width:\s*8px[^}]*height:\s*8px/);
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
  assert.match(pageSource, /开始创作/);
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
  assert.match(composer, /border-radius:\s*9px/);
  assert.match(composer, /background:\s*var\(--agent-panel\)/);
  const actions = cssBlock('.agent-page .agent-conversation-panel .agent-composer-actions');
  assert.match(actions, /border-top:\s*1px solid var\(--agent-border\)/);
  assert.match(actions, /padding-top:\s*7px/);
});

test('进入画布页后会重新读取当前项目画布', () => {
  const canvasLoadEffect = pageSource.match(/const canvasRequest = syncBookThenLoadCanvas\(\{[\s\S]*?\n\s*}, \[([^\]]+)\]\);/);
  assert.ok(canvasLoadEffect, 'missing canvas load effect');
  assert.match(canvasLoadEffect[0], /taskId: activeTaskId/);
  assert.match(canvasLoadEffect[0], /getCanvas: getAgentCanvas/);
  assert.match(canvasLoadEffect[1], /activeTaskId/);
  assert.match(canvasLoadEffect[1], /isCanvasWorkspace/);
  assert.match(canvasLoadEffect[1], /requestedBatchBookId/);
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
  assert.match(buttons, /width:\s*32px/);
  assert.match(buttons, /height:\s*32px/);
  assert.match(buttons, /border-radius:\s*9px/);
});

test('画布底部显示、缩放和项目对话入口合并为一个居中工具条', () => {
  const dock = cssBlock('.agent-page .agent-canvas-bottom-dock');
  assert.match(dock, /position:\s*absolute/);
  assert.match(dock, /bottom:\s*12px/);
  assert.match(dock, /left:\s*50%/);
  assert.match(dock, /display:\s*flex/);
  assert.match(dock, /transform:\s*translateX\(-50%\)/);

  const displayControls = cssBlock('.agent-page .agent-canvas-display-controls');
  const zoomControls = cssBlock('.agent-page .agent-canvas-zoom-controls');
  const conversationToggle = cssBlock('.agent-page .agent-canvas-conversation-toggle');
  assert.doesNotMatch(displayControls, /position:\s*absolute/);
  assert.doesNotMatch(zoomControls, /position:\s*absolute/);
  assert.doesNotMatch(conversationToggle, /position:\s*absolute/);
  assert.match(controlsSource, /agent-canvas-bottom-dock[\s\S]*agent-canvas-conversation-toggle/);
});

test('工作流和故事版不显示浅色 Logo、小地图节点和音频控件', () => {
  assert.match(headerSource, /<BrandLogo \/>/);
  const projectMenu = cssBlock('.agent-page .agent-canvas-brand-menu');
  assert.match(projectMenu, /border:\s*0/);
  assert.match(projectMenu, /background:\s*transparent/);
  assert.match(projectMenu, /color:\s*var\(--agent-secondary\)/);
  const viewSwitchButton = cssBlock('.agent-page .agent-canvas-view-switch button');
  assert.match(viewSwitchButton, /border:\s*0/);
  assert.match(viewSwitchButton, /background:\s*transparent/);
  const audioEdit = cssBlock('.agent-page .agent-canvas-node-audio-preview > button');
  assert.match(audioEdit, /border:\s*1px solid var\(--agent-border\)/);
  assert.match(audioEdit, /background:\s*var\(--agent-panel\)/);
  assert.match(css, /\.agent-page :is\(\.agent-canvas, \.agent-storyboard-workspace\) audio\s*\{[^}]*color-scheme:\s*dark/);
  assert.match(canvasSource, /<MiniMap className="agent-canvas-minimap" nodeColor="#526277" maskColor="rgba\(7, 13, 21, 0\.72\)" pannable zoomable/);
});

test('文本节点预览保持可读尺寸，不与媒体节点共用过窄卡片', () => {
  assert.match(css, /\.agent-page \.agent-canvas-node--text,[\s\S]{0,80}\{ width: 280px; \}/);
  assert.match(css, /\.agent-page \.agent-canvas-node-text-preview\s*\{[^}]*min-height:\s*36px/);
  assert.match(css, /\.agent-page \.agent-canvas-node-text-preview\s*\{[^}]*overflow:\s*hidden/);
});

test('项目抽屉从顶部工具带下方展开，不遮挡画布工具条所在条带', () => {
  const drawer = cssBlock('.agent-page .agent-project-drawer');
  const toolrail = cssBlock('.agent-page .agent-canvas-toolrail');
  // 抽屉改为从 top:44px 的工具带下方展开，因此工具条不再靠更高的 z-index 抢占；
  // 断言改为校验实际布局约束（抽屉起点与工具条位置），不是功能改动。
  assert.match(drawer, /top:\s*44px/);
  assert.match(toolrail, /top:\s*12px/);
  assert.match(drawer, /z-index:\s*20/);
});

test('窄屏编辑工具可横向滚动且缩放栏保持有界', () => {
  assert.match(css, /@media\s*\(max-width:\s*560px\)[\s\S]*?\.agent-page \.agent-canvas-editbar\s*\{[^}]*max-width:\s*calc\(100% - 72px\)[^}]*overflow-x:\s*auto/);
  assert.match(css, /@media\s*\(max-width:\s*560px\)[\s\S]*?\.agent-page \.agent-canvas-bottom-dock\s*\{[^}]*max-width:\s*calc\(100% - 24px\)[^}]*overflow-x:\s*auto/);
  // 对话面板改为浮层后不再下压底部工具条，旧的位移规则已移除。
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
  // 连线保持无箭头的中性样式（与上一条“中性连线”用例一致），不再断言 markerEnd。
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
