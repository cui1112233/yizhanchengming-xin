import assert from 'node:assert/strict';
import test from 'node:test';
import { generationHistoryRows } from './agent-generation-history.js';

test('generation history keeps every reusable media result with its source lineage', () => {
  const rows = generationHistoryRows([
    {
      id: 'task-image', projectId: 'project-a', kind: 'image', modelId: 'image-model', state: 'completed',
      results: [
        { id: 'result-1', assetId: 'asset-1', kind: 'image', asset: { name: 'first.png' } },
        { id: 'result-2', assetId: 'asset-2', kind: 'image', asset: { name: 'second.png' } }
      ]
    },
    { id: 'task-failed', projectId: 'project-b', kind: 'video', modelId: 'video-model', state: 'failed', results: [] }
  ]);

  assert.deepEqual(rows.map(row => ({ taskId: row.taskId, resultId: row.resultId, projectId: row.projectId, kind: row.kind, state: row.state, name: row.name })), [
    { taskId: 'task-image', resultId: 'result-1', projectId: 'project-a', kind: 'image', state: 'completed', name: 'first.png' },
    { taskId: 'task-image', resultId: 'result-2', projectId: 'project-a', kind: 'image', state: 'completed', name: 'second.png' }
  ]);
});
