function editingReason(state, requiresSelection = false) {
  if (state.busy) return '正在处理时间线，请稍后';
  if (requiresSelection && !state.hasSelection) return '请先选择片段';
  if (requiresSelection && state.selectionLocked) return '所选片段所在轨道已锁定';
  return '';
}

/** Only supplied command functions become buttons. Capability flags come from the active timeline. */
export function buildTimelinePrimaryActions(state = {}, commands = {}) {
  const edit = editingReason(state, true);
  const busy = editingReason(state);
  const trim = edit || ((state.canTrimLeft === false || state.canTrimRight === false) ? '播放头需要位于所选片段内部' : '');
  const specs = [
    ['undo', '撤销', '撤销 Ctrl+Z', busy || (!state.canUndo ? '没有可撤销的操作' : '')],
    ['redo', '重做', '重做 Ctrl+Shift+Z', busy || (!state.canRedo ? '没有可重做的操作' : '')],
    ['trimLeft', '向左裁到播放头', '删除所选片段在播放头左侧的部分', edit || (state.canTrimLeft === false ? trim : '')],
    ['trimRight', '向右裁到播放头', '删除所选片段在播放头右侧的部分', edit || (state.canTrimRight === false ? trim : '')],
    ['split', '播放头分割', '在播放头处分割', busy || (!state.canSplit ? '播放头处没有可分割的未锁定片段' : '')],
    ['delete', '删除', '删除所选片段 Delete', edit],
    ['cut', '剪切', '剪切所选片段 Ctrl+X', edit],
    ['copy', '复制', '复制所选片段 Ctrl+C', busy || (!state.hasSelection ? '请先选择片段' : '')],
    ['duplicate', '复制副本', '在时间线上复制所选片段', edit],
    ['addKeyframe', '添加关键帧', '在播放头添加关键帧', edit || (state.canAddKeyframe === false ? '播放头需要位于所选片段内部' : '')],
    ['addMarker', '添加标记', '在播放头添加标记', busy],
  ];
  return specs.filter(([key]) => typeof commands[key] === 'function').map(([key, label, hint, disabledReason]) => ({
    key, label, title: disabledReason ? `${label}：${disabledReason}` : hint,
    disabled: Boolean(disabledReason), disabledReason, danger: key === 'delete',
    onClick: (...args) => { if (!disabledReason) return commands[key](...args); },
  }));
}

export function buildTimelineAdvancedItems(state = {}, commands = {}, trackItems = [], primaryKeys = []) {
  const { hasSelection = false, selectionCount = 0, hasLink = false, hasCompound = false, consolidateBusy = false } = state;
  const primary = new Set(primaryKeys);
  const selectionActions = new Set(['cut', 'duplicate', 'consolidate', 'nudgeBack', 'nudgeForward', 'addKeyframe', 'removeKeyframe', 'link', 'unlink', 'createCompound', 'expandCompound', 'openCompound', 'dissolveCompound']);
  const action = (key, label, unavailable = false, extra = {}) => {
    if (typeof commands[key] !== 'function' || primary.has(key)) return null;
    const cancelling = key === 'consolidate' && consolidateBusy;
    const reason = cancelling ? '' : editingReason(state, selectionActions.has(key));
    const disabledReason = reason || (unavailable ? ({
      consolidate: '请至少选择两个片段', createCompound: '请至少选择两个片段',
      link: '所选片段已有联动', unlink: '所选片段没有联动',
      expandCompound: '请选择复合片段', openCompound: '请选择复合片段', dissolveCompound: '请选择复合片段',
    }[key] || '当前没有可操作的片段') : '');
    return { key, label, disabled: Boolean(disabledReason), disabledReason, title: disabledReason || label, onClick: commands[key], ...extra };
  };
  const group = (key, label, children) => ({ key, label, type: 'group', children: children.filter(Boolean) });
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
  return items.filter(item => item.children?.length);
}
