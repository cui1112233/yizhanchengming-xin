export function batchFactoryPriorityAction(automation, bookId) {
  const unavailable = reason => ({enabled:false,label:'插队',reason});
  if (!bookId) return unavailable('请先选中一本书');
  if (!['running','paused'].includes(automation?.state)) return unavailable('请先启动当前批次的自动生产');
  const book = (automation.books || []).find(item => String(item.bookId) === String(bookId));
  if (book && !['queued','paused'].includes(book.status)) return unavailable(book.status === 'running' ? '这本书正在处理，无需插队' : '这本书已结束，请使用执行或重试');
  if (book?.priorityRequested) return {enabled:false,label:'已插队',reason:'已插队，等待当前书完成后优先处理'};
  return {enabled:true,label:'插队',reason:'不打断正在处理的书，优先取得下一个生产空位'};
}
