import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const dirname = path.dirname(fileURLToPath(import.meta.url));
const css = fs.readFileSync(path.resolve(dirname, '../../shared/styles/agent-workspace.css'), 'utf8');
const pageSource = fs.readFileSync(path.resolve(dirname, './AgentPageV2.jsx'), 'utf8');
const canvasSource = fs.readFileSync(path.resolve(dirname, '../components/AgentCanvas.jsx'), 'utf8');

function cssBlock(selector) {
  const escapedSelector = selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const match = css.match(new RegExp(`${escapedSelector}\\s*\\{([^}]*)\\}`, 's'));
  assert.ok(match, `missing CSS block for ${selector}`);
  return match[1];
}

test('Agent 创作首页覆盖旧版双列和双行布局', () => {
  const page = cssBlock('.agent-page.utility-workbench');
  assert.match(page, /grid-template-columns:\s*minmax\(0,\s*1fr\)/);
  assert.match(page, /grid-template-rows:\s*minmax\(0,\s*1fr\)/);
});

test('Agent 创作选项可换行并使用清晰的主题按钮颜色', () => {
  const actions = cssBlock('.agent-page .agent-composer-actions-left');
  assert.match(actions, /display:\s*flex/);
  assert.match(actions, /flex-wrap:\s*wrap/);

  const chip = cssBlock('.agent-page .agent-context-chip');
  assert.match(chip, /background:\s*var\(--agent-surface\)/);
  assert.match(chip, /color:\s*var\(--agent-text\)/);
  assert.match(chip, /border:\s*1px solid var\(--agent-border\)/);
});

test('Agent 窄屏对话区使用有界底部面板', () => {
  assert.match(css, /@media\s*\(max-width:\s*560px\)[\s\S]*?\.agent-page \.agent-conversation-panel\s*\{[^}]*top:\s*auto[^}]*height:\s*min\(72dvh,\s*620px\)[^}]*width:\s*100%/);
});

test('Agent 窄屏画布隐藏会遮挡输入按钮的桌面宠物', () => {
  assert.match(css, /@media\s*\(max-width:\s*560px\)[\s\S]*?\.legacy-shell:has\(\.agent-page\.agent-workspace-v3\)\s*>\s*\.cm-penguin-shell\s*\{[^}]*display:\s*none/);
});

test('进入画布页后会重新读取当前项目画布', () => {
  const canvasLoadEffect = pageSource.match(/Promise\.all\(\[getAgentCanvas\(activeTaskId\)[\s\S]*?\n\s*}, \[([^\]]+)\]\);/);
  assert.ok(canvasLoadEffect, 'missing canvas load effect');
  assert.match(canvasLoadEffect[1], /activeTaskId/);
  assert.match(canvasLoadEffect[1], /isCanvasWorkspace/);
});

test('A 方案画布工具使用固定尺寸且不会因禁用状态位移', () => {
  for (const selector of ['.agent-page .agent-canvas-toolrail', '.agent-page .agent-canvas-editbar', '.agent-page .agent-canvas-zoombar']) {
    const block = cssBlock(selector);
    assert.match(block, /position:\s*absolute/);
    assert.match(block, /z-index:\s*1[0-9]/);
  }
  const buttons = cssBlock('.agent-page .agent-canvas :is(.agent-canvas-toolrail, .agent-canvas-editbar, .agent-canvas-zoombar) .ant-btn');
  assert.match(buttons, /width:\s*36px/);
  assert.match(buttons, /height:\s*36px/);
  assert.match(buttons, /border-radius:\s*[0-8]px/);
});

test('窄屏编辑工具可横向滚动且缩放栏保持有界', () => {
  assert.match(css, /@media\s*\(max-width:\s*560px\)[\s\S]*?\.agent-page \.agent-canvas-editbar\s*\{[^}]*max-width:\s*calc\(100% - 72px\)[^}]*overflow-x:\s*auto/);
  assert.match(css, /@media\s*\(max-width:\s*560px\)[\s\S]*?\.agent-page \.agent-canvas-zoombar\s*\{[^}]*max-width:\s*calc\(100% - 24px\)/);
  assert.match(css, /@media\s*\(max-width:\s*560px\)[\s\S]*?\.agent-canvas-workspace:not\(\.is-conversation-collapsed\) \.agent-canvas-zoombar\s*\{[^}]*bottom:\s*calc\(min\(72dvh,\s*620px\) \+ 12px\)/);
});

test('节点点击明确支持 Ctrl Shift 和 Command 多选', () => {
  assert.match(canvasSource, /onNodeClick=\{handleNodeClick\}/);
  assert.match(canvasSource, /event\.ctrlKey\s*\|\|\s*event\.metaKey\s*\|\|\s*event\.shiftKey/);
  assert.match(canvasSource, /selectCanvasNodes\([\s\S]*?nodesRef\.current,[\s\S]*?node\.id,[\s\S]*?selectedIds/);
});

test('画布提供持久外观设置和按方向流动的连线', () => {
  for (const label of ['连线颜色', '连线动效', '连线粗细', '背景网点', '辅助线对齐']) {
    assert.match(canvasSource, new RegExp(label));
  }
  assert.match(canvasSource, /label="画布外观"/);
  assert.match(canvasSource, /edgeTypes=\{canvasEdgeTypes\}/);
  assert.match(canvasSource, /edgeAnimated/);
  assert.match(canvasSource, /markerEnd/);
  assert.match(canvasSource, /appearance:\s*normalizeCanvasAppearance\(appearance\)/);
  assert.match(canvasSource, /onAppearanceChange/);
});

test('画布外观包含轻量动效和窄屏安全规则', () => {
  assert.match(css, /\.agent-page \.agent-canvas-edge-flow\s*\{[^}]*animation:/);
  assert.match(css, /@keyframes agent-canvas-edge-flow/);
  assert.match(css, /\.agent-page \.agent-canvas-appearance-panel\s*\{/);
  assert.match(css, /\.agent-canvas-appearance-popover\s*\{[^}]*--agent-edge-coral:\s*#f07167/);
  assert.match(css, /@media \(max-width: 560px\)[\s\S]*?\.agent-canvas-appearance-popover[^}]*max-width:\s*calc\(100vw - 24px\)/);
  assert.match(css, /@media \(prefers-reduced-motion: reduce\)[\s\S]*?\.agent-canvas-edge-flow\s*\{[^}]*animation:\s*none/);
});
