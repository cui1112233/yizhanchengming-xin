import { describe, expect, it } from 'vitest'
import { joinShotCards, replaceInText, splitShotCards } from './scriptWorkbenchState.js'

describe('scriptWorkbenchState', () => {
  it('splits and rejoins editable shot cards', () => {
    const source = '### 分镜一\n00:00-00:03 | 推门\n\n---\n\n### 分镜二\n00:00-00:02 | 回头'
    const cards = splitShotCards(source)
    expect(cards).toHaveLength(2)
    expect(cards[0].title).toBe('分镜一')
    expect(cards[1].body).toContain('回头')
    expect(joinShotCards(cards)).toContain('### 分镜二')
  })

  it('supports one or all replacements', () => {
    expect(replaceInText('她看他，他回头', '他', '林川', false)).toBe('她看林川，他回头')
    expect(replaceInText('她看他，他回头', '他', '林川', true)).toBe('她看林川，林川回头')
  })
})
