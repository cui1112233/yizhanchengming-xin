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
import { canExtractPrivateComicAssets, ensurePrivateComicTask, preparePrivateComicTaskSettings } from './privateComicSingleBookFlow.js';

const singleBookSource = '赵家堂屋的门帘是新换的，红布面压着一道黄边，屋里没有一个人。';

test('asset extraction is available from pasted content before a server task exists', () => {
  assert.equal(canExtractPrivateComicAssets({ sourceText: singleBookSource, busy: false, action: '' }), true);
  assert.equal(canExtractPrivateComicAssets({ sourceText: '太短', busy: false, action: '' }), false);
  assert.equal(canExtractPrivateComicAssets({ sourceText: singleBookSource, busy: true, action: '' }), false);
});

test('single-book flow creates one task without requiring an automation preset', async () => {
  const calls = [];
  const result = await ensurePrivateComicTask({
    task: null,
    draft: {
      title: '红门帘', sourceText: singleBookSource, comic: { enabled: true },
      textModelId: 'text-model-1', imageModelId: 'image-model-1', imageProvider: 'personal_api'
    },
    createTask: async payload => {
      calls.push(payload);
      return { batch: { id: 'batch-1' }, book: { id: 'book-1' } };
    }
  });
  assert.equal(calls.length, 1);
  assert.equal(calls[0].title, '红门帘');
  assert.equal(calls[0].sourceText, singleBookSource);
  assert.equal(calls[0].initialBatchSettings.productionMode, 'comic');
  assert.equal(calls[0].initialBatchSettings.assetImagesEnabled, false);
  assert.deepEqual(result.task, { batchId: 'batch-1', bookId: 'book-1', status: 'created', sourceText: singleBookSource });
  assert.equal(result.created, true);
});

test('single-book comic keeps asset images optional and uses text asset prompts as fallback', () => {
  const settings = buildPrivateComicSettings({ sourceText: singleBookSource });
  assert.equal(settings.assetImagesEnabled, false);
});

test('single-book flow reuses the current book when the source has not changed', async () => {
  const current = { batchId: 'batch-1', bookId: 'book-1', sourceText: singleBookSource };
  const result = await ensurePrivateComicTask({
    task: current,
    draft: { sourceText: singleBookSource, comic: {}, textModelId: 'text-model-1' },
    createTask: async () => { throw new Error('不应重复创建任务'); }
  });
  assert.equal(result.task, current);
  assert.equal(result.created, false);
});

test('single-book flow repairs settings on a reused legacy task before running a stage', async () => {
  const calls = [];
  const ensured = {
    task: { batchId: 'batch-legacy', bookId: 'book-legacy' },
    created: false,
    settings: { productionMode: 'comic', aiPromptConfig: { comic: { enabled: true, scope: 'private_customization' } } }
  };
  await preparePrivateComicTaskSettings({
    ensured,
    saveSettings: async (batchId, settings) => calls.push({ batchId, settings })
  });
  assert.deepEqual(calls, [{ batchId: 'batch-legacy', settings: ensured.settings }]);
});

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
