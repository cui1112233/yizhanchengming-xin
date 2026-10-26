import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

import {
  comicStyleSelection,
  comicStylePreviewReference,
  normalizeComicReference
} from './batchFactoryComicStyle.js';

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

test('comic style grid keeps every preview card at its natural 3:4 height', () => {
  const stylesheet = fs.readFileSync(new URL('../shuihuo-production.css', import.meta.url), 'utf8');
  assert.match(stylesheet, /\.batch-factory-comic-style-grid\s*\{[^}]*grid-auto-rows:\s*max-content/s);
  assert.match(stylesheet, /\.batch-factory-comic-style-image\s*\{[^}]*aspect-ratio:\s*3\s*\/\s*4/s);
});

test('comic style names stay readable on the white preview cards', () => {
  const stylesheet = fs.readFileSync(new URL('../shuihuo-production.css', import.meta.url), 'utf8');
  assert.match(stylesheet, /\.batch-factory-comic-style-copy b\s*\{[^}]*color:\s*#1f1f1f[^}]*font-size:\s*13px/s);
});
