import { Button, ConfigProvider, Result, Spin, message } from 'antd';
import { useEffect, useState } from 'react';
import { BrandLogo } from '../components/BrandLogo';
import { Link } from '../components/Link';
import { ThemePicker } from '../components/ThemePicker';
import { getCurrentAccount, logout } from '../api/auth';
import { updateMemberProfile } from '../api/member';
import { createAntTheme, resolveSchemeId, resolveThemeId } from '../styles/theme';
import { loadCachedTheme, persistCachedTheme, selectThemePreference } from '../styles/theme-preference.js';

const navItems = [
  { href: '/admin', icon: '▦', label: '概览' },
  { href: '/admin/presets', icon: '≡', label: '提示词库' },
  { href: '/admin/prompts', icon: '⌁', label: 'Prompt 策略' },
  { href: '/admin/agent-skills', icon: '◇', label: 'CM 平台技能' },
  { href: '/admin/shuihuo-models', icon: '◇', label: '水货生产模型' },
  { href: '/admin/accounts', icon: '◎', label: '账号与授权' },
  { href: '/admin/error-logs', icon: '!', label: '错误日志' }
];

function initialTheme() {
  return loadCachedTheme();
}

function initialSystemMode() {
  return window.matchMedia?.('(prefers-color-scheme: light)').matches ? 'light' : 'dark';
}

function pageTitle(pathname) {
  return navItems.find(item => item.href === pathname)?.label || '管理后台';
}

export function AdminLayout({ children }) {
  const [account, setAccount] = useState(null);
  const [accessState, setAccessState] = useState('checking');
  const [theme, setTheme] = useState(initialTheme);
  const [systemMode, setSystemMode] = useState(initialSystemMode);
  const pathname = window.location.pathname;
  const resolvedTheme = resolveThemeId(theme, systemMode);
  const themeScheme = resolveSchemeId(theme);

  useEffect(() => {
    getCurrentAccount().then(nextAccount => {
      setAccount(nextAccount);
      const accountTheme = selectThemePreference({ loggedIn: true, accountThemeId: nextAccount.themeId, cachedThemeId: theme });
      setTheme(accountTheme);
      persistCachedTheme(accountTheme);
      if (nextAccount?.role === 'dev') {
        setAccessState('allowed');
        return;
      }
      setAccessState('denied');
      window.location.replace('/profile');
    }).catch(() => {
      setAccessState('denied');
      window.location.replace('/');
    });
  }, []);

  useEffect(() => {
    document.documentElement.dataset.theme = resolvedTheme;
    document.documentElement.dataset.scheme = themeScheme;
  }, [resolvedTheme, themeScheme]);

  useEffect(() => {
    const query = window.matchMedia?.('(prefers-color-scheme: light)');
    if (!query) return undefined;
    const receive = event => setSystemMode(event.matches ? 'light' : 'dark');
    query.addEventListener?.('change', receive);
    return () => query.removeEventListener?.('change', receive);
  }, []);

  useEffect(() => {
    document.documentElement.dataset.shell = 'admin';
    document.body.classList.add('admin-theme-active');
    return () => {
      document.body.classList.remove('admin-theme-active');
      delete document.documentElement.dataset.shell;
      delete document.documentElement.dataset.theme;
      delete document.documentElement.dataset.scheme;
    };
  }, []);

  async function handleLogout() {
    await logout();
    window.location.assign('/');
  }

  async function changeTheme(themeId) {
    const nextThemeId = persistCachedTheme(themeId);
    setTheme(nextThemeId);
    try {
      const result = await updateMemberProfile({ themeId: nextThemeId });
      setAccount(current => current ? { ...current, themeId: result.member?.themeId || nextThemeId } : current);
    } catch {
      message.warning('主题已在本机切换，但未同步到账号');
    }
  }

  if (accessState === 'checking') {
    return <ConfigProvider theme={createAntTheme(theme, systemMode)}><div className="admin-access-screen"><Spin size="large" tip="正在验证开发者权限" /></div></ConfigProvider>;
  }

  if (accessState !== 'allowed') {
    return <ConfigProvider theme={createAntTheme(theme, systemMode)}><div className="admin-access-screen"><Result status="403" title="仅 DEV 可访问管理后台" subTitle="账号与组员授权请在个人中心完成。" /></div></ConfigProvider>;
  }

  return (
    <ConfigProvider theme={createAntTheme(theme, systemMode)}>
      <div className="admin-shell">
      <aside className="admin-sidebar">
        <div className="admin-brand"><BrandLogo className="admin-brand-logo" /><span>管理后台</span></div>
        <p className="admin-nav-group">运营管理</p>
        <nav className="admin-nav">
          {navItems.map(item => <Link key={item.href} href={item.href} className={pathname === item.href ? 'active' : ''}><span>{item.icon}</span>{item.label}</Link>)}
        </nav>
        <div className="admin-sidebar-footer">
          <Link href="/profile" reload className="admin-return-link">← 返回个人中心</Link>
          <div className="admin-user-summary"><span>{account?.username || '正在验证身份'}</span>{account?.isOwner ? <small>主管理员</small> : <small>管理员</small>}</div>
          <ThemePicker value={theme} onChange={changeTheme} />
          <Button size="small" onClick={handleLogout}>退出登录</Button>
        </div>
      </aside>
      <main className="admin-main">
        <header className="admin-topbar"><span>{pageTitle(pathname)}</span></header>
        <section className="admin-content">{children}</section>
      </main>
      </div>
    </ConfigProvider>
  );
}
