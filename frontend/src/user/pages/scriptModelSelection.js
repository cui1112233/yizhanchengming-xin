function available(ids, selectedId) {
  return Boolean(selectedId) && ids.has(selectedId);
}

function firstEnabledModelId(models) {
  const item = (Array.isArray(models) ? models : []).find(model => model?.enabled !== false && String(model?.id || '').trim());
  return String(item?.id || '').trim();
}

const preferencePrefix = 'qiantie:script-model-selection:';
function preferenceKey(username) { return `${preferencePrefix}${encodeURIComponent(String(username || 'guest'))}`; }

export function loadScriptModelSelection(storage, username) {
  try {
    const value = JSON.parse(storage?.getItem(preferenceKey(username)) || '{}');
    return { textModelId: String(value?.textModelId || ''), imageModelId: String(value?.imageModelId || ''), videoModelKey: String(value?.videoModelKey || '') };
  } catch { return { textModelId: '', imageModelId: '', videoModelKey: '' }; }
}

export function saveScriptModelSelection(storage, username, selection = {}) {
  try {
    storage?.setItem(preferenceKey(username), JSON.stringify({ textModelId: String(selection.textModelId || ''), imageModelId: String(selection.imageModelId || ''), videoModelKey: String(selection.videoModelKey || '') }));
    return true;
  } catch { return false; }
}

export function reconcileScriptModelSelection(selection = {}, modelsByKind = {}) {
  const textModelId = String(selection.textModelId || '').trim();
  const imageModelId = String(selection.imageModelId || '').trim();
  const textIds = new Set((Array.isArray(modelsByKind.text) ? modelsByKind.text : [])
    .filter(model => model?.enabled !== false)
    .map(model => String(model?.id || '').trim())
    .filter(Boolean));
  const imageIds = new Set((Array.isArray(modelsByKind.image) ? modelsByKind.image : [])
    .filter(model => model?.enabled !== false)
    .map(model => String(model?.id || '').trim())
    .filter(Boolean));
  const unavailableKinds = [];
  if (textModelId && !available(textIds, textModelId)) unavailableKinds.push('text');
  if (imageModelId && !available(imageIds, imageModelId)) unavailableKinds.push('image');
  return {
    textModelId: available(textIds, textModelId) ? textModelId : '',
    imageModelId: available(imageIds, imageModelId) ? imageModelId : '',
    unavailableKinds
  };
}

export function defaultEnabledModelSelection(selection = {}, modelsByKind = {}) {
  const reconciled = reconcileScriptModelSelection(selection, modelsByKind);
  return {
    textModelId: reconciled.textModelId || firstEnabledModelId(modelsByKind.text),
    imageModelId: reconciled.imageModelId || firstEnabledModelId(modelsByKind.image)
  };
}
