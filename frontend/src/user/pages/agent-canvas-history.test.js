import test from 'node:test';
import assert from 'node:assert/strict';
import { createCanvasHistory } from './agent-canvas-history.js';

test('undo and redo treat one canvas command as one history step', () => {
  const history = createCanvasHistory({ nodes: [{ id: 'a', position: { x: 0, y: 0 }, data: {} }], edges: [] });
  history.push({ nodes: [{ id: 'a', position: { x: 40, y: 0 }, data: {} }], edges: [] });
  assert.equal(history.canUndo(), true);
  assert.equal(history.undo().nodes[0].position.x, 0);
  assert.equal(history.canRedo(), true);
  assert.equal(history.redo().nodes[0].position.x, 40);
});

test('pushing after undo clears redo branch', () => {
  const history = createCanvasHistory({ nodes: [], edges: [] });
  history.push({ nodes: [{ id: 'first', position: { x: 0, y: 0 }, data: {} }], edges: [] });
  history.undo();
  history.push({ nodes: [{ id: 'replacement', position: { x: 0, y: 0 }, data: {} }], edges: [] });
  assert.equal(history.canRedo(), false);
  assert.equal(history.current().nodes[0].id, 'replacement');
});

test('reset switches project baseline and clears old history', () => {
  const history = createCanvasHistory({ nodes: [], edges: [] });
  history.push({ nodes: [{ id: 'old', position: { x: 0, y: 0 }, data: {} }], edges: [] });
  history.reset({ nodes: [{ id: 'new', position: { x: 1, y: 2 }, data: {} }], edges: [] });
  assert.equal(history.canUndo(), false);
  assert.equal(history.canRedo(), false);
  assert.equal(history.current().nodes[0].id, 'new');
});

test('history strips selection and measured render state', () => {
  const history = createCanvasHistory({
    nodes: [{ id: 'a', selected: true, dragging: true, measured: { width: 200, height: 100 }, position: { x: 0, y: 0 }, data: { title: '节点' } }],
    edges: [{ id: 'e1', source: 'a', target: 'b', selected: true }]
  });
  const snapshot = history.current();
  assert.equal(snapshot.nodes[0].selected, undefined);
  assert.equal(snapshot.nodes[0].dragging, undefined);
  assert.equal(snapshot.nodes[0].measured, undefined);
  assert.equal(snapshot.edges[0].selected, undefined);
  assert.equal(snapshot.nodes[0].data.title, '节点');
});

test('returned snapshots cannot mutate stored history', () => {
  const history = createCanvasHistory({ nodes: [{ id: 'a', position: { x: 0, y: 0 }, data: {} }], edges: [] });
  const snapshot = history.current();
  snapshot.nodes[0].position.x = 999;
  assert.equal(history.current().nodes[0].position.x, 0);
});
