import test from 'node:test';
import assert from 'node:assert/strict';
import { shouldAutoClassifyBook, automationBookForCard, automationBookIsActive, automationBookNeedsRetry, batchFactoryAutomationRetryAction, batchFactoryBatchProgress, batchFactoryBookState, batchFactoryBookTimeline, batchFactoryNovelTableRow, batchFactoryVideoProgress, batchFactoryVisibleError, bookFailureIsHistorical } from './batchFactoryBookState.js';

const textStages = [
  ['text_process', '文本处理', '文本处理失败', '文本处理完成'],
  ['text_paginate', '正在分页', '分页失败', '分页完成'],
  ['text_render', '正在渲染', '渲染失败', '渲染完成'],
  ['text_validate', '正在校验', '校验失败', '校验完成'],
  ['text_collection', '正在整理合集', '合集整理失败', '合集就绪'],
  ['text_publish', '正在上传', '上传失败', '已上传']
];
const verifiedTextPublish = { status: 'published', carouselId: 42, imageIds: [17, 19] };

test('text timeline follows each backend attempt and displays the stage outcome', () => {
  const book = { id: 'book', sourceText: '正文' };
  const timeline = (stageSummary = {}) => batchFactoryBookTimeline(book, { productionMode: 'text', stageSummaries: stageSummary });
  assert.deepEqual(timeline().map(step => step.key), textStages.map(([key]) => key));
  assert.equal(batchFactoryBookState(book, { productionMode: 'text' }).label, '等待处理');
  for (const [stage, running, failed, completed] of textStages) {
    const runs = [];
    const summary = { runs };
    const current = () => timeline(summary).find(step => step.key === stage);
    runs.push({ stage, status: 'running' });
    assert.equal(current().status, 'running');
    assert.equal(current().label, running);
    assert.equal(batchFactoryBookState(book, { productionMode: 'text', stageSummary: summary }).label, running);
    runs.push({ stage, status: 'failed', errorMessage: '本次任务失败' });
    assert.equal(current().status, 'failed');
    assert.equal(current().label, failed);
    assert.equal(current().detail, '本次任务失败');
    runs.push({ stage, status: 'succeeded' });
    if (stage === 'text_publish') summary.textPublishRecord = verifiedTextPublish;
    assert.equal(current().status, 'completed');
    assert.equal(current().label, completed);
  }
});

test('text upload completion requires a confirmed text receipt for this book', () => {
  const book = { id: 'book', status: 'completed', stage: 'published', sourceMetadata: { websiteSubmitStatus: 'uploaded' } };
  const stageSummary = {
    runs: [{ stage: 'text_collection', status: 'succeeded' }, { stage: 'text_publish', status: 'succeeded' }],
    comicPublishedCollectionCount: 1
  };
  const runtime = { productionMode: 'text', runMode: 'full_submit', textAutoUpload: true, stageSummary };
  for (const record of [null, { ...verifiedTextPublish, status: 'accepted' }, { ...verifiedTextPublish, carouselId: 0 },
    { ...verifiedTextPublish, imageIds: [] }, { ...verifiedTextPublish, kind: 'comic' },
    { ...verifiedTextPublish, bookId: 'other-book' }]) {
    stageSummary.textPublishRecord = record;
    assert.equal(batchFactoryBookState(book, runtime).terminal, false);
    assert.equal(batchFactoryBookState(book, runtime).label, '正在上传');
    assert.equal(batchFactoryNovelTableRow(book, 0, '', runtime).websiteSubmit, '待回读');
    assert.equal(batchFactoryBatchProgress([book], runtime).uploaded, 0);
  }
  stageSummary.textPublishRecord = verifiedTextPublish;
  assert.equal(batchFactoryBookState(book, runtime).label, '已上传');
  assert.equal(batchFactoryBookState(book, runtime).terminal, true);
  assert.equal(batchFactoryNovelTableRow(book, 0, '', runtime).websiteSubmit, '已回读');
  assert.equal(batchFactoryBatchProgress([book], runtime).uploaded, 1);
});

test('a previous text receipt preserves republish eligibility while the new attempt runs or fails', () => {
  const book = { id: 'book' };
  const previous = { stage: 'text_publish', status: 'succeeded', attempt: 1, requestId: 'first' };
  const current = { stage: 'text_publish', status: 'running', attempt: 2, requestId: 'second', createdAt: '2026-11-24T11:00:00Z' };
  const stageSummary = {
    runs: [previous, current],
    textPublishRecord: { ...verifiedTextPublish, requestKey: 'first', attempt: 1, publishedAt: '2026-11-24T11:01:00Z' }
  };
  const runtime = { productionMode: 'text', stageSummary };
  const timeline = () => batchFactoryBookTimeline(book, runtime).at(-1);
  const state = () => batchFactoryBookState(book, runtime);
  const progress = () => batchFactoryBatchProgress([book], runtime);
  assert.equal(timeline().label, '正在上传');
  assert.equal(state().published, true);
  assert.equal(state().terminal, false);
  assert.equal(progress().running, 1);
  assert.equal(progress().uploaded, 0);
  current.status = 'failed';
  current.errorMessage = '第二次上传超时';
  assert.equal(timeline().label, '上传失败');
  assert.equal(state().detail, '第二次上传超时');
  assert.equal(state().published, true);
  assert.equal(state().terminal, false);
  assert.equal(progress().failed, 1);
  assert.equal(progress().completionPercent, 0);
  stageSummary.textPublishRecord = { ...verifiedTextPublish, requestKey: 'second', attempt: 2 };
  assert.equal(timeline().label, '已上传');
  assert.equal(state().terminal, true);
  assert.equal(progress().completionPercent, 100);
});

test('text receipt comparison uses attempt identity or publication time when a request key is missing', () => {
  const run = { stage: 'text_publish', status: 'running', attempt: 2, createdAt: '2026-11-24T11:00:00Z' };
  const stageSummary = { runs: [run] };
  const state = () => batchFactoryBookState({ id: 'book' }, { productionMode: 'text', stageSummary });
  for (const [fields, terminal] of [
    [{ attempt: 1 }, false], [{ attempt: 2 }, true],
    [{ publishedAt: '2026-11-24T10:59:00Z' }, false], [{ publishedAt: '2026-11-24T11:01:00Z' }, true],
    [{}, false]
  ]) {
    stageSummary.textPublishRecord = { ...verifiedTextPublish, ...fields };
    assert.equal(state().terminal, terminal);
  }
  run.status = 'succeeded';
  assert.equal(state().terminal, false);
  stageSummary.runs = [
    { stage: 'text_publish', status: 'succeeded' },
    { stage: 'text_publish', status: 'succeeded' }
  ];
  assert.equal(state().terminal, false);
});

test('a ready text collection is terminal only for a non-submit job and remains distinct from an upload', () => {
  const book = { id: 'book' };
  const stageSummary = { runs: [{ stage: 'text_collection', status: 'succeeded' }] };
  const runtime = { productionMode: 'text', stageSummary };
  assert.equal(batchFactoryBookState(book, runtime).terminal, false);
  stageSummary.runMode = 'full_submit';
  assert.equal(batchFactoryBookState(book, runtime).terminal, false);
  assert.equal(batchFactoryBookState(book, runtime).status, 'pending');
  assert.equal(batchFactoryBatchProgress([book], runtime).awaitingUpload, 1);
  stageSummary.runMode = 'video_no_submit';
  stageSummary.textAutoUpload = true;
  assert.equal(batchFactoryBookState(book, runtime).terminal, false);
  stageSummary.textAutoUpload = false;
  const state = batchFactoryBookState(book, runtime);
  assert.equal(state.label, '合集就绪');
  assert.equal(state.terminal, true);
  assert.equal(state.published, false);
  const batch = batchFactoryBatchProgress([book], runtime);
  assert.equal(batch.uploaded, 0);
  assert.equal(batch.completionPercent, 100);
});

test('a pending receipt keeps the book uploading while stage logs catch up', () => {
  const runtime = {
    productionMode: 'text',
    stageSummary: { textPublishRecord: { kind: 'text', bookId: 'book', status: 'accepted' } }
  };
  const state = batchFactoryBookState({ id: 'book' }, runtime);
  assert.equal(state.label, '正在上传');
  assert.equal(state.terminal, false);
  assert.equal(batchFactoryBookState({ id: 'unrelated' }, runtime).label, '等待处理');
});

test('text preflight failures are visible before stage logs are available', () => {
  const state = batchFactoryBookState({ id: 'book' }, {
    productionMode: 'text', automationBook: { stage: 'preflight', status: 'failed', error: '缺少原文' }
  });
  assert.equal(state.tone, 'red');
  assert.match(state.detail, /缺少原文/);
  assert.equal(state.failedStage.stage, 'preflight');
});

test('a recovered failure is historical and must not show a retry action', () => {
  const lastFailed = { stage: 'video', updatedAt: '2026-11-04T10:00:00Z', errorMessage: '视频生成失败' };
  const runs = [
    { stage: 'video', status: 'succeeded', updatedAt: '2026-11-04T10:30:00Z' }
  ];
  assert.equal(bookFailureIsHistorical({ lastFailed, runs }), true);
  assert.equal(bookFailureIsHistorical({ lastFailed, runs: [{ stage: 'video', status: 'succeeded', updatedAt: '2026-11-04T09:00:00Z' }] }), false);
  assert.equal(bookFailureIsHistorical({ lastFailed, runs: [{ stage: 'merge', status: 'succeeded', updatedAt: '2026-11-04T11:00:00Z' }] }), false);
  assert.equal(bookFailureIsHistorical({ automationBook: { status: 'completed' }, lastFailed }), true);
  assert.equal(bookFailureIsHistorical({ automationBook: { status: 'failed', stage: 'video' }, lastFailed, runs }), false);
  assert.equal(bookFailureIsHistorical({}), false);
});

test('other batches legacy failure never renders on this book card', () => {
  const status = {
    batchId: 'batch-old',
    state: 'failed',
    books: [{ bookId: 'book-1', status: 'failed', stage: 'audio_planning', error: 'TTS 时长测量失败：exec: "ffprobe"' }]
  };
  assert.equal(automationBookForCard(status, 'batch-current', 'book-1'), null);
});

test('same-batch failure older than threshold is marked as history, not current state', () => {
  const status = {
    batchId: 'batch-current',
    books: [{ bookId: 'book-1', status: 'failed', stage: 'audio_planning', updatedAt: '2026-10-14T11:16:45Z' }]
  };
  const book = automationBookForCard(status, 'batch-current', 'book-1', Date.parse('2026-11-04T12:00:00Z'));
  assert.equal(book.stale, true);
  const fresh = automationBookForCard(status, 'batch-current', 'book-1', Date.parse('2026-10-15T12:00:00Z'));
  assert.equal(fresh.stale, false);
  const otherBook = automationBookForCard(status, 'batch-current', 'book-2', Date.parse('2026-10-15T12:00:00Z'));
  assert.equal(otherBook, null);
});

test('does not block a failed book behind a stale running automation job', () => {
  const status = { state: 'running', books: [{ bookId: 'book-1', status: 'failed', stage: 'comic_images' }] };
  assert.equal(automationBookIsActive(status, 'book-1'), false);
  assert.deepEqual(automationBookNeedsRetry(status, 'book-1'), { stage: 'comic_images' });
});

test('keeps an actually running book protected from manual duplicate retries', () => {
  const status = { state: 'running', books: [{ bookId: 'book-1', status: 'running', stage: 'comic_images' }] };
  assert.equal(automationBookIsActive(status, 'book-1'), true);
  assert.equal(automationBookNeedsRetry(status, 'book-1'), null);
});

test('surfaces an automation failure from a stage outside the timeline instead of showing 待开始', () => {
  const book = { id: 'book-1', sourceText: '正文', sourceMetadata: {} };
  assert.equal(batchFactoryBookState(book).label, '待开始');
  const state = batchFactoryBookState(book, {
    automationBook: {
      bookId: 'book-1',
      status: 'failed',
      stage: 'audio_planning',
      message: '配音与逐行时长失败',
      error: 'TTS 时长测量失败：invalid input: audio probe failed: exec: "ffprobe": executable file not found in $PATH'
    }
  });
  assert.equal(state.label, '异常');
  assert.equal(state.tone, 'red');
  assert.match(state.detail, /配音与逐行时长失败/);
  assert.match(state.detail, /ffprobe/);
  assert.equal(state.failedStage.stage, 'audio_planning');
});

test('allows a queued retry to restart a failed automation run after service recovery', () => {
  assert.deepEqual(batchFactoryAutomationRetryAction({
    state: 'failed',
    counts: { failed: 1, blocked: 0 },
    books: [{ status: 'failed', retryRequested: true }]
  }), {
    label: '重新启动重试（1）',
    disabled: false
  });
});

test('does not count an already published book as a queued retry', () => {
  assert.deepEqual(batchFactoryAutomationRetryAction({
    state: 'failed',
    counts: { failed: 1, blocked: 0 },
    books: [
      { status: 'completed', stage: 'published', retryRequested: true },
      { status: 'failed', stage: 'comic_images', retryRequested: true }
    ]
  }), {
    label: '重新启动重试（1）',
    disabled: false
  });
});

test('translates provider payment failures into an actionable status without exposing HTTP codes', () => {
  const message = batchFactoryVisibleError('YFAI Seedance returned HTTP 402');
  assert.equal(message, '当前视频模型的账户额度或权限不足，请更换视频模型，或检查该模型账户的额度与权限。');
  assert.doesNotMatch(message, /402|YFAI|Seedance/i);
});

test('retries incomplete classification once but leaves complete labels alone', () => {
  const book = sourceMetadata => ({ sourceText: '正文', sourceMetadata });
  assert.equal(shouldAutoClassifyBook(book({ classifyStatus: 'failed', gender: '女频', style: '' })), true);
  assert.equal(shouldAutoClassifyBook(book({ classifyStatus: 'failed', gender: '女频', style: '待判定' })), true);
  assert.equal(shouldAutoClassifyBook(book({ classifyStatus: 'failed', gender: '', style: '现代悬疑' })), true);
  assert.equal(shouldAutoClassifyBook(book({ classifyStatus: 'pending_confirmation', gender: '女频', style: '' })), true);
  assert.equal(shouldAutoClassifyBook(book({ classifyStatus: 'manual', gender: '女频', style: '' })), false);
  assert.equal(shouldAutoClassifyBook(book({ classifyStatus: 'classified', gender: '女频', style: '现代悬疑' })), false);
  assert.equal(shouldAutoClassifyBook(book({ classifyStatus: 'failed', gender: '女频', style: '现代悬疑' })), false);
  assert.equal(shouldAutoClassifyBook({ sourceText: '', sourceMetadata: { classifyStatus: 'failed' } }), false);
});

test('does not turn an incomplete classification into a blocking red state', () => {
  const state = batchFactoryBookState({
    id: 'book-classification-failed',
    sourceText: '正文已经获取',
    sourceMetadata: { classifyStatus: 'failed', classifyError: '文本模型未返回完整的男女频和风格' }
  });
  assert.equal(state.label, '待分类');
  assert.equal(state.tone, 'blue');
  assert.match(state.detail, /男女频\/风格分类/);
  assert.match(state.detail, /文本模型未返回/);
});

test('an incomplete classification no longer asks for manual confirmation', () => {
  const state = batchFactoryBookState({
    id: 'book-classification-confirmation',
    sourceText: '正文已经获取',
    sourceMetadata: { classifyStatus: 'pending_confirmation', classifyError: '已自动判断 2 次仍未完成：文本模型未返回有效风格' }
  });
  assert.equal(state.label, '待分类');
  assert.doesNotMatch(state.detail, /人工确认/);
  assert.match(state.detail, /已自动判断 2 次/);
});

test('complete labels flow straight to production stages regardless of status', () => {
  const state = batchFactoryBookState({
    id: 'book-classification-complete',
    sourceText: '正文已经获取',
    sourceMetadata: { classifyStatus: 'failed', gender: '女频', style: '现代悬疑' }
  });
  assert.notEqual(state.label, '待分类');
  assert.doesNotMatch(state.detail, /男女频\/风格分类/);
});

test('shows persisted classification processing before production status', () => {
  const state = batchFactoryBookState({
    id: 'book-classification-processing',
    sourceText: '正文已经获取',
    sourceMetadata: { classifyStatus: 'processing' }
  });
  assert.deepEqual({ label: state.label, detail: state.detail }, { label: '执行中…', detail: '男女频/风格分类：正在识别' });
});

test('a failed VIDEO overrides manual changes with an exception state', () => {
  const state = batchFactoryBookState({ id: 'book-1', sourceText: '正文', settingsState: { patch: { textModelId: 'text-a' } } }, {
    productionStatus: { jobs: [{ bookId: 'book-1', tasks: [{ videoId: 'video-4', status: 'failed', errorMessage: 'provider rejected' }] }] }
  });
  assert.equal(state.label, '异常');
  assert.match(state.detail, /视频生成/);
  assert.equal(state.manual, true);
});
test('completed batch merge makes a ready book wait for upload instead of completed', () => {
  const state = batchFactoryBookState({ id: 'book-1', sourceText: '正文' }, {
    mergeStatus: { jobs: [{ bookId: 'book-1', status: 'succeeded', outputUrl: 'https://media.example/final.mp4' }] }
  });
  assert.deepEqual({ label: state.label, detail: state.detail }, { label: '待上传', detail: '视频已合并，等待视频管理系统提交' });
});

test('treats manually generated comic page primaries as image-ready', () => {
  const book = {
    id: 'book-comic',
    sourceText: '正文',
    settingsState: { patch: { productionMode: 'comic' } },
    assetRecords: [{ source: 'comic', prompt: '人物' }],
    videos: [
      { id: 'page-1', comicPrompt: '画面一', comicImagePrompt: '图片一' },
      { id: 'page-2', comicPrompt: '画面二', comicImagePrompt: '图片二' }
    ]
  };
  const timeline = batchFactoryBookTimeline(book, {
    productionMode: 'comic',
    stageSummary: { runs: [], comicImageReady: true, comicImageReadyCount: 2, comicImageTotalCount: 2 }
  });
  assert.equal(timeline.find(item => item.key === 'video')?.status, 'completed');
  assert.match(timeline.find(item => item.key === 'video')?.detail || '', /2\/2/);
  assert.equal(batchFactoryBookState(book, {
    stageSummary: { runs: [], comicImageReady: true, comicImageReadyCount: 2, comicImageTotalCount: 2 },
    productionMode: 'comic'
  }).label, '待整理');
});

test('does not mark comic prompts ready when the independent image prompt is missing', () => {
  const book = {
    id: 'book-comic-missing-image-prompt',
    sourceText: '正文',
    settingsState: { patch: { productionMode: 'comic' } },
    assetRecords: [{ source: 'comic', prompt: '人物' }],
    videos: [{ id: 'page-1', comicPrompt: '画面一', comicImagePrompt: '' }]
  };
  const timeline = batchFactoryBookTimeline(book, { productionMode: 'comic' });
  assert.equal(timeline.find(item => item.key === 'director')?.status, 'pending');
  assert.match(timeline.find(item => item.key === 'director')?.detail || '', /等待漫画分页与提示词/);
});

test('marks a persisted comic collection as complete after all page images are ready', () => {
  const book = {
    id: 'book-comic',
    sourceText: '正文',
    settingsState: { patch: { productionMode: 'comic' } },
    assetRecords: [{ source: 'comic', prompt: '人物' }],
    videos: [{ id: 'page-1', comicPrompt: '画面一', comicImagePrompt: '图片一' }]
  };
  const timeline = batchFactoryBookTimeline(book, {
    productionMode: 'comic',
    stageSummary: { runs: [], comicImageReady: true, comicImageReadyCount: 1, comicImageTotalCount: 1, comicCollectionReady: true, comicCollectionCount: 1, comicCollectionPageCount: 1 }
  });
  assert.equal(timeline.find(item => item.key === 'complete')?.status, 'completed');
  assert.match(timeline.find(item => item.key === 'complete')?.detail || '', /1 个合集/);
  assert.equal(batchFactoryBookState(book, {
    stageSummary: { runs: [], comicImageReady: true, comicImageReadyCount: 1, comicImageTotalCount: 1, comicCollectionReady: true, comicCollectionCount: 1, comicCollectionPageCount: 1 },
    productionMode: 'comic'
  }).label, '完成');
});

test('treats a published comic automation book as fully complete', () => {
  const book = {
    id: 'published-comic',
    sourceText: '正文',
    status: 'completed',
    stage: 'published',
    settingsState: { patch: { productionMode: 'comic' } },
    assetRecords: [{ source: 'comic', prompt: '人物' }],
    videos: [{ id: 'page-1', comicPrompt: '画面一', comicImagePrompt: '图片一' }]
  };
  const timeline = batchFactoryBookTimeline(book, { productionMode: 'comic' });
  assert.equal(timeline.every(item => item.status === 'completed'), true);
  assert.equal(timeline.find(item => item.key === 'complete')?.detail, '漫画合集已发布');
  const state = batchFactoryBookState(book, { productionMode: 'comic' });
  assert.equal(state.label, '完成');
  assert.equal(state.published, true);
});

test('treats a manually published comic collection as published even without book status writeback', () => {
  const book = {
    id: 'manual-published-comic',
    sourceText: '正文',
    status: 'running',
    stage: 'comic_publish',
    settingsState: { patch: { productionMode: 'comic' } },
    assetRecords: [{ source: 'comic', prompt: '人物' }],
    videos: [{ id: 'page-1', comicPrompt: '画面一', comicImagePrompt: '图片一' }]
  };
  const stageSummary = { runs: [], comicCollectionReady: true, comicCollectionCount: 1, comicCollectionPageCount: 1, comicPublishedCollectionCount: 1 };
  const timeline = batchFactoryBookTimeline(book, { stageSummary, productionMode: 'comic' });
  assert.equal(timeline.find(item => item.key === 'complete')?.detail, '漫画合集已发布');
  const state = batchFactoryBookState(book, { stageSummary, productionMode: 'comic' });
  assert.equal(state.label, '完成');
  assert.equal(state.published, true);
});

test('keeps an unpublished comic collection out of the published state', () => {
  const book = {
    id: 'comic-collection-ready',
    sourceText: '正文',
    settingsState: { patch: { productionMode: 'comic' } },
    assetRecords: [{ source: 'comic', prompt: '人物' }],
    videos: [{ id: 'page-1', comicPrompt: '画面一', comicImagePrompt: '图片一' }]
  };
  const stageSummary = { runs: [], comicCollectionReady: true, comicCollectionCount: 1, comicCollectionPageCount: 1, comicPublishedCollectionCount: 0 };
  const state = batchFactoryBookState(book, { stageSummary, productionMode: 'comic' });
  assert.notEqual(state.published, true);
});

test('builds a readable per-book production timeline and keeps a failed stage actionable', () => {
  const book = { id: 'book-1', sourceText: '正文', assetRecords: [{ id: 'character-1' }], videos: [{ id: 'video-1', videoPrompt: '镜头提示词' }] };
  const runtime = {
    productionStatus: { jobs: [{ bookId: 'book-1', tasks: [{ videoId: 'video-1', status: 'failed', errorMessage: 'provider rejected' }] }] },
    stageSummary: { runs: [{ stage: 'assets', status: 'succeeded' }, { stage: 'director', status: 'succeeded' }] }
  };
  const timeline = batchFactoryBookTimeline(book, runtime);
  assert.deepEqual(timeline.slice(0, 3).map(item => item.label), ['原文已就绪', '资产（人物/场景/图片）已就绪', '文本提示词已就绪']);
  assert.equal(timeline.find(item => item.key === 'video').status, 'failed');
  assert.match(timeline.find(item => item.key === 'video').detail, /provider rejected/);
  assert.equal(batchFactoryBookState(book, runtime).label, '异常');
});

test('uses comic production labels instead of video merge and upload labels', () => {
  const book = {
    id: 'comic-book',
    sourceText: '正文',
    assetRecords: [{ id: 'comic-asset', source: 'comic' }],
    videos: [{ id: 'comic-page-1', comicPrompt: '画面提示词', comicImagePrompt: '出图提示词', comicSourceSegment: { sourceText: '正文片段' } }]
  };
  const timeline = batchFactoryBookTimeline(book, { productionMode: 'comic' });
  assert.deepEqual(timeline.map(item => item.label), ['原文已就绪', '漫画资产已就绪', '漫画分页与提示词已就绪', '漫画图片生成', '漫画合集']);
  assert.equal(timeline.find(item => item.key === 'video').label, '漫画图片生成');
  assert.equal(batchFactoryBookState(book, { productionMode: 'comic' }).detail, '漫画分页与提示词已就绪，等待生成漫画图片');
});

test('a video batch keeps the video timeline even when books carry legacy comic signals', () => {
  // 批次切回视频生产后，书级遗留信号（旧漫画 patch、漫画资产、comicPrompt）
  // 不允许把列表口径翻回漫画：统计、时间线都必须按视频语义显示。
  const book = {
    id: 'legacy-comic-in-video-batch',
    sourceText: '正文',
    settingsState: { patch: { productionMode: 'comic', aiPromptConfig: { comic: { enabled: true, body: '漫画提示词' } } } },
    assetRecords: [{ id: 'comic-asset', source: 'comic' }],
    videos: [{ id: 'comic-page-1', comicPrompt: '画面提示词', comicImagePrompt: '出图提示词' }]
  };
  const timeline = batchFactoryBookTimeline(book, { productionMode: 'video' });
  assert.deepEqual(timeline.map(item => item.label), ['原文已就绪', '资产（人物/场景/图片）已就绪', '文本提示词已就绪', '视频生成', '上传网络', '完成']);
  const state = batchFactoryBookState(book, { productionMode: 'video' });
  assert.equal(state.label, '待开始');
  assert.equal(state.detail, '原文已就绪');
});

test('counts only successfully uploaded books as batch completion', () => {
  const books = [
    { id: 'uploaded', sourceText: '正文', sourceMetadata: { publishStatus: 'uploaded' } },
    { id: 'failed', sourceText: '正文', sourceMetadata: { publishStatus: 'failed', publishError: 'network rejected' } },
    { id: 'running', sourceText: '正文', sourceMetadata: { publishStatus: 'uploading' } },
    { id: 'awaiting-upload', sourceText: '正文' },
    { id: 'pending', sourceText: '正文' }
  ];
  const runtime = {
    mergeStatus: { jobs: [{ bookId: 'awaiting-upload', status: 'succeeded', outputUrl: 'https://media.example/final.mp4' }] }
  };

  assert.deepEqual(batchFactoryBatchProgress(books, runtime), {
    total: 5,
    uploaded: 1,
    failed: 1,
    running: 1,
    awaitingUpload: 1,
    pending: 1,
    completionPercent: 20
  });
});


test('maps active production tasks to their storyboard cards instead of exposing provider video IDs', () => {
  const book = {
    id: 'book-1',
    videos: [{ id: 'video-1' }, { id: 'video-2' }, { id: 'video-3' }]
  };
  const progress = batchFactoryVideoProgress(book, {
    jobs: [{ bookId: 'book-1', directorRevisionId: 'director-current', tasks: [
      { id: 'task-1', videoId: 'video-1', status: 'succeeded', updatedAt: '2026-09-16T00:00:01Z' },
      { id: 'task-2', videoId: 'video-2', status: 'running', updatedAt: '2026-09-16T00:00:02Z' },
      { id: 'task-3', videoId: 'video-3', status: 'queued', updatedAt: '2026-09-16T00:00:03Z' }
    ] }]
  });

  assert.equal(progress.byVideo.get('video-2').message, '正在生成分镜 2…');
  assert.equal(progress.byVideo.get('video-3').message, '分镜 3 已排队，等待生成…');
  assert.equal(progress.summary, '正在生成分镜 2；分镜 3 已排队');
});

test('ignores stale tasks from an older director revision when current storyboard tasks exist', () => {
  const book = { id: 'book-1', directorRevision: { id: 'director-current' }, videos: [{ id: 'video-1' }] };
  const progress = batchFactoryVideoProgress(book, {
    jobs: [
      { bookId: 'book-1', directorRevisionId: 'director-old', tasks: [{ videoId: 'video-1', status: 'running' }] },
      { bookId: 'book-1', directorRevisionId: 'director-current', tasks: [{ videoId: 'video-1', status: 'succeeded' }] }
    ]
  });

  assert.equal(progress.active.length, 0);
  assert.equal(progress.byVideo.get('video-1').status, 'succeeded');
});


test('shows a durable 121 readback state in the novel list after upload', () => {
  const book = {
    id: 'book-1',
    bookId: '2080310440299710279',
    title: '已上传小说',
    sourceText: '正文',
    sourceMetadata: {
      websiteSubmitStatus: 'uploaded',
      websiteSubmitReceipt: { status: 'confirmed', remoteId: '212960' }
    }
  };
  const row = batchFactoryNovelTableRow(book, 0);
  assert.equal(row.websiteSubmit, '已回读');
  assert.equal(batchFactoryBookState(book).label, '完成');
});

test('a later successful retry clears an earlier failed task for the same storyboard', () => {
  const book = {
    id: 'book-1',
    sourceText: '正文',
    directorRevision: { id: 'director-current' },
    videos: [{ id: 'video-1' }, { id: 'video-2' }]
  };
  const runtime = {
    productionStatus: {
      jobs: [
        { bookId: 'book-1', directorRevisionId: 'director-current', tasks: [
          { videoId: 'video-1', status: 'succeeded', updatedAt: '2026-09-16T11:58:01Z' },
          { videoId: 'video-2', status: 'failed', updatedAt: '2026-09-16T12:43:20Z', errorMessage: 'old provider failure' }
        ] },
        { bookId: 'book-1', directorRevisionId: 'director-current', tasks: [
          { videoId: 'video-2', status: 'succeeded', updatedAt: '2026-09-16T13:49:43Z' }
        ] }
      ]
    }
  };
  const video = batchFactoryBookTimeline(book, runtime).find(item => item.key === 'video');
  assert.equal(video.status, 'completed');
  assert.equal(batchFactoryBookState(book, runtime).label, '待合成');
});

test('an uploaded book stays completed even when a redundant retry leaves one storyboard failed', () => {
  const book = {
    id: 'book-1',
    sourceText: '正文',
    directorRevision: { id: 'director-current' },
    videos: [{ id: 'video-1' }, { id: 'video-2' }, { id: 'video-3' }],
    sourceMetadata: { websiteSubmitStatus: 'uploaded', publishStatus: 'uploaded' }
  };
  const runtime = {
    productionStatus: {
      jobs: [
        { bookId: 'book-1', directorRevisionId: 'director-current', tasks: [
          { videoId: 'video-1', status: 'succeeded', updatedAt: '2026-09-28T15:04:22Z' },
          { videoId: 'video-2', status: 'succeeded', updatedAt: '2026-09-28T15:04:32Z' },
          { videoId: 'video-3', status: 'succeeded', updatedAt: '2026-09-28T15:04:35Z' }
        ] },
        { bookId: 'book-1', directorRevisionId: 'director-current', tasks: [
          { videoId: 'video-1', status: 'failed', updatedAt: '2026-09-28T15:11:01Z', errorMessage: 'redundant retry failed' },
          { videoId: 'video-2', status: 'succeeded', updatedAt: '2026-09-28T15:04:11Z' },
          { videoId: 'video-3', status: 'succeeded', updatedAt: '2026-09-28T15:06:23Z' }
        ] }
      ]
    }
  };

  const state = batchFactoryBookState(book, runtime);
  assert.equal(state.label, '完成');
  assert.equal(batchFactoryBatchProgress([book], runtime).failed, 0);
  assert.equal(batchFactoryBatchProgress([book], runtime).uploaded, 1);
});
