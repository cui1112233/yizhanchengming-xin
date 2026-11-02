const SMART_UNIFIED_PREFIX_PRESET_ID = 'script-constraint-prefix-smart-unified';
const DEFAULT_ENABLED_CATEGORIES = new Set(['prefix', 'quality', 'restriction', 'negative']);

function selectedLayer(constraints, category) {
  const enabledCategories = Array.isArray(constraints?.enabledCategories)
    ? new Set(constraints.enabledCategories)
    : DEFAULT_ENABLED_CATEGORIES;
  const enabled = enabledCategories.has(category);
  if (!enabled) return null;
  return (Array.isArray(constraints?.selections) ? constraints.selections : [])
    .find(item => item?.constraintCategory === category) || null;
}

function text(value) {
  return typeof value === 'string' ? value.trim() : '';
}

export function buildBatchFactoryH3Constraints(constraints) {
  const prefix = selectedLayer(constraints, 'prefix');
  const quality = selectedLayer(constraints, 'quality');
  const restriction = selectedLayer(constraints, 'restriction');
  const negative = selectedLayer(constraints, 'negative');
  const smartUnified = prefix?.presetId === SMART_UNIFIED_PREFIX_PRESET_ID;
  return {
    prefix_text: smartUnified ? '' : text(prefix?.body),
    quality_text: text(quality?.body),
    visual_restriction_text: text(restriction?.body),
    negative_text: text(negative?.body),
    switches: {
      smart_unified: smartUnified,
      base_setup: constraints?.baseSetup?.enabled !== false,
      prefix: Boolean(prefix && !smartUnified && text(prefix?.body)),
      quality: Boolean(quality && text(quality?.body)),
      visual_restriction: Boolean(restriction && text(restriction?.body)),
      negative: Boolean(negative && text(negative?.body))
    }
  };
}
