const test = require('node:test');
const assert = require('node:assert/strict');
const { determineSubmissionOutcome } = require('../src/doubao-acceptance');

test('new prompt-bound message identity is positive acceptance evidence', () => {
  const outcome = determineSubmissionOutcome({
    prompt: '雨夜城市镜头缓缓推进',
    before: { identityNodes: [{ identities: ['msg-old'], text: '旧内容' }] },
    after: { identityNodes: [
      { identities: ['msg-old'], text: '旧内容' },
      { identities: ['msg-42', 'task-9'], text: '雨夜城市镜头缓缓推进' }
    ] }
  });
  assert.equal(outcome.status, 'accepted');
  assert.equal(outcome.submissionId, 'msg-42');
  assert.deepEqual(outcome.identities, ['msg-42', 'task-9']);
});

test('redacted network identity can prove acceptance while DOM is still catching up', () => {
  const outcome = determineSubmissionOutcome({
    prompt: '人物转身',
    before: { identityNodes: [] },
    after: { identityNodes: [] },
    networkEvidence: { accepted: true, identities: ['conversation-1', 'message-7', 'task-3'] }
  });
  assert.equal(outcome.status, 'accepted');
  assert.equal(outcome.submissionId, 'message-7');
});

test('explicit submit failure with no new identity can be treated as definitely not accepted', () => {
  const outcome = determineSubmissionOutcome({
    prompt: '人物转身',
    before: { identityNodes: [] },
    after: { visibleText: '提交失败，请重试', identityNodes: [] }
  });
  assert.deepEqual(outcome, { status: 'not_accepted' });
});

test('lack of proof remains unknown and never becomes not accepted by timeout alone', () => {
  const outcome = determineSubmissionOutcome({
    prompt: '人物转身',
    before: { identityNodes: [] },
    after: { visibleText: '正在加载', identityNodes: [] }
  });
  assert.deepEqual(outcome, { status: 'unknown' });
});

test('ambiguous multiple new prompt-bound identities remains unknown', () => {
  const outcome = determineSubmissionOutcome({
    prompt: '人物转身',
    before: { identityNodes: [] },
    after: { identityNodes: [
      { identities: ['msg-a'], text: '人物转身' },
      { identities: ['msg-b'], text: '人物转身' }
    ] }
  });
  assert.deepEqual(outcome, { status: 'unknown' });
});
