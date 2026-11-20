import assert from 'node:assert/strict'
import test from 'node:test'

import { CREATE_PROJECT_FIELDS, buildCreateProjectPayload } from './createProjectModel.js'

test('create project fields keep the original-use order', () => {
  assert.deepEqual(CREATE_PROJECT_FIELDS.map(field => field.label), ['工程名称', '画面比例', '分辨率', '帧率'])
})

test('create payload retains the selected canvas and retry key', () => {
  assert.deepEqual(buildCreateProjectPayload({
    name: '  竖屏口播  ',
    ratio: '9:16',
    resolution: '1080x1920',
    frameRate: 30,
    idempotencyKey: 'create-1',
  }), {
    name: '竖屏口播',
    width: 1080,
    height: 1920,
    frameRate: 30,
    idempotencyKey: 'create-1',
  })
})
