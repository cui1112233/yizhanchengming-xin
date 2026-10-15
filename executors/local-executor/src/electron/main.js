const path = require('node:path');
const os = require('node:os');
const { app, BrowserWindow, ipcMain, safeStorage, session } = require('electron');
const { DeviceStore } = require('../device-store');
const { AccountStore } = require('../account-store');
const { UpdatePreferencesStore } = require('../update-preferences-store');
const { DesktopRuntime } = require('../desktop-runtime');
const { DesktopController } = require('../desktop-controller');
const { DoubaoAdapter } = require('../doubao-adapter');
const { AccountWindows } = require('./account-windows');
const { UpdateManager } = require('./update-manager');

let mainWindow = null;
let controller = null;
let updateManager = null;
let startupUpdateTimer = null;

function createMainWindow() {
  const win = new BrowserWindow({
    width: 920,
    height: 780,
    minWidth: 760,
    minHeight: 640,
    show: false,
    title: '一战晟铭豆包执行器',
    icon: path.join(__dirname, '..', '..', 'build', 'icon.png'),
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      nodeIntegration: false,
      contextIsolation: true,
      sandbox: true,
      webSecurity: true
    }
  });
  win.webContents.setWindowOpenHandler(() => ({ action: 'deny' }));
  win.once('ready-to-show', () => win.show());
  win.loadFile(path.join(__dirname, '..', '..', 'ui', 'index.html'));
  mainWindow = win;
  return win;
}

function emitState(state) {
  if (mainWindow && !mainWindow.isDestroyed()) {
    mainWindow.webContents.send('executor:state', state);
  }
  updateManager?.refreshTaskState();
}

function emitUpdateState(state) {
  if (mainWindow && !mainWindow.isDestroyed()) {
    mainWindow.webContents.send('updater:state', state);
  }
}

function registerIpc() {
  ipcMain.handle('executor:get-state', () => controller.getState());
  ipcMain.handle('executor:pair', async (_event, input) => controller.pair(input || {}));
  ipcMain.handle('executor:add-account', () => controller.addAccount());
  ipcMain.handle('executor:open-account', (_event, id) => controller.openAccount(id));
  ipcMain.handle('executor:mark-account-available', (_event, id) => controller.markAccountAvailable(id));
  ipcMain.handle('executor:set-automation-enabled', (_event, enabled) => controller.setAutomationEnabled(Boolean(enabled)));

  ipcMain.handle('updater:get-state', () => updateManager.getState());
  ipcMain.handle('updater:check', () => updateManager.checkForUpdates({ autoDownload: true }));
  ipcMain.handle('updater:set-channel', (_event, channel) => updateManager.setChannel(channel));
  ipcMain.handle('updater:install', () => updateManager.installDownloaded());
}

function buildController() {
  const userData = app.getPath('userData');
  const deviceStore = new DeviceStore({
    safeStorage,
    filePath: path.join(userData, 'device.json')
  });
  const accountStore = new AccountStore({
    filePath: path.join(userData, 'accounts.json')
  });
  let accounts = [];
  try {
    accounts = accountStore.load();
  } catch {
    accounts = [];
  }

  const accountWindows = new AccountWindows({ BrowserWindow, session });
  const adapter = new DoubaoAdapter({
    accountWindows,
    downloadDir: path.join(userData, 'returned-videos')
  });
  const runtimeOptions = {
    deviceName: os.hostname(),
    platform: process.platform,
    version: app.getVersion(),
    accounts,
    adapter
  };

  let runtime;
  try {
    runtime = new DesktopRuntime({ deviceStore, ...runtimeOptions });
  } catch (error) {
    runtime = new DesktopRuntime({
      deviceStore: {
        load: () => null,
        save: value => deviceStore.save(value)
      },
      ...runtimeOptions
    });
    runtime.recordError(error);
  }

  return new DesktopController({
    runtime,
    accountStore,
    accountWindows,
    onState: emitState
  });
}

function buildUpdateManager() {
  const userData = app.getPath('userData');
  const preferencesStore = new UpdatePreferencesStore({
    filePath: path.join(userData, 'update-preferences.json')
  });
  return new UpdateManager({
    currentVersion: app.getVersion(),
    platform: process.platform,
    arch: process.arch,
    userDataDir: userData,
    preferencesStore,
    currentTask: () => controller?.getState()?.currentTask || null,
    updateBaseUrl: () => {
      const explicit = String(process.env.YIZHAN_EXECUTOR_UPDATE_BASE_URL || '').trim();
      if (explicit) return explicit;
      const paired = String(controller?.getState()?.pairing?.baseUrl || '').replace(/\/+$/, '');
      return paired ? `${paired}/downloads/local-executor/updates` : '';
    },
    quitApp: () => app.quit(),
    onState: emitUpdateState
  });
}

function scheduleStartupUpdateCheck(delayMs = 2500) {
  if (startupUpdateTimer) clearTimeout(startupUpdateTimer);
  startupUpdateTimer = setTimeout(() => {
    updateManager?.checkForUpdates({ autoDownload: true }).catch(error => {
      console.error('[updater] startup check failed:', error?.message || error);
    });
  }, delayMs);
  startupUpdateTimer.unref?.();
}

app.whenReady().then(() => {
  controller = buildController();
  updateManager = buildUpdateManager();
  registerIpc();
  createMainWindow();
  controller.startHeartbeat(15000);
  controller.startJobPolling(2000);
  if (process.platform === 'win32' && app.isPackaged) scheduleStartupUpdateCheck();

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createMainWindow();
  });
});

app.on('before-quit', () => {
  if (startupUpdateTimer) clearTimeout(startupUpdateTimer);
  startupUpdateTimer = null;
  controller?.shutdown();
});
app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit();
});
