import { normalizeScriptConstraints } from './scriptConstraints.js';
import { normalizeExtractInfo } from './scriptEntities.js';
import { normalizeExtractionPresetId } from './scriptExtractionPresets.js';
import { normalizeShotVideoTaskHistory } from './scriptShotVideoTasks.js';

const storagePrefix = 'qiantie:script-draft:';
const tabStorageKey = 'qiantie:script-draft-tab-id';
const draftVersion = 4;

function usernameKey(username) {
  return `${storagePrefix}${encodeURIComponent(String(username || 'guest'))}`;
}

function tabDraftKey(username, tabId) {
  return `${usernameKey(username)}:${encodeURIComponent(String(tabId || ''))}`;
}

function createFallbackTabId() {
  return `tab-${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`;
}

export function getScriptDraftTabId(sessionStorageLike, random = () => crypto.randomUUID()) {
  try {
    const existing = sessionStorageLike?.getItem(tabStorageKey);
    if (typeof existing === 'string' && existing) return existing;
  } catch {}

  let tabId;
  try {
    tabId = String(random() || '');
  } catch {
    tabId = '';
  }
  tabId = tabId || createFallbackTabId();

  try {
    sessionStorageLike?.setItem(tabStorageKey, tabId);
  } catch {}
  return tabId;
}

function normalizeDraft(draft) {
  if (!draft?.values || typeof draft.values !== 'object') return null;
  if (![1, 2, 3, draftVersion].includes(draft.version)) return null;
  const constraints = normalizeScriptConstraints(draft.constraints);
  // v3 drafts predate the base-setup switch; preserve their stored shape and
  // let the current page defaults add the new layer when it is needed.
  if (draft.version <= 3 && !Object.hasOwn(draft.constraints || {}, 'baseSetup')) {
    delete constraints.baseSetup;
  }
  const shotVideoTasks = {};
  if (draft.shotVideoTasks && typeof draft.shotVideoTasks === 'object' && !Array.isArray(draft.shotVideoTasks)) {
    for (const [index, task] of Object.entries(draft.shotVideoTasks)) {
      const taskId = typeof task?.taskId === 'string' ? task.taskId.trim() : '';
      const status = ['processing', 'succeeded', 'failed'].includes(task?.status) ? task.status : '';
      if (!/^\d+$/.test(index) || !taskId || !status) continue;
      const entry = { taskId, status };
      if (typeof task.prompt === 'string' && task.prompt.trim()) entry.prompt = task.prompt.trim().slice(0, 20000);
      if (typeof task.error === 'string' && task.error.trim()) entry.error = task.error.trim();
      if (typeof task.videoUrl === 'string' && /^(?:https?:\/\/|\/api\/)/i.test(task.videoUrl.trim())) entry.videoUrl = task.videoUrl.trim();
      shotVideoTasks[index] = entry;
    }
  }
  return {
    ...draft,
    version: draftVersion,
    values: {
      ...draft.values,
      extractionPreset: normalizeExtractionPresetId(draft.values.extractionPreset)
    },
    extractInfo: normalizeExtractInfo(draft.extractInfo),
    constraints,
    shotVideoTasks,
    videoTaskHistory: normalizeShotVideoTaskHistory(draft.videoTaskHistory)
  };
}

export function loadScriptDraft(storage, username, tabId) {
  if (!tabId) return null;

  let tabRaw;
  try {
    tabRaw = storage?.getItem(tabDraftKey(username, tabId));
  } catch {
    return null;
  }
  if (tabRaw !== null && tabRaw !== undefined) {
    try {
      return normalizeDraft(JSON.parse(tabRaw));
    } catch {
      return null;
    }
  }

  let legacyRaw;
  try {
    legacyRaw = storage?.getItem(usernameKey(username));
  } catch {
    return null;
  }
  if (!legacyRaw) return null;

  try {
    const legacyDraft = normalizeDraft(JSON.parse(legacyRaw));
    if (!legacyDraft) return null;
    saveScriptDraft(storage, username, tabId, legacyDraft);
    return legacyDraft;
  } catch {
    return null;
  }
}

export function storageFailureReason(error) {
  const name = String(error?.name || '');
  if (name === 'QuotaExceededError' || name === 'NS_ERROR_DOM_QUOTA_REACHED') return '本地存储空间已满';
  if (name === 'SecurityError') return '浏览器已禁用本地存储（隐私模式）';
  return String(error?.message || '浏览器本地存储不可用');
}

export function saveScriptDraft(storage, username, tabId, draft) {
  if (!tabId) return { ok: false, reason: '缺少草稿标识，无法保存' };
  if (!storage) return { ok: false, reason: '浏览器本地存储不可用' };
  try {
    const normalized = normalizeDraft({ ...draft, version: draftVersion });
    if (!normalized) return { ok: false, reason: '草稿内容不完整，无法保存' };
    const savedAt = new Date();
    storage.setItem(tabDraftKey(username, tabId), JSON.stringify({
      ...normalized,
      savedAt: savedAt.toISOString()
    }));
    return { ok: true, savedAt };
  } catch (error) {
    return { ok: false, reason: storageFailureReason(error) };
  }
}
