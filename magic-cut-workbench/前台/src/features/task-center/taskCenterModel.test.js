import assert from 'node:assert/strict'
import test from 'node:test'

import { TASK_COLUMNS, taskAction, taskDownloadHref, taskStatus } from './taskCenterModel.js'

test('task center keeps the original-use columns in order', () => {
  assert.deepEqual(TASK_COLUMNS, ['任务', '工程', '状态', '进度', '更新时间', '操作'])
})

test('task status and available action are explicit', () => {
  assert.deepEqual(taskStatus('running'), { label: '处理中', color: 'processing' })
  assert.equal(taskAction({ status: 'running' }), 'cancel')
  assert.equal(taskAction({ status: 'failed' }), 'retry')
  assert.equal(taskAction({ status: 'succeeded' }), '')
})

test('successful export exposes an owned backend download URL', () => {
  assert.equal(taskDownloadHref({ type: 'export_video', status: 'succeeded', input: { exportId: 'mce_1' } }), '/api/magic-cut/exports/mce_1/file')
  assert.equal(taskDownloadHref({ type: 'material_probe', status: 'succeeded' }), '')
  assert.equal(taskDownloadHref({ type: 'export_video', status: 'running', input: { exportId: 'mce_1' } }), '')
})
