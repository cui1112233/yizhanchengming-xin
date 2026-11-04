import { batchFactorySourceText } from './batchFactoryContentRange.js';
import automationStagesContract from './automationStages.json' with { type: 'json' };

function taskState(task) { return String(task?.status || '').trim().toLowerCase(); }

// 阶段文案单一来源：automationStages.json（Go 侧由 TestAutomationStagesContract
// 守护一致）。契约缺失或缺条目时回退到本地文案，保证构建与旧数据不受影响。
const automationStageLabels = (() => {
  const map = {};
  for (const stage of automationStagesContract?.stages || []) {
    if (stage?.id && stage?.mode && stage?.label) map[`${stage.mode}:${stage.id}`] = stage.label;
  }
  return map;
})();

function stageLabel(mode, id, fallback) {
  return automationStageLabels[`${mode}:${id}`] || fallback;
}

const terminalAutomationBookStates = new Set(['failed', 'blocked', 'stopped', 'completed']);

// 超过该时长的失败记录只作为历史记录展示，不再冒充当前自动化状态。
export const AUTOMATION_BOOK_ERROR_MAX_AGE_MS = 7 * 24 * 60 * 60 * 1000;

// 卡片只能消费“属于当前批次 + 属于本书”的自动化记录。
// 跨批次的历史 job 即便 bookId 撞上，也不允许渲染到当前书卡。
export function automationBookForCard(status, batchId, bookId, now = Date.now()) {
  const statusBatch = String(status?.batchId || '').trim();
  const currentBatch = String(batchId || '').trim();
  if (statusBatch && currentBatch && statusBatch !== currentBatch) return null;
  const wanted = String(bookId || '').trim();
  if (!wanted) return null;
  const book = (Array.isArray(status?.books) ? status.books : [])
    .find(item => String(item?.bookId || '').trim() === wanted);
  if (!book) return null;
  const state = String(book.status || '').trim().toLowerCase();
  const stamp = Date.parse(String(book.updatedAt || book.finishedAt || book.startedAt || ''));
  const stale = terminalAutomationBookStates.has(state) && state !== 'completed'
    && Number.isFinite(stamp) && (now - stamp) > AUTOMATION_BOOK_ERROR_MAX_AGE_MS;
  return stale ? { ...book, stale: true } : { ...book, stale: false };
}

export function automationBookNeedsRetry(status, bookId) {
  const wanted = String(bookId || '').trim();
  if (!wanted) return null;
  const book = (Array.isArray(status?.books) ? status.books : []).find(item => String(item?.bookId || '').trim() === wanted);
  const state = String(book?.status || '').trim().toLowerCase();
  return book && terminalAutomationBookStates.has(state) && state !== 'completed'
    ? { stage: String(book.stage || '').trim() }
    : null;
}

export function automationBookIsActive(status, bookId) {
  const state = String(status?.state || '').trim().toLowerCase();
  if (!['running', 'scheduled', 'paused', 'needs_attention'].includes(state)) return false;
  const wanted = String(bookId || '').trim();
  if (!wanted) return false;
  const book = (Array.isArray(status?.books) ? status.books : []).find(item => String(item?.bookId || '').trim() === wanted);
  if (!book) return true;
  return !terminalAutomationBookStates.has(String(book.status || '').trim().toLowerCase());
}

export function batchFactoryAutomationRetryAction(status) {
  const queued = (Array.isArray(status?.books) ? status.books : []).filter(book => book?.retryRequested === true && taskState(book) !== 'completed').length;
  const failed = Number(status?.counts?.failed || 0) + Number(status?.counts?.blocked || 0);
  const terminal = ['failed', 'completed', 'stopped'].includes(String(status?.state || '').trim().toLowerCase());
  if (queued && terminal) {
    return { label: `重新启动重试（${queued}）`, disabled: false };
  }
  return {
    label: queued ? `重试队列中（${queued}）` : `重试失败小说（${failed}）`,
    disabled: queued > 0 || failed < 1
  };
}

// Provider responses are retained by the server for diagnosis, but raw HTTP
// responses are not useful instructions for a production user.
export function batchFactoryVisibleError(error) {
  const raw = String(error || '').trim();
  if (!raw) return '';
  if (/\b402\b|insufficient[_\s-]?funds|payment required/i.test(raw)) {
    return '当前视频模型的账户额度或权限不足，请更换视频模型，或检查该模型账户的额度与权限。';
  }
  if (/\b40[13]\b|invalid[_\s-]?api[_\s-]?key|unauthori[sz]ed|密钥无效|无权限/i.test(raw)) {
    return '视频模型的 API Key 无效或无权限，请到「API 配置」检查或重建密钥。';
  }
  if (/\b429\b|too many requests|rate limit|限流/i.test(raw)) {
    return '视频模型接口限流，请稍后再试；系统会按上游 Retry-After 自动退避。';
  }
  return raw;
}

function bookTasks(book, productionStatus) {
  const jobs = (Array.isArray(productionStatus?.jobs) ? productionStatus.jobs : []).filter(job => job?.bookId === book?.id);
  const directorRevisionID = String(book?.directorRevision?.id || book?.directorRevisionId || '').trim();
  const currentRevisionJobs = directorRevisionID ? jobs.filter(job => String(job?.directorRevisionId || '').trim() === directorRevisionID) : [];
  return (currentRevisionJobs.length ? currentRevisionJobs : jobs).flatMap(job => Array.isArray(job?.tasks) ? job.tasks : []);
}
function taskUpdatedAt(task, index) {
  const timestamp = Date.parse(task?.updatedAt || task?.createdAt || '');
  return Number.isFinite(timestamp) ? timestamp : index;
}

// Production providers expose task state by VIDEO id. Convert that transport-level
// state into the book's storyboard order before it reaches any UI surface.
export function batchFactoryVideoProgress(book, productionStatus) {
  const videos = Array.isArray(book?.videos) ? book.videos : [];
  const latestByVideo = new Map();
  bookTasks(book, productionStatus).forEach((task, index) => {
    const videoID = String(task?.videoId || '').trim();
    if (!videoID) return;
    const previous = latestByVideo.get(videoID);
    if (!previous || taskUpdatedAt(task, index) >= taskUpdatedAt(previous.task, previous.index)) latestByVideo.set(videoID, { task, index });
  });

  const byVideo = new Map();
  const active = [];
  const failed = [];
  for (const [index, video] of videos.entries()) {
    const videoID = String(video?.id || '').trim();
    const task = latestByVideo.get(videoID)?.task || null;
    const status = taskState(task);
    const storyboardLabel = `分镜 ${index + 1}`;
    let message = '';
    if (status === 'running') message = `正在生成${storyboardLabel}…`;
    else if (status === 'queued') message = `${storyboardLabel} 已排队，等待生成…`;
    else if (status === 'failed') message = `${storyboardLabel} 生成失败${batchFactoryVisibleError(task?.errorMessage) ? `：${batchFactoryVisibleError(task?.errorMessage)}` : ''}`;
    else if (status === 'succeeded') message = `${storyboardLabel} 已生成`;
    const row = { videoID, index, task, status, storyboardLabel, message };
    byVideo.set(videoID, row);
    if (status === 'running' || status === 'queued') active.push(row);
    if (status === 'failed') failed.push(row);
  }
  return { byVideo, active, failed, summary: active.map(item => item.status === 'running' ? `正在生成${item.storyboardLabel}` : `${item.storyboardLabel} 已排队`).join('；') };
}

function hasManualAdjustment(book) {
  if (Object.keys(book?.settingsState?.patch || {}).length) return true;
  if ((book?.assetRecords || []).some(asset => asset?.source === 'manual')) return true;
  return (book?.videos || []).some(video => Object.keys(video?.settingsState?.patch || {}).length > 0);
}

function latestStageRun(summary, stage) {
  return (summary?.runs || []).filter(run => run?.stage === stage).at(-1) || null;
}

function statusFromRun(run, fallback = 'pending') {
  const status = taskState(run);
  if (status === 'succeeded') return 'completed';
  if (['queued', 'running'].includes(status)) return 'running';
  if (status === 'failed') return 'failed';
  return fallback;
}

function websiteSubmitLabel(book) {
  const value = String(book?.sourceMetadata?.websiteSubmitStatus || book?.sourceMetadata?.publishStatus || '').trim().toLowerCase();
  if (['succeeded', 'success', 'completed', 'uploaded'].includes(value)) return '已回读';
  if (value === 'upload_pending_confirm') return '已上传待确认';
  if (['running', 'uploading', 'queued', 'submitted'].includes(value)) return '待回读';
  if (['failed', 'error'].includes(value)) return '失败';
  return '—';
}

function uploadStatus(book) {
  const value = String(book?.sourceMetadata?.publishStatus || book?.sourceMetadata?.websiteSubmitStatus || '').trim().toLowerCase();
  if (['running', 'uploading', 'queued', 'submitted', 'upload_pending_confirm'].includes(value)) return 'running';
  if (['succeeded', 'success', 'completed', 'uploaded'].includes(value)) return 'completed';
  if (['failed', 'error'].includes(value)) return 'failed';
  return 'pending';
}

export function batchFactoryBookTimeline(book, { productionStatus, mergeStatus, stageSummary, stageSummaries, productionMode = 'video' } = {}) {
  // The detail modal historically passed the per-book summary under the
  // plural key. Accept both shapes so its timeline uses the same live status
  // as the list row.
  stageSummary = stageSummary || stageSummaries;
  const comicMode = productionMode === 'comic' || book?.settingsState?.patch?.productionMode === 'comic' || book?.settingsState?.patch?.aiPromptConfig?.comic?.enabled === true || (book?.assetRecords || []).some(asset => asset?.source === 'comic') || (book?.videos || []).some(video => String(video?.comicPrompt || '').trim());
  const sourceReady = Boolean(batchFactorySourceText(book?.sourceText));
  const assets = latestStageRun(stageSummary, 'assets');
  const director = latestStageRun(stageSummary, 'director');
  const image = latestStageRun(stageSummary, 'image');
  const videoRun = latestStageRun(stageSummary, 'video');
  const assetsReady = statusFromRun(assets, (book?.assetRecords || []).length ? 'completed' : 'pending');
  const promptReady = statusFromRun(director, (book?.videos || []).some(video => String(video?.videoPrompt || video?.description || '').trim()) ? 'completed' : 'pending');
  const allVideoTasks = bookTasks(book, productionStatus);
  const videoProgress = batchFactoryVideoProgress(book, productionStatus);
  // For a storyboard with retries, the latest task is its authoritative
  // state. Historical failures remain visible in logs but cannot override a
  // later successful retry in the book-level workflow card.
  const hasStoryboards = (book?.videos || []).length > 0;
  const failedVideo = videoProgress.failed[0]?.task || (!hasStoryboards ? allVideoTasks.find(task => taskState(task) === 'failed') : null);
  const activeVideo = videoProgress.active[0]?.task || (!hasStoryboards ? allVideoTasks.find(task => ['queued', 'running'].includes(taskState(task))) : null);
  const completedVideos = (book?.videos || []).length > 0 && (book?.videos || []).every(video => videoProgress.byVideo.get(String(video?.id || ''))?.status === 'succeeded');
  const merge = (mergeStatus?.jobs || []).filter(job => job?.bookId === book?.id).at(-1);
  const upload = uploadStatus(book);
  const videoStatus = failedVideo ? 'failed' : activeVideo ? 'running' : completedVideos ? 'completed' : statusFromRun(videoRun);
  const mergeStatusValue = taskState(merge);
  const merged = mergeStatusValue === 'succeeded' && String(merge?.outputUrl || '').trim();
  const mergeFailed = mergeStatusValue === 'failed';
  const mergeRunning = ['queued', 'running'].includes(mergeStatusValue);
  const uploadDetail = upload === 'running' ? '正在上传网络' : upload === 'completed' ? '上传成功' : upload === 'failed' ? (batchFactoryVisibleError(book?.sourceMetadata?.publishError) || '上传失败') : merged ? '等待上传网络' : '完成合成后可上传';
  if (comicMode) {
    // 手动发布会把记录写进合集发布草稿，而不会回写 book.status/stage；
    // 运行时汇总里的已发布合集数量同样是权威信号，二者取其一即视为已发布。
    const comicPublished = (String(book?.status || '').toLowerCase() === 'completed'
      && String(book?.stage || '').toLowerCase() === 'published')
      || Number(stageSummary?.comicPublishedCollectionCount || 0) > 0;
    const comicAssetsReady = comicPublished || (book?.assetRecords || []).some(asset => asset?.source === 'comic' && String(asset?.prompt || asset?.visualPrompt || '').trim());
    const comicPromptsReady = comicPublished || ((book?.videos || []).length > 0 && (book?.videos || []).every(video => String(video?.comicPrompt || '').trim() && String(video?.comicImagePrompt || '').trim()));
      const comicImageRun = latestStageRun(stageSummary, 'image');
      const comicImageReady = comicPublished || stageSummary?.comicImageReady === true;
      const comicImageStatus = comicImageReady ? 'completed' : statusFromRun(comicImageRun, 'pending');
    const comicCollectionReady = stageSummary?.comicCollectionReady === true || comicPublished;
    const comicCollectionCount = Number(stageSummary?.comicCollectionCount || 0);
    const comicCollectionPageCount = Number(stageSummary?.comicCollectionPageCount || 0);
    return [
      { key: 'source', label: stageLabel('comic', 'source', '原文已就绪'), status: sourceReady ? 'completed' : 'pending', detail: sourceReady ? '原文已保存' : '等待原文' },
      { key: 'assets', label: stageLabel('comic', 'assets', '漫画资产已就绪'), status: comicAssetsReady ? 'completed' : 'pending', detail: batchFactoryVisibleError(assets?.errorMessage) || (comicAssetsReady ? '人物、场景、道具已准备' : '等待提取漫画资产') },
      { key: 'director', label: stageLabel('comic', 'director', '漫画分页与提示词已就绪'), status: comicPromptsReady ? 'completed' : 'pending', detail: batchFactoryVisibleError(director?.errorMessage) || (comicPromptsReady ? '分页与每页分镜提示词已生成' : '等待漫画分页与提示词') },
      { key: 'video', label: stageLabel('comic', 'video', '漫画图片生成'), status: comicImageStatus, detail: batchFactoryVisibleError(comicImageRun?.errorMessage) || (comicImageStatus === 'completed' ? `漫画页图片已生成${stageSummary?.comicImageTotalCount ? `（${stageSummary.comicImageReadyCount}/${stageSummary.comicImageTotalCount} 页）` : ''}` : comicImageStatus === 'running' ? '正在逐页生成漫画图片' : '等待生成漫画图片') },
      { key: 'complete', label: stageLabel('comic', 'complete', '漫画合集'), status: comicCollectionReady ? 'completed' : 'pending', detail: comicPublished ? '漫画合集已发布' : comicCollectionReady ? `已整理 ${comicCollectionCount || 1} 个合集${comicCollectionPageCount ? `（${comicCollectionPageCount} 页）` : ''}` : '等待全部漫画页主图生成后整理合集' }
    ];
  }
  return [
    { key: 'source', label: stageLabel('video', 'source', '原文已就绪'), status: sourceReady ? 'completed' : 'pending', detail: sourceReady ? '原文已保存' : '等待原文' },
    { key: 'assets', label: stageLabel('video', 'assets', '资产（人物/场景/图片）已就绪'), status: assetsReady, detail: batchFactoryVisibleError(assets?.errorMessage) || (assetsReady === 'completed' ? '资产已生成或已维护' : '等待资产处理') },
    { key: 'director', label: stageLabel('video', 'director', '文本提示词已就绪'), status: promptReady, detail: batchFactoryVisibleError(director?.errorMessage) || (promptReady === 'completed' ? '分镜与提示词已生成' : '等待生成文本提示词') },
    { key: 'video', label: stageLabel('video', 'video', '视频生成'), status: videoStatus, detail: batchFactoryVisibleError(failedVideo?.errorMessage) || videoProgress.summary || batchFactoryVisibleError(videoRun?.errorMessage) || (videoStatus === 'completed' ? 'VIDEO 已生成' : '等待生成视频') },
    { key: 'upload', label: upload === 'running' ? '正在上传网络' : upload === 'completed' ? '上传成功' : stageLabel('video', 'upload', '上传网络'), status: upload, detail: uploadDetail },
    { key: 'complete', label: stageLabel('video', 'complete', '完成'), status: upload === 'completed' ? 'completed' : (mergeFailed ? 'failed' : mergeRunning ? 'running' : 'pending'), detail: batchFactoryVisibleError(merge?.errorMessage) || (upload === 'completed' ? '本书生产与上传均已完成' : mergeRunning ? '正在合成最终视频' : merged ? '等待上传成功回执' : '等待最终视频合成') }
  ];
}

export function shouldAutoClassifyBook(book) {
  if (!String(book?.sourceText || '').trim()) return false;
  const metadata = book?.sourceMetadata || {};
  const status = String(metadata.classifyStatus || '').trim().toLowerCase();
  if (['classified', 'manual'].includes(status)) return false;
  const gender = String(metadata.gender || metadata.gender_name || '').trim();
  const style = String(metadata.style || metadata.styleName || metadata.style_name || '').trim();
  return !gender || gender === '待确认' || !style || ['待判定', '未判定'].includes(style);
}

function classificationReady(metadata = {}) {
  const gender = String(metadata.gender || metadata.gender_name || '').trim();
  const style = String(metadata.style || metadata.styleName || metadata.style_name || '').trim();
  return Boolean(gender && gender !== '待确认' && style && !['待判定', '未判定'].includes(style));
}

export function batchFactoryBookState(book, { productionStatus, mergeStatus, stageSummary, productionMode = 'video', automationBook } = {}) {
  const sourceText = batchFactorySourceText(book?.sourceText);
  const queueStatus = String(book?.sourceMetadata?.queueStatus || '').trim();
  const manual = hasManualAdjustment(book);
  const classifyStatus = String(book?.sourceMetadata?.classifyStatus || '').trim().toLowerCase();
  const classifyError = String(book?.sourceMetadata?.classifyError || '').trim();
  // Once the backend has explicitly recorded the classification stage, show it
  // before production stages. Legacy books without this field keep their old
  // display and can be classified from the detail modal.
  if (sourceText && ['failed', 'pending_confirmation'].includes(classifyStatus) && !classificationReady(book?.sourceMetadata)) {
    return { label: '待分类', detail: `男女频/风格分类：${classifyError || '正在自动重试'}`, tone: 'blue', manual };
  }
  if (sourceText && ['processing', 'queued'].includes(classifyStatus)) {
    return { label: '执行中…', detail: '男女频/风格分类：正在识别', tone: 'blue', manual };
  }
  const comicMode = productionMode === 'comic' || book?.settingsState?.patch?.productionMode === 'comic' || book?.settingsState?.patch?.aiPromptConfig?.comic?.enabled === true || (book?.assetRecords || []).some(asset => asset?.source === 'comic') || (book?.videos || []).some(video => String(video?.comicPrompt || '').trim());
  const timeline = batchFactoryBookTimeline(book, { productionStatus, mergeStatus, stageSummary, productionMode: comicMode ? 'comic' : productionMode });
  const completeStep = timeline.find(item => item.key === 'complete');
  // 漫画模式下“完成”只代表合集已整理；只有发布记录落地（或状态回写）才算发布完成。
  const published = comicMode
    ? Number(stageSummary?.comicPublishedCollectionCount || 0) > 0
      || (String(book?.status || '').toLowerCase() === 'completed' && String(book?.stage || '').toLowerCase() === 'published')
    : completeStep?.status === 'completed';
  // 已经回读确认上传成功，就以“完成”收口；中间某次重试遗留的分镜失败
  // 不能把一件已经干完的活重新标红。
  if (completeStep?.status === 'completed') return { label: '完成', detail: completeStep.detail || '本书生产与上传均已完成', tone: 'green', manual, published };
  const failedStep = timeline.find(item => item.status === 'failed');
  if (failedStep) return { label: '异常', detail: `${failedStep.label}：${failedStep.detail}`, tone: 'red', manual, failedStep };
  const activeStep = timeline.find(item => item.status === 'running');
  if (activeStep) return { label: '执行中…', detail: `${activeStep.label}：${activeStep.detail}`, tone: 'blue', manual };
  // 自动化任务可能在时间线之外的阶段失败（例如 audio_planning 的配音时长测量）。
  // 这种情况过去会一路落到“待开始/原文已就绪”，用户只能看到“等待”却查不到原因。
  if (!failedStep && taskState(automationBook) === 'failed') {
    const stage = String(automationBook?.stage || '').trim();
    const stageText = stageLabel('video', stage, stage === 'audio_planning' ? '配音与逐行时长' : '自动化');
    const message = String(automationBook?.message || '').trim();
    const error = String(automationBook?.error || '').trim();
    const detail = message || `${stageText}失败`;
    return { label: '异常', detail: error && error !== detail ? `${detail}（${batchFactoryVisibleError(error)}）` : detail, tone: 'red', manual, failedStage: { stage, message: detail, errorMessage: error } };
  }
  if (comicMode) {
    if (timeline.find(item => item.key === 'video')?.status === 'completed') return { label: '待整理', detail: '漫画页已生成，等待整理漫画合集', tone: 'cyan', manual };
    if (timeline.find(item => item.key === 'director')?.status === 'completed') return { label: '待生成', detail: '漫画分页与提示词已就绪，等待生成漫画图片', tone: 'cyan', manual };
  }
  const tasks = bookTasks(book, productionStatus);
  const videoProgress = batchFactoryVideoProgress(book, productionStatus);
  const failed = videoProgress.failed[0]?.task;
  if (failed) return { label: '异常', detail: videoProgress.failed[0].message, tone: 'red', manual };
  if (videoProgress.active.length) return { label: '处理中', detail: videoProgress.summary, tone: 'blue', manual };
  const cancelled = tasks.find(task => taskState(task) === 'cancelled');
  if (cancelled) return { label: '已取消', detail: `${cancelled.videoId || 'VIDEO'} 已取消`, tone: 'default', manual };

  const merges = (Array.isArray(mergeStatus?.jobs) ? mergeStatus.jobs : [])
    .filter(job => job?.bookId === book?.id);
  const failedMerge = merges.find(job => taskState(job) === 'failed');
  if (failedMerge) return { label: '异常', detail: `合并失败${batchFactoryVisibleError(failedMerge.errorMessage) ? `：${batchFactoryVisibleError(failedMerge.errorMessage)}` : ''}`, tone: 'red', manual };
  if (merges.some(job => ['queued', 'running'].includes(taskState(job)))) return { label: '处理中', detail: '合并中', tone: 'blue', manual };
  if (merges.some(job => taskState(job) === 'succeeded' && String(job?.outputUrl || '').trim())) return { label: '待上传', detail: '视频已合并，等待视频管理系统提交', tone: 'cyan', manual };
  if (timeline.find(item => item.key === 'video')?.status === 'completed') return { label: '待合成', detail: '所有分镜视频已生成，等待合成', tone: 'cyan', manual };

  if (queueStatus === 'scheduled_waiting') return { label: '待开始', detail: '定时待执行，等待设定时间释放', tone: 'default', manual };
  if (sourceText) return { label: '待开始', detail: '原文已就绪', tone: 'default', manual };
  return { label: '待开始', detail: '等待按书城与 bookId 获取原文', tone: 'default', manual };
}

export function batchFactoryBatchProgress(books, runtime = {}) {
  const result = {
    total: 0,
    uploaded: 0,
    failed: 0,
    running: 0,
    awaitingUpload: 0,
    pending: 0,
    completionPercent: 0
  };
  const values = Array.isArray(books) ? books : [];
  result.total = values.length;
  for (const book of values) {
    const state = batchFactoryBookState(book, {
      ...runtime,
      stageSummary: runtime.stageSummaries?.[book?.id] || runtime.stageSummary
    });
    if (state.label === '完成') result.uploaded += 1;
    else if (state.label === '异常') result.failed += 1;
    else if (state.label === '执行中…' || state.label === '处理中') result.running += 1;
    else if (state.label === '待上传') result.awaitingUpload += 1;
    else result.pending += 1;
  }
  result.completionPercent = result.total ? Math.round((result.uploaded / result.total) * 100) : 0;
  return result;
}

export function batchFactoryNovelTableRow(book, index, createdAt = '', runtime = {}) {
  const sourceText = batchFactorySourceText(book?.sourceText);
  const state = batchFactoryBookState(book, runtime);
  return {
    id: index + 1,
    title: String(book?.title || '').trim() || `小说 ${index + 1}`,
    bookId: String(book?.bookId || '').trim() || '—',
    original: sourceText ? '✓' : '—',
    ai1: '—',
    websiteSubmit: websiteSubmitLabel(book),
    status: state.label,
    state,
    chars: sourceText ? String(sourceText.length) : '—',
    createdAt: String(createdAt || '').trim() || '—'
  };
}
