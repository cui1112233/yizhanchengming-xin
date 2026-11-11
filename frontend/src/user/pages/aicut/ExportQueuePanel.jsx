import { Alert, Button, Empty, Progress, Tag } from 'antd';

const exportStateLabels = {
  queued: '排队中',
  running: '渲染中',
  cancel_requested: '正在取消',
  succeeded: '已完成',
  failed: '失败',
  canceled: '已取消'
};

function exportSummary(job) {
  if (job.format === 'mp3') return 'MP3 音轨';
  const options = job.options || {};
  return [String(job.format || '').toUpperCase(), String(options.resolution || '').toUpperCase(), options.frameRate ? `${options.frameRate} fps` : '', String(options.codec || '').toUpperCase()].filter(Boolean).join(' · ');
}

export default function ExportQueuePanel({ jobs = [], error = '', onReload, onCancel, onDownload, onRetry, busyJobId = '' }) {
  if (!jobs.length) {
    if (error) {
      return <Alert
        type="error"
        showIcon
        message="导出任务队列读取失败"
        description={error}
        action={<Button size="small" onClick={() => onReload?.()}>重试</Button>}
      />;
    }
    return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="还没有导出任务" />;
  }
  return <section className="aicut-export-queue" aria-label="导出任务队列">
    <header><strong>导出任务队列</strong><small>MySQL 保存任务与时间线快照，Redis 按顺序唤醒服务器渲染。</small></header>
    {error ? <Alert
      type="warning"
      showIcon
      message="队列刷新失败，当前显示的是最近一次成功读取的任务"
      description={error}
      action={<Button size="small" onClick={() => onReload?.()}>重试</Button>}
      style={{ marginBottom: 8 }}
    /> : null}
    <div className="aicut-export-queue-list">
      {jobs.map(job => {
        const active = ['queued', 'running', 'cancel_requested'].includes(job.state);
        return <article key={job.id} className={`is-${job.state || 'unknown'}`}>
          <div className="aicut-export-job-main">
            <div><strong>{exportSummary(job)}</strong><small>{job.error || `任务 ${job.id}`}</small></div>
            <Tag>{exportStateLabels[job.state] || job.state}</Tag>
          </div>
          <Progress percent={Number(job.progress || 0)} size="small" status={job.state === 'failed' ? 'exception' : job.state === 'succeeded' ? 'success' : 'active'} />
          <div className="aicut-export-job-actions">
            {active ? <Button size="small" danger disabled={job.state === 'cancel_requested'} loading={busyJobId === job.id} onClick={() => onCancel?.(job)}>取消</Button> : null}
            {job.state === 'succeeded' ? <Button size="small" loading={busyJobId === job.id} onClick={() => onDownload?.(job)}>下载</Button> : null}
            {['failed', 'canceled'].includes(job.state) ? <Button size="small" loading={busyJobId === job.id} onClick={() => onRetry?.(job)}>重新导出</Button> : null}
          </div>
        </article>;
      })}
    </div>
  </section>;
}

export { exportStateLabels, exportSummary };
