import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

import {
  comicStyleCoverFlowStyle,
  comicStyleIndexFromDrag,
  comicStyleRelativeOffset,
  comicStyleSelection,
  comicStylePreviewReference,
  normalizeComicReference,
  wrapComicStyleIndex
} from './batchFactoryComicStyle.js';

test('cover flow wraps at both ends and uses the shortest cyclic offset', () => {
  assert.equal(wrapComicStyleIndex(-1, 5), 4);
  assert.equal(wrapComicStyleIndex(5, 5), 0);
  assert.equal(comicStyleRelativeOffset(4, 0, 5), -1);
  assert.equal(comicStyleRelativeOffset(0, 4, 5), 1);
});

test('cover flow drag changes one style only after crossing the threshold', () => {
  assert.equal(comicStyleIndexFromDrag(2, -50, 5), 3);
  assert.equal(comicStyleIndexFromDrag(2, 50, 5), 1);
  assert.equal(comicStyleIndexFromDrag(2, 20, 5), 2);
});

test('cover flow gives the centered card priority and mirrors both sides', () => {
  assert.deepEqual(comicStyleCoverFlowStyle(0), {
    '--comic-style-offset': 0,
    '--comic-style-distance': 0,
    '--comic-style-brightness': 1,
    '--comic-style-rotation': '0deg',
    '--comic-style-scale': 1.08,
    '--comic-style-translate-x': '0px',
    '--comic-style-translate-z': '90px',
    opacity: 1,
    pointerEvents: 'auto',
    zIndex: 20
  });
  assert.equal(comicStyleCoverFlowStyle(-2)['--comic-style-rotation'], '42deg');
  assert.equal(comicStyleCoverFlowStyle(-2)['--comic-style-translate-x'], '-264px');
  assert.equal(comicStyleCoverFlowStyle(2)['--comic-style-rotation'], '-42deg');
  assert.equal(comicStyleCoverFlowStyle(2)['--comic-style-brightness'], 0.78);
  assert.equal(comicStyleCoverFlowStyle(2)['--comic-style-scale'], 0.9);
});

test('cover flow hides distant cards so they cannot block the visible controls', () => {
  assert.equal(comicStyleCoverFlowStyle(5).pointerEvents, 'none');
  assert.equal(comicStyleCoverFlowStyle(5).opacity, 0);
});

test('comic style selection stores metadata without treating its preview as an AI reference', () => {
  const selected = comicStyleSelection({
    id: 'batch-comic-style-01',
    name: '顶级韩漫风格',
    slot: 'batch.comic-style',
    version: 2,
    previewUrl: '/api/presets/comic-style/batch-comic-style-01/preview'
  });

  assert.deepEqual(selected, {
    presetId: 'batch-comic-style-01',
    presetName: '顶级韩漫风格',
    presetSlot: 'batch.comic-style',
    presetVersion: 2
  });
  assert.equal(Object.hasOwn(selected, 'previewUrl'), false);
});

test('style preview becomes a global reference only after the explicit command', () => {
  assert.deepEqual(comicStylePreviewReference({
    id: 'batch-comic-style-01',
    previewUrl: '/api/presets/comic-style/batch-comic-style-01/preview'
  }), {
    assetId: 'preset:batch-comic-style-01',
    url: '/api/local-executor/v1/batch-comic-style-previews/batch-comic-style-01',
    previewUrl: '/api/presets/comic-style/batch-comic-style-01/preview',
    enabled: true,
    applyToAssets: true,
    applyToPages: true
  });
});

test('uploaded global reference keeps independent asset and page scope switches', () => {
  assert.deepEqual(normalizeComicReference({ assetId: 'lea_reference', url: '/executor/ref', previewUrl: '/preview/ref', enabled: true, applyToAssets: false }), {
    assetId: 'lea_reference',
    url: '/executor/ref',
    previewUrl: '/preview/ref',
    enabled: true,
    applyToAssets: false,
    applyToPages: true
  });
});

test('comic style cover flow keeps every preview at its natural 3:4 height', () => {
  const stylesheet = fs.readFileSync(new URL('../shuihuo-production.css', import.meta.url), 'utf8');
  assert.match(stylesheet, /\.batch-factory-comic-style-image\s*\{[^}]*aspect-ratio:\s*3\s*\/\s*4/s);
});

test('comic style names stay readable on the white preview cards', () => {
  const stylesheet = fs.readFileSync(new URL('../shuihuo-production.css', import.meta.url), 'utf8');
  assert.match(stylesheet, /\.batch-factory-comic-style-copy b\s*\{[^}]*color:\s*#1f1f1f[^}]*font-size:\s*13px/s);
});
