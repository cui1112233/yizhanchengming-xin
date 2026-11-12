import test from 'node:test';
import assert from 'node:assert/strict';
import { executeAICutCommand } from '../aicutCommands.js';
import { applyColorGradeToRGB } from './colorGrade.js';
const timeline = () => ({ tracks: [{ id: 'v', kind: 'video', clips: [{ id: 'c', kind: 'video', durationMs: 1000 }] }] });
const apply = color => executeAICutCommand(timeline(), { type: 'clip.visual_processing_set', trackId: 'v', clipId: 'c', processing: { color } });
test('RGB curves and color wheels survive the real command and preserve caller independence', () => {
  const color = { rgbCurves: { red: [{ x: 0, y: 0 }, { x: .5, y: 1 }, { x: 1, y: 1 }] }, wheels: { shadows: { luminance: 10 } } };
  const result = apply(color).timeline.tracks[0].clips[0].visualProcessing.color;
  assert.ok(applyColorGradeToRGB([64, 100, 200, 77], result)[0] > 128);
  color.rgbCurves.red[1].y = 0;
  color.wheels.shadows.luminance = -100;
  assert.equal(result.rgbCurves.red[1].y, 1);
  assert.equal(result.wheels.shadows.luminance, 10);
});
test('strict command rejects malformed curves points wheel ranges and unknown grade keys', () => {
  for (const color of [
    { rgbCurves: null }, { rgbCurves: { unknown: [] } }, { rgbCurves: { red: [{ x: 0, y: 0 }, { x: 0, y: 1 }] } },
    { rgbCurves: { red: [{ x: 0, y: 0 }, { x: 1, y: NaN }] } }, { rgbCurves: { red: [{ x: 0, y: 0, extra: 1 }, { x: 1, y: 1 }] } },
    { wheels: [] }, { wheels: { unknown: {} } }, { wheels: { midtones: { extra: 1 } } }, { wheels: { midtones: { hue: 181 } } },
    { wheels: { shadows: { saturation: -1 } } }, { wheels: { highlights: { luminance: Infinity } } }, { wheels: { shadows: { hue: '1' } } },
  ]) assert.throws(() => apply(color), /RGB|色轮/);
});
