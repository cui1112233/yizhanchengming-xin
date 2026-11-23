const chunkBytes = 512 * 1024;
const maxBookBytes = 24 * 1024 * 1024;
const maxTotalBytes = 48 * 1024 * 1024;
const maxFieldBytes = 2 ** 24 - 1;
const encoder = new TextEncoder();
let sessionNonce;

function hex(bytes) {
  return Array.from(bytes, value => value.toString(16).padStart(2, '0')).join('');
}

// A reload in the same tab can repeat the first request after losing its reply.
// Keep only the random session nonce in the browser, never a book's contents.
async function bookUploadId(platform, bookId, bytes) {
  if (!sessionNonce) {
    sessionNonce = hex(crypto.getRandomValues(new Uint8Array(24)));
    try {
      const key = 'batch-factory.intake-session';
      sessionNonce = globalThis.sessionStorage?.getItem(key) || sessionNonce;
      globalThis.sessionStorage?.setItem(key, sessionNonce);
    } catch {
      // In a storage-restricted tab, in-memory retries still share the nonce.
    }
  }
  const digest = hex(new Uint8Array(await crypto.subtle.digest('SHA-256', bytes)));
  const identity = encoder.encode(JSON.stringify([sessionNonce, platform, bookId, digest]));
  return hex(new Uint8Array(await crypto.subtle.digest('SHA-256', identity)).subarray(0, 24));
}

function base64(bytes) {
  let binary = '';
  for (let i = 0; i < bytes.length; i += 8192) binary += String.fromCharCode(...bytes.subarray(i, i + 8192));
  return btoa(binary);
}

export async function stageLargeIntake(payload, putChunk) {
  if (!payload || new Blob([JSON.stringify(payload)]).size <= 16 * 1024 * 1024) return payload;
  const staged = { ...payload };
  let total = 0;
  async function stageBook(platform, bookId, source) {
    const largestFieldBytes = Math.max(0, ...Object.values(source).map(value => typeof value === 'string' ? new Blob([value]).size : 0));
    if (largestFieldBytes > maxFieldBytes) throw new Error(`原文和生产稿每项最多 ${maxFieldBytes} 个 UTF-8 字节`);
    const bytes = encoder.encode(JSON.stringify(source));
    total += bytes.length;
    if (bytes.length > maxBookBytes || total > maxTotalBytes) throw new Error('正文超过暂存容量（单书 24 MiB、每批 48 MiB），请减少本批书数或单书内容');
    let ref = '';
    const uploadId = await bookUploadId(platform, bookId, bytes);
    for (let offset = 0; offset < bytes.length; offset += chunkBytes) {
      const result = await putChunk({ platform: String(platform || ''), bookId, bytes: bytes.length,
        index: offset / chunkBytes, ref, uploadId, maxFieldBytes: largestFieldBytes, data: base64(bytes.subarray(offset, offset + chunkBytes)) });
      if (!result?.ref || (ref && result.ref !== ref)) throw new Error('正文暂存失败，请重新提交');
      ref = result.ref;
    }
    return ref;
  }
  async function stageGroup(group) {
    const { sourceTextByBookId = {}, rawSourceTextByBookId = {}, ...rest } = group;
    const refs = {};
    for (const id of new Set([...Object.keys(sourceTextByBookId), ...Object.keys(rawSourceTextByBookId)])) {
      refs[id] = await stageBook(group.platformId, id, { sourceText: sourceTextByBookId[id], originalRaw: rawSourceTextByBookId[id] });
    }
    return { ...rest, sourceRefsByBookId: refs };
  }
  if (Array.isArray(payload.books)) {
    staged.books = [];
    for (const book of payload.books) {
      const { sourceText, originalRaw, txtText, ...rest } = book;
      const sourceRef = await stageBook(book.platform, String(book.bookId || book.id || ''), { sourceText, originalRaw, txtText });
      staged.books.push({ ...rest, sourceRef });
    }
    return staged;
  }
  const manifest = await stageGroup(staged);
  if (Array.isArray(payload.groups)) {
    manifest.groups = [];
    for (const group of payload.groups) manifest.groups.push(await stageGroup(group));
  }
  return manifest;
}
