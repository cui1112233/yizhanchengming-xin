import test from 'node:test';
import assert from 'node:assert/strict';
import { shouldAutoClassifyBook, automationBookIsActive, automationBookNeedsRetry, batchFactoryAutomationRetryAction, batchFactoryBatchProgress, batchFactoryBookState, batchFactoryBookTimeline, batchFactoryNovelTableRow, batchFactoryVideoProgress, batchFactoryVisibleError } from './batchFactoryBookState.js';

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
  assert.deepEqual({ label: state.label, detail: state.detail }, { label: '待上传', detail: '视频已合并，等待 121 提交' });
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
  assert.equal(timeline.find(item => item.key === 'complete')?.detail, '已整理 1 个合集');
  assert.equal(batchFactoryBookState(book, { productionMode: 'comic' }).label, '完成');
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
