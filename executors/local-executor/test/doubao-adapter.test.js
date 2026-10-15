const test = require('node:test');
const assert = require('node:assert/strict');
const { DoubaoAdapter, DoubaoAccountError } = require('../src/doubao-adapter');

function availableSnapshot(extra = {}) {
  return {
    visibleText: '视频生成',
    controls: [
      { text: 'Seedance 2.0 Fast' },
      { text: '10秒' },
      { text: '9:16' },
      { text: '生成视频' }
    ],
    promptInputs: [{ kind: 'textarea', placeholder: '输入提示词' }],
    fileInputs: [{ accept: 'image/*', multiple: true }],
    identityNodes: [],
    videos: [],
    ...extra
  };
}

function harness({ snapshots, evidence, mediaCandidates = [], downloadPath = '/tmp/exact.mp4' } = {}) {
  const calls = [];
  const queue = [...(snapshots || [availableSnapshot(), availableSnapshot()])];
  const webContents = { id: 7, session: {} };
  const tracker = {
    async startAttempt(input) { calls.push(['startAttempt', input]); },
    getEvidence() { return evidence || { accepted: true, identities: ['msg-1', 'task-1'] }; },
    getMediaCandidates() { return mediaCandidates; },
    stop() { calls.push(['stop']); }
  };
  const adapter = new DoubaoAdapter({
    accountWindows: { getWebContents(id) { calls.push(['getWebContents', id]); return webContents; } },
    pageProbe: { async capture() { calls.push(['capture']); return queue.shift() || availableSnapshot(); } },
    pageActions: {
      async clickExactControl(_wc, text) { calls.push(['click', text]); },
      async setReferenceImages(_wc, paths) { calls.push(['images', paths]); },
      async setPrompt(_wc, prompt) { calls.push(['prompt', prompt]); },
      async submit() { calls.push(['submit']); },
      async confirmNormal() { calls.push(['confirm']); return false; }
    },
    trackerFactory: () => tracker,
    downloader: async input => { calls.push(['download', input.url, input.media.mediaId]); return { filePath: downloadPath }; },
    sleep: async () => {},
    maxCompletionPolls: 2,
    completionPollMs: 0
  });
  return { adapter, calls, webContents };
}

test('prepare validates live page then sets exact options, reference images and prompt', async () => {
  const { adapter, calls } = harness();
  await adapter.prepare({
    job: { id: 'job-1', payload: { prompt: '雨夜追车', images: ['/tmp/ref.png'], model: 'Seedance 2.0 Fast', duration: 10, ratio: '9:16' } },
    account: { id: 'acct-1' }
  });
  assert.deepEqual(calls.filter(x => x[0] === 'click'), [
    ['click', 'Seedance 2.0 Fast'], ['click', '10秒'], ['click', '9:16']
  ]);
  assert.deepEqual(calls.find(x => x[0] === 'images'), ['images', ['/tmp/ref.png']]);
  assert.deepEqual(calls.find(x => x[0] === 'prompt'), ['prompt', '雨夜追车']);
});

test('submit records acceptance once and recovery never clicks submit again', async () => {
  const { adapter, calls } = harness();
  const job = { id: 'job-2', payload: { prompt: '城市航拍' } };
  const account = { id: 'acct-1' };
  await adapter.prepare({ job, account });
  const accepted = await adapter.submit({ job, account });
  assert.equal(accepted.status, 'accepted');
  assert.equal(accepted.submissionId, 'msg-1');
  const recovered = await adapter.recoverAcceptance({ job, account });
  assert.equal(recovered.status, 'accepted');
  assert.equal(calls.filter(x => x[0] === 'submit').length, 1);
});

test('completion and download are bound to the accepted submission, never the newest unrelated video', async () => {
  const snapshots = [
    availableSnapshot(),
    availableSnapshot(),
    availableSnapshot({ videos: [
      { mediaId: 'media-newest', identities: ['other-task'], srcKind: 'https' },
      { mediaId: 'media-exact', identities: ['msg-1', 'media-exact'], srcKind: 'https' }
    ] })
  ];
  const mediaCandidates = [
    { mediaId: 'media-newest', identities: ['other-task'], downloadUrl: 'https://cdn.example/newest.mp4' },
    { mediaId: 'media-exact', identities: ['msg-1', 'media-exact'], downloadUrl: 'https://cdn.example/exact.mp4' }
  ];
  const { adapter, calls } = harness({ snapshots, mediaCandidates });
  const job = { id: 'job-3', payload: { prompt: '精确视频' } };
  const account = { id: 'acct-1' };
  await adapter.prepare({ job, account });
  const accepted = await adapter.submit({ job, account });
  const completion = await adapter.waitForCompletion({ job, account, submissionId: accepted.submissionId });
  assert.equal(completion.mediaId, 'media-exact');
  const artifact = await adapter.fetchArtifact({ job, account, completion });
  assert.equal(artifact.filePath, '/tmp/exact.mp4');
  assert.deepEqual(calls.find(x => x[0] === 'download'), ['download', 'https://cdn.example/exact.mp4', 'media-exact']);
});

test('human verification is a typed account hold and never submits', async () => {
  const { adapter, calls } = harness({ snapshots: [availableSnapshot({ visibleText: '请完成安全验证后继续' })] });
  await assert.rejects(
    () => adapter.prepare({ job: { id: 'job-4', payload: { prompt: '测试' } }, account: { id: 'acct-1' } }),
    error => error instanceof DoubaoAccountError && error.accountState === 'human_verification'
  );
  assert.equal(calls.some(x => x[0] === 'submit'), false);
});
