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

test('shares the local login token across the 3001 and 4001 frontends', () => {
  const previousStorage = globalThis.localStorage;
  const previousDocument = globalThis.document;
  const values = new Map([['auth_token', 'node-session-token']]);
  let cookie = '';
  globalThis.localStorage = {
    getItem(key) { return values.get(key) || null; },
    setItem(key, value) { values.set(key, value); },
    removeItem(key) { values.delete(key); }
  };
  globalThis.document = {
    get cookie() { return cookie; },
    set cookie(value) { cookie = value; }
  };
  try {
    assert.equal(getToken(), 'node-session-token');
    assert.match(cookie, /qiantie_auth_token=node-session-token/);
    values.delete('auth_token');
    assert.equal(getToken(), 'node-session-token');
    setToken('');
    assert.match(cookie, /Max-Age=0/);
  } finally {
    globalThis.localStorage = previousStorage;
    globalThis.document = previousDocument;
  }
});
