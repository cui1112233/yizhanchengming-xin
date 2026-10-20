import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import {
  appendReferenceAsset, attachImageAsset, copyCanvasSelection,
  createLinkedNode,
  createNodeAt,
  pasteCanvasSelection,
  removeCanvasNodes,
  shouldOpenConnectionDrop,
  syncNodeReferencesFromEdges
} from './agent-canvas-context-menu.js';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const canvasSource = fs.readFileSync(path.resolve(dirname, '../components/AgentCanvas.jsx'), 'utf8');
const canvasStyles = fs.readFileSync(path.resolve(dirname, '../../shared/styles/agent-workspace.css'), 'utf8');

const imageNode = {
  id: 'image-1', type: 'agent', position: { x: 20, y: 40 }, selected: true,
  data: { type: 'image', title: '封面', summary: '', version: 1, sourceNodeIds: [], status: 'confirmed', assetRef: { id: 'asset-1', kind: 'image' } }
};
const videoNode = {
  id: 'video-1', type: 'agent', position: { x: 260, y: 80 }, selected: true,
  data: { type: 'video', title: '预告', summary: '', version: 1, sourceNodeIds: ['image-1'], status: 'draft', assetRef: { id: 'asset-2', kind: 'video' } }
};
const externalNode = {
  id: 'text-1', type: 'agent', position: { x: 0, y: 0 }, selected: false,
  data: { type: 'text', title: '文案', summary: '正文', version: 1, sourceNodeIds: [], status: 'draft', assetRef: null }
};
const internalEdge = { id: 'edge-image-video', source: 'image-1', target: 'video-1', data: { relation: 'reference' } };
const externalEdge = { id: 'edge-text-image', source: 'text-1', target: 'image-1', data: { relation: 'reference' } };

test('connection drop menu opens only for a genuine empty-canvas drop', () => {
  assert.equal(shouldOpenConnectionDrop({ sourceID: '', connectionCompleted: false, droppedOnHandle: false }), false);
  assert.equal(shouldOpenConnectionDrop({ sourceID: 'image-1', connectionCompleted: true, droppedOnHandle: false }), false);
  assert.equal(shouldOpenConnectionDrop({ sourceID: 'image-1', connectionCompleted: false, droppedOnHandle: true }), false);
  assert.equal(shouldOpenConnectionDrop({ sourceID: 'image-1', connectionCompleted: false, droppedOnHandle: false }), true);
});

test('pasting a copied pair recreates only its internal edge at the click position', () => {
  const clipboard = copyCanvasSelection({ nodes: [imageNode, videoNode, externalNode], edges: [internalEdge, externalEdge], selectedIds: ['image-1', 'video-1'] });
  const pasted = pasteCanvasSelection({ clipboard, position: { x: 500, y: 240 }, nextID: value => `copy-${value}` });
  assert.deepEqual(pasted.nodes.map(node => node.id), ['copy-image-1', 'copy-video-1']);
  assert.deepEqual(pasted.nodes[0].position, { x: 500, y: 240 });
  assert.equal(pasted.nodes[0].data.title, '封面副本');
  assert.equal(pasted.edges.length, 1);
  assert.equal(pasted.edges[0].source, 'copy-image-1');
  assert.equal(pasted.edges[0].target, 'copy-video-1');
});

test('removing a node also removes every attached edge', () => {
  const result = removeCanvasNodes({ nodes: [externalNode, imageNode, videoNode], edges: [externalEdge, internalEdge], ids: ['image-1'] });
  assert.deepEqual(result.nodes.map(node => node.id), ['text-1', 'video-1']);
  assert.deepEqual(result.edges, []);
});

test('creating a media node keeps only a serializable asset reference', () => {
  const node = createNodeAt({ id: 'node-1', kind: 'image', position: { x: 12, y: 18 }, asset: { id: 'asset-1', name: '封面.png', kind: 'image', url: '/api/agent/assets/asset-1/content' } });
  assert.deepEqual(node.position, { x: 12, y: 18 });
  assert.equal(node.data.type, 'image');
  assert.equal(node.data.title, '封面.png');
  assert.equal(node.data.assetRef.id, 'asset-1');
  assert.equal(node.data.status, 'draft');
});

test('creating a blank image node keeps the prompt, model, and ratio fields ready for generation', () => {
  const node = createNodeAt({ id: 'image-blank', kind: 'image', position: { x: 12, y: 18 } });
  assert.equal(node.data.type, 'image');
  assert.equal(node.data.assetRef, null);
  assert.deepEqual(node.data.config, { prompt: '', model: '', ratio: '16:9', referenceAssets: [] });
});

test('adding image references keeps image assets only, deduplicates them, and caps the set at ten', () => {
  const image = index => ({ id: `image-${index}`, kind: 'image', name: `参考图 ${index}`, url: `/assets/${index}.png` });
  const node = createNodeAt({ id: 'image-node', kind: 'image', position: { x: 0, y: 0 } });
  const withFirst = appendReferenceAsset({ node, asset: image(1) });
  const repeated = appendReferenceAsset({ node: withFirst, asset: image(1) });
  const withVideo = appendReferenceAsset({ node: repeated, asset: { id: 'video-1', kind: 'video', name: '不应加入.mp4' } });
  const complete = Array.from({ length: 12 }, (_, index) => image(index + 2))
    .reduce((current, asset) => appendReferenceAsset({ node: current, asset }), withVideo);

  assert.deepEqual(withFirst.data.config.referenceAssets.map(asset => asset.id), ['image-1']);
  assert.deepEqual(repeated.data.config.referenceAssets.map(asset => asset.id), ['image-1']);
  assert.deepEqual(withVideo.data.config.referenceAssets.map(asset => asset.id), ['image-1']);
  assert.equal(complete.data.config.referenceAssets.length, 10);
  assert.equal(complete.data.config.referenceAssets.at(-1).id, 'image-10');
});

test('connecting existing media automatically fills eligible generation references without losing manual selections', () => {
  const sourceImage = createNodeAt({ id: 'source-image', kind: 'image', position: { x: 0, y: 0 }, asset: { id: 'asset-image', kind: 'image', name: '熊猫.png', url: '/assets/panda.png' } });
  const sourceVideo = createNodeAt({ id: 'source-video', kind: 'video', position: { x: 0, y: 0 }, asset: { id: 'asset-video', kind: 'video', name: '熊猫.mp4', url: '/assets/panda.mp4' } });
  const targetVideo = createNodeAt({ id: 'target-video', kind: 'video', position: { x: 0, y: 0 } });
  const withManualReference = {
    ...targetVideo,
    data: {
      ...targetVideo.data,
      config: { ...targetVideo.data.config, referenceAssets: [{ id: 'asset-manual', kind: 'image', name: '手动.png', url: '/assets/manual.png' }] }
    }
  };
  const synced = syncNodeReferencesFromEdges({
    nodes: [sourceImage, sourceVideo, withManualReference],
    edges: [
      { id: 'edge-image-video', source: 'source-image', target: 'target-video' },
      { id: 'edge-video-video', source: 'source-video', target: 'target-video' }
    ]
  });

  assert.deepEqual(synced.find(node => node.id === 'target-video').data.config.referenceAssets.map(asset => asset.id), ['asset-manual', 'asset-image', 'asset-video']);
  assert.deepEqual(synced.find(node => node.id === 'target-video').data.config.autoReferenceAssetIDs, ['asset-image', 'asset-video']);
});

test('image generation ignores connected video assets while accepting connected images', () => {
  const sourceImage = createNodeAt({ id: 'source-image', kind: 'image', position: { x: 0, y: 0 }, asset: { id: 'asset-image', kind: 'image', name: '熊猫.png' } });
  const sourceVideo = createNodeAt({ id: 'source-video', kind: 'video', position: { x: 0, y: 0 }, asset: { id: 'asset-video', kind: 'video', name: '熊猫.mp4' } });
  const targetImage = createNodeAt({ id: 'target-image', kind: 'image', position: { x: 0, y: 0 } });
  const synced = syncNodeReferencesFromEdges({
    nodes: [sourceImage, sourceVideo, targetImage],
    edges: [
      { id: 'edge-image-image', source: 'source-image', target: 'target-image' },
      { id: 'edge-video-image', source: 'source-video', target: 'target-image' }
    ]
  });

  assert.deepEqual(synced.find(node => node.id === 'target-image').data.config.referenceAssets.map(asset => asset.id), ['asset-image']);
});

test('selecting a material from the image card fills the image node without turning it into a reference image', () => {
  const node = createNodeAt({ id: 'image-node', kind: 'image', position: { x: 0, y: 0 } });
  const asset = { id: 'asset-cover', kind: 'image', name: '封面.png', url: '/assets/cover.png' };
  const result = attachImageAsset({ node, asset });

  assert.equal(result.data.assetRef.id, 'asset-cover');
  assert.equal(result.data.title, '封面.png');
  assert.deepEqual(result.data.config.referenceAssets, []);
  assert.equal(attachImageAsset({ node, asset: { id: 'video-1', kind: 'video' } }), node);
});

test('quick adding from a node creates one linked downstream node without mutating the source', () => {
  const source = createNodeAt({ id: 'source-image', kind: 'image', position: { x: 80, y: 140 } });
  const result = createLinkedNode({ source, id: 'next-video', kind: 'video' });

  assert.equal(result.node.id, 'next-video');
  assert.equal(result.node.data.type, 'video');
  assert.deepEqual(result.node.position, { x: 336, y: 140 });
  assert.deepEqual(result.edge, {
    id: 'edge-source-image-next-video-reference',
    source: 'source-image',
    target: 'next-video',
    data: { relation: 'reference' }
  });
  assert.deepEqual(source.position, { x: 80, y: 140 });
});

test('creating an audio node keeps generation-ready business defaults', () => {
  const node = createNodeAt({ id: 'audio-1', kind: 'audio', position: { x: 12, y: 18 } });
  assert.equal(node.data.type, 'audio');
  assert.equal(node.data.title, '新音频');
  assert.deepEqual(node.data.config, { prompt: '', voice: '', speed: 1 });
});

test('creating a video node is allowed before any video asset exists', () => {
  const node = createNodeAt({ id: 'video-1', kind: 'video', position: { x: 12, y: 18 } });
  assert.equal(node.data.type, 'video');
  assert.equal(node.data.title, '新视频');
  assert.equal(node.data.assetRef, null);
  assert.deepEqual(node.data.config, { prompt: '', model: '', ratio: '', durationSeconds: 0, referenceAssets: [] });
});

test('canvas registers pane and node context-menu handlers', () => {
  assert.match(canvasSource, /onPaneContextMenu=\{handlePaneContextMenu\}/);
  assert.match(canvasSource, /onNodeContextMenu=\{handleNodeContextMenu\}/);
  assert.match(canvasSource, /handleContextPaste/);
  assert.match(canvasSource, /handleContextDelete/);
});

test('canvas loads or uploads a project asset before committing a media node', () => {
  assert.match(canvasSource, /listAgentAssets\(taskId, assetKind\)/);
  assert.match(canvasSource, /uploadAgentAsset\(taskId, file\)/);
  assert.match(canvasSource, /createNodeAt\(\{[\s\S]*asset/);
});

test('canvas exposes a standalone video node action alongside asset video import', () => {
  assert.match(canvasSource, /handleContextCreate\('video'\)/);
  assert.match(canvasSource, /新建视频/);
  assert.match(canvasSource, /从资产库添加/);
  assert.match(canvasSource, /openAssetPanel\('video'\)/);
});

test('canvas groups new node choices before paste and history actions', () => {
  assert.match(canvasSource, /新建节点/);
  assert.match(canvasSource, /新建图片/);
  assert.match(canvasSource, /新建视频/);
  assert.match(canvasSource, /从资产库添加/);
  assert.match(canvasSource, /本地上传/);
  assert.match(canvasSource, /contextMenuSubmenu/);
});

test('clicking a context-menu parent keeps its child choices open', () => {
  assert.match(canvasSource, /onClick=\{\(\) => setContextMenuSubmenu\('create'\)\}/);
  assert.match(canvasSource, /onClick=\{\(\) => setContextMenuSubmenu\('assets'\)\}/);
  assert.match(canvasSource, /onClick=\{\(\) => setContextMenuSubmenu\('upload'\)\}/);
  assert.doesNotMatch(canvasSource, /current => current === 'create'/);
});

test('node plus control is a real source handle for dragging a connection', () => {
  assert.match(canvasSource, /<Handle type="source" position=\{Position\.Right\} className="agent-canvas-node-quick-add" aria-label="从此节点拉出连线" \/>/);
  assert.doesNotMatch(canvasSource, /onQuickAdd=/);
  assert.doesNotMatch(canvasSource, /createLinkedNode\(/);
  assert.match(canvasStyles, /\.agent-page \.agent-canvas-node-quick-add::before \{ content: '\+';/);
});

test('empty connection drops offer text and image nodes instead of silently creating a default node', () => {
  assert.match(canvasSource, /onConnectStart=\{handleConnectStart\}/);
  assert.match(canvasSource, /onConnectEnd=\{handleConnectEnd\}/);
  assert.match(canvasSource, /添加参考节点/);
  assert.match(canvasSource, /handleConnectionDropCreate\('text'\)/);
  assert.match(canvasSource, /handleConnectionDropCreate\('image'\)/);
});

test('canvas exposes visible target ports and a forgiving connection hit area', () => {
  assert.match(canvasSource, /<Handle type="target" position=\{Position\.Left\} className="agent-canvas-node-target" aria-label="连接到此节点" \/>/);
  assert.match(canvasSource, /connectionRadius=\{24\}/);
  assert.match(canvasSource, /isValidConnection=\{isValidCanvasConnection\}/);
  assert.match(canvasStyles, /\.agent-page \.agent-canvas-node-target/);
});

test('successful connections are remembered so the empty-drop menu does not also open', () => {
  assert.match(canvasSource, /connectionCompletedRef/);
  assert.match(canvasSource, /shouldOpenConnectionDrop\(\{ sourceID, connectionCompleted, droppedOnHandle \}\)/);
  assert.match(canvasSource, /setConnectingSourceID\(''\)/);
});

test('image and audio generation controls disclose that their executors are not connected', () => {
  assert.match(canvasSource, /暂未接入生成执行器/);
  assert.match(canvasSource, /const generationAvailable = data\.type === 'video'/);
  assert.match(canvasSource, /disabled=\{!generationAvailable \|\| !String\(config\.prompt \|\| ''\)\.trim\(\)\}/);
});

test('opened media editors expose direct prompt, reference, and setting controls below their preview', () => {
  assert.match(canvasSource, /const generationPanel = isEditorOpen/);
  assert.match(canvasSource, /\{generationPanel\}/);
  assert.match(canvasSource, /agent-canvas-generation-panel/);
  assert.match(canvasSource, /aria-label="添加参考素材"/);
  assert.match(canvasSource, /onPickReferenceAsset/);
  assert.match(canvasSource, /图片提示词/);
  assert.match(canvasSource, /视频提示词/);
  assert.match(canvasSource, /onNodeConfigChange/);
  assert.match(canvasStyles, /\.agent-page \.agent-canvas-generation-panel/);
});

test('opened media editors use the wide reference-style panel instead of compact form controls', () => {
  assert.match(canvasSource, /const generationPanel = isEditorOpen/);
  assert.match(canvasSource, /agent-canvas-generation-panel/);
  assert.match(canvasSource, /agent-canvas-generation-reference-strip/);
  assert.match(canvasSource, /agent-canvas-generation-prompt/);
  assert.match(canvasSource, /agent-canvas-generation-toolbar/);
  assert.match(canvasSource, /agent-canvas-generation-submit/);
  assert.match(canvasStyles, /\.agent-page \.agent-canvas-generation-panel/);
  assert.match(canvasStyles, /\.agent-page \.agent-canvas-generation-reference-strip/);
});

test('generation panel uses the workspace palette with restrained spacing instead of the copied green treatment', () => {
  const panelRule = canvasStyles.match(/\.agent-page \.agent-canvas-generation-panel \{([\s\S]*?)\n\}/);
  assert.ok(panelRule, 'generation panel rule should exist');
  assert.match(panelRule[1], /background:\s*var\(--agent-panel\)/);
  assert.match(panelRule[1], /border-radius:\s*10px/);
  assert.doesNotMatch(panelRule[1], /#193727|#315b43/);
  assert.match(canvasStyles, /color:\s*var\(--agent-accent\)/);
});

test('connecting edges synchronizes upstream media into node reference panels before saving', () => {
  assert.match(canvasSource, /syncNodeReferencesFromEdges/);
  assert.match(canvasSource, /const nextNodes = syncNodeReferencesFromEdges\(\{ nodes: nodesRef\.current, edges: next \}\)/);
});

test('context-menu submenus are allowed to extend beyond the parent menu', () => {
  const rule = canvasStyles.match(/\.agent-page \.agent-canvas-context-menu \{([\s\S]*?)\n\}/);
  assert.ok(rule, 'context menu CSS rule should exist');
  assert.match(rule[1], /overflow:\s*visible;/);
});

test('canvas opens the manual editor on text double click and saves node content through commit', () => {
  assert.match(canvasSource, /onNodeDoubleClick=\{handleNodeDoubleClick\}/);
  assert.match(canvasSource, /if \(normalizeCanvasNodeType\(node\.data\?\.type\) !== 'text'\) return;/);
  assert.match(canvasSource, /handleManualNodeSave/);
  assert.match(canvasSource, /title="手动编辑节点"/);
});

test('opened media editor saves generation settings directly by node type', () => {
  assert.match(canvasSource, /node\.data\?\.config \|\| \{\}/);
  assert.match(canvasSource, /图片提示词/);
  assert.match(canvasSource, /视频时长（秒）/);
  assert.match(canvasSource, /音色/);
  assert.match(canvasSource, /onNodeConfigChange/);
});

test('canvas keeps media asset actions in the opened editor instead of the manual modal', () => {
  assert.match(canvasSource, /aria-label=\{data\.assetRef\?\.url \? '更换图片素材' : '添加图片素材'\}/);
  assert.match(canvasSource, /assetPanel\.mode === 'image-source'/);
  assert.doesNotMatch(canvasSource, /manualNode\?\.data\?\.type === 'image'/);
  assert.doesNotMatch(canvasSource, /manualNode\?\.data\?\.type === 'video'/);
  assert.doesNotMatch(canvasSource, /manualNode\?\.data\?\.type === 'audio'/);
});

test('expanded video editor exposes create cancel retry confirm and per-node errors', () => {
  assert.match(canvasSource, /const videoJob = videoJobs\[props\.id\]\?\.projectId === taskId/);
  assert.match(canvasSource, /videoJob=\{videoJob\}/);
  assert.match(canvasSource, /videoState=\{videoNodeViewState\(videoJob\)\}/);
  assert.match(canvasSource, /videoBusy=\{videoJobsLoadedTaskID !== taskId \|\| Boolean\(videoBusyByNode\[props\.id\]\)\}/);
  assert.match(canvasSource, /videoActionError=\{videoActionErrors\[props\.id\] \|\| \(videoJobsLoadedTaskID !== taskId \? videoJobsLoadError : ''\)\}/);
  for (const action of ['生成视频', '取消视频任务', '重试视频任务', '确认视频结果']) {
    assert.match(canvasSource, new RegExp(`aria-label="${action}"`));
  }
  assert.match(canvasSource, /onCancelVideo\?\.\(id\)/);
  assert.match(canvasSource, /onRetryVideo\?\.\(id\)/);
  assert.match(canvasSource, /onConfirmVideo\?\.\(id\)/);
  assert.doesNotMatch(canvasSource, /const activeVideoJob = manualNode/);
});

test('expanded video editor previews the pending result and reports provider failures', () => {
  assert.match(canvasSource, /videoResultPreview\(videoJob\)/);
  assert.match(canvasSource, /<AuthenticatedAgentVideo source=\{videoResult\.url\}/);
  assert.match(canvasSource, /videoJob\?\.errorMessage/);
  assert.match(canvasSource, /\{videoJob\.errorMessage\}/);
});

test('formal and pending Agent videos use the authenticated media component', () => {
  assert.match(canvasSource, /import \{ AuthenticatedAgentVideo \}/);
  assert.equal((canvasSource.match(/<AuthenticatedAgentVideo/g) || []).length, 2);
  assert.doesNotMatch(canvasSource, /<video className="agent-canvas-(?:node-media|generation-result)"/);
});

test('canvas keeps middle mouse panning and advertises only implemented undo redo keys', () => {
  assert.match(canvasSource, /panOnDrag=\{interactionMode === 'pan' \? \[0, 1\] : \[1\]\}/);
  assert.match(canvasSource, /case 'undo':\s*case 'redo':/);
  assert.doesNotMatch(canvasSource, /aria-label="快捷键"/);
  assert.doesNotMatch(canvasSource, /title="快捷键操作"/);
  assert.doesNotMatch(canvasSource, /case 'select_all':/);
  assert.doesNotMatch(canvasSource, /case 'copy':/);
  assert.doesNotMatch(canvasSource, /case 'delete':/);
  assert.doesNotMatch(canvasSource, /case 'focus_selection':/);
});

test('canvas reads image and video models from the shared API configuration instead of hard-coding its own catalog', () => {
  assert.match(canvasSource, /listAvailableModels\('image'\)/);
  assert.match(canvasSource, /listAvailableModels\('video'\)/);
  assert.match(canvasSource, /configuredImageModelOptions/);
  assert.match(canvasSource, /configuredVideoModelOptions/);
  assert.doesNotMatch(canvasSource, /value=\{imageConfig\.model \|\| 'seedream-5-flash-local'\}/);
});
