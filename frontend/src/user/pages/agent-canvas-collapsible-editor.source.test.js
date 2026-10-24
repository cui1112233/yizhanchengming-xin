import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const frontendRoot = path.resolve(dirname, '../../..');
const canvasSource = fs.readFileSync(path.resolve(dirname, '../components/AgentCanvas.jsx'), 'utf8');
const controlsSource = fs.readFileSync(path.resolve(dirname, '../components/AgentCanvasControls.jsx'), 'utf8');
const composerSource = fs.readFileSync(path.resolve(dirname, '../components/AgentNodeComposer.jsx'), 'utf8');
const scriptComposerSource = fs.readFileSync(path.resolve(dirname, '../components/AgentScriptNodeComposer.jsx'), 'utf8');
const authenticatedImagePath = path.resolve(dirname, '../components/AuthenticatedAgentImage.jsx');
const authenticatedImageSource = fs.existsSync(authenticatedImagePath) ? fs.readFileSync(authenticatedImagePath, 'utf8') : '';
const workspaceStyles = fs.readFileSync(path.resolve(dirname, '../../shared/styles/agent-workspace.css'), 'utf8');
const packageLock = JSON.parse(fs.readFileSync(path.resolve(frontendRoot, 'package-lock.json'), 'utf8'));
const reactFlowSource = fs.readFileSync(path.resolve(frontendRoot, 'node_modules/@xyflow/react/dist/esm/index.js'), 'utf8');
const reactFlowStyles = fs.readFileSync(path.resolve(frontendRoot, 'node_modules/@xyflow/react/dist/style.css'), 'utf8');

const classSpecificity = selector => (selector.match(/\.[a-z0-9_-]+/gi) || []).length;

test('ordinary nodes render one temporary inline editor controlled by preview activation', () => {
  assert.match(canvasSource, /const \[activeEditorNodeID, setActiveEditorNodeID\] = useState\(''\)/);
  assert.match(canvasSource, /isEditorOpen=\{activeEditorNodeID === props\.id\}/);
  assert.match(canvasSource, /onOpenEditor=\{openNodeComposer\}/);
  assert.match(canvasSource, /const isPinnedBatchFactoryComposer = usesPinnedBatchFactoryComposer\(data\);/);
  assert.match(canvasSource, /const composer = \(isEditorOpen \|\| isPinnedBatchFactoryComposer\) \? <AgentNodeComposer/);
  assert.match(canvasSource, /\{composer\}/);
  assert.match(canvasSource, /isEditorOpen \? ' is-editor-open' : ''/);
  assert.match(workspaceStyles, /\.react-flow__node:has\(\.agent-canvas-node\.is-editor-open\)[^{]*\{[^}]*z-index:/);
  assert.match(canvasSource, /setActiveEditorNodeID\(''\)/);
});

test('mapped Batch Factory scripts use a compact visible card instead of pinning the full editor', () => {
  assert.match(canvasSource, /<AgentNodeComposer\s+compact=\{isPinnedBatchFactoryComposer\}/);
  assert.match(canvasSource, /: <div className="agent-canvas-node-card">/);
  assert.match(workspaceStyles, /\.agent-page \.agent-canvas-node-pinned-composer\s*\{/);
  assert.match(composerSource, /compact = false/);
  assert.match(scriptComposerSource, /agent-script-node-composer--compact/);
  assert.match(scriptComposerSource, /autoSize=\{\{ minRows: compact \? 3 : 4, maxRows: compact \? 5 : 8 \}\}/);
  assert.match(scriptComposerSource, /const directActionIDs = compact\s*\? actionPlan\.primary\.filter\(actionID => actionID === 'prompt-templates'\)\s*: actionPlan\.primary/);
});

test('mapped Batch Factory script previews only its selected production lines without replacing the full source', () => {
  assert.match(scriptComposerSource, /import \{ batchFactoryPreviewText \} from '\.\.\/pages\/shuihuo\/batchFactoryContentRange';/);
  assert.match(scriptComposerSource, /const batchFactoryPreview = isBatchFactoryScript \? batchFactoryPreviewText\(draftSource, \{ sourceMetadata: \{ contentRangeLines: batchFactoryRangeLines \} \}\) : draftSource;/);
  assert.match(scriptComposerSource, /value=\{compact \? batchFactoryPreview : draftSource\}/);
  assert.match(scriptComposerSource, /readOnly=\{compact\}/);
});

test('mapped Batch Factory script remains a titled canvas card instead of an unframed editor', () => {
  assert.match(canvasSource, /<div className="agent-canvas-node-pinned-composer">\s*<header className="agent-canvas-node-card-header agent-canvas-node-pinned-header"[^>]*>/);
  assert.match(canvasSource, /<span>剧本<\/span><strong>\{data\.title\}<\/strong><Tag color="blue">批量工厂<\/Tag>/);
  assert.match(canvasSource, /agent-canvas-node-pinned-composer__body/);
  assert.match(workspaceStyles, /\.agent-page \.agent-canvas-node-pinned-header\s*\{/);
  assert.match(workspaceStyles, /\.agent-page \.agent-canvas-node \.agent-canvas-node-pinned-footer\s*\{[^}]*display:\s*flex/);
});

test('mapped Batch Factory script uses its card as the only visible node frame', () => {
  assert.match(workspaceStyles, /\.agent-page \.agent-canvas-node-preview-shell:has\(\.agent-canvas-node-pinned-composer\)\s*\{[^}]*border:\s*0[^}]*padding:\s*0[^}]*background:\s*transparent[^}]*box-shadow:\s*none/);
});

test('opening a node editor leaves the viewport still so its prompt can retain keyboard focus', () => {
  assert.doesNotMatch(canvasSource, /flowInstance\.fitView\(\{ nodes: \[\{ id: nodeID \}\]/);
  assert.doesNotMatch(canvasSource, /canvasEditorFocusOptions\(viewportRef\.current\?\.zoom\)/);
});

test('leaving a media editor also closes its anchored asset library', () => {
  assert.match(canvasSource, /function closeNodeComposer\([^)]*\) \{[\s\S]*?setActiveEditorNodeID\(''\);[\s\S]*?setAssetPanel\(null\);[\s\S]*?\}/);
  assert.match(canvasSource, /if \(event\.key === 'Escape'\) \{[\s\S]*?closeNodeComposer\(\);/);
  assert.match(canvasSource, /onPaneClick=\{closeNodeComposer\}/);
});

test('Escape can dismiss an open editor or connection-drop menu without a pane click event', () => {
  assert.match(canvasSource, /function closeNodeComposer\(event\) \{[\s\S]*?if \(event && !event\.target\?\.classList\?\.contains\('react-flow__pane'\)\) return;/);
});

test('Escape is captured by the canvas before a focused node input can swallow it', () => {
  assert.match(canvasSource, /className=\{`agent-canvas\$\{connectingSourceID \? ' is-connecting' : ''\}`\}[\s\S]*?onKeyDownCapture=\{handleKeyDown\}/);
});

test('asset picker does not present an empty library while reporting a request failure', () => {
  assert.match(canvasSource, /!assetLoading && !assetError && !assets\.length/);
  assert.match(canvasSource, /readableAssetRequestError\(error, 'upload'\)/);
  assert.match(canvasSource, /readableAssetRequestError\(error, 'load'\)/);
});

test('image preview opens editing while a named editor action opens the asset picker', () => {
  assert.match(canvasSource, /aria-label="展开图片编辑"/);
  assert.match(composerSource, /node\.data\?\.assetRef\?\.url \? '更换当前图片' : '从收藏添加图片'/);
  assert.match(composerSource, /onPickImageAsset\?\.\(node\.id, event\)/);
  assert.doesNotMatch(canvasSource, /aria-label="选择图片素材"[\s\S]*?onClick=\{event => \{ stopCanvasEvent\(event\); onPickImageAsset/);
});

test('reference add button keeps its asset picker open after the canvas click settles', () => {
  const referencePicker = canvasSource.match(/function openReferenceAssetPanel\(nodeID, event\) \{([\s\S]*?)\n  \}/);
  assert.ok(referencePicker, 'reference asset picker should exist');
  assert.match(referencePicker[1], /window\.requestAnimationFrame\(\(\) => setAssetPanel\(panel\)\)/);
});

test('only text nodes use the manual double-click editor', () => {
  assert.match(canvasSource, /if \(normalizeCanvasNodeType\(node\.data\?\.type\) !== 'text'\) return;/);
});

test('loaded and serialized nodes both use the canonical type', () => {
  assert.match(canvasSource, /type:\s*normalizeCanvasNodeType\(node\.type\)/);
  assert.match(canvasSource, /type:\s*normalizeCanvasNodeType\(node\.data\.type\)/);
  assert.ok(workspaceStyles.includes('.agent-page .agent-node-composer'));
});

test('generation editor expands horizontally while the prompt stays compact', () => {
  const generationPanelRule = workspaceStyles.match(/\.agent-page \.agent-node-composer\s*\{([^}]*)\}/);
  const promptRule = workspaceStyles.match(/\.agent-page \.agent-node-composer__prompt\s*\{([^}]*)\}/);
  const primarySelectRule = workspaceStyles.match(/\.agent-page \.agent-node-composer__model\s*\{([^}]*)\}/);
  assert.match(composerSource, /autoSize=\{\{ minRows: 2, maxRows: 6 \}\}/);
  assert.ok(generationPanelRule, 'the expanded editor should have its own layout rule');
  assert.match(generationPanelRule[1], /position:\s*relative/);
  assert.match(generationPanelRule[1], /left:\s*50%/);
  assert.match(generationPanelRule[1], /width:\s*min\(680px, calc\(\(100vw - 96px\) \/ max\(var\(--agent-canvas-zoom, 1\), 0\.25\)\)\)/);
  assert.match(generationPanelRule[1], /transform:\s*translateX\(-50%\)/);
  assert.ok(promptRule, 'the prompt should have its own layout rule');
  assert.doesNotMatch(promptRule[1], /min-height:\s*96px/);
  assert.match(workspaceStyles, /\.agent-page \.agent-node-composer__toolbar\s*\{[^}]*flex-wrap:\s*nowrap/);
  assert.ok(primarySelectRule, 'the primary model selector should have its own layout rule');
  assert.match(primarySelectRule[1], /min-width:\s*220px/);
  assert.match(primarySelectRule[1], /max-width:\s*280px/);
  assert.match(primarySelectRule[1], /flex:\s*1 1 240px/);
});

test('prompt typing stays local until focus leaves, then saves without per-keystroke canvas writes', () => {
  assert.match(composerSource, /const \[draftPrompt, setDraftPrompt\] = useState\(prompt\)/);
  assert.match(composerSource, /useEffect\(\(\) => \{\s*setDraftPrompt\(prompt\);\s*\}, \[node\?\.id, prompt\]\)/);
  assert.match(composerSource, /value=\{draftPrompt\}/);
  assert.match(composerSource, /onChange=\{event => setDraftPrompt\(event\.target\.value\)\}/);
  assert.match(composerSource, /const persistPrompt = \(\) => \{/);
  assert.match(composerSource, /onBlur=\{persistPrompt\}/);
  assert.match(composerSource, /onPointerDown=\{stop\}/);
  assert.match(composerSource, /prompt: draftPrompt/);
});

test('panning only updates the local viewport and does not enqueue a canvas save', () => {
  const moveEnd = canvasSource.match(/function handleMoveEnd\(_event, nextViewport\) \{([\s\S]*?)\n  \}/);
  assert.ok(moveEnd, 'canvas move-end handler should exist');
  assert.match(moveEnd[1], /onViewportChange\?\.\(nextViewport\)/);
  assert.doesNotMatch(moveEnd[1], /emitChange\(/);
});

test('an entirely off-screen saved workflow viewport is recovered once without persisting the automatic move', () => {
  assert.match(canvasSource, /shouldRecoverCanvasViewport/);
  assert.match(canvasSource, /viewportRecoveryKeyRef/);
  assert.match(canvasSource, /suppressViewportCommitRef\.current = true/);
  assert.match(canvasSource, /flowInstance\.fitView\(\{ nodes: recoverableNodes, padding: 0\.18, duration: 260 \}\)/);
  assert.match(canvasSource, /if \(!suppressViewportCommit\) onViewportCommit\?\.\(nextViewport\)/);
});

test('connection handles use the actual Batch Factory card edge and stay hidden until relevant', () => {
  assert.match(canvasSource, /className="agent-canvas-node-preview-shell"[\s\S]*?<Handle type="target"/);
  assert.match(canvasSource, /isPinnedBatchFactoryComposer \? <div className="agent-canvas-node-pinned-composer">[\s\S]*?<Handle type="source" position=\{Position\.Right\} className="agent-canvas-node-quick-add"/);
  assert.match(canvasSource, /\{!isPinnedBatchFactoryComposer \? <Handle type="source" position=\{Position\.Right\} className="agent-canvas-node-quick-add"/);
  assert.match(canvasSource, /selected,\s*isEditorOpen/);
  assert.match(canvasSource, /useLayoutEffect\(\(\) => \{\s*updateNodeInternals\(id\);\s*\}, \[id, isEditorOpen, isPinnedBatchFactoryComposer, updateNodeInternals\]\)/);
  assert.match(workspaceStyles, /\.agent-page \.agent-canvas-node-preview-shell\s*\{[^}]*position:\s*relative/);
  assert.match(workspaceStyles, /\.agent-page \.agent-canvas-node-pinned-composer\s*\{[^}]*position:\s*relative[^}]*overflow:\s*visible/);
  assert.match(workspaceStyles, /\.agent-page \.agent-canvas-node--image,[\s\S]*?\.agent-canvas-node--audio\s*\{[^}]*width:\s*340px/);
  assert.match(workspaceStyles, /\.agent-page \.agent-node-composer\s*\{[^}]*width:\s*min\(680px, calc\(\(100vw - 96px\) \/ max\(var\(--agent-canvas-zoom, 1\), 0\.25\)\)\)/);
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
  const visibleSourceRule = workspaceStyles.match(/\.agent-page \.agent-canvas-node-preview-shell:hover \.agent-canvas-node-quick-add,[\s\S]*?\{([^}]*pointer-events:\s*auto[^}]*)\}/);

  assert.ok(dependencyRule, 'React Flow pointer rule should be present');
  assert.match(dependencyRule[2], /pointer-events:\s*all/);
  assert.ok(hiddenTargetRule, 'target handle rule should include the React Flow handle class');
  assert.ok(classSpecificity(hiddenTargetRule[1]) > classSpecificity(dependencyRule[1]));
  assert.match(hiddenTargetRule[2], /pointer-events:\s*none/);
  assert.ok(hiddenSourceRule, 'source handle rule should include the React Flow handle class');
  assert.ok(classSpecificity(hiddenSourceRule[1]) > classSpecificity(dependencyRule[1]));
  assert.match(hiddenSourceRule[2], /pointer-events:\s*none/);
  assert.match(visibleTargetRule?.[1] || '', /pointer-events:\s*auto/);
  assert.match(visibleSourceRule?.[1] || '', /pointer-events:\s*auto/);
});

test('edge visibility stays in the bottom display control without duplicating the left toolrail', () => {
  assert.match(controlsSource, /<ControlButton label=\{workspaceView\.showConnections \? '隐藏连线' : '显示连线'\}/);
  assert.doesNotMatch(canvasSource, /<IconTool label=\{workspaceView\.showConnections \? '隐藏连线' : '显示连线'\}/);
  assert.match(canvasSource, /const restoredEdges = snapshot\.edges\.map\(edge => toFlowEdge\(edge, appearanceRef\.current\)\)/);
  assert.match(canvasSource, /commit\(snapshot\.nodes, restoredEdges, \{ record: false \}\)/);
  assert.doesNotMatch(canvasSource, /hidden:\s*!appearance\.edgesVisible/);
});

test('expanded video editor retains reachable create cancel retry and confirm actions', () => {
  assert.match(composerSource, /aria-label=\{`生成\$\{/);
  assert.match(composerSource, /kind === 'video' \? '视频'/);
  for (const label of ['取消', '重试', '确认']) assert.match(composerSource, new RegExp(`>${label}<`));
  assert.match(composerSource, /onPointerDown=\{stop\}/);
  assert.match(composerSource, /submitNodeGeneration\(\{ event, onSubmit, nodeID: node\.id, modelID: selectedModel \}\)/);
  assert.match(composerSource, /onCancelVideo\?\.\(node\.id\)/);
  assert.match(composerSource, /onRetryVideo\?\.\(node\.id\)/);
  assert.match(composerSource, /onConfirmVideo\?\.\(node\.id\)/);
  assert.match(canvasSource, /videoState=\{videoState\}/);
  assert.match(composerSource, /videoState\.label/);
});

test('expanded image editor shows persisted candidates and requires explicit confirmation', () => {
  assert.match(composerSource, /agent-node-composer__candidates/);
  assert.match(composerSource, /aria-label=\{`选择候选图片/);
  assert.match(composerSource, />使用这张</);
  assert.match(composerSource, />重新生成</);
  assert.match(canvasSource, /confirmAgentImageGeneration/);
  assert.match(canvasSource, /nodeExecutionReview/);
  assert.match(canvasSource, /canvas, executions = \[\]/);
});

test('image generation keeps a visible node-level status from submission through failure or review', () => {
  const submitImage = canvasSource.match(/async function submitImageForNode\(nodeID, selectedModel, draftPrompt\) \{([\s\S]*?)\n  \}/);
  assert.ok(submitImage, 'image generation entry should exist');
  assert.match(composerSource, /agent-node-composer__generation-status/);
  assert.match(composerSource, /正在提交生成请求/);
  assert.match(composerSource, /正在生成图片/);
  assert.match(composerSource, /等待选择图片/);
  assert.match(submitImage[1], /phase: '正在提交生成请求'/);
  assert.match(submitImage[1], /void refreshGenerationTasks\(requestedTaskId\)/);
  assert.match(workspaceStyles, /\.agent-page \.agent-node-composer__generation-status\s*\{/);
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
