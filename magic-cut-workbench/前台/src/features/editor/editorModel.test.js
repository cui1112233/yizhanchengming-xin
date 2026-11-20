import assert from 'node:assert/strict'
import test from 'node:test'

import { EDITOR_REGIONS, commandForShortcut, projectIDFromPath } from './editorModel.js'

test('editor keeps the five original work regions in order', () => {
  assert.deepEqual(EDITOR_REGIONS, ['素材区', '播放器', '检查器', '工具栏', '时间线'])
})

test('editor keyboard shortcuts map to typed commands', () => {
  assert.equal(commandForShortcut({ code: 'Space' }), 'toggle-playback')
  assert.equal(commandForShortcut({ code: 'Delete' }), 'delete-selection')
  assert.equal(commandForShortcut({ code: 'KeyZ', ctrlKey: true }), 'undo')
  assert.equal(commandForShortcut({ code: 'KeyZ', ctrlKey: true, shiftKey: true }), 'redo')
  assert.equal(commandForShortcut({ code: 'KeyC', ctrlKey: true }), 'copy')
  assert.equal(commandForShortcut({ code: 'KeyV', ctrlKey: true }), 'paste')
})

test('project editor route extracts only an opaque project id', () => {
  assert.equal(projectIDFromPath('/projects/mcp_abc/edit'), 'mcp_abc')
  assert.equal(projectIDFromPath('/projects/../../edit'), '')
  assert.equal(projectIDFromPath('/'), '')
})
