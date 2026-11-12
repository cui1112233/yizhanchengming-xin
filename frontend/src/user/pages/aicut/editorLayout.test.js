import assert from 'node:assert/strict';
import test from 'node:test';
import { defaultEditorLayout, normalizeEditorLayout, toggleEditorPane } from './editorLayout.js';

test('默认布局按桌面视口生成三块可用尺寸', () => {
  const layout = normalizeEditorLayout(defaultEditorLayout, { width: 1600, height: 1000 });
  assert.deepEqual(layout, {
    resourceWidth: 320,
    inspectorWidth: 320,
    timelineHeight: 340,
    resourceCollapsed: false,
    inspectorCollapsed: false,
    timelineCollapsed: false,
    playheadFollow: true,
    inspectorMode: 'panel',
  });
});

test('布局尺寸会限制到可操作边界并修复非法持久化值', () => {
  assert.deepEqual(
    normalizeEditorLayout({ resourceWidth: 2, inspectorWidth: 9999, timelineHeight: 'bad' }, { width: 1400, height: 800 }),
    {
      resourceWidth: 260,
      inspectorWidth: 400,
      timelineHeight: 272,
      resourceCollapsed: false,
      inspectorCollapsed: false,
      timelineCollapsed: false,
      playheadFollow: true,
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

test('播放头跟随默认开启并保留用户关闭选择', () => {
  assert.equal(normalizeEditorLayout({}, { width: 1440, height: 900 }).playheadFollow, true);
  assert.equal(normalizeEditorLayout({ playheadFollow: false }, { width: 1440, height: 900 }).playheadFollow, false);
});

test('未知区域不会修改布局对象', () => {
  const layout = normalizeEditorLayout({}, { width: 1440, height: 900 });
  assert.strictEqual(toggleEditorPane(layout, 'unknown'), layout);
});

test('oversized saved panes leave at least 480px for the desktop player', () => {
  for (const width of [1280, 1440, 1600]) {
    const layout = normalizeEditorLayout({ resourceWidth: 9999, inspectorWidth: 9999 }, { width, height: 800 });
    assert.ok(width - layout.resourceWidth - layout.inspectorWidth >= 480);
  }
  assert.equal(normalizeEditorLayout({}, { width: 1200, height: 800 }).inspectorMode, 'drawer');
});
