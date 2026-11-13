import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { parse } from '@babel/parser';
import { createCanvasWriteGate } from './canvas-pet-service.js';
import { persistClipboardImage } from './agent-canvas-asset-input.js';
import { runAfterDurableCanvasSave } from './agent-script-shots.js';

const source = readFileSync(new URL('../components/AgentCanvas.jsx', import.meta.url), 'utf8');
const ast = parse(source, { sourceType: 'module', plugins: ['jsx'] });
function actualHandler(name, bindings) {
  let found;
  const walk = node => { if (!node || typeof node !== 'object') return; if (node.type === 'FunctionDeclaration' && node.id?.name === name) found = node; for (const value of Object.values(node)) { if (Array.isArray(value)) value.forEach(walk); else if (value && typeof value === 'object') walk(value); } };
  walk(ast); assert.ok(found, name);
  return Function(...Object.keys(bindings), `return (${source.slice(found.start, found.end)})`)(...Object.values(bindings));
}

test('actual screenshot paste keeps approval blocked until upload and node insertion finish', async () => {
  const gate = createCanvasWriteGate(); let complete; let added = false;
  const handler = actualHandler('handleCanvasPaste', {
    canvasWrites: { current: gate }, firstClipboardImage: () => ({ name: 'screenshot.png', type: 'image/png' }), shouldHandleCanvasImagePaste: () => true,
    taskIdRef: { current: 'one' }, lastCanvasPointerRef: { current: null }, flowInstance: null,
    canvasCenterPosition: () => ({ x: 1, y: 2 }), message: { loading() {}, success() {}, error() {} },
    persistClipboardImage, uploadAgentAsset: () => new Promise(resolve => { complete = resolve; }), addAssetNode: () => { added = true; }
  });
  const pending = handler({ preventDefault() {}, target: { tagName: 'DIV' } });
  assert.throws(() => gate.beginApply(), /仍在处理/);
  complete({ asset: { id: 'uploaded', kind: 'image' } }); await pending;
  assert.equal(added, true); gate.beginApply()();
});

test('actual script split keeps approval blocked through save and result handling', async () => {
  const gate = createCanvasWriteGate(); let complete; let written = false;
  const handler = actualHandler('splitScriptForNode', {
    canvasWrites: { current: gate }, taskId: 'one', taskIdRef: { current: 'one' }, setScriptShotsByNode() {},
    runAfterDurableCanvasSave, onFlushCanvas: async () => ({ revision: 1 }), nodesRef: { current: [{ id: 's1', data: { type: 'script', version: 1, config: { sourceText: '故事' } } }] },
    createAgentScriptSplit: () => new Promise(resolve => { complete = resolve; }), normalizeScriptShots: () => ({ state: 'completed', shots: [] }),
    onScriptShotsChange() {}, mergeGenerationForProject() {}, handleNodeConfigChange: () => { written = true; }, scriptSourceMarker: () => 'same', refreshGenerationTasks() {}
  });
  const pending = handler('s1', { sourceText: '故事', modelId: 'test-only' });
  await Promise.resolve(); await Promise.resolve();
  assert.throws(() => gate.beginApply(), /仍在处理/);
  assert.equal(typeof complete, 'function'); complete({}); await pending;
  assert.equal(written, true); gate.beginApply()();
});
