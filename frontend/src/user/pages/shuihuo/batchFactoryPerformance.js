const RUNTIME_BACKOFF_MS = [5000, 15000, 30000];
const ACTIVE_RUNTIME_POLL_MS = 8000;
const IDLE_RUNTIME_POLL_MS = 60000;

export function nextViewportLoaded(current, entries = []) {
  return Boolean(current || entries.some(entry => entry?.isIntersecting));
}

export function batchFactoryRuntimePollDelay(state, failures = 0) {
  const failureCount = Math.max(0, Number(failures) || 0);
  if (failureCount > 0) {
    return RUNTIME_BACKOFF_MS[Math.min(failureCount - 1, RUNTIME_BACKOFF_MS.length - 1)];
  }
  return state === 'running' || state === 'scheduled'
    ? ACTIVE_RUNTIME_POLL_MS
    : IDLE_RUNTIME_POLL_MS;
}

export function comicRowResultPollDelay(pages = []) {
  const hasUnfinishedGeneration = (Array.isArray(pages) ? pages : []).some(page => (
    page?.generationState === 'generating'
      && !(Array.isArray(page?.images) ? page.images : []).some(image => image?.isPrimary)
  ));
  return hasUnfinishedGeneration ? 1000 : 0;
}

// 图片已经落库时先把当前页展示出来。整批详情有数 MB，刷新它只能作为
// 后台同步，不能继续卡住用户看到刚生成图片的时间。
export async function finishComicPageGeneration({ waitForResult, revealResult, refreshBatch }) {
  const page = await waitForResult();
  revealResult?.(page);
  Promise.resolve().then(() => refreshBatch?.()).catch(() => {});
  return page;
}
