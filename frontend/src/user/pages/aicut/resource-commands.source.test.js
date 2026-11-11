import assert from 'node:assert/strict';
import fs from 'node:fs';
import test from 'node:test';

const page = fs.readFileSync(new URL('../AICutPage.jsx', import.meta.url), 'utf8');

test('目录资源应用使用声明参数而不是写死特效和转场参数', () => {
  assert.match(page, /normalizeCatalogParameters\(resource\.id, suppliedParameters\)/);
  assert.match(page, /transitionMs:\s*Number\(parameters\.durationMs\s*\?\?\s*500\)/);
  assert.match(page, /parameters,\s*\n\s*},/);
});

test('时间线接收资源库拖拽并根据资源类型执行真实命令', () => {
  assert.match(page, /application\/x-aicut-resource/);
  assert.match(page, /dropCatalogResourceOnTrack/);
  assert.match(page, /resource\.category === 'effect'/);
  assert.match(page, /resource\.category === 'transition'/);
});

test('旧转场入口与统一资源目录使用同一组完整转场', () => {
  assert.match(page, /catalogForTab\('transition'\)/);
  assert.doesNotMatch(page, /if \(!transitionCatalog\.some\(\(item\) => item\.id === kind\)\)/);
});
