export const BATCH_FACTORY_ACTIVE_BATCH_STORAGE_KEY = 'qiantie:batch-factory:active-batch-id';

export function batchFactoryNovelFetchHandoffPath(intakeId) {
  const normalized = String(intakeId || '').trim();
  return normalized ? `/shuihuo-production?intake=${encodeURIComponent(normalized)}` : '/shuihuo-production';
}

export function pendingNovelFetchIntakeId(search = '') {
  return String(new URLSearchParams(search).get('intake') || '').trim();
}

export function novelFetchIntakeBooks(intake) {
  if (Array.isArray(intake?.books)) return intake.books;
  const payload = intake?.payload;
  if (payload && typeof payload === 'object' && Array.isArray(payload.books)) return payload.books;
  if (typeof payload !== 'string' || !payload.trim()) return [];
  try {
    const parsed = JSON.parse(payload);
    return Array.isArray(parsed?.books) ? parsed.books : [];
  } catch (_) {
    return [];
  }
}
