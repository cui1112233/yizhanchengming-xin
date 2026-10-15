import test from 'node:test';
import assert from 'node:assert/strict';
import { normalizeNovelHistory } from './historyPanelModel.js';

test('normalizes history records for display and keeps newest first', () => {
  const result = normalizeNovelHistory({ history: [
    { id: 'old', updated_at: '2026-10-01T00:00:00Z', status: 'done' },
    { id: 'new', updated_at: '2026-10-02T00:00:00Z', status: 'failed' }
  ] });
  assert.deepEqual(result.map(item => item.id), ['new', 'old']);
});
