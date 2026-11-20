import assert from 'node:assert/strict'
import test from 'node:test'

import { createWorkModeLinks, createWorkModeModel } from './workModeLinks.js'

test('work mode links keep the dedicated server hostname', () => {
  assert.deepEqual(createWorkModeLinks({ protocol: 'http:', hostname: '192.168.1.8' }), {
    studio: 'http://192.168.1.8:4001/',
    agent: 'http://192.168.1.8:4001/agent/chat',
    magicCut: '/',
  })
})

test('magic cut keeps the yizhan logo, labels the current area, and exposes the complete mode switcher', () => {
  const model = createWorkModeModel({ protocol: 'http:', hostname: '192.168.1.8' })

  assert.equal(model.brandLabel, '魔剪')
  assert.equal(model.logoSrc, '/assets/brand-logo-white.png')
  assert.deepEqual(model.modes.map(mode => mode.label), ['一战晟铭', 'Agent 代理', '魔剪', 'Duet'])
  assert.equal(model.modes.find(mode => mode.active)?.id, 'magic-cut')
  assert.equal(model.modes.find(mode => mode.id === 'duet')?.disabled, true)
})
