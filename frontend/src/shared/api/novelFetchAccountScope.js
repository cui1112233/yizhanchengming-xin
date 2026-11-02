const storagePrefix = 'novel_fetch_view_account:';

function currentLoginUsername() {
  return typeof localStorage === 'undefined' ? '' : (localStorage.getItem('auth_username') || '');
}

function storageKey() {
  return `${storagePrefix}${currentLoginUsername()}`;
}

export function getSelectedNovelFetchAccount() {
  if (typeof localStorage === 'undefined') return '';
  return localStorage.getItem(storageKey()) || currentLoginUsername();
}

export function setSelectedNovelFetchAccount(username) {
  if (typeof localStorage === 'undefined') return;
  const value = String(username || '').trim();
  if (value) localStorage.setItem(storageKey(), value);
  else localStorage.removeItem(storageKey());
}

export function isNovelFetchScopedPath(path) {
  const value = String(path || '');
  return ['/api/novel-fetch-workshop', '/api/novel-fetch/', '/api/batch-rewrite', '/api/web-submit'].some(prefix => value.startsWith(prefix));
}

export function novelFetchAccountOptions(self, teamMembers = [], globalAccounts = []) {
  const username = String(self?.username || '').trim();
  if (!username) return [];
  let rows = [self];
  if (self.role === 'dev') {
    rows = [self, ...(Array.isArray(globalAccounts) ? globalAccounts : [])];
  } else if (self.role === 'manager') {
    rows = [self, ...(Array.isArray(teamMembers) ? teamMembers.filter(item => item?.role === 'member' && item?.boundTo === username) : [])];
  }
  const unique = new Map();
  rows.forEach(item => {
    const value = String(item?.username || '').trim();
    if (value) unique.set(value, { value, label: item?.displayName && item.displayName !== value ? `${item.displayName}（${value}）` : value, role: item?.role || 'member' });
  });
  return [...unique.values()].sort((a, b) => a.value === username ? -1 : b.value === username ? 1 : a.value.localeCompare(b.value));
}
