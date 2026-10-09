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

test('account and login styles do not hide business text with transparent color', () => {
  for (const file of [
    'shared/styles/login-card.css',
    'shared/styles/account-center-visual-rebuild.css',
    'shared/styles/member-center.css',
    'shared/styles/team-governance.css',
    'shared/styles/team-collaboration.css'
  ]) {
    const css = fs.readFileSync(path.join(root, file), 'utf8');
    assert.doesNotMatch(css, /(?:color|-webkit-text-fill-color):\s*transparent/);
  }
});

test('login and account center styles use readable semantic text tokens', () => {
  const expectedTokens = {
    'shared/styles/login-card.css': ['primary', 'secondary', 'tertiary', 'danger'],
    'shared/styles/account-center-visual-rebuild.css': ['primary', 'secondary', 'tertiary', 'disabled', 'danger', 'warning', 'success'],
    'shared/styles/member-center.css': ['primary', 'secondary', 'tertiary', 'danger', 'success'],
    'shared/styles/team-governance.css': ['primary', 'secondary', 'disabled', 'danger', 'warning'],
    'shared/styles/team-collaboration.css': ['primary', 'secondary', 'tertiary', 'warning']
  };

  for (const [file, tokens] of Object.entries(expectedTokens)) {
    const css = fs.readFileSync(path.join(root, file), 'utf8');
    for (const token of tokens) {
      assert.match(css, new RegExp(`var\\(--text-${token}(?:[),])`), `${file} must use --text-${token}`);
    }
  }
});

test('disabled team rows use the disabled text token instead of low opacity', () => {
  const css = fs.readFileSync(path.join(root, 'shared/styles/team-governance.css'), 'utf8');
  assert.match(css, /\.ac-team-table-row\.is-disabled\s*\{[^}]*color:\s*var\(--text-disabled\)/s);
  assert.doesNotMatch(css, /\.ac-team-table-row\.is-disabled\s*\{[^}]*opacity:\s*0?\.[0-4]/s);
});
