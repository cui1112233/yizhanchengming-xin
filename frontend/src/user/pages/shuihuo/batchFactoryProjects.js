const BATCH_FACTORY_V11_SOURCE = 'batch_factory_v11';

function text(value) {
  return String(value ?? '').trim();
}

export function createCoverHydrationScheduler() {
  let tail = Promise.resolve();
  return task => {
    const scheduled = tail.catch(() => {}).then(task);
    tail = scheduled;
    return scheduled;
  };
}

export function batchFactoryBatchFromResponse(response) {
  const batch = response?.batch;
  if (!text(batch?.id)) throw new Error('批量工厂作品读取结果无效');
  return batch;
}

export function batchFactoryProjectsFrom(batches) {
  if (!Array.isArray(batches)) return [];
  return batches.flatMap(batch => {
    const batchId = text(batch?.id);
    if (!batchId) return [];
    return [{
      id: `batch:${batchId}`,
      batchId,
      name: text(batch.title) || '未命名批量',
      productionMode: 'batch_factory',
      factoryMode: text(batch.productionMode) === 'comic' ? 'comic' : 'video',
      source: BATCH_FACTORY_V11_SOURCE,
      createdAt: batch.createdAt,
      updatedAt: batch.updatedAt,
      ...(batch.coverMedia?.url ? { coverMedia: batch.coverMedia } : {}),
      batch
    }];
  });
}

export function batchFactoryCoverFrom(batch, productionStatus, imageURL = '') {
  const tasks = (productionStatus?.jobs || []).flatMap(job => job?.tasks || []);
  const firstBook = batch?.books?.[0];
  const firstVideo = firstBook?.videos?.[0];
  const selectedSource = firstBook?.settingsState?.patch?.primaryUploadSource;
  if (selectedSource?.kind === 'video' && selectedSource.videoId === firstVideo?.id && text(selectedSource.taskId)) {
    const selectedTask = tasks.find(item => item?.id === selectedSource.taskId && item?.videoId === firstVideo.id && item?.status === 'succeeded' && text(item?.mediaUrl));
    if (selectedTask) return { kind: 'video', url: text(selectedTask.mediaUrl) };
  }
  for (const book of batch?.books || []) for (const video of book?.videos || []) {
    const task = tasks.find(item => item?.videoId === video.id && item?.status === 'succeeded' && text(item?.mediaUrl));
    if (task) return { kind: 'video', url: text(task.mediaUrl) };
  }
  const image = text(imageURL);
  return image ? { kind: 'image', url: image } : null;
}

// A historical local merge can be purged after its retention window while its
// successful provider videos remain reachable. This is deliberately used only
// after the primary merge-cover request has failed; it is not list-wide cover
// hydration and therefore cannot turn the project library into an N+1 poller.
export function batchFactoryFallbackCoverFromProductionStatus(productionStatus) {
  for (const job of productionStatus?.jobs || []) for (const task of job?.tasks || []) {
    if (task?.status === 'succeeded' && text(task?.mediaUrl)) return { kind: 'video', url: text(task.mediaUrl) };
  }
  return null;
}

export async function enrichBatchProjectCovers(projects, {
  loadProduction,
  loadMerge,
  coverFrom,
  onUpdate,
  shouldContinue = () => true,
  concurrency = 2
} = {}) {
  const queue = Array.isArray(projects) ? projects : [];
  const limit = Math.max(1, Math.min(Number(concurrency) || 1, queue.length || 1));
  let cursor = 0;
  async function worker() {
    while (cursor < queue.length && shouldContinue()) {
      const project = queue[cursor];
      cursor += 1;
      let hadRequestFailure = false;
      const productionStatus = await loadProduction(project).catch(() => {
        hadRequestFailure = true;
        return null;
      });
      if (!shouldContinue()) return;
      const mergeStatus = await loadMerge(project).catch(() => {
        hadRequestFailure = true;
        return null;
      });
      if (!shouldContinue()) return;
      const coverMedia = coverFrom(project, productionStatus, mergeStatus);
      onUpdate({ ...project, coverMedia: coverMedia || null }, { hadRequestFailure });
    }
  }
  await Promise.all(Array.from({ length: limit }, () => worker()));
}

export function isBatchFactoryV11Project(project) {
  return project?.source === BATCH_FACTORY_V11_SOURCE && Boolean(text(project.batchId));
}
