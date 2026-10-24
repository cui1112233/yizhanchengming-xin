import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { connectionDropChoices, createTypedCanvasEdge } from './agent-canvas-connections.js';
import { shouldOpenConnectionDrop } from './agent-canvas-context-menu.js';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const canvasSource = fs.readFileSync(path.resolve(dirname, '../components/AgentCanvas.jsx'), 'utf8');
const styles = fs.readFileSync(path.resolve(dirname, '../../shared/styles/agent-workspace.css'), 'utf8');

test('node editor is embedded inside the node card while handles stay on the preview shell', () => {
  assert.match(canvasSource, /<div className="agent-canvas-node-preview-shell">[\s\S]*?<div className="agent-canvas-node-card">[\s\S]*?<div className="agent-canvas-node-preview-anchor">[\s\S]*?<Handle type="source"[\s\S]*?\{composer\}[\s\S]*?<\/div>/);
  assert.match(canvasSource, /data-editor-open=\{isEditorOpen \? 'true' : 'false'\}/);
  assert.match(styles, /\.agent-page \.agent-canvas-node-card\s*> \.agent-node-composer\s*\{/);
  assert.doesNotMatch(styles, /\.agent-page \.agent-node-composer\s*\{[^}]*left:\s*50%/);
  assert.doesNotMatch(styles, /width:\s*min\(680px, calc\(100vw - 96px\)\)/);
});

test('source and target handles become interactive only for selection or connection work', () => {
  assert.match(styles, /\.agent-page \.react-flow__handle\.agent-canvas-node-target\s*\{[^}]*opacity:\s*0[^}]*pointer-events:\s*none/);
  assert.match(styles, /\.agent-page \.react-flow__handle\.agent-canvas-node-quick-add\s*\{[^}]*opacity:\s*0[^}]*pointer-events:\s*none/);
  assert.match(styles, /\.agent-page \.agent-canvas-node\.is-editor-open[\s\S]*?\.agent-canvas-node-quick-add/);
  assert.match(styles, /\.agent-page \.agent-canvas\.is-connecting[\s\S]*?\.agent-canvas-node-target/);
});

test('an existing target drop does not open the empty-canvas continuation menu', () => {
  assert.equal(shouldOpenConnectionDrop({ sourceID: 'image-1', connectionCompleted: true, droppedOnHandle: true }), false);
  assert.equal(shouldOpenConnectionDrop({ sourceID: 'image-1', connectionCompleted: false, droppedOnHandle: false }), true);
  assert.deepEqual(connectionDropChoices({ data: { type: 'image' } }).map(item => item.kind), ['image', 'video']);
});

test('typed edge relation still follows image-to-video semantics', () => {
  const edge = createTypedCanvasEdge(
    { source: 'image-1', target: 'video-1' },
    [
      { id: 'image-1', data: { type: 'image' } },
      { id: 'video-1', data: { type: 'video' } }
    ]
  );
  assert.equal(edge.data.relation, 'first_frame');
});
