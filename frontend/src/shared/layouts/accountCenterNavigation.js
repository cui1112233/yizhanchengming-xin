export const ACCOUNT_CENTER_ROUTES = [
  '/member',
  '/profile',
  '/security',
  '/advanced-team-admin',
  '/api-config',
  '/usage',
  '/team',
  '/accounts'
];

export function isAccountCenterPath(pathname) {
  return ACCOUNT_CENTER_ROUTES.includes(pathname);
}

export function normalizeAccountCenterReturnPath(value, origin) {
  if (!value || !origin) return '/';
  try {
    const url = new URL(value, origin);
    if (url.origin !== origin || isAccountCenterPath(url.pathname)) return '/';
    return `${url.pathname}${url.search}${url.hash}`;
  } catch {
    return '/';
  }
}

export function captureAccountCenterOrigin(currentUrl, existingOrigin, origin) {
  return normalizeAccountCenterReturnPath(currentUrl, origin);
}
