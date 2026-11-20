import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const source = readFileSync(new URL('./ExportDialog.jsx', import.meta.url), 'utf8')

test('export dialog submits the backend-only render task and links to task center', () => {
  assert.match(source, /\/api\/magic-cut\/exports/)
  assert.match(source, /buildExportPayload\(projectID/)
  assert.match(source, /任务中心/)
})

test('export dialog exposes only the currently supported codecs', () => {
  assert.match(source, /H\.264/)
  assert.match(source, /AAC/)
  assert.match(source, /MP4/)
  assert.doesNotMatch(source, /systemPrompt|apiKey|secret/i)
})
