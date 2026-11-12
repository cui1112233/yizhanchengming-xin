import { Alert, Button, Card, Collapse, Divider, Segmented, Select, Space, Tag, Typography, message } from 'antd';
import { Bug, Combine, Download, Play, RefreshCw, Sparkles } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { getBatchFactoryMergeCapability, getBatchFactoryProductionStatus, mergeBatchFactoryVideos } from '../../../shared/api/batchFactory';
import { downloadMedia, retryTask } from '../../../shared/api/shuihuoProduction';
import { reportClientError } from '../../../shared/error-reporting';

const ACTIVE_TASK_STATUSES = new Set(['draft', 'queued', 'running']);
const PREVIEW_SPEED_OPTIONS = [1, 1.1, 1.2, 1.3, 1.5, 1.7, 2];
const MERGE_SOURCE = 'batch_merge';
const BULK_MERGE_CONCURRENCY = 2;
const AI_ITEM_STATUSES = new Set(['queued_hook', 'hook_generating', 'queued_director', 'director_generating']);

const taskStatusMeta = {
  draft: ['待生成', 'default'],
  queued: ['排队中', 'processing'],
  running: ['生成中', 'processing'],
  succeeded: ['已完成', 'green'],
  failed: ['失败', 'red'],
  cancelled: ['失败', 'red']
};

const bookStatusMeta = [
  { key: 'all', label: '全部' },
  { key: 'pending', label: '待开始' },
  { key: 'review', label: '待审核', color: 'gold' },
  { key: 'ai_processing', label: 'AI处理中', color: 'processing' },
  { key: 'ready_generate', label: '待生成', color: 'blue' },
  { key: 'queued', label: '排队中', color: 'processing' },
  { key: 'generating', label: '视频生成中', color: 'processing' },
  { key: 'failed', label: '异常', color: 'red' },
  { key: 'ready_merge', label: '待合并', color: 'cyan' },
  { key: 'merged', label: '已合并', color: 'green' }
];

function normalizeProjectIds(batch) {
  const ids = (batch?.items || [])
    .map(item => Number(item.production?.projectId))
    .filter(id => Number.isInteger(id) && id > 0);
  return [...new Set(ids)].sort((a, b) => a - b);
}

async function runWithConcurrency(items, limit, worker, onSettled) {
  const results = new Array(items.length);
  let cursor = 0;

  async function runner() {
    while (cursor < items.length) {
      const index = cursor;
      cursor += 1;
      try {
        results[index] = { ok: true, value: await worker(items[index], index) };
      } catch (error) {
        results[index] = { ok: false, error };
      }
      onSettled?.(results[index], items[index], index);
    }
  }

  const workerCount = Math.min(Math.max(1, limit), items.length);
  await Promise.all(Array.from({ length: workerCount }, () => runner()));
  return results;
}

function reportBatchFactoryIssue(kind, { batch, item, video, production, error, source }) {
  reportClientError({
    kind,
    message: error?.message || production?.error || '批量工厂任务异常',
    source: source || '/batch-factory',
    context: {
      batchId: batch?.id || '',
      itemId: item?.id || '',
      bookTitle: item?.title || '',
      bookId: item?.bookId || '',
      videoId: video?.id ?? '',
      projectId: item?.production?.projectId || '',
      taskId: production?.task?.id || '',
      providerTaskId: production?.task?.providerTaskId || '',
      modelName: item?.production?.modelName || batch?.settings?.videoModelName || ''
    }
  });
}

export function useBatchFactoryProductionStatus(batch) {
  const projectIds = useMemo(() => normalizeProjectIds(batch), [batch]);
  const projectKey = projectIds.join(',');
  const [byProjectId, setByProjectId] = useState({});
  const [error, setError] = useState('');
  const [refreshVersion, setRefreshVersion] = useState(0);

  useEffect(() => {
    if (!projectIds.length) {
      setByProjectId({});
      setError('');
      return undefined;
    }

    let cancelled = false;
    let timer = null;

    async function refresh() {
      try {
        const result = await getBatchFactoryProductionStatus(projectIds);
        if (cancelled) return;
        const next = {};
        let hasActive = false;
        for (const project of result.projects || []) {
          next[String(project.projectId)] = project;
          if ((project.tasks || []).some(task => ACTIVE_TASK_STATUSES.has(task.status))) hasActive = true;
        }
        setByProjectId(next);
        setError('');
        if (hasActive) timer = window.setTimeout(refresh, 3000);
      } catch (requestError) {
        if (cancelled) return;
        setError(requestError.message || '读取视频生产状态失败');
        timer = window.setTimeout(refresh, 5000);
      }
    }

    refresh();
    return () => {
      cancelled = true;
      if (timer) window.clearTimeout(timer);
    };
  }, [projectKey, refreshVersion]);

  return {
    byProjectId,
    error,
    projectIds,
    refreshNow: () => setRefreshVersion(version => version + 1)
  };
}

export function resolveBatchFactoryVideoProduction(item, videoIndex, projectStatus) {
  if (!item?.production?.projectId) return null;
  const results = Array.isArray(item.productionResults) ? item.productionResults : [];
  const submission = results.find(result => Number(result.index) === videoIndex + 1) || results[videoIndex] || null;
  if (submission?.error && !submission?.task) {
    return { status: 'failed', error: submission.error, task: null, media: null };
  }

  const segmentId = Number(submission?.segmentId || submission?.task?.segmentId || 0);
  const submittedTaskId = Number(submission?.task?.id || 0);
  const tasks = Array.isArray(projectStatus?.tasks) ? projectStatus.tasks : [];
  const matchingTasks = tasks.filter(task => (
    (segmentId > 0 && Number(task.segmentId) === segmentId)
    || (submittedTaskId > 0 && Number(task.id) === submittedTaskId)
  ));
  const task = matchingTasks[matchingTasks.length - 1] || submission?.task || null;
  const mediaList = Array.isArray(projectStatus?.media) ? projectStatus.media : [];
  const media = [...mediaList].reverse().find(entry => (
    entry.source !== MERGE_SOURCE
    && ((task?.id && Number(entry.taskId) === Number(task.id))
      || (segmentId > 0 && Number(entry.segmentId) === segmentId))
  )) || null;
  const status = media ? 'succeeded' : (task?.status || (submission ? 'queued' : 'draft'));
  return {
    status,
    error: task?.errorMessage || submission?.error || '',
    task,
    media
  };
}

function resolveMergeCandidate(item, projectStatus) {
  const videoStates = (item.directorResult?.storyboard || []).map((video, videoIndex) => ({
    video,
    production: resolveBatchFactoryVideoProduction(item, videoIndex, projectStatus)
  }));
  const mediaIds = videoStates.map(entry => Number(entry.production?.media?.id || 0));
  const allVideosReady = videoStates.length > 0
    && mediaIds.every(id => Number.isInteger(id) && id > 0)
    && videoStates.every(entry => entry.production?.status === 'succeeded');
  const mergedMedia = [...(projectStatus?.media || [])].reverse().find(media => media.source === MERGE_SOURCE) || null;
  const validBookId = /^\d{1,128}$/.test(String(item.bookId || ''));
  return {
    item,
    projectStatus,
    videoStates,
    mediaIds,
    allVideosReady,
    mergedMedia,
    validBookId,
    readyForFirstMerge: allVideosReady && validBookId && !mergedMedia
  };
}

function resolveBookStatus(item, projectStatus) {
  if (item?.status === 'failed' || Number(item?.production?.failed || 0) > 0) return 'failed';
  if (item?.status === 'hook_review') return 'review';
  if (AI_ITEM_STATUSES.has(item?.status)) return 'ai_processing';
  if (!item?.production?.projectId) {
    if (item?.status === 'complete' && item?.directorResult?.storyboard?.length) return 'ready_generate';
    return 'pending';
  }

  const candidate = resolveMergeCandidate(item, projectStatus);
  const states = candidate.videoStates.map(entry => entry.production?.status || 'draft');
  if (states.some(status => status === 'failed' || status === 'cancelled')) return 'failed';
  if (states.some(status => status === 'running')) return 'generating';
  if (!projectStatus || states.some(status => status === 'queued' || status === 'draft')) return 'queued';
  if (candidate.allVideosReady) return candidate.mergedMedia ? 'merged' : 'ready_merge';
  return 'queued';
}

function ProductionStatusTag({ production }) {
  if (!production) return null;
  const [label, color] = taskStatusMeta[production.status] || [production.status || '待生成', 'default'];
  return <Tag color={color}>{label}</Tag>;
}

function VideoResultPreview({ production, downloadName }) {
  const mediaId = Number(production?.media?.id || 0);
  const [previewUrl, setPreviewUrl] = useState('');
  const [previewLoading, setPreviewLoading] = useState(false);
  const [downloadLoading, setDownloadLoading] = useState(false);
  const previewUrlRef = useRef('');

  useEffect(() => {
    if (previewUrlRef.current) URL.revokeObjectURL(previewUrlRef.current);
    previewUrlRef.current = '';
    setPreviewUrl('');
    return () => {
      if (previewUrlRef.current) URL.revokeObjectURL(previewUrlRef.current);
      previewUrlRef.current = '';
    };
  }, [mediaId]);

  if (!mediaId) return null;

  async function loadPreview() {
    setPreviewLoading(true);
    try {
      const blob = await downloadMedia(mediaId);
      const objectUrl = URL.createObjectURL(blob);
      if (previewUrlRef.current) URL.revokeObjectURL(previewUrlRef.current);
      previewUrlRef.current = objectUrl;
      setPreviewUrl(objectUrl);
    } catch (error) {
      message.error(error.message || '加载视频预览失败');
    } finally {
      setPreviewLoading(false);
    }
  }

  async function downloadResult() {
    setDownloadLoading(true);
    try {
      const blob = await downloadMedia(mediaId);
      const objectUrl = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = objectUrl;
      link.download = downloadName || `批量工厂-VIDEO-${mediaId}.mp4`;
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.setTimeout(() => URL.revokeObjectURL(objectUrl), 60000);
    } catch (error) {
      message.error(error.message || '下载视频成品失败');
    } finally {
      setDownloadLoading(false);
    }
  }

  return <Space direction="vertical" size={10} style={{ width: '100%' }}>
    {previewUrl ? <video controls preload="metadata" src={previewUrl} style={{ width: '100%', maxWidth: 560, borderRadius: 8 }} /> : <Button icon={<Play size={15} />} loading={previewLoading} onClick={loadPreview}>加载视频预览</Button>}
    <Button icon={<Download size={15} />} loading={downloadLoading} onClick={downloadResult}>下载成品</Button>
  </Space>;
}

function MergeTimingPanel({ batch, item, videoStates, projectStatus, mergeCapability, refreshNow, bulkMerging }) {
  const [speed, setSpeed] = useState(1.5);
  const [merging, setMerging] = useState(false);
  const storyboardTotal = (item.directorResult?.storyboard || []).reduce((sum, video) => sum + Number(video.duration_sec || 0), 0);
  const completedMediaDurations = videoStates
    .map(entry => Number(entry.production?.media?.durationMs || 0))
    .filter(duration => duration > 0);
  const completedTotalMs = completedMediaDurations.length === videoStates.length && videoStates.length
    ? completedMediaDurations.reduce((sum, duration) => sum + duration, 0)
    : 0;
  const sourceDuration = completedTotalMs > 0 ? completedTotalMs / 1000 : storyboardTotal;
  const estimatedDuration = sourceDuration > 0 ? sourceDuration / speed : 0;
  const mediaIds = videoStates.map(entry => Number(entry.production?.media?.id || 0));
  const allVideosReady = videoStates.length > 0 && mediaIds.every(id => Number.isInteger(id) && id > 0)
    && videoStates.every(entry => entry.production?.status === 'succeeded');
  const mergedMedia = [...(projectStatus?.media || [])].reverse().find(media => media.source === MERGE_SOURCE) || null;
  const validBookId = /^\d{1,128}$/.test(String(item.bookId || ''));
  const canMerge = allVideosReady && validBookId && mergeCapability?.ready === true && !bulkMerging;

  async function mergeVideos() {
    if (!canMerge) return;
    setMerging(true);
    try {
      const result = await mergeBatchFactoryVideos({
        projectId: Number(item.production.projectId),
        bookId: String(item.bookId),
        mediaIds,
        speed
      });
      message.success(`合并完成：${result.filename || `${item.bookId}.mp4`}`);
      refreshNow();
    } catch (error) {
      reportBatchFactoryIssue('batch-factory.merge-failed', { batch, item, error, source: '/api/shuihuo-production/batch-factory/merge-videos' });
      message.error(error.message || '合并视频失败');
    } finally {
      setMerging(false);
    }
  }

  return <Collapse
    size="small"
    items={[{
      key: 'merge',
      label: <Space wrap><Combine size={16} /><span>合并成品</span>{mergedMedia ? <Tag color="green">已生成</Tag> : <Tag>待合并</Tag>}</Space>,
      children: <Space direction="vertical" size={12} style={{ width: '100%' }}>
        <Typography.Text type="secondary">按导演顺序合并：{videoStates.map(({ video }) => `VIDEO ${video.id}`).join(' → ') || '暂无 VIDEO'}。原始 VIDEO 会继续保留。</Typography.Text>
        <div>
          <Typography.Text strong>成品时长处理</Typography.Text>
          <div style={{ marginTop: 8 }}>
            <Segmented
              value="speed"
              options={[
                { value: 'speed', label: '倍率' },
                { value: 'audio', label: '跟随音频时长 · 即将支持', disabled: true }
              ]}
            />
          </div>
        </div>
        <Space wrap>
          <Typography.Text strong>倍率</Typography.Text>
          <Select
            value={speed}
            onChange={setSpeed}
            style={{ width: 140 }}
            options={PREVIEW_SPEED_OPTIONS.map(value => ({ value, label: `${value.toFixed(1)}x` }))}
          />
          <Tag>原始总时长 {sourceDuration ? `${sourceDuration.toFixed(1)}s` : '—'}</Tag>
          <Tag color="processing">预计成品 {estimatedDuration ? `${estimatedDuration.toFixed(1)}s` : '—'}</Tag>
          {validBookId ? <Tag color="blue">文件名 {item.bookId}.mp4</Tag> : <Tag color="red">缺少书ID</Tag>}
        </Space>

        {!allVideosReady ? <Alert type="info" showIcon message="等待全部 VIDEO 生成完成" description="只有当前小说的每个 VIDEO 都成功产生视频文件后，才允许合并，避免缺段或串书。" /> : null}
        {mergeCapability && mergeCapability.ready === false ? <Alert type="warning" showIcon message="视频合并服务未就绪" description={mergeCapability.reason || '服务器未检测到 FFmpeg'} /> : null}
        {!validBookId ? <Alert type="warning" showIcon message="当前小说没有可用于最终文件名的书ID" description="小说获取正常转入的任务会携带书ID；合并成品固定使用 {书ID}.mp4，避免后续上传时串书。" /> : null}
        <Typography.Text type="secondary">跟随音频时长目前只预留入口，配音流程接入后再启用；这里的倍率也与外部后台“解压倍速”保持独立。</Typography.Text>

        <Space wrap>
          <Button type="primary" icon={<Combine size={15} />} loading={merging} disabled={!canMerge} onClick={mergeVideos}>
            {mergedMedia ? '重新合并' : '合并视频'}
          </Button>
          {bulkMerging ? <Tag color="processing">批量合并进行中</Tag> : null}
          {mergedMedia ? <Tag color="green">最终上传候选：{item.bookId}.mp4</Tag> : null}
        </Space>

        {mergedMedia ? <>
          <Divider style={{ margin: '4px 0' }} />
          <Typography.Text strong>合并成品预览</Typography.Text>
          {mergedMedia.durationMs ? <Tag style={{ alignSelf: 'flex-start' }}>{(Number(mergedMedia.durationMs) / 1000).toFixed(1)}s</Tag> : null}
          <VideoResultPreview production={{ media: mergedMedia }} downloadName={`${item.bookId}.mp4`} />
        </> : null}
      </Space>
    }]}
  />;
}

export function BatchFactoryBatchProductionStatus({ batch }) {
  const { byProjectId, error, projectIds, refreshNow } = useBatchFactoryProductionStatus(batch);
  const [retryingTaskId, setRetryingTaskId] = useState(null);
  const [mergeCapability, setMergeCapability] = useState(null);
  const [bulkMergeSpeed, setBulkMergeSpeed] = useState(1.5);
  const [bulkMerging, setBulkMerging] = useState(false);
  const [bulkMergeProgress, setBulkMergeProgress] = useState(null);
  const [statusFilter, setStatusFilter] = useState('all');

  useEffect(() => { setStatusFilter('all'); }, [batch?.id]);

  useEffect(() => {
    if (!projectIds.length) return undefined;
    let active = true;
    getBatchFactoryMergeCapability()
      .then(result => { if (active) setMergeCapability(result); })
      .catch(error => { if (active) setMergeCapability({ ready: false, reason: error.message || '无法检测 FFmpeg 合并能力' }); });
    return () => { active = false; };
  }, [projectIds.join(',')]);

  const statusRows = (batch?.items || []).map(item => {
    const projectStatus = item.production?.projectId ? byProjectId[String(item.production.projectId)] : null;
    return { item, projectStatus, status: resolveBookStatus(item, projectStatus) };
  });
  const statusCounts = statusRows.reduce((result, row) => {
    result[row.status] = (result[row.status] || 0) + 1;
    return result;
  }, {});
  const selectedMeta = bookStatusMeta.find(entry => entry.key === statusFilter) || bookStatusMeta[0];
  const matchedRows = statusFilter === 'all' ? statusRows : statusRows.filter(row => row.status === statusFilter);
  const matchedProducedIds = new Set(matchedRows.filter(row => row.item.production?.projectId).map(row => row.item.id));
  const allProducedItems = (batch?.items || []).filter(item => item.production?.projectId);
  const producedItems = statusFilter === 'all' ? allProducedItems : allProducedItems.filter(item => matchedProducedIds.has(item.id));

  const mergeCandidates = producedItems.map(item => {
    const projectStatus = byProjectId[String(item.production.projectId)];
    return resolveMergeCandidate(item, projectStatus);
  });
  const allMergeCandidates = allProducedItems.map(item => {
    const projectStatus = byProjectId[String(item.production.projectId)];
    return resolveMergeCandidate(item, projectStatus);
  });
  const firstMergeCandidates = allMergeCandidates.filter(candidate => candidate.readyForFirstMerge);
  const alreadyMergedCount = allMergeCandidates.filter(candidate => candidate.mergedMedia).length;

  const states = mergeCandidates.flatMap(candidate => candidate.videoStates.map(entry => entry.production).filter(Boolean));
  const counts = states.reduce((result, state) => {
    const key = state.status || 'draft';
    result[key] = (result[key] || 0) + 1;
    return result;
  }, {});

  async function retry(production) {
    const taskId = Number(production?.task?.id || 0);
    if (!taskId) return message.warning('这个 VIDEO 没有可重试的正式任务，请检查首次提交错误。');
    setRetryingTaskId(taskId);
    try {
      await retryTask(taskId);
      message.success('VIDEO 已重新进入生成队列');
      refreshNow();
    } catch (retryError) {
      message.error(retryError.message || '重试视频任务失败');
    } finally {
      setRetryingTaskId(null);
    }
  }

  async function mergeAllCompleted() {
    if (mergeCapability?.ready !== true) {
      return message.warning(mergeCapability?.reason || '视频合并服务尚未就绪');
    }
    if (!firstMergeCandidates.length) {
      return message.info('当前没有需要首次合并的已完成小说');
    }

    const total = firstMergeCandidates.length;
    let completed = 0;
    let succeeded = 0;
    const failed = [];
    setBulkMerging(true);
    setBulkMergeProgress({ total, completed: 0, succeeded: 0, failed: [] });

    try {
      await runWithConcurrency(
        firstMergeCandidates,
        BULK_MERGE_CONCURRENCY,
        candidate => mergeBatchFactoryVideos({
          projectId: Number(candidate.item.production.projectId),
          bookId: String(candidate.item.bookId),
          mediaIds: candidate.mediaIds,
          speed: bulkMergeSpeed
        }),
        (result, candidate) => {
          completed += 1;
          if (result.ok) {
            succeeded += 1;
          } else {
            failed.push({
              title: candidate.item.title || String(candidate.item.bookId),
              error: result.error?.message || '合并失败'
            });
            reportBatchFactoryIssue('batch-factory.merge-failed', {
              batch,
              item: candidate.item,
              error: result.error,
              source: '/api/shuihuo-production/batch-factory/merge-videos'
            });
          }
          setBulkMergeProgress({ total, completed, succeeded, failed: [...failed] });
        }
      );

      if (failed.length) {
        message.warning(`批量合并完成：成功 ${succeeded} 本，失败 ${failed.length} 本`);
      } else {
        message.success(`批量合并完成：${succeeded} 本全部成功`);
      }
      refreshNow();
    } finally {
      setBulkMerging(false);
    }
  }

  return <Space direction="vertical" size={12} style={{ width: '100%' }}>
    <Card title="批次状态中心" extra={<Typography.Text type="secondary">按小说计数 · 点击筛选</Typography.Text>}>
      <Space direction="vertical" size={12} style={{ width: '100%' }}>
        <Space wrap>
          {bookStatusMeta.map(meta => {
            const count = meta.key === 'all' ? statusRows.length : (statusCounts[meta.key] || 0);
            const active = statusFilter === meta.key;
            return <Button
              key={meta.key}
              type={active ? 'primary' : 'default'}
              danger={meta.key === 'failed'}
              onClick={() => setStatusFilter(meta.key)}
            >
              {meta.label} {count}
            </Button>;
          })}
        </Space>
        {statusFilter !== 'all' ? <>
          <Divider style={{ margin: '0' }} />
          <Space wrap>
            <Typography.Text strong>当前筛选：{selectedMeta.label}</Typography.Text>
            <Tag color={selectedMeta.color}>{matchedRows.length} 本</Tag>
            <Button type="link" size="small" onClick={() => setStatusFilter('all')}>清除筛选</Button>
          </Space>
          {matchedRows.length ? <div style={{ maxHeight: 260, overflow: 'auto', border: '1px solid var(--border-default)', borderRadius: 8 }}>
            {matchedRows.map((row, index) => <div
              key={row.item.id}
              style={{ padding: '9px 12px', borderBottom: index === matchedRows.length - 1 ? 'none' : '1px solid var(--border-default)' }}
            >
              <Space wrap>
                <Typography.Text strong>{row.item.title}</Typography.Text>
                {row.item.bookId ? <Tag>书ID {row.item.bookId}</Tag> : null}
                {row.item.platform ? <Tag color="blue">{row.item.platform}</Tag> : null}
                {row.item.manuallyEdited || row.item.settingOverrides ? <Tag color="purple">单书已调整</Tag> : null}
                {row.status === 'failed' ? <Tag color="red">需要处理</Tag> : null}
              </Space>
            </div>)}
          </div> : <Typography.Text type="secondary">当前没有属于这个状态的小说。</Typography.Text>}
        </> : <Typography.Text type="secondary">异常状态优先级最高；一本小说即使有多个 VIDEO 失败，在这里也只计算为 1 本异常小说。</Typography.Text>}
      </Space>
    </Card>

    {projectIds.length && allProducedItems.length ? <Card title="视频生成进度">
      <Space direction="vertical" size={12} style={{ width: '100%' }}>
        {error ? <Alert type="warning" showIcon message="生产状态暂时读取失败，系统会自动重试" description={error} /> : null}
        {statusFilter !== 'all' && !producedItems.length ? <Alert type="info" showIcon message={`当前“${selectedMeta.label}”筛选中没有已提交视频的小说`} description="匹配小说已列在上方状态中心；待审核、AI处理中、待生成等状态尚未进入视频生产，所以这里不会重复显示。" /> : <>
          <Space wrap>
            <Tag>当前显示 VIDEO {states.length}</Tag>
            <Tag>待生成 {counts.draft || 0}</Tag>
            <Tag color="processing">排队中 {counts.queued || 0}</Tag>
            <Tag color="processing">生成中 {counts.running || 0}</Tag>
            <Tag color="green">已完成 {counts.succeeded || 0}</Tag>
            <Tag color="red">失败 {(counts.failed || 0) + (counts.cancelled || 0)}</Tag>
            <Typography.Text type="secondary">活动任务每 3 秒整批刷新一次；全部结束后自动停止。</Typography.Text>
          </Space>

          <Collapse
            size="small"
            items={[{
              key: 'bulk-merge',
              label: <Space wrap><Combine size={16} /><span>批量合并</span><Tag color="blue">可合并 {firstMergeCandidates.length}</Tag><Tag color="green">已合并 {alreadyMergedCount}</Tag></Space>,
              children: <Space direction="vertical" size={10} style={{ width: '100%' }}>
                <Space wrap>
                  <Typography.Text strong>批量倍率</Typography.Text>
                  <Select
                    value={bulkMergeSpeed}
                    onChange={setBulkMergeSpeed}
                    disabled={bulkMerging}
                    style={{ width: 140 }}
                    options={PREVIEW_SPEED_OPTIONS.map(value => ({ value, label: `${value.toFixed(1)}x` }))}
                  />
                  <Button
                    type="primary"
                    icon={<Combine size={15} />}
                    loading={bulkMerging}
                    disabled={mergeCapability?.ready !== true || !firstMergeCandidates.length}
                    onClick={mergeAllCompleted}
                  >
                    合并全部已完成小说
                  </Button>
                </Space>
                <Typography.Text type="secondary">只处理全部 VIDEO 已成功、书ID有效且尚未生成合并成品的小说；已有合并成品不会重复执行。服务器同时最多处理 {BULK_MERGE_CONCURRENCY} 本。</Typography.Text>
                {mergeCapability && mergeCapability.ready === false ? <Alert type="warning" showIcon message="视频合并服务未就绪" description={mergeCapability.reason || '服务器未检测到 FFmpeg'} /> : null}
                {bulkMergeProgress ? <Alert
                  type={bulkMerging ? 'info' : (bulkMergeProgress.failed.length ? 'warning' : 'success')}
                  showIcon
                  message={bulkMerging
                    ? `批量合并中：${bulkMergeProgress.completed}/${bulkMergeProgress.total}`
                    : `批量合并完成：成功 ${bulkMergeProgress.succeeded} 本，失败 ${bulkMergeProgress.failed.length} 本`}
                  description={bulkMergeProgress.failed.length
                    ? `失败：${bulkMergeProgress.failed.map(entry => `${entry.title}（${entry.error}）`).join('；')}`
                    : `倍率 ${bulkMergeSpeed.toFixed(1)}x；已有合并成品不会重复处理。`}
                /> : null}
              </Space>
            }]}
          />

          <Collapse items={mergeCandidates.map((candidate, itemIndex) => {
            const { item, projectStatus, videoStates, mergedMedia } = candidate;
            const completed = videoStates.filter(entry => entry.production?.status === 'succeeded').length;
            const failed = videoStates.filter(entry => ['failed', 'cancelled'].includes(entry.production?.status)).length;
            return {
              key: item.id,
              label: <Space wrap>
                <Typography.Text strong>{String(itemIndex + 1).padStart(2, '0')} · {item.title}</Typography.Text>
                <Tag color={completed === videoStates.length && videoStates.length ? 'green' : 'processing'}>{completed}/{videoStates.length} 完成</Tag>
                {failed ? <Tag color="red">{failed} 失败</Tag> : null}
                {mergedMedia ? <Tag color="cyan">已合并</Tag> : null}
                {item.manuallyEdited || item.settingOverrides ? <Tag color="purple">单书已调整</Tag> : null}
              </Space>,
              children: <Space direction="vertical" size={10} style={{ width: '100%' }}>
                <Typography.Text type="secondary">生产项目 #{item.production.projectId}</Typography.Text>
                <Collapse
                  size="small"
                  items={videoStates.map(({ video, production }) => ({
                    key: `${item.id}-${video.id}`,
                    label: <Space wrap>
                      <Typography.Text strong>VIDEO {video.id} · {video.duration_sec}秒</Typography.Text>
                      <ProductionStatusTag production={production} />
                      {production?.media?.durationMs ? <Tag>{Math.round(production.media.durationMs / 1000)}秒成品</Tag> : null}
                    </Space>,
                    children: <Space direction="vertical" size={9} style={{ width: '100%' }}>
                      {production?.status === 'draft' ? <Alert type="info" showIcon message="待生成" description="导演方案已经存在，但这个 VIDEO 尚未进入正式视频生成任务。" /> : null}
                      {production?.status === 'queued' ? <Alert type="info" showIcon message="排队中" description="任务已经成功提交，正在等待视频生成服务处理。" /> : null}
                      {production?.status === 'running' ? <Alert type="info" showIcon message="生成中" description="视频生成服务正在处理这个 VIDEO，状态会自动刷新。" /> : null}
                      {['failed', 'cancelled'].includes(production?.status) ? <Alert type="error" showIcon message="视频生成失败" description={production?.error || '生成任务未成功完成'} /> : null}
                      <Space wrap>
                        {production?.task?.providerTaskId ? <Tag>提供方任务 {production.task.providerTaskId}</Tag> : null}
                        {['failed', 'cancelled'].includes(production?.status) && production?.task?.id ? <Button
                          size="small"
                          icon={<RefreshCw size={14} />}
                          loading={retryingTaskId === Number(production.task.id)}
                          onClick={() => retry(production)}
                        >重试这个 VIDEO</Button> : null}
                        {['failed', 'cancelled'].includes(production?.status) ? <Button
                          size="small"
                          icon={<Bug size={14} />}
                          onClick={() => {
                            reportBatchFactoryIssue('batch-factory.video-failed', { batch, item, video, production });
                            message.info('已提交到问题日志，开发者可按书ID和 VIDEO 定位。');
                          }}
                        >标记问题</Button> : null}
                      </Space>
                      {production?.error && !['failed', 'cancelled'].includes(production?.status) ? <Typography.Text type="danger">{production.error}</Typography.Text> : null}
                      <VideoResultPreview production={production} />
                    </Space>
                  }))}
                />
                <MergeTimingPanel
                  batch={batch}
                  item={item}
                  videoStates={videoStates}
                  projectStatus={projectStatus}
                  mergeCapability={mergeCapability}
                  refreshNow={refreshNow}
                  bulkMerging={bulkMerging}
                />
              </Space>
            };
          })} />
        </>}
      </Space>
    </Card> : null}
  </Space>;
}

export function BatchFactoryVideoPlanCard({ item, video, videoIndex, projectStatus, onCompile }) {
  const production = resolveBatchFactoryVideoProduction(item, videoIndex, projectStatus);
  return <Card
    size="small"
    title={<Space wrap><span>VIDEO {video.id} · {video.duration_sec}秒</span><ProductionStatusTag production={production} /></Space>}
    extra={<Button icon={<Sparkles size={15} />} onClick={onCompile}>查看最终上传 Prompt</Button>}
  >
    <Space wrap style={{ marginBottom: 8 }}>
      <Tag>{video.scene || '未指定场景'}</Tag>
      <Tag>{video.prefix_key || 'general_anime'}</Tag>
      <Tag>{video.characters?.join('、') || '无人'}</Tag>
      {production?.task?.providerTaskId ? <Tag>提供方任务 {production.task.providerTaskId}</Tag> : null}
    </Space>
    <Typography.Paragraph style={{ whiteSpace: 'pre-wrap' }}>{video.video_desc}</Typography.Paragraph>
    <Divider style={{ margin: '12px 0' }} />
    {(video.shots || []).map((shot, shotIndex) => <div key={`${video.id}-${shotIndex}`} style={{ marginBottom: 8 }}>
      <Typography.Text strong>{shot.start_sec}-{shot.end_sec}秒 {shot.shot_type ? `· ${shot.shot_type}` : ''} {shot.camera ? `· ${shot.camera}` : ''}</Typography.Text>
      <div>{shot.description}</div>
    </div>)}

    {production?.error ? <Typography.Paragraph type="danger" style={{ marginTop: 12, marginBottom: 0 }}>{production.error}</Typography.Paragraph> : null}
    {production?.media ? <>
      <Divider style={{ margin: '14px 0' }} />
      <Typography.Text strong>视频成品</Typography.Text>
      <div style={{ marginTop: 10 }}><VideoResultPreview production={production} /></div>
    </> : null}
  </Card>;
}