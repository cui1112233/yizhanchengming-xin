const messages = Object.freeze({
  GENERATION_INVALID: '生成参数无效，请检查后重试',
  GENERATION_NOT_FOUND: '生成记录不存在，请刷新后重试',
  GENERATION_CONFLICT: '当前生成状态不允许此操作，请刷新后重试',
  AUDIO_MEASUREMENT_REQUIRED: '请先生成或检测音频',
  AUDIO_PROBE_UNAVAILABLE: '音频检测服务暂不可用，请稍后重试',
  GENERATION_TIMELINE_INVALID: '导演分镜时长校验失败，请重试导演阶段',
  GENERATION_UNAVAILABLE: '生成服务暂不可用，请稍后重试',
  GENERATION_FAILED: '生成阶段执行失败，请稍后重试',
})

// Only fixed catalogue copy may become generation feedback. Unknown legacy
// diagnostics and codes never become display text.
export function generationOutcomeMessage(outcome) {
  if (typeof outcome?.errorCode === 'string' && Object.hasOwn(messages, outcome.errorCode)) return messages[outcome.errorCode]
  const message = outcome?.errorMessage || outcome?.error
  if (!outcome?.errorCode && Object.values(messages).includes(message)) return message
  return messages.GENERATION_FAILED
}

export function generationValidationSummary(raw) {
  if (typeof raw !== 'string') return ''
  try {
    const value = JSON.parse(raw)
    if (!value || typeof value !== 'object' || Array.isArray(value)) return ''
    const facts = []
    if (typeof value.valid === 'boolean') facts.push(value.valid ? '校验通过' : '校验未通过')
    if (typeof value.repaired === 'boolean') facts.push(value.repaired ? '已修复' : '未修复')
    if (typeof value.durationMs === 'number' && Number.isFinite(value.durationMs) && value.durationMs >= 0) facts.push(`${(value.durationMs / 1000).toFixed(2)} 秒`)
    return facts.join(' · ')
  } catch {
    return ''
  }
}
