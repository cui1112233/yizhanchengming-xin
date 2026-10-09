import assert from 'node:assert/strict';
import test from 'node:test';

import { captureAccountCenterOrigin, normalizeAccountCenterReturnPath } from './accountCenterNavigation.js';

const ORIGIN = 'http://localhost:3001';

test('captures a business URL once and retains its query and hash', () => {
  assert.equal(
    captureAccountCenterOrigin('/shuihuo-production?book=42#stage', null, ORIGIN),
    '/shuihuo-production?book=42#stage'
  );
  assert.equal(
    captureAccountCenterOrigin('/script?draft=next', '/shuihuo-production?book=42#stage', ORIGIN),
    '/shuihuo-production?book=42#stage'
  );
});

test('returns home for direct entry, a foreign origin, or an account-center origin', () => {
  assert.equal(normalizeAccountCenterReturnPath(null, ORIGIN), '/');
  assert.equal(normalizeAccountCenterReturnPath('https://example.com/', ORIGIN), '/');
  assert.equal(normalizeAccountCenterReturnPath('/security', ORIGIN), '/');
  assert.equal(normalizeAccountCenterReturnPath('/shuihuo-production?book=42#stage', ORIGIN), '/shuihuo-production?book=42#stage');
});
