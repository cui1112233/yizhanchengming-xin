import test from 'node:test';
import assert from 'node:assert/strict';
import { buildSelectedNodeContext, canvasSaveAction, createCanvasSaveQueue, shouldKeepPendingCreatedTask, shouldPersistFlowChanges } from './agent-canvas-state.js';

function deferred() {
  let resolve;
  let reject;
  const promise = new Promise((yes, no) => { resolve = yes; reject = no; });
  return { promise, resolve, reject };
}

function waitForTurn() {
  return new Promise(resolve => setTimeout(resolve, 0));
}

test('buildSelectedNodeContext returns only selected public node fields', () => {
  const context = buildSelectedNodeContext({ nodes: [
    { id: 'n1', type: 'input', title: '小说原文', summary: '第一章', version: 2, status: 'confirmed', assetRef: { assetId: 'a1' }, secret: 'no' },
    { id: 'n2', type: 'video', title: '视频', summary: '完成', version: 1, status: 'ready' }
  ] }, ['n1']);
  assert.deepEqual(context, [{ id: 'n1', type: 'input', title: '小说原文', summary: '第一章', version: 2, status: 'confirmed', assetRef: { assetId: 'a1' } }]);
});

test('selection-only flow changes do not create canvas revisions', () => {
  assert.equal(shouldPersistFlowChanges([{ type: 'select', id: 'n1', selected: true }]), false);
  assert.equal(shouldPersistFlowChanges([{ type: 'dimensions', id: 'n1' }]), false);
  assert.equal(shouldPersistFlowChanges([{ type: 'position', id: 'n1', position: { x: 2, y: 3 } }]), true);
});

test('newly created task stays active before the project list refresh finishes', () => {
  assert.equal(shouldKeepPendingCreatedTask({ requestedTaskId: 'new-1', activeTaskId: 'new-1', knownTaskIds: ['old-1'] }), true);
  assert.equal(shouldKeepPendingCreatedTask({ requestedTaskId: 'missing', activeTaskId: 'old-1', knownTaskIds: ['old-1'] }), false);
});

test('canvas save queue keeps the newest change and advances server revision', async () => {
  const first = deferred();
  const calls = [];
  const saved = [];
  const queue = createCanvasSaveQueue({
    revision: 0,
    delay: 0,
    save: (canvas, revision) => {
      calls.push({ canvas, revision });
      if (calls.length === 1) return first.promise;
      return Promise.resolve({ canvas: { ...canvas, revision: 2 } });
    },
    reload: async () => ({ canvas: { version: 1, revision: 9, nodes: [], edges: [] } }),
    onCanvas: canvas => saved.push(canvas),
    onStatus: () => undefined
  });

  queue.enqueue({ version: 1, revision: 0, nodes: [{ id: 'first' }], edges: [] });
  await waitForTurn();
  queue.enqueue({ version: 1, revision: 0, nodes: [{ id: 'latest' }], edges: [] });
  first.resolve({ canvas: { version: 1, revision: 1, nodes: [{ id: 'first' }], edges: [] } });
  await waitForTurn();
  await waitForTurn();

  assert.equal(calls.length, 2);
  assert.equal(calls[0].revision, 0);
  assert.equal(calls[1].revision, 1);
  assert.equal(calls[1].canvas.nodes[0].id, 'latest');
  assert.equal(saved.at(-1).revision, 2);
  queue.dispose();
});

test('canvas save queue reloads authoritative canvas on revision conflict', async () => {
  const statuses = [];
  const canvases = [];
  const fresh = { version: 1, revision: 7, nodes: [{ id: 'server' }], edges: [] };
  const queue = createCanvasSaveQueue({
    revision: 1,
    delay: 0,
    save: async () => { throw Object.assign(new Error('conflict'), { status: 409 }); },
    reload: async () => ({ canvas: fresh }),
    onCanvas: canvas => canvases.push(canvas),
    onStatus: status => statuses.push(status)
  });

  queue.enqueue({ version: 1, revision: 1, nodes: [{ id: 'local' }], edges: [] });
  await waitForTurn();
  await waitForTurn();

  assert.equal(statuses.at(-1), 'conflict');
  assert.deepEqual(canvases.at(-1), fresh);
  queue.dispose();
});

test('canvas save queue flush skips debounce and saves newest canvas', async () => {
  const calls = [];
  const queue = createCanvasSaveQueue({
    revision: 0,
    delay: 5000,
    save: async canvas => {
      calls.push(canvas);
      return { canvas: { ...canvas, revision: 1 } };
    },
    reload: async () => ({ canvas: { version: 1, revision: 0, nodes: [], edges: [] } })
  });

  queue.enqueue({ version: 1, revision: 0, nodes: [{ id: 'latest' }], edges: [] });
  await queue.flush();

  assert.equal(calls.length, 1);
  assert.equal(calls[0].nodes[0].id, 'latest');
  queue.dispose();
});

test('canvas save queue retry resubmits canvas after ordinary failure', async () => {
  const calls = [];
  const statuses = [];
  const queue = createCanvasSaveQueue({
    revision: 3,
    delay: 0,
    save: async canvas => {
      calls.push(canvas);
      if (calls.length === 1) throw new Error('offline');
      return { canvas: { ...canvas, revision: 4 } };
    },
    reload: async () => ({ canvas: { version: 1, revision: 3, nodes: [], edges: [] } }),
    onStatus: status => statuses.push(status)
  });

  queue.enqueue({ version: 1, revision: 3, nodes: [{ id: 'kept' }], edges: [] });
  await queue.flush();
  assert.equal(statuses.at(-1), 'error');
  await queue.retry();

  assert.equal(calls.length, 2);
  assert.equal(calls[1].nodes[0].id, 'kept');
  assert.equal(statuses.at(-1), 'saved');
  queue.dispose();
});

test('manual save action uses flush, retry, or disabled state correctly', () => {
  assert.deepEqual(canvasSaveAction('dirty'), { method: 'flush', label: '保存', disabled: false });
  assert.deepEqual(canvasSaveAction('saving'), { method: 'flush', label: '保存', disabled: false });
  assert.deepEqual(canvasSaveAction('error'), { method: 'retry', label: '重试保存', disabled: false });
  assert.deepEqual(canvasSaveAction('saved'), { method: null, label: '已保存', disabled: true });
  assert.deepEqual(canvasSaveAction('idle'), { method: null, label: '已保存', disabled: true });
  assert.deepEqual(canvasSaveAction('conflict'), { method: null, label: '已重新加载', disabled: true });
});
