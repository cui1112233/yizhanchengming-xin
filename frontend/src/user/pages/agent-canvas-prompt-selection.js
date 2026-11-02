const PROMPT_TYPES_BY_RESOURCE = {
  book_script: ['hook', 'director'],
  book_asset: ['visual'],
  video: ['video'],
  audio: ['audio_match'],
  merge: ['book_merge']
};

function settingsState(batch) {
  const state = batch?.settingsState || batch?.settings || {};
  return state && typeof state === 'object' ? state : {};
}

export function batchFactoryPromptSelectorTypes(config = {}) {
  if (String(config.source || '').trim() !== 'batch_factory') return [];
  return [...(PROMPT_TYPES_BY_RESOURCE[String(config.resourceType || '').trim()] || [])];
}

export async function saveBatchFactoryPromptSelection({ batch, type, promptId, save }) {
  const batchId = String(batch?.id || '').trim();
  const promptType = String(type || '').trim();
  if (!batchId || !promptType || typeof save !== 'function') throw new Error('提示词配置保存条件不完整');
  const state = settingsState(batch);
  const existing = state.patch?.promptSelections && typeof state.patch.promptSelections === 'object'
    ? state.patch.promptSelections
    : {};
  return save(batchId, {
    patch: { promptSelections: { ...existing, [promptType]: String(promptId || '').trim() } },
    expectedRevision: Number(state.revision || 0)
  });
}
