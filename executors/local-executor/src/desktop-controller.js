const { randomUUID } = require('node:crypto');

class DesktopController {
  constructor({ runtime, accountStore, accountWindows, onState = () => {}, makeAccountId = randomUUID }) {
    if (!runtime || !accountStore || !accountWindows) throw new Error('runtime, accountStore and accountWindows are required');
    this.runtime = runtime;
    this.accountStore = accountStore;
    this.accountWindows = accountWindows;
    this.onState = onState;
    this.makeAccountId = makeAccountId;
    this.heartbeatTimer = null;
    this.jobTimer = null;
    this.jobPollBusy = false;
  }

  getState() {
    return this.runtime.getState();
  }

  broadcast() {
    const state = this.getState();
    this.onState(state);
    return state;
  }

  async pair(input) {
    await this.runtime.pair(input);
    this.broadcast();
    return this.getState();
  }

  addAccount() {
    const id = this.makeAccountId();
    const number = this.runtime.getState().accounts.length + 1;
    const account = this.runtime.addAccount({ id, name: `豆包账号 ${number}` });
    this.persistAccounts();
    this.accountWindows.open(account.id);
    this.broadcast();
    return account;
  }

  openAccount(id) {
    const account = this.runtime.getAccount?.(id) || null;
    if (!account) throw new Error('account not found');
    this.accountWindows.open(account.id);
    return account;
  }

  markAccountAvailable(id) {
    const account = this.runtime.markAccountAvailable(id);
    this.persistAccounts();
    this.broadcast();
    return account;
  }

  setAutomationEnabled(enabled) {
    const state = this.runtime.setAutomationEnabled(Boolean(enabled));
    this.broadcast();
    if (enabled) this.pollJobOnce();
    return state;
  }

  persistAccounts() {
    this.accountStore.save(this.runtime.getState().accounts);
  }

  async heartbeatOnce() {
    try {
      await this.runtime.heartbeat();
      this.broadcast();
      return true;
    } catch (error) {
      this.runtime.recordError(error);
      this.broadcast();
      return false;
    }
  }

  startHeartbeat(intervalMs = 15000) {
    this.stopHeartbeat();
    this.heartbeatOnce();
    this.heartbeatTimer = setInterval(() => this.heartbeatOnce(), intervalMs);
    this.heartbeatTimer.unref?.();
  }

  stopHeartbeat() {
    if (this.heartbeatTimer) clearInterval(this.heartbeatTimer);
    this.heartbeatTimer = null;
  }

  async pollJobOnce() {
    if (this.jobPollBusy || !this.runtime.getState().automation.enabled) return null;
    this.jobPollBusy = true;
    try {
      const result = await this.runtime.claimOnce();
      this.persistAccounts();
      this.broadcast();
      return result;
    } catch (error) {
      this.runtime.recordError(error);
      this.persistAccounts();
      this.broadcast();
      return null;
    } finally {
      this.jobPollBusy = false;
    }
  }

  startJobPolling(intervalMs = 2000) {
    this.stopJobPolling();
    this.jobTimer = setInterval(() => this.pollJobOnce(), intervalMs);
    this.jobTimer.unref?.();
  }

  stopJobPolling() {
    if (this.jobTimer) clearInterval(this.jobTimer);
    this.jobTimer = null;
  }

  shutdown() {
    this.stopHeartbeat();
    this.stopJobPolling();
    this.accountWindows.closeAll?.();
  }
}

module.exports = { DesktopController };
