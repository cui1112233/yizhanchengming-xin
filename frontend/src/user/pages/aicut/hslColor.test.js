import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { normalizeHSLAdjustments, applyHSLToRGB, applyHSLToPixels, createHSLLUTAtlas, hasHSLAdjustments } from './hslColor.js';

test('Go and JavaScript shared reference vectors', () => {
  const vectors = JSON.parse(readFileSync(new URL('./hslColor.reference.json', import.meta.url), 'utf8'));
  for (const { pixel, settings, want } of vectors) assert.deepEqual(applyHSLToRGB(pixel, settings), want);
});

test('normalization clamps finite numbers and rejects coercible or unknown values', () => {
  const value = normalizeHSLAdjustments({ red: { hue: 900, saturation: -200, luminance: 12.5 }, orange: { hue: '10', saturation: NaN, luminance: Infinity }, unknown: { hue: 30 } });
  assert.deepEqual(value.red, { hue: 100, saturation: -100, luminance: 12.5 });
  assert.deepEqual(value.orange, { hue: 0, saturation: 0, luminance: 0 });
  assert.equal(Object.keys(value).length, 8);
  assert.equal(value.unknown, undefined);
  assert.deepEqual(normalizeHSLAdjustments(null), normalizeHSLAdjustments({}));
  value.red.hue = 2;
  assert.equal(normalizeHSLAdjustments().red.hue, 0);
});

test('neutral settings are exact identity and never mutate a pixel', () => {
  const pixel = [201, 73, 19, 101];
  assert.deepEqual(applyHSLToRGB(pixel, {}), pixel);
  assert.notEqual(applyHSLToRGB(pixel), pixel);
  assert.deepEqual(pixel, [201, 73, 19, 101]);
});

test('red hue changes red while preserving green and alpha', () => {
  const settings = { red: { hue: 100 } };
  assert.deepEqual(applyHSLToRGB([255, 0, 0, 77], settings), [255, 128, 0, 77]);
  assert.deepEqual(applyHSLToRGB([0, 255, 0, 43], settings), [0, 255, 0, 43]);
  assert.deepEqual(applyHSLToRGB([255, 0, 0], { red: { hue: -100 } }), [255, 0, 128]);
});

test('saturation and luminance modify their actual HSL channels', () => {
  assert.deepEqual(applyHSLToRGB([255, 0, 0, 255], { red: { saturation: -100 } }), [128, 128, 128, 255]);
  assert.deepEqual(applyHSLToRGB([255, 0, 0], { red: { luminance: 50 } }), [255, 128, 128]);
  assert.deepEqual(applyHSLToRGB([255, 0, 0], { red: { luminance: -50 } }), [128, 0, 0]);
  assert.deepEqual(applyHSLToRGB([192, 64, 64], { red: { saturation: 100 } }), [255, 1, 1]);
});

test('all eight bands operate on their own center', () => {
  for (const [band, rgb] of [['red', [255, 0, 0]], ['orange', [255, 128, 0]], ['yellow', [255, 255, 0]], ['green', [0, 255, 0]], ['cyan', [0, 255, 255]], ['blue', [0, 0, 255]], ['purple', [128, 0, 255]], ['magenta', [255, 0, 255]]]) {
    const result = applyHSLToRGB(rgb, { [band]: { saturation: -100 } });
    assert.ok(Math.max(...result) - Math.min(...result) <= 1, `${band} must desaturate its center`);
  }
});

test('soft overlapping bands affect skin orange and remain continuous at red wrap', () => {
  // HSL (15 degrees, 100%, 50%) lies halfway between red and orange.
  const halfway = applyHSLToRGB([255, 64, 0], { red: { saturation: -100 } });
  assert.ok(halfway[0] >= 190 && halfway[0] <= 193);
  assert.ok(halfway[2] >= 62 && halfway[2] <= 65);
  const skin = [210, 150, 110];
  const orange = applyHSLToRGB(skin, { orange: { saturation: -100 } });
  assert.ok(Math.max(...orange) - Math.min(...orange) < 30);
  assert.deepEqual(applyHSLToRGB(skin, { blue: { saturation: -100 } }), skin);
  const above = applyHSLToRGB([255, 1, 0], { red: { saturation: -100 } });
  const below = applyHSLToRGB([255, 0, 1], { red: { saturation: -100 } });
  assert.ok(Math.max(...above) - Math.min(...above) <= 1);
  assert.ok(Math.max(...below) - Math.min(...below) <= 1);
});

test('black white gray and transparent pixels remain untouched', () => {
  const settings = Object.fromEntries(['red', 'orange', 'yellow', 'green', 'cyan', 'blue', 'purple', 'magenta'].map(band => [band, { hue: 100, saturation: 100, luminance: 100 }]));
  for (const pixel of [[0, 0, 0, 255], [255, 255, 255, 255], [83, 83, 83, 177], [255, 0, 0, 0]]) {
    assert.deepEqual(applyHSLToRGB(pixel, settings), pixel);
  }
});

test('pixel buffer returns a separate RGBA buffer with exact alpha', () => {
  const input = new Uint8ClampedArray([255, 0, 0, 9, 0, 255, 0, 200]);
  const output = applyHSLToPixels(input, { red: { hue: 100 } });
  assert.ok(output instanceof Uint8ClampedArray);
  assert.deepEqual([...output], [255, 128, 0, 9, 0, 255, 0, 200]);
  assert.deepEqual([...input], [255, 0, 0, 9, 0, 255, 0, 200]);
  assert.notEqual(applyHSLToPixels(input), input);
  assert.deepEqual([...applyHSLToPixels(new Uint8ClampedArray())], []);
});

test('malformed pixels and incomplete buffers fail clearly', () => {
  for (const pixel of [null, [], [1, 2], [1, NaN, 3], [1, 2, 3, Infinity], [1, 2, '3'], [256, 0, 0]]) assert.throws(() => applyHSLToRGB(pixel), TypeError);
  assert.throws(() => applyHSLToPixels([1, 2, 3, 4]), TypeError);
  assert.throws(() => applyHSLToPixels(new Uint8ClampedArray(3)), TypeError);
});

test('LUT atlas stores red-fastest blue-slice RGB with exact opaque alpha', () => {
  assert.equal(hasHSLAdjustments({ red: { hue: Infinity } }), false);
  assert.equal(hasHSLAdjustments({ red: { hue: 1 } }), true);
  const lut = createHSLLUTAtlas({ red: { hue: 100 } });
  assert.equal(lut.width, 1089);
  assert.equal(lut.height, 33);
  const at = (r, g, b) => [...lut.pixels.slice(((g * 1089) + b * 33 + r) * 4, ((g * 1089) + b * 33 + r) * 4 + 4)];
  assert.deepEqual(at(32, 0, 0), [255, 128, 0, 255]);
  assert.deepEqual(at(0, 32, 0), [0, 255, 0, 255]);
  assert.deepEqual(at(16, 16, 16), [128, 128, 128, 255]);
});
