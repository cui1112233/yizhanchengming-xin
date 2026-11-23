const chunkBytes = 512 * 1024;
const maxBookBytes = 24 * 1024 * 1024;
const maxTotalBytes = 48 * 1024 * 1024;
const encoder = new TextEncoder();

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
    const bytes = encoder.encode(JSON.stringify(source));
    total += bytes.length;
    if (bytes.length > maxBookBytes || total > maxTotalBytes) throw new Error('正文超过暂存容量（单书 24 MiB、每批 48 MiB），请减少本批书数或单书内容');
    let ref = '';
    for (let offset = 0; offset < bytes.length; offset += chunkBytes) {
      const result = await putChunk({ platform: String(platform || ''), bookId, bytes: bytes.length,
        index: offset / chunkBytes, ref, data: base64(bytes.subarray(offset, offset + chunkBytes)) });
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
