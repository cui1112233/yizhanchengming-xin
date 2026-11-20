import assert from 'node:assert/strict'
import test from 'node:test'

import { ACCOUNT_SERVICE_UNAVAILABLE_MESSAGE, loadSession } from './session.js'

test('loadSession reuses the HttpOnly account cookie and returns the account', async () => {
  const requests = []
  const result = await loadSession(async (url, options) => {
    requests.push({ url, options })
    return {
      ok: true,
      status: 200,
      async json() { return { username: 'owner', displayName: '主账号', active: true } },
    }
  })

  assert.deepEqual(requests, [{ url: '/api/login/session', options: { credentials: 'include' } }])
  assert.equal(result.authenticated, true)
  assert.equal(result.account.username, 'owner')
})

test('loadSession gives a fixed recovery message when port 4001 is unavailable', async () => {
  const result = await loadSession(async () => ({
    ok: false,
    status: 503,
    async json() { return { code: 'ACCOUNT_SERVICE_UNAVAILABLE' } },
  }))

  assert.deepEqual(result, {
    authenticated: false,
    account: null,
    errorCode: 'ACCOUNT_SERVICE_UNAVAILABLE',
    message: ACCOUNT_SERVICE_UNAVAILABLE_MESSAGE,
  })
})
