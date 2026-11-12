import test from 'node:test';
import assert from 'node:assert/strict';
import { addStickerLayer } from './stickerLayer.js';

test('sticker adds an independent scaled image overlay above existing locked footage', () => {
  const original = { revision: 4, tracks: [{ id: 't', order: 20, locked: true, kind: 'video', clips: [{ id: 'c', assetId: 'video' }] }] };
  let id = 0;
  const result = addStickerLayer(original, { id: 'own-image', kind: 'image', name: '透明贴纸' }, 1500, 5000, () => `new-${++id}`);
  assert.strictEqual(result.timeline.tracks[0], original.tracks[0]);
  assert.equal(result.timeline.tracks[1].order, 21);
  assert.equal(result.timeline.tracks[1].clips[0].canvasTransform.scaleX, .35);
  assert.equal(result.timeline.tracks[1].clips[0].startMs, 1500);
  assert.equal(result.timeline.tracks[1].clips[0].assetId, 'own-image');
  assert.equal(result.timeline.revision, 4);
  assert.equal(original.tracks.length, 1);
});
test('invalid input cannot create a fake overlay or alter the source timeline', () => {
  const timeline = { tracks: [] };
  assert.throws(() => addStickerLayer(timeline, { id: 'v', kind: 'video' }), /图片贴纸/);
  assert.throws(() => addStickerLayer(timeline, { id: 'i', kind: 'image' }, -1), /时间范围/);
  assert.deepEqual(timeline, { tracks: [] });
});
