import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { batchCreationSettings } from './batchFactoryProductionMode.js';
import { buildGiantMaterialPlaceholderIntakes } from './batchFactoryGiantMaterialImport.js';

const here = path.dirname(fileURLToPath(import.meta.url));
const read = name => fs.readFileSync(path.join(here, name), 'utf8');

test('the selected 8000-character limit reaches both intake and first direct read in all three modes', async () => {
  const source = read('BatchFactoryCreateModal.jsx');
  const selectLimit = source.match(/const contentCaptureCharacters = [^;]+;/)[0];
  const buildIntake = source.match(/const intakePayload = buildGiantMaterialPlaceholderIntakes\([\s\S]*?\n      \);/)[0];
  const firstRead = source.match(/await fetchBookOriginal\(batchId, book\.id[^;]*;/)[0];
  const material = { books: [{ title: '测试', platformBookId: '123', platformName: '常读' }] };
  const selected = [{ item: { id: '7689285397448523826', material }, book: material.books[0] }];
  const run = new Function('batchMode', 'initialBatchSettings', 'captureCharacters', 'selected',
    'buildGiantMaterialPlaceholderIntakes', 'fetchBookOriginal', `
      return (async () => {
        const contentRangeLines = 5, giantOriginalReadStrategy = 'direct_first', importedAt = 'now';
        ${selectLimit}
        ${buildIntake}
        const batchId = 'batch', book = intakePayload.books[0];
        ${firstRead}
        return intakePayload;
      })();
    `);
  for (const mode of ['text', 'comic', 'video']) {
    const calls = [];
    const settings = batchCreationSettings({ productionMode: mode, contentRange: mode === 'text' ? 8000 : 4000 });
    const intake = await run(mode, settings, mode === 'text' ? 4000 : 8000, selected,
      buildGiantMaterialPlaceholderIntakes, (...args) => calls.push(args));
    assert.deepEqual(calls, [['batch', '123', { maxTxt: 8000 }]], mode);
    assert.equal(intake.books[0].sourceMetadata.contentCaptureCharacters, 8000, mode);
    assert.equal(intake.metadata.contentCaptureCharacters, 8000, mode);
  }
});

test('giant jobs bind against the latest persisted book revision', () => {
  const source = read('BatchFactoryCreateModal.jsx');
  assert.match(source, /getBatch\(batchId\)/);
  assert.match(source, /findRegisteredGiantMaterialBook\(latestBatch\?\.books/);
  assert.match(source, /executorJobId: createdJob\.id/);
  assert.match(source, /giantAutomationPlan/);
});

test('giant OCR progress is displayed in the production-content editor', () => {
  const source = read('BatchFactoryNovelList.jsx');
  const editor = source.slice(source.indexOf('编辑生产内容'));
  assert.match(editor, /BatchFactoryGiantMaterialPendingProgress/);
  assert.match(editor, /onContentReady=\{refreshBatch\}/);
  assert.match(read('BatchFactoryGiantMaterialPendingProgress.jsx'), /startBatchAutomation/);
});

test('giant original reads fall back safely and expose both recovery actions', () => {
  const progress = read('BatchFactoryGiantMaterialPendingProgress.jsx');
  assert.match(progress, /originalReadStrategy/);
  assert.match(progress, /originalReadStrategy === 'ocr_first'/);
  assert.match(progress, /自动改用书城获取正文/);
  assert.match(progress, /获取正文/);
  assert.match(progress, />OCR</);
  assert.match(progress, /createGiantMaterialJob/);
  assert.match(progress, /persistedOriginalReadError/);
});

test('direct-first reports a failed book-city read without automatically queueing OCR', () => {
  const source = read('BatchFactoryCreateModal.jsx');
  assert.match(source, /await fetchBookOriginal\(batchId, book\.id, \{ maxTxt: contentCaptureCharacters \}\)/);
  assert.match(source, /\{ contentRangeLines, contentCaptureCharacters, importedAt, originalReadStrategy: giantOriginalReadStrategy \}/);
  const directReadBlock = source.slice(
    source.indexOf("if (giantOriginalReadStrategy === 'direct_first')"),
    source.indexOf('const durationSeconds', source.indexOf("if (giantOriginalReadStrategy === 'direct_first')"))
  );
  assert.match(directReadBlock, /originalReadStage: 'failed'/);
  assert.match(directReadBlock, /originalReadError: directReadError/);
  assert.doesNotMatch(directReadBlock, /queueGiantOcrFallback/);
  assert.doesNotMatch(directReadBlock, /自动转为滚屏 OCR/);
  assert.match(source, /await startBatchAutomation\(batchId, giantAutomationPlan\)/);
});

test('a giant batch does not start automation when a direct source read fails', () => {
  const source = read('BatchFactoryCreateModal.jsx');
  const directReadBlock = source.slice(
    source.indexOf("if (giantOriginalReadStrategy === 'direct_first')"),
    source.indexOf('const durationSeconds', source.indexOf("if (giantOriginalReadStrategy === 'direct_first')"))
  );
  const dispatchLoop = source.indexOf('for (const entry of selected)');
  const automationStart = source.indexOf('await startBatchAutomation(batchId, giantAutomationPlan)');

  assert.match(source, /let hasDirectReadFailure = false/);
  assert.match(directReadBlock, /hasDirectReadFailure = true/);
  assert.match(directReadBlock, /contentPending: false/);
  assert.match(source, /if \(giantAutomationPlan\?\.presetId && !hasDirectReadFailure\)/);
  assert.doesNotMatch(directReadBlock, /await startBatchAutomation\(batchId, giantAutomationPlan\)/);
  assert.ok(automationStart > dispatchLoop);
  assert.equal((source.match(/await startBatchAutomation\(batchId, giantAutomationPlan\)/g) || []).length, 1);
});
