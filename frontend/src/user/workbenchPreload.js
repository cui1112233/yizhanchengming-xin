const preloaders = {
  '/novel-fetch': () => import('./pages/NovelFetchPage'),
  '/batch-factory': () => import('./pages/ShuihuoProductionPage')
};

export function preloadRoute(href) {
  return preloaders[href]?.();
}
