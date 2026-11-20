import assert from 'node:assert/strict'
import test from 'node:test'

import { createWorkModeLinks } from './workModeLinks.js'

test('work mode links keep the dedicated server hostname', () => {
  assert.deepEqual(createWorkModeLinks({ protocol: 'http:', hostname: '192.168.1.8' }), {
    studio: 'http://192.168.1.8:4001/',
    agent: 'http://192.168.1.8:4001/agent',
    magicCut: '/',
  })
})
