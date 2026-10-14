import test from 'node:test';
import assert from 'node:assert/strict';

import { getToken, isSessionAuthFailure, setToken } from './client.js';

function responseWithHeader(value) {
  return {
    headers: {
      get(name) {
        return name.toLowerCase() === 'x-qiantie-auth-failure' ? value : null;
      }
    }
  };
}

test('recognizes the primary application session-expired response', () => {
  assert.equal(isSessionAuthFailure(responseWithHeader('session')), true);
});

test('does not treat downstream bridge 401 responses as session expiry', () => {
  assert.equal(isSessionAuthFailure(responseWithHeader(null)), false);
  assert.equal(isSessionAuthFailure(responseWithHeader('bridge')), false);
});

test('reads and clears the Go application session token from local storage', () => {
  const previousStorage = globalThis.localStorage;
  const values = new Map([['auth_token', 'go-session-token']]);
  globalThis.localStorage = {
    getItem(key) { return values.get(key) || null; },
    setItem(key, value) { values.set(key, value); },
    removeItem(key) { values.delete(key); }
  };
  try {
    assert.equal(getToken(), 'go-session-token');
    values.delete('auth_token');
    assert.equal(getToken(), '');
    setToken('go-session-token-2');
    assert.equal(getToken(), 'go-session-token-2');
    setToken('');
    assert.equal(getToken(), '');
  } finally {
    globalThis.localStorage = previousStorage;
  }
});
