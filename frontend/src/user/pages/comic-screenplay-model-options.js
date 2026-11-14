// Derive choices from the account catalog. Never import another application's providers.
export function availableComicImageModels(models = [], mode = 'generate') {
  return models.filter(model => model?.id && (!model.kind || model.kind === 'image') && model.available !== false && model.enabled !== false)
    .filter(model => {
      const caps = model.capabilities || {};
      const modes = caps.generationModes;
      if (Array.isArray(modes) && modes.length) return modes.includes(mode === 'edit' ? 'image_to_image' : 'text_to_image');
      return mode === 'edit' ? caps.acceptsImageReferences !== false : caps.acceptsTextPrompt !== false;
    });
}

export function comicImageModelFamily(model) {
  // Prefer the real upstream model ID: legacy catalog IDs may contain old names.
  const id = String(model.modelId || model.displayName || model.name || model.id || '').replace(/\[[^\]]*\]|【[^】]*】/g, '').trim();
  if (/gpt[-_ ]?image/i.test(id)) return {key:'gpt-image',label:'GPT 图片'};
  if (/seedream/i.test(id)) return {key:'seedream',label:'Seedream'};
  if (/nano[-_ ]?banana|gemini.*image/i.test(id)) return {key:'gemini-image',label:'Gemini 图片'};
  if (/flux/i.test(id)) return {key:'flux',label:'FLUX'};
  return {key:`model:${id}`,label:id || '已配置图片模型'};
}

export function comicImageModelGroups(models = []) {
  const groups = new Map();
  for (const model of models) {
    const family = comicImageModelFamily(model);
    const group = groups.get(family.key) || {...family,count:0};
    group.count++; groups.set(family.key,group);
  }
  return [{key:'all',label:'全部',count:models.length},...groups.values()];
}

export function comicImageModelChange(field, id, models) {
  if (field !== 'image_model_id') return {[field]:id};
  const editSupported = availableComicImageModels(models, 'edit').some(model => model.id === id);
  return {image_model_id:id,...(editSupported?{image_edit_model_id:id}:{})};
}
