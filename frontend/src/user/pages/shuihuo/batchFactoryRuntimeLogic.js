import { batchFactoryCleanSourceText, batchFactoryPreviewText, batchFactorySourceText } from './batchFactoryContentRange.js';

function normalizedText(value) {
  return String(value ?? '').trim();
}

export function resolveBookProductionText(book) {
  const working = batchFactorySourceText(book?.workingFrontContent);
  if (working) return working;
  return normalizedText(batchFactoryPreviewText(book?.sourceText, book));
}

export function resolveBookEditableContent(book, productionMode) {
  const working = batchFactorySourceText(book?.workingFrontContent);
  if (working) return working;
  if (productionMode === 'comic') return batchFactoryCleanSourceText(book?.sourceText);
  return resolveBookProductionText(book);
}

function stableHash(input) {
  let hash = 0x811c9dc5;
  for (let index = 0; index < input.length; index += 1) {
    hash ^= input.charCodeAt(index);
    hash = Math.imul(hash, 0x01000193);
  }
  return (hash >>> 0).toString(16).padStart(8, '0');
}

export function audioDurationFingerprint(book, tts = {}) {
  return stableHash(JSON.stringify({
    text: resolveBookProductionText(book),
    voice: normalizedText(tts.voice),
    style: normalizedText(tts.style),
    speed: Number(tts.speed ?? 1),
    pitch: Number(tts.pitch ?? 0)
  }));
}

export function batchMediaCounts(books = [], productionStatus = {}, mergeStatus = {}) {
  const storyboards = (Array.isArray(books) ? books : []).reduce(
    (sum, book) => sum + (Array.isArray(book?.videos) ? book.videos.length : 0),
    0
  );

  const playable = new Set();
  for (const job of productionStatus?.jobs || []) {
    for (const task of job?.tasks || []) {
      if (String(task?.status || '').toLowerCase() !== 'succeeded') continue;
      const mediaUrl = normalizedText(task?.mediaUrl);
      if (!mediaUrl) continue;
      playable.add(String(task?.id || `${task?.videoId || ''}:${mediaUrl}`));
    }
  }

  const merges = new Set();
  for (const job of mergeStatus?.jobs || []) {
    if (String(job?.status || '').toLowerCase() !== 'succeeded') continue;
    const outputUrl = normalizedText(job?.outputUrl);
    if (!outputUrl) continue;
    merges.add(String(job?.id || `${job?.bookId || ''}:${outputUrl}`));
  }

  return { storyboards, playable: playable.size, merges: merges.size };
}
