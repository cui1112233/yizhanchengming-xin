import test from 'node:test';
import assert from 'node:assert/strict';
import { buildSelectedNodeContext, canvasSaveAction, createCanvasSaveQueue, mergePendingCanvas, retainCanvasAssetPanel, retainCanvasEditorNodeID, shouldAcceptCanvasRefresh, shouldKeepPendingCreatedTask, shouldPersistFlowChanges, viewportHasChanged } from './agent-canvas-state.js';

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

test('viewport state only updates when React Flow reports a real movement', () => {
  const current = { x: 24, y: -12, zoom: 0.75 };
  assert.equal(viewportHasChanged(current, { x: 24, y: -12, zoom: 0.75 }), false);
  assert.equal(viewportHasChanged(current, { x: 25, y: -12, zoom: 0.75 }), true);
  assert.equal(viewportHasChanged(current, { x: 24, y: -12, zoom: 0.8 }), true);
});

test('newly created task stays active before the project list refresh finishes', () => {
  assert.equal(shouldKeepPendingCreatedTask({ requestedTaskId: 'new-1', activeTaskId: 'new-1', knownTaskIds: ['old-1'] }), true);
  assert.equal(shouldKeepPendingCreatedTask({ requestedTaskId: 'missing', activeTaskId: 'old-1', knownTaskIds: ['old-1'] }), false);
});

test('server canvas refresh keeps node editing surfaces only while their node still exists', () => {
  const nodes = [{ id: 'image-1' }, { id: 'video-1' }];
  assert.equal(retainCanvasEditorNodeID('image-1', nodes), 'image-1');
  assert.equal(retainCanvasEditorNodeID('missing', nodes), '');
  assert.deepEqual(retainCanvasAssetPanel({ mode: 'node-reference', nodeID: 'image-1' }, nodes), { mode: 'node-reference', nodeID: 'image-1' });
  assert.equal(retainCanvasAssetPanel({ mode: 'node-reference', nodeID: 'missing' }, nodes), null);
  assert.equal(retainCanvasAssetPanel({ mode: 'create', position: { x: 1, y: 2 } }, nodes), null);
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

test('canvas save queue batches continuous edits into one trailing save', async () => {
  const calls = [];
  const queue = createCanvasSaveQueue({
    revision: 2,
    delay: 30,
    save: async (canvas, revision) => {
      calls.push({ canvas, revision });
      return { canvas: { ...canvas, revision: revision + 1 } };
    },
    reload: async () => ({ canvas: { version: 1, revision: 2, nodes: [], edges: [] } })
  });

  for (let index = 0; index < 10; index += 1) {
    queue.enqueue({ version: 1, revision: 2, nodes: [{ id: `node-${index}` }], edges: [] });
    await new Promise(resolve => setTimeout(resolve, 4));
  }
  await new Promise(resolve => setTimeout(resolve, 50));

  assert.equal(calls.length, 1);
  assert.equal(calls[0].canvas.nodes[0].id, 'node-9');
  queue.dispose();
});

test('canvas save queue reports the exact remote snapshot that completed synchronization', async () => {
  const synced = [];
  const queue = createCanvasSaveQueue({
    revision: 2,
    delay: 0,
    save: async canvas => ({ canvas: { ...canvas, revision: 3 } }),
    reload: async () => ({ canvas: { version: 1, revision: 2, nodes: [], edges: [] } }),
    onSynced: canvas => synced.push(canvas)
  });

  queue.enqueue({ version: 1, revision: 2, nodes: [{ id: 'latest' }], edges: [] });
  await queue.flush();

  assert.equal(synced.at(-1).revision, 3);
  assert.equal(synced.at(-1).nodes[0].id, 'latest');
  queue.dispose();
});

test('canvas save queue accepts a confirmed server canvas before the next edit', async () => {
  const calls = [];
  const accepted = [];
  const queue = createCanvasSaveQueue({
    revision: 2,
    delay: 5000,
    save: async (canvas, revision) => {
      calls.push({ canvas, revision });
      return { canvas: { ...canvas, revision: revision + 1 } };
    },
    reload: async () => ({ canvas: { version: 1, revision: 2, nodes: [], edges: [] } }),
    onCanvas: canvas => accepted.push(canvas)
  });

  const confirmed = { version: 1, revision: 3, nodes: [{ id: 'video-1', version: 2, status: 'confirmed' }], edges: [] };
  await queue.acceptServerCanvas(confirmed);
  queue.enqueue({ ...confirmed, nodes: [...confirmed.nodes, { id: 'text-1', version: 1 }] });
  await queue.flush();

  assert.equal(calls.length, 1);
  assert.equal(calls[0].revision, 3);
  assert.equal(calls[0].canvas.nodes.at(-1).id, 'text-1');
  assert.equal(accepted.at(-1).revision, 4);
  queue.dispose();
});

test('canvas save queue ignores older server revisions and keeps the newest revision for the next edit', async () => {
  const calls = [];
  const accepted = [];
  const queue = createCanvasSaveQueue({
    revision: 2,
    delay: 5000,
    save: async (canvas, revision) => {
      calls.push({ canvas, revision });
      return { canvas: { ...canvas, revision: revision + 1 } };
    },
    reload: async () => ({ canvas: { version: 1, revision: 5, nodes: [], edges: [] } }),
    onCanvas: canvas => accepted.push(canvas)
  });

  assert.equal(await queue.acceptServerCanvas({ version: 1, revision: 5, nodes: [{ id: 'newest' }], edges: [] }), true);
  assert.equal(await queue.acceptServerCanvas({ version: 1, revision: 4, nodes: [{ id: 'stale' }], edges: [] }), false);
  queue.enqueue({ version: 1, revision: 5, nodes: [{ id: 'edited' }], edges: [] });
  await queue.flush();

  assert.deepEqual(accepted.slice(0, 1).map(canvas => canvas.nodes[0].id), ['newest']);
  assert.equal(calls[0].revision, 5);
  queue.dispose();
});

test('canvas save queue keeps the highest revision when confirmation rereads resolve out of order', async () => {
  const older = deferred();
  const newer = deferred();
  const accepted = [];
  const queue = createCanvasSaveQueue({
    revision: 2,
    delay: 5000,
    save: async canvas => ({ canvas }),
    reload: async () => ({ canvas: { version: 1, revision: 2, nodes: [], edges: [] } }),
    onCanvas: canvas => accepted.push(canvas)
  });
  const acceptOlder = older.promise.then(canvas => queue.acceptServerCanvas(canvas));
  const acceptNewer = newer.promise.then(canvas => queue.acceptServerCanvas(canvas));

  newer.resolve({ version: 1, revision: 4, nodes: [{ id: 'newest' }], edges: [] });
  assert.equal(await acceptNewer, true);
  older.resolve({ version: 1, revision: 3, nodes: [{ id: 'older' }], edges: [] });
  assert.equal(await acceptOlder, false);

  assert.deepEqual(accepted.map(canvas => canvas.nodes[0].id), ['newest']);
  queue.dispose();
});

test('canvas refresh belongs to the task and save queue that requested it', () => {
  const requestedQueue = {};
  assert.equal(shouldAcceptCanvasRefresh({ requestedTaskId: 'task-1', currentTaskId: 'task-1', requestedQueue, currentQueue: requestedQueue }), true);
  assert.equal(shouldAcceptCanvasRefresh({ requestedTaskId: 'task-1', currentTaskId: 'task-2', requestedQueue, currentQueue: requestedQueue }), false);
  assert.equal(shouldAcceptCanvasRefresh({ requestedTaskId: 'task-1', currentTaskId: 'task-1', requestedQueue, currentQueue: {} }), false);
  assert.equal(shouldAcceptCanvasRefresh({ requestedTaskId: 'task-1', currentTaskId: 'task-1', requestedQueue, currentQueue: requestedQueue, mounted: false }), false);
});

test('storyboard batch response is rejected after its submitting project or queue changes', () => {
  const requestedQueue = {};
  assert.equal(shouldAcceptCanvasRefresh({
    requestedTaskId: 'project-1',
    currentTaskId: 'project-2',
    requestedQueue,
    currentQueue: {}
  }), false);
});

test('canvas save queue merges a pending edit with a confirmed server node', async () => {
  const calls = [];
  const queue = createCanvasSaveQueue({
    revision: 2,
    delay: 5000,
    save: async (canvas, revision) => {
      calls.push({ canvas, revision });
      return { canvas: { ...canvas, revision: revision + 1 } };
    },
    reload: async () => ({ canvas: { version: 1, revision: 2, nodes: [], edges: [] } })
  });
  queue.enqueue({ version: 1, revision: 2, nodes: [{ id: 'video-1', version: 1, status: 'draft', config: { prompt: '本地新提示词' } }], edges: [] });

  await queue.acceptServerCanvas({ version: 1, revision: 3, nodes: [{ id: 'video-1', version: 2, status: 'confirmed', assetRef: { id: 'result-1', kind: 'video' }, config: { prompt: '旧提示词' } }], edges: [] });
  await queue.flush();

  assert.equal(calls.length, 1);
  assert.equal(calls[0].revision, 3);
  assert.equal(calls[0].canvas.nodes[0].config.prompt, '本地新提示词');
  assert.equal(calls[0].canvas.nodes[0].status, 'confirmed');
  assert.equal(calls[0].canvas.nodes[0].version, 2);
  assert.equal(calls[0].canvas.nodes[0].assetRef.id, 'result-1');
  queue.dispose();
});

test('pending canvas merge retains local workspace view state', () => {
  const merged = mergePendingCanvas(
    { version: 1, revision: 4, viewState: { activeView: 'workflow', drawerWidth: 288 }, nodes: [], edges: [] },
    { version: 1, revision: 3, viewState: { activeView: 'storyboard', drawerWidth: 340 }, nodes: [], edges: [] }
  );
  assert.equal(merged.revision, 4);
  assert.deepEqual(merged.viewState, { activeView: 'storyboard', drawerWidth: 340 });
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

test('canvas save queue preserves a newly drawn edge after a revision conflict', async () => {
  const calls = [];
  const statuses = [];
  const canvases = [];
  const local = {
    version: 1,
    revision: 1,
    nodes: [{ id: 'image-1', version: 1 }, { id: 'video-1', version: 1 }],
    edges: [{ id: 'edge-image-1-video-1-first_frame', source: 'image-1', target: 'video-1', data: { relation: 'first_frame' } }]
  };
  const queue = createCanvasSaveQueue({
    revision: 1,
    delay: 0,
    save: async (canvas, revision) => {
      calls.push({ canvas, revision });
      if (calls.length === 1) throw Object.assign(new Error('conflict'), { status: 409 });
      return { canvas: { ...canvas, revision: revision + 1 } };
    },
    reload: async () => ({ canvas: { version: 1, revision: 7, nodes: [{ id: 'image-1', version: 2 }, { id: 'video-1', version: 2 }], edges: [] } }),
    onCanvas: canvas => canvases.push(canvas),
    onStatus: status => statuses.push(status)
  });

  queue.enqueue(local);
  await waitForTurn();
  await waitForTurn();
  await waitForTurn();

  assert.equal(calls.length, 2);
  assert.equal(calls[1].revision, 7);
  assert.deepEqual(calls[1].canvas.edges, local.edges);
  assert.deepEqual(canvases.at(-1).edges, local.edges);
  assert.equal(statuses.at(-1), 'saved');
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
