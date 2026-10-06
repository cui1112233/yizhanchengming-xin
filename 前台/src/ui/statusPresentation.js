export const STATUS_PRESENTATION = Object.freeze({
  pending: { label: '待执行', color: 'default' },
  queued: { label: '排队中', color: 'processing' },
  running: { label: '执行中', color: 'processing' },
  succeeded: { label: '已完成', color: 'success' },
  completed: { label: '已完成', color: 'success' },
  partial_failed: { label: '部分失败', color: 'warning' },
  retryable_failed: { label: '可重试失败', color: 'warning' },
  failed: { label: '失败', color: 'error' },
  cancelled: { label: '已取消', color: 'default' },
  skipped: { label: '已跳过', color: 'default' },
  fetched: { label: '已获取', color: 'success' },
})

export function getStatusPresentation(status) {
  const normalized = String(status || '').trim().toLowerCase()
  return STATUS_PRESENTATION[normalized] || {
    label: normalized ? `未知状态（${normalized}）` : '未知状态',
    color: 'default',
  }
}
