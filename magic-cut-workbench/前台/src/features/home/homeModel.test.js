import assert from 'node:assert/strict'
import test from 'node:test'

import { createHomeModel } from './homeModel.js'

test('home model exposes the original first-use actions in order', () => {
  assert.deepEqual(createHomeModel(), {
    productName: '魔剪',
    sectionTitle: '近期工程',
    primaryAction: '新建自由创作',
    emptyTitle: '还没有工程',
    emptyDescription: '新建一个工程后，就可以导入素材并开始剪辑。',
  })
})
