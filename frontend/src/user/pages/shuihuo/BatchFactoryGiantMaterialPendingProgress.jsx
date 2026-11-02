import { useEffect, useRef, useState } from 'react';
import { Alert, Button, message, Progress, Space, Tooltip } from 'antd';
import { InfoCircleOutlined } from '@ant-design/icons';
import { classifyBookPublishMetadata, fetchBookOriginal, getBatch, getBatchAutomationStatus, startBatchAutomation, updateBookMetadata, updateBookSource } from '../../../shared/api/batchFactory.js';
import { findRegisteredGiantMaterialBook } from './batchFactoryGiantMaterialImport.js';
import { cancelGiantMaterialJob, createGiantMaterialJob, getGiantMaterialJob, listGiantMaterialExecutors } from '../../../shared/api/giantMaterialExecutorPublic.js';
import { giantMaterialOcrStatus } from './giantMaterialOcrStatus.js';

// 旧批量在浏览器页面里创建，可能已经留下“direct_first + 待读取”的占位书。
// 同一页面只允许一条直取请求在飞行，避免一次打开 20 本书把书城接口打爆。
let directFirstReadTail = Promise.resolve();
const giantAutomationStartClaims = new Set();
function enqueueDirectFirstRead(task) {
  const next = directFirstReadTail.then(task, task);
  directFirstReadTail = next.catch(() => undefined);
  return next;
}

async function startSavedGiantAutomation(batchId, plan) {
  if (!batchId || !plan?.presetId) return { started: false };
  // 同批多本书可能在相邻秒数里完成 OCR；只允许其中一条回填继续启动生产。
  // 生产控制器的 start 会替换已有任务，因此这里既要本页去重，也要先看服务端真实状态。
  if (giantAutomationStartClaims.has(batchId)) return { started: false };
  giantAutomationStartClaims.add(batchId);
  try {
    const statusResponse = await getBatchAutomationStatus(batchId);
    const automation = statusResponse?.automation || statusResponse?.data?.automation || statusResponse?.data || statusResponse || {};
    if (['running', 'scheduled', 'paused'].includes(String(automation.state || '').toLowerCase())) return { started: false };
    await startBatchAutomation(batchId, {
      scheduledAt: plan.scheduledAt || '',
      scheduleMode: plan.scheduleMode || 'execute',
      presetId: plan.presetId,
      runMode: plan.runMode || 'video_no_submit',
      autoPublish: plan.autoPublish === true,
      concurrency: plan.concurrency
    });
    return { started: true, scheduled: Boolean(plan.scheduledAt) };
  } catch (error) {
    giantAutomationStartClaims.delete(batchId);
    throw error;
  }
}

// 占位书正文进度条：巨量素材登记后正文为空，Windows 执行器读取期间在书卡“小说正文”位置
// 显示实时进度（视频 x/y 秒 + 百分比）。读取完成后在这里直接回填正文并刷新，弹窗关了也不丢。
// 读取失败或未绑定任务时提供“原文获取”：用书卡已有的书名 + Book ID + 书城直接拉正文兜底。
// variant='banner'：小说列表弹窗顶部的通栏提示条，一行展示，不占表格格子。
export function BatchFactoryGiantMaterialPendingProgress({ book, batchId, onContentReady, visible = true, variant = 'inline' }) {
  const metadata = book?.sourceMetadata || {};
  const jobId = String(metadata.executorJobId || '').trim();
  const pending = Boolean(metadata.contentPending);
  const originalReadStrategy = String(metadata.originalReadStrategy || 'ocr_first');
  const persistedOriginalReadError = String(metadata.originalReadError || '').trim();
  const [progress, setProgress] = useState(null);
  const [jobState, setJobState] = useState(null);
  const [error, setError] = useState('');
  const [refetching, setRefetching] = useState(false);
  const [ocrStarting, setOcrStarting] = useState(false);
  const doneRef = useRef(false);
  const fallbackRef = useRef('');
  const autoStartRef = useRef('');

  useEffect(() => {
    doneRef.current = false;
    setError('');
    setProgress(null);
    setJobState(null);
    if (!pending || !jobId || !batchId) return undefined;
    let stopped = false;
    const tick = async () => {
      try {
        const response = await getGiantMaterialJob(jobId);
        const job = response?.job || response?.data?.job || response;
        if (stopped || !job) return;
        setJobState(job);
        const state = String(job.state || '').toLowerCase();
        if (state === 'queued' && job.waitingForExecutor === true) {
          setProgress(job.progress || null);
          return;
        }
        if (state === 'succeeded') {
          setProgress({ completed: 100, total: 100, percent: 100 });
          if (!doneRef.current) {
            doneRef.current = true;
            // OCR 正文由服务端自动回填并继续调度；这里仅展示执行器状态。
            // 不能再让“某个浏览器标签页是否开着”决定正文和自动化是否落库。
            onContentReady?.();
          }
          return;
        }
        if (state === 'failed' || state === 'cancelled') {
          const failure = state === 'cancelled' ? '正文读取已取消。' : '正文读取失败。';
          if (originalReadStrategy === 'ocr_first' && fallbackRef.current !== jobId) {
            fallbackRef.current = jobId;
            setError(`${failure} 正在自动改用书城获取正文…`);
            refetchOriginal(true);
            return;
          }
          setError(failure);
          return;
        }
        setProgress(job.progress || null);
      } catch (_) { /* 网络抖动下一轮再试 */ }
    };
    tick();
    const timer = setInterval(tick, 5000);
    return () => { stopped = true; clearInterval(timer); };
  }, [jobId, pending, batchId, originalReadStrategy]);

  async function fetchBookOriginalWithRetry(batchId, bookId) {
    try {
      return await fetchBookOriginal(batchId, bookId);
    } catch (error) {
      if (!/failed to fetch|networkerror|网络/i.test(String(error?.message || ''))) throw error;
      return fetchBookOriginal(batchId, bookId);
    }
  }

  async function refetchOriginal(automatic = false) {
    if (!batchId || !book?.id || refetching) return;
    setRefetching(true);
    try {
      const latestResponse = await getBatch(batchId);
      const latestBatch = latestResponse?.batch || latestResponse?.data?.batch || latestResponse?.data || latestResponse;
      const latestBook = findRegisteredGiantMaterialBook(latestBatch?.books, metadata.giantMaterialId) || (latestBatch?.books || []).find(item => item?.id === book.id);
      if (!latestBook?.id) throw new Error('GIANT_MATERIAL_BOOK_NOT_FOUND');
      if (String(latestBook.sourceText || '').trim()) {
        await updateBookMetadata(batchId, latestBook.id, {
          metadata: { ...(latestBook.sourceMetadata || {}), contentPending: false, originalReadStage: 'direct', originalReadError: '' },
          expectedRevision: Number(latestBook.revision || 0)
        });
        try { await classifyBookPublishMetadata(batchId, latestBook.id, { force: true }); } catch (_) { /* 分类可在工作区单独重试 */ }
        const existingPlan = latestBook.sourceMetadata?.giantAutomationPlan;
        if (existingPlan?.presetId) {
          try { await startSavedGiantAutomation(batchId, existingPlan); } catch (_) { /* 自动生产可在工作区重试 */ }
        }
        onContentReady?.();
        return;
      }
      if (jobId && jobState?.state === 'queued' && jobState?.waitingForExecutor === true) {
        try { await cancelGiantMaterialJob(jobId); } catch (_) { /* A claimed job is allowed to finish independently. */ }
      }
      await updateBookMetadata(batchId, latestBook.id, {
        metadata: { ...(latestBook.sourceMetadata || {}), executorJobId: '', originalReadStage: 'direct', originalReadError: '', contentPending: true },
        expectedRevision: Number(latestBook.revision || 0)
      });
      await fetchBookOriginalWithRetry(batchId, latestBook.id);
      const refreshedResponse = await getBatch(batchId);
      const refreshedBatch = refreshedResponse?.batch || refreshedResponse?.data?.batch || refreshedResponse?.data || refreshedResponse;
      const refreshedBook = findRegisteredGiantMaterialBook(refreshedBatch?.books, metadata.giantMaterialId) || (refreshedBatch?.books || []).find(item => item?.id === book.id);
      if (refreshedBook?.id) {
        await updateBookMetadata(batchId, refreshedBook.id, {
          metadata: { ...(refreshedBook.sourceMetadata || {}), contentPending: false, originalReadStage: 'direct', originalReadError: '' },
          expectedRevision: Number(refreshedBook.revision || 0)
        });
      }
      try { await classifyBookPublishMetadata(batchId, latestBook.id, { force: true }); } catch (_) { /* 分类失败不阻断生产，可单书重试 */ }
      const plan = latestBook.sourceMetadata?.giantAutomationPlan;
      if (plan?.presetId) {
        try {
          await startSavedGiantAutomation(batchId, plan);
        } catch (automationError) {
          message.warning(`正文已回填；自动生产请在工作区重试：${automationError?.message || '启动失败'}`);
        }
      }
      message.success('已通过书城获取正文');
      onContentReady?.();
    } catch (fetchError) {
      const failure = String(fetchError?.message || '获取正文失败');
      try {
        const refreshedResponse = await getBatch(batchId);
        const refreshedBatch = refreshedResponse?.batch || refreshedResponse?.data?.batch || refreshedResponse?.data || refreshedResponse;
        const refreshedBook = findRegisteredGiantMaterialBook(refreshedBatch?.books, metadata.giantMaterialId) || (refreshedBatch?.books || []).find(item => item?.id === book.id);
        if (refreshedBook?.id) {
          await updateBookMetadata(batchId, refreshedBook.id, {
            metadata: { ...(refreshedBook.sourceMetadata || {}), originalReadStage: 'failed', originalReadError: failure, contentPending: false },
            expectedRevision: Number(refreshedBook.revision || 0)
          });
        }
      } catch (_) { /* 原错误优先展示；用户仍可手动重试 */ }
      setError(failure);
      if (!automatic) message.error(failure);
    } finally {
      setRefetching(false);
    }
  }

  async function startOcrFallback() {
    if (!batchId || !book?.id) return;
    setOcrStarting(true);
    try {
      // 读取最新 revision，避免“直接获取原文”失败后再次派发 OCR 时覆盖其它书卡更新。
      const latestResponse = await getBatch(batchId);
      const latestBatch = latestResponse?.batch || latestResponse?.data?.batch || latestResponse?.data || latestResponse;
      const latestBook = findRegisteredGiantMaterialBook(latestBatch?.books, metadata.giantMaterialId) || (latestBatch?.books || []).find(item => item?.id === book.id);
      if (!latestBook?.id) throw new Error('GIANT_MATERIAL_BOOK_NOT_FOUND');
      if (String(latestBook.sourceText || '').trim()) {
        onContentReady?.();
        return;
      }
      const latestMetadata = latestBook.sourceMetadata || metadata;
      const executorsResponse = await listGiantMaterialExecutors();
      const executors = Array.isArray(executorsResponse?.executors)
        ? executorsResponse.executors
        : Array.isArray(executorsResponse?.data?.executors) ? executorsResponse.data.executors : [];
      if (!executors.some(executor => executor?.online === true)) {
        throw new Error('OCR 执行器当前未在线，未创建队列任务；请启动执行器后重试。');
      }
      const durationSeconds = Number(latestMetadata.videoDurationSeconds || 0);
      if (!(durationSeconds > 0)) throw new Error('素材缺少有效视频时长，无法改用滚屏 OCR。');
      const job = await createGiantMaterialJob({ materialId: latestMetadata.giantMaterialId, platformBookId: latestMetadata.platformBookId, title: latestMetadata.sourceBookTitle || latestBook.title, videoUrl: latestMetadata.videoUrl, durationSeconds, contentRangeLines: latestMetadata.contentRangeLines });
      const created = job?.job || job?.data?.job || job;
      if (!created?.id) throw new Error('GIANT_EXECUTOR_FAILED');
      await updateBookMetadata(batchId, latestBook.id, {
        metadata: { ...latestMetadata, executorJobId: created.id, originalReadStage: 'ocr', originalReadError: '', contentPending: true },
        expectedRevision: Number(latestBook.revision || 0)
      });
      setError(''); onContentReady?.();
    } catch (startError) { setError(String(startError?.message || '等待执行器')); } finally { setOcrStarting(false); }
  }

  useEffect(() => {
    const key = String(book?.id || '');
    const shouldResume = Boolean(
      visible
      && pending
      && batchId
      && book?.id
      && !jobId
      && !String(book?.sourceText || '').trim()
      && originalReadStrategy === 'direct_first'
    );
    if (!shouldResume || autoStartRef.current === key) return;
    autoStartRef.current = key;
    // Retry direct-first books once per visible book mount, including records that failed before this page opened.
    enqueueDirectFirstRead(() => refetchOriginal(true));
  }, [visible, pending, batchId, book?.id, book?.sourceText, jobId, originalReadStrategy]);

  if (!pending || !visible) return null;
  // 直取优先时，任务尚未绑定 OCR 并不代表失败；这里正在自动请求书城正文。
  // 只在入厂内容格里显示状态，避免顶部批量列表出现一串重复按钮。
  if (!jobId && originalReadStrategy === 'direct_first' && !error && !persistedOriginalReadError) {
    return <span className="batch-factory-giant-pending-compact is-direct-reading" onClick={event => event.stopPropagation()}><i>书城获取中</i></span>;
  }
  if (error || persistedOriginalReadError || !jobId) {
    // 窄列（小说列表弹窗“内容”格）放不下整块 Alert：错误全文进悬浮提示，行内只留
    // 一行红字 + 横排两个恢复按钮，避免文字/按钮被挤断行。
    const detail = String(error || persistedOriginalReadError || 'OCR 任务未派发');
    return (
      <Tooltip title={detail}>
        <span className="batch-factory-giant-pending-compact" onClick={event => event.stopPropagation()}>
          <i className="is-error-text">{error || persistedOriginalReadError ? '读取失败' : '未派发OCR'}</i>
          <Space size={4} className="is-actions">
            <Button size="small" type="primary" loading={refetching} onClick={() => refetchOriginal(false)}>获取正文</Button>
            <Button size="small" loading={ocrStarting} disabled={refetching} onClick={startOcrFallback}>OCR</Button>
          </Space>
        </span>
      </Tooltip>
    );
  }
  if (originalReadStrategy === 'direct_first' && jobState?.state === 'queued' && jobState?.waitingForExecutor === true) {
    const detail = 'OCR 执行器当前未在线；可直接重试书城获取正文，也可启动执行器后再做 OCR。';
    return (
      <Tooltip title={detail}>
        <span className="batch-factory-giant-pending-compact" onClick={event => event.stopPropagation()}>
          <i className="is-error-text">OCR 执行器未在线</i>
          <Space size={4} className="is-actions">
            <Button size="small" type="primary" loading={refetching} disabled={ocrStarting} onClick={() => refetchOriginal(false)}>获取正文</Button>
            <Button size="small" loading={ocrStarting} disabled={refetching} onClick={startOcrFallback}>OCR</Button>
          </Space>
        </span>
      </Tooltip>
    );
  }
  const percent = Math.max(0, Math.min(100, Number(progress?.percent || 0)));
  const ocrStatus = giantMaterialOcrStatus({ ...jobState, progress });
  if (variant === 'banner') {
    return <div className="batch-factory-giant-pending-banner" onClick={event => event.stopPropagation()}>
      <span className="is-title"><InfoCircleOutlined />{ocrStatus.title} · {book?.title || '当前小说'}</span>
      <Progress percent={percent} size="small" status={ocrStatus.progressStatus} />
      <span className="is-detail">{ocrStatus.detail}</span>
    </div>;
  }
  return <Alert className="batch-factory-giant-pending" type="info" showIcon message={ocrStatus.title} description={<><Progress percent={percent} size="small" status={ocrStatus.progressStatus} /><span>{ocrStatus.detail}</span></>} onClick={event => event.stopPropagation()} />;
}
