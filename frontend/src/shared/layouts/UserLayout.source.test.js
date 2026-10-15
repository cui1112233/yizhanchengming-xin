import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { fileURLToPath } from 'node:url';

const source = fs.readFileSync(path.join(path.dirname(fileURLToPath(import.meta.url)), 'UserLayout.jsx'), 'utf8');
const stylesheet = fs.readFileSync(path.join(path.dirname(fileURLToPath(import.meta.url)), '..', 'styles', 'global.css'), 'utf8');

test('global header status expires instead of permanently retaining API polling failures', () => {
  assert.match(source, /expiresAt/);
  assert.match(source, /window\.setTimeout\(\(\) =>/);
  assert.match(source, /setGlobalStatus\(current => current\.expiresAt === expiresAt \? \{ text: '', tone: 'idle'/);
});

test('侧边栏不再显示历史和问题日志入口', () => {
  assert.doesNotMatch(source, /href: '\/history'/);
  assert.doesNotMatch(source, /href: '\/issues'/);
  assert.doesNotMatch(source, /label: '历史'/);
  assert.doesNotMatch(source, /label: '问题日志'/);
});

test('批量工厂工作区沿用水货生产主入口名称', () => {
  assert.match(source, /pathname === '\/batch-factory' \|\| pathname\.startsWith\('\/batch-factory\/'\).*return '水货生产'/s);
});

test('小说获取侧边栏进入带导航的 Go 工作台外壳', () => {
  assert.match(source, /href: '\/novel-fetch', icon: BookOpen, label: '小说获取'/);
  assert.match(source, /<Link key=\{item\.href\} href=\{item\.href\} className=/);
});

test('avatar enters profile through the client route without opening the menu first', () => {
  const avatarButton = source.slice(
    source.indexOf('className="legacy-sidebar-avatar"'),
    source.indexOf('</button>', source.indexOf('className="legacy-sidebar-avatar"'))
  );

  assert.match(avatarButton, /window\.history\.pushState\(\{\}, '', '\/profile'\)/);
  assert.match(avatarButton, /window\.dispatchEvent\(new PopStateEvent\('popstate'\)\)/);
  assert.doesNotMatch(avatarButton, /setAccountCenterOpen\(true\)/);
  assert.doesNotMatch(avatarButton, /window\.location\.assign\('\/profile'\)/);
});

test('routes account-center pages through the dedicated layout while the avatar records a business origin', () => {
  assert.match(source, /captureAccountCenterOrigin/);
  assert.match(source, /<AccountCenterLayout/);
  assert.doesNotMatch(source, /accountCenterOpen/);
});

test('keeps theme and logout controls compact inside the collapsed sidebar', () => {
  assert.match(source, /className="legacy-sidebar-logout"/);
  assert.match(source, /<LogOut/);
  assert.match(stylesheet, /\.legacy-theme-toggle\s*\{[^}]*border:\s*0;/s);
  assert.match(stylesheet, /\.legacy-sidebar\.collapsed \.legacy-sidebar-user\s*\{[^}]*flex-direction:\s*column;/s);
  assert.match(stylesheet, /\.legacy-sidebar\.collapsed \.legacy-sidebar-user \.legacy-sidebar-logout-label\s*\{[^}]*display:\s*none;/s);
});
