import test from 'node:test';
import assert from 'node:assert/strict';
import { saveTimelineDraftCopy } from './aicutDraftRecovery.js';

test('a conflicting draft is saved to a new timeline without overwriting either original', async () => {
  const draft = { revision: 8, tracks: [{ id: 'own', clips: [] }] };
  let creates = 0;
  const api = { createProjectTimeline: async () => { creates++; return { id: 'copy' }; }, saveProjectTimeline: async (project, id, value) => { assert.equal(project, 'p'); assert.equal(id, 'copy'); assert.equal(value.revision, 0); return { ...value, revision: 1 }; } };
  const result = await saveTimelineDraftCopy('p', '恢复副本', draft, '', api);
  assert.equal(result.id, 'copy'); assert.equal(creates, 1); assert.equal(draft.revision, 8);
});
test('retry keeps the created ID and recognizes a previously completed save', async () => {
  const draft = { revision: 8, tracks: [] };
  await assert.rejects(() => saveTimelineDraftCopy('p', '副本', draft, '', { createProjectTimeline: async () => ({ id: 'copy' }), saveProjectTimeline: async () => { throw new Error('网络断开'); } }), error => error.createdTimelineId === 'copy');
  const result = await saveTimelineDraftCopy('p', '副本', draft, 'copy', { loadProjectTimeline: async () => ({ ...draft, revision: 1 }), createProjectTimeline: async () => { throw new Error('不得重复建立'); } });
  assert.equal(result.timeline.revision, 1);
});
