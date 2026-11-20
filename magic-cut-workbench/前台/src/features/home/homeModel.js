export function createHomeModel() {
  return {
    productName: '魔剪',
    sectionTitle: '近期工程',
    primaryAction: '新建自由创作',
    emptyTitle: '还没有工程',
    emptyDescription: '新建一个工程后，就可以导入素材并开始剪辑。',
    primaryTools: [
      { id: 'rough-cut', title: '原片顺剪', description: '基于单部剧快速出片', action: 'create-project', available: true, tone: 'purple' },
      { id: 'ai-prepend', title: 'AI前贴', description: '参考爆款生成全新前贴', action: 'navigate', available: false, tone: 'blue' },
      { id: 'drama-picker', title: '选剧助手', description: '一页看全部剧 + 爆剧率', action: 'navigate', available: false, tone: 'green', badge: 'NEW·限免' },
    ],
    secondaryTools: [
      { id: 'batch-cut', title: '批量顺剪', description: '多部剧集批量顺剪', action: 'navigate', available: false },
      { id: 'new-drama', title: '新剧抢发', description: '新剧集快速首发', action: 'navigate', available: false },
      { id: 'remix', title: '爆剧仿剪', description: '模仿爆款剧集结构', action: 'navigate', available: false },
      { id: 'highlight', title: '高光短切', description: '智能截取高光片段', action: 'navigate', available: false },
      { id: 'split', title: '视频裂变', description: '批量裂变爆款视频', action: 'navigate', available: false },
      { id: 'cover', title: '封面素材', description: '批量生成封面素材', action: 'navigate', available: false },
      { id: 'comic', title: '漫剧转译', description: '批量素材一键转译', action: 'navigate', available: false, badge: '限免' },
    ],
    sidebarGroups: [
      { label: '', items: [
        { title: '创作', icon: 'layout', active: true },
        { title: '原片顺剪', indent: true, action: 'create-project', available: true },
        { title: 'AI前贴', indent: true, available: false },
        { title: '选剧助手', indent: true, available: false },
        { title: '历史工程', indent: true, available: false },
      ] },
      { label: '素材资产', items: [
        { title: '魔剪云空间', icon: 'cloud', available: false },
        { title: '创量成片库', icon: 'video', available: false },
        { title: '创量灵感库', icon: 'bulb', available: false },
      ] },
      { label: '数据与任务', items: [
        { title: '数据报表', icon: 'chart', available: false },
        { title: '任务中心', icon: 'tasks', href: '/tasks', available: true },
      ] },
    ],
    stats: [
      { label: '魔剪素材数', value: '0' },
      { label: '已使用素材数', value: '0' },
      { label: '有消耗素材数', value: '0' },
      { label: '消耗', value: '¥0.00' },
    ],
  }
}
