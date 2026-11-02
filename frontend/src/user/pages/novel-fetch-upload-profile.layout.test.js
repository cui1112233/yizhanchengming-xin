import test from 'node:test';
import assert from 'node:assert/strict';

import { mergeDefaultUploadProfile, uploadSessionDisplayName } from './novel-fetch-upload-profile.js';

test('ready upload session shows the saved video-management username', () => {
  assert.equal(uploadSessionDisplayName({ status: 'ready', username: ' operator ' }), 'operator');
  assert.equal(uploadSessionDisplayName({ status: 'ready' }), '已登录');
  assert.equal(uploadSessionDisplayName({ status: 'offline', username: 'operator' }), '未登录');
});

test('saving the default website profile preserves advanced submission settings', () => {
  const original = {
    selected_profile: 'profile-old',
    profile_bindings: { original: 'profile-original', ai1: 'profile-ai1' },
    submit_versions: ['original', 'ai1'],
    retry_times: 3,
    skip_submitted: true,
    nested: { keep: 'yes' }
  };

  const result = mergeDefaultUploadProfile(original, ' profile-new ');

  assert.deepEqual(result, {
    selected_profile: 'profile-new',
    profile_bindings: { original: 'profile-original', ai1: 'profile-ai1' },
    submit_versions: ['original', 'ai1'],
    retry_times: 3,
    skip_submitted: true,
    nested: { keep: 'yes' }
  });
  assert.notStrictEqual(result, original);
  assert.deepEqual(original, {
    selected_profile: 'profile-old',
    profile_bindings: { original: 'profile-original', ai1: 'profile-ai1' },
    submit_versions: ['original', 'ai1'],
    retry_times: 3,
    skip_submitted: true,
    nested: { keep: 'yes' }
  });
});

test('blank default website profile is normalized without dropping other settings', () => {
  const result = mergeDefaultUploadProfile({ profile_bindings: { ai2: 'profile-ai2' } }, '   ');

  assert.deepEqual(result, {
    selected_profile: '',
    profile_bindings: { ai2: 'profile-ai2' }
  });
});
