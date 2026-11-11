export function shouldAutoCollapseAgentNavigation(pathname) {
  return pathname === '/agent' || pathname === '/agent/canvas' || pathname === '/agent/chat';
}
