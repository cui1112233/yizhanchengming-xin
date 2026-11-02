import test from 'node:test';
import assert from 'node:assert/strict';
import { isNovelFetchScopedPath, novelFetchAccountOptions } from '../../shared/api/novelFetchAccountScope.js';

test('Novel Fetch account options follow role visibility', () => {
  const team = [
    { username: 'manager', role: 'manager', displayName: '主管' },
    { username: 'member-a', role: 'member', boundTo: 'manager', displayName: '组员甲' },
    { username: 'member-b', role: 'member', boundTo: 'other', displayName: '组员乙' }
  ];
  assert.deepEqual(
    novelFetchAccountOptions({ username: 'manager', role: 'manager', displayName: '主管' }, team, []).map(item => item.value),
    ['manager', 'member-a']
  );
  assert.deepEqual(
    novelFetchAccountOptions({ username: 'member-a', role: 'member' }, team, []).map(item => item.value),
    ['member-a']
  );
  assert.deepEqual(
    novelFetchAccountOptions({ username: 'developer', role: 'dev' }, [], [{ username: 'member-b' }, { username: 'manager' }]).map(item => item.value),
    ['developer', 'manager', 'member-b']
  );
});

test('only Novel Fetch related requests receive the selected account', () => {
  for (const path of ['/api/novel-fetch-workshop/tasks', '/api/batch-rewrite/config', '/api/web-submit/preview', '/api/novel-fetch/upload-session']) {
    assert.equal(isNovelFetchScopedPath(path), true, path);
  }
  assert.equal(isNovelFetchScopedPath('/api/batch-factory/v12/batches'), false);
  assert.equal(isNovelFetchScopedPath('/api/member/team'), false);
});
