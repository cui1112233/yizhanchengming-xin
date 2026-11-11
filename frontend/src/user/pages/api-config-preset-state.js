// 平台预设卡没有启用开关：是否生效只由凭证决定，输入框在未保存时保持可编辑。
export function hasPresetConfiguration(preset, model) {
  if (model?.hasCredential !== true) return false;
  if (preset?.id === 'minimax-h3-video') return Boolean(String(model.workflowId || '').trim());
  return true;
}

export function getPresetCardState(preset, model, saving = false) {
  return {
    configured: hasPresetConfiguration(preset, model),
    editable: !saving
  };
}
