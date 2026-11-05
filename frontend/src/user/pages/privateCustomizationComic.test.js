import test from 'node:test';
import assert from 'node:assert/strict';

import {
  buildPrivateComicSettings,
  mergePrivateComicReference,
  privateComicDraftKey,
  privateComicTaskKey,
  privateComicTaskSnapshot,
  privateComicTitle,
  shouldReusePrivateComicTask
} from './privateCustomizationComic.js';

test('uses the first non-empty source line as the editable comic title fallback', () => {
  assert.equal(privateComicTitle('\n  雨夜归人  \n第二行正文', ''), '雨夜归人');
  assert.equal(privateComicTitle('正文', '  自定义标题  '), '自定义标题');
});

test('rejects an empty pasted source with a user-readable message', () => {
  assert.equal(privateComicTitle('   ', ''), '未命名漫画');
  assert.throws(() => buildPrivateComicSettings({ sourceText: '  ' }), /请先粘贴漫画正文/);
});

test('preserves comic style and reference settings in the Batch Factory comic patch', () => {
  const settings = buildPrivateComicSettings({
    sourceText: '林晚走进无人的雨夜，远处的灯光在积水里慢慢晃动。',
    comic: {
      enabled: true,
      style: { presetId: 'batch-comic-style-01', presetName: '顶级韩漫风格' },
      globalReferenceImage: { assetId: 'style:batch-comic-style-01', enabled: true, applyToAssets: true, applyToPages: true },
      imagePromptPrefix: '统一前缀',
      imagePromptSuffix: '统一后缀',
      constraints: [{ presetId: 'comic-quality', constraintCategory: 'comic_quality' }]
    },
    textModelId: 'text-1',
    imageModelId: 'image-1',
    imageProvider: 'openai'
  });

  assert.equal(settings.productionMode, 'comic');
  assert.equal(settings.textModelId, 'text-1');
  assert.equal(settings.imageModelId, 'image-1');
  assert.equal(settings.aiPromptConfig.comic.style.presetId, 'batch-comic-style-01');
  assert.equal(settings.aiPromptConfig.comic.globalReferenceImage.applyToPages, true);
  assert.equal(settings.aiPromptConfig.comic.imagePromptPrefix, '统一前缀');
  assert.equal(settings.aiPromptConfig.comic.imagePromptSuffix, '统一后缀');
  assert.equal(settings.aiPromptConfig.comic.constraints[0].constraintCategory, 'comic_quality');
});

test('stores only the private comic task identity needed to restore progress', () => {
  assert.equal(privateComicDraftKey, 'private-customization-comic-draft');
  assert.equal(privateComicTaskKey, 'private-customization-comic-task');
  assert.deepEqual(privateComicTaskSnapshot({ batch: { id: 'batch-1' }, book: { id: 'book-1' } }), {
    batchId: 'batch-1',
    bookId: 'book-1',
    status: 'created'
  });
});

test('reuses an existing comic task instead of creating duplicate batches', () => {
  const task = { batchId: 'batch-1', bookId: 'book-1', sourceText: '第一段正文' };
  assert.equal(shouldReusePrivateComicTask(task, '第一段正文'), true);
  assert.equal(shouldReusePrivateComicTask(task, '  第一段正文  '), true);
  assert.equal(shouldReusePrivateComicTask(task, '完全不同的正文'), false);
  assert.equal(shouldReusePrivateComicTask({ batchId: 'batch-1', bookId: 'book-1' }, ''), true);
  assert.equal(shouldReusePrivateComicTask({ batchId: '', bookId: '' }, '第一段正文'), false);
});

test('merges an uploaded reference into the first automation settings without losing user flags', () => {
  const settings = buildPrivateComicSettings({
    sourceText: '这是足够长的一段漫画正文，用来验证上传参考图后设置仍然完整。',
    comic: { globalReferenceImage: { enabled: false, applyToAssets: true, applyToPages: false } }
  });
  const merged = mergePrivateComicReference(settings, {
    assetId: 'asset-uploaded',
    url: '/api/local-executor/v1/reference/asset-uploaded',
    previewUrl: '/preview/asset-uploaded'
  });
  assert.equal(merged.aiPromptConfig.comic.globalReferenceImage.assetId, 'asset-uploaded');
  assert.equal(merged.aiPromptConfig.comic.globalReferenceImage.enabled, false);
  assert.equal(merged.aiPromptConfig.comic.globalReferenceImage.applyToPages, false);
  assert.equal(merged.aiPromptConfig.comic.globalReferenceImage.source, 'manual-upload');
});
