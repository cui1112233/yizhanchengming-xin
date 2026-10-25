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

export function assetReferenceKey(asset) {
  return String(asset?.id || asset?.assetId || '').trim();
}

function domMentionIdentity(node) {
  const label = String(node?.dataset?.assetMentionValue || '').trim();
  if (!label) return null;
  return {
    assetKey: String(node?.dataset?.assetId || '').trim(),
    label
  };
}

export function assetMentionAtDOMCaret(anchorNode, anchorOffset, root, direction) {
  const offset = Math.max(0, Number(anchorOffset) || 0);
  if (anchorNode === root) {
    const candidate = direction === 'backspace'
      ? root?.childNodes?.[offset - 1]
      : direction === 'delete'
        ? root?.childNodes?.[offset]
        : null;
    return domMentionIdentity(candidate);
  }
  if (anchorNode?.nodeType !== 3) return null;
  const text = String(anchorNode.nodeValue || '');
  if (direction === 'backspace' && offset > 0 && /^[^\S\r\n]+$/u.test(text.slice(0, offset))) {
    return domMentionIdentity(anchorNode.previousSibling);
  }
  if (direction === 'delete' && offset < text.length && /^[^\S\r\n]+$/u.test(text.slice(offset))) {
    return domMentionIdentity(anchorNode.nextSibling);
  }
  return null;
}

function isMentionBoundary(character) {
  return !character || /[\s\p{P}\p{S}]/u.test(character);
}

export function promptMentionRanges(value, references) {
  const text = String(value || '');
  const seen = new Set();
  const candidates = [];
  for (const asset of Array.isArray(references) ? references : []) {
    const key = assetReferenceKey(asset) || `name:${String(asset?.name || '').trim()}`;
    const name = String(asset?.name || '').trim();
    if (!name || seen.has(key)) continue;
    seen.add(key);
    candidates.push({ asset, label: `@${name}` });
  }
  candidates.sort((left, right) => right.label.length - left.label.length);

  const ranges = [];
  for (let cursor = 0; cursor < text.length;) {
    if (text[cursor] !== '@') {
      cursor += 1;
      continue;
    }
    const match = candidates.find(candidate => (
      text.startsWith(candidate.label, cursor)
      && isMentionBoundary(text[cursor + candidate.label.length])
    ));
    if (!match) {
      cursor += 1;
      continue;
    }
    ranges.push({ start: cursor, end: cursor + match.label.length, label: match.label, asset: match.asset });
    cursor += match.label.length;
  }
  return ranges;
}

export function assetMentionDeletionAtCaret(value, references, caret, direction) {
  const text = String(value || '');
  const cursor = Math.max(0, Math.min(Number.isFinite(caret) ? caret : 0, text.length));
  return promptMentionRanges(text, references).find(range => {
    if (direction === 'delete') return cursor === range.start;
    if (direction !== 'backspace') return false;
    return cursor === range.end || (
      cursor === range.end + 1
      && /[^\S\r\n]/u.test(text[range.end] || '')
    );
  }) || null;
}

export function removeAssetMention(value, asset) {
  const text = String(value || '');
  const ranges = promptMentionRanges(text, [asset]);
  let next = text;
  for (const range of ranges.reverse()) {
    let start = range.start;
    let end = range.end;
    const before = next[start - 1] || '';
    const after = next[end] || '';
    if (/[^\S\r\n]/u.test(after) && (start === 0 || /[^\S\r\n]/u.test(before))) end += 1;
    else if (/[^\S\r\n]/u.test(before) && isMentionBoundary(after)) start -= 1;
    next = `${next.slice(0, start)}${next.slice(end)}`;
  }
  return next;
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
