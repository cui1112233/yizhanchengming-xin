export function updateSensitiveRule(rule, key, value) {
  return { ...rule, [key]: value, ...(key === 'replace' ? { replace_explicit: true } : {}) };
}

export function parseSensitiveBulkLines(text) {
  return String(text || '').split(/\r?\n/).map(line => line.trim()).filter(Boolean).map(line => {
    const separator = line.indexOf('=>');
    const find = (separator < 0 ? line : line.slice(0, separator)).trim();
    const replace = separator < 0 ? '' : line.slice(separator + 2).trim();
    return { find, replace, enabled: true, ...(separator < 0 ? {} : { replace_explicit: true }) };
  }).filter(rule => rule.find);
}
