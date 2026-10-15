const test = require('node:test');
const assert = require('node:assert/strict');
const { AccountWindows } = require('../src/electron/account-windows');

function fakeBrowserWindow() {
  return class FakeWindow {
    constructor(options) {
      this.options = options;
      this.webContents = { id: `wc:${options.webPreferences.partition}`, setWindowOpenHandler() {} };
      this.handlers = new Map();
      this.destroyed = false;
    }
    isDestroyed() { return this.destroyed; }
    show() {}
    focus() {}
    on(name, fn) { this.handlers.set(name, fn); }
    loadURL(url) { this.url = url; }
    close() { this.destroyed = true; this.handlers.get('closed')?.(); }
  };
}

test('registered account window exposes only its own webContents', () => {
  const BrowserWindow = fakeBrowserWindow();
  const manager = new AccountWindows({ BrowserWindow, session: { fromPartition() {} } });
  const win = manager.open('account-1');
  assert.equal(manager.getWebContents('account-1'), win.webContents);
  assert.match(win.options.webPreferences.partition, /^persist:yizhan-doubao-account-1$/);
});

test('unknown or closed account window cannot be accessed by automation', () => {
  const BrowserWindow = fakeBrowserWindow();
  const manager = new AccountWindows({ BrowserWindow, session: { fromPartition() {} } });
  assert.throws(() => manager.getWebContents('missing'), /not open/i);
  const win = manager.open('account-1');
  win.close();
  assert.throws(() => manager.getWebContents('account-1'), /not open/i);
});
