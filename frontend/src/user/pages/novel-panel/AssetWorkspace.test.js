import test from 'node:test';
import assert from 'node:assert/strict';
import { buildAssetCards } from './assetWorkspaceModel.js';

test('builds an image preview card when an asset has a usable image url', () => {
  const [card] = buildAssetCards([{ type: 'character', id: 'c-1', name: '主角', image_url: '/img/main.png' }]);
  assert.deepEqual(card, {
    assetType: 'character',
    assetId: 'c-1',
    name: '主角',
    imageUrl: '/img/main.png',
    preview: { kind: 'image', src: '/img/main.png' },
    revision: 0,
    error: '',
    retryable: false
  });
});

test('uses a type placeholder when an asset has no image', () => {
  const [card] = buildAssetCards([{ asset_type: 'scene', asset_id: 's-1', display_name: '客厅' }]);
  assert.equal(card.preview.kind, 'placeholder');
  assert.equal(card.preview.label, '场景');
});

test('keeps a failed asset retryable without dropping its metadata', () => {
  const [card] = buildAssetCards([{ type: 'prop', id: 'p-1', name: '茶盏', error: '生成失败' }]);
  assert.equal(card.error, '生成失败');
  assert.equal(card.retryable, true);
});
