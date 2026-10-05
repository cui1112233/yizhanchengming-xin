import { describe, expect, it } from 'vitest'

import { safeRequestIdFromHeaders } from '../e2e/support/request-id.js'

describe('Task16 request-id evidence', () => {
  it('captures only a safe single X-Request-ID value', () => {
    expect(safeRequestIdFromHeaders({ 'x-request-id': 'req_E2E-1:stage.2' })).toBe('req_E2E-1:stage.2')
    expect(safeRequestIdFromHeaders({ 'x-request-id': 'bad/request?id=1' })).toBe('')
    expect(safeRequestIdFromHeaders({ 'x-request-id': 'bad\nforged' })).toBe('')
    expect(safeRequestIdFromHeaders({ 'x-request-id': 'a'.repeat(129) })).toBe('')
    expect(safeRequestIdFromHeaders({ 'x-request-id': ['one', 'two'] })).toBe('')
  })
})
