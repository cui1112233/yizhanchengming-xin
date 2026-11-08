// Product update ("产品更新") helpers.
//
// This module is intentionally free of React so the release-note rules can be
// unit tested with node --test:
// - a note only shows once per releaseId;
// - missing localStorage degrades to "do not show" instead of throwing;
// - a 404 or an empty payload produces null so the modal stays silent.

export const PRODUCT_UPDATE_SEEN_KEY = 'qiantie:product-update-seen';

const SECTION_TYPES = ['feature', 'improvement', 'fix'];

/**
 * Accessing window.localStorage itself can throw (cookies disabled, sandboxed
 * iframe), not just getItem. Resolve it once behind a guard so every caller can
 * degrade to "do not show" without an uncaught error.
 */
export function resolveProductUpdateStorage() {
  try {
    if (typeof window === 'undefined') return null;
    return window.localStorage || null;
  } catch (_) {
    return null;
  }
}

function text(value) {
  return typeof value === 'string' ? value.trim() : '';
}

function normalizeTags(tags) {
  if (!Array.isArray(tags)) return [];
  return tags.map(tag => text(tag)).filter(Boolean);
}

export function normalizeProductUpdate(payload) {
  if (!payload || typeof payload !== 'object' || Array.isArray(payload)) return null;
  const releaseId = text(payload.releaseId);
  const title = text(payload.title);
  if (!releaseId || !title) return null;

  const sections = (Array.isArray(payload.sections) ? payload.sections : [])
    .map(section => {
      const type = text(section?.type);
      if (!SECTION_TYPES.includes(type)) return null;
      const items = (Array.isArray(section?.items) ? section.items : [])
        .map(item => ({
          title: text(item?.title),
          desc: text(item?.desc),
          tags: normalizeTags(item?.tags),
          link: text(item?.link)
        }))
        .filter(item => item.title);
      if (items.length === 0) return null;
      return { type, label: text(section?.label) || type, items };
    })
    .filter(Boolean);

  if (sections.length === 0) return null;

  return {
    releaseId,
    version: text(payload.version),
    releasedAt: text(payload.releasedAt),
    title,
    summary: text(payload.summary),
    sections
  };
}

/**
 * Returns the releaseId already acknowledged on this browser, or '' when the
 * storage is missing/blocked (private mode, disabled cookies).
 */
export function readSeenReleaseId(storage) {
  try {
    if (!storage) return '';
    return text(storage.getItem(PRODUCT_UPDATE_SEEN_KEY));
  } catch (_) {
    return '';
  }
}

/**
 * Persists the acknowledged releaseId. Returns false when storage is
 * unavailable; the caller must not surface an error to the user.
 */
export function markProductUpdateSeen(storage, releaseId) {
  const id = text(releaseId);
  if (!id) return false;
  try {
    if (!storage) return false;
    storage.setItem(PRODUCT_UPDATE_SEEN_KEY, id);
    return true;
  } catch (_) {
    return false;
  }
}

/**
 * True only for a usable releaseId that this browser has not acknowledged yet.
 * When localStorage itself is unavailable we never show the modal.
 */
export function shouldShowProductUpdate(releaseId, storage) {
  const id = text(releaseId);
  if (!id) return false;
  let seen;
  try {
    if (!storage) return false;
    seen = storage.getItem(PRODUCT_UPDATE_SEEN_KEY);
  } catch (_) {
    return false;
  }
  return text(seen) !== id;
}
