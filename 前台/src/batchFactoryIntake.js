export const BATCH_FACTORY_PLATFORMS = Object.freeze([
  { id: '2', name: '番茄付费' },
  { id: '7', name: '番茄免费' },
  { id: '3', name: '七猫付费' },
  { id: '4', name: '点众付费' },
  { id: '1', name: '黑岩付费' },
  { id: '6', name: '阅文付费' },
  { id: '15', name: '知乎付费' },
  { id: '20', name: '掌阅付费' },
  { id: '26', name: '卓越付费' },
  { id: '29', name: '九州书城' },
  { id: '31', name: '掌文付费' },
]);

function text(value) {
  return String(value ?? '').trim();
}

function positiveInt(value, fallback = 4000) {
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
}

function canonicalGender(value) {
  const normalized = text(value).toLowerCase();
  if (!normalized) return '';
  if (['male', '男', '男频', '男生', '男性', '男向'].includes(normalized)) return 'male';
  if (['female', '女', '女频', '女生', '女性', '女向'].includes(normalized)) return 'female';
  throw new Error('男女频只能选择自动判断、男频或女频');
}

export function parseBookIds(raw) {
  const seen = new Set();
  const result = [];
  for (const token of String(raw ?? '').split(/[\s,，;；]+/u)) {
    const id = token.trim();
    if (!id || seen.has(id)) continue;
    seen.add(id);
    result.push(id);
  }
  return result;
}

export function addBookstoreGroup(groups, input = {}) {
  const current = Array.isArray(groups) ? groups : [];
  const platformId = text(input.platformId);
  const platformName = text(input.platformName);
  if (!platformId || !platformName) throw new Error('请选择书城');

  const ids = parseBookIds(input.rawBookIds);
  if (!ids.length) throw new Error('至少输入一个书籍ID');

  const existing = new Set(
    current
      .filter(group => text(group?.platform_id) === platformId)
      .flatMap(group => Array.isArray(group?.books) ? group.books : [])
      .map(book => text(book?.book_id))
      .filter(Boolean),
  );
  for (const id of ids) {
    if (existing.has(id)) {
      throw new Error(`书籍ID ${id} 已存在于 ${platformName} 分组中`);
    }
  }

  const manualGender = canonicalGender(input.manualGender);
  const style = text(input.style);
  const group = {
    platform_id: platformId,
    platform_name: platformName,
    max_txt: positiveInt(input.maxTxt),
    books: ids.map(bookId => ({
      book_id: bookId,
      manual_gender: manualGender,
      style,
    })),
  };
  return [...current, group];
}

export function groupLabel(group) {
  const name = text(group?.platform_name) || '未命名书城';
  const count = Array.isArray(group?.books) ? group.books.length : 0;
  return `${name}${count}本`;
}

export function buildIntakePayload({ title = '', groups = [] } = {}) {
  if (!Array.isArray(groups) || groups.length === 0) {
    throw new Error('请先添加至少一个书城');
  }
  return {
    title: text(title),
    groups,
  };
}

async function readError(response, fallback) {
  try {
    const body = await response.text();
    if (body && body.trim()) return body.trim();
  } catch {
    // Ignore secondary error while reading an error response.
  }
  return fallback;
}

async function postJSON(fetchImpl, url, body) {
  const response = await fetchImpl(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!response?.ok) {
    throw new Error(await readError(response, `${url} 请求失败`));
  }
  return response.json();
}

export async function submitIntakeAndStartJob({
  fetchImpl = globalThis.fetch,
  title = '',
  groups = [],
  runAt = '',
} = {}) {
  if (typeof fetchImpl !== 'function') throw new Error('fetch is required');
  const intake = await postJSON(
    fetchImpl,
    '/api/batch-factory/intakes',
    buildIntakePayload({ title, groups }),
  );
  const intakeId = text(intake?.intake_id);
  if (!intakeId) throw new Error('创建 Intake 成功但未返回 intake_id');

  const jobPayload = { intake_id: intakeId };
  if (text(runAt)) jobPayload.run_at = text(runAt);
  const job = await postJSON(fetchImpl, '/api/batch-factory/jobs', jobPayload);
  return { intake, job };
}
