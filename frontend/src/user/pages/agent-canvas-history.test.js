import test from 'node:test';
import assert from 'node:assert/strict';
import { createCanvasHistory } from './agent-canvas-history.js';
import * as canvasAppearance from './agent-canvas-appearance.js';

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

test('current edge visibility survives node edit undo and redo without changing edge identity or data', () => {
  assert.equal(typeof canvasAppearance.applyCanvasEdgeAppearance, 'function');
  const applyAppearance = canvasAppearance.applyCanvasEdgeAppearance;
  const visibleAppearance = canvasAppearance.normalizeCanvasAppearance({ edgesVisible: true });
  const hiddenAppearance = canvasAppearance.normalizeCanvasAppearance({ edgesVisible: false });
  const edge = {
    id: 'edge-a-b',
    source: 'a',
    target: 'b',
    hidden: false,
    data: { relation: 'reference', businessLabel: 'keep-me', appearance: visibleAppearance }
  };
  const history = createCanvasHistory({
    nodes: [{ id: 'a', position: { x: 0, y: 0 }, data: { title: 'before' } }],
    edges: [edge]
  });

  const hiddenEdges = history.current().edges.map(item => applyAppearance(item, hiddenAppearance));
  history.push({
    nodes: [{ id: 'a', position: { x: 0, y: 0 }, data: { title: 'after' } }],
    edges: hiddenEdges
  });

  const undoneEdges = history.undo().edges.map(item => applyAppearance(item, hiddenAppearance));
  const redoneEdges = history.redo().edges.map(item => applyAppearance(item, hiddenAppearance));
  for (const restoredEdges of [undoneEdges, redoneEdges]) {
    assert.equal(restoredEdges.length, 1);
    assert.deepEqual(restoredEdges.map(item => item.id), ['edge-a-b']);
    assert.equal(restoredEdges[0].data.appearance.edgesVisible, false);
    assert.equal(restoredEdges[0].data.relation, 'reference');
    assert.equal(restoredEdges[0].data.businessLabel, 'keep-me');
  }

  const shownEdges = redoneEdges.map(item => applyAppearance(item, visibleAppearance));
  assert.deepEqual(shownEdges.map(item => item.id), ['edge-a-b']);
  assert.equal(shownEdges.length, 1);
  assert.equal(shownEdges[0].data.appearance.edgesVisible, true);
  assert.equal(shownEdges[0].data.businessLabel, 'keep-me');
});
