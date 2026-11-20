import assert from 'node:assert/strict'
import test from 'node:test'

import { resolveHomeToolAction } from './homeInteractionModel.js'

test('home tools resolve to create, unavailable, or navigation behavior', () => {
  assert.equal(resolveHomeToolAction({ action: 'create-project', available: true }), 'create-project')
  assert.equal(resolveHomeToolAction({ action: 'navigate', available: false }), 'unavailable')
  assert.equal(resolveHomeToolAction({ action: 'navigate', available: true }), 'navigate')
})

test('missing destinations never pretend to navigate', () => {
  assert.equal(resolveHomeToolAction({ action: '', available: true }), 'unavailable')
  assert.equal(resolveHomeToolAction(null), 'unavailable')
})
