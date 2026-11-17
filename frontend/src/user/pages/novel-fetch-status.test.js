import test from 'node:test';
import assert from 'node:assert/strict';
import * as statusHelpers from './novel-fetch-status.js';

test('scheduled books with preserved prose still display waiting for scheduled execution', () => {
  const state = statusHelpers.novelFetchTaskStatus({status: 'scheduled_waiting', originalStatus: 'done', aiStatus: 'done'});
  assert.equal(state.status, 'scheduled_waiting');
  assert.equal(state.label, '等待定时执行');
});

test('queued retry acknowledgement does not claim processing has finished', () => {
  const text = statusHelpers.novelFetchBatchResultSummary({ queued: 2, failed: 1, skipped: 1, already_running: 1 });
  assert.match(text, /加入后台重试 2 项/);
  assert.match(text, /失败 1 项/);
  assert.doesNotMatch(text, /已处理|已完成/);
});

test('raw original file is not an AI1 result and failed version metadata is not content', () => {
  assert.equal(statusHelpers.novelFetchVersionReady({ aiFiles: ['original_raw'], aiStatus: 'failed' }), false);
  assert.equal(statusHelpers.novelFetchVersionReady({ versions: { ai1: { state: 'failed' } }, aiStatus: 'failed' }), false);
  assert.equal(statusHelpers.novelFetchVersionReady({ versions: { ai1: false }, aiStatus: 'failed' }), false);
  assert.equal(statusHelpers.novelFetchVersionReady({ aiFiles: ['ai2'] }, 'ai1'), false);
  assert.equal(statusHelpers.novelFetchVersionReady({ aiFiles: ['ai2'] }, 'ai2'), true);
  assert.equal(statusHelpers.novelFetchVersionReady({ versions: { ai1: { state: 'ready' } } }), true);
});

test('backend effective status wins over conflicting legacy stage fields', () => {
  const state = statusHelpers.novelFetchTaskStatus({
    effective_status: 'processing',
    status: 'completed',
    aiStatus: 'failed'
  });
  assert.equal(state.status, 'processing');
  assert.equal(state.label, '正在执行中…');
});

test('upload-only failure keeps completed body status and actionable next step', () => {
  assert.equal(statusHelpers.novelFetchTaskStatus({ status: 'upload_pending', originalStatus: 'done', aiStatus: 'done', siteSubmitStatus: 'failed' }).label, '正文已完成，待重试上传');
});

const {
  beijingDateKey, currentBatchBookIds, filterNovelFetchTasks, novelFetchCharCount,
  novelFetchStageLabel, novelFetchTaskDateKey, novelFetchTaskStatus,
  retainVisibleTaskIds, stoppableCurrentBatchBookIds
} = statusHelpers;
const shouldPollNovelFetchRuntime = statusHelpers.shouldPollNovelFetchRuntime || (() => false);
const novelFetchCurrentBatchTaskDetail = statusHelpers.novelFetchCurrentBatchTaskDetail || (() => null);
const novelFetchQueueSummary = statusHelpers.novelFetchQueueSummary || (() => ({ label: '', running: 0, queued: 0, waitingRetry: 0, updatedAt: '' }));
const novelFetchOriginalFetchRetry = statusHelpers.novelFetchOriginalFetchRetry || (() => ({}));
const novelFetchSubmissionLabel = statusHelpers.novelFetchSubmissionLabel || (() => '未提交');
const novelFetchPollingPlan = statusHelpers.novelFetchPollingPlan || (() => ({}));
const novelFetchVersionReady = statusHelpers.novelFetchVersionReady || (() => false);
const mergeNovelFetchQueueStatus = statusHelpers.mergeNovelFetchQueueStatus;
const novelFetchBatchResultSummary = statusHelpers.novelFetchBatchResultSummary || (() => '');

test('batch action summary distinguishes actual updates from skipped tasks', () => {
  assert.equal(
    novelFetchBatchResultSummary({ requested: 3, updated: 1, skipped: 1, failed: 1 }),
    '请求 3 项，已处理 1 项，跳过 1 项，失败 1 项'
  );
});

test('process and task pages poll for scheduled jobs before a current batch exists', () => {
  assert.equal(shouldPollNovelFetchRuntime({ activeTab: 'process', processing: false, currentBatchCount: 0 }), true);
  assert.equal(shouldPollNovelFetchRuntime({ activeTab: 'tasks', processing: false, currentBatchCount: 0 }), true);
  assert.equal(shouldPollNovelFetchRuntime({ activeTab: 'knowledge', processing: false, currentBatchCount: 0 }), false);
  assert.equal(shouldPollNovelFetchRuntime({ activeTab: 'knowledge', processing: true, currentBatchCount: 0 }), true);
});

test('task page separates realtime, queue, and full task refresh intervals', () => {
  assert.deepEqual(
    novelFetchPollingPlan({ activeTab: 'tasks', documentHidden: false }),
    { realtimeMs: 2000, queueMs: 5000, tasksMs: 10000 }
  );
  assert.deepEqual(
    novelFetchPollingPlan({ activeTab: 'knowledge', documentHidden: false }),
    { realtimeMs: 0, queueMs: 0, tasksMs: 0 }
  );
});

test('hidden pages pause runtime polling and failures slow the affected request', () => {
  assert.deepEqual(
    novelFetchPollingPlan({ activeTab: 'tasks', documentHidden: true }),
    { realtimeMs: 0, queueMs: 0, tasksMs: 0 }
  );
  assert.equal(novelFetchPollingPlan({ activeTab: 'tasks', taskFailures: 1 }).tasksMs, 20000);
  assert.equal(novelFetchPollingPlan({ activeTab: 'tasks', taskFailures: 2 }).tasksMs, 30000);
  assert.equal(novelFetchPollingPlan({ activeTab: 'tasks', realtimeFailures: 1 }).realtimeMs, 5000);
  assert.equal(novelFetchPollingPlan({ activeTab: 'tasks', realtimeFailures: 2 }).realtimeMs, 15000);
});

test('current batch view opens one task detail without applying task-page filters', () => {
  assert.deepEqual(novelFetchCurrentBatchTaskDetail({ book_id: 'book-1' }), { bookId: 'book-1', initialTab: 'summary' });
  assert.equal(novelFetchCurrentBatchTaskDetail({}), null);
});

test('realtime queue counts are presented as one readable task summary', () => {
  assert.deepEqual(novelFetchQueueSummary({
    counts: { running: 2, queued: 3, waiting_retry: 1 },
    updated_at: '2026-10-18T12:00:00Z'
  }), {
    label: '队列运行中',
    enabled: false,
    activeJobId: '',
    latestJob: {},
    running: 2,
    queued: 3,
    waitingRetry: 1,
    scopeDate: '',
    scopeBatchId: '',
    scopeLabel: '全部任务',
    matchedDocuments: 0,
    updatedAt: '2026-10-18T12:00:00Z'
  });
  assert.equal(novelFetchQueueSummary({ counts: {} }).label, '队列空闲');
});

test('realtime queue summary states which batch or date the counts belong to', () => {
  const scoped = novelFetchQueueSummary({
    counts: { running: 1, queued: 2 },
    scope: { batch_id: 'batch-1758', current_batch_id: 'batch-1758', matched_documents: 12 }
  });
  assert.equal(scoped.scopeLabel, '当前批次');
  assert.equal(scoped.scopeBatchId, 'batch-1758');
  assert.equal(scoped.matchedDocuments, 12);
  assert.equal(novelFetchQueueSummary({ counts: {}, scope: { date: '2026-09-24' } }).scopeLabel, '2026-09-24');
});

test('realtime queue summary exposes full automation state', () => {
  const summary = novelFetchQueueSummary({
    auto_queue_enabled: true,
    active_job_id: 'queue_1792315781954_1d8ed3fc',
    pending_count: 4,
    processing_count: 2,
    latest_job: { id: 'queue_previous', status: 'done' },
    updated_at: '2026-10-18T19:50:24+08:00'
  });
  assert.equal(summary.enabled, true);
  assert.equal(summary.activeJobId, 'queue_1792315781954_1d8ed3fc');
  assert.equal(summary.running, 2);
  assert.equal(summary.queued, 4);
  assert.equal(summary.latestJob.id, 'queue_previous');
  assert.equal(summary.latestJob.status, 'done');
});

test('queue polling failures keep the last confirmed automation switch state', () => {
  const previous = { enabled: true, auto_queue_enabled: true, counts: { running: 0 } };
  assert.equal(mergeNovelFetchQueueStatus(previous, { counts: {} }).enabled, true);
  assert.equal(mergeNovelFetchQueueStatus(previous, { enabled: false, auto_queue_enabled: false }).enabled, false);
});

test('original fetch retry exposes the next automatic attempt and its reason', () => {
  assert.deepEqual(novelFetchOriginalFetchRetry({
    original_fetch_retry_count: 2,
    original_fetch_next_retry_at: '2026-10-18T21:00:00+08:00',
    original_fetch_last_error: '上游暂时没有正文'
  }), {
    attempts: 2,
    nextRetryAt: '2026-10-18T21:00:00+08:00',
    error: '上游暂时没有正文'
  });
});

test('submission status stays empty until an AI version is ready to submit', () => {
  assert.equal(novelFetchSubmissionLabel({ ai_status: 'waiting_ai_config' }), '—');
  assert.equal(novelFetchSubmissionLabel({ ai_status: 'ai_done', ai_generated_count: 1 }), '待提交');
  assert.equal(novelFetchSubmissionLabel({ ai_status: 'ai_done', site_submit_status: 'submitted' }), '已提交');
});

test('submission status explains version-level blockers from the legacy contract', () => {
  assert.equal(novelFetchSubmissionLabel({ site_submit_status: 'unknown', site_submit_expected_versions: ['ai1', 'ai2'], site_submit_pending_versions: ['ai2'] }), '待远端核验 · 待提交 AI2');
  assert.equal(novelFetchSubmissionLabel({ site_submit_status: 'partial', site_submit_partial_versions: ['ai1'], site_submit_failed_versions: ['ai2'] }), '部分接收 · 失败 AI2');
  assert.equal(novelFetchSubmissionLabel({ site_submit_status: 'waiting_generate', site_submit_waiting_generate_versions: ['ai1'] }), '待生成 AI1');
});

test('current batch marks AI1 complete only when AI1 content is actually available', () => {
  assert.equal(novelFetchVersionReady({ ai_status: 'waiting_ai_config' }, 'ai1'), false);
  assert.equal(novelFetchVersionReady({ ai_status: 'ai_done' }, 'ai1'), true);
  assert.equal(novelFetchVersionReady({ ai_generated_count: 1 }, 'ai1'), true);
});

test('completed Go task status is shown as completed', () => {
  assert.deepEqual(novelFetchTaskStatus({ status: 'completed' }), { status: 'completed', label: '完成', color: 'green' });
});

test('processed but unsubmitted task status is shown as pending submit', () => {
  assert.deepEqual(novelFetchTaskStatus({
    status: 'completed', original_status: 'done', ai_status: 'ai_done'
  }), { status: 'pending_submit', label: '处理完成，待提交', color: 'cyan' });
});

test('legacy submitted AI task falls back to completed', () => {
  assert.deepEqual(novelFetchTaskStatus({ original_status: 'done', ai_status: 'ai_done', site_submit_status: 'submitted' }), { status: 'completed', label: '完成', color: 'green' });
});

test('processing and input-ready use confirmed Chinese labels', () => {
  assert.equal(novelFetchTaskStatus({ status: 'processing' }).label, '正在执行中…');
  assert.equal(novelFetchTaskStatus({ status: 'input_ready' }).label, '分类信息已就绪');
});

test('sensitive AI wait is shown separately from active processing', () => {
  assert.equal(novelFetchTaskStatus({ original_status: 'waiting_sensitive_ai' }).status, 'waiting_sensitive_ai');
  assert.equal(novelFetchTaskStatus({ original_status: 'waiting_sensitive_ai' }).label, '等待敏感词 AI');
  assert.equal(novelFetchTaskStatus({ status: 'waiting_retry' }).status, 'waiting_retry');
});

test('classification wait is shown as a Chinese task state', () => {
  assert.deepEqual(novelFetchTaskStatus({ status: 'waiting_classification' }), {
    status: 'waiting_classification', label: '等待确认分类', color: 'orange'
  });
  assert.equal(novelFetchStageLabel('waiting_classification'), '等待确认分类');
});

test('processed and original counts are displayed together', () => {
  assert.equal(novelFetchCharCount({ processed_chars: 4000, original_raw_chars: 27831 }), '4000/27831');
  assert.equal(novelFetchCharCount({}), '-');
});

test('task stage values are presented as Chinese workflow labels', () => {
  assert.equal(novelFetchStageLabel('done'), '已完成');
  assert.equal(novelFetchStageLabel('ai_done'), '已生成');
  assert.equal(novelFetchStageLabel('pending_upload'), '待提交');
});

test('Beijing date does not fall back to the previous UTC day', () => {
  assert.equal(beijingDateKey(new Date('2026-10-16T16:30:00.000Z')), '2026-10-17');
});

test('task timestamps are filtered by their Beijing calendar date', () => {
  assert.equal(novelFetchTaskDateKey({ createdAt: '2026-10-17T21:00:10.8503469Z' }), '2026-10-18');
  assert.equal(novelFetchTaskDateKey({ updated_at: '2026-10-18T00:30:00+08:00' }), '2026-10-18');
});

test('hidden task selections are removed after date filtering', () => {
  assert.deepEqual(retainVisibleTaskIds(['book-a', 'old-book'], [{ bookId: 'book-a' }, { book_id: 'book-b' }]), ['book-a']);
  assert.deepEqual(retainVisibleTaskIds(['old-book'], []), []);
});

test('batch actions resolve only current batch book ids', () => {
  assert.deepEqual(currentBatchBookIds({
    taskIds: ['book-a', 'book-b', 'book-a'],
    taskStates: [{ bookId: 'historical-book' }]
  }), ['book-a', 'book-b']);
  assert.deepEqual(currentBatchBookIds({ task_states: [{ book_id: 'book-c' }] }), ['book-c']);
});

test('stop action excludes completed and pending-submit tasks', () => {
  assert.deepEqual(stoppableCurrentBatchBookIds({ taskStates: [
    { bookId: 'queued', status: 'queued' },
    { bookId: 'running', status: 'processing' },
    { bookId: 'ready', status: 'pending_submit' },
    { bookId: 'done', status: 'completed' }
  ] }), ['queued', 'running']);
});

test('default task view keeps recent completed tasks and all unfinished tasks', () => {
  const rows = [
    { book_id: 'recent-ok', status: 'completed', created_at: '2026-10-17T08:00:00+08:00' },
    { book_id: 'old-ok', status: 'completed', created_at: '2026-10-10T08:00:00+08:00' },
    { book_id: 'old-failed', status: 'failed', created_at: '2026-09-01T08:00:00+08:00' },
    { book_id: 'old-running', status: 'processing', created_at: '2026-09-02T08:00:00+08:00' }
  ];

  assert.deepEqual(
    filterNovelFetchTasks(rows, { mode: 'default', now: new Date('2026-10-18T12:00:00+08:00') }).map(row => row.book_id),
    ['recent-ok', 'old-failed', 'old-running']
  );
});

test('default task view keeps unfinished books from the current batch only', () => {
  const rows = [
    { book_id: 'current-failed', status: 'failed', created_at: '2026-09-01T08:00:00+08:00' },
    { book_id: 'current-running', status: 'processing', created_at: '2026-09-02T08:00:00+08:00' },
    { book_id: 'history-queued', status: 'queued', created_at: '2026-09-03T08:00:00+08:00' },
    { book_id: 'recent-ok', status: 'completed', created_at: '2026-10-17T08:00:00+08:00' }
  ];

  assert.deepEqual(
    filterNovelFetchTasks(rows, {
      mode: 'default',
      batchIds: ['current-failed', 'current-running'],
      now: new Date('2026-10-18T12:00:00+08:00')
    }).map(row => row.book_id),
    ['current-failed', 'current-running', 'recent-ok']
  );
});

test('task filtering combines an exact date with book id or title search', () => {
  const rows = [
    { book_id: '10001', book_name: '第一本书', status: 'completed', created_at: '2026-10-18T08:00:00+08:00' },
    { book_id: '20002', book_name: '目标小说', status: 'failed', created_at: '2026-10-18T09:00:00+08:00' },
    { book_id: '30003', book_name: '目标旧书', status: 'failed', created_at: '2026-10-17T09:00:00+08:00' }
  ];

  assert.deepEqual(
    filterNovelFetchTasks(rows, { mode: 'date', date: '2026-10-18', query: '目标' }).map(row => row.book_id),
    ['20002']
  );
  assert.deepEqual(
    filterNovelFetchTasks(rows, { mode: 'date', date: '2026-10-18', query: '10001' }).map(row => row.book_id),
    ['10001']
  );
});
