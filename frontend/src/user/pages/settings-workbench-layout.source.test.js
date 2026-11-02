import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const source = fs.readFileSync(path.join(here, 'SettingsPage.jsx'), 'utf8');
const css = fs.readFileSync(path.join(here, '../../shared/styles/global.css'), 'utf8');

test('工作台设置为上排配置和执行器操作提供对齐容器', () => {
  assert.match(source, /className="settings-top-grid"/);
  assert.match(source, /className="settings-section-footer settings-workspace-actions"/);
  assert.match(source, /className="settings-section-footer settings-storage-actions"/);
  assert.match(source, /className="settings-executor-actions settings-executor-actions-doubao"/);
  assert.match(source, /className="settings-executor-actions settings-executor-actions-giant"/);
  assert.match(source, /className="settings-executor-action-row settings-executor-action-row-top"/);
  assert.match(source, /className="settings-executor-action-row settings-executor-action-row-downloads"/);
});

test('工作台设置在浅色主题下使用浅色卡片和可读文字', () => {
  assert.match(css, /\[data-theme='light'\] \.settings-section\s*\{[^}]*background:\s*var\(--surface-panel\)/s);
  assert.match(css, /\[data-theme='light'\] \.settings-section h2\s*\{[^}]*color:\s*var\(--text-primary\)/s);
  assert.match(css, /\[data-theme='light'\] \.settings-page \.ant-btn\s*\{[^}]*background:\s*var\(--surface-panel\)/s);
});

test('工作台设置样式包含统一对齐、按钮层级和窄屏规则', () => {
  assert.match(css, /\.settings-top-grid\s*\{/);
  assert.match(css, /\.settings-section-footer\s*\{/);
  assert.match(css, /\.settings-page \.ant-btn:focus-visible\s*\{/);
  assert.match(css, /\.settings-executor-actions-doubao\s*\{/);
  assert.match(css, /\.settings-executor-actions-giant\s*\{/);
  assert.match(css, /\.settings-executor-action-row \.ant-btn\s*\{[^}]*width:\s*184px/);
  assert.match(css, /\.settings-page \.ant-switch-checked\s*\{/);
});

test('窄屏下载按钮逐行显示，长标题不会撑出设置卡片', () => {
  assert.match(css, /\.settings-executor-action-row-downloads\s*\{[^}]*grid-template-columns:\s*repeat\(auto-fit,\s*minmax\(200px,\s*1fr\)\)/);
  assert.match(css, /\.settings-executor-action-row-downloads \.ant-btn\s*\{[^}]*width:\s*100%/);
});
