const {
  PRIMARY_USER,
  rateLimitMap,
  RATE_LIMIT_MAX,
  RATE_LIMIT_WINDOW,
  createAuthRuntime
} = require('../lib/shared');
const { getPersistentSession, revokePersistentSession } = require('../lib/session-store');

const GO_AUTH_BASE_URL = String(process.env.QIANTIE_GO_AUTH_BASE_URL || 'http://127.0.0.1:4001').replace(/\/$/, '');

async function resolveGoSession(token) {
  if (!token || !GO_AUTH_BASE_URL || typeof fetch !== 'function') return null;
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), 1500);
  try {
    const response = await fetch(`${GO_AUTH_BASE_URL}/api/login/session`, {
      method: 'GET',
      headers: { Authorization: `Bearer ${token}`, 'X-Qiantie-Auth-Bridge': '1' },
      signal: controller.signal
    });
    if (!response.ok) return null;
    const payload = await response.json();
    if (!payload?.active || typeof payload.username !== 'string' || !payload.username.trim()) return null;
    return { username: payload.username.trim(), isOwner: payload.isOwner === true };
  } catch (_) {
    return null;
  } finally {
    clearTimeout(timer);
  }
}

function getRuntime(req) {
  // Mounted sub-applications inherit their runtime from the parent app.
  for (let app = req.app; app; app = app.parent) {
    if (app.locals?.authRuntime) return app.locals.authRuntime;
  }
  return createAuthRuntime();
}

function checkRateLimit(ip) {
  const now = Date.now();
  const entry = rateLimitMap.get(ip);
  if (!entry || now > entry.resetTime) {
    rateLimitMap.set(ip, { count: 1, resetTime: now + RATE_LIMIT_WINDOW });
    return true;
  }
  if (entry.count >= RATE_LIMIT_MAX) return false;
  entry.count++;
  return true;
}

async function authenticateRequest(req) {
  const authHeader = req.headers['authorization'] || '';
  const match = authHeader.match(/^Bearer\s+(.+)$/i);
  if (!match) return null;
  const runtime = getRuntime(req);
  const token = match[1];
  const storedSession = runtime.tokenMap.get(token);
  const persistentSession = !storedSession ? getPersistentSession(runtime.sessionsPath, token) : null;
  const username = storedSession?.username || persistentSession?.username;
  let bridgeIdentity = null;
  if (typeof username !== 'string') bridgeIdentity = await resolveGoSession(token);
  const resolvedUsername = username || bridgeIdentity?.username;
  if (typeof resolvedUsername !== 'string') return null;

  try {
    const account = runtime.accountStore.getAccount(resolvedUsername);
    if (!account || !account.active || account.pending === true) {
      runtime.tokenMap.delete(token);
      revokePersistentSession(runtime.sessionsPath, token);
      return null;
    }
    return { username: account.username, account, effectivePermissions: runtime.accountStore.effectivePermissions(account) };
  } catch (error) {
    return null;
  }
}

function applyAuthentication(req, auth) {
  req.username = auth.username;
  req.auth = auth;
}

async function apiAuth(req, res, next) {
  const auth = await authenticateRequest(req);
  if (!auth) {
    res.set('X-Qiantie-Auth-Failure', 'session');
    return res.status(401).json({ error: 'Unauthorized — invalid or expired token' });
  }
  applyAuthentication(req, auth);
  next();
}

async function optionalApiAuth(req, res, next) {
  const auth = await authenticateRequest(req);
  if (auth) applyAuthentication(req, auth);
  next();
}

function requireCapability(capability, getScope = () => '*') {
  return (req, res, next) => {
    const runtime = getRuntime(req);
    try {
      if (req.auth && req.app?.locals?.memberStore?.effectiveRole(req.auth.username) === 'dev') return next();
    } catch {
      // Keep legacy account capability checks as the safe fallback.
    }
    const adminAccess = capability !== 'account:review' && capability !== 'admin:access'
      && runtime.accountStore.can(req.auth?.username, 'admin:access', '*');
    if (!req.auth || (!runtime.accountStore.can(req.auth.username, capability, getScope(req)) && !adminAccess)) {
      return res.status(403).json({ error: 'Forbidden' });
    }
    next();
  };
}

function requireOwner(req, res, next) {
  const runtime = getRuntime(req);
  const account = req.auth && runtime.accountStore.getAccount(req.auth.username);
  try {
    if (account?.active && req.app?.locals?.memberStore?.effectiveRole(account.username) === 'dev') return next();
  } catch {
    // Keep the existing Owner-only behaviour if membership data cannot be read.
  }
  if (!account || !account.active || account.username !== PRIMARY_USER || !account.isOwner) {
    return res.status(403).json({ error: 'Forbidden' });
  }
  next();
}

module.exports = { checkRateLimit, apiAuth, optionalApiAuth, requireCapability, requireOwner };
