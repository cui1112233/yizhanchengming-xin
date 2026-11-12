import test from 'node:test';
import assert from 'node:assert/strict';
import { createProjectTimeline, loadProjectTimeline, saveProjectTimeline, moveProjectTimelineHistory } from './aicutTimelines.js';
import { scopedAICutOptions } from './aicutTimelineScope.js';

test('named snapshots keep project and timeline identity plus authoritative revision', async () => {
  const calls = [];
  const request = async (path, options) => { calls.push({ path, options }); return { timeline: { revision: 9, tracks: [] } }; };
  await loadProjectTimeline('p-1', 'tl-2', request);
  await saveProjectTimeline('p-1', 'tl-2', { revision: 8, tracks: [] }, request);
  assert.equal(calls[0].path, '/api/ai-cut/projects/p-1/timelines/tl-2/snapshot');
  assert.equal(JSON.parse(calls[1].options.body).expectedRevision, 8);
  await assert.rejects(() => loadProjectTimeline('../p', 'tl-2', request), /无效/);
  await assert.rejects(() => moveProjectTimelineHistory('p-1', 'tl-2', 'erase', 1, request), /无效/);
});
test('timeline creation sends one explicit request without implicit retries', async () => {
  let count = 0;
  await createProjectTimeline('p-1', '分段二', async (_, options) => { count++; assert.equal(JSON.parse(options.body).title, '分段二'); return { timeline: { id: 'tl-2' } }; });
  assert.equal(count, 1);
});
test('legacy tool scope header cannot leak to another project or application', () => {
  const scope = { projectId: 'p', timelineId: 'tl2' };
  const base = { headers: { 'X-Custom': 'retained' } };
  const next = scopedAICutOptions('/api/ai-cut/projects/p/export-jobs', base, scope);
  assert.equal(next.headers.get('X-AICut-Timeline-Id'), 'tl2');
  assert.equal(next.headers.get('X-Custom'), 'retained');
  assert.strictEqual(scopedAICutOptions('/api/ai-cut/projects/p2/export-jobs', base, scope), base);
  assert.strictEqual(scopedAICutOptions('/api/config', base, scope), base);
  assert.equal(base.headers['X-AICut-Timeline-Id'], undefined);
});
