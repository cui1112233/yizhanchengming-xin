import test from 'node:test';
import assert from 'node:assert/strict';
import { createLocalExecutorRefreshGate } from './useLocalExecutorStatus.js';

function deferred() {
  let resolve;
  const promise = new Promise(nextResolve => { resolve = nextResolve; });
  return { promise, resolve };
}

test('local executor refresh gate reuses the request already in flight', async () => {
  const pending = deferred();
  let calls = 0;
  const gate = createLocalExecutorRefreshGate(() => {
    calls += 1;
    return pending.promise;
  });

  const first = gate.refresh();
  const second = gate.refresh();
  assert.equal(first, second);
  assert.equal(calls, 1);

  pending.resolve(['ready']);
  await first;
  await gate.refresh();
  assert.equal(calls, 2);
});
