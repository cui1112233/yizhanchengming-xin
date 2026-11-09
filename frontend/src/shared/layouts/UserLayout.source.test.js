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
  assert.match(source, /pathname === '\/batch-factory' \|\| pathname\.startsWith\('\/batch-factory\/'\) \|\| pathname === '\/shuihuo-production'.*return '水货生产'/s);
});

test('小说获取侧边栏进入带导航的 Go 工作台外壳', () => {
  assert.match(source, /href: '\/novel-fetch', icon: BookOpen, label: '小说获取', reload: true/);
  assert.match(source, /<Link key=\{item\.href\} href=\{item\.href\} reload=\{item\.reload\} className=/);
});

test('账号信息异步更新不会卸载正在使用的业务工作台', () => {
  // 账号资料是异步读取的；不能把用户名当作主内容容器的 React key，
  // 否则资料回来时会连同小说获取 iframe 一起卸载并重新加载，表现为闪烁。
  assert.doesNotMatch(source, /<Fragment key=\{accountSessionKey\}>/);
  assert.match(source, /<Fragment>\s*<main className="legacy-main">/s);
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

test('keeps the shared multi-theme picker and logout controls compact inside the collapsed sidebar', () => {
  assert.match(source, /className="legacy-sidebar-logout"/);
  assert.match(source, /<LogOut/);
  assert.match(source, /<ThemePicker value=\{theme\} onChange=\{changeTheme\} compact \/>/);
  assert.match(source, /const nextThemeId = persistCachedTheme\(themeId\)/);
  assert.match(source, /updateMemberProfile\(\{ themeId: nextThemeId \}\)/);
  assert.match(source, /主题已在本机切换，但未同步到账号/);
  assert.match(stylesheet, /\.legacy-sidebar\.collapsed \.legacy-sidebar-user\s*\{[^}]*flex-direction:\s*column;/s);
  assert.match(stylesheet, /\.legacy-sidebar\.collapsed \.legacy-sidebar-user \.legacy-sidebar-logout-label\s*\{[^}]*display:\s*none;/s);
});

test('logged-in account theme overrides cache and system mode resolves before rendering', () => {
  assert.match(source, /selectThemePreference\(\{ loggedIn: true, accountThemeId: account\.themeId, cachedThemeId: theme \}\)/);
  assert.match(source, /resolveThemeId\(theme, systemMode\)/);
  assert.match(source, /document\.documentElement\.dataset\.theme = resolvedTheme/);
});

test('窗口背景不能把固定桌宠变成占据网页宽度的 flex 子项', () => {
  assert.doesNotMatch(stylesheet, /\.legacy-shell\.has-window-background\s*>\s*\*\s*\{/);
  assert.match(stylesheet, /\.legacy-shell\.has-window-background\s*>\s*\.legacy-sidebar,\s*\.legacy-shell\.has-window-background\s*>\s*\.legacy-main,\s*\.home-shell\.has-window-background\s*>\s*\.home-page\s*\{[^}]*position:\s*relative;[^}]*z-index:\s*1;/);
});
