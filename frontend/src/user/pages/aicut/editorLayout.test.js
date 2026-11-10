import assert from 'node:assert/strict';
import test from 'node:test';
import { defaultEditorLayout, normalizeEditorLayout, toggleEditorPane } from './editorLayout.js';

test('默认布局按桌面视口生成三块可用尺寸', () => {
  const layout = normalizeEditorLayout(defaultEditorLayout, { width: 1600, height: 1000 });
  assert.deepEqual(layout, {
    resourceWidth: 352,
    inspectorWidth: 352,
    timelineHeight: 400,
    resourceCollapsed: false,
    inspectorCollapsed: false,
    timelineCollapsed: false,
    inspectorMode: 'panel',
  });
});

test('布局尺寸会限制到可操作边界并修复非法持久化值', () => {
  assert.deepEqual(
    normalizeEditorLayout({ resourceWidth: 2, inspectorWidth: 9999, timelineHeight: 'bad' }, { width: 1400, height: 800 }),
    {
      resourceWidth: 260,
      inspectorWidth: 520,
      timelineHeight: 320,
      resourceCollapsed: false,
      inspectorCollapsed: false,
      timelineCollapsed: false,
      inspectorMode: 'panel',
    },
  );
});

test('窄窗口把属性栏切换为抽屉但保留用户宽度', () => {
  const layout = normalizeEditorLayout({ inspectorWidth: 410 }, { width: 1024, height: 768 });
  assert.equal(layout.inspectorMode, 'drawer');
  assert.equal(layout.inspectorWidth, 410);
});

test('三个区域可独立收起和恢复且不改变尺寸', () => {
  let layout = normalizeEditorLayout({ resourceWidth: 330, timelineHeight: 280 }, { width: 1440, height: 900 });
  layout = toggleEditorPane(layout, 'resource');
  layout = toggleEditorPane(layout, 'timeline');
  assert.equal(layout.resourceCollapsed, true);
  assert.equal(layout.timelineCollapsed, true);
  assert.equal(layout.resourceWidth, 330);
  assert.equal(layout.timelineHeight, 280);
  layout = toggleEditorPane(layout, 'resource');
  assert.equal(layout.resourceCollapsed, false);
});

test('未知区域不会修改布局对象', () => {
  const layout = normalizeEditorLayout({}, { width: 1440, height: 900 });
  assert.strictEqual(toggleEditorPane(layout, 'unknown'), layout);
});
