const VERSION_LIMIT = 30

export function workbenchStorageKey(projectId, bookId) {
  return 'ycm:script-workbench:' + String(projectId) + ':' + String(bookId)
}

function safeParse(raw) {
  try {
    const value = JSON.parse(raw || '{}')
    return value && typeof value === 'object' ? value : {}
  } catch {
    return {}
  }
}

export function loadWorkbenchRecord(storage, projectId, bookId) {
  if (!storage) return { draft: null, versions: [] }
  const value = safeParse(storage.getItem(workbenchStorageKey(projectId, bookId)))
  return {
    draft: value.draft && typeof value.draft === 'object' ? value.draft : null,
    versions: Array.isArray(value.versions) ? value.versions : [],
  }
}

export function saveWorkbenchDraft(storage, projectId, bookId, draft) {
  if (!storage) return
  const current = loadWorkbenchRecord(storage, projectId, bookId)
  storage.setItem(workbenchStorageKey(projectId, bookId), JSON.stringify({ ...current, draft }))
}

export function saveWorkbenchVersion(storage, projectId, bookId, draft, label = '手工保存') {
  if (!storage) return []
  const current = loadWorkbenchRecord(storage, projectId, bookId)
  const version = {
    id: String(Date.now()) + '-' + String(current.versions.length + 1),
    label,
    savedAt: new Date().toISOString(),
    draft,
  }
  const versions = [version, ...current.versions].slice(0, VERSION_LIMIT)
  storage.setItem(workbenchStorageKey(projectId, bookId), JSON.stringify({ draft, versions }))
  return versions
}

export function splitShotCards(value) {
  const text = String(value || '').trim()
  if (!text) return []
  const matches = [...text.matchAll(/^###\s*分镜[^\n]*$/gm)]
  if (matches.length === 0) return [{ title: '分镜一', body: text }]
  return matches.map((match, index) => {
    const start = match.index + match[0].length
    const end = index + 1 < matches.length ? matches[index + 1].index : text.length
    return {
      title: match[0].replace(/^###\s*/, '').trim() || '分镜' + String(index + 1),
      body: text.slice(start, end).replace(/^\s*---\s*/m, '').trim(),
    }
  })
}

export function joinShotCards(cards) {
  return (cards || [])
    .map((card, index) => '### ' + (card.title || '分镜' + String(index + 1)) + '\n' + String(card.body || '').trim())
    .join('\n\n---\n\n')
}

export function replaceInText(value, find, replacement, replaceAll = true) {
  const source = String(value || '')
  const needle = String(find || '')
  if (!needle) return source
  if (!replaceAll) return source.replace(needle, String(replacement || ''))
  return source.split(needle).join(String(replacement || ''))
}
