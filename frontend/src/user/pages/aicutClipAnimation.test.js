import test from 'node:test';
import assert from 'node:assert/strict';
import { normalizeClipAnimation, validateClipAnimation, clipAnimationFrame, clipAnimationPreviewStyle } from './aicutClipAnimation.js';
import { buildAICutPreviewFrame } from './aicutPreview.js';

test('old clips stay neutral and malformed settings normalize safely', () => {
  assert.deepEqual(clipAnimationFrame(null, 1000, 500), { x: 0, y: 0, scale: 1, opacity: 1 });
  const normalized = normalizeClipAnimation({ entrance: { preset: 'bad', durationMs: -1 }, loop: { preset: 'float', durationMs: Infinity } });
  assert.equal(normalized.entrance.preset, 'none');
  assert.equal(normalized.loop.durationMs, 2000);
});
test('short clip entrance and exit shrink proportionally without overlap', () => {
  const value = { entrance: { preset: 'fade', durationMs: 800 }, exit: { preset: 'slide', durationMs: 800 } };
  assert.equal(clipAnimationFrame(value, 1000, 250).opacity, .5);
  assert.deepEqual(clipAnimationFrame(value, 1000, 500), { x: 0, y: 0, scale: 1, opacity: 1 });
  assert.equal(clipAnimationFrame(value, 1000, 750).x, .5);
});
test('independent loop and zoom sample output time and preserve base style', () => {
  const value = { entrance: { preset: 'zoom', durationMs: 1000 }, loop: { preset: 'float', durationMs: 2000 } };
  const frame = clipAnimationFrame(value, 3000, 500);
  assert.equal(frame.scale, .9);
  assert.equal(frame.y, .015);
  const style = clipAnimationPreviewStyle({ entrance: { preset: 'fade', durationMs: 1000 } }, 2000, 500, { opacity: .4, transform: 'rotate(20deg)' });
  assert.equal(style.opacity, .2);
  assert.match(style.transform, /rotate\(20deg\)$/);
});

test('command validation rejects unsupported presets and unsafe durations', () => {
  for (const value of [{ entrance: { preset: 'bad' } }, { loop: { preset: 'fade' } }, ...[-1, Infinity, NaN, 10001, 50.5, '500'].map(durationMs => ({ exit: { preset: 'fade', durationMs } }))]) assert.throws(() => validateClipAnimation(value));
  assert.equal(validateClipAnimation(null), null);
  assert.equal(validateClipAnimation({ entrance: { preset: 'fade' } }).entrance.durationMs, 500);
});

test('preview carries output-time animation through reversed frozen media', () => {
  const animation = { entrance: { preset: 'fade', durationMs: 500 } };
  const timeline = { tracks: [{ id: 'v', kind: 'video', clips: [{ id: 'c', kind: 'video', assetId: 'a', startMs: 1000, durationMs: 2000, freezeSourceMs: 5000, reversed: true, animation }] }] };
  const item = buildAICutPreviewFrame(timeline, [{ id: 'a', kind: 'video' }], 1250).visuals[0];
  assert.equal(item.sourceTimeMs, 5000);
  assert.equal(item.outputOffsetMs, 250);
  assert.equal(clipAnimationFrame(item.animation, item.durationMs, item.outputOffsetMs).opacity, .5);
});
