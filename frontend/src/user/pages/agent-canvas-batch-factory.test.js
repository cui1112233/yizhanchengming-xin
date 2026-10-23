import test from 'node:test';
import assert from 'node:assert/strict';
import { agentCanvasURL, bindBatchFactoryBookAndNavigate, bookGroupId, focusRequestedBookGroup, projectedNodePresentation, removeBatchFactoryCanvasNodes, syncBookThenLoadCanvas } from './agent-canvas-batch-factory.js';

test('opens the bound project at the requested book group', () => {
  assert.equal(
    agentCanvasURL({ projectId: 'agent-1', batchId: 'batch-1', bookId: 'book-1' }),
    '/agent/canvas?taskId=agent-1&batchId=batch-1&bookId=book-1&view=workflow'
  );
});

test('uses existing node types inside a book group', () => {
  assert.deepEqual(
    projectedNodePresentation({ type: 'image', groupId: bookGroupId('book-1') }),
    { groupId: 'book:book-1', groupLabel: '书籍素材', sourceLabel: '批量工厂' }
  );
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
