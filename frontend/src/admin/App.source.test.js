import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const appSource = fs.readFileSync(path.join(here, 'App.jsx'), 'utf8');
const layoutSource = fs.readFileSync(path.join(here, '../shared/layouts/AdminLayout.jsx'), 'utf8');

test('管理后台提供账号与授权页面，不再把入口重定向到个人中心', () => {
  assert.match(appSource, /AccountGovernancePage/);
  assert.match(appSource, /pathname === '\/admin\/accounts'/);
  assert.doesNotMatch(appSource, /window\.location\.replace\('\/accounts'\)/);
  assert.match(layoutSource, /href: '\/admin\/accounts'.*label: '账号与授权'/s);
});
