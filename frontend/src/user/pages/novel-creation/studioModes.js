export const isStudioMode = mode => mode === 'long_original' || mode === 'short_original';

// Preserve existing long-discussion cache keys; short discussions get their own
// selection and unbound draft. Bound sessions already have unique server IDs.
export const studioSelectionKey = (owner, mode) => mode === 'short_original'
  ? `novel:studio:short-original:selected:${encodeURIComponent(owner)}`
  : `novel:studio:selected:${owner}`;
export const studioDraftKey = (owner, id, mode) => `novel:studio:${encodeURIComponent(owner)}:${encodeURIComponent(id || 'new')}${!id && mode === 'short_original' ? ':short_original' : ''}`;
export const studioCreationKey = (owner, mode) => `novel:studio:create-intent:${owner}${mode === 'short_original' ? ':short_original' : ''}`;
