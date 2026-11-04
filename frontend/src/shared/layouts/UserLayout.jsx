import { Avatar, Button, Checkbox, ConfigProvider, Form, Input, message, Modal } from 'antd';
import { AudioLines, BookOpen, Bot, ChartNoAxesCombined, Clapperboard, FilePenLine, Fingerprint, Home, LogOut, Moon, PanelLeftClose, PanelLeftOpen, Settings2, Sun } from 'lucide-react';
import { cloneElement, Fragment, isValidElement, lazy, Suspense, useEffect, useRef, useState } from 'react';
import { BrandLogo } from '../components/BrandLogo';
import { Link } from '../components/Link';
import { getCurrentAccount, getCurrentUsername, login, loginWithPasskey, logout } from '../api/auth';
import { getToken } from '../api/client';
import { getConfig } from '../api/config';
import { avatarDisplay } from '../avatars';
import { CmPenguinCompanion } from '../pet/CmPenguinCompanion';
import { dispatchPetContext } from '../pet/stacky';
import { GLOBAL_STATUS_EVENT, GLOBAL_TASK_NOTIFICATION_EVENT, normalizeGlobalTaskNotification } from '../notifications/globalTaskCenter.js';
import { createAntTheme } from '../styles/theme';
import { getRouteAccessState, shouldPromptLoginForApiFailure } from './routeAccess.js';
import { AccountCenterLayout } from './AccountCenterLayout';
import { captureAccountCenterOrigin, isAccountCenterPath, normalizeAccountCenterReturnPath } from './accountCenterNavigation.js';
import { shouldAutoCollapseAgentNavigation } from './agentNavigation.js';
import { preloadRoute } from '../../user/workbenchPreload.js';

// 动态 Logo 包含 WebGL shader，不能阻塞任何已登录业务页的首屏，按真正使用时再下载。
const SuperOpcLiquidMetalLogo = lazy(() => import('../components/SuperOpcLiquidMetalLogo'));

const navItems = [
  { href: '/', icon: Home, label: '首页' },
  { href: '/script', icon: FilePenLine, label: '私人定制' },
  // Novel Fetch is served as the Go-embedded canonical workbench. Force a
  // document navigation so the SPA cannot swap in a stale React-only page.
  { href: '/novel-fetch', icon: BookOpen, label: '小说获取', reload: true },
  { href: '/batch-factory', icon: Clapperboard, label: '水货生产' },
  { href: '/agent', icon: Bot, label: 'Agent 工作区' },
  { href: '/tts', icon: AudioLines, label: '配音' }
];

const THEME_STORAGE_KEY = 'yizhan-theme';
const ACCOUNT_CENTER_RETURN_PATH_KEY = 'qiantie:account-center-return-path';
const ACCOUNT_CENTER_ENTRY_KEY = 'qiantie:account-center-entry-active';

function initialTheme() {
  return localStorage.getItem(THEME_STORAGE_KEY) === 'light' ? 'light' : 'dark';
}

function pageTitle(pathname) {
  if (isAccountCenterPath(pathname)) return '个人中心';
  if (pathname === '/batch-factory' || pathname.startsWith('/batch-factory/') || pathname === '/shuihuo-production') return '水货生产';
  const item = navItems.find(nav => nav.href === pathname);
  return item ? item.label : '一战晟铭';
}

function compactHeaderStatus(value) {
  return String(value || '').replace(/\s+/g, ' ').trim().slice(0, 140);
}

function headerStatusTone(value, suggested = 'info') {
  const text = String(value || '');
  if (/处理完成/.test(text) && !/(失败|错误|异常|超时)\s*[：:]?\s*[1-9]\d*/.test(text)) return 'success';
  if (/(失败|错误|异常|超时)/.test(text)) return 'error';
  return suggested;
}

function expiringHeaderStatus(text, tone = 'info', durationMs) {
  const duration = Number(durationMs) || (tone === 'error' ? 8000 : tone === 'working' ? 20000 : 4500);
  return { text: compactHeaderStatus(text), tone, expiresAt: Date.now() + duration };
}

export function UserLayout({ children }) {
  const [username, setUsername] = useState(getCurrentUsername());
  const [account, setAccount] = useState(null);
  const [loading, setLoading] = useState(false);
  const [passkeyLoading, setPasskeyLoading] = useState(false);
  const [loginForm] = Form.useForm();
  const [loginDialogOpen, setLoginDialogOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(() => shouldAutoCollapseAgentNavigation(window.location.pathname));
  const [theme, setTheme] = useState(initialTheme);
  const [petVisible, setPetVisible] = useState(true);
  const [globalStatus, setGlobalStatus] = useState({ text: '', tone: 'idle', expiresAt: 0 });
  const [avatar, setAvatar] = useState(null);
  const accountSessionGenerationRef = useRef(0);
  const loginCardRef = useRef(null);
  const accountCenterReturnPathRef = useRef(null);
  const pathname = window.location.pathname;
  const isAccountCenterRoute = isAccountCenterPath(pathname);
  const isLoggedIn = Boolean(username);
  const isHome = pathname === '/';
  const isAgentRoute = pathname === '/agent' || pathname === '/agent/canvas';
  const isAgentCanvasRoute = pathname === '/agent/canvas';
  // 全站统一：页面标题只在内容区出现一次；顶栏只显示真实全局状态，没有状态时整条收起。
  const routeAccess = getRouteAccessState({ pathname, isLoggedIn });
  const displayAvatar = avatarDisplay(avatar, username);
  // 悬浮助手需要用它区分账号会话；它不能再作为主页面容器的 React key。
  const accountSessionKey = username || 'anonymous';
  const visibleNavItems = navItems;
  const content = isValidElement(children)
    ? cloneElement(
      children,
      { theme },
      isValidElement(children.props.children) ? cloneElement(children.props.children, { theme }) : children.props.children
    )
    : children;

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    localStorage.setItem(THEME_STORAGE_KEY, theme);
  }, [theme]);

  useEffect(() => {
    if (shouldAutoCollapseAgentNavigation(pathname)) setSidebarCollapsed(true);
  }, [pathname]);

  useEffect(() => {
    if (isAccountCenterRoute) return;
    const currentUrl = `${window.location.pathname}${window.location.search}${window.location.hash}`;
    const returnPath = captureAccountCenterOrigin(currentUrl, null, window.location.origin);
    accountCenterReturnPathRef.current = returnPath;
    sessionStorage.setItem(ACCOUNT_CENTER_RETURN_PATH_KEY, returnPath);
    sessionStorage.removeItem(ACCOUNT_CENTER_ENTRY_KEY);
  }, [isAccountCenterRoute, pathname]);

  function enterAccountCenter() {
    const currentUrl = `${window.location.pathname}${window.location.search}${window.location.hash}`;
    const returnPath = captureAccountCenterOrigin(currentUrl, null, window.location.origin);
    accountCenterReturnPathRef.current = returnPath;
    sessionStorage.setItem(ACCOUNT_CENTER_RETURN_PATH_KEY, returnPath);
    sessionStorage.setItem(ACCOUNT_CENTER_ENTRY_KEY, '1');
  }

  function returnToWorkbench() {
    const hasEntryOrigin = sessionStorage.getItem(ACCOUNT_CENTER_ENTRY_KEY) === '1';
    const returnPath = hasEntryOrigin
      ? normalizeAccountCenterReturnPath(accountCenterReturnPathRef.current || sessionStorage.getItem(ACCOUNT_CENTER_RETURN_PATH_KEY), window.location.origin)
      : '/';
    window.location.assign(returnPath);
  }

  useEffect(() => {
    if (!isLoggedIn) return undefined;
    const appendTaskNotification = event => {
      const notification = normalizeGlobalTaskNotification(event.detail || {});
      const notificationText = `${notification.title}${notification.detail ? `：${notification.detail}` : ''}`;
      setGlobalStatus(expiringHeaderStatus(notificationText, headerStatusTone(notificationText, notification.status === 'error' ? 'error' : notification.status === 'working' ? 'working' : 'success')));
      if (notification.status === 'working') return;
      const kind = notification.status === 'error' ? 'error' : 'success';
      const targetPath = notification.pagePath && notification.pagePath !== window.location.pathname ? notification.pagePath : '';
      message.open({
        key: `task-${notification.id}`,
        type: kind,
        content: <div className="global-task-toast" role={targetPath ? 'link' : undefined} tabIndex={targetPath ? 0 : undefined} onClick={() => targetPath && window.location.assign(targetPath)} onKeyDown={event => { if (targetPath && (event.key === 'Enter' || event.key === ' ')) window.location.assign(targetPath); }}><span><strong>{notification.title}</strong>{notification.detail ? `：${notification.detail}` : ''}</span>{targetPath ? <Button type="link" size="small" onClick={event => { event.stopPropagation(); window.location.assign(targetPath); }}>返回查看</Button> : null}</div>,
        duration: notification.status === 'error' ? 6 : 4,
        style: { marginTop: 12 },
        onClick: () => targetPath && window.location.assign(targetPath)
      });
    };
    const receiveEmbeddedTaskNotification = event => {
      if (event.origin !== window.location.origin) return;
      if (event.data?.type !== 'qiantie:task-notification') return;
      appendTaskNotification({ detail: event.data });
    };
    window.addEventListener(GLOBAL_TASK_NOTIFICATION_EVENT, appendTaskNotification);
    window.addEventListener('message', receiveEmbeddedTaskNotification);
    return () => {
      window.removeEventListener(GLOBAL_TASK_NOTIFICATION_EVENT, appendTaskNotification);
      window.removeEventListener('message', receiveEmbeddedTaskNotification);
    };
  }, [isLoggedIn, username]);

  useEffect(() => {
    const receive = event => {
      const detail = event.detail || {};
      if (!detail.text) {
        // 业务页可以发空文本主动清空状态，避免顶栏保留已经过期的队列文案。
        setGlobalStatus({ text: '', tone: 'idle', expiresAt: 0 });
        return;
      }
      setGlobalStatus(expiringHeaderStatus(detail.text, detail.tone || 'info', detail.durationMs));
    };
    window.addEventListener(GLOBAL_STATUS_EVENT, receive);
    return () => window.removeEventListener(GLOBAL_STATUS_EVENT, receive);
  }, []);

  useEffect(() => {
    const expiresAt = Number(globalStatus.expiresAt || 0);
    if (!globalStatus.text || !expiresAt) return undefined;
    const timer = window.setTimeout(() => {
      setGlobalStatus(current => current.expiresAt === expiresAt ? { text: '', tone: 'idle', expiresAt: 0 } : current);
    }, Math.max(0, expiresAt - Date.now()));
    return () => window.clearTimeout(timer);
  }, [globalStatus]);

  useEffect(() => {
    document.body.classList.add('user-theme-active');
    return () => document.body.classList.remove('user-theme-active');
  }, []);

  useEffect(() => {
    if (!isLoggedIn) return undefined;
    const refreshProfile = () => {
      getCurrentAccount().then(currentAccount => {
        setUsername(currentAccount.username);
        setAccount(currentAccount);
      }).catch(() => {});
    };
    window.addEventListener('qiantie:profile-updated', refreshProfile);
    return () => window.removeEventListener('qiantie:profile-updated', refreshProfile);
  }, [isLoggedIn]);

  useEffect(() => {
    function updatePetVisibility(event) {
      setPetVisible(event.detail?.petVisible !== false);
    }
    window.addEventListener('qiantie:notifications-updated', updatePetVisibility);
    if (isLoggedIn) {
      getConfig().then(config => {
        setPetVisible(config.notifications?.petVisible !== false);
        setAvatar(config.avatar || null);
      }).catch(() => {});
    }
    return () => window.removeEventListener('qiantie:notifications-updated', updatePetVisibility);
  }, [isLoggedIn]);

  useEffect(() => {
    let dialogOpen = false;
    function showApiFailure(event) {
      if (dialogOpen) return;
      dialogOpen = true;
      const { source, method, status, message: detail, sessionAuthFailure } = event.detail || {};
      if (shouldPromptLoginForApiFailure({ status, sessionAuthFailure })) {
        setGlobalStatus(expiringHeaderStatus('登录已失效，请重新登录后继续使用', 'error'));
        setUsername('');
        setAccount(null);
        setLoginDialogOpen(true);
        Modal.warning({
          className: 'auth-expired-modal',
          title: '登录已失效',
          content: '当前登录状态已失效，请重新登录后继续使用。登录后将返回当前页面。',
          okText: '重新登录',
          onOk: () => { setLoginDialogOpen(true); dialogOpen = false; },
          afterClose: () => { dialogOpen = false; }
        });
        return;
      }
      const sourceLabel = source ? `${method || 'GET'} ${source}` : '服务请求';
      const statusLabel = status ? `（${status}）` : '';
      setGlobalStatus(expiringHeaderStatus(`${sourceLabel}${statusLabel}：${detail || '请求失败，请稍后重试。'}`, 'error'));
      dialogOpen = false;
    }
    window.addEventListener('qiantie:api-error', showApiFailure);
    return () => window.removeEventListener('qiantie:api-error', showApiFailure);
  }, []);

  useEffect(() => {
    dispatchPetContext({
      page: pageTitle(pathname),
      pagePath: pathname,
      workspace: 'qiantie'
    });
  }, [pathname]);

  useEffect(() => {
    let cancelled = false;
    const sessionGeneration = ++accountSessionGenerationRef.current;
    const sessionToken = getToken();

    function isCurrentSession() {
      return !cancelled
        && accountSessionGenerationRef.current === sessionGeneration
        && getToken() === sessionToken;
    }

    function clearSession() {
      setUsername('');
      setAccount(null);
    }

    function showLogin(event) {
      const expiredToken = event?.detail?.token;
      const currentToken = getToken();
      // Modern clients dispatch the token used by the failed request. Accept
      // only an event for the currently active session. Legacy clients do not
      // carry a token and can only be accepted after they cleared this effect's
      // still-current session.
      if (expiredToken) {
        if (!currentToken || expiredToken !== currentToken) return;
      } else if (!sessionToken || currentToken || cancelled || accountSessionGenerationRef.current !== sessionGeneration) {
        return;
      }
      accountSessionGenerationRef.current += 1;
      clearSession();
      setLoginDialogOpen(true);
    }

    window.addEventListener('qiantie:auth-expired', showLogin);
    if (!sessionToken) {
      clearSession();
    } else {
      getCurrentAccount().then(currentAccount => {
        if (!isCurrentSession()) return;
        setUsername(currentAccount.username);
        setAccount(currentAccount);
      }).catch(error => {
        if (!isCurrentSession()) return;
        if (String(error.message).includes('登录已失效')) showLogin();
        else setAccount(null);
      });
    }
    return () => {
      cancelled = true;
      accountSessionGenerationRef.current += 1;
      window.removeEventListener('qiantie:auth-expired', showLogin);
    };
  }, []);

  useEffect(() => {
    if (routeAccess.shouldPromptLogin) setLoginDialogOpen(true);
  }, [routeAccess.shouldPromptLogin]);

  async function handleLogin(values) {
    accountSessionGenerationRef.current += 1;
    setLoading(true);
    try {
      const data = await login(values.username, values.password, values.remember);
      setUsername(data.username);
      setAccount(data);
      setLoginDialogOpen(false);
      message.success('登录成功');
    } catch (error) {
      message.error(error.message || '登录失败');
    } finally {
      setLoading(false);
    }
  }

  async function handlePasskeyLogin() {
    const values = loginForm.getFieldsValue(['username', 'remember']);
    if (!values.username) return message.warning('请先填写账号，再使用 Passkey 登录');
    accountSessionGenerationRef.current += 1;
    setPasskeyLoading(true);
    try {
      const data = await loginWithPasskey(values.username, values.remember !== false);
      setUsername(data.username);
      setAccount(data);
      setLoginDialogOpen(false);
      message.success('Passkey 登录成功');
    } catch (error) {
      message.error(error.message || 'Passkey 登录失败');
    } finally {
      setPasskeyLoading(false);
    }
  }

  async function handleLogout() {
    accountSessionGenerationRef.current += 1;
    await logout();
    setUsername('');
    setAccount(null);
    setAvatar(null);
    setLoginDialogOpen(true);
    message.success('已退出');
  }

  function openLoginDialog() {
    setLoginDialogOpen(true);
  }

  function updateLoginCardParallax(event) {
    const card = loginCardRef.current;
    if (!card) return;
    const rect = card.getBoundingClientRect();
    const x = (event.clientX - rect.left) / rect.width - 0.5;
    const y = (event.clientY - rect.top) / rect.height - 0.5;
    card.style.setProperty('--login-card-rotate-x', `${-y * 5}deg`);
    card.style.setProperty('--login-card-rotate-y', `${x * 5}deg`);
  }

  function resetLoginCardParallax() {
    const card = loginCardRef.current;
    if (!card) return;
    card.style.removeProperty('--login-card-rotate-x');
    card.style.removeProperty('--login-card-rotate-y');
  }

  function toggleSidebar() {
    setSidebarCollapsed(collapsed => !collapsed);
  }

  function toggleTheme() {
    setTheme(current => current === 'dark' ? 'light' : 'dark');
  }

  const showLoginCard = !isLoggedIn && (isHome || loginDialogOpen);
  const pageContent = routeAccess.canRenderPage ? content : null;
  const loginOverlay = showLoginCard ? (
    <div className="legacy-login-overlay nebula-login-overlay">
      <div className="login-modal nebula-login-modal" ref={loginCardRef} onPointerMove={updateLoginCardParallax} onPointerLeave={resetLoginCardParallax}>
        <div className="nebula-login-brand"><BrandLogo /></div>
        <h1>一战晟铭登录</h1>
        <p className="nebula-login-subtitle">继续你的创作工作流</p>
        <div className="login-form-wrap nebula-login-form-wrap">
          <Form form={loginForm} layout="vertical" initialValues={{ remember: true }} onFinish={handleLogin}>
            <Form.Item label="账号" name="username" rules={[{ required: true, message: '请输入账号' }]}>
              <Input placeholder="请输入账号" autoComplete="username" />
            </Form.Item>
            <Form.Item label="密码" name="password" rules={[{ required: true, message: '请输入密码' }]}>
              <Input.Password
                placeholder="请输入密码"
                autoComplete="current-password"
                styles={{ input: { backgroundColor: '#727786', color: '#fff', boxShadow: 'none' } }}
              />
            </Form.Item>
            <Form.Item name="remember" valuePropName="checked">
              <Checkbox>30 天保持登录</Checkbox>
            </Form.Item>
            <Button block type="primary" htmlType="submit" loading={loading}>登录并进入工作台</Button>
            <Button block icon={<Fingerprint size={17} />} loading={passkeyLoading} onClick={handlePasskeyLogin}>使用 Passkey 登录</Button>
          </Form>
          <p className="login-hint"><a href="/recover">忘记密码？</a>　提示：请联系管理员获取账号</p>
        </div>
      </div>
    </div>
  ) : null;

  if (isHome) {
    return (
      <ConfigProvider theme={createAntTheme(theme)}>
        <div className="home-shell">
          {isValidElement(children) ? cloneElement(children, { theme, isLoggedIn, onOpenLogin: openLoginDialog }) : content}
          {loginOverlay}
        </div>
      </ConfigProvider>
    );
  }

  if (isAccountCenterRoute) {
    return (
      <ConfigProvider theme={createAntTheme(theme)}>
        <AccountCenterLayout account={account} username={username} avatar={displayAvatar} pathname={pathname} onReturn={returnToWorkbench} onLogout={handleLogout}>{pageContent}</AccountCenterLayout>
        {loginOverlay}
      </ConfigProvider>
    );
  }

  return (
    <ConfigProvider theme={createAntTheme(theme)}>
      <div className="legacy-shell">
      <aside className={`legacy-sidebar${sidebarCollapsed ? ' collapsed' : ''}`}>
        <div className="legacy-brand">
          <button
            className="legacy-sidebar-toggle"
            type="button"
            aria-label={sidebarCollapsed ? '展开导航' : '收起导航'}
            aria-expanded={!sidebarCollapsed}
            onClick={toggleSidebar}
          >
            {sidebarCollapsed ? <PanelLeftOpen size={18} aria-hidden="true" /> : <PanelLeftClose size={18} aria-hidden="true" />}
          </button>
          <Link href="/" className="legacy-brand-link">
            <Suspense fallback={<BrandLogo className="legacy-brand-logo" />}>
              <SuperOpcLiquidMetalLogo theme={theme} scale={1} className="legacy-brand-logo" />
            </Suspense>
            <span className="legacy-brand-title shine">一战晟铭</span>
          </Link>
        </div>
        <nav className="legacy-nav">
          {visibleNavItems.map(item => {
            const Icon = item.icon;
            return (
              <Link key={item.href} href={item.href} reload={item.reload} className={pathname === item.href ? 'active' : ''} onIntent={() => { void preloadRoute(item.href); }}>
                <span className="legacy-nav-icon"><Icon size={18} strokeWidth={1.8} aria-hidden="true" /></span>
                <span className="legacy-nav-label">{item.label}</span>
                <span className="legacy-nav-tooltip" aria-hidden="true">{item.label}</span>
              </Link>
            );
          })}
        </nav>
        <div className="legacy-sidebar-tools">
          <label className="legacy-theme-switch" title={theme === 'dark' ? '切换至浅色主题' : '切换至深色主题'}>
            <input
              type="checkbox"
              checked={theme === 'dark'}
              aria-label={theme === 'dark' ? '切换至浅色主题' : '切换至深色主题'}
              onChange={toggleTheme}
            />
            <span className="legacy-theme-switch__track" aria-hidden="true">
              <span className="legacy-theme-switch__clouds" />
              <span className="legacy-theme-switch__stars" />
              <span className="legacy-theme-switch__thumb"><Sun className="legacy-theme-switch__sun" size={17} strokeWidth={1.8} /><Moon className="legacy-theme-switch__moon" size={16} strokeWidth={1.8} /></span>
            </span>
          </label>
          <Link href="/settings" className="legacy-sidebar-tool" title="设置">
            <span className="legacy-nav-icon"><Settings2 size={18} strokeWidth={1.8} aria-hidden="true" /></span>
            <span className="legacy-nav-label">设置</span>
          </Link>
        </div>
        {isLoggedIn ? (
          <div className="legacy-sidebar-user">
            <button
              type="button"
              className="legacy-sidebar-avatar"
              title="打开个人中心"
              aria-label="进入个人中心"
              onClick={() => {
                if (!isAccountCenterRoute) {
                  enterAccountCenter();
                }
                window.history.pushState({}, '', '/profile');
                window.dispatchEvent(new PopStateEvent('popstate'));
              }}
            >
              <Avatar size={28} src={account?.avatarUrl} style={{ backgroundColor: displayAvatar.background }}>{displayAvatar.emoji}</Avatar>
            </button>
            <span className="legacy-sidebar-username" title={account?.displayName || username}>{account?.displayName || username}</span>
            <Button className="legacy-sidebar-logout" size="small" type="text" aria-label="退出登录" onClick={handleLogout}>
              <LogOut size={16} strokeWidth={1.9} aria-hidden="true" />
              <span className="legacy-sidebar-logout-label">退出</span>
            </Button>
          </div>
        ) : null}
      </aside>
      <Fragment>
        <main className="legacy-main">
          {!isAgentCanvasRoute && globalStatus.text ? <header className="legacy-topbar">
            <div className="legacy-userbar" aria-hidden="true" />
            <div className={`legacy-global-status legacy-global-status--${globalStatus.tone}`} role="status" aria-live="polite">{globalStatus.text}</div>
          </header> : null}
          <section className={`legacy-content${isAgentRoute ? ' legacy-content--agent' : ''}${isAgentCanvasRoute ? ' legacy-content--agent-canvas' : ''}`}>{pageContent}</section>
        </main>
        {isLoggedIn && pathname !== '/' && petVisible ? <CmPenguinCompanion username={username} accountSessionKey={accountSessionKey} /> : null}
      </Fragment>
      {loginOverlay}
      </div>
    </ConfigProvider>
  );
}
