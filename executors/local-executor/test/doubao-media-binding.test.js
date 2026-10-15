const test = require('node:test');
const assert = require('node:assert/strict');
const { bindExactMedia, MediaBindingError } = require('../src/doubao-media-binding');

test('binds media only when it contains the accepted submission identity', () => {
  const media = bindExactMedia(
    { submissionId: 'msg-42', identities: ['msg-42', 'task-9'] },
    [
      { mediaId: 'video-old', identities: ['msg-41', 'task-8'], createdAt: 999 },
      { mediaId: 'video-42', identities: ['task-9'], createdAt: 100 },
      { mediaId: 'video-newest', identities: ['msg-99'], createdAt: 1000 }
    ]
  );
  assert.equal(media.mediaId, 'video-42');
});

test('never falls back to newest or last video when no exact identity matches', () => {
  assert.throws(() => bindExactMedia(
    { submissionId: 'msg-42', identities: ['msg-42'] },
    [
      { mediaId: 'video-a', identities: ['msg-a'], createdAt: 1 },
      { mediaId: 'video-latest', identities: ['msg-latest'], createdAt: 999 }
    ]
  ), error => error instanceof MediaBindingError && error.code === 'EXACT_MEDIA_NOT_FOUND');
});

test('ambiguous exact matches fail closed instead of choosing one', () => {
  assert.throws(() => bindExactMedia(
    { submissionId: 'task-9', identities: ['task-9'] },
    [
      { mediaId: 'video-a', identities: ['task-9'] },
      { mediaId: 'video-b', identities: ['task-9'] }
    ]
  ), error => error instanceof MediaBindingError && error.code === 'EXACT_MEDIA_AMBIGUOUS');
});

test('empty or missing accepted identity is rejected', () => {
  assert.throws(() => bindExactMedia({ submissionId: '' }, []), error => error?.code === 'SUBMISSION_IDENTITY_REQUIRED');
});
