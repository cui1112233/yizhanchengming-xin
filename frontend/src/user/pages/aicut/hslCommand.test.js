import test from 'node:test';
import assert from 'node:assert/strict';
import { executeAICutCommand } from '../aicutCommands.js';
import { applyHSLToRGB } from './hslColor.js';

const timeline = () => ({ tracks: [{ id: 'v', kind: 'video', clips: [{ id: 'c', kind: 'video', durationMs: 1000, sourceDurationMs: 1000 }] }] });
const apply = (source, hsl) => executeAICutCommand(source, { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { color: { hsl } } });

test('visual command saves actual HSL effect without aliasing caller settings or source timeline', () => {
  const source = timeline();
  const hsl = { red: { hue: 100, saturation: 0, luminance: 0 } };
  const result = apply(source, hsl);
  const saved = result.timeline.tracks[0].clips[0].visualProcessing.color.hsl;
  assert.deepEqual(applyHSLToRGB([255, 0, 0, 77], saved), [255, 128, 0, 77]);
  assert.deepEqual(applyHSLToRGB([0, 255, 0, 43], saved), [0, 255, 0, 43]);
  hsl.red.hue = -100;
  assert.equal(saved.red.hue, 100);
  assert.equal(source.tracks[0].clips[0].visualProcessing, undefined);
  const again = apply(result.timeline, saved);
  assert.deepEqual(again.timeline.tracks[0].clips[0].visualProcessing.color.hsl, saved);
});

test('HSL command resets active colors and supports sparse valid band adjustments', () => {
  const active = apply(timeline(), { orange: { saturation: -25 } }).timeline;
  const reset = apply(active, {}).timeline.tracks[0].clips[0].visualProcessing.color.hsl;
  assert.deepEqual(applyHSLToRGB([210, 150, 110, 255], reset), [210, 150, 110, 255]);
  assert.equal(active.tracks[0].clips[0].visualProcessing.color.hsl.orange.saturation, -25);
});

test('HSL command rejects malformed bands channels nonfinite values and out-of-range values atomically', () => {
  for (const hsl of [null, [], 5, 'red', { unknown: { hue: 1 } }, { red: null }, { red: [] }, { red: 10 }, { red: { unsupported: 1 } }, { red: { hue: '20' } }, { red: { hue: NaN } }, { red: { saturation: Infinity } }, { red: { luminance: -101 } }, { red: { hue: 101 } }, { red: { hue: undefined } }]) {
    const source = timeline();
    assert.throws(() => apply(source, hsl), /HSL/);
    assert.equal(source.tracks[0].clips[0].visualProcessing, undefined);
  }
});

test('legacy visual command retains absent HSL field shape', () => {
  const result = executeAICutCommand(timeline(), { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { color: { exposure: .5 } } });
  assert.equal(Object.hasOwn(result.timeline.tracks[0].clips[0].visualProcessing.color, 'hsl'), false);
});
