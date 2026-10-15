const fs = require('node:fs');
const path = require('node:path');

class AccountStore {
  constructor({ filePath, fsImpl = fs }) {
    if (!filePath) throw new Error('filePath is required');
    this.filePath = filePath;
    this.fs = fsImpl;
  }

  load() {
    let raw;
    try {
      raw = this.fs.readFileSync(this.filePath, 'utf8');
    } catch (error) {
      if (error?.code === 'ENOENT') return [];
      throw error;
    }
    const data = JSON.parse(raw);
    if (!Array.isArray(data.accounts)) throw new Error('invalid account registry');
    return data.accounts.map(normalizeAccount);
  }

  save(accounts) {
    const safeAccounts = (Array.isArray(accounts) ? accounts : []).map(normalizeAccount);
    this.fs.mkdirSync(path.dirname(this.filePath), { recursive: true });
    const tempPath = `${this.filePath}.tmp`;
    this.fs.writeFileSync(tempPath, JSON.stringify({ version: 1, accounts: safeAccounts }, null, 2), {
      encoding: 'utf8',
      mode: 0o600
    });
    this.fs.renameSync(tempPath, this.filePath);
    return safeAccounts;
  }
}

function normalizeAccount(account) {
  const id = String(account?.id || '').trim();
  if (!id) throw new Error('account id is required');
  return {
    id,
    name: String(account?.name || id),
    state: normalizeState(account?.state)
  };
}

function normalizeState(state) {
  const value = String(state || 'auth_required');
  if (value === 'busy') return 'available';
  const allowed = new Set(['available', 'quota_exhausted', 'auth_required', 'human_verification', 'cooldown', 'disabled']);
  return allowed.has(value) ? value : 'auth_required';
}

module.exports = { AccountStore, normalizeAccount };
