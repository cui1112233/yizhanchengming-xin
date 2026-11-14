export function assetEditorDraft(asset, type = 'character', degree = 'standard') {
  return { type, degree, name: '', identity: '', description: '', prompt: '', voice: '', episodes: '', tag: '', ...asset, extra: { ...(asset?.extra || {}) }, alias: asset?.extra?.alias || '' };
}
export function assetEditorPayload(draft) {
  if (!String(draft.name || '').trim()) throw new Error('请输入资产名称');
  return { type: draft.type, degree: draft.degree, name: draft.name.trim(), identity: draft.identity || '', description: draft.description || '', prompt: draft.prompt || '', voice: draft.voice || '', episodes: draft.episodes || '', tag: draft.tag || '', extra: { ...(draft.extra || {}), alias: draft.alias || '' }, expectedRevision: Number(draft.revision || 0) };
}
export function voiceDraftTarget(row) { return row.asset_id || `voice-draft:${row.id}`; }
export function voiceDesignAudio(items, row) {
  return items.filter(item => item.kind === 'audio' && item.voiceId && item.voiceEngine && item.targetId === voiceDraftTarget(row) && ['succeeded', 'ready', 'completed'].includes(item.status)).at(-1);
}
export function validateVoiceDesign(row, engine) {
  if (!['minimax', 'qwen-tts'].includes(engine)) return '请在模型配置中选择 MiniMax 或 Qwen 音色设计引擎';
  const prompt = String(row.prompt || '').trim(), text = String(row.preview_text || '').trim(), length = Array.from(text).length;
  if (!prompt) return '请填写音色描述';
  if (engine === 'qwen-tts' && Array.from(prompt).length > 500) return 'Qwen 音色描述不能超过500字';
  if (engine === 'qwen-tts' && (length < 15 || length > 200)) return 'Qwen 试听文本须为15到200字';
  if (engine === 'minimax' && (!length || length > 500)) return 'MiniMax 试听文本须为1到500字';
  return '';
}

export async function comicMediaServerRequestId(action, key, { engine = '', targetId = '' } = {}, cryptoProvider = globalThis.crypto) {
  const input = action === 'voice_analyze' ? key : action === 'voice_design' ? `${engine}:${targetId}:${key}` : `${action}:${key}`;
  const digest = await cryptoProvider.subtle.digest('SHA-256', new TextEncoder().encode(input));
  const hex = Array.from(new Uint8Array(digest)).map(byte => byte.toString(16).padStart(2, '0')).join('');
  return `${action === 'voice_analyze' ? 'voice-analysis-' : action === 'voice_design' ? 'sp-voice-' : 'sp-attempt-'}${hex}`;
}
