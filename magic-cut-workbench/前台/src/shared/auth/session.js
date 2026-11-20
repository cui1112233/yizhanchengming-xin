export const ACCOUNT_SERVICE_UNAVAILABLE_MESSAGE = '无法连接一战晟铭账号服务，请确认 4001 已启动'

export async function loadSession(fetchImpl = fetch) {
  let response
  try {
    response = await fetchImpl('/api/login/session', { credentials: 'include' })
  } catch (_) {
    return failure('ACCOUNT_SERVICE_UNAVAILABLE', ACCOUNT_SERVICE_UNAVAILABLE_MESSAGE)
  }

  if (response.ok) {
    return {
      authenticated: true,
      account: await response.json(),
      errorCode: '',
      message: '',
    }
  }

  let payload = {}
  try {
    payload = await response.json()
  } catch (_) {
    payload = {}
  }
  const code = payload.code || (response.status === 401 ? 'AUTH_REQUIRED' : 'ACCOUNT_REQUEST_FAILED')
  if (code === 'ACCOUNT_SERVICE_UNAVAILABLE' || response.status === 503) {
    return failure('ACCOUNT_SERVICE_UNAVAILABLE', ACCOUNT_SERVICE_UNAVAILABLE_MESSAGE)
  }
  if (response.status === 401) {
    return failure('AUTH_REQUIRED', '请先从一战晟铭登录后再进入魔剪')
  }
  return failure(code, payload.error || '账号信息读取失败，请稍后重试')
}

function failure(errorCode, message) {
  return { authenticated: false, account: null, errorCode, message }
}
