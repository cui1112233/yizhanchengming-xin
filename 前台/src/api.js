const API_PREFIX = '/api/v1'
const REFRESH_PATH = '/api/auth/refresh'
const AUTH_PATHS = new Set(['/api/auth/login', REFRESH_PATH, '/api/auth/logout'])
const AUTH_SESSION_HINT_COOKIE = 'ycm_auth_session_hint'
const AUTH_SESSION_HINT_MAX_AGE = 30 * 24 * 60 * 60

let refreshFlight = null
let authFailureNotified = false

export class APIError extends Error {
  constructor(message, { status = 0, code = 'API_ERROR', payload = null, requestId = '' } = {}) {
    super(message)
    this.name = 'APIError'
    this.status = status
    this.code = code
    this.payload = payload
    this.requestId = requestId
  }
}

async function parsePayload(response) {
  try {
    return await response.json()
  } catch {
    return null
  }
}

function responseRequestId(response, payload) {
  const fromHeader = response?.headers?.get?.('X-Request-ID') || response?.headers?.get?.('x-request-id') || ''
  return fromHeader || payload?.request_id || payload?.requestId || ''
}

function errorFrom(response, payload) {
  const message = payload?.message || payload?.error || `请求失败（HTTP ${response.status}）`
  return new APIError(message, {
    status: response.status,
    code: payload?.code || 'API_ERROR',
    payload,
    requestId: responseRequestId(response, payload),
  })
}

function browserCookies() {
  if (typeof document === 'undefined') return ''
  try {
    return document.cookie || ''
  } catch {
    return ''
  }
}

function hasKnownSession() {
  return browserCookies()
    .split(';')
    .some((cookie) => cookie.trim() === `${AUTH_SESSION_HINT_COOKIE}=1`)
}

function rememberKnownSession() {
  if (typeof document !== 'undefined') {
    try {
      document.cookie = `${AUTH_SESSION_HINT_COOKIE}=1; Path=/; Max-Age=${AUTH_SESSION_HINT_MAX_AGE}; SameSite=Lax`
    } catch {
      // The hint is UX-only. Authentication remains server-authoritative.
    }
  }
  authFailureNotified = false
}

function forgetKnownSession() {
  if (typeof document === 'undefined') return
  try {
    document.cookie = `${AUTH_SESSION_HINT_COOKIE}=; Path=/; Max-Age=0; SameSite=Lax`
  } catch {
    // Ignore hint cleanup failures; logout remains server-authoritative.
  }
}

function notifyUnauthenticated(error) {
  if (authFailureNotified) return
  authFailureNotified = true
  if (typeof window !== 'undefined' && typeof window.dispatchEvent === 'function') {
    window.dispatchEvent(new CustomEvent('ycm:auth-unauthenticated', { detail: { error } }))
  }
}

async function refreshSession() {
  if (!refreshFlight) {
    refreshFlight = (async () => {
      const response = await fetch(REFRESH_PATH, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
      })
      const payload = await parsePayload(response)
      if (!response.ok) {
        const error = errorFrom(response, payload)
        notifyUnauthenticated(error)
        throw error
      }
      rememberKnownSession()
      return payload
    })().finally(() => {
      refreshFlight = null
    })
  }
  return refreshFlight
}

export async function requestJSON(path, options = {}) {
  const { skipAuthRecovery = false, ...fetchOptions } = options
  const response = await fetch(path, {
    ...fetchOptions,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...(fetchOptions.headers || {}),
    },
  })

  const payload = await parsePayload(response)
  if (response.ok) return payload

  if (response.status === 401 && !skipAuthRecovery && !AUTH_PATHS.has(path)) {
    await refreshSession()
    return requestJSON(path, { ...options, skipAuthRecovery: true })
  }

  throw errorFrom(response, payload)
}

export function __resetAuthRecoveryForTests() {
  refreshFlight = null
  authFailureNotified = false
  forgetKnownSession()
}

export async function login(input) {
  const payload = await requestJSON('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify(input),
    skipAuthRecovery: true,
  })
  if (payload?.user) rememberKnownSession()
  return payload
}

export async function getCurrentUser() {
  const payload = await requestJSON('/api/auth/current-user', {
    skipAuthRecovery: !hasKnownSession(),
  })
  if (payload?.user) rememberKnownSession()
  return payload
}

export async function logout() {
  const payload = await requestJSON('/api/auth/logout', {
    method: 'POST',
    skipAuthRecovery: true,
  })
  forgetKnownSession()
  return payload
}

export function listIntakes() {
  return requestJSON(`${API_PREFIX}/intakes`)
}

export function createIntake(input) {
  return requestJSON(`${API_PREFIX}/intakes`, {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function executeIntake(intakeId, maxText) {
  return requestJSON(`${API_PREFIX}/intakes/${intakeId}/execute`, {
    method: 'POST',
    body: JSON.stringify({ maxText }),
  })
}

export function restoreIntakeBook(intakeId, bookId, maxText = 4000) {
  return requestJSON(`${API_PREFIX}/intakes/${intakeId}/books/${bookId}/restore`, {
    method: 'POST',
    body: JSON.stringify({ maxText }),
  })
}

export function listBooks(intakeId) {
  return requestJSON(`${API_PREFIX}/intakes/${intakeId}/books`)
}

export function getNovelFetchWorkshop(intakeId) {
  return requestJSON(`${API_PREFIX}/intakes/${intakeId}/workshop`)
}

export function saveNovelFetchWorkshop(intakeId, settings) {
  return requestJSON(`${API_PREFIX}/intakes/${intakeId}/workshop`, {
    method: 'PUT',
    body: JSON.stringify({ settings }),
  })
}

export function createBatchProject(intakeId, input) {
  return requestJSON(`${API_PREFIX}/intakes/${intakeId}/batch-projects`, {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function listBatchProjects() {
  return requestJSON(`${API_PREFIX}/batch-projects`)
}
export function listWorkspaceHistory(input = {}) { const query = new URLSearchParams(); for (const [key,value] of Object.entries(input)) if (value !== undefined && value !== null && value !== '') query.set(key,String(value)); return requestJSON(`${API_PREFIX}/history${query.size ? `?${query}` : ''}`) }
export function getWorkspaceSettings(){return requestJSON(`${API_PREFIX}/workspace/settings`)}
export function saveWorkspaceSettings(settings, { skipAuthRecovery = false } = {}) {
  return requestJSON(`${API_PREFIX}/workspace/settings`, { method: 'PUT', body: JSON.stringify(settings), skipAuthRecovery })
}

export function getBatchProject(projectId) {
  return requestJSON(`${API_PREFIX}/batch-projects/${projectId}`)
}

export function getNovelPanel(projectId) { return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/novel-panel`) }
export function saveNovelPanel(projectId, input) { return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/novel-panel`, { method: 'PUT', body: JSON.stringify(input) }) }
export function listNovelPanelHistory(projectId) { return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/novel-panel/history`) }
export function restoreNovelPanelHistory(projectId, historyId, input = {}) { return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/novel-panel/history/${encodeURIComponent(historyId)}/restore`, { method: 'POST', body: JSON.stringify({ expectedRevision: Number(input.expectedRevision || 0) }) }) }

export function getUnifiedSettings(projectId) {
  return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/settings`)
}

export function saveProductionSettings(projectId, settings) {
  return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/settings/production`, {
    method: 'PUT',
    body: JSON.stringify(settings),
  })
}

export function savePublishingSettings(projectId, settings) {
  return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/settings/publishing`, {
    method: 'PUT',
    body: JSON.stringify(settings),
  })
}

export function getVersionProfile(projectId) {
  return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/version-profile`)
}

export function saveVersionProfile(projectId, profile) {
  return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/version-profile`, {
    method: 'PUT',
    body: JSON.stringify(profile),
  })
}

export function sync121Config(projectId) {
  return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/version-profile/sync-121`, { method: 'POST' })
}

export function syncStyleTypes(projectId) {
  return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/version-profile/sync-style-types`, { method: 'POST' })
}

export function getProjectGeneration(projectId) {
  return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/generation`)
}

export function runProjectGeneration(projectId, input) {
  return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/generation`, {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function getBookGeneration(projectId, bookId) {
  return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/books/${bookId}/generation`)
}

export function saveScriptOriginalText(projectId, bookId, originalText) {
  return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/books/${bookId}/original-text`, {
    method: 'PUT', body: JSON.stringify({ originalText }),
  })
}

export function runBookGeneration(projectId, bookId, input) {
  return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/books/${bookId}/generation`, {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function getAudioMeasurement(projectId, bookId) {
  return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/books/${bookId}/audio-measurement`)
}

export function measureAudio(projectId, bookId, audioAsset) {
  return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/books/${bookId}/audio-measurement`, {
    method: 'POST',
    body: JSON.stringify({ audioAsset }),
  })
}

export function retryGenerationStage(projectId, bookId, stage, requestId) {
  return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/books/${bookId}/generation/stages/${stage}/retry`, {
    method: 'POST',
    body: JSON.stringify({ requestId }),
  })
}

export function getGenerationStage(projectId, bookId, stage) {
  return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/books/${bookId}/generation/stages/${stage}`)
}

export function getScriptStoryboard(projectId, bookId) { return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/books/${bookId}/storyboard`) }
export function saveScriptStoryboardCard(projectId, bookId, input) { const id = input.card?.id; return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/books/${bookId}/storyboard/cards${id ? `/${id}` : ''}`, { method: id ? 'PUT' : 'POST', body: JSON.stringify(input) }) }
export function deleteScriptStoryboardCard(projectId, bookId, cardId, expectedVersion) { return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/books/${bookId}/storyboard/cards/${cardId}`, { method: 'DELETE', body: JSON.stringify({ expectedVersion }) }) }
export function reorderScriptStoryboard(projectId, bookId, cardIds, expectedVersion) { return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/books/${bookId}/storyboard/reorder`, { method: 'POST', body: JSON.stringify({ cardIds, expectedVersion }) }) }
export function recompileScriptStoryboard(projectId, bookId, requestId) { return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/books/${bookId}/storyboard/recompile`, { method: 'POST', body: JSON.stringify({ requestId }) }) }

export function listGenerationPrompts() {
  return requestJSON(`${API_PREFIX}/generation/prompts`)
}

// Script Workspace is deliberately a projection of the existing persisted
// intake / BatchProject / generation graph.  It does not introduce a second
// browser-side project or task store.
export function getScriptWorkspace(projectId) {
  return Promise.all([
    getBatchProject(projectId),
    getProjectGeneration(projectId),
    getUnifiedSettings(projectId),
    getProjectVideoStatus(projectId),
    listGenerationPrompts(),
  ]).then(([project, generation, settings, video, prompts]) => ({ project, generation, settings, video, prompts }))
}

export function getProjectVideoStatus(projectId) {
  return requestJSON(`${API_PREFIX}/batch-projects/${encodeURIComponent(projectId)}/video`)
}

export function startVideoTask(projectId, bookId, input) {
  return requestJSON(`${API_PREFIX}/batch-projects/${encodeURIComponent(projectId)}/books/${encodeURIComponent(bookId)}/video`, {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function retryVideoTask(taskId, requestId) {
  return requestJSON(`${API_PREFIX}/video-tasks/${encodeURIComponent(taskId)}/retry`, {
    method: 'POST',
    body: JSON.stringify({ requestId }),
  })
}

export function cancelVideoTask(taskId) {
  return requestJSON(`${API_PREFIX}/video-tasks/${encodeURIComponent(taskId)}/cancel`, {
    method: 'POST',
  })
}

// Shuihuo 5B uses the persisted Task 12B media graph. These calls do not
// manufacture browser-side tasks or assets; MySQL remains the source of truth.
function shuihuoBase(projectId, bookId) {
  return `${API_PREFIX}/batch-projects/${encodeURIComponent(projectId)}/books/${encodeURIComponent(bookId)}/shuihuo`
}
export function listShuihuoSegments(projectId, bookId) { return requestJSON(`${shuihuoBase(projectId, bookId)}/segments`) }
export function updateShuihuoSegment(projectId, bookId, segmentId, input) { return requestJSON(`${shuihuoBase(projectId, bookId)}/segments/${encodeURIComponent(segmentId)}`, { method: 'PUT', body: JSON.stringify(input) }) }
export function reorderShuihuoSegments(projectId, bookId, segmentIds) { return requestJSON(`${shuihuoBase(projectId, bookId)}/segments/reorder`, { method: 'POST', body: JSON.stringify({ segmentIds }) }) }
export function listShuihuoAssets(projectId, bookId) { return requestJSON(`${shuihuoBase(projectId, bookId)}/assets`) }
export function shuihuoAssetContentURL(projectId, bookId, assetId) { return `${shuihuoBase(projectId, bookId)}/assets/${encodeURIComponent(assetId)}/content` }
export function listShuihuoMediaTasks(projectId, bookId) { return requestJSON(`${shuihuoBase(projectId, bookId)}/media-tasks`) }
export function createShuihuoMediaTask(projectId, bookId, input) { return requestJSON(`${shuihuoBase(projectId, bookId)}/media-tasks`, { method: 'POST', body: JSON.stringify(input) }) }
export function listShuihuoCandidates(projectId, bookId, taskId) { return requestJSON(`${shuihuoBase(projectId, bookId)}/media-tasks/${encodeURIComponent(taskId)}/candidates`) }
export function selectShuihuoCandidate(projectId, bookId, taskId, candidateId) { return requestJSON(`${shuihuoBase(projectId, bookId)}/media-tasks/${encodeURIComponent(taskId)}/candidates/${encodeURIComponent(candidateId)}/select`, { method: 'POST' }) }
export function retryShuihuoMediaTask(projectId, bookId, taskId) { return requestJSON(`${shuihuoBase(projectId, bookId)}/media-tasks/${encodeURIComponent(taskId)}/retry`, { method: 'POST' }) }
