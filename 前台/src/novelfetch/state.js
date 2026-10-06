export function parseSensitiveReplacementLines(raw) {
  const result = {}
  for (const line of String(raw || '').split(/\r?\n/)) {
    const trimmed = line.trim()
    if (!trimmed) continue
    const index = trimmed.indexOf('=>')
    if (index <= 0) continue
    const from = trimmed.slice(0, index).trim()
    const to = trimmed.slice(index + 2).trim()
    if (from) result[from] = to
  }
  return result
}

export function formatSensitiveReplacementLines(value) {
  return Object.entries(value || {})
    .sort(([a], [b]) => b.length - a.length || a.localeCompare(b, 'zh-CN'))
    .map(([from, to]) => `${from}=>${to}`)
    .join('\n')
}

export function parseChapterPrefixLines(raw) {
  return [...new Set(String(raw || '').split(/\r?\n/).map((value) => value.trim()).filter(Boolean))]
}
