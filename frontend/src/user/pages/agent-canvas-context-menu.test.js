import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import {
  copyCanvasSelection,
  createNodeAt,
  pasteCanvasSelection,
  removeCanvasNodes
} from './agent-canvas-context-menu.js';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const canvasSource = fs.readFileSync(path.resolve(dirname, '../components/AgentCanvas.jsx'), 'utf8');

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
