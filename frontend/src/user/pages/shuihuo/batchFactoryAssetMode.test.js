import test from 'node:test';
import assert from 'node:assert/strict';
import { assetsForProductionMode } from './batchFactoryAssetMode.js';

test('comic mode shows only the latest replaceable comic assets', () => {
  const assets = [
    { id: 'manual-old', source: 'manual', name: '豆包资产图测试' },
    { id: 'video', source: 'director', name: '视频人物' },
    { id: 'comic', source: 'comic', name: '我' }
  ];
  assert.deepEqual(assetsForProductionMode(assets, 'comic').map(asset => asset.id), ['comic']);
});

test('video mode excludes comic extraction rows', () => {
  const assets = [
    { id: 'manual', source: 'manual' },
    { id: 'director', source: 'director' },
    { id: 'comic', source: 'comic' }
  ];
  assert.deepEqual(assetsForProductionMode(assets, 'video').map(asset => asset.id), ['manual', 'director']);
});

test('legacy comic assets accidentally marked manual are recovered by their extraction preset', () => {
  const assets = [
    { id: 'legacy-comic', source: 'manual', extractionPresetId: 'vip-smart-character-derive' },
    { id: 'ordinary-manual', source: 'manual' }
  ];
  assert.deepEqual(assetsForProductionMode(assets, 'comic').map(asset => asset.id), ['legacy-comic']);
  assert.deepEqual(assetsForProductionMode(assets, 'video').map(asset => asset.id), ['ordinary-manual']);
});
