async function readPayload(response) {
  const contentType = response.headers.get('content-type') || ''
  return contentType.includes('application/json') ? response.json() : response.text()
}

export async function loginMagicCut(values, fetchImpl = fetch) {
  const loginResponse = await fetchImpl('/api/login', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(values),
  })
  const loginPayload = await readPayload(loginResponse)
  if (!loginResponse.ok) throw new Error(loginPayload?.error || '一战晟铭登录失败')
  if (!loginPayload?.token) throw new Error('登录结果缺少临时凭据')
  const bootstrapResponse = await fetchImpl('/api/auth/bootstrap', {
    method: 'POST',
    credentials: 'include',
    headers: { Authorization: `Bearer ${loginPayload.token}` },
  })
  const bootstrapPayload = await readPayload(bootstrapResponse)
  if (!bootstrapResponse.ok) throw new Error(bootstrapPayload?.error || '魔剪会话建立失败')
  return bootstrapPayload.account || loginPayload
}
