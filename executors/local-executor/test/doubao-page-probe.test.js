const test = require('node:test');
const assert = require('node:assert/strict');
const { DoubaoPageProbe, buildSnapshotScript, sanitizeSnapshot } = require('../src/doubao-page-probe');

test('page probe script never reads browser credential stores', () => {
  const script = buildSnapshotScript();
  assert.equal(/document\.cookie/i.test(script), false);
  assert.equal(/localStorage/i.test(script), false);
  assert.equal(/sessionStorage/i.test(script), false);
});

test('page probe includes ARIA tabs in semantic control discovery', () => {
  const script = buildSnapshotScript();
  assert.equal(script.includes('[role="tab"]'), true);
});

test('capture returns only redacted semantic page evidence', async () => {
  let calledWithUserGesture = null;
  const webContents = {
    async executeJavaScript(_script, userGesture) {
      calledWithUserGesture = userGesture;
      return {
        visibleText: 'Seedance 2.0 5秒 生成视频',
        controls: [{ role: 'button', text: '生成视频', aria: '', secret: 'drop-me' }],
        promptInputs: [{ kind: 'contenteditable', placeholder: '输入提示词', value: 'private prompt value' }],
        fileInputs: [{ accept: 'image/png', multiple: true, path: '/private/file.png' }],
        identityNodes: [{ identities: ['msg-1', 'task-1'], text: '结果卡片', html: '<secret>' }],
        videos: [{ mediaId: 'media-1', identities: ['task-1'], srcKind: 'https' }],
        cookie: 'must-not-return'
      };
    }
  };
  const snapshot = await new DoubaoPageProbe().capture(webContents);
  assert.equal(calledWithUserGesture, false);
  assert.equal(JSON.stringify(snapshot).includes('private prompt value'), false);
  assert.equal(JSON.stringify(snapshot).includes('/private/file.png'), false);
  assert.equal(JSON.stringify(snapshot).includes('must-not-return'), false);
  assert.equal(snapshot.controls[0].text, '生成视频');
  assert.deepEqual(snapshot.videos[0], { mediaId: 'media-1', identities: ['task-1'], srcKind: 'https' });
});

test('sanitizeSnapshot bounds large page text and arrays', () => {
  const snapshot = sanitizeSnapshot({
    visibleText: 'x'.repeat(50000),
    controls: Array.from({ length: 500 }, (_, i) => ({ text: String(i) }))
  });
  assert.ok(snapshot.visibleText.length <= 20000);
  assert.ok(snapshot.controls.length <= 200);
});
