const test = require('node:test');
const assert = require('node:assert/strict');
const { AccountWindows } = require('../src/electron/account-windows');

test('ensureWebContents opens a missing account window and waits for its initial load', async () => {
  let releaseLoad;
  class FakeWindow {
    constructor(options) {
      this.options = options;
      this.destroyed = false;
      this.webContents = { id: 'wc-1', setWindowOpenHandler() {}, isDestroyed: () => false };
    }
    isDestroyed() { return this.destroyed; }
    show() {}
    focus() {}
    on() {}
    loadURL(url) {
      this.url = url;
      return new Promise(resolve => { releaseLoad = resolve; });
    }
    close() { this.destroyed = true; }
  }
  const manager = new AccountWindows({ BrowserWindow: FakeWindow, session: { fromPartition() {} } });
  let settled = false;
  const pending = manager.ensureWebContents('account-1').then(value => { settled = true; return value; });
  await new Promise(resolve => setImmediate(resolve));
  assert.equal(settled, false);
  releaseLoad();
  const webContents = await pending;
  assert.equal(webContents.id, 'wc-1');
  assert.match(manager.get('account-1').url, /^https:\/\/www\.doubao\.com\//);
});
