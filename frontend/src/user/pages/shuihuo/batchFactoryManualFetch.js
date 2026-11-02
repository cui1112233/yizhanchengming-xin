export function manualBookIDsFromInput(inputText = '') {
  const seen = new Set();
  const ids = [];
  for (const line of String(inputText).split(/\r?\n/)) {
    const id = line.match(/^\s*(\d{6,25})(?=\s|$)/)?.[1];
    if (!id) continue;
    if (!seen.has(id)) {
      seen.add(id);
      ids.push(id);
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

export function mergeGroupLines(oldText = '', newText = '') {
  const byId = new Map();
  const order = [];
  for (const line of nonEmptyLines(oldText)) {
    const id = line.match(/^\s*(\d{6,25})(?=\s|$)/)?.[1];
    if (!id) continue;
    if (!byId.has(id)) order.push(id);
    byId.set(id, line);
  }
  for (const line of nonEmptyLines(newText)) {
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
