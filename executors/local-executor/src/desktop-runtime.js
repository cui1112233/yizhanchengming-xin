const { ExecutorApiClient } = require('./api-client');
const { AccountPool } = require('./account-pool');
const { JobRunner } = require('./job-runner');

class DesktopRuntime {
  constructor({
    deviceStore,
    apiFactory = baseUrl => new ExecutorApiClient({ baseUrl }),
    deviceName,
    platform,
    version,
    accounts = [],
    adapter = null,
    runnerOptions = {}
  }) {
    if (!deviceStore) throw new Error('deviceStore is required');
    this.deviceStore = deviceStore;
    this.apiFactory = apiFactory;
    this.deviceName = String(deviceName || 'YiZhan Executor');
    this.os = normalizeOs(platform);
    this.version = String(version || '0.0.0');
    this.accountPool = new AccountPool(accounts);
    this.adapter = adapter;
    this.runnerOptions = { ...runnerOptions };
    this.credential = null;
    this.api = null;
    this.automationEnabled = false;
    this.currentTask = null;
    this.claimPromise = null;
    this.lastHeartbeatAt = null;
    this.lastError = null;
    this.restorePairing();
  }

  restorePairing() {
    const stored = this.deviceStore.load();
    if (!stored) return null;
    if (!stored.baseUrl || !stored.executorId || !stored.token) throw new Error('stored pairing is incomplete');
    this.credential = {
      baseUrl: normalizeBaseUrl(stored.baseUrl),
      executorId: String(stored.executorId),
      token: String(stored.token)
    };
    this.api = this.apiFactory(this.credential.baseUrl);
    return this.getState().pairing;
  }

  async pair({ baseUrl, code }) {
    const normalizedBaseUrl = normalizeBaseUrl(baseUrl);
    const normalizedCode = String(code || '').trim();
    if (!normalizedCode) throw new Error('pairing code is required');
    const api = this.apiFactory(normalizedBaseUrl);
    const response = await api.pair({
      code: normalizedCode,
      deviceName: this.deviceName,
      platform: 'doubao',
      os: this.os,
      version: this.version
    });
    if (!response?.executorId || !response?.token) throw new Error('pair response is incomplete');
    this.deviceStore.save({
      baseUrl: normalizedBaseUrl,
      executorId: response.executorId,
      token: response.token
    });
    this.credential = {
      baseUrl: normalizedBaseUrl,
      executorId: String(response.executorId),
      token: String(response.token)
    };
    this.api = api;
    this.lastError = null;
    return this.getState();
  }

  installAdapter(adapter) {
    if (!adapter) throw new Error('live Doubao adapter is required');
    this.adapter = adapter;
    return this.getState();
  }

  addAccount({ id, name }) {
    const accountId = String(id || '').trim();
    if (!accountId) throw new Error('account id is required');
    this.accountPool.upsert({
      id: accountId,
      name: String(name || accountId),
      state: 'auth_required'
    });
    return this.getAccount(accountId);
  }

  markAccountAvailable(id) {
    if (!this.accountPool.setState(id, 'available')) throw new Error('account not found');
    return this.getAccount(id);
  }

  setAccountState(id, state) {
    if (!this.accountPool.setState(id, state)) throw new Error('account not found');
    return this.getAccount(id);
  }

  getAccount(id) {
    return this.accountPool.list().find(item => item.id === id) || null;
  }

  setAutomationEnabled(enabled) {
    if (!enabled) {
      this.automationEnabled = false;
      return this.getState();
    }
    if (!this.adapter) throw new Error('live Doubao adapter is not installed; automation stays disabled');
    if (!this.credential || !this.api) throw new Error('pair V78 before enabling automation');
    if (this.accountPool.summary().available < 1) throw new Error('at least one available Doubao account is required');
    this.automationEnabled = true;
    this.lastError = null;
    return this.getState();
  }

  async claimOnce() {
    if (!this.automationEnabled || !this.adapter || !this.credential || !this.api) return null;
    if (this.accountPool.summary().available < 1) return null;
    if (this.claimPromise) return this.claimPromise;

    this.claimPromise = (async () => {
      const claim = await this.api.claim(this.credential.token);
      if (!claim?.job?.id) return null;
      this.currentTask = {
        id: claim.job.id,
        sourceTaskId: claim.job.sourceTaskId || null,
        state: claim.job.state || 'leased'
      };
      const runner = new JobRunner({
        api: this.api,
        token: this.credential.token,
        accountPool: this.accountPool,
        adapter: this.adapter,
        ...this.runnerOptions
      });
      try {
        const result = await runner.runClaim(claim);
        this.lastError = null;
        return result;
      } finally {
        this.currentTask = null;
      }
    })();

    try {
      return await this.claimPromise;
    } catch (error) {
      this.recordError(error);
      throw error;
    } finally {
      this.claimPromise = null;
    }
  }

  async heartbeat() {
    if (!this.credential || !this.api) return null;
    const response = await this.api.heartbeat(this.credential.token, {
      deviceName: this.deviceName,
      os: this.os,
      version: this.version,
      accounts: this.accountPool.summary()
    });
    this.lastHeartbeatAt = new Date().toISOString();
    return response;
  }

  recordError(error) {
    this.lastError = error?.message || String(error);
  }

  getState() {
    const summary = this.accountPool.summary();
    const automation = automationState({
      enabled: this.automationEnabled,
      paired: Boolean(this.credential && this.api),
      adapter: Boolean(this.adapter),
      availableAccounts: summary.available
    });
    return {
      pairing: this.credential
        ? { paired: true, baseUrl: this.credential.baseUrl, executorId: this.credential.executorId }
        : { paired: false, baseUrl: '', executorId: '' },
      accounts: this.accountPool.list().map(({ id, name, state, jobId }) => ({ id, name, state, jobId: jobId || null })),
      automation,
      currentTask: this.currentTask,
      lastHeartbeatAt: this.lastHeartbeatAt,
      lastError: this.lastError
    };
  }
}

function automationState({ enabled, paired, adapter, availableAccounts }) {
  if (!adapter) return { enabled: false, ready: false, reason: '真实豆包适配器尚未安装' };
  if (!paired) return { enabled: false, ready: false, reason: '请先绑定一战晟铭网页版' };
  if (availableAccounts < 1) return { enabled: Boolean(enabled), ready: false, reason: '没有可用豆包账号，请先登录或处理验证/额度问题' };
  return {
    enabled: Boolean(enabled),
    ready: true,
    reason: enabled ? '自动任务已开启，会自动领取新的 VIDEO 任务' : '已就绪，开启后会自动领取新的 VIDEO 任务'
  };
}

function normalizeBaseUrl(value) {
  const text = String(value || '').trim().replace(/\/+$/, '');
  if (!/^https?:\/\//i.test(text)) throw new Error('V78 server address must start with http:// or https://');
  return text;
}

function normalizeOs(platform) {
  if (platform === 'win32') return 'windows';
  if (platform === 'darwin') return 'macos';
  return String(platform || 'unknown');
}

module.exports = { DesktopRuntime, automationState, normalizeBaseUrl, normalizeOs };
