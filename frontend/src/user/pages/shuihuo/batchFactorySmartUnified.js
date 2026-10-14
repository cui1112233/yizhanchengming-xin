const SMART_UNIFIED_PREFIX_PRESET_ID = 'script-constraint-prefix-smart-unified';

function object(value) {
  return value && typeof value === 'object' && !Array.isArray(value) ? value : {};
}

function savedAnalysisFromBook(book) {
  const patch = object(object(book?.settingsState).patch);
  let value = patch.h3StyleAnalysis;
  if (typeof value === 'string') {
    try { value = JSON.parse(value); } catch { return null; }
  }
  if (typeof value === 'string') {
    try { value = JSON.parse(value); } catch { return null; }
  }
  const analysis = object(value);
  return String(analysis.prompt || '').trim() ? analysis : null;
}

function effectivePromptConfig(batch, book) {
  const batchConfig = object(object(batch?.settingsState?.patch).aiPromptConfig);
  const bookConfig = object(object(book?.settingsState?.patch).aiPromptConfig);
  const batchConstraints = object(batchConfig.constraints);
  const bookConstraints = object(bookConfig.constraints);
  return {
    ...batchConfig,
    ...bookConfig,
    constraints: {
      ...batchConstraints,
      ...bookConstraints,
      selections: Array.isArray(bookConstraints.selections) ? bookConstraints.selections : batchConstraints.selections
    }
  };
}

export function smartUnifiedDisplayEnabled(batch, book) {
  const selections = effectivePromptConfig(batch, book).constraints?.selections;
  return Array.isArray(selections) && selections.some(selection => (
    selection?.presetId === SMART_UNIFIED_PREFIX_PRESET_ID && selection?.constraintCategory === 'prefix'
  ));
}

export function smartUnifiedAnalysisForBook(book) {
  // h3StyleAnalysis is the current book-scoped result. A director revision is
  // a frozen historical snapshot and must only be used when no current result
  // exists; otherwise a refresh would show two different unified styles.
  const saved = savedAnalysisFromBook(book);
  if (saved) return saved;
  const output = object(book?.directorRevision?.output);
  const analysis = object(output.smart_unified_analysis || output.smartUnifiedAnalysis);
  if (String(analysis.prompt || '').trim()) return analysis;
  const legacyPrompt = String(output.smart_unified_style || output.smartUnifiedStyle || '').trim();
  if (legacyPrompt) return { prompt: legacyPrompt, fields: {}, preset: {} };
  return null;
}

export function prependSmartUnifiedStyle(prompt, style) {
  const body = String(prompt || '').trim();
  const unified = String(style || '').trim();
  if (!unified || body.includes('【统一风格】')) return body;
  return `【统一风格】\n${unified}${body ? `\n\n${body}` : ''}`;
}
