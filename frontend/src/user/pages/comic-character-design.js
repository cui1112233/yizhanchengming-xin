function cloneSelection(value) {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return {};
  return JSON.parse(JSON.stringify(value));
}
export function characterDesignInitialSelection(catalog, initial) {
  const defaults = {};
  for (const category of catalog) for (const field of category.fields) defaults[field.key] = field.multiple ? [] : '';
  return { ...defaults, ...cloneSelection(initial) };
}
export function characterDesignOptionValue(option) {
  return typeof option === 'string' ? option : String(option?.label || option?.value || '');
}
export function toggleCharacterDesignOption(selection, field, option) {
  const value = characterDesignOptionValue(option);
  if (!value) return { ...selection };
  if (!field.multiple) return { ...selection, [field.key]: selection[field.key] === value ? '' : value };
  const current = Array.isArray(selection[field.key]) ? selection[field.key] : [];
  return { ...selection, [field.key]: current.includes(value) ? current.filter(item => item !== value) : [...current, value] };
}
export function characterDesignSelectionText(selection, catalog) {
  const values = [];
  for (const category of catalog) for (const field of category.fields) {
    const value = selection[field.key];
    for (const item of Array.isArray(value) ? value : [value]) {
      const text = String(item || '').trim();
      if (text && !values.includes(text)) values.push(text);
    }
  }
  return values.join('，');
}
export function characterDesignSavedSelection(asset) {
  try { return cloneSelection(JSON.parse(asset?.extra?.character_design_selection || '{}')); } catch { return {}; }
}
// Update only the exact design suffix that we previously applied. Manual edits
// outside that suffix are preserved. An unchanged design is never appended twice.
export function applyCharacterDesign(asset, { text, selection }) {
  const next = String(text || '').trim(), previous = String(asset?.extra?.character_design_text || '').trim();
  let prompt = String(asset?.prompt || '');
  const extra = { ...asset?.extra, character_design_selection: JSON.stringify(cloneSelection(selection)) };
  if (!next) {
    const current = prompt.trimEnd();
    if (previous && current.trim() === previous) prompt = '';
    else if (previous && current.endsWith(`，${previous}`)) prompt = current.slice(0, -(`，${previous}`).length) + prompt.slice(current.length);
    delete extra.character_design_text;
  } else if (next !== previous) {
    const current = prompt.trim();
    if (previous && current === previous) prompt = next;
    else if (previous && current.endsWith(`，${previous}`)) prompt = current.slice(0, -previous.length) + next;
    else prompt = current ? `${current}，${next}` : next;
  }
  if (next) extra.character_design_text = next;
  return { ...asset, prompt, extra };
}
