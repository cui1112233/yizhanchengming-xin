// Keep all login entry points on the public homepage without a full reload.
export function navigateToHomeLogin(openDialog, browser = window) {
  openDialog();
  if (browser.location.pathname === '/') return;
  browser.history.replaceState({}, '', '/');
  // A child layout's mount effect runs before the app installs its route listener.
  browser.queueMicrotask(() => browser.dispatchEvent(new browser.PopStateEvent('popstate')));
}
