import assert from 'node:assert/strict';
import test from 'node:test';

import {
  assetMentionTrigger,
  filterMentionAssets,
  firstClipboardImage,
  insertAssetMention,
  persistClipboardImage,
  referenceKindsForPromptNode,
  shouldCloseCanvasEditorOnEscape,
  shouldHandleCanvasImagePaste
} from './agent-canvas-asset-input.js';

test('asset mention opens only for a standalone at query at the caret', () => {
  assert.deepEqual(assetMentionTrigger('让 @夜景 更明亮', 5), { start: 2, end: 5, query: '夜景' });
  assert.deepEqual(assetMentionTrigger('@', 1), { start: 0, end: 1, query: '' });
  assert.equal(assetMentionTrigger('mail@example.com', 16), null);
  assert.equal(assetMentionTrigger('上一行 @夜景\n下一行', 11), null);
});

test('mention candidates respect node capabilities, query, and asset identity', () => {
  const assets = [
    { id: 'image-1', kind: 'image', name: '夜景.png' },
    { id: 'video-1', kind: 'video', name: '夜景运镜.mp4' },
    { id: 'audio-1', kind: 'audio', name: '雨声.mp3' },
    { id: 'image-1', kind: 'image', name: '重复夜景.png' }
  ];

  assert.deepEqual(referenceKindsForPromptNode('image'), ['image']);
  assert.deepEqual(referenceKindsForPromptNode('video'), ['image', 'video']);
  assert.deepEqual(referenceKindsForPromptNode('audio'), ['audio']);
  assert.deepEqual(referenceKindsForPromptNode('text'), ['image', 'video', 'audio']);
  assert.deepEqual(filterMentionAssets(assets, ['image', 'video'], '夜景').map(asset => asset.id), ['image-1', 'video-1']);
});

test('selecting an asset replaces the active at query and returns the next caret', () => {
  assert.deepEqual(
    insertAssetMention('参考 @夜', { start: 3, end: 5, query: '夜' }, { name: '夜景.png' }),
    { value: '参考 @夜景.png ', caret: 11, label: '@夜景.png' }
  );
});

test('canvas paste accepts the first clipboard image and leaves editable fields alone', () => {
  const image = { name: 'clipboard.png', type: 'image/png' };
  const text = { name: 'notes.txt', type: 'text/plain' };
  const items = [
    { kind: 'string', type: 'text/plain', getAsFile: () => null },
    { kind: 'file', type: 'text/plain', getAsFile: () => text },
    { kind: 'file', type: 'image/png', getAsFile: () => image }
  ];

  assert.equal(firstClipboardImage(items), image);
  assert.equal(shouldHandleCanvasImagePaste({ hasImage: true, targetTagName: 'DIV' }), true);
  assert.equal(shouldHandleCanvasImagePaste({ hasImage: true, targetTagName: 'TEXTAREA' }), false);
  assert.equal(shouldHandleCanvasImagePaste({ hasImage: true, targetTagName: 'DIV', isContentEditable: true }), false);
  assert.equal(shouldHandleCanvasImagePaste({ hasImage: false, targetTagName: 'DIV' }), false);
});

test('pasted screenshots are uploaded before a durable image node is added', async () => {
  const image = { name: 'clipboard.png', type: 'image/png' };
  const asset = { id: 'asset-pasted', kind: 'image', name: 'clipboard.png', url: '/asset-pasted' };
  const calls = [];

  const result = await persistClipboardImage({
    file: image,
    taskId: 'project-1',
    position: { x: 120, y: 240 },
    upload: async (taskId, file) => {
      calls.push(['upload', taskId, file]);
      return { asset };
    },
    add: (created, position) => calls.push(['add', created, position])
  });

  assert.equal(result, asset);
  assert.deepEqual(calls, [
    ['upload', 'project-1', image],
    ['add', asset, { x: 120, y: 240 }]
  ]);
});

test('pasted screenshot does not add a node when upload returns a non-image asset', async () => {
  let added = false;
  await assert.rejects(() => persistClipboardImage({
    file: { name: 'clipboard.png', type: 'image/png' },
    taskId: 'project-1',
    position: { x: 0, y: 0 },
    upload: async () => ({ asset: { id: 'wrong-kind', kind: 'video' } }),
    add: () => { added = true; }
  }), /图片资产/);
  assert.equal(added, false);
});

test('escape closes an open asset mention before it closes the node editor', () => {
  assert.equal(shouldCloseCanvasEditorOnEscape({ insideComposer: true, mentionOpen: true }), false);
  assert.equal(shouldCloseCanvasEditorOnEscape({ insideComposer: true, mentionOpen: false }), true);
  assert.equal(shouldCloseCanvasEditorOnEscape({ insideComposer: false, mentionOpen: false }), true);
});
