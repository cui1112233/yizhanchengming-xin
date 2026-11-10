import test from 'node:test';
import assert from 'node:assert/strict';

import { clampAICutLayout, resolveAICutShortcut, timelineDropTime } from './aicutWorkbench.js';

test('工作台尺寸超出边界时夹取到仍可操作的范围', () => {
  assert.deepEqual(clampAICutLayout({}, { width: 1440, height: 900 }), { agentWidth: 350, mediaWidth: 450, topHeight: 520 });
  assert.deepEqual(clampAICutLayout(
    { agentWidth: 20, mediaWidth: 900, topHeight: 9999 },
    { width: 1440, height: 900 }
  ), { agentWidth: 56, mediaWidth: 620, topHeight: 680 });
  assert.deepEqual(clampAICutLayout(
    { agentWidth: 380, mediaWidth: 420, topHeight: 440 },
    { width: 1440, height: 900 }
  ), { agentWidth: 380, mediaWidth: 420, topHeight: 440 });
});

test('素材落点把可视坐标和横向滚动换算成时间线毫秒', () => {
  assert.equal(timelineDropTime(250, 100, 72, 1), 3083);
  assert.equal(timelineDropTime(80, 100, 0, 2), 0);
});

test('快捷键在输入文字时不截获普通按键但保留组合操作和关闭', () => {
  assert.equal(resolveAICutShortcut({ key: ' ', code: 'Space', targetTag: 'TEXTAREA' }, { hasProject: true }), null);
  assert.equal(resolveAICutShortcut({ key: 'Delete', targetTag: 'INPUT' }, { hasSelection: true }), null);
  assert.equal(resolveAICutShortcut({ key: 's', ctrlKey: true, targetTag: 'INPUT' }, { hasProject: true }), 'save');
  assert.equal(resolveAICutShortcut({ key: 'Escape', targetTag: 'TEXTAREA' }, {}), 'close');
});

test('快捷键只在具备项目或选中片段时返回可执行动作', () => {
  assert.equal(resolveAICutShortcut({ key: ' ', code: 'Space' }, { hasProject: true }), 'play');
  assert.equal(resolveAICutShortcut({ key: 'e', ctrlKey: true }, { hasProject: true }), 'export');
  assert.equal(resolveAICutShortcut({ key: 'b', ctrlKey: true }, { hasProject: true, hasSelection: true }), 'split');
  assert.equal(resolveAICutShortcut({ key: 'b' }, { hasProject: true }), 'blade');
  assert.equal(resolveAICutShortcut({ key: 'Delete' }, { hasProject: true, hasSelection: true }), 'delete');
  assert.equal(resolveAICutShortcut({ key: 'z', ctrlKey: true }, { hasProject: true }), 'undo');
  assert.equal(resolveAICutShortcut({ key: 'Z', ctrlKey: true, shiftKey: true }, { hasProject: true }), 'redo');
  assert.equal(resolveAICutShortcut({ key: '/', ctrlKey: true }, {}), 'shortcuts');
  assert.equal(resolveAICutShortcut({ key: 'n', ctrlKey: true }, {}), 'new-project');
  assert.equal(resolveAICutShortcut({ key: 'Delete' }, { hasProject: true, hasSelection: false }), null);
});

test('快捷键可以在AI创作画布和专业剪辑之间切换', () => {
  assert.equal(resolveAICutShortcut({ key: '1', ctrlKey: true }), 'canvas-view');
  assert.equal(resolveAICutShortcut({ key: '2', ctrlKey: true }), 'editor-view');
});
