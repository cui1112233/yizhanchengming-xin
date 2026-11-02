export const YIZHANSHENGMING_EXECUTOR_PROVIDER = 'doubao_local_executor';

export function usesYizhanshengmingExecutor(settings = {}) {
  return settings.imageProvider === YIZHANSHENGMING_EXECUTOR_PROVIDER
    || settings.videoProvider === YIZHANSHENGMING_EXECUTOR_PROVIDER;
}

export function batchExecutorKinds(batch = {}) {
  const batchSettings = batch?.settingsState?.patch || {};
  const books = Array.isArray(batch?.books) && batch.books.length ? batch.books : [{}];
  const settings = books.map(book => ({ ...batchSettings, ...(book?.settingsState?.patch || {}) }));
  return ['image', 'video'].filter(kind => settings.some(item => item[`${kind}Provider`] === YIZHANSHENGMING_EXECUTOR_PROVIDER));
}

export function localExecutorTaskToast(task = {}) {
  switch (String(task.status || '').toLowerCase()) {
    case 'queued':
      return { type: 'loading', text: '一战晟铭执行器：任务已排队' };
    case 'running':
      return { type: 'loading', text: '一战晟铭执行器：正在生成' };
    case 'succeeded':
      return { type: 'success', text: '一战晟铭执行器：生成完成' };
    case 'failed':
      return { type: 'error', text: `一战晟铭执行器：生成失败${task.errorMessage ? `：${task.errorMessage}` : ''}` };
    case 'cancelled':
      return { type: 'warning', text: '一战晟铭执行器：任务已取消' };
    default:
      return null;
  }
}
