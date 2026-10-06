import { describe, expect, it } from 'vitest'
import { getStatusPresentation } from './statusPresentation.js'

describe('getStatusPresentation', () => {
  it.each([
    ['pending', '待执行', 'default'],
    ['queued', '排队中', 'processing'],
    ['running', '执行中', 'processing'],
    ['succeeded', '已完成', 'success'],
    ['completed', '已完成', 'success'],
    ['partial_failed', '部分失败', 'warning'],
    ['retryable_failed', '可重试失败', 'warning'],
    ['failed', '失败', 'error'],
    ['cancelled', '已取消', 'default'],
    ['skipped', '已跳过', 'default'],
    ['fetched', '已获取', 'success'],
  ])('presents %s canonically', (status, label, color) => {
    expect(getStatusPresentation(status)).toMatchObject({ label, color })
  })

  it('does not present an unknown status as pending', () => {
    expect(getStatusPresentation('upstream_waiting')).toMatchObject({
      label: '未知状态（upstream_waiting）',
      color: 'default',
    })
  })
})
