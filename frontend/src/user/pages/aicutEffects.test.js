import assert from 'node:assert/strict';
import test from 'node:test';
import { normalizeVisualEffect, subtitleAnimationCatalog, subtitleStyleCatalog, textStyleCatalog, transitionCatalog, visualEffectCatalog, visualEffectCSS, visualEffectPreset, visualEffectPreviewStyle } from './aicutEffects.js';

test('normalizes missing visual effect to original', () => {
  assert.deepEqual(normalizeVisualEffect(), { preset: 'original', brightness: 0, contrast: 1, saturation: 1, blur: 0, grayscale: false });
});

test('maps visual effect to browser filter', () => {
  const css = visualEffectCSS({ preset: 'custom', brightness: .2, contrast: 1.1, saturation: .8, blur: 2, grayscale: false });
  assert.match(css, /brightness\(1\.2\)/);
  assert.match(css, /contrast\(1\.1\)/);
  assert.match(css, /saturate\(0\.8\)/);
  assert.match(css, /blur\(2px\)/);
});

test('presets are deterministic and replace previous settings', () => {
  assert.deepEqual(visualEffectPreset('grayscale'), { preset: 'grayscale', brightness: 0, contrast: 1.05, saturation: 0, blur: 0, grayscale: true });
  assert.equal(visualEffectPreset('warm').preset, 'warm');
  assert.equal(visualEffectPreset('missing').preset, 'original');
});

test('第二批风格滤镜在浏览器产生各自稳定且不同的预览', () => {
  const ids = ['golden-hour', 'noir', 'vintage', 'cyberpunk', 'vaporwave', 'mint-haze', 'bleach-bypass', 'winter-cool', 'rosy', 'film', 'super-8', 'instant-print'];
  const filters = ids.map(id => visualEffectCSS(visualEffectPreset(id)));
  assert.equal(new Set(filters).size, ids.length);
  assert.match(filters[0], /brightness/);
  assert.match(filters[1], /grayscale/);
});

test('first batch catalogs use stable ids and expose every usable family', () => {
  assert.deepEqual(transitionCatalog.map(item => item.id), ['cross-dissolve', 'fade', 'flash-black', 'flash-white', 'slide-left', 'slide-right', 'slide-up', 'slide-down', 'zoom', 'blur']);
  assert.ok(visualEffectCatalog.some(item => item.id === 'vignette'));
  assert.ok(visualEffectCatalog.some(item => item.id === 'mosaic'));
  assert.ok(textStyleCatalog.some(item => item.id === 'vertical'));
  assert.ok(subtitleStyleCatalog.some(item => item.id === 'classic-outline'));
  assert.ok(subtitleAnimationCatalog.some(item => item.id === 'typewriter'));
});

test('advanced visual effects normalize bounded intensity and produce real preview styles', () => {
  const effect = normalizeVisualEffect({ preset: 'shake', intensity: 9, speed: 0 });
  assert.equal(effect.intensity, 1);
  assert.equal(effect.speed, .1);
  const style = visualEffectPreviewStyle(effect, 250);
  assert.match(style.transform, /translate/);
  assert.equal(normalizeVisualEffect({ preset: 'missing' }).preset, 'original');
});
