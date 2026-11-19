import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const homeSource = fs.readFileSync(path.join(here, 'HomePage.jsx'), 'utf8');
const layoutSource = fs.readFileSync(
  path.join(here, '..', '..', 'shared', 'layouts', 'UserLayout.jsx'),
  'utf8'
);

function extractNavHrefs(source) {
  const block = source.match(/const navItems = \[([\s\S]*?)\];/);
  assert.ok(block, 'navItems 定义缺失');
  return [...block[1].matchAll(/href: '([^']+)'/g)].map(match => match[1]);
}

// 首页顶部胶囊导航必须与工作台侧栏提供同一份入口集合（侧栏额外含“首页”，
// 由 logo 承担），避免出现“首页导航缺项、内页有侧栏”的不一致。
test('首页导航入口与工作台侧栏完全一致', () => {
  const homeHrefs = extractNavHrefs(homeSource);
  const sidebarHrefs = extractNavHrefs(layoutSource).filter(href => href !== '/');
  assert.deepEqual(homeHrefs, sidebarHrefs);
  // 批量测书在侧栏强制整页加载，首页胶囊必须保留同一语义。
  assert.match(homeSource, /href: '\/novel-fetch', label: '批量测书', reload: true/);
});
