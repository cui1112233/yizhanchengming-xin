import { describe, expect, it } from 'vitest'
import {
  joinShotCards,
  joinShotDocument,
  replaceInText,
  splitShotCards,
  splitShotDocument,
  workbenchExportArtifact,
} from './scriptWorkbenchState.js'

describe('scriptWorkbenchState', () => {
  it('splits and rejoins editable shot cards', () => {
    const source = '### 分镜一\n00:00-00:03 | 推门\n\n---\n\n### 分镜二\n00:00-00:02 | 回头'
    const cards = splitShotCards(source)
    expect(cards).toHaveLength(2)
    expect(cards[0].title).toBe('分镜一')
    expect(cards[1].body).toContain('回头')
    expect(joinShotCards(cards)).toContain('### 分镜二')
  })

  it('preserves a real compiler-shaped envelope instead of dropping prefix or mixing global rules into the last card', () => {
    const compiled = [
      'SYSTEM PRESET:\nSYS',
      'SCRIPT:\n开场说明\n### 分镜一\n00:00-00:03 | 推门\n\n---\n\n### 分镜二\n00:00-00:02 | 回头',
      'HOOK:\nHOOK',
      'DIRECTOR:\nDIRECTOR',
      'PROCESSING RULES:\nRULES',
    ].join('\n\n')
    const document = splitShotDocument(compiled)
    expect(document.prefix).toContain('SYSTEM PRESET:\nSYS')
    expect(document.prefix).toContain('SCRIPT:\n开场说明')
    expect(document.cards).toHaveLength(2)
    expect(document.cards[1].body).toBe('00:00-00:02 | 回头')
    expect(document.suffix).toContain('HOOK:\nHOOK')
    expect(document.suffix).toContain('PROCESSING RULES:\nRULES')

    document.cards[0] = { ...document.cards[0], body: '00:00-00:03 | 修改后的推门' }
    const rebuilt = joinShotDocument(document)
    for (const needle of ['SYSTEM PRESET:', 'SCRIPT:', '修改后的推门', 'HOOK:', 'DIRECTOR:', 'PROCESSING RULES:', 'RULES']) {
      expect(rebuilt).toContain(needle)
    }
  })

  it('supports replacement and deterministic export artifacts', () => {
    expect(replaceInText('她看他，他回头', '他', '林川', false)).toBe('她看林川，他回头')
    expect(replaceInText('她看他，他回头', '他', '林川', true)).toBe('她看林川，林川回头')
    expect(workbenchExportArtifact({ editorOutput: '成品', sourceText: '原文', bookTitle: '测试书', outputMode: 'script' }))
      .toEqual({ filename: '测试书-script.txt', text: '成品' })
  })
})
