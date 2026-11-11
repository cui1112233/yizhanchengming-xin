// 「自定义预设」只决定哪些预置渠道显示在文本 / 视频 / 图片组里。
// 是否可用仍然只看 Key 是否填写，不引入启用开关。
// 勾选结果存放在服务端 apiIntegrations.channelSelection，形如：
// { text: { enabled: ['deepseek', 'doubao'] } }
// 没有该分组记录 = 默认展示全部预置；记录为空数组 = 用户显式全部取消。

export function explicitPresetIds(channelSelection, groupId) {
  const raw = channelSelection?.[groupId];
  if (!raw || typeof raw !== 'object' || !Array.isArray(raw.enabled)) return null;
  return raw.enabled.map(value => String(value ?? '').trim()).filter(Boolean);
}

export function selectedPresetIds(channelSelection, groupId, allIds) {
  const ids = (Array.isArray(allIds) ? allIds : []).map(value => String(value ?? '').trim()).filter(Boolean);
  const explicit = explicitPresetIds(channelSelection, groupId);
  if (explicit === null) return ids;
  const explicitSet = new Set(explicit);
  // 结果按预置清单顺序返回，避免卡片顺序随勾选顺序变化。
  return ids.filter(id => explicitSet.has(id));
}

export function filterPresetsBySelection(presets, channelSelection, groupId) {
  const list = Array.isArray(presets) ? presets : [];
  const visible = new Set(selectedPresetIds(channelSelection, groupId, list.map(preset => preset.id)));
  return list.filter(preset => visible.has(preset.id));
}

export function buildPresetSelectionPayload(channelSelection, groupId, ids) {
  return {
    ...(channelSelection || {}),
    [groupId]: { enabled: Array.from(new Set((Array.isArray(ids) ? ids : []).map(id => String(id ?? '').trim()).filter(Boolean))) }
  };
}
