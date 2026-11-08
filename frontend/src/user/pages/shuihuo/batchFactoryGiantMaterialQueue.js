import { normalizeGiantMaterialId } from '../giantMaterialTest.js';

function text(value) {
  return String(value ?? '').trim();
}

export function parseGiantMaterialIds(input) {
  const valid = [];
  const invalid = [];
  const seen = new Set();
  const tokens = String(input ?? '').split(/[\s,，;；]+/).map(text).filter(Boolean);
  for (const token of tokens) {
    const id = normalizeGiantMaterialId(token);
    if (!id) {
      if (!invalid.includes(token)) invalid.push(token);
      continue;
    }
    if (!seen.has(id)) {
      seen.add(id);
      valid.push(id);
    }
  }
  return { valid, invalid };
}

export function createGiantMaterialQueue(input) {
  const ids = Array.isArray(input) ? input : parseGiantMaterialIds(input).valid;
  return ids.map((id, index) => ({
    id: text(id),
    index,
    status: 'pending',
    stage: 'idle',
    error: '',
    material: null,
    books: [],
    selectedBookKey: '',
    progress: null,
    result: null,
    registeredBook: null
  }));
}

export function duplicateGiantMaterialBookIndexes(entries = []) {
  const seen = new Set();
  const duplicates = new Set();
  for (const [index, entry] of entries.entries()) {
    const platformBookId = text(entry?.platformBookId || entry?.bookId);
    if (!platformBookId) continue;
    if (seen.has(platformBookId)) duplicates.add(index);
    else seen.add(platformBookId);
  }
  return duplicates;
}

export function createGiantMaterialPlatformBookClaims() {
  const ownerByBookID = new Map();
  return {
    claim(book = {}) {
      const platformBookId = text(book.platformBookId || book.bookId);
      const giantMaterialId = text(book.giantMaterialId);
      const duplicateOfMaterialId = platformBookId ? (ownerByBookID.get(platformBookId) || '') : '';
      if (platformBookId && !duplicateOfMaterialId) ownerByBookID.set(platformBookId, giantMaterialId);
      return {
        accepted: Boolean(platformBookId) && !duplicateOfMaterialId,
        platformBookId,
        duplicateOfMaterialId
      };
    }
  };
}

export function queueSummary(items = []) {
  const summary = { total: items.length, pending: 0, running: 0, success: 0, skipped: 0, error: 0 };
  for (const item of items) {
    if (Object.prototype.hasOwnProperty.call(summary, item?.status)) summary[item.status] += 1;
  }
  return summary;
}

function state(item, patch, onState) {
  Object.assign(item, patch);
  onState?.(item, patch);
}

export async function runSequentialGiantMaterialQueue(items, worker, { onState, signal } = {}) {
  for (const item of items) {
    if (!item || item.status === 'success' || item.status === 'skipped') continue;
    if (signal?.aborted) break;
    state(item, { status: 'running', stage: 'running', error: '' }, onState);
    try {
      const result = await worker(item, {
        signal,
        update: patch => state(item, patch, onState)
      });
      if (result?.status === 'skipped') state(item, { ...result, status: 'skipped', stage: 'complete', error: '' }, onState);
      else state(item, { ...result, status: 'success', stage: 'complete', error: '' }, onState);
    } catch (error) {
      if (signal?.aborted) break;
      const code = text(error?.message || error?.code || 'GIANT_MATERIAL_QUEUE_FAILED');
      state(item, { status: 'error', stage: 'error', error: code }, onState);
    }
  }
  return items;
}
