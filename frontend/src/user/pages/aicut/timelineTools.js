export function buildTimelineAdvancedItems(state = {}, commands = {}, trackItems = []) {
  const { hasSelection = false, selectionCount = 0, hasLink = false, hasCompound = false, consolidateBusy = false } = state;
  const action = (key, label, disabled = false, extra = {}) => ({ key, label, disabled, onClick: commands[key], ...extra });
  const group = (key, label, children) => ({ key, label, type: 'group', children });
  const items = [
    group('clip-actions', '片段操作', [
      action('cut', '剪切', !hasSelection),
      action('duplicate', '复制副本', !hasSelection),
      action('consolidate', consolidateBusy ? '取消合并' : '合并为素材', !consolidateBusy && selectionCount < 2, { danger: consolidateBusy }),
    ]),
    group('navigation-actions', '定位与逐帧', [
      action('previousEdit', '上一剪辑点'),
      action('nextEdit', '下一剪辑点'),
      action('nudgeBack', '向前一帧', !hasSelection),
      action('nudgeForward', '向后一帧', !hasSelection),
    ]),
    group('keyframe-marker-actions', '关键帧与标记', [
      action('addKeyframe', '添加关键帧', !hasSelection),
      action('removeKeyframe', '移除关键帧', !hasSelection),
      action('addMarker', '添加标记'),
      action('removeMarker', '移除标记', false, { danger: true }),
    ]),
    group('linked-compound-actions', '联动与复合片段', [
      action('link', '建立联动', !hasSelection || hasLink),
      action('unlink', '取消联动', !hasSelection || !hasLink),
      action('createCompound', '创建复合片段', selectionCount < 2),
      action('expandCompound', '展开复合成员', !hasSelection || !hasCompound),
      action('openCompound', '打开复合时间线', !hasSelection || !hasCompound),
      action('dissolveCompound', '拆散复合片段', !hasSelection || !hasCompound, { danger: true }),
    ]),
  ];
  if (trackItems.length) items.push({ key: 'track-settings', label: '轨道设置', children: trackItems });
  return items;
}
