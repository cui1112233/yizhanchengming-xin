import test from 'node:test';
import assert from 'node:assert/strict';
import { audioEnvelopePoints, audioEnvelopeVolumeAt } from './aicutVolumeEnvelope.js';

test('音量包络用基础音量补齐首尾并只读取音量关键帧', () => {
  const clip = { durationMs: 1000, volume: 1, keyframes: [
    { id: 'move', offsetMs: 100, values: { x: .2 } },
    { id: 'quiet', offsetMs: 250, values: { volume: .5 } },
    { id: 'loud', offsetMs: 750, values: { volume: 1.5 } }
  ] };
  assert.deepEqual(audioEnvelopePoints(clip), [
    { id: '', offsetMs: 0, position: 0, volume: 1 },
    { id: 'quiet', offsetMs: 250, position: .25, volume: .5 },
    { id: 'loud', offsetMs: 750, position: .75, volume: 1.5 },
    { id: '', offsetMs: 1000, position: 1, volume: 1 }
  ]);
});

test('音量包络按相邻点线性插值并保留静音零值', () => {
  const clip = { durationMs: 1000, volume: 1, keyframes: [
    { id: 'mute', offsetMs: 500, values: { volume: 0 } },
    { id: 'end', offsetMs: 1000, values: { volume: 2 } }
  ] };
  assert.equal(audioEnvelopeVolumeAt(clip, 250), .5);
  assert.equal(audioEnvelopeVolumeAt(clip, 500), 0);
  assert.equal(audioEnvelopeVolumeAt(clip, 750), 1);
});

test('音量包络清理越界、重复和无效点', () => {
  const clip = { durationMs: 1000, volume: .8, keyframes: [
    { id: 'bad', offsetMs: -1, values: { volume: 1 } },
    { id: 'a', offsetMs: 400, values: { volume: .2 } },
    { id: 'b', offsetMs: 400, values: { volume: .6 } },
    { id: 'high', offsetMs: 1200, values: { volume: 4 } }
  ] };
  assert.deepEqual(audioEnvelopePoints(clip), [
    { id: '', offsetMs: 0, position: 0, volume: .8 },
    { id: 'b', offsetMs: 400, position: .4, volume: .6 },
    { id: '', offsetMs: 1000, position: 1, volume: .8 }
  ]);
});
