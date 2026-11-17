import { useEffect, useState } from 'react';
import { getCachedAccount, getCurrentAccount } from '../api/auth';
import { WindowVideoBackground } from '../components/WindowVideoBackground';
import { loadCachedTheme, selectThemePreference } from '../styles/theme-preference.js';
import { resolveSchemeId, resolveThemeId } from '../styles/theme';
import { normalizeWindowBackgroundId, windowBackgroundStyle, windowVideoDefinition } from '../styles/window-background.js';

// Fullscreen editors share appearance without inheriting the workbench navigation.
export function StandaloneThemeWindow({ children }) {
  const [account, setAccount] = useState(getCachedAccount);
  const [systemMode, setSystemMode] = useState(() => window.matchMedia?.('(prefers-color-scheme: light)').matches ? 'light' : 'dark');
  const theme = selectThemePreference({ loggedIn: Boolean(account), accountThemeId: account?.themeId, cachedThemeId: loadCachedTheme() });
  const mode = resolveThemeId(theme, systemMode);
  const scheme = resolveSchemeId(theme);
  const background = normalizeWindowBackgroundId(account?.windowBackgroundId, account?.windowBackgroundUrl);
  useEffect(() => {
    let active = true;
    getCurrentAccount().then(value => { if (active) setAccount(value); }).catch(() => {});
    const query = window.matchMedia?.('(prefers-color-scheme: light)');
    const change = event => setSystemMode(event.matches ? 'light' : 'dark');
    query?.addEventListener('change', change);
    return () => { active = false; query?.removeEventListener('change', change); };
  }, []);
  useEffect(() => {
    document.documentElement.dataset.theme = mode;
    document.documentElement.dataset.scheme = scheme;
  }, [mode, scheme]);
  return <div className={`standalone-theme-window${windowVideoDefinition(background) ? ' has-video-background' : ''}`} style={windowBackgroundStyle(background, account?.windowBackgroundUrl || '', mode)}>
    <WindowVideoBackground backgroundId={background} />
    <div className="standalone-theme-content">{children}</div>
  </div>;
}
