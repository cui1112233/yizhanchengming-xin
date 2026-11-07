import assert from 'node:assert/strict';
import test from 'node:test';

import { inferBatchN8AudioCategory, inferBookN8AudioCategory } from './batchFactoryAudioCategory.js';

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
