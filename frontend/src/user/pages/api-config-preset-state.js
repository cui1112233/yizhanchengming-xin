export function hasPresetConfiguration(preset, model) {
  if (model?.hasCredential !== true) return false;
  if (preset?.id === 'minimax-h3-video') return Boolean(String(model.workflowId || '').trim());
  return true;
}

export function getPresetCardState(preset, model, saving = false) {
  const configured = hasPresetConfiguration(preset, model);
  const active = model?.enabled === true && configured;
  return {
    configured,
    active,
    editable: !active && !saving,
    buttonText: saving ? '保存中' : active ? '配置已保存' : '确定配置',
    buttonDisabled: active || saving
  };
}
