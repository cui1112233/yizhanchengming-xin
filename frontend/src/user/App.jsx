import { lazy, Suspense, useEffect, useState } from 'react';
import { UserLayout } from '../shared/layouts/UserLayout';
import { StandaloneThemeWindow } from '../shared/layouts/StandaloneThemeWindow';
import { HomeRoute } from './pages/HomeRoute';
import GiantMaterialTestPage from './pages/GiantMaterialTestPage';
import { PLUGIN_CATALOG } from './pages/pluginCatalog.js';
import { preloadRoute as preloadWorkbenchRoute } from './workbenchPreload.js';

export function preloadRoute(href) {
  return preloadWorkbenchRoute(href);
}

const loadNovelFetchPage = () => import('./pages/NovelFetchPage');

const ScriptPage = lazy(() => import('./pages/ScriptPage'));
const AgentPage = lazy(() => import('./pages/AgentPageV2').then(module => ({ default: module.AgentPageV2 })));
const AgentChatPage = lazy(() => import('./pages/AgentChatPage'));
const ShuihuoProductionPage = lazy(() => import('./pages/BatchFactoryPageV30'));
const TtsPage = lazy(() => import('./pages/TtsPage'));
const NovelFetchPage = lazy(loadNovelFetchPage);
const SettingsPage = lazy(() => import('./pages/SettingsPage'));
const PluginCenterPage = lazy(() => import('./pages/PluginCenterPage'));
const AICutPage = lazy(() => import('./pages/AICutPage'));
const AICutLibraryPage = lazy(() => import('./pages/AICutLibraryPage'));
const ComicScreenplayPage = lazy(() => import('./pages/ComicScreenplayPage'));
const NovelCreationPage = lazy(() => import('./pages/NovelCreationPage'));
const PluginRewritePage = lazy(() => import('./pages/PluginRewritePage'));
const PluginToolPage = lazy(() => import('./pages/PluginToolPage'));
const MemberCenterPage = lazy(() => import('./pages/MemberCenterPage'));
const ProfilePage = lazy(() => import('./pages/ProfilePage'));
const SecurityPage = lazy(() => import('./pages/SecurityPage'));
const AdvancedTeamAdminPage = lazy(() => import('./pages/AdvancedTeamAdminPage'));
const AccountRolePage = lazy(() => import('./pages/AccountRolePage'));
const TeamPage = lazy(() => import('./pages/TeamPage'));
const ApiConfigPage = lazy(() => import('./pages/ApiConfigPage'));
const UsageStatsPage = lazy(() => import('./pages/UsageStatsPage'));
const HistoryPage = lazy(() => import('./pages/HistoryPage'));
const IssueLogPage = lazy(() => import('./pages/IssueLogPage'));
const InviteAcceptPage = lazy(() => import('./pages/InviteAcceptPage'));
const RecoveryPage = lazy(() => import('./pages/RecoveryPage'));

function CanonicalBatchFactoryRedirect() {
  useEffect(() => {
    const suffix = `${window.location.search}${window.location.hash}`;
    window.location.replace(`/batch-factory${suffix}`);
  }, []);
  return <div className="route-loading" role="status">正在打开批量工厂</div>;
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

// 提示词工具的整页面板路由从插件目录 launch.href 派生，不在这里逐条登记。
const pluginToolRoutes = Object.fromEntries(
  PLUGIN_CATALOG
    .filter(plugin => plugin.launch.kind === 'prompt-tool' && plugin.launch.href)
    .map(plugin => [plugin.launch.href, PluginToolPage])
);

function getPage(pathname) {
  const routes = {
    '/script': ScriptPage,
    // 批量测书正式入口，直接渲染 React 源码工作台。
    '/novel-fetch': NovelFetchPage,
    // 迁移调试地址保留同一份 React 页面，避免出现第二套实现。
    '/novel-fetch-react': NovelFetchPage,
    '/novel-fetch-workshop': NovelFetchPage,
    '/novel-creation': NovelCreationPage,
    // The cover-based project library is the only Batch Factory entry.
    '/batch-factory': ShuihuoProductionPage,
    '/agent': AgentPage,
    '/agent/canvas': AgentPage,
    // 工作区 → 会话 → 消息 的对话式 Agent（四栏布局的其余三栏）。
    '/agent/chat': AgentChatPage,
    // Keep historical addresses as redirects so they cannot open a second
    // Batch Factory implementation.
    '/shuihuo-production': CanonicalBatchFactoryRedirect,
    '/shuihuo-production/creative': CanonicalBatchFactoryRedirect,
    '/tts': TtsPage,
    '/settings': SettingsPage,
    '/plugins': PluginCenterPage,
    '/ai-cut': AICutPage,
    '/comic-screenplay': ComicScreenplayPage,
    '/plugins/ai-rewrite': PluginRewritePage,
    '/api-config': ApiConfigPage,
    '/usage': UsageStatsPage,
    '/history': HistoryPage,
    '/issues': IssueLogPage,
    '/team': TeamPage,
    '/member': MemberCenterPage,
    '/profile': ProfilePage,
    '/security': SecurityPage,
    '/advanced-team-admin': AdvancedTeamAdminPage,
    '/accounts': AccountRolePage,
  };
  const Page = routes[pathname] || pluginToolRoutes[pathname];
  if (Page) {
    // 同一组件承担多个插件面板地址时，按路径重建实例，避免沿用上一条的输入和结果。
    return <Suspense fallback={<div className="route-loading" role="status">正在加载工作台</div>}><Page key={pathname} /></Suspense>;
  }
  return <HomeRoute />;
}

export function UserApp() {
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
  if (pathname === '/ai-cut') {
    return <StandaloneThemeWindow><Suspense fallback={<div className="route-loading" role="status">正在加载 AI-CUT</div>}>{new URLSearchParams(window.location.search).get('project') ? <AICutPage /> : <AICutLibraryPage />}</Suspense></StandaloneThemeWindow>;
  }
  return <UserLayout>{getPage(pathname)}</UserLayout>;
}
