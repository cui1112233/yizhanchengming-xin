import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const inspectorSource = readFileSync(new URL('./AudioInspector.jsx', import.meta.url), 'utf8')
const editorSource = readFileSync(new URL('./EditorPage.jsx', import.meta.url), 'utf8')

test('selected audio opens the original property order in the right inspector', () => {
  const labels = ['音量', '倍速', '时长']
  let previous = -1
  for (const label of labels) {
    const index = inspectorSource.indexOf(`>${label}<`)
    assert.ok(index > previous, `${label} must follow the original property order`)
    previous = index
  }
  assert.match(inspectorSource, /min=\{0\} max=\{200\}/)
  assert.match(inspectorSource, /min=\{0\.1\} max=\{4\}/)
  assert.match(inspectorSource, /0% 即静音/)
  assert.match(editorSource, /selection\?\.trackKind === 'audio'/)
  assert.match(editorSource, /<AudioInspector/)
})
