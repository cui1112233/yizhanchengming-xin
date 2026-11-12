import assert from 'node:assert/strict';
import test from 'node:test';
import { executeAICutCommand, restoreTimelineSnapshot } from '../aicutCommands.js';

test('template command produces persistent tracks as one reversible timeline result', () => {
  const initial = { revision: 7, tracks: [{ id: 'locked', kind: 'video', name: '原片', locked: true, order: 20, clips: [] }], markers: [{ id: 'mark', timeMs: 200 }] };
  let sequence = 0;
  const result = executeAICutCommand(initial, { type: 'timeline.template_apply', templateId: 'template-interview', assets: [{ id: 'image', name: '自有图片', kind: 'image', url: '/assets/own.png' }], options: { visualAssetId: 'image', title: '人物访谈', startMs: 1000, durationMs: 5000 } }, { idFactory: () => `new-${++sequence}` });
  assert.equal(result.changed, true);
  assert.equal(result.timeline.revision, 7);
  assert.equal(result.timeline.tracks.length, 3);
  assert.strictEqual(result.timeline.tracks[0], initial.tracks[0]);
  assert.equal(result.metadata.template.trackCount, 2);
  assert.ok(result.selectionKey);
  assert.ok(result.timeline.tracks[1].order > 20);
  assert.equal(JSON.parse(JSON.stringify(result.timeline)).tracks[2].clips[0].text, '人物访谈');
  assert.deepEqual(restoreTimelineSnapshot(initial, 8), { ...initial, revision: 8 });
  assert.equal(Object.hasOwn(result.timeline, 'editorTemplateId'), false);
});
