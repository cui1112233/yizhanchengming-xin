import test from 'node:test';
import assert from 'node:assert/strict';
import { hasReadyLocalExecutor, localExecutorStatus, parseLocalExecutorsResponse } from './localExecutors.js';

test('parseLocalExecutorsResponse prefers canonical executors and keeps items as compatibility fallback', () => {
  const canonical = [{ id: 'canonical', online: true }];
  const legacy = [{ id: 'legacy', online: true }];
  assert.deepEqual(parseLocalExecutorsResponse({ executors: canonical, items: legacy }), canonical);
  assert.deepEqual(parseLocalExecutorsResponse({ items: legacy }), legacy);
  assert.deepEqual(parseLocalExecutorsResponse(null), []);
});

test('hasReadyLocalExecutor requires at least one executor explicitly online', () => {
  assert.equal(hasReadyLocalExecutor([{ id: 'a', online: false }, { id: 'b', online: true }]), true);
  assert.equal(hasReadyLocalExecutor([{ id: 'a', online: false }]), false);
  assert.equal(hasReadyLocalExecutor([{ id: 'a' }]), false);
  assert.equal(hasReadyLocalExecutor([]), false);
});

test('localExecutorStatus reports paired and online device counts from canonical executors', () => {
  assert.deepEqual(localExecutorStatus([{ id: 'a', online: true }, { id: 'b', online: false }]), {
    online: true,
    paired: 2,
    onlineCount: 1
  });
  assert.deepEqual(localExecutorStatus(null), { online: false, paired: 0, onlineCount: 0 });
});
