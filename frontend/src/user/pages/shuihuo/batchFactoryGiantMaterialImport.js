import { normalizeGiantMaterialId } from '../giantMaterialTest.js';

function text(value) {
  return String(value ?? '').trim();
}

function originalReadStrategy(value) {
  return text(value) === 'direct_first' ? 'direct_first' : 'ocr_first';
}

export function giantMaterialSourceLabel(value) {
  const id = normalizeGiantMaterialId(value);
  return id ? `巨量素材 · ${id}` : '巨量素材';
}

export function giantMaterialBookKey(book = {}) {
  return `${text(book.platformName || book.platform)}:${text(book.platformBookId || book.bookId)}`;
}

export function availableGiantMaterialBooks(material = {}) {
  const listed = Array.isArray(material.books) ? material.books : [];
  if (listed.length) return listed.filter(book => text(book.platformBookId || book.bookId) && text(book.title));
  const platformBookId = text(material.platformBookId);
  const title = text(material.title);
  if (!platformBookId || !title) return [];
  return [{ platformBookId, platformName: text(material.platformName), title }];
}

export function selectGiantMaterialBook(material, requestedKey = '') {
  const books = availableGiantMaterialBooks(material);
  if (books.length === 1) return books[0];
  const key = text(requestedKey);
  return books.find(book => giantMaterialBookKey(book) === key) || null;
}

export function findRegisteredGiantMaterialBook(books, giantMaterialId) {
  const id = normalizeGiantMaterialId(giantMaterialId);
  if (!id) return null;
  return (Array.isArray(books) ? books : []).find(book => text(book?.sourceMetadata?.giantMaterialId) === id) || null;
}

export function giantMaterialClassificationState(book = {}) {
  const metadata = book?.sourceMetadata || {};
  const status = text(metadata.classifyStatus).toLowerCase();
  if (status === 'classified' || status === 'manual') return { status: 'classified', label: 'AI 已判断', error: '' };
  if (status === 'failed') return { status: 'failed', label: 'AI 判断失败', error: text(metadata.classifyError) };
  return { status: 'pending', label: '待 AI 判断', error: '' };
}

export function buildGiantMaterialIntake({ giantMaterialId, material, extraction, book, contentRangeLines = 5, importedAt = new Date().toISOString() } = {}) {
  const normalizedID = normalizeGiantMaterialId(giantMaterialId);
  const selected = book || selectGiantMaterialBook(material);
  const sourceText = text(extraction?.text);
  const platformBookId = text(selected?.platformBookId || selected?.bookId);
  const title = text(selected?.title || material?.title);
  const platformName = text(selected?.platformName || selected?.platform || material?.platformName);
  const platformCode = text(selected?.platformCode || material?.platformCode);
  const platformId = text(selected?.platformId || material?.platformId);
  if (!normalizedID) throw new Error('INVALID_GIANT_MATERIAL_ID');
  if (!selected || !platformBookId || !title) throw new Error('QINGYU_BOOK_METADATA_INCOMPLETE');
  if (!sourceText) throw new Error('GIANT_OCR_EMPTY');
  const sourceLabel = giantMaterialSourceLabel(normalizedID);
  const sourceContentVersion = `giant_material:${normalizedID}`;
  const rangeLines = Math.min(500, Math.max(1, Number(contentRangeLines) || 5));
  return {
    books: [{
      id: platformBookId,
      bookId: platformBookId,
      sourceTaskId: `giant-material:${normalizedID}`,
      title,
      platform: platformName,
      sourceText,
      txtText: sourceText,
      txtFileName: `${platformBookId}.txt`,
      sourceMetadata: {
        sourceMode: 'giant_material',
        sourceLabel,
        giantMaterialId: normalizedID,
        qingyuMaterialId: text(material?.materialId),
        videoUrl: text(material?.videoUrl),
        videoDurationSeconds: Number(material?.durationSeconds || material?.duration || 0),
        platformName,
        platformCode,
        platformId,
        platformBookId,
        sourceBookTitle: title,
        sourceContentVersion,
        sourceImportedAt: text(importedAt) || new Date().toISOString(),
        sourceCompleteness: text(extraction?.sourceCompleteness || 'video_excerpt'),
        contentRangeLines: rangeLines,
        requiresProofreading: extraction?.requiresProofreading !== false
      }
    }],
    metadata: {
      sourceMode: 'giant_material',
      sourceLabel,
      giantMaterialId: normalizedID,
      qingyuMaterialId: text(material?.materialId),
      sourceContentVersion,
      contentRangeLines: rangeLines,
      sourceImportedAt: text(importedAt) || new Date().toISOString()
    }
  };
}

// 占位登记：解析成功后立刻把书登记进批量（正文为空），Windows 执行器读取完成后再回填正文。
// 这样新书会马上出现在制作区，用户不用等 OCR 读完才看到书。
export function buildGiantMaterialPlaceholderIntake({ giantMaterialId, material, book, contentRangeLines = 5, contentCaptureCharacters = 4000, importedAt = new Date().toISOString(), originalReadStrategy: requestedStrategy } = {}) {
  const normalizedID = normalizeGiantMaterialId(giantMaterialId);
  const selected = book || selectGiantMaterialBook(material);
  const platformBookId = text(selected?.platformBookId || selected?.bookId);
  const title = text(selected?.title || material?.title);
  const platformName = text(selected?.platformName || selected?.platform || material?.platformName);
  const platformCode = text(selected?.platformCode || material?.platformCode);
  const platformId = text(selected?.platformId || material?.platformId);
  if (!normalizedID) throw new Error('INVALID_GIANT_MATERIAL_ID');
  if (!selected || !platformBookId || !title) throw new Error('QINGYU_BOOK_METADATA_INCOMPLETE');
  const sourceLabel = giantMaterialSourceLabel(normalizedID);
  const sourceContentVersion = `giant_material:${normalizedID}`;
  const rangeLines = Math.min(500, Math.max(1, Number(contentRangeLines) || 5));
  const strategy = originalReadStrategy(requestedStrategy);
  return {
    books: [{
      id: platformBookId,
      bookId: platformBookId,
      sourceTaskId: `giant-material:${normalizedID}`,
      title,
      platform: platformName,
      sourceText: '',
      txtText: '',
      txtFileName: `${platformBookId}.txt`,
      sourceMetadata: {
        sourceMode: 'giant_material',
        sourceLabel,
        giantMaterialId: normalizedID,
        qingyuMaterialId: text(material?.materialId),
        videoUrl: text(material?.videoUrl),
        videoDurationSeconds: Number(material?.durationSeconds || material?.duration || 0),
        platformName,
        platformCode,
        platformId,
        platformBookId,
        sourceBookTitle: title,
        sourceContentVersion,
        sourceImportedAt: text(importedAt) || new Date().toISOString(),
        originalReadStrategy: strategy,
        originalReadStage: 'pending',
        originalReadError: '',
        contentPending: true,
        contentRangeLines: rangeLines,
        contentCaptureCharacters
      }
    }],
    metadata: {
      sourceMode: 'giant_material',
      sourceLabel,
      giantMaterialId: normalizedID,
      qingyuMaterialId: text(material?.materialId),
      sourceContentVersion,
      contentPending: true,
      contentRangeLines: rangeLines,
      contentCaptureCharacters,
      sourceImportedAt: text(importedAt) || new Date().toISOString()
    }
  };
}

// 多本占位登记：一次解析出的 N 条素材合成一个 intake（books 数组），
// createBatchFromIntake 一次性建出含 N 本占位书的批量。
export function buildGiantMaterialPlaceholderIntakes(entries, { contentRangeLines = 5, contentCaptureCharacters = 4000, importedAt = new Date().toISOString(), originalReadStrategy: requestedStrategy } = {}) {
  const list = (Array.isArray(entries) ? entries : []).filter(Boolean);
  if (!list.length) throw new Error('GIANT_MATERIAL_INTAKE_EMPTY');
  const books = list.map(entry => buildGiantMaterialPlaceholderIntake({
    giantMaterialId: entry.giantMaterialId,
    material: entry.material,
    book: entry.book,
    contentRangeLines,
    contentCaptureCharacters,
    importedAt,
    originalReadStrategy: requestedStrategy
  }).books[0]);
  return {
    books,
    metadata: {
      sourceMode: 'giant_material',
      sourceLabel: `巨量素材 · ${books.length} 本`,
      contentPending: true,
      contentRangeLines: Math.min(500, Math.max(1, Number(contentRangeLines) || 5)),
      contentCaptureCharacters,
      sourceImportedAt: importedAt
    }
  };
}
