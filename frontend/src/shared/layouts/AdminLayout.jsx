import { Button, ConfigProvider, Result, Spin } from 'antd';
import { useEffect, useState } from 'react';
import { BrandLogo } from '../components/BrandLogo';
import { Link } from '../components/Link';
import { getCurrentAccount, logout } from '../api/auth';
import { createAntTheme } from '../styles/theme';

const navItems = [
  { href: '/admin', icon: '▦', label: '概览' },
  { href: '/admin/presets', icon: '≡', label: '提示词库' },
  { href: '/admin/prompts', icon: '⌁', label: 'Prompt 策略' },
  { href: '/admin/agent-skills', icon: '◇', label: 'CM 平台技能' },
  { href: '/admin/shuihuo-models', icon: '◇', label: '水货生产模型' },
  { href: '/admin/accounts', icon: '◎', label: '账号与授权' },
  { href: '/admin/error-logs', icon: '!', label: '错误日志' }
];

const THEME_STORAGE_KEY = 'yizhan-theme';

function initialTheme() {
  return localStorage.getItem(THEME_STORAGE_KEY) === 'light' ? 'light' : 'dark';
}

function pageTitle(pathname) {
  return navItems.find(item => item.href === pathname)?.label || '管理后台';
}

export function AdminLayout({ children }) {
  const [account, setAccount] = useState(null);
  const [accessState, setAccessState] = useState('checking');
  const [theme, setTheme] = useState(initialTheme);
  const pathname = window.location.pathname;

  useEffect(() => {
    getCurrentAccount().then(nextAccount => {
      setAccount(nextAccount);
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
    document.documentElement.dataset.theme = theme;
    localStorage.setItem(THEME_STORAGE_KEY, theme);
  }, [theme]);

  useEffect(() => {
    document.body.classList.add('admin-theme-active');
    return () => document.body.classList.remove('admin-theme-active');
  }, []);

  async function handleLogout() {
    await logout();
    window.location.assign('/');
  }

  function toggleTheme() {
    setTheme(current => current === 'dark' ? 'light' : 'dark');
  }

  if (accessState === 'checking') {
    return <ConfigProvider theme={createAntTheme(theme)}><div className="admin-access-screen"><Spin size="large" tip="正在验证开发者权限" /></div></ConfigProvider>;
  }

  if (accessState !== 'allowed') {
    return <ConfigProvider theme={createAntTheme(theme)}><div className="admin-access-screen"><Result status="403" title="仅 DEV 可访问管理后台" subTitle="账号与组员授权请在个人中心完成。" /></div></ConfigProvider>;
  }

  return (
    <ConfigProvider theme={createAntTheme(theme)}>
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
          <Button
            size="small"
            className="admin-theme-toggle"
            type="button"
            onClick={toggleTheme}
            title={theme === 'dark' ? '切换至浅色主题' : '切换至深色主题'}
            aria-label={theme === 'dark' ? '切换至浅色主题' : '切换至深色主题'}
          >
            {theme === 'dark' ? '☀ 浅色主题' : '☾ 深色主题'}
          </Button>
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
