import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';

const css = fs.readFileSync(path.join(process.cwd(), 'src/user/pages/shuihuo-production.css'), 'utf8');

test('automation dialog portal follows the active Shuihuo theme for every surface', () => {
  assert.match(css, /\.batch-factory-automation-modal \.ant-modal-footer[^{]*\{[^}]*background:\s*var\(--sh-panel\)/s);
  assert.match(css, /\.batch-factory-automation-modal \.ant-modal-content[^{]*\{[^}]*background:\s*var\(--sh-panel\)/s);
  assert.match(css, /\.batch-factory-automation-section[^{]*\{[^}]*background:\s*var\(--sh-panel-strong\)/s);
});

test('batch factory rows and settings cards use shared Shuihuo text and surface tokens', () => {
  assert.match(css, /\.batch-factory-book-content strong\s*\{[^}]*color:\s*var\(--sh-text\)/s);
  assert.match(css, /\.batch-factory-book-config-region\s*\{[^}]*background:\s*var\(--sh-panel-strong\)/s);
  assert.match(css, /\.batch-factory-book-config-region small\s*\{[^}]*color:\s*var\(--sh-muted\)/s);
});

test('light theme variables are applied when the theme attribute is on either document root or body', () => {
  assert.match(css, /html\[data-theme='light'\] body\.shuihuo-theme-active[\s\S]*--sh-panel-strong:\s*#f8fafc/);
  assert.match(css, /body\[data-theme='light'\]\.shuihuo-theme-active[\s\S]*--sh-panel-strong:\s*#f8fafc/);
});

test('light workbench keeps headers, row numbers, and action controls readable', () => {
  assert.match(css, /body\.shuihuo-theme-active \.shuihuo-workbench-head[\s\S]*color:\s*var\(--sh-text\)/);
  assert.match(css, /body\.shuihuo-theme-active \.shuihuo-order-cell strong[\s\S]*color:\s*var\(--sh-text\)/);
  assert.match(css, /body\.shuihuo-theme-active \.batch-factory-action-button-grid \.ant-btn[\s\S]*background:\s*var\(--sh-panel\)/);
});
