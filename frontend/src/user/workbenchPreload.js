const preloaders = {
  '/novel-fetch': () => import('./pages/NovelFetchPage'),
  '/batch-factory': () => import('./pages/BatchFactoryPageV30')
};

export function preloadRoute(href) {
  return preloaders[href]?.();
}
