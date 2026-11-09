import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'ShuihuoProductionPage.jsx'), 'utf8');
const css = fs.readFileSync(path.join(here, 'shuihuo-production-v88.css'), 'utf8');
const projects = fs.readFileSync(path.join(here, 'shuihuo', 'ProjectsView.jsx'), 'utf8');
const studio = fs.readFileSync(path.join(here, 'shuihuo', 'StudioView.jsx'), 'utf8');

test('水货生产接入 V88 主题，且项目库容器使用页面间距令牌', () => {
  assert.match(source, /import \{ ConfigProvider, Modal, Spin, message \} from 'antd'/);
  assert.match(source, /theme = 'dark'/);
  assert.match(source, /createV88AntTheme\(theme\)/);
  assert.match(source, /<ConfigProvider theme=\{antTheme\}>/);
  assert.match(css, /\.shuihuo-project-library\s*\{[^}]*width:\s*100%[^}]*padding:\s*var\(--v88-space-page\)/s);
  assert.match(css, /\.shuihuo-project-library\s*\{[^}]*min-width:\s*0/s);
});

test('水货生产颜色令牌映射到 V88，暗色主题生效', () => {
  assert.match(css, /\[data-theme='dark'\] \.shuihuo-production[\s\S]*?--sh-accent:\s*var\(--v88-color-primary\)/);
  assert.match(css, /--sh-bg:\s*var\(--v88-color-bg\)/);
  assert.match(css, /--sh-danger:\s*var\(--v88-color-danger\)/);
  assert.match(css, /\.ant-btn-dangerous:not\(:disabled\)[\s\S]*?var\(--v88-color-danger\)/);
});

test('水货生产空态与动效统一，且尊重 prefers-reduced-motion', () => {
  assert.match(css, /\.shuihuo-production \.shuihuo-empty[\s\S]*?border-radius:\s*var\(--v88-radius-lg\)/);
  assert.match(css, /transition:[^;]*var\(--v88-motion-base\)[^;]*var\(--v88-ease\)/s);
  assert.match(css, /@media \(prefers-reduced-motion: reduce\)/);
});

test('水货生产只改视觉层，接口与状态读取路径未动', () => {
  // 本批没有改任何请求、状态字段或进度计算；这里只锁定原有加载与状态入口仍在。
  assert.match(source, /setProjects/);
  assert.match(source, /setView/);
  assert.match(projects, /projects/);
});
