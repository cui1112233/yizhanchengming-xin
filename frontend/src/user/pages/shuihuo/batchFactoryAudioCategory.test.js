import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';

import * as audioCategory from './batchFactoryAudioCategory.js';

const { inferBatchN8AudioCategory, inferBookN8AudioCategory } = audioCategory;

const unifiedSettingsSource = fs.readFileSync(new URL('./BatchFactoryUnifiedSettingsModal.jsx', import.meta.url), 'utf8');

const categories = [
  { id: 14, name: '现代悬疑', status: 'ENABLE' },
  { id: 13, name: '男频都市', status: 'ENABLE' },
  { id: 12, name: '古风通用', status: 'ENABLE' },
  { id: 11, name: '古风虐文', status: 'ENABLE' },
  { id: 10, name: '古风甜文', status: 'ENABLE' },
  { id: 9, name: '现代通用', status: 'ENABLE' },
  { id: 8, name: '现代虐文', status: 'ENABLE' }
];

test('maps every known novel style to the agreed N8 audio category', () => {
  const cases = [
    ['现代悬疑', '女频', 14], ['男频都市', '男频', 13],
    ['古风通用', '女频', 12], ['玄幻', '男频', 12], ['历史', '男频', 12],
    ['古风虐文', '女频', 11], ['古风甜文', '女频', 10],
    ['现代虐文', '女频', 8], ['年代虐文', '女频', 8], ['家庭伤感', '女频', 8],
    ['现代甜文', '女频', 9], ['现代通用', '女频', 9], ['年代甜文', '女频', 9],
    ['年代通用', '女频', 9], ['现代女主', '女频', 9], ['家庭奇葩', '女频', 9],
    ['职场打脸', '男频', 13], ['职场打脸', '女频', 9]
  ];
  for (const [style, gender, categoryId] of cases) {
    const result = inferBookN8AudioCategory({ sourceMetadata: { classifyStatus: 'classified', gender, style } }, categories);
    assert.equal(result.category?.id, categoryId, `${style}/${gender}`);
    assert.equal(result.reason, 'matched', `${style}/${gender}`);
  }
});

test('falls back by gender only when style has no explicit mapping', () => {
  assert.equal(inferBookN8AudioCategory({ sourceMetadata: { classifyStatus: 'classified', gender: '男频', style: '未知' } }, categories).category?.id, 13);
  assert.equal(inferBookN8AudioCategory({ sourceMetadata: { classifyStatus: 'classified', gender: '女频', style: '未知' } }, categories).category?.id, 9);
  assert.equal(inferBookN8AudioCategory({ sourceMetadata: { classifyStatus: 'classified', gender: '', style: '未知' } }, categories).reason, 'unclassified');
});

test('keeps batch inference compatible without treating mixed books as one category', () => {
  const result = inferBatchN8AudioCategory([
    { sourceMetadata: { classifyStatus: 'classified', gender: '女频', style: '现代悬疑' } },
    { sourceMetadata: { classifyStatus: 'classified', gender: '男频', style: '男频都市' } }
  ], categories);
  assert.equal(result.reason, 'mixed');
  assert.equal(result.category, null);
});

test('batch and automation preset configs always clear manual audio selection', () => {
  assert.equal(typeof audioCategory.normalizeBatchAutoAudioConfig, 'function');
  const original = {
    productionMode: 'comic',
    comicPublishSettings: {
      organizationId: '1',
      audioCategorySelectionMode: 'manual',
      audioCategoryId: 14,
      audioCategoryName: '现代悬疑',
      audioCandidates: [{ id: 199, name: '旧音频' }],
      specifiedAudioId: 199,
      specifiedAudioName: '旧音频'
    }
  };

  const normalized = audioCategory.normalizeBatchAutoAudioConfig(original);

  assert.equal(normalized.comicPublishSettings.audioCategorySelectionMode, 'auto');
  assert.equal(normalized.comicPublishSettings.audioCategoryId, null);
  assert.equal(normalized.comicPublishSettings.audioCategoryName, '');
  assert.deepEqual(normalized.comicPublishSettings.audioCandidates, []);
  assert.equal(normalized.comicPublishSettings.specifiedAudioId, '');
  assert.equal(normalized.comicPublishSettings.specifiedAudioName, '');
  assert.equal(original.comicPublishSettings.audioCategorySelectionMode, 'manual', 'must not mutate the saved source preset');
});

test('only single-book settings expose manual audio matching', () => {
  assert.equal(typeof audioCategory.audioMatchingModeOptions, 'function');
  assert.deepEqual(audioCategory.audioMatchingModeOptions(false), [
    { value: 'auto', label: '按风格自动匹配' }
  ]);
  assert.deepEqual(audioCategory.audioMatchingModeOptions(true), [
    { value: 'auto', label: '按风格自动匹配' },
    { value: 'manual', label: '手动指定' }
  ]);
});

test('comic publish defaults to per-book automatic audio instead of one batch category', () => {
  assert.match(unifiedSettingsSource, /发布时按每本书的男女频和风格自动匹配并随机选择音频/);
  assert.match(unifiedSettingsSource, /const manualAudioMode = singleBook && comicPublishSettings\.audioCategorySelectionMode === 'manual'/);
  assert.match(unifiedSettingsSource, /options=\{audioMatchingModeOptions\(singleBook\)\}/);
  assert.match(unifiedSettingsSource, /createAutomationPreset\(\{ name, config: normalizeBatchAutoAudioConfig\(value\) \}\)/);
  assert.doesNotMatch(unifiedSettingsSource, /inferBatchN8AudioCategory/);
  assert.doesNotMatch(unifiedSettingsSource, /批次书籍分类不一致，需手动选择音频分类/);
});

test('manual category and audio changes explicitly lock manual mode', () => {
  const manualAssignments = unifiedSettingsSource.match(/audioCategorySelectionMode: 'manual'/g) || [];
  assert.ok(manualAssignments.length >= 3, `manual assignments=${manualAssignments.length}`);
  assert.match(unifiedSettingsSource, /specifiedAudioId: selected\?\.id \|\| '', specifiedAudioName: selected\?\.name \|\| '', audioCategorySelectionMode: 'manual'/);
});
