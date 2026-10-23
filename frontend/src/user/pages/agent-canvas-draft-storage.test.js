import test from 'node:test';
import assert from 'node:assert/strict';
import { canvasDraftKey, clearCanvasDraft, isDraftAtLeastAsFresh, loadCanvasDraft, saveCanvasDraft } from './agent-canvas-draft-storage.js';

function createStorage() {
  const values = new Map();
  return {
    getItem(key) { return values.has(key) ? values.get(key) : null; },
    setItem(key, value) { values.set(key, String(value)); },
    removeItem(key) { values.delete(key); }
  };
}

function canvas(nodeID = 'node-1') {
  return {
    version: 1,
    revision: 3,
    nodes: [{ id: nodeID, type: 'image', title: '图片', selected: true, position: { x: 10, y: 20 } }],
    edges: [],
    viewport: { x: 4, y: 8, zoom: 1 },
    appearance: { edgeColor: 'blue' },
    viewState: { activeView: 'workflow' }
  };
}

test('canvas drafts are isolated by account and project', () => {
  const storage = createStorage();
  assert.equal(saveCanvasDraft(storage, 'alice', 'task-a', canvas('local-a'), 3), true);
  assert.equal(saveCanvasDraft(storage, 'alice', 'task-b', canvas('local-b'), 4), true);

  assert.equal(loadCanvasDraft(storage, 'alice', 'task-a').canvas.nodes[0].id, 'local-a');
  assert.equal(loadCanvasDraft(storage, 'alice', 'task-b').canvas.nodes[0].id, 'local-b');
  assert.equal(loadCanvasDraft(storage, 'bob', 'task-a'), null);
  assert.notEqual(canvasDraftKey('alice', 'task-a'), canvasDraftKey('alice', 'task-b'));
});

test('canvas draft strips transient selected node state', () => {
  const storage = createStorage();
  saveCanvasDraft(storage, 'alice', 'task-a', canvas(), 3);
  const saved = loadCanvasDraft(storage, 'alice', 'task-a');

  assert.equal(Object.hasOwn(saved.canvas.nodes[0], 'selected'), false);
  assert.deepEqual(saved.canvas.nodes[0].position, { x: 10, y: 20 });
});

test('invalid canvas draft data is ignored without throwing', () => {
  const storage = createStorage();
  storage.setItem(canvasDraftKey('alice', 'task-a'), '{broken');
  assert.equal(loadCanvasDraft(storage, 'alice', 'task-a'), null);

  storage.setItem(canvasDraftKey('alice', 'task-a'), JSON.stringify({ version: 1, canvas: { nodes: 'not-an-array' }, baseRevision: 3 }));
  assert.equal(loadCanvasDraft(storage, 'alice', 'task-a'), null);
});

test('only drafts based on the current or newer revision can restore', () => {
  const draft = { canvas: canvas('local'), baseRevision: 4, savedAt: '2026-10-21T00:00:00.000Z' };
  assert.equal(isDraftAtLeastAsFresh(draft, { revision: 4 }), true);
  assert.equal(isDraftAtLeastAsFresh(draft, { revision: 5 }), false);
});

test('clearCanvasDraft removes only its scoped draft', () => {
  const storage = createStorage();
  saveCanvasDraft(storage, 'alice', 'task-a', canvas('a'), 2);
  saveCanvasDraft(storage, 'alice', 'task-b', canvas('b'), 2);

  assert.equal(clearCanvasDraft(storage, 'alice', 'task-a'), true);
  assert.equal(loadCanvasDraft(storage, 'alice', 'task-a'), null);
  assert.equal(loadCanvasDraft(storage, 'alice', 'task-b').canvas.nodes[0].id, 'b');
});
