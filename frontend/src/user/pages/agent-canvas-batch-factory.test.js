import test from 'node:test';
import assert from 'node:assert/strict';
import { agentCanvasURL, bookGroupId, projectedNodePresentation } from './agent-canvas-batch-factory.js';

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
