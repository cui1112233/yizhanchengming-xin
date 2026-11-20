import assert from 'node:assert/strict'
import test from 'node:test'

import { createSessionGateModel } from './sessionGateModel.js'

test('unauthenticated magic cut only offers returning to yizhanchengming', () => {
  assert.deepEqual(createSessionGateModel(
    { errorCode: 'AUTH_REQUIRED', message: 'ignored' },
    { protocol: 'http:', hostname: '192.168.1.20' },
  ), {
    title: '请从一战晟铭进入魔剪',
    message: '当前魔剪会话已失效，请返回一战晟铭后重新切换到魔剪。',
    returnLabel: '返回一战晟铭',
    returnHref: 'http://192.168.1.20:4001/',
  })
})

test('account service failure remains a return-only state', () => {
  const model = createSessionGateModel(
    { errorCode: 'ACCOUNT_SERVICE_UNAVAILABLE', message: 'ignored' },
    { protocol: 'https:', hostname: 'magic.example.test' },
  )

  assert.equal(model.title, '无法连接一战晟铭账号服务')
  assert.equal(model.returnHref, 'https://magic.example.test:4001/')
  assert.equal(Object.hasOwn(model, 'login'), false)
})
