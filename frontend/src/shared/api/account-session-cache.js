export const ACCOUNT_PREVIEW_KEY = 'auth_account_preview';
const PREVIEW_TTL_MS = 24 * 60 * 60 * 1000;
const VERIFIED_TTL_MS = 10 * 1000;
const displayFields = ['username', 'displayName', 'avatarUrl', 'themeId', 'windowBackgroundId', 'windowBackgroundUrl'];

// Presentation only. Never persist credentials, permissions or provider settings.
export function createAccountSessionCache({ storage, request, now = Date.now }) {
  let verified = null, flight = null, generation = 0;
  const token = () => storage.getItem('auth_token') || '';
  function savePreview(account, expectedToken = token()) {
    if (!expectedToken || token() !== expectedToken || account?.username !== storage.getItem('auth_username')) return;
    const preview = {};
    for (const field of displayFields) {
      if (typeof account[field] === 'string' && account[field].length <= 4096) preview[field] = account[field];
    }
    try { storage.setItem(ACCOUNT_PREVIEW_KEY, JSON.stringify({ version: 1, savedAt: now(), account: preview })); } catch { /* Storage quota must not block login. */ }
  }
  function readPreview() {
    try {
      if (!token()) return null;
      const snapshot = JSON.parse(storage.getItem(ACCOUNT_PREVIEW_KEY) || 'null');
      const age = now() - snapshot?.savedAt;
      if (snapshot?.version !== 1 || !Number.isFinite(age) || age < 0 || age > PREVIEW_TTL_MS || snapshot.account?.username !== storage.getItem('auth_username')) return null;
      const account = {};
      for (const field of displayFields) if (typeof snapshot.account[field] === 'string') account[field] = snapshot.account[field];
      return account;
    } catch { return null; }
  }
  function invalidate() {
    generation++;
    verified = null;
    flight = null;
    try { storage.removeItem(ACCOUNT_PREVIEW_KEY); } catch { /* Ignore unavailable storage. */ }
  }
  function load({ force = false } = {}) {
    if (force) { generation++; verified = null; flight = null; }
    const sessionToken = token(), epoch = generation;
    if (!sessionToken) return Promise.reject(new Error('请先登录'));
    if (!force && verified?.token === sessionToken && now() - verified.at < VERIFIED_TTL_MS) return Promise.resolve(verified.account);
    if (flight?.token === sessionToken) return flight.promise;
    const record = { token: sessionToken };
    flight = record;
    record.promise = Promise.resolve(request()).then(account => {
      if (epoch !== generation || token() !== sessionToken) throw new Error('账号会话已切换');
      verified = { token: sessionToken, at: now(), account };
      savePreview(account, sessionToken);
      return account;
    }).finally(() => { if (flight === record) flight = null; });
    return record.promise;
  }
  return { readPreview, savePreview, invalidate, load };
}
