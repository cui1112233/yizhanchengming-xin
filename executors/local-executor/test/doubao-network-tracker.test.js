const test = require('node:test');
const assert = require('node:assert/strict');
const { EventEmitter } = require('node:events');
const {
  DoubaoNetworkTracker,
  extractIdentityEvidence,
  extractMediaCandidates,
  sanitizeNetworkUrl
} = require('../src/doubao-network-tracker');

test('extractIdentityEvidence keeps only stable platform identity keys', () => {
  const ids = extractIdentityEvidence({
    id: 'generic-id-must-drop',
    token: 'secret-must-drop',
    data: {
      conversation_id: 'conv-1',
      messageId: 'msg-2',
      nested: [{ task_id: 'task-3' }, { mediaId: 'media-4' }]
    }
  });
  assert.deepEqual(ids, ['conv-1', 'msg-2', 'task-3', 'media-4']);
});

test('extractMediaCandidates keeps strong video URLs tied to stable identities', () => {
  const candidates = extractMediaCandidates({
    data: {
      message_id: 'msg-42',
      media_id: 'media-9',
      download_url: 'https://cdn.example/exact.mp4?signature=keep-local',
      cover_url: 'https://cdn.example/cover.jpg'
    }
  });
  assert.deepEqual(candidates, [{
    mediaId: 'media-9',
    identities: ['msg-42', 'media-9'],
    downloadUrl: 'https://cdn.example/exact.mp4?signature=keep-local'
  }]);
});

test('tracks only prompt-bound submit request and returns redacted acceptance identities', async () => {
  const debug = new EventEmitter();
  debug.attached = false;
  debug.isAttached = () => debug.attached;
  debug.attach = () => { debug.attached = true; };
  debug.detach = () => { debug.attached = false; };
  const commands = [];
  debug.sendCommand = async (method, params) => {
    commands.push([method, params]);
    if (method === 'Network.getResponseBody') {
      return { body: JSON.stringify({ data: { message_id: 'msg-42', task_id: 'task-9' }, access_token: 'drop-me' }), base64Encoded: false };
    }
    return {};
  };
  const tracker = new DoubaoNetworkTracker({ webContents: { debugger: debug } });
  await tracker.startAttempt({ prompt: '雨夜城市镜头缓缓推进' });

  debug.emit('message', {}, 'Network.requestWillBeSent', {
    requestId: 'r-unrelated',
    request: { url: 'https://www.doubao.com/api/other?token=secret', postData: '{"prompt":"别的任务"}' }
  });
  debug.emit('message', {}, 'Network.requestWillBeSent', {
    requestId: 'r-submit',
    request: {
      url: 'https://www.doubao.com/api/video/create?token=secret',
      headers: { Authorization: 'Bearer secret', Cookie: 'session=secret' },
      postData: JSON.stringify({ prompt: '雨夜城市镜头缓缓推进' })
    }
  });
  debug.emit('message', {}, 'Network.responseReceived', {
    requestId: 'r-submit',
    response: { status: 200, url: 'https://www.doubao.com/api/video/create?token=secret', mimeType: 'application/json' }
  });
  debug.emit('message', {}, 'Network.loadingFinished', { requestId: 'r-submit' });
  await new Promise(resolve => setImmediate(resolve));

  const evidence = tracker.getEvidence();
  assert.equal(evidence.accepted, true);
  assert.deepEqual(evidence.identities, ['msg-42', 'task-9']);
  assert.equal(JSON.stringify(evidence).includes('secret'), false);
  assert.equal(evidence.endpoint, 'https://www.doubao.com/api/video/create');
  assert.ok(commands.some(([method]) => method === 'Network.enable'));
  tracker.stop();
});

test('records only media response that carries an accepted submission identity', async () => {
  const debug = new EventEmitter();
  debug.isAttached = () => true;
  const bodies = new Map([
    ['submit', JSON.stringify({ message_id: 'msg-42', task_id: 'task-9' })],
    ['other', JSON.stringify({ message_id: 'msg-other', media_id: 'media-newest', download_url: 'https://cdn.example/newest.mp4' })],
    ['exact', JSON.stringify({ message_id: 'msg-42', media_id: 'media-exact', download_url: 'https://cdn.example/exact.mp4?sig=local' })]
  ]);
  debug.sendCommand = async (method, params) => method === 'Network.getResponseBody'
    ? { body: bodies.get(params.requestId) || '{}', base64Encoded: false }
    : {};

  const tracker = new DoubaoNetworkTracker({ webContents: { debugger: debug } });
  await tracker.startAttempt({ prompt: '本次任务' });
  debug.emit('message', {}, 'Network.requestWillBeSent', { requestId: 'submit', request: { url: 'https://x.test/create', postData: '{"prompt":"本次任务"}' } });
  debug.emit('message', {}, 'Network.responseReceived', { requestId: 'submit', response: { status: 200, url: 'https://x.test/create', mimeType: 'application/json' } });
  debug.emit('message', {}, 'Network.loadingFinished', { requestId: 'submit' });
  await new Promise(resolve => setImmediate(resolve));
  assert.equal(tracker.getEvidence().accepted, true);

  for (const requestId of ['other', 'exact']) {
    debug.emit('message', {}, 'Network.responseReceived', { requestId, response: { status: 200, url: `https://x.test/status/${requestId}`, mimeType: 'application/json' } });
    debug.emit('message', {}, 'Network.loadingFinished', { requestId });
  }
  await new Promise(resolve => setImmediate(resolve));
  assert.deepEqual(tracker.getMediaCandidates(), [{
    mediaId: 'media-exact',
    identities: ['msg-42', 'media-exact'],
    downloadUrl: 'https://cdn.example/exact.mp4?sig=local'
  }]);
  tracker.stop();
});

test('failed response or unrelated traffic never becomes accepted', async () => {
  const debug = new EventEmitter();
  debug.isAttached = () => true;
  debug.sendCommand = async method => method === 'Network.getResponseBody'
    ? { body: '{"task_id":"task-x"}', base64Encoded: false }
    : {};
  const tracker = new DoubaoNetworkTracker({ webContents: { debugger: debug } });
  await tracker.startAttempt({ prompt: '本次任务' });
  debug.emit('message', {}, 'Network.requestWillBeSent', { requestId: 'r1', request: { url: 'https://x.test/create', postData: '{"prompt":"别的任务"}' } });
  debug.emit('message', {}, 'Network.responseReceived', { requestId: 'r1', response: { status: 200, mimeType: 'application/json' } });
  debug.emit('message', {}, 'Network.requestWillBeSent', { requestId: 'r2', request: { url: 'https://x.test/create', postData: '{"prompt":"本次任务"}' } });
  debug.emit('message', {}, 'Network.responseReceived', { requestId: 'r2', response: { status: 500, mimeType: 'application/json' } });
  debug.emit('message', {}, 'Network.loadingFinished', { requestId: 'r2' });
  await new Promise(resolve => setImmediate(resolve));
  assert.equal(tracker.getEvidence().accepted, false);
  tracker.stop();
});

test('sanitizeNetworkUrl removes query strings and fragments', () => {
  assert.equal(sanitizeNetworkUrl('https://example.com/a/b?token=secret#x'), 'https://example.com/a/b');
});
