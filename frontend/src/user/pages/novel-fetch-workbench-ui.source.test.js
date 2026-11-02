import test from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const source = readFileSync(new URL('./novel-fetch-workbench-ui.jsx', import.meta.url), 'utf8');

test('圆柱列表在滚动停止后以居中条目同步选中状态', () => {
  assert.match(source, /function handleScroll\(\)/);
  assert.match(source, /window\.setTimeout\(commitCenteredItem, 100\)/);
  assert.match(source, /onScroll=\{handleScroll\}/);
  assert.match(source, /onSelectRef\.current\?\.\(nearestId\)/);
});
