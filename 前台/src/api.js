const API_PREFIX = '/api/v1'

async function requestJSON(path, options = {}) {
  const response = await fetch(path, {
    headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
    ...options,
  })
  let payload = null
  try { payload = await response.json() } catch { payload = null }
  if (!response.ok) throw new Error(payload?.error || `请求失败（HTTP ${response.status}）`)
  return payload
}

export function createIntake(input) { return requestJSON(`${API_PREFIX}/intakes`, { method: 'POST', body: JSON.stringify(input) }) }
export function executeIntake(intakeId, maxText) { return requestJSON(`${API_PREFIX}/intakes/${intakeId}/execute`, { method: 'POST', body: JSON.stringify({ maxText }) }) }
export function listBooks(intakeId) { return requestJSON(`${API_PREFIX}/intakes/${intakeId}/books`) }
export function createBatchProject(intakeId, input) { return requestJSON(`${API_PREFIX}/intakes/${intakeId}/batch-projects`, { method: 'POST', body: JSON.stringify(input) }) }
export function listBatchProjects() { return requestJSON(`${API_PREFIX}/batch-projects`) }

export function getUnifiedSettings(projectId) { return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/settings`) }
export function saveProductionSettings(projectId, settings) { return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/settings/production`, { method: 'PUT', body: JSON.stringify(settings) }) }
export function savePublishingSettings(projectId, settings) { return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/settings/publishing`, { method: 'PUT', body: JSON.stringify(settings) }) }
export function getVersionProfile(projectId) { return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/version-profile`) }
export function saveVersionProfile(projectId, profile) { return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/version-profile`, { method: 'PUT', body: JSON.stringify(profile) }) }
export function sync121Config(projectId) { return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/version-profile/sync-121`, { method: 'POST' }) }
export function syncStyleTypes(projectId) { return requestJSON(`${API_PREFIX}/batch-projects/${projectId}/version-profile/sync-style-types`, { method: 'POST' }) }
