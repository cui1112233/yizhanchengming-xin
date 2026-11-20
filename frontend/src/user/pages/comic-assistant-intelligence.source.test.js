import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

const panel = readFileSync(new URL('./ComicScreenplayAuthoringPanels.jsx', import.meta.url), 'utf8');
const page = readFileSync(new URL('./ComicScreenplayPage.jsx', import.meta.url), 'utf8');
const css = readFileSync(new URL('./comic-screenplay.css', import.meta.url), 'utf8');

test('助手面板接收导航回调并传入 onNavigate', () => {
  assert.match(panel, /export function ComicAssistantPanel\(\{[^}]*onNavigate/);
  assert.match(page, /<ComicAssistantPanel[^>]*onNavigate=\{navigateStage\}/);
});

test('助手面板渲染主动追问消息（kind=question）', () => {
  assert.match(panel, /kind==='question'/);
  assert.match(panel, /comic-assistant-question/);
  assert.match(panel, /在下方输入框回答/);
});

test('助手面板渲染会话折叠占位（kind=compacted）', () => {
  assert.match(panel, /kind==='compacted'/);
  assert.match(panel, /comic-assistant-compacted/);
});

test('助手面板渲染导航建议按钮并回调 entry', () => {
  assert.match(panel, /item\.suggestions/);
  assert.match(panel, /onNavigate\?\.\(s\.entry\)/);
  assert.match(panel, /comic-assistant-suggestions/);
});

test('样式定义追问/折叠/建议三类展示', () => {
  assert.match(css, /\.comic-assistant-question/);
  assert.match(css, /\.comic-assistant-compacted/);
  assert.match(css, /\.comic-assistant-suggestions/);
});
