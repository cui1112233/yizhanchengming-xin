import test from 'node:test';
import assert from 'node:assert/strict';
import { agentCanvasURL, batchFactoryGroupLabel, batchFactoryMappedAction, bindBatchFactoryBookAndNavigate, bookGroupId, focusRequestedBookGroup, latestBookRevision, mappedBatchFactoryBooksNeedingRefresh, mappedNodeStatePresentation, projectedNodePresentation, refreshBatchFactoryBookCanvas, removeBatchFactoryCanvasNodes, runMappedBatchFactoryAudioPlan, runMappedBatchFactoryScriptStage, runMappedBatchFactoryVideoProduction, syncBookThenLoadCanvas, videoNodeTiming } from './agent-canvas-batch-factory.js';

test('opens the bound project at the requested book group', () => {
  assert.equal(
    agentCanvasURL({ projectId: 'agent-1', batchId: 'batch-1', bookId: 'book-1' }),
    '/agent/canvas?taskId=agent-1&batchId=batch-1&bookId=book-1&view=workflow'
  );
});

test('opens the bound project in batch mode without a book selection', () => {
  assert.equal(
    agentCanvasURL({ projectId: 'agent-1', batchId: 'batch-1' }),
    '/agent/canvas?taskId=agent-1&batchId=batch-1&view=workflow'
  );
});

test('uses existing node types inside a book group', () => {
  assert.deepEqual(
    projectedNodePresentation({ type: 'image', groupId: bookGroupId('book-1') }),
    { groupId: 'book:book-1', groupLabel: '书籍素材', sourceLabel: '批量工厂' }
  );
});

test('labels a book group with its mapped script title before falling back to the book ID', () => {
  const nodes = [
    { type: 'agent', groupId: 'book:book-1', data: { type: 'image', title: '角色图' } },
    { type: 'agent', groupId: 'book:book-1', data: { type: 'script', title: '小妾造反后' } }
  ];
  assert.equal(batchFactoryGroupLabel('book:book-1', nodes), '书籍分组 小妾造反后');
  assert.equal(batchFactoryGroupLabel('book:book-2', nodes), '书籍分组 book-2');
});

test('binds a book project before navigating to its canvas URL', async () => {
  const calls = [];
  await bindBatchFactoryBookAndNavigate({
    batchId: 'batch-1',
    bookId: 'book-1',
    bind: async batchId => { calls.push(`bind:${batchId}`); return { projectId: 'agent-1' }; },
    navigate: url => calls.push(`navigate:${url}`)
  });
  assert.deepEqual(calls, [
    'bind:batch-1',
    'navigate:/agent/canvas?taskId=agent-1&batchId=batch-1&bookId=book-1&view=workflow'
  ]);
});

test('synchronizes the requested book before loading its canvas and creating the save queue', async () => {
  const calls = [];
  const result = await syncBookThenLoadCanvas({
    taskId: 'agent-1',
    bookId: 'book-1',
    sync: async (taskId, bookId) => { calls.push(`sync:${taskId}:${bookId}`); },
    getCanvas: async taskId => { calls.push(`canvas:${taskId}`); return { canvas: { revision: 4 } }; },
    createSaveQueue: canvas => { calls.push(`queue:${canvas.revision}`); return { revision: canvas.revision }; }
  });
  assert.deepEqual(calls, ['sync:agent-1:book-1', 'canvas:agent-1', 'queue:4']);
  assert.deepEqual(result.saveQueue, { revision: 4 });
});

test('refreshing a mapped book flushes local edits before it reloads the Batch Factory projection', async () => {
  const calls = [];
  const canvas = { revision: 8, nodes: [{ id: 'merge-1' }], edges: [] };
  const result = await refreshBatchFactoryBookCanvas({
    taskId: 'agent-1',
    bookId: 'book-1',
    queue: { flushRequired: async () => { calls.push('flush'); } },
    sync: async (taskId, bookId) => { calls.push(`sync:${taskId}:${bookId}`); },
    getCanvas: async taskId => { calls.push(`canvas:${taskId}`); return { canvas }; }
  });

  assert.deepEqual(calls, ['flush', 'sync:agent-1:book-1', 'canvas:agent-1']);
  assert.equal(result.canvas, canvas);
});

test('refreshing a mapped book stops after a task switch instead of applying the old book response', async () => {
  const calls = [];
  const result = await refreshBatchFactoryBookCanvas({
    taskId: 'agent-1',
    bookId: 'book-1',
    queue: { flushRequired: async () => { calls.push('flush'); } },
    sync: async () => { calls.push('sync'); },
    getCanvas: async () => { calls.push('canvas'); return { canvas: { revision: 8, nodes: [], edges: [] } }; },
    isCurrent: () => false
  });

  assert.deepEqual(calls, ['flush']);
  assert.deepEqual(result, { ignored: true });
});

test('focuses only nodes in the requested book group', () => {
  const focused = [];
  const nodes = [{ id: 'book-node', groupId: 'book:book-1' }, { id: 'other-node', groupId: 'book:book-2' }];
  assert.deepEqual(focusRequestedBookGroup({ bookId: 'book-1', nodes, focus: groupNodes => focused.push(...groupNodes.map(node => node.id)) }).map(node => node.id), ['book-node']);
  assert.deepEqual(focused, ['book-node']);
});

test('removes mapped nodes only after flushing and refreshes the projected canvas', async () => {
  const calls = [];
  await removeBatchFactoryCanvasNodes({
    taskId: 'agent-1',
    nodeIds: ['node-1'],
    flush: async () => calls.push('flush'),
    remove: async (taskId, nodeId) => calls.push(`remove:${taskId}:${nodeId}`),
    refresh: async () => calls.push('refresh')
  });
  assert.deepEqual(calls, ['flush', 'remove:agent-1:node-1', 'refresh']);
});

test('does not remove or refresh after the active task changes during flush', async () => {
  const calls = [];
  let current = true;
  await removeBatchFactoryCanvasNodes({
    taskId: 'agent-1',
    nodeIds: ['node-1'],
    flush: async () => { calls.push('flush'); current = false; },
    remove: async () => calls.push('remove'),
    refresh: async () => calls.push('refresh'),
    isCurrent: () => current
  });
  assert.deepEqual(calls, ['flush']);
});

test('stops before later mapped deletes when an earlier delete changes the active task', async () => {
  const calls = [];
  let current = true;
  await removeBatchFactoryCanvasNodes({
    taskId: 'agent-1',
    nodeIds: ['node-1', 'node-2'],
    remove: async (_taskId, nodeId) => { calls.push(`remove:${nodeId}`); current = false; },
    refresh: async () => calls.push('refresh'),
    isCurrent: () => current
  });
  assert.deepEqual(calls, ['remove:node-1']);
});

test('shows a timing mismatch after manual override', () => {
  const timing = videoNodeTiming({ audioSeconds: 7.4, selectedSeconds: 15, allowedSeconds: [5, 10, 15] });
  assert.match(timing.warning, /与已连接音频时长不一致/);
  assert.equal(timing.requestedSeconds, 15);
});

test('chooses the absolute nearest supported duration for linked audio', () => {
  const timing = videoNodeTiming({ audioSeconds: 11, allowedSeconds: [5, 10, 15] });
  assert.equal(timing.requestedSeconds, 10);
  assert.equal(timing.adjusted, true);
});

test('maps projected audio timing into the Batch Factory video action', () => {
  assert.deepEqual(batchFactoryMappedAction({
    groupId: 'book:book-1',
    config: {
      batchId: 'batch-1', bookId: 'book-1', resourceType: 'video', resourceId: 'video-1', resourceRevision: 7,
      connectedAudioDurationSeconds: 11, allowedDurationSeconds: [5, 10, 15]
    }
  }), {
    type: 'video', batchId: 'batch-1', bookId: 'book-1', resourceId: 'video-1', durationSeconds: 10, expectedRevision: 7
  });
});

test('submits and retries only the selected mapped video through its effective executor', async () => {
  for (const [canRetry, mode] of [[false, 'missing'], [true, 'force']]) {
    const calls = [];
    await runMappedBatchFactoryVideoProduction({
      action: {
        type: 'video', batchId: 'batch-1', bookId: 'book-1', resourceId: 'video-selected',
        provider: 'doubao_local_executor', modelId: 'local-doubao-executor-video', canRetry
      },
      requestId: `request-${mode}`,
      runStage: async (...args) => { calls.push(args); return { state: 'queued' }; }
    });
    assert.deepEqual(calls, [[
      'batch-1', 'book-1', 'video',
      { mode, requestId: `request-${mode}`, videoId: 'video-selected', provider: 'doubao_local_executor', videoModelId: 'local-doubao-executor-video' }
    ]]);
  }
});

test('runs a mapped audio card as the batch factory audio plan, not a merge', async () => {
  const calls = [];
  await runMappedBatchFactoryAudioPlan({
    action: { type: 'audio', batchId: 'batch-1', bookId: 'book-1', resourceId: 'tts-before-merge' },
    requestId: 'canvas-audio-1',
    createAudioPlan: async (...args) => calls.push(args)
  });
  assert.deepEqual(calls, [[
    'batch-1', 'book-1', { requestId: 'canvas-audio-1' }
  ]]);
});

test('presents mapped failures with the safe error and retry state', () => {
  assert.deepEqual(mappedNodeStatePresentation({
    groupId: 'book:book-1',
    status: 'failed',
    config: { mappedState: 'failed', errorMessage: '批量工厂任务失败，请重试', canRetry: true }
  }), {
    state: 'failed', errorMessage: '批量工厂任务失败，请重试', canRetry: true, busy: false, message: '批量工厂任务失败，请重试'
  });
});

test('keeps a mapped pending audio plan actionable before its first audio file exists', () => {
  assert.deepEqual(mappedNodeStatePresentation({
    groupId: 'book:book-1',
    status: 'queued',
    config: { resourceType: 'audio', audioPlanPending: true, mappedState: 'queued' }
  }), {
    state: 'draft', errorMessage: '', canRetry: false, busy: false, message: '待生成配音'
  });
});

test('routes mapped image actions back to the batch factory record', () => {
  assert.deepEqual(batchFactoryMappedAction({
    groupId: 'book:book-1',
    config: { batchId: 'batch-1', bookId: 'book-1', resourceType: 'book_asset', resourceId: 'asset-1' }
  }), { type: 'image', batchId: 'batch-1', bookId: 'book-1', resourceId: 'asset-1' });
});

test('keeps the final merge card actionable before its first merge job exists', () => {
  const node = {
    groupId: 'book:book-1',
    status: 'queued',
    config: {
      batchId: 'batch-1', bookId: 'book-1', resourceType: 'merge', resourceId: 'book-1',
      canMerge: true, canPublish: false, mergeState: 'ready'
    }
  };
  assert.deepEqual(batchFactoryMappedAction(node), {
    type: 'merge', batchId: 'batch-1', bookId: 'book-1', resourceId: 'book-1',
    canMerge: true, canPublish: false, mergeState: 'ready', mergeReason: ''
  });
  assert.deepEqual(mappedNodeStatePresentation(node), {
    state: 'ready', errorMessage: '', canRetry: false, busy: false, message: '所有分镜已就绪，可以合成最终视频'
  });
});

test('runs mapped script stages with the selected production line count', async () => {
  const calls = [];
  await runMappedBatchFactoryScriptStage({
    action: { type: 'script', batchId: 'batch-1', bookId: 'book-1' },
    stage: 'assets',
    requestId: 'canvas-assets-1',
    options: { contentRangeLines: 12 },
    runStage: async (...args) => calls.push(args)
  });
  assert.deepEqual(calls, [[
    'batch-1', 'book-1', 'assets', { mode: 'missing', requestId: 'canvas-assets-1', contentRangeLines: 12 }
  ]]);
});

test('uses the newest book revision returned by each save endpoint', () => {
  assert.equal(latestBookRevision({ revision: 14 }, 9), 14);
  assert.equal(latestBookRevision({ book: { revision: 15 } }, 9), 15);
  assert.equal(latestBookRevision({}, 9), 9);
});

test('refreshes each mapped book once while it has queued or running production', () => {
  assert.deepEqual(mappedBatchFactoryBooksNeedingRefresh([
    { id: 'script', data: { config: { source: 'batch_factory', batchId: 'batch-1', bookId: 'book-1', mappedState: 'succeeded' } } },
    { id: 'video-1', data: { config: { source: 'batch_factory', batchId: 'batch-1', bookId: 'book-1', mappedState: 'running' } } },
    { id: 'video-2', data: { config: { source: 'batch_factory', batchId: 'batch-1', bookId: 'book-1', mappedState: 'queued' } } },
    { id: 'video-3', data: { config: { source: 'batch_factory', batchId: 'batch-1', bookId: 'book-2', mappedState: 'failed' } } },
    { id: 'manual', data: { config: { source: 'manual', batchId: 'batch-1', bookId: 'book-3', mappedState: 'running' } } }
  ]), [{ batchId: 'batch-1', bookId: 'book-1' }]);
});
