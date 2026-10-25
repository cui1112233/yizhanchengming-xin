const REFERENCE_KINDS = Object.freeze({
  text: ['image', 'video', 'audio'],
  image: ['image'],
  video: ['image', 'video'],
  audio: ['audio']
});

export function referenceKindsForPromptNode(kind) {
  return [...(REFERENCE_KINDS[String(kind || '').trim()] || [])];
}

export function assetMentionTrigger(value, caret) {
  const text = String(value || '');
  const end = Math.max(0, Math.min(Number.isFinite(caret) ? caret : text.length, text.length));
  const start = text.lastIndexOf('@', end - 1);
  if (start < 0) return null;
  const query = text.slice(start + 1, end);
  if (/[\r\n]/u.test(query) || query.length > 80) return null;
  const emailLocalPart = text.slice(0, start).match(/[A-Za-z0-9.!#$%&'*+/=?^_`{|}~-]+$/u)?.[0] || '';
  const completeEmailDomain = /^[A-Za-z0-9-]+(?:\.[A-Za-z0-9-]+)+$/u.test(query);
  if (emailLocalPart && completeEmailDomain) return null;
  return { start, end, query };
}

export function filterMentionAssets(assets, kinds, query = '', limit = 8) {
  const allowed = new Set(Array.isArray(kinds) ? kinds : []);
  const needle = String(query || '').trim().toLocaleLowerCase();
  const seen = new Set();
  const matches = [];
  for (const asset of Array.isArray(assets) ? assets : []) {
    const id = String(asset?.id || asset?.assetId || '').trim();
    const kind = String(asset?.kind || '').trim();
    const haystack = `${asset?.name || ''} ${id}`.toLocaleLowerCase();
    if (!id || seen.has(id) || !allowed.has(kind) || (needle && !haystack.includes(needle))) continue;
    seen.add(id);
    matches.push(asset);
    if (matches.length >= limit) break;
  }
  return matches;
}

export function insertAssetMention(value, trigger, asset) {
  const text = String(value || '');
  const start = Math.max(0, Number(trigger?.start) || 0);
  const end = Math.max(start, Number(trigger?.end) || start);
  const name = String(asset?.name || asset?.id || asset?.assetId || '资产').trim();
  const label = `@${name}`;
  const suffix = text.slice(end).startsWith(' ') ? '' : ' ';
  const next = `${text.slice(0, start)}${label}${suffix}${text.slice(end)}`;
  return { value: next, caret: start + label.length + suffix.length, label };
}

export function firstClipboardImage(items) {
  for (const item of Array.from(items || [])) {
    if (item?.kind !== 'file' || !String(item.type || '').toLowerCase().startsWith('image/')) continue;
    const file = item.getAsFile?.();
    if (file) return file;
  }
  return null;
}

export function shouldHandleCanvasImagePaste({ hasImage, targetTagName = '', isContentEditable = false } = {}) {
  const tag = String(targetTagName || '').toUpperCase();
  return Boolean(hasImage && !isContentEditable && !['INPUT', 'TEXTAREA', 'SELECT'].includes(tag));
}

export function shouldCloseCanvasEditorOnEscape({ insideComposer = false, mentionOpen = false } = {}) {
  return !(insideComposer && mentionOpen);
}

export async function persistClipboardImage({ file, taskId, position, upload, add } = {}) {
  if (!file || !String(file.type || '').toLowerCase().startsWith('image/')) throw new Error('剪贴板中没有可用图片');
  if (!String(taskId || '').trim() || typeof upload !== 'function' || typeof add !== 'function') throw new Error('截图粘贴条件不完整');
  const result = await upload(taskId, file);
  const asset = result?.asset;
  if (!asset?.id || asset.kind !== 'image') throw new Error('截图上传没有返回图片资产');
  add(asset, position || { x: 0, y: 0 });
  return asset;
}
