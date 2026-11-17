import assert from 'node:assert/strict';
import { test } from 'node:test';
import { normalizeWindowBackgroundAppearance, readWindowBackgroundAppearance, saveWindowBackgroundAppearance } from './window-background-appearance.js';

test('wallpaper defaults match approved 42px blur and 50% brightness', () => {
  assert.deepEqual(normalizeWindowBackgroundAppearance(), { blur: 42, brightness: 50 });
  assert.deepEqual(normalizeWindowBackgroundAppearance({ blur: null, brightness: '' }), { blur: 42, brightness: 50 });
});
test('wallpaper adjustments are bounded and preserve valid zero blur', () => {
  assert.deepEqual(normalizeWindowBackgroundAppearance({ blur: 0, brightness: 50 }), { blur: 0, brightness: 50 });
  assert.deepEqual(normalizeWindowBackgroundAppearance({ blur: 100, brightness: -10 }), { blur: 42, brightness: 18 });
  assert.deepEqual(normalizeWindowBackgroundAppearance({ blur: 'bad', brightness: Infinity }), { blur: 42, brightness: 50 });
});
test('saved adjustments restore on reload and apply the same visual values', () => {
  const previousWindow = globalThis.window;
  const previousDocument = globalThis.document;
  const saved = new Map(), css = new Map();
  globalThis.window = { localStorage: { getItem: key => saved.get(key) ?? null, setItem: (key, value) => saved.set(key, value) } };
  globalThis.document = { documentElement: { style: { setProperty: (key, value) => css.set(key, value) } } };
  try {
    assert.deepEqual(readWindowBackgroundAppearance(), { blur: 42, brightness: 50 });
    saved.set('v88-window-bg-appearance', JSON.stringify({ blur: 24, brightness: 85 }));
    assert.deepEqual(readWindowBackgroundAppearance(), { blur: 24, brightness: 50 });
    saveWindowBackgroundAppearance({ blur: 30, brightness: 60 });
    assert.deepEqual(readWindowBackgroundAppearance(), { blur: 30, brightness: 60 });
    assert.equal(css.get('--window-video-blur'), '30px');
    assert.equal(css.get('--window-video-brightness'), '0.6');
    saved.set('v88-window-bg-appearance-v2', 'bad json');
    assert.deepEqual(readWindowBackgroundAppearance(), { blur: 42, brightness: 50 });
  } finally {
    if (previousWindow === undefined) delete globalThis.window; else globalThis.window = previousWindow;
    if (previousDocument === undefined) delete globalThis.document; else globalThis.document = previousDocument;
  }
});
