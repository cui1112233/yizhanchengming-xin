export function parseOpeningDraft(text) {
  let parsed;
  try { parsed = JSON.parse(text); } catch { throw new Error('拆解结果不是有效 JSON，请检查后再入库'); }
  const items = Array.isArray(parsed) ? parsed : [parsed];
  if (!items.length || items.some(item => !item || typeof item !== 'object' || Array.isArray(item) || !String(item.text || item.content || '').trim())) {
    throw new Error('每条开头词都需要填写 text 或 content，空结果不能入库');
  }
  return items.map(({ id, ...item }) => item);
}
