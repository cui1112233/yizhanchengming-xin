const test = require('node:test');
const assert = require('node:assert/strict');
const { DoubaoAdapter } = require('../src/doubao-adapter');

function chatSnapshot() {
  return {
    visibleText: '豆包 普通对话 AI创作',
    controls: [{ text: 'AI创作' }, { text: '发送' }],
    promptInputs: [{ kind: 'contenteditable' }], fileInputs: [], identityNodes: [], videos: []
  };
}
function workHomeSnapshot() {
  return {
    visibleText: '豆包工作 今天有什么工作要处理 新工作任务 定时任务 技能 连接器 伙伴 云盘 主对话 内容创作 设计与创意',
    controls: [{ text: '新工作任务' }, { text: '内容创作' }, { text: '设计与创意' }],
    promptInputs: [{ kind: 'contenteditable' }], fileInputs: [], identityNodes: [], videos: []
  };
}
function creationSnapshot() {
  return {
    visibleText: 'AI创作 图片 视频',
    controls: [{ text: '视频' }],
    promptInputs: [], fileInputs: [], identityNodes: [], videos: []
  };
}
function videoSnapshot() {
  return {
    visibleText: '视频生成 Seedance 2.0 Fast',
    controls: [{ text: 'Seedance 2.0 Fast' }, { text: '10秒' }, { text: '9:16' }, { text: '生成视频' }],
    promptInputs: [{ kind: 'textarea' }], fileInputs: [], identityNodes: [], videos: []
  };
}

test('prepare enters AI creation then video mode before touching the prompt', async () => {
  const calls = [];
  const snapshots = [chatSnapshot(), creationSnapshot(), videoSnapshot()];
  const adapter = new DoubaoAdapter({
    accountWindows: {
      async ensureWebContents(id) { calls.push(['ensure', id]); return { id: 1 }; },
      getWebContents() { throw new Error('should use ensureWebContents'); }
    },
    pageProbe: { async capture() { calls.push(['capture']); return snapshots.shift(); } },
    pageActions: {
      async openCreationWorkspace() { calls.push(['navigate', 'AI创作']); return true; },
      async openVideoMode() { calls.push(['navigate', '视频']); return true; },
      async clickExactControl(_wc, text) { calls.push(['click', text]); },
      async setPrompt(_wc, text) { calls.push(['prompt', text]); },
      async setReferenceImages() {}
    },
    sleep: async () => {}, navigationDelayMs: 0
  });

  await adapter.prepare({
    job: { id: 'job-nav', payload: { prompt: '只允许发到视频生成', model: 'Seedance 2.0 Fast', duration: 10, ratio: '9:16' } },
    account: { id: 'acct-1' }
  });
  assert.deepEqual(calls.filter(x => x[0] === 'navigate'), [['navigate', 'AI创作'], ['navigate', '视频']]);
  assert.equal(calls.findIndex(x => x[0] === 'prompt') > calls.findIndex(x => x[0] === 'navigate' && x[1] === '视频'), true);
});

test('prepare leaves Doubao work home through the dedicated creation route and waits for video controls', async () => {
  const calls = [];
  const snapshots = [workHomeSnapshot(), workHomeSnapshot(), workHomeSnapshot(), videoSnapshot()];
  const adapter = new DoubaoAdapter({
    accountWindows: {
      async ensureWebContents(id) { calls.push(['ensure', id]); return { id: 2 }; },
      getWebContents() { throw new Error('should use ensureWebContents'); }
    },
    pageProbe: { async capture() { calls.push(['capture']); return snapshots.shift() || videoSnapshot(); } },
    pageActions: {
      async openVideoWorkspace() { calls.push(['navigate', 'direct-video-workspace']); return true; },
      async openCreationWorkspace() { calls.push(['navigate', 'AI创作']); return false; },
      async openVideoMode() { calls.push(['navigate', '视频']); return false; },
      async clickExactControl() {},
      async setPrompt(_wc, text) { calls.push(['prompt', text]); },
      async setReferenceImages() {}
    },
    sleep: async () => { calls.push(['sleep']); },
    navigationDelayMs: 0,
    videoWorkspacePollMs: 0,
    maxVideoWorkspacePolls: 4
  });

  await adapter.prepare({
    job: { id: 'job-work-home', payload: { prompt: '从豆包工作首页进入视频生成' } },
    account: { id: 'acct-1' }
  });

  assert.equal(calls.some(x => x[0] === 'navigate' && x[1] === 'direct-video-workspace'), true);
  assert.equal(calls.filter(x => x[0] === 'capture').length >= 4, true);
  assert.equal(calls.some(x => x[0] === 'prompt'), true);
});
