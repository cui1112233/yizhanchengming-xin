import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';

const PAGE = readFileSync(new URL('./TtsPage.jsx', import.meta.url), 'utf8');
const PAGE_CSS = readFileSync(new URL('./tts-workbench.css', import.meta.url), 'utf8');
const TOKENS = readFileSync(new URL('../../shared/styles/v88-tokens.css', import.meta.url), 'utf8');
const UI_CSS = readFileSync(new URL('../../shared/styles/v88-ui.css', import.meta.url), 'utf8');
const BUTTON = readFileSync(new URL('../../shared/components/V88Button.jsx', import.meta.url), 'utf8');

test('配音页使用 V88 容器与常量内边距', () => {
  assert.match(PAGE, /className="tts-workbench utility-workbench"/);
  assert.match(PAGE_CSS, /\.tts-workbench\s*\{[\s\S]*padding: var\(--v88-space-page\)/);
  assert.match(PAGE_CSS, /\.tts-workbench\s*\{[\s\S]*max-width: none|\.tts-workbench\s*\{[\s\S]*width: 100%/);
});

test('配音页按钮走 V88Button + ConfigProvider 主题，不再手写 antd Button', () => {
  assert.match(PAGE, /import \{ V88Button \} from '\.\.\/\.\.\/shared\/components\/V88Button'/);
  assert.match(PAGE, /<ConfigProvider theme=\{antTheme\}>/);
  assert.match(PAGE, /createV88AntTheme\(theme\)/);
  assert.doesNotMatch(PAGE, /import \{[^}]*Button[^}]*\} from 'antd'/);
  assert.match(BUTTON, /disabled=\{disabled \|\| loading\}/);
});

test('危险操作使用 danger 变体而不是主色', () => {
  assert.match(PAGE, /variant="danger"/);
  assert.match(PAGE, /onClick=\{onRemove\}/);
});

test('空态使用统一 V88EmptyState 组件且主操作可用', () => {
  assert.match(PAGE, /import \{ V88EmptyState \}/);
  assert.match(PAGE, /<V88EmptyState/);
  assert.match(PAGE, /action=\{<V88Button variant="primary"/);
  assert.match(UI_CSS, /\.v88-empty-title/);
});

test('等待态绑定真实状态：配置读取/保存默认/读文件/批量生成/单卡生成', () => {
  assert.match(PAGE, /setConfigLoading\(true\)/);
  assert.match(PAGE, /loading=\{savingDefaults\}/);
  assert.match(PAGE, /loading=\{readingFile\}/);
  assert.match(PAGE, /loading=\{bulkGenerating\}/);
  assert.match(PAGE, /loading=\{card\.loading\}/);
  assert.match(PAGE, /正在批量生成 \{bulkProgress\.done\}\/\{bulkProgress\.total\} 张/);
  assert.match(PAGE, /正在合成音频…已等待/);
});

test('原有按钮与入口一个不少', () => {
  for (const label of ['添加卡片', '上传小说', '保存为默认配音', '全部生成', '生成语音', '下载', '删除']) {
    assert.match(PAGE, new RegExp(label));
  }
  assert.match(PAGE, /ariaLabel="默认语速"/);
  assert.match(PAGE, /ariaLabel="默认音调"/);
  assert.match(PAGE, /<audio controls src=\{card\.audioUrl\}/);
});

test('令牌文件：主色青绿、三档动效、减少动态降级、私人定制映射', () => {
  assert.match(TOKENS, /--v88-color-primary: #63dfc0;/);
  assert.match(TOKENS, /--v88-color-emphasis: #f07167;/);
  assert.match(TOKENS, /--v88-color-info: #9da8ff;/);
  assert.match(TOKENS, /--v88-radius-md: 9px;/);
  assert.match(TOKENS, /--v88-space-page: 24px;/);
  assert.match(TOKENS, /--v88-motion-base: 160ms;/);
  assert.match(TOKENS, /prefers-reduced-motion: reduce/);
  assert.match(TOKENS, /--ps-mint: var\(--v88-color-primary\)/);
  assert.match(TOKENS, /--ps-indigo: var\(--v88-color-info\)/);
});

test('按钮体系与动效：五变体、焦点环、reduced-motion、无 !important', () => {
  for (const variant of ['primary', 'secondary', 'ghost', 'danger', 'icon']) {
    assert.match(BUTTON, new RegExp(`${variant}: \\{`));
  }
  assert.match(UI_CSS, /\.v88-btn:focus-visible/);
  assert.match(UI_CSS, /prefers-reduced-motion: reduce/);
  // 检查规则本身而不是注释：去掉 /* ... */ 后再断言没有强制覆盖。
  const rulesOnly = css => css.replace(/\/\*[\s\S]*?\*\//g, '');
  assert.doesNotMatch(rulesOnly(UI_CSS), /!important/);
  assert.doesNotMatch(rulesOnly(PAGE_CSS), /!important/);
});
