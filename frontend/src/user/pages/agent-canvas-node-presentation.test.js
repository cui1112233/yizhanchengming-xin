import test from 'node:test';
import assert from 'node:assert/strict';
import {
  canvasEditorFocusOptions,
  canvasNodeStatusLabel,
  canvasNodeTypeLabel,
  isMediaCanvasNodeType,
  normalizeCanvasNodeType,
  usesPinnedBatchFactoryComposer
} from './agent-canvas-node-presentation.js';

test('legacy, blank, and unknown node types become text', () => {
  for (const value of ['copy', '', null, 'unexpected']) {
    assert.equal(normalizeCanvasNodeType(value), 'text');
  }
});

test('supported media types remain media types', () => {
  for (const value of ['image', 'video', 'audio']) {
    assert.equal(normalizeCanvasNodeType(value), value);
    assert.equal(isMediaCanvasNodeType(value), true);
  }
  assert.equal(isMediaCanvasNodeType('copy'), false);
});

test('节点类型和状态使用面向用户的中文标签', () => {
  assert.equal(canvasNodeTypeLabel('text'), '文本');
  assert.equal(canvasNodeTypeLabel('image'), '图片');
  assert.equal(canvasNodeTypeLabel('video'), '视频');
  assert.equal(canvasNodeTypeLabel('audio'), '音频');
  assert.equal(canvasNodeTypeLabel('unknown'), '文本');
  assert.equal(canvasNodeStatusLabel('draft'), '草稿');
  assert.equal(canvasNodeStatusLabel('confirmed'), '已确认');
  assert.equal(canvasNodeStatusLabel('failed'), '需处理');
  assert.equal(canvasNodeStatusLabel('running'), '生成中');
});

test('节点编辑器只在画布过小时返回聚焦参数', () => {
  assert.deepEqual(canvasEditorFocusOptions(0.57), {
    padding: 0.16,
    duration: 260,
    minZoom: 0.75,
    maxZoom: 1
  });
  assert.equal(canvasEditorFocusOptions(0.85), null);
  assert.equal(canvasEditorFocusOptions(1.2), null);
});

test('批量工厂剧本卡固定显示生产控制，不能被折叠编辑器隐藏', () => {
  assert.equal(usesPinnedBatchFactoryComposer({
    type: 'script',
    config: { source: 'batch_factory', resourceType: 'book_script' }
  }), true);
  assert.equal(usesPinnedBatchFactoryComposer({
    type: 'script',
    config: { source: 'batch_factory', resourceType: 'video' }
  }), false);
  assert.equal(usesPinnedBatchFactoryComposer({
    type: 'text',
    config: { source: 'batch_factory', resourceType: 'book_script' }
  }), false);
});
