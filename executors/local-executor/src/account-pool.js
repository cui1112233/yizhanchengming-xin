const ELIGIBLE = new Set(['available']);

class AccountPool {
  constructor(accounts = []) {
    this.accounts = new Map();
    for (const account of accounts) this.upsert(account);
  }

  upsert(account) {
    if (!account?.id) throw new Error('account id is required');
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
    for (const account of this.accounts.values()) {
      if (!excluded.has(account.id) && ELIGIBLE.has(account.state) && !account.jobId) {
        account.state = 'busy';
        account.jobId = jobId;
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
