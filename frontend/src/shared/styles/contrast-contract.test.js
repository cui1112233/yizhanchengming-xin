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
    assert.doesNotMatch(css, /(?:^|[;{}])\s*(?:color|-webkit-text-fill-color):\s*transparent/);
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

test('creative-workspace style files use global readable text tokens', () => {
  for (const file of [
    'shared/styles/agent-workspace.css',
    'user/pages/novel-fetch.css',
    'user/pages/giant-material-test.css',
    'user/pages/GiantMaterialStatusCard.css'
  ]) {
    const css = fs.readFileSync(path.join(root, file), 'utf8');
    assert.match(css, /var\(--text-(?:primary|secondary|tertiary|disabled|danger|warning|success)\)/);
  }
});

function cssBlock(css, selector) {
  // 选择器分组（A, B, C { ... }）与同名规则拆块都是主题收敛后的常态：
  // 按选择器精确匹配，并把命中规则的声明合并后返回。
  const wanted = selector.split(',').map(item => item.trim()).filter(Boolean);
  const bodies = [];
  const source = css.replace(/\/\*[\s\S]*?\*\//g, '');
  for (const match of source.matchAll(/([^{}]+)\{([^}]*)\}/gs)) {
    const selectors = match[1].split(',').map(item => item.trim());
    if (wanted.every(item => selectors.includes(item))) bodies.push(match[2]);
  }
  assert.ok(bodies.length, `missing CSS block for ${selector}`);
  return bodies.join('\n');
}

test('agent task drawer surface follows the active theme', () => {
  const css = fs.readFileSync(path.join(root, 'shared/styles/agent-workspace.css'), 'utf8');
  assert.match(cssBlock(css, '.agent-page .agent-task-drawer'), /background:\s*var\(--surface-panel\)/);
});

function assertThemedSurface(block, selector) {
  assert.match(block, /background:\s*var\(--surface-(?:base|panel|input)\)/, `${selector} must use a theme surface`);
  assert.doesNotMatch(block, /background:\s*#(?:0[0-9a-f]|1[0-9a-f]|2[0-9a-f]|3[0-9a-f]|4[0-9a-f]|5[0-9a-f]|6[0-9a-f])/i, `${selector} must not restore a fixed dark surface`);
}

function assertThemedBorder(block, selector) {
  assert.match(block, /border(?:-color)?:\s*(?:1px solid )?var\(--border-(?:default|strong)\)/, `${selector} must use a theme border`);
}

test('giant material test key containers use themed surfaces and borders', () => {
  const css = fs.readFileSync(path.join(root, 'user/pages/giant-material-test.css'), 'utf8');
  for (const selector of [
    '.giant-material-test-form, .giant-material-test-alert, .giant-material-test-stage',
    '.giant-material-test-input-row input',
    '.giant-material-test-body',
    '.giant-material-test-actions button'
  ]) {
    const block = cssBlock(css, selector);
    assertThemedSurface(block, selector);
    assertThemedBorder(block, selector);
  }
});

test('giant material test status selectors retain themed borders and surfaces', () => {
  const css = fs.readFileSync(path.join(root, 'user/pages/giant-material-test.css'), 'utf8');
  for (const status of ['success', 'error', 'warning', 'loading']) {
    const selector = `.giant-material-test-stage.is-${status}`;
    assertThemedBorder(cssBlock(css, selector), selector);
  }

  for (const status of ['success', 'error', 'warning']) {
    const selector = `.giant-material-test-stage.is-${status} .giant-material-test-stage-heading span`;
    assertThemedSurface(cssBlock(css, selector), selector);
  }
  assertThemedSurface(cssBlock(css, '.giant-material-test-stage-heading span'), '.giant-material-test-stage-heading span');
});

test('giant material status card and its states retain themed surfaces and borders', () => {
  const css = fs.readFileSync(path.join(root, 'user/pages/GiantMaterialStatusCard.css'), 'utf8');
  const cardSelector = '.giant-material-status-card';
  const cardBlock = cssBlock(css, cardSelector);
  assertThemedSurface(cardBlock, cardSelector);
  assertThemedBorder(cardBlock, cardSelector);

  for (const status of ['loading', 'success', 'error', 'warning']) {
    const selector = `.giant-material-status-card.is-${status}`;
    assertThemedBorder(cssBlock(css, selector), selector);
    const badgeSelector = `${selector} .giant-material-status-heading span`;
    assertThemedSurface(cssBlock(css, badgeSelector), badgeSelector);
  }
});

test('shuihuo production key text selectors use semantic text variables', () => {
  const css = fs.readFileSync(path.join(root, 'user/pages/shuihuo-production.css'), 'utf8');
  assert.match(cssBlock(css, '.shuihuo-topbar h1'), /color:\s*var\(--text-primary\)/);
  assert.match(cssBlock(css, '.shuihuo-page-heading p'), /color:\s*var\(--text-secondary\)/);
  assert.match(cssBlock(css, '.shuihuo-project-card-cover-fallback small'), /color:\s*var\(--text-tertiary\)/);
  assert.match(cssBlock(css, '.shuihuo-readiness-strip .missing'), /color:\s*var\(--text-danger\)/);
  assert.match(cssBlock(css, '.shuihuo-readiness-strip .ready'), /color:\s*var\(--text-success\)/);
});

test('batch factory v11 theme key text selectors use semantic variables and themed status surfaces', () => {
  const css = fs.readFileSync(path.join(root, 'user/pages/batch-factory-v11/batch-factory-v11-theme.css'), 'utf8');
  assert.match(cssBlock(css, "[data-bf-v11-ui='final'] .bf11-workbench .ant-typography"), /color:\s*var\(--text-primary\)/);
  assert.match(cssBlock(css, "[data-bf-v11-ui='final'] .bf11-workbench .ant-typography-secondary"), /color:\s*var\(--text-secondary\)/);
  assert.match(cssBlock(css, "[data-bf-v11-ui='final'] .bf11-card-title small"), /color:\s*var\(--text-tertiary\)/);
  assert.match(cssBlock(css, "[data-bf-v11-ui='final'] .bf11-workbench .ant-tag-red"), /color:\s*var\(--text-danger\)/);
  assert.match(cssBlock(css, "[data-bf-v11-ui='final'] .bf11-workbench .ant-tag-green"), /color:\s*var\(--text-success\)/);
  assert.match(cssBlock(css, "[data-bf-v11-ui='final'] .bf11-workbench .ant-tag-gold"), /color:\s*var\(--text-warning\)/);
  assert.match(cssBlock(css, "[data-bf-v11-ui='final'] .bf11-workbench .ant-tag-red"), /background:\s*var\(--surface-panel\)/);
  assert.match(cssBlock(css, "[data-bf-v11-ui='final'] .bf11-workbench .ant-tag-red"), /border-color:\s*var\(--border-default\)/);
});

test('batch factory and shuihuo styles use semantic text tokens for status text', () => {
  const files = [
    'user/pages/shuihuo-production.css',
    'user/pages/batch-factory-workbench.css',
    'user/pages/batch-factory-v11/batch-factory-v11-theme.css',
    'user/pages/batch-factory-v11/batch-factory-v11-workbench.css',
    'user/pages/batch-factory-v11/batch-factory-v11-settings.css',
    'user/pages/batch-factory-v11/batch-factory-v11-scoped.css',
    'user/pages/batch-factory-v11/batch-factory-v11-detail.css',
    'user/pages/batch-factory-v11/batch-factory-v11-batch-manager.css'
  ];
  for (const file of files) {
    const css = fs.readFileSync(path.join(root, file), 'utf8');
    assert.match(css, /var\(--text-(?:primary|secondary|tertiary|disabled|danger|warning|success)\)/, `${file} must use semantic text tokens`);
    // 富文本提示词编辑器的透明 textarea 是刻意的输入覆盖层：文字由下层 rich-view 绘制，
    // 上层只承接光标与选区，不参与可读性判定。
    const withoutInputOverlay = css
      .split('\n')
      .filter(line => !line.includes('.batch-factory-prompt-editor-input-overlay'))
      .join('\n');
    assert.doesNotMatch(withoutInputOverlay, /(?<![\w-])(?:color|-webkit-text-fill-color)\s*:\s*transparent/);
  }
});

test('batch factory workbench covers table, drawer, metadata and state text levels', () => {
  const css = fs.readFileSync(path.join(root, 'user/pages/batch-factory-workbench.css'), 'utf8');
  assert.match(cssBlock(css, '.batch-factory-workbench'), /color:\s*var\(--text-primary\)/);
  assert.match(cssBlock(css, '.batch-factory-status-grid button'), /color:\s*var\(--text-secondary\)/);
  assert.match(cssBlock(css, '.batch-factory-status-grid button strong'), /color:\s*var\(--text-primary\)/);
  assert.match(cssBlock(css, '.batch-factory-abnormal-summary .ant-btn-dangerous'), /color:\s*var\(--text-danger\)/);
});

test('batch factory v11 workbench covers task, metadata and error text levels', () => {
  const css = fs.readFileSync(path.join(root, 'user/pages/batch-factory-v11/batch-factory-v11-workbench.css'), 'utf8');
  assert.match(cssBlock(css, '.bf11-workbench .ant-typography'), /color:\s*var\(--text-primary\)/);
  assert.match(cssBlock(css, '.bf11-workbench .ant-typography-secondary'), /color:\s*var\(--text-secondary\)/);
  assert.match(cssBlock(css, '.bf11-card-title small'), /color:\s*var\(--text-tertiary\)/);
  assert.match(cssBlock(css, '.bf11-book-row.is-error'), /border-inline-start:\s*3px solid var\(--bf11-danger\)/);
});

test('batch factory v11 settings, detail and manager labels use semantic levels', () => {
  const files = [
    ['user/pages/batch-factory-v11/batch-factory-v11-settings.css', '.bf11-settings-drawer'],
    ['user/pages/batch-factory-v11/batch-factory-v11-scoped.css', '.bf11-scoped-settings'],
    ['user/pages/batch-factory-v11/batch-factory-v11-detail.css', '.bf11-book-collapse'],
    ['user/pages/batch-factory-v11/batch-factory-v11-batch-manager.css', '.bf11-batch-manager-pane']
  ];
  for (const [file, selector] of files) {
    const css = fs.readFileSync(path.join(root, file), 'utf8');
    assert.match(cssBlock(css, selector), /color:\s*var\(--text-(?:primary|secondary|tertiary)\)/, `${selector} must define readable text`);
  }
});

test('light theme primary and disabled buttons keep readable text contrast', () => {
  // 主按钮文字与底色都跟风格走：底色 = 当前风格主色，文字 = 该主色上的对比色令牌。
  assert.match(globalCSS, /\[data-theme='light'\]\s+\[data-shell='user'\] \.ant-btn-primary\s*\{[^}]*color:\s*var\(--text-on-accent\)/s);
  assert.match(globalCSS, /\[data-theme='light'\]\s+\[data-shell='user'\] \.ant-btn-primary\s*\{[^}]*background:\s*var\(--theme-primary\)/s);
  assert.match(globalCSS, /\[data-shell='user'\] \.ant-btn:disabled\s*\{[^}]*color:\s*var\(--legacy-dim\)/s);
  assert.match(globalCSS, /\[data-shell='admin'\] \.ant-btn:disabled\s*\{[^}]*color:\s*var\(--admin-faint\)/s);
});

test('global AntD tags use theme-aware contrast without flattening dark status colors', () => {
  assert.match(globalCSS, /\[data-theme\] \.ant-tag\s*\{[^}]*color:\s*var\(--tag-text\)/s);
  assert.match(globalCSS, /\[data-theme\] \.ant-tag\s*\{[^}]*background:\s*var\(--tag-bg\)/s);
  for (const color of ['green', 'blue', 'gold', 'purple', 'red']) {
    assert.match(globalCSS, new RegExp(`\\[data-theme\\] \\.ant-tag-${color}[^}]*color:\\s*var\\(--tag-${color}-text\\)`));
    assert.match(globalCSS, new RegExp(`\\[data-theme\\] \\.ant-tag-${color}[^}]*background:\\s*var\\(--tag-${color}-bg\\)`));
  }
  assert.doesNotMatch(globalCSS, /\[data-theme\] \.ant-tag\s*\{[^}]*color:\s*#fff\s*!important/s);
});
