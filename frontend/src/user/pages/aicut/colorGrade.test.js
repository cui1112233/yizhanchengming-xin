import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { normalizeRGBColorCurves, normalizeColorWheels, applyColorGradeToRGB, hasColorGradeAdjustments, createColorGradeLUTAtlas } from './colorGrade.js';

test('Go and JavaScript share RGB wheel ordering gamut and hue-wrap vectors', () => {
  const vectors = JSON.parse(readFileSync(new URL('./colorGrade.reference.json', import.meta.url), 'utf8'));
  for (const { pixel, color, want } of vectors) assert.deepEqual(applyColorGradeToRGB(pixel, color), want);
});

test('channel curves change only their selected channel and preserve alpha', () => {
  const color = { rgbCurves: { red: [{ x: 0, y: 0 }, { x: .5, y: 1 }, { x: 1, y: 1 }] } };
  assert.deepEqual(applyColorGradeToRGB([64, 100, 200, 77], color), [128, 100, 200, 77]);
  assert.deepEqual(applyColorGradeToRGB([64, 100, 200, 0], color), [64, 100, 200, 0]);
});

test('three wheels apply weighted chroma and luminance from original Rec709 luma', () => {
  assert.deepEqual(applyColorGradeToRGB([128, 128, 128, 77], { wheels: { shadows: { hue: 0, saturation: 100 } } }), [140, 125, 125, 77]);
  const lifted = applyColorGradeToRGB([0, 0, 0, 255], { wheels: { shadows: { luminance: 100 } } });
  assert.deepEqual(lifted, [255, 255, 255, 255]);
  assert.deepEqual(applyColorGradeToRGB([255, 255, 255, 255], { wheels: { highlights: { luminance: -100 } } }), [0, 0, 0, 255]);
  assert.deepEqual(applyColorGradeToRGB([0, 0, 0], { wheels: { highlights: { luminance: 100 } } }), [0, 0, 0]);
});

test('grade identity ignores unused hue and stays immutable', () => {
  const pixel = [17, 83, 219, 101];
  assert.deepEqual(applyColorGradeToRGB(pixel, { wheels: { shadows: { hue: 120 } } }), pixel);
  assert.deepEqual(pixel, [17, 83, 219, 101]);
  assert.equal(hasColorGradeAdjustments({ wheels: { shadows: { hue: 120 } } }), false);
  assert.equal(hasColorGradeAdjustments({ rgbCurves: { blue: [{ x: 0, y: 0 }, { x: .4, y: .4 }, { x: 1, y: 1 }] } }), false);
  assert.equal(hasColorGradeAdjustments({ wheels: { midtones: { saturation: 1 } } }), true);
});

test('safe normalization repairs malformed curves and clamps wheel values', () => {
  assert.deepEqual(normalizeRGBColorCurves({ red: [{ x: 1, y: 0 }, { x: 0, y: 1 }] }).red, [{ x: 0, y: 0 }, { x: 1, y: 1 }]);
  assert.deepEqual(normalizeColorWheels({ shadows: { hue: 1000, saturation: -3, luminance: NaN } }).shadows, { hue: 180, saturation: 0, luminance: 0 });
});

test('combined lattice applies RGB curves then wheels then HSL and stays in gamut', () => {
  const color = { rgbCurves: { green: [{ x: 0, y: 0 }, { x: 1, y: .5 }] }, wheels: { shadows: { hue: -180, saturation: 100, luminance: -100 }, midtones: { hue: 180, saturation: 100, luminance: 100 } }, hsl: { red: { saturation: -25 } } };
  const lut = createColorGradeLUTAtlas(color);
  assert.equal(lut.pixels.length, 33 * 33 * 33 * 4);
  assert.ok([...lut.pixels].every(value => Number.isInteger(value) && value >= 0 && value <= 255));
  assert.deepEqual(applyColorGradeToRGB([255, 0, 0], { rgbCurves: { red: [{ x: 0, y: 0 }, { x: 1, y: .5 }] }, hsl: { red: { luminance: 100 } } }), [255, 255, 255]);
});
