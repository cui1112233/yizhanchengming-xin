import { lazy, Suspense, useEffect, useState } from 'react';
import { UserLayout } from '../shared/layouts/UserLayout';
import { HomeRoute } from './pages/HomeRoute';
import GiantMaterialTestPage from './pages/GiantMaterialTestPage';

const loadNovelPanelPage = () => import('./pages/NovelPanelPage');
const loadNovelFetchPage = () => import('./pages/NovelFetchPage');
const loadBatchFactoryWorkbenchPage = () => import('./pages/BatchFactoryWorkbenchPage');

const ScriptPage = lazy(() => import('./pages/ScriptPage'));
const NovelPanelPage = lazy(loadNovelPanelPage);
const AgentPage = lazy(() => import('./pages/AgentPageV2').then(module => ({ default: module.AgentPageV2 })));
const ShuihuoProductionPage = lazy(() => import('./pages/ShuihuoProductionPage'));
const TtsPage = lazy(() => import('./pages/TtsPage'));
const NovelFetchPage = lazy(loadNovelFetchPage);
const NovelFetchWorkshopPage = lazy(() => import('./pages/NovelFetchWorkshopPage'));
const BatchFactoryPage = lazy(() => import('./pages/BatchFactoryPage'));
const BatchFactoryWorkbenchPage = lazy(loadBatchFactoryWorkbenchPage);
const SettingsPage = lazy(() => import('./pages/SettingsPage'));
const MemberCenterPage = lazy(() => import('./pages/MemberCenterPage'));
const ProfilePage = lazy(() => import('./pages/ProfilePage'));
const SecurityPage = lazy(() => import('./pages/SecurityPage'));
const AdvancedTeamAdminPage = lazy(() => import('./pages/AdvancedTeamAdminPage'));
const AccountRolePage = lazy(() => import('./pages/AccountRolePage'));
const TeamPage = lazy(() => import('./pages/TeamPage'));
const ApiConfigPage = lazy(() => import('./pages/ApiConfigPage'));
const UsageStatsPage = lazy(() => import('./pages/UsageStatsPage'));
const InviteAcceptPage = lazy(() => import('./pages/InviteAcceptPage'));
const RecoveryPage = lazy(() => import('./pages/RecoveryPage'));

export async function preloadWorkbenchPages(loaders = {
  novelFetch: loadNovelFetchPage,
  novelPanel: loadNovelPanelPage,
  batchFactory: loadBatchFactoryWorkbenchPage
}) {
  await Promise.allSettled(Object.values(loaders).map(load => load()));
}

function scheduleWorkbenchPreload() {
  const preload = () => { void preloadWorkbenchPages(); };
  if (typeof window.requestIdleCallback === 'function') {
    const idleId = window.requestIdleCallback(preload, { timeout: 1500 });
    return () => window.cancelIdleCallback?.(idleId);
  }
  const timerId = window.setTimeout(preload, 200);
  return () => window.clearTimeout(timerId);
}

function usePathname() {
  const canonicalPathname = () => {
    const pathname = window.location.pathname;
    return pathname.length > 1 ? pathname.replace(/\/+$/, '') : pathname;
  };
  const [pathname, setPathname] = useState(canonicalPathname);

  useEffect(() => {
    const handlePopState = () => setPathname(canonicalPathname());
    window.addEventListener('popstate', handlePopState);
    return () => window.removeEventListener('popstate', handlePopState);
  }, []);

  return pathname;
}

function getPage(pathname) {
  const routes = {
    '/script': ScriptPage,
    '/novel-fetch': NovelFetchPage,
    '/novel-fetch-workshop': NovelFetchWorkshopPage,
    '/batch-factory': BatchFactoryWorkbenchPage,
    '/batch-factory-preview': BatchFactoryPage,
    '/novel-panel': NovelPanelPage,
    '/agent': AgentPage,
    '/agent/canvas': AgentPage,
    '/shuihuo-production': ShuihuoProductionPage,
    '/shuihuo-production/creative': ShuihuoProductionPage,
    '/tts': TtsPage,
    '/settings': SettingsPage,
    '/api-config': ApiConfigPage,
    '/usage': UsageStatsPage,
    '/team': TeamPage,
    '/member': MemberCenterPage,
    '/profile': ProfilePage,
    '/security': SecurityPage,
    '/advanced-team-admin': AdvancedTeamAdminPage,
    '/accounts': AccountRolePage,
  };
  const Page = routes[pathname];
  if (Page) {
    return <Suspense fallback={<div className="route-loading" role="status">正在加载工作台</div>}><Page /></Suspense>;
  }
  return <HomeRoute />;
}

export function UserApp() {
  useEffect(() => scheduleWorkbenchPreload(), []);

  useEffect(() => {
    const recoverChunk = event => {
      const message = String(event?.reason?.message || event?.error?.message || event?.message || '');
      if (!/Failed to fetch dynamically imported module|Unable to preload CSS/i.test(message)) return;
      const key = 'qiantie:chunk-recovery';
      if (sessionStorage.getItem(key) === window.location.href) return;
      sessionStorage.setItem(key, window.location.href);
      window.location.reload();
    };
    window.addEventListener('error', recoverChunk);
    window.addEventListener('unhandledrejection', recoverChunk);
    return () => { window.removeEventListener('error', recoverChunk); window.removeEventListener('unhandledrejection', recoverChunk); };
  }, []);
  const pathname = usePathname();
  if (pathname === '/giant-material-test') return <GiantMaterialTestPage />;
  if (/^\/invite\/[^/]+$/.test(pathname)) {
    return <Suspense fallback={<div className="route-loading" role="status">正在加载团队邀请</div>}><InviteAcceptPage /></Suspense>;
  }
  if (pathname === '/recover') {
    return <Suspense fallback={<div className="route-loading" role="status">正在加载账号恢复</div>}><RecoveryPage /></Suspense>;
  }
  return <UserLayout>{getPage(pathname)}</UserLayout>;
}
