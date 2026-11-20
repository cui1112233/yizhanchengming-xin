export const TASK_COLUMNS = ['任务', '工程', '状态', '进度', '更新时间', '操作']

const statuses = {
  pending: { label: '等待中', color: 'default' },
  running: { label: '处理中', color: 'processing' },
  succeeded: { label: '已完成', color: 'success' },
  failed: { label: '失败', color: 'error' },
  canceled: { label: '已取消', color: 'warning' },
}
export function taskStatus(status) { return statuses[status] || { label: status || '未知', color: 'default' } }
export function taskAction(task) { if (task?.status === 'pending' || task?.status === 'running') return 'cancel'; if (task?.status === 'failed' || task?.status === 'canceled') return 'retry'; return '' }
export function taskTypeLabel(type) { return { material_probe: '素材解析', export_video: '视频导出' }[type] || type || '任务' }
export function taskDownloadHref(task) {
  const exportId = task?.input?.exportId
  return task?.type === 'export_video' && task?.status === 'succeeded' && exportId ? `/api/magic-cut/exports/${exportId}/file` : ''
}
