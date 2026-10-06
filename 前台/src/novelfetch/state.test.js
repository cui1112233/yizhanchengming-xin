import { describe, expect, it } from 'vitest'
import { formatSensitiveReplacementLines, parseChapterPrefixLines, parseSensitiveReplacementLines } from './state.js'

describe('novel fetch rule editor state', () => {
  it('round-trips replacement and chapter rules deterministically', () => {
    const parsed = parseSensitiveReplacementLines('敏感=>合规\n敏感词=>安全词\n')
    expect(parsed).toEqual({ 敏感: '合规', 敏感词: '安全词' })
    expect(formatSensitiveReplacementLines(parsed)).toBe('敏感词=>安全词\n敏感=>合规')
    expect(parseChapterPrefixLines('第一章\n第1章\n第一章')).toEqual(['第一章', '第1章'])
  })
})
