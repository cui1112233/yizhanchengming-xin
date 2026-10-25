import assert from 'node:assert/strict';
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
