import assert from 'node:assert/strict'
import test from 'node:test'

import { extractCandidates } from './extract-prompt-candidates.mjs'

test('extractor reports prompt-like quoted text with byte range and sha256 without executing code', () => {
  const source = `const safe = "不生成字幕；画面无水印、台标和乱码。"; throw new Error("must not execute")`
  const result = extractCandidates(source, 'fixture.js')

  assert.equal(result.length, 1)
  assert.equal(result[0].file, 'fixture.js')
  assert.equal(result[0].text, '不生成字幕；画面无水印、台标和乱码。')
  assert.ok(result[0].startByte < result[0].endByte)
  assert.match(result[0].sha256, /^[a-f0-9]{64}$/)
})
