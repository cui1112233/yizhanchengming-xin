import assert from 'node:assert/strict'
import test from 'node:test'

import { loginMagicCut } from './sessionLogin.js'

test('login exchanges the short-lived token for the HttpOnly Magic Cut cookie', async () => {
  const calls = []
  const fetchImpl = async (url, options) => {
    calls.push({ url, options })
    if (url === '/api/login') return new Response(JSON.stringify({ token: 'short-lived-token', username: 'alice' }), { status: 200, headers: { 'Content-Type': 'application/json' } })
    return new Response(JSON.stringify({ authenticated: true, account: { username: 'alice' } }), { status: 200, headers: { 'Content-Type': 'application/json' } })
  }
  const account = await loginMagicCut({ username: 'alice', password: 'password', remember: true }, fetchImpl)
  assert.equal(account.username, 'alice')
  assert.equal(calls.length, 2)
  assert.equal(calls[1].url, '/api/auth/bootstrap')
  assert.equal(calls[1].options.headers.Authorization, 'Bearer short-lived-token')
})

test('login never bootstraps when account authentication fails', async () => {
  let calls = 0
  await assert.rejects(() => loginMagicCut({ username: 'alice', password: 'wrong', remember: true }, async () => {
    calls += 1
    return new Response(JSON.stringify({ error: '账号或密码错误' }), { status: 401, headers: { 'Content-Type': 'application/json' } })
  }), /账号或密码错误/)
  assert.equal(calls, 1)
})
