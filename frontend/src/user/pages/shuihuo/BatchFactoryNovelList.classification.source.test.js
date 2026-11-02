import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';

const source = readFileSync(new URL('./BatchFactoryNovelList.jsx', import.meta.url), 'utf8')
  + '\n' + readFileSync(new URL('./batchFactoryNovelListHelpers.js', import.meta.url), 'utf8');

test('分类永远自动兜底，不再要求人工确认', () => {
  assert.match(source, /pending_confirmation/);
  assert.match(source, /系统已自动确定男女频和风格，可在单书配置中修改/);
  assert.doesNotMatch(source, /请人工确认后继续/);
  assert.doesNotMatch(source, /请人工确认男女频和风格/);
  assert.doesNotMatch(source, /待人工确认/);
});

test('分类未完成时显示自动重试而不是等待人工', () => {
  assert.match(source, /正在自动重试分类/);
  assert.doesNotMatch(source, /等待确认/);
});
