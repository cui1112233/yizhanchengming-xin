const API_PREFIX = '/api/v1'
const REFRESH_PATH = '/api/auth/refresh'
const AUTH_PATHS = new Set(['/api/auth/login', REFRESH_PATH, '/api/auth/logout'])

let refreshFlight = null
let authFailureNotified = false

export class APIError extends Error {
  constructor(message, { status = 0, code = 'API_ERROR', payload = null } = {}) {
    super(message)
    this.name = 'APIError'
    this.status = status
    this.code = code
    this.payload = payload
  }
}

async function parsePayload(response) {
  try {
    return await response.json()
  } catch {
    return null
  }
}

function errorFrom(response, payload) {
  const message = payload?.message || payload?.error || `请求失败（HTTP ${response.status}）`
  return new APIError(message, {
    status: response.status,
    code: payload?.code || 'API_ERROR',
    payload,
  })
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
      authFailureNotified = false
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
}

export function login(input) {
  return requestJSON('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify(input),
    skipAuthRecovery: true,
  })
}

export function getCurrentUser() {
  return requestJSON('/api/auth/current-user')
}

export function logout() {
  return requestJSON('/api/auth/logout', {
    method: 'POST',
    skipAuthRecovery: true,
  })
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

export function listBooks(intakeId) {
  return requestJSON(`${API_PREFIX}/intakes/${intakeId}/books`)
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

export function getBatchProject(projectId) {
  return requestJSON(`${API_PREFIX}/batch-projects/${projectId}`)
}

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

export function listGenerationPrompts() {
  return requestJSON(`${API_PREFIX}/generation/prompts`)
}

export function getProjectVideoStatus(projectId) {
  return requestJSON(`${API_PREFIX}/batch-projects/${encodeURIComponent(projectId)}/video`)
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
