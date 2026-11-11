import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'HomePage.jsx'), 'utf8');
const css = fs.readFileSync(path.join(here, 'home-v88.css'), 'utf8');

test('首页接入 V88 主题，antd 按钮不再单独引入', () => {
  assert.match(source, /import \{ ConfigProvider \} from 'antd'/);
  assert.doesNotMatch(source, /import \{ Button \} from 'antd'/);
  assert.match(source, /export function HomePage\(\{ isLoggedIn, onOpenLogin, theme = 'dark' \}\)/);
  assert.match(source, /<ConfigProvider theme=\{createV88AntTheme\(theme\)\}>/);
  assert.match(source, /<V88Button variant="secondary"[\s\S]{0,240}?>打开设置<\/V88Button>/);
});

test('首页容器横向留白使用 V88 页面间距令牌', () => {
  assert.match(css, /\.home-page\s*\{[^}]*--app-page-gutter:\s*var\(--v88-space-page\)/s);
  assert.match(css, /\.home-page\s*\{[^}]*min-width:\s*0/s);
});

test('首页卡片与联系区接入 V88 圆角、表面色与动效', () => {
  assert.match(css, /\.quick-action-card\s*\{[^}]*border-radius:\s*var\(--v88-radius-lg\)/s);
  assert.match(css, /transition:[^;]*var\(--v88-motion-base\)[^;]*var\(--v88-ease\)/s);
  assert.match(css, /\.home-contact-section\s*\{[^}]*background:\s*var\(--v88-color-surface\)/s);
  assert.match(css, /@media \(prefers-reduced-motion: reduce\)/);
});

test('首页原有入口一个不少', () => {
  ['私人定制', '水货生产', 'AI 智能 Agent', '声音配音工坊', '小说获取', 'CM 画布', '配音', '设置', '开始生成', '进入配音', '联系', '打开设置'].forEach(label => {
    assert.ok(source.includes(label), `缺少入口：${label}`);
  });
  assert.match(source, /<GradientButton as="a" href="#contact">联系<\/GradientButton>/);
  assert.match(source, /<GradientButton onClick=\{\(\) => openProtectedPage\('\/script'\)\}>开始生成<\/GradientButton>/);
});
