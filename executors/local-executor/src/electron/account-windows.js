const DEFAULT_DOUBAO_URL = 'https://www.doubao.com/';

class AccountWindows {
  constructor({ BrowserWindow, session, startUrl = DEFAULT_DOUBAO_URL }) {
    this.BrowserWindow = BrowserWindow;
    this.session = session;
    this.startUrl = startUrl;
    this.windows = new Map();
    this.loadPromises = new Map();
  }

  open(accountId) {
    const safeId = safeAccountId(accountId);
    const existing = this.windows.get(safeId);
    if (existing && !existing.isDestroyed()) {
      existing.show();
      existing.focus();
      return existing;
    }

    const partition = `persist:yizhan-doubao-${safeId}`;
    this.session.fromPartition(partition);
    const win = new this.BrowserWindow({
      width: 1280,
      height: 860,
      minWidth: 960,
      minHeight: 680,
      show: true,
      title: `豆包账号 · ${safeId}`,
      webPreferences: {
        partition,
        nodeIntegration: false,
        contextIsolation: true,
        sandbox: true,
        webSecurity: true
      }
    });
    win.webContents.setWindowOpenHandler(() => ({ action: 'deny' }));
    win.on('closed', () => {
      this.windows.delete(safeId);
      this.loadPromises.delete(safeId);
    });
    const loading = Promise.resolve(win.loadURL(this.startUrl));
    this.loadPromises.set(safeId, loading);
    this.windows.set(safeId, win);
    return win;
  }

  get(accountId) {
    return this.windows.get(safeAccountId(accountId)) || null;
  }

  getWebContents(accountId) {
    const win = this.get(accountId);
    if (!win || win.isDestroyed()) throw new Error('Doubao account browser is not open');
    if (!win.webContents || win.webContents.isDestroyed?.()) throw new Error('Doubao account browser is not open');
    return win.webContents;
  }

  async ensureWebContents(accountId) {
    const safeId = safeAccountId(accountId);
    let win = this.windows.get(safeId);
    if (!win || win.isDestroyed()) win = this.open(safeId);
    const loading = this.loadPromises.get(safeId);
    if (loading) await loading;
    if (!win || win.isDestroyed() || !win.webContents || win.webContents.isDestroyed?.()) {
      throw new Error('Doubao account browser is not open');
    }
    return win.webContents;
  }

  closeAll() {
    for (const win of this.windows.values()) {
      if (!win.isDestroyed()) win.close();
    }
    this.windows.clear();
    this.loadPromises.clear();
  }
}

function safeAccountId(value) {
  const safe = String(value || '').replace(/[^a-zA-Z0-9_-]/g, '').slice(0, 64);
  if (!safe) throw new Error('invalid account id');
  return safe;
}

module.exports = { AccountWindows, safeAccountId, DEFAULT_DOUBAO_URL };
