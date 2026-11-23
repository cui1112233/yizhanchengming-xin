import test from 'node:test';
import assert from 'node:assert/strict';
import {
  availableGiantMaterialBooks,
  buildGiantMaterialIntake,
  buildGiantMaterialPlaceholderIntake,
  buildGiantMaterialPlaceholderIntakes,
  findRegisteredGiantMaterialBook,
  giantMaterialClassificationState,
  giantMaterialBookKey,
  giantMaterialSourceLabel,
  selectGiantMaterialBook
} from './batchFactoryGiantMaterialImport.js';

const material = {
  materialId: '10122315',
  videoUrl: 'https://material.hnqingyuwen.top/a.mp4',
  books: [
    { platformBookId: '748725', platformName: '七猫', title: '事不过三，过三遭殃' },
    { platformBookId: '1240820', platformName: '七猫', title: '事不过三，过三遭殃（重版）' }
  ]
};

test('keeps all platform book candidates and does not choose an ambiguous one', () => {
  assert.equal(availableGiantMaterialBooks(material).length, 2);
  assert.equal(selectGiantMaterialBook(material), null);
  assert.equal(selectGiantMaterialBook(material, giantMaterialBookKey(material.books[1])).platformBookId, '1240820');
});

test('builds a source-faithful giant material intake', () => {
  const payload = buildGiantMaterialIntake({
    giantMaterialId: '7689285397448523826',
    material,
    book: material.books[0],
    extraction: { text: '前一分钟字幕正文', sourceCompleteness: 'video_excerpt', requiresProofreading: true },
    importedAt: '2026-09-29T00:00:00.000Z'
  });
  assert.equal(payload.books[0].bookId, '748725');
  assert.equal(payload.books[0].title, '事不过三，过三遭殃');
  assert.equal(payload.books[0].sourceText, '前一分钟字幕正文');
  assert.equal(payload.books[0].sourceMetadata.sourceMode, 'giant_material');
  assert.equal(payload.books[0].sourceMetadata.sourceLabel, '巨量素材 · 7689285397448523826');
  assert.equal(payload.books[0].sourceMetadata.sourceContentVersion, 'giant_material:7689285397448523826');
  assert.equal(payload.metadata.giantMaterialId, '7689285397448523826');
});

test('stores the selected content range with the giant material intake', () => {
  const payload = buildGiantMaterialIntake({
    giantMaterialId: '7689285397448523826',
    material,
    book: material.books[0],
    extraction: { text: '正文' },
    contentRangeLines: 12
  });
  assert.equal(payload.books[0].sourceMetadata.contentRangeLines, 12);
  assert.equal(payload.metadata.contentRangeLines, 12);
});

test('preserves Qingyu platform label, raw code, and verified target platform ID', () => {
  const changdu = {
    materialId: '10122316',
    videoUrl: 'https://material.hnqingyuwen.top/a.mp4',
    books: [{ platformBookId: '7632217270088895550', platformName: '常读', platformCode: 'CD', platformId: '2', title: '常读测试书' }]
  };
  const payload = buildGiantMaterialIntake({
    giantMaterialId: '7683728935785873458',
    material: changdu,
    book: changdu.books[0],
    extraction: { text: '正文' }
  });
  assert.equal(payload.books[0].platform, '常读');
  assert.equal(payload.books[0].sourceMetadata.platformCode, 'CD');
  assert.equal(payload.books[0].sourceMetadata.platformId, '2');
});

test('rejects missing platform identity or empty OCR output', () => {
  assert.throws(() => buildGiantMaterialIntake({ giantMaterialId: '7689285397448523826', material: { title: '书' }, extraction: { text: '正文' } }), { message: 'QINGYU_BOOK_METADATA_INCOMPLETE' });
  assert.throws(() => buildGiantMaterialIntake({ giantMaterialId: '7689285397448523826', material, book: material.books[0], extraction: { text: '' } }), { message: 'GIANT_OCR_EMPTY' });
});

test('finds an existing import by giant material ID for idempotent retry', () => {
  const existing = { id: 'book-1', sourceMetadata: { giantMaterialId: '7689285397448523826' } };
  assert.equal(findRegisteredGiantMaterialBook([existing], '7689285397448523826'), existing);
  assert.equal(findRegisteredGiantMaterialBook([existing], '7689285397448523827'), null);
  assert.equal(giantMaterialSourceLabel('bad-id'), '巨量素材');
});

test('maps persisted AI classification state instead of leaving a registered book pending', () => {
  assert.deepEqual(giantMaterialClassificationState({ sourceMetadata: { classifyStatus: 'classified' } }), {
    status: 'classified',
    label: 'AI 已判断',
    error: ''
  });
  assert.deepEqual(giantMaterialClassificationState({ sourceMetadata: { classifyStatus: 'failed', classifyError: 'TEXT_MODEL_REQUIRED' } }), {
    status: 'failed',
    label: 'AI 判断失败',
    error: 'TEXT_MODEL_REQUIRED'
  });
  assert.deepEqual(giantMaterialClassificationState({ sourceMetadata: {} }), {
    status: 'pending',
    label: '待 AI 判断',
    error: ''
  });
});

test('records OCR-first as the default giant placeholder strategy', () => {
  const payload = buildGiantMaterialPlaceholderIntake({
    giantMaterialId: '7689285397448523826', material, book: material.books[0]
  });
  assert.equal(payload.books[0].sourceMetadata.originalReadStrategy, 'ocr_first');
  assert.equal(payload.books[0].sourceMetadata.originalReadStage, 'pending');
  assert.equal(payload.books[0].sourceMetadata.originalReadError, '');
  assert.equal(payload.books[0].sourceMetadata.contentPending, true);
});

test('records direct-first for every book in a multi-ID giant intake', () => {
  const payload = buildGiantMaterialPlaceholderIntakes([
    { giantMaterialId: '7689285397448523826', material, book: material.books[0] },
    { giantMaterialId: '7689285397448523827', material, book: material.books[1] }
  ], { originalReadStrategy: 'direct_first', contentCaptureCharacters: 8000 });
  assert.equal(payload.metadata.contentCaptureCharacters, 8000);
  assert.deepEqual(payload.books.map(book => book.sourceMetadata.contentCaptureCharacters), [8000, 8000]);
  assert.deepEqual(payload.books.map(book => book.sourceMetadata.originalReadStrategy), ['direct_first', 'direct_first']);
});

test('keeps video duration on a placeholder so OCR fallback can be started later', () => {
  const payload = buildGiantMaterialPlaceholderIntake({
    giantMaterialId: '7689285397448523826',
    material: { ...material, durationSeconds: 123 },
    book: material.books[0]
  });
  assert.equal(payload.books[0].sourceMetadata.videoDurationSeconds, 123);
});
