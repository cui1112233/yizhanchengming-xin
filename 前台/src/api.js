const API_PREFIX = '/api/v1'

async function requestJSON(path, options = {}) {
  const response = await fetch(path, {
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    },
    ...options,
  })

  let payload = null
  try {
    payload = await response.json()
  } catch {
    payload = null
  }

  if (!response.ok) {
    const message = payload?.error || `请求失败（HTTP ${response.status}）`
    throw new Error(message)
  }
  return payload
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
