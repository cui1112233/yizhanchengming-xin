const bookIDPattern = /^\d{6,25}$/;

// 一行里如果所有字段都是纯书籍 ID，说明是从表格里横着复制的一串 ID，
// 每个 ID 都是一本书；只要出现书名等非数字字段，就仍按“ID + 书名 + 字段”解析。
function horizontalBookIDs(line) {
  const parts = String(line || '').trim().split(/[\s,，;；]+/).filter(Boolean);
  return parts.length > 1 && parts.every(part => bookIDPattern.test(part)) ? parts : null;
}

export function manualBookIDsFromInput(inputText = '') {
  const seen = new Set();
  const ids = [];
  for (const line of String(inputText).split(/\r?\n/)) {
    const lineIDs = horizontalBookIDs(line) || [line.match(/^\s*(\d{6,25})(?=\s|$)/)?.[1]].filter(Boolean);
    for (const id of lineIDs) {
      if (!seen.has(id)) {
        seen.add(id);
        ids.push(id);
      }
    }
  }
  return ids;
}

export function manualFetchedSourceText(value, depth = 0) {
  if (typeof value === 'string') {
    const text = value.trim();
    return text === '[object Object]' ? '' : text;
  }
  if (!value || typeof value !== 'object' || Array.isArray(value) || depth >= 3) return '';
  for (const key of ['processedText', 'sourceText', 'text', 'content', 'data', 'rawData']) {
    const text = manualFetchedSourceText(value[key], depth + 1);
    if (text) return text;
  }
  return '';
}

function normalizeSourceTextMap(sourceTextByBookId) {
  if (!sourceTextByBookId || typeof sourceTextByBookId !== 'object' || Array.isArray(sourceTextByBookId)) return {};
  return Object.fromEntries(Object.entries(sourceTextByBookId).flatMap(([bookId, value]) => {
    const text = manualFetchedSourceText(value);
    return text ? [[bookId, text]] : [];
  }));
}

export function hasFetchedManualSources(bookIds, sourceTextByBookId) {
  return bookIds.length > 0 && bookIds.every(bookId => manualFetchedSourceText(sourceTextByBookId?.[bookId]));
}

export function buildManualBatchSubmission({ title, scheduledAt = '', automationEnabled = false, autoPublishEnabled = false, automationConcurrency = 2, runMode = '', ...input }) {
  const enabled = automationEnabled === true;
  const normalizedRunMode = String(runMode || '').trim();
  const groups = Array.isArray(input.groups)
    ? input.groups.map(group => ({ ...group, sourceTextByBookId: normalizeSourceTextMap(group?.sourceTextByBookId) }))
    : input.groups;
  return {
    ...input,
    ...(groups ? { groups } : {}),
    ...(input.sourceTextByBookId ? { sourceTextByBookId: normalizeSourceTextMap(input.sourceTextByBookId) } : {}),
    title: String(title || '').trim(),
    scheduledAt: String(scheduledAt || ''),
    runMode: normalizedRunMode,
    automationEnabled: enabled,
    autoPublishEnabled: enabled && (autoPublishEnabled === true || normalizedRunMode === 'full_submit'),
    automationConcurrency: [1, 2, 4].includes(Number(automationConcurrency)) ? Number(automationConcurrency) : 2
  };
}

function nonEmptyLines(text) {
  return String(text || '').split(/\r?\n/).map(line => line.trimEnd()).filter(line => line.trim());
}

function expandedBookLines(text) {
  const lines = [];
  for (const line of nonEmptyLines(text)) {
    const horizontal = horizontalBookIDs(line);
    if (horizontal) lines.push(...horizontal);
    else lines.push(line);
  }
  return lines;
}

export function mergeGroupLines(oldText = '', newText = '') {
  const byId = new Map();
  const order = [];
  for (const line of expandedBookLines(oldText)) {
    const id = line.match(/^\s*(\d{6,25})(?=\s|$)/)?.[1];
    if (!id) continue;
    if (!byId.has(id)) order.push(id);
    byId.set(id, line);
  }
  for (const line of expandedBookLines(newText)) {
    const id = line.match(/^\s*(\d{6,25})(?=\s|$)/)?.[1];
    if (!id) continue;
    if (!byId.has(id)) order.push(id);
    byId.set(id, line); // 新行优先
  }
  return order.map(id => byId.get(id)).join('\n');
}

export function upsertPlatformGroup(groups = [], { platformId, platformName, inputText } = {}) {
  const id = String(platformId || '').trim();
  const name = String(platformName || '').trim();
  const text = String(inputText || '').trim();
  if (!id || !manualBookIDsFromInput(text).length) return groups;
  const existing = groups.find(group => String(group.platformId) === id);
  if (existing) {
    return groups.map(group => String(group.platformId) === id
      ? { ...group, platformName: name || group.platformName, inputText: mergeGroupLines(group.inputText, text) }
      : group);
  }
  return [...groups, { platformId: id, platformName: name, inputText: text }];
}

export function replacePlatformGroup(groups = [], platformId, inputText = '') {
  const id = String(platformId || '').trim();
  const text = String(inputText || '').trim();
  if (!text || !manualBookIDsFromInput(text).length) {
    return groups.filter(group => String(group.platformId) !== id);
  }
  return groups.map(group => String(group.platformId) === id ? { ...group, inputText: text } : group);
}

export function removePlatformGroup(groups = [], platformId) {
  const id = String(platformId || '').trim();
  return groups.filter(group => String(group.platformId) !== id);
}

export function totalGroupBookCount(groups = []) {
  return groups.reduce((sum, group) => sum + manualBookIDsFromInput(group.inputText).length, 0);
}
