import assert from 'node:assert/strict'
import test from 'node:test'

import { createHomeModel } from './homeModel.js'

test('home model keeps the original primary and secondary tools in screenshot order', () => {
  const model = createHomeModel()

  assert.deepEqual(model.primaryTools.map(item => item.title), ['原片顺剪', 'AI前贴', '选剧助手'])
  assert.deepEqual(model.secondaryTools.map(item => item.title), [
    '批量顺剪', '新剧抢发', '爆剧仿剪', '高光短切', '视频裂变', '封面素材', '漫剧转译',
  ])
  assert.deepEqual(model.stats.map(item => item.label), ['魔剪素材数', '已使用素材数', '有消耗素材数', '消耗'])
})

test('home model keeps the original sidebar hierarchy', () => {
  const model = createHomeModel()

  assert.deepEqual(model.sidebarGroups.map(group => ({
    label: group.label,
    items: group.items.map(item => item.title),
  })), [
    { label: '', items: ['创作', '原片顺剪', 'AI前贴', '选剧助手', '历史工程'] },
    { label: '素材资产', items: ['魔剪云空间', '创量成片库', '创量灵感库'] },
    { label: '数据与任务', items: ['数据报表', '任务中心'] },
  ])
})

test('only implemented home entries claim to be available', () => {
  const model = createHomeModel()

  assert.equal(model.primaryTools.find(item => item.title === '原片顺剪').action, 'create-project')
  assert.equal(model.primaryTools.find(item => item.title === '原片顺剪').available, true)
  assert.equal(model.primaryTools.find(item => item.title === 'AI前贴').available, false)
  assert.equal(model.primaryAction, '新建自由创作')
})
