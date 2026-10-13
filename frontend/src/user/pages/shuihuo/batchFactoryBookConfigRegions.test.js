import assert from 'node:assert/strict';
import test from 'node:test';
import { BOOK_CONFIG_REGIONS, bookAssetSummary, bookConfigRegionStatus } from './batchFactoryBookConfigRegions.js';

test('keeps single-book publishing inside engine configuration instead of a special region', () => {
  assert.deepEqual(BOOK_CONFIG_REGIONS.map(item => item.key), ['engine', 'assets', 'constraints', 'video', 'visual']);
});

test('reports current engine and video overrides without the retired fixed opening', () => {
  const book = { settingsState: { patch: { videoModelId: 'video-a', fixedSingleVideo: true, aiPromptConfig: { constraints: { enabled: false }, video: { presetId: 'batch-video-meta' } } } } };
  assert.equal(bookConfigRegionStatus(book, 'engine').tone, 'overridden');
  assert.equal(bookConfigRegionStatus(book, 'video').tone, 'overridden');
  assert.equal(bookConfigRegionStatus(book, 'constraints').tone, 'disabled');
  assert.equal(bookConfigRegionStatus(book, 'assets').tone, 'inherited');
});

test('ignores a legacy fixed-opening field when reporting current engine overrides', () => {
  const book = { settingsState: { patch: { fixedSingleVideo: true } } };
  assert.equal(bookConfigRegionStatus(book, 'engine').tone, 'inherited');
  assert.equal(bookConfigRegionStatus(book, 'video').tone, 'inherited');
});

test('summarizes real current-book assets instead of inventing an asset state', () => {
  assert.equal(bookAssetSummary({ assets: { characters: [{}], scenes: [{}, {}], props: [] } }), '人物 1 · 场景 2 · 道具 0');
});

test('reports a nested publish mapping as an engine override', () => {
  assert.deepEqual(
    bookConfigRegionStatus({ settingsState: { patch: { publishSettings: { materialReuse: true } } } }, 'engine'),
    { label: '已单书覆盖', tone: 'overridden' }
  );
});

test('recognizes independent book media overrides as engine settings', () => {
  assert.equal(bookConfigRegionStatus({ settingsState: { patch: { imageAspectRatio: '1:1' } } }, 'engine').tone, 'overridden');
  assert.equal(bookConfigRegionStatus({ settingsState: { patch: { videoAspectRatio: '16:9', videoResolution: '1080p' } } }, 'engine').tone, 'overridden');
});

test('shows the actual enabled single-book constraint count in the card summary', () => {
  const book = {
    settingsState: {
      patch: {
        aiPromptConfig: {
          constraints: {
            enabled: true,
            selections: [
              { constraintCategory: 'quality', presetName: '画质约束' },
              { constraintCategory: 'negative', presetName: '负面提示词' }
            ]
          }
        }
      }
    }
  };
  assert.deepEqual(bookConfigRegionStatus(book, 'constraints'), { label: '2 条约束已启用', tone: 'overridden' });
});
