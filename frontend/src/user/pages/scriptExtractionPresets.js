export const EXTRACTION_PRESET_MIGRATIONS = {
  standard: 'script-extract'
};

export function filterExtractionPresets(catalog) {
  return (Array.isArray(catalog) ? catalog : [])
    .filter(item => item?.kind === 'base' && item?.extractionPreset === true);
}

export function normalizeExtractionPresetId(value) {
  return EXTRACTION_PRESET_MIGRATIONS[value] || value || 'script-extract';
}

export function selectAvailableExtractionPreset(value, presets) {
  const normalized = normalizeExtractionPresetId(value);
  return presets.some(item => item.id === normalized) ? normalized : (presets[0]?.id || '');
}
