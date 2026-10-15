const fs = require('node:fs');
const fsp = require('node:fs/promises');
const path = require('node:path');
const crypto = require('node:crypto');
const { spawn } = require('node:child_process');
const { Readable } = require('node:stream');
const { pipeline } = require('node:stream/promises');
const { normalizeChannel } = require('../update-preferences-store');

const UPDATE_INSTALLER_RE = /^yizhan-local-executor-v88-[0-9]+\.[0-9]+\.[0-9]+-win-x64\.exe$/;

class UpdateSecurityError extends Error {
  constructor(code, message) {
    super(message);
    this.name = 'UpdateSecurityError';
    this.code = code;
  }
}

class UpdateManager {
  constructor({
    currentVersion,
    platform = process.platform,
    arch = process.arch,
    updateBaseUrl = '',
    userDataDir,
    currentTask = () => null,
    preferencesStore,
    fetchJson = defaultFetchJson,
    downloadFile = defaultDownloadFile,
    hashFile = sha256File,
    spawnInstaller = defaultSpawnInstaller,
    quitApp = () => {},
    onState = () => {}
  } = {}) {
    if (!currentVersion) throw new Error('currentVersion is required');
    if (!userDataDir) throw new Error('userDataDir is required');
    if (!preferencesStore) throw new Error('preferencesStore is required');
    this.currentVersion = String(currentVersion).trim();
    this.platform = String(platform || '').trim();
    this.arch = String(arch || '').trim();
    this.updateBaseUrl = updateBaseUrl;
    this.userDataDir = userDataDir;
    this.currentTask = currentTask;
    this.preferencesStore = preferencesStore;
    this.fetchJson = fetchJson;
    this.downloadFile = downloadFile;
    this.hashFile = hashFile;
    this.spawnInstaller = spawnInstaller;
    this.quitApp = quitApp;
    this.onState = onState;
    this.manifest = null;
    this.downloaded = null;
    let channel = 'beta';
    try { channel = normalizeChannel(this.preferencesStore.load()?.channel); } catch {}
    this.state = {
      status: 'idle',
      currentVersion: this.currentVersion,
      channel,
      availableVersion: null,
      downloadedVersion: null,
      message: '尚未检查更新',
      errorCode: null,
      updateBaseUrl: ''
    };
  }

  getState() {
    return { ...this.state };
  }

  emit(patch = {}) {
    this.state = { ...this.state, ...patch };
    this.onState(this.getState());
    return this.getState();
  }

  resolveBaseUrl() {
    const raw = typeof this.updateBaseUrl === 'function' ? this.updateBaseUrl() : this.updateBaseUrl;
    return validateUpdateBaseUrl(raw);
  }

  async setChannel(channel) {
    const normalized = normalizeChannel(channel);
    this.preferencesStore.save({ channel: normalized });
    this.manifest = null;
    this.downloaded = null;
    return this.emit({
      status: 'idle',
      channel: normalized,
      availableVersion: null,
      downloadedVersion: null,
      message: normalized === 'beta' ? '已切换到测试版通道' : '已切换到稳定版通道',
      errorCode: null
    });
  }

  async checkForUpdates({ autoDownload = false } = {}) {
    let baseUrl;
    try {
      baseUrl = this.resolveBaseUrl();
    } catch (error) {
      if (error instanceof UpdateSecurityError) {
        return this.emit({
          status: 'blocked',
          message: error.message,
          errorCode: error.code,
          updateBaseUrl: ''
        });
      }
      throw error;
    }

    this.emit({
      status: 'checking',
      message: '正在检查更新…',
      errorCode: null,
      updateBaseUrl: baseUrl
    });

    try {
      const manifestUrl = `${baseUrl}/${this.state.channel}/manifest.json`;
      const raw = await this.fetchJson(manifestUrl);
      const manifest = parseUpdateManifest(raw, {
        channel: this.state.channel,
        platform: this.platform,
        arch: this.arch
      });
      if (compareVersions(manifest.version, this.currentVersion) <= 0) {
        this.manifest = null;
        this.downloaded = null;
        return this.emit({
          status: 'up_to_date',
          availableVersion: null,
          downloadedVersion: null,
          message: `当前已是最新版本 ${this.currentVersion}`,
          errorCode: null
        });
      }
      this.manifest = manifest;
      this.downloaded = null;
      this.emit({
        status: 'available',
        availableVersion: manifest.version,
        downloadedVersion: null,
        message: `发现新版本 ${manifest.version}`,
        errorCode: null
      });
      if (autoDownload) return this.downloadAvailable();
      return this.getState();
    } catch (error) {
      return this.emit({
        status: 'error',
        message: error?.message || String(error),
        errorCode: error?.code || 'UPDATE_CHECK_FAILED'
      });
    }
  }

  async downloadAvailable() {
    if (!this.manifest) throw new Error('no update manifest is available');
    const baseUrl = this.resolveBaseUrl();
    const manifest = this.manifest;
    const updateDir = path.join(this.userDataDir, 'updates');
    const targetPath = safeUpdateTarget(updateDir, manifest.file);
    await fsp.mkdir(updateDir, { recursive: true });
    await cleanupOldUpdateInstallers(updateDir, manifest.file);
    await removeIfExists(targetPath);
    this.emit({
      status: 'downloading',
      message: `正在后台下载 ${manifest.version}…`,
      errorCode: null
    });

    try {
      const fileUrl = `${baseUrl}/${this.state.channel}/${encodeURIComponent(manifest.file)}`;
      const result = await this.downloadFile(fileUrl, targetPath);
      const actualSize = Number(result?.size ?? (await fsp.stat(targetPath)).size);
      if (manifest.size > 0 && actualSize !== manifest.size) {
        throw new UpdateSecurityError('UPDATE_SIZE_MISMATCH', '更新安装包大小校验失败，已拒绝安装');
      }
      const actualHash = String(await this.hashFile(targetPath)).toLowerCase();
      if (actualHash !== manifest.sha256) {
        throw new UpdateSecurityError('UPDATE_HASH_MISMATCH', '更新安装包 SHA-256 校验失败，已拒绝安装');
      }
      this.downloaded = {
        version: manifest.version,
        filePath: targetPath,
        sha256: actualHash
      };
      return this.emit({
        status: this.currentTask() ? 'install_deferred' : 'downloaded',
        downloadedVersion: manifest.version,
        message: this.currentTask()
          ? `新版本 ${manifest.version} 已下载；当前任务结束后可更新`
          : `新版本 ${manifest.version} 已下载并校验完成`,
        errorCode: null
      });
    } catch (error) {
      await removeIfExists(targetPath);
      this.downloaded = null;
      return this.emit({
        status: 'error',
        downloadedVersion: null,
        message: error?.message || String(error),
        errorCode: error?.code || 'UPDATE_DOWNLOAD_FAILED'
      });
    }
  }

  refreshTaskState() {
    if (!this.downloaded) return this.getState();
    const busy = Boolean(this.currentTask());
    if (busy && this.state.status === 'downloaded') {
      return this.emit({
        status: 'install_deferred',
        message: `新版本 ${this.downloaded.version} 已下载；当前任务结束后可更新`
      });
    }
    if (!busy && this.state.status === 'install_deferred') {
      return this.emit({
        status: 'downloaded',
        message: `新版本 ${this.downloaded.version} 已下载并校验完成`
      });
    }
    return this.getState();
  }

  async installDownloaded() {
    if (!this.downloaded?.filePath) throw new Error('no verified update installer is downloaded');
    if (this.platform !== 'win32') throw new Error('automatic installation is currently supported on Windows only');
    if (this.currentTask()) {
      this.emit({
        status: 'install_deferred',
        message: `新版本 ${this.downloaded.version} 已下载；正在执行 VIDEO 任务，暂不重启`
      });
      return { deferred: true };
    }
    this.emit({
      status: 'installing',
      message: `正在安装 ${this.downloaded.version}，执行器将重新启动…`,
      errorCode: null
    });
    await this.spawnInstaller(this.downloaded.filePath, ['--updated', '/S', '--force-run']);
    this.quitApp();
    return { installing: true };
  }

  markDownloadedForTest(candidate) {
    this.downloaded = { ...candidate };
    this.emit({
      status: 'downloaded',
      availableVersion: candidate.version,
      downloadedVersion: candidate.version,
      message: `新版本 ${candidate.version} 已下载并校验完成`,
      errorCode: null
    });
  }
}

function validateUpdateBaseUrl(value) {
  const text = String(value || '').trim().replace(/\/+$/, '');
  if (!text) {
    throw new UpdateSecurityError('UPDATE_ORIGIN_NOT_CONFIGURED', '自动更新源尚未配置 HTTPS 地址');
  }
  let url;
  try { url = new URL(text); } catch {
    throw new UpdateSecurityError('UPDATE_ORIGIN_INVALID', '自动更新源地址无效');
  }
  if (url.protocol !== 'https:') {
    throw new UpdateSecurityError('UPDATE_HTTPS_REQUIRED', '自动更新要求 HTTPS；当前不会下载或执行安装包');
  }
  if (url.username || url.password) {
    throw new UpdateSecurityError('UPDATE_ORIGIN_CREDENTIALS_FORBIDDEN', '自动更新源地址不能包含凭据');
  }
  url.hash = '';
  url.search = '';
  return url.toString().replace(/\/$/, '');
}

function parseUpdateManifest(raw, expected = {}) {
  const manifest = raw && typeof raw === 'object' ? raw : {};
  if (manifest.schemaVersion !== 1) throw new Error('unsupported update manifest schema');
  const channel = normalizeChannel(manifest.channel);
  if (channel !== normalizeChannel(expected.channel)) throw new Error('update manifest channel mismatch');
  if (String(manifest.platform) !== String(expected.platform)) throw new Error('update manifest platform mismatch');
  if (String(manifest.arch) !== String(expected.arch)) throw new Error('update manifest architecture mismatch');
  const version = normalizeVersion(manifest.version);
  const file = String(manifest.file || '').trim();
  if (!UPDATE_INSTALLER_RE.test(file) || path.basename(file) !== file) {
    throw new Error('invalid update installer filename');
  }
  const sha256 = String(manifest.sha256 || '').trim().toLowerCase();
  if (!/^[a-f0-9]{64}$/.test(sha256)) throw new Error('invalid update manifest sha256');
  const size = Number(manifest.size);
  if (!Number.isSafeInteger(size) || size <= 0) throw new Error('invalid update manifest size');
  const publishedAt = String(manifest.publishedAt || '').trim();
  if (!publishedAt || Number.isNaN(Date.parse(publishedAt))) throw new Error('invalid update manifest publishedAt');
  return { schemaVersion: 1, channel, version, platform: String(manifest.platform), arch: String(manifest.arch), file, sha256, size, publishedAt };
}

function compareVersions(a, b) {
  const left = versionParts(a);
  const right = versionParts(b);
  for (let i = 0; i < 3; i++) {
    if (left[i] > right[i]) return 1;
    if (left[i] < right[i]) return -1;
  }
  return 0;
}

function normalizeVersion(value) {
  const text = String(value || '').trim();
  if (!/^\d+\.\d+\.\d+$/.test(text)) throw new Error('invalid update version');
  return text;
}

function versionParts(value) {
  return normalizeVersion(value).split('.').map(Number);
}

function safeUpdateTarget(root, file) {
  const base = path.resolve(root);
  const resolved = path.resolve(root, file);
  if (path.dirname(resolved) !== base) throw new UpdateSecurityError('UPDATE_PATH_INVALID', 'invalid update installer path');
  return resolved;
}

async function cleanupOldUpdateInstallers(updateDir, keepFile) {
  let entries;
  try {
    entries = await fsp.readdir(updateDir, { withFileTypes: true });
  } catch (error) {
    if (error?.code === 'ENOENT') return;
    throw error;
  }
  await Promise.all(entries
    .filter(entry => entry.isFile() && UPDATE_INSTALLER_RE.test(entry.name) && entry.name !== keepFile)
    .map(entry => removeIfExists(path.join(updateDir, entry.name))));
}

async function defaultFetchJson(url) {
  const response = await fetch(url, { headers: { accept: 'application/json' }, redirect: 'error' });
  if (!response.ok) throw new Error(`update manifest request failed: HTTP ${response.status}`);
  return response.json();
}

async function defaultDownloadFile(url, filePath) {
  const response = await fetch(url, { redirect: 'error' });
  if (!response.ok || !response.body) throw new Error(`update download failed: HTTP ${response.status}`);
  const stream = fs.createWriteStream(filePath, { flags: 'wx' });
  await pipeline(Readable.fromWeb(response.body), stream);
  const stat = await fsp.stat(filePath);
  return { size: stat.size };
}

async function sha256File(filePath) {
  const hash = crypto.createHash('sha256');
  await pipeline(fs.createReadStream(filePath), hash);
  return hash.digest('hex');
}

async function defaultSpawnInstaller(filePath, args) {
  const child = spawn(filePath, args, { detached: true, stdio: 'ignore', windowsHide: true });
  child.unref();
}

async function removeIfExists(filePath) {
  try { await fsp.unlink(filePath); } catch (error) { if (error?.code !== 'ENOENT') throw error; }
}

module.exports = {
  UpdateManager,
  UpdateSecurityError,
  validateUpdateBaseUrl,
  parseUpdateManifest,
  compareVersions,
  safeUpdateTarget,
  cleanupOldUpdateInstallers,
  sha256File
};
