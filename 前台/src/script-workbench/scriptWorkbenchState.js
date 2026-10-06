const VERSION_LIMIT = 30
const COMPILER_SUFFIX_SECTION = /^\s*(?:HOOK|DIRECTOR|PROCESSING RULES|KNOWLEDGE BASE|PROJECT CONFIG|USER CONFIG|MODEL CONFIG):\s*$/gm

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

function suffixStartAfter(text, offset) {
  COMPILER_SUFFIX_SECTION.lastIndex = offset
  const match = COMPILER_SUFFIX_SECTION.exec(text)
  COMPILER_SUFFIX_SECTION.lastIndex = 0
  return match ? match.index : text.length
}

export function splitShotDocument(value) {
  const text = String(value || '')
  if (!text.trim()) return { prefix: '', cards: [], suffix: '' }

  const allMatches = [...text.matchAll(/^###\s*分镜[^\n]*$/gm)]
  if (allMatches.length === 0) {
    return { prefix: '', cards: [{ title: '分镜一', body: text.trim() }], suffix: '' }
  }

  const first = allMatches[0]
  const suffixStart = suffixStartAfter(text, first.index + first[0].length)
  const matches = allMatches.filter((match) => match.index < suffixStart)
  const prefix = text.slice(0, first.index).trimEnd()
  const suffix = suffixStart < text.length ? text.slice(suffixStart).trimStart() : ''

  const cards = matches.map((match, index) => {
    const start = match.index + match[0].length
    const end = index + 1 < matches.length ? matches[index + 1].index : suffixStart
    return {
      title: match[0].replace(/^###\s*/, '').trim() || '分镜' + String(index + 1),
      body: text.slice(start, end).replace(/^\s*---\s*/m, '').trim(),
    }
  })
  return { prefix, cards, suffix }
}

export function joinShotDocument(document) {
  const prefix = String(document?.prefix || '').trim()
  const suffix = String(document?.suffix || '').trim()
  const cards = Array.isArray(document?.cards) ? document.cards : []
  const cardText = cards
    .map((card, index) => '### ' + (card.title || '分镜' + String(index + 1)) + '\n' + String(card.body || '').trim())
    .join('\n\n---\n\n')
  return [prefix, cardText, suffix].filter(Boolean).join('\n\n')
}

export function splitShotCards(value) {
  return splitShotDocument(value).cards
}

export function joinShotCards(cards) {
  return joinShotDocument({ cards })
}

export function replaceInText(value, find, replacement, replaceAll = true) {
  const source = String(value || '')
  const needle = String(find || '')
  if (!needle) return source
  if (!replaceAll) return source.replace(needle, String(replacement || ''))
  return source.split(needle).join(String(replacement || ''))
}

export function workbenchExportArtifact({ editorOutput, sourceText, bookTitle, outputMode }) {
  return {
    filename: (bookTitle || '剧本') + '-' + (outputMode || 'canvas') + '.txt',
    text: String(editorOutput || sourceText || ''),
  }
}
