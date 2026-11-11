const preloaders = {
  '/novel-fetch': () => import('./pages/NovelFetchPage'),
  '/batch-factory': () => import('./pages/BatchFactoryPageV30'),
  '/plugins': () => import('./pages/PluginCenterPage'),
  '/plugins/ai-rewrite': () => import('./pages/PluginRewritePage'),
  '/ai-cut': () => import('./pages/AICutPage')
};

export function preloadRoute(href) {
  return preloaders[href]?.();
}
