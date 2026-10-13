import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'HomePage.jsx'), 'utf8');
const userLayout = fs.readFileSync(path.join(here, '../../shared/layouts/UserLayout.jsx'), 'utf8');
const css = fs.readFileSync(path.join(here, '../../shared/styles/global.css'), 'utf8');

test('首页不再显示历史和问题日志入口', () => {
  assert.doesNotMatch(source, /href: '\/history'/);
  assert.doesNotMatch(source, /href: '\/issues'/);
  assert.doesNotMatch(source, /查看全部历史记录/);
  assert.doesNotMatch(source, /最近创作项目/);
});

test('首页导航使用黑白主题液态玻璃层次', () => {
  assert.match(css, /\.home-hero-nav::before\s*\{[^}]*content:\s*['"][^'"]*['"]/s);
  assert.match(css, /\.home-hero-nav::after\s*\{[^}]*content:\s*['"][^'"]*['"]/s);
  assert.match(css, /\.home-hero-nav\s*\{[^}]*backdrop-filter:\s*blur\(/s);
  assert.match(css, /\.home-hero-nav\s*\{[^}]*background:\s*(?:linear-gradient\([^;]*rgba\(|rgba\()/s);
  assert.match(css, /\.home-hero-nav::before\s*\{[^}]*opacity:\s*\.56/s);
  assert.match(css, /\[data-theme='light'\] \.home-hero-nav\s*\{[^}]*background:\s*rgba\(/s);
  assert.match(css, /\[data-theme='light'\] \.home-hero-nav\s*\{[^}]*backdrop-filter:\s*blur\(/s);
  assert.match(css, /\.home-hero-links a\s*\{[^}]*background:\s*transparent/s);
  assert.match(css, /\.home-hero-links a\s*\{[^}]*box-shadow:\s*none/s);
  assert.match(css, /\[data-theme='light'\] \.home-hero-nav\s*\{/);
  assert.match(css, /\[data-theme='light'\] \.home-hero-links a:hover\s*\{/);
  assert.match(css, /\[data-theme='light'\] \.home-hero-links a\s*\{[^}]*color:\s*#182333/s);
  assert.match(css, /\[data-theme='light'\] \.home-hero-brand \.brand-logo-black\s*\{[^}]*display:\s*block/s);
  assert.match(css, /\[data-theme='light'\] \.home-hero-brand \.brand-logo-white\s*\{[^}]*display:\s*none/s);
  assert.match(css, /\.home-hero-nav\s*\{[^}]*box-shadow:\s*inset 4px 4px 10px/s);
  assert.match(css, /\[data-theme='light'\] \.home-hero-nav\s*\{[^}]*box-shadow:\s*inset 4px 4px 10px/s);
});

test('浅色主题下左上角品牌标题保持可见', () => {
  assert.match(css, /\[data-theme='light'\] \.legacy-brand-title\.shine\s*\{[\s\S]*color:\s*var\(--legacy-text\)/);
  assert.match(css, /\[data-theme='light'\] \.legacy-brand-title\.shine\s*\{[\s\S]*background-image:\s*none/);
});

test('用户端品牌文字统一使用扫光样式', () => {
  assert.match(source, /<span className="shine">一战晟铭<\/span>/);
  assert.match(userLayout, /<span className="legacy-brand-title shine">一战晟铭<\/span>/);
  assert.match(css, /\.shine\s*\{[^}]*background-image:\s*linear-gradient\(/s);
  assert.match(css, /\.shine\s*\{[^}]*-webkit-background-clip:\s*text/s);
  assert.match(css, /@keyframes shine-sweep\s*\{/);
});
