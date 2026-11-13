import { bookFailureIsHistorical } from './batchFactoryBookState.js';

export function batchFactoryBookStageBadge(book, { productionStatus, stageSummary = {}, automationState = 'idle', automationBook = null } = {}) {
  if (!book) return null;
  const summary = stageSummary || {};
  const runs = Array.isArray(summary.runs) ? summary.runs : [];
  const state = String(automationBook?.status || '').toLowerCase();
  const stage = String(automationBook?.stage || '');
  const stageActive = name => runs.some(run => run?.stage === name && ['running', 'queued'].includes(String(run?.status || '').toLowerCase()));
  if (summary.lastFailed && !bookFailureIsHistorical({ automationBook, lastFailed: summary.lastFailed, runs })) return { label: '失败', tone: 'failed', detail: '本书有失败步骤，可在生产表重试失败步骤' };
  if (Number(summary.comicPublishedCollectionCount || 0) > 0) return { label: '已发布', tone: 'done', detail: '漫画合集已发布' };
  if (automationBook && state !== 'completed' && automationState === 'paused') return { label: '已暂停', tone: 'queued', detail: '已暂停后续派发，可继续任务；已提交给模型的任务仍可能返回' };
  if (automationBook && state !== 'completed' && ['stopped', 'cancelled'].includes(automationState)) return { label: '本轮已结束', tone: 'queued', detail: '已有成果保留，需要时重新启动或重试未完成任务' };
  if (['failed', 'blocked'].includes(state)) return { label: '失败', tone: 'failed', detail: automationBook.error || '本书执行失败，可重试失败任务' };
  const ready = Number(summary.comicImageReadyCount || 0);
  const total = Number(summary.comicImageTotalCount || 0);
  if (Number(summary.comicImageFailedCount || 0) > 0) return { label: `漫画生成失败 ${ready}/${total}`, tone: 'failed', detail: '有漫画页生成失败，可在生产表重试失败步骤' };
  if (Number(summary.comicImageInterruptedCount || 0) > 0) return { label: `生成已中断 ${ready}/${total}`, tone: 'failed', detail: '漫画页任务长时间未更新，可在生产表重试失败步骤恢复' };
  const submitting = (Array.isArray(productionStatus?.jobs) ? productionStatus.jobs : []).some(job => String(job?.bookId || '') === String(book.id) && ['running', 'queued'].includes(String(job?.status || '').toLowerCase()));
  if (submitting) return { label: '提交中', tone: 'running', detail: '生产任务已提交，等待上游返回' };
  if (stageActive('assets') || (automationState === 'running' && state === 'running' && ['assets', 'comic_assets'].includes(stage))) return { label: '资产生成中', tone: 'running', detail: '正在生成人物、场景或道具资产' };
  if (stageActive('director') || stageActive('comic') || (automationState === 'running' && state === 'running' && ['director', 'comic_bootstrap', 'comic_prompt_recovery'].includes(stage))) return { label: '提示词生产中', tone: 'running', detail: '正在生成分镜与提示词' };
  if (Number(summary.comicImageGeneratingCount || 0) > 0 || stageActive('image') || (automationState === 'running' && state === 'running' && stage === 'comic_images')) return { label: total > 0 ? `漫画生成中 ${ready}/${total}` : '图片生成中', tone: 'running', detail: '有已提交的图片任务，正在等待生成结果' };
  if (summary.comicCollectionReady === true || summary.comicImageReady === true) return { label: '已完成', tone: 'done', detail: '本书当前阶段已产出结果' };
  if (automationBook && ['running', 'scheduled'].includes(automationState) && !['completed', 'failed', 'blocked', 'stopped'].includes(state)) return { label: '排队中', tone: 'queued', detail: '已进入本轮任务队列，等待执行后续步骤' };
  if (total > 0 && ready < total) return { label: `${ready > 0 ? '待继续生成' : '待生成'} ${ready}/${total}`, tone: 'queued', detail: '没有正在出图的任务。可在生产表点击“生成整本漫画”，或在“自动生产”菜单中启动任务' };
  return null;
}

export function manualRetryBooks(books, stageSummaries, automationStatus) {
  return books.filter(book => {
    const summary = stageSummaries?.[book.id] || {};
    const automationBook = automationStatus?.books?.find(item => String(item.bookId) === String(book.id));
    return Number(summary.comicImageFailedCount || 0) > 0 || Number(summary.comicImageInterruptedCount || 0) > 0 || (summary.lastFailed && !bookFailureIsHistorical({ automationBook, lastFailed: summary.lastFailed, runs: summary.runs || [] }));
  });
}
