import { apiRequest, getToken, setToken } from './client';
import { getPasskeyAssertion } from '../webauthn';
import { createAccountSessionCache } from './account-session-cache.js';

const accountSession = createAccountSessionCache({
  storage: { getItem: key => localStorage.getItem(key), setItem: (key, value) => localStorage.setItem(key, value), removeItem: key => localStorage.removeItem(key) },
  request: () => apiRequest('/api/login/session', { suppressGlobalError: true })
});

export function getCachedAccount() { return accountSession.readPreview(); }

function persistSession(data) {
  accountSession.invalidate();
  setToken(data.token);
  localStorage.setItem('auth_username', data.username);
  accountSession.savePreview(data);
  return data;
}

export async function login(username, password, remember = true, mfaCode = '') {
  const data = await apiRequest('/api/login', {
    method: 'POST',
    body: JSON.stringify({ username, password, remember, mfaCode }),
    suppressGlobalError: true
  });
  return persistSession(data);
}

export async function loginWithPasskey(username, remember = true) {
  const options = await apiRequest('/api/login/passkey/options', {
    method: 'POST',
    body: JSON.stringify({ username }),
    suppressGlobalError: true
  });
  const credential = await getPasskeyAssertion(options);
  const data = await apiRequest('/api/login/passkey/verify', {
    method: 'POST',
    body: JSON.stringify({ username, remember, challenge: options.challenge, credential }),
    suppressGlobalError: true
  });
  return persistSession(data);
}

export function getCurrentAccount(options) {
  return accountSession.load(options);
}

export async function logout() {
  const token = getToken();
  accountSession.invalidate();
  setToken('');
  localStorage.removeItem('auth_username');
  if (token) {
    void fetch('/api/login/logout', { method: 'POST', headers: { Authorization: `Bearer ${token}` }, keepalive: true }).catch(() => undefined);
  }
}

export function getCurrentUsername() {
  return getToken() ? localStorage.getItem('auth_username') || '' : '';
}
