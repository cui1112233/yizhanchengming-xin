export function batchError(reason, fallback = '请求失败，请稍后重试。') {
  let message = fallback
  if (reason?.status === 403) message = '项目不可访问'
  else if (reason?.status === 404) message = '项目不存在或不可访问'
  else if (reason?.code === 'BATCH_PROJECT_ACTIVE') message = '项目有进行中的任务，请等待任务结束后再归档。'
  else if (reason?.code === 'BATCH_PROJECT_ARCHIVED') message = '项目已归档，请恢复后再操作。'
  return { message, requestId: reason?.requestId || '' }
}

export function batchUpdatedAt(value) {
  const time = new Date(value)
  return value && Number.isFinite(time.getTime()) ? time.toLocaleString('zh-CN') : '暂无更新时间'
}
