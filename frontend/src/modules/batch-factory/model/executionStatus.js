const STATUS = {
  planned: { label: '等待执行', tone: 'default', terminal: false },
  running: { label: '执行中', tone: 'processing', terminal: false },
  reconciling: { label: '核对中', tone: 'warning', terminal: false },
  retry_wait: { label: '等待重试', tone: 'warning', terminal: false },
  needs_attention: { label: '需要处理', tone: 'error', terminal: true },
  completed: { label: '已完成', tone: 'success', terminal: true },
  cancelled: { label: '已取消', tone: 'default', terminal: true }
};

export function executionViewModel(execution = {}) {
  const status = String(execution.status || 'planned');
  const presentation = STATUS[status] || { label: `未知状态（${status}）`, tone: 'error', terminal: true };
  return { ...execution, id: String(execution.id || execution.executionId || ''), status, ...presentation };
}
