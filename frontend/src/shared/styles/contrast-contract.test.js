import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(dirname, '..', '..');
const globalCSS = fs.readFileSync(path.join(root, 'shared/styles/global.css'), 'utf8');
const darkTheme = globalCSS.match(/:root\s*\{([\s\S]*?)\n\}/)?.[1] ?? '';
const lightTheme = globalCSS.match(/\[data-theme='light'\]\s*\{([\s\S]*?)\n\}/)?.[1] ?? '';

const semanticVariables = [
  '--text-primary', '--text-secondary', '--text-tertiary', '--text-disabled',
  '--text-link', '--text-on-accent', '--text-danger', '--text-warning',
  '--text-success', '--surface-base', '--surface-panel', '--surface-input',
  '--border-default', '--border-strong'
];

for (const name of semanticVariables) {
  test(`dark theme defines ${name}`, () => {
    assert.match(darkTheme, new RegExp(`${name}\\s*:`));
  });

  test(`light theme defines ${name}`, () => {
    assert.match(lightTheme, new RegExp(`${name}\\s*:`));
  });
}

test('legacy and admin colors map to semantic variables', () => {
  assert.match(globalCSS, /--legacy-bg:\s*var\(--surface-base\)/);
  assert.match(globalCSS, /--legacy-input:\s*var\(--surface-input\)/);
  assert.match(globalCSS, /--legacy-text:\s*var\(--text-primary\)/);
  assert.match(globalCSS, /--legacy-muted:\s*var\(--text-secondary\)/);
  assert.match(globalCSS, /--legacy-border:\s*var\(--border-default\)/);
  assert.match(globalCSS, /--admin-bg:\s*var\(--surface-base\)/);
  assert.match(globalCSS, /--admin-input:\s*var\(--surface-input\)/);
  assert.match(globalCSS, /--admin-text:\s*var\(--text-primary\)/);
  assert.match(globalCSS, /--admin-border:\s*var\(--border-default\)/);
});

test('account center does not hide its page title with transparent text', () => {
  const css = fs.readFileSync(path.join(root, 'shared/styles/account-center-visual-rebuild.css'), 'utf8');
  assert.doesNotMatch(css, /\.account-center-workspace \.legacy-page-title\s*\{[^}]*color:\s*transparent/s);
  assert.match(css, /\.account-center-workspace \.legacy-page-title\s*\{[^}]*color:\s*var\(--text-secondary\)/s);
});

test('account center preserves the global text secondary semantic variable', () => {
  const css = fs.readFileSync(path.join(root, 'shared/styles/account-center-visual-rebuild.css'), 'utf8');
  assert.doesNotMatch(css, /\.account-center-workspace\s*\{[^}]*--text-secondary\s*:/s);
});
