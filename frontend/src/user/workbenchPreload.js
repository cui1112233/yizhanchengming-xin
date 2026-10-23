const preloaders = {
  '/novel-fetch': () => import('./pages/NovelFetchPage'),
  '/batch-factory': () => import('./pages/BatchFactoryWorkbenchPage')
};

export function preloadRoute(href) {
  return preloaders[href]?.();
}
