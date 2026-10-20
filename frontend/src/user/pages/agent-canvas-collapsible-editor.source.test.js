import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const frontendRoot = path.resolve(dirname, '../../..');
const canvasSource = fs.readFileSync(path.resolve(dirname, '../components/AgentCanvas.jsx'), 'utf8');
const authenticatedImagePath = path.resolve(dirname, '../components/AuthenticatedAgentImage.jsx');
const authenticatedImageSource = fs.existsSync(authenticatedImagePath) ? fs.readFileSync(authenticatedImagePath, 'utf8') : '';
const workspaceStyles = fs.readFileSync(path.resolve(dirname, '../../shared/styles/agent-workspace.css'), 'utf8');
const packageLock = JSON.parse(fs.readFileSync(path.resolve(frontendRoot, 'package-lock.json'), 'utf8'));
const reactFlowSource = fs.readFileSync(path.resolve(frontendRoot, 'node_modules/@xyflow/react/dist/esm/index.js'), 'utf8');
const reactFlowStyles = fs.readFileSync(path.resolve(frontendRoot, 'node_modules/@xyflow/react/dist/style.css'), 'utf8');

const classSpecificity = selector => (selector.match(/\.[a-z0-9_-]+/gi) || []).length;

test('media nodes render one temporary inline editor controlled by preview activation', () => {
  assert.match(canvasSource, /const \[activeEditorNodeID, setActiveEditorNodeID\] = useState\(''\)/);
  assert.match(canvasSource, /isEditorOpen=\{activeEditorNodeID === props\.id\}/);
  assert.match(canvasSource, /onOpenEditor=\{setActiveEditorNodeID\}/);
  assert.match(canvasSource, /const generationPanel = isEditorOpen/);
  assert.match(canvasSource, /\{generationPanel\}/);
  assert.match(canvasSource, /setActiveEditorNodeID\(''\)/);
});

test('image preview opens editing while a named editor action opens the asset picker', () => {
  assert.match(canvasSource, /aria-label="展开图片编辑"/);
  assert.match(canvasSource, /aria-label=\{data\.assetRef\?\.url \? '更换图片素材' : '添加图片素材'\}/);
  assert.doesNotMatch(canvasSource, /aria-label="选择图片素材"[\s\S]*?onClick=\{event => \{ stopCanvasEvent\(event\); onPickImageAsset/);
});

test('only text nodes use the manual double-click editor', () => {
  assert.match(canvasSource, /if \(normalizeCanvasNodeType\(node\.data\?\.type\) !== 'text'\) return;/);
});

test('loaded and serialized nodes both use the canonical type', () => {
  assert.match(canvasSource, /type:\s*normalizeCanvasNodeType\(node\.type\)/);
  assert.match(canvasSource, /type:\s*normalizeCanvasNodeType\(node\.data\.type\)/);
  assert.ok(workspaceStyles.includes('.agent-page .agent-canvas-generation-panel'));
});

test('generation editor expands horizontally while the prompt stays compact', () => {
  const generationPanelRule = workspaceStyles.match(/\.agent-page \.agent-canvas-generation-panel\s*\{([^}]*)\}/);
  const promptRule = workspaceStyles.match(/\.agent-page \.agent-canvas-generation-prompt\s*\{([^}]*)\}/);
  const primarySelectRule = workspaceStyles.match(/\.agent-page \.agent-canvas-generation-toolbar \.ant-select:first-child\s*\{([^}]*)\}/);
  assert.match(canvasSource, /autoSize=\{\{ minRows: 3, maxRows: 7 \}\}/);
  assert.ok(generationPanelRule, 'the expanded editor should have its own layout rule');
  assert.match(generationPanelRule[1], /width:\s*520px/);
  assert.match(generationPanelRule[1], /max-width:\s*calc\(100vw - 96px\)/);
  assert.ok(promptRule, 'the prompt should have its own layout rule');
  assert.doesNotMatch(promptRule[1], /min-height:\s*96px/);
  assert.match(workspaceStyles, /\.agent-page \.agent-canvas-generation-toolbar\s*\{[^}]*flex-wrap:\s*wrap/);
  assert.ok(primarySelectRule, 'the primary model selector should have its own layout rule');
  assert.match(primarySelectRule[1], /flex-basis:\s*100%/);
  assert.match(primarySelectRule[1], /max-width:\s*none/);
});

test('connection handles belong to the visible preview shell and stay hidden until relevant', () => {
  assert.match(canvasSource, /className="agent-canvas-node-preview-shell"[\s\S]*?<Handle type="target"[\s\S]*?<Handle type="source"/);
  assert.match(canvasSource, /selected,\s*isEditorOpen/);
  assert.match(canvasSource, /useLayoutEffect\(\(\) => \{\s*updateNodeInternals\(id\);\s*\}, \[id, isEditorOpen, updateNodeInternals\]\)/);
  assert.match(workspaceStyles, /\.agent-page \.agent-canvas-node-preview-shell\s*\{[^}]*position:\s*relative/);
  assert.match(workspaceStyles, /\.agent-page \.agent-canvas-node--image,[\s\S]*?\.agent-canvas-node--audio\s*\{[^}]*width:\s*340px/);
  assert.match(workspaceStyles, /\.agent-page \.agent-canvas-generation-panel\s*\{[^}]*width:\s*520px/);
  assert.match(workspaceStyles, /\.agent-page \.agent-canvas-node-quick-add\s*\{[^}]*opacity:\s*0[^}]*pointer-events:\s*none/);
  assert.match(workspaceStyles, /\.react-flow__node\.selected[\s\S]*?\.agent-canvas-node-quick-add/);
  assert.match(workspaceStyles, /\.agent-canvas\.is-connecting[\s\S]*?\.agent-canvas-node-target/);
  assert.equal(packageLock.packages['node_modules/@xyflow/react']?.version, '12.11.6');
  assert.match(reactFlowSource, /connectingto:\s*connectingTo/);
  assert.match(workspaceStyles, /\.agent-canvas-node-target\.connectingto\.valid/);
  assert.doesNotMatch(workspaceStyles, /\.agent-canvas-node-target\.connecting\.valid/);
});

test('hidden handle rules outrank React Flow pointer events and interaction states restore them', () => {
  const dependencyRule = reactFlowStyles.match(/(\.react-flow__handle\.connectionindicator)\s*\{([^}]*)\}/);
  const hiddenTargetRule = workspaceStyles.match(/(\.agent-page \.react-flow__handle\.agent-canvas-node-target)\s*\{([^}]*)\}/);
  const hiddenSourceRule = workspaceStyles.match(/(\.agent-page \.react-flow__handle\.agent-canvas-node-quick-add)\s*\{([^}]*)\}/);
  const visibleTargetRule = workspaceStyles.match(/\.agent-page \.agent-canvas\.is-connecting \.agent-canvas-node-target\s*\{([^}]*)\}/);
  const visibleSourceRule = workspaceStyles.match(/\.agent-page \.agent-canvas-node-preview-shell:hover \.agent-canvas-node-quick-add,[\s\S]*?\.agent-page \.agent-canvas-node-preview-shell:focus-within \.agent-canvas-node-quick-add,[\s\S]*?\.agent-page \.agent-canvas-node-quick-add:focus-visible\s*\{([^}]*)\}/);

  assert.ok(dependencyRule, 'React Flow pointer rule should be present');
  assert.match(dependencyRule[2], /pointer-events:\s*all/);
  for (const rule of [hiddenTargetRule, hiddenSourceRule]) {
    assert.ok(rule, 'hidden handle rule should include the React Flow handle class');
    assert.ok(classSpecificity(rule[1]) > classSpecificity(dependencyRule[1]));
    assert.match(rule[2], /pointer-events:\s*none/);
  }
  assert.match(visibleTargetRule?.[1] || '', /pointer-events:\s*auto/);
  assert.match(visibleSourceRule?.[1] || '', /pointer-events:\s*auto/);
});

test('edge visibility changes rendering without deleting serialized edges', () => {
  assert.match(canvasSource, /hidden:\s*!appearance\.edgesVisible/);
  assert.match(canvasSource, /appearance\.edgesVisible \? '隐藏连线' : '显示连线'/);
  assert.match(canvasSource, /updateAppearance\(\{ edgesVisible: !appearance\.edgesVisible \}\)/);
  assert.match(canvasSource, /appearance\.edgesVisible \? <EyeOff size=\{17\} \/> : <Eye size=\{17\} \/>/);
  assert.match(canvasSource, /const restoredEdges = snapshot\.edges\.map\(edge => toFlowEdge\(edge, appearanceRef\.current\)\)/);
  assert.match(canvasSource, /commit\(snapshot\.nodes, restoredEdges, \{ record: false \}\)/);
  assert.doesNotMatch(canvasSource, /appearance\.edgesVisible \? edges : \[\]/);
});

test('expanded video editor retains reachable create cancel retry and confirm actions', () => {
  for (const label of ['生成视频', '取消生成', '重试生成', '确认回填']) {
    assert.match(canvasSource, new RegExp(`>${label}<`));
  }
  assert.match(canvasSource, /onSubmitVideo\?\.\(id\)/);
  assert.match(canvasSource, /onCancelVideo\?\.\(id\)/);
  assert.match(canvasSource, /onRetryVideo\?\.\(id\)/);
  assert.match(canvasSource, /onConfirmVideo\?\.\(id\)/);
  assert.match(canvasSource, /videoState=\{videoNodeViewState\(/);
});

test('every private Agent image surface uses an authenticated Blob session', () => {
  assert.match(canvasSource, /import \{ AuthenticatedAgentImage \}/);
  assert.match(canvasSource, /<AuthenticatedAgentImage source=\{data\.assetRef\.url\}/);
  assert.match(canvasSource, /<AuthenticatedAgentImage source=\{asset\.url\}/);
  assert.doesNotMatch(canvasSource, /<img\s+src=\{(?:data\.assetRef|asset)\.url\}/);
  assert.match(authenticatedImageSource, /createAuthenticatedMediaSession\(source/);
  assert.match(authenticatedImageSource, /reportPlaybackError\('图片源无法显示'\)/);
});

test('video submission refuses selected video references before calling the API', () => {
  assert.match(canvasSource, /videoReferenceCapabilityError\(node\.data\?\.config/);
  assert.match(canvasSource, /setVideoActionErrors[\s\S]*?return;[\s\S]*?createAgentVideoJob/);
});

test('canvas restores project video jobs and rejects stale project responses', () => {
  assert.match(canvasSource, /listAgentVideoJobs\(requestedTaskId\)/);
  assert.match(canvasSource, /shouldAcceptVideoJobResponse\(\{[\s\S]*?requestedTaskId[\s\S]*?currentTaskId:/);
  assert.match(canvasSource, /mergeLatestVideoJobsByNode\(current,/);
});
