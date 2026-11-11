import assert from 'node:assert/strict';
import test from 'node:test';
import { createAICutCapabilities, filterAICutCapabilities, recordRecentAICutFeature, toggleFavoriteAICutFeature } from './aicutCapabilities.js';

test('does not expose a feature without a real handler', () => {
  const items = createAICutCapabilities({});
  assert.equal(items.some(item => item.id === 'visual-effects'), false);
  assert.equal(items.every(item => typeof item.action === 'function'), true);
});

test('exposes only explicitly wired categories and features', () => {
  const items = createAICutCapabilities({
    openAssets() {},
    uploadAsset() {},
    openVisualEffects() {},
  });
  assert.deepEqual(items.map(item => item.id), ['web-upload', 'project-assets', 'visual-effects']);
  assert.deepEqual([...new Set(items.map(item => item.category))], ['媒体', '特效']);
});

test('filters favorites, recent items and search text', () => {
  const items = createAICutCapabilities({ openAssets() {}, uploadAsset() {}, openVisualEffects() {} });
  assert.deepEqual(filterAICutCapabilities(items, { filter: '收藏', favorites: ['visual-effects'] }).map(item => item.id), ['visual-effects']);
  assert.deepEqual(filterAICutCapabilities(items, { filter: '最近使用', recent: ['project-assets', 'web-upload'] }).map(item => item.id), ['project-assets', 'web-upload']);
  assert.deepEqual(filterAICutCapabilities(items, { search: '画面' }).map(item => item.id), ['visual-effects']);
});

test('favorite toggles and recent list stays unique', () => {
  assert.deepEqual(toggleFavoriteAICutFeature(['web-upload'], 'web-upload'), []);
  assert.deepEqual(toggleFavoriteAICutFeature([], 'web-upload'), ['web-upload']);
  assert.deepEqual(recordRecentAICutFeature(['project-assets', 'web-upload'], 'web-upload'), ['web-upload', 'project-assets']);
});
