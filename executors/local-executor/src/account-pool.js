const ELIGIBLE = new Set(['available']);

class AccountPool {
  constructor(accounts = []) {
    this.accounts = new Map();
    this.order = [];
    this.cursor = 0;
    for (const account of accounts) this.upsert(account);
  }

  upsert(account) {
    if (!account?.id) throw new Error('account id is required');
    if (!this.accounts.has(account.id)) this.order.push(account.id);
    const previous = this.accounts.get(account.id) || {};
    this.accounts.set(account.id, {
      ...previous,
      ...account,
      state: account.state || previous.state || 'available',
      jobId: account.jobId || previous.jobId || null
    });
  }

  acquire({ excludeIds = [], jobId = null } = {}) {
    const excluded = new Set(excludeIds);
    if (this.order.length === 0) return null;
    for (let offset = 0; offset < this.order.length; offset += 1) {
      const index = (this.cursor + offset) % this.order.length;
      const account = this.accounts.get(this.order[index]);
      if (!account) continue;
      if (!excluded.has(account.id) && ELIGIBLE.has(account.state) && !account.jobId) {
        account.state = 'busy';
        account.jobId = jobId;
        this.cursor = (index + 1) % this.order.length;
        return { ...account };
      }
    }
    return null;
  }

  release(id, nextState = 'available') {
    const account = this.accounts.get(id);
    if (!account) return false;
    account.state = nextState;
    account.jobId = null;
    return true;
  }

  setState(id, state) {
    const account = this.accounts.get(id);
    if (!account) return false;
    account.state = state;
    if (state !== 'busy') account.jobId = null;
    return true;
  }

  list() { return [...this.accounts.values()].map(item => ({ ...item })); }

  summary() {
    const items = this.list();
    return {
      total: items.length,
      available: items.filter(x => x.state === 'available').length,
      busy: items.filter(x => x.state === 'busy').length,
      quotaExhausted: items.filter(x => x.state === 'quota_exhausted').length,
      loginError: items.filter(x => x.state === 'auth_required').length,
      humanVerification: items.filter(x => x.state === 'human_verification').length
    };
  }
}

module.exports = { AccountPool };
