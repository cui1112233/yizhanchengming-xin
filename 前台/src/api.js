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

export function getBatchProject(projectId) {
  return requestJSON(`${API_PREFIX}/batch-projects/${encodeURIComponent(projectId)}`)
}
